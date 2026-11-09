package v1alpha1

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/scheme"
)

// GroupVersion captures the API group and version for management-plane types.
var GroupVersion = schema.GroupVersion{Group: "infra.aegis.yourorg.dev", Version: "v1alpha1"}

// SchemeBuilder registers management-plane API types with the scheme.
var SchemeBuilder = &scheme.Builder{GroupVersion: GroupVersion}

// AddToScheme allows callers to register management-plane types on a runtime.Scheme.
var AddToScheme = SchemeBuilder.AddToScheme
