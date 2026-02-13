/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"

	aegisproto "github.com/yourorg/aegis/proto/aegis/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	aegisv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/cpclient"
	builders "github.com/yourorg/aegis/agents/k8s-agent/internal/workload/builders"
	workdiscovery "github.com/yourorg/aegis/agents/k8s-agent/internal/workload/discovery"
	workstatus "github.com/yourorg/aegis/agents/k8s-agent/internal/workload/status"
	workspacecfg "github.com/yourorg/aegis/pkg/workspace"
)

const (
	backendWorkspace = "workspace"
	backendPyTorch   = "pytorch_v1"
	backendJob       = "job_v1"

	requeuePending               = 10 * time.Second
	workspaceJobTTLSeconds int32 = 600

	labelWorkloadID = "aegis.workload/id"
	labelSSHManaged = "aegis.yourorg.dev/ssh-managed"

	// Legacy annotations (kept for backward compatibility during rollout)
	annotationStartAcked              = "aegis.yourorg.dev/start-acked"
	annotationFinalAcked              = "aegis.yourorg.dev/final-acked"
	annotationMaxDurationSeconds      = "aegis.yourorg.dev/maxDurationSeconds"
	annotationTTLSecondsAfterFinished = "aegis.yourorg.dev/ttlSecondsAfterFinished"

	// Event-driven state tracking annotations
	// These track the last-pushed state to platform-api, enabling proper re-sync when jobs are recreated
	annotationLastPushedState  = "aegis.yourorg.dev/last-pushed-state"  // PENDING, RUNNING, SUCCEEDED, FAILED
	annotationLastPushedJobUID = "aegis.yourorg.dev/last-pushed-job-uid" // UID of the job when state was pushed

	annotationSSHAuthorizedKeys = "aegis.yourorg.dev/ssh-authorized-keys"
	annotationSSHTrustedCA      = "aegis.yourorg.dev/ssh-trusted-user-ca"
)

// AegisWorkloadReconciler reconciles a AegisWorkload object.
type AegisWorkloadReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	Dynamic dynamic.Interface
	Config  *rest.Config

	cpClient  *cpclient.Client
	clusterID string

	defaultWorkspaceImage string
	defaultTrainingImage  string
	gpuResourceOverride   string
	dryRun                bool
	kueueEnabled          bool
	kueueQueue            string
	proxyServiceName      string
	proxyServicePort      int32
	proxyIngressHost      string
	proxyURL              string // Full URL for heartbeat reporting (e.g., wss://host:port)
	sshBootstrapImage     string

	workspaceEnvDefaults map[string]string

	pyTorchAPIVersion string
	pyTorchGVR        *schema.GroupVersionResource
}

// +kubebuilder:rbac:groups=aegis.yourorg.dev,resources=aegisworkloads,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aegis.yourorg.dev,resources=aegisworkloads/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=aegis.yourorg.dev,resources=aegisworkloads/finalizers,verbs=update
// +kubebuilder:rbac:groups=batch,resources=jobs;jobs/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=kubeflow.org,resources=pytorchjobs;pytorchjobs/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=training.kubeflow.org,resources=trainjobs;trainjobs/status,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete

