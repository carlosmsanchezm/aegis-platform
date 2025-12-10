package kubeclients

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	aegisv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Manager struct {
	dir     string
	scheme  *runtime.Scheme
	mu      sync.RWMutex
	clients map[string]client.Client
}

func New(dir string) *Manager {
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = aegisv1alpha1.AddToScheme(scheme)

	workspaceGV := schema.GroupVersion{Group: "aegis.yourorg.dev", Version: "v1alpha2"}
	scheme.AddKnownTypeWithName(workspaceGV.WithKind("Workspace"), &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(workspaceGV.WithKind("WorkspaceList"), &unstructured.UnstructuredList{})

	return &Manager{
		dir:     dir,
		scheme:  scheme,
		clients: make(map[string]client.Client),
	}
}

func (m *Manager) ClientFor(clusterID string) (client.Client, error) {
	if clusterID == "" {
		return nil, fmt.Errorf("cluster id is required")
	}

	m.mu.RLock()
	cli, ok := m.clients[clusterID]
	m.mu.RUnlock()
	if ok {
		return cli, nil
	}

	return m.loadClient(clusterID)
}

func (m *Manager) loadClient(clusterID string) (client.Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if cli, ok := m.clients[clusterID]; ok {
		return cli, nil
	}

	cfg, err := m.restConfigFor(clusterID)
	if err != nil {
		return nil, err
	}

	cli, err := client.New(cfg, client.Options{Scheme: m.scheme})
	if err != nil {
		return nil, fmt.Errorf("create kube client for %q: %w", clusterID, err)
	}

	m.clients[clusterID] = cli
	return cli, nil
}

// RestConfigFor returns a REST config for the given cluster ID.
func (m *Manager) RestConfigFor(clusterID string) (*rest.Config, error) {
	return m.restConfigFor(clusterID)
}

func (m *Manager) restConfigFor(clusterID string) (*rest.Config, error) {
	path, err := m.resolvePath(clusterID)
	if err != nil {
		return nil, err
	}

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
