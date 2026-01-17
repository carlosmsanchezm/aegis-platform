package aws

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"hash/fnv"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/exp/slices"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	awsec2 "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ec2"
	awseks "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/eks"
	awsiam "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	kubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	kubecorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optrefresh"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/provisioning"
	"github.com/yourorg/aegis/services/platform-api/internal/provisioning/certmanager"
	"github.com/yourorg/aegis/services/platform-api/internal/provisioning/observability"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

const (
	projectNamePrefix          = "aegis-platform"
	defaultStackPrefix         = "aegis"
	defaultHelmNamespace       = "aegis-system"
	defaultWorkloadsNamespace  = "aegis-workloads"
	defaultSpokeReleaseName    = "aegis-spoke"
	defaultSpokeProxyNodePort  = 31484 // Default NodePort for spoke-proxy (configurable via env)
	envAegisWorkloadsNamespace = "AEGIS_WORKLOADS_NAMESPACE"
	envPulumiBackendURL        = "PULUMI_BACKEND_URL"
	envAegisSpokeChart         = "AEGIS_SPOKE_CHART"
	envAegisSpokeChartVersion  = "AEGIS_SPOKE_CHART_VERSION"
	envAegisSpokeRepo          = "AEGIS_SPOKE_REPO"
	envAegisSpokeValuesFile    = "AEGIS_SPOKE_VALUES_FILE"
	envAegisPlatformEndpoint   = "AEGIS_PLATFORM_API_ENDPOINT"
	envAegisPlatformCABase64   = "AEGIS_PLATFORM_API_CA_B64"
	envAegisPlatformCAFile     = "AEGIS_PLATFORM_API_CA_FILE"
	envAegisPlatformInsecure   = "AEGIS_PLATFORM_API_GRPC_INSECURE"
	envAegisSpokeImageRepo     = "AEGIS_SPOKE_IMAGE_REPO"
	envAegisSpokeImageTag      = "AEGIS_SPOKE_IMAGE_TAG"
	envAegisSpokeProxyNodePort = "AEGIS_SPOKE_PROXY_NODEPORT"
	envAegisSpokeProxyHost     = "AEGIS_SPOKE_PROXY_HOST"   // Stable proxy hostname (e.g., spoke-proxy.52.1.2.3.nip.io:443)
	envAegisSpokeProxyCACert   = "AEGIS_SPOKE_PROXY_CA_CERT" // Path to spoke-proxy CA certificate for signing
	envAegisSpokeProxyCAKey    = "AEGIS_SPOKE_PROXY_CA_KEY"  // Path to spoke-proxy CA private key for signing
	envAegisPulumiSkipRefresh  = "AEGIS_PULUMI_SKIP_REFRESH"
	envAegisPulumiSkipApply    = "AEGIS_PULUMI_SKIP_APPLY"
	envAegisPulumiWorkdir      = "AEGIS_PULUMI_WORKDIR"
	defaultSpokeReleaseTimeout = 15 * time.Minute
)

// Certain AZs are not eligible for EKS control planes even though they may exist in EC2.
var unsupportedControlPlaneAZs = map[string]map[string]bool{
	"us-east-1": {
		"us-east-1e": true,
	},
}

// Runner provisions AWS infrastructure via Pulumi Automation.
type Runner struct {
	log  *zap.Logger
	sink store.ProvisioningLogSink
}

// NewRunner returns a new automation-backed AWS runner.
func NewRunner(log *zap.Logger, sink store.ProvisioningLogSink) *Runner {
	if log == nil {
		log = zap.NewNop()
	}
	return &Runner{log: log.Named("aws-provisioner"), sink: sink}
}

// Provision reconciles AWS infrastructure described by the supplied ProjectInfra.
func (r *Runner) Provision(ctx context.Context, infra *infraapi.ProjectInfra, spec *infraapi.AWSInfraSpec) (*provisioning.ProvisionResult, error) {
	if infra == nil {
		return nil, errors.New("project infra required")
	}
	if spec == nil {
		return nil, errors.New("aws spec required")
	}
	if strings.EqualFold(string(spec.Mode), string(infraapi.AWSProvisionModeImport)) {
		return nil, errors.New("provision called for import-mode spec")
	}

	stack, programCfg, err := r.NewStack(ctx, infra, spec)
	if err != nil {
		return nil, err
	}

	// Best-effort unlock before any operations in case a previous run left a remote lock.
	_ = stack.Cancel(ctx)
	infraName := strings.TrimSpace(infra.Name)
	r.clearPulumiLock(programCfg.ProjectID, r.stackNameForInfra(infraName, programCfg.ProjectID, programCfg.Region))

	jobID, clusterID := jobMetadata(infra)
	progressWriter := &logWriter{
		log:       r.log,
		sink:      r.sink,
		jobID:     jobID,
		projectID: strings.TrimSpace(infra.Spec.ProjectID),
		clusterID: clusterID,
		phase:     "Provisioning",
	}

	if skip := strings.EqualFold(os.Getenv(envAegisPulumiSkipRefresh), "true"); !skip {
		// Use ClearPendingCreates to handle any stale pending CREATE operations from
		// previous interrupted runs. This ensures state is properly reconciled with AWS.
		if _, err := stack.Refresh(ctx,
			optrefresh.ProgressStreams(progressWriter),
			optrefresh.ClearPendingCreates(),
		); err != nil {
			if !isLockError(err) || !r.retryAfterCancel(ctx, stack, func() error {
				_, retryErr := stack.Refresh(ctx,
					optrefresh.ProgressStreams(progressWriter),
					optrefresh.ClearPendingCreates(),
				)
				return retryErr
			}) {
				return nil, fmt.Errorf("pulumi refresh: %w", err)
			}
		}
	}

	if strings.EqualFold(os.Getenv(envAegisPulumiSkipApply), "true") {
		stackName := r.stackName(programCfg.ProjectID, programCfg.Region)
		r.log.Warn("skipping pulumi up due to environment override", zap.String("stack", stackName))
	} else {
		// Use runUpWithRetry which handles pending operations and "already exists" errors
		if err := r.runUpWithRetry(ctx, stack, progressWriter); err != nil {
			return nil, fmt.Errorf("pulumi up: %w", err)
		}
	}

	outputs, err := stack.Outputs(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieve stack outputs: %w", err)
	}

	result, err := r.translateOutputs(outputs, programCfg.Region)
	if err != nil {
		return nil, err
	}

	result.CostHintUSDPerHour = estimateCost(spec)
	r.log.Info("aws provisioning completed",
		zap.String("project", programCfg.ProjectID),
		zap.String("region", programCfg.Region),
		zap.Int("clusters", len(result.Outputs)),
		zap.Float64("cost_hint_usd_per_hour", result.CostHintUSDPerHour),
	)
	return result, nil
}

// Destroy tears down the stack for the specified ProjectInfra.
func (r *Runner) Destroy(ctx context.Context, infra *infraapi.ProjectInfra, spec *infraapi.AWSInfraSpec) error {
	if infra == nil || spec == nil {
		return nil
	}
	projectID := strings.TrimSpace(infra.Spec.ProjectID)
	region := strings.TrimSpace(infra.Spec.Region)
	infraName := strings.TrimSpace(infra.Name)
	if projectID == "" || region == "" {
		return nil
	}

	// Use infra-specific stack name to only destroy this ProjectInfra's resources
	stackName := r.stackNameForInfra(infraName, projectID, region)
	projectName := r.projectName(projectID)
	program := r.buildPulumiProgram(&programInput{
		ProjectID: projectID,
		Region:    region,
		SkipHelm:  true,
	})

	// Try to select the infra-specific stack first
	stack, err := auto.SelectStackInlineSource(ctx, stackName, projectName, program, r.buildWorkspaceOptions()...)
	if auto.IsSelectStack404Error(err) {
		// Fall back to legacy shared stack for backward compatibility
		legacyStackName := r.stackName(projectID, region)
		r.log.Info("infra-specific stack not found, trying legacy shared stack",
			zap.String("infra_stack", stackName),
			zap.String("legacy_stack", legacyStackName),
		)
		stack, err = auto.SelectStackInlineSource(ctx, legacyStackName, projectName, program, r.buildWorkspaceOptions()...)
		if auto.IsSelectStack404Error(err) {
			r.log.Info("no pulumi stack found; skipping destroy", zap.String("stack", stackName))
			return nil
		}
		stackName = legacyStackName // Use legacy name for logging
	}
	if err != nil {
		return fmt.Errorf("select pulumi stack: %w", err)
	}
	if err := r.ensurePlugins(ctx, stack); err != nil {
		return err
	}

	if err := stack.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: region}); err != nil {
		return fmt.Errorf("set stack config: %w", err)
	}

	jobID, clusterID := jobMetadata(infra)
	progressWriter := &logWriter{
		log:       r.log,
		sink:      r.sink,
		jobID:     jobID,
		projectID: strings.TrimSpace(infra.Spec.ProjectID),
		clusterID: clusterID,
		phase:     "Destroy",
	}
	_ = stack.Cancel(ctx) // best-effort unlock before destroy
	destroyRes, err := stack.Destroy(ctx, optdestroy.ProgressStreams(progressWriter))
	if err != nil && isLockError(err) && r.retryAfterCancel(ctx, stack, func() error {
		var destroyErr error
		destroyRes, destroyErr = stack.Destroy(ctx, optdestroy.ProgressStreams(progressWriter))
		return destroyErr
	}) {
		err = nil
	}
	if err != nil {
		return fmt.Errorf("pulumi destroy: %w", err)
	}
	r.log.Info("pulumi destroy completed",
		zap.String("stack", stackName),
		zap.String("result", destroyRes.Summary.Result),
	)

	if err := stack.Workspace().RemoveStack(ctx, stackName); err != nil {
		return fmt.Errorf("remove stack: %w", err)
	}
	return nil
}

// NewStack creates or selects the Pulumi stack for the supplied infrastructure and returns the stack along
// with the computed Pulumi program input.
func (r *Runner) NewStack(ctx context.Context, infra *infraapi.ProjectInfra, spec *infraapi.AWSInfraSpec) (auto.Stack, *programInput, error) {
	if infra == nil {
		return auto.Stack{}, nil, errors.New("project infra required")
	}
	if spec == nil {
		return auto.Stack{}, nil, errors.New("aws spec required")
	}

	projectID := strings.TrimSpace(infra.Spec.ProjectID)
	region := strings.TrimSpace(infra.Spec.Region)
	if projectID == "" {
		return auto.Stack{}, nil, errors.New("spec.projectId is required")
	}
	if region == "" {
		return auto.Stack{}, nil, errors.New("spec.region is required")
	}

	// Pass infra name to make cluster IDs unique per ProjectInfra
	// Also pass existing outputs for backward compatibility with clusters created before the hash suffix was added
	infraName := strings.TrimSpace(infra.Name)
	clusterDefs, err := r.buildClusterDefinitions(projectID, region, infraName, spec, infra.Status.Outputs)
	if err != nil {
		return auto.Stack{}, nil, err
	}
	if len(clusterDefs) == 0 {
		return auto.Stack{}, nil, errors.New("no cluster definitions derived from spec")
	}

	programCfg := &programInput{
		ProjectID:           projectID,
		Region:              region,
		VpcID:               strings.TrimSpace(spec.VpcID),
		SubnetIDs:           sanitizeStrings(spec.SubnetIDs),
		RoleARN:             strings.TrimSpace(spec.RoleARN),
		ExternalID:          strings.TrimSpace(spec.ExternalID),
		Clusters:            clusterDefs,
		Platform:            r.resolvePlatformConfig(),
		SpokeHelm:           r.resolveHelmConfig(),
		Observability:       observability.ResolveFromEnv(r.findRepoRoot()),
		EnableObservability: addonEnabled(infra.Spec.Addons, "observability", true),
		CertManager:         certmanager.ResolveFromEnv(),
		EnableCertManager:   addonEnabled(infra.Spec.Addons, "certmanager", false),
		SkipHelm:            strings.EqualFold(os.Getenv("AEGIS_SKIP_SPOKE_HELM"), "true"),
		EnableCostEstimates: true,
	}

	projectName := r.projectName(projectID)
	program := r.buildPulumiProgram(programCfg)

	// For backward compatibility, check if a legacy shared stack exists with resources.
	// If it does, continue using it instead of creating a new per-infra stack.
	// This prevents creating duplicate resources when upgrading existing deployments.
	legacyStackName := r.stackName(projectID, region)
	infraStackName := r.stackNameForInfra(infraName, projectID, region)
	stackName := infraStackName // default to per-infra stack

	legacyStack, legacyErr := auto.SelectStackInlineSource(ctx, legacyStackName, projectName, program, r.buildWorkspaceOptions()...)
	if legacyErr == nil {
		// Legacy stack exists - check if it has resources by looking at its state
		// If it has resources, use it for backward compatibility
		r.log.Info("legacy shared stack found, using it for backward compatibility",
			zap.String("legacy_stack", legacyStackName),
			zap.String("infra_stack", infraStackName),
		)
		stackName = legacyStackName
	} else if !auto.IsSelectStack404Error(legacyErr) {
		// Unexpected error selecting legacy stack
		return auto.Stack{}, nil, fmt.Errorf("check legacy pulumi stack: %w", legacyErr)
	}
	// If legacy stack doesn't exist (404), we'll use the new per-infra stack name

	var stack auto.Stack
	if legacyErr == nil {
		// Use the already-selected legacy stack
		stack = legacyStack
	} else {
		// Create/select the new per-infra stack
		stack, err = auto.UpsertStackInlineSource(ctx, stackName, projectName, program, r.buildWorkspaceOptions()...)
		if err != nil {
			return auto.Stack{}, nil, fmt.Errorf("upsert pulumi stack: %w", err)
		}
	}
	if err := r.ensurePlugins(ctx, stack); err != nil {
		return auto.Stack{}, nil, err
	}
	if err := stack.SetConfig(ctx, "aws:region", auto.ConfigValue{Value: region}); err != nil {
		return auto.Stack{}, nil, fmt.Errorf("set stack config: %w", err)
	}

	return stack, programCfg, nil
}

