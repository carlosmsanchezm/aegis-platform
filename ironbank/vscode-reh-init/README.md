# aegis-vscode-reh-init

Init container that injects VS Code Remote Extension Host into Aegis workspace pods.

## Description

This init container contains a pre-downloaded, pinned VS Code Remote Extension Host (REH) binary. When used as a Kubernetes init container, it copies the REH binary to a shared emptyDir volume before the workspace container starts.

## Base Image

`registry1.dso.mil/ironbank/redhat/ubi/ubi9-minimal:9.7`

## How It Works

1. Iron Bank pipeline pre-fetches the VS Code Server tarball declared in `hardening_manifest.yaml`
2. Dockerfile extracts the tarball to `/reh/bin/current/`
3. At pod startup, the init container copies `/reh/` to the shared volume at `/shared/reh/`
4. The workspace container reads the VS Code binary from the shared volume

## Kubernetes Usage

```yaml
initContainers:
  - name: vscode-reh-init
    image: registry1.dso.mil/ironbank/aegis/vscode-reh-init:1.0.0
    command: ["cp", "-r", "/reh/.", "/shared/reh/"]
    volumeMounts:
      - name: reh-volume
        mountPath: /shared/reh
```

## Resource Requirements

| Resource | Request | Limit |
|----------|---------|-------|
| CPU | 50m | 200m |
| Memory | 64Mi | 128Mi |

## External Resources (pre-fetched by Iron Bank pipeline)

- VS Code Server REH (pinned commit ce099c1ed25d9eb3076c11e4a280f3eb52b4fbeb)
