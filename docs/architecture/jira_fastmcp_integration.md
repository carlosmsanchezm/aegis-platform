# ÆGIS Jira FastMCP Integration

## Overview
- Local **FastMCP Jira bridge** using Basic auth (service account). (Basic auth with API token + email) :contentReference[oaicite:15]{index=15}
- Headless **coordinator** polls Jira via JQL and drives the existing analyze→implement→review→test→PR loop.

## MCP Server
- Tools:
  - `search_issues` → POST `/rest/api/3/search/jql` (fallback `/search`) :contentReference[oaicite:16]{index=16}
  - `get_issue` → GET `/rest/api/3/issue/{key}` :contentReference[oaicite:17]{index=17}
  - `add_comment` → POST `/rest/api/3/issue/{key}/comment` (ADF body) :contentReference[oaicite:18]{index=18}
  - `transition_issue` → GET/POST `/rest/api/3/issue/{key}/transitions` :contentReference[oaicite:19]{index=19}
  - `set_fields` → PUT `/rest/api/3/issue/{key}` :contentReference[oaicite:20]{index=20}

## Plan Mapping (Jira → Plan)
- `issue_key`: Jira key
- `issue_type`: IssueType.name
- `summary`: Summary
- `acceptance`: parse ADF Description under “Acceptance Criteria”/“AC:”; prefer taskList/taskItem; fallback bullet/ordered lists. :contentReference[oaicite:21]{index=21}
- `component`: Components list
- `file_hints`: fenced “File Hints” block in Description or custom field
- `paths`: file_hints + component→repo path mapping via `.aegis/config.yaml`
- `allow_paths`: normalized globs (default `['services/**','charts/**']`)
- `max_files`: custom field default 8
- `risk`: custom Risk (default Medium); block High unless `auto-approve`
- `tests_required`/`docs_required`: defaults by type; override via labels `no-tests`/`no-docs`
- `breaking_change` / `requires_migration`: detect from labels/summary
- `pr.*`: branch prefix `jira-{key}`, title `[{key}] {summary}`, base from env or config

## Guardrails
- Path guardrails (allow_paths)
- Change budget: `max_files`, `max_loc_delta` (config; default 800). If exceeded: comment + `needs-approval`; do **not** auto-transition.
- Risk gates:
  - Risk=High → require `auto-approve` label
  - Breaking/migration → open PR but leave in **In Progress**; comment “Pending architect approval”.
- Never auto-merge.

## Rate Limits
- Honor `Retry-After` header and use exponential backoff with jitter for 429/5xx. :contentReference[oaicite:22]{index=22}

## Status Workflow
- Configurable (`Backlog → In Progress → In Review → Done`).

## Security
- Service account with least privilege (read/comment/transition/update).
- Redact secrets in logs; write audit JSON to `.aegis/logs/jira_mcp_audit.jsonl`.

## Example JQL
`project = AEG AND labels = auto-implement AND status = Backlog AND issuetype in (Story, Task, Bug, Refactor) AND "Risk" != High`

## Example PR body & Jira comment
- PR title: `"[AEG-123] Add rate limit telemetry"`
- PR body: Summary, Acceptance checklist, Affected paths, Risk, Tests/Docs notes, Evidence list
- Jira comment (ADF): bullets with PR URL, file count, LOC delta, evidence links.
