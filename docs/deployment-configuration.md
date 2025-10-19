# Aegis Deployment Configuration Guide

This document outlines the configuration requirements for deploying Aegis components in different environments.

## Environment Configurations

### Local Development

For local development using Docker Desktop or Kind:

```yaml
# charts/aegis-spoke/values.yaml (default values)
k8sAgent:
  env:
    AEGIS_CLUSTER_ID: "dev-1"
    AEGIS_CP_GRPC: "host.docker.internal:8081"  # Platform API running locally
    AEGIS_PROXY_INGRESS_HOST: "proxy.localtest.me"

proxy:
  ingress:
    hostname: "proxy.localtest.me"
  tls:
    terminateAtIngress: false  # Run TLS directly on proxy for local dev
```

### AWS Cloud Deployment

For production AWS EKS deployment:

```yaml
# charts/aegis-spoke/values-cloud.yaml
k8sAgent:
  image:
    repository: 567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/k8s-agent
    tag: "amd64"
  env:
    AEGIS_CLUSTER_ID: "aws-us-east-1-prod"
    AEGIS_CP_GRPC: "YOUR_PLATFORM_API_ENDPOINT:8081"
    AEGIS_PROXY_INGRESS_HOST: "a9ada300f6ee041f9be0f1ce0b8a759b-1516746039.us-east-1.elb.amazonaws.com"

proxy:
  image:
    repository: 567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/proxy
    tag: "amd64"
  ingress:
    hostname: "a9ada300f6ee041f9be0f1ce0b8a759b-1516746039.us-east-1.elb.amazonaws.com"
  tls:
    terminateAtIngress: true  # Use NGINX ingress for TLS termination
```

## Infrastructure Details

### Current AWS Deployment

- **EKS Cluster**: `aegis-spoke-prod`
- **Region**: `us-east-1`
- **Account ID**: `567751785679`
- **Load Balancer**: `a9ada300f6ee041f9be0f1ce0b8a759b-1516746039.us-east-1.elb.amazonaws.com`
- **ECR Repositories**:
  - `567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/k8s-agent:amd64`
  - `567751785679.dkr.ecr.us-east-1.amazonaws.com/aegis/proxy:amd64`

## Platform API Deployment Options

### Current State: Local Development
The platform-api is currently running locally on your development machine. For cloud deployment, you have several options:

### Option 1: Deploy Platform API to AWS EKS
**Recommended for production**

1. Create a platform-api Helm chart
2. Deploy to the same EKS cluster or a dedicated management cluster
3. Use internal service discovery (e.g., `platform-api.aegis-platform.svc.cluster.local:8081`)
4. Benefits: Better security, service mesh integration, easier scaling

### Option 2: Deploy Platform API to AWS ECS/Fargate
**Good for simpler deployment**

1. Package platform-api as a Docker container
2. Deploy to ECS with Application Load Balancer
3. Use ALB hostname for AEGIS_CP_GRPC
4. Benefits: Simpler infrastructure, managed container platform

### Option 3: Keep Platform API Local (Current)
**Development/testing only**

1. Use ngrok or similar to expose local platform-api
2. Update AEGIS_CP_GRPC to the ngrok URL
3. Benefits: Easy for development, no additional infrastructure

## Required Environment Variables

### K8s-Agent Configuration

| Variable | Local Development | AWS Cloud | Description |
|----------|------------------|-----------|-------------|
| `AEGIS_CLUSTER_ID` | `dev-1` | `aws-us-east-1-prod` | Unique identifier for this cluster |
| `AEGIS_CP_GRPC` | `host.docker.internal:8081` | `YOUR_PLATFORM_API_ENDPOINT:8081` | Platform API gRPC endpoint |
| `AEGIS_PROXY_INGRESS_HOST` | `proxy.localtest.me` | `{LoadBalancer-Hostname}` | Public hostname for proxy access |

### Proxy Configuration

| Variable | Local Development | AWS Cloud | Description |
|----------|------------------|-----------|-------------|
| `AEGIS_PROXY_JWT_SECRET` | `a-very-secret-key-for-local-dev-must-be-32-chars` | `{SECURE_SECRET}` | JWT signing secret |
| `AEGIS_PROXY_TLS_CERT` | `/etc/aegis-proxy/tls.crt` | `/etc/aegis-proxy/tls.crt` | TLS certificate path |
| `AEGIS_PROXY_TLS_KEY` | `/etc/aegis-proxy/tls.key` | `/etc/aegis-proxy/tls.key` | TLS private key path |

## Deployment Commands

### Local Development
```bash
helm install aegis-spoke charts/aegis-spoke -n aegis-spoke --create-namespace
```

### AWS Cloud Deployment

**🎉 NEW: Automated Infrastructure Detection**

The deploy script now automatically detects and configures all infrastructure values:

```bash
# Basic deployment (uses placeholder for platform-api)
./scripts/deploy.sh

# Deployment with custom platform-api endpoint
PLATFORM_API_ENDPOINT="platform-api.example.com:8081" ./scripts/deploy.sh

# With custom AWS profile
AWS_PROFILE="production" ./scripts/deploy.sh
```

The script automatically:
- ✅ Gets the EKS load balancer hostname
- ✅ Updates `charts/aegis-spoke/values-cloud.yaml` with correct ECR repositories
- ✅ Sets the cluster ID based on region (e.g., `aws-us-east-1-prod`)
- ✅ Configures proxy ingress hostname
- ✅ Handles TLS certificate generation and mounting

**Manual Deployment (if needed):**
```bash
# Update AEGIS_CP_GRPC in values-cloud.yaml first!
helm install aegis-spoke-aws charts/aegis-spoke \
    -n aegis-spoke --create-namespace \
    -f charts/aegis-spoke/values-cloud.yaml \
    --set-file proxy.tls.cert=/tmp/new-tls.crt \
    --set-file proxy.tls.key=/tmp/new-tls.key
```

## Troubleshooting

### Common Issues

1. **K8s-Agent CrashLoopBackOff**: Usually due to incorrect AEGIS_CP_GRPC endpoint
2. **Proxy TLS Errors**: Certificate and key don't match - regenerate with matching pairs
3. **ImagePullBackOff**: Images not available in ECR - run build and push steps
4. **DNS Resolution**: Load balancer hostname too long for certificate CN - use shorter hostname

### Verification Commands

```bash
# Check pod status
kubectl get pods -n aegis-spoke

# Check k8s-agent logs
kubectl logs -n aegis-spoke -l app.kubernetes.io/component=k8s-agent

# Check proxy logs
kubectl logs -n aegis-spoke -l app.kubernetes.io/component=proxy

# Get load balancer hostname
kubectl get service ingress-nginx-controller -n kube-system \
    -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'
```

## Next Steps

1. **Deploy Platform API to Cloud**: Choose deployment option and implement
2. **Update K8s-Agent Configuration**: Set correct AEGIS_CP_GRPC endpoint
3. **Test End-to-End Flow**: Deploy a GPU workload through the platform API
4. **Production Hardening**:
   - Use proper secrets management (AWS Secrets Manager)
   - Implement monitoring and logging
   - Set up proper RBAC and network policies
   - Configure backup and disaster recovery