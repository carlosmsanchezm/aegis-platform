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
export AWS_ECR_REGISTRY=471147325433.dkr.ecr.us-east-1.amazonaws.com
# Tag from green GitLab build-images job (required — images never built on laptop by default)
export CLOUD_IMAGE_TAG=<CI_COMMIT_SHORT_SHA>

# 1) Sanity
./scripts/hub-eks.sh preflight

# 2) Full path: terraform → verify ECR tags → app → verify  (no local Docker)
./scripts/hub-eks.sh up

# Or step by step:
./scripts/hub-eks.sh terraform
./scripts/hub-eks.sh app
./scripts/hub-eks.sh verify

# Explicit local image build only if you really need it (needs Docker):
# ./scripts/hub-eks.sh images
# ./scripts/hub-eks.sh up --build-images

# Resume after a failed terraform:
./scripts/hub-eks.sh up --skip-terraform

# Teardown
./scripts/hub-eks.sh down
./scripts/hub-eks.sh status
```

### Image builds (GitLab CI — preferred)

See `docs/CI-GITLAB-IMAGES.md`. Runners build and push; laptop only deploys.

### Iron Bank builds (CI manual job or rare local)

```bash
# Prefer GitLab job build-images-ironbank
# Local only if required:
IMAGE_FLAVOR=ironbank ./scripts/hub-eks.sh images
```

## Phases

| Phase | What it does |
|-------|----------------|
| **preflight** | aws/terraform/kubectl/helm; `sts get-caller-identity` (Docker only if `--build-images`) |
| **terraform** | `terraform init -reconfigure` + `apply`; kubeconfig |
| **images** | **Opt-in only** (`images` cmd or `--build-images`): ECR login + `make push-cloud-images` |
| **app** | Helm hub via `scripts/hub-app/deploy-app.sh` (expects tags already in ECR) |
| **verify** | rollout status + port-forward `/healthz` |
| **down** | helm uninstall + terraform destroy (+ ECR force-delete retry) |

On default `up`, after terraform the script **verifies ECR tags** for `CLOUD_IMAGE_TAG` instead of building.

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