// ----------------------------------------------------------------------------- //
// Helper types

type programInput struct {
	ProjectID           string
	Region              string
	VpcID               string
	SubnetIDs           []string
	RoleARN             string
	ExternalID          string
	Clusters            []clusterDefinition
	Platform            platformConfig
	SpokeHelm           helmConfig
	Observability       observability.Config
	EnableObservability bool
	CertManager         certmanager.Config
	EnableCertManager   bool
	SkipHelm            bool
	EnableCostEstimates bool
}

type clusterDefinition struct {
	Name         string
	ClusterID    string
	Version      string
	NodePools    []infraapi.NodePool
	Labels       map[string]string
	Annotations  map[string]string
	NodeSelector map[string]string
}

type platformConfig struct {
	Endpoint       string
	CABundleBase64 string
	ImageRepo      string
	ImageTag       string
	InsecureGRPC   bool
}

type helmConfig struct {
	ChartPath        string
	Repository       string
	Version          string
	ValuesFile       string
	Namespace        string
	ReleaseName      string
	Timeout          time.Duration
	EnableDependency bool
}

// ----------------------------------------------------------------------------- //
// Workspace helpers

func (r *Runner) buildWorkspaceOptions() []auto.LocalWorkspaceOption {
	opts := []auto.LocalWorkspaceOption{}
	if workdir := strings.TrimSpace(os.Getenv(envAegisPulumiWorkdir)); workdir != "" {
		opts = append(opts, auto.WorkDir(workdir))
	}
	env := map[string]string{}
	if backend := strings.TrimSpace(os.Getenv(envPulumiBackendURL)); backend != "" {
		env["PULUMI_BACKEND_URL"] = backend
	}
	if len(env) > 0 {
		opts = append(opts, auto.EnvVars(env))
	}
	return opts
}

func (r *Runner) ensurePlugins(ctx context.Context, stack auto.Stack) error {
	ws := stack.Workspace()
	type plugin struct {
		Name    string
		Version string
	}
	required := []plugin{
		{Name: "aws", Version: "5.43.0"},
		{Name: "kubernetes", Version: "4.23.0"},
	}
	for _, p := range required {
		if err := ws.InstallPlugin(ctx, p.Name, p.Version); err != nil {
			return fmt.Errorf("install plugin %s@%s: %w", p.Name, p.Version, err)
		}
	}
	return nil
}

// ----------------------------------------------------------------------------- //
// Program construction

func (r *Runner) buildPulumiProgram(input *programInput) pulumi.RunFunc {
	if input == nil {
		input = &programInput{}
	}
	return func(ctx *pulumi.Context) error {
		installer := observability.NewInstaller()
		if input.Region == "" {
			return errors.New("region is required")
		}

		providerArgs := &aws.ProviderArgs{
			Region: pulumi.StringPtr(input.Region),
		}
		if input.RoleARN != "" {
			assume := aws.ProviderAssumeRoleArgs{
				RoleArn: pulumi.StringPtr(input.RoleARN),
			}
			if input.ExternalID != "" {
				assume.ExternalId = pulumi.StringPtr(input.ExternalID)
			}
			providerArgs.AssumeRole = assume.ToProviderAssumeRolePtrOutput()
		}

		awsProvider, err := aws.NewProvider(ctx, "aws", providerArgs)
		if err != nil {
			return fmt.Errorf("aws provider: %w", err)
		}
		providerOpt := pulumi.Provider(awsProvider)

		resolvedVpcID := strings.TrimSpace(input.VpcID)
		resolvedSubnets := sanitizeStrings(input.SubnetIDs)
		if resolvedVpcID == "" || len(resolvedSubnets) == 0 {
			vpcID, subnets, err := r.resolveDefaultSubnets(ctx, awsProvider, input.Region)
			if err != nil {
				return err
			}
			if resolvedVpcID == "" {
				resolvedVpcID = vpcID
			}
			if len(resolvedSubnets) == 0 {
				resolvedSubnets = subnets
			}
		}
		if len(resolvedSubnets) == 0 {
			return fmt.Errorf("no subnets available for region %s; specify VPC/subnet IDs in the cluster profile", input.Region)
		}

		clusterMap := pulumi.Map{}
		kubeconfigMap := pulumi.Map{}

		for idx, clusterDef := range input.Clusters {
			if len(clusterDef.NodePools) == 0 {
				return fmt.Errorf("cluster %q has no node pools configured", clusterDef.ClusterID)
			}
			clusterName := fmt.Sprintf("%s-%d", sanitize(clusterDef.Name), idx)
			resourceName := pulumiResourceName(clusterName, 30)
			if clusterDef.ClusterID != "" {
				clusterName = sanitize(clusterDef.ClusterID)
				resourceName = pulumiResourceName(clusterName, 30)
			}
			clusterIDLabel := clusterDef.ClusterID
			if clusterIDLabel == "" {
				clusterIDLabel = clusterName
			}

			tags := pulumi.StringMap{
				"Project": pulumi.String(input.ProjectID),
				"Cluster": pulumi.String(clusterIDLabel),
			}

			clusterRole, err := r.createClusterRole(ctx, fmt.Sprintf("%s-cluster-role", resourceName), tags, providerOpt)
			if err != nil {
				return err
			}
			nodeRole, err := r.createNodeRole(ctx, fmt.Sprintf("%s-node-role", resourceName), tags, providerOpt)
			if err != nil {
				return err
			}
			clusterSG, nodeSG, err := r.createSecurityGroups(ctx, resourceName, resolvedVpcID, tags, providerOpt)
			if err != nil {
				return err
			}

			lt, ltVersion, err := r.buildNodeLaunchTemplate(ctx, resourceName, clusterSG, nodeSG, tags, providerOpt)
			if err != nil {
				return err
			}

			clusterArgs := &awseks.ClusterArgs{
				Name:    pulumi.StringPtr(clusterName),
				RoleArn: clusterRole.Arn,
				Tags:    tags,
				VpcConfig: &awseks.ClusterVpcConfigArgs{
					SecurityGroupIds: pulumi.StringArray{
						clusterSG.ID().ToStringOutput(),
					},
					SubnetIds: pulumi.ToStringArray(resolvedSubnets),
				},
			}
			if version := strings.TrimSpace(clusterDef.Version); version != "" {
				clusterArgs.Version = pulumi.StringPtr(version)
			}

			cluster, err := awseks.NewCluster(ctx, resourceName, clusterArgs, providerOpt, pulumi.DependsOn([]pulumi.Resource{clusterRole, clusterSG}))
			if err != nil {
				return fmt.Errorf("create eks cluster %q: %w", clusterDef.ClusterID, err)
			}

			kubeconfig := r.buildKubeconfigWithExternalID(cluster, input.Region, input.RoleARN, input.ExternalID)

			kubeProvider, err := kubernetes.NewProvider(ctx, fmt.Sprintf("%s-k8s", sanitize(clusterName)), &kubernetes.ProviderArgs{
				Kubeconfig: kubeconfig,
			}, pulumi.DependsOn([]pulumi.Resource{cluster}))
			if err != nil {
				return fmt.Errorf("create kubernetes provider: %w", err)
			}

			awsAuth, err := r.applyAWSAuthConfig(ctx, kubeProvider, nodeRole.Arn, input.RoleARN, clusterName, []pulumi.Resource{cluster})
			if err != nil {
				return err
			}

			nodeGroupDeps := []pulumi.Resource{cluster, awsAuth, lt}
			nodeGroups, err := r.configureManagedNodeGroups(ctx, clusterDef, cluster, nodeRole, lt, ltVersion, resolvedSubnets, awsProvider, nodeGroupDeps, tags)
			if err != nil {
				return err
			}
			if hasGpuNodePool(clusterDef.NodePools) {
				deps := append([]pulumi.Resource{}, nodeGroups...)
				deps = append(deps, cluster)
				if err := r.installNvidiaDevicePlugin(ctx, clusterDef.ClusterID, clusterDef.NodePools, kubeProvider, deps); err != nil {
					return err
				}
			}

			// Extract AWS account ID from the RoleARN (format: arn:aws:iam::ACCOUNT_ID:role/...)
			accountID := extractAccountIDFromARN(input.RoleARN)
			if accountID == "" {
				if ident, err := aws.GetCallerIdentity(ctx, nil, providerOpt); err == nil && ident != nil {
					accountID = strings.TrimSpace(ident.AccountId)
				} else {
					return fmt.Errorf("resolve AWS account id for IRSA: %w", err)
				}
			}

			// Create OIDC provider for IRSA (IAM Roles for Service Accounts).
			// This allows Kubernetes service accounts to assume IAM roles.
			oidcProvider, err := r.createOIDCProvider(ctx, clusterDef.ClusterID, cluster, accountID, providerOpt)
			if err != nil {
				return err
			}

			// Create IAM role for cluster autoscaler with IRSA trust policy.
			autoscalerRoleArn, err := r.createClusterAutoscalerRole(ctx, clusterDef.ClusterID, oidcProvider, accountID, providerOpt)
			if err != nil {
				return err
			}

			// Install Cluster Autoscaler to enable automatic node scaling.
			// This is essential for on-demand GPU provisioning - when a workspace
			// requests GPU resources, the autoscaler will scale up the GPU node group.
			if err := r.installClusterAutoscaler(ctx, clusterDef.ClusterID, cluster.Name, input.Region, autoscalerRoleArn, kubeProvider, append(nodeGroups, cluster, oidcProvider)); err != nil {
				return err
			}

			var observabilityOutputs pulumi.Map
			if input.EnableObservability && input.Observability.Enable {
				obs, err := installer.Install(ctx, clusterDef.ClusterID, kubeProvider, input.Observability, append(nodeGroups, cluster))
				if err != nil {
					return err
				}
				observabilityOutputs = obs
			}

			// Install cert-manager and step-issuer for automated TLS certificate management.
			// This enables spoke-proxy to get CA-signed certificates from the hub's step-ca.
			r.log.Info("cert-manager installation check",
				zap.Bool("EnableCertManager", input.EnableCertManager),
				zap.Bool("CertManager.Enable", input.CertManager.Enable),
				zap.String("cluster_id", clusterDef.ClusterID))

			// Track cert-manager resources for use as dependencies by the spoke helm chart
			var certMgrResources []pulumi.Resource
			if input.EnableCertManager && input.CertManager.Enable {
				r.log.Info("installing cert-manager on spoke cluster", zap.String("cluster_id", clusterDef.ClusterID))
				certMgrInstaller := certmanager.NewInstaller()
				var err error
				certMgrResources, err = certMgrInstaller.Install(ctx, clusterDef.ClusterID, kubeProvider, input.CertManager, append(nodeGroups, cluster))
				if err != nil {
					return fmt.Errorf("install cert-manager: %w", err)
				}
				r.log.Info("cert-manager installed successfully", zap.String("cluster_id", clusterDef.ClusterID), zap.Int("num_resources", len(certMgrResources)))
			} else {
				r.log.Info("skipping cert-manager installation", zap.Bool("EnableCertManager", input.EnableCertManager), zap.Bool("CertManager.Enable", input.CertManager.Enable))
			}

			if !input.SkipHelm {
				// Include cert-manager resources as dependencies so the spoke helm chart
				// waits for the webhook to be ready before creating Certificate resources
				spokeDeps := append(nodeGroups, cluster)
				spokeDeps = append(spokeDeps, certMgrResources...)
				if err := r.installSpokeHelmChart(ctx, clusterDef.ClusterID, kubeProvider, kubeconfig, input, spokeDeps); err != nil {
					return err
				}
			}

			// Create the workloads namespace where workspace pods will be scheduled.
			// This must exist before the hub can create Workspace CRs in this cluster.
			// Namespace is project-scoped for multi-tenant isolation.
			if _, err := r.createWorkloadsNamespace(ctx, input.ProjectID, clusterDef.ClusterID, kubeProvider, append(nodeGroups, cluster)); err != nil {
				return err
			}

			key := clusterDef.ClusterID
			if key == "" {
				key = clusterName
			}
			kubeconfigKey := fmt.Sprintf("%s.kubeconfig", key)
			kubeconfigMap[key] = kubeconfig
			entry := pulumi.Map{
				"clusterId":           pulumi.String(key),
				"name":                pulumi.String(clusterDef.Name),
				"region":              pulumi.String(input.Region),
				"kubeconfigSecretKey": pulumi.String(kubeconfigKey),
			}
			if observabilityOutputs != nil {
				entry["observability"] = observabilityOutputs
			}
			clusterMap[key] = entry
		}

		ctx.Export("clusters", clusterMap)
		ctx.Export("kubeconfigs", kubeconfigMap)
		return nil
	}
}

