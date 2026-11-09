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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	aegisv1alpha1 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha1"
)

// AegisWorkloadSliceItem defines a single workload template entry.
type AegisWorkloadSliceItem struct {
	// Template references the AegisWorkload specification to instantiate.
	// +kubebuilder:validation:Required
	Template aegisv1alpha1.AegisWorkloadSpec `json:"template"`
}

// AegisWorkloadSliceSpec defines the desired state for a slice of workloads.
type AegisWorkloadSliceSpec struct {
	// Items enumerates templates for the child AegisWorkloads.
	// +kubebuilder:validation:MinItems=1
	Items []AegisWorkloadSliceItem `json:"items"`
}

// AegisWorkloadSliceStatus captures observed execution state for the slice.
type AegisWorkloadSliceStatus struct {
	// ReadyCount captures how many child workloads report readiness.
	// +kubebuilder:validation:Optional
	ReadyCount int32 `json:"readyCount,omitempty"`

	// Conditions surfaces aggregated status from child workloads.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=awls
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyCount`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// AegisWorkloadSlice represents a collection of AegisWorkload templates managed together.
type AegisWorkloadSlice struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AegisWorkloadSliceSpec   `json:"spec,omitempty"`
	Status AegisWorkloadSliceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AegisWorkloadSliceList contains a list of AegisWorkloadSlice.
type AegisWorkloadSliceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AegisWorkloadSlice `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AegisWorkloadSlice{}, &AegisWorkloadSliceList{})
}
