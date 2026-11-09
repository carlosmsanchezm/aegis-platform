package controllers

import (
	"context"
	"testing"
	"time"

	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

func TestAegisClusterStatusSyncCreatesCRDs(t *testing.T) {
	scheme, err := newTestScheme()
	if err != nil {
		t.Fatalf("scheme: %v", err)
	}

	mem := store.NewMemStore()
	mem.UpsertClusterFromRegister(testClusterRegisterRequest())
	mem.UpdateClusterFromHeartbeat(testClusterHeartbeat())

	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "aegis-system"}}
	cl := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&infraapi.AegisCluster{}).
		WithObjects(namespace).
		Build()
	syncer := &AegisClusterStatusSync{
		Client:    cl,
		Log:       zap.NewNop(),
		Store:     mem,
		Interval:  1 * time.Second,
		Namespace: "aegis-system",
	}

	if err := syncer.sync(context.Background()); err != nil {
		t.Fatalf("sync: %v", err)
	}

	cluster := &infraapi.AegisCluster{}
	if err := cl.Get(context.Background(), types.NamespacedName{Name: "cluster-1", Namespace: "aegis-system"}, cluster); err != nil {
		t.Fatalf("expected cluster created: %v", err)
	}
	if cluster.Status.Phase != "HeartbeatOK" {
		t.Fatalf("unexpected phase: %s", cluster.Status.Phase)
	}
}

func TestAegisClusterStatusSyncRemovesStaleCRDs(t *testing.T) {
	scheme, err := newTestScheme()
	if err != nil {
		t.Fatalf("scheme: %v", err)
	}

	mem := store.NewMemStore()
	mem.UpsertClusterFromRegister(testClusterRegisterRequest())
	mem.UpdateClusterFromHeartbeat(testClusterHeartbeat())

	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: "aegis-system"}}
	existing := &infraapi.AegisCluster{
		TypeMeta: metav1.TypeMeta{APIVersion: infraapi.GroupVersion.String(), Kind: "AegisCluster"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "old-cluster",
			Namespace: "aegis-system",
		},
	}

	cl := fake.NewClientBuilder().WithScheme(scheme).
		WithStatusSubresource(&infraapi.AegisCluster{}).
		WithObjects(namespace, existing).
		Build()
	syncer := &AegisClusterStatusSync{
		Client:    cl,
		Log:       zap.NewNop(),
		Store:     mem,
		Namespace: "aegis-system",
	}

	if err := syncer.sync(context.Background()); err != nil {
		t.Fatalf("sync: %v", err)
	}

	if err := cl.Get(context.Background(), types.NamespacedName{Name: "old-cluster", Namespace: "aegis-system"}, &infraapi.AegisCluster{}); !apierrors.IsNotFound(err) {
		t.Fatalf("expected old cluster removed, err=%v", err)
	}
}

func testClusterRegisterRequest() *aegis.ClusterRegisterRequest {
	return &aegis.ClusterRegisterRequest{
		ClusterId: "cluster-1",
		Provider:  "AWS",
		Region:    "us-east-1",
		Labels: map[string]string{
			"aegis.yourorg.dev/projectId": "proj-1",
		},
	}
}

func testClusterHeartbeat() *aegis.ClusterHeartbeat {
	return &aegis.ClusterHeartbeat{
		ClusterId: "cluster-1",
		AvailableFlavors: []*aegis.Flavor{
			{Name: "a10"},
		},
		TtfGpuSecondsP50: 12.0,
	}
}
