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

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	aegisv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
	aegisv1alpha2 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha2"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/cpclient"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/providers"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/workspace/metrics"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/workspace/plan"
	aegisproto "github.com/yourorg/aegis/proto/aegis/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	workspaceControllerName = "workspace"
	requeueWorkspace        = 10 * time.Second
)

// WorkspaceReconciler orchestrates Workspace resources and their realizations.
type WorkspaceReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	planner          *plan.Planner
	recorder         recordEventRecorder
	metricsCollector *metrics.Collector
	providers        *providers.Registry
	cpClient         *cpclient.Client
	clusterID        string
}

type recordEventRecorder interface {
	Event(object runtime.Object, eventtype, reason, message string)
	Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{})
}

// +kubebuilder:rbac:groups=aegis.yourorg.dev,resources=workspaces,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aegis.yourorg.dev,resources=workspaces/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=aegis.yourorg.dev,resources=workspaces/finalizers,verbs=update
// +kubebuilder:rbac:groups=aegis.yourorg.dev,resources=aegisworkloads,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=aegis.yourorg.dev,resources=aegisworkloads/status,verbs=get
// +kubebuilder:rbac:groups=aegis.yourorg.dev,resources=workspaceclusters;workspacenetworks;workspacestorage;workspacegpuprofiles;aegisworkloadslices,verbs=list;watch
// +kubebuilder:rbac:groups="",resources=events,verbs=create;patch

// Reconcile handles Workspace state transitions.
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithValues("workspace", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	if r.planner == nil {
		r.planner = plan.NewPlanner()
	}

	var ws aegisv1alpha2.Workspace
	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !ws.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	planResult, err := r.planner.Plan(&ws)
	if err != nil {
		log.Error(err, "failed to compute workspace plan")
		if patchErr := r.patchStatus(ctx, &ws, func(status *aegisv1alpha2.WorkspaceStatus) {
			status.Conditions = mergeCondition(status.Conditions, metav1.Condition{
				Type:               "Planned",
				Status:             metav1.ConditionFalse,
				Reason:             "PlanFailed",
				Message:            err.Error(),
				ObservedGeneration: ws.GetGeneration(),
			})
			status.Phase = aegisv1alpha2.WorkspacePhasePending
		}); patchErr != nil {
			return ctrl.Result{}, patchErr
		}
		return ctrl.Result{}, nil
	}

	extraConditions := []metav1.Condition{}
	if r.providers != nil && ws.Spec.BackendRef != nil {
		if plugin := r.providers.Resolve(ws.Spec.BackendRef); plugin != nil {
			if err := plugin.Reconcile(ctx, &ws); err != nil {
				cond := metav1.Condition{
					Type:               "ClusterReady",
					Status:             metav1.ConditionFalse,
					Reason:             "ProviderError",
					Message:            err.Error(),
					ObservedGeneration: ws.GetGeneration(),
				}
				extraConditions = append(extraConditions, cond)
				if patchErr := r.patchStatus(ctx, &ws, func(status *aegisv1alpha2.WorkspaceStatus) {
					status.Conditions = mergeCondition(status.Conditions, cond)
				}); patchErr != nil {
					return ctrl.Result{}, patchErr
				}
				return ctrl.Result{RequeueAfter: requeueWorkspace}, nil
			}
			extraConditions = append(extraConditions, metav1.Condition{
				Type:               "ClusterReady",
				Status:             metav1.ConditionTrue,
				Reason:             "ProviderReady",
				ObservedGeneration: ws.GetGeneration(),
			})
		} else {
			extraConditions = append(extraConditions, metav1.Condition{
				Type:               "ClusterReady",
				Status:             metav1.ConditionUnknown,
				Reason:             "ProviderUnavailable",
				Message:            "no provider registered for backendRef",
				ObservedGeneration: ws.GetGeneration(),
			})
		}
	}

	workload, err := r.ensureWorkload(ctx, &ws, planResult)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.rollupStatus(ctx, &ws, workload, extraConditions); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: requeueWorkspace}, nil
}