// Reconcile ensures the workload CR reflects the lifecycle of the underlying compute object.
func (r *AegisWorkloadReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithValues("aegisworkload", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	var aw aegisv1alpha1.AegisWorkload
	if err := r.Get(ctx, req.NamespacedName, &aw); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !aw.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if aw.Status.Phase == "" {
		if err := r.patchStatus(ctx, &aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.Phase = aegisv1alpha1.PhasePending
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	switch {
	case aw.Spec.Workspace != nil:
		return r.reconcileWorkspace(ctx, &aw)
	case aw.Spec.Training != nil:
		return r.reconcileTraining(ctx, &aw)
	default:
		log.Info("no workspace or training spec; skipping")
		return ctrl.Result{}, nil
	}
}

func (r *AegisWorkloadReconciler) reconcileWorkspace(ctx context.Context, aw *aegisv1alpha1.AegisWorkload) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithValues("mode", "workspace")
	spec := aw.Spec.Workspace
	image := spec.Image
	if image == "" {
		image = r.defaultWorkspaceImage
	}

	flavor := spec.Flavor
	workspaceInteractive := spec != nil && spec.Interactive
	ports := effectiveWorkspacePorts(spec)
	var (
		sshSecretName   string
		secretAvailable bool
	)
	if workspaceInteractive {
		if name, exists, mutated, err := r.ensureWorkspaceSSHSecret(ctx, aw); err != nil {
			return ctrl.Result{}, err
		} else {
			sshSecretName = name
			secretAvailable = exists
			if mutated {
				log.V(1).Info("ssh connection secret synced", "secret", name)
			}
		}
	} else {
		if changed, err := r.cleanupWorkspaceSSHSecret(ctx, aw); err != nil {
			return ctrl.Result{}, err
		} else if changed {
			log.V(1).Info("ssh connection secret removed")
		}
	}

	jobName := builders.SanitizeName(aw.Name)
	jobKey := types.NamespacedName{Name: jobName, Namespace: aw.Namespace}

	maxDeadline, mdErr := maxDurationFromAnnotation(aw)
	if mdErr != nil {
		log.Error(mdErr, "invalid max duration annotation", "annotation", aw.GetAnnotations()[annotationMaxDurationSeconds])
	}
	ttlAfterFinished := ptr.To(workspaceJobTTLSeconds)
	if ttl, err := ttlSecondsAfterFinishedFromAnnotation(aw); err != nil {
		log.Error(err, "invalid ttl annotation", "annotation", aw.GetAnnotations()[annotationTTLSecondsAfterFinished])
	} else if ttl != nil {
		ttlAfterFinished = ttl
	}

	var job batchv1.Job
	err := r.Get(ctx, jobKey, &job)
	if apierrors.IsNotFound(err) {
		mergedEnv := workspacecfg.MergeEnv(spec.Env, r.workspaceEnvDefaults)
		opts := builders.WorkspaceOptions{
			Namespace:               aw.Namespace,
			JobName:                 jobName,
			WorkloadID:              aw.Name,
			Image:                   image,
			Command:                 spec.Command,
			DefaultCommand:          workspaceDefaultCommand(image),
			Env:                     mergedEnv,
			Flavor:                  flavor,
			Queue:                   aw.Spec.Queue,
			Hints:                   convertSpecHints(aw.Spec.Hints),
			DryRun:                  r.dryRun,
			KueueEnabled:            r.kueueEnabled,
			KueueQueue:              r.kueueQueue,
			GPUResourceOverride:     r.gpuResourceOverride,
			Interactive:             workspaceInteractive,
			InteractivePorts:        ports,
			SSHBootstrapImage:       r.sshBootstrapImage,
			ActiveDeadlineSeconds:   maxDeadline,
			TTLSecondsAfterFinished: ttlAfterFinished,
		}
		if secretAvailable {
			opts.SSHSecretName = sshSecretName
		}
		obj := builders.BuildWorkspaceJob(opts)
		if err := controllerutil.SetControllerReference(aw, obj, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, obj); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("workspace job created", "job", jobName)
		r.Recorder.Eventf(aw, corev1.EventTypeNormal, "Submitted", "Workspace job %s created", jobName)
		if err := r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.Phase = aegisv1alpha1.PhaseSubmitted
			st.Backend = backendWorkspace
			st.JobRef = &aegisv1alpha1.JobRef{APIVersion: "batch/v1", Kind: "Job", Name: jobName, Namespace: aw.Namespace}
			st.URL = fmt.Sprintf("k8s://%s/job/%s", aw.Namespace, jobName)
		}); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: requeuePending}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	updated := false
	if workspaceInteractive {
		if changed, err := r.ensureInteractiveResources(ctx, aw); err != nil {
			return ctrl.Result{}, err
		} else if changed {
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}
	} else {
		if changed, err := r.cleanupInteractiveResources(ctx, aw); err != nil {
			return ctrl.Result{}, err
		} else if changed {
			return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
		}
	}

	if r.kueueEnabled {
		if ptr.Deref(job.Spec.Suspend, false) {
			if aw.Status.Message != "Queued by Kueue" {
				r.Recorder.Eventf(aw, corev1.EventTypeNormal, "QueuedByKueue",
					"Job %s queued by Kueue (queue=%q)", jobName, job.Labels["kueue.x-k8s.io/queue-name"])
				if err := r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
					st.Message = "Queued by Kueue"
				}); err != nil {
					return ctrl.Result{}, err
				}
			}
		} else if aw.Status.Message == "Queued by Kueue" {
			r.Recorder.Eventf(aw, corev1.EventTypeNormal, "AdmittedByKueue",
				"Job %s admitted by Kueue", jobName)
			if err := r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
				st.Message = ""
			}); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		if job.Spec.Suspend == nil || ptr.Deref(job.Spec.Suspend, false) {
			job.Spec.Suspend = ptr.To(false)
			updated = true
		}
		if job.Labels != nil {
			if _, ok := job.Labels["kueue.x-k8s.io/queue-name"]; ok {
				delete(job.Labels, "kueue.x-k8s.io/queue-name")
				updated = true
			}
		}
		if job.Spec.Template.Labels != nil {
			if _, ok := job.Spec.Template.Labels["kueue.x-k8s.io/queue-name"]; ok {
				delete(job.Spec.Template.Labels, "kueue.x-k8s.io/queue-name")
				updated = true
			}
		}
	}
	if maxDeadline != nil {
		current := ptr.Deref(job.Spec.ActiveDeadlineSeconds, int64(0))
		if job.Spec.ActiveDeadlineSeconds == nil || current != *maxDeadline {
			job.Spec.ActiveDeadlineSeconds = ptr.To(*maxDeadline)
			updated = true
		}
	}
	desiredTTL := ptr.Deref(ttlAfterFinished, workspaceJobTTLSeconds)
	if job.Spec.TTLSecondsAfterFinished == nil || ptr.Deref(job.Spec.TTLSecondsAfterFinished, int32(0)) != desiredTTL {
		job.Spec.TTLSecondsAfterFinished = ptr.To(desiredTTL)
		updated = true
	}
	if updated {
		if err := r.Update(ctx, &job); err != nil {
			return ctrl.Result{}, err
		}
		log.Info("workspace job spec updated", "job", jobName)
		return ctrl.Result{RequeueAfter: 2 * time.Second}, nil
	}

	if aw.Status.JobRef == nil {
		if err := r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.JobRef = &aegisv1alpha1.JobRef{APIVersion: "batch/v1", Kind: "Job", Name: jobName, Namespace: aw.Namespace}
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	r.applyJobTransitions(ctx, aw, &job, backendWorkspace)

	return ctrl.Result{RequeueAfter: requeuePending}, nil
}

