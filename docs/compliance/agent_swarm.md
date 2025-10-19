# Aegis Compliance Swarm

This walkthrough implements the AutoGen-based Analyzer → Implementer → Reviewer → Tester
swarm that wires into the official MCP servers and the Aegis repository.

## Authoritative OSCAL content

Fetch the NIST and FedRAMP artifacts referenced by the MCP servers:

```bash
scripts/fetch_oscal_content.sh
```

By default the Python FastMCP server reads from `third_party/oscal-content` and
`third_party/fedramp-automation`. Override the paths with
`OSCAL_CATALOG` and `FEDRAMP_BASELINES_DIR` if needed.

## MCP servers

| Server | Location | Command | Notes |
| ------ | -------- | ------- | ----- |
| GitHub | official | `docker run -i --rm -e GITHUB_PERSONAL_ACCESS_TOKEN ghcr.io/github/github-mcp-server` | Requires fine-grained PAT with PR scope. |
| Filesystem | official | `mcp-server-filesystem <repo> third_party/oscal-content third_party/fedramp-automation` | Restrict directories to read/write. |
| NIST-OSCAL | `servers/nist_oscal/nist_oscal_server.py` | `uv run servers/nist_oscal/nist_oscal_server.py` | Answers `get_control` and `in_fedramp_baseline`. |

Use `npx @modelcontextprotocol/inspector` to list tools and smoke-test the
servers (`get_control("AC-17")`, `in_fedramp_baseline("ac-17", "high")`).

## AutoGen harness

`agents/swarm/autogen_swarm.py` wires AutoGen GroupChat agents to the MCP
servers. Customize credentials through environment variables:

* `AEGIS_NIST_MCP_CMD`
* `AEGIS_FS_MCP_CMD`
* `AEGIS_GITHUB_MCP_CMD`

Example usage:

```python
from agents.swarm.autogen_swarm import build_swarm

llm = {"model": "gpt-4.1", "api_key": os.environ["OPENAI_API_KEY"]}
swarm = build_swarm(llm)
mission = "Close AC-17 gaps by enforcing mTLS in services/proxy"
swarm.operator.initiate_chat(swarm.manager, message=mission)
```

## Evidence trail

Swarm agents write signed JSON evidence to `.aegis/evidence/`:

```
.aegis/evidence/
  2025-01-12T14-31Z_ac-17_analysis.json
  2025-01-12T14-35Z_diff_summary.json
  2025-01-12T14-40Z_ci_results.json
  2025-01-12T14-45Z_component-snippet.json
  SIGNATURE.sig
```

Align evidence with OSCAL component/assessment schemas for ingestion into
SSP/SAP/SAR/POA&M packages.

## DoD IL-5 overlay

`policy/overlays/dod_il5.yaml` layers DoD Cloud Computing SRG requirements on
top of FedRAMP High (e.g., AC-17 session controls, SC-13 FIPS enforcement,
CM-6 SRG baselines). Expand the map with additional SRG/STIG requirements as
materially interpreted by your ISSM.

## CI policy gates

`.github/workflows/compliance.yml` adds four jobs to run Semgrep, Checkov,
Conftest, and Trivy on PRs. The Trivy job uploads SARIF results to GitHub Code
Scanning so reviewers can triage findings inline.

Pin action SHAs before promoting to production and store scan outputs under
`.aegis/evidence/` for long-term compliance records.
