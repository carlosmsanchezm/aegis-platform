# Convenience tooling for Aegis platform-api service builds.

IMG ?= carlosmsanchez/aegis-platform-api:dev
PLATFORMS ?= linux/amd64,linux/arm64
BUILDX_OUTPUT ?= --push
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

.PHONY: docker-build docker-buildx
# Multi-arch build that pushes to the registry defined by IMG.
docker-build docker-buildx:
	@echo "Building $(IMG) for $(PLATFORMS)"
	docker buildx build \
		--platform $(PLATFORMS) \
		--tag $(IMG) \
		$(BUILDX_OUTPUT) \
		$(BUILD_ARGS) \
		-f $(CURDIR)/Dockerfile \
		$(ROOT)

.PHONY: docker-load
# Multi-arch build but keep the image locally (no push).
docker-load:
	@PLATFORM="$$(docker info --format '{{.OSType}}/{{.Architecture}}' 2>/dev/null || echo linux/amd64)"; \
	echo "Building $(IMG) for $$PLATFORM (local load)"; \
	docker buildx build \
		--platform "$$PLATFORM" \
		--tag $(IMG) \
		--load \
		-f $(CURDIR)/Dockerfile \
		$(ROOT)