func (r *WorkspaceReconciler) ensureWorkload(ctx context.Context, ws *aegisv1alpha2.Workspace, planResult *plan.Result) (*aegisv1alpha1.AegisWorkload, error) {
	child := &aegisv1alpha1.AegisWorkload{ObjectMeta: metav1.ObjectMeta{Name: planResult.WorkloadName, Namespace: ws.Namespace}}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, child, func() error {
		if err := controllerutil.SetControllerReference(ws, child, r.Scheme); err != nil {
			return err
		}

		child.Spec = planResult.WorkloadSpec

		if child.Labels == nil {
			child.Labels = map[string]string{}
		}
		for k, v := range planResult.Labels {
			child.Labels[k] = v
		}

		if len(planResult.Annotations) > 0 {
			if child.Annotations == nil {
				child.Annotations = map[string]string{}
			}
			for k, v := range planResult.Annotations {
				child.Annotations[k] = v
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if r.recorder != nil && op == controllerutil.OperationResultCreated {
		r.recorder.Eventf(ws, corev1.EventTypeNormal, "Submitted", "Workspace workload %s created", planResult.WorkloadName)
	}

	if err := r.Get(ctx, types.NamespacedName{Name: child.Name, Namespace: child.Namespace}, child); err != nil {
		if apierrors.IsNotFound(err) {
			return nil, fmt.Errorf("workload %s/%s disappeared after creation", child.Namespace, child.Name)
		}
		return nil, err
	}

	if err := r.ensureControlPlaneRegistration(ctx, ws, child); err != nil {
		return nil, err
	}

	return child, nil
}

func (r *WorkspaceReconciler) rollupStatus(ctx context.Context, ws *aegisv1alpha2.Workspace, child *aegisv1alpha1.AegisWorkload, extra []metav1.Condition) error {
	var reconciledPhase aegisv1alpha2.WorkspacePhase
	err := r.patchStatus(ctx, ws, func(status *aegisv1alpha2.WorkspaceStatus) {
		if child != nil {
			phase := aegisv1alpha2.WorkspacePhase(child.Status.Phase)
			if phase == "" {
				phase = aegisv1alpha2.WorkspacePhaseSubmitted
			}
			status.Phase = phase
			status.Backend = child.Status.Backend
			status.URL = child.Status.URL
			status.WorkloadRef = &corev1.ObjectReference{
				APIVersion: aegisv1alpha1.GroupVersion.String(),
				Kind:       "AegisWorkload",
				Name:       child.Name,
				Namespace:  child.Namespace,
			}
			status.Conditions = mergeConditionsForWorkspace(ws, child, extra)
		} else {
			status.Phase = aegisv1alpha2.WorkspacePhasePending
			status.Backend = ""
			status.URL = ""
			status.WorkloadRef = nil
			status.Conditions = mergeConditionsForWorkspace(ws, nil, extra)
		}
		reconciledPhase = status.Phase
	})
	if err != nil {
		return err
	}

	if r.metricsCollector != nil {
		phaseValue := string(reconciledPhase)
		if phaseValue == "" {
			phaseValue = string(aegisv1alpha2.WorkspacePhasePending)
		}
		r.metricsCollector.ObservePhase(ws.Spec.ProjectRef, ws.Spec.Queue, phaseValue)
	}

	return nil
}

func (r *WorkspaceReconciler) patchStatus(ctx context.Context, ws *aegisv1alpha2.Workspace, mutate func(*aegisv1alpha2.WorkspaceStatus)) error {
	original := ws.DeepCopy()
	mutate(&ws.Status)
	return r.Status().Patch(ctx, ws, client.MergeFrom(original))
}

func (r *WorkspaceReconciler) ensureControlPlaneRegistration(ctx context.Context, ws *aegisv1alpha2.Workspace, aw *aegisv1alpha1.AegisWorkload) error {
	if r.cpClient == nil || ws == nil || aw == nil {
		return nil
	}

	workloadID := strings.TrimSpace(ws.GetLabels()["aegis.workload/id"])
	for strings.HasPrefix(workloadID, "aegis-") {
		workloadID = strings.TrimPrefix(workloadID, "aegis-")
	}
	if workloadID == "" {
		workloadID = strings.TrimSpace(ws.GetName())
		for strings.HasPrefix(workloadID, "aegis-") {
			workloadID = strings.TrimPrefix(workloadID, "aegis-")
		}
	}
	if workloadID == "" {
		workloadID = strings.TrimSpace(aw.GetLabels()["aegis.workload/id"])
		for strings.HasPrefix(workloadID, "aegis-") {
			workloadID = strings.TrimPrefix(workloadID, "aegis-")
		}
	}
	if workloadID == "" {
		workloadID = strings.TrimSpace(aw.GetName())
		for strings.HasPrefix(workloadID, "aegis-") {
			workloadID = strings.TrimPrefix(workloadID, "aegis-")
		}
	}
	if workloadID == "" {
		ctrl.LoggerFrom(ctx).V(1).Info("workspace missing workload identifier; skipping control-plane registration",
			"workspace", ws.GetName(),
			"namespace", ws.GetNamespace())
		return nil
	}

	if _, err := r.cpClient.GetWorkload(ctx, workloadID); err == nil {
		return nil
	} else if status.Code(err) != codes.NotFound {
		return err
	}

	protoWorkload, err := buildControlPlaneWorkload(workloadID, &aw.Spec)
	if err != nil {
		return err
	}
	if r.clusterID != "" {
		protoWorkload.ClusterId = r.clusterID
	}

	ctrl.LoggerFrom(ctx).Info("registering workspace with control plane", "workloadID", workloadID)

	if _, err := r.cpClient.SubmitWorkload(ctx, protoWorkload); err != nil {
		return err
	}
	return nil
}

func buildControlPlaneWorkload(id string, spec *aegisv1alpha1.AegisWorkloadSpec) (*aegisproto.Workload, error) {
	if spec == nil {
		return nil, fmt.Errorf("workload %s has no specification", id)
	}
	if spec.ProjectID == "" {
		return nil, fmt.Errorf("workload %s missing projectId", id)
	}

	workload := &aegisproto.Workload{
		Id:        id,
		ProjectId: spec.ProjectID,
		Queue:     spec.Queue,
	}

	switch {
	case spec.Workspace != nil:
		wsSpec := spec.Workspace
		workload.Kind = &aegisproto.Workload_Workspace{
			Workspace: &aegisproto.WorkspaceSpec{
				Flavor:      wsSpec.Flavor,
				Image:       wsSpec.Image,
				Env:         copyStringMap(wsSpec.Env),
				Command:     copyStringSlice(wsSpec.Command),
				Interactive: wsSpec.Interactive,
				Ports:       copyInt32Slice(wsSpec.Ports),
			},
		}
	case spec.Training != nil:
		trSpec := spec.Training
		workload.Kind = &aegisproto.Workload_Training{
			Training: &aegisproto.TrainingSpec{
				Flavor:        trSpec.Flavor,
				Workers:       trSpec.Workers,
				GpusPerWorker: trSpec.GpusPerWorker,
				Image:         trSpec.Image,
				Command:       copyStringSlice(trSpec.Command),
				Gang:          trSpec.Gang,
			},
		}
	default:
		return nil, fmt.Errorf("workload %s missing workspace or training specification", id)
	}

	if spec.Hints != nil {
		hints := &aegisproto.ResourceHints{
			ResourceName: spec.Hints.ResourceName,
			GpuCount:     spec.Hints.GpuCount,
		}
		if spec.Hints.CpuCoresRequest != nil {
			hints.CpuCoresRequest = *spec.Hints.CpuCoresRequest
		}
		if spec.Hints.MemoryRequest != nil {
			hints.MemoryRequest = *spec.Hints.MemoryRequest
		}
		workload.Hints = hints
	}

	return workload, nil
}

func copyStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func copyStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func copyInt32Slice(in []int32) []int32 {
	if len(in) == 0 {
		return nil
	}
	out := make([]int32, len(in))
	copy(out, in)
	return out
}

func mergeConditionsForWorkspace(ws *aegisv1alpha2.Workspace, child *aegisv1alpha1.AegisWorkload, extra []metav1.Condition) []metav1.Condition {
	conditions := []metav1.Condition{}

	planned := metav1.Condition{
		Type:               "Planned",
		Status:             metav1.ConditionTrue,
		Reason:             "PlanReady",
		ObservedGeneration: ws.GetGeneration(),
	}
	planned.LastTransitionTime = metav1.Now()
	conditions = append(conditions, planned)

	for _, cond := range extra {
		if cond.ObservedGeneration == 0 {
			cond.ObservedGeneration = ws.GetGeneration()
		}
		conditions = mergeCondition(conditions, cond)
	}

	if child == nil {
		return conditions
	}

	for _, cond := range child.Status.Conditions {
		conditions = append(conditions, cond)
	}

	phaseCondition := metav1.Condition{
		Type:               "Running",
		Status:             metav1.ConditionFalse,
		Reason:             "NotRunning",
		ObservedGeneration: ws.GetGeneration(),
	}
	if child.Status.Phase == aegisv1alpha1.PhaseRunning {
		phaseCondition.Status = metav1.ConditionTrue
		phaseCondition.Reason = "Running"
		phaseCondition.Message = "AegisWorkload running"
	}
	conditions = mergeCondition(conditions, phaseCondition)

	return conditions
}

func mergeCondition(existing []metav1.Condition, update metav1.Condition) []metav1.Condition {
	if update.LastTransitionTime.IsZero() {
		update.LastTransitionTime = metav1.Now()
	}
	found := false
	for i, cond := range existing {
		if cond.Type == update.Type {
			existing[i] = update
			found = true
			break
		}
	}
	if !found {
		existing = append(existing, update)
	}
	return existing
}

// SetupWithManager wires the controller into the manager.
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if r.planner == nil {
		r.planner = plan.NewPlanner()
	}
	if r.recorder == nil {
		r.recorder = mgr.GetEventRecorderFor("workspace-controller")
	}
	if r.Client == nil {
		r.Client = mgr.GetClient()
	}
	if r.Scheme == nil {
		r.Scheme = mgr.GetScheme()
	}
	if r.metricsCollector == nil {
		r.metricsCollector = metrics.NewCollector()
	}
	if r.providers == nil {
		r.providers = providers.NewRegistry()
	}
	if r.clusterID == "" {
		r.clusterID = os.Getenv("AEGIS_CLUSTER_ID")
	}
	if r.cpClient == nil {
		if endpoint := os.Getenv("AEGIS_CP_GRPC"); endpoint != "" {
			client, err := cpclient.New(endpoint)
			if err != nil {
				ctrl.Log.WithName("workspace-controller").Error(err, "failed to create control-plane client", "endpoint", endpoint)
			} else {
				r.cpClient = client
			}
		}
	}

	return ctrl.NewControllerManagedBy(mgr).
		Named(workspaceControllerName).
		For(&aegisv1alpha2.Workspace{}).
		Owns(&aegisv1alpha1.AegisWorkload{}).
		Complete(r)
}

// Ensure interface compliance.
var _ reconcile.Reconciler = (*WorkspaceReconciler)(nil)
