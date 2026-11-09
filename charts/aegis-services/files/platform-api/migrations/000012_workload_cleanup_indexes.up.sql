CREATE INDEX IF NOT EXISTS workloads_by_status_updated
    ON workloads(status, updated_at)
    WHERE status IN ('RUNNING', 'PLACED', 'SUSPENDED');

CREATE INDEX IF NOT EXISTS workloads_terminated_at
    ON workloads(terminated_at)
    WHERE status = 'TERMINATED';
