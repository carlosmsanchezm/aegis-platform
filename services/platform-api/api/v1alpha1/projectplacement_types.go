package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=pplace

// ProjectPlacement encodes placement policy and quotas for a single project.
type ProjectPlacement struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectPlacementSpec   `json:"spec,omitempty"`
	Status ProjectPlacementStatus `json:"status,omitempty"`
}

// ProjectPlacementSpec describes the desired placement policy the chooser should honour.
type ProjectPlacementSpec struct {
	ProjectID        string         `json:"projectId"`
	AllowedRegions   []string       `json:"allowedRegions,omitempty"`
	AllowedProviders []string       `json:"allowedProviders,omitempty"`
	DefaultFlavor    string         `json:"defaultFlavor,omitempty"`
	Quotas           map[string]int `json:"quotas,omitempty"`
	Strategy         string         `json:"strategy,omitempty"`
}

// ProjectPlacementStatus exposes the effective policy once validated by the controller.
type ProjectPlacementStatus struct {
	Effective  ProjectPlacementSpec `json:"effective,omitempty"`
	Conditions []metav1.Condition   `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// ProjectPlacementList contains a collection of ProjectPlacement objects.
type ProjectPlacementList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectPlacement `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProjectPlacement{}, &ProjectPlacementList{})
}

func (in *ProjectPlacement) DeepCopyInto(out *ProjectPlacement) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *ProjectPlacement) DeepCopy() *ProjectPlacement {
	if in == nil {
		return nil
	}
	out := new(ProjectPlacement)
	in.DeepCopyInto(out)
	return out
}

func (in *ProjectPlacement) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *ProjectPlacementSpec) DeepCopyInto(out *ProjectPlacementSpec) {
	*out = *in
	if in.AllowedRegions != nil {
		out.AllowedRegions = make([]string, len(in.AllowedRegions))
		copy(out.AllowedRegions, in.AllowedRegions)
	}
	if in.AllowedProviders != nil {
		out.AllowedProviders = make([]string, len(in.AllowedProviders))
		copy(out.AllowedProviders, in.AllowedProviders)
	}
	if in.Quotas != nil {
		out.Quotas = make(map[string]int, len(in.Quotas))
		for k, v := range in.Quotas {
			out.Quotas[k] = v
		}
	}
}

func (in *ProjectPlacementSpec) DeepCopy() *ProjectPlacementSpec {
	if in == nil {
		return nil
	}
	out := new(ProjectPlacementSpec)
	in.DeepCopyInto(out)
	return out
}

func (in *ProjectPlacementStatus) DeepCopyInto(out *ProjectPlacementStatus) {
	*out = *in
	in.Effective.DeepCopyInto(&out.Effective)
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
}

func (in *ProjectPlacementStatus) DeepCopy() *ProjectPlacementStatus {
	if in == nil {
		return nil
	}
	out := new(ProjectPlacementStatus)
	in.DeepCopyInto(out)
	return out
}

func (in *ProjectPlacementList) DeepCopyInto(out *ProjectPlacementList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]ProjectPlacement, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *ProjectPlacementList) DeepCopy() *ProjectPlacementList {
	if in == nil {
		return nil
	}
	out := new(ProjectPlacementList)
	in.DeepCopyInto(out)
	return out
}

func (in *ProjectPlacementList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
