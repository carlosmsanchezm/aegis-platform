package placement

import "testing"

func TestChooseClusterPrefersLowestTTFG(t *testing.T) {
	candidates := []Candidate{
		{
			ClusterID:   "slow-cluster",
			Region:      "us-central",
			TTFGSeconds: 50,
			Flavors: map[string]bool{
				"a10-mig-1g": true,
			},
		},
		{
			ClusterID:   "fast-cluster",
			Region:      "us-central",
			TTFGSeconds: 10,
			Flavors: map[string]bool{
				"a10-mig-1g": true,
			},
		},
	}

	id, err := ChooseCluster(candidates, PolicyDomain{Regions: []string{"us-central"}}, "a10-mig-1g")
	if err != nil {
		t.Fatalf("ChooseCluster returned error: %v", err)
	}
	if want := "fast-cluster"; id != want {
		t.Fatalf("expected %q, got %q", want, id)
	}
}

func TestChooseClusterRespectsPolicyRegions(t *testing.T) {
	candidates := []Candidate{
		{
			ClusterID:   "blocked",
			Region:      "us-west",
			TTFGSeconds: 5,
			Flavors: map[string]bool{
				"a10-mig-1g": true,
			},
		},
		{
			ClusterID:   "allowed",
			Region:      "us-east",
			TTFGSeconds: 15,
			Flavors: map[string]bool{
				"a10-mig-1g": true,
			},
		},
	}

	id, err := ChooseCluster(candidates, PolicyDomain{Regions: []string{"us-east"}}, "a10-mig-1g")
	if err != nil {
		t.Fatalf("ChooseCluster returned error: %v", err)
	}
	if want := "allowed"; id != want {
		t.Fatalf("expected %q, got %q", want, id)
	}
}

func TestChooseClusterErrorsWhenNoEligibleCluster(t *testing.T) {
	candidates := []Candidate{
		{
			ClusterID:   "no-flavor",
			Region:      "us-east",
			TTFGSeconds: 42,
			Flavors:     map[string]bool{},
		},
	}

	if _, err := ChooseCluster(candidates, PolicyDomain{Regions: []string{"us-east"}}, "a10-mig-1g"); err == nil {
		t.Fatalf("expected ChooseCluster to return error when no candidates match")
	}
}
