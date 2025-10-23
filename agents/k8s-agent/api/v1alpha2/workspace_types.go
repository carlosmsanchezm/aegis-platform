/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha2

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkspacePhase captures the coarse lifecycle for Workspace resources.
type WorkspacePhase string

const (
	// WorkspacePhasePending indicates the Workspace is being prepared.
	WorkspacePhasePending WorkspacePhase = "Pending"
	// WorkspacePhaseSubmitted indicates execution resources have been created.
	WorkspacePhaseSubmitted WorkspacePhase = "Submitted"
	// WorkspacePhaseAdmitted indicates the Workspace has been admitted for execution.
	WorkspacePhaseAdmitted WorkspacePhase = "Admitted"
	// WorkspacePhaseRunning indicates the associated execution is active.
	WorkspacePhaseRunning WorkspacePhase = "Running"
	// WorkspacePhaseSucceeded indicates execution completed successfully.
	WorkspacePhaseSucceeded WorkspacePhase = "Succeeded"
	// WorkspacePhaseFailed indicates execution terminated unsuccessfully.
	WorkspacePhaseFailed WorkspacePhase = "Failed"
)

// WorkspaceSpec defines the desired state of a Workspace.
type WorkspaceSpec struct {
	// ProjectRef ties the Workspace to an Aegis project identifier.
	// +kubebuilder:validation:Required
	ProjectRef string `json:"projectRef"`

	// Queue optionally specifies the scheduling queue.
	// +kubebuilder:validation:Optional
	Queue string `json:"queue,omitempty"`

	// Persona captures persona metadata that may drive defaults.
	// +kubebuilder:validation:Optional
	Persona string `json:"persona,omitempty"`

	// ProfileRef references a WorkspaceClass providing defaults and admission rules.
	// +kubebuilder:validation:Required
	ProfileRef string `json:"profileRef"`

	// BackendRef points to a provider-specific backend configuration.
	// +kubebuilder:validation:Optional
	BackendRef *corev1.TypedLocalObjectReference `json:"backendRef,omitempty"`

	// GPUProfile describes GPU flavor selection and hints for execution.
	// +kubebuilder:validation:Optional
	GPUProfile *WorkspaceGPUProfileSpec `json:"gpuProfile,omitempty"`

	// Execution specifies the container image, command, and runtime parameters.
	// +kubebuilder:validation:Optional
	Execution *WorkspaceExecution `json:"execution,omitempty"`

	// Network captures desired isolation and egress guardrails.
	// +kubebuilder:validation:Optional
	Network *WorkspaceNetworkSpec `json:"network,omitempty"`

	// Storage describes persistence requirements for the workspace.
	// +kubebuilder:validation:Optional
	Storage *WorkspaceStorageSpec `json:"storage,omitempty"`

	// MaxDurationSeconds defines a hard cap on runtime duration.
	// +kubebuilder:validation:Optional
	MaxDurationSeconds *int64 `json:"maxDurationSeconds,omitempty"`

	// TTLSecondsAfterFinished defines when to garbage collect execution artifacts.
	// +kubebuilder:validation:Optional
	TTLSecondsAfterFinished *int32 `json:"ttlSecondsAfterFinished,omitempty"`
}

// WorkspaceExecution captures runtime intent for the interactive workload.
type WorkspaceExecution struct {
	// Image is the container image to launch for the workspace.
	// +kubebuilder:validation:Optional
	Image string `json:"image,omitempty"`

	// Env defines environment variables for the workspace container.
	// +kubebuilder:validation:Optional
	Env map[string]string `json:"env,omitempty"`

	// Command overrides the container entrypoint.
	// +kubebuilder:validation:Optional
	Command []string `json:"command,omitempty"`

	// Interactive indicates whether SSH and interactive services should be exposed.
	// +kubebuilder:validation:Optional
	Interactive bool `json:"interactive,omitempty"`

	// Ports enumerates additional service ports to expose.
	// +kubebuilder:validation:Optional
	Ports []int32 `json:"ports,omitempty"`
}

// WorkspaceGPUProfileSpec declares desired GPU flavor and optional hints.
type WorkspaceGPUProfileSpec struct {
	// Flavor names the GPU flavor or profile.
	// +kubebuilder:validation:Optional
	Flavor string `json:"flavor,omitempty"`

	// Hints provide supplemental resource hints derived from the control plane.
	// +kubebuilder:validation:Optional
	Hints *ResourceHints `json:"hints,omitempty"`
}

