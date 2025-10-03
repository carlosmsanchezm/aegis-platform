# Aegis Terraform Infrastructure

This directory contains Terraform configurations for deploying the Aegis platform to AWS EKS.

## 📚 Documentation

- **[DEPLOYMENT.md](./DEPLOYMENT.md)** - Complete step-by-step deployment guide
- **[QUICKREF.md](./QUICKREF.md)** - Quick reference for common commands

## 🚀 Quick Start

```bash
# 1. Deploy infrastructure
terraform init
terraform apply

# 2. Get credentials
./show-credentials.sh

# 3. Deploy applications
./generate-helm-values.sh
```

## 🔐 Accessing Credentials

### Quick Command
```bash
./show-credentials.sh
```

### Individual Values
```bash
# Database password
terraform output -raw db_password_secret_value

# Complete database info (JSON)
terraform output -json database_connection_info | jq

# JWT secret for proxy
terraform output -raw jwt_secret_value
```

## 📝 For Next Time

When you need credentials for deployment:

1. **Go to terraform directory**: `cd terraform`
2. **Run**: `./show-credentials.sh`
3. **Copy the database password** from the output
4. **Use in Helm deployment** (handled automatically by `generate-helm-values.sh`)

The password and all connection info are stored in Terraform state and can be retrieved anytime with the commands above.

## 📦 What Gets Created

- **EKS Cluster** (`aegis-spoke-prod`)
- **RDS PostgreSQL** with auto-generated password
- **VPC & Networking**
- **ECR Repositories**
- **AWS Secrets Manager** secrets

See [DEPLOYMENT.md](./DEPLOYMENT.md) for complete details.
