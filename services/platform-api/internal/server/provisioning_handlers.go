package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

type provisioningLogView struct {
	Timestamp string `json:"timestamp"`
	Phase     string `json:"phase,omitempty"`
	Type      string `json:"type,omitempty"`
	Message   string `json:"message"`
}

type nodePoolView struct {
	Name         string `json:"name"`
	InstanceType string `json:"instanceType"`
	MinSize      int32  `json:"minSize"`
	MaxSize      int32  `json:"maxSize"`
}

type clusterConfigView struct {
	Provider    string         `json:"provider,omitempty"`
	Region      string         `json:"region,omitempty"`
	K8sVersion  string         `json:"k8sVersion,omitempty"`
	ClusterName string         `json:"clusterName,omitempty"`
	NodePools   []nodePoolView `json:"nodePools,omitempty"`
	Autoscaling bool           `json:"autoscaling"`
}

type provisioningLogsResponse struct {
	JobID         string             `json:"jobId"`
	ProjectID     string             `json:"projectId,omitempty"`
	ClusterID     string             `json:"clusterId,omitempty"`
	Phase         string             `json:"phase,omitempty"`
	StartedAt     string             `json:"startedAt,omitempty"`
	CompletedAt   string             `json:"completedAt,omitempty"`
	Logs          []provisioningLogView `json:"logs,omitempty"`
	NextCursor    string             `json:"nextCursor,omitempty"`
	ClusterConfig *clusterConfigView `json:"clusterConfig,omitempty"`
}

func registerProvisioningRoutes(mux *runtime.ServeMux, srv *Server) {
	if mux == nil || srv == nil {
		return
	}
	_ = mux.HandlePath(http.MethodGet, "/api/v1/provisioning/jobs/{jobId}/logs", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		srv.handleProvisioningLogs(w, r, pathParams["jobId"], false)
	})
	_ = mux.HandlePath(http.MethodGet, "/api/v1/provisioning/jobs/{jobId}/logs/stream", func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
		srv.handleProvisioningLogs(w, r, pathParams["jobId"], true)
	})
}

func (s *Server) handleProvisioningLogs(w http.ResponseWriter, r *http.Request, jobID string, forceStream bool) {
	ctx := r.Context()
	jobID = strings.TrimSpace(jobID)
	if jobID == "" {
		writeWizardError(w, status.Error(codes.InvalidArgument, "jobId is required"))
		return
	}

	var clusterConfig *clusterConfigView
	run, _ := s.store.GetProvisioningRun(jobID)
	s.log.Info("provisioning logs request",
		zap.String("job_id", jobID),
		zap.Bool("has_run", run != nil),
		zap.Bool("has_infra_client", s.infraClient != nil))
	if run == nil && s.infraClient != nil {
		if infra, err := s.fetchInfra(ctx, jobID); err == nil && infra != nil {
			clusterID := ""
			if infra.Annotations != nil {
				clusterID = strings.TrimSpace(infra.Annotations["aegis.yourorg.dev/clusterId"])
			}
			if clusterID == "" && infra.Spec.Aws != nil {
				clusterID = strings.TrimSpace(infra.Spec.Aws.ClusterName)
			}
			run = &store.ProvisioningRun{
				JobID:     jobID,
				ProjectID: strings.TrimSpace(infra.Spec.ProjectID),
				ClusterID: clusterID,
				Phase:     strings.TrimSpace(infra.Status.Phase),
			}
			// Try to get actual start time from provisioning logs first
			firstLogs := s.store.ListProvisioningLogs(jobID, time.Time{}, 0, 1)
			if len(firstLogs) > 0 {
				run.StartedAt = firstLogs[0].CreatedAt
			} else if infra.Status.LastSyncTime != nil {
				run.StartedAt = infra.Status.LastSyncTime.Time
			} else if !infra.CreationTimestamp.IsZero() {
				run.StartedAt = infra.CreationTimestamp.Time
			}
			if strings.EqualFold(run.Phase, "Ready") || strings.EqualFold(run.Phase, "Error") {
				// Get completion time from last log entry or lastSyncTime
				lastLogs := s.store.ListProvisioningLogs(jobID, time.Time{}, 0, 1000)
				if len(lastLogs) > 0 {
					ts := lastLogs[len(lastLogs)-1].CreatedAt
					run.CompletedAt = &ts
				} else {
					ts := time.Now().UTC()
					if infra.Status.LastSyncTime != nil {
						ts = infra.Status.LastSyncTime.Time
					}
					run.CompletedAt = &ts
				}
			}
			s.store.UpsertProvisioningRun(*run)

			// Extract cluster configuration from infra spec
			clusterConfig = extractClusterConfig(infra)
		}
	} else if s.infraClient != nil {
		// Also fetch cluster config for existing runs
		if infra, err := s.fetchInfra(ctx, jobID); err == nil && infra != nil {
			clusterConfig = extractClusterConfig(infra)
			s.log.Info("extracted cluster config for existing run",
				zap.String("job_id", jobID),
				zap.Bool("has_config", clusterConfig != nil))
		} else if err != nil {
			s.log.Info("failed to fetch infra for cluster config",
				zap.String("job_id", jobID),
				zap.Error(err))
		}
	} else {
		s.log.Info("infra client not available for cluster config",
			zap.String("job_id", jobID))
	}
	if run == nil {
		writeWizardError(w, status.Errorf(codes.NotFound, "provisioning job %q not found", jobID))
		return
	}

	if err := s.authorize(ctx, run.ProjectID, "", "getProvisioningLogs"); err != nil {
		writeWizardError(w, err)
		return
	}

	limit := parseLimit(r.URL.Query().Get("limit"), 200, 1000)
	cursorTS, cursorSeq := parseCursor(r.URL.Query().Get("since"))
	stream := forceStream || parseBool(r.URL.Query().Get("stream"))

	logs := s.store.ListProvisioningLogs(jobID, cursorTS, cursorSeq, limit)
	if stream && len(logs) == 0 && (run.CompletedAt == nil || cursorTS.Before(*run.CompletedAt)) {
		if run.CompletedAt == nil {
			logs = s.waitForLogs(ctx, jobID, cursorTS, cursorSeq, limit)
		}
	}
	if len(logs) > 0 {
		last := logs[len(logs)-1]
		cursorTS = last.CreatedAt
		cursorSeq = last.Sequence
	}

	resp := provisioningLogsResponse{
		JobID:         jobID,
		ProjectID:     run.ProjectID,
		ClusterID:     run.ClusterID,
		Phase:         strings.TrimSpace(run.Phase),
		Logs:          make([]provisioningLogView, 0, len(logs)),
		ClusterConfig: clusterConfig,
	}
	if !run.StartedAt.IsZero() {
		resp.StartedAt = run.StartedAt.UTC().Format(time.RFC3339Nano)
	}
	if run.CompletedAt != nil {
		resp.CompletedAt = run.CompletedAt.UTC().Format(time.RFC3339Nano)
	}
	for _, entry := range logs {
		resp.Logs = append(resp.Logs, provisioningLogView{
			Timestamp: entry.CreatedAt.UTC().Format(time.RFC3339Nano),
			Phase:     entry.Phase,
			Type:      entry.Type,
			Message:   entry.Message,
		})
	}
	if len(logs) > 0 {
		resp.NextCursor = formatCursor(cursorTS, cursorSeq)
	} else if run.CompletedAt != nil {
		resp.NextCursor = formatCursor(*run.CompletedAt, 0)
	}

	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) waitForLogs(ctx context.Context, jobID string, since time.Time, sinceSeq int64, limit int) []store.ProvisioningLogEntry {
	deadline := time.Now().Add(20 * time.Second)
	cursor := since
	cursorSeq := sinceSeq
	for time.Now().Before(deadline) {
		if ctx.Err() != nil {
			return nil
		}
		entries := s.store.ListProvisioningLogs(jobID, cursor, cursorSeq, limit)
		if len(entries) > 0 {
			return entries
		}
		select {
		case <-time.After(500 * time.Millisecond):
		case <-ctx.Done():
			return nil
		}
	}
	return nil
}

