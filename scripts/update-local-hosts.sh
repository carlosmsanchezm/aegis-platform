#!/usr/bin/env bash
set -euo pipefail

HOSTS_FILE=${HOSTS_FILE:-/etc/hosts}
BLOCK_START="# Aegis Local Hosts"
BLOCK_END="# End Aegis Local Hosts"
ENTRIES="127.0.0.1 platform-api.localtest.me platform-api-grpc.localtest.me proxy.localtest.me"

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

current="$tmp_dir/current"
filtered="$tmp_dir/filtered"

sudo cat "$HOSTS_FILE" > "$current"

# Strip any existing block
sudo awk -v start="$BLOCK_START" -v end="$BLOCK_END" '
  $0==start {skip=1; next}
  $0==end {skip=0; next}
  skip==1 {next}
  {print}
' "$current" > "$filtered"

{
  echo "$BLOCK_START"
  echo "$ENTRIES"
  echo "$BLOCK_END"
} | sudo tee -a "$filtered" >/dev/null

sudo mv "$filtered" "$HOSTS_FILE"
sudo chmod 644 "$HOSTS_FILE"

echo "✓ Updated $HOSTS_FILE with Aegis local hosts"
