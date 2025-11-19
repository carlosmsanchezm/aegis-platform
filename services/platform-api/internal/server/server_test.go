package server

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	workspacecfg "github.com/yourorg/aegis/pkg/workspace"
	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/placement"
	mw "github.com/yourorg/aegis/services/platform-api/internal/server/mw"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	t.Setenv("AEGIS_PROXY_BASE_URL", "https://proxy.test")
	t.Setenv("AEGIS_PROXY_EXPECTED_AUDIENCE", "aegis-proxy")
	t.Setenv("AEGIS_PROXY_JWT_SECRET", "super-secret-key")
	t.Setenv("AEGIS_PROXY_TOKEN_TTL_SECONDS", "600")

	logger := zap.NewNop()
	st := store.NewMemStore()

	srv := New(logger, st, nil, "default", placement.NewPolicyOverlay(), nil, "")
	return srv
}

func contextWithSubject(subject string) context.Context {
	identity := &mw.Identity{
		Subject:  subject,
		Roles:    []string{"default-roles-aegis", "workspace-admin"},
		ClientID: "backstage",
	}
	return mw.ContextWithIdentity(context.Background(), identity)
}

func stubWorkspace(id string, interactive bool, env map[string]string) *aegis.Workload {
	return &aegis.Workload{
		Id:        id,
		ProjectId: "proj-1",
		Queue:     "queue-a",
		ClusterId: "cluster-1",
		Status:    statusRunning,
		Kind: &aegis.Workload_Workspace{
			Workspace: &aegis.WorkspaceSpec{
				Interactive: interactive,
				Ports:       []int32{workspacecfg.DefaultVSCodePort},
				Env:         env,
			},
		},
	}
}

func TestApplyWorkspaceDefaults(t *testing.T) {
	srv := newTestServer(t)
	ws := &aegis.WorkspaceSpec{
		Ports: []int32{-1},
		Env: map[string]string{
			workspacecfg.EnvVSCodeCommit: "custom",
			"CUSTOM":                     "1",
		},
	}

	srv.applyWorkspaceDefaults(ws)

	if len(ws.Ports) != 1 {
		t.Fatalf("expected one port, got %v", ws.Ports)
	}
	if ws.Ports[0] != workspacecfg.DefaultVSCodePort {
		t.Fatalf("expected port %d, got %v", workspacecfg.DefaultVSCodePort, ws.Ports)
	}
	if ws.Env[workspacecfg.EnvVSCodeCommit] != "custom" {
		t.Fatalf("expected user commit to be preserved, got %q", ws.Env[workspacecfg.EnvVSCodeCommit])
	}
	if ws.Env[workspacecfg.EnvVSCodeQuality] == "" {
		t.Fatalf("expected default quality to be populated, env=%v", ws.Env)
	}
	if ws.Env["CUSTOM"] != "1" {
		t.Fatalf("expected custom env to survive, env=%v", ws.Env)
	}
	if ws.Env[workspacecfg.EnvPasswordAccess] != workspacecfg.DefaultPasswordAccess {
		t.Fatalf("expected password access default, env=%v", ws.Env)
	}
}

func TestCreateConnectionSession_Success(t *testing.T) {
	srv := newTestServer(t)
	srv.store.PutWorkload(stubWorkspace("w-123", true, nil))

	ctx := contextWithSubject("alice@example.com")

	req := &aegis.CreateConnectionSessionRequest{WorkloadId: "w-123", Client: "vscode"}
	resp, err := srv.CreateConnectionSession(ctx, req)
	if err != nil {
		t.Fatalf("CreateConnectionSession returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected session response")
	}
	if resp.GetSessionId() == "" {
		t.Fatal("session id should not be empty")
	}
	if resp.GetToken() == "" {
		t.Fatal("token should not be empty")
	}
	if !resp.GetOneTime() {
		t.Fatal("one_time flag should be true")
	}
	if resp.GetSshConfig() == "" || resp.GetProxyUrl() == "" {
		t.Fatal("ssh config and proxy url are required")
	}
	if !contains(resp.GetSshConfig(), "aegis-connect") {
		t.Errorf("ssh config should reference aegis-connect; got %q", resp.GetSshConfig())
	}

	stored, ok := srv.store.ConnectionSession(resp.GetSessionId())
	if !ok {
		t.Fatalf("session %s not persisted", resp.GetSessionId())
	}
	if stored.Client != "vscode" {
		t.Errorf("expected client \"vscode\", got %q", stored.Client)
	}
	if stored.WorkloadID != "w-123" {
		t.Errorf("expected workload id \"w-123\", got %q", stored.WorkloadID)
	}
	if stored.Subject == "" {
		t.Error("stored subject should not be empty")
	}
	if stored.ExpiresAt.Sub(time.Now()) > maxSessionTTL+time.Minute {
		t.Errorf("expires_at exceeded maximum ttl: %v", stored.ExpiresAt.Sub(time.Now()))
	}
}

