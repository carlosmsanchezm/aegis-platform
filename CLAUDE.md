# Aegis Platform -- AI Agent Guide

> **How to use this file**: `CLAUDE.md` is the **entry point** for all AI agents. It is loaded
> automatically every session. It contains critical rules inline (image registry, commands,
> architecture facts) and routes to deeper docs via the "Documentation Map" table below.
> Never duplicate content that belongs in a specialized doc -- instead, add a row to the table.

## Doc Ownership

**Authoritative for:** AI agent entrypoint rules, documentation routing, image registry decisions.

**Not authoritative for:** deployment workflow, production rollout steps, command sequencing.

**See also:**
- `AGENT_DEPLOYMENT_GUIDE.md` -- authoritative deployment workflow for local, hybrid, and full cloud
- `docs/PRODUCTION_DEPLOYMENT.md` -- production-only deltas (RDS, HA, rollback posture)
- `terraform/README.md` -- infrastructure scope and outputs
- `docs/make-commands.md` -- make target contracts

## System Overview

Hub-and-spoke Kubernetes platform for multi-cluster workload scheduling.
Hub runtime spans `aegis-system` (platform-api, proxy, backstage, ingress) and `keycloak` (Keycloak).
Spoke (k8s-agent, spoke-proxy) on remote EKS or co-located on docker-desktop.

All work happens from `~/code/`. The primary repo is `aegis-platform/`.

## Repository Map

| Directory | Purpose |
|-----------|---------|
| `aegis-platform/` | Primary repo: Go services, Helm charts, Terraform, scripts, deployment guides |
| `aegis-ui/` | UI source repo used to build the cloud UI image; deployment/runtime wiring lives in `aegis-platform/` |
| `aegis-compliance-evidence/` | FedRAMP/NIST compliance evidence vault |
| `aegis-vscode-customization/` | VS Code extension for remote workspaces |

## Documentation Map -- Read What You Need

All paths below are relative to `aegis-platform/` unless noted otherwise.

**Always read first**: `AGENTS.md` (operating rules), `AGENT_DEPLOYMENT_GUIDE.md` (deployment — all modes)

**By task type**:

| Task | Read these docs |
|------|----------------|
| Deployment (all modes) | `AGENT_DEPLOYMENT_GUIDE.md` |
| Deployment (cloud/prod with RDS) | `AGENT_DEPLOYMENT_GUIDE.md` § Cloud, `docs/PRODUCTION_DEPLOYMENT.md` |
| Cloud hub deployment | `AGENT_DEPLOYMENT_GUIDE.md` § Cloud, `terraform/README.md` |
| Helm chart configuration | `charts/README.md` |
| Image registry / which image to use | This file (Image Registry section below), `AGENT_DEPLOYMENT_GUIDE.md` |
| Networking / architecture | `docs/networking-architecture.md`, `docs/infrastructure-reference.md` |
| TLS / PKI / certs | `docs/keycloak-auth-readme.md` |
| Auth / OIDC / Keycloak | `docs/security/auth.md`, `docs/keycloak-auth-readme.md` |
| Spoke provisioning | `charts/aegis-spoke/README.md`, `docs/spoke-provisioning-reference.md` |
| Spoke proxy / NLB | `docs/spoke-proxy-automation.md` |
| Hub chart config | `charts/aegis-services/README.md` |
| CRDs / controller | `docs/k8s-crd-and-controller-reference.md` |
| Testing | `docs/testing-guide.md` |
| Makefile targets | `docs/make-commands.md` |
| gRPC debugging | `docs/grpcurl-cheatsheet.md` |
| VS Code extension | `BACKSTAGE-VSCODE-INTEGRATION.md`, `VSCODE-SSH-SETUP.md` |
| Teardown | `terraform/DESTROY_CHECKLIST.md` |
| Terraform / infra | `terraform/README.md` |
| Pulumi provisioning | `docs/pulumi-preview.md` |
| Spoke connectivity / heartbeat / gRPC errors | `docs/aws-tunnel-dev-setup.md`, `docs/networking-architecture.md` |
| Tunnels (AWS NLB relay for gRPC + OIDC) | `docs/aws-tunnel-dev-setup.md`, `CLOUD-PROXY-SETUP.md` |
| Compliance evidence | `aegis-compliance-evidence/README.md` (external repo) |
| Preview CI workflow | `docs/preview-workflow.md` |
| Workspace images | `docs/workspace-images.md` |
| Workspace storage & custom images | `docs/workspace-images.md` (Persistent Storage + Custom Workspace Images sections) |
| GPT-5 advisor | `docs/gpt5_advisor.md` |
| UI frontend source | `aegis-ui/README.md` (external repo) |
| Cloud UI auth/runtime | `docs/security/auth.md`, `AGENT_DEPLOYMENT_GUIDE.md` |
| Architecture overview | `README.md` |
| Local auth / cluster launch | `docs/local-auth-and-cluster-launch-runbook.md` |
| Troubleshooting | `docs/TROUBLESHOOTING-CLUSTER-DROPDOWN.md` |
| Simple mode | `~/code/AEGIS_SIMPLE_V1_EXECUTION_RUNBOOK_2026-02-25.md`, `~/code/AEGIS_SIMPLE_V1_AWS_EXISTING_CLUSTER_PLAYBOOK_2026-02-25.md` |
| Dev mode vs production | `docs/dev-mode-vs-production.md` |

