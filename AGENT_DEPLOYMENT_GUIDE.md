# Aegis Platform -- Agent Deployment Guide

Machine-readable deployment reference for AI agents operating the Aegis V1 platform.
This is the **authoritative guide** for all deployment modes (local, local-tls, cloud).

## Quick Reference

| Command | Effect |
|---|---|
| `make deploy-local` | Deploy hub locally without TLS |
| `make deploy-local-tls` | Deploy hub locally with TLS (default dev workflow) |
| `make deploy-local-tls DEPLOY_SPOKE=true` | Deploy hub + spoke locally with TLS |
| `make deploy-cloud` | Deploy hub to cloud EKS (runs `generate-cloud-deployment.sh --non-interactive`) |
| `make deploy-cloud-full` | One-shot: ECR login + push images + terraform apply + deploy to EKS |
| `make ecr-login` | Authenticate Docker to ECR |
| `make push-cloud-images` | Build + push all images to ECR with git SHA tag |
| `make build-local-all` | Build all 3 service images for local Docker Desktop |
| `make clean-local` | Remove Helm releases |
| `make port-forward` | Port-forward platform-api, proxy, keycloak |
| `./scripts/aegis.sh deploy` | TLS deploy, hub only, dev profile (default) |
| `./scripts/aegis.sh deploy --spoke` | TLS deploy, hub + spoke, dev profile |
| `./scripts/aegis.sh deploy --hardening standard` | TLS deploy + NIST 800-171 hardening |
| `./scripts/aegis.sh status` | Pods, services, network policies |
| `./scripts/aegis.sh sync-certs` | Refresh CA bundles to ~/aegis-*.pem |
| `./scripts/aegis.sh clean` | Uninstall all Helm releases |

Full make targets reference: see `docs/make-commands.md`.

## Environment Matrix

| Environment | Cluster Context | Deploy Command | Notes |
|---|---|---|---|
| **Local (no TLS)** | `docker-desktop` | `make deploy-local` | ClusterIP services, use port-forward |
| **Local (TLS)** | `docker-desktop` | `make deploy-local-tls` | Internal PKI (step-ca + cert-manager) |
| **Cloud (TLS)** | `arn:aws:eks:...aegis-hub-prod` | `make deploy-cloud` | In-cluster Postgres, Cloudflare DNS |

## Local Development (No TLS)

1. **Switch context**
   ```bash
   kubectl config use-context docker-desktop
   ```

2. **Deploy the stack**
   ```bash
   make deploy-local
   ```

3. **Port-forward (recommended)**
   ```bash
   make port-forward
   # Or manually:
   PF_PLATFORM_HTTP_PORT=10080 PF_PLATFORM_GRPC_PORT=10081 \
     kubectl -n aegis-system port-forward svc/aegis-services-platform-api \
     $PF_PLATFORM_HTTP_PORT:8080 $PF_PLATFORM_GRPC_PORT:8081 &
   ```

4. **Run the workspace smoke test**
   ```bash
   WORKSPACE_IMAGE=195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:stable \
   GRPC_ADDR=localhost:10081 \
   ./scripts/test-workspace-connection.sh
   ```

5. **Clean up**
   ```bash
   make clean-local
   ```

## Local Development (TLS)

1. **Switch context**
   ```bash
   kubectl config use-context docker-desktop
   ```

2. **Deploy the TLS stack**
   ```bash
   make deploy-local-tls
   ```
   The make target also writes the CA bundle to `~/aegis-platform-api-ca.crt`.

3. **Trust the CA (macOS)**
   ```bash
   sudo security add-trust -d -r trustRoot \
     -k /Library/Keychains/System.keychain \
     "$HOME/aegis-platform-api-ca.crt"
   ```

4. **Run the workspace smoke test**
   ```bash
   WORKSPACE_IMAGE=195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:stable \
   GRPC_ADDR=platform-api-grpc.localtest.me:443 \
   GRPC_TLS=1 \
   GRPC_CA="$HOME/aegis-platform-api-ca.crt" \
   GRPC_TLS_SERVER_NAME=platform-api-grpc.localtest.me \
   ./scripts/test-workspace-connection.sh
   ```

5. **Manual checks**
   - Backstage: `cd aegis-platform && yarn dev:cloud-tls`
   - Verify the proxy endpoint `https://proxy.localtest.me`

6. **Clean up**
   ```bash
   make clean-local
   ```

## Cloud Hub Deployment

Deploys the hub control plane to EKS with in-cluster Postgres. For full production with RDS and HA, see `docs/PRODUCTION_DEPLOYMENT.md`.

### Prerequisites