func (r *AegisWorkloadReconciler) applyJobTransitions(ctx context.Context, aw *aegisv1alpha1.AegisWorkload, job *batchv1.Job, backend string) {
	if job == nil {
		return
	}

	// Get the job UID for state tracking - this enables proper re-sync when jobs are recreated
	jobUID := string(job.UID)

	phase := aw.Status.Phase
	// Only mark as running when pods are actually Ready, not just Active (which includes Pending pods).
	// This ensures the UI shows accurate status - users shouldn't see "Running" until the container
	// is truly running and ready (node provisioned, image pulled, container started).
	podsReady := r.isJobPodReady(ctx, job)
	observedRunning := (job.Status.Active > 0 && podsReady) || job.Status.Succeeded > 0 || job.Status.Failed > 0

	// Check if this is a recreated job (different UID than what we last pushed)
	lastPushedJobUID := aw.GetAnnotations()[annotationLastPushedJobUID]
	jobRecreated := lastPushedJobUID != "" && lastPushedJobUID != jobUID

	// If job was recreated, we need to re-push status even if phase hasn't changed
	if jobRecreated {
		ctrl.LoggerFrom(ctx).Info("detected job recreation, will re-sync status",
			"workload", aw.Name, "oldJobUID", lastPushedJobUID, "newJobUID", jobUID)
	}

	if observedRunning && (phase != aegisv1alpha1.PhaseRunning || jobRecreated) {
		prev := aw.Status.Phase
		_ = r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.Phase = aegisv1alpha1.PhaseRunning
			st.Backend = backend
			// If job was recreated, reset the start time and clear completion time
			if jobRecreated {
				st.CompletionTime = nil
				st.Message = ""
			}
			if st.StartTime == nil || jobRecreated {
				if job.Status.StartTime != nil {
					st.StartTime = job.Status.StartTime.DeepCopy()
				} else {
					now := metav1.NewTime(time.Now())
					st.StartTime = &now
				}
			}
		})
		if prev != aegisv1alpha1.PhaseRunning || jobRecreated {
			r.startWorkloadBridge(ctx, aw, jobUID)
		}
	}

	if workstatus.IsJobComplete(job) {
		prev := aw.Status.Phase
		_ = r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.Phase = aegisv1alpha1.PhaseSucceeded
			st.Backend = backend
			if st.CompletionTime == nil {
				if job.Status.CompletionTime != nil {
					st.CompletionTime = job.Status.CompletionTime.DeepCopy()
				} else {
					now := metav1.NewTime(time.Now())
					st.CompletionTime = &now
				}
			}
			st.Message = ""
		})
		if prev != aegisv1alpha1.PhaseSucceeded || jobRecreated {
			r.ackWorkloadBridge(ctx, aw, "SUCCEEDED", jobUID)
		}
		return
	}

	if workstatus.IsJobFailed(job) {
		failureMsg := deriveJobFailureMessage(job)
		prev := aw.Status.Phase
		_ = r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.Phase = aegisv1alpha1.PhaseFailed
			st.Backend = backend
			st.Message = failureMsg
			if st.CompletionTime == nil {
				if job.Status.CompletionTime != nil {
					st.CompletionTime = job.Status.CompletionTime.DeepCopy()
				} else {
					now := metav1.NewTime(time.Now())
					st.CompletionTime = &now
				}
			}
		})
		if prev != aegisv1alpha1.PhaseFailed || jobRecreated {
			r.ackWorkloadBridge(ctx, aw, "FAILED", jobUID)
		}
	}
}

// isJobPodReady checks if at least one pod owned by the job is in Ready condition.
// This ensures we only report Running status when the container is actually running,
// not when the pod is still Pending (waiting for node, pulling image, etc.).
func (r *AegisWorkloadReconciler) isJobPodReady(ctx context.Context, job *batchv1.Job) bool {
	if job == nil {
		return false
	}

	var pods corev1.PodList
	if err := r.List(ctx, &pods,
		client.InNamespace(job.Namespace),
		client.MatchingLabels{"job-name": job.Name}); err != nil {
		ctrl.LoggerFrom(ctx).V(1).Info("failed to list job pods for readiness check", "job", job.Name, "error", err)
		return false
	}

	for _, pod := range pods.Items {
		for _, cond := range pod.Status.Conditions {
			if cond.Type == corev1.PodReady && cond.Status == corev1.ConditionTrue {
				return true
			}
		}
	}
	return false
}

func (r *AegisWorkloadReconciler) applyTrainingTransitions(ctx context.Context, aw *aegisv1alpha1.AegisWorkload, obj *unstructured.Unstructured) {
	if obj == nil {
		return
	}

	// Get the training object UID for state tracking
	objUID := string(obj.GetUID())

	// Check if this is a recreated training object
	lastPushedJobUID := aw.GetAnnotations()[annotationLastPushedJobUID]
	objRecreated := lastPushedJobUID != "" && lastPushedJobUID != objUID

	if workstatus.HasCondition(obj, "Running") && (aw.Status.Phase != aegisv1alpha1.PhaseRunning || objRecreated) {
		prev := aw.Status.Phase
		_ = r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.Phase = aegisv1alpha1.PhaseRunning
			st.Backend = backendPyTorch
			if objRecreated {
				st.CompletionTime = nil
				st.Message = ""
			}
			if st.StartTime == nil || objRecreated {
				now := metav1.NewTime(time.Now())
				st.StartTime = &now
			}
		})
		if prev != aegisv1alpha1.PhaseRunning || objRecreated {
			r.startWorkloadBridge(ctx, aw, objUID)
		}
	}

	if workstatus.HasAnySuccess(obj) {
		prev := aw.Status.Phase
		_ = r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.Phase = aegisv1alpha1.PhaseSucceeded
			st.Backend = backendPyTorch
			if st.CompletionTime == nil {
				now := metav1.NewTime(time.Now())
				st.CompletionTime = &now
			}
		})
		if prev != aegisv1alpha1.PhaseSucceeded || objRecreated {
			r.ackWorkloadBridge(ctx, aw, "SUCCEEDED", objUID)
		}
		return
	}

	if workstatus.HasAnyFailure(obj) {
		prev := aw.Status.Phase
		_ = r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.Phase = aegisv1alpha1.PhaseFailed
			st.Backend = backendPyTorch
			if st.CompletionTime == nil {
				now := metav1.NewTime(time.Now())
				st.CompletionTime = &now
			}
		})
		if prev != aegisv1alpha1.PhaseFailed || objRecreated {
			r.ackWorkloadBridge(ctx, aw, "FAILED", objUID)
		}
	}
}

