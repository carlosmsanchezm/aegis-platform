#!/usr/bin/env bash
# aegis-setup-vscode.sh — Automates installation of the Aegis VS Code Remote extension.
#
# Idempotent, works on macOS and Linux, prints all actions for auditability.
# Supports online (download from platform API) and offline (local VSIX) modes.
#
# Usage:
#   bash aegis-setup-vscode.sh --url https://platform.aegis-platform.tech
#   bash aegis-setup-vscode.sh --vsix /path/to/aegis-remote.vsix
#   bash aegis-setup-vscode.sh --help

set -euo pipefail

# ---------------------------------------------------------------------------
# Color helpers (degrade gracefully when not connected to a TTY)
# ---------------------------------------------------------------------------
if [ -t 1 ]; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  BLUE='\033[0;34m'
  BOLD='\033[1m'
  RESET='\033[0m'
else
  RED='' GREEN='' YELLOW='' BLUE='' BOLD='' RESET=''
fi

info()    { printf "${BLUE}[INFO]${RESET}  %s\n" "$*"; }
ok()      { printf "${GREEN}[OK]${RESET}    %s\n" "$*"; }
warn()    { printf "${YELLOW}[WARN]${RESET}  %s\n" "$*"; }
err()     { printf "${RED}[ERROR]${RESET} %s\n" "$*" >&2; }
die()     { err "$@"; exit 1; }
step()    { printf "\n${BOLD}==> %s${RESET}\n" "$*"; }

# ---------------------------------------------------------------------------
# Cleanup trap — remove temp files on any exit
# ---------------------------------------------------------------------------
TMPFILES=()
cleanup() {
  for f in "${TMPFILES[@]+"${TMPFILES[@]}"}"; do
    if [ -f "$f" ]; then
      rm -f "$f" && info "Cleaned up temp file: $f"
    elif [ -d "$f" ]; then
      rm -rf "$f" && info "Cleaned up temp dir: $f"
    fi
  done
}
trap cleanup EXIT

# ---------------------------------------------------------------------------
# Usage
# ---------------------------------------------------------------------------
usage() {
  cat <<EOF
${BOLD}Aegis VS Code Extension Setup${RESET}

Usage:
  $(basename "$0") [OPTIONS]

Options:
  --url URL       Platform API base URL (e.g. https://platform.aegis-platform.tech).
                  Falls back to \$AEGIS_PLATFORM_URL env var.
  --vsix PATH     Local path to VSIX file (offline mode — skips download and checksum).
  --insiders      Force use of VS Code Insiders.
  --help          Show this help message.

Examples:
  # Online — download from platform API
  $(basename "$0") --url https://platform.aegis-platform.tech

  # Offline — use a local VSIX file
  $(basename "$0") --vsix ./aegis-remote-1.0.0.vsix
EOF
  exit 0
}

# ---------------------------------------------------------------------------
# Parse arguments
# ---------------------------------------------------------------------------
PLATFORM_URL=""
VSIX_PATH=""
FORCE_INSIDERS=false

parse_args() {
  while [ $# -gt 0 ]; do
    case "$1" in
      --url)
        [ -z "${2:-}" ] && die "--url requires a value"
        PLATFORM_URL="$2"; shift 2 ;;
      --vsix)
        [ -z "${2:-}" ] && die "--vsix requires a value"
        VSIX_PATH="$2"; shift 2 ;;
      --insiders)
        FORCE_INSIDERS=true; shift ;;
      --help|-h)
        usage ;;
      *)
        die "Unknown option: $1  (use --help for usage)" ;;
    esac
  done

  # Fall back to env var for URL
  if [ -z "$PLATFORM_URL" ]; then
    PLATFORM_URL="${AEGIS_PLATFORM_URL:-}"
  fi

  # Validate: in online mode we need a URL
  if [ -z "$VSIX_PATH" ] && [ -z "$PLATFORM_URL" ]; then
    die "Either --url / \$AEGIS_PLATFORM_URL or --vsix must be provided. Use --help for usage."
  fi

  # Validate: offline VSIX must exist
  if [ -n "$VSIX_PATH" ] && [ ! -f "$VSIX_PATH" ]; then
    die "VSIX file not found: $VSIX_PATH"
  fi
}

