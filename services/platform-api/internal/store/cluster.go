package store

import (
	"sync"
	"time"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

// ClusterInfo represents the latest known snapshot of a cluster.
type ClusterInfo struct {
	ID                  string
	ProjectID           string // Direct FK to projects table for multi-tenancy
	Provider            string
	Region              string
	Labels              map[string]string
	AvailableFlavorSet  map[string]bool
	TTFGSecondsP50      float64
	LastHeartbeat       time.Time
	ProxyURL            string // spoke proxy URL for this cluster (e.g., "wss://proxy.cluster.example.com")
	CreatedAt           time.Time
	DeletedAt           *time.Time // Soft delete timestamp
	ImportMethod        string
	ImportedAt          time.Time
	KubeconfigSecretRef string
	AssumeRoleARN       string
}

// ClusterImport captures the metadata required to register an existing cluster
// in the control plane before the spoke agent is installed.
type ClusterImport struct {
	ClusterID           string
	ProjectID           string
	Provider            string
	Region              string
	Labels              map[string]string
	ImportMethod        string
	ImportedAt          time.Time
	KubeconfigSecretRef string
	AssumeRoleARN       string
}

type clusterState struct {
	mu       sync.RWMutex
	clusters map[string]*ClusterInfo
}

func newClusterState() *clusterState {
	return &clusterState{clusters: map[string]*ClusterInfo{}}
}

func (cs *clusterState) upsertFromRegister(req *aegis.ClusterRegisterRequest) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	ci, ok := cs.clusters[req.ClusterId]
	if !ok {
		ci = &ClusterInfo{ID: req.ClusterId}
		cs.clusters[req.ClusterId] = ci
		ci.CreatedAt = time.Now()
	}
	if ci.ImportMethod == "" {
		ci.ImportMethod = "provisioned"
	}
	ci.Provider = req.GetProvider()
	ci.Region = req.GetRegion()
	if ci.Labels == nil {
		ci.Labels = map[string]string{}
	}
	for k, v := range req.GetLabels() {
		if k == "" {
			continue
		}
		ci.Labels[k] = v
	}
	// Set proxy URL if provided during registration (useful for spoke clusters)
	if proxyURL := req.GetProxyUrl(); proxyURL != "" {
		ci.ProxyURL = proxyURL
	}
	// Do not touch flavors/TTFG here; those arrive in heartbeat.
}

func (cs *clusterState) updateFromHeartbeat(hb *aegis.ClusterHeartbeat) {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	ci, ok := cs.clusters[hb.ClusterId]
	if !ok {
		ci = &ClusterInfo{ID: hb.ClusterId}
		cs.clusters[hb.ClusterId] = ci
		ci.CreatedAt = time.Now()
	}
	newSet := map[string]bool{}
	for _, f := range hb.GetAvailableFlavors() {
		name := f.GetName()
		if name != "" {
			newSet[name] = true
		}
	}
	if ci.AvailableFlavorSet != nil {
		for k := range ci.AvailableFlavorSet {
			newSet[k] = true
		}
	}
	if len(newSet) == 0 && len(ci.AvailableFlavorSet) > 0 {
		newSet = ci.AvailableFlavorSet
	}
	ci.AvailableFlavorSet = newSet
	ci.TTFGSecondsP50 = hb.GetTtfGpuSecondsP50()
	ci.LastHeartbeat = time.Now()
	// Update proxy URL if provided in heartbeat (spoke proxy)
	if proxyURL := hb.GetProxyUrl(); proxyURL != "" {
		ci.ProxyURL = proxyURL
	}
}

func (cs *clusterState) get(clusterID string) *ClusterInfo {
	if clusterID == "" {
		return nil
	}
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	ci, ok := cs.clusters[clusterID]
	if !ok {
		return nil
	}
	// Shallow copy to avoid external mutation
	cpy := *ci
	return &cpy
}

func (cs *clusterState) list() []*ClusterInfo {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	out := make([]*ClusterInfo, 0, len(cs.clusters))
	for _, ci := range cs.clusters {
		// Shallow copy to avoid external mutation
		cpy := *ci
		out = append(out, &cpy)
	}
	return out
}

func (cs *clusterState) listByProject(projectID string) []*ClusterInfo {
	if projectID == "" {
		return nil
	}
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	out := make([]*ClusterInfo, 0)
	for _, ci := range cs.clusters {
		// Check the ProjectID field first (new FK-based approach)
		if ci.ProjectID == projectID {
			cpy := *ci
			out = append(out, &cpy)
			continue
		}
		// Fall back to checking the label for backward compatibility
		if ci.Labels != nil {
			if labelProjID, ok := ci.Labels["aegis.yourorg.dev/projectId"]; ok && labelProjID == projectID {
				cpy := *ci
				out = append(out, &cpy)
			}
		}
	}
	return out
}

func (cs *clusterState) setProjectID(clusterID, projectID string) {
	if clusterID == "" || projectID == "" {
		return
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	ci, ok := cs.clusters[clusterID]
	if !ok {
		ci = &ClusterInfo{ID: clusterID, Labels: map[string]string{}}
		cs.clusters[clusterID] = ci
	}
	// Set the direct ProjectID field (used by new FK-based code)
	ci.ProjectID = projectID
	// Also maintain label for backward compatibility
	if ci.Labels == nil {
		ci.Labels = map[string]string{}
	}
	ci.Labels["aegis.yourorg.dev/projectId"] = projectID
}

func (cs *clusterState) delete(clusterID string) {
	if clusterID == "" {
		return
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	delete(cs.clusters, clusterID)
}

// preRegister creates or updates a cluster entry with provisioning-time information.
// This is called before the k8s-agent connects to ensure project_id and proxy_url are set.
func (cs *clusterState) preRegister(clusterID, projectID, provider, region, proxyURL string) {
	if clusterID == "" {
		return
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	ci, ok := cs.clusters[clusterID]
	if !ok {
		ci = &ClusterInfo{
			ID:                 clusterID,
			Labels:             map[string]string{},
			AvailableFlavorSet: map[string]bool{},
			CreatedAt:          time.Now(),
		}
		cs.clusters[clusterID] = ci
	}
	if projectID != "" {
		ci.ProjectID = projectID
		ci.Labels["aegis.yourorg.dev/projectId"] = projectID
	}
	if provider != "" {
		ci.Provider = provider
	}
	if region != "" {
		ci.Region = region
	}
	if proxyURL != "" {
		ci.ProxyURL = proxyURL
	}
}
