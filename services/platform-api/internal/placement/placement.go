package placement

import (
	"fmt"
	"math"
	"strings"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

// Candidate captures the data required to select a target cluster.
type Candidate struct {
	ClusterID   string
	Provider    string
	Region      string
	Labels      map[string]string
	TTFGSeconds float64
	Flavors     map[string]bool // set for quick lookup
}

// PolicyDomain conveys placement policy constraints.
type PolicyDomain struct {
	Regions   []string
	Providers []string
	Strategy  string
}

// ChooseCluster filters candidates by policy constraints and selects a cluster
// using the specified strategy. clusterLoads may be nil; when provided it should
// contain the number of active workloads assigned per cluster.
func ChooseCluster(cands []Candidate, pd PolicyDomain, reqFlavor string, clusterLoads map[string]int) (string, error) {
	filtered := make([]Candidate, 0, len(cands))
	for _, c := range cands {
		if !regionAllowed(pd.Regions, c.Region) {
			continue
		}
		if !providerAllowed(pd.Providers, c.Provider) {
			continue
		}
		if reqFlavor != "" && !c.Flavors[reqFlavor] {
			continue
		}
		filtered = append(filtered, c)
	}
	if len(filtered) == 0 {
		return "", fmt.Errorf("no eligible cluster for flavor %s", reqFlavor)
	}

	switch strings.ToLower(strings.TrimSpace(pd.Strategy)) {
	case "spread":
		return chooseSpread(filtered, clusterLoads), nil
	case "cost-aware":
		// Placeholder: default to lowest TTFG until cost hints are available.
		fallthrough
	default:
		return chooseLowestTTFG(filtered), nil
	}
}

func chooseLowestTTFG(cands []Candidate) string {
	best := ""
	bestTTFG := math.MaxFloat64
	for _, c := range cands {
		if c.TTFGSeconds < bestTTFG {
			bestTTFG = c.TTFGSeconds
			best = c.ClusterID
		}
	}
	if best == "" && len(cands) > 0 {
		return cands[0].ClusterID
	}
	return best
}

func chooseSpread(cands []Candidate, clusterLoads map[string]int) string {
	if len(cands) == 0 {
		return ""
	}
	best := cands[0].ClusterID
	bestLoad := loadForCluster(clusterLoads, best)
	bestTTFG := cands[0].TTFGSeconds
	for _, c := range cands[1:] {
		load := loadForCluster(clusterLoads, c.ClusterID)
		if load < bestLoad {
			best = c.ClusterID
			bestLoad = load
			bestTTFG = c.TTFGSeconds
			continue
		}
		if load == bestLoad && c.TTFGSeconds < bestTTFG {
			best = c.ClusterID
			bestTTFG = c.TTFGSeconds
		}
	}
	return best
}

func loadForCluster(loads map[string]int, id string) int {
	if loads == nil {
		return 0
	}
	if v, ok := loads[id]; ok {
		return v
	}
	return 0
}

func regionAllowed(allowed []string, region string) bool {
	if len(allowed) == 0 {
		return true
	}
	region = strings.ToLower(strings.TrimSpace(region))
	for _, r := range allowed {
		if region == strings.ToLower(strings.TrimSpace(r)) {
			return true
		}
	}
	return false
}

func providerAllowed(allowed []string, provider string) bool {
	if len(allowed) == 0 {
		return true
	}
	provider = strings.ToLower(strings.TrimSpace(provider))
	for _, p := range allowed {
		if provider == strings.ToLower(strings.TrimSpace(p)) {
			return true
		}
	}
	return false
}

// ClusterSnapshot is kept for backwards compatibility with existing callers.
type ClusterSnapshot interface {
	CandidatesForWorkload(*aegis.Workload) []Candidate
}
