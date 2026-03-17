# Fresh Deployment Issues Log

**Date:** 2026-03-17
**Branch:** `feature/ironbank-compliance`
**Cluster:** `aegis-hub-prod` (EKS, us-east-1)

## Summary

After terraform destroy + fresh terraform apply + generate-cloud-deployment.sh, the following issues required manual intervention. Each needs to be codified so the next fresh deploy works without patches.

---

## Issue 1: ECR Images Deleted by Terraform Destroy

**What happened:** `terraform destroy` deleted ECR repository contents. After `terraform apply` recreated the repos, they were empty. The deploy script's image verification step failed.

**Manual fix:** Rebuilt and pushed all 6 images from local Docker cache. Proxy image wasn't in cache and needed a fresh build.

**Codify:** ECR repositories should be excluded from terraform destroy (already partially done with `manage_ecr_repositories` variable). Or: the build-and-push script should run before deploy, not assumed to have been run separately.

---

## Issue 2: Third-Party Iron Bank Image Overrides Caused ImagePullBackOff

**What happened:** `cloud.yaml` had Iron Bank image references for Keycloak, PostgreSQL, and Ingress NGINX (`registry1.dso.mil/...`). The fresh cluster has no registry1 auth, so all third-party pods failed with ImagePullBackOff.

**Manual fix:** Reverted `cloud.yaml` to use public images (quay.io, docker.io, registry.k8s.io). Iron Bank references kept as comments for when registry1 auth is available.

**Committed:** `0ca4312` — `fix: revert third-party Iron Bank image overrides to public images`

**Codify:** Third-party images should default to public registries. Iron Bank overrides should be in a separate values file (`values/ironbank.yaml`) that's only applied when registry1 credentials are configured on the cluster.

---

## Issue 3: Cloudflare DNS Not Updated (CLOUDFLARE_API_TOKEN required)

**What happened:** Deploy script requires `CLOUDFLARE_API_TOKEN` env var. First run with `SKIP_DNS_UPDATE=1` succeeded but DNS pointed to old NLBs.

**Manual fix:** Ran deploy again with the token, then manually updated DNS via Cloudflare API when the deploy script created new NLBs.

**Codify:** The Cloudflare API token should be stored in AWS Secrets Manager or as a Kubernetes secret, not passed as an env var. Or: use Terraform to manage DNS records (already has `cloudflare.tf`).

---

## Issue 4: Deploy Script Re-ran Created New NLBs (DNS Out of Sync)

**What happened:** Each time `generate-cloud-deployment.sh` runs, it deletes and recreates the Helm release, which creates new NLB endpoints. Cloudflare DNS records set in a previous run become stale.

**Manual fix:** Had to update Cloudflare DNS records 3 times during testing as NLBs changed.

**Codify:** The deploy script should use `helm upgrade --install` (not delete + install) to preserve existing LoadBalancer services and their NLB endpoints. Or: use stable NLB DNS names via Terraform-managed NLBs.

---

## Issue 5: Backstage CrashLoopBackOff — Keycloak TLS Not Trusted

**What happened:** Backstage's Node.js process connects to Keycloak via the public URL. Keycloak's ingress serves a cert signed by the internal CA. Backstage didn't have the CA in its trust store.

**Error:** `OPError: expected 200 OK, got: 530 undefined` (Cloudflare origin error while DNS was stale), then `Error: unable to get local issuer certificate` after DNS resolved.

**Manual fix:** Patched the Backstage deployment to mount `aegis-trust-bundle` secret and set `NODE_EXTRA_CA_CERTS`.

**Committed:** `38c3c4e` — `fix: mount trust bundle into Backstage for Keycloak TLS`

**Helm value added to `cloud.yaml`:**
```yaml
backstage:
  caBundle:
    secretName: aegis-trust-bundle
```

**Codify:** Already codified in cloud.yaml. The Helm template already supported this — just needed the value set. Should also be in `common.yaml` as default since ANY deployment with internal PKI needs this.

---

## Issue 6: Platform-API UNAUTHENTICATED — OIDC JWKS TLS Not Trusted

**What happened:** Platform-api validates JWT tokens by fetching Keycloak's JWKS endpoint. The fetch failed with `x509: certificate signed by unknown authority` because the Go HTTP client didn't trust the internal CA.

**Error:** `token is unverifiable: error while executing keyfunc: failed to fetch JWKS: tls: failed to verify certificate: x509: certificate signed by unknown authority`

**Manual fix:** None needed beyond committing the Helm value.

**Committed:** `b879491` — `fix: mount trust bundle into platform-api for OIDC/JWKS TLS verification`

**Helm value added to `cloud.yaml`:**
```yaml
platformApi:
  auth:
    oidc:
      caBundle:
        secretName: aegis-trust-bundle
```

**Codify:** Already codified in cloud.yaml. The Helm template already supported this (`OIDC_CA_BUNDLE` env var + volume mount). Should also be in `common.yaml` as default.

---

## Issue 7: DNS Cache on Mac (Not a Platform Issue)

**What happened:** After updating Cloudflare DNS records, the Mac's DNS cache retained old records. `curl` and the Sovran extension couldn't resolve the new endpoints.

**Fix:** `sudo dscacheutil -flushcache && sudo killall -HUP mDNSResponder`

**Not a codification issue** — this is client-side DNS caching, not a platform problem.

---

## Issue 8: Spoke Agent Not Deployed (No k8s-agent in Pods)

**What happened:** The deploy script's Step 8 (spoke deployment) ran but the spoke agent pod wasn't visible initially. It appeared after a few minutes.

**Root cause:** The deploy script re-creates the Helm release on each run, and pod scheduling takes time.

**No fix needed** — just timing.

---

## Commits Made During This Session

| Commit | Description |
|--------|-------------|
| `0ca4312` | Revert third-party Iron Bank image overrides to public images |
| `38c3c4e` | Mount trust bundle into Backstage for Keycloak TLS |
| `b879491` | Mount trust bundle into platform-api for OIDC/JWKS TLS verification |

---

## What Needs to Be in `common.yaml` (Default for ALL Deployments)

These trust bundle settings should be defaults, not cloud-only overrides, because any deployment with internal PKI needs them:

```yaml
# common.yaml additions needed:
platformApi:
  auth:
    oidc:
      caBundle:
        secretName: aegis-trust-bundle

backstage:
  caBundle:
    secretName: aegis-trust-bundle
```

---

## What's Still Not Automated

1. **ECR image availability** — Need to verify images exist before deploy, or build-and-push as part of deploy
2. **Cloudflare API token** — Should be in secrets manager, not manual env var
3. **NLB stability** — Deploy script recreates services, changing NLB endpoints. Need stable endpoints.
4. **Trust bundle as default** — Move from `cloud.yaml` to `common.yaml` so all deployments get internal CA trust
