# Compliance Implementation Context Document

This document contains all files needed for the AI Expert to understand the current state and implement the compliance feature.

---

## SECTION 1: FILES TO PROVIDE TO THE AI EXPERT (Current State)

---

### FILE: services/platform-api/internal/store/store.go
**Purpose:** The interface we must extend

```go
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
	// PreRegisterCluster creates a placeholder cluster row during provisioning,
	// before the k8s-agent connects. This ensures project_id and proxy_url are set
	// when the agent's RegisterCluster call updates the row.
	PreRegisterCluster(clusterID, projectID, provider, region, proxyURL string) error
	UpsertClusterFromRegister(*aegis.ClusterRegisterRequest)
	UpdateClusterFromHeartbeat(*aegis.ClusterHeartbeat)
	GetClusterInfo(clusterID string) *ClusterInfo
	ListClusterInfos() []*ClusterInfo
	ListClustersByProject(projectID string) []*ClusterInfo // Multi-tenancy: list clusters for a specific project
	SetClusterProjectID(clusterID, projectID string)
	DeleteCluster(clusterID string)
	// CleanupStaleClusters soft-deletes clusters with heartbeats older than the threshold.
	// Returns the number of clusters cleaned up.
	CleanupStaleClusters(staleThreshold string) int64

	// provisioning logs
	AppendProvisioningLog(entry ProvisioningLogEntry)
	ListProvisioningLogs(jobID string, since time.Time, sinceSeq int64, limit int) []ProvisioningLogEntry
	ClearProvisioningLogs(jobID string)                // Clear logs when new provisioning starts
	DeleteOldProvisioningLogs(olderThan time.Time) int64 // Retention policy cleanup
	UpsertProvisioningRun(run ProvisioningRun)
	GetProvisioningRun(jobID string) (*ProvisioningRun, bool)
	ListProvisioningRuns(projectID string) []*ProvisioningRun // List runs for a project (or all if empty)
}
```

---

### FILE: services/platform-api/api/v1alpha1/projectinfra_types.go
**Purpose:** The "Intent" source for compliance checks

```go
package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=pinfra

// ProjectInfra represents a declarative request to provision or import
// infrastructure for a specific Aegis project.
type ProjectInfra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectInfraSpec   `json:"spec,omitempty"`
	Status ProjectInfraStatus `json:"status,omitempty"`
}

// ProjectInfraSpec defines the desired state for project-scoped infrastructure.
type ProjectInfraSpec struct {
	ProjectID                 string                  `json:"projectId"`
	Provider                  string                  `json:"provider"`
	Region                    string                  `json:"region"`
	Strategy                  string                  `json:"strategy,omitempty"`
	ImportKubeconfigSecretRef *corev1.SecretReference `json:"importKubeconfigSecretRef,omitempty"`
	Aws                       *AWSInfraSpec           `json:"aws,omitempty"`
	Addons                    map[string]bool         `json:"addons,omitempty"`
	Labels                    map[string]string       `json:"labels,omitempty"`
}

// AWSProvisionMode declares how the automation runner should reconcile AWS resources.
type AWSProvisionMode string

const (
	AWSProvisionModeProvision AWSProvisionMode = "Provision"
	AWSProvisionModeImport    AWSProvisionMode = "Import"
)

// SecretKeyReference describes a namespaced secret key selector.
type SecretKeyReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Key       string `json:"key"`
}

// AWSImportSpec describes existing infrastructure to import into the management plane.
type AWSImportSpec struct {
	ClusterID           string             `json:"clusterId"`
	KubeconfigSecretRef SecretKeyReference `json:"kubeconfigSecretRef"`
}

// AWSInfraSpec captures AWS-specific provisioning parameters.
type AWSInfraSpec struct {
	AccountID          string           `json:"accountId,omitempty"`
	Mode               AWSProvisionMode `json:"mode,omitempty"`
	RoleARN            string           `json:"roleArn,omitempty"`
	ExternalID         string           `json:"externalId,omitempty"`
	SubnetIDs          []string         `json:"subnetIds,omitempty"`
	VpcID              string           `json:"vpcId,omitempty"`
	ClusterName        string           `json:"clusterName,omitempty"`
	Version            string           `json:"version,omitempty"`
	NodePools          []NodePool       `json:"nodePools,omitempty"`
	AdditionalClusters []AWSClusterSpec `json:"additionalClusters,omitempty"`
	Imports            []AWSImportSpec  `json:"imports,omitempty"`
}

// AWSClusterSpec describes a secondary cluster to provision in AWS.
type AWSClusterSpec struct {
	ClusterName string     `json:"clusterName"`
	Version     string     `json:"version,omitempty"`
	NodePools   []NodePool `json:"nodePools,omitempty"`
}

// NodePool represents a logical node group definition for a cluster.
type NodePool struct {
	Name         string            `json:"name"`
	InstanceType string            `json:"instanceType"`
	MinSize      int32             `json:"minSize"`
	MaxSize      int32             `json:"maxSize"`
	Labels       map[string]string `json:"labels,omitempty"`
	Taints       []corev1.Taint    `json:"taints,omitempty"`
}

// ProjectInfraStatus reports provisioning progress and outputs.
type ProjectInfraStatus struct {
	Phase              string             `json:"phase,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	Outputs            []ClusterOutput    `json:"outputs,omitempty"`
	CostHintUSDPerHour float64            `json:"costHintUsdPerHour,omitempty"`
	LastSyncTime       *metav1.Time       `json:"lastSyncTime,omitempty"`
}

// ClusterOutput captures key material for a provisioned cluster.
type ClusterOutput struct {
	ClusterID           string              `json:"clusterId"`
	Name                string              `json:"name"`
	Region              string              `json:"region"`
	KubeconfigSecretKey string              `json:"kubeconfigSecretKey"`
	Observability       ObservabilityOutput `json:"observability,omitempty"`
}

// ObservabilityOutput holds optional telemetry endpoints for a cluster.
type ObservabilityOutput struct {
	Namespace                string `json:"namespace,omitempty"`
	PrometheusService        string `json:"prometheusService,omitempty"`
	PrometheusPort           int32  `json:"prometheusPort,omitempty"`
	AlertmanagerService      string `json:"alertmanagerService,omitempty"`
	AlertmanagerPort         int32  `json:"alertmanagerPort,omitempty"`
	AlertmanagerConfigSecret string `json:"alertmanagerConfigSecret,omitempty"`
	MetricsServerService     string `json:"metricsServerService,omitempty"`
	MetricsServerPort        int32  `json:"metricsServerPort,omitempty"`
	LokiNamespace            string `json:"lokiNamespace,omitempty"`
	LokiService              string `json:"lokiService,omitempty"`
	LokiPort                 int32  `json:"lokiPort,omitempty"`
	LokiAuthSecret           string `json:"lokiAuthSecret,omitempty"`
	OtelEndpoint             string `json:"otelEndpoint,omitempty"`
	MetricsURL               string `json:"metricsUrl,omitempty"`
	TracingNamespace         string `json:"tracingNamespace,omitempty"`
	TempoService             string `json:"tempoService,omitempty"`
	TempoPort                int32  `json:"tempoPort,omitempty"`
	OtelService              string `json:"otelService,omitempty"`
	OtelGrpcPort             int32  `json:"otelGrpcPort,omitempty"`
	OtelHttpPort             int32  `json:"otelHttpPort,omitempty"`
}

// +kubebuilder:object:root=true

// ProjectInfraList contains a list of ProjectInfra objects.
type ProjectInfraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectInfra `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProjectInfra{}, &ProjectInfraList{})
}

// DeepCopy methods omitted for brevity - see original file for full implementation
```

---

### FILE: services/platform-api/api/v1alpha1/aegiscluster_types.go
**Purpose:** The "Target" resource we are validating

```go
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=aegisc

// AegisCluster mirrors the authoritative cluster state from the control plane store.
type AegisCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AegisClusterSpec   `json:"spec,omitempty"`
	Status AegisClusterStatus `json:"status,omitempty"`
}

// AegisClusterSpec holds immutable cluster metadata.
type AegisClusterSpec struct {
	ClusterID string `json:"clusterId"`
	ProjectID string `json:"projectId"`
	Provider  string `json:"provider"`
	Region    string `json:"region"`
}

// AegisClusterStatus reflects runtime signal sourced from the database.
type AegisClusterStatus struct {
	Phase         string             `json:"phase,omitempty"`
	Flavors       []string           `json:"flavors,omitempty"`
	Capacity      map[string]string  `json:"capacity,omitempty"`
	Conditions    []metav1.Condition `json:"conditions,omitempty"`
	LastHeartbeat *metav1.Time       `json:"lastHeartbeat,omitempty"`
}

// +kubebuilder:object:root=true

// AegisClusterList contains a list of AegisCluster objects.
type AegisClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AegisCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AegisCluster{}, &AegisClusterList{})
}

// DeepCopy methods omitted for brevity - see original file for full implementation
```

