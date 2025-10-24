SHELL := /bin/bash
PROTO_DIR := proto
PROTO_FILES := aegis/v1/platform.proto
API_MOD := ./services/platform-api
AGENT_MOD := ./agents/k8s-agent
BIN_DIR := $(abspath ./bin)
BUF_CACHE_DIR := $(abspath ./.bufcache)
GO_CACHE_DIR := $(abspath ./.gocache)
GO_MOD_CACHE := $(abspath ./.gomodcache)
ALLOW_NET ?= 0
ALLOW_SOCKETS ?= 0
GRPC_ADDR ?= :8081
HTTP_ADDR ?= :8080
AEGIS_CP_GRPC ?= localhost:8081
AEGIS_CLUSTER_ID ?= dev-1
AEGIS_REGION ?= us-local
AEGIS_PROVIDER ?= DEV
AEGIS_DISABLE_KUEUE ?= 1
HEALTH_PROBE_BIND_ADDRESS ?= :8081
AEGIS_FLAVORS ?=

PF_PLATFORM_HTTP_PORT ?= 10080
PF_PLATFORM_GRPC_PORT ?= 10081
PF_PROXY_HTTP_PORT ?= 10085

.PHONY: all proto tidy build test verify run-api run-operator stop \
	setup-local deploy-local deploy-local-tls port-forward \
	dev-backstage dev-backstage-cloud dev-backstage-cloud-tls clean-local

all: proto tidy build

proto:
	@echo ">> buf generate"
	@mkdir -p $(BUF_CACHE_DIR)
	@cd $(PROTO_DIR) && PATH=$(BIN_DIR):$$PATH BUF_CACHE_DIR=$(BUF_CACHE_DIR) buf generate --path $(PROTO_FILES)

tidy:
ifeq ($(ALLOW_NET),1)
	@mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE)
	@cd $(API_MOD) && GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE) go mod tidy
	@cd $(AGENT_MOD) && GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE) go mod tidy
else
	@echo "ALLOW_NET=0: skipping tidy"
endif

build:
ifeq ($(ALLOW_NET),1)
	@mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE)
	@cd $(API_MOD) && GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE) go build ./...
	@cd $(AGENT_MOD) && GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE) go build ./...
else
	@echo "ALLOW_NET=0: skipping build"
endif

test:
ifeq ($(ALLOW_NET),1)
	@mkdir -p $(GO_CACHE_DIR) $(GO_MOD_CACHE)
	@cd $(API_MOD) && GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE) go test ./...
	@cd $(AGENT_MOD) && GOCACHE=$(GO_CACHE_DIR) GOMODCACHE=$(GO_MOD_CACHE) go test ./...
else
	@echo "ALLOW_NET=0: skipping test"
endif

verify:
	@PATH="$(BIN_DIR):$$PATH" BUF_CACHE_DIR="$(BUF_CACHE_DIR)" $(MAKE) proto
	@$(MAKE) tidy ALLOW_NET=1
	@$(MAKE) build ALLOW_NET=1
	@$(MAKE) test ALLOW_NET=1
	@(cd $(API_MOD) && go list ./...)
	@(cd $(AGENT_MOD) && go list ./...)

run-api:
ifeq ($(ALLOW_SOCKETS),1)
	@cd $(API_MOD) && GRPC_ADDR=$(GRPC_ADDR) HTTP_ADDR=$(HTTP_ADDR) go run ./...
else
	@echo "ALLOW_SOCKETS=0: disabled here"
endif

run-operator:
ifeq ($(ALLOW_SOCKETS),1)
	@cd $(AGENT_MOD) && \
	AEGIS_CP_GRPC=$(AEGIS_CP_GRPC) \
	AEGIS_CLUSTER_ID=$(AEGIS_CLUSTER_ID) \
	AEGIS_REGION=$(AEGIS_REGION) \
	AEGIS_PROVIDER=$(AEGIS_PROVIDER) \
	AEGIS_DISABLE_KUEUE=$(AEGIS_DISABLE_KUEUE) \
	HEALTH_PROBE_BIND_ADDRESS=$(HEALTH_PROBE_BIND_ADDRESS) \
	AEGIS_FLAVORS="$(AEGIS_FLAVORS)" \
	KUBECONFIG=$(KUBECONFIG) \
	go run ./cmd
else
	@echo "ALLOW_SOCKETS=0: disabled here"
endif

stop:
	@pkill -f services/platform-api || true
	@pkill -f "k8s-agent/cmd" || true

