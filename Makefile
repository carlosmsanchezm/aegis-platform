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
HARDENING ?= dev
DEPLOY_SPOKE ?= false

PF_PLATFORM_HTTP_PORT ?= 10080
PF_PLATFORM_GRPC_PORT ?= 10081
PF_PROXY_HTTP_PORT ?= 10085
PF_KEYCLOAK_HTTPS_PORT ?= 10443

.PHONY: all proto tidy build test verify run-api run-operator stop \
	setup-local deploy-local deploy-local-tls port-forward \
	dev-backstage dev-backstage-cloud dev-backstage-cloud-tls clean-local \
	rerun-preview-failures roll-platform-api

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
	@docker buildx build --no-cache --platform linux/amd64,linux/arm64 \
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
	@$(MAKE) -C services/platform-api docker-build IMG=$(PLATFORM_API_IMAGE) $(if $(BUILDX_OUTPUT),BUILDX_OUTPUT=$(BUILDX_OUTPUT)) $(if $(PLATFORMS),PLATFORMS=$(PLATFORMS))
	@if [ "$(BUILDX_OUTPUT)" != "--push" ]; then $(MAKE) kind-load-platform; fi

.PHONY: kind-load-platform
kind-load-platform:
	@if kind get clusters 2>/dev/null | grep -q desktop; then \
		echo "Loading $(PLATFORM_API_IMAGE) into kind cluster 'desktop'..."; \
		kind load docker-image $(PLATFORM_API_IMAGE) --name desktop; \
		echo "Image loaded into kind cluster"; \
	else \
		echo "No kind cluster named 'desktop' found, skipping kind load"; \
	fi

.PHONY: build-agent
build-agent:
	@echo "Building k8s-agent image $(K8S_AGENT_IMAGE)"
ifeq ($(BUILDX_OUTPUT),--push)
	@cd agents/k8s-agent && docker buildx build --no-cache --platform $(or $(PLATFORMS),linux/amd64) --tag $(K8S_AGENT_IMAGE) --push -f Dockerfile ../..
else
	@$(MAKE) -C agents/k8s-agent docker-build IMG=$(K8S_AGENT_IMAGE)
	@$(MAKE) kind-load-agent
endif

.PHONY: build-agent-multi
build-agent-multi:
	@echo "Building k8s-agent multi-arch image $(K8S_AGENT_IMAGE)"
	@$(MAKE) -C agents/k8s-agent docker-build-multi IMG=$(K8S_AGENT_IMAGE)

.PHONY: kind-load-agent
kind-load-agent:
	@if kind get clusters 2>/dev/null | grep -q desktop; then \
		echo "Loading $(K8S_AGENT_IMAGE) into kind cluster 'desktop'..."; \
		kind load docker-image $(K8S_AGENT_IMAGE) --name desktop; \
		echo "Image loaded into kind cluster"; \
	else \
		echo "No kind cluster named 'desktop' found, skipping kind load"; \
	fi

.PHONY: build-workspace
build-workspace:
	@echo "Building workspace image $(WORKSPACE_IMAGE)"
	@docker buildx build --no-cache --platform linux/amd64,linux/arm64 \
		-t $(WORKSPACE_IMAGE) \
		--push \
		workspace-images/openssh-vscode

.PHONY: build-proxy
build-proxy:
	@echo "Building proxy image $(PROXY_IMAGE)"
	@docker buildx build --no-cache --platform linux/amd64,linux/arm64 \
		-t $(PROXY_IMAGE) \
		--push \
		-f services/proxy/Dockerfile \
		.

.PHONY: build-proxy-local
build-proxy-local:
	@echo "Building proxy image $(PROXY_IMAGE) for local platform"
	@docker build -f services/proxy/Dockerfile -t $(PROXY_IMAGE) .

PULUMI_VERSION ?= 3.226.0

