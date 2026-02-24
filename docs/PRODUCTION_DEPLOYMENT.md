# Aegis Platform - Production Deployment Guide

This document provides step-by-step instructions for deploying the Aegis platform to a production environment on AWS EKS. It covers infrastructure provisioning, image building, secret management, Helm-based service deployment, Keycloak SSO configuration, post-deployment verification, rollback procedures, and troubleshooting.

---

## Prerequisites

### Required Tools

| Tool | Minimum Version | Purpose |
|------|----------------|---------|
| AWS CLI | 2.x | AWS resource management |
| Terraform | 1.6+ | Infrastructure provisioning |
| kubectl | 1.27+ | Kubernetes cluster management |
| Helm | 3.12+ | Chart-based deployment |
| Docker (with buildx) | 24+ | Container image builds |
| grpcurl | 1.8+ | gRPC endpoint testing |
| jq | 1.7+ | JSON processing |
| Python 3 | 3.8+ | URL encoding in deployment scripts |

### Infrastructure Requirements

- **AWS Account** with appropriate IAM permissions (or GovCloud for IL4/IL5 environments)
- **EKS Cluster** (Kubernetes 1.27+; Terraform defaults to 1.33)
- **RDS PostgreSQL 15+** instance (Terraform provisions `db.t3.micro` by default; size up for production)
- **ECR Repositories** for container images (`aegis/platform-api`, `aegis/proxy`, `aegis/k8s-agent`, `aegis/workspace-vscode`)
- **VPC** with public, private, and database subnets across 3 AZs
- **Route53 Hosted Zone** for DNS (e.g., `aegist.dev`)
- **S3 Bucket + DynamoDB Table** for Terraform remote state
- **Keycloak Instance** for OIDC authentication (deployed as part of the aegis-services chart, or bring your own)
- **Domain Name** with DNS control for platform-api, proxy, and Keycloak endpoints

### Network Requirements

- EKS worker nodes must be able to reach RDS on port 5432 (database security group)
- LoadBalancers (NLB) must be internet-facing or internal based on your access model
- Port 8080 (HTTP gateway / proxy) and 8081 (gRPC) must be accessible to clients
- Keycloak must be reachable by both platform-api (for OIDC token validation) and end users (for login flows)

---

## Architecture Overview

Aegis uses a **hub-and-spoke** model:

```
                    ┌────────────────────────────────────────────┐
                    │          Hub Cluster (aegis-services)       │
                    │                                            │
  Users/Backstage ──┤  ┌──────────────┐   ┌──────────────┐      │
                    │  │ platform-api │   │    proxy     │      │
                    │  │  (gRPC+HTTP) │   │  (WebSocket) │      │
                    │  └──────┬───────┘   └──────┬───────┘      │
                    │         │                  │              │
                    │  ┌──────┴──────┐   ┌──────┴──────┐      │
                    │  │  PostgreSQL  │   │  Keycloak   │      │
                    │  │   (RDS)      │   │   (OIDC)    │      │
                    │  └─────────────┘   └─────────────┘      │
                    └──────────┬─────────────────┬─────────────┘
                               │                 │
              ┌────────────────┘                 └────────────────┐
              │                                                  │
    ┌─────────┴──────────┐                          ┌─────────────┴──────────┐
    │  Spoke Cluster A   │                          │  Spoke Cluster B       │
    │  ┌──────────────┐  │                          │  ┌──────────────┐      │
    │  │  k8s-agent   │  │                          │  │  k8s-agent   │      │
    │  └──────────────┘  │                          │  └──────────────┘      │
    │  Workload pods     │                          │  Workload pods         │
    └────────────────────┘                          └────────────────────────┘
```

- **Hub (aegis-services chart)**: Central control plane. Runs `platform-api` (gRPC + REST gateway) for workload management, placement decisions, audit logging, and compliance enforcement. Runs `proxy` for WebSocket-based access to workloads. Optionally runs Keycloak for OIDC.
- **Spoke (aegis-spoke chart)**: Deployed once per workload cluster. Runs `k8s-agent` which registers the cluster with the hub, reports capacity/flavors, reconciles `AegisWorkload` CRDs, creates and manages Kubernetes Jobs, and streams status updates back to the hub.

---

## Step 1: Infrastructure Provisioning

All infrastructure is defined in `terraform/` and uses AWS provider ~5.0.

### 1.1 Bootstrap Remote State

If this is a first-time setup, create the S3 backend for Terraform state:

```bash
cd terraform

# The backend is configured in main.tf:
#   bucket         = "aegis-platform-tf-state-bucket"
#   key            = "aegis/prod/terraform.tfstate"
#   region         = "us-east-1"
#   dynamodb_table = "aegis-terraform-locks"

# Create the state bucket and lock table manually or use the bootstrap script:
./bootstrap-remote-state.sh \
  --bucket aegis-platform-tf-state-bucket \
  --region us-east-1 \
  --dynamodb-table aegis-terraform-locks
```

### 1.2 Configure Variables

Create a `terraform.tfvars` file (or pass `-var` flags):

