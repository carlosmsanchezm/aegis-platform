# Workspace Images and Versioning

This document tracks the images we publish for interactive VS Code workspaces,
where they live, and how they are used across environments.

## Image Matrix

| Purpose | Registry / Tag | Architecture | Consumers | Notes |
|---------|----------------|--------------|-----------|-------|
| **Production / shared clusters** | `567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:<version>` | `linux/amd64` (multi-arch planned) | Platform API when launching workspaces via Backstage or VS Code against AWS | Tagged and pushed via CI. Update Terraform/Helm values when the tag changes. |
| **Local development default** | `aegis-workspace:latest` (local Docker daemon) | host arch | `SubmitWorkload` from a developer machine pointing at Docker Desktop’s Kubernetes | Rebuilt with `docker build … workspace-images/openssh-vscode`. Suitable for rapid iterations without pushing. |
| **Ephemeral published build** | `ttl.sh/aegis-workspace-<timestamp>:24h` | `linux/amd64`, `linux/arm64` | Ad hoc local testing where nodes need to pull from a registry | Built with `docker buildx build --platform linux/amd64,linux/arm64 … --push`. Expires automatically after 24h. |

All variants share the same Dockerfile (`workspace-images/openssh-vscode/Dockerfile`)
and scripts (`entrypoint.sh`, `start-reh.sh`). Any functional change should land in
the repository first, then be promoted to remote registries.

## Update Workflow

1. **Develop locally**
   ```bash
   docker build -t aegis-workspace:latest workspace-images/openssh-vscode
   ```
   Deploy a workspace pointing at `aegis-workspace:latest` to verify behaviour.

2. **Publish a multi-architecture image (for testing or release)**
   ```bash
   docker buildx create --name aegis-multi --driver docker-container --use
   docker buildx build \
     --platform linux/amd64,linux/arm64 \
     -t 567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:<newtag> \
     workspace-images/openssh-vscode \
     --push
   ```
   For quick local sharing, push to `ttl.sh/<name>:24h` instead.

3. **Update consumers**
   - **Local dev** – point `SubmitWorkload` requests (or Backstage overrides) at the new tag.
   - **Cloud** – update Terraform `generate-helm-values` outputs or Helm overrides and redeploy.

4. **Document the tag**
   Record the new tag in release notes or the deployment summary so QA knows which image to verify.

## Smoke Testing

### Container bootstrap

Run the helper script to ensure the VS Code Remote Extension Host (REH) starts for a given commit:

```bash
scripts/test-workspace-smoke.sh --image aegis-workspace:latest
```

If `VSCODE_COMMIT` is not set the script will try `code-insiders` and `code` to discover the running
version automatically. The container runs in isolation, logs the “Extension host agent listening”
banner, and the script tears it down.

### End-to-end workspace connectivity

To validate Kubernetes + platform API integration end-to-end:

```bash
export GRPC_ADDR=localhost:10081        # platform API port-forward
export WORKSPACE_IMAGE=ttl.sh/aegis-workspace-<timestamp>:24h   # or your local tag

scripts/test-workspace-connection.sh [image]
```

For TLS endpoints supply the additional parameters (note the default port is 443 when using the ingress):

```bash
GRPC_ADDR=platform-api-grpc.localtest.me:443 \
GRPC_TLS=1 \
GRPC_CA="$HOME/aegis-platform-api-ca.crt" \
GRPC_TLS_SERVER_NAME=platform-api-grpc.localtest.me \
scripts/test-workspace-connection.sh [image]
```

What the script does:

1. Idempotently ensures the `p-demo` project, `cpu-small` flavor, and `default` queue exist.
2. Submits a workspace pointing at `WORKSPACE_IMAGE` (positional argument or default) with the detected VS Code commit.
3. Waits for the pod to reach `Running`, and asserts the REH emitted “Extension host agent listening”.
4. Calls `CreateConnectionSession` and prints the returned `vscodeUri` / `sshConfig`.
5. Executes an in-cluster `curl http://127.0.0.1:11111` to confirm the VS Code server responds.
6. Writes a JSON bundle to `.aegis/workspace-session.json` so you can reuse the credentials.

By default the workspace is left running so you can finish the test manually from Backstage or the VS Code
extension. When you are done:

```bash
kubectl delete aegisworkload $(jq -r '.workloadId' .aegis/workspace-session.json) \
  -n aegis-workloads-local
```

For automated suites set `CLEANUP=1 scripts/test-workspace-connection.sh` to tear everything down once
the connection session has been validated.

To run the same check via Makefile (useful in CI), call:

```bash
make test-workspace
```

Use this in CI or before demos to guarantee the stack produces connectable workspaces.

## Troubleshooting

- **Image pulls succeed but VS Code server exits immediately**
  - Ensure `VSCODE_COMMIT` is injected (Backstage and VS Code clients do this automatically).
  - Verify the commit exists: `curl -I https://update.code.visualstudio.com/commit:<commit>/server-linux-x64/insider`.

- **Local pods fail with `ErrImagePull`**
  - Docker Desktop can only pull from local tags; push to `ttl.sh` or another registry if
    the cluster is remote.
  - Confirm you set `imagePullSecrets` or used a public registry.

- **Mismatch between production and local behaviour**
  - Confirm both reference the same tag table above.
  - `kubectl get deploy -n aegis-system aigis-spoke-k8s-agent -o yaml | grep image:` to inspect runtime images.

Keep this document current whenever workspace images or workflows change.

### VS Code extension settings

Both the smoke scripts and the extension assume the same core configuration. An example
`settings.json` snippet that matches the defaults used here:

```jsonc
{
  "aegis.remote": {
    "platform": {
      "grpcEndpoint": "localhost:10081",       // or platform-api-grpc.localtest.me:8081 via ingress
      "projectId": "p-demo",
      "namespace": "aegis-workloads-local",
      "authScope": "aegis-platform",
      "rejectUnauthorized": false,             // set true when using TLS + real certs
      "mtlsSource": "platform"                 // default; switch to "custom" for user-provided certs
    },
    "heartbeatIntervalMs": 15000,
    "idleTimeoutMs": 45000
  }
}
```

Keep the namespace/project values aligned between tests, scripts, and your VS Code profile so that
connectivity checks exercise the same path end-to-end.