// ----------------------------------------------------------------------------- //
// AWS resource helpers

func (r *Runner) createClusterRole(ctx *pulumi.Context, name string, tags pulumi.StringMap, opts pulumi.ResourceOption) (*awsiam.Role, error) {
	assume := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":["eks.amazonaws.com"]},"Action":["sts:AssumeRole"]}]}`
	role, err := awsiam.NewRole(ctx, name, &awsiam.RoleArgs{
		AssumeRolePolicy: pulumi.String(assume),
		Tags:             tags,
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("create eks service role: %w", err)
	}

	policies := []string{
		"arn:aws:iam::aws:policy/AmazonEKSClusterPolicy",
		"arn:aws:iam::aws:policy/AmazonEKSServicePolicy",
		"arn:aws:iam::aws:policy/AmazonEKSVPCResourceController",
	}
	for i, policy := range policies {
		if _, err := awsiam.NewRolePolicyAttachment(ctx, fmt.Sprintf("%s-attach-%d", name, i), &awsiam.RolePolicyAttachmentArgs{
			PolicyArn: pulumi.String(policy),
			Role:      role.Name,
		}, opts); err != nil {
			return nil, fmt.Errorf("attach policy %s: %w", policy, err)
		}
	}
	return role, nil
}

func (r *Runner) createNodeRole(ctx *pulumi.Context, name string, tags pulumi.StringMap, opts pulumi.ResourceOption) (*awsiam.Role, error) {
	assume := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":["ec2.amazonaws.com"]},"Action":["sts:AssumeRole"]}]}`
	role, err := awsiam.NewRole(ctx, name, &awsiam.RoleArgs{
		AssumeRolePolicy: pulumi.String(assume),
		Tags:             tags,
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("create eks node role: %w", err)
	}

	policies := []string{
		"arn:aws:iam::aws:policy/AmazonEKSWorkerNodePolicy",
		"arn:aws:iam::aws:policy/AmazonEKS_CNI_Policy",
		"arn:aws:iam::aws:policy/AmazonEC2ContainerRegistryReadOnly",
		"arn:aws:iam::aws:policy/AmazonSSMManagedInstanceCore",
	}
	for i, policy := range policies {
		if _, err := awsiam.NewRolePolicyAttachment(ctx, fmt.Sprintf("%s-attach-%d", name, i), &awsiam.RolePolicyAttachmentArgs{
			PolicyArn: pulumi.String(policy),
			Role:      role.Name,
		}, opts); err != nil {
			return nil, fmt.Errorf("attach policy %s: %w", policy, err)
		}
	}

	// Cluster Autoscaler IAM policy - allows autoscaler to modify ASGs
	// This is required for on-demand GPU node provisioning (Kubeflow pattern).
	// TODO: For FedRAMP, consider migrating to IRSA with a dedicated service account.
	autoscalerPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"autoscaling:DescribeAutoScalingGroups",
					"autoscaling:DescribeAutoScalingInstances",
					"autoscaling:DescribeLaunchConfigurations",
					"autoscaling:DescribeScalingActivities",
					"autoscaling:DescribeTags",
					"ec2:DescribeImages",
					"ec2:DescribeInstanceTypes",
					"ec2:DescribeLaunchTemplateVersions",
					"ec2:GetInstanceTypesFromInstanceRequirements",
					"eks:DescribeNodegroup"
				],
				"Resource": ["*"]
			},
			{
				"Effect": "Allow",
				"Action": [
					"autoscaling:SetDesiredCapacity",
					"autoscaling:TerminateInstanceInAutoScalingGroup"
				],
				"Resource": ["*"],
				"Condition": {
					"StringEquals": {
						"autoscaling:ResourceTag/k8s.io/cluster-autoscaler/enabled": "true"
					}
				}
			}
		]
	}`
	if _, err := awsiam.NewRolePolicy(ctx, fmt.Sprintf("%s-autoscaler", name), &awsiam.RolePolicyArgs{
		Role:   role.Name,
		Policy: pulumi.String(autoscalerPolicy),
	}, opts); err != nil {
		return nil, fmt.Errorf("attach cluster autoscaler policy: %w", err)
	}

	return role, nil
}

// createOIDCProvider creates an IAM OIDC identity provider for EKS IRSA.
// This allows Kubernetes service accounts to assume IAM roles via web identity federation.
func (r *Runner) createOIDCProvider(ctx *pulumi.Context, clusterID string, cluster *awseks.Cluster, accountID string, opts pulumi.ResourceOption) (*awsiam.OpenIdConnectProvider, error) {
	key := sanitize(clusterID)
	if key == "" {
		key = "aegis"
	}
	name := pulumiResourceName(fmt.Sprintf("%s-oidc", key), 53)

	// Extract the OIDC issuer URL from the cluster and compute its thumbprint.
	// The thumbprint is required by AWS IAM to verify the OIDC provider's certificate.
	oidcURL := cluster.Identities.Index(pulumi.Int(0)).Oidcs().Index(pulumi.Int(0)).Issuer().Elem()

	// AWS EKS OIDC issuers use Amazon's root CA, which has a well-known thumbprint.
	// This is stable across all EKS clusters in a region.
	// See: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_create_oidc_verify-thumbprint.html
	eksOIDCThumbprint := "9e99a48a9960b14926bb7f3b02e22da2b0ab7280"

	provider, err := awsiam.NewOpenIdConnectProvider(ctx, name, &awsiam.OpenIdConnectProviderArgs{
		Url: oidcURL,
		ClientIdLists: pulumi.StringArray{
			pulumi.String("sts.amazonaws.com"),
		},
		ThumbprintLists: pulumi.StringArray{
			pulumi.String(eksOIDCThumbprint),
		},
	}, opts)
	if err != nil {
		return nil, fmt.Errorf("create OIDC provider: %w", err)
	}

	r.log.Info("OIDC provider created for IRSA",
		zap.String("cluster_id", clusterID))
	return provider, nil
}

// createClusterAutoscalerRole creates an IAM role for the cluster autoscaler with IRSA trust policy.
// The role allows the cluster-autoscaler service account to assume it via web identity federation.
func (r *Runner) createClusterAutoscalerRole(ctx *pulumi.Context, clusterID string, oidcProvider *awsiam.OpenIdConnectProvider, accountID string, opts pulumi.ResourceOption) (pulumi.StringOutput, error) {
	key := sanitize(clusterID)
	if key == "" {
		key = "aegis"
	}
	name := pulumiResourceName(fmt.Sprintf("cluster-autoscaler-%s", key), 53)

	// Build the trust policy that allows the cluster-autoscaler service account
	// to assume this role via OIDC web identity federation.
	trustPolicy := oidcProvider.Url.ApplyT(func(url string) string {
		// Remove https:// prefix for the OIDC provider identifier
		oidcID := strings.TrimPrefix(url, "https://")
		return fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {
					"Federated": "arn:aws:iam::%s:oidc-provider/%s"
				},
				"Action": "sts:AssumeRoleWithWebIdentity",
				"Condition": {
					"StringEquals": {
						"%s:sub": "system:serviceaccount:kube-system:cluster-autoscaler",
						"%s:aud": "sts.amazonaws.com"
					}
				}
			}]
		}`, accountID, oidcID, oidcID, oidcID)
	}).(pulumi.StringOutput)

	role, err := awsiam.NewRole(ctx, name, &awsiam.RoleArgs{
		AssumeRolePolicy: trustPolicy,
		Description:      pulumi.String("IAM role for EKS cluster autoscaler (IRSA)"),
	}, opts)
	if err != nil {
		return pulumi.StringOutput{}, fmt.Errorf("create cluster autoscaler role: %w", err)
	}

	// Attach the cluster autoscaler policy
	autoscalerPolicy := `{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Action": [
					"autoscaling:DescribeAutoScalingGroups",
					"autoscaling:DescribeAutoScalingInstances",
					"autoscaling:DescribeLaunchConfigurations",
					"autoscaling:DescribeScalingActivities",
					"autoscaling:DescribeTags",
					"ec2:DescribeImages",
					"ec2:DescribeInstanceTypes",
					"ec2:DescribeLaunchTemplateVersions",
					"ec2:GetInstanceTypesFromInstanceRequirements",
					"eks:DescribeNodegroup"
				],
				"Resource": ["*"]
			},
			{
				"Effect": "Allow",
				"Action": [
					"autoscaling:SetDesiredCapacity",
					"autoscaling:TerminateInstanceInAutoScalingGroup"
				],
				"Resource": ["*"],
				"Condition": {
					"StringEquals": {
						"autoscaling:ResourceTag/k8s.io/cluster-autoscaler/enabled": "true"
					}
				}
			}
		]
	}`
	if _, err := awsiam.NewRolePolicy(ctx, fmt.Sprintf("%s-policy", name), &awsiam.RolePolicyArgs{
		Role:   role.Name,
		Policy: pulumi.String(autoscalerPolicy),
	}, opts); err != nil {
		return pulumi.StringOutput{}, fmt.Errorf("attach cluster autoscaler policy: %w", err)
	}

	r.log.Info("cluster autoscaler IRSA role created",
		zap.String("cluster_id", clusterID))
	return role.Arn, nil
}

