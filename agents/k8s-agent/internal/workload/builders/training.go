package builders

import (
	"strconv"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// TrainingOptions describes inputs for building a PyTorchJob manifest.
type TrainingOptions struct {
	APIVersion    string
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
