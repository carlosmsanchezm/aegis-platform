-- Migration: Backfill cluster project_id from labels and cluster ID patterns
-- This is idempotent — safe to re-run.

-- 1. Backfill project_id from cluster_labels where the label exists but column is NULL
UPDATE clusters c
SET project_id = cl.v
FROM cluster_labels cl
WHERE cl.cluster_id = c.id
  AND cl.k = 'aegis.yourorg.dev/projectId'
  AND cl.v IS NOT NULL
  AND cl.v != ''
  AND c.project_id IS NULL;

-- 2. Backfill using cluster ID region-pattern parsing, only where derived value
--    exists in the projects table (FK safe)
UPDATE clusters c
SET project_id = derived.pid
FROM (
    SELECT c2.id AS cluster_id,
           CASE
               WHEN position('-us-east-' in c2.id) > 0 THEN left(c2.id, position('-us-east-' in c2.id) - 1)
               WHEN position('-us-west-' in c2.id) > 0 THEN left(c2.id, position('-us-west-' in c2.id) - 1)
               WHEN position('-eu-west-' in c2.id) > 0 THEN left(c2.id, position('-eu-west-' in c2.id) - 1)
               WHEN position('-eu-central-' in c2.id) > 0 THEN left(c2.id, position('-eu-central-' in c2.id) - 1)
               WHEN position('-ap-southeast-' in c2.id) > 0 THEN left(c2.id, position('-ap-southeast-' in c2.id) - 1)
               WHEN position('-ap-northeast-' in c2.id) > 0 THEN left(c2.id, position('-ap-northeast-' in c2.id) - 1)
               WHEN position('-ap-south-' in c2.id) > 0 THEN left(c2.id, position('-ap-south-' in c2.id) - 1)
               WHEN position('-sa-east-' in c2.id) > 0 THEN left(c2.id, position('-sa-east-' in c2.id) - 1)
               WHEN position('-ca-central-' in c2.id) > 0 THEN left(c2.id, position('-ca-central-' in c2.id) - 1)
               WHEN position('-me-south-' in c2.id) > 0 THEN left(c2.id, position('-me-south-' in c2.id) - 1)
               WHEN position('-af-south-' in c2.id) > 0 THEN left(c2.id, position('-af-south-' in c2.id) - 1)
               WHEN position('-il-central-' in c2.id) > 0 THEN left(c2.id, position('-il-central-' in c2.id) - 1)
               ELSE NULL
           END AS pid
    FROM clusters c2
    WHERE c2.project_id IS NULL
) derived
WHERE c.id = derived.cluster_id
  AND derived.pid IS NOT NULL
  AND EXISTS (SELECT 1 FROM projects p WHERE p.id = derived.pid);

-- 3. Insert aegis.yourorg.dev/projectId label for any backfilled rows that don't have it
INSERT INTO cluster_labels (cluster_id, k, v)
SELECT c.id, 'aegis.yourorg.dev/projectId', c.project_id
FROM clusters c
WHERE c.project_id IS NOT NULL
  AND NOT EXISTS (
      SELECT 1 FROM cluster_labels cl
      WHERE cl.cluster_id = c.id
        AND cl.k = 'aegis.yourorg.dev/projectId'
  )
ON CONFLICT DO NOTHING;
