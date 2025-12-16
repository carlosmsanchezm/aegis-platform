-- +migrate Up
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS suspended_at TIMESTAMPTZ NULL;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS suspend_reason TEXT NULL;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS resume_count INTEGER NOT NULL DEFAULT 0;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS terminated_at TIMESTAMPTZ NULL;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS terminate_reason TEXT NULL;

-- +migrate Down
ALTER TABLE workloads DROP COLUMN IF EXISTS terminate_reason;
ALTER TABLE workloads DROP COLUMN IF EXISTS terminated_at;
ALTER TABLE workloads DROP COLUMN IF EXISTS resume_count;
ALTER TABLE workloads DROP COLUMN IF EXISTS suspend_reason;
ALTER TABLE workloads DROP COLUMN IF EXISTS suspended_at;
