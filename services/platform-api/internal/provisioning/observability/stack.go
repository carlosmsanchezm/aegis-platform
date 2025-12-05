package observability

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

const (
	envObservabilityChart           = "AEGIS_OBSERVABILITY_CHART"
	envObservabilityChartVersion    = "AEGIS_OBSERVABILITY_CHART_VERSION"
	envObservabilityRepo            = "AEGIS_OBSERVABILITY_REPO"
	envObservabilityValuesFile      = "AEGIS_OBSERVABILITY_VALUES_FILE"
	envObservabilityEnabled         = "AEGIS_OBSERVABILITY_ENABLED"
	envMetricsServerChart           = "AEGIS_METRICS_SERVER_CHART"
	envMetricsServerVersion         = "AEGIS_METRICS_SERVER_CHART_VERSION"
	envMetricsServerRepo            = "AEGIS_METRICS_SERVER_REPO"
	envMetricsServerValuesFile      = "AEGIS_METRICS_SERVER_VALUES_FILE"
	envLoggingLokiChart             = "AEGIS_LOGGING_LOKI_CHART"
	envLoggingLokiChartVersion      = "AEGIS_LOGGING_LOKI_CHART_VERSION"
	envLoggingLokiRepo              = "AEGIS_LOGGING_LOKI_REPO"
	envLoggingLokiValuesFile        = "AEGIS_LOGGING_LOKI_VALUES_FILE"
	envLoggingFluentBitChart        = "AEGIS_LOGGING_FLUENT_BIT_CHART"
	envLoggingFluentBitVersion      = "AEGIS_LOGGING_FLUENT_BIT_CHART_VERSION"
	envLoggingFluentBitRepo         = "AEGIS_LOGGING_FLUENT_BIT_REPO"
	envLoggingFluentBitValuesFile   = "AEGIS_LOGGING_FLUENT_BIT_VALUES_FILE"
	envTracingTempoChart            = "AEGIS_TRACING_TEMPO_CHART"
	envTracingTempoChartVersion     = "AEGIS_TRACING_TEMPO_CHART_VERSION"
	envTracingTempoRepo             = "AEGIS_TRACING_TEMPO_REPO"
	envTracingTempoValuesFile       = "AEGIS_TRACING_TEMPO_VALUES_FILE"
	envTracingCollectorChart        = "AEGIS_TRACING_COLLECTOR_CHART"
	envTracingCollectorChartVersion = "AEGIS_TRACING_COLLECTOR_CHART_VERSION"
	envTracingCollectorRepo         = "AEGIS_TRACING_COLLECTOR_REPO"
	envTracingCollectorValuesFile   = "AEGIS_TRACING_COLLECTOR_VALUES_FILE"
	defaultObservabilityNamespace   = "aegis-observability"
	defaultObservabilityRelease     = "aegis-obsv"
	defaultObservabilityBaseName    = "aegis-obsv"
	defaultMetricsRelease           = "aegis-metrics"
	defaultLoggingNamespace         = "aegis-logging"
	defaultLoggingBaseName          = "aegis-logging"
	defaultLoggingLokiRelease       = "aegis-loki"
	defaultLoggingFluentBitRelease  = "aegis-fluentbit"
	defaultTracingNamespace         = "aegis-tracing"
	defaultTracingBaseName          = "aegis-tracing"
	defaultTempoRelease             = "aegis-tempo"
	defaultCollectorRelease         = "aegis-otel"
	defaultLokiPort                 = 3100
	defaultTempoPort                = 3200
	defaultTempoGrpcPort            = 4317
	defaultTempoHttpPort            = 4318
	defaultCollectorGrpcPort        = 4317
	defaultCollectorHttpPort        = 4318
	defaultLokiRetention            = "168h"
	defaultTempoRetention           = "48h"
	defaultPrometheusPort           = 9090
	defaultAlertmanagerPort         = 9093
	maxObservabilityNameLength      = 40
	maxLoggingNameLength            = 40
	maxTracingNameLength            = 40
	defaultHelmTimeout              = 15 * time.Minute
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
	Logging          LoggingConfig
	Tracing          TracingConfig
	Namespace        string
	BaseName         string
	PrometheusPort   int
	AlertmanagerPort int
	Enable           bool
}

