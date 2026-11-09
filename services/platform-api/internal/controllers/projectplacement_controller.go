package controllers

import (
	"context"
	"strings"
	"time"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/placement"
)

const projectPlacementFinalizer = "infra.aegis.yourorg.dev/projectplacement-finalizer"

// ProjectPlacementReconciler validates placement policy CRDs and publishes them to the in-memory overlay.
type ProjectPlacementReconciler struct {
	client.Client
	Log           *zap.Logger
	PolicyOverlay *placement.PolicyOverlay
}

// SetupWithManager registers the reconciler with the manager.
func (r *ProjectPlacementReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infraapi.ProjectPlacement{}).
		Complete(r)
}

// Reconcile pushes effective placement policy into the overlay and updates status.
func (r *ProjectPlacementReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.logger().With(zap.String("projectplacement", req.NamespacedName.String()))

	var resource infraapi.ProjectPlacement
	if err := r.Get(ctx, req.NamespacedName, &resource); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	overlay := r.overlay()
	projectID := strings.TrimSpace(resource.Spec.ProjectID)

	if !resource.DeletionTimestamp.IsZero() {
		overlay.Delete(projectID)
		if controllerutil.ContainsFinalizer(&resource, projectPlacementFinalizer) {
			patched := resource.DeepCopy()
			controllerutil.RemoveFinalizer(patched, projectPlacementFinalizer)
			if err := r.Client.Patch(ctx, patched, client.MergeFrom(&resource)); err != nil {
				return ctrl.Result{}, err
			}
		}
		return ctrl.Result{}, nil
	}

	if !controllerutil.ContainsFinalizer(&resource, projectPlacementFinalizer) {
		patched := resource.DeepCopy()
		controllerutil.AddFinalizer(patched, projectPlacementFinalizer)
		if err := r.Client.Patch(ctx, patched, client.MergeFrom(&resource)); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	cond := metav1.Condition{
		Type:               "Ready",
		Status:             metav1.ConditionTrue,
		Reason:             "PolicyValid",
		LastTransitionTime: metav1.NewTime(time.Now()),
		Message:            "Placement policy validated",
	}

	if projectID == "" {
		cond.Status = metav1.ConditionFalse
		cond.Reason = "MissingProjectID"
		cond.Message = "spec.projectId is required"
		if err := r.updateStatus(ctx, &resource, infraapi.ProjectPlacementSpec{}, cond); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// Persist effective policy into the overlay.
	overlay.Set(&resource.Spec)

	if err := r.updateStatus(ctx, &resource, resource.Spec, cond); err != nil {
		return ctrl.Result{}, err
	}

	log.Debug("project placement reconciled", zap.String("project_id", projectID))
	return ctrl.Result{}, nil
}

func (r *ProjectPlacementReconciler) updateStatus(ctx context.Context, resource *infraapi.ProjectPlacement, effective infraapi.ProjectPlacementSpec, cond metav1.Condition) error {
	patched := resource.DeepCopy()
	patched.Status.Effective = effective
	patched.Status.Conditions = append([]metav1.Condition(nil), resource.Status.Conditions...)
	updateCondition(&patched.Status.Conditions, cond)
	return r.Client.Status().Patch(ctx, patched, client.MergeFrom(resource))
}

func updateCondition(conds *[]metav1.Condition, cond metav1.Condition) {
	if conds == nil {
		return
	}
	replaced := false
	for i, existing := range *conds {
		if existing.Type == cond.Type {
			if existing.Status == cond.Status {
				cond.LastTransitionTime = existing.LastTransitionTime
			}
			(*conds)[i] = cond
			replaced = true
			break
		}
	}
	if !replaced {
		*conds = append(*conds, cond)
	}
}

func (r *ProjectPlacementReconciler) overlay() *placement.PolicyOverlay {
	if r.PolicyOverlay == nil {
		r.PolicyOverlay = placement.NewPolicyOverlay()
	}
	return r.PolicyOverlay
}

func (r *ProjectPlacementReconciler) logger() *zap.Logger {
	if r.Log != nil {
		return r.Log
	}
	return zap.NewNop()
}
