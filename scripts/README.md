# Aegis AWS Deployment Scripts

Simple scripts to deploy and destroy Aegis infrastructure on AWS.

## Quick Start

### Deploy Everything
```bash
./scripts/deploy.sh
```
Creates:
- EKS cluster with GPU support
- ECR repositories
- Docker images
- NGINX Ingress
- Aegis spoke deployment
- Auto-scales GPU nodes to 0 for cost savings

**Time**: ~20 minutes
**Cost**: ~$0.15/hour (scales to ~$1.65/hour with GPU)

### Destroy Everything
```bash
./scripts/destroy.sh
```
Deletes:
- EKS cluster
- ECR repositories
- All AWS resources

**Time**: ~10 minutes
**Cost**: $0/hour

## Prerequisites

- AWS CLI configured with profile `myclaude`
- Docker running
- kubectl, helm, eksctl installed

## Configuration

The deployment will output connection details for your platform-api:
- Cluster ID: `aws-us-east-1-prod`
- Proxy URL: `http://<load-balancer-hostname>`

## Cost Management

- GPU nodes auto-scale from 0 to save money
- Will scale up automatically when workloads are scheduled
- Run `./scripts/destroy.sh` when not in use to avoid charges