package workspace

import "sort"

const (
	DefaultVSCodePort int32 = 11111

	EnvVSCodeCommit   = "VSCODE_COMMIT"
	EnvVSCodeQuality  = "VSCODE_QUALITY"
	EnvPUID           = "PUID"
	EnvPGID           = "PGID"
	EnvPasswordAccess = "PASSWORD_ACCESS"
	EnvUserName       = "USER_NAME"
	EnvUserPassword   = "USER_PASSWORD"

	// DefaultVSCodeQuality is provided for environments that explicitly pin VS Code builds.
	DefaultVSCodeQuality = "stable"
	DefaultPUID           = "1000"
	DefaultPGID           = "1000"
	DefaultPasswordAccess = "true"
	DefaultUserName       = "aegis"
	DefaultUserPassword   = "aegis123"
	DefaultWorkspaceImage = "195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:latest"
)

// DefaultEnv returns a writable map containing the baseline VS Code server
// environment expected by remote workspaces.
func DefaultEnv() map[string]string {
	return map[string]string{
		EnvVSCodeQuality:  DefaultVSCodeQuality,
		EnvPUID:           DefaultPUID,
		EnvPGID:           DefaultPGID,
		EnvPasswordAccess: DefaultPasswordAccess,
		EnvUserName:       DefaultUserName,
		EnvUserPassword:   DefaultUserPassword,
	}
}

// MergeEnv combines default workspace environment variables with user-provided
// overrides. User keys win when non-empty. Empty strings are ignored so that
// callers can omit variables without clearing defaults unintentionally.
func MergeEnv(user, defaults map[string]string) map[string]string {
	if len(user) == 0 && len(defaults) == 0 {
		return nil
	}

	size := len(defaults)
	if len(user) > size {
		size = len(user)
	}
	merged := make(map[string]string, size)

	for k, v := range defaults {
		if v == "" {
			continue
		}
		merged[k] = v
	}
	for k, v := range user {
		if v == "" {
			continue
		}
		merged[k] = v
	}

	if len(merged) == 0 {
		return nil
	}
	return merged
}

// EnsureDefaultPorts normalizes the requested port list, guaranteeing the VS
// Code port is always present and falling back to the SSH port when nothing is
// provided. Negative and zero ports are dropped, and the returned slice is
// sorted for stability.
func EnsureDefaultPorts(input []int32) []int32 {
	portsSet := make(map[int32]struct{}, len(input)+1)
	for _, port := range input {
		if port <= 0 {
			continue
		}
		portsSet[port] = struct{}{}
	}
	portsSet[DefaultVSCodePort] = struct{}{}
	ports := make([]int32, 0, len(portsSet))
	for port := range portsSet {
		ports = append(ports, port)
	}
	sort.Slice(ports, func(i, j int) bool { return ports[i] < ports[j] })
	return ports
}

// CopyEnv returns a shallow copy of the provided environment map.
func CopyEnv(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}
