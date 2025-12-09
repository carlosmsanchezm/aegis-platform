package aws

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
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
	"github.com/yourorg/aegis/services/platform-api/internal/provisioning/observability"
)

const (
	projectNamePrefix          = "aegis-platform"
	defaultStackPrefix         = "aegis"
	defaultHelmNamespace       = "aegis-system"
	defaultWorkloadsNamespace  = "aegis-workloads"
	defaultSpokeReleaseName    = "aegis-spoke"
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
	log *zap.Logger
}

// NewRunner returns a new automation-backed AWS runner.
func NewRunner(log *zap.Logger) *Runner {
	if log == nil {
		log = zap.NewNop()
	}
	return &Runner{log: log.Named("aws-provisioner")}
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
	r.clearPulumiLock(programCfg.ProjectID, r.stackName(programCfg.ProjectID, programCfg.Region))

	progressWriter := &logWriter{log: r.log}

	if skip := strings.EqualFold(os.Getenv(envAegisPulumiSkipRefresh), "true"); !skip {
		if _, err := stack.Refresh(ctx, optrefresh.ProgressStreams(progressWriter)); err != nil {
			if !isLockError(err) || !r.retryAfterCancel(ctx, stack, func() error {
				_, retryErr := stack.Refresh(ctx, optrefresh.ProgressStreams(progressWriter))
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
		if _, err := stack.Up(ctx, optup.ProgressStreams(progressWriter)); err != nil {
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

// Destroy tears down the stack for the specified project/region combination.
func (r *Runner) Destroy(ctx context.Context, infra *infraapi.ProjectInfra, spec *infraapi.AWSInfraSpec) error {
	if infra == nil || spec == nil {
		return nil
	}
	projectID := strings.TrimSpace(infra.Spec.ProjectID)
	region := strings.TrimSpace(infra.Spec.Region)
	if projectID == "" || region == "" {
		return nil
	}

	stackName := r.stackName(projectID, region)
	projectName := r.projectName(projectID)
	program := r.buildPulumiProgram(&programInput{
		ProjectID: projectID,
		Region:    region,
		SkipHelm:  true,
	})
	stack, err := auto.SelectStackInlineSource(ctx, stackName, projectName, program, r.buildWorkspaceOptions()...)
	if auto.IsSelectStack404Error(err) {
		r.log.Info("pulumi stack not present; skipping destroy", zap.String("stack", stackName))
		return nil
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

	progressWriter := &logWriter{log: r.log}
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

	clusterDefs, err := r.buildClusterDefinitions(projectID, region, spec)
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
		SkipHelm:            strings.EqualFold(os.Getenv("AEGIS_SKIP_SPOKE_HELM"), "true"),
		EnableCostEstimates: true,
	}

	stackName := r.stackName(projectID, region)
	projectName := r.projectName(projectID)
	program := r.buildPulumiProgram(programCfg)
	stack, err := auto.UpsertStackInlineSource(ctx, stackName, projectName, program, r.buildWorkspaceOptions()...)
	if err != nil {
		return auto.Stack{}, nil, fmt.Errorf("upsert pulumi stack: %w", err)
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

			// Install Cluster Autoscaler to enable automatic node scaling.
			// This is essential for on-demand GPU provisioning - when a workspace
			// requests GPU resources, the autoscaler will scale up the GPU node group.
			if err := r.installClusterAutoscaler(ctx, clusterDef.ClusterID, cluster.Name, input.Region, kubeProvider, append(nodeGroups, cluster)); err != nil {
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

			if !input.SkipHelm {
				if err := r.installSpokeHelmChart(ctx, clusterDef.ClusterID, kubeProvider, kubeconfig, input, append(nodeGroups, cluster)); err != nil {
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
			instanceType = "m6i.large"
		}
		gpuPool := isGpuNodePool(pool)
		scaling := &awseks.NodeGroupScalingConfigArgs{
			DesiredSize: pulumi.Int(int(pool.MinSize)),
			MinSize:     pulumi.Int(int(pool.MinSize)),
			MaxSize:     pulumi.Int(int(pool.MaxSize)),
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
	k8sAgentValues := pulumi.Map{
		"env": envValues,
	}
	if input.Platform.ImageRepo != "" {
		k8sAgentValues["image"] = pulumi.Map{
			"repository": pulumi.String(input.Platform.ImageRepo),
			"tag":        pulumi.String(input.Platform.ImageTag),
		}
	}
	values := pulumi.Map{"k8sAgent": k8sAgentValues}

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
func (r *Runner) installClusterAutoscaler(ctx *pulumi.Context, clusterID string, clusterName pulumi.StringInput, region string, kubeProvider *kubernetes.Provider, depends []pulumi.Resource) error {
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
			"scale-down-enabled":            pulumi.Bool(true),
			"scale-down-delay-after-add":    pulumi.String("5m"),
			"scale-down-unneeded-time":      pulumi.String("5m"),
			"scale-down-utilization-threshold": pulumi.String("0.5"),
			"skip-nodes-with-local-storage": pulumi.Bool(false),
			"skip-nodes-with-system-pods":   pulumi.Bool(false),
			"balance-similar-node-groups":   pulumi.Bool(true),
			"expander":                      pulumi.String("least-waste"),
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
		"rbac": pulumi.Map{
			"create": pulumi.Bool(true),
			"serviceAccount": pulumi.Map{
				"create": pulumi.Bool(true),
				"name":   pulumi.String("cluster-autoscaler"),
				// In production, use IRSA instead of node IAM role
				"annotations": pulumi.Map{},
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

func (r *Runner) buildClusterDefinitions(projectID, region string, spec *infraapi.AWSInfraSpec) ([]clusterDefinition, error) {
	var defs []clusterDefinition

	appendCluster := func(name, version string, pools []infraapi.NodePool) {
		if strings.TrimSpace(name) == "" {
			return
		}
		clusterID := buildClusterID(projectID, region, name)
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
		InstanceType: "m6i.large",
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

// ----------------------------------------------------------------------------- //
// Utility helpers

type logWriter struct {
	log *zap.Logger
}

func (w *logWriter) Write(p []byte) (int, error) {
	msg := strings.TrimSpace(string(p))
	if msg != "" {
		w.log.Info("pulumi progress", zap.String("line", msg))
	}
	return len(p), nil
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

func sanitize(in string) string {
	trimmed := strings.ToLower(strings.TrimSpace(in))
	replacer := strings.NewReplacer(" ", "-", "_", "-", ".", "-")
	trimmed = replacer.Replace(trimmed)
	for strings.Contains(trimmed, "--") {
		trimmed = strings.ReplaceAll(trimmed, "--", "-")
	}
	return strings.Trim(trimmed, "-")
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
