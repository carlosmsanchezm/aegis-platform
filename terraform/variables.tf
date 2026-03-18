# Variables for Aegis Infrastructure

variable "aws_region" {
  description = "AWS region for resources"
  type        = string
  default     = "us-east-1"
}

variable "aws_profile" {
  description = "AWS profile name for AWS CLI (leave blank when using environment-based auth)"
  type        = string
  default     = ""
}

variable "environment" {
  description = "Environment name (dev, staging, prod)"
  type        = string
  default     = "prod"
}

variable "cluster_name_prefix" {
  description = "Prefix for EKS cluster name"
  type        = string
  default     = "aegis-hub"
}

variable "helm_release_name" {
  description = "Helm release name for aegis-services (used to derive K8s service names)"
  type        = string
  default     = "aegis"
}

variable "k8s_namespace" {
  description = "Kubernetes namespace for aegis-services deployment"
  type        = string
  default     = "aegis-system"
}

variable "cluster_version" {
  description = "EKS cluster version"
  type        = string
  default     = "1.33"
}

# VPC Configuration
variable "vpc_cidr" {
  description = "CIDR block for VPC"
  type        = string
  default     = "10.0.0.0/16"
}

variable "availability_zones" {
  description = "Availability zones to use"
  type        = list(string)
  default     = ["us-east-1a", "us-east-1b", "us-east-1c"]
}

# EKS Node Groups
variable "cpu_instance_type" {
  description = "Instance type for CPU workers"
  type        = string
  default     = "t3.large"
}

variable "cpu_desired_capacity" {
  description = "Desired capacity for CPU worker nodes"
  type        = number
  default     = 2
}

# Database Configuration (in-cluster Postgres defaults; override for RDS)
variable "db_host" {
  description = "Database host. Leave empty for in-cluster Postgres (Helm managed)."
  type        = string
  default     = ""
}

variable "db_port" {
  description = "Database port"
  type        = string
  default     = "5432"
}

variable "db_name" {
  description = "Database name"
  type        = string
  default     = "aegis_platform"
}

variable "db_user" {
  description = "Database username"
  type        = string
  default     = "aegis_platform"
}

# ECR Configuration
variable "ecr_repositories" {
  description = "List of ECR repositories to create"
  type        = list(string)
  default     = ["aegis/k8s-agent", "aegis/proxy", "aegis/platform-api", "aegis/workspace-vscode", "aegis/vscode-reh-init", "aegis/ui"]
}

variable "manage_ecr_repositories" {
  description = "Whether Terraform should create and manage ECR repositories"
  type        = bool
  default     = false
}

# Secrets Configuration
variable "create_secrets" {
  description = "Create AWS Secrets Manager secrets"
  type        = bool
  default     = true
}

# DNS / Load Balancer Configuration
variable "platform_api_lb_hostname" {
  description = "Platform API LoadBalancer hostname (populated after K8s deployment)"
  type        = string
  default     = ""
}

variable "proxy_lb_hostname" {
  description = "Proxy LoadBalancer hostname (populated after K8s deployment)"
  type        = string
  default     = ""
}

variable "ui_lb_hostname" {
  description = "Aegis UI LoadBalancer hostname (populated after K8s deployment)"
  type        = string
  default     = ""
}

variable "keycloak_lb_hostname" {
  description = "Keycloak ingress LoadBalancer hostname (populated after K8s deployment)"
  type        = string
  default     = ""
}

# Cloudflare Configuration
variable "cloudflare_api_token" {
  description = "Cloudflare API token for DNS management (sourced from CLOUDFLARE_API_TOKEN env var)"
  type        = string
  sensitive   = true
  default     = ""
}

# Spoke project role (optional, for single-account dev/pilot testing)

variable "create_spoke_project_role" {
  description = "Create a spoke project IAM role in this account for dev/pilot testing"
  type        = bool
  default     = false
}

variable "spoke_project_id" {
  description = "Project ID for the spoke role (e.g. 'e2e-pilot-test'). Only used when create_spoke_project_role = true."
  type        = string
  default     = "e2e-pilot-test"
}

variable "hub_platform_api_role_arn" {
  description = "ARN of the hub account's platform-api IRSA role. Only used when create_spoke_project_role = true."
  type        = string
  default     = ""
}

variable "spoke_oidc_client_secret" {
  description = "OIDC client secret for spoke-agent (must match Keycloak realm config)"
  type        = string
  default     = "rEC99sBBWQAbRgg0xRQFBsMC8rt6pZOB"
  sensitive   = true
}

variable "default_project_id" {
  description = "Default project ID for the co-located spoke cluster. Used in the cluster ID ({project}-{region}-{env}) and project bootstrap."
  type        = string
  default     = "default"
}

variable "spoke_external_id" {
  description = "External ID for the spoke project role trust policy. Only used when create_spoke_project_role = true."
  type        = string
  default     = "aegis-pilot-2026"
}

# Spoke proxy NLB
variable "enable_spoke_proxy_nlb" {
  description = "Enable the spoke-proxy NLB for VS Code remote connections"
  type        = bool
  default     = true
}

variable "spoke_proxy_port" {
  description = "Port for spoke-proxy service (NodePort in EKS)"
  type        = number
  default     = 31484
}