func (r *AegisWorkloadReconciler) startClusterPresence(ctx context.Context) {
	if r.cpClient == nil || r.clusterID == "" {
		return
	}
	zapLogger, err := zap.NewProduction()
	if err != nil {
		ctrl.Log.WithName("cluster-presence").Error(err, "failed to initialise zap logger")
		return
	}
	defer func() { _ = zapLogger.Sync() }()

	// Build labels map for cluster registration
	labels := map[string]string{}
	if projectID := os.Getenv("AEGIS_PROJECT_ID"); projectID != "" {
		labels["aegis.yourorg.dev/projectId"] = projectID
	}
	if ilLevel := os.Getenv("AEGIS_IL_LEVEL"); ilLevel != "" {
		labels["aegis.yourorg.dev/ilLevel"] = ilLevel
	}
	// Add any custom labels from environment (format: KEY1=VAL1,KEY2=VAL2)
	if customLabels := os.Getenv("AEGIS_CLUSTER_LABELS"); customLabels != "" {
		for _, pair := range strings.Split(customLabels, ",") {
			if kv := strings.SplitN(strings.TrimSpace(pair), "=", 2); len(kv) == 2 {
				labels[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}

	// Build proxy URL before registration so it's included in the initial request
	proxyURL := r.buildProxyURL()

	registerReq := &aegisproto.ClusterRegisterRequest{
		ClusterId: r.clusterID,
		Provider:  os.Getenv("AEGIS_PROVIDER"),
		Region:    os.Getenv("AEGIS_REGION"),
		IlLevel:   os.Getenv("AEGIS_IL_LEVEL"),
		Labels:    labels,
		ProxyUrl:  proxyURL, // Include proxy URL in registration for immediate availability
	}

	for {
		if err := r.cpClient.Register(ctx, registerReq); err != nil {
			zapLogger.Warn("cluster registration failed", zap.String("cluster_id", r.clusterID), zap.Error(err))
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
			continue
		}
		zapLogger.Info("cluster registered", zap.String("cluster_id", r.clusterID), zap.String("proxy_url", proxyURL))
		break
	}

	flavors := r.discoverFlavors()
	r.cpClient.HeartbeatLoop(ctx, zapLogger, r.clusterID, flavors, proxyURL)
}

// buildProxyURL returns the spoke proxy URL for heartbeat reporting.
// If AEGIS_PROXY_URL is set, it's used directly. Otherwise, constructs from AEGIS_PROXY_INGRESS_HOST.
// If neither is set, tries to auto-discover the LoadBalancer hostname from the proxy Service.
// Falls back to node public IP with nip.io for NodePort access.
func (r *AegisWorkloadReconciler) buildProxyURL() string {
	// Prefer explicit proxy URL if set (supports custom port for NodePort access)
	if r.proxyURL != "" {
		url := r.proxyURL
		// Ensure scheme is present
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "wss://") && !strings.HasPrefix(url, "ws://") {
			url = "wss://" + url
		}
		return url
	}
	// Fall back to constructing from ingress host
	host := r.proxyIngressHost
	if host == "" {
		// Try to discover LoadBalancer hostname from the proxy Service (NLB)
		if lbHost := r.discoverLoadBalancerHost(); lbHost != "" {
			host = lbHost
		} else if ip := discoverNodePublicIP(); ip != "" {
			// Fallback: Auto-discover node public IP for AWS/cloud deployments
			// Construct nip.io URL with default NodePort
			host = fmt.Sprintf("spoke-proxy.%s.nip.io:31484", ip)
		}
	}
	if host == "" {
		return ""
	}
	// If no scheme specified, default to wss://
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") && !strings.HasPrefix(host, "wss://") && !strings.HasPrefix(host, "ws://") {
		host = "wss://" + host
	}
	return host
}

// discoverLoadBalancerHost queries the aegis-spoke-proxy Service and returns the LoadBalancer
// hostname or IP if available. Returns empty string if not a LoadBalancer or not yet provisioned.
func (r *AegisWorkloadReconciler) discoverLoadBalancerHost() string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Try common service names and namespaces for the spoke proxy
	serviceNames := []string{"aegis-spoke-proxy", "spoke-proxy"}
	namespaces := []string{"aegis-system", "default"}

	for _, ns := range namespaces {
		for _, name := range serviceNames {
			var svc corev1.Service
			if err := r.Client.Get(ctx, types.NamespacedName{Name: name, Namespace: ns}, &svc); err != nil {
				continue
			}
			if svc.Spec.Type != corev1.ServiceTypeLoadBalancer {
				continue
			}
			// Check if LoadBalancer has been provisioned
			for _, ingress := range svc.Status.LoadBalancer.Ingress {
				if ingress.Hostname != "" {
					// AWS NLB uses hostname
					port := int32(31484) // default proxy port
					for _, p := range svc.Spec.Ports {
						if p.Name == "proxy" || p.Name == "https" || p.Port == 8443 || p.Port == 31484 {
							port = p.Port
							break
						}
					}
					return fmt.Sprintf("%s:%d", ingress.Hostname, port)
				}
				if ingress.IP != "" {
					// GCP/Azure use IP
					port := int32(31484)
					for _, p := range svc.Spec.Ports {
						if p.Name == "proxy" || p.Name == "https" || p.Port == 8443 || p.Port == 31484 {
							port = p.Port
							break
						}
					}
					return fmt.Sprintf("%s:%d", ingress.IP, port)
				}
			}
		}
	}
	return ""
}

// discoverNodePublicIP attempts to discover the node's public IP address.
// Tries AWS EC2 metadata service first, then falls back to external IP check services.
func discoverNodePublicIP() string {
	// Try AWS EC2 metadata service (IMDSv1 for simplicity)
	if ip := fetchURL("http://169.254.169.254/latest/meta-data/public-ipv4", 2*time.Second); ip != "" {
		return ip
	}
	// Fallback: try common external IP check services
	services := []string{
		"https://api.ipify.org",
		"https://checkip.amazonaws.com",
		"https://ifconfig.me/ip",
	}
	for _, svc := range services {
		if ip := fetchURL(svc, 3*time.Second); ip != "" {
			return ip
		}
	}
	return ""
}

