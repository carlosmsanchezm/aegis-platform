package kubeclients

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	aegisv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/kubeclients/ekstoken"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const defaultClientTTL = 10 * time.Minute

// ClusterAuthInfo contains the information needed to generate programmatic
// EKS tokens for a cluster, bypassing exec-based kubeconfig auth.
type ClusterAuthInfo struct {
	Endpoint    string // EKS API server URL
	CAData      string // base64-encoded cluster CA certificate
	Region      string
	ClusterName string
	RoleARN     string // optional — for cross-account access
	ExternalID  string // optional — paired with RoleARN
}

// ClusterAuthProvider resolves authentication info for a cluster,
// combining cluster metadata (endpoint, CA, region) with project
// credentials (RoleARN, ExternalID).
type ClusterAuthProvider interface {
	GetClusterAuthInfo(clusterID string) (*ClusterAuthInfo, error)
}

type cachedClient struct {
	client    client.Client
	createdAt time.Time
}

type Manager struct {
	dir             string
	scheme          *runtime.Scheme
	mu              sync.RWMutex
	clients         map[string]*cachedClient
	clientTTL       time.Duration
	infraClient     client.Client // optional, for K8s Secret API access
	secretName      string
	secretNamespace string
	authProvider    ClusterAuthProvider // optional, for programmatic EKS token auth
}

func New(dir string) *Manager {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = aegisv1alpha1.AddToScheme(scheme)

	workspaceGV := schema.GroupVersion{Group: "aegis.yourorg.dev", Version: "v1alpha2"}
	scheme.AddKnownTypeWithName(workspaceGV.WithKind("Workspace"), &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(workspaceGV.WithKind("WorkspaceList"), &unstructured.UnstructuredList{})

	ttl := defaultClientTTL
	if v, err := strconv.Atoi(os.Getenv("AEGIS_KUBECLIENT_TTL_SECONDS")); err == nil && v > 0 {
		ttl = time.Duration(v) * time.Second
	}

	return &Manager{
		dir:       dir,
		scheme:    scheme,
		clients:   make(map[string]*cachedClient),
		clientTTL: ttl,
	}
}

// WithSecretFallback configures the Manager to read kubeconfigs from a K8s
// Secret as a fallback when the filesystem path doesn't contain one. This
// avoids the ~60s kubelet sync delay and survives Helm upgrades that wipe
// dynamically-added keys from the Helm-managed Secret.
func (m *Manager) WithSecretFallback(infraClient client.Client, secretName, secretNamespace string) *Manager {
	m.infraClient = infraClient
	m.secretName = secretName
	m.secretNamespace = secretNamespace
	return m
}

// WithAuthProvider configures the Manager to use programmatic EKS token
// generation as a fallback when no kubeconfig file or Secret entry exists.
func (m *Manager) WithAuthProvider(p ClusterAuthProvider) *Manager {
	m.authProvider = p
	return m
}

func (m *Manager) ClientFor(clusterID string) (client.Client, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("cluster id is required")
	}

	m.mu.RLock()
	cached, ok := m.clients[clusterID]
	m.mu.RUnlock()
	if ok && time.Since(cached.createdAt) < m.clientTTL {
		return cached.client, nil
	}

	return m.loadClient(clusterID)
}

// EvictClient removes a cached client, forcing re-creation on next access.
// Used by retry logic when an auth error indicates the client is stale.
func (m *Manager) EvictClient(clusterID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clients, clusterID)
}

func (m *Manager) loadClient(clusterID string) (client.Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check: another goroutine may have loaded it while we waited for the lock
	if cached, ok := m.clients[clusterID]; ok && time.Since(cached.createdAt) < m.clientTTL {
		return cached.client, nil
	}

	cfg, err := m.restConfigFor(clusterID)
	if err != nil {
		return nil, err
	}

	cli, err := client.New(cfg, client.Options{Scheme: m.scheme})
	if err != nil {
		return nil, fmt.Errorf("create kube client for %q: %w", clusterID, err)
	}

	m.clients[clusterID] = &cachedClient{client: cli, createdAt: time.Now()}
	return cli, nil
}

