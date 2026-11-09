// Package plan converts Workspace intent into child execution resources.
package plan

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	aegisv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
	aegisv1alpha2 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha2"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/workload/builders"
	"github.com/yourorg/aegis/agents/k8s-agent/internal/workspace/catalog"
)

// Planner renders Workspace intent into an AegisWorkload child resource.
type Planner struct{}

// NewPlanner returns a Planner instance.
func NewPlanner() *Planner {
	return &Planner{}
}

// Result captures the computed realization for a Workspace.
type Result struct {
	// WorkloadName is the sanitized child AegisWorkload name.
	WorkloadName string
	// WorkloadSpec is the spec to apply to the child AegisWorkload.
	WorkloadSpec aegisv1alpha1.AegisWorkloadSpec
	// Labels to project onto the child workload.
	Labels map[string]string
	// Annotations to project onto the child workload.
	Annotations map[string]string
}

const (
	annotationMaxDurationSeconds   = "aegis.yourorg.dev/maxDurationSeconds"
	annotationTTLSecondsAfterClose = "aegis.yourorg.dev/ttlSecondsAfterFinished"
)

// Plan builds the workload realization for the provided Workspace.
func (p *Planner) Plan(ws *aegisv1alpha2.Workspace) (*Result, error) {
	if ws == nil {
		return nil, fmt.Errorf("workspace is nil")
	}
	if ws.Spec.ProjectRef == "" {
		return nil, fmt.Errorf("workspace %q missing spec.projectRef", ws.Name)
	}
	if ws.Spec.ProfileRef == "" {
		return nil, fmt.Errorf("workspace %q missing spec.profileRef", ws.Name)
	}

	template, ok := catalog.LookupTemplate(ws.Spec.ProfileRef)
	if !ok {
		return nil, fmt.Errorf("workspace %q references unknown profile %q", ws.Name, ws.Spec.ProfileRef)
	}

	exec := resolveExecution(template, ws.Spec.Execution)
	if exec == nil {
		return nil, fmt.Errorf("workspace %q requires execution settings (profile %q)", ws.Name, ws.Spec.ProfileRef)
	}

	gpuProfile := ws.Spec.GPUProfile
	if gpuProfile == nil && template.Flavor != "" {
		gpuProfile = &aegisv1alpha2.WorkspaceGPUProfileSpec{Flavor: template.Flavor}
	} else if gpuProfile != nil && gpuProfile.Flavor == "" && template.Flavor != "" {
		gpuProfile.Flavor = template.Flavor
	}

	queue := ws.Spec.Queue
	if queue == "" {
		queue = template.Queue
	}

	workloadSpec := aegisv1alpha1.AegisWorkloadSpec{
		ProjectID: ws.Spec.ProjectRef,
		Queue:     queue,
		Workspace: &aegisv1alpha1.WorkspaceSpec{
			Image:       exec.Image,
			Env:         exec.Env,
			Command:     exec.Command,
			Interactive: exec.Interactive,
			Ports:       exec.Ports,
		},
	}

	if gpuProfile != nil {
		workloadSpec.Workspace.Flavor = gpuProfile.Flavor
		if gpuProfile.Hints != nil {
			workloadSpec.Hints = &aegisv1alpha1.ResourceHints{
				ResourceName:    gpuProfile.Hints.ResourceName,
				GpuCount:        gpuProfile.Hints.GPUCount,
				CpuCoresRequest: gpuProfile.Hints.CpuCoresRequest,
				MemoryRequest:   gpuProfile.Hints.MemoryRequest,
			}
		}
	}

	workloadName := builders.SanitizeName("aegis-" + ws.Name)
	annotations := map[string]string{}

	if ws.Spec.MaxDurationSeconds != nil && *ws.Spec.MaxDurationSeconds > 0 {
		annotations[annotationMaxDurationSeconds] = fmt.Sprintf("%d", *ws.Spec.MaxDurationSeconds)
	}
	if ws.Spec.TTLSecondsAfterFinished != nil && *ws.Spec.TTLSecondsAfterFinished > 0 {
		annotations[annotationTTLSecondsAfterClose] = fmt.Sprintf("%d", *ws.Spec.TTLSecondsAfterFinished)
	}

	labels := map[string]string{
		"aegis.yourorg.dev/workspace": ws.Name,
	}

	return &Result{
		WorkloadName: workloadName,
		WorkloadSpec: workloadSpec,
		Labels:       labels,
		Annotations:  annotations,
	}, nil
}

// OwnerReference returns a controller owner reference for the Workspace.
func OwnerReference(ws *aegisv1alpha2.Workspace) metav1.OwnerReference {
	return metav1.OwnerReference{
		APIVersion: aegisv1alpha2.GroupVersion.String(),
		Kind:       "Workspace",
		Name:       ws.GetName(),
		UID:        ws.GetUID(),
	}
}

func resolveExecution(template catalog.Template, user *aegisv1alpha2.WorkspaceExecution) *aegisv1alpha1.WorkspaceSpec {
	baseEnv := map[string]string{}
	for k, v := range template.Execution.Env {
		baseEnv[k] = v
	}

	exec := &aegisv1alpha1.WorkspaceSpec{
		Image:       template.Execution.Image,
		Env:         baseEnv,
		Command:     append([]string(nil), template.Execution.Command...),
		Interactive: template.Execution.Interactive,
		Ports:       append([]int32(nil), template.Execution.Ports...),
	}

	if template.AllowOverride {
		if user != nil {
			if user.Image != "" {
				exec.Image = user.Image
			} else if exec.Image == "" {
				return nil
			}
			if len(user.Command) > 0 {
				exec.Command = append([]string(nil), user.Command...)
			}
			if user.Env != nil {
				if exec.Env == nil {
					exec.Env = map[string]string{}
				}
				for k, v := range user.Env {
					exec.Env[k] = v
				}
			}
			if user.Interactive {
				exec.Interactive = true
			}
			if len(user.Ports) > 0 {
				exec.Ports = append([]int32(nil), user.Ports...)
			} else if len(exec.Ports) == 0 {
				exec.Ports = []int32{22}
			}
		} else {
			if exec.Image == "" {
				return nil
			}
			if len(exec.Ports) == 0 {
				exec.Ports = []int32{22}
			}
		}
		return exec
	}

	if user != nil && user.Env != nil {
		if exec.Env == nil {
			exec.Env = map[string]string{}
		}
		for k, v := range user.Env {
			exec.Env[k] = v
		}
	}

	if exec.Image == "" {
		return nil
	}
	if len(exec.Ports) == 0 {
		exec.Ports = []int32{22}
	}

	return exec
}