.PHONY: platform-api-docker-build
platform-api-docker-build:
	@echo "Building multi-arch platform-api image $(PLATFORM_API_IMAGE)"
	@docker buildx build --platform linux/amd64,linux/arm64 \
		-f services/platform-api/Dockerfile \
		-t $(PLATFORM_API_IMAGE) \
		--push \
		.

.PHONY: platform-api-docker-push
platform-api-docker-push:
	@echo "platform-api-docker-build already pushes the manifest (no-op)"

.PHONY: build-platform
build-platform:
	@echo "Building platform-api image $(PLATFORM_API_IMAGE)"
	@$(MAKE) -C services/platform-api docker-build IMG=$(PLATFORM_API_IMAGE)

.PHONY: build-agent
build-agent:
	@echo "Building k8s-agent image $(K8S_AGENT_IMAGE)"
	@$(MAKE) -C agents/k8s-agent docker-build-multi IMG=$(K8S_AGENT_IMAGE)

.PHONY: build-workspace
build-workspace:
	@echo "Building workspace image $(WORKSPACE_IMAGE)"
	@docker buildx build --platform linux/amd64,linux/arm64 \
		-t $(WORKSPACE_IMAGE) \
		--push \
		workspace-images/openssh-vscode

.PHONY: build-images
build-images: build-platform build-agent build-workspace

test-workspace:
	@echo "Running workspace connectivity smoke test..."
	@GRPC_ADDR=$${GRPC_ADDR:-localhost:10081} \
	 WORKSPACE_IMAGE=$${WORKSPACE_IMAGE:-aegis-workspace:latest} \
	 CLEANUP=1 \
	 scripts/test-workspace-connection.sh

setup-local:
	@echo "Switching to docker-desktop context..."
	@kubectl config use-context docker-desktop

K8S_AGENT_IMAGE ?= carlosmsanchez/aegis-k8s-agent:dev
PLATFORM_API_IMAGE ?= carlosmsanchez/aegis-platform-api:dev
WORKSPACE_IMAGE ?= carlosmsanchez/aegis-workspace-vscode:latest

AWS_ECR_REGISTRY ?= 567751785679.dkr.ecr.us-east-1.amazonaws.com
CLOUD_IMAGE_TAG ?= $(shell git rev-parse --short HEAD)
CLOUD_PLATFORM_API_IMAGE ?= $(AWS_ECR_REGISTRY)/aegis/platform-api:$(CLOUD_IMAGE_TAG)
CLOUD_K8S_AGENT_IMAGE ?= $(AWS_ECR_REGISTRY)/aegis/k8s-agent:$(CLOUD_IMAGE_TAG)
CLOUD_WORKSPACE_IMAGE ?= $(AWS_ECR_REGISTRY)/aegis/workspace-vscode:$(CLOUD_IMAGE_TAG)

.PHONY: push-cloud-images
push-cloud-images:
	@echo "Building and pushing cloud images with tag $(CLOUD_IMAGE_TAG) to $(AWS_ECR_REGISTRY)"
	@$(MAKE) PLATFORM_API_IMAGE=$(CLOUD_PLATFORM_API_IMAGE) build-platform
	@$(MAKE) K8S_AGENT_IMAGE=$(CLOUD_K8S_AGENT_IMAGE) build-agent
	@$(MAKE) WORKSPACE_IMAGE=$(CLOUD_WORKSPACE_IMAGE) build-workspace
ifeq ($(PUSH_LATEST),1)
	@echo "Promoting images to :latest"
	@docker buildx imagetools create --tag $(AWS_ECR_REGISTRY)/aegis/platform-api:latest $(CLOUD_PLATFORM_API_IMAGE)
	@docker buildx imagetools create --tag $(AWS_ECR_REGISTRY)/aegis/k8s-agent:latest $(CLOUD_K8S_AGENT_IMAGE)
	@docker buildx imagetools create --tag $(AWS_ECR_REGISTRY)/aegis/workspace-vscode:latest $(CLOUD_WORKSPACE_IMAGE)
endif

.PHONY: deploy-local
.PHONY: deploy-local
deploy-local: setup-local
<<<<<<< HEAD
	@echo "Ensuring chart dependencies (ingress-nginx) are up to date..."
	@helm dependency update charts/aegis-services >/dev/null
	@./scripts/update-local-hosts.sh
	@echo "Deploying Aegis services locally (no TLS)..."
	@helm upgrade --install aegis-services charts/aegis-services \
