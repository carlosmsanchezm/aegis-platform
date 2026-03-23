# Aegis Infrastructure - AWS EKS + Cloud Hub
# Terraform configuration for complete Aegis cloud deployment

terraform {
  required_version = ">= 1.6"
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.4"
    }
    tls = {
      source  = "hashicorp/tls"
      version = "~> 4.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = "~> 4.0"
    }
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.25"
    }
  }

  backend "s3" {
    bucket         = "aegis-tf-state-195714074609"
    key            = "aegis/prod/terraform.tfstate"
    region         = "us-east-1"
    encrypt        = true
    dynamodb_table = "aegis-terraform-locks"
  }
}

provider "aws" {
  region  = var.aws_region
  profile = var.aws_profile

  default_tags {
    tags = {
      Project     = "aegis"
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}

provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

provider "kubernetes" {
  host                   = aws_eks_cluster.main.endpoint
  cluster_ca_certificate = base64decode(aws_eks_cluster.main.certificate_authority[0].data)
  exec {
    api_version = "client.authentication.k8s.io/v1beta1"
    command     = "aws"
    args = compact([
      "eks", "get-token",
      "--cluster-name", aws_eks_cluster.main.name,
      "--region", var.aws_region,
      var.aws_profile != "" ? "--profile" : "",
      var.aws_profile != "" ? var.aws_profile : "",
    ])
  }
}

# Generate random suffix for unique resource naming
resource "random_string" "suffix" {
  length  = 8
  special = false
  upper   = false
}

# Get current AWS account ID
data "aws_caller_identity" "current" {}

# Common locals used across modules
locals {
  cluster_name = "${var.cluster_name_prefix}-${var.environment}"
  common_tags = {
    Environment = var.environment
    Project     = "aegis"
    ManagedBy   = "terraform"
  }
}

################################################################################
# Secrets (DB password + JWT secret for proxy)
# These were previously in rds.tf; kept here for in-cluster Postgres deployments
################################################################################

resource "random_password" "db_password" {
  length           = 32
  special          = true
  override_special = "!#$%&*()-_=+[]{}<>"

  keepers = {
    reset = "2026-03-08"
  }
}

resource "aws_secretsmanager_secret" "db_password" {
  count = var.create_secrets ? 1 : 0

  name                    = "aegis/${var.environment}/db-password"
  description             = "PostgreSQL database password for Aegis platform-api"
  recovery_window_in_days = 0

  tags = local.common_tags
}

resource "aws_secretsmanager_secret_version" "db_password" {
  count = var.create_secrets ? 1 : 0

  secret_id     = aws_secretsmanager_secret.db_password[0].id
  secret_string = random_password.db_password.result
}

resource "random_password" "jwt_secret" {
  length  = 64
  special = false
}

resource "aws_secretsmanager_secret" "jwt_secret" {
  count = var.create_secrets ? 1 : 0

  name                    = "aegis/${var.environment}/proxy-jwt-secret"
  description             = "JWT secret for Aegis proxy service"
  recovery_window_in_days = 0

  tags = local.common_tags
}

resource "aws_secretsmanager_secret_version" "jwt_secret" {
  count = var.create_secrets ? 1 : 0

  secret_id     = aws_secretsmanager_secret.jwt_secret[0].id
  secret_string = random_password.jwt_secret.result
}

################################################################################
# Platform API IRSA (IAM Roles for Service Accounts)
# Lets platform-api assume spoke project roles via STS
################################################################################

module "platform_api_irsa" {
  source            = "./modules/platform-api-irsa"
  oidc_provider_arn = aws_iam_openid_connect_provider.cluster.arn
  oidc_issuer_url   = aws_eks_cluster.main.identity[0].oidc[0].issuer
  aws_account_id    = data.aws_caller_identity.current.account_id
}

################################################################################
# Spoke project role (optional, for single-account dev/pilot testing)
################################################################################

module "spoke_project_role" {
  count  = var.create_spoke_project_role ? 1 : 0
  source = "./modules/spoke-project-role"

  project_id                = var.spoke_project_id
  hub_platform_api_role_arn = var.hub_platform_api_role_arn
  external_id               = var.spoke_external_id
  tags                      = local.common_tags
}
