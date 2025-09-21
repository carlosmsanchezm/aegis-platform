package discovery

import (
	"context"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

type flavorKey struct {
	name string
}

const (
	labelGPUProduct   = "nvidia.com/gpu.product"
	labelMIGProfile   = "nvidia.com/mig.profile"
	labelRDMA         = "nvidia.com/rdma.present"
	resourceGPUPrefix = "nvidia.com/mig-"
)

// DiscoverFlavors inspects the Kubernetes node inventory and derives a list of
// advertised flavors. It falls back to the static defaults if no GPU metadata
// is present so that sandbox environments with fake nodes still return
// reasonable values.
func DiscoverFlavors(ctx context.Context, cs *kubernetes.Clientset) ([]*aegis.Flavor, error) {
	if cs == nil {
		return StaticFlavors(), nil
	}

	nodeList, err := cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{
		LabelSelector: labels.Everything().String(),
	})
	if err != nil {
		return nil, err
	}

	seen := map[flavorKey]*aegis.Flavor{}

	for _, node := range nodeList.Items {
		captureNodeFlavors(seen, &node)
	}

	if len(seen) == 0 {
		return StaticFlavors(), nil
	}

	flavors := make([]*aegis.Flavor, 0, len(seen))
	for _, fl := range seen {
		flavors = append(flavors, fl)
	}
	sort.Slice(flavors, func(i, j int) bool { return flavors[i].GetName() < flavors[j].GetName() })
	return flavors, nil
}

func captureNodeFlavors(acc map[flavorKey]*aegis.Flavor, node *corev1.Node) {
	if node == nil {
		return
	}

	chip := normalizeName(node.Labels[labelGPUProduct])
	migProfile := normalizeProfile(node.Labels[labelMIGProfile])
	rdma := strings.EqualFold(node.Labels[labelRDMA], "true")

	if migProfile != "" {
		name := chip
		if name != "" && migProfile != "" {
			name = name + "-" + migProfile
		} else if migProfile != "" {
			name = migProfile
		}
		acc[flavorKey{name: name}] = &aegis.Flavor{Name: name, Chip: chip, MigProfile: migProfile, RdmaRequired: rdma}
	}

	for resName := range node.Status.Capacity {
		res := string(resName)
		if strings.HasPrefix(res, resourceGPUPrefix) {
			profile := strings.TrimPrefix(res, resourceGPUPrefix)
			normalized := normalizeProfile(profile)
			name := normalized
			if chip != "" {
				name = chip + "-" + normalized
			}
			acc[flavorKey{name: name}] = &aegis.Flavor{Name: name, Chip: chip, MigProfile: profile, RdmaRequired: rdma}
		}
	}

}

func normalizeName(in string) string {
	if in == "" {
		return ""
	}
	cleaned := strings.ToLower(in)
	cleaned = strings.ReplaceAll(cleaned, "nvidia-", "")
	cleaned = strings.ReplaceAll(cleaned, "nvidia ", "")
	cleaned = strings.ReplaceAll(cleaned, " ", "-")
	cleaned = strings.ReplaceAll(cleaned, "_", "-")
	return cleaned
}

func normalizeProfile(in string) string {
	if in == "" {
		return ""
	}
	cleaned := strings.ToLower(in)
	cleaned = strings.ReplaceAll(cleaned, ".", "-")
	cleaned = strings.ReplaceAll(cleaned, "_", "-")
	cleaned = strings.TrimPrefix(cleaned, "mig-")
	return cleaned
}
