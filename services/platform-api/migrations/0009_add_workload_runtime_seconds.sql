-- +migrate Up
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS runtime_seconds BIGINT NOT NULL DEFAULT 0;

-- +migrate Down
ALTER TABLE workloads DROP COLUMN IF EXISTS runtime_seconds;