func (r *Runner) createSecurityGroups(ctx *pulumi.Context, baseName, vpcID string, tags pulumi.StringMap, opts pulumi.ResourceOption) (*awsec2.SecurityGroup, *awsec2.SecurityGroup, error) {
	clusterSG, err := awsec2.NewSecurityGroup(ctx, fmt.Sprintf("%s-cluster-sg", baseName), &awsec2.SecurityGroupArgs{
		VpcId:       pulumi.String(vpcID),
		Description: pulumi.String("EKS control plane security group"),
		Tags:        tags,
		Egress: awsec2.SecurityGroupEgressArray{
			&awsec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
	}, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("create cluster security group: %w", err)
	}

	nodeSG, err := awsec2.NewSecurityGroup(ctx, fmt.Sprintf("%s-node-sg", baseName), &awsec2.SecurityGroupArgs{
		VpcId:       pulumi.String(vpcID),
		Description: pulumi.String("EKS managed node group security group"),
		Tags:        tags,
		Egress: awsec2.SecurityGroupEgressArray{
			&awsec2.SecurityGroupEgressArgs{
				Protocol:   pulumi.String("-1"),
				FromPort:   pulumi.Int(0),
				ToPort:     pulumi.Int(0),
				CidrBlocks: pulumi.StringArray{pulumi.String("0.0.0.0/0")},
			},
		},
	}, opts, pulumi.DependsOn([]pulumi.Resource{clusterSG}))
	if err != nil {
		return nil, nil, fmt.Errorf("create node security group: %w", err)
	}

	// Allow nodes to reach the control plane and vice versa; allow node-to-node traffic.
	if _, err := awsec2.NewSecurityGroupRule(ctx, fmt.Sprintf("%s-cluster-from-nodes", baseName), &awsec2.SecurityGroupRuleArgs{
		Type:                  pulumi.String("ingress"),
		Protocol:              pulumi.String("-1"),
		FromPort:              pulumi.Int(0),
		ToPort:                pulumi.Int(0),
		SecurityGroupId:       clusterSG.ID(),
		SourceSecurityGroupId: nodeSG.ID(),
		Description:           pulumi.String("allow node groups to reach control plane"),
	}, opts); err != nil {
		return nil, nil, fmt.Errorf("authorize nodes to control plane: %w", err)
	}

	if _, err := awsec2.NewSecurityGroupRule(ctx, fmt.Sprintf("%s-nodes-from-cluster", baseName), &awsec2.SecurityGroupRuleArgs{
		Type:                  pulumi.String("ingress"),
		Protocol:              pulumi.String("-1"),
		FromPort:              pulumi.Int(0),
		ToPort:                pulumi.Int(0),
		SecurityGroupId:       nodeSG.ID(),
		SourceSecurityGroupId: clusterSG.ID(),
		Description:           pulumi.String("allow control plane to reach nodes"),
	}, opts); err != nil {
		return nil, nil, fmt.Errorf("authorize control plane to nodes: %w", err)
	}

	if _, err := awsec2.NewSecurityGroupRule(ctx, fmt.Sprintf("%s-nodes-self", baseName), &awsec2.SecurityGroupRuleArgs{
		Type:                  pulumi.String("ingress"),
		Protocol:              pulumi.String("-1"),
		FromPort:              pulumi.Int(0),
		ToPort:                pulumi.Int(0),
		SecurityGroupId:       nodeSG.ID(),
		SourceSecurityGroupId: nodeSG.ID(),
		Description:           pulumi.String("allow node-to-node communication"),
	}, opts); err != nil {
		return nil, nil, fmt.Errorf("authorize node self traffic: %w", err)
	}

	// Allow inbound traffic to spoke-proxy NodePort from anywhere (for workspace connections)
	nodePort := getSpokeProxyNodePort()
	if _, err := awsec2.NewSecurityGroupRule(ctx, fmt.Sprintf("%s-spoke-proxy-nodeport", baseName), &awsec2.SecurityGroupRuleArgs{
		Type:            pulumi.String("ingress"),
		Protocol:        pulumi.String("tcp"),
		FromPort:        pulumi.Int(nodePort),
		ToPort:          pulumi.Int(nodePort),
		SecurityGroupId: nodeSG.ID(),
		CidrBlocks:      pulumi.StringArray{pulumi.String("0.0.0.0/0")},
		Description:     pulumi.String(fmt.Sprintf("allow spoke-proxy NodePort %d for workspace connections", nodePort)),
	}, opts); err != nil {
		return nil, nil, fmt.Errorf("authorize spoke-proxy nodeport: %w", err)
	}

	return clusterSG, nodeSG, nil
}

func (r *Runner) buildNodeLaunchTemplate(ctx *pulumi.Context, baseName string, clusterSG, nodeSG *awsec2.SecurityGroup, tags pulumi.StringMap, opts pulumi.ResourceOption) (*awsec2.LaunchTemplate, pulumi.StringOutput, error) {
	lt, err := awsec2.NewLaunchTemplate(ctx, fmt.Sprintf("%s-lt", baseName), &awsec2.LaunchTemplateArgs{
		NamePrefix: pulumi.StringPtr(fmt.Sprintf("%s-", baseName)),
		Tags:       tags,
		VpcSecurityGroupIds: pulumi.StringArray{
			clusterSG.ID().ToStringOutput(),
			nodeSG.ID().ToStringOutput(),
		},
		TagSpecifications: awsec2.LaunchTemplateTagSpecificationArray{
			&awsec2.LaunchTemplateTagSpecificationArgs{
				ResourceType: pulumi.String("instance"),
				Tags:         tags,
			},
		},
	}, opts, pulumi.DependsOn([]pulumi.Resource{clusterSG, nodeSG}))
	if err != nil {
		return nil, pulumi.StringOutput{}, fmt.Errorf("create launch template: %w", err)
	}
	version := lt.LatestVersion.ApplyT(func(v int) string {
		return fmt.Sprintf("%d", v)
	}).(pulumi.StringOutput)
	return lt, version, nil
}

func (r *Runner) configureManagedNodeGroups(ctx *pulumi.Context, clusterDef clusterDefinition, cluster *awseks.Cluster, nodeRole *awsiam.Role, lt *awsec2.LaunchTemplate, ltVersion pulumi.StringOutput, subnets []string, provider *aws.Provider, depends []pulumi.Resource, tags pulumi.StringMap) ([]pulumi.Resource, error) {
	if len(clusterDef.NodePools) == 0 {
		return nil, nil
	}

	nodeGroups := []pulumi.Resource{}
	for _, pool := range clusterDef.NodePools {
		name := strings.TrimSpace(pool.Name)
		if name == "" {
			name = fmt.Sprintf("%s-nodepool", sanitize(clusterDef.ClusterID))
		}
		name = pulumiResourceName(sanitize(name), 30)
		instanceType := normalizeInstanceType(pool.InstanceType)
		if instanceType == "" {
			instanceType = "t3.small" // Default: 2 vCPU, 2GB - sufficient for most system workloads
		}
		gpuPool := isGpuNodePool(pool)
		// AWS EKS requires MaxSize >= 1 for node groups.
		// Ensure MaxSize is at least 1 even if profile has 0.
		maxSize := int(pool.MaxSize)
		if maxSize < 1 {
			maxSize = 5 // Default to 5 for autoscaling headroom
		}
		scaling := &awseks.NodeGroupScalingConfigArgs{
			DesiredSize: pulumi.Int(int(pool.MinSize)),
			MinSize:     pulumi.Int(int(pool.MinSize)),
			MaxSize:     pulumi.Int(maxSize),
		}

		// Add Cluster Autoscaler discovery tags to enable automatic scaling.
		// These tags allow the autoscaler to identify and manage the ASGs.
		ngTags := pulumi.StringMap{}
		for k, v := range tags {
			ngTags[k] = v
		}
		// Required tags for Cluster Autoscaler autodiscovery
		ngTags["k8s.io/cluster-autoscaler/enabled"] = pulumi.String("true")
		ngTags["k8s.io/cluster-autoscaler/"+clusterDef.ClusterID] = pulumi.String("owned")

		ngArgs := &awseks.NodeGroupArgs{
			ClusterName:   pulumi.StringInput(cluster.Name),
			NodeRoleArn:   nodeRole.Arn,
			NodeGroupName: pulumi.StringPtr(name),
			SubnetIds:     pulumi.ToStringArray(subnets),
			InstanceTypes: pulumi.StringArray{
				pulumi.String(instanceType),
			},
			ScalingConfig: scaling,
			Taints:        convertTaints(pool.Taints),
			Tags:          ngTags,
			LaunchTemplate: &awseks.NodeGroupLaunchTemplateArgs{
				Id:      lt.ID().ToStringPtrOutput(),
				Version: ltVersion,
			},
		}

		if gpuPool {
			ngArgs.AmiType = pulumi.StringPtr(gpuAmiType(clusterDef.Version))
		}

		if len(pool.Labels) > 0 {
			labels := pulumi.StringMap{}
			for k, v := range pool.Labels {
				labels[k] = pulumi.String(v)
			}
			ngArgs.Labels = labels
		}

		opts := []pulumi.ResourceOption{pulumi.Provider(provider), pulumi.DependsOn(depends)}
		ng, err := awseks.NewNodeGroup(ctx, fmt.Sprintf("%s-nodegroup", name), ngArgs, opts...)
		if err != nil {
			return nil, fmt.Errorf("create managed node group %q: %w", name, err)
		}
		nodeGroups = append(nodeGroups, ng)
	}
	return nodeGroups, nil
}

func (r *Runner) applyAWSAuthConfig(ctx *pulumi.Context, kubeProvider *kubernetes.Provider, nodeRoleArn pulumi.StringInput, adminRoleArn string, clusterName string, depends []pulumi.Resource) (pulumi.Resource, error) {
	mapRoles := pulumi.Sprintf(`- rolearn: %s
  username: system:node:{{EC2PrivateDNSName}}
  groups:
    - system:bootstrappers
    - system:nodes
`, nodeRoleArn)

	adminRole := strings.TrimSpace(adminRoleArn)
	if adminRole != "" {
		mapRoles = pulumi.Sprintf(`%s
- rolearn: %s
  username: admin:{{SessionName}}
  groups:
    - system:masters
`, mapRoles, pulumi.String(adminRole))
	}

	cm, err := kubecorev1.NewConfigMap(ctx, fmt.Sprintf("%s-aws-auth", sanitize(clusterName)), &kubecorev1.ConfigMapArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("aws-auth"),
			Namespace: pulumi.String("kube-system"),
		},
		Data: pulumi.StringMap{
			"mapRoles": mapRoles,
		},
	}, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
	if err != nil {
		return nil, fmt.Errorf("create aws-auth configmap: %w", err)
	}
	return cm, nil
}

func (r *Runner) installSpokeHelmChart(ctx *pulumi.Context, clusterID string, kubeProvider *kubernetes.Provider, kubeconfig pulumi.StringOutput, input *programInput, depends []pulumi.Resource) error {
	envValues := pulumi.Map{
		"AEGIS_CLUSTER_ID": pulumi.String(clusterID),
		"AEGIS_REGION":     pulumi.String(input.Region),
		"AEGIS_PROVIDER":   pulumi.String("aws"),
	}
	if input.Platform.Endpoint != "" {
		envValues["AEGIS_PLATFORM_API_ENDPOINT"] = pulumi.String(input.Platform.Endpoint)
		if endpoint := grpcEndpointHostPort(input.Platform.Endpoint); endpoint != "" {
			envValues["AEGIS_CP_GRPC"] = pulumi.String(endpoint)
			if !input.Platform.InsecureGRPC {
				envValues["AEGIS_CP_GRPC_INSECURE"] = pulumi.String("false")
			}
		}
	}
	if input.Platform.CABundleBase64 != "" {
		envValues["AEGIS_PLATFORM_CA_B64"] = pulumi.String(input.Platform.CABundleBase64)
	}

	// If AEGIS_SPOKE_PROXY_HOST is set, use it as the stable proxy URL.
	// This allows the operator to configure a stable NLB/EIP-based URL.
	if proxyHost := strings.TrimSpace(os.Getenv(envAegisSpokeProxyHost)); proxyHost != "" {
		envValues["AEGIS_PROXY_INGRESS_HOST"] = pulumi.String(proxyHost)
		r.log.Info("using stable spoke proxy host from environment",
			zap.String("proxy_host", proxyHost))
	}

	k8sAgentValues := pulumi.Map{
		"env": envValues,
	}
	if input.Platform.ImageRepo != "" {
		k8sAgentValues["image"] = pulumi.Map{
			"repository": pulumi.String(input.Platform.ImageRepo),
			"tag":        pulumi.String(input.Platform.ImageTag),
		}
	}

	// Configure spoke-proxy for remote workspace connections.
	// Uses NodePort (31484) with nip.io for dynamic DNS based on node public IP.
	// The k8s-agent will discover the node public IP and register the proxy_url.
	nodePort := getSpokeProxyNodePort()
	proxyValues := pulumi.Map{
		"enabled": pulumi.Bool(true),
		"service": pulumi.Map{
			"type":     pulumi.String("NodePort"),
			"nodePort": pulumi.Int(nodePort),
			"port":     pulumi.Int(443),
		},
		"ingress": pulumi.Map{
			"enabled": pulumi.Bool(false), // Using NodePort, not ingress
		},
	}
	if input.Platform.ImageRepo != "" {
		proxyValues["image"] = pulumi.Map{
			"repository": pulumi.String(strings.Replace(input.Platform.ImageRepo, "aegis-k8s-agent", "aegis-proxy", 1)),
			"tag":        pulumi.String(input.Platform.ImageTag),
		}
	}

	// Configure TLS for spoke-proxy.
	// If cert-manager is enabled, use it for automated certificate management.
	// Otherwise, generate a self-signed certificate (or CA-signed if CA is configured).
	if input.EnableCertManager && input.CertManager.Enable {
		// Use cert-manager to get certificates from step-ca
		proxyValues["tls"] = pulumi.Map{
			"terminateAtIngress": pulumi.Bool(false),
			"certManager": pulumi.Map{
				"enabled": pulumi.Bool(true),
				"issuerRef": pulumi.Map{
					"name":  pulumi.String(input.CertManager.ClusterIssuerName),
					"kind":  pulumi.String("StepClusterIssuer"),
					"group": pulumi.String("certmanager.step.sm"),
				},
				"dnsNames": pulumi.StringArray{
					pulumi.String("*.nip.io"),
					pulumi.String(fmt.Sprintf("spoke-proxy-%s.nip.io", clusterID)),
				},
			},
		}
		r.log.Info("spoke-proxy will use cert-manager for TLS certificates",
			zap.String("cluster_id", clusterID),
			zap.String("issuer", input.CertManager.ClusterIssuerName))
	} else {
		// Fallback: Generate self-signed TLS certificate for spoke-proxy with *.nip.io wildcard
		certPEM, keyPEM, err := generateSpokeProxyCert()
		if err != nil {
			return fmt.Errorf("generate spoke-proxy TLS cert: %w", err)
		}
		proxyValues["tls"] = pulumi.Map{
			"cert": pulumi.String(certPEM),
			"key":  pulumi.String(keyPEM),
		}
	}

	values := pulumi.Map{
		"k8sAgent": k8sAgentValues,
		"proxy":    proxyValues,
	}

	helmCfg := input.SpokeHelm
	if helmCfg.Namespace == "" {
		helmCfg.Namespace = defaultHelmNamespace
	}
	if helmCfg.ReleaseName == "" {
		helmCfg.ReleaseName = pulumiResourceName(fmt.Sprintf("%s-%s", defaultSpokeReleaseName, sanitize(clusterID)), 53)
	}
	if helmCfg.Timeout == 0 {
		helmCfg.Timeout = defaultSpokeReleaseTimeout
	}

	releaseArgs := &helm.ReleaseArgs{
		Name:            pulumi.StringPtr(helmCfg.ReleaseName),
		Namespace:       pulumi.StringPtr(helmCfg.Namespace),
		Chart:           pulumi.String(helmCfg.ChartPath),
		Values:          values,
		Timeout:         pulumi.IntPtr(int(helmCfg.Timeout.Seconds())),
		CreateNamespace: pulumi.BoolPtr(true),
	}

	if helmCfg.Repository != "" {
		opts := helm.RepositoryOptsArgs{
			Repo: pulumi.StringPtr(helmCfg.Repository),
		}
		releaseArgs.RepositoryOpts = opts.ToRepositoryOptsPtrOutput()
	}
	if helmCfg.Version != "" {
		releaseArgs.Version = pulumi.StringPtr(helmCfg.Version)
	}
	if helmCfg.ValuesFile != "" {
		releaseArgs.ValueYamlFiles = pulumi.AssetOrArchiveArray{
			pulumi.NewFileAsset(helmCfg.ValuesFile),
		}
	}
	if helmCfg.EnableDependency {
		releaseArgs.DependencyUpdate = pulumi.BoolPtr(true)
	}

	if _, err := helm.NewRelease(ctx, fmt.Sprintf("%s-release", sanitize(clusterID)), releaseArgs, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends)); err != nil {
		return fmt.Errorf("install aegis-spoke for cluster %q: %w", clusterID, err)
	}
	return nil
}

