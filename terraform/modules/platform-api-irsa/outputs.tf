output "role_arn" {
  description = "ARN of the IAM role for platform-api IRSA. Pass this to Helm as platformApi.irsaRoleArn."
  value       = aws_iam_role.platform_api.arn
}
