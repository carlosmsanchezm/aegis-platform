package controller

import (
	"context"
	"os"
	"strconv"
	"strings"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	aegisv1alpha2 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha2"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/cpclient"
)

const (
	workloadGCControllerName = "workload-gc"

	defaultWorkloadGCInterval     = 30 * time.Second
	defaultWorkloadGCMaxDeletions = 25
)

type workloadLister interface {
	ListClusterWorkloadIDs(ctx context.Context, clusterID string) ([]string, error)
}

// WorkloadGCController periodically reconciles the local Workspace CRDs with hub workload IDs and deletes
// any orphaned Workspaces (present on the spoke but missing from the hub database).
type WorkloadGCController struct {
	client.Client

	cpClient        workloadLister
	clusterID       string
	interval        time.Duration
	maxDeletions    int
	dryRun          bool
	createdCPClient bool
}

func (c *WorkloadGCController) SetupWithManager(mgr ctrl.Manager) error {
	logger := ctrl.Log.WithName(workloadGCControllerName)

	if c.Client == nil {
		c.Client = mgr.GetClient()
	}

	enabled := parseEnvBool("AEGIS_WORKLOAD_GC_ENABLED", true)
	if !enabled {
		logger.Info("workload GC disabled (AEGIS_WORKLOAD_GC_ENABLED=false)")
		return nil
	}

	if c.clusterID == "" {
		c.clusterID = strings.TrimSpace(os.Getenv("AEGIS_CLUSTER_ID"))
	}
	if c.clusterID == "" {
		logger.Info("workload GC disabled (AEGIS_CLUSTER_ID not set)")
		return nil
	}

	if c.interval <= 0 {
		c.interval = parseEnvDuration("AEGIS_WORKLOAD_GC_INTERVAL", defaultWorkloadGCInterval)
	}
	if c.maxDeletions == 0 {
		c.maxDeletions = parseEnvInt("AEGIS_WORKLOAD_GC_MAX_DELETIONS", defaultWorkloadGCMaxDeletions)
	}
	c.dryRun = parseEnvBool("AEGIS_WORKLOAD_GC_DRY_RUN", false)

	if c.cpClient == nil {
		if endpoint := strings.TrimSpace(os.Getenv("AEGIS_CP_GRPC")); endpoint != "" {
			client, err := cpclient.New(endpoint)
			if err != nil {
				logger.Error(err, "failed to create control-plane client", "endpoint", endpoint)
			} else {
				c.cpClient = client
				c.createdCPClient = true
			}
		}
	}
	if c.cpClient == nil {
		logger.Info("workload GC disabled (AEGIS_CP_GRPC not configured)")
		return nil
	}

	logger.Info("workload GC enabled",
		"cluster_id", c.clusterID,
		"interval", c.interval.String(),
		"max_deletions", c.maxDeletions,
		"dry_run", c.dryRun,
	)

	return mgr.Add(manager.RunnableFunc(func(ctx context.Context) error {
		if c.createdCPClient {
			defer func() {
				if closer, ok := c.cpClient.(interface{ Close() error }); ok {
					_ = closer.Close()
				}
			}()
		}
		c.run(ctx)
		return nil
	}))
}

func (c *WorkloadGCController) run(ctx context.Context) {
	logger := ctrl.Log.WithName(workloadGCControllerName)
	ctx = ctrl.LoggerInto(ctx, logger)

	// Run immediately on startup, then at the configured interval.
	c.runOnce(ctx)

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.runOnce(ctx)
		}
	}
}

func (c *WorkloadGCController) runOnce(ctx context.Context) {
	logger := ctrl.LoggerFrom(ctx)

	hubIDs, err := c.cpClient.ListClusterWorkloadIDs(ctx, c.clusterID)
	if err != nil {
		logger.Error(err, "workload GC: failed to list hub workload IDs", "cluster_id", c.clusterID)
		return
	}

	hubSet := make(map[string]struct{}, len(hubIDs))
	for _, id := range hubIDs {
		if normalized := normalizeWorkloadID(id); normalized != "" {
			hubSet[normalized] = struct{}{}
		}
	}

	var workspaces aegisv1alpha2.WorkspaceList
	if err := c.List(ctx, &workspaces); err != nil {
		logger.Error(err, "workload GC: failed to list local Workspaces")
		return
	}

	orphans := make([]*aegisv1alpha2.Workspace, 0, len(workspaces.Items))
	for i := range workspaces.Items {
		ws := &workspaces.Items[i]
		if ws.GetDeletionTimestamp() != nil {
			continue
		}
		normalizedID := normalizeWorkloadID(ws.GetLabels()[labelWorkloadID])
		if normalizedID == "" {
			continue
		}
		if _, ok := hubSet[normalizedID]; ok {
			continue
		}
		orphans = append(orphans, ws)
	}

	if len(orphans) == 0 {
		logger.V(1).Info("workload GC: no orphaned Workspaces detected", "cluster_id", c.clusterID, "hub_workloads", len(hubSet))
		return
	}

	deleted := 0
	for _, ws := range orphans {
		if c.maxDeletions > 0 && deleted >= c.maxDeletions {
			logger.Info("workload GC: reached max deletions per run; remaining orphans will be retried",
				"cluster_id", c.clusterID,
				"deleted", deleted,
				"orphans_total", len(orphans),
				"max_deletions", c.maxDeletions,
			)
			break
		}

		workloadID := normalizeWorkloadID(ws.GetLabels()[labelWorkloadID])
		if c.dryRun {
			logger.Info("workload GC: would delete orphaned Workspace (dry-run)",
				"cluster_id", c.clusterID,
				"namespace", ws.GetNamespace(),
				"workspace", ws.GetName(),
				"workload_id", workloadID,
			)
			deleted++
			continue
		}

		if err := c.Delete(ctx, ws, client.PropagationPolicy(metav1.DeletePropagationBackground)); err != nil {
			if apierrors.IsNotFound(err) {
				continue
			}
			logger.Error(err, "workload GC: failed to delete orphaned Workspace",
				"cluster_id", c.clusterID,
				"namespace", ws.GetNamespace(),
				"workspace", ws.GetName(),
				"workload_id", workloadID,
			)
			continue
		}

		logger.Info("workload GC: deleted orphaned Workspace",
			"cluster_id", c.clusterID,
			"namespace", ws.GetNamespace(),
			"workspace", ws.GetName(),
			"workload_id", workloadID,
		)
		deleted++
	}
}

func normalizeWorkloadID(id string) string {
	id = strings.TrimSpace(id)
	for strings.HasPrefix(id, "aegis-") {
		id = strings.TrimPrefix(id, "aegis-")
	}
	return id
}

func parseEnvBool(key string, def bool) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	val, err := strconv.ParseBool(raw)
	if err != nil {
		return def
	}
	return val
}

func parseEnvDuration(key string, def time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	val, err := time.ParseDuration(raw)
	if err != nil || val <= 0 {
		return def
	}
	return val
}

func parseEnvInt(key string, def int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	val, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return val
}
