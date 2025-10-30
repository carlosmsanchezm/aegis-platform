# Authentication & Access Control Mapping

| Control | Capability | Implementation Evidence |
|---------|------------|-------------------------|
| AC-02, AC-06, AC-17 | Enforce authenticated, role scoped access to Platform API workloads and budgets | gRPC/HTTP middleware validates Keycloak JWTs (`services/platform-api/internal/server/mw/auth.go`) and applies policy bindings from `AUTHZ_ROLE_BINDINGS_JSON` before mutating or reading project/queue resources. Session provisioning paths call `authorize` prior to minting connection tokens. |
| IA-2, IA-2(1/2), IA-11 | MFA and phishing-resistant authenticators for privileged and user access | Backstage OIDC provider auto-redirects to Keycloak with PKCE. Platform API rejects tokens lacking allowed `amr` values when `REQUIRE_PHISHING_RESISTANT_MFA=true`. |
| AU-2, AU-3, AU-8, AU-11 | Structured audit logging with ≥90 day retention baseline | Backstage `auth-event` logs capture subject, decision, IP, client metadata in JSON. Platform API authentication middleware logs `auth_login`/`auth_failure` with issuer, client_id, amr/acr, token and key identifiers, enabling forwarding to central SIEM with retention ≥90 days per logging pipeline policy. |
| AU-5 | Alerting on auth back-pressure/failure | Prometheus counter `aegis_auth_failures_total` increments per failure reason enabling alerting on authentication or MFA enforcement errors. |
| SC-8, SC-12, SC-13 | TLS, key management, and FIPS-aligned crypto | Backstage pods mount trust bundle and set `NODE_EXTRA_CA_CERTS`. Platform API Dockerfile now ships on UBI9 minimal with `update-crypto-policies --set FIPS`, ingress enforces TLS 1.2+ with curated cipher suites, and JWKS caching honours issuer trust. |
| SC-10, AC-12 | Session termination / idle timeout | Platform API connection sessions minted via `mintConnectionSession` expire after a maximum of five minutes (configurable via `AEGIS_PROXY_TOKEN_TTL_SECONDS` but clamped to 5m). Renewals require the same authenticated subject and session revocation clears tokens immediately. |

For configuration instructions and environment variable descriptions see `docs/security/auth.md`. The Helm chart exposes the relevant knobs under `charts/aegis-services/values/*` via `platformApi.env` and the `authz-role-bindings.json` secret.