- Terraform infrastructure applied (`cd terraform && terraform apply`)
- AWS CLI configured with `aegis-new` profile
- Docker Desktop running (for buildx cross-compilation)
- kubectl context set to the EKS cluster

### One-Shot Deploy (recommended)

Single command that does everything — ECR login, image push, terraform apply, full deploy:

```bash
export AWS_PROFILE=aegis-new
export CLOUDFLARE_API_TOKEN="<your-token>"
make deploy-cloud-full
```

Override the image tag (defaults to `git rev-parse --short HEAD`):
```bash
CLOUD_IMAGE_TAG=v1.0.0 make deploy-cloud-full
```

### Step-by-Step Deploy

1. **Login to ECR**
   ```bash
   make ecr-login
   ```

2. **Push images to ECR**
   ```bash
   make push-cloud-images
   ```

3. **Deploy to cloud EKS**
   ```bash
   make deploy-cloud
   ```

4. **Verify**
   ```bash
   kubectl get pods -n aegis-system
   ```

### What `generate-cloud-deployment.sh` Does (9 Steps)

Both `deploy-cloud` and `deploy-cloud-full` run `terraform/generate-cloud-deployment.sh --non-interactive`, which executes:

| Step | Action | Fatal? |
|------|--------|--------|
| 1 | Configure kubectl for EKS cluster | Yes |
| 2 | Create namespace `aegis-system` + Kubernetes secrets | Yes |
| 3 | Defer migrations to Step 6b (in-cluster Postgres mode) | Yes |
| 4 | Install internal PKI (cert-manager + step-ca + step-issuer) | Yes |
| 5 | Apply CRDs (AegisWorkload) | Yes |
| 6 | Deploy aegis-services Helm chart (platform-api, proxy, keycloak) | Yes |
| 6b | Run in-cluster Postgres migrations | Yes |
| 7 | Wait for LoadBalancers, update Cloudflare + Route53 DNS, update /etc/hosts | Non-fatal (/etc/hosts) |
| 7b | Write Backstage config to `aegis-ui/app-config.cloud.yaml` | Non-fatal |
| 8 | Deploy aegis-spoke (k8s-agent) Helm chart with OIDC credentials | Yes |
| 9 | Build + push + deploy aegis-ui (Backstage) | Non-fatal (subshell) |

### Key Environment Variables for Cloud Deploy

| Variable | Default | Description |
|---|---|---|
| `AWS_PROFILE` | `aegis-new` | AWS CLI profile |
| `CLOUD_IMAGE_TAG` | `git rev-parse --short HEAD` | Image tag for all ECR images |
| `CLOUDFLARE_API_TOKEN` | (empty) | Required for Cloudflare DNS updates in Step 7 |
| `SKIP_ROUTE53_UPDATE` | `0` | Set to `1` to skip DNS updates |
| `SKIP_MIGRATION_PLACEHOLDER` | `1` (set by Make) | Defers migrations to Step 6b |

### Key notes

- Uses **in-cluster Postgres** (no RDS). Migrations run in Step 6b.
- DNS zones: `aegist.dev` (Route53), `aegis-platform.tech` (Cloudflare — `terraform/cloudflare.tf`)
- Hub cluster name: `aegis-hub-prod` (configurable via `cluster_name_prefix` terraform variable)
- `CLOUD_IMAGE_TAG` uses simple expansion (`:=`) — tag is locked at Make parse time to prevent mismatch between push and deploy
- For infrastructure details, see `terraform/README.md`

## Hardening Profiles

The `hardeningProfile` Helm value controls security posture. It is a first-class chart value
set via `--set hardeningProfile=<profile>` or the `--hardening` flag in `aegis.sh`.

### `dev` (default)

- NetworkPolicies: **not rendered** (even if `networkPolicy.enabled: true` in values)
- Security contexts: **not rendered** (pods run as root on local dev)
- FIPS: **off** (no GODEBUG env var)
- Egress: **unrestricted**
- Use for: local development, fast iteration, Docker Desktop

### `standard`

- NetworkPolicies: **rendered** when `networkPolicy.enabled: true`
  - Ingress deny-by-default (from common.yaml rules)
  - Egress deny-by-default (new default-deny policy)
  - Per-service egress allowlists (platform-api, proxy, keycloak, k8s-agent)
- Security contexts: **rendered** from values (runAsNonRoot, drop ALL, readOnlyRootFilesystem)
- FIPS: **enabled** (GODEBUG=fips140=only on platform-api)
- Use for: enterprise deployments, NIST 800-171 Rev 3 compliance, production

### What Each Profile Enables/Disables

