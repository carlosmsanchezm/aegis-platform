# Images on GitHub Container Registry (not ECR)

## Model

| Layer | Where |
|-------|--------|
| **Source code** | GitHub (`aegis-platform`, `aegis-ui`) |
| **Container images** | **ghcr.io** (GitHub Packages / Container Registry) |
| **Cluster / network / DB** | AWS (EKS, VPC, RDS-shaped resources) |
| **Laptop** | `aws`, `terraform`, `kubectl`, `helm` only — **no Docker** |

App images do **not** need to live in ECR. ECR is AWS-specific and couples “build” to one account.  
Because code is already on GitHub, **GitHub Actions builds → push to GHCR**; EKS **pulls** those tags.

> “GitHub Releases” in the product sense = release **tags** that also tag GHCR images  
> (`v1.2.3` + short SHA). The registry itself is **ghcr.io** (OCI), which is what Kubernetes pulls.

## Image names

| Image | GHCR path |
|-------|-----------|
| platform-api | `ghcr.io/carlosmsanchezm/aegis/platform-api:<sha>` |
| proxy | `ghcr.io/carlosmsanchezm/aegis/proxy:<sha>` |
| k8s-agent | `ghcr.io/carlosmsanchezm/aegis/k8s-agent:<sha>` |
| workspace (optional) | `ghcr.io/carlosmsanchezm/aegis/workspace-vscode:<sha>` |
| ui | `ghcr.io/carlosmsanchezm/aegis/ui:<sha>` (built from **aegis-ui** repo) |

## CI

Workflow: `.github/workflows/build-images.yml`

- Trigger: push to `main` / `feature/ironbank-compliance` (paths that affect images), or **workflow_dispatch**
- Runner: GitHub-hosted (`ubuntu-latest`), `linux/amd64`
- Auth: `GITHUB_TOKEN` → `packages: write`
- Push: multi-image packages under `ghcr.io/<owner>/aegis/*`

### First-time package visibility

New packages default to **private**. Either:

1. GitHub → Packages → each package → **Change visibility → Public** (simplest for lab), or  
2. Keep private and set on the laptop / cluster:

```bash
export GHCR_PULL_TOKEN=<PAT with read:packages>
export GHCR_USERNAME=carlosmsanchezm
```

Deploy creates `ghcr-registry-secret` in the cluster for `imagePullSecrets`.

## Deploy from laptop

```bash
export AWS_PROFILE=aegis-lab
export IMAGE_REGISTRY=ghcr.io/carlosmsanchezm/aegis
export CLOUD_IMAGE_TAG=<short_sha_from_Actions>   # e.g. 049cd32
# only if packages are private:
# export GHCR_PULL_TOKEN=ghp_...

./scripts/hub-eks.sh up
```

This does **not** build images. It:

1. Ensures AWS infra (terraform)  
2. Checks GHCR for tags  
3. Helm-installs using those image refs  
4. Verifies health (port-forward)

## UI repo

`aegis-ui` should have the same pattern: Actions → `ghcr.io/.../aegis/ui:<tag>`.  
Platform deploy points `UI_IMAGE` at that tag when present.

## What about ECR?

| Use | Role |
|-----|------|
| **GHCR** | Source of truth for **Aegis app** images (platform, proxy, agent, ui) |
| **ECR** | Optional / legacy; not required for lab hub if using GHCR |
| **AWS** | Runtime only (EKS pulls from ghcr.io) |

Terraform may still create ECR repos if `manage_ecr_repositories=true`; they are unused by the GHCR path.

## Related

- `docs/HUB-EKS-DEPLOY.md` — hub-eks phases  
- `.github/workflows/build-images.yml` — CI definition  
- `docs/CI-GITLAB-IMAGES.md` — older GitLab/ECR notes (superseded for lab by this doc)
