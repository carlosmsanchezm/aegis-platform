package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
	"google.golang.org/protobuf/encoding/protojson"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

const (
	statusPlaced     = "PLACED"
	statusRunning    = "RUNNING"
	statusSuspended  = "SUSPENDED"
	statusTerminated = "TERMINATED"
)

var (
	protoMarshal   = protojson.MarshalOptions{EmitUnpopulated: true}
	protoUnmarshal = protojson.UnmarshalOptions{DiscardUnknown: true}
)

type rowScanner interface {
	Scan(dest ...any) error
}

func (s *PostgresStore) PutWorkload(w *aegis.Workload) {
	if w == nil || w.GetId() == "" {
		return
	}
	var (
		workspaceJSON []byte
		trainingJSON  []byte
		err           error
	)
	kind := "unknown"
	if wk, ok := w.GetKind().(*aegis.Workload_Workspace); ok && wk.Workspace != nil {
		workspaceJSON, err = protoMarshal.Marshal(wk.Workspace)
		if err != nil {
			s.logExecError("marshal_workspace", err, zap.String("workload_id", w.GetId()))
		}
		kind = "workspace"
	} else if tr, ok := w.GetKind().(*aegis.Workload_Training); ok && tr.Training != nil {
		trainingJSON, err = protoMarshal.Marshal(tr.Training)
		if err != nil {
			s.logExecError("marshal_training", err, zap.String("workload_id", w.GetId()))
		}
		kind = "training"
	}

	var hintsResource interface{}
	var hintsGPU interface{}
	var hintsCPU interface{}
	var hintsMem interface{}
	if hints := w.GetHints(); hints != nil {
		hintsResource = nullableString(hints.GetResourceName())
		if hints.GetGpuCount() != 0 {
			hintsGPU = hints.GetGpuCount()
		}
		hintsCPU = nullableString(hints.GetCpuCoresRequest())
		hintsMem = nullableString(hints.GetMemoryRequest())
	}

	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	_, execErr := s.pool.Exec(ctx, `
INSERT INTO workloads (
    id, project_id, queue, cluster_id, status, ui_status, url, message, kind,
    hints_resource_name, hints_gpu_count, hints_cpu_request, hints_mem_request,
    workspace_json, training_json, created_at, updated_at
) VALUES (
    $1, $2, COALESCE($3, ''), $4, $5, NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, ''), $9,
    $10, $11, $12, $13, $14, $15, now(), now()
)
ON CONFLICT (id) DO UPDATE SET
    project_id = EXCLUDED.project_id,
    queue = EXCLUDED.queue,
    cluster_id = EXCLUDED.cluster_id,
    status = EXCLUDED.status,
    ui_status = EXCLUDED.ui_status,
    url = EXCLUDED.url,
    message = EXCLUDED.message,
    kind = EXCLUDED.kind,
    hints_resource_name = EXCLUDED.hints_resource_name,
    hints_gpu_count = EXCLUDED.hints_gpu_count,
    hints_cpu_request = EXCLUDED.hints_cpu_request,
    hints_mem_request = EXCLUDED.hints_mem_request,
    workspace_json = EXCLUDED.workspace_json,
    training_json = EXCLUDED.training_json,
    updated_at = now()
`,
		w.GetId(),
		w.GetProjectId(),
		nullableString(w.GetQueue()),
		nullableString(w.GetClusterId()),
		w.GetStatus(),
		w.GetUiStatus(),
		w.GetUrl(),
		w.GetMessage(),
		kind,
		hintsResource,
		hintsGPU,
		hintsCPU,
		hintsMem,
		bytesOrNil(workspaceJSON),
		bytesOrNil(trainingJSON),
	)
	s.logExecError("upsert_workload", execErr, zap.String("workload_id", w.GetId()))
}

func (s *PostgresStore) GetWorkload(id string) *aegis.Workload {
	if id == "" {
		return nil
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	row := s.pool.QueryRow(ctx, workloadSelect("WHERE id=$1"), id)
	w, err := s.scanWorkload(row)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.logExecError("get_workload", err, zap.String("workload_id", id))
		}
		return nil
	}
	return w
}

