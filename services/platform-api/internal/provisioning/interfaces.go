package provisioning

import (
	"context"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
)

// AWSProvisioner encapsulates the infrastructure lifecycle for AWS-based clusters.
type AWSProvisioner interface {
	Provision(ctx context.Context, infra *infraapi.ProjectInfra, spec *infraapi.AWSInfraSpec) (*ProvisionResult, error)
	Destroy(ctx context.Context, infra *infraapi.ProjectInfra, spec *infraapi.AWSInfraSpec) error
}

// ProvisionResult captures the outputs produced by the provisioner.
type ProvisionResult struct {
	Outputs            []infraapi.ClusterOutput
	Kubeconfigs        map[string][]byte
	CostHintUSDPerHour float64
}
