# Aegis Platform (Backstage UI)

Backstage-based frontend for the Aegis platform.

## Quick Start

### Local Development

```bash
# From repository root
make deploy-local && make port-forward && make dev-backstage
# Access at http://localhost:3000
```

### Keycloak SSO setup

```bash
# Install the Keycloak CRDs (done automatically when you deploy `aegis-services`)
kubectl get crd keycloaks.k8s.keycloak.org

# Enable Keycloak in the chart with forceRender for the first install
helm upgrade --install aegis-services charts/aegis-services \
  -f charts/aegis-services/values/common.yaml \
  -f charts/aegis-services/values/local.yaml \
  -f charts/aegis-services/values/local-tls.yaml \
  --set keycloak.forceRender=true \
  --namespace aegis-system --create-namespace --wait

# Populate Backstage environment variables
cp aegis-platform/.env.development aegis-platform/.env
```

The `.env.development` file points Backstage at the local Keycloak instance (`https://keycloak.localtest.me`). Adjust the values if you deploy Keycloak to a different hostname or realm.

### Cloud Development

```bash
make dev-backstage-cloud
# Access at http://localhost:3000
```

## Configuration Modes

| Command                        | Backend                            | Port-Forward? |
| ------------------------------ | ---------------------------------- | ------------- |
| `make dev-backstage`           | localhost:10080                    | Yes ✅        |
| `make dev-backstage-cloud`     | platform-api.aegist.dev:8080       | No ❌         |
| `make dev-backstage-cloud-tls` | platform-api.aegist.dev:8080 (TLS) | No ❌         |

## Config Files

- `app-config.local-dev.yaml` → `http://localhost:10080` (local mode)
- `app-config.cloud.yaml` → `http://platform-api.aegist.dev:8080` (cloud mode)
- `app-config.cloud-tls.yaml` → Cloud with TLS
- `app-config.local.yaml` → Active config (gitignored, auto-copied)

## Workflow

**Local:**

```bash
# Terminal 1
make deploy-local
make port-forward

# Terminal 2
make dev-backstage
```

**Cloud:**

```bash
make dev-backstage-cloud
```
