package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"

	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

func (s *PostgresStore) ReserveIfAllowed(projectID, queue string, estimateUSD float64) (bool, string, string, store.BudgetUsageView) {
	projectID = strings.TrimSpace(projectID)
	queue = strings.TrimSpace(queue)
	if projectID == "" || estimateUSD <= 0 {
		return true, "", "", store.BudgetUsageView{}
	}

	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logExecError("reserve_begin", err, zap.String("project_id", projectID), zap.String("queue", queue))
		return false, "", "internal_error", store.BudgetUsageView{}
	}
	defer tx.Rollback(ctx)

	budget, err := s.lockBudget(ctx, tx, projectID, queue)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// No budget configured: allow by default.
			return true, "", "", store.BudgetUsageView{}
		}
		s.logExecError("reserve_lock_budget", err, zap.String("project_id", projectID), zap.String("queue", queue))
		return false, "", "internal_error", store.BudgetUsageView{}
	}

	period := monthStartUTC(time.Now().UTC())
	usage, err := s.lockUsage(ctx, tx, budget.projectID, budget.queue, period)
	if err != nil {
		s.logExecError("reserve_lock_usage", err, zap.String("project_id", budget.projectID), zap.String("queue", budget.queue))
		return false, strings.ToUpper(budget.policyMode), "internal_error", store.BudgetUsageView{}
	}

	next := usage.actualUSD + usage.reservedUSD + estimateUSD
	policy := strings.ToUpper(budget.policyMode)
	if policy == "HARD" && next > budget.limitUSD {
		view := usage.toView(period)
		return false, policy, "insufficient_funds", view
	}

	if _, err := tx.Exec(ctx, `UPDATE budget_usage SET reserved_usd = reserved_usd + $1, updated_at = now() WHERE project_id=$2 AND queue=$3 AND period_start_utc=$4`, estimateUSD, budget.projectID, budget.queue, period); err != nil {
		s.logExecError("reserve_update_usage", err, zap.String("project_id", budget.projectID), zap.String("queue", budget.queue))
		return false, policy, "internal_error", usage.toView(period)
	}

	if err := tx.Commit(ctx); err != nil {
		s.logExecError("reserve_commit", err, zap.String("project_id", budget.projectID), zap.String("queue", budget.queue))
		return false, policy, "internal_error", store.BudgetUsageView{}
	}

	usage.reservedUSD += estimateUSD
	return true, policy, "", usage.toView(period)
}

func (s *PostgresStore) ReconcileOnAck(projectID, queue string, estUSD, actualUSD float64) store.BudgetUsageView {
	projectID = strings.TrimSpace(projectID)
	queue = strings.TrimSpace(queue)
	if projectID == "" {
		return store.BudgetUsageView{}
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		s.logExecError("reconcile_begin", err, zap.String("project_id", projectID), zap.String("queue", queue))
		return store.BudgetUsageView{}
	}
	defer tx.Rollback(ctx)

	budget, err := s.lockBudget(ctx, tx, projectID, queue)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.BudgetUsageView{}
		}
		s.logExecError("reconcile_lock_budget", err, zap.String("project_id", projectID), zap.String("queue", queue))
		return store.BudgetUsageView{}
	}

	period := monthStartUTC(time.Now().UTC())
	usage, err := s.lockUsage(ctx, tx, budget.projectID, budget.queue, period)
	if err != nil {
		s.logExecError("reconcile_lock_usage", err, zap.String("project_id", budget.projectID), zap.String("queue", budget.queue))
		return store.BudgetUsageView{}
	}

	usage.reservedUSD -= estUSD
	if usage.reservedUSD < 0 {
		usage.reservedUSD = 0
	}
	usage.actualUSD += actualUSD
	if _, err := tx.Exec(ctx, `UPDATE budget_usage SET reserved_usd=$1, actual_usd=$2, updated_at=now() WHERE project_id=$3 AND queue=$4 AND period_start_utc=$5`, usage.reservedUSD, usage.actualUSD, budget.projectID, budget.queue, period); err != nil {
		s.logExecError("reconcile_update_usage", err, zap.String("project_id", budget.projectID), zap.String("queue", budget.queue))
		return store.BudgetUsageView{}
	}

	if err := tx.Commit(ctx); err != nil {
		s.logExecError("reconcile_commit", err, zap.String("project_id", budget.projectID), zap.String("queue", budget.queue))
		return store.BudgetUsageView{}
	}

	return usage.toView(period)
}