```hcl
# terraform.tfvars
aws_region           = "us-east-1"
aws_profile          = ""                 # Leave blank for env-based auth; set for named profiles
environment          = "prod"
cluster_name_prefix  = "aegis-spoke"
cluster_version      = "1.33"

# VPC
vpc_cidr             = "10.0.0.0/16"
availability_zones   = ["us-east-1a", "us-east-1b", "us-east-1c"]

# EKS Node Groups
cpu_instance_type    = "t3.medium"        # Size up for production (e.g., m5.xlarge)
cpu_desired_capacity = 2
gpu_instance_type    = "g4dn.xlarge"
gpu_desired_capacity = 0                  # Set >0 when GPU workloads are needed
gpu_max_capacity     = 2
use_spot_instances   = false              # Use on-demand for production

# RDS
db_instance_class       = "db.t3.micro"  # Size up for production (e.g., db.r6g.large)
db_allocated_storage    = 20
db_max_allocated_storage = 100
db_postgres_version     = "15.12"
db_backup_retention     = 7
db_backup_window        = "03:00-04:00"
db_maintenance_window   = "sun:04:00-sun:05:00"
db_skip_final_snapshot  = false
db_deletion_protection  = true            # Enable for production

# ECR
manage_ecr_repositories = true
ecr_repositories        = ["aegis/k8s-agent", "aegis/proxy", "aegis/platform-api", "aegis/workspace-vscode"]
```

### 1.3 Initialize and Apply

```bash
cd terraform

terraform init -backend-config=backend.hcl  # Or rely on the backend block in main.tf

terraform plan -out=tfplan
terraform apply tfplan
```

### 1.4 Record Key Outputs

After `terraform apply` completes, note these outputs:

```bash
# Database connection info
terraform output -json database_connection_info | jq
terraform output -raw rds_endpoint              # e.g., aegis-prod-rds.xxxx.us-east-1.rds.amazonaws.com:5432
terraform output -raw db_password_secret_value   # Database password
terraform output rds_database_name               # "aegis"
terraform output rds_username                    # "aegis_api"

# ECR registry
terraform output -raw ecr_registry_url           # e.g., 567751785679.dkr.ecr.us-east-1.amazonaws.com

# Cluster info
terraform output -raw cluster_name
terraform output -raw kubectl_config_command

# Secrets
terraform output -raw jwt_secret_value

# VPC/networking
terraform output vpc_id
terraform output private_subnet_ids
```

---

## Step 2: Build and Push Images

The platform has three services, each with its own Dockerfile:

| Service | Dockerfile | Ports | Base Image |
|---------|-----------|-------|------------|
| platform-api | `services/platform-api/Dockerfile` | 8080 (HTTP), 8081 (gRPC) | UBI9 Minimal |
| proxy | `services/proxy/Dockerfile` | 8080 | distroless |
| k8s-agent | `agents/k8s-agent/Dockerfile` | - | distroless |

### 2.1 Authenticate with ECR

```bash
ECR_REGISTRY=$(terraform -chdir=terraform output -raw ecr_registry_url)

aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin "${ECR_REGISTRY}"
```

### 2.2 Build and Push

All builds must target `linux/amd64` for EKS:

```bash
# Platform API
docker buildx build --platform linux/amd64 \
  -t "${ECR_REGISTRY}/aegis/platform-api:0.1.0" \
  -f services/platform-api/Dockerfile . --push

# Proxy
docker buildx build --platform linux/amd64 \
  -t "${ECR_REGISTRY}/aegis/proxy:0.1.0" \
  -f services/proxy/Dockerfile . --push

# K8s Agent
docker buildx build --platform linux/amd64 \
  -t "${ECR_REGISTRY}/aegis/k8s-agent:0.1.0" \
  -f agents/k8s-agent/Dockerfile . --push
```

> **Tip**: Tag images with semver (e.g., `0.1.0`). The common.yaml defaults to `0.1.0` for platform-api. Avoid `latest` in production.

---

## Step 3: Configure Secrets

### 3.1 Configure kubectl

```bash
# Use the Terraform-generated command
eval "$(terraform -chdir=terraform output -raw kubectl_config_command)"

# Or manually:
aws eks update-kubeconfig \
  --region us-east-1 \
  --name aegis-spoke-prod
```

### 3.2 Create Namespace

```bash
kubectl create namespace aegis-system --dry-run=client -o yaml | kubectl apply -f -
```

### 3.3 Create Platform Secrets

```bash
DB_PASSWORD=$(terraform -chdir=terraform output -raw db_password_secret_value)
JWT_SECRET=$(terraform -chdir=terraform output -raw jwt_secret_value)

kubectl create secret generic aegis-platform-secrets \
  --from-literal=db-password="${DB_PASSWORD}" \
  --from-literal=proxy-jwt-secret="${JWT_SECRET}" \
  --namespace aegis-system \
  --dry-run=client -o yaml | kubectl apply -f -
```

### 3.4 Create Keycloak Secrets

