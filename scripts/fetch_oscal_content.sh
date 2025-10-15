#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
TARGET_DIR="${REPO_ROOT}/third_party"

NIST_SRC="https://github.com/usnistgov/oscal-content.git"
FEDRAMP_SRC="https://github.com/GSA/fedramp-automation.git"

mkdir -p "${TARGET_DIR}"

clone_or_update() {
  local url="$1"
  local dest="$2"

  if [[ -d "${dest}/.git" ]]; then
    echo "Updating ${dest}" >&2
    git -C "${dest}" pull --ff-only
  else
    echo "Cloning ${url} into ${dest}" >&2
    git clone --depth=1 "${url}" "${dest}"
  fi
}

clone_or_update "${NIST_SRC}" "${TARGET_DIR}/oscal-content"
clone_or_update "${FEDRAMP_SRC}" "${TARGET_DIR}/fedramp-automation"

cat <<'EOF'
Authoritative OSCAL datasets are now available under third_party/:
- oscal-content (NIST SP 800-53 Rev5, SP 800-171, etc.)
- fedramp-automation (FedRAMP baselines Rev5)

Update the environment variables below if you store the data elsewhere:
  export OSCAL_CATALOG=${REPO_ROOT}/third_party/oscal-content/nist.gov/SP800-53/rev5/json/sp800-53rev5-catalog.json
  export FEDRAMP_BASELINES_DIR=${REPO_ROOT}/third_party/fedramp-automation/dist/content/baselines/rev5/json
EOF
