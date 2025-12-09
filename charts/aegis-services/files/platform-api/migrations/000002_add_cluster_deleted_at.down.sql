DROP INDEX IF EXISTS clusters_deleted_at_idx;
ALTER TABLE clusters DROP COLUMN IF EXISTS deleted_at;
