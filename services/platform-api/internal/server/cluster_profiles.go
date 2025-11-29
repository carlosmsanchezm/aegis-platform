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
		"eks-gpu-train": {
			ID:       "eks-gpu-train",
			Version:  "1.4.0",
			Provider: "aws",
			AWS: &infraapi.AWSInfraSpec{
				Mode:        infraapi.AWSProvisionModeProvision,
				ClusterName: "",
				Version:     "1.34",
				NodePools: []infraapi.NodePool{
					{
						Name:         "system",
						InstanceType: "m6i.large",
						MinSize:      1,
						MaxSize:      2,
						Labels: map[string]string{
							"aegis.dev/purpose": "system",
						},
					},
					{
						Name:         "gpu",
						InstanceType: "g5.2xlarge",
						MinSize:      0,
						MaxSize:      1,
						Labels: map[string]string{
							"aegis.dev/purpose": "training",
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
			Version:  "2.1.0",
			Provider: "aws",
			AWS: &infraapi.AWSInfraSpec{
				Mode:        infraapi.AWSProvisionModeProvision,
				ClusterName: "",
				Version:     "1.27",
				NodePools: []infraapi.NodePool{
					{
						Name:         "general",
						InstanceType: "m6i.2xlarge",
						MinSize:      2,
						MaxSize:      10,
						Labels: map[string]string{
							"aegis.dev/purpose": "general",
						},
					},
					{
						Name:         "analytics",
						InstanceType: "r6i.2xlarge",
						MinSize:      1,
						MaxSize:      5,
						Labels: map[string]string{
							"aegis.dev/purpose": "analytics",
						},
					},
				},
			},
		},
		"eks-secure": {
			ID:       "eks-secure",
			Version:  "1.2.0",
			Provider: "aws",
			AWS: &infraapi.AWSInfraSpec{
				Mode:        infraapi.AWSProvisionModeProvision,
				ClusterName: "",
				Version:     "1.28",
				NodePools: []infraapi.NodePool{
					{
						Name:         "control",
						InstanceType: "m6i.4xlarge",
						MinSize:      3,
						MaxSize:      6,
						Labels: map[string]string{
							"aegis.dev/purpose": "control",
						},
					},
					{
						Name:         "gpu",
						InstanceType: "p4d.24xlarge",
						MinSize:      1,
						MaxSize:      3,
						Labels: map[string]string{
							"aegis.dev/purpose": "secure-gpu",
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
	if val == 0 {
		pool.MaxSize = 0
	} else if pool.MaxSize < val {
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
	case "G5", "SMALL":
		spec.NodePools[idx].InstanceType = "g5.2xlarge"
	case "H100":
		// Request the smallest available GPU-capable family by default to avoid vCPU quota exhaustion.
		spec.NodePools[idx].InstanceType = "g5.2xlarge"
	case "A100":
		spec.NodePools[idx].InstanceType = "p4d.24xlarge"
	case "A10G":
		spec.NodePools[idx].InstanceType = "g5.2xlarge"
	case "NONE":
		spec.NodePools[idx].InstanceType = "m6i.4xlarge"
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
