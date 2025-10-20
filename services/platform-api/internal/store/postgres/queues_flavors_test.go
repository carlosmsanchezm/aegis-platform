package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

const catalogDDL = `
CREATE TABLE IF NOT EXISTS flavors (
    name TEXT PRIMARY KEY,
    chip TEXT NULL,
    mig_profile TEXT NULL,
    rdma_required BOOLEAN NOT NULL DEFAULT FALSE,
    gpu_count INT NOT NULL DEFAULT 0,
    memory_gib DOUBLE PRECISION NULL,
    resource_name TEXT NULL,
    cpu_cores_request TEXT NULL,
    memory_request TEXT NULL,
    price_usd_per_gpu_hour DOUBLE PRECISION NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS queues (
    name TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    priority_tier TEXT NULL,
    allowed_flavors TEXT[] NULL,
    default_max_duration_seconds BIGINT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`

func TestCatalogUpsertsAreIdempotent(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	container, dsn := startPostgresContainer(ctx, t)
	t.Cleanup(func() {
		_ = container.Terminate(ctx)
	})

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("connect schema pool: %v", err)
	}
	t.Cleanup(pool.Close)

	var schemaErr error
	for attempt := 0; attempt < 5; attempt++ {
		if _, schemaErr = pool.Exec(ctx, catalogDDL); schemaErr == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if schemaErr != nil {
		t.Fatalf("apply catalog schema: %v", schemaErr)
	}

	store, err := New(dsn, zap.NewNop())
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	t.Cleanup(store.Close)

	queue := &aegis.Queue{
		Name:                      "default",
		ProjectId:                 "p-demo",
		DefaultMaxDurationSeconds: 600,
	}
	store.PutQueue(queue)

	queue.DefaultMaxDurationSeconds = 900
	queue.AllowedFlavors = []string{"cpu-small"}
	store.PutQueue(queue)

	fetchedQueue := store.GetQueue("default")
	if fetchedQueue == nil {
		t.Fatal("queue not found after upsert")
	}
	if fetchedQueue.GetProjectId() != "p-demo" {
		t.Fatalf("queue project mismatch, got %s", fetchedQueue.GetProjectId())
	}
	if fetchedQueue.GetDefaultMaxDurationSeconds() != 900 {
		t.Fatalf("expected queue max duration 900, got %d", fetchedQueue.GetDefaultMaxDurationSeconds())
	}
	if len(fetchedQueue.GetAllowedFlavors()) != 1 || fetchedQueue.GetAllowedFlavors()[0] != "cpu-small" {
		t.Fatalf("queue allowed flavors mismatch: %v", fetchedQueue.GetAllowedFlavors())
	}

	var queueCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM queues WHERE name='default'`).Scan(&queueCount); err != nil {
		t.Fatalf("count queues: %v", err)
	}
	if queueCount != 1 {
		t.Fatalf("expected single queue row, got %d", queueCount)
	}

	flavor := &aegis.Flavor{
		Name:            "cpu-small",
		CpuCoresRequest: "2",
		MemoryRequest:   "4Gi",
	}
	store.PutFlavor(flavor)

	flavor.MemoryRequest = "8Gi"
	flavor.ResourceName = "nvidia.com/gpu"
	flavor.GpuCount = 1
	store.PutFlavor(flavor)

	fetchedFlavor := store.GetFlavor("cpu-small")
	if fetchedFlavor == nil {
		t.Fatal("flavor not found after upsert")
	}
	if fetchedFlavor.GetMemoryRequest() != "8Gi" {
		t.Fatalf("flavor memory request mismatch, got %s", fetchedFlavor.GetMemoryRequest())
	}
	if fetchedFlavor.GetResourceName() != "nvidia.com/gpu" {
		t.Fatalf("flavor resource name mismatch, got %s", fetchedFlavor.GetResourceName())
	}
	if fetchedFlavor.GetGpuCount() != 1 {
		t.Fatalf("flavor gpu count mismatch, got %d", fetchedFlavor.GetGpuCount())
	}

	var flavorCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM flavors WHERE name='cpu-small'`).Scan(&flavorCount); err != nil {
		t.Fatalf("count flavors: %v", err)
	}
	if flavorCount != 1 {
		t.Fatalf("expected single flavor row, got %d", flavorCount)
	}
}