// createWorkloadsNamespace creates the namespace where workspace pods will be scheduled.
// This namespace is required for the hub's platform-api to create Workspace CRs in spoke clusters.
// For multi-tenancy, namespaces are project-scoped: aegis-workloads-{projectId}
func (r *Runner) createWorkloadsNamespace(ctx *pulumi.Context, projectID, clusterID string, kubeProvider *kubernetes.Provider, depends []pulumi.Resource) (*kubecorev1.Namespace, error) {
	nsName := WorkloadsNamespaceForProject(projectID)
	resourceName := fmt.Sprintf("%s-workloads-ns", sanitize(clusterID))

	ns, err := kubecorev1.NewNamespace(ctx, resourceName, &kubecorev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String(nsName),
			Labels: pulumi.StringMap{
				"app.kubernetes.io/name":       pulumi.String("aegis-workloads"),
				"app.kubernetes.io/component":  pulumi.String("workloads"),
				"app.kubernetes.io/managed-by": pulumi.String("aegis-platform"),
				"aegis.yourorg.dev/cluster-id": pulumi.String(clusterID),
				"aegis.yourorg.dev/project-id": pulumi.String(projectID),
			},
			Annotations: pulumi.StringMap{
				"aegis.yourorg.dev/purpose": pulumi.String("Namespace for aegis workspace pods"),
			},
		},
	}, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
	if err != nil {
		return nil, fmt.Errorf("create workloads namespace %q: %w", nsName, err)
	}
	return ns, nil
}

// WorkloadsNamespaceForProject returns the namespace name for a project's workspace pods.
// Pattern: aegis-workloads-{projectId} for multi-tenant isolation.
// Can be overridden globally via AEGIS_WORKLOADS_NAMESPACE for single-tenant deployments.
// This function is exported so the platform-api server can use the same logic.
func WorkloadsNamespaceForProject(projectID string) string {
	// Allow global override for single-tenant or dev deployments
	if ns := strings.TrimSpace(os.Getenv(envAegisWorkloadsNamespace)); ns != "" {
		return ns
	}
	// Multi-tenant pattern: project-scoped namespaces
	sanitizedProject := sanitize(projectID)
	if sanitizedProject == "" {
		sanitizedProject = "default"
	}
	return fmt.Sprintf("%s-%s", defaultWorkloadsNamespace, sanitizedProject)
}