func (s *PostgresStore) ListWorkloads(projectID string) []*aegis.Workload {
	if projectID == "" {
		return nil
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	rows, err := s.pool.Query(ctx, workloadSelect("WHERE project_id=$1 ORDER BY created_at DESC"), projectID)
	if err != nil {
		s.logExecError("list_workloads", err, zap.String("project_id", projectID))
		return nil
	}
	defer rows.Close()
	var items []*aegis.Workload
	for rows.Next() {
		w, scanErr := s.scanWorkload(rows)
		if scanErr != nil {
			s.logExecError("scan_workload", scanErr)
			break
		}
		items = append(items, w)
	}
	return items
}

func (s *PostgresStore) ListClusterWorkloadIDs(clusterID string) ([]string, error) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return nil, fmt.Errorf("cluster id required")
	}

	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	rows, err := s.pool.Query(ctx, `SELECT id FROM workloads WHERE cluster_id=$1 ORDER BY id`, clusterID)
	if err != nil {
		s.logExecError("list_cluster_workload_ids_query", err, zap.String("cluster_id", clusterID))
		return nil, err
	}
	defer rows.Close()

	ids := []string{}
	for rows.Next() {
		var id string
		if scanErr := rows.Scan(&id); scanErr != nil {
			s.logExecError("list_cluster_workload_ids_scan", scanErr, zap.String("cluster_id", clusterID))
			return nil, scanErr
		}
		ids = append(ids, id)
	}
	if rowsErr := rows.Err(); rowsErr != nil {
		s.logExecError("list_cluster_workload_ids_rows", rowsErr, zap.String("cluster_id", clusterID))
		return nil, rowsErr
	}
	return ids, nil
}

func (s *PostgresStore) MarkPlaced(id string) {
	if id == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	_, err := s.pool.Exec(ctx, `UPDATE workloads SET placed_at = now(), updated_at = now() WHERE id=$1`, id)
	s.logExecError("mark_placed", err, zap.String("workload_id", id))
}

func (s *PostgresStore) GetPlacedAt(id string) (time.Time, bool) {
	if id == "" {
		return time.Time{}, false
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	row := s.pool.QueryRow(ctx, `SELECT placed_at FROM workloads WHERE id=$1`, id)
	var placed sql.NullTime
	if err := row.Scan(&placed); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.logExecError("get_placed_at", err, zap.String("workload_id", id))
		}
		return time.Time{}, false
	}
	if placed.Valid {
		return placed.Time, true
	}
	return time.Time{}, false
}

func (s *PostgresStore) ClearPlacedAt(id string) {
	if id == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	_, err := s.pool.Exec(ctx, `UPDATE workloads SET placed_at = NULL, updated_at = now() WHERE id=$1`, id)
	s.logExecError("clear_placed_at", err, zap.String("workload_id", id))
}

func (s *PostgresStore) MarkStarted(id string) {
	if id == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	_, err := s.pool.Exec(ctx, `UPDATE workloads SET started_at = now(), updated_at = now() WHERE id=$1`, id)
	s.logExecError("mark_started", err, zap.String("workload_id", id))
}

func (s *PostgresStore) GetStartedAt(id string) (time.Time, bool) {
	if id == "" {
		return time.Time{}, false
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	row := s.pool.QueryRow(ctx, `SELECT started_at FROM workloads WHERE id=$1`, id)
	var started sql.NullTime
	if err := row.Scan(&started); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.logExecError("get_started_at", err, zap.String("workload_id", id))
		}
		return time.Time{}, false
	}
	if started.Valid {
		return started.Time, true
	}
	return time.Time{}, false
}

func (s *PostgresStore) GetRuntimeSeconds(id string) (int64, bool) {
	if id == "" {
		return 0, false
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	row := s.pool.QueryRow(ctx, `SELECT runtime_seconds FROM workloads WHERE id=$1`, id)
	var secs int64
	if err := row.Scan(&secs); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.logExecError("get_runtime_seconds", err, zap.String("workload_id", id))
		}
		return 0, false
	}
	if secs < 0 {
		secs = 0
	}
	return secs, true
}

func (s *PostgresStore) SetEstimateUSD(id string, usd float64) {
	if id == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	_, err := s.pool.Exec(ctx, `
INSERT INTO workload_estimates (workload_id, estimate_usd)
VALUES ($1, $2)
ON CONFLICT (workload_id) DO UPDATE SET estimate_usd = EXCLUDED.estimate_usd
`, id, usd)
	s.logExecError("set_estimate", err, zap.String("workload_id", id))
}

