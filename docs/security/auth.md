# Backstage Authentication Posture

This Backstage deployment now relies on Keycloak-backed SSO. The control
implementation aligns with NIST SP 800-171r3 and FedRAMP Moderate guidance
for identity, transport protection, and audit logging.

## Identity & Authentication

- **MFA (03.05.03 / IA-2)** – Keycloak realm configuration must enforce MFA
  for every user. TOTP and WebAuthn authenticators are supported. Enrollment
  is required prior to granting access to Backstage or downstream systems.
- **Replay Resistance (03.05.04 / IA-11)** – The Keycloak provider reuses the
  hardened OIDC authenticator from Backstage, which enables PKCE (S256),
  state, and nonce on every authorization code exchange.
- **Subject Propagation (AC-02 / AC-06)** – Successful logins produce signed
  Backstage identity tokens that are forwarded to the Aegis proxy. Guest
  identities are rejected whenever `requireAuth` is set, preventing silent
  privilege escalation.
- **Token flow** – Browser → Keycloak (Auth Code + PKCE) → Backstage frontend
  → Backstage backend OIDC provider → Platform API. Access tokens issued at the
  Backstage layer are forwarded verbatim to the Platform API gateway where they
  are re-validated against the issuer.

### Runtime Configuration

Platform API authentication is controlled with environment variables (or the
`AUTH_CONFIG_JSON` overlay if you prefer a single blob):

- `OIDC_ISSUER_URL` – Keycloak issuer, e.g. `https://keycloak.localtest.me/realms/aegis`
- `OIDC_AUDIENCE` – Expected audience claim (Backstage client or API audience).
- `OIDC_JWKS_URL` *(optional)* – Override JWKS endpoint; defaults to
  `${OIDC_ISSUER_URL}/protocol/openid-connect/certs`.
- `OIDC_JWKS_CACHE_TTL` / `OIDC_JWKS_REFRESH_INTERVAL` – Cache behaviour for
  JWKS retrieval (Go duration strings).
- `REQUIRE_PHISHING_RESISTANT_MFA` – When `true`, API calls are rejected unless
  the token’s `amr` claim includes one of the factors below.
- `ALLOWED_PHISHING_RESISTANT_AMR` – Comma separated list of AMR values that
  qualify as phishing-resistant (defaults to `hwk,webauthn,piv,piv-cac`).

Authorization bindings are provided through `AUTHZ_ROLE_BINDINGS_JSON`, a JSON
array with entries such as:

```json
[
  {
    "roles": ["aegis-admin"],
    "projects": ["*"],
    "queues": ["*"]
  }
]
```

In Helm the value is sourced from the `authz-role-bindings.json` secret. Leaving
the array empty permits all requests, which is acceptable for development but
must be tightened for production.

## Cryptography & Transport Protection

- **TLS Scope (SC-8)** – Treat all traffic between browser ⇄ Backstage ⇄
  Keycloak ⇄ platform APIs as in-scope "data-in-transit" that must be
  protected end-to-end.
- **FIPS Expectations (SC-13 / FedRAMP SC-13)** – Terminate every TLS hop with
  FIPS 140-3 validated cryptographic modules where available. This covers
  HTTPS endpoints, Keycloak token issuance, and MFA token generation.
- **Configuration** – The OIDC provider requires the following environment
  variables (kept out of git):
  `KEYCLOAK_METADATA_URL`, `KEYCLOAK_AUTH_URL`, `KEYCLOAK_TOKEN_URL`,
  `KEYCLOAK_USERINFO_URL`, `KEYCLOAK_LOGOUT_URL`, `KEYCLOAK_CLIENT_ID`,
  `KEYCLOAK_CLIENT_SECRET`, `BACKEND_SECRET`, and `BACKEND_BASE_URL`.

## Session Authenticity

- **Cookie Hardening (SC-23)** – Session cookies are marked `Secure`,
  `HttpOnly`, and `SameSite=Lax` at the ingress/controller layer. Backstage’s
  backend sessions are signed with `BACKEND_SECRET`, ensuring integrity.
