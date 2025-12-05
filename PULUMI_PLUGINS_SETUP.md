# Pulumi Plugins Setup Instructions

## Current Status

The platform-api Docker image has been configured to pre-bake Pulumi plugins (AWS v5.70.0, Kubernetes v4.24.0) to avoid runtime downloads and GitHub rate limits. This is essential for air-gapped/DoD deployments.

**However**, plugin download is currently blocked by GitHub/Pulumi CDN rate limits.

## Next Steps

### Option 1: Wait and Retry (Recommended)

GitHub rate limits reset after ~1 hour. Simply wait and run:

```bash
cd services/platform-api/pulumi-plugins
export GITHUB_TOKEN=REDACTED_TOKEN
./download-plugins.sh
```

Once successful, build and deploy:

```bash
# From project root
make build-platform
kubectl -n aegis-system rollout restart deploy/aegis-services-platform-api
```

### Option 2: Download from Different Network

If you have access to another machine/network that isn't rate limited:

```bash
# On other machine
git clone <your-repo>
cd services/platform-api/pulumi-plugins
export GITHUB_TOKEN=REDACTED_TOKEN
./download-plugins.sh

# Transfer the bundle back
scp pulumi-plugins-bundle.tar.gz <your-machine>:~/code/aegis-platform/services/platform-api/pulumi-plugins/
```

Then build as normal.

### Option 3: Temporary Workaround (Runtime Download)

If you need to proceed immediately, you can temporarily allow runtime plugin downloads by reverting the Dockerfile changes and setting a GitHub token in the pod:

**Not recommended for production**, but for testing:

1. Create a Kubernetes secret with your GitHub token:
   ```bash
   kubectl -n aegis-system create secret generic pulumi-github-token \
     --from-literal=token=REDACTED_TOKEN
   ```

2. Update Helm values (`charts/aegis-services/values/local.yaml`):
   ```yaml
   platformApi:
     pulumi:
       githubTokenSecret:
         name: pulumi-github-token
         key: token
   ```

3. Redeploy:
   ```bash
   make upgrade-platform
   ```

This allows runtime plugin downloads but defeats the air-gap readiness goal.

## What Was Changed

### Files Modified

1. **Dockerfile** (`services/platform-api/Dockerfile`)
   - Modified to COPY pre-downloaded plugin bundle instead of downloading at build time
   - Plugins extracted to `/root/.pulumi/plugins/` during image build

2. **Download Script** (`services/platform-api/pulumi-plugins/download-plugins.sh`)
   - Automates plugin download using Pulumi Docker image
   - Supports GitHub token for higher rate limits
   - Creates `pulumi-plugins-bundle.tar.gz` (required for build)

3. **Documentation** (`services/platform-api/pulumi-plugins/README.md`)
   - Comprehensive guide for plugin management
   - Air-gap deployment instructions

4. **Helm Chart** (`charts/aegis-services/`)
   - Already configured with optional `githubTokenSecret` support (for runtime downloads if needed)

### Architecture

```
Build Time:
  download-plugins.sh → pulumi-plugins-bundle.tar.gz → Docker build → Image with plugins

Runtime:
  Pulumi checks /root/.pulumi/plugins/ → Finds plugins → No download needed
```

## Verification

Once the bundle is downloaded and image built, verify:

```bash
# Check image contains plugins
docker run --rm --entrypoint=sh carlosmsanchez/aegis-platform-api:dev -c "ls -la /root/.pulumi/plugins/"

# Should show:
# resource-aws-v5.70.0/
# resource-kubernetes-v4.24.0/
```

## Air-Gap Deployment Checklist

- [x] Dockerfile configured for pre-baked plugins
- [x] Download script created
- [x] Documentation added
- [x] .gitignore configured (bundle won't be committed)
- [ ] **Plugins downloaded** (blocked by rate limit - run `download-plugins.sh`)
- [ ] Image built with plugins
- [ ] Image pushed to registry
- [ ] Deployment tested

## Troubleshooting

### "File not found: pulumi-plugins-bundle.tar.gz"

You must run `download-plugins.sh` successfully before building. If rate limited, wait ~1 hour.

### "403 error fetching plugin"

**At build time**: The bundle is missing or extraction failed. Check Dockerfile COPY step.

**At runtime**: Plugins aren't in the image. Rebuild after running `download-plugins.sh`.

### Plugin Version Mismatch

Ensure versions match across:
- `download-plugins.sh` (AWS_VERSION, K8S_VERSION)
- `internal/provisioning/pulumi/aws/runner.go` (PluginVersion fields)
- Dockerfile comments

## Future Updates

When updating Pulumi or plugins:

1. Update `PULUMI_VERSION` in Dockerfile
2. Update versions in `download-plugins.sh`
3. Update versions in Go code
4. Run `./download-plugins.sh`
5. Rebuild image
6. Test thoroughly

## Questions?

- Plugin download issues: Check `services/platform-api/pulumi-plugins/README.md`
- Air-gap requirements: See README section on "Air-Gapped Deployments"
- Rate limits: Use GitHub token or wait for reset

## Contact

For DoD/air-gap specific requirements, ensure:
- All plugin tarballs are scanned/approved before bundling
- Image layers are scanned for vulnerabilities
- SBOM generated for compliance
- Update procedures documented for classified environments
