# Aegis Terraform Infrastructure

This Terraform configuration creates the complete AWS infrastructure for the Aegis platform, including:

- **EKS Cluster** with CPU and GPU node groups
- **RDS PostgreSQL** database for persistent storage
- **VPC** with public, private, and database subnets
- **ECR Repositories** for container images
- **Security Groups** and networking
- **AWS Secrets Manager** secrets for database and JWT

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     AWS VPC (10.0.0.0/16)                  │
├─────────────────────────────────────────────────────────────┤
│ Public Subnets          │ Private Subnets                   │
│ ┌─────────────────────┐ │ ┌─────────────────────────────────┐ │
│ │   Load Balancer     │ │ │         EKS Cluster             │ │
│ │   (NGINX Ingress)   │ │ │  ┌─────────────────────────────┐ │ │
│ └─────────────────────┘ │ │  │     CPU Workers (t3.medium) │ │ │
│                         │ │  │     GPU Workers (g4dn.xlarge)│ │ │
│                         │ │  └─────────────────────────────┘ │ │
│                         │ └─────────────────────────────────┘ │
│                         │                                     │
│                         │ Database Subnets                   │
│                         │ ┌─────────────────────────────────┐ │
│                         │ │    RDS PostgreSQL               │ │
│                         │ │    (Multi-AZ)                   │ │
│                         │ └─────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

## 📋 Prerequisites

1. **AWS CLI** configured with appropriate credentials
2. **Terraform** >= 1.6 installed
3. **kubectl** installed for cluster management
4. **AWS Account** with appropriate permissions

## 🚀 Quick Start

### 1. Configure Variables

```bash
# Copy the example variables file
cp terraform.tfvars.example terraform.tfvars

# Edit with your specific values
vim terraform.tfvars
```

Key variables to set:
- `aws_profile`: Your AWS CLI profile
- `aws_region`: Target AWS region
- `environment`: Environment name (dev/staging/prod)

### 2. Initialize and Deploy

```bash
# Initialize Terraform
terraform init

# Review the plan
terraform plan

# Deploy infrastructure
terraform apply
```

### 3. Configure kubectl

```bash
# Get the kubectl configuration command from Terraform output
terraform output kubectl_config_command

# Run the command (example)
aws eks update-kubeconfig --region us-east-1 --name aegis-spoke-prod --profile myclaude
```

### 4. Install NGINX Ingress Controller

```bash
# Add Helm repo
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update

# Install NGINX Ingress
helm install ingress-nginx ingress-nginx/ingress-nginx -n kube-system
```

### 5. Deploy Aegis Services

```bash
# Get Helm values from Terraform output
terraform output -json helm_values > helm_values.json

# Create namespace
kubectl create namespace aegis-services

# Deploy with Helm (you'll need to create a values file from the output)
helm install aegis-services ../charts/aegis-services \
  -n aegis-services \
  -f aegis-services-values.yaml \
  --set-string platformApi.secrets.db-password="$(terraform output -raw db_password)" \
  --set-string proxy.jwtSecret="$(terraform output -raw jwt_secret)"
```

## 📊 Terraform Outputs

This configuration provides comprehensive outputs for integration:

### Cluster Information
- `cluster_name`: EKS cluster name
- `cluster_endpoint`: Kubernetes API endpoint
- `kubectl_config_command`: Command to configure kubectl

### Database Information
- `rds_endpoint`: PostgreSQL database endpoint
- `rds_port`: Database port (5432)
- `rds_database_name`: Database name

### Container Registries
- `ecr_repositories`: Map of ECR repository URLs
- `ecr_registry_url`: Base ECR registry URL

### Helper Values
- `helm_values`: Pre-formatted values for Helm charts

## 💰 Cost Optimization

The configuration includes several cost optimizations:

1. **Spot Instances**: Uses spot instances by default (`use_spot_instances = true`)
2. **GPU Scaling**: GPU nodes start at 0 capacity
3. **Single NAT Gateway**: Uses one NAT gateway instead of one per AZ
4. **Small Database**: Starts with `db.t3.micro` instance

### Estimated Costs (us-east-1):
- **EKS Control Plane**: ~$72/month
- **t3.medium (CPU worker)**: ~$30/month (spot: ~$15/month)
- **g4dn.xlarge (GPU worker)**: ~$380/month when running (spot: ~$190/month)
- **db.t3.micro RDS**: ~$15/month
- **NAT Gateway**: ~$45/month
- **Load Balancer**: ~$22/month

**Total (without GPU)**: ~$150/month
**With GPU running**: ~$530/month (spot: ~$340/month)

## 🔧 Configuration

### Environment-Specific Configurations

**Development:**
```hcl
environment = "dev"
db_instance_class = "db.t3.micro"
cpu_desired_capacity = 1
gpu_desired_capacity = 0
db_skip_final_snapshot = true
```

**Production:**
```hcl
environment = "prod"
db_instance_class = "db.t3.small"  # More resources
cpu_desired_capacity = 2
gpu_desired_capacity = 1
db_skip_final_snapshot = false  # Keep backups
```

### Scaling GPU Workers

To scale GPU workers up/down:

```bash
# Scale up for workloads
aws eks update-nodegroup-config \
  --cluster-name aegis-spoke-prod \
  --nodegroup-name gpu-workers-g4 \
  --scaling-config minSize=0,maxSize=2,desiredSize=1

# Scale down to save costs
aws eks update-nodegroup-config \
  --cluster-name aegis-spoke-prod \
  --nodegroup-name gpu-workers-g4 \
  --scaling-config minSize=0,maxSize=2,desiredSize=0
```

## 🔐 Security

### Secrets Management
- Database passwords stored in AWS Secrets Manager
- JWT secrets automatically generated and stored securely
- No sensitive data in Terraform state (except password hashes)

### Network Security
- Database in private subnets only
- Security groups restrict access between services
- EKS uses private subnets for worker nodes

### Access Control
- EKS RBAC integration with AWS IAM
- Minimal required permissions for node groups
- Database access restricted to EKS nodes only

## 🧹 Cleanup

To destroy all resources:

```bash
# WARNING: This will delete everything including data!
terraform destroy
```

**Note**: If `db_skip_final_snapshot = false`, RDS will create a final snapshot before deletion.

## 📝 Notes

1. **First Run**: Initial deployment takes ~20-30 minutes
2. **GPU Drivers**: The EKS GPU AMI includes NVIDIA drivers
3. **Database Migrations**: Run database migrations after deployment
4. **DNS**: You'll need to configure DNS to point to the load balancer
5. **Certificates**: Consider using cert-manager for automatic TLS certificates

## 🆘 Troubleshooting

### Common Issues

**EKS Authentication:**
```bash
# Update kubeconfig
aws eks update-kubeconfig --region us-east-1 --name aegis-spoke-prod --profile myclaude

# Verify access
kubectl get nodes
```

**Database Connection:**
```bash
# Test from EKS cluster
kubectl run -it --rm debug --image=postgres:15 -- psql -h <rds-endpoint> -U aegis_api -d aegis
```

**ECR Authentication:**
```bash
# Login to ECR
aws ecr get-login-password --region us-east-1 --profile myclaude | docker login --username AWS --password-stdin <account-id>.dkr.ecr.us-east-1.amazonaws.com
```