---

### FILE: proto/aegis/v1/platform.proto
**Purpose:** To understand the Project object for multi-tenancy

```protobuf
syntax = "proto3";
package aegis.v1;
option go_package = "github.com/yourorg/aegis/proto/aegis/v1;aegisv1";

import "google/api/annotations.proto";
import "google/protobuf/empty.proto";

// ---- Core resource messages ----
message Project {
  string id = 1;
  string display_name = 2;
  string owner_group = 3;
  PolicyDomain policy = 4;
  map<string, string> annotations = 5;
  ProjectAwsCredentials aws = 6;
}

message ProjectAwsCredentials {
  string account_id = 1;
  string role_arn = 2;
  string external_id = 3;
}
message PolicyDomain { repeated string regions = 1; string data_level = 2; bool deny_egress_by_default = 3; }

message Budget {
  string project_id = 1;
  string queue = 2;
  double limit_usd = 3;
  string policy_mode = 4;
}

message Flavor {
  string name = 1;
  string chip = 2;
  string mig_profile = 3;
  bool rdma_required = 4;
  string resource_name = 5;
  int32 gpu_count = 6;
  double memory_gib = 7;
  double price_usd_per_gpu_hour = 8;
  string cpu_cores_request = 9;
  string memory_request = 10;
}

message Queue {
  string name = 1; string project_id = 2; string priority_tier = 3;
  repeated string allowed_flavors = 4;
  int64 default_max_duration_seconds = 5;
}

message Workload {
  string id = 1; string project_id = 2; string queue = 3; string cluster_id = 4; string status = 5; string url = 6;
  oneof kind { WorkspaceSpec workspace = 10; TrainingSpec training = 11; }
  ResourceHints hints = 12;
  string ui_status = 13;
  string message = 14;
}

// Cluster handshake
message ClusterRegisterRequest {
  string cluster_id = 1;
  string provider = 2;
  string region = 3;
  string il_level = 4;
  map<string,string> labels = 5;
  string proxy_url = 6;
}

message ClusterHeartbeat {
  string cluster_id = 1;
  double ttf_gpu_seconds_p50 = 2;
  repeated Flavor available_flavors = 3;
  string proxy_url = 4;
}

// ... (additional messages and service definition - see full file)

service AegisPlatform {
  rpc CreateProject(CreateProjectRequest) returns (Project);
  rpc ListProjects(ListProjectsRequest) returns (ListProjectsResponse);
  rpc RegisterCluster(ClusterRegisterRequest) returns (ClusterRegisterResponse);
  rpc Heartbeat(ClusterHeartbeat) returns (ClusterHeartbeatAck);
  // ... additional RPCs
}
```

---

### FILE: services/platform-api/internal/controllers/aegiscluster_statussync.go
**Purpose:** CRITICAL - This is where the overwrite trap lives

```go
package controllers

import (
	"context"
	"sort"
	"time"

	"go.uber.org/zap"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

const (
	heartbeatGrace = 90 * time.Second
)

// AegisClusterStatusSync mirrors store state into the management-plane AegisCluster CRDs.
type AegisClusterStatusSync struct {
	client.Client
	Log       *zap.Logger
	Store     store.Store
	Namespace string
	Interval  time.Duration
}

// NeedLeaderElection indicates this runnable does not require leader election.
func (s *AegisClusterStatusSync) NeedLeaderElection() bool { return false }

// Start begins the periodic sync loop.
func (s *AegisClusterStatusSync) Start(ctx context.Context) error {
	interval := s.Interval
	if interval <= 0 {
		interval = 30 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	if err := s.sync(ctx); err != nil {
		s.logger().Warn("initial cluster sync failed", zap.Error(err))
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.sync(ctx); err != nil {
				s.logger().Warn("cluster sync failed", zap.Error(err))
			}
		}
	}
}

func (s *AegisClusterStatusSync) sync(ctx context.Context) error {
	if s.Store == nil {
		return nil
	}

	namespace := s.Namespace
	if namespace == "" {
		namespace = "aegis-system"
	}

	infos := s.Store.ListClusterInfos()
	existing := map[string]struct{}{}
	for _, info := range infos {
		name := info.ID
		existing[name] = struct{}{}
		if err := s.upsertCluster(ctx, namespace, info); err != nil {
			return err
		}
	}

	return s.gcOrphaned(ctx, namespace, existing)
}

func (s *AegisClusterStatusSync) upsertCluster(ctx context.Context, namespace string, info *store.ClusterInfo) error {
	cluster := &infraapi.AegisCluster{}
	key := types.NamespacedName{Name: info.ID, Namespace: namespace}
	err := s.Client.Get(ctx, key, cluster)
	if err != nil && !apierrors.IsNotFound(err) {
		return err
	}

	desiredSpec := infraapi.AegisClusterSpec{
		ClusterID: info.ID,
		ProjectID: info.Labels["aegis.yourorg.dev/projectId"],
		Provider:  info.Provider,
		Region:    info.Region,
	}

	phase, cond := clusterPhase(info)
	desiredStatus := infraapi.AegisClusterStatus{
		Phase:         phase,
		Flavors:       flattenFlavorSet(info.AvailableFlavorSet),
		Capacity:      nil,
		Conditions:    []metav1.Condition{cond},
		LastHeartbeat: cond.LastTransitionTime.DeepCopy(),
	}

	if apierrors.IsNotFound(err) {
		newCluster := &infraapi.AegisCluster{
			TypeMeta: metav1.TypeMeta{
				APIVersion: infraapi.GroupVersion.String(),
				Kind:       "AegisCluster",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      key.Name,
				Namespace: key.Namespace,
			},
			Spec:   desiredSpec,
			Status: desiredStatus,
		}
		if err := s.Client.Create(ctx, newCluster); err != nil {
			return err
		}
		return s.Client.Status().Update(ctx, newCluster)
	}

	specPatched := cluster.DeepCopy()
	specPatched.Spec = desiredSpec
	if err := s.Client.Patch(ctx, specPatched, client.MergeFrom(cluster)); err != nil {
		return err
	}
	cluster = specPatched

	statusPatched := cluster.DeepCopy()
	statusPatched.Status = desiredStatus
	return s.Client.Status().Patch(ctx, statusPatched, client.MergeFrom(cluster))
}

func (s *AegisClusterStatusSync) gcOrphaned(ctx context.Context, namespace string, active map[string]struct{}) error {
	var list infraapi.AegisClusterList
	if err := s.Client.List(ctx, &list, client.InNamespace(namespace)); err != nil {
		return err
	}
	for i := range list.Items {
		name := list.Items[i].Name
		if _, ok := active[name]; ok {
			continue
		}
		if err := s.Client.Delete(ctx, &list.Items[i]); err != nil && !apierrors.IsNotFound(err) {
			return err
		}
	}
	return nil
}

func clusterPhase(info *store.ClusterInfo) (string, metav1.Condition) {
	if info == nil {
		return "Pending", metav1.Condition{Type: "Heartbeat", Status: metav1.ConditionUnknown, Reason: "Unknown", LastTransitionTime: metav1.NewTime(time.Now())}
	}
	age := time.Since(info.LastHeartbeat)
	cond := metav1.Condition{Type: "Heartbeat", LastTransitionTime: metav1.NewTime(info.LastHeartbeat)}
	if info.LastHeartbeat.IsZero() {
		cond.Status = metav1.ConditionFalse
		cond.Reason = "NeverSeen"
		cond.Message = "cluster has not heartbeated yet"
		return "Pending", cond
	}
	if age <= heartbeatGrace {
		cond.Status = metav1.ConditionTrue
		cond.Reason = "HeartbeatOK"
		cond.Message = "cluster heartbeat within grace interval"
		return "HeartbeatOK", cond
	}
	cond.Status = metav1.ConditionFalse
	cond.Reason = "HeartbeatStale"
	cond.Message = "cluster heartbeat stale"
	return "Degraded", cond
}

func flattenFlavorSet(set map[string]bool) []string {
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for k, ok := range set {
		if ok {
			out = append(out, k)
		}
	}
	sort.Strings(out)
	return out
}

func (s *AegisClusterStatusSync) logger() *zap.Logger {
	if s.Log != nil {
		return s.Log
	}
	return zap.NewNop()
}

var _ manager.LeaderElectionRunnable = (*AegisClusterStatusSync)(nil)
```

---

### FILE: services/platform-api/internal/controllers/projectinfra_controller.go
**Purpose:** Reference for how provisioning triggers

