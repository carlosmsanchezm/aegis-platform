package postgres

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

const defaultQueryTimeout = 5 * time.Second

// PostgresStore implements the store.Store interface backed by PostgreSQL.
type PostgresStore struct {
	pool    *pgxpool.Pool
	log     *zap.Logger
	timeout time.Duration
}

var _ store.Store = (*PostgresStore)(nil)

// New establishes a connection pool using the provided DSN and applies optional
// tuning sourced from environment variables (PG_MAX_OPEN_CONNS, PG_MAX_IDLE_CONNS,
// PG_CONN_MAX_LIFETIME, PG_DEFAULT_QUERY_TIMEOUT).
func New(dsn string, log *zap.Logger) (*PostgresStore, error) {
	if dsn == "" {
		return nil, fmt.Errorf("postgres store requires non-empty DSN")
	}
	if log == nil {
		log = zap.NewNop()
	}

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}

	if maxConns := envInt("PG_MAX_OPEN_CONNS", 0); maxConns > 0 {
		cfg.MaxConns = maxConns
	}
	if minConns := envInt("PG_MAX_IDLE_CONNS", 0); minConns > 0 {
		cfg.MinConns = minConns
	}
	if lifetime := envDuration("PG_CONN_MAX_LIFETIME", 0); lifetime > 0 {
		cfg.MaxConnLifetime = lifetime
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	timeout := envDuration("PG_DEFAULT_QUERY_TIMEOUT", defaultQueryTimeout)
	if timeout <= 0 {
		timeout = defaultQueryTimeout
	}

	store := &PostgresStore{pool: pool, log: log, timeout: timeout}
	if err := store.ensureProvisioningTables(context.Background()); err != nil {
		return nil, fmt.Errorf("ensure provisioning tables: %w", err)
	}

	return store, nil
}

// Close releases all pooled connections.
func (s *PostgresStore) Close() {
	if s == nil {
		return
	}
	s.pool.Close()
}

func (s *PostgresStore) logExecError(action string, err error, fields ...zap.Field) {
	if err == nil {
		return
	}
	if s.log == nil {
		return
	}
	fields = append(fields, zap.Error(err))
	s.log.Warn("postgres action failed", append([]zap.Field{zap.String("action", action)}, fields...)...)
}

func (s *PostgresStore) withTimeout(ctx context.Context) (context.Context, context.CancelFunc) {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithTimeout(ctx, s.timeout)
}

func envInt(key string, def int32) int32 {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if parsed, err := strconv.ParseInt(v, 10, 32); err == nil {
			return int32(parsed)
		}
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	if t.IsZero() {
		return nil
	}
	return t.UTC()
}
