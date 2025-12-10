package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

type provisioningLogView struct {
	Timestamp string `json:"timestamp"`
	Phase     string `json:"phase,omitempty"`
	Type      string `json:"type,omitempty"`
	Message   string `json:"message"`
}

type provisioningLogsResponse struct {
	JobID       string                `json:"jobId"`
	ProjectID   string                `json:"projectId,omitempty"`
	ClusterID   string                `json:"clusterId,omitempty"`
	Phase       string                `json:"phase,omitempty"`
	StartedAt   string                `json:"startedAt,omitempty"`
	CompletedAt string                `json:"completedAt,omitempty"`
	Logs        []provisioningLogView `json:"logs,omitempty"`
	NextCursor  string                `json:"nextCursor,omitempty"`
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

	run, _ := s.store.GetProvisioningRun(jobID)
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
			if infra.Status.LastSyncTime != nil {
				run.StartedAt = infra.Status.LastSyncTime.Time
			} else if !infra.CreationTimestamp.IsZero() {
				run.StartedAt = infra.CreationTimestamp.Time
			}
			if strings.EqualFold(run.Phase, "Ready") || strings.EqualFold(run.Phase, "Error") {
				ts := time.Now().UTC()
				if infra.Status.LastSyncTime != nil {
					ts = infra.Status.LastSyncTime.Time
				}
				run.CompletedAt = &ts
			}
			s.store.UpsertProvisioningRun(*run)
		}
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
		logs = s.waitForLogs(ctx, jobID, cursorTS, cursorSeq, limit)
	}
	if len(logs) > 0 {
		last := logs[len(logs)-1]
		cursorTS = last.CreatedAt
		cursorSeq = last.Sequence
	}

	resp := provisioningLogsResponse{
		JobID:     jobID,
		ProjectID: run.ProjectID,
		ClusterID: run.ClusterID,
		Phase:     strings.TrimSpace(run.Phase),
		Logs:      make([]provisioningLogView, 0, len(logs)),
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
		return time.Time{}, 0
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
		return time.Time{}, 0
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
