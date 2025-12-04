package controller

import (
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	aegisv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
)

// ApplyResourceHints mutates the provided PodSpec with GPU related hints.
func ApplyResourceHints(pod *corev1.PodSpec, hints *aegisv1alpha1.ResourceHints) {
	if pod == nil || hints == nil {
		return
	}

	needsGPU := hints.GpuCount > 0 || hints.ResourceName != ""
	if !needsGPU {
		return
	}

	resourceName := hints.ResourceName
	if resourceName == "" {
		resourceName = "nvidia.com/gpu"
	}

	if len(pod.Containers) > 0 {
		container := &pod.Containers[0]
		if container.Resources.Requests == nil {
			container.Resources.Requests = corev1.ResourceList{}
		}
		if container.Resources.Limits == nil {
			container.Resources.Limits = corev1.ResourceList{}
		}

		qty := resourceQuantity(int64(hints.GpuCount))
		container.Resources.Requests[corev1.ResourceName(resourceName)] = *qty
		container.Resources.Limits[corev1.ResourceName(resourceName)] = *qty
	}

	if pod.NodeSelector == nil {
		pod.NodeSelector = map[string]string{}
	}
	flavorLabel := "nvidia-tesla-t4"
	cpuHint := ""
	if hints.CpuCoresRequest != nil {
		cpuHint = strings.TrimSpace(*hints.CpuCoresRequest)
	}
	memHint := ""
	if hints.MemoryRequest != nil {
		memHint = strings.ToLower(strings.TrimSpace(*hints.MemoryRequest))
	}
	if strings.Contains(resourceName, "mig-1g.10gb") || strings.Contains(resourceName, "a10") || cpuHint == "8" || strings.HasPrefix(memHint, "32g") {
		flavorLabel = "nvidia-a10g-mig"
	}
	pod.NodeSelector["aegis.io/gpu-flavor"] = flavorLabel

	if strings.Contains(resourceName, "mig-1g.10gb") {
		if !hasToleration(pod.Tolerations, "nvidia.com/mig-1g.10gb") {
			pod.Tolerations = append(pod.Tolerations, corev1.Toleration{
				Key:      "nvidia.com/mig-1g.10gb",
				Operator: corev1.TolerationOpExists,
				Effect:   corev1.TaintEffectNoSchedule,
			})
		}
	} else if !hasToleration(pod.Tolerations, "nvidia.com/gpu") {
		pod.Tolerations = append(pod.Tolerations, corev1.Toleration{
			Key:      "nvidia.com/gpu",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoSchedule,
		})
	}
}

func resourceQuantity(count int64) *resource.Quantity {
	if count <= 0 {
		count = 1
	}
	qty := resource.NewQuantity(count, resource.DecimalSI)
	return qty
}

func hasToleration(tolerations []corev1.Toleration, key string) bool {
	for _, tol := range tolerations {
		if tol.Key == key {
			return true
		}
	}
	return false
}
