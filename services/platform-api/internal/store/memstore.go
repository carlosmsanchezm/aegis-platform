package store

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

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
	placedAt  map[string]time.Time
	startedAt map[string]time.Time
	estUSD    map[string]float64

	// cluster state is managed via clusterState for fine-grained locking
	cstate *clusterState

	usage map[string]*budgetUsage

	sessions           map[string]*ConnectionSession
	sessionsByWorkload map[string]map[string]struct{}
	sessionJTIs        map[string]*jtiRecord

	provisioningLogs map[string][]ProvisioningLogEntry
	provisioningRuns map[string]*ProvisioningRun
	provisioningSeq  int64
}

func NewMemStore() *MemStore {
	return &MemStore{
		projects:           map[string]*aegis.Project{},
		budgets:            map[string]*aegis.Budget{},
		flavors:            map[string]*aegis.Flavor{},
		queues:             map[string]*aegis.Queue{},
		workloads:          map[string]*aegis.Workload{},
		placedAt:           map[string]time.Time{},
		startedAt:          map[string]time.Time{},
		estUSD:             map[string]float64{},
		cstate:             newClusterState(),
		usage:              map[string]*budgetUsage{},
		sessions:           map[string]*ConnectionSession{},
		sessionsByWorkload: map[string]map[string]struct{}{},
		sessionJTIs:        map[string]*jtiRecord{},
		provisioningLogs:   map[string][]ProvisioningLogEntry{},
		provisioningRuns:   map[string]*ProvisioningRun{},
	}
}

// memstore implementation uses the ConnectionSession type from store.go

type jtiRecord struct {
	SessionID string
	Used      bool
	ExpiresAt time.Time
}

func copySession(in *ConnectionSession) *ConnectionSession {
	if in == nil {
		return nil
	}
	out := *in
	return &out
}