```bash
# Admin credentials
kubectl create secret generic keycloak-admin-secret \
  --from-literal=username="admin" \
  --from-literal=password="<KEYCLOAK_ADMIN_PASSWORD>" \
  --namespace aegis-system \
  --dry-run=client -o yaml | kubectl apply -f -

# Database credentials (for Keycloak's own PostgreSQL)
kubectl create secret generic keycloak-db-secret \
  --from-literal=username="keycloak" \
  --from-literal=password="<KEYCLOAK_DB_PASSWORD>" \
  --namespace aegis-system \
  --dry-run=client -o yaml | kubectl apply -f -

# Backstage OIDC client secret
kubectl create secret generic keycloak-backstage-client-secret \
  --from-literal=clientSecret="<BACKSTAGE_CLIENT_SECRET>" \
  --namespace aegis-system \
  --dry-run=client -o yaml | kubectl apply -f -
```

### Required Secrets Summary

| Secret Name | Keys | Purpose |
|-------------|------|---------|
| `aegis-platform-secrets` | `db-password`, `proxy-jwt-secret` | Platform API DB auth + Proxy JWT signing |
| `keycloak-admin-secret` | `username`, `password` | Keycloak admin console access |
| `keycloak-db-secret` | `username`, `password` | Keycloak's internal PostgreSQL credentials |
| `keycloak-backstage-client-secret` | `clientSecret` | OIDC client secret for Backstage |
| `aegis-kubeconfigs` | (cluster kubeconfigs) | Managed by Helm; for multi-cluster workload placement |
| `aegis-trust-bundle` | `ca.crt` | Internal PKI CA certificate (created by PKI installer) |

> **Note**: For production, use AWS Secrets Manager or HashiCorp Vault with an External Secrets Operator rather than creating secrets directly. Credentials are also stored in AWS Secrets Manager at `aegis/prod/db-password` and `aegis/prod/proxy-jwt-secret`.

---

## Step 4: Deploy Hub (aegis-services)

The `aegis-services` chart deploys the hub control plane: `platform-api`, `proxy`, and optionally `Keycloak`.

### 4.1 Generate Terraform-Derived Values

```bash
cd terraform

# Generate Helm values from Terraform outputs
terraform output -raw helm_values_aegis_services > ../charts/aegis-services/values-cloud-generated.yaml
terraform output -raw helm_values_aegis_spoke > ../charts/aegis-spoke/values-cloud-generated.yaml
```

### 4.2 Create Production Overrides File

Create an `overrides.yaml` file that sets image tags, secrets, and environment-specific configuration:

```yaml
# overrides.yaml
platformApi:
  image:
    repository: <ECR_REGISTRY>/aegis/platform-api
    tag: "0.1.0"
  env:
    DATABASE_URL: "postgres://aegis_api:<URL_ENCODED_PASSWORD>@<RDS_HOST>:5432/aegis?sslmode=require"
    OIDC_ISSUER_URL: "https://aegis-keycloak-service.aegis-system.svc.cluster.local:8443/realms/aegis"
    OIDC_AUDIENCE: "backstage"
    OIDC_JWKS_URL: "https://aegis-keycloak-service.aegis-system.svc.cluster.local:8443/realms/aegis/protocol/openid-connect/certs"
    AEGIS_PROXY_BASE_URL: "wss://proxy.aegist.dev:8080"
    LOG_LEVEL: "info"
    ENVIRONMENT: "production"
    AEGIS_STORE_BACKEND: "postgres"
    DB_SSLMODE: "require"
    PG_MAX_OPEN_CONNS: "50"
    PG_MAX_IDLE_CONNS: "10"
    PG_CONN_MAX_LIFETIME: "30m"
  secrets:
    db-password: "<DB_PASSWORD>"
    proxy-jwt-secret: "<JWT_SECRET>"
  tls:
    certManager:
      enabled: true
      dnsNames:
        - "platform-api-grpc.aegist.dev"
        - "platform-api.aegist.dev"

proxy:
  image:
    repository: <ECR_REGISTRY>/aegis/proxy
    tag: "0.1.0"
  jwtSecret: "<JWT_SECRET>"
  publicHost: "proxy.aegist.dev"
  tls:
    enabled: true
    certManager:
      enabled: true
      dnsNames:
        - "proxy.aegist.dev"

keycloak:
  enabled: true
  forceRender: true
  namespace: aegis-system
  hostname:
    hostname: "https://aegis-keycloak-service.aegis-system.svc.cluster.local:8443"
    admin: "https://aegis-keycloak-service.aegis-system.svc.cluster.local:8443"
    strict: false
  tls:
    secret:
      name: keycloak-tls
      create: false
    certManager:
      enabled: true
```

### 4.3 Run Database Migrations

The deployment script runs migrations via a Kubernetes Job. If deploying manually:

```bash
DB_PASSWORD=$(terraform -chdir=terraform output -raw db_password_secret_value)
DB_HOST=$(terraform -chdir=terraform output -raw rds_endpoint | cut -d: -f1)

# Create a ConfigMap from the migration SQL
kubectl -n aegis-system create configmap aegis-migrations \
  --from-file=0001_init.sql=services/platform-api/migrations/0001_init.sql \
  --dry-run=client -o yaml | kubectl apply -f -

# Run migration via a temporary pod
kubectl run aegis-migrate --rm -i --restart=Never \
  --image=postgres:16-alpine \
  --namespace=aegis-system \
  --env="PGPASSWORD=${DB_PASSWORD}" \
  -- psql -h "${DB_HOST}" -U aegis_api -d aegis \
  -v ON_ERROR_STOP=1 -f /dev/stdin < services/platform-api/migrations/0001_init.sql
```

