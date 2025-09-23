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
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
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
)

const (
	backendWorkspace = "workspace"
	backendPyTorch   = "pytorch_v1"

	requeuePending = 10 * time.Second

	annotationStartAcked = "aegis.yourorg.dev/start-acked"
	annotationFinalAcked = "aegis.yourorg.dev/final-acked"
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
	jobName := builders.SanitizeName("aegis-" + aw.Name)
	jobKey := types.NamespacedName{Name: jobName, Namespace: aw.Namespace}

	var job batchv1.Job
	err := r.Get(ctx, jobKey, &job)
	if apierrors.IsNotFound(err) {
		opts := builders.WorkspaceOptions{
			Namespace:           aw.Namespace,
			JobName:             jobName,
			WorkloadID:          aw.Name,
			Image:               image,
			Command:             spec.Command,
			DefaultCommand:      workspaceDefaultCommand(image),
			Env:                 spec.Env,
			Flavor:              flavor,
			Queue:               aw.Spec.Queue,
			Hints:               convertSpecHints(aw.Spec.Hints),
			DryRun:              r.dryRun,
			KueueEnabled:        r.kueueEnabled,
			KueueQueue:          r.kueueQueue,
			GPUResourceOverride: r.gpuResourceOverride,
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

	if !r.kueueEnabled && job.Spec.Suspend != nil && *job.Spec.Suspend {
		switch err := r.unsuspendJob(ctx, jobKey); {
		case err == nil:
			log.Info("unsuspended job", "job", jobName)
		case apierrors.IsNotFound(err):
			return ctrl.Result{}, nil
		default:
			return ctrl.Result{}, err
		}
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

	phase := aw.Status.Phase
	observedRunning := job.Status.Active > 0 || job.Status.Succeeded > 0 || job.Status.Failed > 0

	if observedRunning && phase != aegisv1alpha1.PhaseRunning {
		prev := aw.Status.Phase
		_ = r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.Phase = aegisv1alpha1.PhaseRunning
			st.Backend = backend
			if st.StartTime == nil {
				if job.Status.StartTime != nil {
					st.StartTime = job.Status.StartTime.DeepCopy()
				} else {
					now := metav1.NewTime(time.Now())
					st.StartTime = &now
				}
			}
		})
		if prev != aegisv1alpha1.PhaseRunning {
			r.startWorkloadBridge(ctx, aw)
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
		if prev != aegisv1alpha1.PhaseSucceeded {
			r.ackWorkloadBridge(ctx, aw, "SUCCEEDED")
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
		if prev != aegisv1alpha1.PhaseFailed {
			r.ackWorkloadBridge(ctx, aw, "FAILED")
		}
	}
}

func (r *AegisWorkloadReconciler) applyTrainingTransitions(ctx context.Context, aw *aegisv1alpha1.AegisWorkload, obj *unstructured.Unstructured) {
	if obj == nil {
		return
	}

	if workstatus.HasCondition(obj, "Running") && aw.Status.Phase != aegisv1alpha1.PhaseRunning {
		prev := aw.Status.Phase
		_ = r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.Phase = aegisv1alpha1.PhaseRunning
			st.Backend = backendPyTorch
			if st.StartTime == nil {
				now := metav1.NewTime(time.Now())
				st.StartTime = &now
			}
		})
		if prev != aegisv1alpha1.PhaseRunning {
			r.startWorkloadBridge(ctx, aw)
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
		if prev != aegisv1alpha1.PhaseSucceeded {
			r.ackWorkloadBridge(ctx, aw, "SUCCEEDED")
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
		if prev != aegisv1alpha1.PhaseFailed {
			r.ackWorkloadBridge(ctx, aw, "FAILED")
		}
	}
}

func (r *AegisWorkloadReconciler) unsuspendJob(ctx context.Context, key types.NamespacedName) error {
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		var latest batchv1.Job
		if err := r.Get(ctx, key, &latest); err != nil {
			return err
		}
		if latest.Spec.Suspend != nil && !*latest.Spec.Suspend {
			return nil
		}
		latest.Spec.Suspend = ptr.To(false)
		return r.Update(ctx, &latest)
	})
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

	registerReq := &aegisproto.ClusterRegisterRequest{
		ClusterId: r.clusterID,
		Provider:  os.Getenv("AEGIS_PROVIDER"),
		Region:    os.Getenv("AEGIS_REGION"),
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
		zapLogger.Info("cluster registered", zap.String("cluster_id", r.clusterID))
		break
	}

	flavors := r.discoverFlavors()
	r.cpClient.HeartbeatLoop(ctx, zapLogger, r.clusterID, flavors)
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
	gpusPer := int(spec.GpusPerWorker)
	if gpusPer <= 0 {
		gpusPer = 1
	}

	if err := r.ensureTrainingDiscovery(); err != nil {
		log.Error(err, "training CRDs unavailable")
		return ctrl.Result{}, err
	}

	name := builders.SanitizeName("aegis-" + aw.Name)
	resource := r.Dynamic.Resource(*r.pyTorchGVR).Namespace(aw.Namespace)

	pyJob, err := resource.Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		obj := builders.BuildPyTorchJob(builders.TrainingOptions{
			APIVersion:    r.pyTorchAPIVersion,
			Namespace:     aw.Namespace,
			Name:          name,
			Image:         image,
			Command:       spec.Command,
			Flavor:        spec.Flavor,
			Workers:       workers,
			GPUsPerWorker: gpusPer,
			DryRun:        r.dryRun,
			Hints:         convertSpecHints(aw.Spec.Hints),
		})
		if err := controllerutil.SetControllerReference(aw, obj, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}
		if _, err := resource.Create(ctx, obj, metav1.CreateOptions{}); err != nil {
			return ctrl.Result{}, err
		}

		log.Info("pytorchjob created", "name", name)
		r.Recorder.Eventf(aw, corev1.EventTypeNormal, "Submitted", "PyTorchJob %s created", name)

		if err := r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.Phase = aegisv1alpha1.PhaseSubmitted
			st.Backend = backendPyTorch
			st.JobRef = &aegisv1alpha1.JobRef{APIVersion: r.pyTorchAPIVersion, Kind: "PyTorchJob", Name: name, Namespace: aw.Namespace}
			st.URL = fmt.Sprintf("k8s://%s/pytorchjob/%s", aw.Namespace, name)
		}); err != nil {
			return ctrl.Result{}, err
		}

		return ctrl.Result{RequeueAfter: requeuePending}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	if aw.Status.JobRef == nil {
		if err := r.patchStatus(ctx, aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.JobRef = &aegisv1alpha1.JobRef{APIVersion: r.pyTorchAPIVersion, Kind: "PyTorchJob", Name: name, Namespace: aw.Namespace}
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	r.applyTrainingTransitions(ctx, aw, pyJob)

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

func (r *AegisWorkloadReconciler) startWorkloadBridge(ctx context.Context, aw *aegisv1alpha1.AegisWorkload) {
	if r.cpClient == nil || aw == nil {
		return
	}
	if r.clusterID == "" {
		ctrl.LoggerFrom(ctx).Info("cluster id missing; skipping StartWorkload bridge", "workload", aw.Name)
		return
	}
	if aw.GetAnnotations()[annotationStartAcked] == "true" {
		return
	}
	if err := r.cpClient.Start(ctx, aw.Name, r.clusterID); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "StartWorkload bridge failed", "workload", aw.Name, "clusterID", r.clusterID)
		return
	}
	if err := r.markAnnotation(ctx, aw, annotationStartAcked); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed to persist start acknowledgement", "workload", aw.Name)
		return
	}
	ctrl.LoggerFrom(ctx).Info("StartWorkload bridge sent", "workload", aw.Name, "clusterID", r.clusterID)
}

