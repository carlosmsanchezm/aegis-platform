# Software Bill of Materials (SBOM)

## Overview

This directory contains tools and documentation for generating and maintaining Software Bills of Materials for Aegis Platform. SBOMs are required for supply chain transparency and compliance with federal requirements (EO 14028).

## SBOM Format

Aegis uses **CycloneDX** format (version 1.5) for SBOMs. CycloneDX is:
- OWASP maintained
- JSON and XML support
- Widely supported by security tools
- Compatible with federal requirements

## Generated SBOMs

| Component | SBOM Location | Update Frequency |
|-----------|--------------|-----------------|
| platform-api | `sbom/platform-api-sbom.json` | Each release |
| backstage | `sbom/backstage-sbom.json` | Each release |
| k8s-agent | `sbom/k8s-agent-sbom.json` | Each release |
| spoke-proxy | `sbom/spoke-proxy-sbom.json` | Each release |
| keycloak | `sbom/keycloak-sbom.json` | Each release |

## Generation Instructions

### Prerequisites

```bash
# Install CycloneDX tools
npm install -g @cyclonedx/cdxgen

# For Go projects
go install github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@latest

# For container images
curl -sSfL https://raw.githubusercontent.com/anchore/syft/main/install.sh | sh -s -- -b /usr/local/bin
```

### Generate SBOMs

Run the generation script:

```bash
./generate-sbom.sh
```

Or generate individually:

```bash
# Go projects (platform-api, k8s-agent, spoke-proxy)
cd services/platform-api
cyclonedx-gomod mod -json -output ../../docs/compliance/sbom/platform-api-sbom.json

# Node.js projects (backstage)
cd services/backstage
cdxgen -o ../../docs/compliance/sbom/backstage-sbom.json

# Container images (full dependency tree)
syft your-registry/aegis-platform-api:latest -o cyclonedx-json > platform-api-container-sbom.json
```

## SBOM Contents

Each SBOM includes:

| Field | Description | Example |
|-------|-------------|---------|
| `bomFormat` | Format identifier | CycloneDX |
| `specVersion` | Spec version | 1.5 |
| `serialNumber` | Unique identifier | urn:uuid:... |
| `version` | SBOM version | 1 |
| `metadata` | Creation info | timestamp, tool |
| `components` | Dependency list | See below |

### Component Information

For each dependency:
- Name and version
- Package URL (purl)
- License(s)
- SHA-256 hash
- Supplier info (when available)
- External references

## Vulnerability Scanning

Use SBOMs with vulnerability scanners:

```bash
# Grype (Anchore)
grype sbom:platform-api-sbom.json

# Trivy
trivy sbom platform-api-sbom.json

# OSV-Scanner (Google)
osv-scanner --sbom=platform-api-sbom.json
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
name: Generate SBOM
on:
  release:
    types: [published]

jobs:
  sbom:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Generate SBOM
        uses: CycloneDX/gh-gomod-generate-sbom@v2
        with:
          version: v1
          args: mod -json -output sbom.json

      - name: Upload SBOM
        uses: actions/upload-artifact@v3
        with:
          name: sbom
          path: sbom.json

      - name: Attach SBOM to release
        uses: softprops/action-gh-release@v1
        with:
          files: sbom.json
```

## Customer Delivery

SBOMs should be delivered to customers:

1. **With each release** - Attach to release artifacts
2. **On request** - Available via customer portal
3. **Automated** - API endpoint for SBOM retrieval (optional)

### SBOM Attestation

For higher assurance, sign SBOMs:

```bash
# Using cosign
cosign sign-blob --key cosign.key \
  --output-signature sbom.sig \
  platform-api-sbom.json

# Verify
cosign verify-blob --key cosign.pub \
  --signature sbom.sig \
  platform-api-sbom.json
```

## Compliance Requirements

| Framework | SBOM Requirement |
|-----------|-----------------|
| EO 14028 | Required for software sold to federal government |
| FedRAMP | Recommended, becoming required |
| NIST 800-171 | Supports CM-8 (Component Inventory) |
| SOC 2 | Supports CC7.1 (Vulnerability Management) |

## File Inventory

```
sbom/
├── README.md                      # This file
├── generate-sbom.sh               # Generation script
├── sbom-template.json             # CycloneDX template
├── platform-api-sbom.json         # Generated (gitignored)
├── backstage-sbom.json            # Generated (gitignored)
├── k8s-agent-sbom.json            # Generated (gitignored)
├── spoke-proxy-sbom.json          # Generated (gitignored)
└── keycloak-sbom.json             # Generated (gitignored)
```

## Related Documentation

- [Configuration Hardening Guide](../customer-docs/configuration-hardening-guide.md)
- [Control Implementation Statements](../customer-docs/control-implementation-statements.md) (CM-8)