# ---------------------------------------------------------------------------
# Step 1 — Detect VS Code binary
# ---------------------------------------------------------------------------
CODE_BIN=""
CODE_VARIANT=""  # "stable" or "insiders"

detect_vscode() {
  step "Step 1: Detecting VS Code binary"

  local has_stable=false has_insiders=false

  if command -v code >/dev/null 2>&1; then
    has_stable=true
    info "Found: code (stable)"
  fi
  if command -v code-insiders >/dev/null 2>&1; then
    has_insiders=true
    info "Found: code-insiders"
  fi

  if $FORCE_INSIDERS; then
    if $has_insiders; then
      CODE_BIN="code-insiders"
      CODE_VARIANT="insiders"
      info "Using VS Code Insiders (--insiders flag)"
    else
      die "VS Code Insiders requested (--insiders) but 'code-insiders' not found in PATH."
    fi
  elif $has_stable; then
    CODE_BIN="code"
    CODE_VARIANT="stable"
    info "Using VS Code (stable)"
  elif $has_insiders; then
    CODE_BIN="code-insiders"
    CODE_VARIANT="insiders"
    warn "VS Code stable not found; falling back to VS Code Insiders."
  else
    die "Neither 'code' nor 'code-insiders' found in PATH. Install VS Code first."
  fi

  ok "VS Code binary: $CODE_BIN"
}

# ---------------------------------------------------------------------------
# Step 2 — Download VSIX (online mode only)
# ---------------------------------------------------------------------------
download_vsix() {
  if [ -n "$VSIX_PATH" ]; then
    step "Step 2: Using local VSIX (offline mode)"
    info "VSIX path: $VSIX_PATH"
    ok "Skipping download and checksum verification."
    return
  fi

  step "Step 2: Downloading VSIX from platform API"

  # Strip trailing slash from URL
  PLATFORM_URL="${PLATFORM_URL%/}"

  local metadata_url="${PLATFORM_URL}/api/v1/extension/metadata"
  local vsix_url="${PLATFORM_URL}/api/v1/extension/vsix"

  # --- Fetch metadata ---
  info "Fetching extension metadata from: $metadata_url"
  local metadata_file
  metadata_file="$(mktemp)"
  TMPFILES+=("$metadata_file")

  if ! curl -fsSL --retry 3 --retry-delay 2 -o "$metadata_file" "$metadata_url"; then
    die "Failed to fetch extension metadata from $metadata_url"
  fi
  info "Metadata downloaded to: $metadata_file"

  # --- Parse expected checksum from metadata ---
  local expected_sha256=""
  if command -v python3 >/dev/null 2>&1; then
    expected_sha256="$(python3 -c "
import json, sys
with open('$metadata_file') as f:
    data = json.load(f)
print(data.get('sha256', data.get('checksum', '')))
" 2>/dev/null || true)"
  elif command -v jq >/dev/null 2>&1; then
    expected_sha256="$(jq -r '.sha256 // .checksum // ""' "$metadata_file" 2>/dev/null || true)"
  else
    warn "Neither python3 nor jq available — cannot parse metadata for checksum."
  fi

  if [ -z "$expected_sha256" ]; then
    die "Could not extract SHA-256 checksum from metadata."
  fi
  info "Expected SHA-256: $expected_sha256"

  # --- Download VSIX ---
  local vsix_file
  local tmpdir
  tmpdir="$(mktemp -d)"
  vsix_file="${tmpdir}/aegis-remote.vsix"
  TMPFILES+=("$vsix_file" "$tmpdir")

  info "Downloading VSIX from: $vsix_url"
  if ! curl -fsSL --retry 3 --retry-delay 2 -o "$vsix_file" "$vsix_url"; then
    die "Failed to download VSIX from $vsix_url"
  fi
  info "VSIX downloaded to: $vsix_file"

  # --- Compute actual checksum ---
  local actual_sha256=""
  case "$(uname -s)" in
    Darwin)
      actual_sha256="$(shasum -a 256 "$vsix_file" | awk '{print $1}')"
      ;;
    Linux)
      actual_sha256="$(sha256sum "$vsix_file" | awk '{print $1}')"
      ;;
    *)
      die "Unsupported OS for checksum: $(uname -s)"
      ;;
  esac
  info "Actual SHA-256:   $actual_sha256"

  # --- Compare checksums ---
  if [ "$expected_sha256" != "$actual_sha256" ]; then
    err "Checksum mismatch!"
    err "  Expected: $expected_sha256"
    err "  Actual:   $actual_sha256"
    die "VSIX integrity check failed. Aborting installation."
  fi
  ok "Checksum verified."

  VSIX_PATH="$vsix_file"
}

