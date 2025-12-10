-- +migrate Up
-- Add soft delete support for clusters table
-- This preserves audit trail and allows for potential recovery

ALTER TABLE clusters ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS deleted_by TEXT NULL;
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS deletion_reason TEXT NULL;

-- Index for efficient queries on non-deleted clusters
CREATE INDEX IF NOT EXISTS clusters_active ON clusters(id) WHERE deleted_at IS NULL;

-- Index for audit queries on deleted clusters
CREATE INDEX IF NOT EXISTS clusters_deleted ON clusters(deleted_at) WHERE deleted_at IS NOT NULL;

-- +migrate Down
DROP INDEX IF EXISTS clusters_deleted;
DROP INDEX IF EXISTS clusters_active;
ALTER TABLE clusters DROP COLUMN IF EXISTS deletion_reason;
ALTER TABLE clusters DROP COLUMN IF EXISTS deleted_by;
ALTER TABLE clusters DROP COLUMN IF EXISTS deleted_at;