## Critical: After Rebuilding/Restarting Platform-API

When you restart the platform-api pod (e.g., after `make build-platform-local` + `kubectl rollout restart`), you **must re-establish port-forwards** or the EKS spoke will lose connectivity and clusters will go unhealthy:

```bash
# Re-check and restart port-forwards after any pod restart
lsof -i :8081 || kubectl -n aegis-system port-forward svc/aegis-services-platform-api 8081:8081 &
lsof -i :8443 || kubectl -n keycloak port-forward svc/aegis-services-keycloak 8443:8443 &
lsof -i :10080 || kubectl -n aegis-system port-forward svc/aegis-services-platform-api 10080:8080 10081:8081 &
```

See `docs/aws-tunnel-dev-setup.md` → "Cluster shows unhealthy after rebuilding/restarting platform-api" for full explanation.

## Key Architecture Facts

- PKI is ALWAYS external (installed by `scripts/install-internal-pki.sh`)
- Helm chart consumes PKI via `pki-secrets-init` post-install hook
- Namespaces: aegis-system (hub), aegis-pki (step-ca), cert-manager, keycloak
- step-ca at aegis-pki namespace
- Trust bundle: `aegis-trust-bundle` secret in aegis-system + keycloak
- Spoke-proxy CA: independent self-signed CA, auto-generated by Helm hook
- cert-manager, step-ca, step-issuer are NOT subcharts; installed externally
- Hybrid dev (local hub + cloud spoke): both gRPC and OIDC go through AWS NLB relay (SSH tunnel). See `docs/aws-tunnel-dev-setup.md`
- Hybrid dev stays on the local/tunnel workflow; do not rewrite local configs to use the public cloud hostnames
- Full-cloud public access is Cloudflare-only on `ui.aegis-platform.tech`, `platform-api.aegis-platform.tech`, `proxy.aegis-platform.tech`, and `keycloak.aegis-platform.tech`
- In full cloud, UI and Keycloak are served through the shared public ingress; platform-api and proxy stay on direct public service endpoints
- The external `aegis-ui/` repo is the image source, not the source of truth for the live cloud auth/runtime contract
- Hub cluster default: `aegis-hub-prod` (configurable via `cluster_name_prefix` terraform variable)

## Project Structure

```
~/code/aegis-platform/              # Primary repo
  charts/
    aegis-services/                 #   Hub chart (platform-api, proxy, keycloak, ingress)
    aegis-spoke/                    #   Spoke chart (k8s-agent, spoke-proxy)
  services/
    platform-api/                   #   Go gRPC+HTTP API server
    proxy/                          #   Reverse proxy for workspace access
  agents/
    k8s-agent/                      #   Kubernetes operator (controller-runtime)
  scripts/                          #   Deployment and test scripts
  terraform/                        #   AWS infrastructure (EKS, VPC, ECR)
  docs/                             #   Architecture and operational docs
```

## Image Registry -- Decision Rules for AI Agents

**The one rule**: If the target Kubernetes cluster is **EKS (cloud)**, use **ECR**. If it is **Docker Desktop (local)**, use **DockerHub**.

### Registry Decision Matrix

