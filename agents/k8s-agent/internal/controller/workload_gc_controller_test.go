package controller

import (
	"context"
	"errors"
	"testing"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	aegisv1alpha2 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha2"
)

type staticWorkloadLister struct {
	ids []string
	err error
}

func (s staticWorkloadLister) ListClusterWorkloadIDs(_ context.Context, _ string) ([]string, error) {
	return s.ids, s.err
}

func TestWorkloadGCControllerDeletesOrphanedWorkspaces(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := aegisv1alpha2.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	ws1 := &aegisv1alpha2.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ws-1",
			Namespace: "default",
			Labels:    map[string]string{labelWorkloadID: "aegis-w-1"},
		},
	}
	ws2 := &aegisv1alpha2.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ws-2",
			Namespace: "default",
			Labels:    map[string]string{labelWorkloadID: "w-2"},
		},
	}
	wsNoLabel := &aegisv1alpha2.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ws-nolabel",
			Namespace: "default",
		},
	}

	cli := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ws1, ws2, wsNoLabel).Build()
	gc := &WorkloadGCController{
		Client:       cli,
		cpClient:     staticWorkloadLister{ids: []string{"w-1"}},
		clusterID:    "cluster-1",
		dryRun:       false,
		maxDeletions: 10,
	}

	gc.runOnce(context.Background())

	var got aegisv1alpha2.Workspace
	if err := cli.Get(context.Background(), client.ObjectKeyFromObject(ws1), &got); err != nil {
		t.Fatalf("expected ws-1 to remain, got error: %v", err)
	}
	if err := cli.Get(context.Background(), client.ObjectKeyFromObject(wsNoLabel), &got); err != nil {
		t.Fatalf("expected ws-nolabel to remain, got error: %v", err)
	}
	if err := cli.Get(context.Background(), client.ObjectKeyFromObject(ws2), &got); !apierrors.IsNotFound(err) {
		t.Fatalf("expected ws-2 to be deleted, got error: %v", err)
	}
}

func TestWorkloadGCControllerSkipsDeletionOnHubError(t *testing.T) {
	scheme := runtime.NewScheme()
	if err := aegisv1alpha2.AddToScheme(scheme); err != nil {
		t.Fatalf("add scheme: %v", err)
	}

	ws := &aegisv1alpha2.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ws-1",
			Namespace: "default",
			Labels:    map[string]string{labelWorkloadID: "w-1"},
		},
	}

	cli := fake.NewClientBuilder().WithScheme(scheme).WithObjects(ws).Build()
	gc := &WorkloadGCController{
		Client:    cli,
		cpClient:  staticWorkloadLister{err: errors.New("hub unavailable")},
		clusterID: "cluster-1",
	}

	gc.runOnce(context.Background())

	var got aegisv1alpha2.Workspace
	if err := cli.Get(context.Background(), client.ObjectKeyFromObject(ws), &got); err != nil {
		t.Fatalf("expected workspace to remain when hub call fails, got: %v", err)
	}
}