```go
package controllers

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/provisioning"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

const (
	projectInfraFinalizer = "infra.aegis.yourorg.dev/finalizer"
	conditionReady        = "Ready"
)

// ProjectInfraReconciler handles ProjectInfra lifecycle orchestration.
type ProjectInfraReconciler struct {
	client.Client
	APIReader                 client.Reader
	Scheme                    *runtime.Scheme
	Log                       *zap.Logger
	Provisioner               provisioning.AWSProvisioner
	Store                     store.Store
	KubeconfigSecretName      string
	KubeconfigSecretNamespace string
	LocalLocks                map[string]*sync.Mutex
	LocalLocksMu              sync.Mutex
	HolderIdentity            string
}

// Reconcile implements the controller-runtime reconciliation contract.
func (r *ProjectInfraReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.logger().With(zap.String("projectinfra", req.NamespacedName.String()))

	var infra infraapi.ProjectInfra
	if err := r.Get(ctx, req.NamespacedName, &infra); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	releaseLock, res, err := r.acquireStackLock(ctx, &infra)
	if err != nil {
		return res, err
	}
	if res.Requeue || res.RequeueAfter > 0 {
		if releaseLock != nil {
			releaseLock()
		}
		return res, nil
	}
	if releaseLock != nil {
		defer releaseLock()
	}

	// Re-read the object after acquiring the lock
	reader := r.APIReader
	if reader == nil {
		reader = r.Client
	}
	if err := reader.Get(ctx, req.NamespacedName, &infra); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !infra.DeletionTimestamp.IsZero() {
		return r.reconcileDelete(ctx, log, &infra)
	}

	if err := r.ensureFinalizer(ctx, &infra); err != nil {
		return ctrl.Result{}, err
	}

	return r.reconcileNormal(ctx, log, &infra)
}

func (r *ProjectInfraReconciler) reconcileNormal(ctx context.Context, log *zap.Logger, infra *infraapi.ProjectInfra) (ctrl.Result, error) {
	r.syncProjectToStore(infra)

	if strings.EqualFold(infra.Status.Phase, "Ready") {
		log.Info("skipping reconcile; infrastructure is already ready")
		return ctrl.Result{}, nil
	}

	if strings.EqualFold(infra.Status.Phase, "Error") {
		log.Info("skipping reconcile; infrastructure is in error state")
		return ctrl.Result{}, nil
	}

	if infra.Spec.Aws != nil {
		if strings.EqualFold(string(infra.Spec.Aws.Mode), string(infraapi.AWSProvisionModeImport)) {
			return r.handleAWSImport(ctx, log, infra)
		}
	}
	if infra.Spec.ImportKubeconfigSecretRef != nil {
		return r.handleImport(ctx, log, infra)
	}
	if infra.Spec.Aws != nil {
		return r.handleAWSProvision(ctx, log, infra)
	}
	reason := "InvalidSpec"
	err := fmt.Errorf("spec.aws or spec.importKubeconfigSecretRef required")
	cond := newCondition(metav1.ConditionFalse, reason, err.Error())
	_ = r.setStatus(ctx, infra, "Error", &cond, nil, infra.Status.CostHintUSDPerHour)
	return ctrl.Result{}, err
}

// ... (additional methods - see full file for complete implementation)

func (r *ProjectInfraReconciler) ensureAegisCluster(ctx context.Context, infra *infraapi.ProjectInfra, output infraapi.ClusterOutput) error {
	clusterID := strings.TrimSpace(output.ClusterID)
	if clusterID == "" {
		return fmt.Errorf("cluster ID required")
	}

	projectID := strings.TrimSpace(infra.Spec.ProjectID)
	provider := strings.TrimSpace(infra.Spec.Provider)
	region := strings.TrimSpace(output.Region)

	// Pre-register the cluster in the store
	if r.Store != nil && projectID != "" {
		proxyURL := ""
		if host := strings.TrimSpace(os.Getenv("AEGIS_SPOKE_PROXY_HOST")); host != "" {
			if !strings.HasPrefix(host, "wss://") && !strings.HasPrefix(host, "ws://") {
				proxyURL = "wss://" + host
			} else {
				proxyURL = host
			}
		}
		if err := r.Store.PreRegisterCluster(clusterID, projectID, provider, region, proxyURL); err != nil {
			r.Log.Warn("failed to pre-register cluster in store",
				zap.String("cluster_id", clusterID),
				zap.String("project_id", projectID),
				zap.Error(err))
		}
	}

	// ... rest of implementation
	return nil
}
```

---

### FILE: services/platform-api/internal/store/postgres/store.go
**Purpose:** Base Postgres implementation

```go
package postgres

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

const defaultQueryTimeout = 5 * time.Second

// PostgresStore implements the store.Store interface backed by PostgreSQL.
type PostgresStore struct {
	pool    *pgxpool.Pool
	log     *zap.Logger
	timeout time.Duration
}

var _ store.Store = (*PostgresStore)(nil)

// New establishes a connection pool using the provided DSN and applies optional
// tuning sourced from environment variables (PG_MAX_OPEN_CONNS, PG_MAX_IDLE_CONNS,
// PG_CONN_MAX_LIFETIME, PG_DEFAULT_QUERY_TIMEOUT).
func New(dsn string, log *zap.Logger) (*PostgresStore, error) {
	if dsn == "" {
		return nil, fmt.Errorf("postgres store requires non-empty DSN")
	}
	if log == nil {
		log = zap.NewNop()
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	if maxConns := envInt("PG_MAX_OPEN_CONNS", 0); maxConns > 0 {
		cfg.MaxConns = maxConns
	}
	if minConns := envInt("PG_MAX_IDLE_CONNS", 0); minConns > 0 {
		cfg.MinConns = minConns
	}
	if lifetime := envDuration("PG_CONN_MAX_LIFETIME", 0); lifetime > 0 {
		cfg.MaxConnLifetime = lifetime
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	timeout := envDuration("PG_DEFAULT_QUERY_TIMEOUT", defaultQueryTimeout)
	if timeout <= 0 {
		timeout = defaultQueryTimeout
	}

	return &PostgresStore{pool: pool, log: log, timeout: timeout}, nil
}

// Close releases all pooled connections.
func (s *PostgresStore) Close() {
	if s == nil {
		return
	}
	s.pool.Close()
}

func (s *PostgresStore) logExecError(action string, err error, fields ...zap.Field) {
	if err == nil {
		return
	}
	if s.log == nil {
		return
	}
	fields = append(fields, zap.Error(err))
	s.log.Warn("postgres action failed", append([]zap.Field{zap.String("action", action)}, fields...)...)
}

func (s *PostgresStore) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, s.timeout)
}

func envInt(key string, def int32) int32 {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 32); err == nil {
			return int32(parsed)
		}
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	if t.IsZero() {
		return nil
	}
	return t.UTC()
}
```

---

### FILE: charts/aegis-services/files/platform-api/migrations/000001_init.up.sql
**Purpose:** Example of existing Helm migration path

