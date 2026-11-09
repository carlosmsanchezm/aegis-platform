-- Migration: Multi-tenancy and referential integrity fixes
-- This migration addresses critical database design issues for scalability and multi-tenancy

-- ============================================================================
-- 1. Add project_id column to clusters table (CRITICAL)
-- ============================================================================
-- Previously, project association was only stored as a label in cluster_labels.
-- This violated multi-tenancy isolation and prevented proper cascade deletes.

ALTER TABLE clusters ADD COLUMN IF NOT EXISTS project_id TEXT;

-- Migrate existing project associations from labels to the new column
UPDATE clusters c
SET project_id = cl.v
FROM cluster_labels cl
WHERE cl.cluster_id = c.id
  AND cl.k = 'aegis.yourorg.dev/projectId'
  AND c.project_id IS NULL;

-- Add foreign key constraint (with SET NULL on delete to preserve cluster history)
-- Note: Using SET NULL instead of CASCADE because clusters may have workloads
-- that need to be cleaned up first
ALTER TABLE clusters
    ADD CONSTRAINT clusters_project_id_fkey
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;

-- Add index for efficient project-based queries
CREATE INDEX IF NOT EXISTS clusters_by_project ON clusters(project_id);

-- ============================================================================
-- 2. Add foreign key constraint on workloads.cluster_id
-- ============================================================================
-- Previously, workloads could reference non-existent clusters

-- First, clean up any orphaned workloads referencing non-existent clusters
UPDATE workloads w
SET cluster_id = NULL
WHERE w.cluster_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM clusters c WHERE c.id = w.cluster_id);

-- Add foreign key (SET NULL on delete - workload record preserved for history)
ALTER TABLE workloads
    ADD CONSTRAINT workloads_cluster_id_fkey
    FOREIGN KEY (cluster_id) REFERENCES clusters(id) ON DELETE SET NULL;

-- ============================================================================
-- 3. Fix queue primary key for multi-tenancy (CRITICAL)
-- ============================================================================
-- Currently queue names are globally unique, which prevents multi-tenancy.
-- Two projects cannot both have a queue named "default".

-- Step 1: Add a new composite unique constraint
ALTER TABLE queues ADD CONSTRAINT queues_project_name_unique UNIQUE (project_id, name);

-- Step 2: Update workloads to reference queues properly
-- Add queue_name column to distinguish from the global name
-- (The queue column currently stores the queue name)

-- Step 3: Update budget_usage to handle the new queue structure
-- The existing FK references budgets(project_id, queue) which is correct

-- ============================================================================
-- 4. Add foreign keys to provisioning tables
-- ============================================================================
-- These tables had no referential integrity, causing orphaned data

-- Clean up orphaned provisioning_runs
DELETE FROM provisioning_runs pr
WHERE pr.project_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM projects p WHERE p.id = pr.project_id);

DELETE FROM provisioning_runs pr
WHERE pr.cluster_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM clusters c WHERE c.id = pr.cluster_id);

-- Clean up orphaned provisioning_logs (this may take a while for large datasets)
DELETE FROM provisioning_logs pl
WHERE pl.project_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM projects p WHERE p.id = pl.project_id);

DELETE FROM provisioning_logs pl
WHERE pl.cluster_id IS NOT NULL
  AND NOT EXISTS (SELECT 1 FROM clusters c WHERE c.id = pl.cluster_id);

-- Add foreign keys with CASCADE delete
-- NOTE: Only add project_id FK, NOT cluster_id FK
-- Reason: Provisioning logs are written BEFORE the cluster exists in the database.
-- The cluster_id in provisioning logs is a logical reference that may not exist yet.
ALTER TABLE provisioning_runs
    ADD CONSTRAINT provisioning_runs_project_id_fkey
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

-- REMOVED: provisioning_runs_cluster_id_fkey - cluster doesn't exist during provisioning

ALTER TABLE provisioning_logs
    ADD CONSTRAINT provisioning_logs_project_id_fkey
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE;

-- REMOVED: provisioning_logs_cluster_id_fkey - cluster doesn't exist during provisioning

-- ============================================================================
-- 5. Add missing indexes for multi-tenant query performance
-- ============================================================================

-- Index for workloads by project and queue (budget calculations)
CREATE INDEX IF NOT EXISTS workloads_by_project_queue ON workloads(project_id, queue);

-- Index for workloads by status (scheduler queries)
CREATE INDEX IF NOT EXISTS workloads_by_status ON workloads(status);

-- Index for provisioning logs by project
CREATE INDEX IF NOT EXISTS provisioning_logs_by_project ON provisioning_logs(project_id);

-- Index for provisioning runs by project and phase
CREATE INDEX IF NOT EXISTS provisioning_runs_by_project_phase ON provisioning_runs(project_id, phase);

-- ============================================================================
-- 6. Add soft delete columns to workloads for consistency with clusters
-- ============================================================================

ALTER TABLE workloads ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ NULL;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS deleted_by TEXT NULL;
ALTER TABLE workloads ADD COLUMN IF NOT EXISTS deletion_reason TEXT NULL;

CREATE INDEX IF NOT EXISTS workloads_deleted_at_idx ON workloads(deleted_at);

-- ============================================================================
-- 7. Add annotation validation helper (stored as comment for documentation)
-- ============================================================================
-- Valid annotation keys use camelCase after the prefix:
--   aegis.yourorg.dev/awsRoleArn     (correct)
--   aegis.yourorg.dev/aws-role-arn   (incorrect - kebab-case)
--
-- Validation is enforced at the application layer in catalog.go

COMMENT ON COLUMN projects.annotations IS
'JSONB annotations. Keys must use camelCase format after the domain prefix (e.g., aegis.yourorg.dev/awsRoleArn). Kebab-case keys like aws-role-arn are invalid and will be rejected.';