// fetchURL fetches a URL and returns the trimmed body, or empty string on error.
func fetchURL(url string, timeout time.Duration) string {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(body))
}

func (r *AegisWorkloadReconciler) discoverFlavors() []*aegisproto.Flavor {
	raw := strings.Split(os.Getenv("AEGIS_FLAVORS"), ",")
	var flavors []*aegisproto.Flavor
	for _, token := range raw {
		name := strings.TrimSpace(token)
		if name == "" {
			continue
		}
		flavors = append(flavors, &aegisproto.Flavor{Name: name})
	}
	if len(flavors) == 0 {
		flavors = []*aegisproto.Flavor{
			{Name: "a10-mig-1g"},
			{Name: "a100-8x"},
		}
	}
	return flavors
}

func (r *AegisWorkloadReconciler) reconcileTraining(ctx context.Context, aw *aegisv1alpha1.AegisWorkload) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithValues("mode", "training")
	spec := aw.Spec.Training
	image := spec.Image
	if image == "" {
		image = r.defaultTrainingImage
	}

	workers := int(spec.Workers)
	if workers <= 0 {
		workers = 1
	}

	name := builders.SanitizeName(aw.Name)

	// Use simple Kubernetes Jobs instead of PyTorchJob
	var existingJob batchv1.Job
	err := r.Get(ctx, types.NamespacedName{Name: name, Namespace: aw.Namespace}, &existingJob)
	if apierrors.IsNotFound(err) {
		// Create new Job
		job := builders.BuildTrainingJob(builders.TrainingOptions{
			Namespace: aw.Namespace,
			Name:      name,
			Image:     image,
			Command:   spec.Command,
			Flavor:    spec.Flavor,
			Workers:   workers,
			DryRun:    r.dryRun,
			Hints:     convertSpecHints(aw.Spec.Hints),
		})

		if err := controllerutil.SetControllerReference(aw, job, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if err := r.Create(ctx, job); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("training job created", "name", name)
		r.Recorder.Eventf(aw, corev1.EventTypeNormal, "Submitted", "Training Job %s created", name)

		if err := r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.Phase = aegisv1alpha1.PhaseSubmitted
			st.Backend = backendJob
			st.JobRef = &aegisv1alpha1.JobRef{APIVersion: "batch/v1", Kind: "Job", Name: name, Namespace: aw.Namespace}
			st.URL = fmt.Sprintf("k8s://%s/job/%s", aw.Namespace, name)
		}); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: requeuePending}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// Job exists, check status
	if aw.Status.JobRef == nil {
		if err := r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.JobRef = &aegisv1alpha1.JobRef{APIVersion: "batch/v1", Kind: "Job", Name: name, Namespace: aw.Namespace}
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Apply status transitions based on Job state
	r.applyJobTransitions(ctx, aw, &existingJob, backendJob)

	return ctrl.Result{RequeueAfter: requeuePending}, nil
}

func (r *AegisWorkloadReconciler) ensureTrainingDiscovery() error {
	if r.pyTorchGVR == nil {
		gv, gvr, err := workdiscovery.DetectPyTorchGVR(r.Config)
		if err != nil {
			return err
		}
		r.pyTorchAPIVersion = gv
		r.pyTorchGVR = &gvr
	}
	return nil
}

func (r *AegisWorkloadReconciler) patchStatus(ctx context.Context, aw *aegisv1alpha1.AegisWorkload, mutate func(*aegisv1alpha1.AegisWorkloadStatus)) error {
	original := aw.DeepCopy()
	mutate(&aw.Status)
	return r.Status().Patch(ctx, aw, client.MergeFrom(original))
}

func controlPlaneWorkloadID(name string) string {
	id := strings.TrimSpace(name)
	for strings.HasPrefix(id, "aegis-") {
		id = strings.TrimPrefix(id, "aegis-")
	}
	return id
}

func workspaceAliasName(name string) string {
	cpID := controlPlaneWorkloadID(name)
	if cpID == "" {
		cpID = strings.TrimSpace(name)
	}
	return fmt.Sprintf("aegis-w-%s", cpID)
}

func (r *AegisWorkloadReconciler) startWorkloadBridge(ctx context.Context, aw *aegisv1alpha1.AegisWorkload, jobUID string) {
	if r.cpClient == nil || aw == nil {
		return
	}
	if r.clusterID == "" {
		ctrl.LoggerFrom(ctx).Info("cluster id missing; skipping StartWorkload bridge", "workload", aw.Name)
		return
	}

	annotations := aw.GetAnnotations()
	lastPushedState := annotations[annotationLastPushedState]
	lastPushedJobUID := annotations[annotationLastPushedJobUID]

	// Only push if: state changed to RUNNING, OR job was recreated (different UID)
	stateChanged := lastPushedState != "RUNNING"
	jobRecreated := jobUID != "" && lastPushedJobUID != "" && lastPushedJobUID != jobUID

	if !stateChanged && !jobRecreated {
		// Already pushed RUNNING for this job
		return
	}

	cpID := controlPlaneWorkloadID(aw.Name)
	if err := r.cpClient.Start(ctx, cpID, r.clusterID); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "StartWorkload bridge failed", "workload", aw.Name, "cpWorkloadID", cpID, "clusterID", r.clusterID)
		return
	}

	// Update annotations to track what we just pushed
	if err := r.updatePushedStateAnnotations(ctx, aw, "RUNNING", jobUID); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed to persist state tracking", "workload", aw.Name)
		return
	}

	if jobRecreated {
		ctrl.LoggerFrom(ctx).Info("StartWorkload bridge sent (job recreated)", "workload", aw.Name, "cpWorkloadID", cpID, "clusterID", r.clusterID, "oldJobUID", lastPushedJobUID, "newJobUID", jobUID)
	} else {
		ctrl.LoggerFrom(ctx).Info("StartWorkload bridge sent", "workload", aw.Name, "cpWorkloadID", cpID, "clusterID", r.clusterID)
	}
}

