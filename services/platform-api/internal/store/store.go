package store

import (
	"errors"
	"time"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

// ErrSessionNotFound is returned when a requested connection session cannot be located.
var ErrSessionNotFound = errors.New("connection session not found")

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
	VSCodeURI    string
	ExpiresAt    time.Time
	OneTime      bool
	Used         bool
	Revoked      bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Store defines the contract satisfied by all persistence backends used by the
// platform API. Implementations must provide their own concurrency controls and
// guarantee that multi-step operations documented as atomic remain so.
type Store interface {
	// catalog
	PutProject(*aegis.Project)
	GetProject(id string) *aegis.Project
	ListProjects() []*aegis.Project

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
	MarkPlaced(id string)
	GetPlacedAt(id string) (time.Time, bool)
	ClearPlacedAt(id string)
	MarkStarted(id string)
	GetStartedAt(id string) (time.Time, bool)
	SetEstimateUSD(id string, usd float64)
	PopEstimateUSD(id string) float64
	LeaseWorkloads(clusterID string, max int) []*aegis.Workload

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
	UpsertClusterFromRegister(*aegis.ClusterRegisterRequest)
	UpdateClusterFromHeartbeat(*aegis.ClusterHeartbeat)
	GetClusterInfo(clusterID string) *ClusterInfo
	ListClusterInfos() []*ClusterInfo
	SetClusterProjectID(clusterID, projectID string)
	DeleteCluster(clusterID string)
	// CleanupStaleClusters soft-deletes clusters with heartbeats older than the threshold.
	// Returns the number of clusters cleaned up.
	CleanupStaleClusters(staleThreshold string) int64
}
