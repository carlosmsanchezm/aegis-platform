package server

import (
	"strconv"
	"strings"

	corev1 "k8s.io/api/core/v1"

	infraapi "github.com/yourorg/aegis/services/platform-api/api/v1alpha1"
)

type clusterProfileTemplate struct {
	ID       string
	Version  string
	Provider string
	AWS      *infraapi.AWSInfraSpec
}

func defaultClusterProfiles() map[string]*clusterProfileTemplate {
	return map[string]*clusterProfileTemplate{
		// eks-train: Training cluster with system nodes and GPU ASG (MinSize=0).
		// GPU ASG exists but has no running nodes until a workspace is requested.
		// This follows the Kubeflow pattern: Cluster Autoscaler scales from 0 on-demand.
		"eks-train": {
			ID:       "eks-train",
			Version:  "1.5.0",
			Provider: "aws",
			AWS: &infraapi.AWSInfraSpec{
				Mode:        infraapi.AWSProvisionModeProvision,
				ClusterName: "",
				Version:     "1.34",
				NodePools: []infraapi.NodePool{
					{
						Name:         "system",
						InstanceType: "t3.large", // 2 vCPU, 8GB, 35 pod limit - fits all system + observability pods on one node
						MinSize:      1,
						MaxSize:      3,
						Labels: map[string]string{
							"aegis.dev/purpose": "system",
						},
					},
					{
						// GPU ASG with MinSize=0: no nodes running until workspace requested.
						// Cluster Autoscaler will scale up when it sees pending GPU pods.
						Name:         "gpu",
						InstanceType: "g4dn.xlarge",
						MinSize:      0,
						MaxSize:      5,
						Labels: map[string]string{
							"aegis.dev/purpose":     "training",
							"aegis.io/gpu-flavor":   "nvidia-tesla-t4",
							"aegis.dev/spotAllowed": "false",
						},
						Taints: []corev1.Taint{{
							Key:    "nvidia.com/gpu",
							Value:  "true",
							Effect: corev1.TaintEffectNoSchedule,
						}},
					},
				},
			},
		},
		"eks-general": {
			ID:       "eks-general",
			Version:  "2.2.0",
			Provider: "aws",
			AWS: &infraapi.AWSInfraSpec{
				Mode:        infraapi.AWSProvisionModeProvision,
				ClusterName: "",
				Version:     "1.27",
				NodePools: []infraapi.NodePool{
					{
						Name:         "general",
						InstanceType: "t3.large", // 2 vCPU, 8GB, 35 pod limit - fits observability stack on one node
						MinSize:      2,
						MaxSize:      10,
						Labels: map[string]string{
							"aegis.dev/purpose": "general",
						},
					},
				},
			},
		},
		// eks-system: Minimal cluster with only system nodes - NO GPU ASG.
		// Use this when you don't need GPU capability at all.
		"eks-system": {
			ID:       "eks-system",
			Version:  "1.0.0",
			Provider: "aws",
			AWS: &infraapi.AWSInfraSpec{
				Mode:        infraapi.AWSProvisionModeProvision,
				ClusterName: "",
				Version:     "1.34",
				NodePools: []infraapi.NodePool{
					{
						Name:         "system",
						InstanceType: "t3.large", // 2 vCPU, 8GB, 35 pod limit - fits all system + observability pods on one node
						MinSize:      1,
						MaxSize:      5,
						Labels: map[string]string{
							"aegis.dev/purpose": "system",
						},
					},
				},
			},
		},
		"eks-secure": {
			ID:       "eks-secure",
			Version:  "1.3.0",
			Provider: "aws",
			AWS: &infraapi.AWSInfraSpec{
				Mode:        infraapi.AWSProvisionModeProvision,
				ClusterName: "",
				Version:     "1.28",
				NodePools: []infraapi.NodePool{
					{
						Name:         "control",
						InstanceType: "t3.large", // 2 vCPU, 8GB, 35 pod limit - fits all system + observability pods on one node
						MinSize:      3,
						MaxSize:      6,
						Labels: map[string]string{
							"aegis.dev/purpose": "control",
						},
					},
					{
						// GPU ASG with MinSize=0 for secure workloads
						Name:         "gpu",
						InstanceType: "g4dn.xlarge",
						MinSize:      0,
						MaxSize:      5,
						Labels: map[string]string{
							"aegis.dev/purpose":     "secure-gpu",
							"aegis.io/gpu-flavor":   "nvidia-tesla-t4",
							"aegis.dev/spotAllowed": "false",
						},
						Taints: []corev1.Taint{{
							Key:    "nvidia.com/gpu",
							Value:  "true",
							Effect: corev1.TaintEffectNoSchedule,
						}},
					},
				},
			},
		},
	}
}

