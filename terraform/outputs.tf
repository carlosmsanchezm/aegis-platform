# Terraform Outputs for Aegis Infrastructure

################################################################################
# EKS Cluster Information
################################################################################

output "cluster_name" {
  description = "Name of the EKS cluster"
  value       = aws_eks_cluster.main.name
}

output "node_security_group_id" {
  description = "Primary security group used by EKS worker nodes"
  value       = aws_security_group.node.id
}

output "eks_cluster_security_group_id" {
  description = "AWS-managed cluster security group created by EKS"
  value       = aws_eks_cluster.main.vpc_config[0].cluster_security_group_id
}

output "node_groups" {
  description = "EKS node groups information"
  value = {
    cpu_workers = {
      node_group_id  = aws_eks_node_group.cpu_workers.id
      node_group_arn = aws_eks_node_group.cpu_workers.arn
    }
  }
}

output "platform_api_irsa_role_arn" {
  description = "ARN of the IRSA role for platform-api (pass to Helm as platformApi.irsaRoleArn)"
  value       = module.platform_api_irsa.role_arn
}

################################################################################
# VPC Information
################################################################################

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

################################################################################
# ECR Information
################################################################################

output "ecr_repository_urls" {
  description = "Map of ECR repository names to URLs"
  value       = var.manage_ecr_repositories ? { for name, repo in module.ecr : name => repo.repository_url } : {}
}

output "ecr_registry_url" {
  description = "ECR registry URL"
  value       = "${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com"
}

################################################################################
# Database stub outputs (in-cluster Postgres — no RDS)
# These satisfy generate-cloud-deployment.sh expectations.
# Set SKIP_MIGRATION_PLACEHOLDER=1 when using in-cluster Postgres.
################################################################################

output "rds_endpoint" {
  description = "Database endpoint (in-cluster Postgres service)"
  value       = var.db_host != "" ? "${var.db_host}:${var.db_port}" : ""
}

output "rds_port" {
  description = "Database port"
  value       = var.db_port
}

output "rds_database_name" {
  description = "Database name"
  value       = var.db_name
}

output "rds_username" {
  description = "Database username"
  value       = var.db_user
  sensitive   = true
}

################################################################################
# Secrets
################################################################################

output "db_password_secret_value" {
  description = "Database password"
  value       = nonsensitive(random_password.db_password.result)
  sensitive   = true
}

output "jwt_secret_value" {
  description = "JWT secret for proxy"
  value       = nonsensitive(random_password.jwt_secret.result)
  sensitive   = true
}

output "spoke_oidc_client_secret" {
  description = "OIDC client secret for spoke-agent"
  value       = var.spoke_oidc_client_secret
  sensitive   = true
}

output "secrets" {
  description = "AWS Secrets Manager secret information"
  value = var.create_secrets ? {
    db_password_secret_arn  = aws_secretsmanager_secret.db_password[0].arn
    db_password_secret_name = aws_secretsmanager_secret.db_password[0].name
    jwt_secret_secret_arn   = aws_secretsmanager_secret.jwt_secret[0].arn
    jwt_secret_secret_name  = aws_secretsmanager_secret.jwt_secret[0].name
  } : null
}

output "k8s_secret_commands" {
  description = "Commands to create Kubernetes secrets from Terraform outputs"
  value       = <<-EOT
  # Create secrets in Kubernetes from Terraform outputs:
  kubectl create secret generic aegis-platform-secrets \
    --from-literal=db-password="$(terraform output -raw db_password_secret_value)" \
    --from-literal=proxy-jwt-secret="$(terraform output -raw jwt_secret_value)" \
    --namespace aegis-system
  EOT
}

################################################################################
# Helm Values (generated for aegis-services and aegis-spoke charts)
################################################################################

output "helm_values_aegis_services" {
  description = "Ready-to-use values for aegis-services Helm chart (control plane)"
  sensitive   = true
  value       = <<-EOT
  # Auto-generated from Terraform - aegis-services (Control Plane / Hub)
  # Cluster: ${aws_eks_cluster.main.name}
  # Region: ${var.aws_region}
  # Generated: ${timestamp()}

  platformApi:
    image:
      repository: ${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/aegis/platform-api
    serviceAccount:
      annotations:
        eks.amazonaws.com/role-arn: "${module.platform_api_irsa.role_arn}"

  proxy:
    image:
      repository: ${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/aegis/proxy
  EOT
}

output "helm_values_aegis_spoke" {
  description = "Ready-to-use values for aegis-spoke Helm chart (workload cluster)"
  sensitive   = true
  value       = <<-EOT
  # Auto-generated from Terraform - aegis-spoke (Workload Cluster)
  # Cluster: ${aws_eks_cluster.main.name}
  # Region: ${var.aws_region}
  # Generated: ${timestamp()}

  k8sAgent:
    image:
      repository: ${data.aws_caller_identity.current.account_id}.dkr.ecr.${var.aws_region}.amazonaws.com/aegis/k8s-agent

    env:
      AEGIS_CP_GRPC: "${var.helm_release_name}-platform-api.${var.k8s_namespace}.svc.cluster.local:8081"
      AEGIS_PROXY_INGRESS_HOST: "${cloudflare_record.proxy.hostname}"
      AEGIS_CLUSTER_ID: "${var.default_project_id}-${var.aws_region}-${var.environment}"
      AEGIS_CP_OIDC_TOKEN_URL: "https://${var.helm_release_name}-keycloak-service.${var.k8s_namespace}.svc.cluster.local:8443/realms/aegis/protocol/openid-connect/token"
      AEGIS_CP_OIDC_CLIENT_ID: "spoke-agent"
      AEGIS_CP_OIDC_CLIENT_SECRET: "${var.spoke_oidc_client_secret}"
      AEGIS_CP_OIDC_AUDIENCE: "aegis-platform"

  proxy:
    enabled: false
  EOT
}

################################################################################
# AWS Account & Region
################################################################################

output "aws_account_id" {
  description = "AWS Account ID"
  value       = data.aws_caller_identity.current.account_id
}

output "aws_region" {
  description = "AWS Region"
  value       = var.aws_region
}

################################################################################
# kubectl Configuration
################################################################################

output "kubectl_config_command" {
  description = "Command to configure kubectl"
  value       = var.aws_profile != "" ? "aws eks update-kubeconfig --region ${var.aws_region} --name ${aws_eks_cluster.main.name} --profile ${var.aws_profile}" : "aws eks update-kubeconfig --region ${var.aws_region} --name ${aws_eks_cluster.main.name}"
}

################################################################################
# Public DNS
################################################################################

output "dns_platform_api" {
  description = "Public DNS hostname for platform-api"
  value       = cloudflare_record.platform_api.hostname
}

output "dns_keycloak" {
  description = "Public DNS hostname for Keycloak"
  value       = cloudflare_record.keycloak.hostname
}

output "dns_proxy" {
  description = "Public DNS hostname for proxy service"
  value       = cloudflare_record.proxy.hostname
}

output "dns_ui" {
  description = "Public DNS hostname for UI service"
  value       = cloudflare_record.ui.hostname
}
