-- Add soft-delete support for clusters
ALTER TABLE clusters
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS deleted_by TEXT NULL,
    ADD COLUMN IF NOT EXISTS deletion_reason TEXT NULL;

-- Optional: index to speed up deleted vs active lookups
CREATE INDEX IF NOT EXISTS clusters_deleted_at_idx ON clusters (deleted_at);
