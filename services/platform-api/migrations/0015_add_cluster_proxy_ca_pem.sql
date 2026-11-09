-- Add proxy_ca_pem column to store PEM-encoded spoke-proxy CA certificates.
-- Sent by the k8s-agent during cluster registration so clients can verify
-- the spoke-proxy's TLS certificate without manual CA distribution.
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS proxy_ca_pem TEXT NOT NULL DEFAULT '';
