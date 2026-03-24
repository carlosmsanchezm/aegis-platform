package store

import "testing"

func TestDeriveProjectIDFromClusterID(t *testing.T) {
	tests := []struct {
		name      string
		clusterID string
		want      string
	}{
		{"us-east-1 standard", "db-1-us-east-1-atlas-train-govcloud", "db-1"},
		{"us-west-2 standard", "myproj-us-west-2-worker-abc", "myproj"},
		{"eu-west-1", "team-alpha-eu-west-1-gpu-cluster", "team-alpha"},
		{"eu-central-1", "prod-eu-central-1-main", "prod"},
		{"ap-southeast-1", "asia-team-ap-southeast-1-train", "asia-team"},
		{"ap-northeast-1", "jp-cluster-ap-northeast-1-gpu", "jp-cluster"},
		{"ap-south-1", "india-ap-south-1-test", "india"},
		{"sa-east-1", "latam-sa-east-1-dev", "latam"},
		{"ca-central-1", "canada-ca-central-1-prod", "canada"},
		{"me-south-1", "mideast-me-south-1-staging", "mideast"},
		{"af-south-1", "africa-af-south-1-test", "africa"},
		{"il-central-1", "israel-il-central-1-main", "israel"},
		{"docker-desktop no match", "docker-desktop", ""},
		{"empty string", "", ""},
		{"no region pattern", "local-cluster", ""},
		{"single word", "standalone", ""},
		{"whitespace padded", "  db-1-us-east-1-foo  ", "db-1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DeriveProjectIDFromClusterID(tt.clusterID)
			if got != tt.want {
				t.Errorf("DeriveProjectIDFromClusterID(%q) = %q, want %q", tt.clusterID, got, tt.want)
			}
		})
	}
}
