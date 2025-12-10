-- Idempotent soft-delete columns for clusters (apply even if previous migration numbers were skipped).
ALTER TABLE clusters
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS deleted_by TEXT NULL,
    ADD COLUMN IF NOT EXISTS deletion_reason TEXT NULL;

CREATE INDEX IF NOT EXISTS clusters_deleted_at_idx ON clusters (deleted_at);
