# Aegis Agent Swarm - Complete File Structure & Organization

## Core Agent Swarm Files

### 1. **Agent Implementations**

#### `/agents/strands/` - Strands-based Agent Framework
- **`run_team_ac17.py`** - Main AC-17 compliance team orchestrator
  - Uses Strands framework with OpenAI models
  - Orchestrates multi-agent workflow for AC-17 (Remote Access) compliance
  - Integrates MCP (Model Context Protocol) clients
  - Generates evidence and opens PRs

#### `/agents/claude/` - Claude-based Agent Implementation
- **`run_fedramp_swarm.py`** - FedRAMP compliance swarm using Claude
  - Uses Claude Agent SDK
  - Implements tool-calling architecture
  - Evidence generation and repo operations

### 2. **Agent Tools & Utilities**

#### `/agents/strands/tools/`
- **`composite_github.py`** - GitHub operations (PR creation, file commits)
- **`evidence.py`** - Evidence file generation and management
- **`__init__.py`** - Package initialization

### 3. **OSCAL Data Sources**

#### `/docs/oscal/` - NIST & FedRAMP Control Catalogs
- **`nist_sp_800_53_rev5_catalog.json`** (5MB) - Full NIST SP 800-53 Rev 5 catalog
- **`nist_sp_800_53b_rev5_moderate_baseline.json`** (1.8MB) - Moderate baseline
- **`nist_sp_800_171_rev3_catalog.json`** (524KB) - CUI requirements (DoD IL-5)
- **`fedramp_rev5_moderate_baseline.json`** (5MB) - FedRAMP Moderate baseline
- **`README.md`** - Documentation on updating/adding catalogs

#### `/third_party/` - External OSCAL Content
- **`oscal-content/`** - NIST official OSCAL content repository
- **`fedramp-automation/`** - GSA FedRAMP automation artifacts

### 4. **MCP Servers**

#### `/servers/nist_oscal/`
- **`nist_oscal_server.py`** - Custom MCP server for OSCAL operations
  - Tools: `get_control()`, `in_fedramp_baseline()`
  - Parses OSCAL JSON catalogs
  - Provides compliance control lookups

### 5. **Plans & Mission Files**

#### `/plans/`
- **`sc10.json`** - SC-10 (Session Termination) compliance plan
  - Defines baseline, controls, file hints
  - Expert guidance for implementation
  - PR configuration (branch, title, base)

### 6. **Evidence & Output**

#### `/.aegis/evidence/` - Compliance Evidence Trail
- Timestamped JSON evidence files
- Format: `{timestamp}_{control-id}_{type}.json`
- Examples:
  - `2025-01-12T14-31Z_ac-17_analysis.json`
  - `2025-01-12T14-35Z_diff_summary.json`
  - `2025-01-12T14-40Z_ci_results.json`
- **`SIGNATURE.sig`** - Digital signatures for evidence

### 7. **Documentation**

#### `/docs/compliance/`
- **`agent_swarm.md`** - Complete swarm documentation
  - MCP server configuration
  - AutoGen harness usage
  - Evidence trail format
  - DoD IL-5 overlay info
  - CI policy gates

---

## Directory Structure

```
aegis/
├── agents/
│   ├── claude/                    # Claude SDK-based agents
│   │   ├── run_fedramp_swarm.py  # FedRAMP compliance swarm
│   │   └── requirements.txt
│   │
│   ├── strands/                   # Strands framework agents
│   │   ├── run_team_ac17.py      # AC-17 compliance team
│   │   ├── tools/
│   │   │   ├── composite_github.py  # GitHub PR tools
│   │   │   ├── evidence.py          # Evidence generation
│   │   │   └── __init__.py
│   │   ├── requirements.txt
│   │   └── __init__.py
│   │
│   └── k8s-agent/                # Kubernetes operator (not swarm-related)
│
├── servers/
│   └── nist_oscal/
│       └── nist_oscal_server.py  # Custom OSCAL MCP server
│
├── docs/
│   ├── oscal/                     # OSCAL data sources
│   │   ├── nist_sp_800_53_rev5_catalog.json
│   │   ├── nist_sp_800_53b_rev5_moderate_baseline.json
│   │   ├── nist_sp_800_171_rev3_catalog.json
│   │   ├── fedramp_rev5_moderate_baseline.json
│   │   └── README.md
│   │
│   └── compliance/
│       ├── agent_swarm.md         # Main swarm documentation
│       └── AGENT_SWARM_STRUCTURE.md  # This file
│
├── plans/
│   └── sc10.json                  # Control implementation plans
│
├── third_party/
│   ├── oscal-content/             # NIST OSCAL repository
│   └── fedramp-automation/        # FedRAMP automation artifacts
│
└── .aegis/
    └── evidence/                  # Compliance evidence output
        ├── {timestamp}_{control}_*.json
        └── SIGNATURE.sig
```

