# Troubleshooting: Cluster Dropdown Not Showing Clusters

This document describes how to fix the issue where the "Target Cluster" dropdown in the workspace creation page (`/aegis/workspaces/create`) shows no clusters.

## Symptoms

- The "Target Cluster" dropdown is empty or disabled
- Browser console shows: `{"code":12, "message":"Method Not Allowed", "details":[]}`
- HTTP status 501 Not Implemented on `/api/proxy/aegis/api/v1/clusters`

## Root Causes & Fixes

### 1. Missing Project Label on Cluster

**Most common cause.** Clusters must have the `aegis.yourorg.dev/projectId` label to appear in the dropdown for a specific project.

#### How to Check

```bash
# Check cluster labels in the database
kubectl exec -n aegis-system platform-postgres-0 -- \
  psql -U aegis_platform -d aegis_platform -c "SELECT * FROM cluster_labels;"
```

Look for a row with:
- `cluster_id`: your cluster ID (e.g., `db-1-us-east-1-atlas-train-govcloud`)
- `k`: `aegis.yourorg.dev/projectId`
- `v`: your project ID (e.g., `db-1`)

#### How to Fix

```bash
# Add the projectId label to the cluster
kubectl exec -n aegis-system platform-postgres-0 -- \
  psql -U aegis_platform -d aegis_platform -c \
  "INSERT INTO cluster_labels (cluster_id, k, v) VALUES ('<CLUSTER_ID>', 'aegis.yourorg.dev/projectId', '<PROJECT_ID>') ON CONFLICT DO NOTHING;"
```

**Example:**
```bash
kubectl exec -n aegis-system platform-postgres-0 -- \
  psql -U aegis_platform -d aegis_platform -c \
  "INSERT INTO cluster_labels (cluster_id, k, v) VALUES ('db-1-us-east-1-atlas-train-govcloud', 'aegis.yourorg.dev/projectId', 'db-1') ON CONFLICT DO NOTHING;"
```

### Auto-Derivation from Cluster ID

As of the project-ID auto-derivation feature, the server automatically extracts and persists the project association from the cluster ID format (`{projectId}-{region}-{clusterName}-{suffix}`) at registration and heartbeat time. This means:

- **Manual label insertion is no longer needed** for clusters with standard AWS-format IDs (e.g., `db-1-us-east-1-atlas-train-govcloud` → project `db-1`)
- The server validates that the derived project exists in the `projects` table before persisting
- Heartbeat acts as self-healing: if a cluster registered before its project was created, the next heartbeat will backfill the association
- Explicit `AEGIS_PROJECT_ID` env var or `aegis.yourorg.dev/projectId` label still takes priority over auto-derivation

If a cluster still doesn't appear in the dropdown after auto-derivation was deployed, check:
1. The cluster ID follows the `{projectId}-{region}-...` format
2. The project exists in the `projects` table
3. The cluster has sent at least one heartbeat since the feature was deployed

### 2. Query Parameter Case Mismatch

The frontend must send `project_id` (snake_case), not `projectId` (camelCase).

#### How to Check

In browser DevTools Network tab, check the request URL:
- **Correct:** `/api/v1/clusters?project_id=db-1`
- **Wrong:** `/api/v1/clusters?projectId=db-1`

#### How to Fix

In `aegis-ui/plugins/aegis/src/api/aegisClient.ts`, ensure the `listClusters` function uses snake_case:

```typescript
if (options?.projectId) {
  params.append('project_id', options.projectId);  // Must be snake_case
}
```

### 3. Route Conflict in Platform API

The wizard handlers must NOT register a handler for `/api/v1/clusters` as this conflicts with the grpc-gateway.

#### How to Check

```bash
kubectl logs -n aegis-system deploy/aegis-services-platform-api | grep "workspace wizard routes"
```

The output should NOT include `/api/v1/clusters`:
```
"routes":["/api/projects","/api/clusters","/api/workspaces","/aegis/api/projects","/aegis/api/clusters","/aegis/api/workspaces"]
```

#### How to Fix

In `aegis-platform/services/platform-api/internal/server/wizard_handlers.go`, ensure there is NO `HandlePath` registration for `/api/v1/clusters`. The grpc-gateway handles this route automatically.

### 4. Cluster Not Registered or Heartbeat Stale

Clusters must be registered and sending heartbeats to appear.

#### How to Check

```bash
# Check clusters in database
kubectl exec -n aegis-system platform-postgres-0 -- \
  psql -U aegis_platform -d aegis_platform -c "SELECT id, provider, region, last_heartbeat FROM clusters;"
```

Verify:
- The cluster exists
- `last_heartbeat` is recent (within last 5 minutes for "Ready" status)

## Registering a New Cluster with Project Label

When registering a new cluster, ensure the spoke agent sends the project ID label in the heartbeat. The label should be set in the Helm values:

```yaml
# charts/aegis-spoke/values.yaml
k8sAgent:
  env:
    AEGIS_CLUSTER_ID: "my-cluster-id"
    AEGIS_REGION: "us-east-1"
    AEGIS_PROVIDER: "aws"
  labels:
    aegis.yourorg.dev/projectId: "my-project-id"
```

## Quick Diagnostic Commands

```bash
# 1. Check if clusters exist
kubectl exec -n aegis-system platform-postgres-0 -- \
  psql -U aegis_platform -d aegis_platform -c "SELECT id, last_heartbeat FROM clusters;"

# 2. Check cluster labels
kubectl exec -n aegis-system platform-postgres-0 -- \
  psql -U aegis_platform -d aegis_platform -c "SELECT * FROM cluster_labels WHERE k = 'aegis.yourorg.dev/projectId';"

# 3. Check platform-api logs for errors
kubectl logs -n aegis-system deploy/aegis-services-platform-api --tail=50 | grep -E "error|Error|ERROR"

# 4. Test the API directly (should return Unauthenticated, not Method Not Allowed)
curl -s http://localhost:10080/api/v1/clusters

# 5. Verify port-forward is running
lsof -i :10080
```

## Related Files

- `aegis-ui/plugins/aegis/src/api/aegisClient.ts` - Frontend API client
- `aegis-ui/plugins/aegis/src/components/LaunchWorkspacePage.tsx` - Workspace creation page
- `aegis-platform/services/platform-api/internal/server/server.go` - ListClusters implementation
- `aegis-platform/services/platform-api/internal/server/wizard_handlers.go` - HTTP route handlers
- `aegis-platform/proto/aegis/v1/platform.proto` - Proto definitions
