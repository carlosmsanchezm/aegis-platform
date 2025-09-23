package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	aegisv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	"github.com/yourorg/aegis/services/platform-api/internal/kubeclients"
	"github.com/yourorg/aegis/services/platform-api/internal/placement"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Server struct {
	aegis.UnimplementedAegisPlatformServer
	log             *zap.Logger
	store           *store.MemStore
	kubeClients     *kubeclients.Manager
	targetNamespace string
}

const (
	statusPlaced  = "PLACED"
	statusRunning = "RUNNING"

	heartbeatTTL = 45 * time.Second
)

var (
	mPlaced = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "aegis_workload_placed_total", Help: "Workloads placed"},
		[]string{"flavor"},
	)
	mLeased = promauto.NewCounter(
		prometheus.CounterOpts{Name: "aegis_workload_leased_total", Help: "Workloads leased"},
	)
	mAcked = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "aegis_workload_acked_total", Help: "Workloads acked by final status"},
		[]string{"status", "backend"},
	)
	mQueueWait = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "aegis_workload_queue_wait_seconds",
			Help:    "Time between placement and lease per queue/flavor",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"queue", "flavor"},
	)
	mBudgetActual = promauto.NewGaugeVec(
		prometheus.GaugeOpts{Name: "aegis_budget_actual_usd", Help: "Actual spend in USD for current UTC month"},
		[]string{"project", "queue"},
	)
	mBudgetReserved = promauto.NewGaugeVec(
		prometheus.GaugeOpts{Name: "aegis_budget_reserved_usd", Help: "Reserved (estimated) USD for current UTC month"},
		[]string{"project", "queue"},
	)
	mBudgetDenied = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "aegis_budget_denied_total", Help: "Budget denials at submit"},
		[]string{"project", "queue", "reason"},
	)
	mBudgetOverrun = promauto.NewCounterVec(
		prometheus.CounterOpts{Name: "aegis_budget_overrun_total", Help: "Budget soft-policy overruns"},
		[]string{"project", "queue"},
	)
	mEstimateHist = promauto.NewHistogram(
		prometheus.HistogramOpts{Name: "aegis_workload_estimated_cost_usd", Help: "Estimated workload USD", Buckets: prometheus.ExponentialBuckets(0.01, 2, 15)},
	)
)

func New(log *zap.Logger, st *store.MemStore, clients *kubeclients.Manager, namespace string) *Server {
	if namespace == "" {
		namespace = "default"
	}
	return &Server{log: log, store: st, kubeClients: clients, targetNamespace: namespace}
}

