package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	agentv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
	workspacecfg "github.com/yourorg/aegis/pkg/workspace"
	aegis "github.com/yourorg/aegis/proto/aegis/v1"
	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
	"github.com/yourorg/aegis/services/platform-api/internal/placement"
	mw "github.com/yourorg/aegis/services/platform-api/internal/server/mw"
	"github.com/yourorg/aegis/services/platform-api/internal/store"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

type staticKubeClient struct {
	cli client.Client
}

type failingStore struct {
	*store.MemStore
	failResume    bool
	failTerminate bool
}

type fixedRuntimeStore struct {
	*store.MemStore
	startedAt      time.Time
	startedAtOK    bool
	runtimeSeconds int64
}

func (s *fixedRuntimeStore) GetStartedAt(id string) (time.Time, bool) {
	if !s.startedAtOK {
		return time.Time{}, false
	}
	return s.startedAt, true
}

func (s *fixedRuntimeStore) GetRuntimeSeconds(id string) (int64, bool) {
	return s.runtimeSeconds, true
}

func (s *failingStore) ResumeWorkload(id string) (*aegis.Workload, error) {
	if s.failResume {
		return nil, errors.New("simulated store failure")
	}
	return s.MemStore.ResumeWorkload(id)
}

func (s *failingStore) TerminateWorkload(id, reason string) (*aegis.Workload, error) {
	if s.failTerminate {
		return nil, errors.New("simulated store failure")
	}
	return s.MemStore.TerminateWorkload(id, reason)
}

type deleteFailingClient struct {
	client.Client
	failWorkspaceDelete bool
}

func (c deleteFailingClient) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) error {
	if c.failWorkspaceDelete {
		if u, ok := obj.(*unstructured.Unstructured); ok && u.GetKind() == "Workspace" {
			return errors.New("simulated kube delete failure")
		}
	}
	return c.Client.Delete(ctx, obj, opts...)
}

func (s staticKubeClient) ClientFor(clusterID string) (client.Client, error) {
	return s.cli, nil
}

func (s staticKubeClient) RestConfigFor(clusterID string) (*rest.Config, error) {
	return &rest.Config{}, nil
}

func (s staticKubeClient) HasKubeconfig(clusterID string) bool {
	return true // Test mock always has kubeconfig
}

func (s staticKubeClient) EvictClient(clusterID string) {}

func (s staticKubeClient) Dir() string {
	return "/tmp/test-kubeconfigs"
}

func newFakeWorkspaceClient(t *testing.T) client.Client {
	t.Helper()
	scheme := runtime.NewScheme()
	_ = clientgoscheme.AddToScheme(scheme)
	_ = agentv1alpha1.AddToScheme(scheme)
	workspaceGV := schema.GroupVersion{Group: "aegis.yourorg.dev", Version: "v1alpha2"}
	scheme.AddKnownTypeWithName(workspaceGV.WithKind("Workspace"), &unstructured.Unstructured{})
	scheme.AddKnownTypeWithName(workspaceGV.WithKind("WorkspaceList"), &unstructured.UnstructuredList{})
	return fake.NewClientBuilder().WithScheme(scheme).Build()
}

func newTestServer(t *testing.T) *Server {
	t.Helper()
	return newTestServerWithStore(t, store.NewMemStore())
}

func newTestServerWithStore(t *testing.T, st store.Store) *Server {
	t.Helper()
	t.Setenv("AEGIS_PROXY_BASE_URL", "https://proxy.test")
	t.Setenv("AEGIS_PROXY_EXPECTED_AUDIENCE", "aegis-proxy")
	t.Setenv("AEGIS_PROXY_JWT_SECRET", "super-secret-key")
	t.Setenv("AEGIS_PROXY_TOKEN_TTL_SECONDS", "600")

	logger := zap.NewNop()

	srv := New(logger, st, nil, "default", placement.NewPolicyOverlay(), nil, "")
	srv.kubeClients = staticKubeClient{cli: newFakeWorkspaceClient(t)}
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

func TestResumeWorkload_RollbackJobWhenStoreUpdateFails(t *testing.T) {
	fs := &failingStore{MemStore: store.NewMemStore(), failResume: true}
	srv := newTestServerWithStore(t, fs)

	w := stubWorkspace("w-123", true, nil)
	w.Status = statusSuspended
	srv.store.PutWorkload(w)

	cli, err := srv.kubeClients.ClientFor(w.GetClusterId())
	if err != nil {
		t.Fatalf("ClientFor returned error: %v", err)
	}
	targetNS := srv.namespaceForProject(w.GetProjectId())

	suspend := true
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job-w-123",
			Namespace: targetNS,
			Labels:    map[string]string{labelWorkloadID: w.GetId()},
			Annotations: map[string]string{
				annotationSuspendReason: "idle_timeout",
			},
		},
		Spec: batchv1.JobSpec{Suspend: &suspend},
	}
	if err := cli.Create(context.Background(), job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	_, err = srv.ResumeWorkload(contextWithSubject("alice@example.com"), &aegis.ResumeWorkloadRequest{Id: w.GetId()})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", err)
	}

	gotJob := &batchv1.Job{}
	if err := cli.Get(context.Background(), types.NamespacedName{Name: job.GetName(), Namespace: targetNS}, gotJob); err != nil {
		t.Fatalf("expected job to remain, got error: %v", err)
	}
	if gotJob.Spec.Suspend == nil || *gotJob.Spec.Suspend != true {
		t.Fatalf("expected job to remain suspended, suspend=%v", gotJob.Spec.Suspend)
	}
	if gotJob.Annotations == nil || gotJob.Annotations[annotationSuspendReason] == "" {
		t.Fatalf("expected suspend reason annotation to remain, annotations=%v", gotJob.Annotations)
	}
	if _, ok := gotJob.Annotations[annotationResumedAt]; ok {
		t.Fatalf("expected resumed-at annotation to be rolled back, annotations=%v", gotJob.Annotations)
	}

	if still := srv.store.GetWorkload(w.GetId()); still == nil || still.GetStatus() != statusSuspended {
		t.Fatalf("expected workload status to remain SUSPENDED, got %v", still)
	}
}

