# Convenience tooling for Aegis platform-api service builds.

IMG ?= carlosmsanchez/aegis-platform-api:dev
# Default to the host platform for faster local builds; override to linux/amd64,linux/arm64 when needed.
PLATFORMS ?= $(shell docker info --format '{{.OSType}}/{{.Architecture}}' 2>/dev/null || echo linux/arm64)
# Default to loading locally; override to --push when you need to publish.
BUILDX_OUTPUT ?= --load
GO_FLAGS ?=
ROOT := $(abspath $(CURDIR)/../..)

SHELL := /usr/bin/env bash
.SHELLFLAGS := -euo pipefail -c

.PHONY: help
help:
	@echo "Targets:"
	@echo "  build           Build the platform-api Go binary"
	@echo "  run             Run the service locally (go run ./...)"
	@echo "  pulumi-preview  Run the pulumi preview CLI (set PULUMI_PREVIEW_ARGS=...)"
	@echo "  docker-build    Build & push multi-arch image (set IMG=…)"
	@echo "  docker-buildx   Alias of docker-build"
	@echo "  docker-load     Build multi-arch image and load into local Docker"

.PHONY: build
build:
	GO111MODULE=on go build $(GO_FLAGS) ./...

.PHONY: run
run:
	GO111MODULE=on go run $(GO_FLAGS) ./...

.PHONY: pulumi-preview
pulumi-preview:
	GO111MODULE=on go run $(GO_FLAGS) ./cmd/pulumi-preview -- $(PULUMI_PREVIEW_ARGS)

SOVRAN_DIR ?= $(HOME)/code/sovran
VSIX_SRC   ?= $(SOVRAN_DIR)/aegis-vscode-remote/extension/aegis-remote.vsix

.PHONY: docker-build docker-buildx
# Build the image (defaults to host arch + local load; override PLATFORMS/BUILDX_OUTPUT for multi-arch + push).
# Automatically bundles the VS Code extension VSIX if found in the sovran repo or build context.
docker-build docker-buildx:
	@echo "Building $(IMG) for $(PLATFORMS) with output $(BUILDX_OUTPUT)"
	@if [ ! -f "$(ROOT)/aegis-remote.vsix" ] && [ -f "$(VSIX_SRC)" ]; then \
		echo "Staging VS Code extension VSIX from $(VSIX_SRC)"; \
		cp "$(VSIX_SRC)" "$(ROOT)/aegis-remote.vsix"; \
	fi
	docker buildx build --no-cache \
		--platform $(PLATFORMS) \
		--tag $(IMG) \
		$(BUILDX_OUTPUT) \
		$(BUILD_ARGS) \
		-f $(CURDIR)/Dockerfile \
		$(ROOT)
	@rm -f "$(ROOT)/aegis-remote.vsix"

.PHONY: docker-load
# Build for host arch and load locally (no push).
docker-load:
	@if [ ! -f "$(ROOT)/aegis-remote.vsix" ] && [ -f "$(VSIX_SRC)" ]; then \
		echo "Staging VS Code extension VSIX from $(VSIX_SRC)"; \
		cp "$(VSIX_SRC)" "$(ROOT)/aegis-remote.vsix"; \
	fi
	@PLATFORM="$$(docker info --format '{{.OSType}}/{{.Architecture}}' 2>/dev/null || echo linux/amd64)"; \
	echo "Building $(IMG) for $$PLATFORM (local load)"; \
	docker buildx build --no-cache \
		--platform "$$PLATFORM" \
		--tag $(IMG) \
		--load \
		-f $(CURDIR)/Dockerfile \
		$(ROOT)
	@rm -f "$(ROOT)/aegis-remote.vsix"