- **Workspace Session TTL** – Platform API connection sessions issue one-time
  tokens with a five-minute default TTL (`AEGIS_PROXY_TOKEN_TTL_SECONDS`),
  capped at five minutes even if callers request longer durations. Renewals
  require the same authenticated subject, satisfying AC-12 / SC-10 idle timeout
  expectations.

## Audit & Accountability

- **Auth Event Logging (AU-02 / AU-03 / AU-08 / AU-12)** – Every request to
  `/api/auth/**` now emits a structured `auth-event` log at INFO level with:
  UTC timestamp, event type (start, callback, refresh, logout, failure), HTTP
  method, response status, Keycloak provider, resolved subject (when
  available), client IP (preferring `X-Forwarded-For`), user agent, request
  latency in milliseconds, and the `decision` flag (`allow`/`deny`). Successful
  authentications also include `client_id`, `amr`, `acr`, `token_id` (JWT JTI),
  and `key_id` (JWKS kid) so downstream tooling can correlate events.
- **Log Handling** – Ship these entries to the central SIEM with retention
  that satisfies AU-09 and AU-11. Downstream alerting should cover repeated
  failures and anomalous logout patterns.

## Operational Checklist

1. Provision the Keycloak client (`backstage`) with PKCE enabled and the
   redirect URI `https://<backstage>/api/auth/keycloak/handler/*`.
2. Require MFA for all users in the Keycloak realm.
3. Set the environment variables above via Helm values or secret managers –
   never commit credentials.
4. Ensure all ingress controllers enforce TLS 1.2+ and HSTS.
5. Forward Backstage backend logs to the compliance logging pipeline and
   verify the presence of `auth-event` entries during smoke tests.

## Preview CI behaviour

The `preview-deployment.yml` “tests-only” path provisions the control plane in
an ephemeral namespace (for example `preview-29`), deploys Keycloak, and runs the
Platform API E2E suite against the public DNS endpoints. The flow is:

1. **Ensure Keycloak comes online**
   - Helm installs both the Keycloak operator and in-cluster RHBK instance.
   - We bind the PostgreSQL PVC to the EKS `gp2` storage class and, on every
     run, delete any existing Keycloak stateful sets and PVCs so RHBK can
     provision storage cleanly. (See
     `charts/aegis-services/values/cloud.yaml` and
     `terraform/generate-cloud-deployment.sh`.)
   - The workflow step **Wait for Keycloak readiness** blocks until the pods
     with `app=keycloak` report `Ready`. If they never reach that state the job
     fails early, which prevents the E2Es from running with a half-configured
     SSO stack.

2. **Obtain a real bearer token**
   - `scripts/e2e-platform-api.sh` exports the preview namespace/release
     (e.g. `KEYCLOAK_NAMESPACE=preview-29`,
     `KEYCLOAK_SERVICE_NAME=preview-29-keycloak-service`) and calls
     `scripts/keycloak-token.sh`.
   - `keycloak-token.sh` discovers the service, waits for the Keycloak pods,
     establishes a port-forward if the service is internal-only, and requests
     an access token via the Resource Owner Password grant using the automation
     user (`cloud@test.com`) and the Backstage client secret.
   - Successful responses include a non-empty `access_token`, which the script
     prints to stdout. The E2E harness captures it and exports
     `AEGIS_BEARER_TOKEN`.

3. **Exercise the Platform API with OIDC**
   - All subsequent gRPC calls in the suite add
     `Authorization: Bearer ${AEGIS_BEARER_TOKEN}`. No `x-aegis-user` fallback
     is accepted.
   - The suite performs the following checks end to end:
       * `SubmitWorkload` issues a ticket for a real `workspace` job in the
         preview cluster.
       * The new workload transitions through `PLACED` to `SUCCEEDED`, proving
         the control plane, kube-agent, and namespace wiring honour the token.
       * `AckWorkload` confirms the caller received the ticket and the job can
         be cleaned up.

4. **Success criteria**
   - Token exchange returns HTTP 200 and the bearer token is present in logs.
   - The workload status observed by the suite reaches `SUCCEEDED` (the logs
     show `status":"SUCCEEDED", "uiStatus":"SUCCEEDED"` for the workload).
   - The job exits 0; the GitHub Actions run is marked `conclusion: success`
     and linked from the PR/Jira updates.