func TestTerminateWorkload_PreconditionDoesNotDeleteResources(t *testing.T) {
	srv := newTestServer(t)
	w := stubWorkspace("w-123", true, nil)
	w.Status = "SUCCEEDED"
	srv.store.PutWorkload(w)

	cli, err := srv.kubeClients.ClientFor(w.GetClusterId())
	if err != nil {
		t.Fatalf("ClientFor returned error: %v", err)
	}

	targetNS := srv.namespaceForProject(w.GetProjectId())
	workspace := &unstructured.Unstructured{}
	workspace.SetAPIVersion("aegis.yourorg.dev/v1alpha2")
	workspace.SetKind("Workspace")
	workspace.SetName(w.GetId())
	workspace.SetNamespace(targetNS)
	if err := cli.Create(context.Background(), workspace); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	aw := &agentv1alpha1.AegisWorkload{ObjectMeta: metav1.ObjectMeta{Name: w.GetId(), Namespace: targetNS}}
	if err := cli.Create(context.Background(), aw); err != nil {
		t.Fatalf("failed to create aegisworkload: %v", err)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job-w-123",
			Namespace: targetNS,
			Labels:    map[string]string{labelWorkloadID: w.GetId()},
		},
	}
	if err := cli.Create(context.Background(), job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	_, err = srv.TerminateWorkload(contextWithSubject("alice@example.com"), &aegis.TerminateWorkloadRequest{
		Id:     w.GetId(),
		Reason: "cleanup",
	})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", err)
	}

	gotWS := &unstructured.Unstructured{}
	gotWS.SetAPIVersion("aegis.yourorg.dev/v1alpha2")
	gotWS.SetKind("Workspace")
	if err := cli.Get(context.Background(), types.NamespacedName{Name: w.GetId(), Namespace: targetNS}, gotWS); err != nil {
		t.Fatalf("expected workspace to remain, got error: %v", err)
	}

	gotAW := &agentv1alpha1.AegisWorkload{}
	if err := cli.Get(context.Background(), types.NamespacedName{Name: w.GetId(), Namespace: targetNS}, gotAW); err != nil {
		t.Fatalf("expected aegisworkload to remain, got error: %v", err)
	}

	gotJob := &batchv1.Job{}
	if err := cli.Get(context.Background(), types.NamespacedName{Name: job.GetName(), Namespace: targetNS}, gotJob); err != nil {
		t.Fatalf("expected job to remain, got error: %v", err)
	}

	if still := srv.store.GetWorkload(w.GetId()); still == nil || still.GetStatus() != "SUCCEEDED" {
		t.Fatalf("expected workload status to remain SUCCEEDED, got %v", still)
	}
}

func TestTerminateWorkload_DeleteFailureRollsBackStore(t *testing.T) {
	st := store.NewMemStore()
	srv := newTestServerWithStore(t, st)

	cli := deleteFailingClient{Client: newFakeWorkspaceClient(t), failWorkspaceDelete: true}
	srv.kubeClients = staticKubeClient{cli: cli}

	w := stubWorkspace("w-123", true, nil)
	w.Status = statusRunning
	srv.store.PutWorkload(w)

	targetNS := srv.namespaceForProject(w.GetProjectId())
	workspace := &unstructured.Unstructured{}
	workspace.SetAPIVersion("aegis.yourorg.dev/v1alpha2")
	workspace.SetKind("Workspace")
	workspace.SetName(w.GetId())
	workspace.SetNamespace(targetNS)
	if err := cli.Create(context.Background(), workspace); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	_, err := srv.TerminateWorkload(contextWithSubject("alice@example.com"), &aegis.TerminateWorkloadRequest{
		Id:     w.GetId(),
		Reason: "cleanup",
	})
	if status.Code(err) != codes.Internal {
		t.Fatalf("expected Internal, got %v", err)
	}

	if still := srv.store.GetWorkload(w.GetId()); still == nil || still.GetStatus() != statusRunning || still.GetTerminatedAtUtc() != "" {
		t.Fatalf("expected workload status to rollback to RUNNING with no terminated_at, got %v", still)
	}

	gotWS := &unstructured.Unstructured{}
	gotWS.SetAPIVersion("aegis.yourorg.dev/v1alpha2")
	gotWS.SetKind("Workspace")
	if err := cli.Get(context.Background(), types.NamespacedName{Name: w.GetId(), Namespace: targetNS}, gotWS); err != nil {
		t.Fatalf("expected workspace to remain, got error: %v", err)
	}
}

func TestTerminateWorkload_StoreFailureDoesNotDeleteResources(t *testing.T) {
	fs := &failingStore{MemStore: store.NewMemStore(), failTerminate: true}
	srv := newTestServerWithStore(t, fs)
	w := stubWorkspace("w-123", true, nil)
	w.Status = statusRunning
	srv.store.PutWorkload(w)

	cli, err := srv.kubeClients.ClientFor(w.GetClusterId())
	if err != nil {
		t.Fatalf("ClientFor returned error: %v", err)
	}
	targetNS := srv.namespaceForProject(w.GetProjectId())

	workspace := &unstructured.Unstructured{}
	workspace.SetAPIVersion("aegis.yourorg.dev/v1alpha2")
	workspace.SetKind("Workspace")
	workspace.SetName(w.GetId())
	workspace.SetNamespace(targetNS)
	if err := cli.Create(context.Background(), workspace); err != nil {
		t.Fatalf("failed to create workspace: %v", err)
	}

	aw := &agentv1alpha1.AegisWorkload{ObjectMeta: metav1.ObjectMeta{Name: w.GetId(), Namespace: targetNS}}
	if err := cli.Create(context.Background(), aw); err != nil {
		t.Fatalf("failed to create aegisworkload: %v", err)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "job-w-123",
			Namespace: targetNS,
			Labels:    map[string]string{labelWorkloadID: w.GetId()},
		},
	}
	if err := cli.Create(context.Background(), job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	_, err = srv.TerminateWorkload(contextWithSubject("alice@example.com"), &aegis.TerminateWorkloadRequest{Id: w.GetId()})
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("expected FailedPrecondition, got %v", err)
	}

	gotWS := &unstructured.Unstructured{}
	gotWS.SetAPIVersion("aegis.yourorg.dev/v1alpha2")
	gotWS.SetKind("Workspace")
	if err := cli.Get(context.Background(), types.NamespacedName{Name: w.GetId(), Namespace: targetNS}, gotWS); err != nil {
		t.Fatalf("expected workspace to remain, got error: %v", err)
	}

	gotAW := &agentv1alpha1.AegisWorkload{}
	if err := cli.Get(context.Background(), types.NamespacedName{Name: w.GetId(), Namespace: targetNS}, gotAW); err != nil {
		t.Fatalf("expected aegisworkload to remain, got error: %v", err)
	}

	gotJob := &batchv1.Job{}
	if err := cli.Get(context.Background(), types.NamespacedName{Name: job.GetName(), Namespace: targetNS}, gotJob); err != nil {
		t.Fatalf("expected job to remain, got error: %v", err)
	}

	if still := srv.store.GetWorkload(w.GetId()); still == nil || still.GetStatus() != statusRunning {
		t.Fatalf("expected workload status to remain RUNNING, got %v", still)
	}
}

