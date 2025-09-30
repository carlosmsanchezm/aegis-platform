# Terraform Outputs for Aegis Infrastructure

# EKS Cluster Information
output "cluster_name" {
  description = "Name of the EKS cluster"
  value       = module.eks.cluster_name
}

output "cluster_endpoint" {
  description = "Endpoint for EKS control plane"
  value       = module.eks.cluster_endpoint
}

output "cluster_security_group_id" {
  description = "Security group ID attached to the EKS cluster"
  value       = module.eks.cluster_security_group_id
}

output "cluster_iam_role_arn" {
  description = "IAM role ARN associated with EKS cluster"
  value       = module.eks.cluster_iam_role_arn
}

output "cluster_certificate_authority_data" {
  description = "Base64 encoded certificate data required to communicate with the cluster"
  value       = module.eks.cluster_certificate_authority_data
}

output "cluster_oidc_issuer_url" {
  description = "The URL on the EKS cluster OIDC Issuer"
  value       = module.eks.cluster_oidc_issuer_url
}

# Node Groups
output "node_groups" {
  description = "EKS node groups information"
  value = {
    cpu_workers = {
      node_group_id   = module.eks.eks_managed_node_groups["cpu_workers"].node_group_id
      node_group_arn  = module.eks.eks_managed_node_groups["cpu_workers"].node_group_arn
      autoscaling_group_names = module.eks.eks_managed_node_groups["cpu_workers"].node_group_autoscaling_group_names
    }
    gpu_workers = {
      node_group_id   = module.eks.eks_managed_node_groups["gpu_workers"].node_group_id
      node_group_arn  = module.eks.eks_managed_node_groups["gpu_workers"].node_group_arn
      autoscaling_group_names = module.eks.eks_managed_node_groups["gpu_workers"].node_group_autoscaling_group_names
    }
  }
}

# VPC Information
output "vpc_id" {
  description = "VPC ID"
  value       = module.vpc.vpc_id
}

output "vpc_cidr_block" {
  description = "VPC CIDR block"
  value       = module.vpc.vpc_cidr_block
}

output "private_subnet_ids" {
  description = "Private subnet IDs"
  value       = module.vpc.private_subnets
}

output "public_subnet_ids" {
  description = "Public subnet IDs"
  value       = module.vpc.public_subnets
}

output "database_subnet_ids" {
  description = "Database subnet IDs"
  value       = module.vpc.database_subnets
}

# RDS Information
output "rds_endpoint" {
  description = "RDS instance endpoint"
  value       = module.rds.db_instance_endpoint
  sensitive   = false
}

output "rds_port" {
  description = "RDS instance port"
  value       = module.rds.db_instance_port
}

output "rds_database_name" {
  description = "Name of the database"
  value       = module.rds.db_instance_name
}

output "rds_username" {
  description = "Master username for the database"
  value       = module.rds.db_instance_username
  sensitive   = true
}

output "rds_arn" {
  description = "RDS instance ARN"
  value       = module.rds.db_instance_arn
}

# Secrets Manager
output "secrets" {
  description = "AWS Secrets Manager secret information"
  value = var.create_secrets ? {
    db_password_secret_arn = aws_secretsmanager_secret.db_password[0].arn
    db_password_secret_name = aws_secretsmanager_secret.db_password[0].name
    jwt_secret_secret_arn  = aws_secretsmanager_secret.jwt_secret[0].arn
    jwt_secret_secret_name = aws_secretsmanager_secret.jwt_secret[0].name
  } : null
}

# ECR Repositories
output "ecr_repositories" {
  description = "ECR repository URLs"
  value       = local.ecr_repositories
}

output "ecr_registry_url" {
  description = "ECR registry URL"
  value       = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com"
}

# Helper outputs for Helm values
output "helm_values" {
  description = "Generated values for Helm chart deployment"
  sensitive   = true
  value = {
    # Platform API configuration
    platform_api = {
      image_repository = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/aegis/platform-api"
      db_host         = split(":", module.rds.db_instance_endpoint)[0]
      db_port         = module.rds.db_instance_port
      db_name         = module.rds.db_instance_name
      db_user         = module.rds.db_instance_username
    }

    # Proxy configuration
    proxy = {
      image_repository = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/aegis/proxy"
    }

    # K8s Agent configuration
    k8s_agent = {
      image_repository = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/aegis/k8s-agent"
      cluster_id      = "aws-${var.aws_region}-${var.environment}"
    }

    # Secrets configuration
    secrets = var.create_secrets ? {
      db_password_secret_name = aws_secretsmanager_secret.db_password[0].name
      jwt_secret_secret_name  = aws_secretsmanager_secret.jwt_secret[0].name
    } : null
  }
}

# AWS Account Information
data "aws_caller_identity" "current" {}

output "aws_account_id" {
  description = "AWS Account ID"
  value       = data.aws_caller_identity.current.account_id
}

output "aws_region" {
  description = "AWS Region"
  value       = var.aws_region
}

# Kubectl configuration command
output "kubectl_config_command" {
  description = "Command to configure kubectl"
  value       = "aws eks update-kubeconfig --region ${var.aws_region} --name ${module.eks.cluster_name} --profile ${var.aws_profile}"
}