The migration files are (applied in order):
1. `0001_init.sql` - Core schema (projects, clusters, workloads, queues, etc.)
2. `0002_add_project_annotations.sql` - Project annotation support
3. `0003_add_cluster_soft_delete.sql` - Cluster soft-delete fields
4. `0004_provisioning_logs.sql` - Provisioning log table
5. `0005_add_cluster_proxy_url.sql` - Cluster proxy URL field
6. `0006_add_cluster_import_fields.sql` - Cluster import fields
7. `0007_multi_tenancy_fixes.sql` - Multi-tenancy improvements
8. `0008_add_workload_session_management.sql` - Session management
9. `0009_add_workload_runtime_seconds.sql` - Runtime tracking
10. `0010_audit_events.sql` - Audit event table

### 4.4 Install Internal PKI (cert-manager + step-ca)

The platform uses an internal PKI for pod-to-pod TLS. Install it before deploying Helm charts:

```bash
# The PKI installer script handles cert-manager, step-ca, and step-issuer
TRUST_BUNDLE_NAMESPACES="aegis-system" \
PKI_NAMESPACE="aegis-pki" \
CERT_MANAGER_NAMESPACE="cert-manager" \
  ./scripts/install-internal-pki.sh
```

After installation, the `aegis-trust-bundle` secret (containing `ca.crt`) is distributed to the `aegis-system` namespace.

### 4.5 Apply CRDs

```bash
# Install k8s-agent CRDs before Helm upgrade
kubectl apply -f agents/k8s-agent/config/crd/bases/
```

### 4.6 Deploy with Helm

```bash
cd charts

helm upgrade --install aegis ./aegis-services \
  -f ./aegis-services/values/common.yaml \
  -f ./aegis-services/values/cloud.yaml \
  -f ./aegis-services/values-cloud-generated.yaml \
  -f overrides.yaml \
  --namespace aegis-system \
  --create-namespace \
  --timeout 10m
```

**Values file layering order** (later files override earlier):

| Layer | File | Purpose |
|-------|------|---------|
| 1 | `values/common.yaml` | Base defaults for all environments (images, resources, security contexts, FedRAMP controls) |
| 2 | `values/cloud.yaml` | Cloud-specific: LoadBalancer services, production replicas, RDS config, pod anti-affinity |
| 3 | `values-cloud-generated.yaml` | Auto-generated from Terraform outputs (RDS endpoint, DNS names, ECR URLs) |
| 4 | `overrides.yaml` | Your secrets, image tags, and cert-manager DNS SANs |

### 4.7 Verify Hub Deployment

```bash
# Wait for rollout
kubectl rollout status deployment/aegis-platform-api -n aegis-system --timeout=5m
kubectl rollout status deployment/aegis-proxy -n aegis-system --timeout=5m

# Check pods
kubectl get pods -n aegis-system

# Expected output:
# NAME                                      READY   STATUS    RESTARTS   AGE
# aegis-platform-api-xxxxx                  1/1     Running   0          2m
# aegis-proxy-xxxxx                         1/1     Running   0          2m
# aegis-keycloak-0                          1/1     Running   0          3m
# aegis-keycloak-db-0                       1/1     Running   0          3m
```

### 4.8 Wait for Load Balancers and Update DNS

```bash
# Get platform-api LoadBalancer hostname
PLATFORM_API_LB=$(kubectl get svc aegis-platform-api -n aegis-system \
  -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

# Get proxy LoadBalancer hostname
PROXY_LB=$(kubectl get svc aegis-proxy -n aegis-system \
  -o jsonpath='{.status.loadBalancer.ingress[0].hostname}')

echo "Platform API LB: ${PLATFORM_API_LB}"
echo "Proxy LB: ${PROXY_LB}"

# Update Route53 DNS records
cd terraform
terraform apply -auto-approve \
  -var="platform_api_lb_hostname=${PLATFORM_API_LB}" \
  -var="proxy_lb_hostname=${PROXY_LB}" \
  -target=aws_route53_record.platform_api_grpc \
  -target=aws_route53_record.platform_api_http \
  -target=aws_route53_record.proxy
```

---

## Step 5: Deploy Spoke (aegis-spoke)

The `aegis-spoke` chart deploys the `k8s-agent` on each workload cluster. If the spoke runs in the same cluster as the hub, deploy it in the same namespace. For remote spoke clusters, configure kubectl to point to that cluster first.

### 5.1 Deploy with Helm