func getEnvInt(key string, def int64) int64 {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

func (s *Server) defaultMaxRuntimeSeconds(q *aegis.Queue) int64 {
	if q != nil && q.GetDefaultMaxDurationSeconds() > 0 {
		return q.GetDefaultMaxDurationSeconds()
	}
	return getEnvInt("AEGIS_DEFAULT_MAX_RUNTIME_SECONDS", 3600)
}

func (s *Server) CreateProject(ctx context.Context, req *aegis.CreateProjectRequest) (*aegis.Project, error) {
	if req == nil || req.Project == nil {
		err := status.Error(codes.InvalidArgument, "project payload required")
		s.log.Warn("create project failed", zap.Error(err))
		return nil, err
	}
	p := req.Project
	s.store.PutProject(p)
	s.log.Info("project upserted", zap.String("project_id", p.GetId()), zap.String("owner_group", p.GetOwnerGroup()))
	return p, nil
}

func (s *Server) UpsertBudget(ctx context.Context, req *aegis.UpsertBudgetRequest) (*aegis.Budget, error) {
	if req == nil || req.Budget == nil {
		err := status.Error(codes.InvalidArgument, "budget payload required")
		s.log.Warn("upsert budget failed", zap.Error(err))
		return nil, err
	}
	b := req.Budget
	s.store.PutBudget(b)
	s.log.Info("budget upserted",
		zap.String("project_id", b.GetProjectId()),
		zap.String("queue", b.GetQueue()),
		zap.Float64("limit_usd", b.GetLimitUsd()),
		zap.String("policy_mode", b.GetPolicyMode()),
	)
	return b, nil
}

func (s *Server) UpsertFlavor(ctx context.Context, req *aegis.UpsertFlavorRequest) (*aegis.Flavor, error) {
	if req == nil || req.Flavor == nil {
		err := status.Error(codes.InvalidArgument, "flavor payload required")
		s.log.Warn("upsert flavor failed", zap.Error(err))
		return nil, err
	}
	f := req.Flavor
	s.store.PutFlavor(f)
	s.log.Info("flavor upserted", zap.String("flavor", f.GetName()), zap.String("chip", f.GetChip()), zap.String("mig_profile", f.GetMigProfile()), zap.Bool("rdma_required", f.GetRdmaRequired()))
	return f, nil
}

func (s *Server) UpsertQueue(ctx context.Context, req *aegis.UpsertQueueRequest) (*aegis.Queue, error) {
	if req == nil || req.Queue == nil {
		err := status.Error(codes.InvalidArgument, "queue payload required")
		s.log.Warn("upsert queue failed", zap.Error(err))
		return nil, err
	}
	q := req.Queue
	s.store.PutQueue(q)
	s.log.Info("queue upserted", zap.String("queue", q.GetName()), zap.String("project_id", q.GetProjectId()), zap.String("priority_tier", q.GetPriorityTier()), zap.Int("allowed_flavors", len(q.GetAllowedFlavors())))
	return q, nil
}

func (s *Server) RegisterCluster(ctx context.Context, r *aegis.ClusterRegisterRequest) (*aegis.ClusterRegisterResponse, error) {
	if r == nil {
		err := status.Error(codes.InvalidArgument, "cluster payload required")
		s.log.Warn("register cluster failed", zap.Error(err))
		return nil, err
	}
	s.store.UpsertClusterFromRegister(r)
	s.log.Info("cluster registered", zap.String("cluster_id", r.GetClusterId()), zap.String("provider", r.GetProvider()), zap.String("region", r.GetRegion()), zap.Int("label_count", len(r.GetLabels())))
	return &aegis.ClusterRegisterResponse{Ok: true, Message: "registered"}, nil
}

func (s *Server) Heartbeat(ctx context.Context, hb *aegis.ClusterHeartbeat) (*aegis.ClusterHeartbeatAck, error) {
	if hb == nil {
		err := status.Error(codes.InvalidArgument, "heartbeat payload required")
		s.log.Warn("cluster heartbeat failed", zap.Error(err))
		return nil, err
	}
	flavorNames := make([]string, 0, len(hb.GetAvailableFlavors()))
	for _, f := range hb.GetAvailableFlavors() {
		flavorNames = append(flavorNames, f.GetName())
	}
	s.store.UpdateClusterFromHeartbeat(hb)
	s.log.Debug("cluster heartbeat",
		zap.String("cluster_id", hb.GetClusterId()),
		zap.Float64("ttf_gpu_seconds_p50", hb.GetTtfGpuSecondsP50()),
		zap.Strings("available_flavors", flavorNames),
	)
	return &aegis.ClusterHeartbeatAck{Ok: true}, nil
}

func (s *Server) SubmitWorkload(ctx context.Context, req *aegis.SubmitWorkloadRequest) (*aegis.Workload, error) {
	if req == nil || req.Workload == nil {
		err := status.Error(codes.InvalidArgument, "workload payload required")
		s.log.Warn("submit workload failed", zap.Error(err))
		return nil, err
	}
	w := req.Workload
	if w.ProjectId == "" {
		return nil, status.Error(codes.InvalidArgument, "projectId required on workload")
	}
	if w.Id == "" {
		w.Id = "w-" + RandID()
	}
	p := s.store.GetProject(w.ProjectId)
	if p == nil {
		return nil, status.Errorf(codes.InvalidArgument, "unknown project %q", w.ProjectId)
	}
	reqFlavor, err := requiredFlavor(w)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	flavorObj := s.store.GetFlavor(reqFlavor)
	if flavorObj == nil {
		s.log.Warn("submit workload rejected; unknown flavor",
			zap.String("workload_id", w.GetId()),
			zap.String("flavor", reqFlavor),
		)
		return nil, status.Error(codes.FailedPrecondition, "unknown flavor: "+reqFlavor)
	}
	queueObj := s.store.GetQueue(w.GetQueue())

	estimateUSD, maxSecs, estErr := s.estimateWorkloadUSD(w, flavorObj, queueObj)
	if estErr != nil {
		s.log.Debug("workload cost estimate fallback",
			zap.String("workload_id", w.GetId()),
			zap.Error(estErr),
		)
	} else {
		s.log.Debug("workload cost estimate",
			zap.String("workload_id", w.GetId()),
			zap.Float64("estimate_usd", estimateUSD),
			zap.Int64("max_runtime_secs", maxSecs),
		)
	}
	mEstimateHist.Observe(estimateUSD)
	allowed, policy, reason, usage := s.store.ReserveIfAllowed(w.GetProjectId(), w.GetQueue(), estimateUSD)
	if !allowed {
		if reason == "" {
			reason = "budget_denied"
		}
		mBudgetDenied.WithLabelValues(w.GetProjectId(), w.GetQueue(), reason).Inc()
		s.log.Warn("budget reservation denied",
			zap.String("workload_id", w.GetId()),
			zap.String("project_id", w.GetProjectId()),
			zap.String("queue", w.GetQueue()),
			zap.Float64("estimate_usd", estimateUSD),
			zap.String("reason", reason),
		)
		return nil, status.Error(codes.FailedPrecondition, "budget exceeded: "+reason)
	}
	if policy != "" {
		if flavorObj.GetPriceUsdPerGpuHour() <= 0 {
			mBudgetDenied.WithLabelValues(w.GetProjectId(), w.GetQueue(), "missing_price").Inc()
			s.log.Warn("budget reservation denied: missing price",
				zap.String("workload_id", w.GetId()),
				zap.String("project_id", w.GetProjectId()),
				zap.String("queue", w.GetQueue()),
				zap.String("flavor", reqFlavor),
			)
			return nil, status.Error(codes.FailedPrecondition, "missing flavor pricing for budgeted project")
		}
		b := s.store.GetBudgetExact(w.GetProjectId(), w.GetQueue())
		if b == nil {
			b = s.store.GetBudgetExact(w.GetProjectId(), "")
		}
		if b != nil && strings.EqualFold(policy, "SOFT") && (usage.ActualUSD+usage.ReservedUSD) > b.GetLimitUsd() {
			mBudgetOverrun.WithLabelValues(w.GetProjectId(), w.GetQueue()).Inc()
			s.log.Warn("budget soft overrun",
				zap.String("workload_id", w.GetId()),
				zap.String("project_id", w.GetProjectId()),
				zap.String("queue", w.GetQueue()),
				zap.Float64("estimate_usd", estimateUSD),
				zap.Float64("limit_usd", b.GetLimitUsd()),
			)
		}
		mBudgetReserved.WithLabelValues(w.GetProjectId(), w.GetQueue()).Set(usage.ReservedUSD)
		mBudgetActual.WithLabelValues(w.GetProjectId(), w.GetQueue()).Set(usage.ActualUSD)
	}
	if estimateUSD > 0 {
		s.store.SetEstimateUSD(w.GetId(), estimateUSD)
	}

	// Build candidates from current cluster snapshots
	infos := s.store.ListClusterInfos()
	cands := make([]placement.Candidate, 0, len(infos))
	now := time.Now()
	for _, ci := range infos {
		if ci.LastHeartbeat.IsZero() || now.Sub(ci.LastHeartbeat) > heartbeatTTL {
			s.log.Debug("skipping stale cluster",
				zap.String("cluster_id", ci.ID),
				zap.Time("last_heartbeat", ci.LastHeartbeat),
			)
			continue
		}
		cands = append(cands, placement.Candidate{
			ClusterID:   ci.ID,
			Region:      ci.Region,
			Labels:      ci.Labels,
			TTFGSeconds: ci.TTFGSecondsP50,
			Flavors:     ci.AvailableFlavorSet,
		})
	}

	var regions []string
	if p.GetPolicy() != nil {
		regions = p.GetPolicy().GetRegions()
	}
	pd := placement.PolicyDomain{Regions: regions}
	placementFlavor := reqFlavor
	if flavorObj.GetGpuCount() == 0 && flavorObj.GetResourceName() == "" {
		placementFlavor = ""
	}
	chosen, perr := placement.ChooseCluster(cands, pd, placementFlavor)
	if perr != nil {
		s.log.Warn("placement failed",
			zap.String("workload_id", w.GetId()),
			zap.String("project_id", w.GetProjectId()),
			zap.String("flavor", reqFlavor),
			zap.Strings("regions", regions),
			zap.Int("candidate_count", len(cands)),
			zap.Error(perr),
		)
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("no eligible cluster for flavor=%s in policy regions=%v", reqFlavor, regions))
	}

	w.ClusterId = chosen
	w.Status = statusPlaced

	if fl := s.store.GetFlavor(reqFlavor); fl != nil {
		w.Hints = &aegis.ResourceHints{
			ResourceName: fl.GetResourceName(),
			GpuCount:     fl.GetGpuCount(),
		}
	}

	if s.kubeClients == nil {
		err := status.Error(codes.Internal, "kubernetes client manager not configured")
		s.log.Error("submit workload failed", zap.Error(err))
		return nil, err
	}

	kubeClient, err := s.kubeClients.ClientFor(chosen)
	if err != nil {
		s.log.Error("submit workload failed: kube client", zap.String("cluster_id", chosen), zap.Error(err))
		return nil, status.Error(codes.Internal, "failed to resolve cluster client")
	}

	cr := buildAegisWorkloadCR(w, s.targetNamespace)
	if err := kubeClient.Create(ctx, cr); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			s.log.Error("failed to create aegis workload CR",
				zap.String("workload_id", w.GetId()),
				zap.String("cluster_id", chosen),
				zap.String("namespace", s.targetNamespace),
				zap.Error(err),
			)
			return nil, status.Error(codes.Internal, "failed to create AegisWorkload in target cluster")
		}
		s.log.Info("aegis workload CR already exists",
			zap.String("workload_id", w.GetId()),
			zap.String("cluster_id", chosen),
			zap.String("namespace", s.targetNamespace),
		)
	} else {
		s.log.Info("aegis workload CR created",
			zap.String("workload_id", w.GetId()),
			zap.String("cluster_id", chosen),
			zap.String("namespace", s.targetNamespace),
		)
	}

	s.store.PutWorkload(w)
	s.store.MarkPlaced(w.GetId())
	mPlaced.WithLabelValues(reqFlavor).Inc()
	s.log.Info("workload placed",
		zap.String("workload_id", w.GetId()),
		zap.String("project_id", w.GetProjectId()),
		zap.String("flavor", reqFlavor),
		zap.String("cluster_id", w.GetClusterId()),
		zap.String("status", w.GetStatus()),
		zap.String("namespace", s.targetNamespace),
	)
	return w, nil
}

