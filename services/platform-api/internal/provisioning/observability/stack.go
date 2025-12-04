package observability

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	envObservabilityChart         = "AEGIS_OBSERVABILITY_CHART"
	envObservabilityChartVersion  = "AEGIS_OBSERVABILITY_CHART_VERSION"
	envObservabilityRepo          = "AEGIS_OBSERVABILITY_REPO"
	envObservabilityValuesFile    = "AEGIS_OBSERVABILITY_VALUES_FILE"
	envObservabilityEnabled       = "AEGIS_OBSERVABILITY_ENABLED"
	envMetricsServerChart         = "AEGIS_METRICS_SERVER_CHART"
	envMetricsServerVersion       = "AEGIS_METRICS_SERVER_CHART_VERSION"
	envMetricsServerRepo          = "AEGIS_METRICS_SERVER_REPO"
	envMetricsServerValuesFile    = "AEGIS_METRICS_SERVER_VALUES_FILE"
	defaultObservabilityNamespace = "aegis-observability"
	defaultObservabilityRelease   = "aegis-obsv"
	defaultObservabilityBaseName  = "aegis-obsv"
	defaultMetricsRelease         = "aegis-metrics"
	defaultPrometheusPort         = 9090
	defaultAlertmanagerPort       = 9093
	maxObservabilityNameLength    = 40
	defaultHelmTimeout            = 15 * time.Minute
)

// HelmConfig describes a helm release configuration.
type HelmConfig struct {
	ChartPath        string
	Repository       string
	Version          string
	ValuesFile       string
	Namespace        string
	ReleaseName      string
	Timeout          time.Duration
	EnableDependency bool
}

// Config captures the observability stack settings.
type Config struct {
	Stack            HelmConfig
	MetricsServer    HelmConfig
	Namespace        string
	BaseName         string
	PrometheusPort   int
	AlertmanagerPort int
	Enable           bool
}

// Installer allows swapping observability implementations.
type Installer interface {
	Install(ctx *pulumi.Context, clusterID string, kubeProvider *kubernetes.Provider, cfg Config, depends []pulumi.Resource) (pulumi.Map, error)
}

type defaultInstaller struct{}

// NewInstaller returns the default installer implementation.
func NewInstaller() Installer {
	return defaultInstaller{}
}

// ResolveFromEnv builds a Config using environment variables and defaults.
func ResolveFromEnv(repoRoot string) Config {
	if repoRoot == "" {
		repoRoot = findRepoRoot()
	}

	enabled := strings.EqualFold(strings.TrimSpace(os.Getenv(envObservabilityEnabled)), "true")

	stackChart := strings.TrimSpace(os.Getenv(envObservabilityChart))
	if stackChart == "" {
		stackChart = "kube-prometheus-stack"
	}
	stackRepo := strings.TrimSpace(os.Getenv(envObservabilityRepo))
	if stackRepo == "" {
		stackRepo = "https://prometheus-community.github.io/helm-charts"
	}
	stackValues := strings.TrimSpace(os.Getenv(envObservabilityValuesFile))
	if stackValues == "" {
		stackValues = filepath.Join(repoRoot, "services", "platform-api", "config", "observability", "values-mvp.yaml")
		if !fileExists(stackValues) {
			stackValues = ""
		}
	}

	msChart := strings.TrimSpace(os.Getenv(envMetricsServerChart))
	if msChart == "" {
		msChart = "metrics-server"
	}
	msRepo := strings.TrimSpace(os.Getenv(envMetricsServerRepo))
	if msRepo == "" {
		msRepo = "https://kubernetes-sigs.github.io/metrics-server/"
	}
	msValues := strings.TrimSpace(os.Getenv(envMetricsServerValuesFile))
	if msValues == "" {
		msValues = filepath.Join(repoRoot, "services", "platform-api", "config", "observability", "metrics-server-values.yaml")
		if !fileExists(msValues) {
			msValues = ""
		}
	}

	timeout := defaultHelmTimeout
	return Config{
		Namespace:        defaultObservabilityNamespace,
		BaseName:         defaultObservabilityBaseName,
		PrometheusPort:   defaultPrometheusPort,
		AlertmanagerPort: defaultAlertmanagerPort,
		Enable:           enabled,
		Stack: HelmConfig{
			ChartPath:        stackChart,
			Repository:       stackRepo,
			Version:          strings.TrimSpace(os.Getenv(envObservabilityChartVersion)),
			ValuesFile:       stackValues,
			Namespace:        defaultObservabilityNamespace,
			ReleaseName:      defaultObservabilityRelease,
			Timeout:          timeout,
			EnableDependency: true,
		},
		MetricsServer: HelmConfig{
			ChartPath:        msChart,
			Repository:       msRepo,
			Version:          strings.TrimSpace(os.Getenv(envMetricsServerVersion)),
			ValuesFile:       msValues,
			Namespace:        defaultObservabilityNamespace,
			ReleaseName:      defaultMetricsRelease,
			Timeout:          timeout,
			EnableDependency: true,
		},
	}
}