If any of the steps above fail (e.g. Keycloak pods crash, PVC cannot bind, the
token endpoint returns 4xx), the workflow stops and surfaces the failure so we
can debug before merging.

### Manual validation against a running preview

When a preview environment such as `preview-29` is already deployed you can
replicate the workflow end-to-end:

1. **Point kubectl at the preview cluster**

   ```bash
   cd terraform
   eval "$(terraform output -raw kubectl_config_command)"
   export KUBE_CONTEXT=$(kubectl config current-context)
   echo "Using kube context: ${KUBE_CONTEXT}"
   ```

2. **Verify Keycloak health and storage**

   ```bash
   PREVIEW_NS=preview-29

   kubectl -n "${PREVIEW_NS}" get pods -l app=keycloak
   kubectl -n "${PREVIEW_NS}" get pvc -l app.kubernetes.io/component=keycloak-postgres -o wide
   # View service endpoints and TLS
   kubectl -n "${PREVIEW_NS}" get svc | grep -i keycloak
   ```

   You should see the Keycloak pod `Ready` and the PVC bound to the `gp2`
   storage class.

3. **Mint a bearer token manually**

   ```bash
   KEYCLOAK_NAMESPACE=${PREVIEW_NS} \
   KEYCLOAK_SERVICE_NAME=${PREVIEW_NS}-keycloak-service \
   KEYCLOAK_PORT_FORWARD=1 \
   KEYCLOAK_USERNAME=cloud@test.com \
   KEYCLOAK_PASSWORD=password \
   KEYCLOAK_CLIENT_ID=backstage \
   KEYCLOAK_CLIENT_SECRET=$(kubectl -n "${PREVIEW_NS}" get secret ${PREVIEW_NS}-keycloak-backstage-client-secret -o jsonpath='{.data.clientSecret}' | base64 --decode) \
   scripts/keycloak-token.sh > /tmp/keycloak-token
   TOKEN=$(cat /tmp/keycloak-token)
   ```

   The helper logs should show `http_status=200` and an `access_token` in the
   response body.

4. **Exercise the Platform API using the token**

   ```bash
   GRPC_HOST=platform-api-grpc.aegist.dev
   GRPC_PORT=443
   CA_CERT=/path/to/aegis-platform-api-ca.crt   # download from workflow artifact

   grpcurl \
     -d '{"project_id":"p-e2e-manual","queue":"default","cluster_id":"aws-us-east-1-prod","workspace":{"flavor":"a10-mig-1g","image":"alpine:3.19","command":["sh","-c","echo hello"],"env":{"USER_NAME":"aegis","USER_PASSWORD":"aegis123"}}}' \
     -H "authorization: Bearer ${TOKEN}" \
     -cacert "${CA_CERT}" \
     ${GRPC_HOST}:${GRPC_PORT} \
     aegis.platform.v1alpha.WorkloadsService/SubmitWorkload
   ```

   Capture the returned `workload_id`, then poll `GetWorkload` and call
   `AckWorkload` using the same bearer token to mirror what the CI job does.

5. **Connect from VS Code / CLI**

   Add or override the Aegis extension settings to target the preview:

   ```jsonc
   {
     "aegis.remote": {
       "platform": {
         "grpcEndpoint": "platform-api-grpc.aegist.dev:443",
         "projectId": "p-e2e-manual",
         "namespace": "aegis-workloads",
         "authScope": "aegis-platform",
         "rejectUnauthorized": true,
         "mtlsSource": "platform",
         "caCertPath": "/path/to/aegis-platform-api-ca.crt"
       }
     }
   }
   ```

   Launch VS Code with these settings (or re-run `scripts/test-workspace-connection.sh` \
   with the same `GRPC_*` environment variables). You should reach the workspace
   pod deployed to the preview cluster using the Keycloak-issued token.

This process mirrors each stage of the CI workflow and confirms you can obtain
bearer tokens, issue workloads, and connect to compute from a local environment.