func (r *AegisWorkloadReconciler) ackWorkloadBridge(ctx context.Context, aw *aegisv1alpha1.AegisWorkload, status string, jobUID string) {
	if r.cpClient == nil || aw == nil {
		return
	}

	annotations := aw.GetAnnotations()
	lastPushedState := annotations[annotationLastPushedState]
	lastPushedJobUID := annotations[annotationLastPushedJobUID]

	// Only push if: state changed, OR job was recreated (different UID)
	stateChanged := lastPushedState != status
	jobRecreated := jobUID != "" && lastPushedJobUID != "" && lastPushedJobUID != jobUID

	if !stateChanged && !jobRecreated {
		// Already pushed this final state for this job
		return
	}

	backend := backendForWorkload(aw)
	url := aw.Status.URL
	if url == "" {
		url = fmt.Sprintf("k8s://%s/%s", aw.Namespace, aw.Name)
	}
	cpID := controlPlaneWorkloadID(aw.Name)
	if err := r.cpClient.Ack(ctx, cpID, status, backend, url); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "AckWorkload bridge failed", "status", status, "workload", aw.Name, "cpWorkloadID", cpID, "clusterID", r.clusterID)
		return
	}

	// Update annotations to track what we just pushed
	if err := r.updatePushedStateAnnotations(ctx, aw, status, jobUID); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed to persist state tracking", "workload", aw.Name, "status", status)
		return
	}

	if jobRecreated {
		ctrl.LoggerFrom(ctx).Info("AckWorkload bridge sent (job recreated)", "status", status, "clusterID", r.clusterID, "oldJobUID", lastPushedJobUID, "newJobUID", jobUID)
	} else {
		ctrl.LoggerFrom(ctx).Info("AckWorkload bridge sent", "status", status, "clusterID", r.clusterID)
	}
}

func (r *AegisWorkloadReconciler) markAnnotation(ctx context.Context, aw *aegisv1alpha1.AegisWorkload, key string) error {
	original := aw.DeepCopy()
	if aw.Annotations == nil {
		aw.Annotations = map[string]string{}
	}
	aw.Annotations[key] = "true"
	return r.Patch(ctx, aw, client.MergeFrom(original))
}

// updatePushedStateAnnotations updates annotations to track the last-pushed state and job UID.
// This enables proper re-sync when jobs are recreated - we compare against these values
// to determine if a status push is needed.
func (r *AegisWorkloadReconciler) updatePushedStateAnnotations(ctx context.Context, aw *aegisv1alpha1.AegisWorkload, state string, jobUID string) error {
	original := aw.DeepCopy()
	if aw.Annotations == nil {
		aw.Annotations = map[string]string{}
	}
	aw.Annotations[annotationLastPushedState] = state
	if jobUID != "" {
		aw.Annotations[annotationLastPushedJobUID] = jobUID
	}
	// Also set legacy annotations for backward compatibility during rollout
	if state == "RUNNING" {
		aw.Annotations[annotationStartAcked] = "true"
	} else if state == "SUCCEEDED" || state == "FAILED" {
		aw.Annotations[annotationFinalAcked] = "true"
	}
	return r.Patch(ctx, aw, client.MergeFrom(original))
}

func (r *AegisWorkloadReconciler) ensureInteractiveResources(ctx context.Context, aw *aegisv1alpha1.AegisWorkload) (bool, error) {
	ports := effectiveWorkspacePorts(aw.Spec.Workspace)
	changed := false

	if _, _, mutated, err := r.ensureWorkspaceSSHSecret(ctx, aw); err != nil {
		return false, err
	} else if mutated {
		changed = true
	}

	if svcChanged, err := r.ensureWorkspaceService(ctx, aw, ports); err != nil {
		return false, err
	} else if svcChanged {
		changed = true
	}

	if ingChanged, err := r.ensureWorkspaceIngress(ctx, aw); err != nil {
		return false, err
	} else if ingChanged {
		changed = true
	}

	return changed, nil
}

func (r *AegisWorkloadReconciler) cleanupInteractiveResources(ctx context.Context, aw *aegisv1alpha1.AegisWorkload) (bool, error) {
	changed := false

	if secretDeleted, err := r.cleanupWorkspaceSSHSecret(ctx, aw); err != nil {
		return false, err
	} else if secretDeleted {
		changed = true
	}

	svcName := workspaceAliasName(aw.Name)
	svc := &corev1.Service{}
	if err := r.Get(ctx, types.NamespacedName{Name: svcName, Namespace: aw.Namespace}, svc); err == nil {
		if metav1.IsControlledBy(svc, aw) {
			if err := r.Delete(ctx, svc); err != nil && !apierrors.IsNotFound(err) {
				return false, err
			}
			changed = true
		}
	} else if !apierrors.IsNotFound(err) {
		return false, err
	}

	ingName := workspaceAliasName(aw.Name)
	ing := &networkingv1.Ingress{}
	if err := r.Get(ctx, types.NamespacedName{Name: ingName, Namespace: aw.Namespace}, ing); err == nil {
		if metav1.IsControlledBy(ing, aw) {
			if err := r.Delete(ctx, ing); err != nil && !apierrors.IsNotFound(err) {
				return false, err
			}
			changed = true
		}
	} else if !apierrors.IsNotFound(err) {
		return false, err
	}

	return changed, nil
}

func workspaceSSHSecretName(workloadName string) string {
	return fmt.Sprintf("%s-ssh", workloadName)
}

