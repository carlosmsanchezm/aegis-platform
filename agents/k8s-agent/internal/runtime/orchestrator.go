package runtime

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"

	"github.com/yourorg/aegis/agents/k8s-agent/internal/cpclient"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/executor"
	execjob "github.com/yourorg/aegis/agents/k8s-agent/internal/executor/job"
	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

type Orchestrator struct {
	log       *zap.Logger
	clusterID string
	client    *cpclient.Client
	exec      executor.Executor
	parallel  int
}

func NewOrchestrator(log *zap.Logger, clusterID string, client *cpclient.Client) (*Orchestrator, error) {
	exec, err := execjob.New(log)
	if err != nil {
		return nil, err
	}
	parallel := getenvInt("AEGIS_MAX_PARALLEL", 4)
	return &Orchestrator{
		log:       log,
		clusterID: clusterID,
		client:    client,
		exec:      exec,
		parallel:  parallel,
	}, nil
}

func (o *Orchestrator) Run(ctx context.Context) {
	o.log.Info("orchestrator started", zap.String("cluster_id", o.clusterID), zap.Int("parallel", o.parallel))
	free := make(chan struct{}, o.parallel)
	for i := 0; i < o.parallel; i++ {
		free <- struct{}{}
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			o.log.Info("orchestrator stopping", zap.String("cluster_id", o.clusterID))
			return
		case <-ticker.C:
			slots := len(free)
			if slots == 0 {
				o.log.Debug("no available executor slots; waiting", zap.String("cluster_id", o.clusterID))
				continue
			}
			workloads, err := o.client.Lease(ctx, o.clusterID, int32(slots))
			if err != nil {
				o.log.Warn("lease failed", zap.Error(err))
				continue
			}
			if len(workloads) == 0 {
				o.log.Debug("lease returned zero workloads", zap.String("cluster_id", o.clusterID))
				continue
			}
			for _, wl := range workloads {
				<-free
				go o.execute(ctx, wl, free)
			}
		}
	}
}

func (o *Orchestrator) execute(ctx context.Context, wl *aegis.Workload, free chan<- struct{}) {
	defer func() { free <- struct{}{} }()

	kind := kindOf(wl)
	o.log.Info("executing workload", zap.String("workload_id", wl.GetId()), zap.String("kind", kind))
	if h := wl.GetHints(); h != nil {
		o.log.Info("workload hints",
			zap.String("workload_id", wl.GetId()),
			zap.String("resource_name", h.GetResourceName()),
			zap.Int32("gpu_count", h.GetGpuCount()),
		)
	}

	var res executor.Result
	switch wl.GetKind().(type) {
	case *aegis.Workload_Workspace:
		res = o.exec.RunWorkspace(ctx, wl)
	case *aegis.Workload_Training:
		res = o.exec.RunTraining(ctx, wl)
	default:
		res = executor.Result{Status: "FAILED", Err: fmt.Errorf("unsupported workload kind")}
	}

	status := res.Status
	if status == "" {
		status = "FAILED"
	}
	backend := res.Backend
	if backend == "" {
		backend = kind
	}
	url := res.URL
	maxRetries := 6
	backoff := time.Second
	var lastErr error
	acked := false
	for attempt := 1; attempt <= maxRetries; attempt++ {
		if err := o.client.Ack(ctx, wl.GetId(), status, backend, url); err != nil {
			lastErr = err
			o.log.Warn("ack failed (retrying)",
				zap.String("workload_id", wl.GetId()),
				zap.String("status", status),
				zap.String("backend", backend),
				zap.String("url", url),
				zap.Int("attempt", attempt),
				zap.Duration("backoff", backoff),
				zap.Error(err),
			)
			time.Sleep(backoff)
			backoff *= 2
			continue
		}
		o.log.Info("ack succeeded", zap.String("workload_id", wl.GetId()), zap.String("status", status), zap.String("backend", backend), zap.String("url", url), zap.Int("attempt", attempt))
		acked = true
		break
	}
	if !acked && lastErr != nil {
		o.log.Error("ack exhausted retries", zap.String("workload_id", wl.GetId()), zap.String("status", status), zap.String("backend", backend), zap.String("url", url), zap.Error(lastErr))
	}
}

func kindOf(w *aegis.Workload) string {
	switch w.GetKind().(type) {
	case *aegis.Workload_Workspace:
		return "workspace"
	case *aegis.Workload_Training:
		return "training"
	default:
		return "unknown"
	}
}

func getenvInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if val, err := strconv.Atoi(v); err == nil && val > 0 {
			return val
		}
	}
	return def
}
