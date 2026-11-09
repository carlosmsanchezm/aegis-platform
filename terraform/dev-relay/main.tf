# AWS Development Relay for Aegis Platform
#
# Creates an EC2 instance + internal NLB that acts as a reverse SSH tunnel
# relay, allowing remote EKS spoke clusters to reach the local platform-api.
#
# Architecture:
#   Local Machine (platform-api :8081, keycloak :8443)
#         |
#         | SSH Reverse Tunnel (-R 8081 -R 8443)
#         v
#   EC2 Instance (relay, Elastic IP for stable SSH target)
#         |
#         | Internal NLB (TCP passthrough)
#         v
#   EKS Spoke Cluster (k8s-agent connects to NLB:8081 for gRPC, NLB:8443 for OIDC)
#
# Usage:
#   cd terraform/dev-relay
#   terraform init
#   terraform apply
#
#   # Then start the SSH tunnel:
#   ../../scripts/start-aws-tunnel.sh

terraform {
  required_version = ">= 1.6.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.60"
    }
    local = {
      source  = "hashicorp/local"
      version = "~> 2.5"
    }
  }
}

provider "aws" {
  region  = var.aws_region
  profile = var.aws_profile
}
