-- +migrate Up
ALTER TABLE clusters
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS deleted_by TEXT NULL,
    ADD COLUMN IF NOT EXISTS deletion_reason TEXT NULL;

CREATE INDEX IF NOT EXISTS clusters_deleted_at_idx ON clusters (deleted_at);

-- +migrate Down
DROP INDEX IF EXISTS clusters_deleted_at_idx;
ALTER TABLE clusters
    DROP COLUMN IF EXISTS deletion_reason,
    DROP COLUMN IF EXISTS deleted_by,
    DROP COLUMN IF EXISTS deleted_at;