func (s *Server) GetWorkload(ctx context.Context, req *aegis.GetWorkloadRequest) (*aegis.Workload, error) {
	if req == nil || req.Id == "" {
		err := status.Error(codes.InvalidArgument, "workload id required")
		s.log.Warn("get workload failed", zap.Error(err))
		return nil, err
	}
	w := s.store.GetWorkload(req.Id)
	if w == nil {
		err := status.Error(codes.NotFound, "workload not found")
		s.log.Warn("workload not found", zap.String("workload_id", req.GetId()))
		return nil, err
	}
	s.log.Info("workload retrieved", zap.String("workload_id", w.GetId()), zap.String("status", w.GetStatus()), zap.String("project_id", w.GetProjectId()))
	return w, nil
}

func (s *Server) ListWorkloads(ctx context.Context, req *aegis.ListWorkloadsRequest) (*aegis.ListWorkloadsResponse, error) {
	if req == nil || req.ProjectId == "" {
		err := status.Error(codes.InvalidArgument, "project id required")
		s.log.Warn("list workloads failed", zap.Error(err))
		return nil, err
	}
	items := s.store.ListWorkloads(req.ProjectId)
	s.log.Info("workloads listed", zap.String("project_id", req.GetProjectId()), zap.Int("count", len(items)))
	return &aegis.ListWorkloadsResponse{Items: items}, nil
}