```bash
cd charts

helm upgrade --install aegis-spoke ./aegis-spoke \
  -f ./aegis-spoke/values-cloud-generated.yaml \
  -f ./aegis-spoke/values-cloud-tls.yaml \
  --set k8sAgent.enabled=true \
  --set k8sAgent.replicaCount=1 \
  --set k8sAgent.image.repository=<ECR_REGISTRY>/aegis/k8s-agent \
  --set k8sAgent.image.tag=0.1.0 \
  --set k8sAgent.trustBundle.enabled=true \
  --set k8sAgent.trustBundle.secretName=aegis-trust-bundle \
  --set proxy.enabled=false \
  --namespace aegis-system \
  --create-namespace \
  --timeout 5m
```

### 5.2 Key k8s-agent Environment Variables

These are typically set in `values-cloud-generated.yaml` but can be overridden:

| Variable | Description | Example |
|----------|-------------|---------|
| `AEGIS_CLUSTER_ID` | Unique identifier for this spoke cluster | `aws-us-east-1-prod` |
| `AEGIS_CP_GRPC` | Hub platform-api gRPC endpoint | `platform-api-grpc.aegist.dev:8081` |
| `AEGIS_CP_GRPC_INSECURE` | Set to `"false"` for TLS connections | `"false"` |
| `AEGIS_REGION` | AWS region where this cluster is deployed | `us-east-1` |
| `AEGIS_PROVIDER` | Cloud provider | `aws` |
| `AEGIS_FLAVORS` | Comma-separated compute flavors this cluster offers | `cpu-small,cpu-large,gpu-a100` |
| `AEGIS_DEFAULT_IMAGE` | Default workspace image | `<ECR>/aegis/workspace-vscode:latest` |
| `AEGIS_PROXY_INGRESS_HOST` | Public hostname for spoke proxy (if enabled) | `proxy.spoke-a.aegist.dev` |
| `AEGIS_WORKLOAD_GC_ENABLED` | Enable orphaned workload garbage collection | `"true"` |

### 5.3 Verify Spoke Registration

```bash
# Check k8s-agent pod
kubectl get pods -n aegis-system -l app.kubernetes.io/component=k8s-agent

# Check logs for successful registration
kubectl logs -n aegis-system -l app.kubernetes.io/component=k8s-agent --tail=50

# Look for "cluster registered" in platform-api logs
kubectl logs -n aegis-system -l app.kubernetes.io/component=platform-api | grep "cluster registered"
```

### 5.4 Deploying Additional Spokes

For each additional workload cluster:

1. Switch kubectl context to the target cluster
2. Copy the `aegis-trust-bundle` secret (CA cert) to the new cluster's namespace
3. Run the Helm install with the cluster-specific `AEGIS_CLUSTER_ID`, `AEGIS_REGION`, and `AEGIS_FLAVORS`
4. Verify the cluster appears in the hub's cluster list

---

## Step 6: Configure Keycloak

Keycloak is deployed as part of the `aegis-services` chart when `keycloak.enabled=true`. It uses the `aegis-realm.json` file for initial realm import.

### 6.1 Realm Configuration (Auto-Imported)

The chart automatically imports the `aegis` realm with:

- **Realm**: `aegis` (display name: "Aegis")
- **Client**: `backstage` (public client for the Backstage UI)
- **Token Exchange**: Enabled (for VS Code extension token flow)
- **Access Token Lifespan**: 300 seconds (5 minutes)
- **OTP Policy**: TOTP, HmacSHA1, 6 digits, 30s period

The redirect URIs and web origins default to local development URLs. For production, update them:

### 6.2 Update Client Redirect URIs

Access the Keycloak admin console and update the `backstage` client:

```
Redirect URIs:
  - https://backstage.yourdomain.com/*
  - http://localhost:7008/*       (for local Backstage development)

Web Origins:
  - https://backstage.yourdomain.com
  - http://localhost:3000          (for local Backstage development)
```

### 6.3 Create Users

In the Keycloak admin console (`https://<keycloak-host>:8443`):

1. Navigate to **Aegis realm** > **Users** > **Add user**
2. Set username, email, first name, last name
3. Under **Credentials**, set a temporary password
4. Under **Role mapping**, assign appropriate roles

### 6.4 Platform-API OIDC Configuration

The platform-api connects to Keycloak for token validation. The relevant environment variables are:

```yaml
platformApi:
  env:
    OIDC_ISSUER_URL: "https://<keycloak-host>:8443/realms/aegis"
    OIDC_AUDIENCE: "backstage"
    OIDC_JWKS_URL: "https://<keycloak-host>:8443/realms/aegis/protocol/openid-connect/certs"
    REQUIRE_PHISHING_RESISTANT_MFA: "true"          # Set to "false" for pilot
    ALLOWED_PHISHING_RESISTANT_AMR: "hwk,webauthn,piv,piv-cac"
```

When Keycloak runs in-cluster, the internal hostname is:
```
aegis-keycloak-service.aegis-system.svc.cluster.local:8443
```

### 6.5 VS Code Extension OIDC (Service Account)

For the spoke k8s-agent to authenticate with the hub (optional, recommended for production):

```yaml
k8sAgent:
  env:
    AEGIS_CP_OIDC_TOKEN_URL: "https://<keycloak-host>:8443/realms/aegis/protocol/openid-connect/token"
    AEGIS_CP_OIDC_CLIENT_ID: "<service-account-client-id>"
    AEGIS_CP_OIDC_CLIENT_SECRET: "<service-account-client-secret>"
    AEGIS_CP_OIDC_AUDIENCE: "backstage"
```

