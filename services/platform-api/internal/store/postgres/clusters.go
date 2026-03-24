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
  cluster_endpoint,
  cluster_ca,
  created_at,
  updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now(), now())
ON CONFLICT (id) DO UPDATE SET
  provider=EXCLUDED.provider,
  region=EXCLUDED.region,
  import_method=EXCLUDED.import_method,
  imported_at=EXCLUDED.imported_at,
  kubeconfig_secret_ref=EXCLUDED.kubeconfig_secret_ref,
  assume_role_arn=EXCLUDED.assume_role_arn,
  cluster_endpoint=COALESCE(NULLIF(EXCLUDED.cluster_endpoint, ''), clusters.cluster_endpoint),
  cluster_ca=COALESCE(NULLIF(EXCLUDED.cluster_ca, ''), clusters.cluster_ca),
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
		nullableString(req.ClusterEndpoint),
		nullableString(req.ClusterCA),
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

func (s *PostgresStore) GetClusterProjectID(clusterID string) (string, bool) {
	clusterID = strings.TrimSpace(clusterID)
	if clusterID == "" {
		return "", false
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	var projectID string
	err := s.pool.QueryRow(ctx, `SELECT cl.v
FROM cluster_labels cl
JOIN clusters c ON cl.cluster_id = c.id
WHERE c.deleted_at IS NULL AND cl.cluster_id=$1 AND cl.k=$2`,
		clusterID, "aegis.yourorg.dev/projectId",
	).Scan(&projectID)
	if err != nil {
		return "", false
	}
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return "", false
	}
	return projectID, true
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

	// Extract project_id from labels if present (for multi-tenancy FK)
	var projectID *string
	if labels := req.GetLabels(); labels != nil {
		if pid, ok := labels["aegis.yourorg.dev/projectId"]; ok && pid != "" {
			projectID = &pid
		}
	}

	// Insert or update cluster, including proxy_url and project_id if provided.
	// Use COALESCE to handle empty proxy_url gracefully (column has NOT NULL constraint with default '').
	// Clear deleted_at and stale_since so re-provisioned clusters are resurrected.
	proxyURL := strings.TrimSpace(req.GetProxyUrl())
	proxyCAPem := strings.TrimSpace(req.GetProxyCaPem())
	_, err = tx.Exec(ctx, `INSERT INTO clusters (id, project_id, provider, region, proxy_url, proxy_ca_pem, created_at, updated_at)
VALUES ($1, $5, $2, $3, COALESCE(NULLIF($4, ''), ''), COALESCE(NULLIF($6, ''), ''), now(), now())
ON CONFLICT (id) DO UPDATE SET
    provider = COALESCE(NULLIF(EXCLUDED.provider, ''), clusters.provider),
    region = COALESCE(NULLIF(EXCLUDED.region, ''), clusters.region),
    project_id = COALESCE(EXCLUDED.project_id, clusters.project_id),
    proxy_url = CASE WHEN $4 <> '' THEN $4 ELSE clusters.proxy_url END,
    proxy_ca_pem = CASE WHEN $6 <> '' THEN $6 ELSE clusters.proxy_ca_pem END,
    deleted_at = NULL,
    stale_since = NULL,
    updated_at = now()`,
		req.GetClusterId(), nullableString(req.GetProvider()), nullableString(req.GetRegion()), proxyURL, projectID, proxyCAPem)
	if err != nil {
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

	// Persist IL level as a label so it survives round-trips through the DB.
	if ilLevel := strings.TrimSpace(req.GetIlLevel()); ilLevel != "" {
		if _, err := tx.Exec(ctx, `INSERT INTO cluster_labels (cluster_id, k, v) VALUES ($1, $2, $3)
ON CONFLICT (cluster_id, k) DO UPDATE SET v = EXCLUDED.v`,
			req.GetClusterId(), "aegis.yourorg.dev/ilLevel", strings.ToUpper(ilLevel)); err != nil {
			s.logExecError("cluster_register_insert_il_level", err, zap.String("cluster_id", req.GetClusterId()))
			return
		}
	}

	if err := tx.Commit(ctx); err != nil {
		s.logExecError("cluster_register_commit", err, zap.String("cluster_id", req.GetClusterId()))
	}
}

// PreRegisterCluster creates a placeholder cluster row during provisioning.
// This is called by the Pulumi runner after cluster creation but before the k8s-agent connects.
// The k8s-agent's RegisterCluster call will then update this row rather than failing.
func (s *PostgresStore) PreRegisterCluster(clusterID, projectID, provider, region, proxyURL, endpoint, ca string) error {
	if clusterID == "" {
		return nil
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	// Use ON CONFLICT to handle race conditions if cluster already exists
	_, err := s.pool.Exec(ctx, `INSERT INTO clusters (id, project_id, provider, region, proxy_url, cluster_endpoint, cluster_ca, created_at, updated_at)
VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''), COALESCE(NULLIF($5, ''), ''), NULLIF($6, ''), NULLIF($7, ''), now(), now())
ON CONFLICT (id) DO UPDATE SET
    project_id = COALESCE(clusters.project_id, EXCLUDED.project_id),
    provider = COALESCE(NULLIF(EXCLUDED.provider, ''), clusters.provider),
    region = COALESCE(NULLIF(EXCLUDED.region, ''), clusters.region),
    proxy_url = CASE WHEN EXCLUDED.proxy_url <> '' THEN EXCLUDED.proxy_url ELSE clusters.proxy_url END,
    cluster_endpoint = COALESCE(NULLIF(EXCLUDED.cluster_endpoint, ''), clusters.cluster_endpoint),
    cluster_ca = COALESCE(NULLIF(EXCLUDED.cluster_ca, ''), clusters.cluster_ca),
    deleted_at = NULL,
    stale_since = NULL,
    updated_at = now()`,
		clusterID, projectID, provider, region, proxyURL, endpoint, ca)
	// Note: proxy_ca_pem is populated later via RegisterCluster when the k8s-agent sends it.
	if err != nil {
		s.logExecError("cluster_pre_register", err, zap.String("cluster_id", clusterID), zap.String("project_id", projectID))
		return err
	}

	// Also insert the project label for this cluster
	if projectID != "" {
		_, err = s.pool.Exec(ctx, `INSERT INTO cluster_labels (cluster_id, k, v)
VALUES ($1, 'aegis.yourorg.dev/projectId', $2)
ON CONFLICT (cluster_id, k) DO UPDATE SET v = EXCLUDED.v`,
			clusterID, projectID)
		if err != nil {
			s.logExecError("cluster_pre_register_label", err, zap.String("cluster_id", clusterID), zap.String("project_id", projectID))
			// Don't fail for label errors
		}
	}

	s.log.Info("pre-registered cluster",
		zap.String("cluster_id", clusterID),
		zap.String("project_id", projectID),
		zap.String("provider", provider),
		zap.String("region", region),
		zap.String("proxy_url", proxyURL))
	return nil
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

	// Update TTFG metric, proxy URL, and last heartbeat timestamp.
	// Clear stale_since so the two-phase cleanup won't delete this cluster.
	// Do not resurrect soft-deleted clusters.
	res, err := tx.Exec(ctx, `UPDATE clusters
SET ttf_gpu_seconds_p50=$1,
    proxy_url = CASE WHEN $3 <> '' THEN $3 ELSE proxy_url END,
    last_heartbeat=now(),
    stale_since=NULL,
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
		id              string
		projectID       sql.NullString
		provider        string
		region          string
		importMethod    string
		importedAt      sql.NullTime
		kubeconfig      sql.NullString
		assumeRole      sql.NullString
		clusterEndpoint sql.NullString
		clusterCA       sql.NullString
		ttf             float64
		proxyURL        sql.NullString
		proxyCAPem      sql.NullString
		heartbeat       sql.NullTime
		createdAt       time.Time
		deletedAt       sql.NullTime
	)
	err := s.pool.QueryRow(ctx, `SELECT
  id,
  project_id,
  COALESCE(provider, ''),
  COALESCE(region, ''),
  COALESCE(import_method, 'provisioned'),
  imported_at,
  kubeconfig_secret_ref,
  assume_role_arn,
  cluster_endpoint,
  cluster_ca,
  ttf_gpu_seconds_p50,
  COALESCE(proxy_url, ''),
  COALESCE(proxy_ca_pem, ''),
  last_heartbeat,
  created_at,
  deleted_at
FROM clusters WHERE id=$1 AND deleted_at IS NULL`, clusterID).Scan(
		&id,
		&projectID,
		&provider,
		&region,
		&importMethod,
		&importedAt,
		&kubeconfig,
		&assumeRole,
		&clusterEndpoint,
		&clusterCA,
		&ttf,
		&proxyURL,
		&proxyCAPem,
		&heartbeat,
		&createdAt,
		&deletedAt,
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
	if clusterEndpoint.Valid {
		info.ClusterEndpoint = clusterEndpoint.String
	}
	if clusterCA.Valid {
		info.ClusterCA = clusterCA.String
	}
	if projectID.Valid {
		info.ProjectID = projectID.String
	}
	if heartbeat.Valid {
		info.LastHeartbeat = heartbeat.Time.UTC()
	}
	if proxyURL.Valid {
		info.ProxyURL = proxyURL.String
	}
	if proxyCAPem.Valid {
		info.ProxyCAPem = proxyCAPem.String
	}
	if deletedAt.Valid {
		t := deletedAt.Time.UTC()
		info.DeletedAt = &t
	}

	// Load labels for the cluster.
	labelRows, lErr := s.pool.Query(ctx, `SELECT k, v FROM cluster_labels WHERE cluster_id=$1`, clusterID)
	if lErr == nil {
		defer labelRows.Close()
		for labelRows.Next() {
			var k, v string
			if err := labelRows.Scan(&k, &v); err != nil {
				break
			}
			info.Labels[k] = v
			if k == "aegis.yourorg.dev/ilLevel" && strings.TrimSpace(v) != "" {
				info.ILLevel = strings.ToUpper(strings.TrimSpace(v))
			}
		}
	}

	return info
}

func (s *PostgresStore) ListClusterInfos() []*store.ClusterInfo {
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	rows, err := s.pool.Query(ctx, `SELECT
  id,
  project_id,
  COALESCE(provider, ''),
  COALESCE(region, ''),
  COALESCE(import_method, 'provisioned'),
  imported_at,
  kubeconfig_secret_ref,
  assume_role_arn,
  cluster_endpoint,
  cluster_ca,
  ttf_gpu_seconds_p50,
  COALESCE(proxy_url, ''),
  COALESCE(proxy_ca_pem, ''),
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
			id              string
			projectID       sql.NullString
			provider        string
			region          string
			importMethod    string
			importedAt      sql.NullTime
			kubeconfig      sql.NullString
			assumeRole      sql.NullString
			clusterEndpoint sql.NullString
			clusterCA       sql.NullString
			ttf             float64
			proxyURL        sql.NullString
			proxyCAPem      sql.NullString
			heartbeat       sql.NullTime
			createdAt       time.Time
		)
		if err := rows.Scan(&id, &projectID, &provider, &region, &importMethod, &importedAt, &kubeconfig, &assumeRole, &clusterEndpoint, &clusterCA, &ttf, &proxyURL, &proxyCAPem, &heartbeat, &createdAt); err != nil {
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
		if clusterEndpoint.Valid {
			info.ClusterEndpoint = clusterEndpoint.String
		}
		if clusterCA.Valid {
			info.ClusterCA = clusterCA.String
		}
		if projectID.Valid {
			info.ProjectID = projectID.String
		}
		if heartbeat.Valid {
			info.LastHeartbeat = heartbeat.Time.UTC()
		}
		if proxyURL.Valid {
			info.ProxyURL = proxyURL.String
		}
		if proxyCAPem.Valid {
			info.ProxyCAPem = proxyCAPem.String
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
				// Hydrate ILLevel from the well-known label.
				if k == "aegis.yourorg.dev/ilLevel" && strings.TrimSpace(v) != "" {
					info.ILLevel = strings.ToUpper(strings.TrimSpace(v))
				}
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

// ListClustersByProject returns all non-deleted clusters for a specific project.
// This is the primary multi-tenancy query for cluster isolation.
func (s *PostgresStore) ListClustersByProject(projectID string) []*store.ClusterInfo {
	if projectID == "" {
		return nil
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	rows, err := s.pool.Query(ctx, `SELECT id, project_id, COALESCE(provider, ''), COALESCE(region, ''), ttf_gpu_seconds_p50, COALESCE(proxy_url, ''), cluster_endpoint, cluster_ca, last_heartbeat, created_at FROM clusters WHERE project_id = $1 AND deleted_at IS NULL`, projectID)
	if err != nil {
		s.logExecError("cluster_list_by_project", err, zap.String("project_id", projectID))
		return nil
	}
	defer rows.Close()

	clusters := map[string]*store.ClusterInfo{}
	for rows.Next() {
		var (
			id              string
			projID          sql.NullString
			provider        string
			region          string
			ttf             float64
			proxyURL        sql.NullString
			clusterEndpoint sql.NullString
			clusterCA       sql.NullString
			heartbeat       sql.NullTime
			createdAt       time.Time
		)
		if err := rows.Scan(&id, &projID, &provider, &region, &ttf, &proxyURL, &clusterEndpoint, &clusterCA, &heartbeat, &createdAt); err != nil {
			s.logExecError("cluster_list_by_project_scan", err)
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
		}
		if projID.Valid {
			info.ProjectID = projID.String
		}
		if heartbeat.Valid {
			info.LastHeartbeat = heartbeat.Time.UTC()
		}
		if proxyURL.Valid {
			info.ProxyURL = proxyURL.String
		}
		if clusterEndpoint.Valid {
			info.ClusterEndpoint = clusterEndpoint.String
		}
		if clusterCA.Valid {
			info.ClusterCA = clusterCA.String
		}
		clusters[id] = info
	}

	// Load labels for the filtered clusters
	if len(clusters) > 0 {
		clusterIDs := make([]string, 0, len(clusters))
		for id := range clusters {
			clusterIDs = append(clusterIDs, id)
		}

		labelRows, err := s.pool.Query(ctx, `SELECT cluster_id, k, v FROM cluster_labels WHERE cluster_id = ANY($1)`, clusterIDs)
		if err == nil {
			defer labelRows.Close()
			for labelRows.Next() {
				var id, k, v string
				if err := labelRows.Scan(&id, &k, &v); err != nil {
					s.logExecError("cluster_list_by_project_scan_labels", err)
					break
				}
				if info, ok := clusters[id]; ok {
					info.Labels[k] = v
					if k == "aegis.yourorg.dev/ilLevel" && strings.TrimSpace(v) != "" {
						info.ILLevel = strings.ToUpper(strings.TrimSpace(v))
					}
				}
			}
		}

		// Load flavors for the filtered clusters
		flavorRows, err := s.pool.Query(ctx, `SELECT cluster_id, flavor FROM cluster_flavors WHERE cluster_id = ANY($1)`, clusterIDs)
		if err == nil {
			defer flavorRows.Close()
			for flavorRows.Next() {
				var id, flavor string
				if err := flavorRows.Scan(&id, &flavor); err != nil {
					s.logExecError("cluster_list_by_project_scan_flavors", err)
					break
				}
				if info, ok := clusters[id]; ok {
					info.AvailableFlavorSet[flavor] = true
				}
			}
		}
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

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logExecError("set_cluster_project_begin", err)
		return
	}
	defer tx.Rollback(ctx)

	// Update the project_id column directly (new FK-based approach)
	if _, err := tx.Exec(ctx, `UPDATE clusters SET project_id = $2, updated_at = now() WHERE id = $1`,
		clusterID, projectID); err != nil {
		s.logExecError("set_cluster_project_update", err, zap.String("cluster_id", clusterID), zap.String("project_id", projectID))
		return
	}

	// Also maintain the label for backward compatibility
	if _, err := tx.Exec(ctx, `INSERT INTO cluster_labels (cluster_id, k, v) VALUES ($1, $2, $3)
ON CONFLICT (cluster_id, k) DO UPDATE SET v = EXCLUDED.v`,
		clusterID, "aegis.yourorg.dev/projectId", projectID); err != nil {
		s.logExecError("set_cluster_project_label", err, zap.String("cluster_id", clusterID), zap.String("project_id", projectID))
		return
	}

	if err := tx.Commit(ctx); err != nil {
		s.logExecError("set_cluster_project_commit", err, zap.String("cluster_id", clusterID), zap.String("project_id", projectID))
	}
}

func (s *PostgresStore) SetClusterLabel(clusterID, key, value string) {
	if clusterID == "" || key == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	if _, err := s.pool.Exec(ctx, `INSERT INTO cluster_labels (cluster_id, k, v) VALUES ($1, $2, $3)
ON CONFLICT (cluster_id, k) DO UPDATE SET v = EXCLUDED.v`,
		clusterID, key, value); err != nil {
		s.logExecError("set_cluster_label", err, zap.String("cluster_id", clusterID), zap.String("key", key))
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

// CleanupStaleClusters uses two-phase detection to avoid false positives during
// hub pod restarts (when spokes temporarily can't heartbeat).
//
// Phase 1 — Mark: clusters whose last_heartbeat is older than the threshold AND
// that are not yet marked get stale_since set to now(). This starts the grace clock.
//
// Phase 2 — Delete: clusters that have been continuously stale (stale_since is set
// AND stale_since itself is older than the threshold) get soft-deleted.
//
// If a heartbeat arrives between Phase 1 and Phase 2 (see UpdateClusterFromHeartbeat),
// stale_since is cleared and the cluster survives.
func (s *PostgresStore) CleanupStaleClusters(staleThreshold string) int64 {
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	if staleThreshold == "" {
		staleThreshold = "1 hour"
	}

	// Phase 1: Mark newly stale clusters (heartbeat old, not yet marked).
	if _, err := s.pool.Exec(ctx,
		`UPDATE clusters
		 SET stale_since = now(), updated_at = now()
		 WHERE deleted_at IS NULL
		   AND stale_since IS NULL
		   AND last_heartbeat IS NOT NULL
		   AND last_heartbeat < NOW() - $1::interval`,
		staleThreshold); err != nil {
		s.logExecError("mark_stale_clusters", err)
	}

	// Phase 2: Delete clusters that have been stale for the full threshold.
	result, err := s.pool.Exec(ctx,
		`UPDATE clusters
		 SET deleted_at = now(), updated_at = now(),
		     deleted_by = 'system', deletion_reason = 'stale heartbeat'
		 WHERE deleted_at IS NULL
		   AND stale_since IS NOT NULL
		   AND stale_since < NOW() - $1::interval`,
		staleThreshold)
	if err != nil {
		s.logExecError("cleanup_stale_clusters", err)
		return 0
	}
	return result.RowsAffected()
}