// PutConnectionSession upserts the provided session and returns an immutable copy of the stored value.
func (s *MemStore) PutConnectionSession(sess *ConnectionSession) *ConnectionSession {
	if sess == nil || sess.SessionID == "" {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	existing, found := s.sessions[sess.SessionID]
	if found {
		sess.CreatedAt = existing.CreatedAt
	} else if sess.CreatedAt.IsZero() {
		sess.CreatedAt = now
	}
	sess.UpdatedAt = now

	stored := copySession(sess)
	s.sessions[sess.SessionID] = stored

	if _, ok := s.sessionsByWorkload[sess.WorkloadID]; !ok {
		s.sessionsByWorkload[sess.WorkloadID] = map[string]struct{}{}
	}
	s.sessionsByWorkload[sess.WorkloadID][sess.SessionID] = struct{}{}

	if sess.JTI != "" {
		s.sessionJTIs[sess.JTI] = &jtiRecord{SessionID: sess.SessionID, Used: sess.Used, ExpiresAt: sess.ExpiresAt}
	}

	return copySession(stored)
}

// ConnectionSession returns a copy of the stored session for the provided identifier.
func (s *MemStore) ConnectionSession(id string) (*ConnectionSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	return copySession(sess), ok
}

// ConnectionSessionByJTI resolves a session by JWT ID for audit and invalidation flows.
func (s *MemStore) ConnectionSessionByJTI(jti string) (*ConnectionSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	rec, ok := s.sessionJTIs[jti]
	if !ok {
		return nil, false
	}
	sess, ok := s.sessions[rec.SessionID]
	return copySession(sess), ok
}

// UpdateConnectionSession applies a mutation function while holding the lock to ensure consistency.
func (s *MemStore) UpdateConnectionSession(id string, mutate func(*ConnectionSession) error) (*ConnectionSession, error) {
	if mutate == nil {
		return nil, errors.New("mutate function required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, ok := s.sessions[id]
	if !ok {
		return nil, ErrSessionNotFound
	}

	prevJTI := existing.JTI
	if err := mutate(existing); err != nil {
		return nil, err
	}
	existing.UpdatedAt = time.Now()

	s.sessions[id] = existing

	if prevJTI != existing.JTI {
		delete(s.sessionJTIs, prevJTI)
	}
	if existing.JTI != "" {
		s.sessionJTIs[existing.JTI] = &jtiRecord{SessionID: existing.SessionID, Used: existing.Used, ExpiresAt: existing.ExpiresAt}
	}

	return copySession(existing), nil
}

// MarkSessionUsed toggles the session usage flag and returns whether the session was present.
func (s *MemStore) MarkSessionUsed(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return false
	}
	sess.Used = true
	sess.UpdatedAt = time.Now()
	s.sessions[id] = sess
	if rec, ok := s.sessionJTIs[sess.JTI]; ok {
		rec.Used = true
	}
	return true
}

// MarkJTIUsed marks the JTI as consumed; returns false if not tracked.
func (s *MemStore) MarkJTIUsed(jti string) bool {
	if jti == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	rec, ok := s.sessionJTIs[jti]
	if !ok {
		return false
	}
	rec.Used = true
	if sess, ok := s.sessions[rec.SessionID]; ok {
		sess.Used = true
		sess.UpdatedAt = time.Now()
		s.sessions[rec.SessionID] = sess
	}
	return true
}

// DeleteConnectionSession removes the session and its indexes.
func (s *MemStore) DeleteConnectionSession(id string) (*ConnectionSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, false
	}
	delete(s.sessions, id)
	if sess.JTI != "" {
		delete(s.sessionJTIs, sess.JTI)
	}
	if assoc, ok := s.sessionsByWorkload[sess.WorkloadID]; ok {
		delete(assoc, id)
		if len(assoc) == 0 {
			delete(s.sessionsByWorkload, sess.WorkloadID)
		}
	}
	return copySession(sess), true
}

// SessionsForWorkload returns active sessions for a workload ID.
func (s *MemStore) SessionsForWorkload(workloadID string) []*ConnectionSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	assoc := s.sessionsByWorkload[workloadID]
	if len(assoc) == 0 {
		return nil
	}
	out := make([]*ConnectionSession, 0, len(assoc))
	for sessionID := range assoc {
		if sess, ok := s.sessions[sessionID]; ok {
			out = append(out, copySession(sess))
		}
	}
	return out
}

// PurgeExpiredSessions deletes sessions and JTIs that are past expiry.
func (s *MemStore) PurgeExpiredSessions(now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, sess := range s.sessions {
		if sess.ExpiresAt.IsZero() {
			continue
		}
		if now.After(sess.ExpiresAt.Add(5 * time.Minute)) { // grace period for audit
			delete(s.sessions, id)
			if sess.JTI != "" {
				delete(s.sessionJTIs, sess.JTI)
			}
			if assoc, ok := s.sessionsByWorkload[sess.WorkloadID]; ok {
				delete(assoc, id)
				if len(assoc) == 0 {
					delete(s.sessionsByWorkload, sess.WorkloadID)
				}
			}
		}
	}
	for jti, rec := range s.sessionJTIs {
		if now.After(rec.ExpiresAt.Add(5 * time.Minute)) {
			delete(s.sessionJTIs, jti)
		}
	}
}

type budgetUsage struct {
	periodStart time.Time
	reservedUSD float64
	actualUSD   float64
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
func (s *MemStore) ListProjects() []*aegis.Project {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*aegis.Project, 0, len(s.projects))
	for _, p := range s.projects {
		if p == nil {
			continue
		}
		copy := *p
		out = append(out, &copy)
	}
	return out
}

// DeleteProject removes a project from the store.
// Returns an error if the project has active clusters attached.
func (s *MemStore) DeleteProject(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return fmt.Errorf("project id is required")
	}

	// Check for active clusters (via clusterState)
	clusters := s.cstate.listByProject(id)
	if len(clusters) > 0 {
		return fmt.Errorf("cannot delete project %q: %d active cluster(s) still attached - delete clusters first to avoid incurring costs", id, len(clusters))
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.projects[id]; !exists {
		return fmt.Errorf("project %q not found", id)
	}

	delete(s.projects, id)
	return nil
}