func (t *clusterProfileTemplate) instantiate(clusterName string, params map[string]string) *infraapi.AWSInfraSpec {
	if t == nil || t.AWS == nil {
		return nil
	}
	spec := t.AWS.DeepCopy()
	spec.Mode = infraapi.AWSProvisionModeProvision
	spec.ClusterName = sanitizeClusterName(clusterName)
	spec.NodePools = cloneNodePools(spec.NodePools)
	applyClusterParameters(spec, params)
	return spec
}

func cloneNodePools(pools []infraapi.NodePool) []infraapi.NodePool {
	if len(pools) == 0 {
		return nil
	}
	cloned := make([]infraapi.NodePool, len(pools))
	for i := range pools {
		cloned[i] = pools[i]
		if pools[i].Labels != nil {
			labels := make(map[string]string, len(pools[i].Labels))
			for k, v := range pools[i].Labels {
				labels[k] = v
			}
			cloned[i].Labels = labels
		}
		if pools[i].Taints != nil {
			cloned[i].Taints = make([]corev1.Taint, len(pools[i].Taints))
			copy(cloned[i].Taints, pools[i].Taints)
		}
	}
	return cloned
}

func applyClusterParameters(spec *infraapi.AWSInfraSpec, params map[string]string) {
	if spec == nil || len(params) == 0 {
		return
	}
	if version := strings.TrimSpace(params["k8s.version"]); version != "" {
		spec.Version = version
	}
	if vpc := strings.TrimSpace(params["network.vpcId"]); vpc != "" {
		spec.VpcID = vpc
	}
	if rawSubnets := strings.TrimSpace(params["network.subnetIds"]); rawSubnets != "" {
		parts := strings.Split(rawSubnets, ",")
		subnets := make([]string, 0, len(parts))
		for _, part := range parts {
			val := strings.TrimSpace(part)
			if val != "" {
				subnets = append(subnets, val)
			}
		}
		spec.SubnetIDs = subnets
	}
	if count := strings.TrimSpace(params["gpu.count"]); count != "" {
		if n, err := strconv.Atoi(count); err == nil {
			adjustGpuPoolSize(spec, n)
		}
	}
	if gpuType := strings.TrimSpace(params["gpu.type"]); gpuType != "" {
		adjustGpuInstanceType(spec, gpuType)
	}
	if spot := strings.TrimSpace(params["nodePool.spotAllowed"]); spot != "" {
		if allowed, err := strconv.ParseBool(spot); err == nil {
			labelGpuPool(spec, "aegis.dev/spotAllowed", strconv.FormatBool(allowed))
		}
	}
}

func adjustGpuPoolSize(spec *infraapi.AWSInfraSpec, count int) {
	if count < 0 {
		count = 0
	}
	idx := gpuPoolIndex(spec.NodePools)
	if idx < 0 {
		return
	}
	pool := spec.NodePools[idx]
	val := int32(count)
	pool.MinSize = val
	// Don't set MaxSize to 0 when count is 0 - this would prevent autoscaling.
	// The Kubeflow pattern requires MinSize=0 but MaxSize>0 so the Cluster
	// Autoscaler can scale up GPU nodes on-demand when workspaces are requested.
	// Only increase MaxSize if the requested count exceeds current MaxSize.
	if val > 0 && pool.MaxSize < val {
		pool.MaxSize = val
	}
	spec.NodePools[idx] = pool
}

func adjustGpuInstanceType(spec *infraapi.AWSInfraSpec, gpuType string) {
	idx := gpuPoolIndex(spec.NodePools)
	if idx < 0 {
		return
	}
	switch strings.ToUpper(gpuType) {
	case "T4", "G4DN", "SMALL":
		// T4 GPU - most cost-effective option (~$0.50/hr for g4dn.xlarge)
		spec.NodePools[idx].InstanceType = "g4dn.xlarge"
	case "H100":
		// H100 instances: p5.48xlarge is the only p5 size available.
		// For cost-effective testing, use g5.xlarge (A10G) instead.
		// p5.48xlarge has 8x H100 GPUs and costs ~$98/hr.
		spec.NodePools[idx].InstanceType = "p5.48xlarge"
	case "A100":
		spec.NodePools[idx].InstanceType = "p4d.24xlarge"
	case "A10G":
		spec.NodePools[idx].InstanceType = "g5.xlarge"
	case "NONE":
		spec.NodePools[idx].InstanceType = "m6i.large"
	}
}

func labelGpuPool(spec *infraapi.AWSInfraSpec, key, value string) {
	idx := gpuPoolIndex(spec.NodePools)
	if idx < 0 {
		return
	}
	if spec.NodePools[idx].Labels == nil {
		spec.NodePools[idx].Labels = map[string]string{}
	}
	spec.NodePools[idx].Labels[key] = value
}

func gpuPoolIndex(pools []infraapi.NodePool) int {
	for i, pool := range pools {
		name := strings.ToLower(strings.TrimSpace(pool.Name))
		if strings.Contains(name, "gpu") {
			return i
		}
		if purpose := strings.ToLower(pool.Labels["aegis.dev/purpose"]); purpose == "training" || purpose == "secure-gpu" {
			return i
		}
	}
	return -1
}
