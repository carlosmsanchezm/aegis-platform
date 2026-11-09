package authz

import (
	"testing"

	mw "github.com/yourorg/aegis/services/platform-api/internal/server/mw"
)

func TestAuthorize_AllowsMatchingRoleAndClient(t *testing.T) {
	policy := &Policy{bindings: []Binding{{
		Roles:    []string{"workspace-admin"},
		Projects: []string{"*"},
		Queues:   []string{"*"},
		Clients:  []string{"backstage"},
	}}}

	identity := &mw.Identity{Roles: []string{"workspace-admin"}, ClientID: "Backstage"}
	if !policy.Authorize(identity, "proj-1", "default") {
		t.Fatalf("expected authorization to succeed for matching role and client")
	}
}

func TestAuthorize_DeniesMismatchedClient(t *testing.T) {
	policy := &Policy{bindings: []Binding{{
		Roles:    []string{"workspace-admin"},
		Projects: []string{"*"},
		Queues:   []string{"*"},
		Clients:  []string{"backstage"},
	}}}

	identity := &mw.Identity{Roles: []string{"workspace-admin"}, ClientID: "vscode-extension"}
	if policy.Authorize(identity, "proj-1", "default") {
		t.Fatalf("expected authorization to fail when client does not match")
	}
}

func TestAuthorize_AllowsWhenClientsUnset(t *testing.T) {
	policy := &Policy{bindings: []Binding{{
		Roles:    []string{"workspace-admin"},
		Projects: []string{"*"},
		Queues:   []string{"*"},
		Clients:  nil,
	}}}

	identity := &mw.Identity{Roles: []string{"workspace-admin"}, ClientID: "other"}
	if !policy.Authorize(identity, "proj-1", "default") {
		t.Fatalf("expected authorization to allow when clients list empty")
	}
}

func TestAuthorize_DeniesWhenNoRoles(t *testing.T) {
	policy := &Policy{bindings: []Binding{{
		Roles:    []string{"workspace-admin"},
		Projects: []string{"*"},
		Queues:   []string{"*"},
		Clients:  []string{"backstage"},
	}}}

	identity := &mw.Identity{ClientID: "backstage"}
	if policy.Authorize(identity, "proj-1", "default") {
		t.Fatalf("expected authorization to fail when identity lacks roles")
	}
}
