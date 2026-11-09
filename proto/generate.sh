#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${ROOT}"

export PATH="${ROOT}/../bin:${PATH}"
export BUF_CACHE_DIR="${ROOT}/../.bufcache"

buf generate