func (s *PostgresStore) PopEstimateUSD(id string) float64 {
	if id == "" {
		return 0
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	row := s.pool.QueryRow(ctx, `DELETE FROM workload_estimates WHERE workload_id=$1 RETURNING estimate_usd`, id)
	var est float64
	if err := row.Scan(&est); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.logExecError("pop_estimate", err, zap.String("workload_id", id))
		}
		return 0
	}
	return est
}

func (s *PostgresStore) StartWorkload(id string) (*aegis.Workload, time.Duration, bool, error) {
	if id == "" {
		return nil, 0, false, fmt.Errorf("workload id required")
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, 0, false, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT status, placed_at FROM workloads WHERE id=$1 FOR UPDATE`, id)
	var (
		status   string
		placedAt sql.NullTime
	)
	if err := row.Scan(&status, &placedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, 0, false, fmt.Errorf("workload %s not found", id)
		}
		return nil, 0, false, err
	}

	now := time.Now().UTC()
	switch status {
	case statusPlaced:
		if _, err := tx.Exec(ctx, `UPDATE workloads SET status=$2, started_at=$3, updated_at=now() WHERE id=$1`, id, statusRunning, now); err != nil {
			return nil, 0, false, err
		}
	case statusRunning:
		// already running - no op
	default:
		return nil, 0, false, fmt.Errorf("workload %s not in a startable state", id)
	}

	w, err := s.getWorkloadTx(ctx, tx, id)
	if err != nil {
		return nil, 0, false, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, 0, false, err
	}

	if status != statusPlaced || !placedAt.Valid {
		return w, 0, false, nil
	}
	wait := time.Since(placedAt.Time)
	if wait < 0 {
		wait = 0
	}
	return w, wait, true, nil
}

func (s *PostgresStore) AckWorkload(id, nextStatus, url string) (*aegis.Workload, error) {
	if id == "" {
		return nil, fmt.Errorf("workload id required")
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT status FROM workloads WHERE id=$1 FOR UPDATE`, id)
	var status string
	if err := row.Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("workload %s not found", id)
		}
		return nil, err
	}
	if status != statusRunning {
		return nil, fmt.Errorf("workload %s not in RUNNING state", id)
	}

	if strings.EqualFold(nextStatus, statusRunning) {
		_, err = tx.Exec(ctx, `UPDATE workloads SET url = NULLIF($2, ''), updated_at = now() WHERE id=$1`, id, url)
	} else if strings.EqualFold(nextStatus, statusSuspended) {
		_, err = tx.Exec(ctx, `UPDATE workloads
SET status=$2,
    url = NULLIF($3, ''),
    runtime_seconds = runtime_seconds + COALESCE(GREATEST(0, EXTRACT(EPOCH FROM (now() - started_at))), 0)::BIGINT,
    started_at = NULL,
    suspended_at = now(),
    suspend_reason = $4,
    updated_at = now()
WHERE id=$1`, id, nextStatus, url, "idle_timeout")
	} else {
		_, err = tx.Exec(ctx, `UPDATE workloads
SET status=$2,
    url = NULLIF($3, ''),
    runtime_seconds = runtime_seconds + COALESCE(GREATEST(0, EXTRACT(EPOCH FROM (now() - started_at))), 0)::BIGINT,
    started_at = NULL,
    updated_at = now()
WHERE id=$1`, id, nextStatus, url)
	}
	if err != nil {
		return nil, err
	}

	w, err := s.getWorkloadTx(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *PostgresStore) ResumeWorkload(id string) (*aegis.Workload, error) {
	if id == "" {
		return nil, fmt.Errorf("workload id required")
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT status, resume_count FROM workloads WHERE id=$1 FOR UPDATE`, id)
	var (
		status      string
		resumeCount int32
	)
	if err := row.Scan(&status, &resumeCount); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("workload %s not found", id)
		}
		return nil, err
	}
	if status != statusSuspended {
		return nil, fmt.Errorf("workload %s not in SUSPENDED state", id)
	}

	_, err = tx.Exec(ctx, `UPDATE workloads
SET status=$2,
    resume_count=$3,
    started_at=now(),
    suspended_at=NULL,
    suspend_reason=NULL,
    updated_at=now()
WHERE id=$1`,
		id,
		statusRunning,
		resumeCount+1,
	)
	if err != nil {
		return nil, err
	}

	w, err := s.getWorkloadTx(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *PostgresStore) TerminateWorkload(id, reason string) (*aegis.Workload, error) {
	if id == "" {
		return nil, fmt.Errorf("workload id required")
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT status FROM workloads WHERE id=$1 FOR UPDATE`, id)
	var status string
	if err := row.Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("workload %s not found", id)
		}
		return nil, err
	}

	switch status {
	case statusRunning, statusSuspended:
		// ok
	case statusTerminated:
		if reason != "" {
			_, _ = tx.Exec(ctx, `UPDATE workloads
SET terminate_reason = COALESCE(NULLIF(terminate_reason, ''), NULLIF($2, '')),
    terminated_at = COALESCE(terminated_at, now()),
    updated_at = now()
WHERE id=$1`, id, reason)
		}
		w, err := s.getWorkloadTx(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return w, nil
	default:
		return nil, fmt.Errorf("workload %s not in a terminable state", id)
	}

	_, err = tx.Exec(ctx, `UPDATE workloads
SET status=$2,
    terminated_at = now(),
    terminate_reason = NULLIF($3, ''),
    updated_at = now()
WHERE id=$1`, id, statusTerminated, reason)
	if err != nil {
		return nil, err
	}

	w, err := s.getWorkloadTx(ctx, tx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *PostgresStore) RollbackTerminateWorkload(id, previousStatus string) (*aegis.Workload, error) {
	if id == "" {
		return nil, fmt.Errorf("workload id required")
	}
	if previousStatus != statusRunning && previousStatus != statusSuspended {
		return nil, fmt.Errorf("invalid rollback target status %q for workload %s", previousStatus, id)
	}

	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT status FROM workloads WHERE id=$1 FOR UPDATE`, id)
	var status string
	if err := row.Scan(&status); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("workload %s not found", id)
		}
		return nil, err
	}
	if status != statusTerminated {
		return nil, fmt.Errorf("workload %s not in TERMINATED state", id)
	}

	_, err = tx.Exec(ctx, `UPDATE workloads
SET status=$2,
    terminated_at = NULL,
    terminate_reason = NULL,
    updated_at = now()
WHERE id=$1`, id, previousStatus)
	if err != nil {
		return nil, err
	}

	w, err := s.getWorkloadTx(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return w, nil
}

func (s *PostgresStore) LeaseWorkloads(clusterID string, max int) []*aegis.Workload {
	if clusterID == "" || max <= 0 {
		return nil
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logExecError("lease_workloads_begin", err, zap.String("cluster_id", clusterID))
		return nil
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
SELECT id FROM workloads
WHERE cluster_id=$1 AND status=$2
ORDER BY created_at ASC
LIMIT $3
FOR UPDATE SKIP LOCKED
`, clusterID, statusPlaced, max)
	if err != nil {
		s.logExecError("lease_workloads_select", err, zap.String("cluster_id", clusterID))
		return nil
	}
	ids := make([]string, 0, max)
	for rows.Next() {
		var id string
		if scanErr := rows.Scan(&id); scanErr != nil {
			s.logExecError("lease_workloads_scan_id", scanErr)
			rows.Close()
			return nil
		}
		ids = append(ids, id)
	}
	rows.Close()

	if len(ids) == 0 {
		return nil
	}

	now := time.Now().UTC()
	for _, id := range ids {
		if _, err := tx.Exec(ctx, `UPDATE workloads SET status=$2, started_at=$3, updated_at = now() WHERE id=$1`, id, statusRunning, now); err != nil {
			s.logExecError("lease_workloads_update", err, zap.String("workload_id", id))
			return nil
		}
	}

	out := make([]*aegis.Workload, 0, len(ids))
	for _, id := range ids {
		w, err := s.getWorkloadTx(ctx, tx, id)
		if err != nil {
			s.logExecError("lease_workloads_fetch", err, zap.String("workload_id", id))
			return nil
		}
		out = append(out, w)
	}

	if err := tx.Commit(ctx); err != nil {
		s.logExecError("lease_workloads_commit", err, zap.String("cluster_id", clusterID))
		return nil
	}
	return out
}

