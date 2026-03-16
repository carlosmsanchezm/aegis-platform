package server

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
)

const (
	// Environment variables for PKI configuration
	envStepCaRootCAB64  = "AEGIS_STEP_CA_ROOT_CA_B64"
	envStepCaRootCAFile = "AEGIS_STEP_CA_ROOT_CA_FILE"
	envStepCaURL        = "AEGIS_STEP_CA_URL"
	envStepProvKID      = "AEGIS_STEP_PROVISIONER_KID"
	envStepProvName     = "AEGIS_STEP_PROVISIONER_NAME"
	envClusterIssuer    = "AEGIS_CLUSTER_ISSUER_NAME"
)

// PKIConfig contains the step-ca configuration for spoke clusters
type PKIConfig struct {
	// RootCA is the PEM-encoded root CA certificate
	RootCA string `json:"rootCA"`
	// RootCABase64 is the base64-encoded root CA (for helm values)
	RootCABase64 string `json:"rootCABase64"`
	// StepCaURL is the URL of the step-ca server
	StepCaURL string `json:"stepCaURL,omitempty"`
	// ProvisionerName is the name of the JWK provisioner
	ProvisionerName string `json:"provisionerName,omitempty"`
	// ProvisionerKID is the key ID of the provisioner
	ProvisionerKID string `json:"provisionerKID,omitempty"`
	// ClusterIssuerName is the name of the StepClusterIssuer
	ClusterIssuerName string `json:"clusterIssuerName,omitempty"`
}

// registerPKIRoutes registers the PKI-related HTTP endpoints.
func registerPKIRoutes(mux *runtime.ServeMux, srv *Server) {
	if mux == nil || srv == nil {
		return
	}

	// GET /api/v1/pki/root-ca - Returns the root CA certificate (PEM)
	if err := mux.HandlePath(http.MethodGet, "/api/v1/pki/root-ca", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		srv.handleGetRootCA(w, r)
	}); err != nil {
		srv.log.Error("failed to register /api/v1/pki/root-ca route", zap.Error(err))
	}

	// GET /api/v1/pki/config - Returns full PKI config for spoke cert-manager
	if err := mux.HandlePath(http.MethodGet, "/api/v1/pki/config", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		srv.handleGetPKIConfig(w, r)
	}); err != nil {
		srv.log.Error("failed to register /api/v1/pki/config route", zap.Error(err))
	}

	srv.log.Info("PKI routes registered", zap.Strings("routes", []string{
		"/api/v1/pki/root-ca",
		"/api/v1/pki/config",
	}))
}

// handleGetRootCA returns the step-ca root CA certificate in PEM format.
// This is used by VS Code extension and spoke clusters to trust the CA.
func (s *Server) handleGetRootCA(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("handleGetRootCA called", zap.String("path", r.URL.Path))

	rootCA, err := s.loadRootCA()
	if err != nil {
		s.log.Error("failed to load root CA", zap.Error(err))
		http.Error(w, "PKI not configured: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	// Check Accept header for format preference
	accept := r.Header.Get("Accept")
	if strings.Contains(accept, "application/json") {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"rootCA":       rootCA,
			"rootCABase64": base64.StdEncoding.EncodeToString([]byte(rootCA)),
		})
		return
	}

	// Default: return PEM directly (for curl, wget, etc.)
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Header().Set("Content-Disposition", "attachment; filename=\"aegis-root-ca.crt\"")
	w.Write([]byte(rootCA))
}

// handleGetPKIConfig returns the full PKI configuration needed by spoke clusters.
// This includes the root CA, step-ca URL, and provisioner details.
func (s *Server) handleGetPKIConfig(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("handleGetPKIConfig called", zap.String("path", r.URL.Path))

	rootCA, err := s.loadRootCA()
	if err != nil {
		s.log.Error("failed to load root CA", zap.Error(err))
		http.Error(w, "PKI not configured: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	config := PKIConfig{
		RootCA:            rootCA,
		RootCABase64:      base64.StdEncoding.EncodeToString([]byte(rootCA)),
		StepCaURL:         os.Getenv(envStepCaURL),
		ProvisionerName:   getEnvOrDefault(envStepProvName, "aegis"),
		ProvisionerKID:    os.Getenv(envStepProvKID),
		ClusterIssuerName: getEnvOrDefault(envClusterIssuer, "aegis-internal"),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(config); err != nil {
		s.log.Error("failed to encode PKI config", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// loadRootCA loads the root CA certificate from environment or file.
func (s *Server) loadRootCA() (string, error) {
	// Try base64-encoded env var first
	if b64 := os.Getenv(envStepCaRootCAB64); b64 != "" {
		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return "", err
		}
		return string(decoded), nil
	}

	// Try file path
	if path := os.Getenv(envStepCaRootCAFile); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	// Try the spoke-proxy CA (fallback for current setup)
	if path := os.Getenv("AEGIS_SPOKE_PROXY_CA_CERT"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	// Auto-load from mounted TLS secret (platform-api's own cert chain).
	// The TLS secret contains:
	//   ca.crt  = root CA
	//   tls.crt = server cert + intermediate CA
	// Combine root + intermediate for a complete trust chain.
	tlsMountPath := os.Getenv("AEGIS_TLS_MOUNT_PATH")
	if tlsMountPath == "" {
		tlsMountPath = "/etc/aegis-platform-api/tls"
	}
	rootCAPath := tlsMountPath + "/ca.crt"
	tlsCertPath := tlsMountPath + "/tls.crt"

	rootCA, rootErr := os.ReadFile(rootCAPath)
	if rootErr == nil && len(rootCA) > 0 {
		// Try to extract intermediate CA from tls.crt (second cert in chain)
		tlsCert, tlsErr := os.ReadFile(tlsCertPath)
		if tlsErr == nil {
			certs := extractPEMCerts(string(tlsCert))
			if len(certs) > 1 {
				// Combine root + intermediate
				combined := strings.TrimSpace(string(rootCA)) + "\n" + strings.TrimSpace(certs[1]) + "\n"
				s.log.Info("auto-loaded CA chain from TLS mount",
					zap.String("path", tlsMountPath),
					zap.Int("cert_count", strings.Count(combined, "BEGIN CERTIFICATE")),
				)
				return combined, nil
			}
		}
		// Return root CA alone if intermediate not available
		s.log.Info("auto-loaded root CA from TLS mount", zap.String("path", rootCAPath))
		return string(rootCA), nil
	}

	return "", &pkiNotConfiguredError{}
}

type pkiNotConfiguredError struct{}

func (e *pkiNotConfiguredError) Error() string {
	return "PKI not configured. Set AEGIS_STEP_CA_ROOT_CA_B64 or AEGIS_STEP_CA_ROOT_CA_FILE"
}

// extractPEMCerts splits a PEM bundle into individual certificates.
func extractPEMCerts(pem string) []string {
	var certs []string
	var current []string
	inCert := false
	for _, line := range strings.Split(pem, "\n") {
		if strings.Contains(line, "BEGIN CERTIFICATE") {
			inCert = true
			current = nil
		}
		if inCert {
			current = append(current, line)
		}
		if strings.Contains(line, "END CERTIFICATE") {
			inCert = false
			certs = append(certs, strings.Join(current, "\n"))
		}
	}
	return certs
}

func getEnvOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}