---

## Component Breakdown

### Agent Frameworks Used

1. **Strands** (`agents/strands/`)
   - Multi-agent orchestration
   - MCP client integration
   - OpenAI model integration
   - Used for: AC-17 compliance workflow

2. **Claude Agent SDK** (`agents/claude/`)
   - Tool-calling architecture
   - Evidence generation
   - Repo operations
   - Used for: FedRAMP compliance swarm

### MCP (Model Context Protocol) Servers

| Server | Location | Command | Purpose |
|--------|----------|---------|---------|
| **GitHub** | Official | `docker run ghcr.io/github/github-mcp-server` | PR creation, file operations |
| **Filesystem** | Official | `mcp-server-filesystem <paths>` | Read/write repo files |
| **NIST-OSCAL** | Custom | `uv run servers/nist_oscal/nist_oscal_server.py` | OSCAL control lookups |

### Plan File Format (`plans/*.json`)

```json
{
  "baseline": "moderate|high|low",
  "controls": ["SC-10", "AC-17"],
  "file_hints": ["path/to/file.go"],
  "allow_paths": ["services/**"],
  "max_files": 6,
  "pr": {
    "branch_prefix": "frp-mod",
    "title": "FedRAMP: implement SC-10",
    "base": "main"
  },
  "expert_notes": "Implementation guidance..."
}
```

---

## Data Flow

```
┌─────────────────┐
│  plans/*.json   │  Mission definition
└────────┬────────┘
         │
         v
┌─────────────────────────────────────────┐
│  Agent Swarm (strands/claude)           │
│  - Reads mission from plan              │
│  - Queries OSCAL via MCP server         │
│  - Reads/writes files via MCP           │
└────────┬────────────────────────────────┘
         │
         v
┌──────────────────────────────────────────┐
│  MCP Servers                             │
│  1. nist_oscal_server.py                 │
│     - get_control("AC-17")               │
│     - in_fedramp_baseline("ac-17", "mod")│
│                                          │
│  2. filesystem (official)                │
│     - read_file(), write_file()          │
│                                          │
│  3. github (official)                    │
│     - create_pr(), commit_files()        │
└────────┬─────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────┐
│  Evidence Output                         │
│  .aegis/evidence/                        │
│  - {timestamp}_{control}_analysis.json   │
│  - {timestamp}_diff_summary.json         │
│  - SIGNATURE.sig                         │
└──────────────────────────────────────────┘
         │
         v
┌─────────────────────────────────────────┐
│  GitHub PR                               │
│  - Branch: {prefix}-{control}-{timestamp}│
│  - Files: Implementation changes         │
│  - Body: Evidence references             │
└──────────────────────────────────────────┘
```

---

## Proposed Improvements

### 🎯 Organizational Enhancements

#### 1. **Consolidate Agent Implementations**
**Current State**: Two separate agent directories (`agents/strands/`, `agents/claude/`)

**Proposed**:
```
agents/
├── swarm/                        # Main swarm orchestration
│   ├── frameworks/
│   │   ├── strands/             # Strands-based agents
│   │   └── claude/              # Claude SDK agents
│   ├── orchestrator.py          # Unified orchestration layer
│   └── config.py                # Swarm configuration
│
├── tools/                        # Shared tools (moved from strands/tools)
│   ├── github.py
│   ├── evidence.py
│   └── mcp_clients.py
│
└── plans/                        # Move plans here for locality
    ├── sc10.json
    └── ac17.json
```

**Benefits**:
- Clearer separation between framework implementations
- Shared tooling reduces duplication
- Plans co-located with agents that use them

#### 2. **Standardize Evidence Structure**
**Current**: `.aegis/evidence/` with mixed file naming

**Proposed**:
```
.aegis/
├── evidence/
│   ├── controls/                 # Organized by control
│   │   ├── ac-17/
│   │   │   ├── 2025-01-12_analysis.json
│   │   │   ├── 2025-01-12_implementation.json
│   │   │   └── metadata.json
│   │   └── sc-10/
│   │       └── ...
│   ├── runs/                     # Organized by swarm run
│   │   └── 2025-01-12T14-30Z/
│   │       ├── manifest.json
│   │       ├── ac-17_evidence.json
│   │       └── SIGNATURE.sig
│   └── archive/                  # Historical evidence
└── reports/                      # Human-readable reports
    └── 2025-01-12_compliance_summary.md
```

#### 3. **Create Unified Configuration**
**New File**: `agents/swarm/config.yaml`

