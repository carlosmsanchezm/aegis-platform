package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

// Annotation key validation constants
const (
	annotationPrefix = "aegis.yourorg.dev/"
)

// knownAnnotationKeys maps valid camelCase annotation keys to their purpose.
// This ensures consistency across the codebase.
var knownAnnotationKeys = map[string]string{
	"aegis.yourorg.dev/awsRoleArn":    "AWS IAM role ARN for assuming cross-account access",
	"aegis.yourorg.dev/awsAccountId":  "AWS account ID for the project",
	"aegis.yourorg.dev/awsExternalId": "External ID for STS AssumeRole",
	"aegis.yourorg.dev/projectId":     "Project ID for cluster association",
	"aegis.yourorg.dev/ilLevel":       "Information Level (IL) classification",
	"aegis.yourorg.dev/environment":   "Environment (dev, staging, prod)",
	"aegis.yourorg.dev/description":   "Human-readable description",
	"aegis.yourorg.dev/enable_fips":   "Enable FIPS 140-2 compliance",
	"aegis/description":               "Legacy description annotation",
	"aegis/environment":               "Legacy environment annotation",
	"aegis/enable_fips":               "Legacy FIPS annotation",
	"aegis/monthly_budget":            "Monthly budget in USD",
	"aegis/network_isolation":         "Network isolation enabled",
	"aegis/monthly_alert_percent":     "Budget alert threshold percentage",
	"aegis/compute_profiles":          "JSON array of compute profile configurations",
}

// invalidAnnotationKeyPatterns detects common mistakes in annotation keys
var invalidAnnotationKeyPatterns = []*regexp.Regexp{
	// Kebab-case after prefix (should be camelCase)
	regexp.MustCompile(`^aegis\.yourorg\.dev/[a-z]+-[a-z]+`),
}

// normalizeAnnotationKey attempts to fix common annotation key mistakes.
// Returns the normalized key and whether normalization was applied.
func normalizeAnnotationKey(key string) (string, bool) {
	// Check for kebab-case AWS keys and convert to camelCase
	kebabToCamel := map[string]string{
		"aegis.yourorg.dev/aws-role-arn":    "aegis.yourorg.dev/awsRoleArn",
		"aegis.yourorg.dev/aws-account-id":  "aegis.yourorg.dev/awsAccountId",
		"aegis.yourorg.dev/aws-external-id": "aegis.yourorg.dev/awsExternalId",
		"aegis.yourorg.dev/project-id":      "aegis.yourorg.dev/projectId",
		"aegis.yourorg.dev/il-level":        "aegis.yourorg.dev/ilLevel",
	}

	if normalized, ok := kebabToCamel[key]; ok {
		return normalized, true
	}
	return key, false
}

// normalizeAnnotations fixes common annotation key mistakes and returns the normalized map.
// It also logs warnings for normalized keys.
func normalizeAnnotations(annotations map[string]string, log *zap.Logger, projectID string) map[string]string {
	if len(annotations) == 0 {
		return annotations
	}

	normalized := make(map[string]string, len(annotations))
	for key, value := range annotations {
		newKey, wasNormalized := normalizeAnnotationKey(key)
		if wasNormalized && log != nil {
			log.Warn("normalized annotation key from kebab-case to camelCase",
				zap.String("project_id", projectID),
				zap.String("original_key", key),
				zap.String("normalized_key", newKey))
		}
		normalized[newKey] = value
	}
	return normalized
}

func (s *PostgresStore) PutProject(p *aegis.Project) {
	if p == nil || strings.TrimSpace(p.GetId()) == "" {
		return
	}
	policy := p.GetPolicy()
	regions := []string(nil)
	dataLevel := ""
	denyEgress := false
	if policy != nil {
		regions = append(regions, policy.GetRegions()...)
		dataLevel = policy.GetDataLevel()
		denyEgress = policy.GetDenyEgressByDefault()
	}

	// Normalize annotation keys (fix kebab-case to camelCase)
	annotations := normalizeAnnotations(p.GetAnnotations(), s.log, p.GetId())

	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	_, err := s.pool.Exec(ctx, `
INSERT INTO projects (id, display_name, owner_group, policy_regions, policy_data_level, policy_deny_egress_by_default, annotations, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, now(), now())
ON CONFLICT (id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    owner_group = EXCLUDED.owner_group,
    policy_regions = EXCLUDED.policy_regions,
    policy_data_level = EXCLUDED.policy_data_level,
    policy_deny_egress_by_default = EXCLUDED.policy_deny_egress_by_default,
    annotations = EXCLUDED.annotations,
    updated_at = now()
`, p.GetId(), nullableString(p.GetDisplayName()), p.GetOwnerGroup(), regions, nullableString(dataLevel), denyEgress, mapToJSONB(annotations))
	s.logExecError("upsert_project", err, zap.String("project_id", p.GetId()))
}