func (r *AegisWorkloadReconciler) ackWorkloadBridge(ctx context.Context, aw *aegisv1alpha1.AegisWorkload, status string) {
	if r.cpClient == nil || aw == nil {
		return
	}
	if aw.GetAnnotations()[annotationFinalAcked] == "true" {
		return
	}
	backend := backendForWorkload(aw)
	url := aw.Status.URL
	if url == "" {
		url = fmt.Sprintf("k8s://%s/%s", aw.Namespace, aw.Name)
	}
	if err := r.cpClient.Ack(ctx, aw.Name, status, backend, url); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "AckWorkload bridge failed", "status", status, "clusterID", r.clusterID)
		return
	}
	if err := r.markAnnotation(ctx, aw, annotationFinalAcked); err != nil {
		ctrl.LoggerFrom(ctx).Error(err, "failed to persist final acknowledgement", "workload", aw.Name, "status", status)
		return
	}
	ctrl.LoggerFrom(ctx).Info("AckWorkload bridge sent", "status", status, "clusterID", r.clusterID)
}

func (r *AegisWorkloadReconciler) markAnnotation(ctx context.Context, aw *aegisv1alpha1.AegisWorkload, key string) error {
	original := aw.DeepCopy()
	if aw.Annotations == nil {
		aw.Annotations = map[string]string{}
	}
	aw.Annotations[key] = "true"
	return r.Patch(ctx, aw, client.MergeFrom(original))
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
	return &builders.GPUHints{ResourceName: h.ResourceName, GPUCount: h.GpuCount}
}

func workspaceDefaultCommand(image string) string {
	return fmt.Sprintf(`echo "[AEGIS] start workload=$AEGIS_WORKLOAD_ID pod=$POD_NAME ns=$POD_NAMESPACE node=$NODE_NAME";
	date;
	echo "[AEGIS] image=%s";
	echo "[AEGIS] running task...";
	sleep 3;
	echo "[AEGIS] done";`, image)
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
	r.defaultWorkspaceImage = envOrDefault("AEGIS_DEFAULT_IMAGE", "alpine:3.19")
	r.defaultTrainingImage = envOrDefault("AEGIS_TRAINING_DEFAULT_IMAGE", "pytorch/pytorch:2.4.0-cuda11.8-cudnn8-runtime")
	r.gpuResourceOverride = envOrDefault("AEGIS_GPU_RESOURCE_NAME", "")
	r.dryRun = os.Getenv("AEGIS_DRY_RUN") == "1"
	r.kueueEnabled = os.Getenv("AEGIS_KUEUE_ENABLED") == "1"
	r.kueueQueue = envOrDefault("AEGIS_KUEUE_QUEUE", "")
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
