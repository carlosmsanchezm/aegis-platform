package discovery

import aegis "github.com/yourorg/aegis/proto/aegis/v1"

func StaticFlavors() []*aegis.Flavor {
	return []*aegis.Flavor{
		{Name: "a10-mig-1g", Chip: "A10", MigProfile: "1g.10gb"},
		{Name: "a100-8x", Chip: "A100", MigProfile: "", RdmaRequired: true},
	}
}
