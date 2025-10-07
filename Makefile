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

.PHONY: all proto tidy build test verify run-api run-operator stop \
	setup-local deploy-local port-forward dev-backstage clean-local

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

setup-local:
	@echo "Switching to docker-desktop context..."
	@kubectl config use-context docker-desktop

deploy-local: setup-local
	@echo "Deploying Aegis services locally..."
	@helm upgrade --install aegis-services charts/aegis-services \
	  -f charts/aegis-services/values/common.yaml \
	  -f charts/aegis-services/values/local.yaml \
	  --namespace aegis-system --create-namespace
	@helm upgrade --install aegis-spoke charts/aegis-spoke \
	  -f charts/aegis-spoke/values.yaml \
	  -f charts/aegis-spoke/values-local.yaml \
	  --namespace aegis-system --create-namespace

port-forward:
	@echo "Stopping any existing port-forwards..."
	@while pgrep -f "kubectl port-forward .*aegis-system" >/dev/null; do \
		pkill -f "kubectl port-forward .*aegis-system" || true; \
		sleep 1; \
	done
	@echo "Setting up port-forwarding..."
	@kubectl -n aegis-system port-forward svc/aegis-services-platform-api 10080:8080 10081:8081 &
	@kubectl -n aegis-system port-forward svc/aegis-services-proxy 10085:8085 &
	@echo "Port-forwarding started. Platform API on 10080/10081, proxy on 10085. Use 'pkill -f \"kubectl port-forward\"' to stop."

dev-backstage:
	@echo "Starting Backstage development server..."
	@cd aegis-platform && yarn dev

clean-local:
	@echo "Uninstalling Aegis services..."
	@helm uninstall aegis-services aegis-spoke -n aegis-system || true
