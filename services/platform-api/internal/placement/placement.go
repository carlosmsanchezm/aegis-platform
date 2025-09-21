package placement

import (
	"fmt"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

type Candidate struct {
	ClusterID   string
	Region      string
	Labels      map[string]string
	TTFGSeconds float64
	Flavors     map[string]bool // set for quick lookup
}

type PolicyDomain struct {
	Regions []string
}

func ChooseCluster(cands []Candidate, pd PolicyDomain, reqFlavor string) (string, error) {
	// MVP: filter by region + flavor, then pick lowest TTFG (or first)
	best := ""
	bestTTFG := 1e18
	regionOK := func(r string) bool {
		if len(pd.Regions) == 0 {
			return true
		}
		for _, x := range pd.Regions {
			if x == r {
				return true
			}
		}
		return false
	}
	for _, c := range cands {
		if !regionOK(c.Region) {
			continue
		}
		if !c.Flavors[reqFlavor] {
			continue
		}
		if c.TTFGSeconds < bestTTFG {
			bestTTFG = c.TTFGSeconds
			best = c.ClusterID
		}
	}
	if best == "" {
		return "", fmt.Errorf("no eligible cluster for flavor %s", reqFlavor)
	}
	return best, nil
}

type ClusterSnapshot interface {
	CandidatesForWorkload(*aegis.Workload) []Candidate
}
