package builders

import (
	"strconv"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// TrainingOptions describes inputs for building training job manifests.
type TrainingOptions struct {
	APIVersion    string // For PyTorchJob compatibility (deprecated)
	Namespace     string
	Name          string
	Image         string
	Command       []string
	Flavor        string
	Workers       int
	GPUsPerWorker int
	DryRun        bool
	Hints         *GPUHints
}

// BuildTrainingJob creates a simple Kubernetes batch/v1 Job for training workloads.
// This is the preferred method for training workloads when PyTorchJob/distributed training is not needed.
func BuildTrainingJob(opts TrainingOptions) *batchv1.Job {
	container := corev1.Container{
		Name:  "training",
		Image: opts.Image,
	}

	if len(opts.Command) > 0 {
		container.Command = opts.Command
	}

	// Add GPU resources only if explicitly requested
	if !opts.DryRun {
		resources := corev1.ResourceRequirements{
			Limits:   corev1.ResourceList{},
			Requests: corev1.ResourceList{},
		}
		resourcesRequested := false

		// Check if GPU count from hints is > 0
		gpuCount := int32(0)
		if opts.Hints != nil {
			gpuCount = opts.Hints.GPUCount
		}

		// Only add GPU requests if count > 0
		if gpuCount > 0 {
			resName := ""
			if opts.Hints != nil && opts.Hints.ResourceName != "" {
				resName = opts.Hints.ResourceName
			}
			if resName == "" {
				resName = AutoGPUResource(opts.Flavor)
			}
			if resName == "" {
				resName = "nvidia.com/gpu"
			}
			quantity := resource.MustParse(strconv.Itoa(int(gpuCount)))
			resources.Limits[corev1.ResourceName(resName)] = quantity
			resources.Requests[corev1.ResourceName(resName)] = quantity
			resourcesRequested = true
		}

		// Add CPU/Memory hints if provided
		if opts.Hints != nil && opts.Hints.CpuCoresRequest != nil && *opts.Hints.CpuCoresRequest != "" {
			cpuQty := resource.MustParse(*opts.Hints.CpuCoresRequest)
			resources.Requests[corev1.ResourceCPU] = cpuQty
			resources.Limits[corev1.ResourceCPU] = cpuQty
			resourcesRequested = true
		}
		if opts.Hints != nil && opts.Hints.MemoryRequest != nil && *opts.Hints.MemoryRequest != "" {
			memQty := resource.MustParse(*opts.Hints.MemoryRequest)
			resources.Requests[corev1.ResourceMemory] = memQty
			resources.Limits[corev1.ResourceMemory] = memQty
			resourcesRequested = true
		}

		if resourcesRequested {
			container.Resources = resources
		}
	}

	jobLabels := map[string]string{
		"aegis.workload/id": opts.Name,
		"aegis.job/type":    "training",
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      opts.Name,
			Namespace: opts.Namespace,
			Labels:    jobLabels,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: int32Ptr(0),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: jobLabels},
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					Containers:    []corev1.Container{container},
				},
			},
		},
	}
}

// BuildPyTorchJob produces an unstructured PyTorchJob manifest using the shared executor logic.
func BuildPyTorchJob(opts TrainingOptions) *unstructured.Unstructured {
	container := map[string]interface{}{
		"name":  "pytorch",
		"image": opts.Image,
	}
	if len(opts.Command) > 0 {
		container["command"] = stringSliceToAny(opts.Command)
	}
	if !opts.DryRun {
		resName := ""
		count := max(1, opts.GPUsPerWorker)
		if opts.Hints != nil {
			if rn := opts.Hints.ResourceName; rn != "" {
				resName = rn
			}
			if opts.Hints.GPUCount > 0 {
				count = int(opts.Hints.GPUCount)
			}
		}
		if resName == "" {
			resName = AutoGPUResource(opts.Flavor)
			if resName == "" {
				resName = "nvidia.com/gpu"
			}
		}
		quantity := resource.MustParse(strconv.Itoa(max(1, count)))
		limits := map[string]interface{}{resName: quantity.String()}
		container["resources"] = map[string]interface{}{
			"limits":   limits,
			"requests": limits,
		}
	}

	template := map[string]interface{}{
		"spec": map[string]interface{}{
			"restartPolicy": "Never",
			"containers":    []interface{}{container},
		},
	}

	replicas := map[string]interface{}{
		"Master": map[string]interface{}{
			"replicas": int64(1),
			"template": template,
		},
	}

	if opts.Workers > 1 {
		replicas["Worker"] = map[string]interface{}{
			"replicas": int64(opts.Workers - 1),
			"template": template,
		}
	}

	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": opts.APIVersion,
			"kind":       "PyTorchJob",
			"metadata": map[string]interface{}{
				"name":      opts.Name,
				"namespace": opts.Namespace,
				"labels": map[string]interface{}{
					"aegis.workload/id": opts.Name,
				},
			},
			"spec": map[string]interface{}{
				"pytorchReplicaSpecs": replicas,
			},
		},
	}
}

func stringSliceToAny(in []string) []interface{} {
	out := make([]interface{}, len(in))
	for i := range in {
		out[i] = in[i]
	}
	return out
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
