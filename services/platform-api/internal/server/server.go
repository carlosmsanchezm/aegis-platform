package server

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
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

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	"github.com/yourorg/aegis/services/platform-api/internal/placement"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

type Server struct {
	aegis.UnimplementedAegisPlatformServer
	log   *zap.Logger
	store *store.MemStore
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
)

func New(log *zap.Logger, st *store.MemStore) *Server { return &Server{log: log, store: st} }

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
	s.log.Info("budget upserted", zap.String("project_id", b.GetProjectId()), zap.Float64("gpu_hours_cap", b.GetGpuHoursCap()), zap.Float64("warn_threshold_pct", b.GetWarnThresholdPct()))
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
	s.log.Info("cluster heartbeat",
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
	if s.store.GetFlavor(reqFlavor) == nil {
		s.log.Warn("submit workload rejected; unknown flavor",
			zap.String("workload_id", w.GetId()),
			zap.String("flavor", reqFlavor),
		)
		return nil, status.Error(codes.FailedPrecondition, "unknown flavor: "+reqFlavor)
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
	chosen, perr := placement.ChooseCluster(cands, pd, reqFlavor)
	if perr != nil {
		s.log.Warn("placement failed",
			zap.String("workload_id", w.GetId()),
			zap.String("project_id", w.GetProjectId()),
			zap.String("flavor", reqFlavor),
			zap.Strings("regions", regions),
			zap.Error(perr),
		)
		return nil, status.Error(codes.FailedPrecondition, fmt.Sprintf("no eligible cluster for flavor=%s in policy regions=%v", reqFlavor, regions))
	}

	w.ClusterId = chosen
	w.Status = statusPlaced
	s.store.PutWorkload(w)
	s.store.MarkPlaced(w.GetId())
	mPlaced.WithLabelValues(reqFlavor).Inc()
	s.log.Info("workload placed",
		zap.String("workload_id", w.GetId()),
		zap.String("project_id", w.GetProjectId()),
		zap.String("flavor", reqFlavor),
		zap.String("cluster_id", w.GetClusterId()),
		zap.String("status", w.GetStatus()),
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
	return &aegis.AckWorkloadResponse{Workload: updated}, nil
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