func (r *AegisWorkloadReconciler) ensureWorkspaceSSHSecret(ctx context.Context, aw *aegisv1alpha1.AegisWorkload) (string, bool, bool, error) {
	secretName := workspaceSSHSecretName(aw.Name)
	annotations := aw.GetAnnotations()
	authorized := strings.TrimSpace(annotations[annotationSSHAuthorizedKeys])
	trusted := strings.TrimSpace(annotations[annotationSSHTrustedCA])
	if authorized == "" && trusted == "" {
		deleted, err := r.cleanupWorkspaceSSHSecret(ctx, aw)
		return "", false, deleted, err
	}

	secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: secretName, Namespace: aw.Namespace}}
	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, secret, func() error {
		if err := controllerutil.SetControllerReference(aw, secret, r.Scheme); err != nil {
			return err
		}
		if secret.Labels == nil {
			secret.Labels = map[string]string{}
		}
		secret.Labels[labelWorkloadID] = aw.Name
		secret.Labels[labelSSHManaged] = "true"
		secret.Type = corev1.SecretTypeOpaque
		if secret.Data == nil {
			secret.Data = map[string][]byte{}
		} else {
			for k := range secret.Data {
				delete(secret.Data, k)
			}
		}
		if authorized != "" {
			secret.Data["authorized_keys"] = []byte(authorized)
		}
		if trusted != "" {
			secret.Data["trusted_user_ca_keys"] = []byte(trusted)
		}
		return nil
	})
	if err != nil {
		return "", false, false, err
	}
	return secretName, true, result != controllerutil.OperationResultNone, nil
}

func (r *AegisWorkloadReconciler) cleanupWorkspaceSSHSecret(ctx context.Context, aw *aegisv1alpha1.AegisWorkload) (bool, error) {
	secretName := workspaceSSHSecretName(aw.Name)
	secret := &corev1.Secret{}
	if err := r.Get(ctx, types.NamespacedName{Name: secretName, Namespace: aw.Namespace}, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	if !metav1.IsControlledBy(secret, aw) {
		return false, nil
	}
	if err := r.Delete(ctx, secret); err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *AegisWorkloadReconciler) ensureWorkspaceService(ctx context.Context, aw *aegisv1alpha1.AegisWorkload, ports []int32) (bool, error) {
	svcName := workspaceAliasName(aw.Name)
	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: svcName, Namespace: aw.Namespace}}
	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		if err := controllerutil.SetControllerReference(aw, svc, r.Scheme); err != nil {
			return err
		}
		if svc.Labels == nil {
			svc.Labels = map[string]string{}
		}
		svc.Labels[labelWorkloadID] = aw.Name
		svc.Spec.Selector = map[string]string{labelWorkloadID: aw.Name}
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		servicePorts := make([]corev1.ServicePort, 0, len(ports))
		for _, p := range ports {
			if p <= 0 {
				continue
			}
			servicePorts = append(servicePorts, corev1.ServicePort{
				Name:       fmt.Sprintf("tcp-%d", p),
				Port:       p,
				TargetPort: intstr.FromInt(int(p)),
				Protocol:   corev1.ProtocolTCP,
			})
		}
		svc.Spec.Ports = servicePorts
		return nil
	})
	if err != nil {
		return false, err
	}
	return result != controllerutil.OperationResultNone, nil
}