| Feature | dev | standard |
|---|---|---|
| NetworkPolicy ingress deny | Off | On |
| NetworkPolicy egress deny | Off | On |
| Per-service egress allowlists | Off | On |
| Spoke agent NetworkPolicy | Off | On |
| Pod securityContext (runAsNonRoot) | Off | On |
| Container securityContext (drop ALL) | Off | On |
| FIPS 140 mode (GODEBUG) | Off | On |
| TLS (cert-manager + step-ca) | On | On |
| mTLS client certs | Configurable | Configurable |
| Audit logging | On | On |

## Verification Steps

### After `dev` deploy

```bash
# All pods should be Running
kubectl get pods -n aegis-system

# No NetworkPolicies should exist
kubectl get networkpolicy -n aegis-system
# Expected: "No resources found"

# Platform API should be reachable
curl -sk https://platform-api.localtest.me/healthz
```

### After `standard` deploy

```bash
# All pods should be Running
kubectl get pods -n aegis-system

# NetworkPolicies should exist (ingress + egress deny, per-service egress)
kubectl get networkpolicy -n aegis-system
# Expected: default-deny-egress, platform-api-egress, proxy-egress, plus ingress policies

# Verify egress deny-by-default
kubectl get networkpolicy -n aegis-system -o name | grep deny-egress

# Platform API can reach DB, Keycloak, DNS (allowed by egress rules)
# Spoke agent heartbeats still work

# Full V1 loop still works
# Import cluster -> create workspace -> VS Code connect
```

## PKI Secrets Required by platform-api

The platform-api deployment mounts three secrets. `aegis.sh deploy` creates these
automatically, but if deploying manually via `helm upgrade` you must ensure they exist:

| Secret | Namespace | Created by | Contents |
|---|---|---|---|
| `aegis-trust-bundle` | `aegis-system` | `install-internal-pki.sh` | Root CA cert from step-ca (`ca.crt` key) |
| `step-ca-credentials` | `aegis-system` | `aegis.sh` (copies from `aegis-pki`) | step-ca provisioner password (`provisioner-password` key) |
| `spoke-proxy-ca` | `aegis-system` | `generate-spoke-proxy-ca.sh --install` | CA cert+key for signing spoke-proxy TLS certs (`ca.crt`, `ca.key` keys) |

If platform-api is stuck in `Init:0/1` with `FailedMount`, one of these secrets is missing.

The step-ca provisioner password originates as `step-certificates-provisioner-password`
in the `aegis-pki` namespace (created by the step-certificates Helm chart). The deploy
script copies it to `aegis-system` as `step-ca-credentials` with the key renamed from
`password` to `provisioner-password`.

## Order of Operations -- Full V1 Loop

### Local

1. **Deploy**: `./scripts/aegis.sh deploy` (or with `--hardening standard`)
2. **Port-forward** (if not using ingress): `./scripts/aegis.sh port-forward`
3. **Import cluster**: Via Backstage UI or platform-api gRPC
4. **Create workspace**: Via Backstage UI or platform-api gRPC
5. **Connect VS Code**: `NODE_EXTRA_CA_CERTS=~/aegis-local-trust.pem code .`
6. **Verify**: `./scripts/aegis.sh status`

### Cloud

1. **Infra**: `cd terraform && terraform apply`
2. **One-shot deploy**: `AWS_PROFILE=aegis-new CLOUDFLARE_API_TOKEN=<token> make deploy-cloud-full`
   (This handles: ECR login, image push, terraform apply, deploy steps 1-9 including DNS)
3. **Verify**: `kubectl get pods -n aegis-system`
4. **Import cluster**: Via platform-api gRPC or Backstage UI

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `HARDENING` | `dev` | Hardening profile |
| `PLATFORM_API_IMAGE` | `carlosmsanchez/aegis-platform-api:dev` | Platform API image (DockerHub = local, ECR = cloud) |
| `PROXY_IMAGE` | `carlosmsanchez/aegis-proxy:dev` | Proxy image (DockerHub = local, ECR = cloud) |
| `K8S_AGENT_IMAGE` | `carlosmsanchez/aegis-k8s-agent:dev` | K8s agent image (DockerHub = local, ECR = cloud) |
| `NAMESPACE` | `aegis-system` | Target K8s namespace |
| `DEPLOY_SPOKE` | `false` | Set to `true` to deploy aegis-spoke locally |
| `KUBE_CONTEXT` | `docker-desktop` | kubectl context |
| `RHBK_USERNAME` | (empty) | Red Hat registry username |
| `RHBK_PASSWORD` | (empty) | Red Hat registry password |
| `SKIP_MIGRATION_PLACEHOLDER` | `0` | Set to `1` to skip DB migration job (cloud with in-cluster Postgres) |
| `SKIP_ROUTE53_UPDATE` | `0` | Set to `1` to skip Route53 DNS updates |
| `CLOUDFLARE_API_TOKEN` | (empty) | Cloudflare API token for DNS management (Step 7) |
| `CLOUD_IMAGE_TAG` | `git rev-parse --short HEAD` | Image tag for all ECR images; locked at Make parse time |
| `AWS_PROFILE` | `aegis-new` | AWS CLI profile for ECR, EKS, Route53 |

