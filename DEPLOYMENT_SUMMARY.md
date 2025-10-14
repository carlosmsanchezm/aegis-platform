# Aegis Deployment Guide

This doc captures how to run Aegis in three modes:

1. **Local development** on Docker Desktop (no TLS)
2. **Cloud / EKS without TLS** (plain LoadBalancers)
3. **Cloud / EKS with TLS** (Ingress + cert-manager)

Everything is value-file driven—Helm commands simply choose the right overlays.

---

## 1. Local Development (Docker Desktop)

### Prerequisites
- Docker Desktop with Kubernetes enabled
- `kubectl`, `helm`, and `yarn`
- Repo checked out locally

### Switch kubeconfig
```bash
kubectl config use-context docker-desktop
kubectl get nodes   # should show docker-desktop
```

### Deploy with Make
```bash
make deploy-local
```

### Access services via localtest.me (no port-forward needed)

Both make targets configure ingress-nginx with hostnames that resolve to `127.0.0.1`, so you can reach the services directly:

| Service        | Hostname                          | Protocol |
|----------------|-----------------------------------|----------|
| Platform API   | `platform-api-grpc.localtest.me`  | gRPC     |
| Proxy          | `proxy.localtest.me`              | HTTPS/WS |

If you still prefer a tunnel, you can port-forward manually:

```bash
# Platform API on http://localhost:10080 (HTTP) and :10081 (gRPC)
PF_PLATFORM_HTTP_PORT=10080 PF_PLATFORM_GRPC_PORT=10081 \
kubectl -n aegis-system port-forward svc/aegis-services-platform-api $PF_PLATFORM_HTTP_PORT:8080 $PF_PLATFORM_GRPC_PORT:8081

# Proxy tunnel on http://localhost:10085/proxy/
PF_PROXY_HTTP_PORT=10085 \
kubectl -n aegis-system port-forward svc/aegis-services-proxy $PF_PROXY_HTTP_PORT:8085

# Stop existing forwards if needed
pkill -f "kubectl port-forward"  # optional cleanup
```

### Start Backstage against local services
```bash
cd aegis-platform
yarn dev          # copies app-config.local-dev.yaml
```

### Enable TLS locally (auto-refreshes CA)
```bash
make deploy-local-tls
```
This command also saves the latest ingress certificate to `~/aegis-platform-api-ca.crt`. After it finishes:

```bash
# trust the new CA
sudo security add-trust -d -r trustRoot \
  -k /Library/Keychains/System.keychain \
  ~/aegis-platform-api-ca.crt

# launch VS Code with the CA
NODE_EXTRA_CA_CERTS=~/aegis-platform-api-ca.crt \
  /Applications/Visual\ Studio\ Code.app/Contents/MacOS/Electron \
  --enable-proposed-api aegis.aegis-remote \
  ~/code/sovran
```
> Tip: wrap the launch command in a script so the env var is always present.

---

## 2. Cloud / EKS (plain HTTP)

### Provision infra
```bash
cd terraform
# One-time bootstrap (creates the S3 bucket / DynamoDB table if needed)
./bootstrap-remote-state.sh \\
  --bucket <your-state-bucket> \\
  --region us-east-1 \\
  --dynamodb-table terraform-lock-table

# Configure remote state (use backend.hcl with your bucket/key)
terraform init -backend-config=backend.hcl
terraform apply
```

### Generate value overlays
```bash
./generate-helm-values.sh
```
This creates/updates:
- `charts/aegis-services/values-cloud-generated.yaml`
- `charts/aegis-spoke/values-cloud-generated.yaml`
- Prints a sample override file for secrets/image tags

### Configure kubeconfig
```bash
aws eks update-kubeconfig \
  --region us-east-1 \
  --name aegis-spoke-prod \
  --profile myclaude
```

### Generate Helm values (script can deploy for you)
```bash
./generate-helm-values.sh
```
When the script finishes it prompts `Do you want to deploy now?`:

