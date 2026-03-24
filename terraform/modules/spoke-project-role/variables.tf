variable "project_id" {
  description = "Aegis project identifier (used in IAM role name: aegis-project-{project_id})"
  type        = string

  validation {
    condition     = can(regex("^[a-z0-9][a-z0-9-]{1,61}[a-z0-9]$", var.project_id))
    error_message = "project_id must be lowercase alphanumeric with hyphens, 3-63 characters."
  }
}

variable "hub_platform_api_role_arn" {
  description = "ARN of the hub account's platform-api IRSA role (the principal allowed to assume this spoke role)"
  type        = string

  validation {
    condition     = can(regex("^arn:aws:iam::", var.hub_platform_api_role_arn))
    error_message = "hub_platform_api_role_arn must be a valid IAM role ARN."
  }
}

variable "external_id" {
  description = "External ID for the trust policy condition (prevents confused deputy attacks)"
  type        = string

  validation {
    condition     = length(var.external_id) >= 4
    error_message = "external_id must be at least 4 characters."
  }
}

variable "tags" {
  description = "Additional tags to apply to the IAM role"
  type        = map(string)
  default     = {}
}
