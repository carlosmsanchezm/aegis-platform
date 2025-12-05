# ProjectInfra Deletion Behavior

Expected lifecycle when deleting a ProjectInfra:
- **Normal delete:** `kubectl delete projectinfra <name>` should leave the finalizer in place. The controller runs Pulumi **destroy**, tearing down the full EKS stack (cluster, nodegroups, roles, SGs, etc.).
- **Do not remove finalizers or force-delete:** Stripping the finalizer or force-deleting the CR bypasses destroy and leaves the cluster in AWS (nodegroups may be removed manually, but the cluster persists).
- **Values file requirement:** The platform-api deployment must have a valid `AEGIS_SPOKE_VALUES_FILE` path. Current value: `/root/charts/aegis-spoke/values-cloud-remote.yaml`. A missing/incorrect path causes apply/destroy to fail.

Frontend/Backend integration notes:
- UI-driven deletes should perform a normal CR delete, letting the finalizer run Pulumi destroy. This ensures the entire EKS cluster is removed.
- If a ProjectInfra was force-deleted, the controller cannot destroy the cluster; a manual AWS delete (or recreate-and-delete with finalizer intact) is needed.

Current state reminders:
- Platform-api is rolled to `carlosmsanchez/aegis-platform-api:dev` and has `AEGIS_SPOKE_VALUES_FILE=/root/charts/aegis-spoke/values-cloud-remote.yaml`.
- For new provisions/deletes, keep the finalizer intact so the controller performs full cleanup.