func (r *Runner) installNvidiaDevicePlugin(ctx *pulumi.Context, clusterID string, nodePools []infraapi.NodePool, kubeProvider *kubernetes.Provider, depends []pulumi.Resource) error {
	key := sanitize(clusterID)
	if key == "" {
		key = "aegis"
	}
	// Helm release names must be <= 53 chars. Build the full name first and then constrain it.
	name := pulumiResourceName(fmt.Sprintf("%s-nvidia-device-plugin", key), 53)
	values := pulumi.Map{
		"tolerations": pulumi.Array{
			pulumi.Map{
				"key":      pulumi.String("nvidia.com/gpu"),
				"operator": pulumi.String("Exists"),
				"effect":   pulumi.String("NoSchedule"),
			},
			pulumi.Map{
				"key":      pulumi.String("CriticalAddonsOnly"),
				"operator": pulumi.String("Exists"),
			},
		},
	}

	// Restrict scheduling to GPU nodegroups. If there is exactly one GPU flavor, use a nodeSelector.
	// If multiple GPU flavors exist, use affinity with an "In" matcher so the plugin lands on all GPU pools.
	if flavors := gpuFlavors(nodePools); len(flavors) == 1 {
		values["nodeSelector"] = pulumi.Map{"aegis.io/gpu-flavor": pulumi.String(flavors[0])}
	} else if len(flavors) > 1 {
		vals := pulumi.StringArray{}
		for _, f := range flavors {
			vals = append(vals, pulumi.String(f))
		}
		values["affinity"] = pulumi.Map{
			"nodeAffinity": pulumi.Map{
				"requiredDuringSchedulingIgnoredDuringExecution": pulumi.Map{
					"nodeSelectorTerms": pulumi.Array{
						pulumi.Map{
							"matchExpressions": pulumi.Array{
								pulumi.Map{
									"key":      pulumi.String("aegis.io/gpu-flavor"),
									"operator": pulumi.String("In"),
									"values":   vals,
								},
							},
						},
					},
				},
			},
		}
	} else if sel := gpuNodeSelector(nodePools); len(sel) > 0 {
		ns := pulumi.Map{}
		for k, v := range sel {
			ns[k] = pulumi.String(v)
		}
		values["nodeSelector"] = ns
	}

	if mig := gpuMigStrategy(nodePools); mig != "" {
		values["args"] = pulumi.Array{pulumi.String(fmt.Sprintf("--mig-strategy=%s", mig))}
	}

	_, err := helm.NewRelease(ctx, name, &helm.ReleaseArgs{
		Name:      pulumi.StringPtr(name),
		Namespace: pulumi.StringPtr("kube-system"),
		Chart:     pulumi.String("nvidia-device-plugin"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.StringPtr("https://nvidia.github.io/k8s-device-plugin"),
		},
		Version: pulumi.StringPtr("0.14.5"),
		Values:  values,
	}, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
	if err != nil {
		return fmt.Errorf("install nvidia device plugin: %w", err)
	}
	return nil
}

// installClusterAutoscaler deploys the Kubernetes Cluster Autoscaler to enable automatic
// node scaling based on pending pod resource requests. This is essential for on-demand
// GPU node provisioning - when a workspace requests GPU resources, the autoscaler will
// scale up the appropriate node group.
func (r *Runner) installClusterAutoscaler(ctx *pulumi.Context, clusterID string, clusterName pulumi.StringInput, region string, roleArn pulumi.StringOutput, kubeProvider *kubernetes.Provider, depends []pulumi.Resource) error {
	key := sanitize(clusterID)
	if key == "" {
		key = "aegis"
	}
	name := pulumiResourceName(fmt.Sprintf("%s-cluster-autoscaler", key), 53)

	// Cluster Autoscaler needs the cluster name to discover ASGs
	values := pulumi.Map{
		"autoDiscovery": pulumi.Map{
			"clusterName": clusterName,
		},
		"awsRegion": pulumi.String(region),
		// Scale down settings for cost optimization
		"extraArgs": pulumi.Map{
			"scale-down-enabled":               pulumi.Bool(true),
			"scale-down-delay-after-add":       pulumi.String("5m"),
			"scale-down-unneeded-time":         pulumi.String("5m"),
			"scale-down-utilization-threshold": pulumi.String("0.5"),
			"skip-nodes-with-local-storage":    pulumi.Bool(false),
			"skip-nodes-with-system-pods":      pulumi.Bool(false),
			"balance-similar-node-groups":      pulumi.Bool(true),
			"expander":                         pulumi.String("least-waste"),
		},
		// Resource requests for the autoscaler pod
		"resources": pulumi.Map{
			"requests": pulumi.Map{
				"cpu":    pulumi.String("100m"),
				"memory": pulumi.String("300Mi"),
			},
			"limits": pulumi.Map{
				"cpu":    pulumi.String("100m"),
				"memory": pulumi.String("300Mi"),
			},
		},
		// Run on system nodes, not GPU nodes
		"nodeSelector": pulumi.Map{
			"kubernetes.io/os": pulumi.String("linux"),
		},
		"tolerations": pulumi.Array{
			pulumi.Map{
				"key":      pulumi.String("CriticalAddonsOnly"),
				"operator": pulumi.String("Exists"),
			},
		},
		// RBAC is required for the autoscaler to modify ASGs
		// Use IRSA (IAM Roles for Service Accounts) for secure credential management
		"rbac": pulumi.Map{
			"create": pulumi.Bool(true),
			"serviceAccount": pulumi.Map{
				"create": pulumi.Bool(true),
				"name":   pulumi.String("cluster-autoscaler"),
				// IRSA annotation allows the service account to assume the IAM role
				"annotations": pulumi.Map{
					"eks.amazonaws.com/role-arn": roleArn,
				},
			},
		},
	}

	_, err := helm.NewRelease(ctx, name, &helm.ReleaseArgs{
		Name:      pulumi.StringPtr("cluster-autoscaler"),
		Namespace: pulumi.StringPtr("kube-system"),
		Chart:     pulumi.String("cluster-autoscaler"),
		RepositoryOpts: &helm.RepositoryOptsArgs{
			Repo: pulumi.StringPtr("https://kubernetes.github.io/autoscaler"),
		},
		Version: pulumi.StringPtr("9.37.0"),
		Values:  values,
	}, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
	if err != nil {
		return fmt.Errorf("install cluster autoscaler: %w", err)
	}

	r.log.Info("cluster autoscaler installed",
		zap.String("cluster_id", clusterID),
		zap.String("region", region))
	return nil
}

func (r *Runner) buildKubeconfigWithExternalID(cluster *awseks.Cluster, region, roleARN, externalID string) pulumi.StringOutput {
	if cluster == nil {
		return pulumi.Sprintf("")
	}

	hasExternal := strings.TrimSpace(roleARN) != "" && strings.TrimSpace(externalID) != ""
	if hasExternal {
		return pulumi.Sprintf(`apiVersion: v1
clusters:
- cluster:
    server: %s
    certificate-authority-data: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s
kind: Config
preferences: {}
users:
- name: %s
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: /usr/local/bin/aegis-eks-token
      env:
      - name: AEGIS_EKS_CLUSTER_NAME
        value: %s
      - name: AEGIS_EKS_REGION
        value: %s
      - name: AEGIS_EKS_ROLE_ARN
        value: %s
      - name: AEGIS_EKS_EXTERNAL_ID
        value: %s
      interactiveMode: IfAvailable
`, cluster.Endpoint, cluster.CertificateAuthority.Data().Elem(), cluster.Name,
			cluster.Name, cluster.Name, cluster.Name, cluster.Name, cluster.Name,
			cluster.Name, pulumi.String(region), pulumi.String(roleARN), pulumi.String(externalID))
	}

	return pulumi.Sprintf(`apiVersion: v1
clusters:
- cluster:
    server: %s
    certificate-authority-data: %s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s
kind: Config
preferences: {}
users:
- name: %s
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: aws
      args:
      - eks
      - get-token
      - --cluster-name
      - %s
      - --region
      - %s
      interactiveMode: IfAvailable
`, cluster.Endpoint, cluster.CertificateAuthority.Data().Elem(), cluster.Name,
		cluster.Name, cluster.Name, cluster.Name, cluster.Name, cluster.Name,
		cluster.Name, cluster.Name, pulumi.String(region))
}

// ----------------------------------------------------------------------------- //
// VPC helpers

func (r *Runner) resolveDefaultSubnets(ctx *pulumi.Context, provider *aws.Provider, region string) (string, []string, error) {
	azs, err := aws.GetAvailabilityZones(ctx, &aws.GetAvailabilityZonesArgs{}, pulumi.Provider(provider))
	if err != nil {
		return "", nil, fmt.Errorf("discover availability zones: %w", err)
	}
	eligibleAZs := map[string]bool{}
	for _, name := range azs.Names {
		eligibleAZs[strings.TrimSpace(name)] = true
	}
	if u := unsupportedControlPlaneAZs[strings.TrimSpace(region)]; len(u) > 0 {
		for az := range u {
			delete(eligibleAZs, az)
		}
	}

	vpcs, err := awsec2.GetVpcs(ctx, &awsec2.GetVpcsArgs{
		Filters: []awsec2.GetVpcsFilter{
			{
				Name:   "is-default",
				Values: []string{"true"},
			},
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return "", nil, fmt.Errorf("discover default vpc: %w", err)
	}
	if len(vpcs.Ids) == 0 {
		return "", nil, fmt.Errorf("no default VPC found in region %s; set VPC/subnet IDs in the profile", region)
	}
	vpcID := vpcs.Ids[0]
	subnets, err := awsec2.GetSubnets(ctx, &awsec2.GetSubnetsArgs{
		Filters: []awsec2.GetSubnetsFilter{
			{
				Name:   "vpc-id",
				Values: []string{vpcID},
			},
		},
	}, pulumi.Provider(provider))
	if err != nil {
		return "", nil, fmt.Errorf("discover subnets for vpc %s: %w", vpcID, err)
	}
	filtered := []string{}
	azSet := map[string]bool{}
	for _, subnetID := range subnets.Ids {
		subnet, err := awsec2.LookupSubnet(ctx, &awsec2.LookupSubnetArgs{Id: pulumi.StringRef(subnetID)}, pulumi.Provider(provider))
		if err != nil {
			return "", nil, fmt.Errorf("inspect subnet %s: %w", subnetID, err)
		}
		az := strings.TrimSpace(subnet.AvailabilityZone)
		if !eligibleAZs[az] {
			continue
		}
		filtered = append(filtered, subnetID)
		azSet[az] = true
	}

	if len(azSet) < 2 {
		return "", nil, fmt.Errorf("default VPC %s has %d supported AZs for region %s; specify VPC/subnet IDs in the cluster profile (e.g., subnets in %v)", vpcID, len(azSet), region, azs.Names)
	}
	if len(filtered) == 0 {
		return "", nil, fmt.Errorf("no eligible subnets discovered in default VPC %s for region %s", vpcID, region)
	}
	ctx.Log.Info(fmt.Sprintf("using filtered default VPC subnets for VPC %s (subnets=%d, azs=%d)", vpcID, len(filtered), len(azSet)), &pulumi.LogArgs{})
	return vpcID, filtered, nil
}

// ----------------------------------------------------------------------------- //
// Output translation

func (r *Runner) translateOutputs(outputs map[string]auto.OutputValue, region string) (*provisioning.ProvisionResult, error) {
	rawClusters, ok := outputs["clusters"]
	if !ok {
		return nil, errors.New("pulumi output missing 'clusters'")
	}
	rawKubeconfigs, ok := outputs["kubeconfigs"]
	if !ok {
		return nil, errors.New("pulumi output missing 'kubeconfigs'")
	}

	clusterEntries, ok := rawClusters.Value.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected pulumi output format for clusters")
	}
	kubeconfigEntries, ok := rawKubeconfigs.Value.(map[string]interface{})
	if !ok {
		return nil, errors.New("unexpected pulumi output format for kubeconfigs")
	}

	result := &provisioning.ProvisionResult{
		Outputs:     make([]infraapi.ClusterOutput, 0, len(clusterEntries)),
		Kubeconfigs: make(map[string][]byte, len(kubeconfigEntries)),
	}

	for clusterID, val := range clusterEntries {
		entry, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("pulumi cluster entry %q malformed", clusterID)
		}
		name := stringFromEntry(entry, "name", clusterID)
		secretKey := stringFromEntry(entry, "kubeconfigSecretKey", fmt.Sprintf("%s.kubeconfig", clusterID))
		regionOut := stringFromEntry(entry, "region", region)
		observability := observabilityFromEntry(entry["observability"])

		result.Outputs = append(result.Outputs, infraapi.ClusterOutput{
			ClusterID:           clusterID,
			Name:                name,
			Region:              regionOut,
			KubeconfigSecretKey: secretKey,
			Observability:       observability,
		})
	}

	for clusterID, val := range kubeconfigEntries {
		cfg, ok := val.(string)
		if !ok {
			return nil, fmt.Errorf("pulumi kubeconfig entry %q malformed", clusterID)
		}
		result.Kubeconfigs[clusterID] = []byte(cfg)
	}

	return result, nil
}

func stringFromEntry(entry map[string]interface{}, key, defaultVal string) string {
	if raw, ok := entry[key]; ok {
		if s, ok := raw.(string); ok && s != "" {
			return s
		}
	}
	return defaultVal
}

func intFromEntry(entry map[string]interface{}, key string, defaultVal int) int {
	if raw, ok := entry[key]; ok {
		switch v := raw.(type) {
		case float64:
			return int(v)
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case string:
			if parsed, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
				return parsed
			}
		}
	}
	return defaultVal
}

func observabilityFromEntry(val interface{}) infraapi.ObservabilityOutput {
	obs := infraapi.ObservabilityOutput{}
	if val == nil {
		return obs
	}
	raw, ok := val.(map[string]interface{})
	if !ok {
		return obs
	}

	obs.Namespace = stringFromEntry(raw, "namespace", "")
	obs.PrometheusService = stringFromEntry(raw, "prometheusService", "")
	obs.PrometheusPort = int32(intFromEntry(raw, "prometheusPort", 0))
	obs.AlertmanagerService = stringFromEntry(raw, "alertmanagerService", "")
	obs.AlertmanagerPort = int32(intFromEntry(raw, "alertmanagerPort", 0))
	obs.AlertmanagerConfigSecret = stringFromEntry(raw, "alertmanagerConfigSecret", "")
	obs.MetricsServerService = stringFromEntry(raw, "metricsServerService", "")
	obs.MetricsServerPort = int32(intFromEntry(raw, "metricsServerPort", 0))
	obs.LokiNamespace = stringFromEntry(raw, "lokiNamespace", "")
	obs.LokiService = stringFromEntry(raw, "lokiService", "")
	obs.LokiPort = int32(intFromEntry(raw, "lokiPort", 0))
	obs.LokiAuthSecret = stringFromEntry(raw, "lokiAuthSecret", "")
	obs.OtelEndpoint = stringFromEntry(raw, "otelEndpoint", "")
	obs.MetricsURL = stringFromEntry(raw, "metricsUrl", "")
	obs.TracingNamespace = stringFromEntry(raw, "tracingNamespace", "")
	obs.TempoService = stringFromEntry(raw, "tempoService", "")
	obs.TempoPort = int32(intFromEntry(raw, "tempoPort", 0))
	obs.OtelService = stringFromEntry(raw, "otelService", "")
	obs.OtelGrpcPort = int32(intFromEntry(raw, "otelGrpcPort", 0))
	obs.OtelHttpPort = int32(intFromEntry(raw, "otelHttpPort", 0))
	return obs
}

// ----------------------------------------------------------------------------- //
// Spec translation helpers

func (r *Runner) buildClusterDefinitions(projectID, region, infraName string, spec *infraapi.AWSInfraSpec, existingOutputs []infraapi.ClusterOutput) ([]clusterDefinition, error) {
	var defs []clusterDefinition

	// Build a map of existing cluster IDs from the status outputs for backward compatibility.
	// If a cluster was already created with a specific ID, we must continue using that ID.
	existingClusterIDs := make(map[string]string) // name -> clusterId
	for _, out := range existingOutputs {
		if out.Name != "" && out.ClusterID != "" {
			existingClusterIDs[out.Name] = out.ClusterID
		}
	}

	// Generate a short unique suffix from the infra name for NEW clusters only.
	// This prevents conflicts when multiple ProjectInfras use the same base cluster name.
	infraSuffix := shortHash(infraName)

	appendCluster := func(name, version string, pools []infraapi.NodePool) {
		if strings.TrimSpace(name) == "" {
			return
		}
		// Check if this cluster already exists with a known ID (backward compatibility)
		clusterID := existingClusterIDs[name]
		if clusterID == "" {
			// Also check legacy format without suffix (e.g., demo-1-us-east-1-demo-3)
			legacyID := buildClusterID(projectID, region, name)
			for _, out := range existingOutputs {
				if out.ClusterID == legacyID {
					clusterID = legacyID
					break
				}
			}
		}
		if clusterID == "" {
			// New cluster - use suffix for uniqueness
			clusterID = buildClusterIDWithSuffix(projectID, region, name, infraSuffix)
		}
		defs = append(defs, clusterDefinition{
			Name:      name,
			ClusterID: clusterID,
			Version:   strings.TrimSpace(version),
			NodePools: pools,
		})
	}

	if primary := strings.TrimSpace(spec.ClusterName); primary != "" {
		pools := normalizeNodePools(spec.NodePools)
		if len(pools) == 0 {
			pools = defaultNodePools()
		}
		appendCluster(primary, spec.Version, pools)
	}
	for _, additional := range spec.AdditionalClusters {
		pools := normalizeNodePools(additional.NodePools)
		if len(pools) == 0 {
			pools = defaultNodePools()
		}
		appendCluster(additional.ClusterName, additional.Version, pools)
	}

	return defs, nil
}

func defaultNodePools() []infraapi.NodePool {
	return []infraapi.NodePool{{
		Name:         "default",
		InstanceType: "t3.small", // 2 vCPU, 2GB - cost-effective default
		MinSize:      1,
		MaxSize:      3,
	}}
}

func normalizeNodePools(pools []infraapi.NodePool) []infraapi.NodePool {
	if len(pools) == 0 {
		return pools
	}

	normalized := make([]infraapi.NodePool, 0, len(pools))
	for _, pool := range pools {
		cloned := pool
		rawInstance := strings.ToLower(strings.TrimSpace(pool.InstanceType))
		cloned.InstanceType = normalizeInstanceType(pool.InstanceType)
		if isGpuNodePool(cloned) {
			if cloned.Labels == nil {
				cloned.Labels = map[string]string{}
			}
			flavorLabel := "nvidia-tesla-t4"
			if strings.Contains(rawInstance, "g5") || hasMigTaint(pool.Taints) {
				flavorLabel = "nvidia-a10g-mig"
			}
			if val := strings.TrimSpace(cloned.Labels["aegis.io/gpu-flavor"]); val == "" {
				cloned.Labels["aegis.io/gpu-flavor"] = flavorLabel
			}
		}
		normalized = append(normalized, cloned)
	}

	return normalized
}

func normalizeInstanceType(instanceType string) string {
	trimmed := strings.TrimSpace(instanceType)
	if trimmed == "" {
		return trimmed
	}
	lower := strings.ToLower(trimmed)
	if lower == "g5" {
		// If only the family is provided, default to a concrete g5 size; otherwise preserve the requested type.
		return "g5.xlarge"
	}
	return trimmed
}

func isGpuNodePool(pool infraapi.NodePool) bool {
	inst := strings.ToLower(strings.TrimSpace(pool.InstanceType))
	if strings.HasPrefix(inst, "g4") || strings.HasPrefix(inst, "g5") || strings.HasPrefix(inst, "p2") || strings.HasPrefix(inst, "p3") || strings.HasPrefix(inst, "p4") || strings.HasPrefix(inst, "p5") {
		return true
	}
	for _, t := range pool.Taints {
		key := strings.ToLower(strings.TrimSpace(t.Key))
		if key == "" {
			continue
		}
		if strings.Contains(key, "nvidia.com/gpu") || strings.Contains(key, "mig") {
			return true
		}
	}
	for k, v := range pool.Labels {
		key := strings.ToLower(strings.TrimSpace(k))
		val := strings.ToLower(strings.TrimSpace(v))
		if strings.Contains(key, "gpu") || strings.Contains(val, "gpu") {
			return true
		}
	}
	return false
}

func hasMigTaint(taints []corev1.Taint) bool {
	for _, t := range taints {
		key := strings.ToLower(strings.TrimSpace(t.Key))
		if strings.Contains(key, "mig") {
			return true
		}
	}
	return false
}

func hasGpuNodePool(pools []infraapi.NodePool) bool {
	return slices.ContainsFunc(pools, isGpuNodePool)
}

// gpuNodeSelector derives a nodeSelector from GPU pool labels to target the device plugin to GPU nodes.
// Prefer a specific GPU flavor label; fall back to a generic purpose label if present.
func gpuNodeSelector(pools []infraapi.NodePool) map[string]string {
	flavors := gpuFlavors(pools)
	if len(flavors) == 1 {
		return map[string]string{"aegis.io/gpu-flavor": flavors[0]}
	}
	for _, p := range pools {
		if !isGpuNodePool(p) {
			continue
		}
		if purpose, ok := p.Labels["aegis.dev/purpose"]; ok && strings.TrimSpace(purpose) != "" {
			return map[string]string{"aegis.dev/purpose": strings.TrimSpace(purpose)}
		}
	}
	return nil
}

// gpuFlavors collects unique GPU flavor labels from GPU node pools.
func gpuFlavors(pools []infraapi.NodePool) []string {
	seen := map[string]struct{}{}
	out := []string{}
	for _, p := range pools {
		if !isGpuNodePool(p) {
			continue
		}
		if flavor, ok := p.Labels["aegis.io/gpu-flavor"]; ok {
			trimmed := strings.TrimSpace(flavor)
			if trimmed == "" {
				continue
			}
			if _, exists := seen[trimmed]; !exists {
				seen[trimmed] = struct{}{}
				out = append(out, trimmed)
			}
		}
	}
	return out
}

// gpuMigStrategy returns the desired MIG strategy for the NVIDIA device plugin.
// If any GPU pool is tagged for MIG, return "mixed" so MIG resources are advertised.
func gpuMigStrategy(pools []infraapi.NodePool) string {
	flavors := gpuFlavors(pools)
	for _, f := range flavors {
		if strings.Contains(strings.ToLower(f), "mig") {
			return "mixed"
		}
	}
	for _, p := range pools {
		if !isGpuNodePool(p) {
			continue
		}
		if hasMigTaint(p.Taints) {
			return "mixed"
		}
		for _, v := range p.Labels {
			if strings.Contains(strings.ToLower(strings.TrimSpace(v)), "mig") {
				return "mixed"
			}
		}
	}
	return ""
}

func gpuAmiType(clusterVersion string) string {
	// AL2 GPU AMIs are only supported up to K8s 1.32. Use AL2023 GPU for newer clusters.
	major, minor := parseK8sVersion(clusterVersion)
	if major == 0 && minor == 0 {
		// Unknown/unspecified version defaults to current EKS default (>=1.33) -> use Bottlerocket NVIDIA for GPU.
		return "BOTTLEROCKET_x86_64_NVIDIA"
	}
	if major > 1 || (major == 1 && minor >= 33) {
		// Managed nodegroups don’t expose an AL2023 GPU amiType; use Bottlerocket NVIDIA for 1.33+ GPU pools.
		return "BOTTLEROCKET_x86_64_NVIDIA"
	}
	return "AL2_x86_64_GPU"
}

func parseK8sVersion(version string) (int, int) {
	trimmed := strings.TrimSpace(version)
	if trimmed == "" {
		return 0, 0
	}
	parts := strings.SplitN(strings.TrimPrefix(trimmed, "v"), ".", 3)
	if len(parts) < 2 {
		return 0, 0
	}
	major, _ := strconv.Atoi(parts[0])
	minor, _ := strconv.Atoi(parts[1])
	return major, minor
}

func (r *Runner) resolvePlatformConfig() platformConfig {
	cfg := platformConfig{
		Endpoint:       strings.TrimSpace(os.Getenv(envAegisPlatformEndpoint)),
		CABundleBase64: strings.TrimSpace(os.Getenv(envAegisPlatformCABase64)),
		ImageRepo:      strings.TrimSpace(os.Getenv(envAegisSpokeImageRepo)),
		ImageTag:       strings.TrimSpace(os.Getenv(envAegisSpokeImageTag)),
	}
	if insecure := strings.TrimSpace(os.Getenv(envAegisPlatformInsecure)); insecure != "" {
		if parsed, err := strconv.ParseBool(insecure); err == nil {
			cfg.InsecureGRPC = parsed
		}
	}

	if cfg.CABundleBase64 == "" {
		if caPath := strings.TrimSpace(os.Getenv(envAegisPlatformCAFile)); caPath != "" {
			if raw, err := os.ReadFile(caPath); err == nil {
				cfg.CABundleBase64 = base64.StdEncoding.EncodeToString(raw)
			}
		}
	}
	if cfg.CABundleBase64 == "" {
		if raw, err := os.ReadFile("/etc/aegis/platform-ca.crt"); err == nil {
			cfg.CABundleBase64 = base64.StdEncoding.EncodeToString(raw)
		}
	}
	return cfg
}

func (r *Runner) resolveHelmConfig() helmConfig {
	chart := strings.TrimSpace(os.Getenv(envAegisSpokeChart))
	repo := strings.TrimSpace(os.Getenv(envAegisSpokeRepo))
	if chart == "" {
		repoRoot := r.findRepoRoot()
		chart = filepath.Join(repoRoot, "charts", "aegis-spoke")
	}
	timeout := defaultSpokeReleaseTimeout
	if override := strings.TrimSpace(os.Getenv("AEGIS_SPOKE_HELM_TIMEOUT_SECONDS")); override != "" {
		if v, err := strconv.Atoi(override); err == nil && v > 0 {
			timeout = time.Duration(v) * time.Second
		}
	}
	return helmConfig{
		ChartPath:        chart,
		Repository:       repo,
		Version:          strings.TrimSpace(os.Getenv(envAegisSpokeChartVersion)),
		ValuesFile:       strings.TrimSpace(os.Getenv(envAegisSpokeValuesFile)),
		Namespace:        defaultHelmNamespace,
		ReleaseName:      defaultSpokeReleaseName,
		Timeout:          timeout,
		EnableDependency: true,
	}
}

func (r *Runner) findRepoRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		return "."
	}
	for {
		if _, err := os.Stat(filepath.Join(cwd, ".git")); err == nil {
			return cwd
		}
		parent := filepath.Dir(cwd)
		if parent == cwd {
			break
		}
		cwd = parent
	}
	return "."
}