```yaml
# Swarm configuration
frameworks:
  strands:
    model: gpt-4
    temperature: 0.7
  claude:
    model: claude-sonnet-4
    max_tokens: 8000

mcp_servers:
  nist_oscal:
    command: uv run servers/nist_oscal/nist_oscal_server.py
    timeout: 30
  github:
    command: docker run -i ghcr.io/github/github-mcp-server
    env:
      GITHUB_TOKEN: ${GITHUB_TOKEN}
  filesystem:
    command: mcp-server-filesystem
    args:
      - /repo/path
      - /oscal/data/path

evidence:
  output_dir: .aegis/evidence
  signing: true
  formats: [json, oscal]

compliance:
  baselines: [moderate, high]
  default_baseline: moderate
```

#### 4. **Add Plan Schema & Validation**
**New File**: `agents/plans/schema.json`

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["baseline", "controls", "pr"],
  "properties": {
    "baseline": {"enum": ["low", "moderate", "high"]},
    "controls": {"type": "array", "items": {"type": "string"}},
    "file_hints": {"type": "array", "items": {"type": "string"}},
    "allow_paths": {"type": "array", "items": {"type": "string"}},
    "max_files": {"type": "integer", "minimum": 1},
    "pr": {
      "type": "object",
      "required": ["branch_prefix", "title", "base"],
      "properties": {
        "branch_prefix": {"type": "string"},
        "title": {"type": "string"},
        "base": {"type": "string"}
      }
    },
    "expert_notes": {"type": "string"}
  }
}
```

**Validator**: `agents/plans/validate.py`

#### 5. **Better Tool Organization**
**Current**: Tools scattered across agent implementations

**Proposed**: Shared tool library
```
agents/tools/
├── mcp/
│   ├── __init__.py
│   ├── nist.py              # NIST OSCAL operations
│   ├── github.py            # GitHub operations
│   └── filesystem.py        # File operations
│
├── evidence/
│   ├── __init__.py
│   ├── generator.py         # Evidence generation
│   ├── signer.py            # Digital signatures
│   └── validator.py         # Evidence validation
│
├── compliance/
│   ├── __init__.py
│   ├── control_parser.py    # Parse OSCAL controls
│   ├── gap_analyzer.py      # Gap analysis
│   └── mapper.py            # Map controls to code
│
└── utils/
    ├── diff.py              # Code diff utilities
    └── pr_builder.py        # PR creation helpers
```

#### 6. **Documentation Improvements**

**New Structure**:
```
docs/compliance/
├── README.md                    # Overview & quick start
├── architecture/
│   ├── agent_swarm.md          # Current doc (moved here)
│   ├── data_flow.md            # Visual data flow diagrams
│   └── mcp_integration.md      # MCP server integration
│
├── guides/
│   ├── adding_agents.md        # How to add new agents
│   ├── creating_plans.md       # Plan file creation guide
│   ├── evidence_format.md      # Evidence specifications
│   └── testing.md              # Testing agent swarms
│
└── reference/
    ├── plan_schema.md          # Plan JSON schema reference
    ├── mcp_tools.md            # Available MCP tools
    └── file_structure.md       # This document
```

### 🔧 Implementation Recommendations

1. **Phase 1: Consolidation** (Week 1)
   - Move plans to `agents/plans/`
   - Create shared `agents/tools/` directory
   - Consolidate duplicate tooling

2. **Phase 2: Standardization** (Week 2)
   - Implement evidence structure
   - Add plan validation
   - Create unified config

3. **Phase 3: Documentation** (Week 3)
   - Reorganize docs
   - Add architecture diagrams
   - Create developer guides

4. **Phase 4: Testing** (Week 4)
   - Add integration tests
   - CI/CD for swarm validation
   - Evidence format validation

---

## Quick Reference

### Running Agent Swarms

#### Strands-based (AC-17):
```bash
cd agents/strands
uv run run_team_ac17.py
```

#### Claude-based (FedRAMP):
```bash
cd agents/claude
export GITHUB_TOKEN=xxx
uv run run_fedramp_swarm.py
```

### Updating OSCAL Data:
```bash
scripts/fetch_oscal_content.sh
```

### Validating Evidence:
```bash
python agents/tools/evidence/validator.py .aegis/evidence/
```

---

## Key Contacts & Resources

- **OSCAL Content**: https://github.com/usnistgov/oscal-content
- **FedRAMP Automation**: https://github.com/GSA/fedramp-automation
- **MCP Specification**: https://modelcontextprotocol.io
- **Strands Framework**: (internal)
- **Claude Agent SDK**: https://github.com/anthropics/anthropic-sdk-python
