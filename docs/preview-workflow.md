# Preview Deployment Workflow Guide

The `Preview: Provision, Test & Optional Destroy` workflow ( `.github/workflows/preview-deployment.yml` ) is our end-to-end validation pipeline for the cloud preview environment.  It can lint and unit-test the codebase, build and publish container images, deploy or refresh a preview stack, exercise the end-to-end suites, promote images, and optionally tear everything down.

Understanding the available run subsets (and when to use them) keeps remote validation fast and predictable—especially for AI coding agents that need clear instructions on what to trigger.

---

## Quick Reference

| Run Subset             | Builds Images | Deploys / Refreshes Preview | Runs E2E Suites | Promotes Images | Tears Down Preview | Primary Use Case |
|------------------------|---------------|-----------------------------|-----------------|-----------------|--------------------|------------------|
| `all` (default)        | ✅             | ✅                           | ✅               | ✅               | ✅                 | Final validation before merge. Ensures fresh images, promotion, and cleanup. |
| `deploy`               | ❌             | ✅                           | ✅               | ❌               | ✅                 | Smoke test a fresh deploy using previously published images. |
| `deploy-no-teardown`   | ❌             | ✅                           | ✅               | ❌               | ❌                 | Stand up / refresh preview and leave it running for manual QA. |
| `tests-only`           | ❌             | ♻️ (reuses existing stack)   | ✅               | ❌               | ❌                 | Fast remote verification after local tests. Recommended for day-to-day automation. |
| `provision-only`       | ❌             | ✅ (infra only)              | ❌               | ❌               | ❌                 | Rerun Terraform when investigating provisioning issues. |
| `destroy-only`         | ❌             | ❌                           | ❌               | ❌               | ✅                 | Tear down an existing preview namespace/infra when you’re done. |

Legend:

- ✅ — Step always runs in this subset.
- ❌ — Step is skipped.
- ♻️ — Deploy step reuses the existing preview unless the stack is missing (in which case it is recreated with cached images).

---

## Recommended Flows for Automation Agents

1. **Local validation:** run `make clean-local`, `make deploy-local-tls`, and `scripts/test-local-with-tls.sh` until green.
2. **Remote smoke test:** trigger the workflow with `run_subset=tests-only`. This keeps turnaround under ~25 minutes because it reuses registries and the existing preview. Supply `preview_number` so the agent targets the PR’s namespace.
3. **Full promotion cycle:** once code is ready to merge, dispatch `run_subset=all` (or rely on the PR-triggered run). This rebuilds/pushes images, exercises the full matrix, promotes `latest`, and tears down the preview so we leave no infra behind.
4. **Cleanup if needed:** if a preview must be removed manually, dispatch `run_subset=destroy-only` with the appropriate `preview_number`.

> **Tip:** Only use `deploy` / `deploy-no-teardown` if you explicitly need to redeploy without rebuilding images. Most remote testing should stick to `tests-only` (fast) or `all` (final check).

---

## Dispatch Inputs

All manual runs accept the following inputs.  When in doubt, set them explicitly so agents don’t guess defaults.

| Input            | Required | Description |
|------------------|----------|-------------|
| `run_subset`     | ✅        | One of the subsets above (`all`, `tests-only`, etc.). |
| `preview_number` | ⚪️        | Preview identifier. If omitted, the workflow uses PR number (for pull_request events) or the workflow run id. For deterministic cleanup, set this to the PR number. |
| `image_tag`      | ⚪️        | Commit/tag to test for `deploy`, `deploy-no-teardown`, or `tests-only`. Must already exist in ECR. Leave blank to use the baked-in default SHA (`DEFAULT_E2E_IMAGE_TAG`). |
| `suites`         | ⚪️        | Comma/space-separated list (`all`, `platform-api`, `workspace`, `k8s-agent`). Leave blank for `all`. |
| `teardown_policy`| ⚪️        | Controls auto teardown in non destroy-only runs (`auto-on-success`, `always`, `never`, `manual`). Usually leave at default. |
| `keep_on_failure`| ⚪️        | If `true`, leaves preview up when E2E suites fail. Agents normally keep this at `true` so humans can debug live systems. |

---

## Triggering the Workflow

### GitHub UI

1. Navigate to **Actions → Preview: Provision, Test & Optional Destroy**.
2. Click **Run workflow**.
3. Choose the branch (e.g. `aegis-ci/MVP-21-preview-workflow`).
4. Fill in the inputs (at minimum `run_subset`).
5. Click **Run workflow**.

### GitHub CLI Examples

