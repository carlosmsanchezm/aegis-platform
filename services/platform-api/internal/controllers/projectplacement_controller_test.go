package controllers

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	placementlib "github.com/yourorg/aegis/services/platform-api/internal/placement"
)

func TestProjectPlacementReconcileCreatesPolicy(t *testing.T) {
	t.Parallel()

	scheme, err := newTestScheme()
	if err != nil {
		t.Fatalf("scheme: %v", err)
	}

	placement := &infraapi.ProjectPlacement{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "proj-pp",
			Namespace: "aegis-system",
		},
		Spec: infraapi.ProjectPlacementSpec{
			ProjectID:        "proj-pp",
			AllowedRegions:   []string{"US-East"},
			AllowedProviders: []string{"AWS"},
			DefaultFlavor:    "a10",
			Quotas: map[string]int{
				"A10": 3,
			},
			Strategy: "Spread",
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&infraapi.ProjectPlacement{}).
		WithObjects(placement).
		Build()

	overlay := placementlib.NewPolicyOverlay()
	reconciler := &ProjectPlacementReconciler{
		Client:        cl,
		Log:           zaptest.NewLogger(t).Named("projectplacement"),
		PolicyOverlay: overlay,
	}

	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: placement.Name, Namespace: placement.Namespace}}

	// First reconcile adds finalizer, second populates overlay/status.
	if _, err := reconciler.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("reconcile add finalizer: %v", err)
	}
	if _, err := reconciler.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("reconcile populate policy: %v", err)
	}

	policy, ok := overlay.Effective("proj-pp")
	if !ok {
		t.Fatalf("expected policy overlay entry")
	}
	if policy.DefaultFlavor != "a10" || policy.Strategy != "Spread" {
		t.Fatalf("unexpected policy defaults: %+v", policy)
	}
	if got := policy.Quotas["a10"]; got != 3 {
		t.Fatalf("unexpected quota: %d", got)
	}
	if len(policy.AllowedRegions) != 1 || policy.AllowedRegions[0] != "us-east" {
		t.Fatalf("unexpected normalized regions: %+v", policy.AllowedRegions)
	}
	if len(policy.AllowedProviders) != 1 || policy.AllowedProviders[0] != "aws" {
		t.Fatalf("unexpected providers: %+v", policy.AllowedProviders)
	}

	var fetched infraapi.ProjectPlacement
	if err := cl.Get(context.Background(), req.NamespacedName, &fetched); err != nil {
		t.Fatalf("get placement: %v", err)
	}
	if !containsString(fetched.Finalizers, projectPlacementFinalizer) {
		t.Fatalf("expected finalizer %s", projectPlacementFinalizer)
	}
	if cond := findCondition(fetched.Status.Conditions, "Ready"); cond == nil || cond.Status != metav1.ConditionTrue {
		t.Fatalf("expected Ready condition true, got %+v", fetched.Status.Conditions)
	}
}

func TestProjectPlacementReconcileHandlesDeletion(t *testing.T) {
	t.Parallel()

	scheme, err := newTestScheme()
	if err != nil {
		t.Fatalf("scheme: %v", err)
	}

	deletionTime := metav1.NewTime(time.Now())
	placement := &infraapi.ProjectPlacement{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "proj-delete",
			Namespace:         "aegis-system",
			Finalizers:        []string{projectPlacementFinalizer},
			DeletionTimestamp: &deletionTime,
		},
		Spec: infraapi.ProjectPlacementSpec{
			ProjectID: "proj-delete",
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&infraapi.ProjectPlacement{}).
		WithObjects(placement).
		Build()

	overlay := placementlib.NewPolicyOverlay()
	reconciler := &ProjectPlacementReconciler{
		Client:        cl,
		Log:           zaptest.NewLogger(t).Named("projectplacement"),
		PolicyOverlay: overlay,
	}
	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: placement.Name, Namespace: placement.Namespace}}

	overlay.Set(&infraapi.ProjectPlacementSpec{ProjectID: "proj-delete"})

	if _, err := reconciler.Reconcile(context.Background(), req); err != nil {
		t.Fatalf("reconcile delete: %v", err)
	}

	if _, ok := overlay.Effective("proj-delete"); ok {
		t.Fatalf("expected overlay entry removed")
	}

	var updated infraapi.ProjectPlacement
	err = cl.Get(context.Background(), req.NamespacedName, &updated)
	if err != nil && !apierrors.IsNotFound(err) {
		t.Fatalf("unexpected get after deletion: %v", err)
	}
}
func containsString(values []string, target string) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}

func findCondition(conds []metav1.Condition, condType string) *metav1.Condition {
	for i := range conds {
		if conds[i].Type == condType {
			return &conds[i]
		}
	}
	return nil
}
