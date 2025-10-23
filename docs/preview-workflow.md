# Preview Deployment Workflow Guide

This document explains how our `Build, Test, and Deploy Preview` GitHub Actions workflow operates, what each job covers, and how to run specific portions of the workflow when iterating on fixes.

## Workflow Overview

The workflow lives at `.github/workflows/preview-deployment.yml` and is triggered:

- on every pull request, running the full pipeline by default, and
- on manual `workflow_dispatch` runs (supports running subsets of the pipeline).

It drives four major stages:

1. **Lint & Unit Tests (`lint-and-unit-test`)**  
   Runs Go and Node lint/unit suites across agents, services, proxy, workspace, and frontend packages.

2. **Build & Push Images (`build-and-push`)**  
   Builds the platform API, proxy, k8s-agent, and workspace images, pushes them to AWS ECR, and promotes the new commit SHA to the `latest` tag on success.

3. **Deploy Preview & Run E2E Tests (`deploy-preview-and-test`)**  
   Provisions/updates the preview stack via Terraform and Helm, generates TLS certs, updates Route53, waits for load balancers, then runs:
   - Platform API end-to-end suite (`scripts/e2e-platform-api.sh`),
   - Workspace connectivity smoke test (`scripts/test-workspace-connection.sh`),
   - K8s-agent end-to-end suite (`make test-e2e` in `agents/k8s-agent`).

4. **Teardown Preview (`teardown-preview`)**  
   Cleans up the preview stack (namespace, Helm releases, Terraform state) unless the run subset explicitly requests to skip teardown.

## Run Subset Modes

You can select a subset of the workflow when triggering manually:

| Mode                  | Description                                                                 |
|-----------------------|-----------------------------------------------------------------------------|
| `all`                 | Full pipeline: lint/unit, build & push, deploy, e2e tests, teardown.        |
| `deploy`              | Skip lint/unit/build; deploy stack and run e2e tests + teardown.            |
| `deploy-no-teardown`  | Same as `deploy` but leaves the preview environment running.               |
| `e2e-only`            | Assume preview already exists; run e2e suites without redeploy/teardown.    |

Pull requests now default to `run_subset=all` for complete coverage. Manual `workflow_dispatch` runs may pick any subset.

## Running Manually

### GitHub Actions UI

1. Go to **Actions** → **Build, Test, and Deploy Preview** → **Run workflow**.
2. Choose the branch to test (e.g., `MVP-10`).
3. Select a `run_subset`.
4. Optional: provide `image_tag` if skipping the build step (`deploy`, `e2e-only`). Use an image tag that already exists in ECR.
5. Click **Run workflow**.

### GitHub CLI

```bash
# Full pipeline (default)
gh workflow run preview-deployment.yml --ref MVP-10

# As-needed e2e without rebuild
gh workflow run preview-deployment.yml \
  --ref MVP-10 \
  -f run_subset=e2e-only \
  -f image_tag=4439705b81f1478e1b7e38529a171a9232feb691

# Deploy without teardown (keep preview running)
gh workflow run preview-deployment.yml \
  --ref MVP-10 \
  -f run_subset=deploy-no-teardown \
  -f image_tag=<your-tag>
```

## Image Promotion

When `build-and-push` runs it tags images with the commit SHA and, on success, promotes them to `latest` in:  
- `aegis/platform-api`  
- `aegis/proxy`  
- `aegis/k8s-agent`  
- `aegis/workspace-vscode`

That keeps downstream environments on a fresh `latest` after a successful run.

## Terraform & Helm

- Terraform provisions the preview infrastructure and outputs Helm values for the control plane and spoke.  
- Helm deploys the `preview-<id>` release, generating fresh TLS certs and updating Route53 so `platform-api-grpc.aegist.dev` and `proxy.aegist.dev` resolve via HTTPS/TLS.  
- For now we redeploy each run to avoid TLS secret drift. A reuse path may come back once cert persistence is automated.

## Troubleshooting

- **Helm TLS errors (`proxy.tls.cert must be provided`)** – ensure the workflow ran the TLS generation step. Manual runs should not delete the TLS secret unless they regenerate the cert.
- **Workspace session stuck in `PLACED`** – the smoke test sends `StartWorkload` using the assigned cluster ID before requesting a connection session; do the same if running locally.
- **Need faster iteration** – `run_subset=e2e-only` with a known `image_tag` avoids rebuilding or redeploying the control plane when the preview environment already exists.

### Inspecting Runs

- `gh run list --workflow preview-deployment.yml --branch MVP-10`  
- `gh run view <run-id> --log --exit-status`  
- `kubectl get pods -n preview-<id>` if you want to inspect the preview namespace manually.

## Future Enhancements

- Reintroduce a safe Helm/TLS reuse path once cert secret handling is automated.
- Optionally add UI/Backstage smoke tests or workspace persistence checks.
- Consider nightly `run_subset=deploy-no-teardown` runs to keep a long-lived preview for manual QA.

For questions or contributions, check the history of `.github/workflows/preview-deployment.yml` or reach out to the platform team.
