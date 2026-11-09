# MCP Integration Playbook

This repo ships everything needed for Codex to talk to remote MCP servers (Atlassian, Figma, AWS) and the local STDIO servers that live under `servers/`. Follow this guide when you want the agent to run a ticket end-to-end, execute local CI, and stage a GitHub Actions dispatch.

---

## 1. Environment & Secrets

Copy `.env.example` to `.env` (or export the variables in your shell) and fill in production values. At a minimum you will need:

- `ATLASSIAN_SITE`, `JIRA_BASE_URL`, `JIRA_EMAIL`, `JIRA_API_TOKEN`, `JIRA_PROJECT_KEY` – Jira Cloud OAuth token with _Jira Software_, _Jira Service Management_, and _offline_access_ scopes.
- `GITHUB_PERSONAL_ACCESS_TOKEN`, `AEGIS_GITHUB_OWNER`, `AEGIS_GITHUB_REPO` – GitHub PAT with `repo`, `workflow`, and `pull_request:write` scopes so the agent can stage PRs and monitor workflow runs.
- `ANTHROPIC_API_KEY` – Claude API key with access to the models you plan to call (e.g. Claude 3.5 Sonnet).
- `OPENAI_API_KEY`, `AEGIS_OPENAI_MODEL` – OpenAI Responses API key capable of running the GPT-5 advisor MCP server (`gpt-5-pro-2025-10-06` by default).
- `AWS_PROFILE` **or** `AWS_ACCESS_KEY_ID`/`AWS_SECRET_ACCESS_KEY` (+ optional session token) plus `AWS_DEFAULT_REGION` – lets the AWS MCP servers enumerate services and docs.
- `OSCAL_CATALOG`, `FEDRAMP_BASELINES_DIR` – absolute paths to NIST SP 800-53 Rev5 JSON catalog and FedRAMP baseline JSON files. The repo does **not** include these datasets; fetch them and point the variables at the downloaded content.
- Optional helpers: `AEGIS_TEST_CMD` overrides the default `go test ./...` for the local CI server, `AEGIS_JIRA_AUDIT` customises where Jira audit logs are written.

Keep secrets out of version control. Use `direnv`, `dotenvx`, or your preferred secret store to surface them at runtime.

---

## 2. Codex configuration

Codex reads global MCP entries from `~/.codex/config.toml`. Ensure it contains:

- `experimental_use_rmcp_client = true`
- One `[mcp_servers.<name>]` table per entry (remote + local), matching the commands shown in `.mcp.json`.

After updating, restart Codex (or `codex mcp reload`) and verify with `codex mcp list`.

For project-level overrides, drop `.mcp.json` (see example below) into the repo root. Codex will prefer this when running with `--cd` inside the project.

```json
{
  "mcpServers": {
    "atlassian": {
      "type": "command",
      "command": "npx",
      "args": ["-y", "mcp-remote", "https://mcp.atlassian.com/v1/sse"]
    },
    "figma-remote": {
      "type": "command",
      "command": "npx",
      "args": ["-y", "mcp-remote", "https://mcp.figma.com/mcp"]
    },
    "aws-api": {
      "type": "command",
      "command": "uvx",
      "args": ["awslabs.aws-api-mcp-server@latest"]
    },
    "aws-docs": {
      "type": "command",
      "command": "uvx",
      "args": ["awslabs.aws-documentation-mcp-server@latest"]
    },
    "nist": {
      "type": "command",
      "command": "uv",
      "args": ["run", "--env-file", ".env", "--python", "3.11", "--with", "mcp", "servers/nist_oscal/nist_oscal_server.py"]
    },
    "jira-local": {
      "type": "command",
      "command": "uv",
      "args": ["run", "--env-file", ".env", "--python", "3.11", "--with", "mcp", "--with", "httpx", "servers/jira/jira_mcp_server.py"]
    },
    "gpt5": {
      "type": "command",
      "command": "uv",
      "args": ["run", "--env-file", ".env", "--python", "3.11", "--with", "mcp", "--with", "openai", "servers/gpt5_pro/gpt5_pro_mcp_server.py"]
    },
    "aegis": {
      "type": "command",
      "command": "uv",
      "args": ["run", "--env-file", ".env", "--python", "3.11", "--with", "mcp", "servers/aegis/aegis_mcp_server.py"]
    },
    "repo-files": {
      "type": "command",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/Users/carlossanchez/code/aegis-final-ui", "/Users/carlossanchez/code/aegis-final-ui/third_party/oscal-content", "/Users/carlossanchez/code/aegis-final-ui/third_party/fedramp-automation"]
    }
  }
}
```

---

## 3. Verification checklist

Once the servers register successfully, have the agent perform these smoke checks:

1. `tools/list` for each server and record `server → sample tools`. Failures usually mean missing environment variables or datasets.
2. **Jira:** `mcp__jira-local__get_issue` for a known key (no transitions without explicit instruction).
3. **NIST:** `mcp__nist__get_control` (e.g., `AC-2`) and `mcp__nist__in_fedramp_baseline` (`level="moderate"`). Missing OSCAL files cause these calls to fail.
4. **Local CI:** `mcp__aegis__go_build`, then `mcp__aegis__run_tests`. Capture stdout/stderr summaries for reporting.
5. **GitHub Actions:** On push/pr open `preview-deployment.yml` runs automatically. Monitor with `gh run list` / `gh run watch`. Use `github-actions.prepare_dispatch` only if manual dispatch is required.

Document the outputs in task logs so regressions are easy to trace.

---

## 4. Jira workflow (mirrors AGENTS.md)

- Transition tickets to “In Progress” when work starts; label with `codex-in-progress`.
- On completion, replace with `automation-complete` and comment with summary + PR link.

---

## 5. Misc. troubleshooting

- **Workspace tests timing out** – the MCP wrapper enforces a 10-minute limit; ensure `AEGIS_TEST_CMD` points to the correct script and that dependencies (Go, node, yarn, kubectl, grpcurl, jq) are installed.
- **Figma local** – optional. Remove the `figma-local` entry if the desktop bridge isn’t running.
- **Atlassian/Figma OAuth** – run `codex mcp login <server>` and complete the browser flow.
- **GitHub MCP** – requires `GITHUB_PERSONAL_ACCESS_TOKEN`, owner, repo in `.env`. If the server exits immediately, run the command manually to see the precise error.

With these pieces wired together, Codex can take a Jira ticket (e.g., “Implement MVP-5”), update the codebase via the filesystem server, run local CI, open a PR, and monitor GitHub Actions until it succeeds—all without manual babysitting.
