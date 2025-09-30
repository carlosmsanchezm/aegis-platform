package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

func (s *PostgresStore) PutConnectionSession(sess *store.ConnectionSession) *store.ConnectionSession {
	if sess == nil || sess.SessionID == "" {
		return nil
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logExecError("session_put_begin", err)
		return nil
	}
	defer tx.Rollback(ctx)

	existing, err := s.getSessionTx(ctx, tx, sess.SessionID, true)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		s.logExecError("session_put_fetch", err, zap.String("session_id", sess.SessionID))
		return nil
	}

	now := time.Now().UTC()
	createdAt := sess.CreatedAt
	if existing != nil {
		createdAt = existing.CreatedAt
	}
	if createdAt.IsZero() {
		createdAt = now
	}
	sess.CreatedAt = createdAt
	sess.UpdatedAt = now

	_, err = tx.Exec(ctx, `
INSERT INTO connection_sessions (
    session_id, workload_id, subject, client, jti, token, ssh_user, ssh_host_alias,
    internal_host, port, ssh_config, proxy_url, vscode_uri, expires_at, one_time, used, revoked, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19
)
ON CONFLICT (session_id) DO UPDATE SET
    workload_id = EXCLUDED.workload_id,
    subject = EXCLUDED.subject,
    client = EXCLUDED.client,
    jti = EXCLUDED.jti,
    token = EXCLUDED.token,
    ssh_user = EXCLUDED.ssh_user,
    ssh_host_alias = EXCLUDED.ssh_host_alias,
    internal_host = EXCLUDED.internal_host,
    port = EXCLUDED.port,
    ssh_config = EXCLUDED.ssh_config,
    proxy_url = EXCLUDED.proxy_url,
    vscode_uri = EXCLUDED.vscode_uri,
    expires_at = EXCLUDED.expires_at,
    one_time = EXCLUDED.one_time,
    used = EXCLUDED.used,
    revoked = EXCLUDED.revoked,
    created_at = $18,
    updated_at = EXCLUDED.updated_at
`,
		sess.SessionID, sess.WorkloadID, sess.Subject, sess.Client, sess.JTI, sess.Token, sess.SSHUser, sess.SSHHostAlias,
		sess.InternalHost, sess.Port, sess.SSHConfig, sess.ProxyURL, sess.VSCodeURI, sess.ExpiresAt.UTC(), sess.OneTime, sess.Used, sess.Revoked, createdAt, sess.UpdatedAt,
	)
	if err != nil {
		s.logExecError("session_upsert", err, zap.String("session_id", sess.SessionID))
		return nil
	}

	if existing != nil && existing.JTI != "" && existing.JTI != sess.JTI {
		if _, err := tx.Exec(ctx, `DELETE FROM session_jtis WHERE jti=$1`, existing.JTI); err != nil {
			s.logExecError("session_delete_prev_jti", err, zap.String("jti", existing.JTI))
			return nil
		}
	}

	if sess.JTI != "" {
		_, err = tx.Exec(ctx, `INSERT INTO session_jtis (jti, session_id, used, expires_at) VALUES ($1, $2, $3, $4)
ON CONFLICT (jti) DO UPDATE SET session_id = EXCLUDED.session_id, used = EXCLUDED.used, expires_at = EXCLUDED.expires_at`, sess.JTI, sess.SessionID, sess.Used, sess.ExpiresAt.UTC())
		if err != nil {
			s.logExecError("session_upsert_jti", err, zap.String("jti", sess.JTI))
			return nil
		}
	}

	stored, err := s.getSessionTx(ctx, tx, sess.SessionID, false)
	if err != nil {
		s.logExecError("session_reload", err, zap.String("session_id", sess.SessionID))
		return nil
	}

	if err := tx.Commit(ctx); err != nil {
		s.logExecError("session_put_commit", err, zap.String("session_id", sess.SessionID))
		return nil
	}

	return stored
}

func (s *PostgresStore) ConnectionSession(id string) (*store.ConnectionSession, bool) {
	if id == "" {
		return nil, false
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	sess, err := s.getSession(ctx, id)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.logExecError("session_get", err, zap.String("session_id", id))
		}
		return nil, false
	}
	return sess, true
}

func (s *PostgresStore) ConnectionSessionByJTI(jti string) (*store.ConnectionSession, bool) {
	if jti == "" {
		return nil, false
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	row := s.pool.QueryRow(ctx, `
SELECT cs.session_id FROM session_jtis sj
JOIN connection_sessions cs ON cs.session_id = sj.session_id
WHERE sj.jti=$1
`, jti)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.logExecError("session_get_by_jti", err, zap.String("jti", jti))
		}
		return nil, false
	}
	sess, err := s.getSession(ctx, sessionID)
	if err != nil {
		s.logExecError("session_get_by_jti_fetch", err, zap.String("session_id", sessionID))
		return nil, false
	}
	return sess, true
}

