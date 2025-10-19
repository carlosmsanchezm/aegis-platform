# Aegis Deployment & Workspace Testing Guide

This document replaces `DEPLOYMENT.md`, `DEPLOYMENT_SUMMARY.md`, `DEPLOYMENT-MODES.md`, and `QUICKREF.md`. It captures everything you need to:

- Deploy the platform locally (with or without TLS) and in the cloud (TLS enforced)
- Exercise an interactive workspace via `./scripts/test-workspace-connection.sh`
- Launch Backstage and the VS Code extension for end-to-end manual verification

If you need historical details that are not covered here, the legacy docs now simply point back to this guide.

---

## Prerequisites

- **Tooling**: Docker Desktop (with Kubernetes enabled for local runs), kubectl, Helm ≥ 3, Terraform ≥ 1.5, AWS CLI (profile `myclaude`), yarn, `grpcurl`, `jq`
- **AWS Access**: Ability to assume the account that owns `aegis-spoke-prod`
- **VS Code**: Stable build (`/Applications/Visual Studio Code.app`) with the Aegis extension installed
- **ECR Login** *(cloud only)*: `aws ecr get-login-password --region us-east-1 --profile myclaude | docker login --username AWS --password-stdin 567751785679.dkr.ecr.us-east-1.amazonaws.com`

The workspace image baked into the deployment is `567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:stable`. Override it only if you need to test a custom build.

---

## Environment Matrix

| Environment | Cluster Context | Deploy Command(s) | Backstage Command | Workspace Script Example |
|-------------|-----------------|-------------------|-------------------|--------------------------|
| **Local (no TLS)** | `docker-desktop` | `make deploy-local` | `cd aegis-platform && yarn dev` | `WORKSPACE_IMAGE=567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:stable \`<br>`GRPC_ADDR=localhost:10081 \`*¹*<br>`./scripts/test-workspace-connection.sh` |
| **Local (TLS)** | `docker-desktop` | `make deploy-local-tls` | `cd aegis-platform && yarn dev:cloud-tls` | `WORKSPACE_IMAGE=567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:stable \`<br>`GRPC_ADDR=platform-api-grpc.localtest.me:443 \`<br>`GRPC_TLS=1 \`<br>`GRPC_CA="$HOME/aegis-platform-api-ca.crt" \`<br>`GRPC_TLS_SERVER_NAME=platform-api-grpc.localtest.me \`<br>`./scripts/test-workspace-connection.sh` |
| **Cloud (TLS)** | `arn:aws:eks:us-east-1:567751785679:cluster/aegis-spoke-prod` | `cd terraform && ./generate-cloud-deployment.sh` | `cd aegis-platform && yarn dev:cloud-tls` | `WORKSPACE_IMAGE=567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:stable \`<br>`GRPC_ADDR=platform-api-grpc.aegist.dev:8081 \`<br>`GRPC_TLS=1 \`<br>`GRPC_CA="$HOME/aegis-platform-api-ca.crt" \`<br>`GRPC_TLS_SERVER_NAME=platform-api-grpc.aegist.dev \`<br>`WORKSPACE_NAMESPACE=aegis-workloads \`<br>`./scripts/test-workspace-connection.sh` |

*¹* For local (no TLS) the script defaults to `localhost:10081` only if you port-forward the platform API service:

```bash
PF_PLATFORM_HTTP_PORT=10080 PF_PLATFORM_GRPC_PORT=10081 \
  kubectl port-forward -n aegis-system svc/aegis-services-platform-api \
  $PF_PLATFORM_HTTP_PORT:8080 $PF_PLATFORM_GRPC_PORT:8081 &