func (s *PostgresStore) GetProject(id string) *aegis.Project {
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	row := s.pool.QueryRow(ctx, `
SELECT id, COALESCE(display_name, ''), owner_group, policy_regions, COALESCE(policy_data_level, ''), policy_deny_egress_by_default, annotations
FROM projects WHERE id=$1
`, id)
	var (
		projID          string
		display         string
		owner           string
		regions         []string
		dataLevel       string
		denyEgress      bool
		annotationsJSON []byte
	)
	if err := row.Scan(&projID, &display, &owner, &regions, &dataLevel, &denyEgress, &annotationsJSON); err != nil {
		if err != pgx.ErrNoRows {
			s.logExecError("get_project", err, zap.String("project_id", id))
		}
		return nil
	}
	project := &aegis.Project{
		Id:          projID,
		DisplayName: display,
		OwnerGroup:  owner,
	}
	if len(regions) > 0 || dataLevel != "" || denyEgress {
		project.Policy = &aegis.PolicyDomain{
			Regions:             append([]string{}, regions...),
			DataLevel:           dataLevel,
			DenyEgressByDefault: denyEgress,
		}
	}
	if len(annotationsJSON) > 0 {
		var annotations map[string]string
		if err := json.Unmarshal(annotationsJSON, &annotations); err == nil {
			project.Annotations = annotations
		}
	}
	return project
}

func (s *PostgresStore) ListProjects() []*aegis.Project {
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	rows, err := s.pool.Query(ctx, `
SELECT id, COALESCE(display_name, ''), owner_group, policy_regions, COALESCE(policy_data_level, ''), policy_deny_egress_by_default, annotations
FROM projects
ORDER BY id
`)
	if err != nil {
		s.logExecError("list_projects", err)
		return nil
	}
	defer rows.Close()
	var items []*aegis.Project
	for rows.Next() {
		var (
			projID          string
			display         string
			owner           string
			regions         []string
			dataLevel       string
			denyEgress      bool
			annotationsJSON []byte
		)
		if err := rows.Scan(&projID, &display, &owner, &regions, &dataLevel, &denyEgress, &annotationsJSON); err != nil {
			s.logExecError("scan_project", err)
			continue
		}
		project := &aegis.Project{
			Id:          projID,
			DisplayName: display,
			OwnerGroup:  owner,
		}
		if len(regions) > 0 || dataLevel != "" || denyEgress {
			project.Policy = &aegis.PolicyDomain{
				Regions:             append([]string{}, regions...),
				DataLevel:           dataLevel,
				DenyEgressByDefault: denyEgress,
			}
		}
		if len(annotationsJSON) > 0 {
			var annotations map[string]string
			if err := json.Unmarshal(annotationsJSON, &annotations); err == nil {
				project.Annotations = annotations
			}
		}
		items = append(items, project)
	}
	return items
}

// DeleteProject removes a project from the database.
// It returns an error if the project has active (non-deleted) clusters attached,
// as those clusters would incur costs and should be deleted first.
func (s *PostgresStore) DeleteProject(id string) error {
	id = strings.TrimSpace(id)
	if id == "" {
		return errors.New("project id is required")
	}

	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	// Check if project has any active clusters
	var clusterCount int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM clusters WHERE project_id = $1 AND deleted_at IS NULL`, id).Scan(&clusterCount)
	if err != nil {
		s.logExecError("delete_project_check_clusters", err, zap.String("project_id", id))
		return fmt.Errorf("failed to check for active clusters: %w", err)
	}

	if clusterCount > 0 {
		return fmt.Errorf("cannot delete project %q: %d active cluster(s) still attached - delete clusters first to avoid incurring costs", id, clusterCount)
	}

	// Safe to delete - no active clusters
	result, err := s.pool.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		s.logExecError("delete_project", err, zap.String("project_id", id))
		return fmt.Errorf("failed to delete project: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("project %q not found", id)
	}

	s.log.Info("project deleted", zap.String("project_id", id))
	return nil
}

// HasActiveClusters checks if a project has any non-deleted clusters attached.
func (s *PostgresStore) HasActiveClusters(projectID string) bool {
	if projectID == "" {
		return false
	}

	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	var count int
	err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM clusters WHERE project_id = $1 AND deleted_at IS NULL`, projectID).Scan(&count)
	if err != nil {
		s.logExecError("has_active_clusters", err, zap.String("project_id", projectID))
		return false
	}

	return count > 0
}

