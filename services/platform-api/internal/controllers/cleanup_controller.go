package controllers

import (
	"context"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

// CleanupController periodically runs maintenance tasks to clean up stale data.
// It handles:
// - Soft-deleting clusters that haven't sent heartbeats
// - Purging expired connection sessions
// - Future: orphaned provisioning logs, stale budget records, etc.
type CleanupController struct {
	Log                    *zap.Logger
	Store                  store.Store
	StaleClusterThreshold  string        // PostgreSQL interval string, e.g., "1 hour"
	StaleWorkloadThreshold string        // PostgreSQL interval string, e.g., "24 hours"
	CleanupInterval        time.Duration // How often to run cleanup
}

// DefaultCleanupConfig returns a CleanupController with sensible defaults.
// Reads configuration from environment variables if available.
func DefaultCleanupConfig(log *zap.Logger, s store.Store) *CleanupController {
	threshold := os.Getenv("AEGIS_STALE_CLUSTER_THRESHOLD")
	if threshold == "" {
		threshold = "1 hour"
	}

	workloadThreshold := os.Getenv("AEGIS_STALE_WORKLOAD_THRESHOLD")
	if workloadThreshold == "" {
		workloadThreshold = "24 hours"
	}

	interval := 15 * time.Minute
	if v := os.Getenv("AEGIS_CLEANUP_INTERVAL"); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			interval = d
		}
	}

	return &CleanupController{
		Log:                    log.Named("cleanup"),
		Store:                  s,
		StaleClusterThreshold:  threshold,
		StaleWorkloadThreshold: workloadThreshold,
		CleanupInterval:        interval,
	}
}

// Start begins the cleanup loop. It blocks until the context is cancelled.
// Call this in a goroutine: go cleanup.Start(ctx)
func (c *CleanupController) Start(ctx context.Context) {
	if c.Store == nil {
		c.Log.Warn("cleanup controller disabled: no store configured")
		return
	}

	ticker := time.NewTicker(c.CleanupInterval)
	defer ticker.Stop()

	c.Log.Info("cleanup controller started",
		zap.String("stale_threshold", c.StaleClusterThreshold),
		zap.Duration("interval", c.CleanupInterval))

	// Run cleanup immediately on start
	c.runCleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			c.Log.Info("cleanup controller stopping")
			return
		case <-ticker.C:
			c.runCleanup(ctx)
		}
	}
}

// runCleanup executes all cleanup tasks.
func (c *CleanupController) runCleanup(ctx context.Context) {
	start := time.Now()
	c.Log.Debug("starting cleanup cycle")

	// 1. Clean up stale clusters (soft-delete)
	staleClusters := c.Store.CleanupStaleClusters(c.StaleClusterThreshold)
	if staleClusters > 0 {
		c.Log.Info("cleaned up stale clusters",
			zap.Int64("count", staleClusters),
			zap.String("threshold", c.StaleClusterThreshold))
	}

	// 2. Purge expired connection sessions
	c.Store.PurgeExpiredSessions(time.Now())

	// 3. Terminate workloads orphaned by soft-deleted clusters
	orphaned := c.Store.TerminateOrphanedWorkloads()
	if orphaned > 0 {
		c.Log.Info("terminated orphaned workloads",
			zap.Int64("count", orphaned))
	}

	// 4. Terminate workloads stuck in non-terminal state too long
	stale := c.Store.TerminateStaleWorkloads(c.StaleWorkloadThreshold)
	if stale > 0 {
		c.Log.Info("terminated stale workloads",
			zap.Int64("count", stale),
			zap.String("threshold", c.StaleWorkloadThreshold))
	}

	// Log completion
	c.Log.Debug("cleanup cycle completed",
		zap.Duration("duration", time.Since(start)),
		zap.Int64("stale_clusters", staleClusters),
		zap.Int64("orphaned_workloads", orphaned),
		zap.Int64("stale_workloads", stale))
}

// RunOnce executes cleanup tasks once and returns.
// Useful for testing or manual cleanup triggers.
func (c *CleanupController) RunOnce(ctx context.Context) {
	c.runCleanup(ctx)
}