---

## Step 7: Post-Deployment Verification

### 7.1 Health Checks

```bash
# Check all pods are running
kubectl get pods -n aegis-system

# Check services and LoadBalancers
kubectl get svc -n aegis-system

# Check platform-api logs for startup success
kubectl logs deployment/aegis-platform-api -n aegis-system --tail=20

# Check proxy logs
kubectl logs deployment/aegis-proxy -n aegis-system --tail=20

# Check k8s-agent logs
kubectl logs deployment/aegis-spoke-aegis-spoke-k8s-agent -n aegis-system --tail=20
```

### 7.2 Test gRPC Connectivity

```bash
# Export the CA bundle (from the PKI setup)
export CA_BUNDLE="${HOME}/aegis-platform-api-ca.crt"

# Refresh CA bundle from cluster
kubectl get secret aegis-trust-bundle -n aegis-system \
  -o "jsonpath={.data.ca\.crt}" | base64 --decode > "${CA_BUNDLE}"

# Get a Keycloak token (use the keycloak-token.sh script or manually):
export AEGIS_BEARER_TOKEN="<your-bearer-token>"

# Test CreateProject
grpcurl -cacert "${CA_BUNDLE}" \
  -H "authorization: Bearer ${AEGIS_BEARER_TOKEN}" \
  -d '{"project":{"id":"p-test","displayName":"Test Project","ownerGroup":"eng"}}' \
  platform-api-grpc.aegist.dev:8081 aegis.v1.AegisPlatform/CreateProject

# List projects
grpcurl -cacert "${CA_BUNDLE}" \
  -H "authorization: Bearer ${AEGIS_BEARER_TOKEN}" \
  platform-api-grpc.aegist.dev:8081 aegis.v1.AegisPlatform/ListProjects

# List clusters (should show registered spoke)
grpcurl -cacert "${CA_BUNDLE}" \
  -H "authorization: Bearer ${AEGIS_BEARER_TOKEN}" \
  platform-api-grpc.aegist.dev:8081 aegis.v1.AegisPlatform/ListClusters
```

### 7.3 Run the E2E Test Script

```bash
export AEGIS_GRPC_ADDR="platform-api-grpc.aegist.dev:8081"
export GRPC_TLS=1
export GRPC_CA="${HOME}/aegis-platform-api-ca.crt"
export AEGIS_BEARER_TOKEN="<your-token>"

./scripts/e2e-platform-api.sh
```

### 7.4 Submit a Test Workload

```bash
grpcurl -cacert "${CA_BUNDLE}" \
  -H "authorization: Bearer ${AEGIS_BEARER_TOKEN}" \
  -d '{
    "workload": {
      "projectId": "p-test",
      "displayName": "smoke-test",
      "queue": "default",
      "flavor": "cpu-small",
      "image": "busybox:latest",
      "command": ["echo", "hello aegis"]
    }
  }' \
  platform-api-grpc.aegist.dev:8081 aegis.v1.AegisPlatform/SubmitWorkload
```

### 7.5 Verify Audit Events

```bash
# List recent audit events
grpcurl -cacert "${CA_BUNDLE}" \
  -H "authorization: Bearer ${AEGIS_BEARER_TOKEN}" \
  -d '{"pageSize": 10}' \
  platform-api-grpc.aegist.dev:8081 aegis.v1.AegisPlatform/ListAuditEvents
```

---

## Automated Deployment (Recommended)

The `generate-cloud-deployment.sh` script automates Steps 3-5 above:

```bash
cd terraform

# Interactive mode (prompts before deploying)
./generate-cloud-deployment.sh

# Non-interactive mode (for CI/CD)
./generate-cloud-deployment.sh --non-interactive
```

The script will:
1. Generate Helm values from Terraform outputs
2. Configure kubectl for the EKS cluster
3. Create namespace and Kubernetes secrets
4. Run database migrations via a Kubernetes Job
5. Install internal PKI (cert-manager + step-ca)
6. Deploy aegis-services with Helm
7. Wait for LoadBalancers and update Route53 DNS
8. Deploy aegis-spoke with Helm
9. Update Backstage configuration files

Environment variables to customize the script:

| Variable | Default | Purpose |
|----------|---------|---------|
| `PLATFORM_API_IMAGE_TAG` | `v1.0.6-tls2` | Platform API image tag (can include repo:tag) |
| `PROXY_IMAGE_TAG` | `no-client-cert` | Proxy image tag |
| `K8S_AGENT_IMAGE_TAG` | `v1.0.2-tls-20251005-amd64` | K8s agent image tag |
| `K8S_NAMESPACE` | `aegis-system` | Target Kubernetes namespace |
| `HELM_RELEASE` | `aegis` | Helm release name |
| `KEYCLOAK_ADMIN_PASSWORD` | - | Keycloak admin password |
| `KEYCLOAK_DB_PASSWORD` | - | Keycloak DB password |
| `KEYCLOAK_BACKSTAGE_CLIENT_SECRET` | - | OIDC client secret |
| `SKIP_ROUTE53_UPDATE` | `0` | Set to `1` to skip DNS updates |