```bash
# 1. Fast remote verification (reuses existing preview & images)
gh workflow run preview-deployment.yml \
  --ref aegis-ci/MVP-21-preview-workflow \
  -f run_subset=tests-only \
  -f preview_number=26 \
  -f suites=all

# 2. Full promotion cycle (builds, tests, promotes, tears down)
gh workflow run preview-deployment.yml \
  --ref aegis-ci/MVP-21-preview-workflow \
  -f run_subset=all \
  -f preview_number=26

# 3. Deploy and leave preview running for manual QA (no builds)
gh workflow run preview-deployment.yml \
  --ref aegis-ci/MVP-21-preview-workflow \
  -f run_subset=deploy-no-teardown \
  -f preview_number=26 \
  -f image_tag=6d407aa31e3be48ff7d4ad6ae0c98fe2770640c0

# 4. Tear down the preview once testing is done
gh workflow run preview-deployment.yml \
  --ref aegis-ci/MVP-21-preview-workflow \
  -f run_subset=destroy-only \
  -f preview_number=26
```

> When supplying `image_tag`, confirm the tag already exists in `567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/<service>:<tag>`. The workflow never builds images in deploy/tests-only modes.

---

## What Each Job Does

- **resolve-settings** — Normalizes inputs (preview id, run subset, suites) so downstream jobs use consistent values.
- **lint-and-unit-test** — Runs Go and frontend test suites. Skipped whenever `run_subset` isn’t `all`.
- **build-and-push** — Matrix build for platform API, proxy, k8s-agent, workspace images. Stores the resulting tags for later jobs. Skipped outside `all`/`deploy`/`deploy-no-teardown`.
- **collect-built-images** — Aggregates matrix artifacts into single outputs the rest of the workflow can consume safely.
- **provision-infra** — Terraform apply for infra-backed runs (all deploy-related subsets). Sets flags so teardown knows whether infra exists.
- **deploy-control-plane** — Runs Helm/Terraform helper to (re)deploy the preview control plane, generates TLS certs, updates Route53, and optionally runs pulumi preview for the spoke namespace. Reuses deploys for `tests-only`.
- **e2e-endpoints** — Resolves platform/proxy DNS & TLS secrets, exporting the data the E2E suites consume.
- **e2e-tests** — Matrix suites for platform API, workspace, and k8s-agent. In `tests-only` mode this is the primary job you care about.
- **promote-images** — Only runs in `run_subset=all` after successful E2Es. Promotes the freshly built images to `:latest`.
- **teardown-preview** — Destroys the preview stack when policy allows. Automatically runs for `all`/`deploy` (and can be forced manually via `destroy-only`).

---

## Interpreting Run Time

- `all` takes ~80–90 minutes because it includes lint/unit tests, four Docker builds/pushes, AWS deploy (with load balancer + Route53 waits), full E2E matrix, promotion, and teardown.
- `tests-only` usually finishes in ~20–25 minutes. Use this for day-to-day cloud verification.
- Build-heavy subsets (`all`, `deploy`) benefit from Docker caching but still cost 10–20 minutes for the platform API image.

---

## Common Questions

**Which images does `tests-only` use?**

- If you pass `image_tag`, that exact tag in ECR is used for all services.
- If you leave it blank, the workflow falls back to `DEFAULT_E2E_IMAGE_TAG` (a known-good SHA baked into the workflow).

**Does `tests-only` tear down the preview?**

- No. It reuses the existing namespace and leaves it running when done. Use `destroy-only` to clean up, or run `all` when you’re ready for a full promote+teardown cycle.

**When should agents run `destroy-only`?**

- After manual QA or any `deploy-no-teardown` run to ensure we don’t leak resources. Always set the correct `preview_number` so the workflow destroys the intended namespace.

**Where do the image refs end up?**

- Even in matrix builds, the workflow collects them via artifacts and publishes stable outputs in `collect-built-images`. Downstream jobs (deploy/tests/promote) always consume those aggregated outputs.

---

## Monitoring Runs

- List recent executions: `gh run list --workflow preview-deployment.yml --branch <branch>`
- Inspect logs: `gh run view <run-id> --log --exit-status`
- Kubernetes namespace: `kubectl get pods -n preview-<id>`

The workflow posts summary annotations for failed suites, and the workspace smoke test logs helpful kubeconfig diagnostics before trying to connect.

---

## Summary

- **Local first, remote second.** Agents should ship only after local `make` + scripts succeed.
- **Use `tests-only` for quick cloud validation.** It’s the fastest way to prove remote correctness without rebuilding images.
- **Reserve `all` for final promotion.** That run guarantees new images are published, smoke-tested, and the preview is torn down.
- **Clean up with `destroy-only`** whenever you’ve left a preview running for manual checks.

Following these conventions keeps our pipeline fast, predictable, and inexpensive—whether the workflow is triggered by a human or an AI coding agent.
