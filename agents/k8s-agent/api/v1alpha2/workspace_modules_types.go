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

// WorkspaceClusterSpec describes the target cluster provisioning semantics.
type WorkspaceClusterSpec struct {
	// Provider identifies the infrastructure provider (e.g. aws, gcp, azure, onprem).
	// +kubebuilder:validation:Required
	Provider string `json:"provider"`

	// ClassRef references a provider-specific class for dynamic provisioning.
	// +kubebuilder:validation:Optional
	ClassRef *corev1.TypedLocalObjectReference `json:"classRef,omitempty"`

	// ClusterRef references an existing Kubernetes cluster handle.
	// +kubebuilder:validation:Optional
	ClusterRef *corev1.ObjectReference `json:"clusterRef,omitempty"`

	// NodePools describes desired node pools keyed by purpose (gpu/system/etc).
	// +kubebuilder:validation:Optional
	NodePools map[string]WorkspaceNodePool `json:"nodePools,omitempty"`
}

// WorkspaceNodePool configures a logical node pool used by the Workspace.
type WorkspaceNodePool struct {
	// Purpose communicates the intent for this pool (gpu/system/ingress/etc).
	// +kubebuilder:validation:Optional
	Purpose string `json:"purpose,omitempty"`

	// Labels applies Kubernetes node labels to the provisioned pool.
	// +kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`

	// Taints applies Kubernetes taints to the provisioned pool.
	// +kubebuilder:validation:Optional
	Taints []corev1.Taint `json:"taints,omitempty"`
}

// WorkspaceClusterStatus communicates cluster readiness details.
type WorkspaceClusterStatus struct {
	// Endpoint is the Kubernetes API endpoint for the provisioned cluster.
	// +kubebuilder:validation:Optional
	Endpoint string `json:"endpoint,omitempty"`

	// KubeconfigSecretRef references a secret containing kubeconfig material.
	// +kubebuilder:validation:Optional
	KubeconfigSecretRef *corev1.SecretReference `json:"kubeconfigSecretRef,omitempty"`

	// Conditions reports readiness and provisioning state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// WorkspaceCluster orchestrates cluster provisioning for a Workspace.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=wcluster
// +kubebuilder:printcolumn:name="Provider",type=string,JSONPath=`.spec.provider`
// +kubebuilder:printcolumn:name="Endpoint",type=string,JSONPath=`.status.endpoint`,priority=1
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type WorkspaceCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceClusterSpec   `json:"spec,omitempty"`
	Status WorkspaceClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceClusterList contains a list of WorkspaceCluster.
type WorkspaceClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceCluster `json:"items"`
}

// WorkspaceNetworkStatus communicates network provisioning status.
type WorkspaceNetworkStatus struct {
	// NetworkRef references the concrete network or namespace.
	// +kubebuilder:validation:Optional
	NetworkRef *corev1.ObjectReference `json:"networkRef,omitempty"`

	// Conditions reports readiness and policy states.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// WorkspaceNetwork defines network isolation settings for a Workspace.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=wnet
// +kubebuilder:printcolumn:name="Isolation",type=string,JSONPath=`.spec.isolationLevel`
// +kubebuilder:printcolumn:name="Egress",type=string,JSONPath=`.spec.egress`,priority=1
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type WorkspaceNetwork struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceNetworkSpec   `json:"spec,omitempty"`
	Status WorkspaceNetworkStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceNetworkList contains a list of WorkspaceNetwork.
type WorkspaceNetworkList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceNetwork `json:"items"`
}

// WorkspaceStorageStatus communicates storage readiness for the Workspace.
type WorkspaceStorageStatus struct {
	// ClaimRef references the created PersistentVolumeClaim.
	// +kubebuilder:validation:Optional
	ClaimRef *corev1.ObjectReference `json:"claimRef,omitempty"`

	// Conditions reports readiness and attachment state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// WorkspaceStorage defines storage provisioning requirements.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=wstorage
// +kubebuilder:printcolumn:name="Mode",type=string,JSONPath=`.spec.mode`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type WorkspaceStorage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceStorageSpec   `json:"spec,omitempty"`
	Status WorkspaceStorageStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceStorageList contains a list of WorkspaceStorage.
type WorkspaceStorageList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceStorage `json:"items"`
}

// WorkspaceGPUProfileStatus reports resolved GPU scheduling characteristics.
type WorkspaceGPUProfileStatus struct {
	// ResolvedFlavor is the concrete flavor admitted by the control plane/provider.
	// +kubebuilder:validation:Optional
	ResolvedFlavor string `json:"resolvedFlavor,omitempty"`

	// ResolvedHints contains the resolved resource hints.
	// +kubebuilder:validation:Optional
	ResolvedHints *ResourceHints `json:"resolvedHints,omitempty"`

	// Conditions reports readiness or policy state.
	// +kubebuilder:validation:Optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// WorkspaceGPUProfile models GPU intent as a modular CRD.
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=wgpu
// +kubebuilder:printcolumn:name="Flavor",type=string,JSONPath=`.spec.flavor`
// +kubebuilder:printcolumn:name="Ready",type=string,JSONPath=`.status.conditions[?(@.type=="Ready")].status`,priority=1
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type WorkspaceGPUProfile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WorkspaceGPUProfileSpec   `json:"spec,omitempty"`
	Status WorkspaceGPUProfileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// WorkspaceGPUProfileList contains a list of WorkspaceGPUProfile.
type WorkspaceGPUProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []WorkspaceGPUProfile `json:"items"`
}

func init() {
	SchemeBuilder.Register(
		&WorkspaceCluster{}, &WorkspaceClusterList{},
		&WorkspaceNetwork{}, &WorkspaceNetworkList{},
		&WorkspaceStorage{}, &WorkspaceStorageList{},
		&WorkspaceGPUProfile{}, &WorkspaceGPUProfileList{},
	)
}
