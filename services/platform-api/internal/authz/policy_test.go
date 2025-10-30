package authz

import (
	"os"
	"testing"

	mw "github.com/yourorg/aegis/services/platform-api/internal/server/mw"
)

func TestLoadPolicyFromEnvParsesBindings(t *testing.T) {
	t.Setenv("AUTHZ_ROLE_BINDINGS_JSON", `[
		{
			"roles": ["Aegis-Admin", "aegis-admin"],
			"projects": ["*", "PROJECT-123"],
			"queues": ["*", "High-Priority"]
		},
		{
			"roles": ["project-owner"],
			"projects": ["Project-123"],
			"queues": ["queue-a", "QUEUE-A"]
		}
	]`)

	policy, err := LoadPolicyFromEnv()
	if err != nil {
		t.Fatalf("LoadPolicyFromEnv returned error: %v", err)
	}
	if policy == nil || len(policy.bindings) != 2 {
		t.Fatalf("expected 2 bindings, got %#v", policy)
	}

	admin := policy.bindings[0]
	if !contains(admin.Roles, "aegis-admin") {
		t.Fatalf("expected normalized role aegis-admin, got %v", admin.Roles)
	}
	if !contains(admin.Projects, "*") || !contains(admin.Projects, "project-123") {
		t.Fatalf("expected normalized projects, got %v", admin.Projects)
	}
	if !contains(admin.Queues, "*") || !contains(admin.Queues, "high-priority") {
		t.Fatalf("expected normalized queues, got %v", admin.Queues)
	}

	owner := policy.bindings[1]
	if len(owner.Roles) != 1 || owner.Roles[0] != "project-owner" {
		t.Fatalf("unexpected owner roles %v", owner.Roles)
	}
	if len(owner.Projects) != 1 || owner.Projects[0] != "project-123" {
		t.Fatalf("unexpected owner project scope %v", owner.Projects)
	}
	if len(owner.Queues) != 1 || owner.Queues[0] != "queue-a" {
		t.Fatalf("unexpected owner queue scope %v", owner.Queues)
	}
}

func TestAuthorizeEvaluatesRoleAndScopes(t *testing.T) {
	t.Setenv("AUTHZ_ROLE_BINDINGS_JSON", `[
		{
			"roles": ["aegis-admin"],
			"projects": ["*"],
			"queues": ["*"]
		},
		{
			"roles": ["project-owner"],
			"projects": ["project-123"],
			"queues": ["queue-a"]
		}
	]`)
	policy, err := LoadPolicyFromEnv()
	if err != nil {
		t.Fatalf("LoadPolicyFromEnv returned error: %v", err)
	}

	adminIdentity := &mw.Identity{Roles: []string{"Aegis-Admin"}}
	if !policy.Authorize(adminIdentity, "project-999", "queue-x") {
		t.Fatal("expected admin to be authorized for any project/queue")
	}

	projectOwner := &mw.Identity{Roles: []string{"project-owner"}}
	if !policy.Authorize(projectOwner, "project-123", "queue-a") {
		t.Fatal("expected project owner to access bound queue")
	}
	if policy.Authorize(projectOwner, "project-123", "queue-b") {
		t.Fatal("expected project owner to be denied for unmatched queue")
	}
	if policy.Authorize(projectOwner, "project-456", "queue-a") {
		t.Fatal("expected project owner to be denied for unmatched project")
	}

	noRoles := &mw.Identity{Roles: []string{}}
	if policy.Authorize(noRoles, "project-123", "queue-a") {
		t.Fatal("expected identity without roles to be denied")
	}

	if policy.Authorize(nil, "project-123", "queue-a") {
		t.Fatal("expected nil identity to be denied when bindings are configured")
	}
}

func TestAuthorizeWithEmptyPolicyDeniesAll(t *testing.T) {
	t.Setenv("AUTHZ_ROLE_BINDINGS_JSON", "")
	policy, err := LoadPolicyFromEnv()
	if err != nil {
		t.Fatalf("LoadPolicyFromEnv returned error: %v", err)
	}
	if policy.Authorize(&mw.Identity{Roles: []string{"anything"}}, "project-1", "queue-1") {
		t.Fatal("expected empty policy to deny all requests (fail-closed)")
	}
}

func TestAuthorizeWithEmptyArrayDeniesAll(t *testing.T) {
	t.Setenv("AUTHZ_ROLE_BINDINGS_JSON", "[]")
	policy, err := LoadPolicyFromEnv()
	if err != nil {
		t.Fatalf("LoadPolicyFromEnv returned error: %v", err)
	}
	if policy.Authorize(&mw.Identity{Roles: []string{"anything"}}, "project-1", "queue-1") {
		t.Fatal("expected empty array policy to deny all requests (fail-closed)")
	}
}

func contains(values []string, candidate string) bool {
	for _, v := range values {
		if v == candidate {
			return true
		}
	}
	return false
}

func TestMain(m *testing.M) {
	os.Clearenv()
	code := m.Run()
	os.Exit(code)
}
