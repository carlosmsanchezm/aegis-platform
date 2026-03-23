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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// WorkspaceStorageSpec configures persistent storage for workspace pods.
// Data persists across workspace restarts within the same project.
type WorkspaceStorageSpec struct {
	// Persistent enables PVC-backed storage. When false, workspace data is ephemeral.
	// +kubebuilder:validation:Optional
	Persistent bool `json:"persistent,omitempty"`

	// StorageClass is the Kubernetes StorageClass name (e.g. "gp3"). Empty uses the cluster default.
	// +kubebuilder:validation:Optional
	StorageClass string `json:"storageClass,omitempty"`

	// Size is the storage capacity as a Kubernetes quantity (e.g. "50Gi").
	// +kubebuilder:validation:Optional
	Size string `json:"size,omitempty"`

	// MountPath is the path inside the container where storage is mounted. Default: "/home/coder".
	// +kubebuilder:validation:Optional
	MountPath string `json:"mountPath,omitempty"`

	// ExistingClaimName reuses an existing PVC by name for session continuity.
	// +kubebuilder:validation:Optional
	ExistingClaimName string `json:"existingClaimName,omitempty"`
}

// WorkspaceSpec defines properties for interactive workloads.
type WorkspaceSpec struct {
	// Flavor is the GPU flavor requested for the workspace pods.
	// +kubebuilder:validation:Optional
	Flavor string `json:"flavor,omitempty"`

	// Image contains the container image to execute.
	// +kubebuilder:validation:Optional
	Image string `json:"image,omitempty"`

	// Env holds environment variables to project into the container.
	// +kubebuilder:validation:Optional
	Env map[string]string `json:"env,omitempty"`

	// Command overrides the default entrypoint.
	// +kubebuilder:validation:Optional
	Command []string `json:"command,omitempty"`

	// Interactive indicates the workspace expects interactive access (SSH, VS Code, etc).
	// +kubebuilder:validation:Optional
	Interactive bool `json:"interactive,omitempty"`

	// Ports exposes additional container ports for interactive scenarios (defaults to 11111 when empty).
	// +kubebuilder:validation:Optional
	Ports []int32 `json:"ports,omitempty"`

	// Storage configures persistent storage for the workspace. When nil, platform defaults apply.
	// +kubebuilder:validation:Optional
	Storage *WorkspaceStorageSpec `json:"storage,omitempty"`
}

// TrainingSpec captures distributed training configuration details.
type TrainingSpec struct {
	// Flavor identifies the GPU flavor for each training worker.
	// +kubebuilder:validation:Optional
	Flavor string `json:"flavor,omitempty"`

	// Workers represents the number of replicas participating in the training job.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	Workers int32 `json:"workers,omitempty"`

	// GpusPerWorker specifies GPUs to request per replica.
	// +kubebuilder:validation:Optional
	// +kubebuilder:validation:Minimum=1
	GpusPerWorker int32 `json:"gpusPerWorker,omitempty"`

	// Image used for the training containers.
	// +kubebuilder:validation:Optional
	Image string `json:"image,omitempty"`

	// Command overrides the default entrypoint for training containers.
	// +kubebuilder:validation:Optional
	Command []string `json:"command,omitempty"`

	// Gang requests gang scheduling semantics when supported.
	// +kubebuilder:validation:Optional
	Gang bool `json:"gang,omitempty"`
}

// ResourceHints reflects control-plane scheduling hints.
type ResourceHints struct {
	// ResourceName is the GPU resource alias to request (e.g. nvidia.com/mig-1g.10gb).
	// +kubebuilder:validation:Optional
	ResourceName string `json:"resourceName,omitempty"`

	// GpuCount is the number of GPUs to request per container.
	// +kubebuilder:validation:Optional
	GpuCount int32 `json:"gpuCount,omitempty"`

	// CpuCoresRequest is the CPU core quantity to request (e.g. "1", "500m").
	// +kubebuilder:validation:Optional
	CpuCoresRequest *string `json:"cpuCoresRequest,omitempty"`

	// MemoryRequest is the memory quantity to request (e.g. "4Gi", "1024Mi").
	// +kubebuilder:validation:Optional
	MemoryRequest *string `json:"memoryRequest,omitempty"`
}