func (s *PostgresStore) UsageView(projectID, queue string) (store.BudgetUsageView, bool) {
	projectID = strings.TrimSpace(projectID)
	queue = strings.TrimSpace(queue)
	if projectID == "" {
		return store.BudgetUsageView{}, false
	}
	ctx, cancel := s.withTimeout(context.Background())
	defer cancel()

	period := monthStartUTC(time.Now().UTC())
	row := s.pool.QueryRow(ctx, `SELECT reserved_usd, actual_usd FROM budget_usage WHERE project_id=$1 AND queue=$2 AND period_start_utc=$3`, projectID, queue, period)
	var reserved, actual float64
	if err := row.Scan(&reserved, &actual); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return store.BudgetUsageView{ReservedUSD: 0, ActualUSD: 0, PeriodStart: period, PeriodEnd: monthEndUTC(period)}, true
		}
		s.logExecError("usage_view", err, zap.String("project_id", projectID), zap.String("queue", queue))
		return store.BudgetUsageView{}, false
	}
	return store.BudgetUsageView{ReservedUSD: reserved, ActualUSD: actual, PeriodStart: period, PeriodEnd: monthEndUTC(period)}, true
}

// budgetRecord is a lightweight view of an existing budget row.
type budgetRecord struct {
	projectID  string
	queue      string
	limitUSD   float64
	policyMode string
}

// usageRecord mirrors the budget_usage row held under lock.
type usageRecord struct {
	reservedUSD float64
	actualUSD   float64
}

func (u usageRecord) toView(periodStart time.Time) store.BudgetUsageView {
	return store.BudgetUsageView{
		ReservedUSD: u.reservedUSD,
		ActualUSD:   u.actualUSD,
		PeriodStart: periodStart,
		PeriodEnd:   monthEndUTC(periodStart),
	}
}

func (s *PostgresStore) lockBudget(ctx context.Context, tx pgx.Tx, projectID, queue string) (budgetRecord, error) {
	queues := []string{queue}
	if queue != "" {
		queues = append(queues, "")
	}
	for _, q := range queues {
		row := tx.QueryRow(ctx, `SELECT project_id, queue, limit_usd, policy_mode FROM budgets WHERE project_id=$1 AND queue=$2 FOR UPDATE`, projectID, q)
		var rec budgetRecord
		if err := row.Scan(&rec.projectID, &rec.queue, &rec.limitUSD, &rec.policyMode); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				continue
			}
			return budgetRecord{}, err
		}
		return rec, nil
	}
	return budgetRecord{}, pgx.ErrNoRows
}

func (s *PostgresStore) lockUsage(ctx context.Context, tx pgx.Tx, projectID, queue string, period time.Time) (usageRecord, error) {
	row := tx.QueryRow(ctx, `SELECT reserved_usd, actual_usd FROM budget_usage WHERE project_id=$1 AND queue=$2 AND period_start_utc=$3 FOR UPDATE`, projectID, queue, period)
	var rec usageRecord
	if err := row.Scan(&rec.reservedUSD, &rec.actualUSD); err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			return usageRecord{}, err
		}
		if _, err := tx.Exec(ctx, `INSERT INTO budget_usage (project_id, queue, period_start_utc, reserved_usd, actual_usd, updated_at) VALUES ($1, $2, $3, 0, 0, now())`, projectID, queue, period); err != nil {
			return usageRecord{}, err
		}
		row = tx.QueryRow(ctx, `SELECT reserved_usd, actual_usd FROM budget_usage WHERE project_id=$1 AND queue=$2 AND period_start_utc=$3 FOR UPDATE`, projectID, queue, period)
		if err := row.Scan(&rec.reservedUSD, &rec.actualUSD); err != nil {
			return usageRecord{}, err
		}
	}
	return rec, nil
}

func monthStartUTC(t time.Time) time.Time {
	utc := t.UTC()
	return time.Date(utc.Year(), utc.Month(), 1, 0, 0, 0, 0, time.UTC)
}

func monthEndUTC(t time.Time) time.Time {
	return monthStartUTC(t).AddDate(0, 1, 0)
}
