package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:shortName=pinfra

// ProjectInfra represents a declarative request to provision or import
// infrastructure for a specific Aegis project.
type ProjectInfra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ProjectInfraSpec   `json:"spec,omitempty"`
	Status ProjectInfraStatus `json:"status,omitempty"`
}

// ProjectInfraSpec defines the desired state for project-scoped infrastructure.
type ProjectInfraSpec struct {
	ProjectID                 string                  `json:"projectId"`
	Provider                  string                  `json:"provider"`
	Region                    string                  `json:"region"`
	Strategy                  string                  `json:"strategy,omitempty"`
	ImportKubeconfigSecretRef *corev1.SecretReference `json:"importKubeconfigSecretRef,omitempty"`
	Aws                       *AWSInfraSpec           `json:"aws,omitempty"`
	Addons                    map[string]bool         `json:"addons,omitempty"`
	Labels                    map[string]string       `json:"labels,omitempty"`
}

// AWSProvisionMode declares how the automation runner should reconcile AWS resources.
type AWSProvisionMode string

const (
	AWSProvisionModeProvision AWSProvisionMode = "Provision"
	AWSProvisionModeImport    AWSProvisionMode = "Import"
)

// SecretKeyReference describes a namespaced secret key selector.
type SecretKeyReference struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
	Key       string `json:"key"`
}

// AWSImportSpec describes existing infrastructure to import into the management plane.
type AWSImportSpec struct {
	ClusterID           string             `json:"clusterId"`
	KubeconfigSecretRef SecretKeyReference `json:"kubeconfigSecretRef"`
}

// AWSInfraSpec captures AWS-specific provisioning parameters.
type AWSInfraSpec struct {
	AccountID          string           `json:"accountId,omitempty"`
	Mode               AWSProvisionMode `json:"mode,omitempty"`
	RoleARN            string           `json:"roleArn,omitempty"`
	ExternalID         string           `json:"externalId,omitempty"`
	SubnetIDs          []string         `json:"subnetIds,omitempty"`
	VpcID              string           `json:"vpcId,omitempty"`
	ClusterName        string           `json:"clusterName,omitempty"`
	Version            string           `json:"version,omitempty"`
	NodePools          []NodePool       `json:"nodePools,omitempty"`
	AdditionalClusters []AWSClusterSpec `json:"additionalClusters,omitempty"`
	Imports            []AWSImportSpec  `json:"imports,omitempty"`
}

// AWSClusterSpec describes a secondary cluster to provision in AWS.
type AWSClusterSpec struct {
	ClusterName string     `json:"clusterName"`
	Version     string     `json:"version,omitempty"`
	NodePools   []NodePool `json:"nodePools,omitempty"`
}

// NodePool represents a logical node group definition for a cluster.
type NodePool struct {
	Name         string            `json:"name"`
	InstanceType string            `json:"instanceType"`
	MinSize      int32             `json:"minSize"`
	MaxSize      int32             `json:"maxSize"`
	Labels       map[string]string `json:"labels,omitempty"`
	Taints       []corev1.Taint    `json:"taints,omitempty"`
}

// ProjectInfraStatus reports provisioning progress and outputs.
type ProjectInfraStatus struct {
	Phase              string             `json:"phase,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
	Outputs            []ClusterOutput    `json:"outputs,omitempty"`
	CostHintUSDPerHour float64            `json:"costHintUsdPerHour,omitempty"`
	LastSyncTime       *metav1.Time       `json:"lastSyncTime,omitempty"`
}

// ClusterOutput captures key material for a provisioned cluster.
type ClusterOutput struct {
	ClusterID           string              `json:"clusterId"`
	Name                string              `json:"name"`
	Region              string              `json:"region"`
	KubeconfigSecretKey string              `json:"kubeconfigSecretKey"`
	Observability       ObservabilityOutput `json:"observability,omitempty"`
}

// ObservabilityOutput holds optional telemetry endpoints for a cluster.
type ObservabilityOutput struct {
	Namespace                string `json:"namespace,omitempty"`
	PrometheusService        string `json:"prometheusService,omitempty"`
	PrometheusPort           int32  `json:"prometheusPort,omitempty"`
	AlertmanagerService      string `json:"alertmanagerService,omitempty"`
	AlertmanagerPort         int32  `json:"alertmanagerPort,omitempty"`
	AlertmanagerConfigSecret string `json:"alertmanagerConfigSecret,omitempty"`
	MetricsServerService     string `json:"metricsServerService,omitempty"`
	MetricsServerPort        int32  `json:"metricsServerPort,omitempty"`
	OtelEndpoint             string `json:"otelEndpoint,omitempty"`
	MetricsURL               string `json:"metricsUrl,omitempty"`
}

// +kubebuilder:object:root=true

// ProjectInfraList contains a list of ProjectInfra objects.
type ProjectInfraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ProjectInfra `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ProjectInfra{}, &ProjectInfraList{})
}

func (in *ProjectInfra) DeepCopyInto(out *ProjectInfra) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

func (in *ProjectInfra) DeepCopy() *ProjectInfra {
	if in == nil {
		return nil
	}
	out := new(ProjectInfra)
	in.DeepCopyInto(out)
	return out
}

func (in *ProjectInfra) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