```sql
CREATE TABLE IF NOT EXISTS projects (
    id TEXT PRIMARY KEY,
    display_name TEXT NULL,
    owner_group TEXT NOT NULL,
    policy_regions TEXT[] NULL,
    policy_data_level TEXT NULL,
    policy_deny_egress_by_default BOOLEAN NOT NULL DEFAULT FALSE,
    annotations JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS flavors (
    name TEXT PRIMARY KEY,
    chip TEXT NOT NULL,
    mig_profile TEXT NULL,
    rdma_required BOOLEAN NOT NULL DEFAULT FALSE,
    gpu_count INTEGER NOT NULL DEFAULT 0,
    memory_gib DOUBLE PRECISION NULL,
    resource_name TEXT NULL,
    cpu_cores_request TEXT NULL,
    memory_request TEXT NULL,
    price_usd_per_gpu_hour DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS queues (
    name TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    priority_tier TEXT NULL,
    allowed_flavors TEXT[] NULL,
    default_max_duration_seconds BIGINT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS queues_by_project ON queues(project_id);

CREATE TABLE IF NOT EXISTS budgets (
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    queue TEXT NOT NULL DEFAULT '',
    limit_usd DOUBLE PRECISION NOT NULL,
    policy_mode TEXT NOT NULL CHECK (policy_mode IN ('HARD', 'SOFT')),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, queue)
);

CREATE TABLE IF NOT EXISTS workloads (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    queue TEXT NOT NULL DEFAULT '',
    cluster_id TEXT NULL,
    status TEXT NOT NULL,
    ui_status TEXT NULL,
    url TEXT NULL,
    message TEXT NULL,
    kind TEXT NOT NULL,
    hints_resource_name TEXT NULL,
    hints_gpu_count INTEGER NULL,
    hints_cpu_request TEXT NULL,
    hints_mem_request TEXT NULL,
    workspace_json JSONB NULL,
    training_json JSONB NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    placed_at TIMESTAMPTZ NULL,
    started_at TIMESTAMPTZ NULL
);
CREATE INDEX IF NOT EXISTS workloads_by_project ON workloads(project_id);
CREATE INDEX IF NOT EXISTS workloads_by_cluster_status ON workloads(cluster_id, status);

CREATE TABLE IF NOT EXISTS workload_estimates (
    workload_id TEXT PRIMARY KEY REFERENCES workloads(id) ON DELETE CASCADE,
    estimate_usd DOUBLE PRECISION NOT NULL
);

CREATE TABLE IF NOT EXISTS budget_usage (
    project_id TEXT NOT NULL,
    queue TEXT NOT NULL DEFAULT '',
    period_start_utc DATE NOT NULL,
    reserved_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    actual_usd DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (project_id, queue, period_start_utc),
    FOREIGN KEY (project_id, queue) REFERENCES budgets(project_id, queue) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS budget_usage_lookup ON budget_usage(project_id, queue);

CREATE TABLE IF NOT EXISTS connection_sessions (
    session_id TEXT PRIMARY KEY,
    workload_id TEXT NOT NULL REFERENCES workloads(id) ON DELETE CASCADE,
    subject TEXT NOT NULL,
    client TEXT NOT NULL,
    jti TEXT NOT NULL UNIQUE,
    token TEXT NOT NULL,
    ssh_user TEXT NOT NULL,
    ssh_host_alias TEXT NOT NULL,
    internal_host TEXT NOT NULL,
    port INTEGER NOT NULL,
    ssh_config TEXT NOT NULL,
    proxy_url TEXT NOT NULL,
    vscode_uri TEXT NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    one_time BOOLEAN NOT NULL DEFAULT TRUE,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    revoked BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS sessions_by_workload ON connection_sessions(workload_id);
CREATE INDEX IF NOT EXISTS sessions_by_expires ON connection_sessions(expires_at);

CREATE TABLE IF NOT EXISTS session_jtis (
    jti TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES connection_sessions(session_id) ON DELETE CASCADE,
    used BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS clusters (
    id TEXT PRIMARY KEY,
    provider TEXT NULL,
    region TEXT NULL,
    ttf_gpu_seconds_p50 DOUBLE PRECISION NOT NULL DEFAULT 0,
    last_heartbeat TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cluster_labels (
    cluster_id TEXT NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    k TEXT NOT NULL,
    v TEXT NOT NULL,
    PRIMARY KEY (cluster_id, k)
);

CREATE TABLE IF NOT EXISTS cluster_flavors (
    cluster_id TEXT NOT NULL REFERENCES clusters(id) ON DELETE CASCADE,
    flavor TEXT NOT NULL,
    PRIMARY KEY (cluster_id, flavor)
);
```

---

### FILE: terraform/generate-cloud-deployment.sh
**Purpose:** To fix the Terraform migration gap (Note: File is 961 lines - key sections shown)

```bash
#!/bin/bash
# Generate Helm values and optionally deploy to Kubernetes
#
# For complete deployment documentation, see: DEPLOYMENT_AND_TESTING_GUIDE.md

set -e

# ... (configuration variables and helper functions)

# Step 3: Run database migrations manually (to avoid public image pull issues)
echo ""
echo "3️⃣  Running database migrations..."

# ... (DB configuration)

# Run migrations using an in-cluster Job so RDS schema exists before tests
MIGRATION_CONFIGMAP="${HELM_RELEASE}-migrations"
MIGRATION_JOB="${HELM_RELEASE}-migrate"
MIGRATIONS_DIR="${SCRIPT_DIR}/../services/platform-api/migrations"

if [ ! -f "${MIGRATIONS_DIR}/0001_init.sql" ]; then
  echo "❌ Migration file not found at ${MIGRATIONS_DIR}/0001_init.sql"
  exit 1
fi
MIGRATIONS_UP_FILE=$(mktemp)
awk '/^--[[:space:]]+\+migrate[[:space:]]+Down/{exit} {print}' "${MIGRATIONS_DIR}/0001_init.sql" > "${MIGRATIONS_UP_FILE}"

kubectl -n "${K8S_NAMESPACE}" create configmap "${MIGRATION_CONFIGMAP}" \
  --from-file=0001_init.sql="${MIGRATIONS_UP_FILE}" \
  --dry-run=client -o yaml | kubectl apply -f -

# ... (Job creation and execution)

# NOTE: New migration files (like 000007_compliance_schema.up.sql) need to be added
# to the MIGRATIONS_DIR and included in the migration job execution
```

---

## SECTION 2: TARGET FILE LIST (Files to Create or Modify)

---

### NEW FILES TO CREATE:

1. **charts/aegis-services/files/platform-api/migrations/000007_compliance_schema.up.sql**
   - The SQL Schema for compliance tables

2. **charts/aegis-services/files/platform-api/migrations/000007_compliance_schema.down.sql**
   - Rollback migration for compliance schema

3. **services/platform-api/internal/compliance/engine.go**
   - The Compliance Interface definition

4. **services/platform-api/internal/compliance/checks.go**
   - The Logic/Rules for compliance checks

5. **services/platform-api/internal/store/postgres/compliance.go**
   - The DB implementation for compliance store methods

6. **services/platform-api/internal/controllers/compliance_controller.go**
   - The new compliance runner controller (if choosing Option A)

---

### EXISTING FILES TO MODIFY:

1. **services/platform-api/internal/controllers/aegiscluster_statussync.go**
   - Fix the status overwrite issue
   - Preserve compliance-related status fields during sync

2. **services/platform-api/internal/store/store.go**
   - Add `GetComplianceProfile` method
   - Add `SaveComplianceReport` method
   - Add any other compliance-related store interface methods

3. **terraform/generate-cloud-deployment.sh**
   - Add the new migration file (000007_compliance_schema.up.sql) to the cloud deploy list
   - Ensure migration job includes all migration files

---

## SECTION 3: ADDITIONAL FILES FROM ARCHITECTURE REPORT (Spoke/K8s-Agent & Supporting Files)

---

### FILE: agents/k8s-agent/api/v1alpha1/aegisworkload_types.go
**Purpose:** AegisWorkload CRD definition - workload execution on spoke clusters

```go
/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkspaceSpec defines properties for interactive workloads.
type WorkspaceSpec struct {
	// Flavor is the GPU flavor requested for the workspace pods.
	// +kubebuilder:validation:Optional
	Flavor string `json:"flavor,omitempty"`

	// Image contains the container image to execute.
	// +kubebuilder:validation:Optional
	Image string `json:"image,omitempty"`

	// Env holds environment variables to project into the container.
	// +kubebuilder:validation:Optional
	Env map[string]string `json:"env,omitempty"`

	// Command overrides the default entrypoint.
	// +kubebuilder:validation:Optional
	Command []string `json:"command,omitempty"`

	// Interactive indicates the workspace expects interactive access (SSH, VS Code, etc).
	// +kubebuilder:validation:Optional
	Interactive bool `json:"interactive,omitempty"`

	// Ports exposes additional container ports for interactive scenarios (defaults to 11111 when empty).
	// +kubebuilder:validation:Optional
	Ports []int32 `json:"ports,omitempty"`
}

// TrainingSpec captures distributed training configuration details.
type TrainingSpec struct {
	// Flavor identifies the GPU flavor for each training worker.
	// +kubebuilder:validation:Optional
	Flavor string `json:"flavor,omitempty"`

	// Workers represents the number of replicas participating in the training job.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	Workers int32 `json:"workers,omitempty"`

	// GpusPerWorker specifies GPUs to request per replica.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	GpusPerWorker int32 `json:"gpusPerWorker,omitempty"`

	// Image used for the training containers.
	// +kubebuilder:validation:Optional
	Image string `json:"image,omitempty"`

	// Command overrides the default entrypoint for training containers.
	// +kubebuilder:validation:Optional
	Command []string `json:"command,omitempty"`

	// Gang requests gang scheduling semantics when supported.
	// +kubebuilder:validation:Optional
	Gang bool `json:"gang,omitempty"`
}

// ResourceHints reflects control-plane scheduling hints.
type ResourceHints struct {
	// ResourceName is the GPU resource alias to request (e.g. nvidia.com/mig-1g.10gb).
	// +kubebuilder:validation:Optional
	ResourceName string `json:"resourceName,omitempty"`

	// GpuCount is the number of GPUs to request per container.
	// +kubebuilder:validation:Optional
	GpuCount int32 `json:"gpuCount,omitempty"`

	// CpuCoresRequest is the CPU core quantity to request (e.g. "1", "500m").
	// +kubebuilder:validation:Optional
	CpuCoresRequest *string `json:"cpuCoresRequest,omitempty"`

	// MemoryRequest is the memory quantity to request (e.g. "4Gi", "1024Mi").
	// +kubebuilder:validation:Optional
	MemoryRequest *string `json:"memoryRequest,omitempty"`
}

// AegisWorkloadSpec defines the desired state of AegisWorkload.
// +kubebuilder:validation:XValidation:rule="has(self.workspace) || has(self.training)",message="exactly one of workspace or training must be specified"
// +kubebuilder:validation:XValidation:rule="!(has(self.workspace) && has(self.training))",message="exactly one of workspace or training must be specified"
type AegisWorkloadSpec struct {
	// ProjectID associates the workload with an Aegis project.
	// +kubebuilder:validation:Required
	ProjectID string `json:"projectId"`

	// Queue optionally specifies the scheduling queue.
	// +kubebuilder:validation:Optional
	Queue string `json:"queue,omitempty"`

	// Workspace holds the interactive workload specification.
	// +kubebuilder:validation:Optional
	Workspace *WorkspaceSpec `json:"workspace,omitempty"`

	// Training holds the distributed training specification.
	// +kubebuilder:validation:Optional
	Training *TrainingSpec `json:"training,omitempty"`

	// Hints carries scheduling hints propagated from the control plane.
	// +kubebuilder:validation:Optional
	Hints *ResourceHints `json:"hints,omitempty"`
}

// JobRef references the workload realization created by the operator.
type JobRef struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Name       string `json:"name,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
}