---

## Rollback Procedure

### Helm Rollback

```bash
# List release history
helm history aegis -n aegis-system
helm history aegis-spoke -n aegis-system

# Rollback to a previous revision
helm rollback aegis <REVISION> -n aegis-system
helm rollback aegis-spoke <REVISION> -n aegis-system

# Verify rollback
kubectl rollout status deployment/aegis-platform-api -n aegis-system
kubectl rollout status deployment/aegis-proxy -n aegis-system
```

### Database Rollback

RDS automated backups are configured with a 7-day retention window (backup window: 03:00-04:00 UTC):

```bash
# List available snapshots
aws rds describe-db-snapshots \
  --db-instance-identifier aegis-prod-rds \
  --query 'DBSnapshots[].{ID:DBSnapshotIdentifier,Time:SnapshotCreateTime,Status:Status}' \
  --output table

# Point-in-time restore (creates a new RDS instance)
aws rds restore-db-instance-to-point-in-time \
  --source-db-instance-identifier aegis-prod-rds \
  --target-db-instance-identifier aegis-prod-rds-restored \
  --restore-time "2025-01-15T10:00:00Z" \
  --region us-east-1
```

After restoring, update the `DATABASE_URL` / `DB_HOST` in your overrides and redeploy.

### Full Rollback (Nuclear Option)

```bash
# Uninstall all Helm releases
helm uninstall aegis -n aegis-system
helm uninstall aegis-spoke -n aegis-system

# Delete namespace (removes all resources)
kubectl delete namespace aegis-system

# Redeploy from scratch using the steps above
```

---

## Troubleshooting

### Pod CrashLoopBackOff

```bash
# Check pod events
kubectl describe pod <POD_NAME> -n aegis-system

# Check logs (including previous crash)
kubectl logs <POD_NAME> -n aegis-system --previous

# Common causes:
# - Database connection failure (check DB_HOST, DB_PASSWORD, security groups)
# - Missing secrets (check kubectl get secrets -n aegis-system)
# - Image pull errors (check ECR authentication)
# - OOM kills (check resource limits in values)
```

### Database Connection Failures

```bash
# 1. Verify password is correct
terraform -chdir=terraform output -raw db_password_secret_value

# 2. Test connection from within the cluster
kubectl run psql-test --rm -i --restart=Never \
  --image=postgres:16-alpine \
  --namespace=aegis-system \
  --env="PGPASSWORD=<PASSWORD>" \
  -- psql -h <RDS_HOST> -U aegis_api -d aegis -c "SELECT version();"

# 3. Check security groups allow EKS nodes to reach RDS on port 5432
# 4. Check the platform-api pod's environment:
kubectl get pod -n aegis-system -l app.kubernetes.io/component=platform-api \
  -o jsonpath='{.items[0].spec.containers[0].env}' | jq .

# 5. If "relation does not exist" errors, re-run migrations (see Step 4.3)
```

### Authentication Errors (401/403)

```bash
# 1. Check OIDC configuration in platform-api
kubectl logs -n aegis-system -l app.kubernetes.io/component=platform-api | grep -i "oidc\|auth\|token"

# 2. Verify Keycloak is running and accessible from platform-api
kubectl exec -it deployment/aegis-platform-api -n aegis-system -- \
  curl -sk https://aegis-keycloak-service.aegis-system.svc.cluster.local:8443/realms/aegis/.well-known/openid-configuration

# 3. Verify the OIDC CA bundle is mounted
kubectl exec -it deployment/aegis-platform-api -n aegis-system -- \
  ls -la /etc/aegis-platform-api/oidc/

# 4. Check that the bearer token audience matches OIDC_AUDIENCE ("backstage")
# 5. If REQUIRE_PHISHING_RESISTANT_MFA is "true", ensure MFA is configured for the user
```

### Cluster Registration Failures

```bash
# 1. Check k8s-agent logs
kubectl logs -n aegis-system -l app.kubernetes.io/component=k8s-agent --tail=100

# 2. Verify the agent can reach the hub
kubectl exec -it deployment/aegis-spoke-aegis-spoke-k8s-agent -n aegis-system -- \
  wget -q -O- --no-check-certificate https://platform-api-grpc.aegist.dev:8081 || echo "Connection test complete"

# 3. Check AEGIS_CP_GRPC is set correctly
kubectl get deployment aegis-spoke-aegis-spoke-k8s-agent -n aegis-system \
  -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="AEGIS_CP_GRPC")].value}'

# 4. Verify TLS settings (AEGIS_CP_GRPC_INSECURE should be "false" for TLS)
# 5. Check that the trust bundle secret exists:
kubectl get secret aegis-trust-bundle -n aegis-system
```

### Workload Scheduling Issues

