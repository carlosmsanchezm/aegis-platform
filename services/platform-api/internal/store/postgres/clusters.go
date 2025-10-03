package postgres

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

func (s *PostgresStore) UpsertClusterFromRegister(req *aegis.ClusterRegisterRequest) {
	if req == nil || req.GetClusterId() == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logExecError("cluster_register_begin", err)
		return
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `INSERT INTO clusters (id, provider, region, created_at, updated_at)
VALUES ($1, $2, $3, now(), now())
ON CONFLICT (id) DO UPDATE SET provider=EXCLUDED.provider, region=EXCLUDED.region, updated_at=now()`, req.GetClusterId(), nullableString(req.GetProvider()), nullableString(req.GetRegion())); err != nil {
		s.logExecError("cluster_register_upsert", err, zap.String("cluster_id", req.GetClusterId()))
		return
	}

	if _, err := tx.Exec(ctx, `DELETE FROM cluster_labels WHERE cluster_id=$1`, req.GetClusterId()); err != nil {
		s.logExecError("cluster_register_delete_labels", err, zap.String("cluster_id", req.GetClusterId()))
		return
	}
	for k, v := range req.GetLabels() {
		if k == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO cluster_labels (cluster_id, k, v) VALUES ($1, $2, $3)`, req.GetClusterId(), k, v); err != nil {
			s.logExecError("cluster_register_insert_label", err, zap.String("cluster_id", req.GetClusterId()), zap.String("label", k))
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		s.logExecError("cluster_register_commit", err, zap.String("cluster_id", req.GetClusterId()))
	}
}

func (s *PostgresStore) UpdateClusterFromHeartbeat(hb *aegis.ClusterHeartbeat) {
	// Heartbeats are ephemeral operational metrics - no need to persist to database
	// Cluster registration handles persistence of cluster metadata
	return
}

func (s *PostgresStore) ListClusterInfos() []*store.ClusterInfo {
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	rows, err := s.pool.Query(ctx, `SELECT id, COALESCE(provider, ''), COALESCE(region, ''), ttf_gpu_seconds_p50, last_heartbeat FROM clusters`)
	if err != nil {
		s.logExecError("cluster_list", err)
		return nil
	}
	defer rows.Close()

	clusters := map[string]*store.ClusterInfo{}
	for rows.Next() {
		var (
			id        string
			provider  string
			region    string
			ttf       float64
			heartbeat sql.NullTime
		)
		if err := rows.Scan(&id, &provider, &region, &ttf, &heartbeat); err != nil {
			s.logExecError("cluster_list_scan", err)
			return nil
		}
		info := &store.ClusterInfo{
			ID:                 id,
			Provider:           provider,
			Region:             region,
			Labels:             map[string]string{},
			AvailableFlavorSet: map[string]bool{},
			TTFGSecondsP50:     ttf,
		}
		if heartbeat.Valid {
			info.LastHeartbeat = heartbeat.Time.UTC()
		}
		clusters[id] = info
	}

	labelRows, err := s.pool.Query(ctx, `SELECT cluster_id, k, v FROM cluster_labels`)
	if err == nil {
		defer labelRows.Close()
		for labelRows.Next() {
			var (
				id string
				k  string
				v  string
			)
			if err := labelRows.Scan(&id, &k, &v); err != nil {
				s.logExecError("cluster_list_scan_labels", err)
				break
			}
			if info, ok := clusters[id]; ok {
				info.Labels[k] = v
			}
		}
	} else {
		s.logExecError("cluster_list_labels", err)
	}

	flavorRows, err := s.pool.Query(ctx, `SELECT cluster_id, flavor FROM cluster_flavors`)
	if err == nil {
		defer flavorRows.Close()
		for flavorRows.Next() {
			var id, flavor string
			if err := flavorRows.Scan(&id, &flavor); err != nil {
				s.logExecError("cluster_list_scan_flavors", err)
				break
			}
			if info, ok := clusters[id]; ok {
				info.AvailableFlavorSet[flavor] = true
			}
		}
	} else {
		s.logExecError("cluster_list_flavors", err)
	}

	out := make([]*store.ClusterInfo, 0, len(clusters))
	for _, info := range clusters {
		out = append(out, info)
	}
	return out
}