func (r *AegisWorkloadReconciler) ensureWorkspaceIngress(ctx context.Context, aw *aegisv1alpha1.AegisWorkload) (bool, error) {
	ingName := workspaceAliasName(aw.Name)
	ing := &networkingv1.Ingress{ObjectMeta: metav1.ObjectMeta{Name: ingName, Namespace: aw.Namespace}}
	result, err := controllerutil.CreateOrUpdate(ctx, r.Client, ing, func() error {
		if err := controllerutil.SetControllerReference(aw, ing, r.Scheme); err != nil {
			return err
		}
		if ing.Labels == nil {
			ing.Labels = map[string]string{}
		}
		ing.Labels[labelWorkloadID] = aw.Name
		path := fmt.Sprintf("/proxy/%s", aw.Name)
		pathType := networkingv1.PathTypePrefix
		backendPort := networkingv1.ServiceBackendPort{Number: r.proxyServicePort}
		ing.Spec = networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{createIngressRule(r.proxyIngressHost, path, r.proxyServiceName, backendPort, pathType)},
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return result != controllerutil.OperationResultNone, nil
}

func effectiveWorkspacePorts(spec *aegisv1alpha1.WorkspaceSpec) []int32 {
	if spec == nil {
		return workspacecfg.EnsureDefaultPorts(nil)
	}
	return workspacecfg.EnsureDefaultPorts(spec.Ports)
}

func createIngressRule(host, path, serviceName string, backendPort networkingv1.ServiceBackendPort, pathType networkingv1.PathType) networkingv1.IngressRule {
	backend := networkingv1.IngressBackend{
		Service: &networkingv1.IngressServiceBackend{
			Name: serviceName,
			Port: backendPort,
		},
	}
	return networkingv1.IngressRule{
		Host: host,
		IngressRuleValue: networkingv1.IngressRuleValue{
			HTTP: &networkingv1.HTTPIngressRuleValue{
				Paths: []networkingv1.HTTPIngressPath{{
					Path:     path,
					PathType: &pathType,
					Backend:  backend,
				}},
			},
		},
	}
}

func backendForWorkload(aw *aegisv1alpha1.AegisWorkload) string {
	if aw.Status.Backend != "" {
		return aw.Status.Backend
	}
	if aw.Spec.Workspace != nil {
		return backendWorkspace
	}
	if aw.Spec.Training != nil {
		return backendPyTorch
	}
	return "operator"
}

func convertSpecHints(h *aegisv1alpha1.ResourceHints) *builders.GPUHints {
	if h == nil {
		return nil
	}

	return &builders.GPUHints{
		ResourceName:    h.ResourceName,
		GPUCount:        h.GpuCount,
		CpuCoresRequest: h.CpuCoresRequest,
		MemoryRequest:   h.MemoryRequest,
	}
}
func maxDurationFromAnnotation(aw *aegisv1alpha1.AegisWorkload) (*int64, error) {
	if aw == nil {
		return nil, nil
	}
	val := aw.GetAnnotations()[annotationMaxDurationSeconds]
	if val == "" {
		return nil, nil
	}
	secs, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", annotationMaxDurationSeconds, err)
	}
	if secs <= 0 {
		return nil, nil
	}
	return ptr.To(secs), nil
}

func ttlSecondsAfterFinishedFromAnnotation(aw *aegisv1alpha1.AegisWorkload) (*int32, error) {
	if aw == nil {
		return nil, nil
	}
	val := aw.GetAnnotations()[annotationTTLSecondsAfterFinished]
	if val == "" {
		return nil, nil
	}
	secs, err := strconv.ParseInt(val, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse %s: %w", annotationTTLSecondsAfterFinished, err)
	}
	if secs < 0 {
		return nil, nil
	}
	return ptr.To(int32(secs)), nil
}

func workspaceDefaultCommand(image string) string {
	return fmt.Sprintf(`echo "[AEGIS] start workload=$AEGIS_WORKLOAD_ID pod=$POD_NAME ns=$POD_NAMESPACE node=$NODE_NAME";
date;
echo "[AEGIS] image=%s";
echo "[AEGIS] launching VS Code workspace...";
exec /init`, image)
}

func deriveJobFailureMessage(job *batchv1.Job) string {
	if job == nil {
		return ""
	}
	for _, cond := range job.Status.Conditions {
		if cond.Type == batchv1.JobFailed {
			if cond.Message != "" {
				return cond.Message
			}
			if cond.Reason != "" {
				return cond.Reason
			}
		}
	}
	return "job failed"
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// SetupWithManager sets up the controller with the Manager.
func (r *AegisWorkloadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	dyn, err := dynamic.NewForConfig(mgr.GetConfig())
	if err != nil {
		return err
	}

	var cpCli *cpclient.Client
	clusterID := os.Getenv("AEGIS_CLUSTER_ID")
	if endpoint := os.Getenv("AEGIS_CP_GRPC"); endpoint != "" {
		client, err := cpclient.New(endpoint)
		if err != nil {
			ctrl.Log.WithName("operator").Error(err, "failed to create control-plane client", "endpoint", endpoint)
		} else {
			cpCli = client
		}
	}

	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()
	r.Recorder = mgr.GetEventRecorderFor("aegisworkload-controller")
	r.Dynamic = dyn
	r.Config = mgr.GetConfig()
	r.cpClient = cpCli
	r.clusterID = clusterID
	r.defaultWorkspaceImage = envOrDefault("AEGIS_DEFAULT_IMAGE", workspacecfg.DefaultWorkspaceImage)
	r.defaultTrainingImage = envOrDefault("AEGIS_TRAINING_DEFAULT_IMAGE", "pytorch/pytorch:2.4.0-cuda11.8-cudnn8-runtime")
	r.gpuResourceOverride = envOrDefault("AEGIS_GPU_RESOURCE_NAME", "")
	r.dryRun = os.Getenv("AEGIS_DRY_RUN") == "1"
	r.kueueEnabled = os.Getenv("AEGIS_KUEUE_ENABLED") == "1"
	r.kueueQueue = envOrDefault("AEGIS_KUEUE_QUEUE", "")
	r.proxyServiceName = envOrDefault("AEGIS_PROXY_SERVICE_NAME", "aegis-auth-proxy")
	defaults := workspacecfg.DefaultEnv()
	defaults[workspacecfg.EnvVSCodeCommit] = envOrDefault("AEGIS_VSCODE_COMMIT", defaults[workspacecfg.EnvVSCodeCommit])
	defaults[workspacecfg.EnvVSCodeQuality] = envOrDefault("AEGIS_VSCODE_QUALITY", defaults[workspacecfg.EnvVSCodeQuality])
	defaults[workspacecfg.EnvPUID] = envOrDefault("AEGIS_WORKSPACE_PUID", defaults[workspacecfg.EnvPUID])
	defaults[workspacecfg.EnvPGID] = envOrDefault("AEGIS_WORKSPACE_PGID", defaults[workspacecfg.EnvPGID])
	defaults[workspacecfg.EnvPasswordAccess] = envOrDefault("AEGIS_WORKSPACE_PASSWORD_ACCESS", defaults[workspacecfg.EnvPasswordAccess])
	defaults[workspacecfg.EnvUserName] = envOrDefault("AEGIS_WORKSPACE_USER_NAME", defaults[workspacecfg.EnvUserName])
	defaults[workspacecfg.EnvUserPassword] = envOrDefault("AEGIS_WORKSPACE_USER_PASSWORD", defaults[workspacecfg.EnvUserPassword])
	r.workspaceEnvDefaults = defaults
	if portStr := envOrDefault("AEGIS_PROXY_SERVICE_PORT", "8080"); portStr != "" {
		if val, err := strconv.Atoi(portStr); err == nil {
			r.proxyServicePort = int32(val)
		} else {
			r.proxyServicePort = 8080
		}
	} else {
		r.proxyServicePort = 8080
	}
	r.proxyIngressHost = envOrDefault("AEGIS_PROXY_INGRESS_HOST", "")
	r.proxyURL = envOrDefault("AEGIS_PROXY_URL", "")
	r.sshBootstrapImage = envOrDefault("AEGIS_SSH_BOOTSTRAP_IMAGE", "")
	if r.kueueEnabled {
		if ok, err := workdiscovery.DetectKueue(r.Config); err != nil {
			ctrl.Log.WithName("operator").Error(err, "kueue detection failed")
			r.kueueEnabled = false
		} else if !ok {
			ctrl.Log.WithName("operator").Info("kueue CRDs not found; disabling integration")
			r.kueueEnabled = false
		}
	}
	if os.Getenv("AEGIS_DISABLE_KUEUE") == "1" {
		r.kueueEnabled = false
	}

	if r.cpClient != nil && r.clusterID != "" {
		if err := mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
			r.startClusterPresence(ctx)
			return nil
		})); err != nil {
			return err
		}
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&aegisv1alpha1.AegisWorkload{}).
		Owns(&batchv1.Job{}).
		Named("aegisworkload").
		Complete(r)
}
