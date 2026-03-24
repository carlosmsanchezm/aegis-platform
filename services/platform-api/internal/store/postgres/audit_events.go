package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

func (s *PostgresStore) PutAuditEvent(event *store.AuditEvent) error {
	if event == nil {
		return fmt.Errorf("audit event required")
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	detailsJSON := []byte("{}")
	if len(event.Details) > 0 {
		if b, err := json.Marshal(event.Details); err == nil {
			detailsJSON = b
		}
	}

	ts := event.Timestamp
	if ts.IsZero() {
		ts = time.Now().UTC()
	}

	_, err := s.pool.Exec(ctx, `
INSERT INTO audit_events (id, event_type, timestamp, subject, resource_type, resource_id, action, outcome, details, source_ip)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		event.ID,
		event.EventType,
		ts,
		event.Subject,
		event.ResourceType,
		event.ResourceID,
		event.Action,
		event.Outcome,
		detailsJSON,
		event.SourceIP,
	)
	s.logExecError("audit_event_put", err, zap.String("event_id", event.ID), zap.String("event_type", event.EventType))
	return err
}

func (s *PostgresStore) ListAuditEvents(filter store.AuditEventFilter) ([]*store.AuditEvent, error) {
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	limit := filter.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.EventType != "" {
		conditions = append(conditions, fmt.Sprintf("event_type = $%d", argIdx))
		args = append(args, filter.EventType)
		argIdx++
	}
	if filter.Subject != "" {
		conditions = append(conditions, fmt.Sprintf("subject = $%d", argIdx))
		args = append(args, filter.Subject)
		argIdx++
	}
	if filter.ResourceType != "" {
		conditions = append(conditions, fmt.Sprintf("resource_type = $%d", argIdx))
		args = append(args, filter.ResourceType)
		argIdx++
	}
	if filter.ResourceID != "" {
		conditions = append(conditions, fmt.Sprintf("resource_id = $%d", argIdx))
		args = append(args, filter.ResourceID)
		argIdx++
	}
	if !filter.StartTime.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp >= $%d", argIdx))
		args = append(args, filter.StartTime)
		argIdx++
	}
	if !filter.EndTime.IsZero() {
		conditions = append(conditions, fmt.Sprintf("timestamp <= $%d", argIdx))
		args = append(args, filter.EndTime)
		argIdx++
	}

	query := "SELECT id, event_type, timestamp, subject, resource_type, resource_id, action, outcome, details, source_ip FROM audit_events"
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY timestamp DESC"
	query += fmt.Sprintf(" LIMIT $%d", argIdx)
	args = append(args, limit)

	rows, err := s.pool.Query(ctx, query, args...)
	if err != nil {
		s.logExecError("audit_events_list", err)
		return nil, err
	}
	defer rows.Close()

	var out []*store.AuditEvent
	for rows.Next() {
		ev := &store.AuditEvent{}
		var detailsJSON []byte
		if err := rows.Scan(
			&ev.ID, &ev.EventType, &ev.Timestamp, &ev.Subject,
			&ev.ResourceType, &ev.ResourceID, &ev.Action, &ev.Outcome,
			&detailsJSON, &ev.SourceIP,
		); err != nil {
			s.logExecError("audit_events_list_scan", err)
			continue
		}
		if len(detailsJSON) > 0 {
			ev.Details = make(map[string]string)
			_ = json.Unmarshal(detailsJSON, &ev.Details)
		}
		out = append(out, ev)
	}
	return out, nil
}
