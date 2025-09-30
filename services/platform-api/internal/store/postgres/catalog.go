package postgres

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

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
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()
	_, err := s.pool.Exec(ctx, `
INSERT INTO projects (id, display_name, owner_group, policy_regions, policy_data_level, policy_deny_egress_by_default, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, now(), now())
ON CONFLICT (id) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    owner_group = EXCLUDED.owner_group,
    policy_regions = EXCLUDED.policy_regions,
    policy_data_level = EXCLUDED.policy_data_level,
    policy_deny_egress_by_default = EXCLUDED.policy_deny_egress_by_default,
    updated_at = now()
`, p.GetId(), nullableString(p.GetDisplayName()), p.GetOwnerGroup(), regions, nullableString(dataLevel), denyEgress)
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
SELECT id, COALESCE(display_name, ''), owner_group, policy_regions, COALESCE(policy_data_level, ''), policy_deny_egress_by_default
FROM projects WHERE id=$1
`, id)
	var (
		projID     string
		display    string
		owner      string
		regions    []string
		dataLevel  string
		denyEgress bool
	)
	if err := row.Scan(&projID, &display, &owner, &regions, &dataLevel, &denyEgress); err != nil {
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
	return project
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
