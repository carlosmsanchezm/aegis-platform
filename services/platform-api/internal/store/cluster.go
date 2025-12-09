package store

import (
	"sync"
	"time"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

// ClusterInfo represents the latest known snapshot of a cluster.
type ClusterInfo struct {
	ID                 string
	Provider           string
	Region             string
	Labels             map[string]string
	AvailableFlavorSet map[string]bool
	TTFGSecondsP50     float64
	LastHeartbeat      time.Time
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
	}
	ci.Provider = req.GetProvider()
	ci.Region = req.GetRegion()
	ci.Labels = map[string]string{}
	for k, v := range req.GetLabels() {
		ci.Labels[k] = v
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