```bash
# 1. Check workload status via gRPC
grpcurl -cacert "${CA_BUNDLE}" \
  -H "authorization: Bearer ${TOKEN}" \
  -d '{"projectId":"<PROJECT_ID>"}' \
  platform-api-grpc.aegist.dev:8081 aegis.v1.AegisPlatform/ListWorkloads

# 2. Check if the target cluster has the required flavor
# (AEGIS_FLAVORS on the spoke must include the requested flavor)

# 3. Check AegisWorkload CRDs on the spoke cluster
kubectl get aegisworkloads -n aegis-workloads

# 4. Check k8s-agent reconciliation logs
kubectl logs -n aegis-system -l app.kubernetes.io/component=k8s-agent | grep -i "reconcil\|workload"

# 5. Check if workload namespace exists on spoke
kubectl get namespace aegis-workloads
```

### Image Pull Errors

```bash
# Re-authenticate with ECR (tokens expire after 12 hours)
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin \
  $(terraform -chdir=terraform output -raw ecr_registry_url)

# For EKS nodes, ensure the IAM role attached to the node group has
# AmazonEC2ContainerRegistryReadOnly policy

# Check if the image exists
aws ecr describe-images --repository-name aegis/platform-api --region us-east-1
```

### TLS / Certificate Issues

```bash
# 1. Check if cert-manager certificates are issued
kubectl get certificates -n aegis-system
kubectl get certificaterequests -n aegis-system

# 2. Check cert-manager logs
kubectl logs -n cert-manager -l app.kubernetes.io/component=controller --tail=50

# 3. Verify the CA bundle is current
kubectl get secret aegis-trust-bundle -n aegis-system -o jsonpath='{.data.ca\.crt}' | \
  base64 -d | openssl x509 -text -noout | head -20

# 4. Re-export the CA bundle
kubectl get secret aegis-trust-bundle -n aegis-system \
  -o "jsonpath={.data.ca\.crt}" | base64 --decode > ~/aegis-platform-api-ca.crt

# 5. Test TLS connectivity
openssl s_client -connect platform-api-grpc.aegist.dev:8081 \
  -CAfile ~/aegis-platform-api-ca.crt -servername platform-api-grpc.aegist.dev
```

### Keycloak Issues

```bash
# Check Keycloak pod logs
kubectl logs -n aegis-system -l app.kubernetes.io/component=keycloak --tail=50

# Check Keycloak PostgreSQL pod
kubectl logs -n aegis-system -l app.kubernetes.io/component=keycloak-postgres --tail=50

# Verify Keycloak secrets exist
kubectl get secret keycloak-admin-secret -n aegis-system
kubectl get secret keycloak-db-secret -n aegis-system
kubectl get secret keycloak-backstage-client-secret -n aegis-system

# Port-forward to access Keycloak admin console locally
kubectl port-forward svc/aegis-keycloak-service -n aegis-system 8443:8443
# Then open https://localhost:8443 in a browser
```

---

## Quick Reference

### Service Endpoints

| Service | Protocol | Default Endpoint |
|---------|----------|------------------|
| Platform API (gRPC) | gRPC/TLS | `platform-api-grpc.aegist.dev:8081` |
| Platform API (HTTP) | HTTPS | `platform-api.aegist.dev:8080` |
| Proxy (WebSocket) | WSS | `proxy.aegist.dev:8080` |
| Keycloak (OIDC) | HTTPS | `aegis-keycloak-service.aegis-system.svc.cluster.local:8443` |

### Useful Commands

```bash
# View all resources
kubectl get all -n aegis-system

# View pod logs
kubectl logs deployment/aegis-platform-api -n aegis-system -f
kubectl logs deployment/aegis-proxy -n aegis-system -f
kubectl logs deployment/aegis-spoke-aegis-spoke-k8s-agent -n aegis-system -f

# Restart a deployment
kubectl rollout restart deployment/aegis-platform-api -n aegis-system

# Dry-run to see merged Helm values
helm install --dry-run --debug aegis ./aegis-services \
  -f ./aegis-services/values/common.yaml \
  -f ./aegis-services/values/cloud.yaml \
  -f ./aegis-services/values-cloud-generated.yaml \
  -f overrides.yaml \
  --namespace aegis-system

# Check TLS status
kubectl get deployment aegis-platform-api -n aegis-system \
  -o jsonpath='{.spec.template.spec.containers[0].env[?(@.name=="GRPC_TLS_ENABLED")].value}'

# Retrieve credentials from AWS Secrets Manager
aws secretsmanager get-secret-value \
  --secret-id aegis/prod/db-password \
  --region us-east-1 \
  --query SecretString --output text
```

### Helm Chart Reference

| Chart | Version | Description |
|-------|---------|-------------|
| `aegis-services` | 0.1.0 | Hub control plane (platform-api + proxy + Keycloak) |
| `aegis-spoke` | 0.1.1 | Spoke agent (k8s-agent) |

### Security Notes

- All pod-to-pod communication uses TLS via internal PKI (step-ca + cert-manager)
- TLS minimum version: 1.2 with FIPS 140-2 approved cipher suites
- Pods run as non-root (UID 10000) with read-only root filesystems
- Network policies enforce default-deny with explicit allow rules
- Audit logging is set to `strict` enforcement mode (operations fail if audit logging fails)
- FedRAMP Moderate baseline controls are configured in `common.yaml`
- Secret rotation is configured on a monthly schedule for JWT keys