func TestCreateConnectionSession_RespectsWorkspaceUser(t *testing.T) {
	srv := newTestServer(t)
	env := map[string]string{"USER_NAME": "AegisUser"}
	srv.store.PutWorkload(stubWorkspace("w-456", true, env))

	ctx := contextWithSubject("someone@example.com")
	resp, err := srv.CreateConnectionSession(ctx, &aegis.CreateConnectionSessionRequest{WorkloadId: "w-456"})
	if err != nil {
		t.Fatalf("CreateConnectionSession returned error: %v", err)
	}
	config := resp.GetSshConfig()
	if !strings.Contains(config, "User aegisuser") {
		t.Fatalf("expected ssh config to use workspace user; got %q", config)
	}
	if strings.Contains(config, "\"aegis-connect") {
		t.Fatalf("proxy command should not be wrapped in quotes; got %q", config)
	}
	if !strings.Contains(config, "ProxyCommand aegis-connect --proxy=") {
		t.Fatalf("proxy command not updated; got %q", config)
	}
}

func TestCreateConnectionSession_SanitizesWorkspaceUser(t *testing.T) {
	srv := newTestServer(t)
	env := map[string]string{"AEGIS_SSH_USER": "ROOT"}
	srv.store.PutWorkload(stubWorkspace("w-789", true, env))

	ctx := contextWithSubject("another@example.com")
	resp, err := srv.CreateConnectionSession(ctx, &aegis.CreateConnectionSessionRequest{WorkloadId: "w-789"})
	if err != nil {
		t.Fatalf("CreateConnectionSession returned error: %v", err)
	}
	config := resp.GetSshConfig()
	if strings.Contains(config, "User root") {
		t.Fatalf("expected root to be rejected in ssh config; got %q", config)
	}
	if !strings.Contains(config, "User aegis-") {
		t.Fatalf("expected fallback hashed user in ssh config; got %q", config)
	}
}

func TestCreateConnectionSession_InvalidClient(t *testing.T) {
	srv := newTestServer(t)
	srv.store.PutWorkload(stubWorkspace("w-123", true, nil))
	ctx := contextWithSubject("bob@example.com")

	_, err := srv.CreateConnectionSession(ctx, &aegis.CreateConnectionSessionRequest{WorkloadId: "w-123", Client: "win-ssh"})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected invalid argument error, got %v", err)
	}
}

