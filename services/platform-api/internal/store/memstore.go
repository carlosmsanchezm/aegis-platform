package store

import (
	"fmt"
	"sync"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

const (
	statusPlaced  = "PLACED"
	statusRunning = "RUNNING"
)

type MemStore struct {
	mu        sync.RWMutex
	projects  map[string]*aegis.Project
	budgets   map[string]*aegis.Budget
	flavors   map[string]*aegis.Flavor
	queues    map[string]*aegis.Queue
	workloads map[string]*aegis.Workload

	// cluster state is managed via clusterState for fine-grained locking
	cstate *clusterState
}

func NewMemStore() *MemStore {
	return &MemStore{
		projects:  map[string]*aegis.Project{},
		budgets:   map[string]*aegis.Budget{},
		flavors:   map[string]*aegis.Flavor{},
		queues:    map[string]*aegis.Queue{},
		workloads: map[string]*aegis.Workload{},
		cstate:    newClusterState(),
	}
}

// -------- projects/budgets/flavors/queues/workloads --------

func (s *MemStore) PutProject(p *aegis.Project) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.projects[p.Id] = p
}
func (s *MemStore) GetProject(id string) *aegis.Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.projects[id]
}

func (s *MemStore) PutBudget(b *aegis.Budget) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.budgets[b.ProjectId] = b
}
func (s *MemStore) GetBudget(projectID string) *aegis.Budget {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.budgets[projectID]
}

func (s *MemStore) PutFlavor(f *aegis.Flavor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flavors[f.Name] = f
}
func (s *MemStore) GetFlavor(name string) *aegis.Flavor {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.flavors[name]
}

func (s *MemStore) PutQueue(q *aegis.Queue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.queues[q.Name] = q
}
func (s *MemStore) GetQueue(name string) *aegis.Queue {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.queues[name]
}

func (s *MemStore) PutWorkload(w *aegis.Workload) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.workloads[w.Id] = w
}
func (s *MemStore) GetWorkload(id string) *aegis.Workload {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.workloads[id]
}
func (s *MemStore) ListWorkloads(projectID string) []*aegis.Workload {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*aegis.Workload{}
	for _, w := range s.workloads {
		if w.ProjectId == projectID {
			out = append(out, w)
		}
	}
	return out
}

// -------- clusters --------

func (s *MemStore) UpsertClusterFromRegister(req *aegis.ClusterRegisterRequest) {
	s.cstate.upsertFromRegister(req)
}
func (s *MemStore) UpdateClusterFromHeartbeat(hb *aegis.ClusterHeartbeat) {
	s.cstate.updateFromHeartbeat(hb)
}
func (s *MemStore) ListClusterInfos() []*ClusterInfo {
	return s.cstate.list()
}

// LeaseWorkloads moves up to max workloads for the cluster from PLACED to RUNNING.
func (s *MemStore) LeaseWorkloads(clusterID string, max int) []*aegis.Workload {
	if max <= 0 {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	leased := make([]*aegis.Workload, 0, max)
	for _, w := range s.workloads {
		if w.GetClusterId() != clusterID || w.GetStatus() != statusPlaced {
			continue
		}
		w.Status = statusRunning
		leased = append(leased, w)
		if len(leased) >= max {
			break
		}
	}
	return leased
}

// AckWorkload updates the workload status if currently RUNNING.
func (s *MemStore) AckWorkload(id string, nextStatus string) (*aegis.Workload, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.workloads[id]
	if !ok {
		return nil, fmt.Errorf("workload %s not found", id)
	}
	if w.GetStatus() != statusRunning {
		return nil, fmt.Errorf("workload %s not in RUNNING state", id)
	}
	w.Status = nextStatus
	return w, nil
}