func (s *Server) LeaseWorkload(ctx context.Context, req *aegis.LeaseWorkloadRequest) (*aegis.LeaseWorkloadResponse, error) {
	if req == nil || req.GetClusterId() == "" {
		err := status.Error(codes.InvalidArgument, "cluster_id required")
		s.log.Warn("lease workload failed", zap.Error(err))
		return nil, err
	}
	if req.GetMax() <= 0 {
		err := status.Error(codes.InvalidArgument, "max must be > 0")
		s.log.Warn("lease workload failed", zap.Error(err), zap.Int32("max", req.GetMax()))
		return nil, err
	}
	items := s.store.LeaseWorkloads(req.GetClusterId(), int(req.GetMax()))
	now := time.Now()
	for _, w := range items {
		queue := w.GetQueue()
		if queue == "" {
			queue = "unknown"
		}
		flavor, ferr := requiredFlavor(w)
		if ferr != nil {
			s.log.Warn("leased workload missing flavor",
				zap.String("workload_id", w.GetId()),
				zap.Error(ferr),
			)
			continue
		}
		if t0, ok := s.store.GetPlacedAt(w.GetId()); ok {
			wait := now.Sub(t0)
			mQueueWait.WithLabelValues(queue, flavor).Observe(wait.Seconds())
			s.store.ClearPlacedAt(w.GetId())
			s.log.Debug("queue wait observed",
				zap.String("workload_id", w.GetId()),
				zap.String("queue", queue),
				zap.String("flavor", flavor),
				zap.Duration("wait", wait),
			)
		}
		if fl := s.store.GetFlavor(flavor); fl != nil {
			w.Hints = &aegis.ResourceHints{
				ResourceName: fl.GetResourceName(),
				GpuCount:     fl.GetGpuCount(),
			}
			s.log.Debug("resource hints stamped",
				zap.String("workload_id", w.GetId()),
				zap.String("queue", queue),
				zap.String("flavor", flavor),
				zap.String("resource_name", fl.GetResourceName()),
				zap.Int("gpu_count", int(fl.GetGpuCount())),
			)
		} else {
			s.log.Debug("flavor missing in catalog; hints skipped",
				zap.String("workload_id", w.GetId()),
				zap.String("flavor", flavor),
			)
		}
	}
	if len(items) > 0 {
		ids := make([]string, 0, len(items))
		for _, w := range items {
			ids = append(ids, w.GetId())
		}
		s.log.Info("workloads leased",
			zap.String("cluster_id", req.GetClusterId()),
			zap.Strings("workload_ids", ids),
			zap.Int("count", len(items)),
		)
		mLeased.Add(float64(len(items)))
	} else {
		s.log.Debug("no workloads leased", zap.String("cluster_id", req.GetClusterId()))
	}
	return &aegis.LeaseWorkloadResponse{Items: items}, nil
}

