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

## Design System & Themes

- The Frontend theme lives in `packages/app/src/theme/aegisTheme.tsx`. Update this file when adding new palette tokens, typography, or component overrides so the ÆGIS Nightfall/Dawn themes stay in sync.
- Dark mode (`ÆGIS Nightfall`) loads by default. Use the floating sun/moon control in the upper-right corner of the app chrome to flip themes instantly, or open **Settings → Theme & Appearance** for a richer chooser and “follow system” option.
- Theme selections persist to `localStorage` and respect the browser’s `prefers-color-scheme` hint on the first visit.

## Workspace Wizard

- The card-based wizard is implemented under `plugins/aegis/src/wizard`. Add new workspace templates by extending the `templates` array inside `WorkspaceWizard.tsx`, providing defaults for image, queue, ports, and any environment variables.
- Reuse `TemplateCard` for new selectable cards, and update `types.ts` when introducing new template or flavor identifiers.
- Unit coverage: `ThemeToggleButton.test.tsx` covers the global toggle, and `WorkspaceWizard.test.tsx` verifies template selection plus validation. Run `yarn test` from the repository root to execute them.
- End-to-end: `packages/app/e2e-tests/workspace-wizard.spec.ts` exercises Templates → Configure → Review → Launch with a stubbed SubmitWorkload response. Start Backstage (`yarn dev`), then run `yarn test:e2e` to drive the flow with Playwright.