```

---

## Local Development (no TLS)

1. **Switch context**
   ```bash
   kubectl config use-context docker-desktop
   ```

2. **Deploy the stack**
   ```bash
   make deploy-local
   ```

3. **Port-forward (optional but recommended for the workspace script)**
   ```bash
   PF_PLATFORM_HTTP_PORT=10080 PF_PLATFORM_GRPC_PORT=10081 \
     kubectl -n aegis-system port-forward svc/aegis-services-platform-api \
     $PF_PLATFORM_HTTP_PORT:8080 $PF_PLATFORM_GRPC_PORT:8081 &
   PF_PROXY_HTTP_PORT=10085 \
     kubectl -n aegis-system port-forward svc/aegis-services-proxy \
     $PF_PROXY_HTTP_PORT:8085 &
   ```

4. **Run the workspace smoke test**
   ```bash
   WORKSPACE_IMAGE=567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:stable \
   GRPC_ADDR=localhost:10081 \
   ./scripts/test-workspace-connection.sh
   ```
   The script auto-detects your stable VS Code commit and writes session details to `.aegis/workspace-session.json`.

5. **Manual checks**
   - Backstage: `cd aegis-platform && yarn dev`
   - VS Code: use the deep link printed by the script (or the Backstage “Open in VS Code” button)

6. **Clean up**
   ```bash
   # WARNING: operates on your current kubectl context
   make clean-local
   ```

---

## Local Development with TLS

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
   WORKSPACE_IMAGE=567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:stable \
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
   make clean-local   # still uses the current kubectl context
   ```

---

## Cloud Deployment (TLS)

1. **Provision infrastructure (if not already in place)**
   ```bash
   cd terraform
   terraform init
   terraform apply
   ```

2. **Deploy or update the applications**
   ```bash
   ./generate-cloud-deployment.sh
   ```
   The script regenerates Helm values from Terraform output, redeploys both charts, updates Route53 records, refreshes TLS secrets, and writes the latest CA certificate to `~/aegis-platform-api-ca.crt`.

3. **Switch kubeconfig context**
   ```bash
   kubectl config use-context arn:aws:eks:us-east-1:567751785679:cluster/aegis-spoke-prod
   kubectl get pods -n aegis-system
   ```

4. **Run the workspace smoke test**
   ```bash
   WORKSPACE_IMAGE=567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:stable \
   GRPC_ADDR=platform-api-grpc.aegist.dev:8081 \
   GRPC_TLS=1 \
   GRPC_CA="$HOME/aegis-platform-api-ca.crt" \
   GRPC_TLS_SERVER_NAME=platform-api-grpc.aegist.dev \
   WORKSPACE_NAMESPACE=aegis-workloads \
   ./scripts/test-workspace-connection.sh
   ```

5. **Manual checks**
   - Backstage: `cd aegis-platform && yarn dev:cloud-tls`
   - VS Code: use the `vscode://` deep link emitted by the script or Backstage
   - Confirm the AWS Load Balancers via `kubectl get svc -n aegis-system`

6. **Clean up (optional)**
   ```bash
   # Caution: ensure your context is the EKS cluster before running these
   kubectl config use-context arn:aws:eks:us-east-1:567751785679:cluster/aegis-spoke-prod
   helm uninstall aegis -n aegis-system
   # (spoke chart is named aegis-spoke; uninstall if present)
   helm uninstall aegis-spoke -n aegis-system || true
   ```
   To tear down the infrastructure entirely: `terraform destroy` (after removing Helm releases).

---

## Workspace Script Reference

- Defaults to the **stable** channel (`VSCODE_QUALITY=stable`). Override to `insider` only if you intentionally need the Insiders build—set both `VSCODE_QUALITY` and `VSCODE_COMMIT` before running the script.
- The script auto-detects the commit hash using `code --version` (stable first, then Insiders). If no editor is found or you want to lock to a specific build, export `VSCODE_COMMIT=<40-char hash>`.
- `WORKSPACE_IMAGE` defaults to `aegis-workspace:latest`; explicitly set it to the ECR tag above when targeting cloud resources to avoid accidental mismatches.
- `WORKSPACE_NAMESPACE` defaults to `aegis-workloads-local`. Always set it to `aegis-workloads` for the cloud deployment.
- TLS options (`GRPC_TLS`, `GRPC_CA`, `GRPC_TLS_SERVER_NAME`) control whether the script speaks plaintext or TLS. Leaving them unset runs plaintext.
- Each run writes a full JSON payload (including SSH config and VS Code URI) to `.aegis/workspace-session.json` for later reference.

---

## Automated & Manual Test Coverage

Below is how the existing scripts and CI jobs exercise the spoke controller and CRDs in each environment, plus the gaps that still rely on manual inspection.