// LoggingConfig captures the Loki + Fluent Bit settings.
type LoggingConfig struct {
	Loki            HelmConfig
	FluentBit       HelmConfig
	Namespace       string
	BaseName        string
	LokiPort        int
	RetentionPeriod string
	Enable          bool
}

// TracingConfig captures Tempo + OTEL Collector settings.
type TracingConfig struct {
	Tempo             HelmConfig
	Collector         HelmConfig
	Namespace         string
	BaseName          string
	TempoPort         int
	TempoGrpcPort     int
	TempoHttpPort     int
	CollectorGrpcPort int
	CollectorHttpPort int
	RetentionPeriod   string
	Enable            bool
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

	lokiChart := strings.TrimSpace(os.Getenv(envLoggingLokiChart))
	if lokiChart == "" {
		lokiChart = "loki"
	}
	lokiRepo := strings.TrimSpace(os.Getenv(envLoggingLokiRepo))
	if lokiRepo == "" {
		lokiRepo = "https://grafana.github.io/helm-charts"
	}
	lokiValues := strings.TrimSpace(os.Getenv(envLoggingLokiValuesFile))
	if lokiValues == "" {
		lokiValues = filepath.Join(repoRoot, "services", "platform-api", "config", "observability", "loki-values.yaml")
	}

	fluentChart := strings.TrimSpace(os.Getenv(envLoggingFluentBitChart))
	if fluentChart == "" {
		fluentChart = "fluent-bit"
	}
	fluentRepo := strings.TrimSpace(os.Getenv(envLoggingFluentBitRepo))
	if fluentRepo == "" {
		fluentRepo = "https://fluent.github.io/helm-charts"
	}
	fluentValues := strings.TrimSpace(os.Getenv(envLoggingFluentBitValuesFile))
	if fluentValues == "" {
		fluentValues = filepath.Join(repoRoot, "services", "platform-api", "config", "observability", "fluent-bit-values.yaml")
	}

	tempoChart := strings.TrimSpace(os.Getenv(envTracingTempoChart))
	if tempoChart == "" {
		tempoChart = "tempo"
	}
	tempoRepo := strings.TrimSpace(os.Getenv(envTracingTempoRepo))
	if tempoRepo == "" {
		tempoRepo = "https://grafana.github.io/helm-charts"
	}
	tempoValues := strings.TrimSpace(os.Getenv(envTracingTempoValuesFile))
	if tempoValues == "" {
		tempoValues = filepath.Join(repoRoot, "services", "platform-api", "config", "observability", "tempo-values.yaml")
	}

	collectorChart := strings.TrimSpace(os.Getenv(envTracingCollectorChart))
	if collectorChart == "" {
		collectorChart = "opentelemetry-collector"
	}
	collectorRepo := strings.TrimSpace(os.Getenv(envTracingCollectorRepo))
	if collectorRepo == "" {
		collectorRepo = "https://open-telemetry.github.io/opentelemetry-helm-charts"
	}
	collectorValues := strings.TrimSpace(os.Getenv(envTracingCollectorValuesFile))
	if collectorValues == "" {
		collectorValues = filepath.Join(repoRoot, "services", "platform-api", "config", "observability", "otel-collector-values.yaml")
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
		Logging: LoggingConfig{
			Namespace:       defaultLoggingNamespace,
			BaseName:        defaultLoggingBaseName,
			LokiPort:        defaultLokiPort,
			RetentionPeriod: defaultLokiRetention,
			Enable:          true,
			Loki: HelmConfig{
				ChartPath:        lokiChart,
				Repository:       lokiRepo,
				Version:          strings.TrimSpace(os.Getenv(envLoggingLokiChartVersion)),
				ValuesFile:       lokiValues,
				Namespace:        defaultLoggingNamespace,
				ReleaseName:      defaultLoggingLokiRelease,
				Timeout:          timeout,
				EnableDependency: true,
			},
			FluentBit: HelmConfig{
				ChartPath:        fluentChart,
				Repository:       fluentRepo,
				Version:          strings.TrimSpace(os.Getenv(envLoggingFluentBitVersion)),
				ValuesFile:       fluentValues,
				Namespace:        defaultLoggingNamespace,
				ReleaseName:      defaultLoggingFluentBitRelease,
				Timeout:          timeout,
				EnableDependency: true,
			},
		},
		Tracing: TracingConfig{
			Namespace:         defaultTracingNamespace,
			BaseName:          defaultTracingBaseName,
			TempoPort:         defaultTempoPort,
			TempoGrpcPort:     defaultTempoGrpcPort,
			TempoHttpPort:     defaultTempoHttpPort,
			CollectorGrpcPort: defaultCollectorGrpcPort,
			CollectorHttpPort: defaultCollectorHttpPort,
			RetentionPeriod:   defaultTempoRetention,
			Enable:            true,
			Tempo: HelmConfig{
				ChartPath:        tempoChart,
				Repository:       tempoRepo,
				Version:          strings.TrimSpace(os.Getenv(envTracingTempoChartVersion)),
				ValuesFile:       tempoValues,
				Namespace:        defaultTracingNamespace,
				ReleaseName:      defaultTempoRelease,
				Timeout:          timeout,
				EnableDependency: true,
			},
			Collector: HelmConfig{
				ChartPath:        collectorChart,
				Repository:       collectorRepo,
				Version:          strings.TrimSpace(os.Getenv(envTracingCollectorChartVersion)),
				ValuesFile:       collectorValues,
				Namespace:        defaultTracingNamespace,
				ReleaseName:      defaultCollectorRelease,
				Timeout:          timeout,
				EnableDependency: true,
			},
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

	loggingOutputs, err := installLogging(ctx, clusterKey, kubeProvider, cfg, depends)
	if err != nil {
		return nil, err
	}

	tracingOutputs, err := installTracing(ctx, clusterKey, kubeProvider, cfg, depends)
	if err != nil {
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
	for k, v := range loggingOutputs {
		outputs[k] = v
	}
	for k, v := range tracingOutputs {
		outputs[k] = v
	}
	return outputs, nil
}

func installTracing(ctx *pulumi.Context, clusterKey string, kubeProvider *kubernetes.Provider, cfg Config, depends []pulumi.Resource) (pulumi.Map, error) {
	if cfg.Tracing.Enable && kubeProvider == nil {
		return nil, fmt.Errorf("kubernetes provider is required for tracing installs")
	}
	if !cfg.Tracing.Enable {
		return pulumi.Map{}, nil
	}

	namespace := strings.TrimSpace(cfg.Tracing.Namespace)
	if namespace == "" {
		namespace = defaultTracingNamespace
	}
	baseName := strings.TrimSpace(cfg.Tracing.BaseName)
	if baseName == "" {
		baseName = defaultTracingBaseName
	}

	tempoPort := cfg.Tracing.TempoPort
	if tempoPort == 0 {
		tempoPort = defaultTempoPort
	}
	tempoGrpc := cfg.Tracing.TempoGrpcPort
	if tempoGrpc == 0 {
		tempoGrpc = defaultTempoGrpcPort
	}
	tempoHttp := cfg.Tracing.TempoHttpPort
	if tempoHttp == 0 {
		tempoHttp = defaultTempoHttpPort
	}
	collectorGrpc := cfg.Tracing.CollectorGrpcPort
	if collectorGrpc == 0 {
		collectorGrpc = defaultCollectorGrpcPort
	}
	collectorHttp := cfg.Tracing.CollectorHttpPort
	if collectorHttp == 0 {
		collectorHttp = defaultCollectorHttpPort
	}
	retention := strings.TrimSpace(cfg.Tracing.RetentionPeriod)
	if retention == "" {
		retention = defaultTempoRetention
	}

	tempoReleaseName := pulumiResourceName(cfg.Tracing.Tempo.ReleaseName+"-"+clusterKey, 53)
	tempoFullname := pulumiResourceName(baseName+"-"+clusterKey+"-tempo", maxTracingNameLength)
	tempoTimeout := cfg.Tracing.Tempo.Timeout
	if tempoTimeout == 0 {
		tempoTimeout = defaultHelmTimeout
	}

	tempoValues := pulumi.Map{
		"fullnameOverride": pulumi.String(tempoFullname),
		"replicas":         pulumi.Int(1),
		"tempo": pulumi.Map{
			"memBallastSizeMbs":   pulumi.Int(0),
			"multitenancyEnabled": pulumi.Bool(false),
			"reportingEnabled":    pulumi.Bool(false),
			"retention":           pulumi.String(retention),
			"receivers": pulumi.Map{
				"otlp": pulumi.Map{
					"protocols": pulumi.Map{
						"grpc": pulumi.Map{
							"endpoint": pulumi.Sprintf("0.0.0.0:%d", tempoGrpc),
						},
						"http": pulumi.Map{
							"endpoint": pulumi.Sprintf("0.0.0.0:%d", tempoHttp),
						},
					},
				},
			},
			"server": pulumi.Map{
				"http_listen_port": pulumi.Int(tempoPort),
			},
		},
		"service": pulumi.Map{
			"type": pulumi.String("ClusterIP"),
		},
		"persistence": pulumi.Map{
			"enabled": pulumi.Bool(false),
		},
		"tempoQuery": pulumi.Map{
			"enabled": pulumi.Bool(false),
		},
		"serviceMonitor": pulumi.Map{
			"enabled": pulumi.Bool(false),
		},
	}

	tempoArgs := &helm.ReleaseArgs{
		Name:            pulumi.StringPtr(tempoReleaseName),
		Namespace:       pulumi.StringPtr(namespace),
		Chart:           pulumi.String(cfg.Tracing.Tempo.ChartPath),
		Values:          tempoValues,
		Timeout:         pulumi.IntPtr(int(tempoTimeout.Seconds())),
		CreateNamespace: pulumi.BoolPtr(true),
	}
	if cfg.Tracing.Tempo.Repository != "" {
		opts := helm.RepositoryOptsArgs{Repo: pulumi.StringPtr(cfg.Tracing.Tempo.Repository)}
		tempoArgs.RepositoryOpts = opts.ToRepositoryOptsPtrOutput()
	}
	if cfg.Tracing.Tempo.Version != "" {
		tempoArgs.Version = pulumi.StringPtr(cfg.Tracing.Tempo.Version)
	}
	if cfg.Tracing.Tempo.ValuesFile != "" {
		tempoArgs.ValueYamlFiles = pulumi.AssetOrArchiveArray{
			pulumi.NewFileAsset(cfg.Tracing.Tempo.ValuesFile),
		}
	}
	if cfg.Tracing.Tempo.EnableDependency {
		tempoArgs.DependencyUpdate = pulumi.BoolPtr(true)
	}

	tempoRelease, err := helm.NewRelease(ctx, pulumiResourceName(clusterKey+"-tempo", 53), tempoArgs, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
	if err != nil {
		return nil, err
	}

	collectorReleaseName := pulumiResourceName(cfg.Tracing.Collector.ReleaseName+"-"+clusterKey, 53)
	collectorFullname := pulumiResourceName(baseName+"-"+clusterKey+"-otel", maxTracingNameLength)
	collectorTimeout := cfg.Tracing.Collector.Timeout
	if collectorTimeout == 0 {
		collectorTimeout = defaultHelmTimeout
	}

	collectorValues := pulumi.Map{
		"mode":             pulumi.String("deployment"),
		"fullnameOverride": pulumi.String(collectorFullname),
		"service": pulumi.Map{
			"type": pulumi.String("ClusterIP"),
		},
		"ports": pulumi.Map{
			"otlp": pulumi.Map{
				"enabled":       pulumi.Bool(true),
				"servicePort":   pulumi.Int(collectorGrpc),
				"containerPort": pulumi.Int(collectorGrpc),
				"protocol":      pulumi.String("TCP"),
				"appProtocol":   pulumi.String("grpc"),
			},
			"otlp-http": pulumi.Map{
				"enabled":       pulumi.Bool(true),
				"servicePort":   pulumi.Int(collectorHttp),
				"containerPort": pulumi.Int(collectorHttp),
				"protocol":      pulumi.String("TCP"),
			},
			"jaeger-compact": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
			"jaeger-thrift": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
			"jaeger-grpc": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
			"zipkin": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
			"metrics": pulumi.Map{
				"enabled":       pulumi.Bool(true),
				"servicePort":   pulumi.Int(8888),
				"containerPort": pulumi.Int(8888),
				"protocol":      pulumi.String("TCP"),
			},
		},
		"config": pulumi.Map{
			"exporters": pulumi.Map{
				"otlp": pulumi.Map{
					"endpoint": pulumi.Sprintf("%s:%d", tempoFullname, tempoGrpc),
					"tls": pulumi.Map{
						"insecure": pulumi.Bool(true),
					},
				},
			},
			"extensions": pulumi.Map{
				"health_check": pulumi.Map{
					"endpoint": pulumi.Sprintf("${env:MY_POD_IP}:%d", 13133),
				},
			},
			"processors": pulumi.Map{
				"batch": pulumi.Map{},
				"memory_limiter": pulumi.Map{
					"check_interval":         pulumi.String("5s"),
					"limit_percentage":       pulumi.Int(80),
					"spike_limit_percentage": pulumi.Int(25),
				},
			},
			"receivers": pulumi.Map{
				"otlp": pulumi.Map{
					"protocols": pulumi.Map{
						"grpc": pulumi.Map{
							"endpoint": pulumi.Sprintf("${env:MY_POD_IP}:%d", collectorGrpc),
						},
						"http": pulumi.Map{
							"endpoint": pulumi.Sprintf("${env:MY_POD_IP}:%d", collectorHttp),
						},
					},
				},
			},
			"service": pulumi.Map{
				"extensions": pulumi.Array{
					pulumi.String("health_check"),
				},
				"pipelines": pulumi.Map{
					"traces": pulumi.Map{
						"receivers": pulumi.Array{
							pulumi.String("otlp"),
						},
						"processors": pulumi.Array{
							pulumi.String("memory_limiter"),
							pulumi.String("batch"),
						},
						"exporters": pulumi.Array{
							pulumi.String("otlp"),
						},
					},
				},
				"telemetry": pulumi.Map{
					"metrics": pulumi.Map{
						"readers": pulumi.Array{
							pulumi.Map{
								"pull": pulumi.Map{
									"exporter": pulumi.Map{
										"prometheus": pulumi.Map{
											"host": pulumi.String("${env:MY_POD_IP}"),
											"port": pulumi.Int(8888),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	collectorArgs := &helm.ReleaseArgs{
		Name:            pulumi.StringPtr(collectorReleaseName),
		Namespace:       pulumi.StringPtr(namespace),
		Chart:           pulumi.String(cfg.Tracing.Collector.ChartPath),
		Values:          collectorValues,
		Timeout:         pulumi.IntPtr(int(collectorTimeout.Seconds())),
		CreateNamespace: pulumi.BoolPtr(true),
	}
	if cfg.Tracing.Collector.Repository != "" {
		opts := helm.RepositoryOptsArgs{Repo: pulumi.StringPtr(cfg.Tracing.Collector.Repository)}
		collectorArgs.RepositoryOpts = opts.ToRepositoryOptsPtrOutput()
	}
	if cfg.Tracing.Collector.Version != "" {
		collectorArgs.Version = pulumi.StringPtr(cfg.Tracing.Collector.Version)
	}
	if cfg.Tracing.Collector.ValuesFile != "" {
		collectorArgs.ValueYamlFiles = pulumi.AssetOrArchiveArray{
			pulumi.NewFileAsset(cfg.Tracing.Collector.ValuesFile),
		}
	}
	if cfg.Tracing.Collector.EnableDependency {
		collectorArgs.DependencyUpdate = pulumi.BoolPtr(true)
	}

	collectorDeps := append([]pulumi.Resource{}, depends...)
	collectorDeps = append(collectorDeps, tempoRelease)
	if _, err := helm.NewRelease(ctx, pulumiResourceName(clusterKey+"-otel-collector", 53), collectorArgs, pulumi.Provider(kubeProvider), pulumi.DependsOn(collectorDeps)); err != nil {
		return nil, err
	}

	return pulumi.Map{
		"tracingNamespace": pulumi.String(namespace),
		"tempoService":     pulumi.String(tempoFullname),
		"tempoPort":        pulumi.Int(tempoPort),
		"otelService":      pulumi.String(collectorFullname),
		"otelGrpcPort":     pulumi.Int(collectorGrpc),
		"otelHttpPort":     pulumi.Int(collectorHttp),
	}, nil
}

func installLogging(ctx *pulumi.Context, clusterKey string, kubeProvider *kubernetes.Provider, cfg Config, depends []pulumi.Resource) (pulumi.Map, error) {
	if cfg.Logging.Enable && kubeProvider == nil {
		return nil, fmt.Errorf("kubernetes provider is required for logging installs")
	}
	if !cfg.Logging.Enable {
		return pulumi.Map{}, nil
	}

	namespace := strings.TrimSpace(cfg.Logging.Namespace)
	if namespace == "" {
		namespace = defaultLoggingNamespace
	}
	baseName := strings.TrimSpace(cfg.Logging.BaseName)
	if baseName == "" {
		baseName = defaultLoggingBaseName
	}

	lokiPort := cfg.Logging.LokiPort
	if lokiPort == 0 {
		lokiPort = defaultLokiPort
	}
	retention := strings.TrimSpace(cfg.Logging.RetentionPeriod)
	if retention == "" {
		retention = defaultLokiRetention
	}

	lokiReleaseName := pulumiResourceName(cfg.Logging.Loki.ReleaseName+"-"+clusterKey, 53)
	lokiFullname := pulumiResourceName(baseName+"-"+clusterKey+"-loki", maxLoggingNameLength)

	lokiTimeout := cfg.Logging.Loki.Timeout
	if lokiTimeout == 0 {
		lokiTimeout = defaultHelmTimeout
	}

	lokiValues := pulumi.Map{
		"deploymentMode":   pulumi.String("SingleBinary"),
		"fullnameOverride": pulumi.String(lokiFullname),
		"loki": pulumi.Map{
			"auth_enabled": pulumi.Bool(false),
			"analytics": pulumi.Map{
				"reporting_enabled": pulumi.Bool(false),
			},
			"server": pulumi.Map{
				"http_listen_port": pulumi.Int(lokiPort),
			},
			"limits_config": pulumi.Map{
				"retention_period": pulumi.String(retention),
			},
			"commonConfig": pulumi.Map{
				"replication_factor": pulumi.Int(1),
			},
			"storage": pulumi.Map{
				"type": pulumi.String("filesystem"),
				"filesystem": pulumi.Map{
					"chunks_directory": pulumi.String("/var/loki/chunks"),
					"rules_directory":  pulumi.String("/var/loki/rules"),
				},
			},
		},
		"singleBinary": pulumi.Map{
			"replicas": pulumi.Int(1),
			"persistence": pulumi.Map{
				"enabled": pulumi.Bool(false),
			},
			"resources": pulumi.Map{
				"requests": pulumi.Map{
					"cpu":    pulumi.String("200m"),
					"memory": pulumi.String("512Mi"),
				},
				"limits": pulumi.Map{
					"cpu":    pulumi.String("500m"),
					"memory": pulumi.String("1Gi"),
				},
			},
			"service": pulumi.Map{
				"type": pulumi.String("ClusterIP"),
			},
		},
		"gateway": pulumi.Map{
			"enabled": pulumi.Bool(false),
		},
	}

	lokiArgs := &helm.ReleaseArgs{
		Name:            pulumi.StringPtr(lokiReleaseName),
		Namespace:       pulumi.StringPtr(namespace),
		Chart:           pulumi.String(cfg.Logging.Loki.ChartPath),
		Values:          lokiValues,
		Timeout:         pulumi.IntPtr(int(lokiTimeout.Seconds())),
		CreateNamespace: pulumi.BoolPtr(true),
	}
	if cfg.Logging.Loki.Repository != "" {
		opts := helm.RepositoryOptsArgs{Repo: pulumi.StringPtr(cfg.Logging.Loki.Repository)}
		lokiArgs.RepositoryOpts = opts.ToRepositoryOptsPtrOutput()
	}
	if cfg.Logging.Loki.Version != "" {
		lokiArgs.Version = pulumi.StringPtr(cfg.Logging.Loki.Version)
	}
	if cfg.Logging.Loki.ValuesFile != "" {
		lokiArgs.ValueYamlFiles = pulumi.AssetOrArchiveArray{
			pulumi.NewFileAsset(cfg.Logging.Loki.ValuesFile),
		}
	}
	if cfg.Logging.Loki.EnableDependency {
		lokiArgs.DependencyUpdate = pulumi.BoolPtr(true)
	}

	lokiRelease, err := helm.NewRelease(ctx, pulumiResourceName(clusterKey+"-loki", 53), lokiArgs, pulumi.Provider(kubeProvider), pulumi.DependsOn(depends))
	if err != nil {
		return nil, err
	}

	fluentReleaseName := pulumiResourceName(cfg.Logging.FluentBit.ReleaseName+"-"+clusterKey, 53)
	fluentFullname := pulumiResourceName(baseName+"-"+clusterKey+"-fluentbit", maxLoggingNameLength)
	fluentTimeout := cfg.Logging.FluentBit.Timeout
	if fluentTimeout == 0 {
		fluentTimeout = defaultHelmTimeout
	}

	fluentServiceConfig := `[SERVICE]
    Daemon Off
    Flush 1
    Log_Level info
    Parsers_File /fluent-bit/etc/parsers.conf
    Parsers_File /fluent-bit/etc/conf/custom_parsers.conf
    HTTP_Server On
    HTTP_Listen 0.0.0.0
    HTTP_Port 2020
    Health_Check On
`
	fluentInputs := `[INPUT]
    Name tail
    Path /var/log/containers/*.log
    multiline.parser docker, cri
    Tag kube.*
    Mem_Buf_Limit 5MB
    Skip_Long_Lines On

[INPUT]
    Name kubernetes_events
    Tag kube.events
    kube_url https://kubernetes.default.svc:443
    kube_ca_file /var/run/secrets/kubernetes.io/serviceaccount/ca.crt
    kube_token_file /var/run/secrets/kubernetes.io/serviceaccount/token
    tls.verify On
`
	fluentFilters := fmt.Sprintf(`[FILTER]
    Name kubernetes
    Match kube.*
    Merge_Log On
    Keep_Log Off
    K8S-Logging.Parser On
    K8S-Logging.Exclude Off
    Kube_Tag_Prefix kube.var.log.containers.

[FILTER]
    Name modify
    Match kube.*
    Add cluster_id %s

[FILTER]
    Name modify
    Match kube.events
    Add cluster_id %s
`, clusterKey, clusterKey)

	fluentOutputs := fmt.Sprintf(`[OUTPUT]
    Name loki
    Match kube.*
    Host %s
    Port %d
    labels cluster=%s,namespace=$kubernetes['namespace_name'],pod=$kubernetes['pod_name'],container=$kubernetes['container_name'],app=$kubernetes['labels']['app']
    label_keys $kubernetes['namespace_name'],$kubernetes['pod_name'],$kubernetes['container_name'],$kubernetes['labels']['app']
    remove_keys kubernetes,stream
    line_format json
    auto_kubernetes_labels off

[OUTPUT]
    Name loki
    Match kube.events
    Host %s
    Port %d
    labels cluster=%s,namespace=$kubernetes['namespace_name'],event_reason=$reason,event_type=$type
    label_keys $kubernetes['namespace_name'],$reason,$type
    remove_keys kubernetes,stream
    line_format json
    auto_kubernetes_labels off
`, lokiFullname, lokiPort, clusterKey, lokiFullname, lokiPort, clusterKey)

	fluentValues := pulumi.Map{
		"fullnameOverride": pulumi.String(fluentFullname),
		"rbac": pulumi.Map{
			"create":       pulumi.Bool(true),
			"nodeAccess":   pulumi.Bool(true),
			"eventsAccess": pulumi.Bool(true),
		},
		"service": pulumi.Map{
			"type": pulumi.String("ClusterIP"),
		},
		"serviceAccount": pulumi.Map{
			"automountServiceAccountToken": pulumi.BoolPtr(true),
		},
		"env": pulumi.Array{
			pulumi.Map{
				"name":  pulumi.String("CLUSTER_ID"),
				"value": pulumi.String(clusterKey),
			},
			pulumi.Map{
				"name":  pulumi.String("LOKI_HOST"),
				"value": pulumi.String(lokiFullname),
			},
			pulumi.Map{
				"name":  pulumi.String("LOKI_PORT"),
				"value": pulumi.Sprintf("%d", lokiPort),
			},
		},
		"config": pulumi.Map{
			"service": pulumi.String(fluentServiceConfig),
			"inputs":  pulumi.String(fluentInputs),
			"filters": pulumi.String(fluentFilters),
			"outputs": pulumi.String(fluentOutputs),
		},
		"resources": pulumi.Map{
			"requests": pulumi.Map{
				"cpu":    pulumi.String("100m"),
				"memory": pulumi.String("128Mi"),
			},
			"limits": pulumi.Map{
				"cpu":    pulumi.String("400m"),
				"memory": pulumi.String("256Mi"),
			},
		},
		"testFramework": pulumi.Map{
			"enabled": pulumi.Bool(false),
		},
	}

	fluentArgs := &helm.ReleaseArgs{
		Name:            pulumi.StringPtr(fluentReleaseName),
		Namespace:       pulumi.StringPtr(namespace),
		Chart:           pulumi.String(cfg.Logging.FluentBit.ChartPath),
		Values:          fluentValues,
		Timeout:         pulumi.IntPtr(int(fluentTimeout.Seconds())),
		CreateNamespace: pulumi.BoolPtr(true),
	}
	if cfg.Logging.FluentBit.Repository != "" {
		opts := helm.RepositoryOptsArgs{Repo: pulumi.StringPtr(cfg.Logging.FluentBit.Repository)}
		fluentArgs.RepositoryOpts = opts.ToRepositoryOptsPtrOutput()
	}
	if cfg.Logging.FluentBit.Version != "" {
		fluentArgs.Version = pulumi.StringPtr(cfg.Logging.FluentBit.Version)
	}
	if cfg.Logging.FluentBit.ValuesFile != "" {
		fluentArgs.ValueYamlFiles = pulumi.AssetOrArchiveArray{
			pulumi.NewFileAsset(cfg.Logging.FluentBit.ValuesFile),
		}
	}
	if cfg.Logging.FluentBit.EnableDependency {
		fluentArgs.DependencyUpdate = pulumi.BoolPtr(true)
	}

	fluentDeps := append([]pulumi.Resource{}, depends...)
	fluentDeps = append(fluentDeps, lokiRelease)
	if _, err := helm.NewRelease(ctx, pulumiResourceName(clusterKey+"-fluentbit", 53), fluentArgs, pulumi.Provider(kubeProvider), pulumi.DependsOn(fluentDeps)); err != nil {
		return nil, err
	}

	return pulumi.Map{
		"lokiNamespace":  pulumi.String(namespace),
		"lokiService":    pulumi.String(lokiFullname),
		"lokiPort":       pulumi.Int(lokiPort),
		"lokiAuthSecret": pulumi.String(""),
	}, nil
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
