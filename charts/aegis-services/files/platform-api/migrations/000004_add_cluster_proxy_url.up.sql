-- Add per-cluster proxy URL (e.g., wss://spoke-proxy.<ip>.nip.io:<port>) reported by k8s-agent heartbeats.
-- This is required so platform-api can return the correct proxy endpoint for remote EKS workspaces.
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS proxy_url TEXT NULL;

