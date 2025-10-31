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

The preview GitHub Actions workflow now waits for the Keycloak pods to reach
`Ready` status and reuses `scripts/keycloak-token.sh` to discover the exposed
service ports automatically. The helper falls back to HTTP/8080 when TLS is
disabled, probes the `.well-known/openid-configuration` endpoint before
requesting an access token, and retries the token exchange with exponential
backoff. This keeps the platform-api E2E jobs resilient to slow Keycloak
startups while still enforcing TLS whenever the service exposes port 8443.