func mapToJSONB(m map[string]string) []byte {
	if len(m) == 0 {
		return nil
	}
	data, err := json.Marshal(m)
	if err != nil {
		return nil
	}
	return data
}

func (s *PostgresStore) PutBudget(b *aegis.Budget) {
	if b == nil || strings.TrimSpace(b.GetProjectId()) == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	_, err := s.pool.Exec(ctx, `
INSERT INTO budgets (project_id, queue, limit_usd, policy_mode, updated_at)
VALUES ($1, COALESCE($2, ''), $3, UPPER($4), now())
ON CONFLICT (project_id, queue) DO UPDATE SET
    limit_usd = EXCLUDED.limit_usd,
    policy_mode = EXCLUDED.policy_mode,
    updated_at = now()
`, b.GetProjectId(), nullableString(b.GetQueue()), b.GetLimitUsd(), strings.ToUpper(b.GetPolicyMode()))
	s.logExecError("upsert_budget", err, zap.String("project_id", b.GetProjectId()), zap.String("queue", b.GetQueue()))
}

func (s *PostgresStore) GetBudget(projectID string) *aegis.Budget {
	return s.GetBudgetExact(projectID, "")
}

func (s *PostgresStore) GetBudgetExact(projectID, queue string) *aegis.Budget {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	row := s.pool.QueryRow(ctx, `
SELECT project_id, queue, limit_usd, policy_mode
FROM budgets WHERE project_id=$1 AND queue=COALESCE($2, '')
`, projectID, nullableString(queue))
	var (
		pid   string
		q     string
		limit float64
		mode  string
	)
	if err := row.Scan(&pid, &q, &limit, &mode); err != nil {
		if err != pgx.ErrNoRows {
			s.logExecError("get_budget_exact", err, zap.String("project_id", projectID), zap.String("queue", queue))
		}
		return nil
	}
	return &aegis.Budget{ProjectId: pid, Queue: q, LimitUsd: limit, PolicyMode: mode}
}

func (s *PostgresStore) ResolveBudget(projectID, queue string) (*aegis.Budget, string) {
	if resolved := s.GetBudgetExact(projectID, queue); resolved != nil {
		return resolved, budgetKey(projectID, resolved.GetQueue())
	}
	if fallback := s.GetBudgetExact(projectID, ""); fallback != nil {
		return fallback, budgetKey(projectID, fallback.GetQueue())
	}
	return nil, ""
}

func budgetKey(projectID, queue string) string {
	return strings.TrimSpace(projectID) + "|" + strings.TrimSpace(queue)
}

func (s *PostgresStore) ListBudgets(filterProject string) []*aegis.Budget {
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	var (
		rows pgx.Rows
		err  error
	)
	if strings.TrimSpace(filterProject) == "" {
		rows, err = s.pool.Query(ctx, `SELECT project_id, queue, limit_usd, policy_mode FROM budgets ORDER BY project_id, queue`)
	} else {
		rows, err = s.pool.Query(ctx, `SELECT project_id, queue, limit_usd, policy_mode FROM budgets WHERE project_id=$1 ORDER BY project_id, queue`, strings.TrimSpace(filterProject))
	}
	if err != nil {
		s.logExecError("list_budgets", err, zap.String("project_id", filterProject))
		return nil
	}
	defer rows.Close()
	items := []*aegis.Budget{}
	for rows.Next() {
		var (
			pid   string
			q     string
			limit float64
			mode  string
		)
		if scanErr := rows.Scan(&pid, &q, &limit, &mode); scanErr != nil {
			s.logExecError("scan_budget", scanErr)
			return items
		}
		items = append(items, &aegis.Budget{ProjectId: pid, Queue: q, LimitUsd: limit, PolicyMode: mode})
	}
	return items
}

