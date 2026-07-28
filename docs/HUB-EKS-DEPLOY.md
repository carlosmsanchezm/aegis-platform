# Deterministic hub deploy on EKS

**Prefer `./scripts/hub-eks.sh` over long Makefile targets** for cloud hub bring-up and teardown.

The Makefile is still useful for **individual image builds** and local Docker Desktop.  
For **“stand up hub on EKS and prove it”**, use the phased CLI.

## Why not only the Makefile?

| Makefile / old scripts | `hub-eks.sh` |
|------------------------|--------------|
| One long chain; error mid-way is opaque | **Named phases** with separate log files |
| Mixes local + cloud + preview CI | **Cloud hub only** |
| Hard to resume after failure | Re-run single phase: `images`, `app`, … |
| Multiple entrypoints (`deploy.sh`, commercial, make) | **One entrypoint** |

## Commands

```bash
cd ~/code/aegis-platform
export AWS_PROFILE=aegis-lab
export IMAGE_REGISTRY=ghcr.io/carlosmsanchezm/aegis
# Tag from green GitHub Actions "Build images (GHCR)" run
export CLOUD_IMAGE_TAG=<short_sha>
# If GHCR packages are private:
# export GHCR_PULL_TOKEN=<pat with read:packages>

./scripts/hub-eks.sh preflight
./scripts/hub-eks.sh up

# Or step by step:
./scripts/hub-eks.sh terraform
./scripts/hub-eks.sh app
./scripts/hub-eks.sh verify

# Teardown
./scripts/hub-eks.sh down
./scripts/hub-eks.sh status
```

### Image builds (GitHub → GHCR — required)

See **`docs/CI-GHCR-IMAGES.md`**.  
GitHub Actions builds `linux/amd64` and pushes to **ghcr.io**. Laptop never builds.  
App images are **not** stored in ECR.

## Phases

| Phase | What it does |
|-------|----------------|
| **preflight** | aws/terraform/kubectl/helm; `sts get-caller-identity` (no Docker) |
| **terraform** | `terraform init -reconfigure` + `apply`; kubeconfig |
| **images** | **Disabled** — use GHCR CI |
| **app** | Helm hub (image refs from `IMAGE_REGISTRY` / GHCR) |
| **verify** | rollout status + port-forward `/healthz` |
| **down** | helm uninstall + terraform destroy |

On `up`, after terraform the script **verifies GHCR tags** for `CLOUD_IMAGE_TAG`.

## Logs

Every run writes:

```text
.deploy-logs/hub-eks-<timestamp>.log
.deploy-logs/<timestamp>-preflight.log
.deploy-logs/<timestamp>-terraform.log
...
```

On failure the script prints the **phase name** and the **last 40 lines** of that phase log.

Add `.deploy-logs/` to `.gitignore` if not already ignored.

## Thin Makefile wrappers (optional)

```makefile
hub-up:    ; ./scripts/hub-eks.sh up
hub-down:  ; ./scripts/hub-eks.sh down
hub-status:; ./scripts/hub-eks.sh status
```

## App phase (Helm) implementation

| Path | Role |
|------|------|
| `scripts/hub-app/deploy-app.sh` | **Phased** app deploy (preflight → kube → secrets → pki → helm → migrations → LBs → verify) |
| `scripts/hub-app/lib.sh` | Shared helpers (ECR verify, LB wait, Cloudflare) |
| `terraform/generate-cloud-deployment.sh` | Thin **wrapper** → `deploy-app.sh` (old callers still work) |
| `scripts/hub-app/generate-cloud-deployment.sh.orig` | Backup of the previous monolith |

**Lab-friendly defaults in app deploy:**
- `SKIP_DNS_UPDATE=1` — no Cloudflare; port-forward / raw LB for health
- `REUSE_EXISTING=1` — helm upgrade without deleting the namespace
- `AWS_PROFILE=aegis-lab`

## What this does *not* replace yet

- Local Docker Desktop: still `make deploy-local-tls` / `scripts/aegis.sh`
- Preview CI / Graphite pipelines

Long-term: Helmfile or CI calling `hub-eks.sh up` with OIDC; Terraform workspaces per env.

## Related

- `docs/IMAGE-BASES.md` — public vs ironbank
- `docs/business/LAB-AWS-ACCOUNT.md` — lab account
- `AGENT_DEPLOYMENT_GUIDE.md` — low-level deploy reference
