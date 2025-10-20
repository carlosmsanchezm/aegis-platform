package catalog

// TemplateExecution captures execution defaults for a workspace template.
type TemplateExecution struct {
	Image       string
	Ports       []int32
	Command     []string
	Env         map[string]string
	Interactive bool
}

// Template describes workspace defaults for a given profile/persona.
type Template struct {
	ID             string
	WorkspaceTypes []string
	Flavor         string
	Queue          string
	Execution      TemplateExecution
	AllowOverride  bool
}

var templates = map[string]Template{
	"vscode-python": {
		ID:             "vscode-python",
		WorkspaceTypes: []string{"vscode"},
		Flavor:         "cpu-medium",
		Execution: TemplateExecution{
			Image:       "carlosmsanchez/aegis-workspace-vscode:latest",
			Ports:       []int32{22, 11111},
			Interactive: true,
		},
	},
	"vscode-data": {
		ID:             "vscode-data",
		WorkspaceTypes: []string{"vscode"},
		Flavor:         "cpu-large",
		Execution: TemplateExecution{
			Image:       "ghcr.io/aegis/workspace-vscode-data:latest",
			Ports:       []int32{22, 11111},
			Interactive: true,
		},
	},
	"jupyter-pytorch": {
		ID:             "jupyter-pytorch",
		WorkspaceTypes: []string{"jupyter"},
		Flavor:         "gpu-standard",
		Queue:          "gpu",
		Execution: TemplateExecution{
			Image:       "ghcr.io/aegis/workspace-jupyter-pytorch:latest",
			Ports:       []int32{22, 8888},
			Env:         map[string]string{"NOTEBOOK_TOKEN": "aegis"},
			Interactive: true,
		},
	},
	"jupyter-rapids": {
		ID:             "jupyter-rapids",
		WorkspaceTypes: []string{"jupyter"},
		Flavor:         "gpu-large",
		Queue:          "gpu",
		Execution: TemplateExecution{
			Image:       "ghcr.io/aegis/workspace-jupyter-rapids:latest",
			Ports:       []int32{22, 8888},
			Interactive: true,
		},
	},
	"cli-ops": {
		ID:             "cli-ops",
		WorkspaceTypes: []string{"cli"},
		Flavor:         "cpu-small",
		Execution: TemplateExecution{
			Image:       "ghcr.io/aegis/workspace-cli:latest",
			Ports:       []int32{22},
			Interactive: true,
		},
	},
	"custom": {
		ID:            "custom",
		AllowOverride: true,
		Execution: TemplateExecution{
			Ports:       []int32{22},
			Interactive: true,
		},
	},
}

// LookupTemplate fetches a template definition by identifier.
func LookupTemplate(id string) (Template, bool) {
	if tmpl, ok := templates[id]; ok {
		return tmpl, true
	}
	return Template{}, false
}