// Install installs the observability stack and returns the outputs map suitable for ClusterOutput.Observability.
func (defaultInstaller) Install(ctx *pulumi.Context, clusterID string, kubeProvider *kubernetes.Provider, cfg Config, depends []pulumi.Resource) (pulumi.Map, error) {
	if !cfg.Enable {
		return nil, nil
	}
	namespace := strings.TrimSpace(cfg.Namespace)
	if namespace == "" {
		namespace = defaultObservabilityNamespace
	}
	baseName := strings.TrimSpace(cfg.BaseName)
	if baseName == "" {
		baseName = defaultObservabilityBaseName
	}
	clusterKey := sanitize(clusterID)
	if clusterKey == "" {
		clusterKey = "aegis"
	}

	promPort := cfg.PrometheusPort
	if promPort == 0 {
		promPort = defaultPrometheusPort
	}
	alertPort := cfg.AlertmanagerPort
	if alertPort == 0 {
		alertPort = defaultAlertmanagerPort
	}

	stackReleaseName := pulumiResourceName(cfg.Stack.ReleaseName+"-"+clusterKey, 53)
	stackFullname := pulumiResourceName(baseName+"-"+clusterKey, maxObservabilityNameLength)
	metricsFullname := pulumiResourceName(stackFullname+"-ms", maxObservabilityNameLength)
	metricsReleaseName := pulumiResourceName(cfg.MetricsServer.ReleaseName+"-"+clusterKey, 53)

	stackTimeout := cfg.Stack.Timeout
	if stackTimeout == 0 {
		stackTimeout = defaultHelmTimeout
	}
	stackValues := pulumi.Map{
		"fullnameOverride": pulumi.String(stackFullname),
		"alertmanager": pulumi.Map{
			"service": pulumi.Map{
				"type": pulumi.String("ClusterIP"),
				"port": pulumi.Int(alertPort),
			},
		},
		"prometheus": pulumi.Map{
			"service": pulumi.Map{
				"type": pulumi.String("ClusterIP"),
				"port": pulumi.Int(promPort),
			},
		},
	}

	stackArgs := &helm.ReleaseArgs{
		Name:            pulumi.StringPtr(stackReleaseName),
		Namespace:       pulumi.StringPtr(namespace),
		Chart:           pulumi.String(cfg.Stack.ChartPath),
		Values:          stackValues,
		Timeout:         pulumi.IntPtr(int(stackTimeout.Seconds())),
		CreateNamespace: pulumi.BoolPtr(true),
	}
	if cfg.Stack.Repository != "" {
		opts := helm.RepositoryOptsArgs{Repo: pulumi.StringPtr(cfg.Stack.Repository)}
		stackArgs.RepositoryOpts = opts.ToRepositoryOptsPtrOutput()
	}
	if cfg.Stack.Version != "" {
		stackArgs.Version = pulumi.StringPtr(cfg.Stack.Version)
	}
	if cfg.Stack.ValuesFile != "" {
		stackArgs.ValueYamlFiles = pulumi.AssetOrArchiveArray{
			pulumi.NewFileAsset(cfg.Stack.ValuesFile),
		}
	}
	if cfg.Stack.EnableDependency {
		stackArgs.DependencyUpdate = pulumi.BoolPtr(true)
	}

	stackRelease, err := helm.NewRelease(ctx, pulumiResourceName(clusterKey+"-observability", 53), stackArgs, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
	if err != nil {
		return nil, err
	}

	metricsTimeout := cfg.MetricsServer.Timeout
	if metricsTimeout == 0 {
		metricsTimeout = defaultHelmTimeout
	}
	metricsValues := pulumi.Map{
		"fullnameOverride": pulumi.String(metricsFullname),
		"service": pulumi.Map{
			"type": pulumi.String("ClusterIP"),
			"port": pulumi.Int(443),
		},
	}

	metricsArgs := &helm.ReleaseArgs{
		Name:            pulumi.StringPtr(metricsReleaseName),
		Namespace:       pulumi.StringPtr(namespace),
		Chart:           pulumi.String(cfg.MetricsServer.ChartPath),
		Values:          metricsValues,
		Timeout:         pulumi.IntPtr(int(metricsTimeout.Seconds())),
		CreateNamespace: pulumi.BoolPtr(true),
	}
	if cfg.MetricsServer.Repository != "" {
		opts := helm.RepositoryOptsArgs{Repo: pulumi.StringPtr(cfg.MetricsServer.Repository)}
		metricsArgs.RepositoryOpts = opts.ToRepositoryOptsPtrOutput()
	}
	if cfg.MetricsServer.Version != "" {
		metricsArgs.Version = pulumi.StringPtr(cfg.MetricsServer.Version)
	}
	if cfg.MetricsServer.ValuesFile != "" {
		metricsArgs.ValueYamlFiles = pulumi.AssetOrArchiveArray{
			pulumi.NewFileAsset(cfg.MetricsServer.ValuesFile),
		}
	}
	if cfg.MetricsServer.EnableDependency {
		metricsArgs.DependencyUpdate = pulumi.BoolPtr(true)
	}

	metricsDeps := append([]pulumi.Resource{}, depends...)
	metricsDeps = append(metricsDeps, stackRelease)
	if _, err := helm.NewRelease(ctx, pulumiResourceName(clusterKey+"-metrics-server", 53), metricsArgs, pulumi.Provider(kubeProvider), pulumi.DependsOn(metricsDeps)); err != nil {
		return nil, err
	}

	outputs := pulumi.Map{
		"namespace":                pulumi.String(namespace),
		"prometheusService":        pulumi.String(stackFullname + "-prometheus"),
		"prometheusPort":           pulumi.Int(promPort),
		"alertmanagerService":      pulumi.String(stackFullname + "-alertmanager"),
		"alertmanagerPort":         pulumi.Int(alertPort),
		"alertmanagerConfigSecret": pulumi.String("alertmanager-" + stackFullname + "-alertmanager"),
		"metricsServerService":     pulumi.String(metricsFullname),
		"metricsServerPort":        pulumi.Int(443),
	}
	return outputs, nil
}

func findRepoRoot() string {
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
	if max <= 0 || len(name) <= max {
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
	return trimmed + "-" + suffix
}

func fileExists(path string) bool {
	if strings.TrimSpace(path) == "" {
		return false
	}
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}