# ---------------------------------------------------------------------------
# Step 3 — Install extension
# ---------------------------------------------------------------------------
install_extension() {
  step "Step 3: Installing extension"

  info "Running: $CODE_BIN --install-extension $VSIX_PATH --force"
  if ! "$CODE_BIN" --install-extension "$VSIX_PATH" --force; then
    die "Extension installation failed."
  fi

  ok "Extension installed successfully."
}

# ---------------------------------------------------------------------------
# Step 4 — Configure argv.json (enable-proposed-api)
# ---------------------------------------------------------------------------
EXTENSION_ID="aegis.aegis-remote"

configure_argv_json() {
  step "Step 4: Configuring argv.json (enable-proposed-api)"

  # Determine argv.json path based on OS and variant
  local argv_path=""
  local os_name
  os_name="$(uname -s)"

  case "$os_name" in
    Darwin)
      if [ "$CODE_VARIANT" = "insiders" ]; then
        argv_path="$HOME/Library/Application Support/Code - Insiders/argv.json"
      else
        argv_path="$HOME/Library/Application Support/Code/argv.json"
      fi
      ;;
    Linux)
      if [ "$CODE_VARIANT" = "insiders" ]; then
        argv_path="$HOME/.vscode-insiders/argv.json"
      else
        argv_path="$HOME/.vscode/argv.json"
      fi
      ;;
    *)
      warn "Unsupported OS ($os_name) — cannot determine argv.json path."
      warn "Please manually add \"$EXTENSION_ID\" to the enable-proposed-api array in your VS Code argv.json."
      return
      ;;
  esac

  info "argv.json path: $argv_path"

  # --- File does not exist — create it ---
  if [ ! -f "$argv_path" ]; then
    info "argv.json does not exist. Creating it."
    mkdir -p "$(dirname "$argv_path")"
    printf '%s\n' "{\"enable-proposed-api\": [\"$EXTENSION_ID\"]}" > "$argv_path"
    ok "Created argv.json with enable-proposed-api for $EXTENSION_ID."
    return
  fi

  # --- File exists — check if extension is already listed ---
  info "argv.json exists. Checking for $EXTENSION_ID in enable-proposed-api."

  # Try python3 first, then jq, then print manual instructions
  if command -v python3 >/dev/null 2>&1; then
    info "Using python3 for JSON manipulation."
    local result
    result="$(python3 -c "
import json, sys, os

argv_path = sys.argv[1]
ext_id = sys.argv[2]

with open(argv_path, 'r') as f:
    # Handle JSON with comments (VS Code allows them) — strip single-line comments
    lines = f.readlines()
    cleaned = []
    for line in lines:
        stripped = line.lstrip()
        if stripped.startswith('//'):
            continue
        cleaned.append(line)
    content = ''.join(cleaned)

try:
    data = json.loads(content)
except json.JSONDecodeError:
    # If JSON is still invalid, try removing trailing commas
    import re
    content = re.sub(r',\s*([}\]])', r'\1', content)
    data = json.loads(content)

if not isinstance(data, dict):
    data = {}

api_list = data.get('enable-proposed-api', [])
if not isinstance(api_list, list):
    api_list = []

if ext_id in api_list:
    print('ALREADY_PRESENT')
