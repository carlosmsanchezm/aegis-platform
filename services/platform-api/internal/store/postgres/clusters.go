package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

func (s *PostgresStore) UpsertClusterImport(req store.ClusterImport) error {
	clusterID := strings.TrimSpace(req.ClusterID)
	if clusterID == "" {
		return fmt.Errorf("cluster_id required")
	}
	importMethod := strings.TrimSpace(req.ImportMethod)
	if importMethod == "" {
		return fmt.Errorf("import_method required")
	}
	projectID := strings.TrimSpace(req.ProjectID)

	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin cluster import transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `INSERT INTO clusters (
  id,
  provider,
  region,
  import_method,
  imported_at,
  kubeconfig_secret_ref,
  assume_role_arn,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
ON CONFLICT (id) DO UPDATE SET
  provider=EXCLUDED.provider,
  region=EXCLUDED.region,
  import_method=EXCLUDED.import_method,
  imported_at=EXCLUDED.imported_at,
  kubeconfig_secret_ref=EXCLUDED.kubeconfig_secret_ref,
  assume_role_arn=EXCLUDED.assume_role_arn,
  deleted_at=NULL,
  deleted_by=NULL,
  deletion_reason=NULL,
  updated_at=now()`,
		clusterID,
		nullableString(req.Provider),
		nullableString(req.Region),
		importMethod,
		nullableTime(&req.ImportedAt),
		nullableString(req.KubeconfigSecretRef),
		nullableString(req.AssumeRoleARN),
	); err != nil {
		return fmt.Errorf("upsert cluster %q: %w", clusterID, err)
	}

	if projectID != "" {
		var stored string
		err := tx.QueryRow(ctx, `INSERT INTO cluster_labels (cluster_id, k, v) VALUES ($1, $2, $3)
ON CONFLICT (cluster_id, k) DO UPDATE SET v = EXCLUDED.v WHERE cluster_labels.v = EXCLUDED.v
RETURNING v`, clusterID, "aegis.yourorg.dev/projectId", projectID).Scan(&stored)
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%w", store.ErrClusterProjectConflict)
		}
		if err != nil {
			return fmt.Errorf("upsert cluster project label for %q: %w", clusterID, err)
		}
	}

	labels := req.Labels
	if labels == nil {
		labels = map[string]string{}
	}

	for k, v := range labels {
		if strings.TrimSpace(k) == "" {
			continue
		}
		if k == "aegis.yourorg.dev/projectId" {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO cluster_labels (cluster_id, k, v) VALUES ($1, $2, $3)
ON CONFLICT (cluster_id, k) DO UPDATE SET v = EXCLUDED.v`, clusterID, k, v); err != nil {
			return fmt.Errorf("upsert cluster label %q for %q: %w", k, clusterID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit cluster import %q: %w", clusterID, err)
	}
	return nil
}

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
ON CONFLICT (id) DO UPDATE SET provider=EXCLUDED.provider, region=EXCLUDED.region, deleted_at=NULL, updated_at=now()`, req.GetClusterId(), nullableString(req.GetProvider()), nullableString(req.GetRegion())); err != nil {
		s.logExecError("cluster_register_upsert", err, zap.String("cluster_id", req.GetClusterId()))
		return
	}

	for k, v := range req.GetLabels() {
		if k == "" {
			continue
		}
		// projectId label is managed separately; avoid duplicate-key error on re-register
		if k == "aegis.yourorg.dev/projectId" {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO cluster_labels (cluster_id, k, v) VALUES ($1, $2, $3)
ON CONFLICT (cluster_id, k) DO UPDATE SET v = EXCLUDED.v`, req.GetClusterId(), k, v); err != nil {
			s.logExecError("cluster_register_insert_label", err, zap.String("cluster_id", req.GetClusterId()), zap.String("label", k))
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		s.logExecError("cluster_register_commit", err, zap.String("cluster_id", req.GetClusterId()))
	}
}

func (s *PostgresStore) UpdateClusterFromHeartbeat(hb *aegis.ClusterHeartbeat) {
	if hb == nil || hb.GetClusterId() == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logExecError("heartbeat_begin", err)
		return
	}
	defer tx.Rollback(ctx)

	// Update TTFG metric, proxy URL, and last heartbeat timestamp; do not resurrect soft-deleted clusters.
	res, err := tx.Exec(ctx, `UPDATE clusters
SET ttf_gpu_seconds_p50=$1,
    proxy_url = CASE WHEN $3 <> '' THEN $3 ELSE proxy_url END,
    last_heartbeat=now(),
    updated_at=now()
WHERE id=$2 AND deleted_at IS NULL`,
		hb.GetTtfGpuSecondsP50(), hb.GetClusterId(), strings.TrimSpace(hb.GetProxyUrl()))
	if err != nil {
		s.logExecError("heartbeat_update_ttfg", err, zap.String("cluster_id", hb.GetClusterId()))
		return
	}
	if rows := res.RowsAffected(); rows == 0 {
		s.log.Debug("heartbeat ignored for soft-deleted cluster", zap.String("cluster_id", hb.GetClusterId()))
		return
	}

	// Update available flavors
	if _, err := tx.Exec(ctx, `DELETE FROM cluster_flavors WHERE cluster_id=$1`, hb.GetClusterId()); err != nil {
		s.logExecError("heartbeat_delete_flavors", err, zap.String("cluster_id", hb.GetClusterId()))
		return
	}
	for _, f := range hb.GetAvailableFlavors() {
		if f.GetName() == "" {
			continue
		}
		if _, err := tx.Exec(ctx, `INSERT INTO cluster_flavors (cluster_id, flavor) VALUES ($1, $2)`,
			hb.GetClusterId(), f.GetName()); err != nil {
			s.logExecError("heartbeat_insert_flavor", err,
				zap.String("cluster_id", hb.GetClusterId()),
				zap.String("flavor", f.GetName()))
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		s.logExecError("heartbeat_commit", err, zap.String("cluster_id", hb.GetClusterId()))
	}
}

func (s *PostgresStore) GetClusterInfo(clusterID string) *store.ClusterInfo {
	if clusterID == "" {
		return nil
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	var (
		id           string
		provider     string
		region       string
		importMethod string
		importedAt   sql.NullTime
		kubeconfig   sql.NullString
		assumeRole   sql.NullString
		ttf          float64
		proxyURL     sql.NullString
		heartbeat    sql.NullTime
		createdAt    time.Time
	)
	err := s.pool.QueryRow(ctx, `SELECT
  id,
  COALESCE(provider, ''),
  COALESCE(region, ''),
  COALESCE(import_method, 'provisioned'),
  imported_at,
  kubeconfig_secret_ref,
  assume_role_arn,
  ttf_gpu_seconds_p50,
  proxy_url,
  last_heartbeat,
  created_at
FROM clusters WHERE id=$1 AND deleted_at IS NULL`, clusterID).Scan(
		&id,
		&provider,
		&region,
		&importMethod,
		&importedAt,
		&kubeconfig,
		&assumeRole,
		&ttf,
		&proxyURL,
		&heartbeat,
		&createdAt,
	)
	if err != nil {
		return nil
	}
	info := &store.ClusterInfo{
		ID:                 id,
		Provider:           provider,
		Region:             region,
		Labels:             map[string]string{},
		AvailableFlavorSet: map[string]bool{},
		TTFGSecondsP50:     ttf,
		CreatedAt:          createdAt.UTC(),
		ImportMethod:       importMethod,
	}
	if importedAt.Valid {
		info.ImportedAt = importedAt.Time.UTC()
	}
	if kubeconfig.Valid {
		info.KubeconfigSecretRef = kubeconfig.String
	}
	if assumeRole.Valid {
		info.AssumeRoleARN = assumeRole.String
	}
	if heartbeat.Valid {
		info.LastHeartbeat = heartbeat.Time.UTC()
	}
	if proxyURL.Valid {
		info.ProxyURL = proxyURL.String
	}
	return info
}

func (s *PostgresStore) ListClusterInfos() []*store.ClusterInfo {
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	rows, err := s.pool.Query(ctx, `SELECT
  id,
  COALESCE(provider, ''),
  COALESCE(region, ''),
  COALESCE(import_method, 'provisioned'),
  imported_at,
  kubeconfig_secret_ref,
  assume_role_arn,
  ttf_gpu_seconds_p50,
  proxy_url,
  last_heartbeat,
  created_at
FROM clusters WHERE deleted_at IS NULL`)
	if err != nil {
		s.logExecError("cluster_list", err)
		return nil
	}
	defer rows.Close()

	clusters := map[string]*store.ClusterInfo{}
	for rows.Next() {
		var (
			id           string
			provider     string
			region       string
			importMethod string
			importedAt   sql.NullTime
			kubeconfig   sql.NullString
			assumeRole   sql.NullString
			ttf          float64
			proxyURL     sql.NullString
			heartbeat    sql.NullTime
			createdAt    time.Time
		)
		if err := rows.Scan(&id, &provider, &region, &importMethod, &importedAt, &kubeconfig, &assumeRole, &ttf, &proxyURL, &heartbeat, &createdAt); err != nil {
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
			CreatedAt:          createdAt.UTC(),
			ImportMethod:       importMethod,
		}
		if importedAt.Valid {
			info.ImportedAt = importedAt.Time.UTC()
		}
		if kubeconfig.Valid {
			info.KubeconfigSecretRef = kubeconfig.String
		}
		if assumeRole.Valid {
			info.AssumeRoleARN = assumeRole.String
		}
		if heartbeat.Valid {
			info.LastHeartbeat = heartbeat.Time.UTC()
		}
		if proxyURL.Valid {
			info.ProxyURL = proxyURL.String
		}
		clusters[id] = info
	}

	labelRows, err := s.pool.Query(ctx, `SELECT cl.cluster_id, cl.k, cl.v FROM cluster_labels cl JOIN clusters c ON cl.cluster_id = c.id WHERE c.deleted_at IS NULL`)
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

	flavorRows, err := s.pool.Query(ctx, `SELECT cf.cluster_id, cf.flavor FROM cluster_flavors cf JOIN clusters c ON cf.cluster_id = c.id WHERE c.deleted_at IS NULL`)
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

func (s *PostgresStore) SetClusterProjectID(clusterID, projectID string) {
	if clusterID == "" || projectID == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	// Upsert the project label for the cluster
	_, err := s.pool.Exec(ctx, `INSERT INTO cluster_labels (cluster_id, k, v) VALUES ($1, $2, $3)
ON CONFLICT (cluster_id, k) DO UPDATE SET v = EXCLUDED.v`,
		clusterID, "aegis.yourorg.dev/projectId", projectID)
	if err != nil {
		s.logExecError("set_cluster_project", err, zap.String("cluster_id", clusterID), zap.String("project_id", projectID))
	}
}

func (s *PostgresStore) DeleteCluster(clusterID string) {
	if clusterID == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	if _, err := s.pool.Exec(ctx, `UPDATE clusters SET deleted_at=now(), updated_at=now() WHERE id=$1 AND deleted_at IS NULL`, clusterID); err != nil {
		s.logExecError("delete_cluster", err, zap.String("cluster_id", clusterID))
	}
}

// CleanupStaleClusters soft-deletes clusters that haven't sent a heartbeat in the specified duration.
// This prevents stale clusters from accumulating in the database.
func (s *PostgresStore) CleanupStaleClusters(staleThreshold string) int64 {
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	// Default to 1 hour if not specified
	if staleThreshold == "" {
		staleThreshold = "1 hour"
	}

	result, err := s.pool.Exec(ctx, `UPDATE clusters SET deleted_at=now(), updated_at=now(), deleted_by='system', deletion_reason='stale heartbeat' WHERE deleted_at IS NULL AND last_heartbeat IS NOT NULL AND last_heartbeat < NOW() - $1::interval`, staleThreshold)
	if err != nil {
		s.logExecError("cleanup_stale_clusters", err)
		return 0
	}
	return result.RowsAffected()
}
