-- Migration: Multi-tenancy and referential integrity fixes
-- This migration addresses critical database design issues for scalability and multi-tenancy

-- ============================================================================
-- 1. Add project_id column to clusters table (CRITICAL)
-- ============================================================================
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS project_id TEXT;

-- Migrate existing project associations from labels
UPDATE clusters c
SET project_id = cl.v
FROM cluster_labels cl
WHERE cl.cluster_id = c.id
  AND cl.k = 'aegis.yourorg.dev/projectId'
  AND c.project_id IS NULL;

-- Add foreign key constraint
ALTER TABLE clusters
    ADD CONSTRAINT clusters_project_id_fkey
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS clusters_by_project ON clusters(project_id);

-- ============================================================================
-- 2. Add foreign key constraint on workloads.cluster_id
-- ============================================================================
UPDATE workloads w
SET cluster_id = NULL
WHERE w.cluster_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM clusters c WHERE c.id = w.cluster_id);

ALTER TABLE workloads
    ADD CONSTRAINT workloads_cluster_id_fkey
    FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE SET NULL;

-- ============================================================================
-- 3. Fix queue primary key for multi-tenancy
-- ============================================================================
ALTER TABLE queues ADD CONSTRAINT queues_project_name_unique UNIQUE (project_id, name);

-- ============================================================================
-- 4. Add foreign keys to provisioning tables
-- ============================================================================
DELETE FROM provisioning_runs pr
WHERE pr.project_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM projects p WHERE p.id = pr.project_id);

DELETE FROM provisioning_runs pr
WHERE pr.cluster_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM clusters c WHERE c.id = pr.cluster_id);

DELETE FROM provisioning_logs pl
WHERE pl.project_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM projects p WHERE p.id = pl.project_id);

DELETE FROM provisioning_logs pl
WHERE pl.cluster_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM clusters c WHERE c.id = pl.cluster_id);

-- NOTE: Only add project_id FK, NOT cluster_id FK
-- Reason: Provisioning logs are written BEFORE the cluster exists in the database.
ALTER TABLE provisioning_runs
    ADD CONSTRAINT provisioning_runs_project_id_fkey
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

-- REMOVED: provisioning_runs_cluster_id_fkey - cluster doesn't exist during provisioning

ALTER TABLE provisioning_logs
    ADD CONSTRAINT provisioning_logs_project_id_fkey
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

-- REMOVED: provisioning_logs_cluster_id_fkey - cluster doesn't exist during provisioning

-- ============================================================================
-- 5. Add missing indexes
-- ============================================================================
CREATE INDEX IF NOT EXISTS workloads_by_project_queue ON workloads(project_id, queue);
CREATE INDEX IF NOT EXISTS workloads_by_status ON workloads(status);
CREATE INDEX IF NOT EXISTS provisioning_logs_by_project ON provisioning_logs(project_id);
CREATE INDEX IF NOT EXISTS provisioning_runs_by_project_phase ON provisioning_runs(project_id, phase);

-- ============================================================================
-- 6. Add soft delete columns to workloads
-- ============================================================================
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS deleted_by TEXT NULL;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS deletion_reason TEXT NULL;

CREATE INDEX IF NOT EXISTS workloads_deleted_at_idx ON workloads(deleted_at);

COMMENT ON COLUMN projects.annotations IS
'JSONB annotations. Keys must use camelCase format after the domain prefix (e.g., aegis.yourorg.dev/awsRoleArn). Kebab-case keys like aws-role-arn are invalid and will be rejected.';