| Context | Registry | Tag | Pull Policy |
|---------|----------|-----|-------------|
| Local dev (Docker Desktop) | `carlosmsanchez/aegis-*` (DockerHub) | `dev` | `IfNotPresent` |
| Cloud hub (EKS aegis-services) | `195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/*` (ECR) | git SHA | `Always` |
| Cloud spoke (EKS aegis-spoke) | ECR (same) | git SHA | `Always` |
| Workspace pod (AEGIS_DEFAULT_IMAGE) | ECR in cloud, DockerHub locally | `latest` | `Always` (cloud) |

### Image Names

| Image | DockerHub (local) | ECR (cloud) |
|-------|-------------------|-------------|
| platform-api | `carlosmsanchez/aegis-platform-api` | `195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/platform-api` |
| proxy | `carlosmsanchez/aegis-proxy` | `195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/proxy` |
| k8s-agent | `carlosmsanchez/aegis-k8s-agent` | `195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/k8s-agent` |
| workspace | `carlosmsanchez/aegis-workspace-vscode` | `195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode` |

### Pull Policy Rules

- **`IfNotPresent`** = Docker Desktop (daemon is shared with k8s; images built locally are immediately available)
- **`Always`** = EKS (nodes are separate VMs; must pull from registry)

### Command Routing

Use the specialized docs for command behavior instead of this file:

- `docs/make-commands.md` owns exact target semantics.
- `AGENT_DEPLOYMENT_GUIDE.md` owns workflow sequencing and deployment verification.
- `docs/PRODUCTION_DEPLOYMENT.md` owns production-only rollout deltas.

Cloud images use immutable git-SHA tags by default. The deploy script is deploy-only and consumes explicit image refs; do not treat this file as the source of truth for deployment command flow.

### Workspace Image

- Local: `docker.io/carlosmsanchez/aegis-workspace-vscode:latest` (set in `values-local.yaml`)
- Cloud: `195714074609.dkr.ecr.us-east-1.amazonaws.com/aegis/workspace-vscode:latest` (set in cloud values)

### Helm Values File Registry Summary

| Values file | Registry | Pull Policy |
|-------------|----------|-------------|
| `values.yaml` (spoke base) | DockerHub | IfNotPresent |
| `values-local.yaml` (spoke local) | DockerHub | IfNotPresent |
| `values-cloud-remote.yaml` (spoke cloud) | ECR | Always |
| `values-aws-relay.yaml` (spoke relay) | ECR | Always |
| `values/local.yaml` (hub local) | DockerHub | IfNotPresent |
| `values/cloud.yaml` (hub cloud) | ECR | Always |

## Hardening Profiles

- `dev` (default) -- Development
- `standard` -- Enterprise hardening

## Key Namespaces

| Namespace | Contents |
|-----------|----------|
| `aegis-system` | platform-api, proxy, ingress-nginx |
| `aegis-pki` | step-ca, step-issuer |
| `cert-manager` | cert-manager controller |
| `keycloak` | Keycloak SSO + Postgres |

## Agent Workflow (from AGENTS.md)

1. **Jira hygiene** -- Transition ticket to In Progress, add `codex-in-progress` label at start; swap to `automation-complete` on completion with a comment linking the PR.
2. **Local dev** -- Use MCP tools for file ops. Run `mcp__aegis__go_build` before tests. Run `mcp__aegis__run_tests` (10-min timeout). On 3rd failure, call `mcp__gpt5__advise`.
3. **Remote validation** -- Push branch, ensure PR is open. Poll `preview-deployment.yml` with `gh run list/watch`. On 3rd failure, call `mcp__gpt5__advise`.
4. **Git etiquette** -- Branch: `aegis-ci/<JiraKey>-<slug>`. Reference ticket in commits and PRs.
5. **Safety** -- Only dispatch `preview` or `ci` workflows. Abort if a command hangs >10 min.

## Prerequisites

- Go 1.24+
- Docker Desktop (with Kubernetes enabled)
- `kubectl`, `grpcurl`, Helm 3
- AWS CLI profile: `aegis-new` (account `195714074609`). Always use `--profile aegis-new` or `AWS_PROFILE=aegis-new` for AWS commands.

## Quick Local Dev Start

```bash
cd ~/code/aegis-platform
make stop && pkill -f 'k8s-agent/cmd' || true
kubectl delete jobs --all && kubectl delete aegisworkloads --all
kubectl apply -f agents/k8s-agent/config/crd/bases/aegis.yourorg.dev_aegisworkloads.yaml
make deploy-local-tls                          # hub only (default)
# make deploy-local-tls DEPLOY_SPOKE=true      # add spoke if needed
```
