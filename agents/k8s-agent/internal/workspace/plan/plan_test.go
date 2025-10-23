package plan

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	aegisv1alpha2 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha2"
)

func TestPlannerAppliesTemplateDefaults(t *testing.T) {
	ws := &aegisv1alpha2.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "sample"},
		Spec: aegisv1alpha2.WorkspaceSpec{
			ProjectRef: "demo",
			ProfileRef: "vscode-python",
		},
	}

	result, err := NewPlanner().Plan(ws)
	if err != nil {
		t.Fatalf("Plan returned error: %v", err)
	}

	if result.WorkloadSpec.Workspace == nil {
		t.Fatalf("workspace spec not populated")
	}
	got := result.WorkloadSpec.Workspace
	if got.Image != "carlosmsanchez/aegis-workspace-vscode:latest" {
		t.Fatalf("unexpected image %q", got.Image)
	}
	if !got.Interactive {
		t.Fatalf("expected interactive workspace")
	}
	if len(got.Ports) != 2 || got.Ports[0] != 22 || got.Ports[1] != 11111 {
		t.Fatalf("unexpected ports: %#v", got.Ports)
	}
}

func TestPlannerRequiresCustomExecution(t *testing.T) {
	ws := &aegisv1alpha2.Workspace{
		ObjectMeta: metav1.ObjectMeta{Name: "custom"},
		Spec: aegisv1alpha2.WorkspaceSpec{
			ProjectRef: "demo",
			ProfileRef: "custom",
		},
	}

	if _, err := NewPlanner().Plan(ws); err == nil {
		t.Fatalf("expected error when custom profile has no execution overrides")
	}

	ws.Spec.Execution = &aegisv1alpha2.WorkspaceExecution{Image: "alpine:3.19", Ports: []int32{22}}
	if _, err := NewPlanner().Plan(ws); err != nil {
		t.Fatalf("unexpected error for custom execution: %v", err)
	}
}
