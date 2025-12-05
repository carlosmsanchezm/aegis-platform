variable "aws_region" {
  type        = string
  description = "AWS region where IAM resources will be created."
  default     = "us-east-1"
}

variable "target_role_arn" {
  type        = string
  description = "ARN of the IAM role that Pulumi should assume (e.g., arn:aws:iam::567751785679:role/aegis-platform)."
  default     = "arn:aws:iam::567751785679:role/aegis-platform"
}

variable "iam_user_name" {
  type        = string
  description = "Optional override for the IAM user name that will hold the Pulumi access key."
  default     = ""
}

variable "tags" {
  type        = map(string)
  description = "Additional tags to apply to created IAM resources."
  default     = {}
}