// HasActiveClusters checks if a project has any clusters attached.
func (s *MemStore) HasActiveClusters(projectID string) bool {
	clusters := s.cstate.listByProject(projectID)
	return len(clusters) > 0
}

func budgetKey(projectID, queue string) string { return projectID + "|" + queue }

func monthStartUTC(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func monthEndUTC(t time.Time) time.Time { return monthStartUTC(t).AddDate(0, 1, 0) }

// BudgetUsageView is defined in store.go

func (s *MemStore) PutBudget(b *aegis.Budget) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.budgets[budgetKey(b.ProjectId, b.Queue)] = b
}
func (s *MemStore) GetBudget(projectID string) *aegis.Budget {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.budgets[budgetKey(projectID, "")]
}
func (s *MemStore) GetBudgetExact(projectID, queue string) *aegis.Budget {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.budgets[budgetKey(projectID, queue)]
}
func (s *MemStore) ResolveBudget(projectID, queue string) (*aegis.Budget, string) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if b, ok := s.budgets[budgetKey(projectID, queue)]; ok {
		return b, budgetKey(projectID, queue)
	}
	if b, ok := s.budgets[budgetKey(projectID, "")]; ok {
		return b, budgetKey(projectID, "")
	}
	return nil, ""
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

func (s *MemStore) MarkPlaced(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.placedAt[id] = time.Now()
}

func (s *MemStore) GetPlacedAt(id string) (time.Time, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.placedAt[id]
	return t, ok
}

func (s *MemStore) ClearPlacedAt(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.placedAt, id)
}

func (s *MemStore) MarkStarted(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.startedAt[id] = time.Now()
}

func (s *MemStore) GetStartedAt(id string) (time.Time, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.startedAt[id]
	return t, ok
}

func (s *MemStore) SetEstimateUSD(id string, usd float64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.estUSD[id] = usd
}

func (s *MemStore) PopEstimateUSD(id string) float64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	v := s.estUSD[id]
	delete(s.estUSD, id)
	return v
}

// StartWorkload transitions a workload from PLACED to RUNNING. It returns the
// workload alongside the observed queue wait and a boolean indicating whether a
// wait duration was recorded. Subsequent invocations when the workload is
// already RUNNING are treated as no-ops.
func (s *MemStore) StartWorkload(id string) (*aegis.Workload, time.Duration, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	w, ok := s.workloads[id]
	if !ok {
		return nil, 0, false, fmt.Errorf("workload %s not found", id)
	}

	switch w.GetStatus() {
	case statusPlaced:
		w.Status = statusRunning
		s.startedAt[id] = time.Now()
		if t0, ok := s.placedAt[id]; ok {
			delete(s.placedAt, id)
			return w, time.Since(t0), true, nil
		}
		return w, 0, false, nil
	case statusRunning:
		return w, 0, false, nil
	default:
		return nil, 0, false, fmt.Errorf("workload %s not in a startable state", id)
	}
}

// -------- clusters --------

func (s *MemStore) UpsertClusterFromRegister(req *aegis.ClusterRegisterRequest) {
	s.cstate.upsertFromRegister(req)
}
func (s *MemStore) UpdateClusterFromHeartbeat(hb *aegis.ClusterHeartbeat) {
	s.cstate.updateFromHeartbeat(hb)
}
func (s *MemStore) GetClusterInfo(clusterID string) *ClusterInfo {
	return s.cstate.get(clusterID)
}
func (s *MemStore) ListClusterInfos() []*ClusterInfo {
	return s.cstate.list()
}

func (s *MemStore) ListClustersByProject(projectID string) []*ClusterInfo {
	return s.cstate.listByProject(projectID)
}

func (s *MemStore) SetClusterProjectID(clusterID, projectID string) {
	s.cstate.setProjectID(clusterID, projectID)
}