func TestTerminateWorkload_BillingCutsOffAtSuspendedAt(t *testing.T) {
	st := &fixedRuntimeStore{MemStore: store.NewMemStore(), runtimeSeconds: 3600}
	st.PutBudget(&aegis.Budget{ProjectId: "proj-1", Queue: "queue-a", LimitUsd: 1000, PolicyMode: "SOFT"})
	st.PutFlavor(&aegis.Flavor{Name: "f-1", GpuCount: 1, PriceUsdPerGpuHour: 1})

	srv := newTestServerWithStore(t, st)

	w := stubWorkspace("w-123", true, nil)
	w.Status = statusSuspended
	if wk, ok := w.GetKind().(*aegis.Workload_Workspace); ok && wk.Workspace != nil {
		wk.Workspace.Flavor = "f-1"
	}
	st.PutWorkload(w)

	if _, err := srv.TerminateWorkload(contextWithSubject("alice@example.com"), &aegis.TerminateWorkloadRequest{
		Id:     w.GetId(),
		Reason: "cleanup",
	}); err != nil {
		t.Fatalf("TerminateWorkload returned error: %v", err)
	}

	usage, ok := st.UsageView(w.GetProjectId(), w.GetQueue())
	if !ok {
		t.Fatalf("expected usage view to exist")
	}
	if usage.ActualUSD < 0.9 || usage.ActualUSD > 1.1 {
		t.Fatalf("expected actualUSD ~1.0, got %f", usage.ActualUSD)
	}
}

func TestTerminateWorkload_BillingExcludesSuspendedTimeAfterResume(t *testing.T) {
	st := &fixedRuntimeStore{
		MemStore:       store.NewMemStore(),
		runtimeSeconds: 3600, // 1h before suspension
		startedAtOK:    true,
		startedAt:      time.Now().Add(-2 * time.Hour), // 2h since resume
	}
	st.PutBudget(&aegis.Budget{ProjectId: "proj-1", Queue: "queue-a", LimitUsd: 1000, PolicyMode: "SOFT"})
	st.PutFlavor(&aegis.Flavor{Name: "f-1", GpuCount: 1, PriceUsdPerGpuHour: 1})

	srv := newTestServerWithStore(t, st)

	w := stubWorkspace("w-123", true, nil)
	w.Status = statusRunning
	w.SuspendedAtUtc = time.Now().Add(-12 * time.Hour).Format(time.RFC3339Nano) // stale value should not affect billing
	if wk, ok := w.GetKind().(*aegis.Workload_Workspace); ok && wk.Workspace != nil {
		wk.Workspace.Flavor = "f-1"
	}
	st.PutWorkload(w)

	if _, err := srv.TerminateWorkload(contextWithSubject("alice@example.com"), &aegis.TerminateWorkloadRequest{
		Id:     w.GetId(),
		Reason: "cleanup",
	}); err != nil {
		t.Fatalf("TerminateWorkload returned error: %v", err)
	}

	usage, ok := st.UsageView(w.GetProjectId(), w.GetQueue())
	if !ok {
		t.Fatalf("expected usage view to exist")
	}
	if usage.ActualUSD < 2.9 || usage.ActualUSD > 3.1 {
		t.Fatalf("expected actualUSD ~3.0, got %f", usage.ActualUSD)
	}
}

func TestListClusterWorkloadIDs(t *testing.T) {
	srv := newTestServer(t)
	srv.store.PutWorkload(stubWorkspace("w-123", true, nil))
	srv.store.PutWorkload(stubWorkspace("w-456", false, nil))
	srv.store.PutWorkload(&aegis.Workload{Id: "w-other", ProjectId: "proj-1", Queue: "queue-a", ClusterId: "cluster-2", Status: statusRunning})

	resp, err := srv.ListClusterWorkloadIDs(context.Background(), &aegis.ListClusterWorkloadIDsRequest{ClusterId: "cluster-1"})
	if err != nil {
		t.Fatalf("ListClusterWorkloadIDs returned error: %v", err)
	}
	got := resp.GetWorkloadIds()
	if len(got) != 2 {
		t.Fatalf("expected 2 workload IDs, got %v", got)
	}
	if got[0] != "w-123" || got[1] != "w-456" {
		t.Fatalf("unexpected workload IDs: %v", got)
	}
}

