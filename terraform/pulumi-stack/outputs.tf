output "pulumi_iam_user_name" {
  description = "Name of the IAM user provisioned for Pulumi."
  value       = aws_iam_user.pulumi.name
}

output "pulumi_access_key_id" {
  description = "AWS access key ID for the Pulumi IAM user."
  value       = aws_iam_access_key.pulumi.id
}

output "pulumi_secret_access_key" {
  description = "AWS secret access key for the Pulumi IAM user."
  value       = aws_iam_access_key.pulumi.secret
  sensitive   = true
}