.PHONY: stage-build-deps
stage-build-deps:
	@echo "Staging Iron Bank build dependencies for linux/amd64..."
	@if [ ! -f awscli.zip ]; then \
		echo "  Downloading AWS CLI v2..."; \
		curl -fsSL "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o awscli.zip; \
	fi
	@if [ ! -f kubectl ]; then \
		echo "  Downloading kubectl..."; \
		KUBE_VER=$$(curl -fsSL https://dl.k8s.io/release/stable.txt); \
		curl -fsSL "https://dl.k8s.io/release/$${KUBE_VER}/bin/linux/amd64/kubectl" -o kubectl; \
	fi
	@if [ ! -f jq ]; then \
		echo "  Downloading jq..."; \
		curl -fsSL "https://github.com/jqlang/jq/releases/download/jq-1.7.1/jq-linux-amd64" -o jq; \
	fi
	@if [ ! -f pulumi-linux-x64.tar.gz ]; then \
		echo "  Downloading Pulumi v$(PULUMI_VERSION)..."; \
		curl -fsSL "https://get.pulumi.com/releases/sdk/pulumi-v$(PULUMI_VERSION)-linux-x64.tar.gz" -o pulumi-linux-x64.tar.gz; \
	fi
	@echo "All build dependencies staged."

.PHONY: build-platform-local
build-platform-local: stage-build-deps
	@echo "Building platform-api image $(PLATFORM_API_IMAGE) for local platform (native arch)"
	@$(MAKE) -C services/platform-api docker-load IMG=$(PLATFORM_API_IMAGE)

.PHONY: build-agent-local
build-agent-local:
	@echo "Building k8s-agent image $(K8S_AGENT_IMAGE) for local platform (native arch)"
	@docker build -f agents/k8s-agent/Dockerfile -t $(K8S_AGENT_IMAGE) .

.PHONY: build-local-all
build-local-all: build-platform-local build-proxy-local build-agent-local
	@echo "All local images built. Use 'kubectl rollout restart' to pick up changes."

# Build all 4 images for multi-arch (platform-api + agent loaded into kind; proxy + workspace pushed to DockerHub)
.PHONY: build-images
build-images: build-platform build-agent build-proxy build-workspace

test-workspace:
	@echo "Running workspace connectivity smoke test..."
	@GRPC_ADDR=$${GRPC_ADDR:-localhost:10081} \
	 WORKSPACE_IMAGE=$${WORKSPACE_IMAGE:-aegis-workspace:latest} \
	 CLEANUP=1 \
	 scripts/test-workspace-connection.sh

setup-local:
	@echo "Switching to docker-desktop context..."
	@kubectl config use-context docker-desktop

roll-platform-api:
	@echo "Rolling platform-api deployment to $(PLATFORM_API_IMAGE)"
	@$(MAKE) kind-load-platform
	@kubectl set image deployment/aegis-services-platform-api platform-api=$(PLATFORM_API_IMAGE) -n aegis-system
	@kubectl rollout restart deployment/aegis-services-platform-api -n aegis-system
	@echo "Waiting for rollout to complete..."
	@kubectl rollout status deployment/aegis-services-platform-api -n aegis-system

K8S_AGENT_IMAGE ?= carlosmsanchez/aegis-k8s-agent:dev
PLATFORM_API_IMAGE ?= carlosmsanchez/aegis-platform-api:dev
PROXY_IMAGE ?= carlosmsanchez/aegis-proxy:dev
WORKSPACE_IMAGE ?= carlosmsanchez/aegis-workspace-vscode:latest

.PHONY: rerun-preview-failures
rerun-preview-failures:
	@WORKFLOW_FILE="$(WORKFLOW_FILE)" RUN_ID="$(RUN_ID)" BRANCH="$(BRANCH)" WATCH="$(WATCH)" RUN_SUBSET="$(RUN_SUBSET)" SUITES="$(SUITES)" ./scripts/rerun-preview-failures.sh $(TARGET)

RHBK_USERNAME ?= un1cornsl4yer69
RHBK_PASSWORD ?= P1rac1cab@16love
RHBK_EMAIL ?= aegis@local.test

AWS_ECR_REGISTRY ?= 195714074609.dkr.ecr.us-east-1.amazonaws.com
CLOUD_IMAGE_TAG := $(or $(CLOUD_IMAGE_TAG),$(shell git rev-parse --short HEAD))
CLOUD_PLATFORM_API_IMAGE ?= $(AWS_ECR_REGISTRY)/aegis/platform-api:$(CLOUD_IMAGE_TAG)
CLOUD_PROXY_IMAGE ?= $(AWS_ECR_REGISTRY)/aegis/proxy:$(CLOUD_IMAGE_TAG)
CLOUD_K8S_AGENT_IMAGE ?= $(AWS_ECR_REGISTRY)/aegis/k8s-agent:$(CLOUD_IMAGE_TAG)
CLOUD_WORKSPACE_IMAGE ?= $(AWS_ECR_REGISTRY)/aegis/workspace-vscode:$(CLOUD_IMAGE_TAG)
CLOUD_UI_IMAGE ?= $(AWS_ECR_REGISTRY)/aegis/ui:$(CLOUD_IMAGE_TAG)
AEGIS_UI_DIR ?= $(HOME)/code/aegis-ui

.PHONY: push-cloud-images
push-cloud-images:
	@echo "Building and pushing cloud images with tag $(CLOUD_IMAGE_TAG) to $(AWS_ECR_REGISTRY)"
	@$(MAKE) PLATFORM_API_IMAGE=$(CLOUD_PLATFORM_API_IMAGE) BUILDX_OUTPUT=--push PLATFORMS=linux/amd64 build-platform
	@$(MAKE) PROXY_IMAGE=$(CLOUD_PROXY_IMAGE) build-proxy
	@$(MAKE) K8S_AGENT_IMAGE=$(CLOUD_K8S_AGENT_IMAGE) BUILDX_OUTPUT=--push PLATFORMS=linux/amd64 build-agent
	@$(MAKE) WORKSPACE_IMAGE=$(CLOUD_WORKSPACE_IMAGE) build-workspace
	@$(MAKE) CLOUD_UI_IMAGE=$(CLOUD_UI_IMAGE) AEGIS_UI_DIR=$(AEGIS_UI_DIR) push-ui-cloud
ifeq ($(PUSH_LATEST),1)
	@echo "Promoting images to :latest"
	@docker buildx imagetools create --tag $(AWS_ECR_REGISTRY)/aegis/platform-api:latest $(CLOUD_PLATFORM_API_IMAGE)
	@docker buildx imagetools create --tag $(AWS_ECR_REGISTRY)/aegis/proxy:latest $(CLOUD_PROXY_IMAGE)
	@docker buildx imagetools create --tag $(AWS_ECR_REGISTRY)/aegis/k8s-agent:latest $(CLOUD_K8S_AGENT_IMAGE)
	@docker buildx imagetools create --tag $(AWS_ECR_REGISTRY)/aegis/workspace-vscode:latest $(CLOUD_WORKSPACE_IMAGE)
endif

.PHONY: build-ui-cloud
build-ui-cloud:
	@echo "Building Docker image $(CLOUD_UI_IMAGE) (multi-stage, linux/amd64)..."
	docker build --platform linux/amd64 -t $(CLOUD_UI_IMAGE) -f $(AEGIS_UI_DIR)/packages/backend/Dockerfile.cloud $(AEGIS_UI_DIR)

.PHONY: push-ui-cloud
push-ui-cloud: build-ui-cloud
	@echo "Pushing $(CLOUD_UI_IMAGE) to ECR..."
	docker push $(CLOUD_UI_IMAGE)

.PHONY: clean-webhook
clean-webhook:
	@echo "Checking ingress-nginx admission webhook health..."
	@if kubectl get validatingwebhookconfiguration ingress-nginx-admission >/dev/null 2>&1; then \
		echo "Cleaning up potentially stale webhook configuration..."; \
		kubectl delete validatingwebhookconfiguration ingress-nginx-admission --ignore-not-found >/dev/null 2>&1 || true; \
		sleep 2; \
	fi

.PHONY: deploy-local
deploy-local: setup-local clean-webhook
	@ \
	PLATFORM_API_IMAGE="$(PLATFORM_API_IMAGE)"; \
	if [[ "$$PLATFORM_API_IMAGE" == *":"* ]]; then \
	  PLATFORM_API_REPO="$${PLATFORM_API_IMAGE%:*}"; \
	  PLATFORM_API_TAG="$${PLATFORM_API_IMAGE##*:} "; \
	else \
	  PLATFORM_API_REPO="$$PLATFORM_API_IMAGE"; \
	  PLATFORM_API_TAG="latest"; \
	fi; \
	K8S_AGENT_IMAGE="$(K8S_AGENT_IMAGE)"; \
	if [[ "$$K8S_AGENT_IMAGE" == *":"* ]]; then \
	  K8S_AGENT_REPO="$${K8S_AGENT_IMAGE%:*} "; \
	  K8S_AGENT_TAG="$${K8S_AGENT_IMAGE##*:} "; \
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
		  -f charts/aegis-services/values/common.yaml \
		  -f charts/aegis-services/values/local.yaml \
		  --set platformApi.image.repository=$$PLATFORM_API_REPO \
		  --set platformApi.image.tag=$$PLATFORM_API_TAG \
		  --namespace aegis-system --create-namespace \
		  --wait --timeout 5m; \
		if [ "$(DEPLOY_SPOKE)" = "true" ]; then \
		  echo "Deploying Aegis spoke locally (no TLS)..."; \
		  helm upgrade --install aegis-spoke charts/aegis-spoke \
		    -f charts/aegis-spoke/values.yaml \
		    -f charts/aegis-spoke/values-local.yaml \
		    --set k8sAgent.image.repository=$$K8S_AGENT_REPO \
		    --set k8sAgent.image.tag=$$K8S_AGENT_TAG \
		    --set k8sAgent.image.pullPolicy=Always \
		    --namespace aegis-system --create-namespace; \
		  echo "✅ Deployed local stack without TLS (with spoke)"; \
		else \
		  echo "⏭️  Skipping aegis-spoke (DEPLOY_SPOKE=false). Set DEPLOY_SPOKE=true to include it."; \
		  echo "✅ Deployed local hub stack without TLS"; \
		fi; \
		echo "   Platform API gRPC: platform-api-grpc.localtest.me:80"; \
		echo "   Proxy: http://proxy.localtest.me"

.PHONY: deploy-local-tls
deploy-local-tls: setup-local clean-webhook
	PLATFORM_API_IMAGE="$(PLATFORM_API_IMAGE)" \
	K8S_AGENT_IMAGE="$(K8S_AGENT_IMAGE)" \
	RHBK_USERNAME="$(RHBK_USERNAME)" \
	RHBK_PASSWORD="$(RHBK_PASSWORD)" \
	RHBK_EMAIL="$(RHBK_EMAIL)" \
	HARDENING="$(HARDENING)" \
	DEPLOY_SPOKE="$(DEPLOY_SPOKE)" \
	./scripts/aegis.sh deploy --hardening $(HARDENING)


.PHONY: sync-certs
sync-certs:
	@echo "Syncing internal CA bundle to $$HOME/aegis-platform-api-ca.crt ..."; 
	kubectl get secret aegis-trust-bundle -n aegis-system -o "jsonpath={.data.ca\.crt}" | base64 --decode > "$$HOME/aegis-platform-api-ca.crt"; 
	chmod 0644 "$$HOME/aegis-platform-api-ca.crt"; 
	echo "Syncing internal CA bundle to $$HOME/keycloak.localtest.me.crt ..."; 
	kubectl get secret aegis-trust-bundle -n keycloak -o "jsonpath={.data.ca\.crt}" | base64 --decode > "$$HOME/keycloak.localtest.me.crt"; 
	chmod 0644 "$$HOME/keycloak.localtest.me.crt"; 
	cp "$$HOME/aegis-platform-api-ca.crt" "$$HOME/aegis-local-trust.pem"; 
	chmod 0644 "$$HOME/aegis-local-trust.pem"; 
	echo "   CA bundles refreshed."; 
	echo "   Combined trust store: $$HOME/aegis-local-trust.pem";

port-forward:
	@echo "Stopping any existing port-forwards..."
	@pkill -f "kubectl.*port-forward" 2>/dev/null || true
	@sleep 2
	@echo "Setting up port-forwarding..."
	@# Use nohup to keep processes running after make exits
	@nohup kubectl --context docker-desktop -n aegis-system port-forward svc/aegis-services-platform-api $(PF_PLATFORM_HTTP_PORT):8080 $(PF_PLATFORM_GRPC_PORT):8081 >/dev/null 2>&1 &
	@nohup kubectl --context docker-desktop -n aegis-system port-forward svc/aegis-services-proxy $(PF_PROXY_HTTP_PORT):8085 >/dev/null 2>&1 &
	@nohup kubectl --context docker-desktop -n keycloak port-forward svc/aegis-services-keycloak-service $(PF_KEYCLOAK_HTTPS_PORT):8443 >/dev/null 2>&1 &
	@# AWS relay tunnel port-forwards (for remote spoke clusters to reach local platform-api)
	@nohup kubectl --context docker-desktop -n aegis-system port-forward svc/aegis-services-platform-api 8081:8081 >/dev/null 2>&1 &
	@nohup kubectl --context docker-desktop -n keycloak port-forward svc/aegis-services-keycloak-service 8443:8443 >/dev/null 2>&1 &
	@nohup kubectl --context docker-desktop -n aegis-pki port-forward svc/step-certificates 9443:443 >/dev/null 2>&1 &
	@sleep 3
	@# Verify port-forwards are running
	@if ! pgrep -f "kubectl.*port-forward.*aegis-services-platform-api.*$(PF_PLATFORM_HTTP_PORT)" >/dev/null; then \
		echo "ERROR: Platform API port-forward failed to start. Check if pods are running:"; \
		echo "  kubectl --context docker-desktop -n aegis-system get pods"; \
		exit 1; \
	fi
	@echo "Port-forwarding started:"
	@echo "  Platform API: http://localhost:$(PF_PLATFORM_HTTP_PORT) (HTTP) / localhost:$(PF_PLATFORM_GRPC_PORT) (gRPC)"
	@echo "  Proxy: http://localhost:$(PF_PROXY_HTTP_PORT)"
	@echo "  Keycloak: https://localhost:$(PF_KEYCLOAK_HTTPS_PORT)"
	@echo "  AWS Relay: localhost:8081 (gRPC) / localhost:8443 (Keycloak) / localhost:9443 (step-ca)"
	@echo ""
	@# Quick connectivity test
	@if curl -s --max-time 3 http://localhost:$(PF_PLATFORM_HTTP_PORT)/healthz >/dev/null 2>&1; then \
		echo "Connectivity verified: Platform API is reachable"; \
	else \
		echo "Warning: Platform API health check failed (may still be starting)"; \
	fi
	@echo ""
	@echo "Use 'make stop-port-forward' or 'pkill -f \"kubectl.*port-forward\"' to stop all port-forwards."

.PHONY: stop-port-forward
stop-port-forward:
	@echo "Stopping all port-forwards..."
	@pkill -f "kubectl.*port-forward" 2>/dev/null || true
	@echo "All port-forwards stopped."

dev-backstage:
	@echo "Starting Backstage development server (local mode)..."
	@echo "   Backend: https://platform-api.localtest.me (ingress, no port-forward required)"
	@cd aegis-platform && NODE_EXTRA_CA_CERTS="$${NODE_EXTRA_CA_CERTS:-\\$HOME/aegis-local-trust.pem}" yarn dev

dev-backstage-cloud:
	@echo "Starting Backstage development server (cloud mode)..."
	@echo "   Backend: http://platform-api.aegis-platform.tech:8080"
	@cd aegis-platform && yarn dev:cloud

dev-backstage-cloud-tls:
	@echo "Starting Backstage development server (cloud TLS mode)..."
	@echo "   Backend: http://platform-api.aegis-platform.tech:8080"
	@cd aegis-platform && yarn dev:cloud-tls

clean-local:
	./scripts/aegis.sh clean

AWS_PROFILE ?= aegis-new
AWS_REGION ?= us-east-1
SKIP_DNS_UPDATE ?= 0

.PHONY: ecr-login
ecr-login:
	@echo "Logging into ECR..."
	aws ecr get-login-password --region $(AWS_REGION) --profile $(AWS_PROFILE) | \
		docker login --username AWS --password-stdin $(AWS_ECR_REGISTRY)

.PHONY: deploy-cloud
deploy-cloud:
	@echo "Applying Terraform (idempotent)..."
	cd terraform && AWS_PROFILE=$(AWS_PROFILE) terraform apply -auto-approve
	@echo "Deploying Aegis hub to cloud EKS..."
	SKIP_MIGRATION_PLACEHOLDER=1 SKIP_DNS_UPDATE=$(SKIP_DNS_UPDATE) \
		PLATFORM_API_IMAGE_TAG=$(CLOUD_PLATFORM_API_IMAGE) \
		PROXY_IMAGE_TAG=$(CLOUD_PROXY_IMAGE) \
		K8S_AGENT_IMAGE_TAG=$(CLOUD_K8S_AGENT_IMAGE) \
		UI_IMAGE_TAG=$(CLOUD_UI_IMAGE) \
		./terraform/generate-cloud-deployment.sh --non-interactive
