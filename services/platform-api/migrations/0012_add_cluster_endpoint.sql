-- Add cluster endpoint and CA for programmatic EKS token authentication.
-- These are populated during provisioning (from Pulumi outputs) and import
-- (from kubeconfig parsing), enabling token-based auth without kubeconfig files.
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS cluster_endpoint TEXT;
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS cluster_ca TEXT;
