# Pulumi Plugins for Docker Image

This directory contains tooling to pre-bake Pulumi plugins into the platform-api Docker image, avoiding runtime downloads and GitHub rate limits. This is essential for air-gapped/DoD deployments.

## Required Plugins

- **AWS** v5.70.0 (linux-amd64)
- **Kubernetes** v4.24.0 (linux-amd64)

## Quick Start

### 1. Download Plugins

Run the download script to fetch and bundle the required plugins:

```bash
cd services/platform-api/pulumi-plugins
./download-plugins.sh
```

**If you hit rate limits:**

```bash
# Set GitHub token for higher rate limits (5000/hour vs 60/hour)
export GITHUB_TOKEN=ghp_your_token_here
./download-plugins.sh
```

The script will create `pulumi-plugins-bundle.tar.gz` (~450MB), which contains both plugins.

### 2. Build Docker Image

Once the plugin bundle exists, build the image normally:

```bash
# From project root
make build-platform

# Or for specific architecture
PLATFORMS=linux/arm64 make build-platform
```

## How It Works

1. **download-plugins.sh**: Uses a Pulumi Docker container to download plugins and bundle them into a tarball
2. **Dockerfile**: Copies the bundle during build and extracts it to `/root/.pulumi/plugins/`
3. **Runtime**: Pulumi finds pre-installed plugins, no network requests needed

## Troubleshooting

### Rate Limit Errors

```
Error: 403 HTTP error fetching plugin from https://get.pulumi.com/...
```

**Solutions:**
- Wait ~1 hour for rate limit reset
- Use GitHub token: `export GITHUB_TOKEN=...`
- Download from a different network/machine and copy the bundle file

### Plugin Version Mismatch

If you need different plugin versions:

1. Update versions in `download-plugins.sh`
2. Update versions in Go code (`internal/provisioning/pulumi/aws/runner.go`)
3. Re-run `./download-plugins.sh`
4. Rebuild image

### Missing Bundle File

```
Error: COPY failed: file not found: services/platform-api/pulumi-plugins/pulumi-plugins-bundle.tar.gz
```

You must run `./download-plugins.sh` successfully before building the image.

## Air-Gapped Deployments

For fully air-gapped environments:

1. **On internet-connected machine:**
   ```bash
   cd services/platform-api/pulumi-plugins
   export GITHUB_TOKEN=...  # if needed
   ./download-plugins.sh
   ```

2. **Transfer bundle:**
   ```bash
   # Copy pulumi-plugins-bundle.tar.gz to air-gapped environment
   scp pulumi-plugins-bundle.tar.gz user@airgapped-host:/path/to/aegis-platform/services/platform-api/pulumi-plugins/
   ```

3. **Build on air-gapped machine:**
   ```bash
   # On air-gapped machine
   make build-platform
   ```

## Plugin Cache Location

Plugins are installed to `/root/.pulumi/plugins/` with this structure:

```
/root/.pulumi/plugins/
├── resource-aws-v5.70.0/
│   └── pulumi-resource-aws
└── resource-kubernetes-v4.24.0/
    └── pulumi-resource-kubernetes
```

## Future Updates

When updating Pulumi or plugin versions:

1. Update `ARG PULUMI_VERSION` in `Dockerfile`
2. Update versions in `download-plugins.sh`
3. Update versions in Go code plugin requirements
4. Run `./download-plugins.sh` to fetch new versions
5. Rebuild and test

## Environment Variables (Optional)

- `GITHUB_TOKEN`: Increases GitHub API rate limit from 60 to 5000/hour
- `PULUMI_HOME`: Override plugin cache location (default: `/root/.pulumi`)
