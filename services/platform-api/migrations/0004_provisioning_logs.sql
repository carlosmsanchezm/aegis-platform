-- Provisioning run metadata
CREATE TABLE IF NOT EXISTS provisioning_runs (
    job_id TEXT PRIMARY KEY,
    project_id TEXT,
    cluster_id TEXT,
    phase TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Provisioning log lines
CREATE TABLE IF NOT EXISTS provisioning_logs (
    id BIGSERIAL PRIMARY KEY,
    job_id TEXT NOT NULL,
    project_id TEXT,
    cluster_id TEXT,
    phase TEXT,
    entry_type TEXT,
    message TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_provisioning_logs_job_created ON provisioning_logs (job_id, created_at, id);
