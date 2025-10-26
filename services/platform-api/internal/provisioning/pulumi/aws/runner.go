package aws

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws"
	awseks "github.com/pulumi/pulumi-aws/sdk/v5/go/aws/eks"
	"github.com/pulumi/pulumi-eks/sdk/go/eks"
	kubernetes "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/auto"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optdestroy"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optrefresh"
	"github.com/pulumi/pulumi/sdk/v3/go/auto/optup"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/provisioning"
)

const (
	projectNamePrefix          = "aegis-platform"
	defaultStackPrefix         = "aegis"
	defaultHelmNamespace       = "aegis-system"
	defaultSpokeReleaseName    = "aegis-spoke"
	envPulumiBackendURL        = "PULUMI_BACKEND_URL"
	envAegisSpokeChart         = "AEGIS_SPOKE_CHART"
	envAegisSpokeChartVersion  = "AEGIS_SPOKE_CHART_VERSION"
	envAegisSpokeRepo          = "AEGIS_SPOKE_REPO"
	envAegisSpokeValuesFile    = "AEGIS_SPOKE_VALUES_FILE"
	envAegisPlatformEndpoint   = "AEGIS_PLATFORM_API_ENDPOINT"
	envAegisPlatformCABase64   = "AEGIS_PLATFORM_API_CA_B64"
	envAegisSpokeImageRepo     = "AEGIS_SPOKE_IMAGE_REPO"
	envAegisSpokeImageTag      = "AEGIS_SPOKE_IMAGE_TAG"
	envAegisPulumiSkipRefresh  = "AEGIS_PULUMI_SKIP_REFRESH"
	envAegisPulumiSkipApply    = "AEGIS_PULUMI_SKIP_APPLY"
	envAegisPulumiWorkdir      = "AEGIS_PULUMI_WORKDIR"
	defaultSpokeReleaseTimeout = 15 * time.Minute
)

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

	progressWriter := &logWriter{log: r.log}

	if skip := strings.EqualFold(os.Getenv(envAegisPulumiSkipRefresh), "true"); !skip {
		if _, err := stack.Refresh(ctx, optrefresh.ProgressStreams(progressWriter)); err != nil {
			return nil, fmt.Errorf("pulumi refresh: %w", err)
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
	destroyRes, err := stack.Destroy(ctx, optdestroy.ProgressStreams(progressWriter))
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
		RoleARN:             strings.TrimSpace(spec.RoleARN),
		ExternalID:          strings.TrimSpace(spec.ExternalID),
		Clusters:            clusterDefs,
		Platform:            r.resolvePlatformConfig(),
		SpokeHelm:           r.resolveHelmConfig(),
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
	RoleARN             string
	ExternalID          string
	Clusters            []clusterDefinition
	Platform            platformConfig
	SpokeHelm           helmConfig
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
		{Name: "eks", Version: "1.0.4"},
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

		clusterMap := pulumi.Map{}
		kubeconfigMap := pulumi.Map{}

		for idx, clusterDef := range input.Clusters {
			clusterName := fmt.Sprintf("%s-%d", sanitize(clusterDef.Name), idx)
			if clusterDef.ClusterID != "" {
				clusterName = sanitize(clusterDef.ClusterID)
			}
			skipDefault := true
			clusterIDLabel := clusterDef.ClusterID
			if clusterIDLabel == "" {
				clusterIDLabel = clusterName
			}
			clusterArgs := &eks.ClusterArgs{
				SkipDefaultNodeGroup: &skipDefault,
				Version:              pulumi.StringPtr(strings.TrimSpace(clusterDef.Version)),
				Tags: pulumi.StringMap{
					"Project": pulumi.String(input.ProjectID),
					"Cluster": pulumi.String(clusterIDLabel),
				},
			}
			if input.VpcID != "" {
				clusterArgs.VpcId = pulumi.StringPtr(input.VpcID)
			}

			cluster, err := eks.NewCluster(ctx, clusterName, clusterArgs, providerOpt)
			if err != nil {
				return fmt.Errorf("create eks cluster %q: %w", clusterDef.ClusterID, err)
			}

			if err := r.configureManagedNodeGroups(ctx, clusterDef, cluster, awsProvider); err != nil {
				return err
			}

			if !input.SkipHelm {
				if err := r.installSpokeHelmChart(ctx, clusterDef.ClusterID, cluster, input); err != nil {
					return err
				}
			}

			key := clusterDef.ClusterID
			if key == "" {
				key = clusterName
			}
			kubeconfigKey := fmt.Sprintf("%s.kubeconfig", key)
			kubeconfigMap[key] = cluster.KubeconfigJson
			clusterMap[key] = pulumi.Map{
				"clusterId":           pulumi.String(key),
				"name":                pulumi.String(clusterDef.Name),
				"region":              pulumi.String(input.Region),
				"kubeconfigSecretKey": pulumi.String(kubeconfigKey),
			}
		}

		ctx.Export("clusters", clusterMap)
		ctx.Export("kubeconfigs", kubeconfigMap)
		return nil
	}
}

