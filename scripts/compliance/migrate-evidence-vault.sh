#!/bin/bash
#
# migrate-evidence-vault.sh
#
# Migrates compliance evidence from the main repository to the external evidence vault.
# This script moves evidence out of the product repo to maintain separation of concerns.
#
# Usage: ./scripts/compliance/migrate-evidence-vault.sh [target_vault_path]
#
# Default target: ~/code/aegis-compliance-evidence

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log_info() { echo -e "${GREEN}[INFO]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_step() { echo -e "${BLUE}[STEP]${NC} $1"; }

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SOURCE_DIR="$REPO_ROOT/compliance-evidence"
TARGET_DIR="${1:-$HOME/code/aegis-compliance-evidence}"

echo ""
echo "=============================================="
echo "   Compliance Evidence Vault Migration"
echo "=============================================="
echo ""
echo "Source: $SOURCE_DIR"
echo "Target: $TARGET_DIR"
echo ""

# Check if source exists
if [[ ! -d "$SOURCE_DIR" ]]; then
    log_warn "No evidence found in repository at: $SOURCE_DIR"
    log_info "Nothing to migrate. Creating empty vault structure instead."

    # Create the vault structure anyway
    log_step "Creating evidence vault structure..."
    mkdir -p "$TARGET_DIR"/{soc2,iso27001}
    mkdir -p "$TARGET_DIR/soc2/2025"/{retroactive,Q4,governance,risk-management,vendor-soc2-reports}
    mkdir -p "$TARGET_DIR/soc2/2026"/{Q1,Q2,Q3,Q4,governance,risk-management,vendor-soc2-reports}
    mkdir -p "$TARGET_DIR/iso27001/2026"/{internal-audits,management-reviews,risk-assessments,corrective-actions}

    # Create month directories for 2026
    for month in 01 02 03 04 05 06 07 08 09 10 11 12; do
        mkdir -p "$TARGET_DIR/soc2/2026/2026-$month"/{access-reviews,ci-cd-security,vuln-management,incident-response,logging-monitoring,change-management,vendor-management,backups-dr,training-policy-ack}
    done

    log_info "Empty vault created at: $TARGET_DIR"
    echo ""
    log_info "Next steps:"
    echo "  1. Run monthly evidence collection scripts"
    echo "  2. Add to shell: export EVIDENCE_VAULT=\"$TARGET_DIR\""
    exit 0
fi

# Count files to migrate
FILE_COUNT=$(find "$SOURCE_DIR" -type f ! -name ".DS_Store" | wc -l | tr -d ' ')
log_info "Found $FILE_COUNT files to migrate"

if [[ "$FILE_COUNT" -eq 0 ]]; then
    log_warn "No files to migrate (only .DS_Store files found)"
    exit 0
fi

# Confirm migration
echo ""
read -p "Proceed with migration? (y/n) " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    log_warn "Migration cancelled"
    exit 1
fi

# Create target directory
log_step "Creating target directory structure..."
mkdir -p "$TARGET_DIR"

# Move evidence
log_step "Moving evidence files..."
if [[ -d "$SOURCE_DIR/soc2" ]]; then
    # Create parent if needed
    mkdir -p "$TARGET_DIR/soc2"

    # Copy preserving structure
    cp -R "$SOURCE_DIR/soc2/"* "$TARGET_DIR/soc2/" 2>/dev/null || true
    log_info "Copied SOC 2 evidence"
fi

if [[ -d "$SOURCE_DIR/iso27001" ]]; then
    mkdir -p "$TARGET_DIR/iso27001"
    cp -R "$SOURCE_DIR/iso27001/"* "$TARGET_DIR/iso27001/" 2>/dev/null || true
    log_info "Copied ISO 27001 evidence"
fi

# Create additional structure that may not exist yet
log_step "Ensuring complete vault structure..."
mkdir -p "$TARGET_DIR/soc2/2025"/{retroactive,Q4,governance,risk-management,vendor-soc2-reports}
mkdir -p "$TARGET_DIR/soc2/2026"/{Q1,Q2,Q3,Q4,governance,risk-management,vendor-soc2-reports}
mkdir -p "$TARGET_DIR/iso27001/2026"/{internal-audits,management-reviews,risk-assessments,corrective-actions}

# Create month directories for 2026 if missing
for month in 01 02 03 04 05 06 07 08 09 10 11 12; do
    mkdir -p "$TARGET_DIR/soc2/2026/2026-$month"/{access-reviews,ci-cd-security,vuln-management,incident-response,logging-monitoring,change-management,vendor-management,backups-dr,training-policy-ack}
done

# Verify migration
MIGRATED_COUNT=$(find "$TARGET_DIR" -type f ! -name ".DS_Store" | wc -l | tr -d ' ')
log_info "Migrated $MIGRATED_COUNT files to vault"

# Create vault README
log_step "Creating vault README..."
cat > "$TARGET_DIR/README.md" << 'EOF'
# Aegis Compliance Evidence Vault

This directory contains sensitive compliance evidence for SOC 2 Type II and ISO 27001 certification.

## Structure

```
.
├── soc2/                    # SOC 2 Type II evidence
│   ├── 2025/
│   │   ├── retroactive/     # Historical evidence (Sept-Dec 2025)
│   │   └── 2025-MM/         # Monthly evidence
│   └── 2026/
│       ├── 2026-MM/         # Monthly evidence
│       ├── QX/              # Quarterly access reviews
│       ├── governance/      # Annual governance docs
│       └── vendor-soc2-reports/
│
└── iso27001/                # ISO 27001-specific evidence
    └── 2026/
        ├── internal-audits/
        ├── management-reviews/
        ├── risk-assessments/
        └── corrective-actions/
```

## Usage

Set environment variable for scripts:
```bash
export EVIDENCE_VAULT="$HOME/code/aegis-compliance-evidence"
```

## Security

- This vault contains sensitive information
- Do NOT commit to public repositories
- Consider encryption for highly sensitive data
- Retain per compliance requirements (typically 7 years)

## Collection

Run monthly evidence collection:
```bash
cd ~/code/aegis-platform-observability-integration
./scripts/compliance/export_github_security_baseline.sh "$EVIDENCE_VAULT/soc2/$(date +%Y)/$(date +%Y-%m)/access-reviews"
./scripts/compliance/export_aws_security_baseline.sh "$EVIDENCE_VAULT/soc2/$(date +%Y)/$(date +%Y-%m)/access-reviews"
```

## Related Documentation

- Evidence Map: `docs/compliance/soc2/evidence-map.md`
- ISMS Operating Plan: `docs/compliance/iso27001/10-isms-operating-plan.md`
- Control Matrix: `docs/compliance/soc2/control-matrix.csv`
EOF

log_info "Created vault README"

# Remove source directory from repo
log_step "Cleaning up repository..."
if [[ -d "$SOURCE_DIR" ]]; then
    rm -rf "$SOURCE_DIR"
    log_info "Removed evidence from repository"
fi

# Add to .gitignore if not already there
GITIGNORE="$REPO_ROOT/.gitignore"
if [[ -f "$GITIGNORE" ]]; then
    if ! grep -q "^compliance-evidence/" "$GITIGNORE" 2>/dev/null; then
        echo "" >> "$GITIGNORE"
        echo "# Compliance evidence (stored in separate vault)" >> "$GITIGNORE"
        echo "compliance-evidence/" >> "$GITIGNORE"
        log_info "Added compliance-evidence/ to .gitignore"
    else
        log_info ".gitignore already excludes compliance-evidence/"
    fi
else
    echo "# Compliance evidence (stored in separate vault)" > "$GITIGNORE"
    echo "compliance-evidence/" >> "$GITIGNORE"
    log_info "Created .gitignore with compliance-evidence/ exclusion"
fi

# Summary
echo ""
echo "=============================================="
echo "   Migration Complete"
echo "=============================================="
echo ""
echo "Evidence vault location: $TARGET_DIR"
echo "Files migrated: $MIGRATED_COUNT"
echo ""
log_info "Next steps:"
echo ""
echo "  1. Add to your shell profile (~/.zshrc or ~/.bashrc):"
echo "     export EVIDENCE_VAULT=\"$TARGET_DIR\""
echo ""
echo "  2. Optionally create a private git repo for the vault:"
echo "     cd $TARGET_DIR"
echo "     git init"
echo "     git add ."
echo "     git commit -m 'Initial evidence vault'"
echo "     gh repo create aegis-compliance-evidence --private --source=. --push"
echo ""
echo "  3. Run the monthly evidence collection per the ISMS Operating Plan"
echo ""
log_info "See docs/compliance/soc2/evidence-map.md for complete documentation"
