package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func testServer(t *testing.T) *Server {
	t.Helper()
	return &Server{log: zap.NewNop()}
}

func TestHandleGetExtensionMetadata(t *testing.T) {
	srv := testServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/extension/metadata", nil)
	w := httptest.NewRecorder()
	srv.handleGetExtensionMetadata(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected application/json, got %s", ct)
	}

	var meta ExtensionMetadata
	if err := json.NewDecoder(w.Body).Decode(&meta); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if meta.Version == "" {
		t.Error("version should not be empty")
	}
	if meta.VSIXUrl != "/api/v1/extension/vsix" {
		t.Errorf("expected vsixUrl /api/v1/extension/vsix, got %s", meta.VSIXUrl)
	}
	if meta.SetupScriptUrl != "/api/v1/extension/setup-script" {
		t.Errorf("expected setupScriptUrl /api/v1/extension/setup-script, got %s", meta.SetupScriptUrl)
	}
	if meta.RequiredVSCodeVersion == "" {
		t.Error("requiredVscodeVersion should not be empty")
	}
}

func TestHandleGetExtensionMetadata_SHA256FromEnv(t *testing.T) {
	t.Setenv(envExtensionSHA256, "abc123deadbeef")

	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/extension/metadata", nil)
	w := httptest.NewRecorder()
	srv.handleGetExtensionMetadata(w, req)

	var meta ExtensionMetadata
	json.NewDecoder(w.Body).Decode(&meta)
	if meta.SHA256 != "abc123deadbeef" {
		t.Errorf("expected SHA256 from env, got %s", meta.SHA256)
	}
}

func TestHandleGetExtensionVSIX_FileNotFound(t *testing.T) {
	t.Setenv(envExtensionVSIXPath, "/nonexistent/path/to.vsix")

	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/extension/vsix", nil)
	w := httptest.NewRecorder()
	srv.handleGetExtensionVSIX(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when file missing, got %d", w.Code)
	}
}

func TestHandleGetExtensionVSIX_ServesFile(t *testing.T) {
	tmp := t.TempDir()
	vsixPath := filepath.Join(tmp, "test.vsix")
	content := []byte("fake-vsix-content")
	os.WriteFile(vsixPath, content, 0644)
	t.Setenv(envExtensionVSIXPath, vsixPath)

	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/extension/vsix", nil)
	w := httptest.NewRecorder()
	srv.handleGetExtensionVSIX(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/octet-stream" {
		t.Errorf("expected application/octet-stream, got %s", ct)
	}
	if cd := w.Header().Get("Content-Disposition"); cd == "" {
		t.Error("expected Content-Disposition header")
	}
	if w.Body.String() != string(content) {
		t.Error("response body does not match file content")
	}
}

func TestHandleGetExtensionSetupScript_FileNotFound(t *testing.T) {
	t.Setenv(envExtensionSetupScript, "/nonexistent/script.sh")

	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/extension/setup-script", nil)
	w := httptest.NewRecorder()
	srv.handleGetExtensionSetupScript(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503 when file missing, got %d", w.Code)
	}
}

func TestHandleGetExtensionSetupScript_ServesFile(t *testing.T) {
	tmp := t.TempDir()
	scriptPath := filepath.Join(tmp, "setup.sh")
	content := []byte("#!/bin/bash\necho hello")
	os.WriteFile(scriptPath, content, 0755)
	t.Setenv(envExtensionSetupScript, scriptPath)

	srv := testServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/v1/extension/setup-script", nil)
	w := httptest.NewRecorder()
	srv.handleGetExtensionSetupScript(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/x-shellscript" {
		t.Errorf("expected text/x-shellscript, got %s", ct)
	}
	if w.Body.String() != string(content) {
		t.Error("response body does not match file content")
	}
}
