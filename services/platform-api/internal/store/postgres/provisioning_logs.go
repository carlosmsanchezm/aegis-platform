package postgres

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

func (s *PostgresStore) ensureProvisioningTables(ctx context.Context) error {
	ctx, cancel := s.withTimeout(ctx)
	defer cancel()
	statements := []string{
		`CREATE TABLE IF NOT EXISTS provisioning_runs (
    job_id TEXT PRIMARY KEY,
    project_id TEXT,
    cluster_id TEXT,
    phase TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE TABLE IF NOT EXISTS provisioning_logs (
    id BIGSERIAL PRIMARY KEY,
    job_id TEXT NOT NULL,
    project_id TEXT,
    cluster_id TEXT,
    phase TEXT,
    entry_type TEXT,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
)`,
		`CREATE INDEX IF NOT EXISTS idx_provisioning_logs_job_created ON provisioning_logs (job_id, created_at, id)`,
	}
	for _, stmt := range statements {
		if _, err := s.pool.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *PostgresStore) AppendProvisioningLog(entry store.ProvisioningLogEntry) {
	if strings.TrimSpace(entry.JobID) == "" || strings.TrimSpace(entry.Message) == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	created := entry.CreatedAt
	if created.IsZero() {
		created = time.Now().UTC()
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO provisioning_logs (job_id, project_id, cluster_id, phase, entry_type, message, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.JobID,
		nullableString(entry.ProjectID),
		nullableString(entry.ClusterID),
		nullableString(entry.Phase),
		nullableString(entry.Type),
		entry.Message,
		created,
	)
	s.logExecError("provisioning_log_insert", err, zap.String("job_id", entry.JobID))
}

func (s *PostgresStore) ListProvisioningLogs(jobID string, since time.Time, sinceSeq int64, limit int) []store.ProvisioningLogEntry {
	if strings.TrimSpace(jobID) == "" {
		return nil
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	if limit <= 0 {
		limit = 200
	}
	if limit > 1000 {
		limit = 1000
	}

	query := `SELECT job_id, COALESCE(project_id, ''), COALESCE(cluster_id, ''), COALESCE(phase, ''), COALESCE(entry_type, ''), message, created_at, id FROM provisioning_logs WHERE job_id=$1`
	args := []any{jobID}
	if !since.IsZero() {
		if sinceSeq > 0 {
			query += " AND (created_at > $" + strconv.Itoa(len(args)+1) + " OR (created_at = $" + strconv.Itoa(len(args)+1) + " AND id > $" + strconv.Itoa(len(args)+2) + "))"
			args = append(args, since, sinceSeq)
		} else {
			query += " AND created_at > $" + strconv.Itoa(len(args)+1)
			args = append(args, since)
		}
	}
	query += " ORDER BY created_at ASC, id ASC LIMIT $" + strconv.Itoa(len(args)+1)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		s.logExecError("provisioning_log_list", err, zap.String("job_id", jobID))
		return nil
	}
	defer rows.Close()

	out := make([]store.ProvisioningLogEntry, 0, limit)
	for rows.Next() {
		var entry store.ProvisioningLogEntry
		var projectID, clusterID, phase, entryType string
		if scanErr := rows.Scan(&entry.JobID, &projectID, &clusterID, &phase, &entryType, &entry.Message, &entry.CreatedAt, &entry.Sequence); scanErr != nil {
			s.logExecError("provisioning_log_scan", scanErr)
			break
		}
		entry.ProjectID = strings.TrimSpace(projectID)
		entry.ClusterID = strings.TrimSpace(clusterID)
		entry.Phase = strings.TrimSpace(phase)
		entry.Type = strings.TrimSpace(entryType)
		out = append(out, entry)
	}
	return out
}

func (s *PostgresStore) ClearProvisioningLogs(jobID string) {
	if strings.TrimSpace(jobID) == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	result, err := s.pool.Exec(ctx, `DELETE FROM provisioning_logs WHERE job_id = $1`, jobID)
	if err != nil {
		s.logExecError("provisioning_logs_clear", err, zap.String("job_id", jobID))
		return
	}
	if s.log != nil {
		s.log.Info("cleared provisioning logs", zap.String("job_id", jobID), zap.Int64("deleted_count", result.RowsAffected()))
	}
}

func (s *PostgresStore) DeleteOldProvisioningLogs(olderThan time.Time) int64 {
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	result, err := s.pool.Exec(ctx, `DELETE FROM provisioning_logs WHERE created_at < $1`, olderThan)
	if err != nil {
		s.logExecError("provisioning_logs_retention_cleanup", err, zap.Time("older_than", olderThan))
		return 0
	}
	return result.RowsAffected()
}

func (s *PostgresStore) UpsertProvisioningRun(run store.ProvisioningRun) {
	if strings.TrimSpace(run.JobID) == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	started := run.StartedAt
	if started.IsZero() {
		started = time.Now().UTC()
	}
	_, err := s.pool.Exec(ctx, `INSERT INTO provisioning_runs (job_id, project_id, cluster_id, phase, started_at, completed_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, now())
ON CONFLICT (job_id) DO UPDATE SET
    project_id = COALESCE(EXCLUDED.project_id, provisioning_runs.project_id),
    cluster_id = COALESCE(EXCLUDED.cluster_id, provisioning_runs.cluster_id),
    phase = COALESCE(EXCLUDED.phase, provisioning_runs.phase),
    started_at = EXCLUDED.started_at,
    completed_at = EXCLUDED.completed_at,
    updated_at = now()
`, run.JobID, nullableString(run.ProjectID), nullableString(run.ClusterID), nullableString(run.Phase), started, nullableTime(run.CompletedAt))
	s.logExecError("provisioning_run_upsert", err, zap.String("job_id", run.JobID))
}

func (s *PostgresStore) GetProvisioningRun(jobID string) (*store.ProvisioningRun, bool) {
	if strings.TrimSpace(jobID) == "" {
		return nil, false
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	row := s.pool.QueryRow(ctx, `SELECT job_id, COALESCE(project_id, ''), COALESCE(cluster_id, ''), COALESCE(phase, ''), started_at, completed_at, updated_at FROM provisioning_runs WHERE job_id=$1`, jobID)
	var (
		run                  store.ProvisioningRun
		projectID, clusterID string
		phase                string
		started, updated     time.Time
		completed            sql.NullTime
	)
	if err := row.Scan(&run.JobID, &projectID, &clusterID, &phase, &started, &completed, &updated); err != nil {
		if err != pgx.ErrNoRows {
			s.logExecError("provisioning_run_get", err, zap.String("job_id", jobID))
		}
		return nil, false
	}
	run.ProjectID = strings.TrimSpace(projectID)
	run.ClusterID = strings.TrimSpace(clusterID)
	run.Phase = strings.TrimSpace(phase)
	run.StartedAt = started
	run.UpdatedAt = updated
	if completed.Valid {
		run.CompletedAt = &completed.Time
	}
	return &run, true
}