func (s *PostgresStore) UpdateConnectionSession(id string, mutate func(*store.ConnectionSession) error) (*store.ConnectionSession, error) {
	if id == "" {
		return nil, store.ErrSessionNotFound
	}
	if mutate == nil {
		return nil, errors.New("mutate function required")
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	sess, err := s.getSessionTx(ctx, tx, id, true)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, store.ErrSessionNotFound
		}
		return nil, err
	}

	mutable := *sess
	if err := mutate(&mutable); err != nil {
		return nil, err
	}
	mutable.UpdatedAt = time.Now().UTC()

	_, err = tx.Exec(ctx, `UPDATE connection_sessions SET
    workload_id=$2, subject=$3, client=$4, jti=$5, token=$6, ssh_user=$7, ssh_host_alias=$8,
    internal_host=$9, port=$10, ssh_config=$11, proxy_url=$12, vscode_uri=$13,
    expires_at=$14, one_time=$15, used=$16, revoked=$17, updated_at=$18
WHERE session_id=$1`,
		mutable.SessionID, mutable.WorkloadID, mutable.Subject, mutable.Client, mutable.JTI, mutable.Token, mutable.SSHUser, mutable.SSHHostAlias,
		mutable.InternalHost, mutable.Port, mutable.SSHConfig, mutable.ProxyURL, mutable.VSCodeURI,
		mutable.ExpiresAt.UTC(), mutable.OneTime, mutable.Used, mutable.Revoked, mutable.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if sess.JTI != mutable.JTI && sess.JTI != "" {
		if _, err := tx.Exec(ctx, `DELETE FROM session_jtis WHERE jti=$1`, sess.JTI); err != nil {
			return nil, err
		}
	}
	if mutable.JTI != "" {
		if _, err := tx.Exec(ctx, `INSERT INTO session_jtis (jti, session_id, used, expires_at) VALUES ($1, $2, $3, $4)
ON CONFLICT (jti) DO UPDATE SET session_id=EXCLUDED.session_id, used=EXCLUDED.used, expires_at=EXCLUDED.expires_at`, mutable.JTI, mutable.SessionID, mutable.Used, mutable.ExpiresAt.UTC()); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &mutable, nil
}

func (s *PostgresStore) MarkSessionUsed(id string) bool {
	if id == "" {
		return false
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	tag, err := s.pool.Exec(ctx, `UPDATE connection_sessions SET used=true, updated_at=now() WHERE session_id=$1`, id)
	if err != nil {
		s.logExecError("session_mark_used", err, zap.String("session_id", id))
		return false
	}
	if tag.RowsAffected() == 0 {
		return false
	}
	if _, err := s.pool.Exec(ctx, `UPDATE session_jtis SET used=true WHERE session_id=$1`, id); err != nil {
		s.logExecError("session_mark_used_jti", err, zap.String("session_id", id))
	}
	return true
}

func (s *PostgresStore) MarkJTIUsed(jti string) bool {
	if jti == "" {
		return false
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logExecError("session_mark_jti_begin", err)
		return false
	}
	defer tx.Rollback(ctx)

	row := tx.QueryRow(ctx, `SELECT session_id FROM session_jtis WHERE jti=$1 FOR UPDATE`, jti)
	var sessionID string
	if err := row.Scan(&sessionID); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			s.logExecError("session_mark_jti_scan", err, zap.String("jti", jti))
		}
		return false
	}

	if _, err := tx.Exec(ctx, `UPDATE session_jtis SET used=true WHERE jti=$1`, jti); err != nil {
		s.logExecError("session_mark_jti_update", err, zap.String("jti", jti))
		return false
	}
	if _, err := tx.Exec(ctx, `UPDATE connection_sessions SET used=true, updated_at=now() WHERE session_id=$1`, sessionID); err != nil {
		s.logExecError("session_mark_jti_update_session", err, zap.String("session_id", sessionID))
		return false
	}

	if err := tx.Commit(ctx); err != nil {
		s.logExecError("session_mark_jti_commit", err, zap.String("jti", jti))
		return false
	}
	return true
}