## Helm Values Layering

For how Helm values files are structured and merged, see `charts/README.md`.

## Building and Pushing Images

> **Registry decision rules**: See `CLAUDE.md` > "Image Registry -- Decision Rules for AI Agents" for the
> complete decision matrix (DockerHub = local, ECR = cloud), pull policy rules, and image name table.

### Local development (Docker Desktop)

Docker Desktop shares the Docker daemon with Kubernetes, so locally built images
are available immediately. Use native architecture (omit `--platform`):

```bash
# Build for local use
docker build -f services/platform-api/Dockerfile -t carlosmsanchez/aegis-platform-api:dev .
docker build -f services/proxy/Dockerfile -t carlosmsanchez/aegis-proxy:dev .
docker build -f agents/k8s-agent/Dockerfile -t carlosmsanchez/aegis-k8s-agent:dev .
```

NOTE: The platform-api and proxy Dockerfiles use `GOARCH=amd64` by default.
On Apple Silicon, the binaries run under Rosetta emulation.

After building, restart deployments to pick up the new image:

```bash
kubectl -n aegis-system rollout restart deployment/aegis-services-platform-api
kubectl -n aegis-system rollout restart deployment/aegis-services-proxy
kubectl -n aegis-system rollout restart deployment/aegis-spoke-k8s-agent
```

### Remote / Cloud (AWS EKS on amd64)

**WARNING: Never push to DockerHub for EKS deployments.** EKS nodes must pull from ECR.

```bash
# Push all images to ECR (canonical command)
make push-cloud-images

# Or push individually:
TAG=$(git rev-parse --short HEAD)
docker buildx build --platform linux/amd64 -f services/platform-api/Dockerfile -t 195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/platform-api:$TAG --push .
docker buildx build --platform linux/amd64 -f services/proxy/Dockerfile -t 195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/proxy:$TAG --push .
docker buildx build --platform linux/amd64 -f agents/k8s-agent/Dockerfile -t 195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/k8s-agent:$TAG --push .
```

### ECR Login

```bash
# Login to ECR (use aegis-new profile for the 195714074609 account)
AWS_PROFILE=aegis-new aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 195714074609.dkr.ecr.us-east-1.amazonaws.com
```

### Upgrading a remote spoke

```bash
# Connect to the EKS cluster
aws eks update-kubeconfig --name <cluster-name> --region us-east-1

# Upgrade spoke chart (values-cloud-remote.yaml uses ECR repos)
helm upgrade aegis-spoke charts/aegis-spoke \
  -n aegis-system \
  -f charts/aegis-spoke/values.yaml \
  -f charts/aegis-spoke/values-cloud-remote.yaml \
  --set k8sAgent.env.AEGIS_CLUSTER_ID=<cluster-id> \
  --set k8sAgent.env.AEGIS_REGION=us-east-1 \
  --reset-values --wait --timeout 5m

# Or use the convenience script:
./scripts/update-spoke-aws.sh <cluster-name> [region]
```

## Workspace Testing Reference

- Defaults to the **stable** channel (`VSCODE_QUALITY=stable`). Override to `insider` only if needed.
- The script auto-detects the commit hash using `code --version`. If no editor is found, export `VSCODE_COMMIT=<40-char hash>`.
- `WORKSPACE_IMAGE` defaults to `aegis-workspace:latest`; set it to the ECR tag for cloud deployments.
- `WORKSPACE_NAMESPACE` defaults to `aegis-workloads-local`. Set to `aegis-workloads` for cloud.
- TLS options (`GRPC_TLS`, `GRPC_CA`, `GRPC_TLS_SERVER_NAME`) control plaintext vs TLS. Unset = plaintext.
- Each run writes a JSON payload to `.aegis/workspace-session.json`.

### Workspace smoke test examples

**Local (no TLS):**
```bash
WORKSPACE_IMAGE=195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:stable \
GRPC_ADDR=localhost:10081 \
./scripts/test-workspace-connection.sh
```