// ----------------------------------------------------------------------------- //
// Pulumi stack helpers

// stackNameForInfra returns a unique stack name per ProjectInfra to ensure each
// infrastructure resource has its own isolated Pulumi state. This prevents
// deleting one ProjectInfra from affecting others in the same project/region.
func (r *Runner) stackNameForInfra(infraName, projectID, region string) string {
	// Use a hash of the infra name to keep stack names short while ensuring uniqueness
	infraID := sanitize(infraName)
	if len(infraID) > 20 {
		// Truncate and add hash suffix for long names
		h := fnv.New32a()
		h.Write([]byte(infraName))
		infraID = fmt.Sprintf("%s-%x", infraID[:12], h.Sum32())
	}
	return fmt.Sprintf("%s-%s-%s-%s", defaultStackPrefix, sanitize(projectID), sanitize(region), infraID)
}

// stackName returns the legacy shared stack name (deprecated - use stackNameForInfra)
// Kept for backward compatibility during migration.
func (r *Runner) stackName(projectID, region string) string {
	return fmt.Sprintf("%s-%s-%s", defaultStackPrefix, sanitize(projectID), sanitize(region))
}

func (r *Runner) projectName(projectID string) string {
	return fmt.Sprintf("%s-%s", projectNamePrefix, sanitize(projectID))
}

// clearPulumiLock removes stale file-backend locks; for remote backends, stack.Cancel is preferred.
func (r *Runner) clearPulumiLock(projectID, stackName string) {
	backend := strings.TrimSpace(os.Getenv(envPulumiBackendURL))
	basePath := "/tmp/pulumi-backend"
	if backend != "" && strings.HasPrefix(backend, "file://") {
		basePath = strings.TrimPrefix(backend, "file://")
	}
	lockDir := filepath.Join(basePath, ".pulumi", "locks", "organization", r.projectName(projectID), stackName)
	if err := os.RemoveAll(lockDir); err != nil {
		r.log.Warn("failed to clear pulumi lock", zap.String("lock_dir", lockDir), zap.Error(err))
	} else {
		r.log.Info("cleared pulumi locks", zap.String("lock_dir", lockDir))
	}
}

func isLockError(err error) bool {
	if err == nil {
		return false
	}
	lockSnippets := []string{
		"stack is currently locked",
		"currently locked by",
		"lock(s)",
	}
	return slices.ContainsFunc(lockSnippets, func(sub string) bool {
		return strings.Contains(strings.ToLower(err.Error()), strings.ToLower(sub))
	})
}

// retryAfterCancel tries to clear a pulumi backend lock via stack.Cancel then executes fn once.
func (r *Runner) retryAfterCancel(ctx context.Context, stack auto.Stack, fn func() error) bool {
	if err := stack.Cancel(ctx); err != nil {
		r.log.Warn("pulumi cancel failed while clearing lock", zap.Error(err))
	}
	if err := fn(); err != nil {
		return false
	}
	return true
}

// isAlreadyExistsError returns true if the error indicates a resource already exists in AWS.
// This typically happens when Pulumi state has a pending CREATE operation but the resource
// was actually created in a previous interrupted run.
func isAlreadyExistsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	existsSnippets := []string{
		"already exists",
		"resourceinuseexception",
		"entityalreadyexists",
		"alreadyexistsexception",
		"conflict",
	}
	return slices.ContainsFunc(existsSnippets, func(sub string) bool {
		return strings.Contains(errStr, sub)
	})
}

// hasPendingOperations returns true if the error message indicates pending operations exist.
func hasPendingOperations(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "pending operations") ||
		strings.Contains(errStr, "pending operation")
}

// parseEKSClusterNameFromError extracts the EKS cluster name from a "ResourceInUseException" error.
// Example error: "creating EKS Cluster (demo-1-us-east-1-demo-2): ResourceInUseException: Cluster already exists with name: demo-1-us-east-1-demo-2"
func parseEKSClusterNameFromError(err error) string {
	if err == nil {
		return ""
	}
	errStr := err.Error()

	// Look for "Cluster already exists with name: <name>"
	if idx := strings.Index(errStr, "Cluster already exists with name:"); idx >= 0 {
		after := errStr[idx+len("Cluster already exists with name:"):]
		// Find the cluster name - it ends at newline, quote, or brace
		name := strings.TrimSpace(after)
		for i, c := range name {
			if c == '\n' || c == '"' || c == '}' || c == ',' {
				name = name[:i]
				break
			}
		}
		return strings.TrimSpace(name)
	}

	// Fallback: look for "creating EKS Cluster (<name>)"
	if idx := strings.Index(errStr, "creating EKS Cluster ("); idx >= 0 {
		after := errStr[idx+len("creating EKS Cluster ("):]
		if endIdx := strings.Index(after, ")"); endIdx >= 0 {
			return strings.TrimSpace(after[:endIdx])
		}
	}

	return ""
}

// importExistingEKSCluster imports an existing EKS cluster into Pulumi state.
// This is needed when a cluster was created but Pulumi was interrupted, leaving the state out of sync.
func (r *Runner) importExistingEKSCluster(ctx context.Context, stack auto.Stack, clusterName string, progressWriter *logWriter) error {
	if clusterName == "" {
		return fmt.Errorf("cluster name is required")
	}

	r.log.Info("importing existing EKS cluster into Pulumi state",
		zap.String("cluster_name", clusterName))

	// Export current state to manipulate it
	deployment, err := stack.Export(ctx)
	if err != nil {
		return fmt.Errorf("export stack state: %w", err)
	}

	// Parse the deployment state
	var state map[string]interface{}
	if err := json.Unmarshal(deployment.Deployment, &state); err != nil {
		return fmt.Errorf("parse stack state: %w", err)
	}

	// Find the resources array in the deployment
	deploymentData, ok := state["deployment"].(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid deployment structure")
	}

	resources, ok := deploymentData["resources"].([]interface{})
	if !ok {
		return fmt.Errorf("no resources array in deployment")
	}

	// Look for pending EKS cluster resource and update it
	modified := false
	for i, res := range resources {
		resMap, ok := res.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if this is the EKS cluster resource
		resType, _ := resMap["type"].(string)
		urn, _ := resMap["urn"].(string)

		if resType != "aws:eks/cluster:Cluster" {
			continue
		}

		// Check if the URN contains our cluster name
		if !strings.Contains(urn, clusterName) {
			continue
		}

		r.log.Info("found EKS cluster resource in state",
			zap.String("urn", urn),
			zap.String("type", resType))

		// Check if it's in a pending state
		if pending, hasPending := resMap["pending"].(string); hasPending && pending != "" {
			r.log.Info("clearing pending state for EKS cluster",
				zap.String("urn", urn),
				zap.String("pending", pending))
			delete(resMap, "pending")
			modified = true
		}

		// Set the ID if not already set
		if _, hasID := resMap["id"].(string); !hasID || resMap["id"] == "" {
			resMap["id"] = clusterName
			modified = true
			r.log.Info("setting resource ID for EKS cluster",
				zap.String("urn", urn),
				zap.String("id", clusterName))
		}

		// Ensure outputs exist
		if outputs, hasOutputs := resMap["outputs"].(map[string]interface{}); !hasOutputs || outputs == nil {
			resMap["outputs"] = map[string]interface{}{
				"name": clusterName,
				"id":   clusterName,
			}
			modified = true
		}

		resources[i] = resMap
	}

	if !modified {
		r.log.Info("no modifications needed to EKS cluster resource")
		return nil
	}

	// Save the modified state back
	deploymentData["resources"] = resources
	state["deployment"] = deploymentData

	modifiedJSON, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal modified state: %w", err)
	}

	deployment.Deployment = modifiedJSON

	// Import the modified state
	if err := stack.Import(ctx, deployment); err != nil {
		return fmt.Errorf("import modified state: %w", err)
	}

	r.log.Info("successfully modified EKS cluster state")
	return nil
}

// clearPendingOperations exports the stack state, removes any pending operations,
// and reimports the cleaned state. This resolves state conflicts from interrupted operations.
func (r *Runner) clearPendingOperations(ctx context.Context, stack auto.Stack) error {
	r.log.Info("checking for pending operations in stack state")

	// Export current stack state
	deployment, err := stack.Export(ctx)
	if err != nil {
		return fmt.Errorf("export stack state: %w", err)
	}

	// The deployment is an apitype.UntypedDeployment which wraps a JSON-serialized deployment
	// We need to unmarshal it, check for pending_operations, and if found, clear them
	var state map[string]interface{}
	if err := json.Unmarshal(deployment.Deployment, &state); err != nil {
		r.log.Warn("failed to parse stack deployment state", zap.Error(err))
		return nil // Not fatal - continue with operation
	}

	// Check if there are pending operations
	pendingOps, hasPending := state["pending_operations"]
	if !hasPending {
		r.log.Info("no pending operations in stack state")
		return nil
	}

	pendingSlice, ok := pendingOps.([]interface{})
	if !ok || len(pendingSlice) == 0 {
		r.log.Info("no pending operations to clear")
		return nil
	}

	r.log.Warn("found pending operations in stack state, clearing them",
		zap.Int("count", len(pendingSlice)))

	// Log details of pending operations for debugging
	for i, op := range pendingSlice {
		if opMap, ok := op.(map[string]interface{}); ok {
			r.log.Info("pending operation details",
				zap.Int("index", i),
				zap.Any("type", opMap["type"]),
				zap.Any("resource", opMap["resource"]))
		}
	}

	// Clear the pending operations
	state["pending_operations"] = []interface{}{}

	// Serialize the cleaned state
	cleanedData, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("marshal cleaned state: %w", err)
	}

	// Create a new deployment with cleaned state
	deployment.Deployment = cleanedData

	// Import the cleaned state back
	if err := stack.Import(ctx, deployment); err != nil {
		return fmt.Errorf("import cleaned stack state: %w", err)
	}

	r.log.Info("successfully cleared pending operations from stack state")
	return nil
}