// AegisWorkloadPhase describes the high-level lifecycle state of a workload.
type AegisWorkloadPhase string

const (
	// PhasePending denotes that the workload has been accepted but not yet submitted to Kubernetes.
	PhasePending AegisWorkloadPhase = "Pending"
	// PhaseSubmitted indicates the operator has created the concrete Job/CRD.
	PhaseSubmitted AegisWorkloadPhase = "Submitted"
	// PhaseAdmitted reflects external admission (e.g. Kueue) allowing execution.
	PhaseAdmitted AegisWorkloadPhase = "Admitted"
	// PhaseRunning represents actively executing workloads.
	PhaseRunning AegisWorkloadPhase = "Running"
	// PhaseSucceeded marks workloads that completed successfully.
	PhaseSucceeded AegisWorkloadPhase = "Succeeded"
	// PhaseFailed marks workloads that terminated unsuccessfully.
	PhaseFailed AegisWorkloadPhase = "Failed"
)

// AegisWorkloadStatus defines the observed state of AegisWorkload.
type AegisWorkloadStatus struct {
	// Phase is the coarse-grained lifecycle state.
	// +kubebuilder:validation:Optional
	Phase AegisWorkloadPhase `json:"phase,omitempty"`

	// Message provides additional diagnostic information.
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`

	// Backend records which executor handled the workload (workspace, pytorch, etc).
	// +kubebuilder:validation:Optional
	Backend string `json:"backend,omitempty"`

	// URL references the running workload (e.g. k8s:// namespace/object).
	// +kubebuilder:validation:Optional
	URL string `json:"url,omitempty"`

	// JobRef captures the Kubernetes object created by the operator.
	// +kubebuilder:validation:Optional
	JobRef *JobRef `json:"jobRef,omitempty"`

	// StartTime indicates when execution began.
	// +kubebuilder:validation:Optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime indicates when execution finished.
	// +kubebuilder:validation:Optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Conditions provide detailed status information.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=awl
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Backend",type=string,JSONPath=`.status.backend`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// AegisWorkload is the Schema for the aegisworkloads API.
type AegisWorkload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AegisWorkloadSpec   `json:"spec,omitempty"`
	Status AegisWorkloadStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AegisWorkloadList contains a list of AegisWorkload.
type AegisWorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AegisWorkload `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AegisWorkload{}, &AegisWorkloadList{})
}
```

---

### FILE: agents/k8s-agent/api/v1alpha2/workspace_types.go
**Purpose:** Workspace CRD definition (v1alpha2) - higher-level workspace abstraction

```go
/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkspacePhase captures the coarse lifecycle for Workspace resources.
type WorkspacePhase string

const (
	// WorkspacePhasePending indicates the Workspace is being prepared.
	WorkspacePhasePending WorkspacePhase = "Pending"
	// WorkspacePhaseSubmitted indicates execution resources have been created.
	WorkspacePhaseSubmitted WorkspacePhase = "Submitted"
	// WorkspacePhaseAdmitted indicates the Workspace has been admitted for execution.
	WorkspacePhaseAdmitted WorkspacePhase = "Admitted"
	// WorkspacePhaseRunning indicates the associated execution is active.
	WorkspacePhaseRunning WorkspacePhase = "Running"
	// WorkspacePhaseSucceeded indicates execution completed successfully.
	WorkspacePhaseSucceeded WorkspacePhase = "Succeeded"
	// WorkspacePhaseFailed indicates execution terminated unsuccessfully.
	WorkspacePhaseFailed WorkspacePhase = "Failed"
)

// WorkspaceSpec defines the desired state of a Workspace.
type WorkspaceSpec struct {
	// ProjectRef ties the Workspace to an Aegis project identifier.
	// +kubebuilder:validation:Required
	ProjectRef string `json:"projectRef"`

	// Queue optionally specifies the scheduling queue.
	// +kubebuilder:validation:Optional
	Queue string `json:"queue,omitempty"`

	// Persona captures persona metadata that may drive defaults.
	// +kubebuilder:validation:Optional
	Persona string `json:"persona,omitempty"`

	// ProfileRef references a WorkspaceClass providing defaults and admission rules.
	// +kubebuilder:validation:Required
	ProfileRef string `json:"profileRef"`

	// BackendRef points to a provider-specific backend configuration.
	// +kubebuilder:validation:Optional
	BackendRef *corev1.TypedLocalObjectReference `json:"backendRef,omitempty"`

	// GPUProfile describes GPU flavor selection and hints for execution.
	// +kubebuilder:validation:Optional
	GPUProfile *WorkspaceGPUProfileSpec `json:"gpuProfile,omitempty"`

	// Execution specifies the container image, command, and runtime parameters.
	// +kubebuilder:validation:Optional
	Execution *WorkspaceExecution `json:"execution,omitempty"`

	// Network captures desired isolation and egress guardrails.
	// +kubebuilder:validation:Optional
	Network *WorkspaceNetworkSpec `json:"network,omitempty"`

	// Storage describes persistence requirements for the workspace.
	// +kubebuilder:validation:Optional
	Storage *WorkspaceStorageSpec `json:"storage,omitempty"`

	// MaxDurationSeconds defines a hard cap on runtime duration.
	// +kubebuilder:validation:Optional
	MaxDurationSeconds *int64 `json:"maxDurationSeconds,omitempty"`

	// TTLSecondsAfterFinished defines when to garbage collect execution artifacts.
	// +kubebuilder:validation:Optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`
}

// WorkspaceExecution captures runtime intent for the interactive workload.
type WorkspaceExecution struct {
	// Image is the container image to launch for the workspace.
	// +kubebuilder:validation:Optional
	Image string `json:"image,omitempty"`

	// Env defines environment variables for the workspace container.
	// +kubebuilder:validation:Optional
	Env map[string]string `json:"env,omitempty"`

	// Command overrides the container entrypoint.
	// +kubebuilder:validation:Optional
	Command []string `json:"command,omitempty"`

	// Interactive indicates whether SSH and interactive services should be exposed.
	// +kubebuilder:validation:Optional
	Interactive bool `json:"interactive,omitempty"`

	// Ports enumerates additional service ports to expose.
	// +kubebuilder:validation:Optional
	Ports []int32 `json:"ports,omitempty"`
}

// WorkspaceGPUProfileSpec declares desired GPU flavor and optional hints.
type WorkspaceGPUProfileSpec struct {
	// Flavor names the GPU flavor or profile.
	// +kubebuilder:validation:Optional
	Flavor string `json:"flavor,omitempty"`

	// Hints provide supplemental resource hints derived from the control plane.
	// +kubebuilder:validation:Optional
	Hints *ResourceHints `json:"hints,omitempty"`
}

// ResourceHints describes control plane derived resource hints.
type ResourceHints struct {
	// ResourceName corresponds to the GPU resource alias (e.g. nvidia.com/mig-1g.10gb).
	// +kubebuilder:validation:Optional
	ResourceName string `json:"resourceName,omitempty"`

	// GPUCount expresses the number of GPUs to request.
	// +kubebuilder:validation:Optional
	GPUCount int32 `json:"gpuCount,omitempty"`

	// CpuCoresRequest expresses CPU request quantity (e.g. "2", "500m").
	// +kubebuilder:validation:Optional
	CpuCoresRequest *string `json:"cpuCoresRequest,omitempty"`

	// MemoryRequest expresses memory request quantity (e.g. "16Gi").
	// +kubebuilder:validation:Optional
	MemoryRequest *string `json:"memoryRequest,omitempty"`
}

