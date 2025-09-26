# Aegis Workspace Image

This image extends `lscr.io/linuxserver/openssh-server` with the extra Alpine
packages Visual Studio Code Remote SSH expects (`libstdc++` and `libgcc`). Build
it locally before submitting interactive workspaces so that new pods have the
runtime VS Code needs without any manual `apk add` patching.

```sh
# from the repo root
# Build locally
docker build -t aegis-workspace:latest workspace-images/openssh-vscode

# Push to Docker Hub (requires prior `docker login`)
docker tag aegis-workspace:latest carlossanchez/aegis-workspace-vscode:latest
docker push carlossanchez/aegis-workspace-vscode:latest

# Optional: push to ttl.sh instead (public, expires after 24h)
TAG=ttl.sh/aegis-workspace-libstdcpp-$(date +%s):24h
docker tag aegis-workspace:latest "$TAG"
docker push "$TAG"
```

When running a local Docker Desktop Kubernetes cluster the tag
`carlossanchez/aegis-workspace-vscode:latest` is immediately pullable. Use the
fully qualified tag in your `SubmitWorkload` request:

```json
{
  "workload": {
    "projectId": "p-demo",
    "queue": "default",
    "workspace": {
      "flavor": "cpu-small",
      "image": "carlossanchez/aegis-workspace-vscode:latest",
      "interactive": true,
      "ports": [2222],
      "command": ["/init"],
      "env": {
        "PUID": "1000",
        "PGID": "1000",
        "PASSWORD_ACCESS": "true",
        "USER_NAME": "aegis",
        "USER_PASSWORD": "aegis123",
        "SSH_PORT": "2222"
      }
    }
  }
}
```

Rebuild and retag whenever you update the base image.