=======
	@ \
	PLATFORM_API_IMAGE="$(PLATFORM_API_IMAGE)"; \
	if [[ "$$PLATFORM_API_IMAGE" == *":"* ]]; then \
	  PLATFORM_API_REPO="$${PLATFORM_API_IMAGE%:*}"; \
	  PLATFORM_API_TAG="$${PLATFORM_API_IMAGE##*:}"; \
	else \
	  PLATFORM_API_REPO="$$PLATFORM_API_IMAGE"; \
	  PLATFORM_API_TAG="latest"; \
	fi; \
	K8S_AGENT_IMAGE="$(K8S_AGENT_IMAGE)"; \
	if [[ "$$K8S_AGENT_IMAGE" == *":"* ]]; then \
	  K8S_AGENT_REPO="$${K8S_AGENT_IMAGE%:*}"; \
	  K8S_AGENT_TAG="$${K8S_AGENT_IMAGE##*:}"; \
	else \
	  K8S_AGENT_REPO="$$K8S_AGENT_IMAGE"; \
	  K8S_AGENT_TAG="latest"; \
	fi; \
	echo "Using k8s-agent image $$K8S_AGENT_IMAGE"; \
	echo "Using platform-api image $$PLATFORM_API_IMAGE"; \
	echo "Ensuring chart dependencies (ingress-nginx) are up to date..."; \
	helm dependency update charts/aegis-services >/dev/null; \
	echo "Deploying Aegis services locally (no TLS)..."; \
	helm upgrade --install aegis-services charts/aegis-services \
>>>>>>> origin/main
	  -f charts/aegis-services/values/common.yaml \
	  -f charts/aegis-services/values/local.yaml \
	  --set platformApi.image.repository=$$PLATFORM_API_REPO \
	  --set platformApi.image.tag=$$PLATFORM_API_TAG \
	  --namespace aegis-system --create-namespace \
	  --wait --timeout 5m; \
	echo "Deploying Aegis spoke locally (no TLS)..."; \
	helm upgrade --install aegis-spoke charts/aegis-spoke \
	  -f charts/aegis-spoke/values.yaml \
	  -f charts/aegis-spoke/values-local.yaml \
	  --set k8sAgent.image.repository=$$K8S_AGENT_REPO \
	  --set k8sAgent.image.tag=$$K8S_AGENT_TAG \
	  --set k8sAgent.image.pullPolicy=Always \
	  --namespace aegis-system --create-namespace; \
	echo "✅ Deployed local stack without TLS"; \
	echo "   Platform API gRPC: platform-api-grpc.localtest.me:80"; \
	echo "   Proxy: http://proxy.localtest.me"

.PHONY: deploy-local-tls
.PHONY: deploy-local-tls
deploy-local-tls: setup-local
<<<<<<< HEAD
	@echo "Ensuring chart dependencies (ingress-nginx) are up to date..."
	@helm dependency update charts/aegis-services >/dev/null
	@./scripts/update-local-hosts.sh
	@echo "Deploying Aegis services locally with TLS..."
	@helm upgrade --install aegis-services charts/aegis-services \
=======
	@ \
	PLATFORM_API_IMAGE="$(PLATFORM_API_IMAGE)"; \
	if [[ "$$PLATFORM_API_IMAGE" == *":"* ]]; then \
	  PLATFORM_API_REPO="$${PLATFORM_API_IMAGE%:*}"; \
	  PLATFORM_API_TAG="$${PLATFORM_API_IMAGE##*:}"; \
	else \
	  PLATFORM_API_REPO="$$PLATFORM_API_IMAGE"; \
	  PLATFORM_API_TAG="latest"; \
	fi; \
	K8S_AGENT_IMAGE="$(K8S_AGENT_IMAGE)"; \
	if [[ "$$K8S_AGENT_IMAGE" == *":"* ]]; then \
	  K8S_AGENT_REPO="$${K8S_AGENT_IMAGE%:*}"; \
	  K8S_AGENT_TAG="$${K8S_AGENT_IMAGE##*:}"; \
	else \
	  K8S_AGENT_REPO="$$K8S_AGENT_IMAGE"; \
	  K8S_AGENT_TAG="latest"; \
	fi; \
	echo "Using k8s-agent image $$K8S_AGENT_IMAGE"; \
	echo "Using platform-api image $$PLATFORM_API_IMAGE"; \
	echo "Ensuring chart dependencies (ingress-nginx) are up to date..."; \
	helm dependency update charts/aegis-services >/dev/null; \
	echo "Deploying Aegis services locally with TLS..."; \
	helm upgrade --install aegis-services charts/aegis-services \