// WorkspaceNetworkSpec models desired network settings for a Workspace.
type WorkspaceNetworkSpec struct {
	// IsolationLevel specifies the isolation class (e.g. IL2, IL4).
	// +kubebuilder:validation:Optional
	IsolationLevel string `json:"isolationLevel,omitempty"`

	// Egress defines default egress behavior (Allow, DenyByDefault, etc).
	// +kubebuilder:validation:Optional
	Egress string `json:"egress,omitempty"`

	// AdditionalCIDRs enumerates CIDRs that should be reachable from the workspace.
	// +kubebuilder:validation:Optional
	AdditionalCIDRs []string `json:"additionalCIDRs,omitempty"`
}

// WorkspaceStorageSpec captures persistence configuration.
type WorkspaceStorageSpec struct {
	// Mode identifies persistent or ephemeral storage.
	// +kubebuilder:validation:Optional
	Mode string `json:"mode,omitempty"`

	// PVCTemplate provides a template used to create a PersistentVolumeClaim.
	// +kubebuilder:validation:Optional
	PVCTemplate *corev1.PersistentVolumeClaimSpec `json:"pvcTemplate,omitempty"`
}

// ConnectionInfo captures connection material surfaced to the user.
type ConnectionInfo struct {
	// ProxyURL exposes the reverse proxy endpoint for browser access.
	// +kubebuilder:validation:Optional
	ProxyURL string `json:"proxyURL,omitempty"`

	// SSHHostAlias captures the SSH host alias configured for bootstrap.
	// +kubebuilder:validation:Optional
	SSHHostAlias string `json:"sshHostAlias,omitempty"`
}

// CostStatus captures cost estimation and accrual details.
type CostStatus struct {
	// AccruedUSD is the total accrued cost in USD.
	// +kubebuilder:validation:Optional
	AccruedUSD string `json:"accruedUSD,omitempty"`

	// EstHourlyRateUSD captures the estimated hourly burn rate in USD.
	// +kubebuilder:validation:Optional
	EstHourlyRateUSD string `json:"estHourlyRateUSD,omitempty"`
}

