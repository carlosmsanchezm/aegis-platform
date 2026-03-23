# Aegis Platform

**Compliance-first GPU control plane for regulated environments**

Aegis is a multi-cluster Kubernetes control plane that orchestrates GPU workloads across distributed clusters while enforcing FedRAMP High and IL-4/5 compliance boundaries. It delivers p95 < 90s time-to-first-GPU through intelligent placement, Kueue-based queuing, and automated infrastructure provisioning.

## Architecture

```
                        ┌─────────────────────────────────────┐
                        │           Aegis Hub Cluster          │
                        │                                     │
  ┌──────────┐          │  ┌──────────────┐  ┌─────────────┐  │
  │ Aegis UI │──REST──▶ │  │ Platform API │  │  PostgreSQL  │  │
  │(Backstage)│         │  │ (gRPC+REST)  │◀─┤             │  │
  └──────────┘          │  └──────┬───────┘  └─────────────┘  │
                        │         │                           │
  ┌──────────┐          │  ┌──────▼───────┐  ┌─────────────┐  │
  │  Sovran  │──gRPC──▶ │  │   Keycloak   │  │ Cert-Manager│  │
  │(VS Code) │          │  │   (OIDC)     │  │  + Step CA  │  │
  └──────────┘          │  └──────────────┘  └─────────────┘  │
                        └─────────────┬───────────────────────┘
                                      │ gRPC (mTLS)
              ┌───────────────────────┼───────────────────────┐
              ▼                       ▼                       ▼
  ┌───────────────────┐  ┌───────────────────┐  ┌───────────────────┐
  │  Spoke Cluster A  │  │  Spoke Cluster B  │  │  Spoke Cluster N  │
  │                   │  │                   │  │                   │
  │  ┌─────────────┐  │  │  ┌─────────────┐  │  │  ┌─────────────┐  │
  │  │  K8s Agent   │  │  │  │  K8s Agent   │  │  │  │  K8s Agent   │  │
  │  │  (Operator)  │  │  │  │  (Operator)  │  │  │  │  (Operator)  │  │
  │  └──────┬──────┘  │  │  └──────┬──────┘  │  │  └──────┬──────┘  │
  │         ▼         │  │         ▼         │  │         ▼         │
  │  ┌─────────────┐  │  │  ┌─────────────┐  │  │  ┌─────────────┐  │
  │  │ Kueue Queue │  │  │  │ Kueue Queue │  │  │  │ Kueue Queue │  │
  │  └──────┬──────┘  │  │  └──────┬──────┘  │  │  └──────┬──────┘  │
  │         ▼         │  │         ▼         │  │         ▼         │
  │  ┌─────────────┐  │  │  ┌─────────────┐  │  │  ┌─────────────┐  │
  │  │  Workload   │  │  │  │  Workload   │  │  │  │  Workload   │  │
  │  │    Pods     │  │  │  │    Pods     │  │  │  │    Pods     │  │
  │  │  (GPU/CPU)  │  │  │  │  (GPU/CPU)  │  │  │  │  (GPU/CPU)  │  │
  │  └─────────────┘  │  │  └─────────────┘  │  │  └─────────────┘  │
  └───────────────────┘  └───────────────────┘  └───────────────────┘
```

## System Components

