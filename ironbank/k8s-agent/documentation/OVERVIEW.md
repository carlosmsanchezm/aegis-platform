# Aegis K8s Agent — Architecture Overview

## Role in Hub-and-Spoke Architecture

The k8s-agent is the spoke-side component of the Aegis Platform. It runs as a Kubernetes operator on each managed cluster and communicates with the hub's platform-api via gRPC.

## Custom Resource Definitions

- **AegisWorkload** — Represents a scheduled workload (workspace or training job)
- **Workspace** — GPU development environment with VS Code Remote Extension Host
- **WorkspaceClass** — Template defining resource limits and GPU flavors

## Reconciliation Loop

1. Hub schedules a workload → creates AegisWorkload CRD on spoke
2. k8s-agent detects the CRD and creates the underlying Kubernetes resources (Pod, Service, PVC)
3. Agent monitors pod status and reports back to hub via heartbeat
4. On workload deletion, agent cleans up all associated resources

## Security

- Runs as non-root (UID 65532)
- FIPS 140-2 compliant (BoringCrypto)
- mTLS communication with hub via internal PKI
- RBAC scoped to aegis-system namespace
