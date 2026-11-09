# aegis-platform-api

Aegis Platform API — gRPC+REST control plane for multi-cluster GPU workload orchestration in secure DoD environments.

## Description

The platform-api is the central hub service of the Aegis Platform. It provides:

- gRPC and REST APIs for cluster registration, heartbeat, and lifecycle management
- Workload scheduling and GPU resource allocation across spoke clusters
- Project and queue-based budget enforcement
- RBAC integration with Keycloak OIDC
- Infrastructure provisioning via Pulumi (EKS, VPC, IAM)
- Interactive GPU development workspace orchestration

## Base Image

`registry1.dso.mil/ironbank/redhat/ubi/ubi9-minimal:9.7`

Builder: `registry1.dso.mil/ironbank/google/golang/ubi9/golang-1.24:1.24.13`

## Ports

| Port | Protocol | Description |
|------|----------|-------------|
| 8080 | HTTP | REST API (gRPC-Gateway) |
| 8081 | gRPC | gRPC API |

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `DATABASE_URL` | Yes | PostgreSQL connection string |
| `KEYCLOAK_URL` | Yes | Keycloak OIDC issuer URL |
| `AEGIS_EKS_CLUSTER_NAME` | No | EKS cluster name for token helper |
| `AEGIS_EKS_REGION` | No | AWS region for EKS operations |
| `AEGIS_EKS_ROLE_ARN` | No | IAM role ARN for cross-account access |

## Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: aegis-platform-api
  namespace: aegis-system
spec:
  replicas: 1
  selector:
    matchLabels:
      app: aegis-platform-api
  template:
    spec:
      containers:
        - name: platform-api
          image: registry1.dso.mil/ironbank/aegis/platform-api:1.0.0
          ports:
            - containerPort: 8080
            - containerPort: 8081
          resources:
            requests:
              cpu: 250m
              memory: 256Mi
            limits:
              cpu: "1"
              memory: 1Gi
          securityContext:
            runAsUser: 1000
            runAsGroup: 1000
            runAsNonRoot: true
            readOnlyRootFilesystem: true
```

## Resource Requirements

| Resource | Request | Limit |
|----------|---------|-------|
| CPU | 250m | 1000m |
| Memory | 256Mi | 1Gi |

## Runtime Dependencies

- PostgreSQL 15+
- Keycloak 24+ (OIDC provider)
- Kubernetes cluster (hub)

## External Resources (pre-fetched by Iron Bank pipeline)

- Pulumi CLI v3.226.0
- AWS CLI v2
- kubectl v1.33.0
- jq v1.8.0