else:
    api_list.append(ext_id)
    data['enable-proposed-api'] = api_list
    with open(argv_path, 'w') as f:
        json.dump(data, f, indent=2)
        f.write('\n')
    print('ADDED')
" "$argv_path" "$EXTENSION_ID" 2>&1)" || true

    case "$result" in
      ALREADY_PRESENT)
        ok "$EXTENSION_ID is already in enable-proposed-api. No changes needed." ;;
      ADDED)
        ok "Added $EXTENSION_ID to enable-proposed-api in argv.json." ;;
      *)
        warn "python3 JSON manipulation returned unexpected result: $result"
        warn "Please manually ensure \"$EXTENSION_ID\" is in the enable-proposed-api array in: $argv_path"
        ;;
    esac

  elif command -v jq >/dev/null 2>&1; then
    info "Using jq for JSON manipulation."

    # Check if key already present
    local already_present
    already_present="$(jq -r \
      --arg ext "$EXTENSION_ID" \
      'if (.["enable-proposed-api"] // []) | index($ext) then "yes" else "no" end' \
      "$argv_path" 2>/dev/null || echo "error")"

    case "$already_present" in
      yes)
        ok "$EXTENSION_ID is already in enable-proposed-api. No changes needed." ;;
      no)
        local tmp_argv
        tmp_argv="$(mktemp)"
        TMPFILES+=("$tmp_argv")
        jq --arg ext "$EXTENSION_ID" \
          '.["enable-proposed-api"] = ((.["enable-proposed-api"] // []) + [$ext])' \
          "$argv_path" > "$tmp_argv"
        mv "$tmp_argv" "$argv_path"
        ok "Added $EXTENSION_ID to enable-proposed-api in argv.json."
        ;;
      *)
        warn "jq could not parse argv.json (it may contain comments)."
        warn "Please manually add \"$EXTENSION_ID\" to the enable-proposed-api array in: $argv_path"
        ;;
    esac

  else
    warn "Neither python3 nor jq available."
    warn "Please manually edit: $argv_path"
    warn "Ensure the file contains:"
    warn "  {\"enable-proposed-api\": [\"$EXTENSION_ID\"]}"
  fi
}

# ---------------------------------------------------------------------------
# Step 5 — Verify installation
# ---------------------------------------------------------------------------
verify_installation() {
  step "Step 5: Verifying installation"

  info "Running: $CODE_BIN --list-extensions"
  local extensions
  extensions="$("$CODE_BIN" --list-extensions 2>/dev/null || true)"

  if echo "$extensions" | grep -qi "aegis.aegis-remote"; then
    ok "Extension '$EXTENSION_ID' is installed."
  else
    warn "Extension '$EXTENSION_ID' not found in --list-extensions output."
    warn "This may be expected if VS Code needs a restart. Listed extensions:"
    echo "$extensions" | sed 's/^/    /'
  fi
}

# ---------------------------------------------------------------------------
# Step 6 — Print final result
# ---------------------------------------------------------------------------
print_result() {
  step "Setup complete"
  printf "\n"
  ok "Aegis VS Code Remote extension has been installed and configured."
  printf "\n"
  info "${BOLD}Restart VS Code to complete setup.${RESET}"
  printf "\n"
}

# ---------------------------------------------------------------------------
# Banner
# ---------------------------------------------------------------------------
print_banner() {
  printf "\n"
  printf "${BOLD}╔══════════════════════════════════════╗${RESET}\n"
  printf "${BOLD}║   Aegis VS Code Extension Setup      ║${RESET}\n"
  printf "${BOLD}╚══════════════════════════════════════╝${RESET}\n"
  printf "\n"
  info "Date: $(date -u '+%Y-%m-%dT%H:%M:%SZ')"
  info "Host: $(hostname)"
  info "OS:   $(uname -s) $(uname -m)"
  printf "\n"
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
main() {
  parse_args "$@"
  print_banner
  detect_vscode
  download_vsix
  install_extension
  configure_argv_json
  verify_installation
  print_result
}

main "$@"
