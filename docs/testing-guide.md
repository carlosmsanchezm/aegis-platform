# Aegis Testing Guide

This guide covers all testing approaches for the Aegis platform, including local tests and remote preview deployment tests.

## Table of Contents

- [Overview](#overview)
- [Local Testing](#local-testing)
  - [Quick Start](#quick-start)
  - [Unit and Integration Tests](#unit-and-integration-tests)
  - [E2E Tests](#e2e-tests)
- [Remote Preview Testing](#remote-preview-testing)
  - [Workflow Overview](#workflow-overview)
  - [Running Preview Deployment](#running-preview-deployment)
  - [Preview E2E Tests](#preview-e2e-tests)
- [Test Coverage](#test-coverage)

---

## Overview

Aegis uses a comprehensive testing strategy that includes:

- **Unit Tests**: Go and frontend component tests
- **Integration Tests**: Cross-module integration testing
- **E2E Tests**: End-to-end testing against deployed environments
- **Remote Preview Tests**: Full deployment testing in AWS via GitHub Actions

---

## Local Testing

### Quick Start

Run all local tests (unit + integration):

```bash
./scripts/test-all-local.sh
```

This mirrors the GitHub Actions workflow and includes:
- Go unit and integration tests
- Frontend unit tests
- Linting (Go and frontend)

### Prerequisites

Install required dependencies:

- **Go**: 1.24+ ([install](https://golang.org/doc/install))
- **Node.js**: 18+ ([install](https://nodejs.org/))
- **Yarn**: `npm install -g yarn`
- **kubectl**: Required for E2E tests ([install](https://kubernetes.io/docs/tasks/tools/))
- **grpcurl**: Required for platform E2E ([install](https://github.com/fullstorydev/grpcurl))
- **jq**: JSON processor ([install](https://stedolan.github.io/jq/))

### Unit and Integration Tests

#### Go Tests

Run all Go unit and integration tests:

```bash
# Sync Go workspace
go work sync

# Test k8s-agent (uses envtest)
cd agents/k8s-agent && make test

# Test platform-api
cd services/platform-api && go test ./...

# Test proxy
cd services/proxy && go test ./...

# Test workspace package
cd pkg/workspace && go test ./...
```

#### Frontend Tests

Run frontend unit tests:

```bash
cd aegis-platform

# Install dependencies (if needed)
yarn install --immutable

# Run tests
yarn test --watchAll=false
```

#### Linting

**Go Linting:**
```bash
cd agents/k8s-agent
go fmt ./...
go vet ./...
make lint  # golangci-lint
```

**Frontend Linting:**
```bash
cd aegis-platform
yarn prettier:check
yarn lint:all
```

### E2E Tests

#### Platform API E2E Tests

Test the platform API gRPC endpoints against a deployed environment.

**Prerequisites:**
- Aegis services deployed (via `make deploy-local` or remote)
- kubectl configured to access the cluster

**Run locally with port-forwarding:**

```bash
# Set up port-forwarding to local cluster
make port-forward

# Run E2E tests (defaults to localhost:10081)
export AEGIS_GRPC_ADDR="127.0.0.1:10081"
export GRPC_TLS=0  # Set to 1 for TLS
./scripts/e2e-platform-api.sh
```

**Environment Variables:**
- `AEGIS_GRPC_ADDR`: Platform API gRPC address (default: `127.0.0.1:10081`)
- `GRPC_TLS`: Enable TLS (`0` or `1`)
- `GRPC_CA`: Path to CA certificate (when TLS enabled)
- `GRPC_TLS_SERVER_NAME`: TLS server name override
- `AEGIS_PLATFORM_NAMESPACE`: Platform namespace (default: `aegis-system`)
- `AEGIS_WORKLOAD_NAMESPACE`: Workload namespace (default: `aegis-workloads-local`)

**What it tests:**
1. UpsertFlavor - Create GPU flavor
2. CreateProject - Create test project
3. UpsertQueue - Create workload queue
4. RegisterCluster - Register K8s cluster
5. Heartbeat - Cluster health check
6. SubmitWorkload - Submit a test workload
7. ListWorkloads - Verify workload listing
8. StartWorkload - Start the workload
9. GetWorkload - Check workload status
10. AckWorkload - Acknowledge completion
11. Verify terminal state (SUCCEEDED)

#### K8s Agent E2E Tests

Test the Kubernetes operator against a cluster.

**Using existing cluster:**
```bash
cd agents/k8s-agent
USE_EXISTING_CLUSTER=true make test-e2e
```

**Using new Kind cluster:**
```bash
cd agents/k8s-agent
USE_EXISTING_CLUSTER=false KIND_CLUSTER=aegis-e2e make test-e2e
```

### Test Script Options

The `test-all-local.sh` script supports various options:

```bash
# Skip specific test suites
./scripts/test-all-local.sh --skip-lint
./scripts/test-all-local.sh --skip-go-tests
./scripts/test-all-local.sh --skip-frontend-tests

# Run with E2E tests
./scripts/test-all-local.sh --with-platform-e2e
./scripts/test-all-local.sh --with-operator-e2e

# Run operator E2E in new Kind cluster
./scripts/test-all-local.sh --with-operator-e2e --use-kind

# Environment variable control
RUN_E2E_PLATFORM=1 ./scripts/test-all-local.sh
RUN_E2E_OPERATOR=1 ./scripts/test-all-local.sh
```

---

## Remote Preview Testing

The preview deployment workflow runs in GitHub Actions and provides full integration testing in AWS.

### Workflow Overview

**Workflow File:** `.github/workflows/preview-deployment.yml`

**Triggers:**
- `workflow_dispatch`: Manual trigger with options
- `pull_request`: Automatic on PRs (coming soon)

**Run Modes:**
- `all`: Full pipeline (lint → test → build → deploy → test → teardown)
- `deploy`: Deploy only (skip lint/test/build)
- `deploy-no-teardown`: Deploy only, leave environment running

### Workflow Jobs

#### 1. Lint & Unit Test

Runs on `workflow_dispatch` with `run_subset=all`:

```yaml
# Go tests
go work sync
cd agents/k8s-agent && make test
cd services/platform-api && go test ./...
cd services/proxy && go test ./...
cd pkg/workspace && go test ./...

# Frontend tests
cd aegis-platform
yarn install --immutable
yarn test --watchAll=false
```

#### 2. Build & Push Images

Builds Docker images and pushes to ECR:

- Platform API: `aegis/platform-api:${GITHUB_SHA}`
- Proxy: `aegis/proxy:${GITHUB_SHA}`
- K8s Agent: `aegis/k8s-agent:${GITHUB_SHA}`

**Optimization:** Checks if images exist before building.

#### 3. Deploy Preview & Run E2E Tests

**Infrastructure Setup:**
1. Terraform provisions AWS resources (EKS, RDS, networking)
2. Helm deploys control plane services
3. Route53 DNS records created
4. LoadBalancers provisioned and mapped to DNS

**E2E Test Execution:**

**Platform API E2E** (via DNS with TLS):
```bash
# Environment
export AEGIS_GRPC_ADDR="preview-${ID}.platform-api.aegis.example.com:8081"
export GRPC_TLS=1
export GRPC_CA="${HOME}/aegis-platform-api-ca.crt"
export GRPC_TLS_SERVER_NAME="preview-${ID}.platform-api.aegis.example.com"

# Run tests
./scripts/e2e-platform-api.sh
```

**K8s Agent E2E** (on EKS cluster):
```bash
cd agents/k8s-agent
export USE_EXISTING_CLUSTER=true
export KIND_CLUSTER=aegis-e2e
export K8S_AGENT_E2E_IMAGE="${ECR_REGISTRY}/aegis/k8s-agent:${GITHUB_SHA}"
export K8S_AGENT_E2E_SKIP_BUILD=true
make test-e2e
```

#### 4. Teardown Preview

Automatic cleanup (unless `deploy-no-teardown`):
1. Uninstall Helm releases
2. Delete Kubernetes namespace
3. Run `terraform destroy`
4. Delete Terraform workspace

### Running Preview Deployment

**From GitHub UI:**

1. Go to Actions → "Build, Test, and Deploy Preview"
2. Click "Run workflow"
3. Select options:
   - **Run subset**: `all` | `deploy` | `deploy-no-teardown`
   - **Preview number**: Custom identifier (optional, defaults to run ID)

**Key Environment Variables:**

```bash
# Set automatically by workflow
PREVIEW_ID="${preview_number or run_id}"
PREVIEW_NAMESPACE="preview-${PREVIEW_ID}"
PREVIEW_RELEASE="preview-${PREVIEW_ID}"

# AWS Configuration
AWS_REGION="us-east-1"
```

### Preview E2E Tests

The preview deployment runs comprehensive E2E tests:

**DNS Configuration:**
- Platform API: `preview-${ID}.platform-api.aegis.example.com:8081`
- Proxy: `preview-${ID}.proxy.aegis.example.com:8080`

**TLS Setup:**
- Uses AWS Certificate Manager for TLS
- CA certificate downloaded to runner
- Configured via `GRPC_CA` and `GRPC_TLS_SERVER_NAME`

**Host Resolution:**
```bash
# LoadBalancer IPs added to /etc/hosts for immediate resolution
PLATFORM_IP=$(dig +short "${PLATFORM_DNS}" | head -1)
PROXY_IP=$(dig +short "${PROXY_DNS}" | head -1)

echo "${PLATFORM_IP} ${PLATFORM_DNS}" >> /etc/hosts
echo "${PROXY_IP} ${PROXY_DNS}" >> /etc/hosts

# Flush DNS cache
sudo resolvectl flush-caches
```

**Test Flow:**
1. Wait for LoadBalancers to be provisioned (max 5 minutes)
2. Update Route53 with LoadBalancer hostnames
3. Resolve and add to /etc/hosts for immediate access
4. Install test tooling (grpcurl, jq)
5. Run platform API E2E tests over TLS
6. Run k8s-agent E2E tests on EKS cluster

### Debugging Preview Deployments

**Check deployment status:**
```bash
kubectl get all -n preview-${ID}
kubectl logs -n preview-${ID} deployment/preview-${ID}-platform-api
kubectl logs -n preview-${ID} deployment/preview-${ID}-proxy
```

**Verify DNS:**
```bash
dig preview-${ID}.platform-api.aegis.example.com
dig preview-${ID}.proxy.aegis.example.com
```

**Check Terraform state:**
```bash
cd terraform
terraform workspace select preview-${ID}
terraform output
```

**Manual teardown:**
```bash
cd terraform
terraform workspace select preview-${ID}
helm uninstall preview-${ID} -n preview-${ID}
helm uninstall preview-${ID}-spoke -n preview-${ID}
kubectl delete namespace preview-${ID}
terraform destroy -auto-approve
terraform workspace select default
terraform workspace delete preview-${ID}
```

---

## Test Coverage

### Go Modules

| Module | Unit Tests | Integration Tests | E2E Tests |
|--------|-----------|-------------------|-----------|
| `agents/k8s-agent` | ✅ | ✅ (envtest) | ✅ (Kind/EKS) |
| `services/platform-api` | ✅ | ✅ | ✅ (gRPC) |
| `services/proxy` | ✅ | ✅ | ✅ (via platform) |
| `pkg/workspace` | ✅ | ✅ | N/A |

### Frontend

| Area | Tests |
|------|-------|
| React Components | ✅ Unit tests |
| Integration | ✅ Via `yarn test` |
| Linting | ✅ ESLint + Prettier |

### E2E Test Coverage

**Platform API E2E** (`scripts/e2e-platform-api.sh`):
- ✅ Project lifecycle (create, manage)
- ✅ Queue management (upsert, configure)
- ✅ Cluster registration and heartbeat
- ✅ Workload submission and lifecycle
- ✅ Flavor management

**K8s Agent E2E** (`agents/k8s-agent/test/e2e/`):
- ✅ Operator deployment and reconciliation
- ✅ CRD validation and processing
- ✅ Webhook configuration
- ✅ Resource management

**Preview Deployment E2E**:
- ✅ Full infrastructure provisioning (Terraform)
- ✅ Helm deployment (control plane + spoke)
- ✅ TLS/mTLS configuration
- ✅ DNS and LoadBalancer setup
- ✅ Database migrations
- ✅ End-to-end workload flow

---

## CI/CD Integration

### Local Development Workflow

```bash
# 1. Make changes
# 2. Run local tests
./scripts/test-all-local.sh

# 3. Run E2E tests locally
make deploy-local
make port-forward
./scripts/test-all-local.sh --with-platform-e2e

# 4. Commit and push
git add .
git commit -m "feature: description"
git push
```

### Remote Testing Workflow

```bash
# 1. Push changes to branch
git push origin feature-branch

# 2. Trigger preview deployment
# Via GitHub Actions UI or:
gh workflow run preview-deployment.yml \
  -f run_subset=all \
  -f preview_number=my-feature

# 3. Monitor deployment
gh run list --workflow=preview-deployment.yml
gh run view <run-id> --log

# 4. Keep environment for debugging (optional)
gh workflow run preview-deployment.yml \
  -f run_subset=deploy-no-teardown
```

---

## Best Practices

1. **Run tests locally first** - Catch issues before pushing
2. **Use preview deployments for integration** - Test full AWS stack
3. **Monitor E2E test output** - Platform E2E creates temporary resources
4. **Clean up preview environments** - Don't use `deploy-no-teardown` unless debugging
5. **Check test coverage** - Ensure new features include tests
6. **Keep tests fast** - Unit tests should run in seconds, E2E in minutes

---

## Troubleshooting

### Local Tests Failing

**Go tests:**
```bash
# Clean build cache
go clean -cache -testcache
go work sync

# Re-run specific module
cd services/platform-api
go test -v ./...
```

**Frontend tests:**
```bash
# Clear node_modules
rm -rf node_modules yarn.lock
yarn install
yarn test
```

### E2E Tests Failing

**Platform API E2E:**
- Verify services are deployed: `kubectl get pods -n aegis-system`
- Check port-forwarding: `lsof -i :10081`
- Verify gRPC connectivity: `grpcurl -plaintext localhost:10081 list`

**K8s Agent E2E:**
- Check cluster access: `kubectl cluster-info`
- Verify CRDs installed: `kubectl get crds | grep aegis`
- Check operator logs: `kubectl logs -n aegis-system deployment/aegis-k8s-agent`

### Preview Deployment Issues

**Terraform errors:**
- Check AWS credentials: `aws sts get-caller-identity`
- Verify backend: `terraform init -backend-config=backend.hcl`
- Check workspace: `terraform workspace list`

**LoadBalancer timeout:**
- Check AWS console for LB status
- Verify security groups allow traffic
- Check VPC and subnet configuration

**DNS resolution:**
- Verify Route53 records exist
- Check TTL (may need to wait)
- Try `dig +short <dns-name>` to verify

**TLS/mTLS issues:**
- Verify CA certificate downloaded
- Check certificate paths in environment vars
- Validate certificate: `openssl x509 -in ${GRPC_CA} -text -noout`
