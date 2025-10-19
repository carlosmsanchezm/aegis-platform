# OSCAL Control Sources

The JSON files in this directory are machine-readable control catalogs sourced
from NIST and FedRAMP. They provide authoritative language for automation and
agent workflows targeting AC-17, AU-2, and related compliance objectives.

| File | Revision | Source | Notes |
| ---- | -------- | ------ | ----- |
| `nist_sp_800_53_rev5_catalog.json` | SP 800-53 Rev. 5 | https://github.com/usnistgov/oscal-content | Full control catalog; ~1,196 controls. |
| `nist_sp_800_53b_rev5_moderate_baseline.json` | SP 800-53B Rev. 5 | https://github.com/usnistgov/oscal-content | FedRAMP Moderate-tailored subset for baseline checks. |
| `nist_sp_800_171_rev3_catalog.json` | SP 800-171 Rev. 3 | https://github.com/usnistgov/oscal-content | CUI-specific requirements mapped to DoD IL-5. |
| `fedramp_rev5_moderate_baseline.json` | FedRAMP Rev. 5 | https://github.com/GSA/fedramp-automation | FedRAMP Moderate resolved profile with assessment guidance. |

## Updating the Catalogs

1. Download the latest OSCAL JSON from the linked repositories.
2. Replace the existing file, keeping the same filename so the MCP tooling can
   auto-discover it.
3. Run the smoke tests in `agents/mcp-server/README.md` to verify the data is
   still parsed correctly.

## Adding New Sources

To add another baseline or profile:

1. Place the OSCAL JSON file in this directory.
2. Update `DATASETS` in `agents/mcp-server/tools/nist.js` with the new filename
   and metadata label.
3. (Optional) Extend the MCP README with usage notes for the new dataset.

## Conversion Workflow (If OSCAL JSON Is Unavailable)

In rare cases you may need to convert a PDF/Excel control list yourself:

1. Install the official OSCAL CLI tooling locally:

   ```bash
   pip install oscal-cli
   ```

2. Convert the source document:

   ```bash
   oscal-cli convert --input path/to/source.xlsx --output docs/oscal/new_catalog.json --format json
   ```

3. Validate the result against the OSCAL schema. The CLI reports any
   structural issues that must be fixed before agents can rely on the file.

Leave placeholders blank if you cannot reach a new data source yet—the MCP
loader will skip unknown datasets and continue serving the existing ones.
