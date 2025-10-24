package controllers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.uber.org/zap/zaptest"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/provisioning"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

type fakeProvisioner struct {
	result provisioning.ProvisionResult
	err    error
	calls  int
}

func (f *fakeProvisioner) Provision(ctx context.Context, infra *infraapi.ProjectInfra, spec *infraapi.AWSInfraSpec) (*provisioning.ProvisionResult, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return &f.result, nil
}

func (f *fakeProvisioner) Destroy(ctx context.Context, infra *infraapi.ProjectInfra, spec *infraapi.AWSInfraSpec) error {
	return nil
}

func TestProjectInfraReconcile_ImportSecret(t *testing.T) {
	t.Parallel()

	scheme, err := newTestScheme()
	if err != nil {
		t.Fatalf("scheme: %v", err)
	}

	ns := fmt.Sprintf("infra-%d", time.Now().UnixNano())
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}

	importSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: "import", Namespace: ns},
		Data:       map[string][]byte{"cluster-a.kubeconfig": []byte("kubeconfig")},
	}

	infra := &infraapi.ProjectInfra{
		ObjectMeta: metav1.ObjectMeta{Name: "proj", Namespace: ns},
		Spec: infraapi.ProjectInfraSpec{
			ProjectID: "proj",
			Provider:  "aws",
			Region:    "us-east-1",
			ImportKubeconfigSecretRef: &corev1.SecretReference{
				Name: importSecret.Name,
			},
		},
	}

	cl := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&infraapi.ProjectInfra{}).
		WithObjects(namespace, importSecret, infra).
		Build()

	reconciler := &ProjectInfraReconciler{
		Client:                    cl,
		Scheme:                    scheme,
		Log:                       zaptest.NewLogger(t).Named("projectinfra"),
		Store:                     store.NewMemStore(),
		KubeconfigSecretName:      "bundle",
		KubeconfigSecretNamespace: ns,
	}

	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: infra.Name, Namespace: infra.Namespace}}
	for i := 0; i < 3; i++ {
		if _, err := reconciler.Reconcile(context.Background(), req); err != nil {
			t.Fatalf("reconcile attempt %d failed: %v", i, err)
		}
	}

	aggregated := &corev1.Secret{}
	if err := cl.Get(context.Background(), types.NamespacedName{Name: "bundle", Namespace: ns}, aggregated); err != nil {
		t.Fatalf("fetch aggregated secret: %v", err)
	}
	if _, ok := aggregated.Data["cluster-a.kubeconfig"]; !ok {
		t.Fatalf("expected kubeconfig key in aggregated secret")
	}

	cluster := &infraapi.AegisCluster{}
	if err := cl.Get(context.Background(), types.NamespacedName{Name: "cluster-a", Namespace: ns}, cluster); err != nil {
		t.Fatalf("expected AegisCluster: %v", err)
	}
	if cluster.Spec.ProjectID != "proj" || cluster.Spec.Region != "us-east-1" {
		t.Fatalf("unexpected cluster spec: %+v", cluster.Spec)
	}
}

func TestProjectInfraReconcile_AWSProvisioner(t *testing.T) {
	t.Parallel()

	scheme, err := newTestScheme()
	if err != nil {
		t.Fatalf("scheme: %v", err)
	}

	ns := fmt.Sprintf("infra-aws-%d", time.Now().UnixNano())
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: ns}}

	infra := &infraapi.ProjectInfra{
		ObjectMeta: metav1.ObjectMeta{Name: "proj-aws", Namespace: ns},
		Spec: infraapi.ProjectInfraSpec{
			ProjectID: "proj-aws",
			Provider:  "aws",
			Region:    "us-west-2",
			Aws: &infraapi.AWSInfraSpec{
				ClusterName: "primary",
			},
		},
	}

	fc := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&infraapi.ProjectInfra{}).
		WithObjects(namespace, infra).
		Build()

	provisioner := &fakeProvisioner{result: provisioning.ProvisionResult{
		Outputs: []infraapi.ClusterOutput{
			{
				ClusterID:           "proj-aws-primary",
				Name:                "primary",
				Region:              "us-west-2",
				KubeconfigSecretKey: "proj-aws-primary.kubeconfig",
			},
		},
		Kubeconfigs: map[string][]byte{
			"proj-aws-primary": []byte("kubeconfig"),
		},
		CostHintUSDPerHour: 42.0,
	}}

	reconciler := &ProjectInfraReconciler{
		Client:                    fc,
		Scheme:                    scheme,
		Log:                       zaptest.NewLogger(t).Named("projectinfra"),
		Provisioner:               provisioner,
		Store:                     store.NewMemStore(),
		KubeconfigSecretName:      "bundle",
		KubeconfigSecretNamespace: ns,
	}

	req := reconcile.Request{NamespacedName: types.NamespacedName{Name: infra.Name, Namespace: infra.Namespace}}
	for i := 0; i < 3; i++ {
		if _, err := reconciler.Reconcile(context.Background(), req); err != nil {
			t.Fatalf("reconcile attempt %d failed: %v", i, err)
		}
	}

	if provisioner.calls == 0 {
		t.Fatalf("expected provisioner to be invoked")
	}

	aggregated := &corev1.Secret{}
	if err := fc.Get(context.Background(), types.NamespacedName{Name: "bundle", Namespace: ns}, aggregated); err != nil {
		t.Fatalf("aggregated secret: %v", err)
	}
	if _, ok := aggregated.Data["proj-aws-primary.kubeconfig"]; !ok {
		t.Fatalf("expected kubeconfig in aggregated secret")
	}

	var updated infraapi.ProjectInfra
	if err := fc.Get(context.Background(), req.NamespacedName, &updated); err != nil {
		t.Fatalf("fetch infra: %v", err)
	}
	if cond := findCondition(updated.Status.Conditions, conditionReady); cond == nil || cond.Status != metav1.ConditionTrue {
		t.Fatalf("expected Ready condition true, got %+v", updated.Status.Conditions)
	}
	if updated.Status.CostHintUSDPerHour != 42.0 {
		t.Fatalf("unexpected cost hint: %f", updated.Status.CostHintUSDPerHour)
	}
}
