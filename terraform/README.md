# Aegis Terraform Infrastructure

This directory contains Terraform configurations for deploying the Aegis hub cluster to AWS EKS.

**Authoritative for:** infrastructure scope, Terraform variables/outputs, and infra bootstrap.

**Not authoritative for:** application deployment workflow, deploy-script steps, or production rollout sequencing.

**See also:**
- `../AGENT_DEPLOYMENT_GUIDE.md` -- standard deployment workflow after infra apply
- `../docs/PRODUCTION_DEPLOYMENT.md` -- production-only overlays
- `../docs/security/auth.md` -- current cloud Backstage/Keycloak auth contract

## What Gets Created

- **EKS Cluster** (`aegis-hub-prod` by default, configurable via `cluster_name_prefix`)
- **VPC** with public, private, and database subnets across 3 AZs
- **IRSA** (IAM Roles for Service Accounts) for platform-api
- **Cloudflare DNS** records for `aegis-platform.tech` (see `cloudflare.tf`)
- **AWS Secrets Manager** secrets (DB password, JWT secret via `random_password`)
- **ECR Repositories** (optional — set `manage_ecr_repositories = true`)
- **Spoke proxy NLB** (optional — set `enable_spoke_proxy_nlb = true`)

**No RDS** is created. The minimal cloud hub uses in-cluster Postgres. For full production with RDS, see `docs/PRODUCTION_DEPLOYMENT.md`.

## Quick Start

```bash
cd terraform

# 1. Initialize (with optional remote state backend)
terraform init -backend-config=backend.hcl

# 2. Deploy infrastructure
terraform apply

# 3. Configure kubectl
eval "$(terraform output -raw kubectl_config_command)"

# 4. Hand off to the deployment guide
cd ..
# Follow AGENT_DEPLOYMENT_GUIDE.md § Cloud Hub Deployment
```

For the full deployment walkthrough, see [AGENT_DEPLOYMENT_GUIDE.md](../AGENT_DEPLOYMENT_GUIDE.md) § Cloud Hub Deployment.

## Key Variables

| Variable | Default | Description |
|---|---|---|
| `cluster_name_prefix` | `aegis-hub` | EKS cluster name prefix (cluster = `{prefix}-prod`) |
| `cpu_instance_type` | `t3.medium` | EC2 instance type for CPU node group |
| `cloudflare_api_token` | (empty) | Cloudflare API token for DNS management |
| `platform_api_lb_hostname` | (empty) | Platform API LoadBalancer hostname (set after deploy) |
| `proxy_lb_hostname` | (empty) | Proxy LoadBalancer hostname (set after deploy) |
| `manage_ecr_repositories` | `false` | Whether Terraform manages ECR repos |
| `enable_spoke_proxy_nlb` | `false` | Whether to create spoke proxy NLB |
| `aws_region` | `us-east-1` | AWS region |
| `aws_profile` | (empty) | AWS CLI profile name |

## Key Outputs

```bash
# Cluster
terraform output -raw cluster_name
terraform output -raw kubectl_config_command
terraform output -raw platform_api_irsa_role_arn

# Credentials (from random_password resources)
terraform output -raw db_password_secret_value
terraform output -raw jwt_secret_value

# Helm values (auto-generated for charts)
terraform output -raw helm_values_aegis_services
terraform output -raw helm_values_aegis_spoke

# DNS
terraform output -raw dns_platform_api
terraform output -raw dns_keycloak
terraform output -raw dns_proxy
terraform output -raw dns_ui

# ECR
terraform output -raw ecr_registry_url
```

## Canonical Public DNS

The deploy workflow updates the canonical Cloudflare records after the public load balancers are ready:

- `platform-api.aegis-platform.tech`
- `keycloak.aegis-platform.tech`
- `proxy.aegis-platform.tech`
- `ui.aegis-platform.tech`

Current topology:

- `platform-api` and `proxy` point at their direct public service load balancers
- `ui` and `keycloak` point at the shared public ingress load balancer

See [AGENT_DEPLOYMENT_GUIDE.md](../AGENT_DEPLOYMENT_GUIDE.md) § Cloud Hub Deployment for the operational flow.

## Remote State (One-Time Setup)

1. Provision an S3 bucket and optional DynamoDB table for locking
2. Copy `backend.hcl.example` to `backend.hcl` and update values
3. Initialize: `terraform init -backend-config=backend.hcl`

## Credentials

```bash
# Quick command to show all credentials
./show-credentials.sh

# Or individually
terraform output -raw db_password_secret_value
terraform output -raw jwt_secret_value
terraform output -json database_connection_info | jq
```

## Teardown

See [DESTROY_CHECKLIST.md](./DESTROY_CHECKLIST.md) for the full teardown procedure.