func (s *Server) StartWorkload(ctx context.Context, req *aegis.StartWorkloadRequest) (*aegis.StartWorkloadResponse, error) {
	if req == nil || req.GetId() == "" {
		err := status.Error(codes.InvalidArgument, "workload id required")
		s.log.Warn("start workload failed", zap.Error(err))
		return nil, err
	}
	if req.GetClusterId() == "" {
		err := status.Error(codes.InvalidArgument, "cluster_id required")
		s.log.Warn("start workload failed", zap.Error(err), zap.String("workload_id", req.GetId()))
		return nil, err
	}

	w := s.store.GetWorkload(req.GetId())
	if w == nil {
		err := status.Error(codes.NotFound, "workload not found")
		s.log.Warn("start workload failed", zap.Error(err), zap.String("workload_id", req.GetId()))
		return nil, err
	}

	if w.GetClusterId() != req.GetClusterId() {
		err := status.Errorf(codes.FailedPrecondition, "workload assigned to cluster %s", w.GetClusterId())
		s.log.Warn("start workload cluster mismatch",
			zap.Error(err),
			zap.String("workload_id", w.GetId()),
			zap.String("expected_cluster", w.GetClusterId()),
			zap.String("requested_cluster", req.GetClusterId()),
		)
		return nil, err
	}

	updated, wait, observed, err := s.store.StartWorkload(req.GetId())
	if err != nil {
		errStatus := status.Error(codes.FailedPrecondition, err.Error())
		s.log.Warn("start workload state transition failed",
			zap.Error(err),
			zap.String("workload_id", req.GetId()),
		)
		return nil, errStatus
	}

	queue := updated.GetQueue()
	if queue == "" {
		queue = "unknown"
	}
	if observed {
		if flavor, ferr := requiredFlavor(updated); ferr == nil {
			mQueueWait.WithLabelValues(queue, flavor).Observe(wait.Seconds())
			s.log.Debug("queue wait observed",
				zap.String("workload_id", updated.GetId()),
				zap.String("queue", queue),
				zap.String("flavor", flavor),
				zap.Duration("wait", wait),
			)
		} else {
			s.log.Debug("start workload missing flavor for metrics",
				zap.String("workload_id", updated.GetId()),
				zap.Error(ferr),
			)
		}
	}

	idempotent := !observed && wait == 0
	s.log.Info("workload start acknowledged",
		zap.String("workload_id", updated.GetId()),
		zap.String("cluster_id", updated.GetClusterId()),
		zap.Bool("idempotent", idempotent),
		zap.Duration("queue_wait", wait),
	)

	return &aegis.StartWorkloadResponse{Workload: updated}, nil
}

