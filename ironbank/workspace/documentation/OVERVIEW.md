# Aegis Workspace — Architecture Overview

## Purpose

The Aegis Workspace image provides interactive GPU development environments for data scientists and ML engineers working in secure DoD environments. Each workspace is a Kubernetes pod with dedicated GPU resources.

## Init Container Pattern

The VS Code Remote Extension Host (REH) binary is NOT baked into the workspace image. Instead, it is injected at pod startup via the `vscode-reh-init` init container:

1. `vscode-reh-init` container starts, copies pre-downloaded VS Code server to shared emptyDir volume
2. Workspace container starts, finds VS Code server at `/reh/bin/current/`
3. `entrypoint.sh` validates the VS Code binary exists and is executable
4. `start-reh.sh` launches the VS Code server on the configured port

This pattern allows independent versioning of the workspace image and VS Code server.

## Process Management

Tini (v0.19.0) is used as PID 1 to handle signal forwarding and zombie process reaping. The entrypoint chain is:

```
tini → entrypoint.sh → start-reh.sh → code-server
```

## Security

- Runs as non-root (UID 1000)
- SUID/SGID bits stripped from all binaries
- No runtime downloads — all software pre-installed or injected via init container
- NVIDIA GPU access controlled via Kubernetes device plugin
