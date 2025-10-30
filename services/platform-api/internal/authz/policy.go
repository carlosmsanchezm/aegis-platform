package authz

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	mw "github.com/yourorg/aegis/services/platform-api/internal/server/mw"
)

// Binding associates one or more roles and client identifiers with permitted project and queue scopes.
type Binding struct {
	Roles    []string `json:"roles"`
	Projects []string `json:"projects"`
	Queues   []string `json:"queues"`
	Clients  []string `json:"clients"`
}

// Policy represents an in-memory authorization map derived from static configuration.
type Policy struct {
	bindings []Binding
}

// LoadPolicyFromEnv loads a policy from the AUTHZ_ROLE_BINDINGS_JSON environment variable.
// The variable accepts a JSON array of bindings. When unset or empty, the resulting policy
// denies all requests (fail-closed for security).
func LoadPolicyFromEnv(defaultJSON string) (*Policy, error) {
	raw := strings.TrimSpace(os.Getenv("AUTHZ_ROLE_BINDINGS_JSON"))
	if raw == "" {
		raw = strings.TrimSpace(defaultJSON)
	}
	if raw == "" {
		return &Policy{}, nil
	}
	var bindings []Binding
	if err := json.Unmarshal([]byte(raw), &bindings); err != nil {
		return nil, fmt.Errorf("failed to parse AUTHZ_ROLE_BINDINGS_JSON: %w", err)
	}
	for i := range bindings {
		bindings[i].Roles = normalize(bindingStrings(bindings[i].Roles))
		bindings[i].Projects = normalize(bindingStrings(bindings[i].Projects))
		bindings[i].Queues = normalize(bindingStrings(bindings[i].Queues))
		bindings[i].Clients = normalize(bindingStrings(bindings[i].Clients))
	}
	return &Policy{bindings: bindings}, nil
}

// Authorize returns true when the supplied identity is allowed to act on the given project/queue.
// When no bindings are configured the policy denies all requests (fail-closed for security).
func (p *Policy) Authorize(identity *mw.Identity, projectID, queue string) bool {
	if len(p.bindings) == 0 {
		return false
	}
	if identity == nil {
		return false
	}
	roles := toSet(identity.Roles)
	if len(roles) == 0 {
		return false
	}
	project := strings.ToLower(strings.TrimSpace(projectID))
	queueID := strings.ToLower(strings.TrimSpace(queue))
	clientID := strings.ToLower(strings.TrimSpace(identity.ClientID))

	for _, binding := range p.bindings {
		if !matchesClient(clientID, binding.Clients) {
			continue
		}
		if !hasRoleIntersection(roles, binding.Roles) {
			continue
		}
		if !matchesScope(project, binding.Projects) {
			continue
		}
		if !matchesScope(queueID, binding.Queues) {
			continue
		}
		return true
	}
	return false
}

func hasRoleIntersection(roles map[string]struct{}, required []string) bool {
	if len(required) == 0 {
		return true
	}
	for _, role := range required {
		if _, ok := roles[role]; ok {
			return true
		}
	}
	return false
}

func matchesScope(value string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, candidate := range allowed {
		switch candidate {
		case "*":
			return true
		case value:
			return true
		}
	}
	return false
}

func matchesClient(client string, allowed []string) bool {
	if len(allowed) == 0 {
		return true
	}
	if client == "" {
		return false
	}
	for _, candidate := range allowed {
		if candidate == client {
			return true
		}
	}
	return false
}

func toSet(values []string) map[string]struct{} {
	set := make(map[string]struct{}, len(values))
	for _, v := range values {
		if trimmed := strings.TrimSpace(strings.ToLower(v)); trimmed != "" {
			set[trimmed] = struct{}{}
		}
	}
	return set
}

func normalize(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	set := make(map[string]struct{}, len(values))
	for _, v := range values {
		trimmed := strings.TrimSpace(strings.ToLower(v))
		if trimmed == "" {
			continue
		}
		set[trimmed] = struct{}{}
	}
	if len(set) == 0 {
		return nil
	}
	out := make([]string, 0, len(set))
	for v := range set {
		out = append(out, v)
	}
	return out
}

func bindingStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}