// AegisWorkloadSpec defines the desired state of AegisWorkload.
// +kubebuilder:validation:XValidation:rule="has(self.workspace) || has(self.training)",message="exactly one of workspace or training must be specified"
// +kubebuilder:validation:XValidation:rule="!(has(self.workspace) && has(self.training))",message="exactly one of workspace or training must be specified"
type AegisWorkloadSpec struct {
	// ProjectID associates the workload with an Aegis project.
	// +kubebuilder:validation:Required
	ProjectID string `json:"projectId"`

	// Queue optionally specifies the scheduling queue.
	// +kubebuilder:validation:Optional
	Queue string `json:"queue,omitempty"`

	// Workspace holds the interactive workload specification.
	// +kubebuilder:validation:Optional
	Workspace *WorkspaceSpec `json:"workspace,omitempty"`

	// Training holds the distributed training specification.
	// +kubebuilder:validation:Optional
	Training *TrainingSpec `json:"training,omitempty"`

	// Hints carries scheduling hints propagated from the control plane.
	// +kubebuilder:validation:Optional
	Hints *ResourceHints `json:"hints,omitempty"`
}

// JobRef references the workload realization created by the operator.
type JobRef struct {
	APIVersion string `json:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Name       string `json:"name,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
}

// AegisWorkloadPhase describes the high-level lifecycle state of a workload.
type AegisWorkloadPhase string

const (
	// PhasePending denotes that the workload has been accepted but not yet submitted to Kubernetes.
	PhasePending AegisWorkloadPhase = "Pending"
	// PhaseSubmitted indicates the operator has created the concrete Job/CRD.
	PhaseSubmitted AegisWorkloadPhase = "Submitted"
	// PhaseAdmitted reflects external admission (e.g. Kueue) allowing execution.
	PhaseAdmitted AegisWorkloadPhase = "Admitted"
	// PhaseRunning represents actively executing workloads.
	PhaseRunning AegisWorkloadPhase = "Running"
	// PhaseSuspended marks workloads that have been policy-suspended (idle timeout, admin action, etc.).
	PhaseSuspended AegisWorkloadPhase = "Suspended"
	// PhaseSucceeded marks workloads that completed successfully.
	PhaseSucceeded AegisWorkloadPhase = "Succeeded"
	// PhaseFailed marks workloads that terminated unsuccessfully.
	PhaseFailed AegisWorkloadPhase = "Failed"
	// PhaseTerminated marks workloads that were intentionally terminated (e.g. user request).
	PhaseTerminated AegisWorkloadPhase = "Terminated"
)

// AegisWorkloadStatus defines the observed state of AegisWorkload.
type AegisWorkloadStatus struct {
	// Phase is the coarse-grained lifecycle state.
	// +kubebuilder:validation:Optional
	Phase AegisWorkloadPhase `json:"phase,omitempty"`

	// Message provides additional diagnostic information.
	// +kubebuilder:validation:Optional
	Message string `json:"message,omitempty"`

	// Backend records which executor handled the workload (workspace, pytorch, etc).
	// +kubebuilder:validation:Optional
	Backend string `json:"backend,omitempty"`

	// URL references the running workload (e.g. k8s:// namespace/object).
	// +kubebuilder:validation:Optional
	URL string `json:"url,omitempty"`

	// JobRef captures the Kubernetes object created by the operator.
	// +kubebuilder:validation:Optional
	JobRef *JobRef `json:"jobRef,omitempty"`

	// StartTime indicates when execution began.
	// +kubebuilder:validation:Optional
	StartTime *metav1.Time `json:"startTime,omitempty"`

	// CompletionTime indicates when execution finished.
	// +kubebuilder:validation:Optional
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`

	// Conditions provide detailed status information.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=awl
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Backend",type=string,JSONPath=`.status.backend`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// AegisWorkload is the Schema for the aegisworkloads API.
type AegisWorkload struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AegisWorkloadSpec   `json:"spec,omitempty"`
	Status AegisWorkloadStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AegisWorkloadList contains a list of AegisWorkload.
type AegisWorkloadList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AegisWorkload `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AegisWorkload{}, &AegisWorkloadList{})
}