// runUpWithRetry runs pulumi up with automatic retry on recoverable errors.
// If the operation fails due to "already exists" errors (from pending CREATE operations),
// it will clear pending operations using ClearPendingCreates and retry once.
//
// IMPORTANT: The key fix here is using optrefresh.ClearPendingCreates() which properly
// handles pending CREATE operations. Simply clearing the pending_operations array is
// NOT sufficient because Pulumi also tracks pending creates in the resource entries
// themselves. ClearPendingCreates drops those resources from state so refresh can
// re-discover them from AWS.
func (r *Runner) runUpWithRetry(ctx context.Context, stack auto.Stack, progressWriter *logWriter) error {
	// First, proactively run a refresh with ClearPendingCreates to handle any
	// leftover state from previous interrupted operations
	r.log.Info("running pre-emptive refresh with ClearPendingCreates to resolve any stale state")
	if _, err := stack.Refresh(ctx,
		optrefresh.ProgressStreams(progressWriter),
		optrefresh.ClearPendingCreates(),
	); err != nil {
		r.log.Warn("pre-emptive refresh with ClearPendingCreates failed", zap.Error(err))
		// Continue anyway - the operation might still succeed
	}

	// Run the first attempt
	_, err := stack.Up(ctx, optup.ProgressStreams(progressWriter))
	if err == nil {
		return nil
	}

	// Check if this is a recoverable error
	if !isAlreadyExistsError(err) && !hasPendingOperations(err) {
		// Not a recoverable error, return as-is
		return err
	}

	r.log.Warn("pulumi up failed with recoverable error, attempting state repair and retry",
		zap.Error(err))

	// Try to import the existing resource if it's an "already exists" error.
	// ClearPendingCreates only removes resources from state - it does NOT re-import them.
	// We need to explicitly import existing AWS resources into Pulumi state.
	if isAlreadyExistsError(err) {
		clusterName := parseEKSClusterNameFromError(err)
		if clusterName != "" {
			r.log.Info("attempting to import existing EKS cluster into Pulumi state",
				zap.String("cluster_name", clusterName))
			if importErr := r.importExistingEKSCluster(ctx, stack, clusterName, progressWriter); importErr != nil {
				r.log.Warn("failed to import existing EKS cluster", zap.Error(importErr))
				// Continue anyway - will try other recovery methods
			} else {
				r.log.Info("successfully imported existing EKS cluster", zap.String("cluster_name", clusterName))
			}
		}
	}

	// Clear pending operations from state file as additional cleanup
	if clearErr := r.clearPendingOperations(ctx, stack); clearErr != nil {
		r.log.Warn("failed to clear pending operations array", zap.Error(clearErr))
		// Continue anyway
	}

	// Run refresh to sync state with AWS after import
	r.log.Info("running refresh to sync state with AWS after import")
	if _, refreshErr := stack.Refresh(ctx,
		optrefresh.ProgressStreams(progressWriter),
	); refreshErr != nil {
		r.log.Warn("refresh failed", zap.Error(refreshErr))
		// Continue anyway - the up might still work
	}

	// Retry the up operation
	r.log.Info("retrying pulumi up after state repair")
	_, retryErr := stack.Up(ctx, optup.ProgressStreams(progressWriter))
	if retryErr != nil {
		return fmt.Errorf("pulumi up retry failed: %w (original error: %v)", retryErr, err)
	}

	r.log.Info("pulumi up succeeded after state repair and retry")
	return nil
}

// ----------------------------------------------------------------------------- //
// Utility helpers

type logWriter struct {
	log       *zap.Logger
	sink      store.ProvisioningLogSink
	jobID     string
	projectID string
	clusterID string
	phase     string
}

func (w *logWriter) Write(p []byte) (int, error) {
	msg := strings.TrimSpace(string(p))
	if msg != "" {
		if w.log != nil {
			w.log.Info("pulumi progress", zap.String("line", msg), zap.String("job", w.jobID))
		}
		if w.sink != nil && strings.TrimSpace(w.jobID) != "" {
			entry := store.ProvisioningLogEntry{
				JobID:     w.jobID,
				ProjectID: w.projectID,
				ClusterID: w.clusterID,
				Phase:     w.phase,
				Type:      store.LogTypeProgress,
				Message:   msg,
				CreatedAt: time.Now().UTC(),
			}
			w.sink.AppendProvisioningLog(entry)
		}
	}
	return len(p), nil
}

func jobMetadata(infra *infraapi.ProjectInfra) (jobID, clusterID string) {
	if infra == nil {
		return "", ""
	}
	jobID = strings.TrimSpace(infra.Name)
	if infra.Annotations != nil {
		if ann := strings.TrimSpace(infra.Annotations["aegis.yourorg.dev/clusterId"]); ann != "" {
			clusterID = ann
		}
	}
	if clusterID == "" && infra.Spec.Aws != nil {
		clusterID = strings.TrimSpace(infra.Spec.Aws.ClusterName)
	}
	return jobID, clusterID
}

func convertTaints(taints []corev1.Taint) awseks.NodeGroupTaintArray {
	if len(taints) == 0 {
		return nil
	}
	out := make(awseks.NodeGroupTaintArray, 0, len(taints))
	for _, t := range taints {
		key := strings.TrimSpace(t.Key)
		if key == "" {
			continue
		}
		value := strings.TrimSpace(t.Value)
		effect := formatTaintEffect(string(t.Effect))
		taintArgs := awseks.NodeGroupTaintArgs{
			Key:    pulumi.String(key),
			Effect: pulumi.String(effect),
		}
		if value != "" {
			taintArgs.Value = pulumi.StringPtr(value)
		}
		out = append(out, taintArgs)
	}
	return out
}

func formatTaintEffect(effect string) string {
	switch strings.ToLower(strings.TrimSpace(effect)) {
	case "noschedule", "no_schedule":
		return "NO_SCHEDULE"
	case "noexecute", "no_execute":
		return "NO_EXECUTE"
	case "prefernoschedule", "prefer_no_schedule":
		return "PREFER_NO_SCHEDULE"
	default:
		return strings.ToUpper(strings.ReplaceAll(effect, "-", "_"))
	}
}

func grpcEndpointHostPort(endpoint string) string {
	trimmed := strings.TrimSpace(endpoint)
	if trimmed == "" {
		return ""
	}
	for _, prefix := range []string{"grpcs://", "grpc://", "https://", "http://"} {
		if strings.HasPrefix(strings.ToLower(trimmed), strings.ToLower(prefix)) {
			trimmed = trimmed[len(prefix):]
			break
		}
	}
	if slash := strings.IndexRune(trimmed, '/'); slash >= 0 {
		trimmed = trimmed[:slash]
	}
	return strings.TrimSuffix(trimmed, "/")
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

// buildClusterIDWithSuffix creates a unique cluster ID by appending a short hash suffix.
// This ensures each ProjectInfra gets its own unique EKS cluster, preventing conflicts
// when multiple ProjectInfras use the same base cluster name.
func buildClusterIDWithSuffix(projectID, region, name, suffix string) string {
	base := buildClusterID(projectID, region, name)
	if suffix == "" {
		return base
	}
	// EKS cluster name limit is 100 chars. Ensure we stay within that.
	// Format: {base}-{suffix} where suffix is 8 chars
	maxBase := 100 - len(suffix) - 1
	if len(base) > maxBase {
		base = base[:maxBase]
	}
	return fmt.Sprintf("%s-%s", base, suffix)
}

// shortHash generates an 8-character hash from the input string.
// Used to create unique suffixes for cluster names.
func shortHash(input string) string {
	if input == "" {
		return ""
	}
	h := fnv.New32a()
	h.Write([]byte(input))
	return fmt.Sprintf("%08x", h.Sum32())
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

// extractAccountIDFromARN extracts the AWS account ID from an IAM ARN.
// ARN format: arn:aws:iam::ACCOUNT_ID:role/role-name
func extractAccountIDFromARN(arn string) string {
	parts := strings.Split(arn, ":")
	if len(parts) >= 5 {
		return parts[4]
	}
	return ""
}

func pulumiResourceName(base string, max int) string {
	name := sanitize(base)
	if max <= 0 {
		return name
	}
	if len(name) <= max {
		return name
	}
	hash := sha1.Sum([]byte(name))
	suffix := hex.EncodeToString(hash[:])[:8]
	trim := max - len(suffix) - 1
	if trim <= 0 {
		return suffix
	}
	if trim > len(name) {
		trim = len(name)
	}
	trimmed := strings.TrimRight(name[:trim], "-")
	if trimmed == "" {
		return suffix
	}
	return fmt.Sprintf("%s-%s", trimmed, suffix)
}

func addonEnabled(addons map[string]bool, key string, defaultVal bool) bool {
	if len(addons) == 0 {
		return defaultVal
	}
	for k, v := range addons {
		if strings.EqualFold(strings.TrimSpace(k), strings.TrimSpace(key)) {
			return v
		}
	}
	return defaultVal
}

func sanitizeStrings(values []string) []string {
	out := []string{}
	for _, v := range values {
		trimmed := strings.TrimSpace(v)
		if trimmed == "" {
			continue
		}
		out = append(out, trimmed)
	}
	return out
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

// getSpokeProxyNodePort returns the NodePort to use for the spoke-proxy service.
// Configurable via AEGIS_SPOKE_PROXY_NODEPORT environment variable, defaults to 31484.
func getSpokeProxyNodePort() int {
	if portStr := strings.TrimSpace(os.Getenv(envAegisSpokeProxyNodePort)); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil && port >= 30000 && port <= 32767 {
			return port
		}
	}
	return defaultSpokeProxyNodePort
}

// generateSpokeProxyCert generates a self-signed TLS certificate for the spoke-proxy.
// The certificate includes wildcard SANs for *.nip.io to work with any node public IP.
// Returns PEM-encoded certificate and key.
func generateSpokeProxyCert() (certPEM, keyPEM string, err error) {
	// Check if we have a CA to sign with
	caCertPath := os.Getenv(envAegisSpokeProxyCACert)
	caKeyPath := os.Getenv(envAegisSpokeProxyCAKey)

	var caCert *x509.Certificate
	var caKey *rsa.PrivateKey

	if caCertPath != "" && caKeyPath != "" {
		// Load CA certificate and key
		caCert, caKey, err = loadCA(caCertPath, caKeyPath)
		if err != nil {
			// Log warning but fall back to self-signed
			fmt.Printf("WARNING: failed to load spoke-proxy CA, falling back to self-signed: %v\n", err)
			caCert, caKey = nil, nil
		}
	}

	// Generate RSA key for the spoke-proxy cert
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", "", fmt.Errorf("generate private key: %w", err)
	}

	// Create certificate template
	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return "", "", fmt.Errorf("generate serial number: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Aegis Platform"},
			CommonName:   "spoke-proxy.aegis.local",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0), // Valid for 1 year
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false, // End-entity certificate, not a CA
		DNSNames: []string{
			"spoke-proxy.aegis.local",
			"*.nip.io",             // Wildcard for any IP.nip.io
			"spoke-proxy.*.nip.io", // spoke-proxy prefix wildcard
			"*.*.nip.io",           // Double wildcard for spoke-proxy.IP.nip.io
			"localhost",
		},
		IPAddresses: []net.IP{
			net.ParseIP("127.0.0.1"),
		},
	}

	var certDER []byte
	if caCert != nil && caKey != nil {
		// Sign with CA
		certDER, err = x509.CreateCertificate(rand.Reader, &template, caCert, &privateKey.PublicKey, caKey)
		if err != nil {
			return "", "", fmt.Errorf("create CA-signed certificate: %w", err)
		}
	} else {
		// Self-signed fallback - make it a CA so it can be added to trust bundle
		template.IsCA = true
		template.KeyUsage |= x509.KeyUsageCertSign
		certDER, err = x509.CreateCertificate(rand.Reader, &template, &template, &privateKey.PublicKey, privateKey)
		if err != nil {
			return "", "", fmt.Errorf("create self-signed certificate: %w", err)
		}
	}

	// Encode to PEM
	certPEMBytes := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEMBytes := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(privateKey)})

	return string(certPEMBytes), string(keyPEMBytes), nil
}

// loadCA loads a CA certificate and private key from PEM files.
func loadCA(certPath, keyPath string) (*x509.Certificate, *rsa.PrivateKey, error) {
	// Read CA certificate
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read CA cert: %w", err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode CA cert PEM")
	}
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("parse CA cert: %w", err)
	}

	// Read CA private key
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read CA key: %w", err)
	}
	block, _ = pem.Decode(keyPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("failed to decode CA key PEM")
	}

	// Try parsing as PKCS1 first, then PKCS8
	caKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		// Try PKCS8
		key, err2 := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err2 != nil {
			return nil, nil, fmt.Errorf("parse CA key (tried PKCS1 and PKCS8): %w", err)
		}
		var ok bool
		caKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return nil, nil, fmt.Errorf("CA key is not RSA")
		}
	}

	return caCert, caKey, nil
}
