# GitLab CI — build images on runner (no local Docker Desktop)

## Goal

| Who | Does what |
|-----|-----------|
| **GitLab Runner** | `docker build` + push to **ECR** |
| **You (laptop)** | `./scripts/hub-eks.sh up` — terraform + helm only (never builds images by default) |

No Docker Desktop required on your Mac for day-to-day hub deploys. Image builds only run if you explicitly pass `--build-images` or run `./scripts/hub-eks.sh images`.

## One-time GitLab setup

1. Push this repo to GitLab (or ensure `.gitlab-ci.yml` is on the default branch).
2. Project → **Settings → CI/CD → Variables** (mask secrets):

| Variable | Example | Notes |
|----------|---------|--------|
| `AWS_ACCESS_KEY_ID` | `AKIA…` | IAM user that can push to ECR + (optional) describe |
| `AWS_SECRET_ACCESS_KEY` | `…` | Masked |
| `AWS_DEFAULT_REGION` | `us-east-1` | Optional if set in YAML |
| `AWS_ECR_REGISTRY` | `471147325433.dkr.ecr.us-east-1.amazonaws.com` | Lab account |
| `AEGIS_UI_REPO_URL` | `https://gitlab.com/you/aegis-ui.git` | Optional; omit to skip UI |
| `AEGIS_UI_REPO_TOKEN` | deploy token | If UI repo is private |

3. Runner must support **Docker-in-Docker** (`docker:dind`) or privileged Docker builds.  
   If your runners use a tag, set it in `.gitlab-ci.yml` under `tags:`.

4. Ensure ECR repos exist (terraform apply once creates them with `manage_ecr_repositories = true`).

### IAM minimum for CI user

- `ecr:GetAuthorizationToken`
- `ecr:BatchCheckLayerAvailability`, `PutImage`, `InitiateLayerUpload`, `UploadLayerPart`, `CompleteLayerUpload`
- `ecr:BatchGetImage`, `DescribeImages`, `DescribeRepositories`

## Run a build

- **CI/CD → Pipelines → Run pipeline** (web), or push a commit that touches services/agents/Makefile.
- Job: **`build-images`** (`IMAGE_FLAVOR=public` by default).
- Manual job **`build-images-ironbank`** when you need Iron Bank bases (runner must reach `registry1.dso.mil`).

On success, artifacts include `build.env` with:

```bash
CLOUD_IMAGE_TAG=<short sha>
PLATFORM_API_IMAGE=.../aegis/platform-api:<sha>
...
```

## Deploy from laptop (no Docker)

```bash
cd ~/code/aegis-platform

export AWS_PROFILE=aegis-lab
export AWS_ECR_REGISTRY=471147325433.dkr.ecr.us-east-1.amazonaws.com
# Use the tag from the green pipeline:
export CLOUD_IMAGE_TAG=<paste CI_COMMIT_SHORT_SHA>

./scripts/hub-eks.sh up
```

`up` **never** builds images on the laptop. It checks ECR for `CLOUD_IMAGE_TAG`, then terraform + helm + verify. Docker is not required.

## Local vs CI

| Task | Local Docker | GitLab Runner |
|------|--------------|---------------|
| Build/push images | only with `--build-images` / `images` | **yes (default path)** |
| `hub-eks.sh up` | **no Docker** — pulls tags from ECR | n/a |
| terraform / helm | laptop or CI later | optional future job |

## Troubleshooting

| Issue | Fix |
|-------|-----|
| `Cannot connect to Docker daemon` on runner | Use `docker:dind` service + privileged runner, or Kaniko executor |
| ECR login failed | Check CI AWS vars; account must match registry |
| UI skipped | Set `AEGIS_UI_REPO_URL` or accept hub without Backstage image |
| Iron Bank job fails | Only for `IMAGE_FLAVOR=ironbank`; use public for lab |

## Related

- `docs/IMAGE-BASES.md` — public vs ironbank
- `docs/HUB-EKS-DEPLOY.md` — hub-eks phases
- `docs/business/LAB-AWS-ACCOUNT.md` — lab account IDs
