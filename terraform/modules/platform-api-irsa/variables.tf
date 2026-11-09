variable "oidc_provider_arn" {
  description = "ARN of the OIDC provider for the EKS cluster"
  type        = string
}

variable "oidc_issuer_url" {
  description = "OIDC issuer URL of the EKS cluster (e.g., https://oidc.eks.us-east-1.amazonaws.com/id/EXAMPLE)"
  type        = string
}

variable "namespace" {
  description = "Kubernetes namespace where platform-api runs"
  type        = string
  default     = "aegis-system"
}

variable "service_account_name" {
  description = "Name of the platform-api ServiceAccount"
  type        = string
  default     = "aegis-platform-api"
}

variable "role_name" {
  description = "Name for the IAM role"
  type        = string
  default     = "aegis-platform-api"
}

variable "aws_account_id" {
  description = "AWS account ID (used for S3 Pulumi state bucket ARN)"
  type        = string
}