func parseLimit(raw string, def, max int) int {
	val := def
	if trimmed := strings.TrimSpace(raw); trimmed != "" {
		if parsed, err := strconv.Atoi(trimmed); err == nil && parsed > 0 {
			val = parsed
		}
	}
	if val <= 0 {
		val = def
	}
	if val > max {
		val = max
	}
	return val
}

func parseCursor(raw string) (time.Time, int64) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		// Default to 30 minutes ago to avoid returning ALL logs on initial load
		return time.Now().UTC().Add(-30 * time.Minute), 0
	}

	var seqPart string
	tsPart := trimmed
	if strings.Contains(trimmed, "|") {
		parts := strings.SplitN(trimmed, "|", 2)
		tsPart = parts[0]
		seqPart = parts[1]
	}

	ts, err := time.Parse(time.RFC3339Nano, tsPart)
	if err != nil {
		return time.Now().UTC().Add(-30 * time.Minute), 0
	}

	var seq int64
	if seqPart != "" {
		if parsed, parseErr := strconv.ParseInt(seqPart, 10, 64); parseErr == nil {
			seq = parsed
		}
	}
	return ts, seq
}

func parseBool(raw string) bool {
	if trimmed := strings.TrimSpace(raw); trimmed != "" {
		if parsed, err := strconv.ParseBool(trimmed); err == nil {
			return parsed
		}
	}
	return false
}

func formatCursor(ts time.Time, seq int64) string {
	if ts.IsZero() {
		return ""
	}
	if seq > 0 {
		return fmt.Sprintf("%s|%d", ts.UTC().Format(time.RFC3339Nano), seq)
	}
	return ts.UTC().Format(time.RFC3339Nano)
}

// extractClusterConfig builds a cluster configuration view from the ProjectInfra spec.
func extractClusterConfig(infra *infraapi.ProjectInfra) *clusterConfigView {
	if infra == nil {
		return nil
	}
	cfg := &clusterConfigView{
		Provider:    strings.TrimSpace(infra.Spec.Provider),
		Region:      strings.TrimSpace(infra.Spec.Region),
		Autoscaling: true, // Default to true since we use cluster autoscaler
	}
	if infra.Spec.Aws != nil {
		cfg.ClusterName = strings.TrimSpace(infra.Spec.Aws.ClusterName)
		cfg.K8sVersion = strings.TrimSpace(infra.Spec.Aws.Version)
		if cfg.K8sVersion == "" {
			cfg.K8sVersion = "1.29" // Default version
		}
		// Extract node pools
		for _, np := range infra.Spec.Aws.NodePools {
			cfg.NodePools = append(cfg.NodePools, nodePoolView{
				Name:         strings.TrimSpace(np.Name),
				InstanceType: strings.TrimSpace(np.InstanceType),
				MinSize:      np.MinSize,
				MaxSize:      np.MaxSize,
			})
		}
	}
	return cfg
}