// RestConfigFor returns a REST config for the given cluster ID.
func (m *Manager) RestConfigFor(clusterID string) (*rest.Config, error) {
	return m.restConfigFor(clusterID)
}

func (m *Manager) restConfigFor(clusterID string) (*rest.Config, error) {
	// 1. Try filesystem first.
	path, fsErr := m.resolvePath(clusterID)
	if fsErr == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read kubeconfig %q: %w", path, err)
		}
		cfg, err := clientcmd.RESTConfigFromKubeConfig(data)
		if err != nil {
			return nil, fmt.Errorf("load kubeconfig %q: %w", path, err)
		}
		cfg.UserAgent = "aegis-platform-api"
		return cfg, nil
	}

	// 2. Fall back to K8s Secret.
	if m.infraClient != nil && m.secretName != "" {
		cfg, err := m.restConfigFromSecret(clusterID)
		if err == nil {
			return cfg, nil
		}
		// Don't return yet — try programmatic token next.
	}

	// 3. Fall back to programmatic EKS token generation.
	if m.authProvider != nil {
		cfg, err := m.tokenRESTConfig(clusterID)
		if err == nil {
			return cfg, nil
		}
	}

	// 4. Fall back to in-cluster config for co-located spokes.
	// If platform-api is on the same cluster as the spoke, the service account
	// already has RBAC access to create workloads in spoke namespaces.
	inClusterCfg, err := rest.InClusterConfig()
	if err == nil {
		inClusterCfg.UserAgent = "aegis-platform-api/in-cluster"
		return inClusterCfg, nil
	}

	return nil, fmt.Errorf("no kubeconfig or auth info for cluster %q: %w", clusterID, fsErr)
}

// tokenRESTConfig builds a rest.Config using programmatic EKS token generation.
// It looks up the cluster's endpoint, CA, and credentials from the auth provider
// and injects a bearer token on each HTTP request.
func (m *Manager) tokenRESTConfig(clusterID string) (*rest.Config, error) {
	if m.authProvider == nil {
		return nil, fmt.Errorf("no auth provider configured")
	}

	info, err := m.authProvider.GetClusterAuthInfo(clusterID)
	if err != nil {
		return nil, fmt.Errorf("get cluster auth info for %q: %w", clusterID, err)
	}
	if info.Endpoint == "" || info.CAData == "" {
		return nil, fmt.Errorf("cluster %q missing endpoint or CA data", clusterID)
	}

	caBytes, err := base64.StdEncoding.DecodeString(info.CAData)
	if err != nil {
		return nil, fmt.Errorf("decode CA data for cluster %q: %w", clusterID, err)
	}

	provider := &ekstoken.Provider{
		Region:      info.Region,
		ClusterName: info.ClusterName,
		RoleARN:     info.RoleARN,
		ExternalID:  info.ExternalID,
	}

	cfg := &rest.Config{
		Host:      info.Endpoint,
		UserAgent: "aegis-platform-api",
		TLSClientConfig: rest.TLSClientConfig{
			CAData: caBytes,
		},
		WrapTransport: func(rt http.RoundTripper) http.RoundTripper {
			return &tokenInjectingRoundTripper{
				delegate: rt,
				provider: provider,
			}
		},
	}
	return cfg, nil
}

// tokenInjectingRoundTripper injects a bearer token from an ekstoken.Provider
// into each HTTP request, refreshing the token when it nears expiry.
type tokenInjectingRoundTripper struct {
	delegate http.RoundTripper
	provider *ekstoken.Provider

	mu     sync.Mutex
	token  string
	expiry time.Time
}

func (t *tokenInjectingRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	token, err := t.getToken(req.Context())
	if err != nil {
		return nil, fmt.Errorf("get EKS token: %w", err)
	}
	req = req.Clone(req.Context())
	req.Header.Set("Authorization", "Bearer "+token)
	return t.delegate.RoundTrip(req)
}

