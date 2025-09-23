package discovery

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

// DetectKueue reports whether any supported Kueue API group/version is present on the cluster.
func DetectKueue(cfg *rest.Config) (bool, error) {
	disc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return false, err
	}
	versions := []string{"kueue.x-k8s.io/v1beta1", "kueue.x-k8s.io/v1alpha2"}
	for _, gv := range versions {
		if _, err := disc.ServerResourcesForGroupVersion(gv); err == nil {
			return true, nil
		}
	}
	return false, nil
}

// DetectPyTorchGVR returns the available PyTorchJob API group version if present.
func DetectPyTorchGVR(cfg *rest.Config) (string, schema.GroupVersionResource, error) {
	disc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return "", schema.GroupVersionResource{}, err
	}
	candidates := []string{"kubeflow.org/v1", "training.kubeflow.org/v1"}
	for _, gv := range candidates {
		if _, err := disc.ServerResourcesForGroupVersion(gv); err == nil {
			parts := strings.Split(gv, "/")
			if len(parts) != 2 {
				continue
			}
			return gv, schema.GroupVersionResource{
				Group:    parts[0],
				Version:  parts[1],
				Resource: "pytorchjobs",
			}, nil
		}
	}
	return "", schema.GroupVersionResource{}, fmt.Errorf("pytorchjobs CRD not found")
}

// DetectTrainerV2 locates the TrainJob v2 API if installed.
func DetectTrainerV2(cfg *rest.Config) (schema.GroupVersionResource, error) {
	disc, err := discovery.NewDiscoveryClientForConfig(cfg)
	if err != nil {
		return schema.GroupVersionResource{}, err
	}
	const gv = "trainer.kubeflow.org/v1alpha1"
	if _, err := disc.ServerResourcesForGroupVersion(gv); err == nil {
		return schema.GroupVersionResource{
			Group:    "trainer.kubeflow.org",
			Version:  "v1alpha1",
			Resource: "trainjobs",
		}, nil
	}
	return schema.GroupVersionResource{}, fmt.Errorf("trainjobs CRD not found")
}
