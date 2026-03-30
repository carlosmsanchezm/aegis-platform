// Package spokeinstall installs the aegis-spoke Helm chart on remote Kubernetes
// clusters during the cluster import flow. It uses the helm.sh/helm/v3 Go library
// to programmatically run helm install/upgrade, reusing the same chart and values
// structure as the Pulumi provisioner.
package spokeinstall

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/discovery/cached/memory"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// SpokeInstallConfig holds all parameters needed to install the aegis-spoke chart
// on a remote cluster. User-provided fields are ClusterID, Provider, Region.
// Everything else is auto-derived from the hub's own configuration.
type SpokeInstallConfig struct {
	// User-provided (from import form)
	ClusterID string
	Provider  string
	Region    string

	// Auto-derived from hub configuration
	Namespace        string // default: aegis-system
	HubGRPC          string
	HubGRPCInsecure  bool
	OIDCTokenURL     string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCAudience     string
	CABundleB64      string
	ProxyJWTSecret   string
	AgentImageRepo   string
	AgentImageTag    string
	ProxyImageRepo   string
	ProxyImageTag    string
	ProxyEnabled     bool
	ProxyNodePort    int32
	Flavors          string
}

// Install installs the aegis-spoke Helm chart on a remote cluster.
// The restConfig provides authenticated access to the target cluster
// (from a kubeconfig or assume-role credential).
func Install(ctx context.Context, log *zap.Logger, restConfig *rest.Config, chartPath string, cfg SpokeInstallConfig) error {
	if cfg.Namespace == "" {
		cfg.Namespace = "aegis-system"
	}
	if cfg.ProxyNodePort == 0 {
		cfg.ProxyNodePort = 31484
	}

	// Resolve chart path — look for the aegis-spoke chart
	absChartPath, err := resolveChartPath(chartPath)
	if err != nil {
		return fmt.Errorf("resolve chart path: %w", err)
	}

	// Load the chart
	chart, err := loader.Load(absChartPath)
	if err != nil {
		return fmt.Errorf("load chart %s: %w", absChartPath, err)
	}

	log.Info("installing aegis-spoke chart on remote cluster",
		zap.String("cluster_id", cfg.ClusterID),
		zap.String("chart", chart.Metadata.Name),
		zap.String("version", chart.Metadata.Version),
		zap.String("namespace", cfg.Namespace),
	)

	// Build Helm values matching charts/aegis-spoke/values.yaml structure
	values := buildValues(cfg)

	// Create Helm action configuration using the remote cluster's restConfig
	getter := &restClientGetter{restConfig: restConfig, namespace: cfg.Namespace}
	actionCfg := new(action.Configuration)
	if err := actionCfg.Init(getter, cfg.Namespace, "secret", func(format string, v ...interface{}) {
		log.Debug(fmt.Sprintf(format, v...), zap.String("component", "helm"))
	}); err != nil {
		return fmt.Errorf("init helm action config: %w", err)
	}

	// Check if release already exists (upgrade vs install)
	releaseName := fmt.Sprintf("aegis-spoke-%s", sanitize(cfg.ClusterID))
	if len(releaseName) > 53 {
		releaseName = releaseName[:53]
	}

	histClient := action.NewHistory(actionCfg)
	histClient.Max = 1
	_, err = histClient.Run(releaseName)
	if err != nil {
		// Release doesn't exist — install
		installClient := action.NewInstall(actionCfg)
		installClient.ReleaseName = releaseName
		installClient.Namespace = cfg.Namespace
		installClient.CreateNamespace = true
		installClient.Timeout = 5 * time.Minute
		installClient.Wait = true

		if _, err := installClient.RunWithContext(ctx, chart, values); err != nil {
			return fmt.Errorf("helm install: %w", err)
		}
		log.Info("aegis-spoke installed successfully", zap.String("cluster_id", cfg.ClusterID), zap.String("release", releaseName))
	} else {
		// Release exists — upgrade
		upgradeClient := action.NewUpgrade(actionCfg)
		upgradeClient.Namespace = cfg.Namespace
		upgradeClient.Timeout = 5 * time.Minute
		upgradeClient.Wait = true

		if _, err := upgradeClient.RunWithContext(ctx, releaseName, chart, values); err != nil {
			return fmt.Errorf("helm upgrade: %w", err)
		}
		log.Info("aegis-spoke upgraded successfully", zap.String("cluster_id", cfg.ClusterID), zap.String("release", releaseName))
	}

	return nil
}

