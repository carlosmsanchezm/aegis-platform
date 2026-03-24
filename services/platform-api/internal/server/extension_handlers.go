package server

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.uber.org/zap"
)

const (
	// Environment variable overrides (optional — defaults work out of the box).
	envExtensionVSIXPath       = "AEGIS_EXTENSION_VSIX_PATH"
	envExtensionVersion        = "AEGIS_EXTENSION_VERSION"
	envExtensionSHA256         = "AEGIS_EXTENSION_SHA256"
	envExtensionSetupScript    = "AEGIS_EXTENSION_SETUP_SCRIPT_PATH"
	envExtensionRequiredVSCode = "AEGIS_EXTENSION_REQUIRED_VSCODE"

	// Defaults — these match the paths baked into the Docker image.
	defaultExtensionVSIXPath       = "/data/extension/aegis-remote.vsix"
	defaultExtensionVersion        = "0.0.1"
	defaultExtensionSetupScript    = "/data/extension/aegis-setup-vscode.sh"
	defaultExtensionRequiredVSCode = "^1.99.0"
)

// ExtensionMetadata contains the VS Code extension metadata returned by the metadata endpoint.
type ExtensionMetadata struct {
	Version               string `json:"version"`
	SHA256                string `json:"sha256,omitempty"`
	VSIXUrl               string `json:"vsixUrl"`
	SetupScriptUrl        string `json:"setupScriptUrl"`
	RequiredVSCodeVersion string `json:"requiredVscodeVersion"`
}

// registerExtensionRoutes registers the VS Code extension HTTP endpoints.
func registerExtensionRoutes(mux *runtime.ServeMux, srv *Server) {
	if mux == nil || srv == nil {
		return
	}

	// GET /api/v1/extension/metadata — returns extension metadata JSON
	if err := mux.HandlePath(http.MethodGet, "/api/v1/extension/metadata", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		srv.handleGetExtensionMetadata(w, r)
	}); err != nil {
		srv.log.Error("failed to register /api/v1/extension/metadata route", zap.Error(err))
	}

	// GET /api/v1/extension/vsix — serves the VSIX file
	if err := mux.HandlePath(http.MethodGet, "/api/v1/extension/vsix", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		srv.handleGetExtensionVSIX(w, r)
	}); err != nil {
		srv.log.Error("failed to register /api/v1/extension/vsix route", zap.Error(err))
	}

	// GET /api/v1/extension/setup-script — serves the setup shell script
	if err := mux.HandlePath(http.MethodGet, "/api/v1/extension/setup-script", func(w http.ResponseWriter, r *http.Request, _ map[string]string) {
		srv.handleGetExtensionSetupScript(w, r)
	}); err != nil {
		srv.log.Error("failed to register /api/v1/extension/setup-script route", zap.Error(err))
	}

	srv.log.Info("Extension routes registered", zap.Strings("routes", []string{
		"/api/v1/extension/metadata",
		"/api/v1/extension/vsix",
		"/api/v1/extension/setup-script",
	}))
}

// vsixSHA256Cache caches the computed SHA-256 so we only hash the file once.
var (
	vsixSHA256Cache string
	vsixSHA256Once  sync.Once
)

// computeVSIXSHA256 returns the SHA-256 hex digest of the VSIX on disk.
// The result is cached after the first successful computation.
func computeVSIXSHA256(path string) string {
	vsixSHA256Once.Do(func() {
		data, err := os.ReadFile(path)
		if err != nil {
			return
		}
		sum := sha256.Sum256(data)
		vsixSHA256Cache = hex.EncodeToString(sum[:])
	})
	return vsixSHA256Cache
}

// handleGetExtensionMetadata returns the VS Code extension metadata as JSON.
func (s *Server) handleGetExtensionMetadata(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("handleGetExtensionMetadata called", zap.String("path", r.URL.Path))

	vsixPath := getEnvOrDefault(envExtensionVSIXPath, defaultExtensionVSIXPath)

	// SHA-256: use explicit env var if set, otherwise compute from file on disk.
	checksum := os.Getenv(envExtensionSHA256)
	if checksum == "" {
		checksum = computeVSIXSHA256(vsixPath)
	}

	metadata := ExtensionMetadata{
		Version:               getEnvOrDefault(envExtensionVersion, defaultExtensionVersion),
		SHA256:                checksum,
		VSIXUrl:               "/api/v1/extension/vsix",
		SetupScriptUrl:        "/api/v1/extension/setup-script",
		RequiredVSCodeVersion: getEnvOrDefault(envExtensionRequiredVSCode, defaultExtensionRequiredVSCode),
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metadata); err != nil {
		s.log.Error("failed to encode extension metadata", zap.Error(err))
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// handleGetExtensionVSIX serves the VSIX file from disk.
func (s *Server) handleGetExtensionVSIX(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("handleGetExtensionVSIX called", zap.String("path", r.URL.Path))

	vsixPath := getEnvOrDefault(envExtensionVSIXPath, defaultExtensionVSIXPath)
	data, err := os.ReadFile(vsixPath)
	if err != nil {
		s.log.Error("failed to read VSIX file", zap.String("path", vsixPath), zap.Error(err))
		http.Error(w, "Extension VSIX not available: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	version := getEnvOrDefault(envExtensionVersion, defaultExtensionVersion)
	filename := fmt.Sprintf("aegis-remote-%s.vsix", version)

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	w.Write(data)
}

// handleGetExtensionSetupScript serves the setup script from disk.
func (s *Server) handleGetExtensionSetupScript(w http.ResponseWriter, r *http.Request) {
	s.log.Debug("handleGetExtensionSetupScript called", zap.String("path", r.URL.Path))

	scriptPath := getEnvOrDefault(envExtensionSetupScript, defaultExtensionSetupScript)
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		s.log.Error("failed to read setup script", zap.String("path", scriptPath), zap.Error(err))
		http.Error(w, "Setup script not available: "+err.Error(), http.StatusServiceUnavailable)
		return
	}

	w.Header().Set("Content-Type", "text/x-shellscript")
	w.Write(data)
}