// ResourceHints describes control plane derived resource hints.
type ResourceHints struct {
	// ResourceName corresponds to the GPU resource alias (e.g. nvidia.com/mig-1g.10gb).
	// +kubebuilder:validation:Optional
	ResourceName string `json:"resourceName,omitempty"`

	// GPUCount expresses the number of GPUs to request.
	// +kubebuilder:validation:Optional
	GPUCount int32 `json:"gpuCount,omitempty"`

	// CpuCoresRequest expresses CPU request quantity (e.g. "2", "500m").
	// +kubebuilder:validation:Optional
	CpuCoresRequest *string `json:"cpuCoresRequest,omitempty"`

	// MemoryRequest expresses memory request quantity (e.g. "16Gi").
	// +kubebuilder:validation:Optional
	MemoryRequest *string `json:"memoryRequest,omitempty"`
}

// WorkspaceNetworkSpec models desired network settings for a Workspace.
type WorkspaceNetworkSpec struct {
	// IsolationLevel specifies the isolation class (e.g. IL2, IL4).
	// +kubebuilder:validation:Optional
	IsolationLevel string `json:"isolationLevel,omitempty"`

	// Egress defines default egress behavior (Allow, DenyByDefault, etc).
	// +kubebuilder:validation:Optional
	Egress string `json:"egress,omitempty"`

	// AdditionalCIDRs enumerates CIDRs that should be reachable from the workspace.
	// +kubebuilder:validation:Optional
	AdditionalCIDRs []string `json:"additionalCIDRs,omitempty"`
}

// WorkspaceStorageSpec captures persistence configuration.
type WorkspaceStorageSpec struct {
	// Mode identifies persistent or ephemeral storage.
	// +kubebuilder:validation:Optional
	Mode string `json:"mode,omitempty"`

	// PVCTemplate provides a template used to create a PersistentVolumeClaim.
	// +kubebuilder:validation:Optional
	PVCTemplate *corev1.PersistentVolumeClaimSpec `json:"pvcTemplate,omitempty"`
}

// ConnectionInfo captures connection material surfaced to the user.
type ConnectionInfo struct {
	// ProxyURL exposes the reverse proxy endpoint for browser access.
	// +kubebuilder:validation:Optional
	ProxyURL string `json:"proxyURL,omitempty"`

	// SSHHostAlias captures the SSH host alias configured for bootstrap.
	// +kubebuilder:validation:Optional
	SSHHostAlias string `json:"sshHostAlias,omitempty"`
}

// CostStatus captures cost estimation and accrual details.
type CostStatus struct {
	// AccruedUSD is the total accrued cost in USD.
	// +kubebuilder:validation:Optional
	AccruedUSD string `json:"accruedUSD,omitempty"`

	// EstHourlyRateUSD captures the estimated hourly burn rate in USD.
	// +kubebuilder:validation:Optional
	EstHourlyRateUSD string `json:"estHourlyRateUSD,omitempty"`
}

// WorkspaceStatus defines the observed state of a Workspace.
type WorkspaceStatus struct {
	// Phase reflects the coarse lifecycle state rolled up from owned resources.
	// +kubebuilder:validation:Optional
	Phase WorkspacePhase `json:"phase,omitempty"`

	// Backend records the execution backend reported by the active workload.
	// +kubebuilder:validation:Optional
	Backend string `json:"backend,omitempty"`

	// URL references the execution object or interactive endpoint.
	// +kubebuilder:validation:Optional
	URL string `json:"url,omitempty"`

	// Cost surfaces estimation and accrual signals.
	// +kubebuilder:validation:Optional
	Cost *CostStatus `json:"cost,omitempty"`

	// Connection contains access material for interactive experiences.
	// +kubebuilder:validation:Optional
	Connection *ConnectionInfo `json:"connection,omitempty"`

	// Conditions provides granular readiness and policy state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// WorkloadRef references the current realization (e.g. AegisWorkload).
	// +kubebuilder:validation:Optional
	WorkloadRef *corev1.ObjectReference `json:"workloadRef,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=ws
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Queue",type=string,JSONPath=`.spec.queue`,priority=1
// +kubebuilder:printcolumn:name="Backend",type=string,JSONPath=`.status.backend`,priority=1
// Workspace is the Schema for the workspaces API.
type Workspace struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceSpec   `json:"spec,omitempty"`
	Status WorkspaceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceList contains a list of Workspace.
type WorkspaceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Workspace `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Workspace{}, &WorkspaceList{})
}
