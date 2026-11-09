# Aegis Platform API — Architecture Overview

## Hub-and-Spoke Architecture

The Aegis Platform API serves as the central control plane (hub) in a hub-and-spoke architecture for multi-cluster GPU workload orchestration.

### Hub Components
- **platform-api** — gRPC+REST API server handling cluster registration, workload scheduling, budget enforcement, and infrastructure provisioning
- **proxy** — Reverse proxy for secure WebSocket tunneling to GPU workspaces
- **Keycloak** — OIDC identity provider for authentication and RBAC
- **Backstage UI** — Developer portal for workload management

### Spoke Components
- **k8s-agent** — Kubernetes operator running on each spoke cluster, reconciling workloads and reporting cluster status via gRPC heartbeats

## API Surface

- **gRPC** (port 8081) — Primary API for spoke agents and internal services
- **REST** (port 8080) — gRPC-Gateway generated REST API for UI and external clients

## Data Flow

1. Spoke clusters register with the hub via gRPC `RegisterCluster`
2. Users submit workloads via REST/gRPC API or Backstage UI
3. Platform-api schedules workloads to spoke clusters based on GPU availability, project quotas, and budget constraints
4. k8s-agent reconciles the workload CRDs on spoke clusters
5. Workspace workloads are accessible via the proxy's WebSocket tunneling

## Security Model

- FIPS 140-2 compliant (BoringCrypto) when built with `FIPS_ENABLED=true`
- Non-root execution (UID 1000)
- mTLS between hub and spoke via internal PKI (step-ca)
- OIDC authentication via Keycloak
- RBAC with project-scoped permissions
