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

### Deploy the hub (platform-api + proxy)
```bash
helm upgrade --install aegis-services charts/aegis-services \
  -f charts/aegis-services/values/common.yaml \
  -f charts/aegis-services/values/local.yaml \
  --namespace aegis-system --create-namespace
```

### Deploy the spoke (k8s-agent)
```bash
helm upgrade --install aegis-spoke charts/aegis-spoke \
  -f charts/aegis-spoke/values.yaml \
  -f charts/aegis-spoke/values-local.yaml \
  --namespace aegis-system
```

> **Re-deploying?** Run `helm uninstall aegis-services aegis-spoke -n aegis-system` first.

### Port-forward for local access
```bash
# Platform API on http://localhost:10080 (HTTP) and :10081 (gRPC)
kubectl -n aegis-system port-forward svc/aegis-services-platform-api 10080:8080 10081:8081

# Proxy tunnel on http://localhost:10085/proxy/
kubectl -n aegis-system port-forward svc/aegis-services-proxy 10085:8085

# Stop existing forwards if needed
pkill -f "kubectl port-forward"  # optional cleanup
```

### Start Backstage against local services
```bash
cd aegis-platform
yarn dev          # copies app-config.local-dev.yaml
```

---

## 2. Cloud / EKS (plain HTTP)

### Provision infra
```bash
cd terraform
terraform init
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
./generate-helm-values.sh --tls
```
The prompt behaves the same way as the non-TLS run (`y` to deploy now, `n` to just write the files). This adds:
- `charts/aegis-services/values-cloud-tls.yaml`
- `tls-overrides.yaml`

### Deploy with TLS overlay (manual path)
```bash
helm upgrade --install aegis-services charts/aegis-services \
  -f charts/aegis-services/values/common.yaml \
  -f charts/aegis-services/values/cloud.yaml \
  -f charts/aegis-services/values-cloud-generated.yaml \
  -f overrides.yaml \
  -f charts/aegis-services/values-cloud-tls.yaml \
  -f tls-overrides.yaml \
  --namespace aegis-system --create-namespace
```
(Spoke deployment is unchanged unless you also apply `charts/aegis-spoke/values-cloud-tls.yaml` to enforce TLS when dialing the hub.)

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
| Cloud Backstage (HTTP) | `yarn dev:cloud` |
| Cloud Backstage (TLS) | `yarn dev:cloud-tls` |
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