func TestListClusterWorkloadIDs_RequiresClusterID(t *testing.T) {
	srv := newTestServer(t)

	_, err := srv.ListClusterWorkloadIDs(context.Background(), &aegis.ListClusterWorkloadIDsRequest{})
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("expected InvalidArgument, got %v", err)
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
	if flavor.GetCpuCoresRequest() != "500m" || flavor.GetMemoryRequest() != "512Mi" {
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

func TestSubmitWorkload_RespectsRequestedCluster(t *testing.T) {
	srv := newTestServer(t)
	srv.autoBootstrap = true

	srv.store.PutProject(&aegis.Project{Id: "proj-1"})
	srv.store.UpsertClusterFromRegister(&aegis.ClusterRegisterRequest{ClusterId: "cluster-1", Provider: "aws", Region: "us-east-1"})
	srv.store.UpdateClusterFromHeartbeat(&aegis.ClusterHeartbeat{
		ClusterId:        "cluster-1",
		AvailableFlavors: []*aegis.Flavor{{Name: "cpu-small"}},
	})

	req := &aegis.SubmitWorkloadRequest{
		Workload: &aegis.Workload{
			ProjectId: "proj-1",
			Queue:     "queue-a",
			ClusterId: "cluster-1",
			Kind: &aegis.Workload_Workspace{
				Workspace: &aegis.WorkspaceSpec{
					Flavor: "cpu-small",
					Image:  "alpine:3.19",
				},
			},
		},
	}

	ctx := contextWithSubject("clustered@example.com")
	res, err := srv.SubmitWorkload(ctx, req)
	if err != nil {
		t.Fatalf("SubmitWorkload returned error: %v", err)
	}
	if res.GetClusterId() != "cluster-1" {
		t.Fatalf("expected workload pinned to requested cluster, got %q", res.GetClusterId())
	}
}

func TestWizardHandlers_ListProjectsAndClusters(t *testing.T) {
	srv := newTestServer(t)
	srv.store.PutProject(&aegis.Project{Id: "proj-1", DisplayName: "Project One"})
	srv.store.UpsertClusterFromRegister(&aegis.ClusterRegisterRequest{
		ClusterId: "proj-1-cluster",
		Provider:  "aws",
		Region:    "us-west-2",
		Labels:    map[string]string{"aegis.yourorg.dev/projectId": "proj-1", "k8sVersion": "1.27"},
	})
	srv.store.UpdateClusterFromHeartbeat(&aegis.ClusterHeartbeat{
		ClusterId:        "proj-1-cluster",
		AvailableFlavors: []*aegis.Flavor{{Name: "a10-1gpu"}},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	req = req.WithContext(contextWithSubject("wizard@example.com"))
	rec := httptest.NewRecorder()

	srv.handleProjects(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 response, got %d", rec.Code)
	}
	var resp struct {
		Projects []projectView `json:"projects"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Projects) != 1 {
		t.Fatalf("expected one project, got %d", len(resp.Projects))
	}
	if resp.Projects[0].ID != "proj-1" {
		t.Fatalf("expected project id proj-1, got %s", resp.Projects[0].ID)
	}
	if len(resp.Projects[0].Clusters) != 1 {
		t.Fatalf("expected one cluster, got %d", len(resp.Projects[0].Clusters))
	}
	cluster := resp.Projects[0].Clusters[0]
	if !cluster.HasGPU {
		t.Fatal("expected cluster to indicate GPU availability")
	}
	if cluster.Status != "ready" {
		t.Fatalf("expected cluster status ready, got %s", cluster.Status)
	}
}

func TestWizardHandlers_CreateWorkspace(t *testing.T) {
	srv := newTestServer(t)
	srv.autoBootstrap = true

	srv.store.PutProject(&aegis.Project{Id: "proj-1", DisplayName: "Project One"})
	srv.store.UpsertClusterFromRegister(&aegis.ClusterRegisterRequest{
		ClusterId: "proj-1-cluster",
		Provider:  "aws",
		Region:    "us-west-2",
		Labels:    map[string]string{"aegis.yourorg.dev/projectId": "proj-1"},
	})
	srv.store.UpdateClusterFromHeartbeat(&aegis.ClusterHeartbeat{
		ClusterId:        "proj-1-cluster",
		AvailableFlavors: []*aegis.Flavor{{Name: "cpu-small"}},
	})

	payload := workspaceCreateRequest{
		ProjectID: "proj-1",
		ClusterID: "proj-1-cluster",
		Name:      "demo-workspace",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/workspaces", bytes.NewReader(body))
	req = req.WithContext(contextWithSubject("wizard@example.com"))
	rec := httptest.NewRecorder()

	srv.handleCreateWorkspace(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201 response, got %d (%s)", rec.Code, rec.Body.String())
	}
	var resp workspaceCreateResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.ProjectID != payload.ProjectID || resp.ClusterID != payload.ClusterID {
		t.Fatalf("unexpected ids in response: %+v", resp)
	}
	if resp.Status == "" || resp.ID == "" {
		t.Fatalf("expected id and status to be populated: %+v", resp)
	}
	if stored := srv.store.GetWorkload(resp.ID); stored == nil {
		t.Fatalf("expected workload %s to be persisted", resp.ID)
	}
}

// ---- Project RPCs ----

func TestCreateProject(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("admin@example.com")

	proj := &aegis.Project{Id: "proj-new", DisplayName: "New Project", OwnerGroup: "team-alpha"}
	resp, err := srv.CreateProject(ctx, &aegis.CreateProjectRequest{Project: proj})
	if err != nil {
		t.Fatalf("CreateProject returned error: %v", err)
	}
	if resp.GetId() != "proj-new" {
		t.Fatalf("expected project id proj-new, got %s", resp.GetId())
	}
	if resp.GetDisplayName() != "New Project" {
		t.Fatalf("expected display name 'New Project', got %s", resp.GetDisplayName())
	}
	stored := srv.store.GetProject("proj-new")
	if stored == nil {
		t.Fatal("expected project to be persisted in store")
	}
	if stored.GetOwnerGroup() != "team-alpha" {
		t.Fatalf("expected owner_group team-alpha, got %s", stored.GetOwnerGroup())
	}
}

func TestListProjects(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("admin@example.com")

	srv.store.PutProject(&aegis.Project{Id: "proj-a", DisplayName: "Alpha"})
	srv.store.PutProject(&aegis.Project{Id: "proj-b", DisplayName: "Beta"})
	srv.store.PutProject(&aegis.Project{Id: "proj-c", DisplayName: "Gamma"})

	resp, err := srv.ListProjects(ctx, &aegis.ListProjectsRequest{})
	if err != nil {
		t.Fatalf("ListProjects returned error: %v", err)
	}
	if len(resp.GetItems()) != 3 {
		t.Fatalf("expected 3 projects, got %d", len(resp.GetItems()))
	}
	ids := map[string]bool{}
	for _, p := range resp.GetItems() {
		ids[p.GetId()] = true
	}
	for _, expected := range []string{"proj-a", "proj-b", "proj-c"} {
		if !ids[expected] {
			t.Fatalf("expected project %s in list, got %v", expected, ids)
		}
	}
}

// ---- Budget RPCs ----

func TestUpsertBudget(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("admin@example.com")

	srv.store.PutProject(&aegis.Project{Id: "proj-1"})

	budget := &aegis.Budget{ProjectId: "proj-1", Queue: "queue-a", LimitUsd: 500, PolicyMode: "HARD"}
	resp, err := srv.UpsertBudget(ctx, &aegis.UpsertBudgetRequest{Budget: budget})
	if err != nil {
		t.Fatalf("UpsertBudget returned error: %v", err)
	}
	if resp.GetProjectId() != "proj-1" {
		t.Fatalf("expected project_id proj-1, got %s", resp.GetProjectId())
	}
	if resp.GetLimitUsd() != 500 {
		t.Fatalf("expected limit_usd 500, got %f", resp.GetLimitUsd())
	}
	if resp.GetPolicyMode() != "HARD" {
		t.Fatalf("expected policy_mode HARD, got %s", resp.GetPolicyMode())
	}
	stored := srv.store.GetBudgetExact("proj-1", "queue-a")
	if stored == nil {
		t.Fatal("expected budget to be persisted in store")
	}
}

func TestGetBudget(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("admin@example.com")

	srv.store.PutBudget(&aegis.Budget{ProjectId: "proj-1", Queue: "", LimitUsd: 1000, PolicyMode: "SOFT"})

	resp, err := srv.GetBudget(ctx, &aegis.GetBudgetRequest{ProjectId: "proj-1"})
	if err != nil {
		t.Fatalf("GetBudget returned error: %v", err)
	}
	if resp.GetBudget().GetLimitUsd() != 1000 {
		t.Fatalf("expected limit_usd 1000, got %f", resp.GetBudget().GetLimitUsd())
	}
	if resp.GetUsage() == nil {
		t.Fatal("expected usage to be populated")
	}
	if resp.GetUsage().GetRemainingUsd() != 1000 {
		t.Fatalf("expected remaining_usd 1000, got %f", resp.GetUsage().GetRemainingUsd())
	}
}

func TestListBudgets(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("admin@example.com")

	srv.store.PutBudget(&aegis.Budget{ProjectId: "proj-1", Queue: "", LimitUsd: 1000, PolicyMode: "SOFT"})
	srv.store.PutBudget(&aegis.Budget{ProjectId: "proj-1", Queue: "queue-a", LimitUsd: 500, PolicyMode: "HARD"})
	srv.store.PutBudget(&aegis.Budget{ProjectId: "proj-2", Queue: "", LimitUsd: 2000, PolicyMode: "SOFT"})

	resp, err := srv.ListBudgets(ctx, &aegis.ListBudgetsRequest{ProjectId: "proj-1"})
	if err != nil {
		t.Fatalf("ListBudgets returned error: %v", err)
	}
	if len(resp.GetItems()) != 2 {
		t.Fatalf("expected 2 budgets for proj-1, got %d", len(resp.GetItems()))
	}
	for _, item := range resp.GetItems() {
		if item.GetBudget().GetProjectId() != "proj-1" {
			t.Fatalf("expected all budgets for proj-1, got %s", item.GetBudget().GetProjectId())
		}
		if item.GetUsage() == nil {
			t.Fatal("expected usage to be populated on each budget")
		}
	}
}

// ---- Flavor/Queue RPCs ----

func TestUpsertFlavor(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("admin@example.com")

	flavor := &aegis.Flavor{Name: "gpu-a100", GpuCount: 8, PriceUsdPerGpuHour: 3.5, Chip: "A100", ResourceName: "nvidia.com/gpu"}
	resp, err := srv.UpsertFlavor(ctx, &aegis.UpsertFlavorRequest{Flavor: flavor})
	if err != nil {
		t.Fatalf("UpsertFlavor returned error: %v", err)
	}
	if resp.GetName() != "gpu-a100" {
		t.Fatalf("expected flavor name gpu-a100, got %s", resp.GetName())
	}
	if resp.GetGpuCount() != 8 {
		t.Fatalf("expected gpu_count 8, got %d", resp.GetGpuCount())
	}
	stored := srv.store.GetFlavor("gpu-a100")
	if stored == nil {
		t.Fatal("expected flavor to be persisted in store")
	}
	if stored.GetPriceUsdPerGpuHour() != 3.5 {
		t.Fatalf("expected price 3.5, got %f", stored.GetPriceUsdPerGpuHour())
	}
}

func TestUpsertQueue(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("admin@example.com")

	srv.store.PutProject(&aegis.Project{Id: "proj-1"})
	srv.store.PutFlavor(&aegis.Flavor{Name: "cpu-small"})

	queue := &aegis.Queue{Name: "queue-main", ProjectId: "proj-1", PriorityTier: "high", AllowedFlavors: []string{"cpu-small"}}
	resp, err := srv.UpsertQueue(ctx, &aegis.UpsertQueueRequest{Queue: queue})
	if err != nil {
		t.Fatalf("UpsertQueue returned error: %v", err)
	}
	if resp.GetName() != "queue-main" {
		t.Fatalf("expected queue name queue-main, got %s", resp.GetName())
	}
	if resp.GetProjectId() != "proj-1" {
		t.Fatalf("expected project_id proj-1, got %s", resp.GetProjectId())
	}
	stored := srv.store.GetQueue("queue-main")
	if stored == nil {
		t.Fatal("expected queue to be persisted in store")
	}
	if stored.GetPriorityTier() != "high" {
		t.Fatalf("expected priority_tier high, got %s", stored.GetPriorityTier())
	}
}

// ---- Cluster RPCs ----

func TestRegisterCluster(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("agent@cluster")

	resp, err := srv.RegisterCluster(ctx, &aegis.ClusterRegisterRequest{
		ClusterId: "cluster-new",
		Provider:  "aws",
		Region:    "us-east-1",
		Labels:    map[string]string{"env": "production"},
	})
	if err != nil {
		t.Fatalf("RegisterCluster returned error: %v", err)
	}
	if !resp.GetOk() {
		t.Fatalf("expected ok=true, got %v", resp.GetOk())
	}
	ci := srv.store.GetClusterInfo("cluster-new")
	if ci == nil {
		t.Fatal("expected cluster to be persisted in store")
	}
	if ci.Provider != "aws" {
		t.Fatalf("expected provider aws, got %s", ci.Provider)
	}
	if ci.Region != "us-east-1" {
		t.Fatalf("expected region us-east-1, got %s", ci.Region)
	}
}

func TestHeartbeat(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("agent@cluster")

	srv.store.UpsertClusterFromRegister(&aegis.ClusterRegisterRequest{
		ClusterId: "cluster-hb",
		Provider:  "aws",
		Region:    "us-west-2",
	})

	resp, err := srv.Heartbeat(ctx, &aegis.ClusterHeartbeat{
		ClusterId:        "cluster-hb",
		TtfGpuSecondsP50: 12.5,
		AvailableFlavors: []*aegis.Flavor{{Name: "gpu-a100"}, {Name: "cpu-small"}},
	})
	if err != nil {
		t.Fatalf("Heartbeat returned error: %v", err)
	}
	if !resp.GetOk() {
		t.Fatalf("expected ok=true, got %v", resp.GetOk())
	}
	ci := srv.store.GetClusterInfo("cluster-hb")
	if ci == nil {
		t.Fatal("expected cluster to exist after heartbeat")
	}
	if ci.LastHeartbeat.IsZero() {
		t.Fatal("expected last_heartbeat to be updated")
	}
	if !ci.AvailableFlavorSet["gpu-a100"] {
		t.Fatalf("expected gpu-a100 in available flavors, got %v", ci.AvailableFlavorSet)
	}
}

func TestListClusters(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("admin@example.com")

	srv.store.UpsertClusterFromRegister(&aegis.ClusterRegisterRequest{ClusterId: "c-1", Provider: "aws", Region: "us-east-1"})
	srv.store.UpdateClusterFromHeartbeat(&aegis.ClusterHeartbeat{ClusterId: "c-1", AvailableFlavors: []*aegis.Flavor{{Name: "cpu-small"}}})

	srv.store.UpsertClusterFromRegister(&aegis.ClusterRegisterRequest{ClusterId: "c-2", Provider: "gcp", Region: "us-central1"})
	srv.store.UpdateClusterFromHeartbeat(&aegis.ClusterHeartbeat{ClusterId: "c-2", AvailableFlavors: []*aegis.Flavor{{Name: "gpu-a100"}}})

	resp, err := srv.ListClusters(ctx, &aegis.ListClustersRequest{})
	if err != nil {
		t.Fatalf("ListClusters returned error: %v", err)
	}
	if len(resp.GetItems()) != 2 {
		t.Fatalf("expected 2 clusters, got %d", len(resp.GetItems()))
	}
	ids := map[string]bool{}
	for _, c := range resp.GetItems() {
		ids[c.GetId()] = true
	}
	if !ids["c-1"] || !ids["c-2"] {
		t.Fatalf("expected clusters c-1 and c-2 in list, got %v", ids)
	}
}

// ---- Workspace RPC ----

func TestCreateWorkspace(t *testing.T) {
	srv := newTestServer(t)
	srv.autoBootstrap = true
	ctx := contextWithSubject("user@example.com")

	srv.store.PutProject(&aegis.Project{Id: "proj-1"})
	srv.store.PutFlavor(&aegis.Flavor{Name: "cpu-small"})
	srv.store.PutQueue(&aegis.Queue{Name: "default", ProjectId: "proj-1"})
	srv.store.UpsertClusterFromRegister(&aegis.ClusterRegisterRequest{ClusterId: "cluster-1", Provider: "aws", Region: "us-east-1"})
	srv.store.UpdateClusterFromHeartbeat(&aegis.ClusterHeartbeat{
		ClusterId:        "cluster-1",
		AvailableFlavors: []*aegis.Flavor{{Name: "cpu-small"}},
	})

	resp, err := srv.CreateWorkspace(ctx, &aegis.CreateWorkspaceRequest{
		ProjectId: "proj-1",
		Queue:     "default",
		Workspace: &aegis.WorkspaceSpec{
			Flavor: "cpu-small",
			Image:  "alpine:3.19",
		},
	})
	if err != nil {
		t.Fatalf("CreateWorkspace returned error: %v", err)
	}
	if resp.GetWorkload() == nil {
		t.Fatal("expected workload in response")
	}
	wl := resp.GetWorkload()
	if wl.GetId() == "" {
		t.Fatal("expected workload id to be assigned")
	}
	if wl.GetProjectId() != "proj-1" {
		t.Fatalf("expected project_id proj-1, got %s", wl.GetProjectId())
	}
	if wl.GetClusterId() != "cluster-1" {
		t.Fatalf("expected cluster_id cluster-1, got %s", wl.GetClusterId())
	}
	if wl.GetStatus() != "PLACED" {
		t.Fatalf("expected status PLACED, got %s", wl.GetStatus())
	}
	if stored := srv.store.GetWorkload(wl.GetId()); stored == nil {
		t.Fatal("expected workload to be persisted in store")
	}
}

// ---- Workload Lifecycle RPCs ----

func TestLeaseWorkload(t *testing.T) {
	srv := newTestServer(t)
	srv.autoBootstrap = true
	ctx := contextWithSubject("user@example.com")

	// Set up prerequisites and submit a workload
	srv.store.PutProject(&aegis.Project{Id: "proj-1"})
	srv.store.PutFlavor(&aegis.Flavor{Name: "cpu-small", GpuCount: 1, PriceUsdPerGpuHour: 1})
	srv.store.PutQueue(&aegis.Queue{Name: "queue-a", ProjectId: "proj-1"})
	srv.store.UpsertClusterFromRegister(&aegis.ClusterRegisterRequest{ClusterId: "cluster-1", Provider: "aws", Region: "us-east-1"})
	srv.store.UpdateClusterFromHeartbeat(&aegis.ClusterHeartbeat{
		ClusterId:        "cluster-1",
		AvailableFlavors: []*aegis.Flavor{{Name: "cpu-small"}},
	})

	submitted, err := srv.SubmitWorkload(ctx, &aegis.SubmitWorkloadRequest{
		Workload: &aegis.Workload{
			ProjectId: "proj-1",
			Queue:     "queue-a",
			ClusterId: "cluster-1",
			Kind: &aegis.Workload_Workspace{
				Workspace: &aegis.WorkspaceSpec{Flavor: "cpu-small", Image: "alpine:3.19"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SubmitWorkload returned error: %v", err)
	}
	if submitted.GetStatus() != "PLACED" {
		t.Fatalf("expected PLACED status, got %s", submitted.GetStatus())
	}

	// Lease it
	leaseResp, err := srv.LeaseWorkload(context.Background(), &aegis.LeaseWorkloadRequest{
		ClusterId: "cluster-1",
		Max:       10,
	})
	if err != nil {
		t.Fatalf("LeaseWorkload returned error: %v", err)
	}
	if len(leaseResp.GetItems()) != 1 {
		t.Fatalf("expected 1 leased workload, got %d", len(leaseResp.GetItems()))
	}
	leased := leaseResp.GetItems()[0]
	if leased.GetId() != submitted.GetId() {
		t.Fatalf("expected leased workload id %s, got %s", submitted.GetId(), leased.GetId())
	}
	if leased.GetStatus() != "RUNNING" {
		t.Fatalf("expected RUNNING status after lease, got %s", leased.GetStatus())
	}
}

func TestStartWorkload(t *testing.T) {
	srv := newTestServer(t)
	srv.autoBootstrap = true
	ctx := contextWithSubject("user@example.com")

	srv.store.PutProject(&aegis.Project{Id: "proj-1"})
	srv.store.PutFlavor(&aegis.Flavor{Name: "cpu-small", GpuCount: 1, PriceUsdPerGpuHour: 1})
	srv.store.PutQueue(&aegis.Queue{Name: "queue-a", ProjectId: "proj-1"})
	srv.store.UpsertClusterFromRegister(&aegis.ClusterRegisterRequest{ClusterId: "cluster-1", Provider: "aws", Region: "us-east-1"})
	srv.store.UpdateClusterFromHeartbeat(&aegis.ClusterHeartbeat{
		ClusterId:        "cluster-1",
		AvailableFlavors: []*aegis.Flavor{{Name: "cpu-small"}},
	})

	submitted, err := srv.SubmitWorkload(ctx, &aegis.SubmitWorkloadRequest{
		Workload: &aegis.Workload{
			ProjectId: "proj-1",
			Queue:     "queue-a",
			ClusterId: "cluster-1",
			Kind: &aegis.Workload_Workspace{
				Workspace: &aegis.WorkspaceSpec{Flavor: "cpu-small", Image: "alpine:3.19"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SubmitWorkload returned error: %v", err)
	}

	// StartWorkload transitions from PLACED to RUNNING
	startResp, err := srv.StartWorkload(context.Background(), &aegis.StartWorkloadRequest{
		Id:        submitted.GetId(),
		ClusterId: "cluster-1",
	})
	if err != nil {
		t.Fatalf("StartWorkload returned error: %v", err)
	}
	if startResp.GetWorkload().GetStatus() != "RUNNING" {
		t.Fatalf("expected RUNNING status after start, got %s", startResp.GetWorkload().GetStatus())
	}
	if startResp.GetWorkload().GetId() != submitted.GetId() {
		t.Fatalf("expected workload id %s, got %s", submitted.GetId(), startResp.GetWorkload().GetId())
	}
}

func TestAckWorkload(t *testing.T) {
	srv := newTestServer(t)
	srv.autoBootstrap = true
	ctx := contextWithSubject("user@example.com")

	srv.store.PutProject(&aegis.Project{Id: "proj-1"})
	srv.store.PutFlavor(&aegis.Flavor{Name: "cpu-small", GpuCount: 1, PriceUsdPerGpuHour: 1})
	srv.store.PutQueue(&aegis.Queue{Name: "queue-a", ProjectId: "proj-1"})
	srv.store.UpsertClusterFromRegister(&aegis.ClusterRegisterRequest{ClusterId: "cluster-1", Provider: "aws", Region: "us-east-1"})
	srv.store.UpdateClusterFromHeartbeat(&aegis.ClusterHeartbeat{
		ClusterId:        "cluster-1",
		AvailableFlavors: []*aegis.Flavor{{Name: "cpu-small"}},
	})

	submitted, err := srv.SubmitWorkload(ctx, &aegis.SubmitWorkloadRequest{
		Workload: &aegis.Workload{
			ProjectId: "proj-1",
			Queue:     "queue-a",
			ClusterId: "cluster-1",
			Kind: &aegis.Workload_Workspace{
				Workspace: &aegis.WorkspaceSpec{Flavor: "cpu-small", Image: "alpine:3.19"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SubmitWorkload returned error: %v", err)
	}

	// Start the workload (PLACED -> RUNNING)
	_, err = srv.StartWorkload(context.Background(), &aegis.StartWorkloadRequest{
		Id:        submitted.GetId(),
		ClusterId: "cluster-1",
	})
	if err != nil {
		t.Fatalf("StartWorkload returned error: %v", err)
	}

	// Ack the workload (RUNNING -> SUCCEEDED)
	ackResp, err := srv.AckWorkload(context.Background(), &aegis.AckWorkloadRequest{
		Id:     submitted.GetId(),
		Status: "SUCCEEDED",
		Url:    "https://workload.example.com",
	})
	if err != nil {
		t.Fatalf("AckWorkload returned error: %v", err)
	}
	if ackResp.GetWorkload().GetStatus() != "SUCCEEDED" {
		t.Fatalf("expected SUCCEEDED status after ack, got %s", ackResp.GetWorkload().GetStatus())
	}
	if ackResp.GetWorkload().GetUrl() != "https://workload.example.com" {
		t.Fatalf("expected url to be set, got %s", ackResp.GetWorkload().GetUrl())
	}
}

// ---- Audit Events ----

func TestAuditEventsOnSubmit(t *testing.T) {
	srv := newTestServer(t)
	srv.autoBootstrap = true
	ctx := contextWithSubject("auditor@example.com")

	srv.store.PutProject(&aegis.Project{Id: "proj-1"})
	srv.store.PutFlavor(&aegis.Flavor{Name: "cpu-small", GpuCount: 1, PriceUsdPerGpuHour: 1})
	srv.store.PutQueue(&aegis.Queue{Name: "queue-a", ProjectId: "proj-1"})
	srv.store.UpsertClusterFromRegister(&aegis.ClusterRegisterRequest{ClusterId: "cluster-1", Provider: "aws", Region: "us-east-1"})
	srv.store.UpdateClusterFromHeartbeat(&aegis.ClusterHeartbeat{
		ClusterId:        "cluster-1",
		AvailableFlavors: []*aegis.Flavor{{Name: "cpu-small"}},
	})

	submitted, err := srv.SubmitWorkload(ctx, &aegis.SubmitWorkloadRequest{
		Workload: &aegis.Workload{
			ProjectId: "proj-1",
			Queue:     "queue-a",
			ClusterId: "cluster-1",
			Kind: &aegis.Workload_Workspace{
				Workspace: &aegis.WorkspaceSpec{Flavor: "cpu-small", Image: "alpine:3.19"},
			},
		},
	})
	if err != nil {
		t.Fatalf("SubmitWorkload returned error: %v", err)
	}

	events, err := srv.store.ListAuditEvents(store.AuditEventFilter{
		EventType:  "workload.submitted",
		ResourceID: submitted.GetId(),
	})
	if err != nil {
		t.Fatalf("ListAuditEvents returned error: %v", err)
	}
	if len(events) == 0 {
		t.Fatal("expected at least one audit event for workload.submitted")
	}
	ev := events[0]
	if ev.Subject != "auditor@example.com" {
		t.Fatalf("expected subject auditor@example.com, got %s", ev.Subject)
	}
	if ev.ResourceType != "workload" {
		t.Fatalf("expected resource_type workload, got %s", ev.ResourceType)
	}
	if ev.Action != "create" {
		t.Fatalf("expected action create, got %s", ev.Action)
	}
	if ev.Outcome != "success" {
		t.Fatalf("expected outcome success, got %s", ev.Outcome)
	}
	if ev.Details["project_id"] != "proj-1" {
		t.Fatalf("expected details.project_id proj-1, got %s", ev.Details["project_id"])
	}
}

// --- Auto-derivation tests ---

func TestRegisterClusterAutoDerivesProjectID(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("admin")

	// Create the project first so derivation succeeds
	srv.store.PutProject(&aegis.Project{Id: "db-1", OwnerGroup: "admin"})

	// Register cluster without project label
	resp, err := srv.RegisterCluster(ctx, &aegis.ClusterRegisterRequest{
		ClusterId: "db-1-us-east-1-atlas-train-govcloud",
		Provider:  "aws",
		Region:    "us-east-1",
	})
	if err != nil {
		t.Fatalf("RegisterCluster returned error: %v", err)
	}
	if !resp.GetOk() {
		t.Fatalf("expected ok=true, got false")
	}

	// Verify project ID was persisted
	pid, ok := srv.store.GetClusterProjectID("db-1-us-east-1-atlas-train-govcloud")
	if !ok || pid != "db-1" {
		t.Fatalf("expected project_id=db-1, got %q (ok=%v)", pid, ok)
	}
}

func TestRegisterClusterExplicitLabelTakesPriority(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("admin")

	srv.store.PutProject(&aegis.Project{Id: "db-1", OwnerGroup: "admin"})
	srv.store.PutProject(&aegis.Project{Id: "explicit-proj", OwnerGroup: "admin"})

	// Register cluster WITH explicit project label (should override derivation)
	_, err := srv.RegisterCluster(ctx, &aegis.ClusterRegisterRequest{
		ClusterId: "db-1-us-east-1-atlas-train-govcloud",
		Provider:  "aws",
		Region:    "us-east-1",
		Labels:    map[string]string{"aegis.yourorg.dev/projectId": "explicit-proj"},
	})
	if err != nil {
		t.Fatalf("RegisterCluster returned error: %v", err)
	}

	pid, ok := srv.store.GetClusterProjectID("db-1-us-east-1-atlas-train-govcloud")
	if !ok || pid != "explicit-proj" {
		t.Fatalf("expected project_id=explicit-proj, got %q (ok=%v)", pid, ok)
	}
}

func TestRegisterClusterProjectNotFoundSkipsDerivation(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("admin")

	// Do NOT create the project — derivation should be skipped
	_, err := srv.RegisterCluster(ctx, &aegis.ClusterRegisterRequest{
		ClusterId: "db-1-us-east-1-atlas-train-govcloud",
		Provider:  "aws",
		Region:    "us-east-1",
	})
	if err != nil {
		t.Fatalf("RegisterCluster returned error: %v", err)
	}

	pid, _ := srv.store.GetClusterProjectID("db-1-us-east-1-atlas-train-govcloud")
	if pid != "" {
		t.Fatalf("expected empty project_id (project not found), got %q", pid)
	}
}

func TestHeartbeatBackfillsProjectID(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("admin")

	// Register cluster without project (project doesn't exist yet)
	_, err := srv.RegisterCluster(ctx, &aegis.ClusterRegisterRequest{
		ClusterId: "db-1-us-east-1-atlas-train-govcloud",
		Provider:  "aws",
		Region:    "us-east-1",
	})
	if err != nil {
		t.Fatalf("RegisterCluster returned error: %v", err)
	}

	pid, _ := srv.store.GetClusterProjectID("db-1-us-east-1-atlas-train-govcloud")
	if pid != "" {
		t.Fatalf("expected empty project_id before project creation, got %q", pid)
	}

	// Now create the project
	srv.store.PutProject(&aegis.Project{Id: "db-1", OwnerGroup: "admin"})

	// Heartbeat should backfill
	_, err = srv.Heartbeat(ctx, &aegis.ClusterHeartbeat{
		ClusterId: "db-1-us-east-1-atlas-train-govcloud",
	})
	if err != nil {
		t.Fatalf("Heartbeat returned error: %v", err)
	}

	pid, ok := srv.store.GetClusterProjectID("db-1-us-east-1-atlas-train-govcloud")
	if !ok || pid != "db-1" {
		t.Fatalf("expected project_id=db-1 after heartbeat, got %q (ok=%v)", pid, ok)
	}
}

func TestHeartbeatDoesNotOverwriteExistingProjectID(t *testing.T) {
	srv := newTestServer(t)
	ctx := contextWithSubject("admin")

	srv.store.PutProject(&aegis.Project{Id: "original-proj", OwnerGroup: "admin"})
	srv.store.PutProject(&aegis.Project{Id: "db-1", OwnerGroup: "admin"})

	// Register cluster with explicit project
	_, err := srv.RegisterCluster(ctx, &aegis.ClusterRegisterRequest{
		ClusterId: "db-1-us-east-1-atlas-train-govcloud",
		Provider:  "aws",
		Region:    "us-east-1",
		Labels:    map[string]string{"aegis.yourorg.dev/projectId": "original-proj"},
	})
	if err != nil {
		t.Fatalf("RegisterCluster returned error: %v", err)
	}

	// Heartbeat should NOT overwrite the existing project ID
	_, err = srv.Heartbeat(ctx, &aegis.ClusterHeartbeat{
		ClusterId: "db-1-us-east-1-atlas-train-govcloud",
	})
	if err != nil {
		t.Fatalf("Heartbeat returned error: %v", err)
	}

	pid, ok := srv.store.GetClusterProjectID("db-1-us-east-1-atlas-train-govcloud")
	if !ok || pid != "original-proj" {
		t.Fatalf("expected project_id=original-proj (unchanged), got %q (ok=%v)", pid, ok)
	}
}
