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
)

// WorkspaceClassCompositionType enumerates supported composition engines.
type WorkspaceClassCompositionType string

const (
	// WorkspaceCompositionCrossplane indicates Crossplane-based composition.
	WorkspaceCompositionCrossplane WorkspaceClassCompositionType = "Crossplane"
	// WorkspaceCompositionTerraform indicates Terraform-based composition.
	WorkspaceCompositionTerraform WorkspaceClassCompositionType = "Terraform"
	// WorkspaceCompositionNative indicates native controller-based composition.
	WorkspaceCompositionNative WorkspaceClassCompositionType = "Native"
)

// WorkspaceClassSpec defines defaults and composition references for Workspaces.
type WorkspaceClassSpec struct {
	// Defaults capture default values applied to Workspaces referencing this class.
	// +kubebuilder:validation:Optional
	Defaults *WorkspaceClassDefaults `json:"defaults,omitempty"`

	// Composition selects the provisioning engine responsible for backends.
	// +kubebuilder:validation:Optional
	Composition *WorkspaceClassComposition `json:"composition,omitempty"`

	// Admission contains guardrails that gate queue and flavor usage.
	// +kubebuilder:validation:Optional
	Admission *WorkspaceClassAdmission `json:"admission,omitempty"`
}

// WorkspaceClassDefaults holds default workspace parameters.
type WorkspaceClassDefaults struct {
	// Execution contains default execution settings.
	// +kubebuilder:validation:Optional
	Execution *WorkspaceExecution `json:"execution,omitempty"`

	// Network contains default network settings.
	// +kubebuilder:validation:Optional
	Network *WorkspaceNetworkSpec `json:"network,omitempty"`

	// Storage contains default storage settings.
	// +kubebuilder:validation:Optional
	Storage *WorkspaceStorageSpec `json:"storage,omitempty"`

	// GPUProfile contains default GPU configuration.
	// +kubebuilder:validation:Optional
	GPUProfile *WorkspaceGPUProfileSpec `json:"gpuProfile,omitempty"`
}

// WorkspaceClassComposition references composition resources.
type WorkspaceClassComposition struct {
	// Type identifies the composition engine.
	// +kubebuilder:validation:Optional
	Type WorkspaceClassCompositionType `json:"type,omitempty"`

	// Ref references the specific composition resource (Composition, TerraformRun, etc).
	// +kubebuilder:validation:Optional
	Ref *CrossNamespaceObjectReference `json:"ref,omitempty"`
}

// WorkspaceClassAdmission defines queue and flavor guards.
type WorkspaceClassAdmission struct {
	// AllowedQueues enumerates queues permitted to reference this class.
	// +kubebuilder:validation:Optional
	AllowedQueues []string `json:"allowedQueues,omitempty"`

	// AllowedFlavors enumerates GPU flavors permitted when using this class.
	// +kubebuilder:validation:Optional
	AllowedFlavors []string `json:"allowedFlavors,omitempty"`
}

// CrossNamespaceObjectReference references namespaced resources across namespaces.
type CrossNamespaceObjectReference struct {
	// APIVersion of the target object.
	// +kubebuilder:validation:Optional
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind of the target object.
	// +kubebuilder:validation:Optional
	Kind string `json:"kind,omitempty"`

	// Name of the target object.
	// +kubebuilder:validation:Optional
	Name string `json:"name,omitempty"`

	// Namespace where the target object resides.
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`
}

// WorkspaceClassStatus captures applied composition state.
type WorkspaceClassStatus struct {
	// ObservedGeneration records the last reconciled spec generation.
	// +kubebuilder:validation:Optional
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions reports validation or provisioning state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,shortName=wclass
// +kubebuilder:printcolumn:name="Composition",type=string,JSONPath=`.spec.composition.type`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// WorkspaceClass defines reusable workspace compositions and defaults.
type WorkspaceClass struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceClassSpec   `json:"spec,omitempty"`
	Status WorkspaceClassStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceClassList contains a list of WorkspaceClass.
type WorkspaceClassList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceClass `json:"items"`
}

func init() {
	SchemeBuilder.Register(&WorkspaceClass{}, &WorkspaceClassList{})
}