### Local Test Helpers (`scripts/test-local-with-tls.sh`, `scripts/test-all-local.sh`)

- ✅ The Makefile applies the CRD before deploying the spoke chart (`make deploy-local` and `deploy-local-tls` targets call `kubectl apply -f charts/aegis-spoke/crds/aegisworkload-crd.yaml`).
- ✅ Both targets now wait for the `aegis-spoke-k8s-agent` pod to become Ready (120s timeout with a helpful error if the pod never starts).
- ❌ These make targets do not assert additional workload functionality once the pod is ready—you still need to run the workspace smoke script or Backstage manually.

### Remote Preview Workflow (`.github/workflows/preview-deployment.yml` via `terraform/generate-cloud-deployment.sh`)

- ✅ The script explicitly applies the CRD before running Helm (search for `aegisworkload-crd.yaml`).
- ✅ Helm installs both the hub and spoke charts.
- ✅ Added health check: waits for the spoke/k8s-agent deployment to report Ready in the target namespace (120s timeout) and surfaces diagnostics if it fails.
- ❌ No additional automated verification that the controller reconciles workloads—the workflow stops once the pod is healthy.

### Go E2E Suite (`agents/k8s-agent/test/e2e/e2e_test.go`)

- ✅ Installs CRDs (`By("installing CRDs") → make install`).
- ✅ Creates `AegisWorkload` resources to verify reconciliation logic end-to-end.
- ✅ Confirms the controller can reconcile lifecycle transitions (scheduling, deletion, etc.).

#### Summary

| Component / Check        | Local Make Targets | Preview Workflow | Go E2E Suite | Notes |
|--------------------------|--------------------|------------------|--------------|-------|
| CRD Application          | ✅                | ✅               | ✅           | Applied before Helm or via `make install`. |
| Spoke Deployment         | ✅                | ✅               | N/A          | Managed by Helm. |
| k8s-agent Pod Health     | ✅                | ✅               | N/A          | 120s readiness check with diagnostic logs. |
| AegisWorkload Reconcile  | ❌ Manual         | ❌ Manual        | ✅           | Use `scripts/test-workspace-connection.sh` or Backstage for manual smoke tests. |

**Benefits of the added health checks**

1. **Early failure detection** – deployments fail fast if the k8s-agent pod never starts (e.g., missing CRD or image pull issues).
2. **Better debugging** – the scripts print a pointer to the relevant `kubectl logs` when the pod fails readiness checks.
3. **Consistent behavior** – local make targets and the remote workflow now follow the same readiness gating.
4. **Prevents silent failures** – the issue we hit previously (controller crash loop) is caught immediately instead of surfacing hours later.

---

## Switching Contexts Quickly

```bash
# Cloud EKS
aws eks update-kubeconfig --region us-east-1 --name aegis-spoke-prod --profile myclaude

# Local Docker Desktop
kubectl config use-context docker-desktop
```

Use `kubectl config current-context` to confirm before running destructive commands like `make clean-local` or `helm uninstall`.

---

## Troubleshooting

| Symptom | Likely Cause | Fix |
|---------|--------------|-----|
| `Client refused: version mismatch` in VS Code | VS Code commit/channel mismatch | Ensure the workspace pod prints `Downloading VS Code server commit <stable hash> (stable, x64)` and that you launch **stable** VS Code. The script already sets `VSCODE_QUALITY=stable`; keep `WORKSPACE_IMAGE` on the stable tag. |
| Workspace never reaches `Running` | Namespace or image misconfigured | Confirm `WORKSPACE_NAMESPACE` matches the deployment (`aegis-workloads` in cloud) and that the image tag is valid in ECR. |
| TLS handshake errors in proxy logs | Incorrect CA bundle or server name | Double-check `GRPC_CA` and `GRPC_TLS_SERVER_NAME`. For local TLS use `platform-api-grpc.localtest.me`; for cloud use `platform-api-grpc.aegist.dev`. |
| `make clean-local` removes cloud resources | Running from the wrong context | Always switch back to `docker-desktop` before running local cleanup targets. |

Need more detail? The legacy docs remain in the repo for historical context, but this guide is the authoritative source going forward.
