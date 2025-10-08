package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	testcontainers "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

const schemaDDL = `
CREATE TABLE IF NOT EXISTS workloads (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    queue TEXT NOT NULL DEFAULT '',
    cluster_id TEXT NULL,
    status TEXT NOT NULL,
    ui_status TEXT NULL,
    url TEXT NULL,
    message TEXT NULL,
    kind TEXT NOT NULL,
    hints_resource_name TEXT NULL,
    hints_gpu_count INT NULL,
    hints_cpu_request TEXT NULL,
    hints_mem_request TEXT NULL,
    workspace_json BYTEA NULL,
    training_json BYTEA NULL,
    placed_at TIMESTAMPTZ NULL,
    started_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS workload_estimates (
    workload_id TEXT PRIMARY KEY REFERENCES workloads(id) ON DELETE CASCADE,
    estimate_usd DOUBLE PRECISION NOT NULL
);
`

func TestWorkloadLifecycle_Postgres(t *testing.T) {
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
		if _, schemaErr = pool.Exec(ctx, schemaDDL); schemaErr == nil {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}
	if schemaErr != nil {
		t.Fatalf("apply schema: %v", schemaErr)
	}

	store, err := New(dsn, zap.NewNop())
	if err != nil {
		t.Fatalf("create store: %v", err)
	}
	t.Cleanup(store.Close)

	projectID := "p-e2e"
	queue := "default"
	workloadID := fmt.Sprintf("w-%d", time.Now().UnixNano())

	original := &aegis.Workload{
		Id:        workloadID,
		ProjectId: projectID,
		Queue:     queue,
		Status:    "SUBMITTED",
		Kind: &aegis.Workload_Workspace{Workspace: &aegis.WorkspaceSpec{
			Flavor:             "a10-mig-1g",
			Image:              "alpine:3.19",
			Command:            []string{"sh", "-c", "echo hi"},
			MaxDurationSeconds: 300,
			Interactive:        false,
		}},
		Hints: &aegis.ResourceHints{
			ResourceName:    "nvidia.com/mig-1g.10gb",
			GpuCount:        1,
			CpuCoresRequest: "1",
			MemoryRequest:   "4Gi",
		},
	}

	store.PutWorkload(original)

	assertWorkloadEqual := func(label string, got *aegis.Workload, want *aegis.Workload) {
		t.Helper()
		if got == nil {
			t.Fatalf("%s workload nil", label)
		}
		if got.GetId() != want.GetId() {
			t.Fatalf("%s id mismatch: got %s want %s", label, got.GetId(), want.GetId())
		}
		if got.GetProjectId() != want.GetProjectId() {
			t.Fatalf("%s project mismatch: got %s want %s", label, got.GetProjectId(), want.GetProjectId())
		}
		if got.GetQueue() != want.GetQueue() {
			t.Fatalf("%s queue mismatch: got %s want %s", label, got.GetQueue(), want.GetQueue())
		}
		if got.GetStatus() != want.GetStatus() {
			t.Fatalf("%s status mismatch: got %s want %s", label, got.GetStatus(), want.GetStatus())
		}

		gHints, wHints := got.GetHints(), want.GetHints()
		if (gHints == nil) != (wHints == nil) {
			t.Fatalf("%s hints presence mismatch", label)
		}
		if gHints != nil {
			if gHints.GetResourceName() != wHints.GetResourceName() {
				t.Fatalf("%s hint resource mismatch: got %s want %s", label, gHints.GetResourceName(), wHints.GetResourceName())
			}
			if gHints.GetGpuCount() != wHints.GetGpuCount() {
				t.Fatalf("%s hint gpu mismatch: got %d want %d", label, gHints.GetGpuCount(), wHints.GetGpuCount())
			}
			if gHints.GetCpuCoresRequest() != wHints.GetCpuCoresRequest() {
				t.Fatalf("%s hint cpu mismatch: got %s want %s", label, gHints.GetCpuCoresRequest(), wHints.GetCpuCoresRequest())
			}
			if gHints.GetMemoryRequest() != wHints.GetMemoryRequest() {
				t.Fatalf("%s hint mem mismatch: got %s want %s", label, gHints.GetMemoryRequest(), wHints.GetMemoryRequest())
			}
		}

		gotWorkspace, okGot := got.GetKind().(*aegis.Workload_Workspace)
		wantWorkspace, okWant := want.GetKind().(*aegis.Workload_Workspace)
		if okGot != okWant {
			t.Fatalf("%s workspace presence mismatch", label)
		}
		if okGot {
			if gotWorkspace.Workspace.GetImage() != wantWorkspace.Workspace.GetImage() {
				t.Fatalf("%s workspace image mismatch: got %s want %s", label, gotWorkspace.Workspace.GetImage(), wantWorkspace.Workspace.GetImage())
			}
			if gotWorkspace.Workspace.GetFlavor() != wantWorkspace.Workspace.GetFlavor() {
				t.Fatalf("%s workspace flavor mismatch: got %s want %s", label, gotWorkspace.Workspace.GetFlavor(), wantWorkspace.Workspace.GetFlavor())
			}
			if fmt.Sprint(gotWorkspace.Workspace.GetCommand()) != fmt.Sprint(wantWorkspace.Workspace.GetCommand()) {
				t.Fatalf("%s workspace command mismatch: got %v want %v", label, gotWorkspace.Workspace.GetCommand(), wantWorkspace.Workspace.GetCommand())
			}
		}
	}

	got := store.GetWorkload(workloadID)
	assertWorkloadEqual("get", got, original)

	list := store.ListWorkloads(projectID)
	if len(list) != 1 {
		t.Fatalf("expected 1 workload in list, got %d", len(list))
	}
	assertWorkloadEqual("list", list[0], original)

	// Update workload fields and ensure they persist.
	original.Status = statusPlaced
	original.ClusterId = "cluster-1"
	original.UiStatus = "Ready"
	original.Url = "grpc://workload"
	original.Message = "waiting"
	store.PutWorkload(original)

	updated := store.GetWorkload(workloadID)
	assertWorkloadEqual("updated", updated, original)
	if updated.GetClusterId() != "cluster-1" {
		t.Fatalf("cluster id not persisted, got %s", updated.GetClusterId())
	}
	if updated.GetUiStatus() != "Ready" {
		t.Fatalf("ui status mismatch, got %s", updated.GetUiStatus())
	}
	if updated.GetMessage() != "waiting" {
		t.Fatalf("message mismatch, got %s", updated.GetMessage())
	}

	store.MarkPlaced(workloadID)
	placedAt, ok := store.GetPlacedAt(workloadID)
	if !ok {
		t.Fatalf("expected placed_at to be set")
	}
	if time.Since(placedAt) > time.Minute {
		t.Fatalf("placed_at timestamp looks stale: %v", placedAt)
	}

	store.ClearPlacedAt(workloadID)
	if _, ok := store.GetPlacedAt(workloadID); ok {
		t.Fatalf("expected placed_at to be cleared")
	}

	store.MarkPlaced(workloadID)
	time.Sleep(20 * time.Millisecond)

	runningW, waitDuration, transitioned, err := store.StartWorkload(workloadID)
	if err != nil {
		t.Fatalf("start workload: %v", err)
	}
	if !transitioned {
		t.Fatalf("expected workload to transition on first start call")
	}
	if waitDuration < 0 {
		t.Fatalf("expected non-negative wait duration, got %v", waitDuration)
	}
	if runningW.GetStatus() != statusRunning {
		t.Fatalf("expected status RUNNING after start, got %s", runningW.GetStatus())
	}

	again, waitAgain, transitionedAgain, err := store.StartWorkload(workloadID)
	if err != nil {
		t.Fatalf("second start workload call failed: %v", err)
	}
	if transitionedAgain {
		t.Fatalf("second start call should not transition")
	}
	if waitAgain != 0 {
		t.Fatalf("expected zero wait on idempotent start, got %v", waitAgain)
	}
	if again.GetStatus() != statusRunning {
		t.Fatalf("idempotent start should keep RUNNING status, got %s", again.GetStatus())
	}

	store.MarkStarted(workloadID)
	if startedAt, ok := store.GetStartedAt(workloadID); !ok || time.Since(startedAt) > time.Minute {
		t.Fatalf("expected started_at to be present and recent, got %v (present=%t)", startedAt, ok)
	}

	store.SetEstimateUSD(workloadID, 42.25)
	if est := store.PopEstimateUSD(workloadID); est < 42.20 || est > 42.30 {
		t.Fatalf("unexpected estimate: %f", est)
	}
	if est := store.PopEstimateUSD(workloadID); est != 0 {
		t.Fatalf("estimate should be cleared after pop, got %f", est)
	}

	// Prepare a second workload for leasing.
	leaseID := fmt.Sprintf("w-%d", time.Now().UnixNano())
	lease := &aegis.Workload{
		Id:        leaseID,
		ProjectId: projectID,
		Queue:     queue,
		ClusterId: "c-1",
		Status:    statusPlaced,
		Kind: &aegis.Workload_Workspace{Workspace: &aegis.WorkspaceSpec{
			Flavor:  "a10-mig-1g",
			Image:   "alpine:3.19",
			Command: []string{"sh", "-c", "echo lease"},
		}},
	}
	store.PutWorkload(lease)
	store.MarkPlaced(leaseID)

	leased := store.LeaseWorkloads("c-1", 2)
	if len(leased) != 1 {
		t.Fatalf("expected exactly one leased workload, got %d", len(leased))
	}
	if leased[0].GetId() != leaseID {
		t.Fatalf("leased workload mismatch: got %s want %s", leased[0].GetId(), leaseID)
	}
	if leased[0].GetStatus() != statusRunning {
		t.Fatalf("leased workload should transition to RUNNING, got %s", leased[0].GetStatus())
	}
	leaseDB := store.GetWorkload(leaseID)
	if leaseDB.GetStatus() != statusRunning {
		t.Fatalf("expected leased workload persisted as RUNNING, got %s", leaseDB.GetStatus())
	}
}

func startPostgresContainer(ctx context.Context, t *testing.T) (testcontainers.Container, string) {
	t.Helper()

	const (
		username = "postgres"
		password = "secret"
		dbName   = "aegis_test"
	)

	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": password,
			"POSTGRES_USER":     username,
			"POSTGRES_DB":       dbName,
		},
		WaitingFor: wait.ForSQL("5432/tcp", "postgres", func(host string, port uint16) string {
			return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable", username, password, host, port, dbName)
		}).WithStartupTimeout(2 * time.Minute),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("container host: %v", err)
	}

	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		container.Terminate(ctx)
		t.Fatalf("container mapped port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", username, password, host, port.Port(), dbName)
	return container, dsn
}
