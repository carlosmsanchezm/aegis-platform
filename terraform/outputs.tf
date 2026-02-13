# Terraform Outputs for Aegis Infrastructure
# Only VPC and spoke-proxy NLB outputs - EKS/RDS managed by Pulumi

# VPC Information
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

# ECR Repository Information
output "ecr_repository_urls" {
  description = "Map of ECR repository names to URLs"
  value       = var.manage_ecr_repositories ? { for name, repo in module.ecr : name => repo.repository_url } : {}
}

# Route53 Information
output "route53_zone_id" {
  description = "Route53 hosted zone ID"
  value       = aws_route53_zone.aegist.zone_id
}

output "route53_name_servers" {
  description = "Route53 name servers"
  value       = aws_route53_zone.aegist.name_servers
}
