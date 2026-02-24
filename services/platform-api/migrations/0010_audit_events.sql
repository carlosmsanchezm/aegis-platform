-- +migrate Up
CREATE TABLE IF NOT EXISTS audit_events (
    id TEXT PRIMARY KEY,
    event_type TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    subject TEXT NOT NULL,
    resource_type TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    action TEXT NOT NULL,
    outcome TEXT NOT NULL,
    details JSONB DEFAULT '{}',
    source_ip TEXT DEFAULT ''
);

CREATE INDEX idx_audit_events_timestamp ON audit_events (timestamp DESC);
CREATE INDEX idx_audit_events_event_type ON audit_events (event_type);
CREATE INDEX idx_audit_events_subject ON audit_events (subject);
CREATE INDEX idx_audit_events_resource ON audit_events (resource_type, resource_id);

-- +migrate Down
DROP TABLE IF EXISTS audit_events;