func (s *Server) AckWorkload(ctx context.Context, req *aegis.AckWorkloadRequest) (*aegis.AckWorkloadResponse, error) {
	if req == nil || req.GetId() == "" {
		err := status.Error(codes.InvalidArgument, "workload id required")
		s.log.Warn("ack workload failed", zap.Error(err))
		return nil, err
	}
	if req.GetStatus() == "" {
		err := status.Error(codes.InvalidArgument, "status required")
		s.log.Warn("ack workload failed", zap.Error(err), zap.String("workload_id", req.GetId()))
		return nil, err
	}
	w := s.store.GetWorkload(req.GetId())
	if w == nil {
		err := status.Error(codes.NotFound, "workload not found")
		s.log.Warn("ack workload failed", zap.Error(err), zap.String("workload_id", req.GetId()))
		return nil, err
	}
	if w.GetStatus() != statusRunning {
		err := status.Errorf(codes.FailedPrecondition, "workload not in %s state", statusRunning)
		s.log.Warn("ack workload failed",
			zap.Error(err),
			zap.String("workload_id", w.GetId()),
			zap.String("current_status", w.GetStatus()),
		)
		return nil, err
	}
	updated, err := s.store.AckWorkload(req.GetId(), req.GetStatus(), req.GetUrl())
	if err != nil {
		s.log.Error("ack workload store update failed",
			zap.Error(err),
			zap.String("workload_id", req.GetId()),
		)
		return nil, status.Error(codes.Internal, "failed to update workload status")
	}
	backend := req.GetBackend()
	if backend == "" {
		switch w.GetKind().(type) {
		case *aegis.Workload_Workspace:
			backend = "workspace"
		case *aegis.Workload_Training:
			backend = "trainer_v2"
		default:
			backend = "unknown"
		}
	}
	mAcked.WithLabelValues(req.GetStatus(), backend).Inc()
	s.log.Info("workload acknowledged",
		zap.String("workload_id", updated.GetId()),
		zap.String("cluster_id", updated.GetClusterId()),
		zap.String("status", updated.GetStatus()),
		zap.String("backend", backend),
		zap.String("url", updated.GetUrl()),
	)
	if b, _ := s.store.ResolveBudget(updated.GetProjectId(), updated.GetQueue()); b != nil {
		reqFlavor, _ := requiredFlavor(updated)
		fl := s.store.GetFlavor(reqFlavor)
		runtimeSecs := int64(0)
		if t0, ok := s.store.GetStartedAt(updated.GetId()); ok {
			runtimeSecs = int64(time.Since(t0).Seconds())
			if runtimeSecs < 0 {
				runtimeSecs = 0
			}
		}
		actualUSD := s.runtimeCostUSD(updated, fl, runtimeSecs)
		estUSD := s.store.PopEstimateUSD(updated.GetId())
		usage := s.store.ReconcileOnAck(updated.GetProjectId(), updated.GetQueue(), estUSD, actualUSD)
		mBudgetReserved.WithLabelValues(updated.GetProjectId(), updated.GetQueue()).Set(usage.ReservedUSD)
		mBudgetActual.WithLabelValues(updated.GetProjectId(), updated.GetQueue()).Set(usage.ActualUSD)
		s.log.Info("budget reconciled",
			zap.String("workload_id", updated.GetId()),
			zap.String("project_id", updated.GetProjectId()),
			zap.String("queue", updated.GetQueue()),
			zap.Float64("released_estimate_usd", estUSD),
			zap.Float64("actual_usd", actualUSD),
			zap.Float64("usage_reserved_usd", usage.ReservedUSD),
			zap.Float64("usage_actual_usd", usage.ActualUSD),
			zap.Int64("runtime_secs", runtimeSecs),
		)
	}
	return &aegis.AckWorkloadResponse{Workload: updated}, nil
}