// WorkspaceStatus defines the observed state of a Workspace.
type WorkspaceStatus struct {
	// Phase reflects the coarse lifecycle state rolled up from owned resources.
	// +kubebuilder:validation:Optional
	Phase WorkspacePhase `json:"phase,omitempty"`

	// Backend records the execution backend reported by the active workload.
	// +kubebuilder:validation:Optional
	Backend string `json:"backend,omitempty"`

	// URL references the execution object or interactive endpoint.
	// +kubebuilder:validation:Optional
	URL string `json:"url,omitempty"`

	// Cost surfaces estimation and accrual signals.
	// +kubebuilder:validation:Optional
	Cost *CostStatus `json:"cost,omitempty"`

	// Connection contains access material for interactive experiences.
	// +kubebuilder:validation:Optional
	Connection *ConnectionInfo `json:"connection,omitempty"`

	// Conditions provides granular readiness and policy state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// WorkloadRef references the current realization (e.g. AegisWorkload).
	// +kubebuilder:validation:Optional
	WorkloadRef *corev1.ObjectReference `json:"workloadRef,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ws
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Queue",type=string,JSONPath=`.spec.queue`,priority=1
// +kubebuilder:printcolumn:name="Backend",type=string,JSONPath=`.status.backend`,priority=1
// Workspace is the Schema for the workspaces API.
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceSpec   `json:"spec,omitempty"`
	Status WorkspaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceList contains a list of Workspace.
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}
```

---

### FILE: agents/k8s-agent/internal/controller/aegisworkload_controller.go
**Purpose:** AegisWorkload controller - reconciles workloads on spoke clusters (1304 lines - key sections shown)

```go
package controller

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"go.uber.org/zap"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	unstructured "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/ptr"

	aegisproto "github.com/yourorg/aegis/proto/aegis/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	aegisv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/cpclient"
	builders "github.com/yourorg/aegis/agents/k8s-agent/internal/workload/builders"
	workdiscovery "github.com/yourorg/aegis/agents/k8s-agent/internal/workload/discovery"
	workstatus "github.com/yourorg/aegis/agents/k8s-agent/internal/workload/status"
	workspacecfg "github.com/yourorg/aegis/pkg/workspace"
)

const (
	backendWorkspace = "workspace"
	backendPyTorch   = "pytorch_v1"
	backendJob       = "job_v1"

	requeuePending               = 10 * time.Second
	workspaceJobTTLSeconds int32 = 600

	labelWorkloadID = "aegis.workload/id"
	labelSSHManaged = "aegis.yourorg.dev/ssh-managed"

	// Event-driven state tracking annotations
	annotationLastPushedState  = "aegis.yourorg.dev/last-pushed-state"
	annotationLastPushedJobUID = "aegis.yourorg.dev/last-pushed-job-uid"
)

// AegisWorkloadReconciler reconciles a AegisWorkload object.
type AegisWorkloadReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder

	Dynamic dynamic.Interface
	Config  *rest.Config

	cpClient  *cpclient.Client
	clusterID string

	defaultWorkspaceImage string
	defaultTrainingImage  string
	gpuResourceOverride   string
	dryRun                bool
	kueueEnabled          bool
	kueueQueue            string
	proxyServiceName      string
	proxyServicePort      int32
	proxyIngressHost      string
	proxyURL              string
	sshBootstrapImage     string

	workspaceEnvDefaults map[string]string

	pyTorchAPIVersion string
	pyTorchGVR        *schema.GroupVersionResource
}

// Reconcile ensures the workload CR reflects the lifecycle of the underlying compute object.
func (r *AegisWorkloadReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithValues("aegisworkload", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	var aw aegisv1alpha1.AegisWorkload
	if err := r.Get(ctx, req.NamespacedName, &aw); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !aw.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	if aw.Status.Phase == "" {
		if err := r.patchStatus(ctx, &aw, func(st *aegisv1alpha1.AegisWorkloadStatus) {
			st.Phase = aegisv1alpha1.PhasePending
		}); err != nil {
			return ctrl.Result{}, err
		}
	}

	switch {
	case aw.Spec.Workspace != nil:
		return r.reconcileWorkspace(ctx, &aw)
	case aw.Spec.Training != nil:
		return r.reconcileTraining(ctx, &aw)
	default:
		log.Info("no workspace or training spec; skipping")
		return ctrl.Result{}, nil
	}
}

func (r *AegisWorkloadReconciler) startClusterPresence(ctx context.Context) {
	if r.cpClient == nil || r.clusterID == "" {
		return
	}
	// ... cluster registration and heartbeat loop
}

// startWorkloadBridge notifies the control plane when a workload starts running
func (r *AegisWorkloadReconciler) startWorkloadBridge(ctx context.Context, aw *aegisv1alpha1.AegisWorkload, jobUID string) {
	if r.cpClient == nil || aw == nil {
		return
	}
	// ... bridge to control plane
}

// ackWorkloadBridge notifies the control plane of workload completion
func (r *AegisWorkloadReconciler) ackWorkloadBridge(ctx context.Context, aw *aegisv1alpha1.AegisWorkload, status string, jobUID string) {
	if r.cpClient == nil || aw == nil {
		return
	}
	// ... bridge to control plane
}

// SetupWithManager sets up the controller with the Manager.
func (r *AegisWorkloadReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// ... setup logic including cpClient initialization from AEGIS_CP_GRPC
	return ctrl.NewControllerManagedBy(mgr).
		For(&aegisv1alpha1.AegisWorkload{}).
		Owns(&batchv1.Job{}).
		Named("aegisworkload").
		Complete(r)
}
```

---

### FILE: agents/k8s-agent/internal/controller/workspace_controller.go
**Purpose:** Workspace controller (v1alpha2) - orchestrates Workspace resources

```go
/*
Copyright 2025.
*/

package controller

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	aegisv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
	aegisv1alpha2 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha2"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/cpclient"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/providers"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/workspace/metrics"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/workspace/plan"
	aegisproto "github.com/yourorg/aegis/proto/aegis/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	workspaceControllerName = "workspace"
	requeueWorkspace        = 10 * time.Second
)

// WorkspaceReconciler orchestrates Workspace resources and their realizations.
type WorkspaceReconciler struct {
	client.Client
	Scheme           *runtime.Scheme
	planner          *plan.Planner
	recorder         recordEventRecorder
	metricsCollector *metrics.Collector
	providers        *providers.Registry
	cpClient         *cpclient.Client
	clusterID        string
}

// Reconcile handles Workspace state transitions.
func (r *WorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.LoggerFrom(ctx).WithValues("workspace", req.NamespacedName)
	ctx = ctrl.LoggerInto(ctx, log)

	if r.planner == nil {
		r.planner = plan.NewPlanner()
	}

	var ws aegisv1alpha2.Workspace
	if err := r.Get(ctx, req.NamespacedName, &ws); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !ws.DeletionTimestamp.IsZero() {
		return ctrl.Result{}, nil
	}

	planResult, err := r.planner.Plan(&ws)
	if err != nil {
		log.Error(err, "failed to compute workspace plan")
		// ... error handling
		return ctrl.Result{}, nil
	}

	workload, err := r.ensureWorkload(ctx, &ws, planResult)
	if err != nil {
		return ctrl.Result{}, err
	}

	if err := r.rollupStatus(ctx, &ws, workload, nil); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: requeueWorkspace}, nil
}

func (r *WorkspaceReconciler) ensureWorkload(ctx context.Context, ws *aegisv1alpha2.Workspace, planResult *plan.Result) (*aegisv1alpha1.AegisWorkload, error) {
	child := &aegisv1alpha1.AegisWorkload{ObjectMeta: metav1.ObjectMeta{Name: planResult.WorkloadName, Namespace: ws.Namespace}}

	op, err := controllerutil.CreateOrUpdate(ctx, r.Client, child, func() error {
		if err := controllerutil.SetControllerReference(ws, child, r.Scheme); err != nil {
			return err
		}
		child.Spec = planResult.WorkloadSpec
		// ... labels and annotations
		return nil
	})
	if err != nil {
		return nil, err
	}

	if r.recorder != nil && op == controllerutil.OperationResultCreated {
		r.recorder.Eventf(ws, corev1.EventTypeNormal, "Submitted", "Workspace workload %s created", planResult.WorkloadName)
	}

	// ... control plane registration
	return child, nil
}

// SetupWithManager wires the controller into the manager.
func (r *WorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// ... setup including cpClient from AEGIS_CP_GRPC
	return ctrl.NewControllerManagedBy(mgr).
		Named(workspaceControllerName).
		For(&aegisv1alpha2.Workspace{}).
		Owns(&aegisv1alpha1.AegisWorkload{}).
		Complete(r)
}
```

---

### FILE: agents/k8s-agent/internal/cpclient/client.go
**Purpose:** Control plane gRPC client - agent-to-platform communication

```go
package cpclient

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

type Client struct {
	api  aegis.AegisPlatformClient
	conn *grpc.ClientConn
}

func New(endpoint string) (*Client, error) {
	var opts []grpc.DialOption

	tlsEnabled := enableTLS()
	if tlsEnabled {
		creds, err := buildTLSCredentials(endpoint)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithTransportCredentials(creds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// OIDC client credentials support
	if oidcCfg, err := loadOIDCConfigFromEnv(); err != nil {
		return nil, err
	} else if oidcCfg != nil {
		tokenSource, err := buildOIDCTokenSource(oidcCfg)
		if err != nil {
			return nil, err
		}
		opts = append(opts, grpc.WithPerRPCCredentials(&bearerTokenCredentials{
			source:     oauth2.ReuseTokenSource(nil, tokenSource),
			requireTLS: tlsEnabled,
		}))
	}

	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		api:  aegis.NewAegisPlatformClient(conn),
		conn: conn,
	}, nil
}

func (c *Client) Register(ctx context.Context, req *aegis.ClusterRegisterRequest) error {
	_, err := c.api.RegisterCluster(ctx, req)
	return err
}

func (c *Client) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// HeartbeatLoop continuously reports cluster health, advertised flavors, and spoke proxy URL.
func (c *Client) HeartbeatLoop(ctx context.Context, logger *zap.Logger, clusterID string, flavors []*aegis.Flavor, proxyURL string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := c.api.Heartbeat(ctx, &aegis.ClusterHeartbeat{
				ClusterId:        clusterID,
				TtfGpuSecondsP50: 60.0,
				AvailableFlavors: flavors,
				ProxyUrl:         proxyURL,
			}); err != nil {
				logger.Warn("heartbeat failed", zap.Error(err))
			}
		}
	}
}

func (c *Client) Lease(ctx context.Context, clusterID string, max int32) ([]*aegis.Workload, error) {
	resp, err := c.api.LeaseWorkload(ctx, &aegis.LeaseWorkloadRequest{ClusterId: clusterID, Max: max})
	if err != nil {
		return nil, err
	}
	return resp.GetItems(), nil
}

func (c *Client) Ack(ctx context.Context, id, status, backend, url string) error {
	_, err := c.api.AckWorkload(ctx, &aegis.AckWorkloadRequest{Id: id, Status: status, Backend: backend, Url: url})
	return err
}

func (c *Client) Start(ctx context.Context, id, clusterID string) error {
	_, err := c.api.StartWorkload(ctx, &aegis.StartWorkloadRequest{Id: id, ClusterId: clusterID})
	return err
}

func (c *Client) SubmitWorkload(ctx context.Context, workload *aegis.Workload) (*aegis.Workload, error) {
	return c.api.SubmitWorkload(ctx, &aegis.SubmitWorkloadRequest{Workload: workload})
}

func (c *Client) GetWorkload(ctx context.Context, id string) (*aegis.Workload, error) {
	return c.api.GetWorkload(ctx, &aegis.GetWorkloadRequest{Id: id})
}

// OIDC configuration loaded from environment variables:
// - AEGIS_CP_OIDC_TOKEN_URL
// - AEGIS_CP_OIDC_CLIENT_ID
// - AEGIS_CP_OIDC_CLIENT_SECRET
// - AEGIS_CP_OIDC_AUDIENCE
```

---

### FILE: services/platform-api/internal/store/postgres/catalog.go
**Purpose:** Project/catalog persistence - where projects.annotations is stored (JSONB)

```go
package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

// Annotation key validation constants
const (
	annotationPrefix = "aegis.yourorg.dev/"
)

// knownAnnotationKeys maps valid camelCase annotation keys to their purpose.
var knownAnnotationKeys = map[string]string{
	"aegis.yourorg.dev/awsRoleArn":    "AWS IAM role ARN for assuming cross-account access",
	"aegis.yourorg.dev/awsAccountId":  "AWS account ID for the project",
	"aegis.yourorg.dev/awsExternalId": "External ID for STS AssumeRole",
	"aegis.yourorg.dev/projectId":     "Project ID for cluster association",
	"aegis.yourorg.dev/ilLevel":       "Information Level (IL) classification",
	"aegis.yourorg.dev/environment":   "Environment (dev, staging, prod)",
	"aegis.yourorg.dev/description":   "Human-readable description",
	"aegis.yourorg.dev/enable_fips":   "Enable FIPS 140-2 compliance",
	// Compliance policy annotations (recommended for per-project policy):
	// "aegis.yourorg.dev/complianceProfile": "fedramp_high|fedramp_moderate|basic_dev"
	// "aegis.yourorg.dev/complianceEnforcement": "alert|block|off"
}

func (s *PostgresStore) PutProject(p *aegis.Project) {
	if p == nil || strings.TrimSpace(p.GetId()) == "" {
		return
	}
	policy := p.GetPolicy()
	regions := []string(nil)
	dataLevel := ""
	denyEgress := false
	if policy != nil {
		regions = append(regions, policy.GetRegions()...)
		dataLevel = policy.GetDataLevel()
		denyEgress = policy.GetDenyEgressByDefault()
	}

	// Normalize annotation keys (fix kebab-case to camelCase)
	annotations := normalizeAnnotations(p.GetAnnotations(), s.log, p.GetId())

	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	_, err := s.pool.Exec(ctx, `
INSERT INTO projects (id, display_name, owner_group, policy_regions, policy_data_level, policy_deny_egress_by_default, annotations, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
ON CONFLICT (id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    owner_group = EXCLUDED.owner_group,
    policy_regions = EXCLUDED.policy_regions,
    policy_data_level = EXCLUDED.policy_data_level,
    policy_deny_egress_by_default = EXCLUDED.policy_deny_egress_by_default,
    annotations = EXCLUDED.annotations,
    updated_at = now()
`, p.GetId(), nullableString(p.GetDisplayName()), p.GetOwnerGroup(), regions, nullableString(dataLevel), denyEgress, mapToJSONB(annotations))
	s.logExecError("upsert_project", err, zap.String("project_id", p.GetId()))
}

func (s *PostgresStore) GetProject(id string) *aegis.Project {
	// ... retrieves project including annotations JSONB
	// annotations stored in projects.annotations column
}

func (s *PostgresStore) ListProjects() []*aegis.Project {
	// ... lists all projects with annotations
}

// DeleteProject removes a project from the database.
// Returns an error if the project has active (non-deleted) clusters attached.
func (s *PostgresStore) DeleteProject(id string) error {
	// ... checks for active clusters before deleting
}

// HasActiveClusters checks if a project has any non-deleted clusters attached.
func (s *PostgresStore) HasActiveClusters(projectID string) bool {
	// ... count query on clusters table
}

func mapToJSONB(m map[string]string) []byte {
	if len(m) == 0 {
		return nil
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return data
}
```

---

### FILE: services/platform-api/internal/server/wizard_handlers.go
**Purpose:** Workspace creation endpoint - enforcement point for compliance blocking

```go
package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	workspacecfg "github.com/yourorg/aegis/pkg/workspace"
	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

type workspaceCreateRequest struct {
	ProjectID   string            `json:"projectId"`
	ClusterID   string            `json:"clusterId"`
	Name        string            `json:"name"`
	Profile     string            `json:"profile"`
	Template    string            `json:"template"`
	Parameters  map[string]string `json:"parameters"`
	RequestedBy string            `json:"requestedBy"`
	Flavor      string            `json:"flavor"`
	Queue       string            `json:"queue"`
	Image       string            `json:"image"`
}

func registerWorkspaceWizardRoutes(mux *runtime.ServeMux, srv *Server) {
	// Registers: /api/projects, /api/clusters, /api/workspaces
	// Also with /aegis prefix for Backstage proxy compatibility
}

func (s *Server) handleCreateWorkspace(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req workspaceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWizardError(w, status.Errorf(codes.InvalidArgument, "invalid request body: %v", err))
		return
	}

	projectID := strings.TrimSpace(req.ProjectID)
	clusterID := strings.TrimSpace(req.ClusterID)
	name := strings.TrimSpace(req.Name)

	// Validate project exists
	project := s.store.GetProject(projectID)
	if project == nil {
		writeWizardError(w, status.Error(codes.NotFound, "project not found"))
		return
	}

	// Authorization check
	if err := s.authorize(ctx, projectID, "", "createWorkspace"); err != nil {
		writeWizardError(w, err)
		return
	}

	// Validate cluster is ready
	clusterInfo := s.clusterInfo(clusterID)
	if clusterInfo == nil {
		writeWizardError(w, status.Errorf(codes.NotFound, "cluster %q not registered", clusterID))
		return
	}
	if !clusterReady(clusterInfo, time.Now()) {
		writeWizardError(w, status.Errorf(codes.FailedPrecondition, "cluster %s not ready", clusterID))
		return
	}

	// TODO: COMPLIANCE ENFORCEMENT POINT
	// If project policy is "block" and compliance check fails:
	// writeWizardError(w, status.Error(codes.FailedPrecondition, "cluster non-compliant"))
	// return

	// Build and submit workload
	workload := &aegis.Workload{
		Id:        buildWorkspaceID(projectID, name),
		ProjectId: projectID,
		Queue:     queue,
		ClusterId: clusterID,
		Kind: &aegis.Workload_Workspace{
			Workspace: &aegis.WorkspaceSpec{
				Flavor:      flavor,
				Image:       image,
				Env:         env,
				Interactive: true,
			},
		},
	}

	res, err := s.SubmitWorkload(ctx, &aegis.SubmitWorkloadRequest{Workload: workload})
	if err != nil {
		writeWizardError(w, err)
		return
	}

	// Return response
	writeJSON(w, http.StatusCreated, workspaceCreateResponse{
		ID:        res.GetId(),
		ProjectID: projectID,
		ClusterID: res.GetClusterId(),
		Name:      name,
		Status:    "pending",
	})
}
```

---

### FILE: charts/aegis-services/values/cloud.yaml
**Purpose:** Cloud-specific Helm values - shows migrations.enabled=false for cloud

```yaml
# Cloud production values for aegis-services
# This configuration is optimized for production-like environments

platformApi:
  enabled: true
  replicaCount: 1
  serviceAccountName: aegis-platform-api

  image:
    repository: 567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/platform-api
    tag: "8afb9e0-vscode-dynamic"
    pullPolicy: IfNotPresent

  env:
    LOG_LEVEL: "info"
    ENVIRONMENT: "production"
    AEGIS_STORE_BACKEND: "postgres"
    AEGIS_OBSERVABILITY_ENABLED: "true"
    DB_HOST: "aegis-prod-rds.cluster-xxxx.us-east-1.rds.amazonaws.com"
    DB_PORT: "5432"
    DB_NAME: "aegis"
    DB_USER: "aegis_api"
    DB_SSLMODE: "require"

  # IMPORTANT: Migrations disabled for cloud - Terraform handles migrations
  migrations:
    enabled: false
    image:
      repository: ghcr.io/golang-migrate/migrate
      tag: v4.16.2
      pullPolicy: IfNotPresent
    ttlSecondsAfterFinished: 300
    args:
      - "-path=/migrations"
      - "-database=$(DATABASE_URL)"
      - "up"

proxy:
  enabled: true
  replicaCount: 1

keycloak:
  enabled: true
```

---

### FILE: charts/aegis-services/templates/platform-api-migrations-configmap.yaml
**Purpose:** Helm migration ConfigMap template - mounts files/platform-api/migrations/*

```yaml
{{- $m := .Values.platformApi.migrations -}}
{{- if and .Values.platformApi.enabled $m $m.enabled }}
apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ include "aegis-services.platformApi.fullname" . }}-migrations
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "aegis-services.labels" . | nindent 4 }}
    app.kubernetes.io/component: platform-api
    aegis.yourorg.dev/component: migrations
  annotations:
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/hook-delete-policy": before-hook-creation
    "helm.sh/hook-weight": "-5"
    aegis.yourorg.dev/migrations-version: "{{ default "unknown" .Chart.AppVersion }}"
data:
{{- $files := .Files.Glob "files/platform-api/migrations/*" }}
{{- if not $files }}
  0001_placeholder.sql: |
    -- No migrations bundled with this chart.
{{- else }}
{{- range $path, $file := $files }}
  {{ base $path }}: |
{{ $.Files.Get $path | indent 4 }}
{{- end }}
{{- end }}
{{- end }}
```

---

### FILE: charts/aegis-services/templates/platform-api-migrations-job.yaml
**Purpose:** Helm migration Job template - runs golang-migrate with mounted migrations

```yaml
{{- $m := .Values.platformApi.migrations -}}
{{- if and .Values.platformApi.enabled $m $m.enabled }}
apiVersion: batch/v1
kind: Job
metadata:
  name: {{ include "aegis-services.platformApi.fullname" . }}-migrate
  namespace: {{ .Release.Namespace }}
  labels:
    {{- include "aegis-services.labels" . | nindent 4 }}
    app.kubernetes.io/component: platform-api
    aegis.yourorg.dev/job: migrate
  annotations:
    "helm.sh/hook": post-install,post-upgrade
    "helm.sh/hook-delete-policy": before-hook-creation,hook-succeeded
    "helm.sh/hook-weight": "0"
spec:
  {{- with $m.ttlSecondsAfterFinished }}
  ttlSecondsAfterFinished: {{ . }}
  {{- end }}
  backoffLimit: 5
  template:
    metadata:
      labels:
        {{- include "aegis-services.platformApi.selectorLabels" . | nindent 8 }}
        aegis.yourorg.dev/job: migrate
    spec:
      restartPolicy: OnFailure
      initContainers:
        - name: wait-for-postgres
          image: busybox:1.36
          command:
            - sh
            - -c
            - |
              echo "Waiting for postgres at $DB_HOST:$DB_PORT..."
              until nc -z -w5 "$DB_HOST" "$DB_PORT"; do
                echo "Postgres not ready, waiting..."
                sleep 2
              done
              echo "Postgres is ready!"
          env:
            - name: DB_HOST
              value: {{ .Values.platformApi.env.DB_HOST | default "platform-postgres" | quote }}
            - name: DB_PORT
              value: {{ .Values.platformApi.env.DB_PORT | default "5432" | quote }}
      containers:
        - name: migrate
          image: "{{ $m.image.repository }}:{{ $m.image.tag }}"
          imagePullPolicy: {{ $m.image.pullPolicy | default "IfNotPresent" }}
          args:
            {{- range $m.args }}
            - {{ . | quote }}
            {{- end }}
          env:
            {{- range $key, $value := .Values.platformApi.env }}
            - name: {{ $key }}
              value: {{ $value | quote }}
            {{- end }}
            - name: DATABASE_URL
              value: "postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)"
          volumeMounts:
            - name: migrations
              mountPath: /migrations
      volumes:
        - name: migrations
          configMap:
            name: {{ include "aegis-services.platformApi.fullname" . }}-migrations
{{- end }}
```

---

## END OF DOCUMENT
