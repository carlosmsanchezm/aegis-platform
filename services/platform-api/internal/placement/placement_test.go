package placement

import "testing"

func TestChooseClusterPrefersLowestTTFG(t *testing.T) {
	candidates := []Candidate{
		{
			ClusterID:   "slow-cluster",
			Provider:    "aws",
			Region:      "us-central",
			TTFGSeconds: 50,
			Flavors: map[string]bool{
				"a10-mig-1g": true,
			},
		},
		{
			ClusterID:   "fast-cluster",
			Provider:    "aws",
			Region:      "us-central",
			TTFGSeconds: 10,
			Flavors: map[string]bool{
				"a10-mig-1g": true,
			},
		},
	}

	id, err := ChooseCluster(candidates, PolicyDomain{Regions: []string{"us-central"}}, "a10-mig-1g", nil)
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
			Provider:    "aws",
			Region:      "us-west",
			TTFGSeconds: 5,
			Flavors: map[string]bool{
				"a10-mig-1g": true,
			},
		},
		{
			ClusterID:   "allowed",
			Provider:    "aws",
			Region:      "us-east",
			TTFGSeconds: 15,
			Flavors: map[string]bool{
				"a10-mig-1g": true,
			},
		},
	}

	id, err := ChooseCluster(candidates, PolicyDomain{Regions: []string{"us-east"}}, "a10-mig-1g", nil)
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
			Provider:    "aws",
			Region:      "us-east",
			TTFGSeconds: 42,
			Flavors:     map[string]bool{},
		},
	}

	if _, err := ChooseCluster(candidates, PolicyDomain{Regions: []string{"us-east"}}, "a10-mig-1g", nil); err == nil {
		t.Fatalf("expected ChooseCluster to return error when no candidates match")
	}
}

func TestChooseClusterFiltersProviders(t *testing.T) {
	candidates := []Candidate{
		{ClusterID: "p1", Provider: "aws", Region: "us-east", TTFGSeconds: 5, Flavors: map[string]bool{"a10": true}},
		{ClusterID: "p2", Provider: "gcp", Region: "us-east", TTFGSeconds: 1, Flavors: map[string]bool{"a10": true}},
	}
	pd := PolicyDomain{Regions: []string{"us-east"}, Providers: []string{"aws"}}
	id, err := ChooseCluster(candidates, pd, "a10", nil)
	if err != nil {
		t.Fatalf("ChooseCluster returned error: %v", err)
	}
	if id != "p1" {
		t.Fatalf("expected provider filter to select aws cluster, got %s", id)
	}
}

func TestChooseClusterSpreadStrategy(t *testing.T) {
	candidates := []Candidate{
		{ClusterID: "a", Provider: "aws", Region: "us-east", TTFGSeconds: 50, Flavors: map[string]bool{"a10": true}},
		{ClusterID: "b", Provider: "aws", Region: "us-east", TTFGSeconds: 20, Flavors: map[string]bool{"a10": true}},
	}
	loads := map[string]int{"a": 5, "b": 1}
	pd := PolicyDomain{Regions: []string{"us-east"}, Providers: []string{"aws"}, Strategy: "spread"}
	id, err := ChooseCluster(candidates, pd, "a10", loads)
	if err != nil {
		t.Fatalf("ChooseCluster returned error: %v", err)
	}
	if id != "b" {
		t.Fatalf("expected spread strategy to favor lowest load cluster, got %s", id)
	}
}

func TestChooseClusterRegionAndProviderCaseInsensitive(t *testing.T) {
	candidates := []Candidate{
		{ClusterID: "c1", Provider: "AWS", Region: "US-EAST", TTFGSeconds: 5, Flavors: map[string]bool{"a10": true}},
	}
	pd := PolicyDomain{
		Regions:   []string{"us-east"},
		Providers: []string{"aws"},
	}
	id, err := ChooseCluster(candidates, pd, "a10", nil)
	if err != nil {
		t.Fatalf("ChooseCluster returned error: %v", err)
	}
	if id != "c1" {
		t.Fatalf("expected case-insensitive match, got %s", id)
	}
}
