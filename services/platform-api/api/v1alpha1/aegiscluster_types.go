package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=aegisc

// AegisCluster mirrors the authoritative cluster state from the control plane store.
type AegisCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AegisClusterSpec   `json:"spec,omitempty"`
	Status AegisClusterStatus `json:"status,omitempty"`
}

// AegisClusterSpec holds immutable cluster metadata.
type AegisClusterSpec struct {
	ClusterID string `json:"clusterId"`
	ProjectID string `json:"projectId"`
	Provider  string `json:"provider"`
	Region    string `json:"region"`
}

// AegisClusterStatus reflects runtime signal sourced from the database.
type AegisClusterStatus struct {
	Phase         string             `json:"phase,omitempty"`
	Flavors       []string           `json:"flavors,omitempty"`
	Capacity      map[string]string  `json:"capacity,omitempty"`
	Conditions    []metav1.Condition `json:"conditions,omitempty"`
	LastHeartbeat *metav1.Time       `json:"lastHeartbeat,omitempty"`
}

// +kubebuilder:object:root=true

// AegisClusterList contains a list of AegisCluster objects.
type AegisClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AegisCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AegisCluster{}, &AegisClusterList{})
}

func (in *AegisCluster) DeepCopyInto(out *AegisCluster) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

func (in *AegisCluster) DeepCopy() *AegisCluster {
	if in == nil {
		return nil
	}
	out := new(AegisCluster)
	in.DeepCopyInto(out)
	return out
}

func (in *AegisCluster) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *AegisClusterStatus) DeepCopyInto(out *AegisClusterStatus) {
	*out = *in
	if in.Flavors != nil {
		out.Flavors = make([]string, len(in.Flavors))
		copy(out.Flavors, in.Flavors)
	}
	if in.Capacity != nil {
		out.Capacity = make(map[string]string, len(in.Capacity))
		for k, v := range in.Capacity {
			out.Capacity[k] = v
		}
	}
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
	if in.LastHeartbeat != nil {
		out.LastHeartbeat = in.LastHeartbeat.DeepCopy()
	}
}

func (in *AegisClusterStatus) DeepCopy() *AegisClusterStatus {
	if in == nil {
		return nil
	}
	out := new(AegisClusterStatus)
	in.DeepCopyInto(out)
	return out
}

func (in *AegisClusterList) DeepCopyInto(out *AegisClusterList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]AegisCluster, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *AegisClusterList) DeepCopy() *AegisClusterList {
	if in == nil {
		return nil
	}
	out := new(AegisClusterList)
	in.DeepCopyInto(out)
	return out
}

func (in *AegisClusterList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
