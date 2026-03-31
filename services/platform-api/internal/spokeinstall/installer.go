// Package spokeinstall installs the aegis-spoke Helm chart on remote Kubernetes
// clusters during the cluster import flow. It uses the helm.sh/helm/v3 Go library
// to programmatically run helm install/upgrade, reusing the same chart and values
// structure as the Pulumi provisioner.
package spokeinstall

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
	ProxyTLSCert           string // PEM-encoded spoke proxy cert (only if not using cert-manager)
	ProxyTLSKey            string // PEM-encoded spoke proxy key (only if not using cert-manager)
	HubCABundle            string // Base64-encoded hub CA for agent to verify hub TLS
	CertManagerEnabled     bool   // Use cert-manager + StepClusterIssuer for proxy TLS
	CertManagerIssuerName  string // e.g., "aegis-internal"
	CertManagerIssuerKind  string // e.g., "StepClusterIssuer"
	CertManagerIssuerGroup string // e.g., "certmanager.step.sm"
	VscodeREHInitImage     string // ECR image for VS Code REH init container (e.g., aegis/vscode-reh-init:1.113.0)
	Flavors                string
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

	// If the restConfig uses exec-based auth (e.g., aws eks get-token), the Helm
	// library's REST client wrapper doesn't always propagate the exec context correctly.
	// Pre-fetch a bearer token and convert to static token auth.
	if restConfig.ExecProvider != nil {
		log.Info("restConfig uses exec-based auth, pre-fetching bearer token")
		resolved, err := resolveBearerToken(restConfig)
		if err != nil {
			return fmt.Errorf("resolve bearer token from exec provider: %w", err)
		}
		restConfig = resolved
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
	if cfg.HubCABundle != "" {
		env["AEGIS_PLATFORM_CA_B64"] = cfg.HubCABundle
	}
	if cfg.VscodeREHInitImage != "" {
		env["AEGIS_VSCODE_REH_INIT_IMAGE"] = cfg.VscodeREHInitImage
	}
	if cfg.ProxyNodePort > 0 {
		env["AEGIS_SPOKE_PROXY_NODEPORT"] = fmt.Sprintf("%d", cfg.ProxyNodePort)
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
		if cfg.CertManagerEnabled && cfg.CertManagerIssuerName != "" {
			// Use cert-manager for automatic TLS cert issuance from step-ca
			proxy["tls"] = map[string]interface{}{
				"terminateAtIngress": false,
				"certManager": map[string]interface{}{
					"enabled": true,
					"issuerRef": map[string]interface{}{
						"name":  cfg.CertManagerIssuerName,
						"kind":  cfg.CertManagerIssuerKind,
						"group": cfg.CertManagerIssuerGroup,
					},
					"dnsNames": []interface{}{
						"*.nip.io",
						fmt.Sprintf("spoke-proxy-%s.nip.io", sanitize(cfg.ClusterID)),
					},
				},
			}
		} else if cfg.ProxyTLSCert != "" && cfg.ProxyTLSKey != "" {
			// Fallback: static TLS cert
			proxy["tls"] = map[string]interface{}{
				"cert": cfg.ProxyTLSCert,
				"key":  cfg.ProxyTLSKey,
			}
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

// resolveBearerToken takes a restConfig with exec-based auth and pre-fetches
// a bearer token, returning a new restConfig with static BearerToken auth.
// This works around Helm's REST client wrapper not propagating exec contexts.
func resolveBearerToken(original *rest.Config) (*rest.Config, error) {
	// Use the original config to make a simple API call — this triggers the
	// exec provider and caches the token in the transport.
	dc, err := discovery.NewDiscoveryClientForConfig(original)
	if err != nil {
		return nil, fmt.Errorf("create discovery client: %w", err)
	}
	// Make a lightweight API call to trigger token fetch
	_, err = dc.ServerVersion()
	if err != nil {
		return nil, fmt.Errorf("fetch server version (token exchange): %w", err)
	}

	// The exec provider caches the token. Extract it by creating a transport
	// and reading the cached credential.
	// Simpler approach: use the exec provider directly to get a token.
	if original.ExecProvider != nil {
		// Build the exec command and run it
		args := original.ExecProvider.Args
		cmd := original.ExecProvider.Command

		// Run the exec command to get credentials
		execCmd := execCommand(cmd, args, original.ExecProvider.Env)
		output, err := execCmd.Output()
		if err != nil {
			return nil, fmt.Errorf("exec credential provider %q: %w", cmd, err)
		}

		// Parse the ExecCredential response
		var cred struct {
			Status struct {
				Token string `json:"token"`
			} `json:"status"`
		}
		if err := json.Unmarshal(output, &cred); err != nil {
			return nil, fmt.Errorf("parse exec credential output: %w", err)
		}
		if cred.Status.Token == "" {
			return nil, fmt.Errorf("exec credential provider returned empty token")
		}

		// Create a new config with the static bearer token
		resolved := rest.CopyConfig(original)
		resolved.ExecProvider = nil
		resolved.BearerToken = cred.Status.Token
		return resolved, nil
	}

	return original, nil
}

// execCommand creates an exec.Cmd from the exec provider config.
func execCommand(command string, args []string, envVars []clientcmdapi.ExecEnvVar) *exec.Cmd {
	cmd := exec.Command(command, args...)
	// Inherit current environment
	cmd.Env = os.Environ()
	// Add exec provider env vars
	for _, e := range envVars {
		cmd.Env = append(cmd.Env, e.Name+"="+e.Value)
	}
	return cmd
}