func (s *Server) estimateWorkloadUSD(w *aegis.Workload, fl *aegis.Flavor, q *aegis.Queue) (float64, int64, error) {
	if fl == nil {
		return 0, 0, fmt.Errorf("missing flavor")
	}
	price := fl.GetPriceUsdPerGpuHour()
	var totalGPUs int32
	var maxSecs int64
	switch wk := w.GetKind().(type) {
	case *aegis.Workload_Workspace:
		totalGPUs = fl.GetGpuCount()
		if wk.Workspace.GetMaxDurationSeconds() > 0 {
			maxSecs = wk.Workspace.GetMaxDurationSeconds()
		}
	case *aegis.Workload_Training:
		workers := wk.Training.GetWorkers()
		if workers <= 0 {
			workers = 1
		}
		gpusPer := wk.Training.GetGpusPerWorker()
		if gpusPer <= 0 {
			gpusPer = fl.GetGpuCount()
		}
		totalGPUs = workers * gpusPer
		if wk.Training.GetMaxDurationSeconds() > 0 {
			maxSecs = wk.Training.GetMaxDurationSeconds()
		}
	default:
	}
	if maxSecs <= 0 {
		maxSecs = s.defaultMaxRuntimeSeconds(q)
	}
	usd := float64(totalGPUs) * price * (float64(maxSecs) / 3600.0)
	return usd, maxSecs, nil
}

func (s *Server) runtimeCostUSD(w *aegis.Workload, fl *aegis.Flavor, runtimeSecs int64) float64 {
	if fl == nil || runtimeSecs <= 0 {
		return 0
	}
	price := fl.GetPriceUsdPerGpuHour()
	var totalGPUs int32
	switch wk := w.GetKind().(type) {
	case *aegis.Workload_Workspace:
		totalGPUs = fl.GetGpuCount()
	case *aegis.Workload_Training:
		workers := wk.Training.GetWorkers()
		if workers <= 0 {
			workers = 1
		}
		gpusPer := wk.Training.GetGpusPerWorker()
		if gpusPer <= 0 {
			gpusPer = fl.GetGpuCount()
		}
		totalGPUs = workers * gpusPer
	default:
	}
	return float64(totalGPUs) * price * (float64(runtimeSecs) / 3600.0)
}

func (s *Server) GetBudget(ctx context.Context, req *aegis.GetBudgetRequest) (*aegis.GetBudgetResponse, error) {
	if req == nil || req.GetProjectId() == "" {
		return nil, status.Error(codes.InvalidArgument, "project_id required")
	}
	b := s.store.GetBudgetExact(req.GetProjectId(), req.GetQueue())
	if b == nil {
		b = s.store.GetBudgetExact(req.GetProjectId(), "")
	}
	if b == nil {
		return nil, status.Error(codes.NotFound, "budget not found")
	}
	view, _ := s.store.UsageView(b.GetProjectId(), b.GetQueue())
	resp := &aegis.GetBudgetResponse{Budget: b, Usage: &aegis.BudgetUsage{
		ActualUsd:      view.ActualUSD,
		ReservedUsd:    view.ReservedUSD,
		RemainingUsd:   b.GetLimitUsd() - (view.ActualUSD + view.ReservedUSD),
		PeriodStartUtc: view.PeriodStart.Format("2006-01-02"),
		PeriodEndUtc:   view.PeriodEnd.Format("2006-01-02"),
	}}
	return resp, nil
}

