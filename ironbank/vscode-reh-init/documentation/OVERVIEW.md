# Aegis VS Code REH Init Container — Architecture Overview

## Purpose

The vscode-reh-init container implements the init container pattern for injecting the VS Code Remote Extension Host (REH) binary into workspace pods. This decouples VS Code server versioning from the workspace image.

## Pinned VS Code Version

The VS Code Server binary is pinned to a specific commit hash to ensure reproducible builds and deterministic behavior. The commit hash and sha256 are declared in `hardening_manifest.yaml`.

## Security

- Runs as non-root (UID 65534)
- No network access required
- Read-only operation (copies files to shared volume)
- No runtime downloads — binary is pre-fetched by Iron Bank pipeline
