DROP INDEX IF EXISTS clusters_deleted_at_idx;
ALTER TABLE clusters
    DROP COLUMN IF EXISTS deletion_reason,
    DROP COLUMN IF EXISTS deleted_by,
    DROP COLUMN IF EXISTS deleted_at;
