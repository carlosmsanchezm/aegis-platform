ALTER TABLE clusters DROP CONSTRAINT IF EXISTS clusters_import_method_check;
ALTER TABLE clusters
    DROP COLUMN IF EXISTS assume_role_arn,
    DROP COLUMN IF EXISTS kubeconfig_secret_ref,
    DROP COLUMN IF EXISTS imported_at,
    DROP COLUMN IF EXISTS import_method;
