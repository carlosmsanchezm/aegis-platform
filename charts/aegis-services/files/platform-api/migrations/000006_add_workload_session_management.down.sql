ALTER TABLE workloads DROP COLUMN IF EXISTS terminate_reason;
ALTER TABLE workloads DROP COLUMN IF EXISTS terminated_at;
ALTER TABLE workloads DROP COLUMN IF EXISTS resume_count;
ALTER TABLE workloads DROP COLUMN IF EXISTS suspend_reason;
ALTER TABLE workloads DROP COLUMN IF EXISTS suspended_at;
