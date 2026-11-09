output "role_arn" {
  description = "ARN of the spoke project IAM role. Use this as the Role ARN when creating the project in Aegis."
  value       = aws_iam_role.spoke_project.arn
}

output "role_name" {
  description = "Name of the spoke project IAM role."
  value       = aws_iam_role.spoke_project.name
}
