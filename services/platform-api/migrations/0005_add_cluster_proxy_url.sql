-- +migrate Up
ALTER TABLE clusters ADD COLUMN IF NOT EXISTS proxy_url TEXT NULL;

-- +migrate Down
ALTER TABLE clusters DROP COLUMN IF EXISTS proxy_url;
