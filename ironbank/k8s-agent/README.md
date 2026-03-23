# aegis-k8s-agent

Aegis Kubernetes operator agent for spoke cluster workload reconciliation.

## Description

The k8s-agent is a Kubernetes operator (built on controller-runtime) that runs on each spoke cluster in the Aegis hub-and-spoke architecture. It:

- Registers the spoke cluster with the hub via gRPC
- Sends periodic heartbeats with cluster capacity and GPU availability
- Reconciles AegisWorkload CRDs (workspaces and training jobs)
- Manages workspace pod lifecycle including VS Code Remote Extension Host injection
- Reports workload status back to the hub

## Base Image

`registry1.dso.mil/ironbank/redhat/ubi/ubi9-minimal:9.7`

Builder: `registry1.dso.mil/ironbank/google/golang/ubi9/golang-1.24:1.24.13`

## Kubernetes Deployment

Deployed as a Deployment in the `aegis-system` namespace on each spoke cluster via the `aegis-spoke` Helm chart.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aegis-k8s-agent
  namespace: aegis-system
spec:
  replicas: 1
  template:
    spec:
      containers:
        - name: manager
          image: registry1.dso.mil/ironbank/aegis/k8s-agent:1.0.0
          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi
          securityContext:
            runAsUser: 65532
            runAsGroup: 65532
            runAsNonRoot: true
```

## Resource Requirements

| Resource | Request | Limit |
|----------|---------|-------|
| CPU | 100m | 500m |
| Memory | 128Mi | 512Mi |