func (s *PostgresStore) getWorkloadTx(ctx context.Context, tx pgx.Tx, id string) (*aegis.Workload, error) {
	row := tx.QueryRow(ctx, workloadSelect("WHERE id=$1"), id)
	w, err := s.scanWorkload(row)
	if err != nil {
		return nil, err
	}
	return w, nil
}

func workloadSelect(clause string) string {
	base := `SELECT id, project_id, queue, cluster_id, status, ui_status, url, message, kind,
        hints_resource_name, hints_gpu_count, hints_cpu_request, hints_mem_request,
        workspace_json, training_json,
        suspended_at, suspend_reason, resume_count,
        terminated_at, terminate_reason
        FROM workloads`
	if clause != "" {
		base += " " + clause
	}
	return base
}

func (s *PostgresStore) scanWorkload(row rowScanner) (*aegis.Workload, error) {
	var (
		id              string
		projectID       string
		queue           string
		cluster         sql.NullString
		status          string
		uiStatus        sql.NullString
		url             sql.NullString
		message         sql.NullString
		kind            string
		hintsRes        sql.NullString
		hintsGPU        sql.NullInt32
		hintsCPU        sql.NullString
		hintsMem        sql.NullString
		workspace       []byte
		training        []byte
		suspendedAt     sql.NullTime
		suspendReason   sql.NullString
		resumeCount     sql.NullInt32
		terminatedAt    sql.NullTime
		terminateReason sql.NullString
	)
	if err := row.Scan(
		&id,
		&projectID,
		&queue,
		&cluster,
		&status,
		&uiStatus,
		&url,
		&message,
		&kind,
		&hintsRes,
		&hintsGPU,
		&hintsCPU,
		&hintsMem,
		&workspace,
		&training,
		&suspendedAt,
		&suspendReason,
		&resumeCount,
		&terminatedAt,
		&terminateReason,
	); err != nil {
		return nil, err
	}

	w := &aegis.Workload{
		Id:        id,
		ProjectId: projectID,
		Queue:     queue,
		Status:    status,
	}
	if cluster.Valid {
		w.ClusterId = cluster.String
	}
	if uiStatus.Valid {
		w.UiStatus = uiStatus.String
	}
	if url.Valid {
		w.Url = url.String
	}
	if message.Valid {
		w.Message = message.String
	}
	if suspendedAt.Valid {
		w.SuspendedAtUtc = suspendedAt.Time.UTC().Format(time.RFC3339Nano)
	}
	if suspendReason.Valid {
		w.SuspendReason = suspendReason.String
	}
	if resumeCount.Valid {
		w.ResumeCount = int32(resumeCount.Int32)
	}
	if terminatedAt.Valid {
		w.TerminatedAtUtc = terminatedAt.Time.UTC().Format(time.RFC3339Nano)
	}
	if terminateReason.Valid {
		w.TerminateReason = terminateReason.String
	}

	if hintsRes.Valid || hintsGPU.Valid || hintsCPU.Valid || hintsMem.Valid {
		w.Hints = &aegis.ResourceHints{
			ResourceName:    hintsRes.String,
			GpuCount:        int32(hintsGPU.Int32),
			CpuCoresRequest: hintsCPU.String,
			MemoryRequest:   hintsMem.String,
		}
	}

	switch kind {
	case "workspace":
		if len(workspace) > 0 {
			ws := &aegis.WorkspaceSpec{}
			if err := protoUnmarshal.Unmarshal(workspace, ws); err != nil {
				s.logExecError("unmarshal_workspace", err, zap.String("workload_id", id))
			} else {
				w.Kind = &aegis.Workload_Workspace{Workspace: ws}
			}
		}
	case "training":
		if len(training) > 0 {
			tr := &aegis.TrainingSpec{}
			if err := protoUnmarshal.Unmarshal(training, tr); err != nil {
				s.logExecError("unmarshal_training", err, zap.String("workload_id", id))
			} else {
				w.Kind = &aegis.Workload_Training{Training: tr}
			}
		}
	}

	return w, nil
}

func bytesOrNil(in []byte) interface{} {
	if len(in) == 0 {
		return nil
	}
	return in
}