- **Type `y`** to let it run Helm immediately, layering the cloud value files with the overrides it just produced.
- **Type `n`** to stop after generation; the script still writes the overlays and prints the Helm command so you can execute it manually later.

Artifacts generated each run:
- `charts/aegis-services/values-cloud-generated.yaml`
- `charts/aegis-spoke/values-cloud-generated.yaml`
- `overrides.yaml` (image tags, DATABASE_URL, secrets)

### Create secrets (once per cluster)
```bash
kubectl create namespace aegis-system --dry-run=client -o yaml | kubectl apply -f -

kubectl create secret generic aegis-platform-secrets \
  --from-literal=db-password="$(terraform output -raw db_password_secret_value)" \
  --from-literal=proxy-jwt-secret="$(terraform output -raw jwt_secret_value)" \
  --namespace aegis-system --dry-run=client -o yaml | kubectl apply -f -
```

### Install/upgrade via Helm
```bash
# Hub
helm upgrade --install aegis-services charts/aegis-services \
  -f charts/aegis-services/values/common.yaml \
  -f charts/aegis-services/values/cloud.yaml \
  -f charts/aegis-services/values-cloud-generated.yaml \
  -f overrides.yaml \
  --namespace aegis-system --create-namespace

# Spoke
helm upgrade --install aegis-spoke charts/aegis-spoke \
  -f charts/aegis-spoke/values-cloud.yaml \
  -f charts/aegis-spoke/values-cloud-generated.yaml \
  -f overrides.yaml \
  --namespace aegis-system
```
`overrides.yaml` is the small file printed by the generation script (it sets image tags, DATABASE_URL, secrets, etc.). If you answered `y` when prompted, the script already ran the Helm commands above.

### Backstage against cloud HTTP
```bash
cd aegis-platform
yarn dev:cloud      # copies app-config.cloud.yaml
```

---

## 3. Cloud / EKS with TLS

Use the same Terraform apply.

### Generate TLS assets + values
```bash
./generate-helm-values.sh
```
The prompt behaves the same way as before (`y` to deploy now, `n` to just write the files). TLS is now enforced by default, and the script emits:
- `tls-overrides.yaml`

### Deploy with TLS-enabled services (manual path)
```bash
helm upgrade --install aegis-services charts/aegis-services \
  -f charts/aegis-services/values/common.yaml \
  -f charts/aegis-services/values/cloud.yaml \
  -f charts/aegis-services/values-cloud-generated.yaml \
  -f overrides.yaml \
  -f tls-overrides.yaml \
  --namespace aegis-system --create-namespace
```
(Apply `charts/aegis-spoke/values-cloud-tls.yaml` alongside the spoke chart so the agent dials the hub over TLS.)

### Backstage against cloud TLS
```bash
cd aegis-platform
yarn dev:cloud-tls   # copies app-config.cloud-tls.yaml
```

---

## Switching environments

| Task | Command |
|------|---------|
| Docker Desktop → EKS | `aws eks update-kubeconfig --region us-east-1 --name aegis-spoke-prod --profile myclaude` |
| EKS → Docker Desktop | `kubectl config use-context docker-desktop` |
| Local Backstage | `yarn dev` |
| Cloud Backstage | `yarn dev:cloud-tls` |
| Remove local stack | `helm uninstall aegis-services aegis-spoke -n aegis-system` |

---

## Handy status commands
```bash
kubectl get pods -n aegis-system
kubectl logs deployment/aegis-services-platform-api -n aegis-system
kubectl logs deployment/aegis-services-proxy -n aegis-system
kubectl logs deployment/aegis-spoke-aegis-spoke-k8s-agent -n aegis-system
```

Need to rotate backplane secrets? Update `overrides.yaml` and rerun the Helm upgrade.

That’s it—you can now develop locally, test cloud HTTP, or flip on TLS without rewriting charts.