func TestRenewConnectionSession(t *testing.T) {
	srv := newTestServer(t)
	srv.store.PutWorkload(stubWorkspace("w-123", true, nil))
	ctx := contextWithSubject("carol@example.com")

	created, err := srv.CreateConnectionSession(ctx, &aegis.CreateConnectionSessionRequest{WorkloadId: "w-123", Client: "cli"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	oldToken := created.GetToken()
	renewed, err := srv.RenewConnectionSession(ctx, &aegis.RenewConnectionSessionRequest{SessionId: created.GetSessionId()})
	if err != nil {
		t.Fatalf("renew failed: %v", err)
	}
	if renewed.GetToken() == oldToken {
		t.Error("renew should issue a new token")
	}

	stored, ok := srv.store.ConnectionSession(created.GetSessionId())
	if !ok {
		t.Fatal("session missing after renew")
	}
	if stored.Used {
		t.Error("session should remain unused after renewal")
	}
}

func TestRenewConnectionSession_WhenUsedFails(t *testing.T) {
	srv := newTestServer(t)
	srv.store.PutWorkload(stubWorkspace("w-123", true, nil))
	ctx := contextWithSubject("dan@example.com")

	created, err := srv.CreateConnectionSession(ctx, &aegis.CreateConnectionSessionRequest{WorkloadId: "w-123"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	srv.store.MarkSessionUsed(created.GetSessionId())

	_, err = srv.RenewConnectionSession(ctx, &aegis.RenewConnectionSessionRequest{SessionId: created.GetSessionId()})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition when renewing used session, got %v", err)
	}
}

func TestRevokeConnectionSession(t *testing.T) {
	srv := newTestServer(t)
	srv.store.PutWorkload(stubWorkspace("w-123", true, nil))
	ctx := contextWithSubject("erin@example.com")

	created, err := srv.CreateConnectionSession(ctx, &aegis.CreateConnectionSessionRequest{WorkloadId: "w-123"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	if _, err := srv.RevokeConnectionSession(ctx, &aegis.RevokeConnectionSessionRequest{SessionId: created.GetSessionId()}); err != nil {
		t.Fatalf("revoke failed: %v", err)
	}

	stored, ok := srv.store.ConnectionSession(created.GetSessionId())
	if !ok {
		t.Fatal("session not found after revoke")
	}
	if !stored.Revoked {
		t.Error("session should be marked revoked")
	}
	if stored.Token != "" {
		t.Error("token should be cleared upon revocation")
	}
}

func TestGetWorkspaceConnectionDetails(t *testing.T) {
	srv := newTestServer(t)
	srv.store.PutWorkload(stubWorkspace("w-123", true, map[string]string{"USER_NAME": "Aegis"}))
	ctx := contextWithSubject("frank@example.com")

	resp, err := srv.GetWorkspaceConnectionDetails(ctx, &aegis.GetWorkspaceConnectionDetailsRequest{Id: "w-123"})
	if err != nil {
		t.Fatalf("get connection details failed: %v", err)
	}
	if resp.GetProxyUrl() == "" || resp.GetToken() == "" {
		t.Fatal("expected proxy url and token in response")
	}
	if resp.GetSshHostAlias() == "" {
		t.Fatal("ssh host alias missing")
	}

	token := resp.GetToken()
	parsed, err := jwt.Parse(token, func(token *jwt.Token) (any, error) { return []byte("super-secret-key"), nil })
	if err != nil || !parsed.Valid {
		t.Fatalf("returned token not valid JWT: %v", err)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestMaybeBootstrapWorkspaceDeps_CreatesCatalogWhenEnabled(t *testing.T) {
	srv := newTestServer(t)
	srv.autoBootstrap = true

	workload := &aegis.Workload{
		Id:        "w-boot",
		ProjectId: "p-demo",
		Queue:     "default",
		Kind: &aegis.Workload_Workspace{
			Workspace: &aegis.WorkspaceSpec{
				Flavor: "cpu-small",
				Image:  "alpine:3.19",
			},
		},
	}

	srv.maybeBootstrapWorkspaceDeps(context.Background(), workload)

	if srv.store.GetProject("p-demo") == nil {
		t.Fatalf("expected project p-demo to be bootstrapped")
	}
	queue := srv.store.GetQueue("default")
	if queue == nil {
		t.Fatalf("expected queue default to be bootstrapped")
	}
	if queue.GetProjectId() != "p-demo" {
		t.Fatalf("queue project mismatch: got %s", queue.GetProjectId())
	}
	if queue.GetDefaultMaxDurationSeconds() != srv.defaultMaxRuntimeSeconds(nil) {
		t.Fatalf("queue default max duration not populated, got %d", queue.GetDefaultMaxDurationSeconds())
	}
	flavor := srv.store.GetFlavor("cpu-small")
	if flavor == nil {
		t.Fatalf("expected flavor cpu-small to be bootstrapped")
	}
	if flavor.GetCpuCoresRequest() != "2" || flavor.GetMemoryRequest() != "4Gi" {
		t.Fatalf("unexpected flavor defaults: cpu=%s mem=%s", flavor.GetCpuCoresRequest(), flavor.GetMemoryRequest())
	}
	// Ensure idempotency
	srv.maybeBootstrapWorkspaceDeps(context.Background(), workload)
}

func TestSubmitWorkload_BootstrapUsesPolicyDefaults(t *testing.T) {
	srv := newTestServer(t)
	srv.autoBootstrap = true

	srv.policyOverlay.Set(&infraapi.ProjectPlacementSpec{
		ProjectID:     "proj-policy",
		DefaultFlavor: "cpu-small",
	})

	req := &aegis.SubmitWorkloadRequest{
		Workload: &aegis.Workload{
			ProjectId: "proj-policy",
			Queue:     "queue-policy",
			Kind: &aegis.Workload_Workspace{
				Workspace: &aegis.WorkspaceSpec{
					Image: "alpine:3.19",
				},
			},
		},
	}

	ctx := contextWithSubject("policy@example.com")

	_, err := srv.SubmitWorkload(ctx, req)
	if code := status.Code(err); code != codes.Internal && code != codes.FailedPrecondition {
		t.Fatalf("unexpected error code %v: %v", code, err)
	}

	if srv.store.GetProject("proj-policy") == nil {
		t.Fatalf("expected project proj-policy to be bootstrapped")
	}
	if srv.store.GetQueue("queue-policy") == nil {
		t.Fatalf("expected queue queue-policy to be bootstrapped")
	}
	if srv.store.GetFlavor("cpu-small") == nil {
		t.Fatalf("expected flavor cpu-small to be bootstrapped")
	}
	if ws, ok := req.Workload.GetKind().(*aegis.Workload_Workspace); !ok || ws.Workspace.GetFlavor() != "cpu-small" {
		t.Fatalf("expected workspace flavor to be defaulted, got %+v", req.Workload.GetKind())
	}
}

func TestSubmitWorkload_BootstrapDisabledRequiresProject(t *testing.T) {
	srv := newTestServer(t)
	srv.autoBootstrap = false

	req := &aegis.SubmitWorkloadRequest{
		Workload: &aegis.Workload{
			ProjectId: "p-missing",
			Kind: &aegis.Workload_Workspace{
				Workspace: &aegis.WorkspaceSpec{
					Flavor: "cpu-small",
					Image:  "alpine:3.19",
				},
			},
		},
	}

	ctx := contextWithSubject("missing@example.com")

	_, err := srv.SubmitWorkload(ctx, req)
	if err == nil {
		t.Fatal("expected error when project is missing")
	}
	if status.Code(err) != codes.InvalidArgument || !strings.Contains(err.Error(), "unknown project") {
		t.Fatalf("expected unknown project error, got %v", err)
	}
}