| Component | Repository | Stack | Purpose |
|-----------|-----------|-------|---------|
| **Platform API** | this repo | Go, gRPC + REST gateway | Central orchestration: workload placement, budget enforcement, cluster management, proxy ticket minting |
| **K8s Agent** | this repo | Go, controller-runtime | Per-cluster operator that reconciles `AegisWorkload` CRDs into Kubernetes Jobs with Kueue integration |
| **Proxy** | this repo | Go, WebSocket | Authenticated reverse proxy for VS Code workspace connections (JWT + mTLS) |
| **Aegis UI** | [aegis-ui](https://github.com/carlosmsanchezm/aegis-ui) | TypeScript, Backstage | Web frontend source for workload submission, monitoring, FinOps dashboards, and administration. The cloud runtime/deployment path is owned by this repo’s `aegis-services` chart. |
| **Sovran** | [sovran](https://github.com/carlosmsanchezm/sovran) | TypeScript, VS Code API | VS Code extension for connecting to remote GPU workspaces via the Aegis proxy |

## Key Features

- **Multi-cluster GPU scheduling** -- Central API places workloads across spoke clusters based on flavor availability, heartbeat freshness, and time-to-first-GPU metrics
- **Budget enforcement** -- Per-queue spending limits with HARD (reject) and SOFT (warn) policy modes, cost estimation from GPU-hour pricing
- **Kueue integration** -- Jobs are created suspended with queue annotations; Kueue handles admission and resource fairness
- **Interactive workspaces** -- Submit workspace workloads with port mappings, connect via VS Code through authenticated WebSocket tunnels
- **NIST 800-171 R3 compliance** -- Session management (AC-11/AC-12), audit logging, OSCAL evidence export, policy domain enforcement with region and IL-level constraints
- **Automated infrastructure** -- Pulumi-based AWS EKS spoke cluster provisioning with cert-manager PKI, AWS NLB relay connectivity
- **Multi-tenant projects** -- Project isolation with per-project AWS credentials, policy domains, and data classification levels

## Deployment

For the current deployment workflows and the verified production-cloud auth/UI surface, use:

- `AGENT_DEPLOYMENT_GUIDE.md` for local, hybrid, and full-cloud rollout steps
- `docs/security/auth.md` for the current cloud Backstage/Keycloak auth contract

### Prerequisites

- Go 1.24+
- Docker Desktop with Kubernetes enabled
- PostgreSQL 15+ (or use in-memory store for development)
- Keycloak (deployed via Helm chart)
- `kubectl`, `helm`, `grpcurl`

### Local Development

```bash
# Apply CRDs and prepare kubeconfig
kubectl apply -f agents/k8s-agent/config/crd/bases/aegis.yourorg.dev_aegisworkloads.yaml
cp ~/.kube/config /tmp/dev-1.kubeconfig

# Terminal 1: Platform API
export KUBECONFIGS_DIR=/tmp
make run-api ALLOW_SOCKETS=1

# Terminal 2: K8s Agent (operator)
AEGIS_DISABLE_KUEUE=1 \
AEGIS_CLUSTER_ID=dev-1 \
AEGIS_FLAVORS="cpu-small" \
make run-operator ALLOW_SOCKETS=1

# Terminal 3: Seed and submit workloads
grpcurl -plaintext -d '{"project":{"id":"p-demo"}}' localhost:8081 aegis.v1.AegisPlatform/CreateProject
grpcurl -plaintext -d '{"flavor":{"name":"cpu-small","gpuCount":0}}' localhost:8081 aegis.v1.AegisPlatform/UpsertFlavor
grpcurl -plaintext -d '{"queue":{"name":"default","projectId":"p-demo","allowedFlavors":["cpu-small"]}}' localhost:8081 aegis.v1.AegisPlatform/UpsertQueue
```

### Helm Deployment (Hub)

```bash
helm upgrade --install aegis-services charts/aegis-services \
  -f charts/aegis-services/values/common.yaml \
  -f charts/aegis-services/values/local.yaml \
  -f charts/aegis-services/values/local-tls.yaml \
  --namespace aegis-system --create-namespace --wait
```

### Helm Deployment (Spoke)

```bash
helm upgrade --install aegis-spoke charts/aegis-spoke \
  -n aegis-spoke --create-namespace \
  -f charts/aegis-spoke/values-local.yaml
```

## Technology Stack

| Layer | Technologies |
|-------|-------------|
| **Language** | Go 1.24, Protocol Buffers |
| **API** | gRPC + gRPC-Gateway (REST), protobuf service definitions |
| **Database** | PostgreSQL (pgx), in-memory store for development |
| **Orchestration** | Kubernetes (client-go, controller-runtime), Kueue |
| **Infrastructure** | Pulumi (AWS EKS), Terraform, Helm |
| **Identity** | Keycloak (OIDC), JWT, mTLS via cert-manager + step-ca |
| **Networking** | AWS NLB, ingress-nginx |
| **Observability** | Prometheus metrics, OpenTelemetry, Zap structured logging |
| **Clients** | Backstage (React/TypeScript), VS Code Extension (TypeScript) |

## Documentation

For detailed architecture documentation, deployment guides, and compliance documentation, visit [aegis-platform.tech](https://aegis-platform.tech).
