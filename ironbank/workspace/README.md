# aegis-workspace

Aegis GPU development workspace with CUDA runtime and VS Code Remote Extension Host support.

## Description

The workspace image provides an interactive GPU development environment for DoD data scientists and ML engineers. It includes:

- NVIDIA CUDA 12.6 runtime for GPU compute workloads
- VS Code Remote Extension Host (injected via init container)
- Tini init process manager for proper signal handling
- Git, bash, and essential development tools

## Base Image

`registry1.dso.mil/ironbank/opensource/nvidia/cuda:12.6`

## Ports

| Port | Protocol | Description |
|------|----------|-------------|
| 11111 | HTTP | VS Code Remote Extension Host server |

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `WORKSPACE_ROOT` | `/home/aegis/work` | Working directory for user files |
| `VSCODE_SERVER_PORT` | `11111` | VS Code server listen port |
| `VSCODE_QUALITY` | `stable` | VS Code release channel |
| `NVIDIA_VISIBLE_DEVICES` | `all` | GPU devices exposed to container |

## Kubernetes Deployment

Deployed as a workspace pod by the k8s-agent operator. Requires the `vscode-reh-init` init container to inject VS Code server binary via shared emptyDir volume.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: workspace-example
  namespace: aegis-system
spec:
  initContainers:
    - name: vscode-reh-init
      image: registry1.dso.mil/ironbank/aegis/vscode-reh-init:1.0.0
      volumeMounts:
        - name: reh-volume
          mountPath: /shared/reh
  containers:
    - name: workspace
      image: registry1.dso.mil/ironbank/aegis/workspace:1.0.0
      ports:
        - containerPort: 11111
      resources:
        requests:
          cpu: "1"
          memory: 4Gi
          nvidia.com/gpu: "1"
        limits:
          cpu: "4"
          memory: 16Gi
          nvidia.com/gpu: "1"
      volumeMounts:
        - name: reh-volume
          mountPath: /reh
  volumes:
    - name: reh-volume
      emptyDir: {}
```

## Resource Requirements

| Resource | Request | Limit |
|----------|---------|-------|
| CPU | 1 | 4 |
| Memory | 4Gi | 16Gi |
| GPU | 1 | 1 |

## External Resources (pre-fetched by Iron Bank pipeline)

- tini v0.19.0 (init process manager)
