-- +migrate Up
ALTER TABLE clusters
    ADD COLUMN IF NOT EXISTS import_method VARCHAR(50) NOT NULL DEFAULT 'provisioned',
    ADD COLUMN IF NOT EXISTS imported_at TIMESTAMPTZ NULL,
    ADD COLUMN IF NOT EXISTS kubeconfig_secret_ref VARCHAR(255) NULL,
    ADD COLUMN IF NOT EXISTS assume_role_arn VARCHAR(255) NULL;

DO $$
BEGIN
    ALTER TABLE clusters
        ADD CONSTRAINT clusters_import_method_check
        CHECK (import_method IN ('provisioned', 'kubeconfig', 'assume_role', 'agent_only'));
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;

-- +migrate Down
ALTER TABLE clusters DROP CONSTRAINT IF EXISTS clusters_import_method_check;
ALTER TABLE clusters
    DROP COLUMN IF EXISTS assume_role_arn,
    DROP COLUMN IF EXISTS kubeconfig_secret_ref,
    DROP COLUMN IF EXISTS imported_at,
    DROP COLUMN IF EXISTS import_method;