func (s *MemStore) DeleteCluster(clusterID string) {
	s.cstate.delete(clusterID)
}

// CleanupStaleClusters is a no-op for memory store (no persistence).
func (s *MemStore) CleanupStaleClusters(staleThreshold string) int64 {
	return 0
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
		s.startedAt[w.Id] = time.Now()
		leased = append(leased, w)
		if len(leased) >= max {
			break
		}
	}
	return leased
}

// AckWorkload updates the workload status if currently RUNNING and stamps optional URL.
func (s *MemStore) AckWorkload(id string, nextStatus string, url string) (*aegis.Workload, error) {
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
	if url != "" {
		w.Url = url
	}
	return w, nil
}

// ---- Budget usage helpers ----

func (s *MemStore) ReserveIfAllowed(projectID, queue string, estimateUSD float64) (allowed bool, policy string, reason string, view BudgetUsageView) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	b, key := s.resolveBudgetLocked(projectID, queue)
	if b == nil {
		return true, "", "", BudgetUsageView{}
	}
	u := s.ensureUsageLocked(key, now)
	next := u.actualUSD + u.reservedUSD + estimateUSD
	policy = b.GetPolicyMode()
	if next > b.GetLimitUsd() && strings.EqualFold(policy, "HARD") {
		return false, policy, "insufficient_funds", BudgetUsageView{ReservedUSD: u.reservedUSD, ActualUSD: u.actualUSD, PeriodStart: u.periodStart, PeriodEnd: monthEndUTC(now)}
	}
	u.reservedUSD += estimateUSD
	return true, policy, "", BudgetUsageView{ReservedUSD: u.reservedUSD, ActualUSD: u.actualUSD, PeriodStart: u.periodStart, PeriodEnd: monthEndUTC(now)}
}

func (s *MemStore) ReconcileOnAck(projectID, queue string, estUSD, actualUSD float64) BudgetUsageView {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	_, key := s.resolveBudgetLocked(projectID, queue)
	if key == "" {
		return BudgetUsageView{}
	}
	u := s.ensureUsageLocked(key, now)
	u.reservedUSD -= estUSD
	if u.reservedUSD < 0 {
		u.reservedUSD = 0
	}
	u.actualUSD += actualUSD
	return BudgetUsageView{ReservedUSD: u.reservedUSD, ActualUSD: u.actualUSD, PeriodStart: u.periodStart, PeriodEnd: monthEndUTC(now)}
}

func (s *MemStore) UsageView(projectID, queue string) (BudgetUsageView, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now().UTC()
	_, key := s.resolveBudgetLocked(projectID, queue)
	if key == "" {
		return BudgetUsageView{}, false
	}
	u, ok := s.usage[key]
	if !ok {
		return BudgetUsageView{ReservedUSD: 0, ActualUSD: 0, PeriodStart: monthStartUTC(now), PeriodEnd: monthEndUTC(now)}, true
	}
	return BudgetUsageView{ReservedUSD: u.reservedUSD, ActualUSD: u.actualUSD, PeriodStart: u.periodStart, PeriodEnd: monthEndUTC(now)}, true
}

func (s *MemStore) ListBudgets(filterProject string) []*aegis.Budget {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := []*aegis.Budget{}
	for _, b := range s.budgets {
		if filterProject == "" || b.GetProjectId() == filterProject {
			out = append(out, b)
		}
	}
	return out
}

func (s *MemStore) ensureUsageLocked(key string, now time.Time) *budgetUsage {
	u, ok := s.usage[key]
	if !ok || !u.periodStart.Equal(monthStartUTC(now)) {
		u = &budgetUsage{periodStart: monthStartUTC(now)}
		s.usage[key] = u
	}
	return u
}

func (s *MemStore) resolveBudgetLocked(projectID, queue string) (*aegis.Budget, string) {
	if b, ok := s.budgets[budgetKey(projectID, queue)]; ok {
		return b, budgetKey(projectID, queue)
	}
	if b, ok := s.budgets[budgetKey(projectID, "")]; ok {
		return b, budgetKey(projectID, "")
	}
	return nil, ""
}

