package placement

import (
	"sort"
	"strings"
	"sync"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
)

// ProjectPolicy captures the effective placement policy.
type ProjectPolicy struct {
	ProjectID        string
	AllowedRegions   []string
	AllowedProviders []string
	DefaultFlavor    string
	Quotas           map[string]int
	Strategy         string
}

// PolicyOverlay offers a threadsafe view of project-level placement policies.
type PolicyOverlay struct {
	mu      sync.RWMutex
	entries map[string]ProjectPolicy
}

// NewPolicyOverlay returns an initialized overlay instance.
func NewPolicyOverlay() *PolicyOverlay {
	return &PolicyOverlay{entries: make(map[string]ProjectPolicy)}
}

// Set updates the effective policy for the provided project.
func (o *PolicyOverlay) Set(spec *infraapi.ProjectPlacementSpec) {
	if spec == nil || strings.TrimSpace(spec.ProjectID) == "" {
		return
	}
	policy := ProjectPolicy{
		ProjectID:        strings.TrimSpace(spec.ProjectID),
		AllowedRegions:   normalizeList(spec.AllowedRegions),
		AllowedProviders: normalizeList(spec.AllowedProviders),
		DefaultFlavor:    strings.TrimSpace(spec.DefaultFlavor),
		Quotas:           copyQuota(spec.Quotas),
		Strategy:         strings.TrimSpace(spec.Strategy),
	}
	o.mu.Lock()
	o.entries[policy.ProjectID] = policy
	o.mu.Unlock()
}

// Delete removes a policy from the overlay.
func (o *PolicyOverlay) Delete(projectID string) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return
	}
	o.mu.Lock()
	delete(o.entries, projectID)
	o.mu.Unlock()
}

// Effective returns a copy of the effective policy if present.
func (o *PolicyOverlay) Effective(projectID string) (ProjectPolicy, bool) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return ProjectPolicy{}, false
	}
	o.mu.RLock()
	policy, ok := o.entries[projectID]
	o.mu.RUnlock()
	if !ok {
		return ProjectPolicy{}, false
	}
	policy.Quotas = copyQuota(policy.Quotas)
	policy.AllowedRegions = append([]string(nil), policy.AllowedRegions...)
	policy.AllowedProviders = append([]string(nil), policy.AllowedProviders...)
	return policy, true
}

// Snapshot returns a copy of all configured policies keyed by project ID.
func (o *PolicyOverlay) Snapshot() map[string]ProjectPolicy {
	o.mu.RLock()
	defer o.mu.RUnlock()
	out := make(map[string]ProjectPolicy, len(o.entries))
	for k, v := range o.entries {
		cp := v
		cp.Quotas = copyQuota(v.Quotas)
		cp.AllowedRegions = append([]string(nil), v.AllowedRegions...)
		cp.AllowedProviders = append([]string(nil), v.AllowedProviders...)
		out[k] = cp
	}
	return out
}

func copyQuota(in map[string]int) map[string]int {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]int, len(in))
	for k, v := range in {
		key := strings.TrimSpace(k)
		if key == "" {
			continue
		}
		out[strings.ToLower(key)] = v
	}
	return out
}

func normalizeList(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	dedup := make(map[string]struct{}, len(in))
	for _, v := range in {
		if trimmed := strings.TrimSpace(v); trimmed != "" {
			dedup[strings.ToLower(trimmed)] = struct{}{}
		}
	}
	if len(dedup) == 0 {
		return nil
	}
	out := make([]string, 0, len(dedup))
	for k := range dedup {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
