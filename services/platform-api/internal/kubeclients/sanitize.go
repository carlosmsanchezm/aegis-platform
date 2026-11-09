package kubeclients

import (
	"encoding/base64"
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// dangerousExecEnvVars are environment variable names that are
// environment-specific and must be stripped from kubeconfig exec
// configs before storing. These work on a developer's machine but
// break inside the platform-api pod.
var dangerousExecEnvVars = map[string]bool{
	"AWS_PROFILE":                  true,
	"AWS_CONFIG_FILE":              true,
	"AWS_SHARED_CREDENTIALS_FILE":  true,
	"AWS_DEFAULT_PROFILE":          true,
	"KUBECONFIG":                   true,
}

// SanitizeKubeconfig strips environment-specific values from kubeconfig
// exec configs that would break inside the platform-api pod.
// Returns the sanitized YAML and a list of warnings for anything removed.
// If parsing fails, returns the original bytes with a warning.
func SanitizeKubeconfig(raw []byte) ([]byte, []string) {
	cfg, err := clientcmd.Load(raw)
	if err != nil {
		return raw, []string{fmt.Sprintf("failed to parse kubeconfig for sanitization: %v", err)}
	}

	var warnings []string
	for userName, authInfo := range cfg.AuthInfos {
		if authInfo == nil || authInfo.Exec == nil {
			continue
		}
		cleaned := make([]clientcmdapi.ExecEnvVar, 0, len(authInfo.Exec.Env))
		for _, env := range authInfo.Exec.Env {
			if dangerousExecEnvVars[env.Name] {
				warnings = append(warnings, fmt.Sprintf(
					"removed %s=%q from exec env in user %q (environment-specific)",
					env.Name, env.Value, userName))
				continue
			}
			cleaned = append(cleaned, env)
		}
		authInfo.Exec.Env = cleaned
	}

	if len(warnings) == 0 {
		return raw, nil
	}

	out, err := clientcmd.Write(*cfg)
	if err != nil {
		return raw, append(warnings, fmt.Sprintf("failed to re-serialize kubeconfig: %v", err))
	}
	return out, warnings
}

// ExtractClusterEndpointCA parses a kubeconfig and returns the server endpoint
// and base64-encoded CA data from the first cluster entry. Returns empty strings
// if parsing fails or no cluster is found.
func ExtractClusterEndpointCA(raw []byte) (endpoint, ca string) {
	cfg, err := clientcmd.Load(raw)
	if err != nil {
		return "", ""
	}
	// Use the current context's cluster if available
	if ctx, ok := cfg.Contexts[cfg.CurrentContext]; ok && ctx != nil {
		if cluster, ok := cfg.Clusters[ctx.Cluster]; ok && cluster != nil {
			ep := cluster.Server
			caData := ""
			if len(cluster.CertificateAuthorityData) > 0 {
				caData = base64.StdEncoding.EncodeToString(cluster.CertificateAuthorityData)
			}
			return ep, caData
		}
	}
	// Fallback: use the first cluster entry
	for _, cluster := range cfg.Clusters {
		if cluster == nil {
			continue
		}
		ep := cluster.Server
		caData := ""
		if len(cluster.CertificateAuthorityData) > 0 {
			caData = base64.StdEncoding.EncodeToString(cluster.CertificateAuthorityData)
		}
		return ep, caData
	}
	return "", ""
}
