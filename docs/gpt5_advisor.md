# GPT-5 Pro Advisor Escalation

This repository keeps the existing Claude Agent SDK swarm intact and adds an escalation path that calls **OpenAI `gpt-5-pro`** via a lightweight Model Context Protocol (MCP) server. The Claude swarm still owns the workflow; GPT-5 Pro is only invoked when automated remediation needs additional guidance.

## Components

- `servers/gpt5_pro/gpt5_pro_mcp_server.py` exposes a single `advise` tool. It forwards diffs, failing test metadata, and guardrails to `gpt-5-pro` using the OpenAI Responses API and returns *strict* JSON `{summary, root_cause, edits, risk, confidence}`.
- `agents/claude/run_fedramp_swarm.py` registers the new MCP server, adds a `run_tests` tool, and introduces retry + escalation logic:
  1. Analyzer/Implementer/Reviewer complete as before.
  2. Tester runs `go build` then `run_tests`.
  3. After `AEGIS_ESCALATE_AFTER` failed attempts, the swarm calls `mcp__gpt5__advise`.
  4. GPT-5 instructions are appended to the Implementer inputs, Reviewer re-validates, and the Tester re-runs.
  5. Advice and CI artifacts are written to `.aegis/evidence/*_gpt5_advice.json` and `*_ci_results.json`.

## Configuration

`.aegis/config.yaml` now includes:

```yaml
openai:
  model: gpt-5-pro
  escalate_after_failures: 2
  max_test_attempts: 3

tests:
  command: "go test ./..."
```

## Required Environment

Set the following before launching the swarm (locally or in CI):

```bash
export OPENAI_API_KEY=sk-...
export AEGIS_OPENAI_MODEL=gpt-5-pro
export AEGIS_ESCALATE_AFTER=2
export AEGIS_MAX_TEST_ATTEMPTS=3
export AEGIS_TEST_CMD="go test ./..."
```

The existing Anthropic and GitHub credentials remain unchanged. The MCP server exits immediately if `OPENAI_API_KEY` is absent to avoid accidental escalation attempts without credentials.

## Operational Notes

- GPT-5 Pro only sees the data passed to the MCP tool (diffs + truncated test logs). Scrub sensitive output upstream if needed.
- Guardrails stay enforced: allowlisted paths, `max_files`, and the LOC budget are respected because Claude still rewrites full files.
- If tests continue failing after `AEGIS_MAX_TEST_ATTEMPTS`, the run aborts with the last tester output to prompt manual investigation.
