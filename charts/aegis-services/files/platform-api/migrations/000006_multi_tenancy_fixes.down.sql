-- Rollback: Multi-tenancy and referential integrity fixes

-- Remove soft delete columns from workloads
DROP INDEX IF EXISTS workloads_deleted_at_idx;
ALTER TABLE workloads DROP COLUMN IF EXISTS deletion_reason;
ALTER TABLE workloads DROP COLUMN IF EXISTS deleted_by;
ALTER TABLE workloads DROP COLUMN IF EXISTS deleted_at;

-- Remove performance indexes
DROP INDEX IF EXISTS provisioning_runs_by_project_phase;
DROP INDEX IF EXISTS provisioning_logs_by_project;
DROP INDEX IF EXISTS workloads_by_status;
DROP INDEX IF EXISTS workloads_by_project_queue;

-- Remove provisioning foreign keys
ALTER TABLE provisioning_logs DROP CONSTRAINT IF EXISTS provisioning_logs_cluster_id_fkey;
ALTER TABLE provisioning_logs DROP CONSTRAINT IF EXISTS provisioning_logs_project_id_fkey;
ALTER TABLE provisioning_runs DROP CONSTRAINT IF EXISTS provisioning_runs_cluster_id_fkey;
ALTER TABLE provisioning_runs DROP CONSTRAINT IF EXISTS provisioning_runs_project_id_fkey;

-- Remove queue composite constraint
ALTER TABLE queues DROP CONSTRAINT IF EXISTS queues_project_name_unique;

-- Remove workloads cluster foreign key
ALTER TABLE workloads DROP CONSTRAINT IF EXISTS workloads_cluster_id_fkey;

-- Remove clusters project foreign key and column
DROP INDEX IF EXISTS clusters_by_project;
ALTER TABLE clusters DROP CONSTRAINT IF EXISTS clusters_project_id_fkey;
ALTER TABLE clusters DROP COLUMN IF EXISTS project_id;

-- Remove annotation comment
COMMENT ON COLUMN projects.annotations IS NULL;
