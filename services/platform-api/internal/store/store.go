package store

import (
	"errors"
	"time"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

// ErrSessionNotFound is returned when a requested connection session cannot be located.
var ErrSessionNotFound = errors.New("connection session not found")

// ErrClusterProjectConflict is returned when an operation would change the
// project ownership of an existing cluster.
var ErrClusterProjectConflict = errors.New("cluster already associated with a different project")

// BudgetUsageView exposes a consistent snapshot of reserved and actual spend
// figures for the current UTC accounting period.
type BudgetUsageView struct {
	ReservedUSD float64
	ActualUSD   float64
	PeriodStart time.Time
	PeriodEnd   time.Time
}

// ConnectionSession captures the persisted state for a single-use remote access token.
type ConnectionSession struct {
	SessionID    string
	WorkloadID   string
	Subject      string
	Client       string
	JTI          string
	Token        string
	SSHUser      string
	SSHHostAlias string
	InternalHost string
	Port         int32
	SSHConfig    string
	ProxyURL     string
	ProxyCAPem    string
	VSCodeURI     string
	WorkspaceRoot string
	ExpiresAt    time.Time
	OneTime      bool
	Used         bool
	Revoked      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// ProvisioningLogEntry captures a single provisioning log line tied to a ProjectInfra job.
type ProvisioningLogEntry struct {
	JobID     string
	ProjectID string
	ClusterID string
	Phase     string
	Type      string
	Message   string
	CreatedAt time.Time
	Sequence  int64
}

const (
	LogTypeProgress = "progress"
	LogTypeEvent    = "event"
	LogTypeError    = "error"
)

// ProvisioningRun tracks coarse-grained lifecycle metadata for a ProjectInfra job.
type ProvisioningRun struct {
	JobID       string
	ProjectID   string
	ClusterID   string
	Phase       string
	StartedAt   time.Time
	CompletedAt *time.Time
	UpdatedAt   time.Time
}

// AuditEvent represents a single auditable action within the platform.
type AuditEvent struct {
	ID           string
	EventType    string    // e.g. "workload.submitted", "workload.rejected", "session.created"
	Timestamp    time.Time
	Subject      string    // user identity from auth token
	ResourceType string    // "workload", "project", "cluster", "session", "budget"
	ResourceID   string
	Action       string    // "create", "read", "update", "delete"
	Outcome      string    // "success", "failure", "denied"
	Details      map[string]string
	SourceIP     string
}

// AuditEventFilter defines optional criteria for listing audit events.
type AuditEventFilter struct {
	EventType    string
	Subject      string
	ResourceType string
	ResourceID   string
	StartTime    time.Time
	EndTime      time.Time
	Limit        int
}

// ProvisioningLogSink exposes the minimal interface required to persist provisioning log lines.
type ProvisioningLogSink interface {
	AppendProvisioningLog(entry ProvisioningLogEntry)
}

// Store defines the contract satisfied by all persistence backends used by the
// platform API. Implementations must provide their own concurrency controls and
// guarantee that multi-step operations documented as atomic remain so.
type Store interface {
	// catalog
	PutProject(*aegis.Project)
	GetProject(id string) *aegis.Project
	ListProjects() []*aegis.Project
	DeleteProject(id string) error      // Delete project (fails if active clusters exist)
	HasActiveClusters(projectID string) bool // Check if project has non-deleted clusters

	PutBudget(*aegis.Budget)
	GetBudget(projectID string) *aegis.Budget
	GetBudgetExact(projectID, queue string) *aegis.Budget
	ResolveBudget(projectID, queue string) (*aegis.Budget, string)

	PutFlavor(*aegis.Flavor)
	GetFlavor(name string) *aegis.Flavor

	PutQueue(*aegis.Queue)
	GetQueue(name string) *aegis.Queue

	// workloads
	PutWorkload(*aegis.Workload)
	GetWorkload(id string) *aegis.Workload
	ListWorkloads(projectID string) []*aegis.Workload
	StartWorkload(id string) (*aegis.Workload, time.Duration, bool, error)
	AckWorkload(id, nextStatus, url string) (*aegis.Workload, error)
	ResumeWorkload(id string) (*aegis.Workload, error)
	TerminateWorkload(id, reason string) (*aegis.Workload, error)
	RollbackTerminateWorkload(id, previousStatus string) (*aegis.Workload, error)
	SetWorkloadURL(id, url string)
	MarkPlaced(id string)
	GetPlacedAt(id string) (time.Time, bool)
	ClearPlacedAt(id string)
	MarkStarted(id string)
	GetStartedAt(id string) (time.Time, bool)
	GetRuntimeSeconds(id string) (int64, bool)
	SetEstimateUSD(id string, usd float64)
	PopEstimateUSD(id string) float64
	LeaseWorkloads(clusterID string, max int) []*aegis.Workload
	ListClusterWorkloadIDs(clusterID string) ([]string, error)

	// budgets usage
	ReserveIfAllowed(projectID, queue string, estimateUSD float64) (bool, string, string, BudgetUsageView)
	ReconcileOnAck(projectID, queue string, estUSD, actualUSD float64) BudgetUsageView
	UsageView(projectID, queue string) (BudgetUsageView, bool)
	ListBudgets(filterProject string) []*aegis.Budget

	// sessions
	PutConnectionSession(*ConnectionSession) *ConnectionSession
	ConnectionSession(id string) (*ConnectionSession, bool)
	ConnectionSessionByJTI(jti string) (*ConnectionSession, bool)
	UpdateConnectionSession(id string, mutate func(*ConnectionSession) error) (*ConnectionSession, error)
	MarkSessionUsed(id string) bool
	MarkJTIUsed(jti string) bool
	DeleteConnectionSession(id string) (*ConnectionSession, bool)
	SessionsForWorkload(workloadID string) []*ConnectionSession
	PurgeExpiredSessions(now time.Time)

	// clusters
	// PreRegisterCluster creates a placeholder cluster row during provisioning,
	// before the k8s-agent connects. This ensures project_id and proxy_url are set
	// when the agent's RegisterCluster call updates the row.
	PreRegisterCluster(clusterID, projectID, provider, region, proxyURL, endpoint, ca string) error
	UpsertClusterFromRegister(*aegis.ClusterRegisterRequest)
	UpdateClusterFromHeartbeat(*aegis.ClusterHeartbeat)
	UpsertClusterImport(ClusterImport) error
	GetClusterProjectID(clusterID string) (string, bool)
	GetClusterInfo(clusterID string) *ClusterInfo
	ListClusterInfos() []*ClusterInfo
	ListClustersByProject(projectID string) []*ClusterInfo // Multi-tenancy: list clusters for a specific project
	SetClusterProjectID(clusterID, projectID string)
	SetClusterLabel(clusterID, key, value string)
	DeleteCluster(clusterID string)
	// CleanupStaleClusters soft-deletes clusters with heartbeats older than the threshold.
	// Returns the number of clusters cleaned up.
	CleanupStaleClusters(staleThreshold string) int64

	// TerminateWorkloadsByCluster terminates all non-terminal workloads on the
	// given cluster. Used for immediate cleanup when a cluster is deleted.
	TerminateWorkloadsByCluster(clusterID string) int64

	// TerminateOrphanedWorkloads terminates all non-terminal workloads whose
	// cluster has been soft-deleted. Returns the number of workloads terminated.
	TerminateOrphanedWorkloads() int64

	// TerminateStaleWorkloads terminates workloads stuck in RUNNING/PLACED
	// state longer than the given PostgreSQL interval threshold (e.g. "24 hours").
	// Returns the number of workloads terminated.
	TerminateStaleWorkloads(staleThreshold string) int64

	// provisioning logs
	AppendProvisioningLog(entry ProvisioningLogEntry)
	ListProvisioningLogs(jobID string, since time.Time, sinceSeq int64, limit int) []ProvisioningLogEntry
	ClearProvisioningLogs(jobID string)                // Clear logs when new provisioning starts
	DeleteOldProvisioningLogs(olderThan time.Time) int64 // Retention policy cleanup
	UpsertProvisioningRun(run ProvisioningRun)
	GetProvisioningRun(jobID string) (*ProvisioningRun, bool)
	ListProvisioningRuns(projectID string) []*ProvisioningRun // List runs for a project (or all if empty)

	// audit events
	PutAuditEvent(event *AuditEvent) error
	ListAuditEvents(filter AuditEventFilter) ([]*AuditEvent, error)
}
