# Iron Bank Image Testing Log

**Branch:** `feature/ironbank-compliance`
**Date started:** 2026-03-15
**EKS Cluster:** `aegis-hub-prod` (us-east-1)

## Test Environment
- Local: Docker Desktop on macOS arm64
- Cloud: EKS `aegis-hub-prod` (amd64 nodes, v1.33.8)
- ECR: `195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/*`

## Issues Found During Testing

### Issue: vscode-reh-init ECR repo doesn't exist
- **Step:** Step 3a (Push to ECR)
- **Environment:** cloud
- **Error:** `RepositoryNotFoundException: The repository with name 'aegis/vscode-reh-init' does not exist`
- **Fix:** `aws ecr create-repository --repository-name aegis/vscode-reh-init --region us-east-1 --profile aegis-new`
- **Recurring:** One-time setup
- **Automate:** Yes — script should create repo if missing

### Issue: Spoke agent requires OIDC client secret for cloud deployment
- **Step:** Step 3b (Deploy to EKS)
- **Environment:** cloud
- **Error:** `AEGIS_CP_OIDC_CLIENT_SECRET is required when AEGIS_CP_OIDC_TOKEN_URL is set`
- **Fix:** Set `--set k8sAgent.env.AEGIS_CP_OIDC_CLIENT_SECRET="rEC99sBBWQAbRgg0xRQFBsMC8rt6pZOB"` (from Keycloak realm JSON)
- **Recurring:** Every cloud spoke deployment
- **Automate:** Yes — script should extract from realm JSON or accept as parameter

### Issue: Spoke proxy misconfiguration guard blocks deployment without proxy
- **Step:** Step 3b (Deploy to EKS)
- **Environment:** cloud
- **Error:** Helm template error requiring proxy.enabled=true for cloud deployments
- **Fix:** Set `--set proxy.enabled=true` even for co-located spoke on hub cluster
- **Recurring:** Every spoke deployment
- **Automate:** Yes — script should always set proxy.enabled=true for cloud

## Cloud Test Results (2026-03-15)

| Component | Image | Status | Notes |
|-----------|-------|--------|-------|
| platform-api | ironbank-test (UBI9-minimal) | RUNNING | gRPC + HTTP serving, OIDC auth enforced, PostgreSQL connected |
| proxy | ironbank-test (distroless) | RUNNING | TLS listening on :8080 |
| k8s-agent | ironbank-test (distroless) | RUNNING | Cluster registered, heartbeating, workload GC enabled |
| spoke-proxy | ironbank-test (distroless) | RUNNING | Co-located on hub cluster |
| workspace | ironbank-test (CUDA UBI9) | RUNNING | VS Code REH starts on port 11111, extension host listening |
| vscode-reh-init | ironbank-test (UBI9-minimal) | COMPLETED (Exit 0) | Copies REH binary to shared volume successfully |

### Issue: Init container `cp -a` fails with permission error
- **Step:** Workspace pod creation
- **Environment:** cloud (EKS)
- **Error:** `cp: preserving times for '/shared/reh/.': Operation not permitted`
- **Fix:** Changed `cp -a` to `cp -r` in workspace.go and vscode-reh-init Dockerfile. Non-root user (UID 65534) cannot preserve timestamps on emptyDir volume.
- **Recurring:** Would happen every deployment
- **Automate:** Fixed in code (committed)

### Issue: Cluster needs kubeconfig for workload placement
- **Step:** Workspace submission
- **Environment:** cloud (co-located spoke on hub)
- **Error:** `no eligible cluster for flavor=cpu-small in policy regions=[us-east-1]`
- **Fix:** Added in-cluster kubeconfig to `aegis-kubeconfigs` secret and set `kubeconfig_secret_ref` in DB
- **Recurring:** Per-cluster setup
- **Automate:** Yes — deploy script should handle kubeconfig injection for co-located spokes

### Issue: Project needs policy regions for placement
- **Step:** Workspace submission
- **Environment:** both
- **Error:** `no eligible cluster for flavor=cpu-small in policy regions=[]`
- **Fix:** Create project with `policy.regions: ["us-east-1"]`
- **Recurring:** Per-project setup
- **Automate:** Yes — test script should create project with regions