>>>>>>> origin/main
	  -f charts/aegis-services/values/common.yaml \
	  -f charts/aegis-services/values/local.yaml \
	  -f charts/aegis-services/values/local-tls.yaml \
	  --set platformApi.image.repository=$$PLATFORM_API_REPO \
	  --set platformApi.image.tag=$$PLATFORM_API_TAG \
	  --namespace aegis-system --create-namespace \
	  --wait --timeout 5m; \
	echo "Deploying Aegis spoke locally with TLS..."; \
	helm upgrade --install aegis-spoke charts/aegis-spoke \
	  -f charts/aegis-spoke/values.yaml \
	  -f charts/aegis-spoke/values-local.yaml \
	  -f charts/aegis-spoke/values-local-tls.yaml \
	  --set k8sAgent.image.repository=$$K8S_AGENT_REPO \
	  --set k8sAgent.image.tag=$$K8S_AGENT_TAG \
	  --set k8sAgent.image.pullPolicy=Always \
	  --namespace aegis-system --create-namespace; \
	echo "Syncing platform API certificate to $$HOME/aegis-platform-api-ca.crt ..."; \
	kubectl get secret aegis-services-platform-api-tls -n aegis-system -o "jsonpath={.data.tls\.crt}" | base64 --decode > "$$HOME/aegis-platform-api-ca.crt"; \
	chmod 0644 "$$HOME/aegis-platform-api-ca.crt"; \
	echo "   CA bundle refreshed."; \
	echo "   To trust it system-wide: sudo security add-trust -d -r trustRoot -k /Library/Keychains/System.keychain $$HOME/aegis-platform-api-ca.crt"; \
	echo "   Launch VS Code with TLS trust:"; \
	echo "     NODE_EXTRA_CA_CERTS=$$HOME/aegis-platform-api-ca.crt \"; \
	echo "       /Applications/Visual\\ Studio\\ Code.app/Contents/MacOS/Electron --enable-proposed-api aegis.aegis-remote $$PWD"; \
	echo "✅ Deployed with TLS using self-signed certificates"; \
	echo "   Platform API gRPC: platform-api-grpc.localtest.me:443"; \
	echo "   Proxy: https://proxy.localtest.me"; \
	echo ""; \
	echo "   For E2E tests with TLS:"; \
	echo "   export GRPC_TLS=1"; \
	echo "   export GRPC_TLS_SKIP_VERIFY=1  # Self-signed certs"; \
	echo "   ./scripts/e2e-platform-api.sh"


port-forward:
	@echo "Stopping any existing port-forwards..."
	@while pgrep -f "kubectl port-forward .*aegis-system" >/dev/null; do \
		pkill -f "kubectl port-forward .*aegis-system" || true; \
		sleep 1; \
	done
	@echo "Setting up port-forwarding..."
	@kubectl -n aegis-system port-forward svc/aegis-services-platform-api $(PF_PLATFORM_HTTP_PORT):8080 $(PF_PLATFORM_GRPC_PORT):8081 &
	@kubectl -n aegis-system port-forward svc/aegis-services-proxy $(PF_PROXY_HTTP_PORT):8085 &
	@echo "Port-forwarding started. Platform API on $(PF_PLATFORM_HTTP_PORT)/$(PF_PLATFORM_GRPC_PORT), proxy on $(PF_PROXY_HTTP_PORT). Use 'pkill -f \"kubectl port-forward\"' to stop."

dev-backstage:
	@echo "Starting Backstage development server (local mode)..."
	@echo "   Backend: http://platform-api.localtest.me (ingress, no port-forward required)"
	@cd aegis-platform && yarn dev

dev-backstage-cloud:
	@echo "Starting Backstage development server (cloud mode)..."
	@echo "   Backend: http://platform-api.aegist.dev:8080"
	@cd aegis-platform && yarn dev:cloud

dev-backstage-cloud-tls:
	@echo "Starting Backstage development server (cloud TLS mode)..."
	@echo "   Backend: http://platform-api.aegist.dev:8080"
	@cd aegis-platform && yarn dev:cloud-tls

clean-local:
	@for release in aegis-services aegis-spoke; do \
		if helm status $$release -n aegis-system >/dev/null 2>&1; then \
			echo "Uninstalling $$release..."; \
			helm uninstall $$release -n aegis-system >/dev/null; \
		else \
			echo "Skipping $$release (not installed)"; \
		fi; \
	done
