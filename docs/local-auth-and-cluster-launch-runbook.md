# Local Auth & Cluster Launch Runbook (TLS, Keycloak, Pulumi)

Use this checklist whenever the UI/cluster launch flow starts failing with 4xx/5xx (Keycloak auth issues, TLS errors, Pulumi token errors, “job already exists”, etc.).

## 1) Reset local stack and deploy with TLS
```bash
make clean-local
# build image (arm64 nodes)
PLATFORMS=linux/arm64 BUILD_ARGS="--platform linux/arm64" make build-platform
make deploy-local-tls
```
- This redeploys aegis-services + aegis-spoke, refreshes ingress/certs, and restarts Keycloak.

## 2) Trust certs locally
The deploy copies certs to `$HOME`:
- `~/aegis-platform-api-ca.crt`
- `~/keycloak.localtest.me.crt`
- Combined: `~/aegis-local-trust.pem`

For the frontend (and VS Code) set:
```bash
export NODE_EXTRA_CA_CERTS=$HOME/aegis-local-trust.pem
```
If using the browser UI, import the combined bundle into your system trust store (or set “Insecure origins” exception for `*.localtest.me`).

## 3) Keycloak sanity checks
```bash
kubectl get pods -n keycloak
```
- All pods should be Running; realm job Completed.
- If Keycloak ingress TLS is broken, redeploy (`make deploy-local-tls`) which re-patches the ingress.

## 4) Platform API env (Helm values)
Ensure these in `charts/aegis-services/values/local.yaml` (already set):
- `platformApi.pulumi.backendUrl`: `s3://aegis-pulumi-state-dev/aegis/pulumi`
- `platformApi.pulumi.secretsProvider`: KMS URI from `terraform/pulumi-stack` outputs.
- `platformApi.tls.enabled: true` (via local-tls overlay).
Plugins are pre-baked into the image; no runtime downloads are needed.

## 5) Pulumi IAM / backend
Terraform module `terraform/pulumi-stack` owns:
- IAM user `aegis-pulumi-provisioner` (admin + assume-role policy)
- S3 state bucket `aegis-pulumi-state-dev` (versioned, encrypted)
- KMS key `alias/aegis-pulumi-secrets-dev`
If state drifts, run:
```bash
cd terraform/pulumi-stack
terraform apply -var aws_profile=aegis-new
```

## 6) Common launch errors and fixes
- **401/403 from UI**: ensure Keycloak/realm is healthy; trust CA locally; set `NODE_EXTRA_CA_CERTS`.
- **504/timeout to platform-api**: ingress/webhook stale → `make clean-local && make deploy-local-tls`.
- **Pulumi “ACCESS_TOKEN” or plugin download errors**: backend/secrets set via Helm; plugins baked. Redeploy and confirm pods on latest image (`kubectl get pods -n aegis-system -l app.kubernetes.io/component=platform-api` and check imageID).
- **“cluster job … already exists”**: we auto-generate unique job IDs now. If old stuck CRs remain, delete them:
  ```bash
  kubectl delete projectinfras.infra.aegis.yourorg.dev --all -n aegis-system --force --grace-period=0
  ```

## 7) Frontend launch (Backstage)
Run with trusted certs:
```bash
cd aegis-platform
NODE_EXTRA_CA_CERTS=$HOME/aegis-local-trust.pem yarn dev
```
Target backend: `https://platform-api.localtest.me`.

## 8) Verify after rollout
```bash
kubectl -n aegis-system get pods
kubectl -n aegis-system logs deploy/aegis-services-platform-api --tail=50
curl -k https://platform-api.localtest.me/healthz
```

Keep this document updated if Helm values or Pulumi backend change. This flow has been validated with the current `:dev` image (`carlosmsanchez/aegis-platform-api:dev@sha256:83ff87f0da637401ad2fef4d0921bde1485c18ce26bc4059b0837c85db8f864a`). 
