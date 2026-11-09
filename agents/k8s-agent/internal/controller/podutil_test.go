package controller

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	aegisv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
)

func TestApplyResourceHints_DefaultGPU(t *testing.T) {
	pod := &corev1.PodSpec{Containers: []corev1.Container{{}}}
	hints := &aegisv1alpha1.ResourceHints{GpuCount: 1}

	ApplyResourceHints(pod, hints)

	if pod.NodeSelector["aegis.io/gpu-flavor"] != "nvidia-tesla-t4" {
		t.Fatalf("expected default gpu flavor label, got %q", pod.NodeSelector["aegis.io/gpu-flavor"])
	}
	if !hasToleration(pod.Tolerations, "nvidia.com/gpu") {
		t.Fatalf("expected nvidia.com/gpu toleration, got %#v", pod.Tolerations)
	}
}

func TestApplyResourceHints_MIG(t *testing.T) {
	pod := &corev1.PodSpec{Containers: []corev1.Container{{}}}
	hints := &aegisv1alpha1.ResourceHints{GpuCount: 1, ResourceName: "nvidia.com/mig-1g.10gb"}

	ApplyResourceHints(pod, hints)

	if pod.NodeSelector["aegis.io/gpu-flavor"] != "nvidia-a10g-mig" {
		t.Fatalf("expected mig gpu flavor label, got %q", pod.NodeSelector["aegis.io/gpu-flavor"])
	}
	if !hasToleration(pod.Tolerations, "nvidia.com/mig-1g.10gb") {
		t.Fatalf("expected mig toleration, got %#v", pod.Tolerations)
	}
	if hasToleration(pod.Tolerations, "nvidia.com/gpu") {
		t.Fatalf("did not expect generic gpu toleration when mig toleration is set")
	}
}