func (s *PostgresStore) PutFlavor(f *aegis.Flavor) {
	if f == nil || strings.TrimSpace(f.GetName()) == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	_, err := s.pool.Exec(ctx, `
INSERT INTO flavors (name, chip, mig_profile, rdma_required, gpu_count, memory_gib, resource_name, cpu_cores_request, memory_request, price_usd_per_gpu_hour, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, now())
ON CONFLICT (name) DO UPDATE SET
    chip = EXCLUDED.chip,
    mig_profile = EXCLUDED.mig_profile,
    rdma_required = EXCLUDED.rdma_required,
    gpu_count = EXCLUDED.gpu_count,
    memory_gib = EXCLUDED.memory_gib,
    resource_name = EXCLUDED.resource_name,
    cpu_cores_request = EXCLUDED.cpu_cores_request,
    memory_request = EXCLUDED.memory_request,
    price_usd_per_gpu_hour = EXCLUDED.price_usd_per_gpu_hour,
    updated_at = now()
`, f.GetName(), f.GetChip(), nullableString(f.GetMigProfile()), f.GetRdmaRequired(), f.GetGpuCount(), f.GetMemoryGib(), nullableString(f.GetResourceName()), nullableString(f.GetCpuCoresRequest()), nullableString(f.GetMemoryRequest()), f.GetPriceUsdPerGpuHour())
	s.logExecError("upsert_flavor", err, zap.String("flavor", f.GetName()))
}

func (s *PostgresStore) GetFlavor(name string) *aegis.Flavor {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	row := s.pool.QueryRow(ctx, `
SELECT name, chip, COALESCE(mig_profile, ''), rdma_required, gpu_count, COALESCE(memory_gib, 0), COALESCE(resource_name, ''), COALESCE(cpu_cores_request, ''), COALESCE(memory_request, ''), price_usd_per_gpu_hour
FROM flavors WHERE name=$1
`, name)
	var (
		fname     string
		chip      string
		mig       string
		rdma      bool
		gpuCount  int32
		memoryGiB float64
		resource  string
		cpuReq    string
		memReq    string
		price     float64
	)
	if err := row.Scan(&fname, &chip, &mig, &rdma, &gpuCount, &memoryGiB, &resource, &cpuReq, &memReq, &price); err != nil {
		if err != pgx.ErrNoRows {
			s.logExecError("get_flavor", err, zap.String("flavor", name))
		}
		return nil
	}
	return &aegis.Flavor{
		Name:               fname,
		Chip:               chip,
		MigProfile:         mig,
		RdmaRequired:       rdma,
		ResourceName:       resource,
		GpuCount:           gpuCount,
		MemoryGib:          memoryGiB,
		CpuCoresRequest:    cpuReq,
		MemoryRequest:      memReq,
		PriceUsdPerGpuHour: price,
	}
}

func (s *PostgresStore) PutQueue(q *aegis.Queue) {
	if q == nil || strings.TrimSpace(q.GetName()) == "" {
		return
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	_, err := s.pool.Exec(ctx, `
INSERT INTO queues (name, project_id, priority_tier, allowed_flavors, default_max_duration_seconds, updated_at)
VALUES ($1, $2, $3, $4, $5, now())
ON CONFLICT (name) DO UPDATE SET
    project_id = EXCLUDED.project_id,
    priority_tier = EXCLUDED.priority_tier,
    allowed_flavors = EXCLUDED.allowed_flavors,
    default_max_duration_seconds = EXCLUDED.default_max_duration_seconds,
    updated_at = now()
`, q.GetName(), q.GetProjectId(), nullableString(q.GetPriorityTier()), stringSliceOrNil(q.GetAllowedFlavors()), nullableInt64Ptr(q.GetDefaultMaxDurationSeconds()))
	s.logExecError("upsert_queue", err, zap.String("queue", q.GetName()))
}

func (s *PostgresStore) GetQueue(name string) *aegis.Queue {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	row := s.pool.QueryRow(ctx, `
SELECT name, project_id, COALESCE(priority_tier, ''), allowed_flavors, COALESCE(default_max_duration_seconds, 0)
FROM queues WHERE name=$1
`, name)
	var (
		qName     string
		projectID string
		tier      string
		allowed   []string
		maxDur    int64
	)
	if err := row.Scan(&qName, &projectID, &tier, &allowed, &maxDur); err != nil {
		if err != pgx.ErrNoRows {
			s.logExecError("get_queue", err, zap.String("queue", name))
		}
		return nil
	}
	return &aegis.Queue{
		Name:                      qName,
		ProjectId:                 projectID,
		PriorityTier:              tier,
		AllowedFlavors:            append([]string{}, allowed...),
		DefaultMaxDurationSeconds: maxDur,
	}
}

func nullableString(in string) interface{} {
	trimmed := strings.TrimSpace(in)
	if trimmed == "" {
		return nil
	}
	return trimmed
}

func stringSliceOrNil(in []string) interface{} {
	if len(in) == 0 {
		return nil
	}
	out := make([]string, len(in))
	copy(out, in)
	return out
}

func nullableInt64Ptr(in int64) interface{} {
	if in == 0 {
		return nil
	}
	return in
}
