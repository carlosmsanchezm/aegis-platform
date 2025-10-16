# Codex Agent Operating Guide

Follow these rules on every task in this repository.

1. Jira hygiene
   - Immediately transition the ticket to **In Progress** via `mcp__jira-local__transition_issue` when work begins.
   - Add a `codex-in-progress` label (alongside any existing labels) using `mcp__jira-local__set_fields`.
   - Upon completion (after PR and validations are green), remove `codex-in-progress`, add `automation-complete`, and add a Jira comment with the result plus PR link.

2. Local development workflow
 - Prefer MCP filesystem tools (`repo-files.*`) for reading and editing files.
  - Preserve existing structure when editing. Avoid wholesale file rewrites unless absolutely necessary; implement targeted changes instead.
  - Run `mcp__aegis__go_build` before any tests.
  - Run `mcp__aegis__run_tests` with the default command (`./scripts/test-local-with-tls.sh`). The MCP wrapper has a 10-minute timeout—do not re-run the script manually unless it fails.
  - If the test tool reports a failure, inspect the MCP output/logs and fix the issue; avoid running the script directly unless you need to capture additional logs (use `timeout_ms` > 600000 when shelling out).
  - Track local failures: retry the build+test loop up to three times. On the third failure, gather the failing stdout/stderr (truncate if necessary), note any diagnostics executed (e.g., `helm` or `kubectl` output), capture the script exit code, and call `mcp__gpt5__advise` with that context plus diffs, allow_paths, and acceptance. Apply the returned guidance before continuing.

3. Remote validation
   - After local tests pass, push the branch and ensure a pull request is open. The preview workflow runs automatically on `pull_request` events; do not trigger it manually unless instructed.
   - Use the GitHub CLI to monitor the run: poll with `gh run list --workflow preview-deployment.yml --branch <branch> --limit 1` until it appears, then `gh run watch --workflow preview-deployment.yml --branch <branch>` to stream logs. Sleep at least 120 seconds between polls (max ~30 minutes).
   - If the run fails, gather diagnostics with `gh run view --workflow preview-deployment.yml --branch <branch> --log --exit-status`, fix the issue, and push again. Repeat at most three times; on the third failure, call `mcp__gpt5__advise` with the failing diff, workflow logs, and a summary of prior remediation attempts before trying another fix.
   - Record the final workflow URL and status in your summary.

4. Git etiquette
   - Create branches named `aegis-ci/<JiraKey>-<slug>` and reference the ticket in commits and PRs.
   - Push the branch and open a PR with validation evidence and noted risks.

5. Safety
   - Only dispatch `preview` or `ci` workflows unless explicitly instructed otherwise.
   - Abort (and report) if a command hangs for more than 10 minutes with no output after retries.
