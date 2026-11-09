/*
Package providers declares the plugin registry used to integrate provider-specific controllers.
*/
package providers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	aegisv1alpha2 "github.com/yourorg/aegis/agents/k8s-agent/api/v1alpha2"
)

// Plugin defines the minimal provider integration contract.
type Plugin interface {
	// Name identifies the provider plugin (e.g. aws, gcp, azure).
	Name() string
	// SupportsBackend returns true when the plugin manages the given backend reference.
	SupportsBackend(ref *corev1.TypedLocalObjectReference) bool
	// Reconcile ensures provider-specific resources for the Workspace.
	Reconcile(ctx context.Context, ws *aegisv1alpha2.Workspace) error
	// Cleanup performs provider-specific teardown on Workspace deletion.
	Cleanup(ctx context.Context, ws *aegisv1alpha2.Workspace) error
}

// Registry holds registered provider plugins.
type Registry struct {
	plugins map[string]Plugin
}

// NewRegistry returns an empty registry.
func NewRegistry() *Registry {
	return &Registry{plugins: map[string]Plugin{}}
}

// Register adds a plugin to the registry.
func (r *Registry) Register(plugin Plugin) {
	if plugin == nil {
		return
	}
	if r.plugins == nil {
		r.plugins = map[string]Plugin{}
	}
	r.plugins[plugin.Name()] = plugin
}

// Resolve returns the first plugin supporting the provided backend reference.
func (r *Registry) Resolve(ref *corev1.TypedLocalObjectReference) Plugin {
	if ref == nil {
		return nil
	}
	for _, plugin := range r.plugins {
		if plugin.SupportsBackend(ref) {
			return plugin
		}
	}
	return nil
}

// DefaultConditions exposes standard provider status conditions.
func DefaultConditions() []metav1.Condition {
	return []metav1.Condition{
		{
			Type:   "ClusterReady",
			Status: metav1.ConditionUnknown,
			Reason: "Pending",
		},
		{
			Type:   "NetworkReady",
			Status: metav1.ConditionUnknown,
			Reason: "Pending",
		},
		{
			Type:   "StorageReady",
			Status: metav1.ConditionUnknown,
			Reason: "Pending",
		},
	}
}