func (s *PostgresStore) DeleteConnectionSession(id string) (*store.ConnectionSession, bool) {
	if id == "" {
		return nil, false
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logExecError("session_delete_begin", err)
		return nil, false
	}
	defer tx.Rollback(ctx)

	sess, err := s.getSessionTx(ctx, tx, id, true)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false
		}
		s.logExecError("session_delete_fetch", err, zap.String("session_id", id))
		return nil, false
	}

	if _, err := tx.Exec(ctx, `DELETE FROM connection_sessions WHERE session_id=$1`, id); err != nil {
		s.logExecError("session_delete_exec", err, zap.String("session_id", id))
		return nil, false
	}
	if sess.JTI != "" {
		if _, err := tx.Exec(ctx, `DELETE FROM session_jtis WHERE jti=$1`, sess.JTI); err != nil {
			s.logExecError("session_delete_jti", err, zap.String("jti", sess.JTI))
			return nil, false
		}
	}

	if err := tx.Commit(ctx); err != nil {
		s.logExecError("session_delete_commit", err, zap.String("session_id", id))
		return nil, false
	}
	return sess, true
}

func (s *PostgresStore) SessionsForWorkload(workloadID string) []*store.ConnectionSession {
	if workloadID == "" {
		return nil
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	rows, err := s.pool.Query(ctx, `SELECT session_id FROM connection_sessions WHERE workload_id=$1`, workloadID)
	if err != nil {
		s.logExecError("session_list_for_workload", err, zap.String("workload_id", workloadID))
		return nil
	}
	defer rows.Close()
	var sessions []*store.ConnectionSession
	for rows.Next() {
		var sessionID string
		if err := rows.Scan(&sessionID); err != nil {
			s.logExecError("session_list_scan_id", err)
			break
		}
		sess, err := s.getSession(ctx, sessionID)
		if err != nil {
			s.logExecError("session_list_fetch", err, zap.String("session_id", sessionID))
			continue
		}
		sessions = append(sessions, sess)
	}
	return sessions
}

func (s *PostgresStore) PurgeExpiredSessions(now time.Time) {
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	grace := now.Add(-5 * time.Minute)

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logExecError("session_purge_begin", err)
		return
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `SELECT session_id, jti FROM connection_sessions WHERE expires_at < $1`, grace)
	if err != nil {
		s.logExecError("session_purge_select", err)
		return
	}
	type victim struct {
		sessionID string
		jti       string
	}
	victims := []victim{}
	for rows.Next() {
		var v victim
		if err := rows.Scan(&v.sessionID, &v.jti); err != nil {
			s.logExecError("session_purge_scan", err)
			rows.Close()
			return
		}
		victims = append(victims, v)
	}
	rows.Close()

	for _, v := range victims {
		if _, err := tx.Exec(ctx, `DELETE FROM connection_sessions WHERE session_id=$1`, v.sessionID); err != nil {
			s.logExecError("session_purge_delete", err, zap.String("session_id", v.sessionID))
			return
		}
		if v.jti != "" {
			if _, err := tx.Exec(ctx, `DELETE FROM session_jtis WHERE jti=$1`, v.jti); err != nil {
				s.logExecError("session_purge_delete_jti", err, zap.String("jti", v.jti))
				return
			}
		}
	}

	if _, err := tx.Exec(ctx, `DELETE FROM session_jtis WHERE expires_at < $1`, grace); err != nil {
		s.logExecError("session_purge_orphaned_jtis", err)
		return
	}

	if err := tx.Commit(ctx); err != nil {
		s.logExecError("session_purge_commit", err)
	}
}

func (s *PostgresStore) getSession(ctx context.Context, id string) (*store.ConnectionSession, error) {
	return s.getSessionTx(ctx, nil, id, false)
}

func (s *PostgresStore) getSessionTx(ctx context.Context, tx pgx.Tx, id string, forUpdate bool) (*store.ConnectionSession, error) {
	var row pgx.Row
	query := `SELECT session_id, workload_id, subject, client, jti, token, ssh_user, ssh_host_alias, internal_host,
        port, ssh_config, proxy_url, vscode_uri, expires_at, one_time, used, revoked, created_at, updated_at
        FROM connection_sessions WHERE session_id=$1`
	if forUpdate {
		query += " FOR UPDATE"
	}
	if tx != nil {
		row = tx.QueryRow(ctx, query, id)
	} else {
		row = s.pool.QueryRow(ctx, query, id)
	}
	sess := &store.ConnectionSession{}
	if err := row.Scan(
		&sess.SessionID, &sess.WorkloadID, &sess.Subject, &sess.Client, &sess.JTI, &sess.Token, &sess.SSHUser, &sess.SSHHostAlias,
		&sess.InternalHost, &sess.Port, &sess.SSHConfig, &sess.ProxyURL, &sess.VSCodeURI, &sess.ExpiresAt, &sess.OneTime,
		&sess.Used, &sess.Revoked, &sess.CreatedAt, &sess.UpdatedAt,
	); err != nil {
		return nil, err
	}
	sess.ExpiresAt = sess.ExpiresAt.UTC()
	sess.CreatedAt = sess.CreatedAt.UTC()
	sess.UpdatedAt = sess.UpdatedAt.UTC()
	return sess, nil
}
