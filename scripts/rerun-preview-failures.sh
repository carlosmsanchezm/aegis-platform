#!/usr/bin/env bash
# Rerun only the failed jobs from the latest preview workflow run (or a specific run).

set -euo pipefail

WORKFLOW_FILE="${WORKFLOW_FILE:-preview-deployment.yml}"
RUN_ID="${RUN_ID:-}"
BRANCH="${BRANCH:-}"
WATCH="${WATCH:-0}"

usage() {
  cat <<'EOF'
Usage: scripts/rerun-preview-failures.sh [run-id|branch]

  With no arguments the script finds the most recent failed preview workflow
  run for the current branch and reruns only its failed jobs.

  You may pass either:
    * a numeric workflow run ID
    * a branch name (to override the current branch)

Environment variables:
  WORKFLOW_FILE   GitHub Actions workflow file name (default: preview-deployment.yml)
  RUN_ID          Explicit workflow run ID to rerun
  BRANCH          Branch to inspect for failed runs (default: current branch)
  WATCH           Set to 1 to watch the rerun to completion
EOF
}

if [[ "${1:-}" =~ ^(-h|--help)$ ]]; then
  usage
  exit 0
fi

if [[ $# -gt 1 ]]; then
  echo "Too many arguments" >&2
  usage
  exit 1
fi

if [[ $# -eq 1 ]]; then
  if [[ "${1}" =~ ^[0-9]+$ ]]; then
    RUN_ID="${1}"
  else
    BRANCH="${1}"
  fi
fi

if [[ -z "${RUN_ID}" ]]; then
  if [[ -z "${BRANCH}" ]]; then
    BRANCH="$(git rev-parse --abbrev-ref HEAD)"
  fi

  echo "Searching for failed runs on branch '${BRANCH}' in ${WORKFLOW_FILE}..." >&2
  RUN_ID="$(gh run list \
    --workflow "${WORKFLOW_FILE}" \
    --branch "${BRANCH}" \
    --limit 20 \
    --json databaseId,status,conclusion \
    --jq 'map(select(.status == "completed" and .conclusion != "success")) | first | .databaseId')" || true

  if [[ -z "${RUN_ID}" || "${RUN_ID}" == "null" ]]; then
    echo "No failed runs found for branch '${BRANCH}'." >&2
    exit 1
  fi
fi

echo "Rerunning failed jobs for workflow run ${RUN_ID}..." >&2
gh run rerun "${RUN_ID}" --failed

if [[ "${WATCH}" == "1" ]]; then
  echo "Watching rerun ${RUN_ID}..." >&2
  gh run watch "${RUN_ID}" --exit-status
fi
