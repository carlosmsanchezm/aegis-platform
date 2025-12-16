# ProjectInfra Deletion Behavior

## Expected Lifecycle

When deleting a ProjectInfra:
- **Normal delete:** `kubectl delete projectinfra <name>` should leave the finalizer in place. The controller runs Pulumi **destroy**, tearing down the full EKS stack (cluster, nodegroups, roles, SGs, etc.).
- **Do not remove finalizers or force-delete:** Stripping the finalizer or force-deleting the CR bypasses destroy and leaves the cluster in AWS (nodegroups may be removed manually, but the cluster persists).
- **Values file requirement:** The platform-api deployment must have a valid `AEGIS_SPOKE_VALUES_FILE` path. Current value: `/root/charts/aegis-spoke/values-cloud-remote.yaml`. A missing/incorrect path causes apply/destroy to fail.

## Architecture (Fixed Dec 2025)

### Per-Infra Pulumi Stacks
Each ProjectInfra now has its **own isolated Pulumi stack** instead of sharing one per project/region:
- **Stack naming:** `aegis-{project}-{region}-{infra-name-hash}`
- **Benefit:** Deleting one ProjectInfra no longer affects others in the same project/region
- **Backward compatibility:** Destroy attempts to find the infra-specific stack first, falling back to legacy shared stack if not found

### Per-Infra Locking
Kubernetes Lease locks are now per-ProjectInfra:
- **Lease naming:** `pi-{project}-{region}-{infra-name}`
- **Benefit:** Provisioning one ProjectInfra doesn't block deletion of another

### Key Files Changed
- `internal/provisioning/pulumi/aws/runner.go`: `stackNameForInfra()` function creates isolated stacks
- `internal/controllers/projectinfra_controller.go`: `stackLockKey()` and `leaseNameForInfra()` use infra-specific keys

## Frontend/Backend Integration

- UI-driven deletes should perform a normal CR delete, letting the finalizer run Pulumi destroy. This ensures the entire EKS cluster is removed.
- If a ProjectInfra was force-deleted, the controller cannot destroy the cluster; a manual AWS delete (or recreate-and-delete with finalizer intact) is needed.

## Troubleshooting

### Cluster not deleting
1. Check if ProjectInfra has `deletionTimestamp` set: `kubectl get projectinfra <name> -o yaml | grep deletionTimestamp`
2. Check platform-api logs for destroy progress: `kubectl logs -n aegis-system -l app.kubernetes.io/name=platform-api`
3. If stuck in provision operation, restart platform-api pod to interrupt it

### Manual cleanup required
If deletion fails and clusters remain in AWS:
1. Delete node groups first: `aws eks list-nodegroups --cluster-name <name> --region <region>` then delete each
2. Delete cluster: `aws eks delete-cluster --name <name> --region <region>`
3. Remove finalizer from ProjectInfra: `kubectl patch projectinfra <name> -n aegis-system -p '{"metadata":{"finalizers":null}}' --type=merge`

## Current State Reminders

- Platform-api is rolled to `carlosmsanchez/aegis-platform-api:dev` and has `AEGIS_SPOKE_VALUES_FILE=/root/charts/aegis-spoke/values-cloud-remote.yaml`.
- For new provisions/deletes, keep the finalizer intact so the controller performs full cleanup.
- Each ProjectInfra should now have isolated Pulumi state - verify by checking stack names in Pulumi backend.