// -------- provisioning logs --------

func (s *MemStore) AppendProvisioningLog(entry ProvisioningLogEntry) {
	if strings.TrimSpace(entry.JobID) == "" || strings.TrimSpace(entry.Message) == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if entry.CreatedAt.IsZero() {
		entry.CreatedAt = time.Now()
	}
	s.provisioningSeq++
	if entry.Sequence == 0 {
		entry.Sequence = s.provisioningSeq
	}
	clone := entry
	s.provisioningLogs[entry.JobID] = append(s.provisioningLogs[entry.JobID], clone)
}

func (s *MemStore) ListProvisioningLogs(jobID string, since time.Time, sinceSeq int64, limit int) []ProvisioningLogEntry {
	if strings.TrimSpace(jobID) == "" {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}
	source := s.provisioningLogs[jobID]
	out := make([]ProvisioningLogEntry, 0, limit)
	for _, entry := range source {
		if !since.IsZero() {
			if entry.CreatedAt.Before(since) {
				continue
			}
			if entry.CreatedAt.Equal(since) {
				if sinceSeq > 0 {
					if entry.Sequence <= sinceSeq {
						continue
					}
				} else {
					// legacy cursor without sequence: mimic strict time cursor to avoid duplicates
					continue
				}
			}
		}
		copy := entry
		out = append(out, copy)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func (s *MemStore) ClearProvisioningLogs(jobID string) {
	if strings.TrimSpace(jobID) == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.provisioningLogs, jobID)
}

func (s *MemStore) DeleteOldProvisioningLogs(olderThan time.Time) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	var deleted int64
	for jobID, logs := range s.provisioningLogs {
		var kept []ProvisioningLogEntry
		for _, entry := range logs {
			if entry.CreatedAt.After(olderThan) {
				kept = append(kept, entry)
			} else {
				deleted++
			}
		}
		if len(kept) == 0 {
			delete(s.provisioningLogs, jobID)
		} else {
			s.provisioningLogs[jobID] = kept
		}
	}
	return deleted
}

func (s *MemStore) UpsertProvisioningRun(run ProvisioningRun) {
	if strings.TrimSpace(run.JobID) == "" {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	existing, found := s.provisioningRuns[run.JobID]
	merged := run
	if merged.StartedAt.IsZero() {
		if found && existing != nil {
			merged.StartedAt = existing.StartedAt
		} else {
			merged.StartedAt = now
		}
	}
	if merged.Phase == "" && found && existing != nil {
		merged.Phase = existing.Phase
	}
	if merged.ProjectID == "" && found && existing != nil {
		merged.ProjectID = existing.ProjectID
	}
	if merged.ClusterID == "" && found && existing != nil {
		merged.ClusterID = existing.ClusterID
	}
	merged.UpdatedAt = now
	s.provisioningRuns[run.JobID] = &merged
}

func (s *MemStore) GetProvisioningRun(jobID string) (*ProvisioningRun, bool) {
	if strings.TrimSpace(jobID) == "" {
		return nil, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	run, ok := s.provisioningRuns[jobID]
	if !ok || run == nil {
		return nil, false
	}
	copy := *run
	if run.CompletedAt != nil {
		ts := *run.CompletedAt
		copy.CompletedAt = &ts
	}
	return &copy, true
}

func (s *MemStore) ListProvisioningRuns(projectID string) []*ProvisioningRun {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*ProvisioningRun, 0, len(s.provisioningRuns))
	for _, run := range s.provisioningRuns {
		if run == nil {
			continue
		}
		if projectID != "" && !strings.EqualFold(run.ProjectID, projectID) {
			continue
		}
		copy := *run
		if run.CompletedAt != nil {
			ts := *run.CompletedAt
			copy.CompletedAt = &ts
		}
		out = append(out, &copy)
	}
	return out
}