func (t *tokenInjectingRoundTripper) getToken(ctx context.Context) (string, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Refresh if token is empty or expires in less than 2 minutes
	if t.token == "" || time.Until(t.expiry) < 2*time.Minute {
		token, expiry, err := t.provider.Token(ctx)
		if err != nil {
			return "", err
		}
		t.token = token
		t.expiry = expiry
	}
	return t.token, nil
}

// HasKubeconfig checks if a kubeconfig is available for the given cluster ID,
// either on the filesystem or in the K8s Secret fallback.
func (m *Manager) HasKubeconfig(clusterID string) bool {
	if clusterID == "" {
		return false
	}
	if m.dir != "" {
		if _, err := m.resolvePath(clusterID); err == nil {
			return true
		}
	}
	// Check Secret fallback.
	if m.infraClient != nil && m.secretName != "" {
		if _, err := m.secretKubeconfigData(clusterID); err == nil {
			return true
		}
	}
	// Check programmatic token auth.
	if m.authProvider != nil {
		if info, err := m.authProvider.GetClusterAuthInfo(clusterID); err == nil && info.Endpoint != "" {
			return true
		}
	}
	// In-cluster fallback: if running inside Kubernetes, in-cluster config
	// is always available and can reach any namespace on the same cluster.
	if _, err := rest.InClusterConfig(); err == nil {
		return true
	}
	return false
}

// Dir returns the kubeconfigs directory path.
func (m *Manager) Dir() string {
	return m.dir
}

func (m *Manager) resolvePath(clusterID string) (string, error) {
	if m.dir == "" {
		return "", fmt.Errorf("kubeconfigs directory not configured")
	}

	candidates := []string{
		filepath.Join(m.dir, clusterID),
		filepath.Join(m.dir, clusterID+".kubeconfig"),
		filepath.Join(m.dir, clusterID+".yaml"),
		filepath.Join(m.dir, clusterID+".yml"),
		filepath.Join(m.dir, clusterID+".conf"),
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	entries, err := os.ReadDir(m.dir)
	if err != nil {
		return "", fmt.Errorf("list kubeconfigs in %q: %w", m.dir, err)
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		base := strings.TrimSuffix(name, filepath.Ext(name))
		if base == clusterID {
			return filepath.Join(m.dir, name), nil
		}
	}

	return "", fmt.Errorf("no kubeconfig found for cluster %q in %q", clusterID, m.dir)
}

// restConfigFromSecret reads a kubeconfig from the K8s Secret and returns a REST config.
func (m *Manager) restConfigFromSecret(clusterID string) (*rest.Config, error) {
	data, err := m.secretKubeconfigData(clusterID)
	if err != nil {
		return nil, err
	}
	cfg, err := clientcmd.RESTConfigFromKubeConfig(data)
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig from secret for %q: %w", clusterID, err)
	}
	cfg.UserAgent = "aegis-platform-api"
	return cfg, nil
}

// secretKubeconfigData reads raw kubeconfig bytes from the K8s Secret, trying
// multiple key patterns to match what upsertClusterKubeconfigSecret writes.
func (m *Manager) secretKubeconfigData(clusterID string) ([]byte, error) {
	secret := &corev1.Secret{}
	key := types.NamespacedName{Name: m.secretName, Namespace: m.secretNamespace}
	if err := m.infraClient.Get(context.Background(), key, secret); err != nil {
		return nil, fmt.Errorf("get kubeconfig secret %s/%s: %w", m.secretNamespace, m.secretName, err)
	}
	if secret.Data == nil {
		return nil, fmt.Errorf("kubeconfig secret %s/%s has no data", m.secretNamespace, m.secretName)
	}
	candidates := []string{
		clusterID,
		clusterID + ".kubeconfig",
		clusterID + ".yaml",
		clusterID + ".yml",
		clusterID + ".conf",
	}
	for _, k := range candidates {
		if v, ok := secret.Data[k]; ok && len(v) > 0 {
			return v, nil
		}
	}
	return nil, fmt.Errorf("no kubeconfig key for cluster %q in secret %s/%s", clusterID, m.secretNamespace, m.secretName)
}