**Local (TLS):**
```bash
WORKSPACE_IMAGE=195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:stable \
GRPC_ADDR=platform-api-grpc.localtest.me:443 \
GRPC_TLS=1 \
GRPC_CA="$HOME/aegis-platform-api-ca.crt" \
GRPC_TLS_SERVER_NAME=platform-api-grpc.localtest.me \
./scripts/test-workspace-connection.sh
```

**Cloud:**
```bash
WORKSPACE_IMAGE=195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:stable \
GRPC_ADDR=platform-api-grpc.aegist.dev:8081 \
GRPC_TLS=1 \
GRPC_CA="$HOME/aegis-platform-api-ca.crt" \
GRPC_TLS_SERVER_NAME=platform-api-grpc.aegist.dev \
WORKSPACE_NAMESPACE=aegis-workloads \
./scripts/test-workspace-connection.sh
```

## Common Failure Modes

| Symptom | Cause | Fix |
|---|---|---|
| Pods CrashLoopBackOff with permission denied | Security contexts active but image needs root | Use `dev` profile or fix image |
| DNS resolution failing in `standard` | Egress deny blocks kube-dns | Verify egress policies allow port 53 |
| Agent can't heartbeat in `standard` | Egress deny blocks gRPC | Verify spoke NetworkPolicy allows 443/8081/8443 |
| cert-manager webhook timeout | Webhook not ready | Script retries 3x; if still failing, check cert-manager pods |
| Helm install fails | Stale webhook config | Run `./scripts/aegis.sh clean` then redeploy |
| Helm upgrade fails: `post-upgrade hooks failed: BackoffLimitExceeded` | Stale/failed migration job | `kubectl -n aegis-system delete job aegis-services-platform-api-migrate` then `helm rollback aegis-services <last-good-revision> -n aegis-system` and retry |
| Helm upgrade fails: `another operation is in progress` | Previous failed upgrade left pending state | `helm rollback aegis-services <last-good-revision> -n aegis-system` then retry |
| Agent logs: `error reading server preface: EOF` | TLS mismatch -- agent connecting without TLS to TLS-enabled platform-api | Ensure helm upgrade includes all four values files: `values.yaml`, `values-local.yaml`, `values-local-tls.yaml`, `values-local-pki.yaml` |
| Agent logs: `Unauthenticated: missing bearer token` | OIDC credentials not configured for spoke agent | Local dev requires Keycloak `spoke-agent` client; set `AEGIS_CP_OIDC_CLIENT_ID`, `AEGIS_CP_OIDC_CLIENT_SECRET`, `AEGIS_CP_OIDC_TOKEN_URL` via `--set` or values file |
| `ImagePullBackOff` after changing image tag | New tag doesn't exist in registry and `imagePullPolicy: IfNotPresent` | Either push the tag to the registry, use `imagePullPolicy: Always`, or use the existing `:dev` tag |
| platform-api stuck in `Init:0/1` with `FailedMount` | Missing `step-ca-credentials` or `spoke-proxy-ca` secrets | Run `./scripts/generate-spoke-proxy-ca.sh --install` and copy step-ca provisioner password: `kubectl get secret step-certificates-provisioner-password -n aegis-pki -o jsonpath='{.data.password}' \| base64 -d \| xargs -I{} kubectl create secret generic step-ca-credentials -n aegis-system --from-literal=provisioner-password={}` |
| Proxy `ImagePullBackOff` on Apple Silicon | Proxy image built for amd64 only | Rebuild with multi-arch: `docker buildx build --platform linux/arm64,linux/amd64 -f services/proxy/Dockerfile -t carlosmsanchez/aegis-proxy:dev --push .` |
| `Client refused: version mismatch` in VS Code | VS Code commit/channel mismatch | Ensure the workspace pod uses the stable VS Code server hash and you launch **stable** VS Code |
| Workspace never reaches `Running` | Namespace or image misconfigured | Confirm `WORKSPACE_NAMESPACE` matches the deployment and that the image tag is valid |
| TLS handshake errors in proxy logs | Incorrect CA bundle or server name | Check `GRPC_CA` and `GRPC_TLS_SERVER_NAME`. Local TLS: `platform-api-grpc.localtest.me`; cloud: `platform-api-grpc.aegist.dev` |
| `make clean-local` removes cloud resources | Running from the wrong context | Always switch back to `docker-desktop` before running local cleanup targets |

## Context Switching

```bash
# Cloud EKS
aws eks update-kubeconfig --region us-east-1 --name aegis-hub-prod --profile aegis-new

# Local Docker Desktop
kubectl config use-context docker-desktop

# Confirm current context
kubectl config current-context
```

Use `kubectl config current-context` to confirm before running destructive commands like `make clean-local` or `helm uninstall`.