func (in *ProjectInfraSpec) DeepCopyInto(out *ProjectInfraSpec) {
	*out = *in
	if in.ImportKubeconfigSecretRef != nil {
		out.ImportKubeconfigSecretRef = &corev1.SecretReference{
			Name:      in.ImportKubeconfigSecretRef.Name,
			Namespace: in.ImportKubeconfigSecretRef.Namespace,
		}
	}
	if in.Aws != nil {
		out.Aws = new(AWSInfraSpec)
		in.Aws.DeepCopyInto(out.Aws)
	}
	if in.Addons != nil {
		out.Addons = make(map[string]bool, len(in.Addons))
		for k, v := range in.Addons {
			out.Addons[k] = v
		}
	}
	if in.Labels != nil {
		out.Labels = make(map[string]string, len(in.Labels))
		for k, v := range in.Labels {
			out.Labels[k] = v
		}
	}
}

func (in *ProjectInfraSpec) DeepCopy() *ProjectInfraSpec {
	if in == nil {
		return nil
	}
	out := new(ProjectInfraSpec)
	in.DeepCopyInto(out)
	return out
}

func (in *AWSInfraSpec) DeepCopyInto(out *AWSInfraSpec) {
	*out = *in
	if in.SubnetIDs != nil {
		out.SubnetIDs = make([]string, len(in.SubnetIDs))
		copy(out.SubnetIDs, in.SubnetIDs)
	}
	if in.NodePools != nil {
		out.NodePools = make([]NodePool, len(in.NodePools))
		for i := range in.NodePools {
			in.NodePools[i].DeepCopyInto(&out.NodePools[i])
		}
	}
	if in.AdditionalClusters != nil {
		out.AdditionalClusters = make([]AWSClusterSpec, len(in.AdditionalClusters))
		for i := range in.AdditionalClusters {
			in.AdditionalClusters[i].DeepCopyInto(&out.AdditionalClusters[i])
		}
	}
	if in.SubnetIDs != nil {
		out.SubnetIDs = make([]string, len(in.SubnetIDs))
		copy(out.SubnetIDs, in.SubnetIDs)
	}
	if in.Imports != nil {
		out.Imports = make([]AWSImportSpec, len(in.Imports))
		for i := range in.Imports {
			in.Imports[i].DeepCopyInto(&out.Imports[i])
		}
	}
}

func (in *AWSInfraSpec) DeepCopy() *AWSInfraSpec {
	if in == nil {
		return nil
	}
	out := new(AWSInfraSpec)
	in.DeepCopyInto(out)
	return out
}

func (in *AWSClusterSpec) DeepCopyInto(out *AWSClusterSpec) {
	*out = *in
	if in.NodePools != nil {
		out.NodePools = make([]NodePool, len(in.NodePools))
		for i := range in.NodePools {
			in.NodePools[i].DeepCopyInto(&out.NodePools[i])
		}
	}
}

func (in *AWSClusterSpec) DeepCopy() *AWSClusterSpec {
	if in == nil {
		return nil
	}
	out := new(AWSClusterSpec)
	in.DeepCopyInto(out)
	return out
}

func (in *AWSImportSpec) DeepCopyInto(out *AWSImportSpec) {
	*out = *in
}

func (in *AWSImportSpec) DeepCopy() *AWSImportSpec {
	if in == nil {
		return nil
	}
	out := new(AWSImportSpec)
	in.DeepCopyInto(out)
	return out
}

func (in *SecretKeyReference) DeepCopyInto(out *SecretKeyReference) {
	*out = *in
}

func (in *SecretKeyReference) DeepCopy() *SecretKeyReference {
	if in == nil {
		return nil
	}
	out := new(SecretKeyReference)
	in.DeepCopyInto(out)
	return out
}

func (in *NodePool) DeepCopyInto(out *NodePool) {
	*out = *in
	if in.Labels != nil {
		out.Labels = make(map[string]string, len(in.Labels))
		for k, v := range in.Labels {
			out.Labels[k] = v
		}
	}
	if in.Taints != nil {
		out.Taints = make([]corev1.Taint, len(in.Taints))
		for i := range in.Taints {
			in.Taints[i].DeepCopyInto(&out.Taints[i])
		}
	}
}

func (in *NodePool) DeepCopy() *NodePool {
	if in == nil {
		return nil
	}
	out := new(NodePool)
	in.DeepCopyInto(out)
	return out
}

func (in *ProjectInfraStatus) DeepCopyInto(out *ProjectInfraStatus) {
	*out = *in
	if in.Conditions != nil {
		out.Conditions = make([]metav1.Condition, len(in.Conditions))
		for i := range in.Conditions {
			in.Conditions[i].DeepCopyInto(&out.Conditions[i])
		}
	}
	if in.Outputs != nil {
		out.Outputs = make([]ClusterOutput, len(in.Outputs))
		for i := range in.Outputs {
			in.Outputs[i].DeepCopyInto(&out.Outputs[i])
		}
	}
	if in.LastSyncTime != nil {
		out.LastSyncTime = in.LastSyncTime.DeepCopy()
	}
}

func (in *ProjectInfraStatus) DeepCopy() *ProjectInfraStatus {
	if in == nil {
		return nil
	}
	out := new(ProjectInfraStatus)
	in.DeepCopyInto(out)
	return out
}

func (in *ClusterOutput) DeepCopyInto(out *ClusterOutput) {
	*out = *in
}

func (in *ClusterOutput) DeepCopy() *ClusterOutput {
	if in == nil {
		return nil
	}
	out := new(ClusterOutput)
	in.DeepCopyInto(out)
	return out
}

func (in *ProjectInfraList) DeepCopyInto(out *ProjectInfraList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]ProjectInfra, len(in.Items))
		for i := range in.Items {
			in.Items[i].DeepCopyInto(&out.Items[i])
		}
	}
}

func (in *ProjectInfraList) DeepCopy() *ProjectInfraList {
	if in == nil {
		return nil
	}
	out := new(ProjectInfraList)
	in.DeepCopyInto(out)
	return out
}

func (in *ProjectInfraList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}
