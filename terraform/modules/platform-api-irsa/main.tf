# IAM role for platform-api pod via IRSA (IAM Roles for Service Accounts).
# This role allows the platform-api to:
#   - Generate EKS bearer tokens (sts:GetCallerIdentity presigned URLs)
#   - Assume per-project IAM roles for multi-tenant cluster access
#   - Describe EKS clusters to fetch endpoint + CA dynamically

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]

    principals {
      type        = "Federated"
      identifiers = [var.oidc_provider_arn]
    }

    condition {
      test     = "StringEquals"
      variable = "${replace(var.oidc_issuer_url, "https://", "")}:sub"
      values   = ["system:serviceaccount:${var.namespace}:${var.service_account_name}"]
    }

    condition {
      test     = "StringEquals"
      variable = "${replace(var.oidc_issuer_url, "https://", "")}:aud"
      values   = ["sts.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "platform_api" {
  name               = var.role_name
  assume_role_policy = data.aws_iam_policy_document.assume_role.json

  tags = {
    Component = "aegis-platform-api"
    ManagedBy = "terraform"
  }
}

data "aws_iam_policy_document" "platform_api" {
  # Allow assuming per-project roles for multi-tenant EKS access
  statement {
    sid    = "AssumeProjectRoles"
    effect = "Allow"
    actions = [
      "sts:AssumeRole",
    ]
    resources = ["arn:aws:iam::*:role/aegis-project-*"]
  }

  # Required for EKS token generation (presigned GetCallerIdentity)
  statement {
    sid    = "GetCallerIdentity"
    effect = "Allow"
    actions = [
      "sts:GetCallerIdentity",
    ]
    resources = ["*"]
  }

  # Optional: describe clusters to dynamically fetch endpoint + CA
  statement {
    sid    = "DescribeEKSClusters"
    effect = "Allow"
    actions = [
      "eks:DescribeCluster",
    ]
    resources = ["*"]
  }

  # Pulumi state backend (S3)
  statement {
    sid    = "PulumiStateBackend"
    effect = "Allow"
    actions = [
      "s3:GetObject",
      "s3:PutObject",
      "s3:DeleteObject",
      "s3:ListBucket",
      "s3:GetBucketLocation",
    ]
    resources = [
      "arn:aws:s3:::aegis-pulumi-state-${var.aws_account_id}",
      "arn:aws:s3:::aegis-pulumi-state-${var.aws_account_id}/*",
    ]
  }
}

resource "aws_iam_role_policy" "platform_api" {
  name   = "${var.role_name}-policy"
  role   = aws_iam_role.platform_api.id
  policy = data.aws_iam_policy_document.platform_api.json
}
