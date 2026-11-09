# IAM role for a spoke/project AWS account.
# This is the role that platform-api assumes via STS to provision EKS clusters
# and related resources in the project's AWS account.
#
# The IRSA module (platform-api-irsa) grants platform-api permission to
# sts:AssumeRole on arn:aws:iam::*:role/aegis-project-*. This module creates
# the target role that matches that pattern.

data "aws_iam_policy_document" "trust" {
  statement {
    sid     = "AllowHubAssumeRole"
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "AWS"
      identifiers = [var.hub_platform_api_role_arn]
    }

    condition {
      test     = "StringEquals"
      variable = "sts:ExternalId"
      values   = [var.external_id]
    }
  }
}

resource "aws_iam_role" "spoke_project" {
  name               = "aegis-project-${var.project_id}"
  assume_role_policy = data.aws_iam_policy_document.trust.json
  tags               = merge(var.tags, { AegisProject = var.project_id })
}

data "aws_iam_policy_document" "spoke_permissions" {
  # EKS cluster and nodegroup management
  statement {
    sid    = "EKSManagement"
    effect = "Allow"
    actions = [
      "eks:CreateCluster",
      "eks:DeleteCluster",
      "eks:DescribeCluster",
      "eks:UpdateClusterConfig",
      "eks:UpdateClusterVersion",
      "eks:TagResource",
      "eks:UntagResource",
      "eks:ListClusters",
      "eks:CreateNodegroup",
      "eks:DeleteNodegroup",
      "eks:DescribeNodegroup",
      "eks:UpdateNodegroupConfig",
      "eks:ListNodegroups",
      "eks:CreateAddon",
      "eks:DeleteAddon",
      "eks:DescribeAddon",
      "eks:UpdateAddon",
      "eks:ListAddons",
      "eks:AssociateEncryptionConfig",
    ]
    resources = ["*"]
  }

  # IAM roles and policies for EKS node groups, IRSA, etc.
  statement {
    sid    = "IAMManagement"
    effect = "Allow"
    actions = [
      "iam:CreateRole",
      "iam:DeleteRole",
      "iam:GetRole",
      "iam:ListRolePolicies",
      "iam:ListAttachedRolePolicies",
      "iam:PutRolePolicy",
      "iam:DeleteRolePolicy",
      "iam:GetRolePolicy",
      "iam:AttachRolePolicy",
      "iam:DetachRolePolicy",
      "iam:PassRole",
      "iam:TagRole",
      "iam:UntagRole",
      "iam:CreateOpenIDConnectProvider",
      "iam:DeleteOpenIDConnectProvider",
      "iam:GetOpenIDConnectProvider",
      "iam:TagOpenIDConnectProvider",
      "iam:CreateInstanceProfile",
      "iam:DeleteInstanceProfile",
      "iam:GetInstanceProfile",
      "iam:AddRoleToInstanceProfile",
      "iam:RemoveRoleFromInstanceProfile",
      "iam:ListInstanceProfilesForRole",
    ]
    resources = ["*"]
  }

  # EC2 for EKS (security groups, VPC lookups, instances, tagging)
  statement {
    sid    = "EC2Networking"
    effect = "Allow"
    actions = [
      "ec2:CreateSecurityGroup",
      "ec2:DeleteSecurityGroup",
      "ec2:AuthorizeSecurityGroupIngress",
      "ec2:AuthorizeSecurityGroupEgress",
      "ec2:RevokeSecurityGroupIngress",
      "ec2:RevokeSecurityGroupEgress",
      "ec2:DescribeSecurityGroups",
      "ec2:DescribeSecurityGroupRules",
      "ec2:DescribeVpcs",
      "ec2:DescribeSubnets",
      "ec2:DescribeAvailabilityZones",
      "ec2:DescribeRouteTables",
      "ec2:DescribeInternetGateways",
      "ec2:DescribeNatGateways",
      "ec2:DescribeInstances",
      "ec2:DescribeImages",
      "ec2:DescribeInstanceTypes",
      "ec2:DescribeKeyPairs",
      "ec2:DescribeVolumes",
      "ec2:DescribeNetworkInterfaces",
      "ec2:RunInstances",
      "ec2:TerminateInstances",
      "ec2:CreateTags",
      "ec2:DeleteTags",
      "ec2:DescribeTags",
      "ec2:CreateLaunchTemplate",
      "ec2:CreateLaunchTemplateVersion",
      "ec2:DeleteLaunchTemplate",
      "ec2:DescribeLaunchTemplates",
      "ec2:DescribeLaunchTemplateVersions",
    ]
    resources = ["*"]
  }

  # STS for identity verification
  statement {
    sid       = "STSIdentity"
    effect    = "Allow"
    actions   = ["sts:GetCallerIdentity"]
    resources = ["*"]
  }

  # CloudFormation (Pulumi uses CF for some resource types)
  statement {
    sid    = "CloudFormation"
    effect = "Allow"
    actions = [
      "cloudformation:CreateStack",
      "cloudformation:DeleteStack",
      "cloudformation:DescribeStacks",
      "cloudformation:DescribeStackEvents",
      "cloudformation:DescribeStackResources",
      "cloudformation:GetTemplate",
      "cloudformation:ListStacks",
      "cloudformation:UpdateStack",
      "cloudformation:ListStackResources",
    ]
    resources = ["*"]
  }

  # KMS for EKS envelope encryption (optional but recommended)
  statement {
    sid    = "KMSForEKS"
    effect = "Allow"
    actions = [
      "kms:CreateKey",
      "kms:DescribeKey",
      "kms:CreateAlias",
      "kms:DeleteAlias",
      "kms:ListAliases",
      "kms:TagResource",
    ]
    resources = ["*"]
  }

  # CloudWatch Logs for EKS control plane logging
  statement {
    sid    = "CloudWatchLogs"
    effect = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:DeleteLogGroup",
      "logs:DescribeLogGroups",
      "logs:PutRetentionPolicy",
      "logs:TagResource",
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "spoke_permissions" {
  name   = "aegis-project-${var.project_id}-permissions"
  role   = aws_iam_role.spoke_project.id
  policy = data.aws_iam_policy_document.spoke_permissions.json
}
