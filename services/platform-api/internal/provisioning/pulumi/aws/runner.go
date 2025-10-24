package aws

import (
	"context"
	"fmt"
	"strings"

	"go.uber.org/zap"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/provisioning"
)

// Runner provides a placeholder AWS infrastructure provisioner backed by Pulumi Automation.
type Runner struct {
	log *zap.Logger
}

// NewRunner returns a new Automation runner.
func NewRunner(log *zap.Logger) *Runner {
	if log == nil {
		log = zap.NewNop()
	}
	return &Runner{log: log.Named("aws-provisioner")}
}

// Provision simulates AWS infrastructure provisioning. The implementation produces deterministic
// kubeconfig stubs so that higher layers can exercise their reconciliation logic without needing
// live AWS credentials.
func (r *Runner) Provision(ctx context.Context, infra *infraapi.ProjectInfra, spec *infraapi.AWSInfraSpec) (*provisioning.ProvisionResult, error) {
	_ = ctx
	if infra == nil {
		return nil, fmt.Errorf("project infra required")
	}
	if spec == nil {
		return nil, fmt.Errorf("aws spec required")
	}

	projectID := strings.TrimSpace(infra.Spec.ProjectID)
	region := strings.TrimSpace(infra.Spec.Region)

	outputs := []infraapi.ClusterOutput{}
	kubeconfigs := map[string][]byte{}

	recordCluster := func(name string) {
		clusterID := buildClusterID(projectID, region, name)
		secretKey := fmt.Sprintf("%s.kubeconfig", clusterID)
		outputs = append(outputs, infraapi.ClusterOutput{
			ClusterID:           clusterID,
			Name:                clusterID,
			Region:              region,
			KubeconfigSecretKey: secretKey,
		})
		kubeconfigs[clusterID] = []byte(renderKubeconfig(clusterID, region))
	}

	if strings.TrimSpace(spec.ClusterName) != "" {
		recordCluster(spec.ClusterName)
	}
	for _, extra := range spec.AdditionalClusters {
		if strings.TrimSpace(extra.ClusterName) == "" {
			continue
		}
		recordCluster(extra.ClusterName)
	}
	if len(outputs) == 0 {
		return nil, fmt.Errorf("no clusters defined in aws spec")
	}

	cost := estimateCost(spec)
	r.log.Info("simulated aws provisioning complete",
		zap.String("project", projectID),
		zap.String("region", region),
		zap.Int("clusters", len(outputs)),
		zap.Float64("cost_hint_usd_per_hour", cost))

	return &provisioning.ProvisionResult{
		Outputs:            outputs,
		Kubeconfigs:        kubeconfigs,
		CostHintUSDPerHour: cost,
	}, nil
}

// Destroy logs the teardown request; no real cloud resources are affected by the stub implementation.
func (r *Runner) Destroy(ctx context.Context, infra *infraapi.ProjectInfra, spec *infraapi.AWSInfraSpec) error {
	_ = ctx
	if infra == nil || spec == nil {
		return nil
	}
	projectID := strings.TrimSpace(infra.Spec.ProjectID)
	r.log.Info("simulated aws destroy invoked",
		zap.String("project", projectID),
		zap.String("region", infra.Spec.Region))
	return nil
}

func buildClusterID(projectID, region, name string) string {
	parts := []string{}
	if projectID != "" {
		parts = append(parts, sanitize(projectID))
	}
	if region != "" {
		parts = append(parts, sanitize(region))
	}
	parts = append(parts, sanitize(name))
	return strings.Join(parts, "-")
}

func sanitize(in string) string {
	trimmed := strings.ToLower(strings.TrimSpace(in))
	replacer := strings.NewReplacer(" ", "-", "_", "-", ".", "-")
	trimmed = replacer.Replace(trimmed)
	for strings.Contains(trimmed, "--") {
		trimmed = strings.ReplaceAll(trimmed, "--", "-")
	}
	return strings.Trim(trimmed, "-")
}

func renderKubeconfig(clusterID, region string) string {
	return fmt.Sprintf(`apiVersion: v1
clusters:
- cluster:
    server: https://%s.%s.aegis.local
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s
kind: Config
users:
- name: %s
  user:
    token: placeholder
`, clusterID, region, clusterID, clusterID, clusterID, clusterID, clusterID, clusterID)
}

func estimateCost(spec *infraapi.AWSInfraSpec) float64 {
	if spec == nil {
		return 0
	}
	cost := 0.0
	for _, pool := range spec.NodePools {
		cost += float64(pool.MaxSize)
	}
	for _, cluster := range spec.AdditionalClusters {
		for _, pool := range cluster.NodePools {
			cost += float64(pool.MaxSize)
		}
	}
	return cost
}
