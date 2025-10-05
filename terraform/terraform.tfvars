# Terraform variables for Aegis infrastructure
# AWS Configuration
aws_region  = "us-east-1"
aws_profile = "myclaude"

# Environment
environment = "prod"

# EKS Configuration
cluster_name_prefix = "aegis-spoke"
cluster_version     = "1.32"

# Instance types and capacity (matching your script setup)
cpu_instance_type    = "t3.medium"
gpu_instance_type    = "g4dn.xlarge"
cpu_desired_capacity = 2
gpu_desired_capacity = 0 # Start with 0 for cost savings
gpu_max_capacity     = 2
use_spot_instances   = true # Cost optimization

# RDS Configuration
db_instance_class        = "db.t3.micro" # Start small
db_allocated_storage     = 20
db_max_allocated_storage = 100
db_postgres_version      = "15.12"
db_backup_retention      = 7
db_skip_final_snapshot   = false # Keep backups

# ECR Repositories
ecr_repositories = [
  "aegis/k8s-agent",
  "aegis/proxy",
  "aegis/platform-api",
  "aegis/workspace-vscode"
]

# Create secrets in AWS Secrets Manager
create_secrets = true