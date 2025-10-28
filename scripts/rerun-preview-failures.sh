#!/usr/bin/env bash
# Rerun only the failed jobs from the latest preview workflow run (or a specific run).

set -euo pipefail

WORKFLOW_FILE="${WORKFLOW_FILE:-preview-deployment.yml}"
RUN_ID="${RUN_ID:-}"
BRANCH="${BRANCH:-}"
WATCH="${WATCH:-0}"
SUITES="${SUITES:-platform-api}"
RUN_SUBSET="${RUN_SUBSET:-tests-only}"

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
  RUN_SUBSET      Workflow input run_subset when dispatching new runs (default: tests-only)
  SUITES          Workflow input suites when dispatching new runs (default: platform-api)
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

current_sha="$(git rev-parse HEAD)"
run_json="$(gh run view "${RUN_ID}" --json status,conclusion,headSha,headBranch,event)"
run_head_sha="$(echo "${run_json}" | jq -r '.headSha')"
run_branch="$(echo "${run_json}" | jq -r '.headBranch')"
pr_number="$(gh pr list --state all --head \"${run_branch}\" --limit 1 --json number --jq '.[0].number' 2>/dev/null || echo \"\")"

if [[ -z "${run_head_sha}" || "${run_head_sha}" == "null" ]]; then
  echo "Unable to determine head SHA for run ${RUN_ID}" >&2
  exit 1
fi

if [[ "${run_head_sha}" == "${current_sha}" ]]; then
  echo "Rerunning failed jobs for workflow run ${RUN_ID} (same commit)..." >&2
  gh run rerun "${RUN_ID}" --failed

  if [[ "${WATCH}" == "1" ]]; then
    echo "Watching rerun ${RUN_ID}..." >&2
    gh run watch "${RUN_ID}" --exit-status
  fi
  exit 0
fi

# Workflow definition changed; dispatch a targeted rerun using the latest workflow file.
if [[ -z "${pr_number}" ]]; then
  echo "Run ${RUN_ID} is not associated with a pull request; cannot compute preview number for dispatch." >&2
  exit 1
fi

echo "Workflow definition changed since run ${RUN_ID}; dispatching ${RUN_SUBSET} run for preview ${pr_number} on branch ${run_branch}..." >&2
gh workflow run "${WORKFLOW_FILE}" \
  --ref "${run_branch}" \
  -f run_subset="${RUN_SUBSET}" \
  -f suites="${SUITES}" \
  -f preview_number="${pr_number}"

# Give GitHub a moment to register the new run, then pick it up for watching.
sleep 10
new_run_id="$(gh run list \
  --workflow "${WORKFLOW_FILE}" \
  --branch "${run_branch}" \
  --limit 10 \
  --json databaseId,status,headSha,createdAt \
  --jq 'map(select(.status != "completed")) | sort_by(.createdAt) | last | .databaseId')" || true

if [[ -z "${new_run_id}" || "${new_run_id}" == "null" ]]; then
  echo "Unable to identify the dispatched run; check GitHub Actions manually." >&2
  exit 1
fi

echo "Dispatched run ${new_run_id} for preview ${pr_number}." >&2
if [[ "${WATCH}" == "1" ]]; then
  echo "Watching run ${new_run_id}..." >&2
  gh run watch "${new_run_id}" --exit-status
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
