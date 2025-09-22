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
	chip string
	mig  string
	rdma bool
}

// DiscoverFlavors queries Nodes and infers available flavors from allocatable resources.
// - MIG: resource like "nvidia.com/mig-1g.10gb" => flavor with MigProfile set.
// - Non-MIG: if "nvidia.com/gpu" > 0 => add a generic flavor with best-effort chip hint.
// Falls back to StaticFlavors() if discovery yields nothing.
func DiscoverFlavors(ctx context.Context, cs *kubernetes.Clientset) ([]*aegis.Flavor, error) {
	if cs == nil {
		return StaticFlavors(), nil
	}

	nodes, err := cs.CoreV1().Nodes().List(ctx, metav1.ListOptions{LabelSelector: labels.Everything().String()})
	if err != nil {
		return nil, err
	}

	seen := map[flavorKey]struct{}{}

	for i := range nodes.Items {
		n := &nodes.Items[i]
		captureNodeFlavors(seen, n)
	}

	if len(seen) == 0 {
		return StaticFlavors(), nil
	}

	flavors := make([]*aegis.Flavor, 0, len(seen))
	for k := range seen {
		flavors = append(flavors, &aegis.Flavor{
			Name:         k.name,
			Chip:         k.chip,
			MigProfile:   k.mig,
			RdmaRequired: k.rdma,
		})
	}
	sort.Slice(flavors, func(i, j int) bool { return flavors[i].GetName() < flavors[j].GetName() })
	return flavors, nil
}

func captureNodeFlavors(acc map[flavorKey]struct{}, node *corev1.Node) {
	if node == nil {
		return
	}

	alloc := node.Status.Allocatable

	// MIG profiles
	for resName, qty := range alloc {
		rn := string(resName)
		if !strings.HasPrefix(rn, "nvidia.com/") || qty.Value() <= 0 {
			continue
		}
		if strings.HasPrefix(rn, "nvidia.com/mig-") {
			profile := strings.TrimPrefix(rn, "nvidia.com/mig-")
			name := "mig-" + profile
			acc[flavorKey{name: name, mig: profile}] = struct{}{}
			continue
		}
		if rn == "nvidia.com/gpu" {
			chip := node.Labels["nvidia.com/gpu.product"]
			if chip == "" {
				chip = node.Labels["feature.node.kubernetes.io/pci-10de.present"]
			}
			acc[flavorKey{name: "gpu-1x", chip: chip}] = struct{}{}
		}
	}
}