func (s *Server) ListBudgets(ctx context.Context, req *aegis.ListBudgetsRequest) (*aegis.ListBudgetsResponse, error) {
	items := []*aegis.BudgetWithUsage{}
	for _, b := range s.store.ListBudgets(req.GetProjectId()) {
		view, _ := s.store.UsageView(b.GetProjectId(), b.GetQueue())
		items = append(items, &aegis.BudgetWithUsage{
			Budget: b,
			Usage: &aegis.BudgetUsage{
				ActualUsd:      view.ActualUSD,
				ReservedUsd:    view.ReservedUSD,
				RemainingUsd:   b.GetLimitUsd() - (view.ActualUSD + view.ReservedUSD),
				PeriodStartUtc: view.PeriodStart.Format("2006-01-02"),
				PeriodEndUtc:   view.PeriodEnd.Format("2006-01-02"),
			},
		})
	}
	return &aegis.ListBudgetsResponse{Items: items}, nil
}

func Run(ctx context.Context, log *zap.Logger, addrGRPC, addrHTTP string, svc *Server) error {
	lis, err := net.Listen("tcp", addrGRPC)
	if err != nil {
		return err
	}
	gs := grpc.NewServer()
	reflection.Register(gs)
	aegis.RegisterAegisPlatformServer(gs, svc)

	log.Info("gRPC server listening", zap.String("addr", addrGRPC))

	// HTTP gateway mux with Prometheus metrics exposed.
	mux := runtime.NewServeMux()
	root := http.NewServeMux()
	root.Handle("/", mux)
	root.Handle("/metrics", promhttp.Handler())
	httpSrv := &http.Server{Addr: addrHTTP, Handler: root}

	go func() {
		log.Info("HTTP gateway listening", zap.String("addr", addrHTTP))
		if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Warn("HTTP gateway stopped", zap.Error(err))
		}
	}()

	if err := gs.Serve(lis); err != nil {
		if errors.Is(err, grpc.ErrServerStopped) {
			log.Info("gRPC server stopped")
			return nil
		}
		return err
	}
	return nil
}

func buildAegisWorkloadCR(w *aegis.Workload, namespace string) *aegisv1alpha1.AegisWorkload {
	spec := aegisv1alpha1.AegisWorkloadSpec{
		ProjectID: w.GetProjectId(),
		Queue:     w.GetQueue(),
	}

	switch wk := w.GetKind().(type) {
	case *aegis.Workload_Workspace:
		ws := wk.Workspace
		if ws != nil {
			spec.Workspace = &aegisv1alpha1.WorkspaceSpec{
				Flavor:  ws.GetFlavor(),
				Image:   ws.GetImage(),
				Env:     cloneStringMap(ws.GetEnv()),
				Command: cloneStringSlice(ws.GetCommand()),
			}
		}
	case *aegis.Workload_Training:
		tr := wk.Training
		if tr != nil {
			spec.Training = &aegisv1alpha1.TrainingSpec{
				Flavor:        tr.GetFlavor(),
				Workers:       tr.GetWorkers(),
				GpusPerWorker: tr.GetGpusPerWorker(),
				Image:         tr.GetImage(),
				Command:       cloneStringSlice(tr.GetCommand()),
				Gang:          tr.GetGang(),
			}
		}
	}

	if hints := w.GetHints(); hints != nil {
		spec.Hints = &aegisv1alpha1.ResourceHints{
			ResourceName: hints.GetResourceName(),
			GpuCount:     hints.GetGpuCount(),
		}
	}

	return &aegisv1alpha1.AegisWorkload{
		TypeMeta: metav1.TypeMeta{
			APIVersion: aegisv1alpha1.GroupVersion.String(),
			Kind:       "AegisWorkload",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      w.GetId(),
			Namespace: namespace,
		},
		Spec: spec,
	}
}

func cloneStringSlice(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func cloneStringMap(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func requiredFlavor(w *aegis.Workload) (string, error) {
	switch wk := w.GetKind().(type) {
	case *aegis.Workload_Workspace:
		if f := wk.Workspace.GetFlavor(); f != "" {
			return f, nil
		}
	case *aegis.Workload_Training:
		if f := wk.Training.GetFlavor(); f != "" {
			return f, nil
		}
	default:
	}
	return "", fmt.Errorf("flavor required on workspace or training")
}

func RandID() string {
	buf := make([]byte, 4)
	if _, err := rand.Read(buf); err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 16)
	}
	return hex.EncodeToString(buf)
}
