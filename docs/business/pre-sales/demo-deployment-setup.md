# Aegis Demo Deployment Setup

**Purpose:** Prepare and rehearse the cloud demo. Uses `make deploy-cloud` for everything — the script is idempotent, so running it again after a `helm uninstall` re-deploys only what's missing.

**Last Updated:** 2026-03-30

---

## Overview

```
Phase 1: make push-cloud-images      (build + push all images to ECR)
Phase 2: make deploy-cloud            (full deploy: terraform + PKI + helm + DNS)
Phase 3: helm uninstall               (tear down helm release only)
Phase 4: make deploy-cloud            (on-camera — reruns, skips what exists, deploys helm + DNS)
```

---

## Phase 1: Build and Push Images

```bash
export AWS_PROFILE=aegis-new
make push-cloud-images TARGET_ARCH=amd64
```

Also ensure the VS Code REH init image exists:
```bash
aws ecr describe-images --repository-name aegis/vscode-reh-init --region us-east-1 \
  --query 'imageDetails[0].imageTags' --output text
```

If missing or outdated, rebuild it:
```bash
# Get your VS Code commit hash
COMMIT=$("/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code" --version | sed -n '2p')
VERSION=$("/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code" --version | head -1)

# Download matching REH binary
curl -sL "https://update.code.visualstudio.com/commit:${COMMIT}/server-linux-x64/stable" \
  -o workspace-images/vscode-reh-init/vscode-server-linux-x64.tar.gz

# Build and push
docker build --platform linux/amd64 \
  -t 195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/vscode-reh-init:${VERSION} \
  -t 195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/vscode-reh-init:latest \
  workspace-images/vscode-reh-init/

aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin \
  195714074609.dkr.ecr.us-east-1.amazonaws.com

docker push 195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/vscode-reh-init:${VERSION}
docker push 195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/vscode-reh-init:latest
```

---

## Phase 2: Full Deploy (off-camera)

```bash
export AWS_PROFILE=aegis-new
export CLOUDFLARE_API_TOKEN="<your-token>"

make deploy-cloud
```

This runs `terraform apply` + `generate-cloud-deployment.sh` which handles:
1. Terraform (EKS, IRSA, DNS zones, secrets) — idempotent
2. Namespace + Kubernetes secrets
3. Database migrations
4. Internal PKI (step-ca NLB + cert-manager + step-issuer)
5. CRDs
6. Helm install (platform-api, proxy, keycloak, backstage, ingress-nginx)
7. Wait for NLBs
8. Update Cloudflare DNS
9. Verify endpoints

### Verify everything works

```bash
# UI loads
open https://ui.aegis-platform.tech

# Discovery endpoint
curl -sk http://platform-api.aegis-platform.tech:8080/api/v1/discovery | python3 -m json.tool | head -5

# Step-CA healthy
curl -sk "https://$(kubectl get svc step-certificates-nlb -n aegis-pki \
  -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')/health"

# Platform-api has PKI env vars
kubectl exec deploy/aegis-platform-api -n aegis-system -- env | grep AEGIS_STEP_CA_URL
```

---

## Phase 3: Tear Down Helm Release (off-camera)

```bash
helm uninstall aegis -n aegis-system --wait
```

### What survives (everything the script needs):

| Kept | Why |
|------|-----|
| EKS cluster, VPC, IRSA | Terraform-managed |
| aegis-system namespace | Created by kubectl |
| All Kubernetes secrets | Created by kubectl, not helm |
| step-ca + cert-manager + step-issuer | Separate helm releases in aegis-pki/cert-manager namespaces |
| StepClusterIssuer (aegis-internal) | Created by install-internal-pki.sh |
| Trust bundle + PKI init secrets | Helm hooks survive uninstall |
| CRDs | Applied via kubectl |
| Step-CA NLB | Created via kubectl |
| ECR images | In registry |

### What gets deleted (helm recreates on next install):

| Deleted | Recreated by |
|---------|-------------|
| Platform-api, proxy, backstage deployments | Helm install |
| Keycloak StatefulSet | Helm install |
| Ingress-nginx controller | Helm install |
| NLBs (platform-api, proxy, ingress) | Helm install (new hostnames) |
| TLS certificates | cert-manager re-issues from step-ca |

---

## Phase 4: Demo (on-camera)

```bash
make deploy-cloud
```

Since terraform and PKI already exist, the script:
- `terraform apply` — no changes (2-3 seconds)
- Secrets — already exist, skipped
- PKI — already installed, skipped
- CRDs — already applied
- **Helm install** — this is the main event (~3-5 min)
- Wait for NLBs (~2 min)
- Update Cloudflare DNS to new NLB hostnames
- Verify endpoints

Total on-camera time: ~5-7 minutes.

### What to narrate (from demo-narration-script.md, Act 2)

> "We're starting from scratch. I have an EKS cluster running in GovCloud. No platform, no auth, no services — just Kubernetes."
>
> "I'm running a single deploy command. Watch what comes up."
>
> As services appear: Keycloak, Internal PKI, Platform API, Proxy, Backstage UI.
>
> "That's it. One command. The entire platform — auth, encryption, scheduling, UI, proxy — is now running."

---

## Post-Demo: VS Code Extension Setup

For the VS Code connection demo (Act 5), update the local CA:

```bash
kubectl exec step-certificates-0 -n aegis-pki -- \
  cat /home/step/certs/root_ca.crt > ~/aegis-cloud-ca.pem
```

Then launch VS Code with the extension:
```bash
AEGIS_TEST_USERNAME="cms553@cornell.edu" \
AEGIS_TEST_PASSWORD="<password>" \
"/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code" \
  --extensionDevelopmentPath="$HOME/code/sovran/aegis-vscode-remote/extension" \
  --enable-proposed-api aegis.aegis-remote \
  "$HOME/code/sovran"
```

---

## Troubleshooting

| Symptom | Cause | Fix |
|---------|-------|-----|
| `make deploy-cloud` fails at terraform | State lock or drift | `cd terraform && terraform plan` to diagnose |
| Helm install times out | Keycloak waiting for TLS cert | Check `kubectl get certificate -n aegis-system` — StepClusterIssuer must be Ready |
| DNS not resolving after deploy | Cloudflare propagation delay | Wait 30-60s, or check `SKIP_DNS_UPDATE` wasn't set |
| VS Code "unable to get local issuer" | Local CA file stale | Re-download: `kubectl exec step-certificates-0 -n aegis-pki -- cat /home/step/certs/root_ca.crt > ~/aegis-cloud-ca.pem` |
| VS Code "version mismatch" | REH init image tag doesn't match local VS Code | Rebuild vscode-reh-init (see Phase 1) |
| Spoke provisioning fails | Missing AEGIS_STEP_CA_URL | Script verifies this — check output for verification step |
| Keycloak PVC stuck | Previous PVC with different StorageClass | `kubectl delete pvc -n aegis-system -l app.kubernetes.io/name=keycloak` then re-run |
