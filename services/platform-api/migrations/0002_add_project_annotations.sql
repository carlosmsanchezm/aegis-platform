-- Add annotations column to projects table (used by platform-api store)
ALTER TABLE projects
ADD COLUMN IF NOT EXISTS annotations JSONB DEFAULT '{}'::jsonb;

