# Swarm Automation Test Configuration

This guide documents the environment settings and workflow needed to run the AI agent swarm end‑to‑end with **one attempt per stage**, full HTTP + TLS coverage locally, and a follow‑up GitHub Actions remote CI run. Use it as a checklist before launching `python -m agents.swarm.coordinator`.

---

## 1. Required Credentials

| Purpose | Variable | Notes |
|---------|----------|-------|
| Anthropic/Claude access | `ANTHROPIC_API_KEY` | Required for Analyzer / Implementer / Reviewer roles. |
| GPT‑5 escalation | `OPENAI_API_KEY` | Required for advisors invoked after escalation. |
| GitHub REST API | `GITHUB_PERSONAL_ACCESS_TOKEN` | Must have `repo` + `workflow` scopes so the swarm can push branches and dispatch Actions workflows. |
| Jira access | `JIRA_BASE_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN`, `JIRA_PROJECT_KEY` | Allows issue polling, state transitions, and comment updates. |

> Keep these secrets in a local `.env` file or secrets manager—never commit them.

---

## 2. Local Test Defaults (HTTP + TLS)

| Variable | Default | Recommended for full coverage |
|----------|---------|-------------------------------|
| `AEGIS_TEST_CMD` | *(unset)* | `./scripts/test-local-with-tls.sh` to run HTTP first, then TLS. |
| `RUN_LINT` | `1` | Keep `1` so Go/Node lint checks run. |
| `HTTP_RUN_E2E_PLATFORM` | `0` | Set to `1` for HTTP platform E2Es. |
| `HTTP_RUN_E2E_OPERATOR` | `0` | Set to `1` for HTTP operator E2Es. |
| `TLS_RUN_E2E_PLATFORM` | `1` | Leave at default for TLS E2Es. |
| `TLS_RUN_E2E_OPERATOR` | `1` | Leave at default for TLS E2Es. |

The helper script writes local evidence to `.aegis/evidence/<timestamp>_ci_results.json`. Review that file (and the console output) to confirm results.

---

## 3. Attempt Budgets & Escalation

Set the attempt ceilings to **one try** for both stages so the swarm either succeeds immediately or escalates back to you:

```bash
export AEGIS_LOCAL_MAX_ATTEMPTS=1
export AEGIS_LOCAL_ESCALATE_AFTER=1
export AEGIS_REMOTE_MAX_ATTEMPTS=1
export AEGIS_REMOTE_ESCALATE_AFTER=1
```

The local loop finishes on success; on failure it captures evidence, invokes the Failure Analyst, and exits because escalation happens after the single allowed attempt. The remote loop mirrors that behaviour for the GitHub Actions run.

---

## 4. Remote CI Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `AEGIS_REMOTE_CI_ENABLE` | `true` | Leave enabled so remote CI is required. |
| `AEGIS_REMOTE_CI_WORKFLOW` | `preview-deployment.yml` | Matches `.github/workflows/preview-deployment.yml`, which already accepts `workflow_dispatch` and `push` on `aegis-ci-*`. |
| `AEGIS_GITHUB_OWNER`, `AEGIS_GITHUB_REPO`, `AEGIS_GITHUB_BASE` | *(required)* | Identifies the repo and base branch (e.g., `main`). |
| `AEGIS_REMOTE_CI_TIMEOUT_SECONDS` | `5400` | Optional. No need to lower it; Actions handles its own timeouts. |

Artifacts from this phase land in `.aegis/evidence/*remote_ci_results.json` and `*remote_ci_logs.zip`. View the zip locally for detailed job logs; the GitHub Actions run is also visible on the repository’s Actions tab.

---

## 5. Suggested Environment Script

```bash
# Core credentials
export ANTHROPIC_API_KEY=...
export OPENAI_API_KEY=...
export GITHUB_PERSONAL_ACCESS_TOKEN=...

# Jira
export JIRA_BASE_URL="https://<your-domain>.atlassian.net"
export JIRA_EMAIL="you@example.com"
export JIRA_API_TOKEN=...
export JIRA_PROJECT_KEY="MVP"

# Local test command & full HTTP/TLS coverage
export AEGIS_TEST_CMD="./scripts/test-local-with-tls.sh"
export RUN_LINT=1
export HTTP_RUN_E2E_PLATFORM=1
export HTTP_RUN_E2E_OPERATOR=1
export TLS_RUN_E2E_PLATFORM=1
export TLS_RUN_E2E_OPERATOR=1

# One attempt per stage
export AEGIS_LOCAL_MAX_ATTEMPTS=1
export AEGIS_LOCAL_ESCALATE_AFTER=1
export AEGIS_REMOTE_MAX_ATTEMPTS=1
export AEGIS_REMOTE_ESCALATE_AFTER=1

# Remote CI wiring
export AEGIS_REMOTE_CI_ENABLE=true
export AEGIS_REMOTE_CI_WORKFLOW="preview-deployment.yml"
export AEGIS_GITHUB_OWNER="carlosmsanchezm"
export AEGIS_GITHUB_REPO="aegis"
export AEGIS_GITHUB_BASE="main"
```

Tip: store these in an `.env.swarm` file and `source` it before running the coordinator.

---

## 6. Running & Reviewing Results

1. Ensure Docker Desktop + Kubernetes are running and clean the environment (`make clean-local` and `scripts/test-cleanup.sh` if needed).
2. `source .env.swarm` (or export the variables manually).
3. `python -m agents.swarm.coordinator`
   - Local tests should complete within ~5 minutes; the script handles its own internal waits, so no extra timeout variable is required.
   - Evidence accumulates under `.aegis/evidence/` for both local and remote stages.
4. On success the swarm opens a PR from an `aegis-ci-*` branch; check Jira for the comment containing the PR link and evidence list.

If any stage fails, inspect the latest `*_ci_results.json` and `*_diff_summary.json` files for diagnostics before re-running.

---

## 7. Troubleshooting Checklist

- **Missing evidence files:** confirm `.aegis/evidence/` exists and the coordinator has write access.
- **GitHub workflow not launching:** verify the PAT scope, workflow name (`preview-deployment.yml`), and that the workflow file has `workflow_dispatch` + `push` triggers.
- **Jira automation errors:** check `servers/jira/jira_mcp_server.py` logs and validate the Jira token/email/base URL.
- **TLS phase failure:** delete stale CA file (`rm ~/aegis-platform-api-ca.crt`) and rerun `make deploy-local-tls` manually to confirm the cluster state.

This setup ensures the swarm covers the full local suite, escalates immediately on failure, runs remote CI exactly once, and leaves a clear evidence trail for both stages.

---

## Jira Ticket Template

Use this structure so the swarm receives clear scope, guardrails, and evidence expectations.

```markdown
h2. Summary
<One sentence describing the change.>

h2. Background
<Why the change is needed; keep it short and factual.>

h2. Scope / Requirements
* <Explicit bullet list of actions the agent must perform.>
* <Call out any configuration or doc updates that are required.>

h2. Guardrails
* <List constraints: files to touch, approaches to avoid, libraries/frameworks that must stay.>
* <Call out anything that must NOT change (schema, dependencies, etc.).>

h2. Non-Goals
* <Bullets for items explicitly out of scope.>

h2. Acceptance Criteria
* <How we’ll verify success: behaviour, tests, evidence.>

h2. Affected Paths
* <Exact file paths/globs the agent may edit.>

h2. Test / Evidence Requirements
* <Specify commands (e.g., scripts/test-local-with-tls.sh) and the artifacts to attach.>
```