// buildValues constructs the Helm values map matching the aegis-spoke chart structure.
// This mirrors the values built by the Pulumi runner in installSpokeHelmChart().
func buildValues(cfg SpokeInstallConfig) map[string]interface{} {
	env := map[string]interface{}{
		"AEGIS_CLUSTER_ID": cfg.ClusterID,
		"AEGIS_REGION":     cfg.Region,
		"AEGIS_PROVIDER":   cfg.Provider,
	}

	if cfg.HubGRPC != "" {
		env["AEGIS_CP_GRPC"] = cfg.HubGRPC
	}
	if cfg.HubGRPCInsecure {
		env["AEGIS_CP_GRPC_INSECURE"] = "true"
	} else {
		env["AEGIS_CP_GRPC_INSECURE"] = "false"
	}
	if cfg.OIDCTokenURL != "" {
		env["AEGIS_CP_OIDC_TOKEN_URL"] = cfg.OIDCTokenURL
	}
	if cfg.OIDCClientID != "" {
		env["AEGIS_CP_OIDC_CLIENT_ID"] = cfg.OIDCClientID
	}
	if cfg.OIDCClientSecret != "" {
		env["AEGIS_CP_OIDC_CLIENT_SECRET"] = cfg.OIDCClientSecret
	}
	if cfg.OIDCAudience != "" {
		env["AEGIS_CP_OIDC_AUDIENCE"] = cfg.OIDCAudience
	}
	if cfg.CABundleB64 != "" {
		env["AEGIS_PLATFORM_CA_B64"] = cfg.CABundleB64
	}
	if cfg.Flavors != "" {
		env["AEGIS_FLAVORS"] = cfg.Flavors
	}

	k8sAgent := map[string]interface{}{
		"env": env,
	}
	if cfg.AgentImageRepo != "" {
		k8sAgent["image"] = map[string]interface{}{
			"repository": cfg.AgentImageRepo,
			"tag":        cfg.AgentImageTag,
		}
	}

	proxy := map[string]interface{}{
		"enabled": cfg.ProxyEnabled,
	}
	if cfg.ProxyEnabled {
		proxy["service"] = map[string]interface{}{
			"type":     "NodePort",
			"nodePort": cfg.ProxyNodePort,
			"port":     443,
		}
		proxy["ingress"] = map[string]interface{}{
			"enabled": false,
		}
		if cfg.ProxyImageRepo != "" {
			proxy["image"] = map[string]interface{}{
				"repository": cfg.ProxyImageRepo,
				"tag":        cfg.ProxyImageTag,
			}
		}
		if cfg.ProxyJWTSecret != "" {
			proxy["jwtSecret"] = cfg.ProxyJWTSecret
		}
	}

	return map[string]interface{}{
		"k8sAgent": k8sAgent,
		"proxy":    proxy,
	}
}

// resolveChartPath finds the aegis-spoke chart directory.
func resolveChartPath(chartPath string) (string, error) {
	if chartPath != "" {
		abs, err := filepath.Abs(chartPath)
		if err != nil {
			return "", err
		}
		if _, err := os.Stat(abs); err == nil {
			return abs, nil
		}
	}

	// Try common locations
	candidates := []string{
		"charts/aegis-spoke",
		"/home/aegis/charts/aegis-spoke",   // container path (platform-api Dockerfile)
		"/app/charts/aegis-spoke",          // alternative container path
		"/workspace/charts/aegis-spoke",    // CI path
		"../../charts/aegis-spoke",         // relative from services/platform-api
	}
	for _, c := range candidates {
		abs, err := filepath.Abs(c)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs, nil
		}
	}

	return "", fmt.Errorf("aegis-spoke chart not found; set AEGIS_SPOKE_CHART_PATH or place chart in charts/aegis-spoke/")
}

func sanitize(s string) string {
	r := strings.NewReplacer("/", "-", ".", "-", "_", "-", " ", "-")
	return strings.ToLower(r.Replace(s))
}

// restClientGetter implements genericclioptions.RESTClientGetter for a rest.Config.
// This allows Helm to use a programmatic kubeconfig instead of a file-based one.
type restClientGetter struct {
	restConfig *rest.Config
	namespace  string
}

func (r *restClientGetter) ToRESTConfig() (*rest.Config, error) {
	return r.restConfig, nil
}

func (r *restClientGetter) ToDiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	config, err := r.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	return memory.NewMemCacheClient(dc), nil
}

func (r *restClientGetter) ToRESTMapper() (meta.RESTMapper, error) {
	config, err := r.ToRESTConfig()
	if err != nil {
		return nil, err
	}
	dc, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return nil, err
	}
	return restmapper.NewDeferredDiscoveryRESTMapper(memory.NewMemCacheClient(dc)), nil
}

func (r *restClientGetter) ToRawKubeConfigLoader() clientcmd.ClientConfig {
	return clientcmd.NewDefaultClientConfig(clientcmdapi.Config{}, &clientcmd.ConfigOverrides{
		Context: clientcmdapi.Context{
			Namespace: r.namespace,
		},
	})
}

// Ensure interface compliance
var _ genericclioptions.RESTClientGetter = (*restClientGetter)(nil)
// Silence unused import warning for cli package
var _ = cli.New
