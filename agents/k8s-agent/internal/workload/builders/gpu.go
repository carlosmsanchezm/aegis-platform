package builders

import "strings"

// AutoGPUResource infers the GPU resource name from a flavor string.
func AutoGPUResource(flavor string) string {
	lower := strings.ToLower(flavor)
	if strings.HasPrefix(lower, "cpu") {
		return ""
	}
	if strings.HasPrefix(lower, "mig-") {
		return "nvidia.com/" + strings.TrimPrefix(flavor, "mig-")
	}
	if flavor != "" {
		return "nvidia.com/gpu"
	}
	return ""
}