func (r *Runner) configureManagedNodeGroups(ctx *pulumi.Context, clusterDef clusterDefinition, cluster *eks.Cluster, provider *aws.Provider) error {
	if len(clusterDef.NodePools) == 0 {
		return nil
	}
	instanceRoles := cluster.Core.InstanceRoles()
	role := instanceRoles.Index(pulumi.Int(0))

	for _, pool := range clusterDef.NodePools {
		name := strings.TrimSpace(pool.Name)
		if name == "" {
			name = fmt.Sprintf("%s-nodepool", sanitize(clusterDef.ClusterID))
		}
		instanceType := strings.TrimSpace(pool.InstanceType)
		if instanceType == "" {
			instanceType = "m6i.large"
		}
		scaling := awseks.NodeGroupScalingConfigArgs{
			DesiredSize: pulumi.Int(int(pool.MinSize)),
			MinSize:     pulumi.Int(int(pool.MinSize)),
			MaxSize:     pulumi.Int(int(pool.MaxSize)),
		}
		mp := &eks.ManagedNodeGroupArgs{
			Cluster:       cluster,
			ClusterName:   cluster.EksCluster.Name().ToStringPtrOutput(),
			NodeGroupName: pulumi.StringPtr(name),
			NodeRole:      role,
			InstanceTypes: pulumi.StringArray{
				pulumi.String(instanceType),
			},
			ScalingConfig: scaling.ToNodeGroupScalingConfigPtrOutput(),
			Taints:        convertTaints(pool.Taints),
		}
		if len(pool.Labels) > 0 {
			labels := pulumi.StringMap{}
			for k, v := range pool.Labels {
				labels[k] = pulumi.String(v)
			}
			mp.Labels = labels
		}

		if _, err := eks.NewManagedNodeGroup(ctx, fmt.Sprintf("%s-nodegroup", name), mp, pulumi.Provider(provider), pulumi.DependsOn([]pulumi.Resource{cluster})); err != nil {
			return fmt.Errorf("create managed node group %q: %w", name, err)
		}
	}
	return nil
}

func (r *Runner) installSpokeHelmChart(ctx *pulumi.Context, clusterID string, cluster *eks.Cluster, input *programInput) error {
	kubeProvider, err := kubernetes.NewProvider(ctx, fmt.Sprintf("%s-k8s", sanitize(clusterID)), &kubernetes.ProviderArgs{
		Kubeconfig: cluster.KubeconfigJson,
	}, pulumi.DependsOn([]pulumi.Resource{cluster}))
	if err != nil {
		return fmt.Errorf("create kubernetes provider: %w", err)
	}

	envValues := pulumi.Map{
		"AEGIS_CLUSTER_ID": pulumi.String(clusterID),
	}
	if input.Platform.Endpoint != "" {
		envValues["AEGIS_PLATFORM_API_ENDPOINT"] = pulumi.String(input.Platform.Endpoint)
	}
	if input.Platform.CABundleBase64 != "" {
		envValues["AEGIS_PLATFORM_CA_B64"] = pulumi.String(input.Platform.CABundleBase64)
	}
	values := pulumi.Map{"env": envValues}
	if input.Platform.ImageRepo != "" {
		values["image"] = pulumi.Map{
			"repository": pulumi.String(input.Platform.ImageRepo),
			"tag":        pulumi.String(input.Platform.ImageTag),
		}
	}

	helmCfg := input.SpokeHelm
	if helmCfg.Namespace == "" {
		helmCfg.Namespace = defaultHelmNamespace
	}
	if helmCfg.ReleaseName == "" {
		helmCfg.ReleaseName = fmt.Sprintf("%s-%s", defaultSpokeReleaseName, sanitize(clusterID))
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

	if _, err := helm.NewRelease(ctx, fmt.Sprintf("%s-release", sanitize(clusterID)), releaseArgs, pulumi.Provider(kubeProvider)); err != nil {
		return fmt.Errorf("install aegis-spoke for cluster %q: %w", clusterID, err)
	}
	return nil
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

		result.Outputs = append(result.Outputs, infraapi.ClusterOutput{
			ClusterID:           clusterID,
			Name:                name,
			Region:              regionOut,
			KubeconfigSecretKey: secretKey,
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
		pools := spec.NodePools
		if len(pools) == 0 {
			pools = defaultNodePools()
		}
		appendCluster(primary, spec.Version, pools)
	}
	for _, additional := range spec.AdditionalClusters {
		pools := additional.NodePools
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

func (r *Runner) resolvePlatformConfig() platformConfig {
	cfg := platformConfig{
		Endpoint:       strings.TrimSpace(os.Getenv(envAegisPlatformEndpoint)),
		CABundleBase64: strings.TrimSpace(os.Getenv(envAegisPlatformCABase64)),
		ImageRepo:      strings.TrimSpace(os.Getenv(envAegisSpokeImageRepo)),
		ImageTag:       strings.TrimSpace(os.Getenv(envAegisSpokeImageTag)),
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

func (r *Runner) stackName(projectID, region string) string {
	return fmt.Sprintf("%s-%s-%s", defaultStackPrefix, sanitize(projectID), sanitize(region))
}

func (r *Runner) projectName(projectID string) string {
	return fmt.Sprintf("%s-%s", projectNamePrefix, sanitize(projectID))
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

// ----------------------------------------------------------------------------- //
// Legacy helpers retained from initial stub

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
