# Pulumi Preview Workflow

This guide explains how to run the Pulumi preview helper locally and how it is used in CI to validate infrastructure changes without performing a full `pulumi up`.

## Local usage

Use `scripts/pulumi-preview.sh` to orchestrate a deterministic preview against a `ProjectInfra` specification.

```bash
# Preview using the bundled sample spec (applies to the cluster, runs preview, then cleans up)
./scripts/pulumi-preview.sh

# Preview an existing ProjectInfra from the cluster
./scripts/pulumi-preview.sh --namespace aegis-system --name my-project

# Preview a manifest from disk without touching the cluster
./scripts/pulumi-preview.sh --file path/to/projectinfra.yaml

# Rebuild platform-api image, restart the deployment, and keep the applied CR for debugging
./scripts/pulumi-preview.sh --build-image --keep
```

Key behaviour:

- The script optionally rebuilds/pushes the platform-api image and rolls the deployment so the Pulumi runner matches the latest code.
- A sample manifest lives at `ops/previews/projectinfra-sample.yaml`; it is applied automatically when no explicit `--file`/`--name` is provided.
- Output from the CLI streams directly to the terminal so you can inspect the Pulumi diff.
- Unless `--keep` is supplied, any ProjectInfra created by the script is deleted before exit.

The script delegates to the make target `make -C services/platform-api pulumi-preview`, which runs the CLI via `go run`. Pass extra arguments by exporting `PULUMI_PREVIEW_ARGS="--kubeconfig /tmp/kubeconfig"`.

## CLI details

The CLI entry point lives at `services/platform-api/cmd/pulumi-preview/main.go`. It reuses the controller's Pulumi runner to build the stack, then calls `stack.Preview` with progress streaming to stdout. The helper supports:

- Loading a ProjectInfra from disk with `--file`.
- Fetching a ProjectInfra from the cluster with `--namespace`/`--name` (honouring `KUBECONFIG` or in-cluster config).
- Emitting a summary of resource operations after a successful preview.

The runner continues to perform `Refresh + Up` when invoked by the controller; the preview tool is strictly a preflight check.

## CI integration

`.github/workflows/preview-deployment.yml` now exercises the Pulumi preview after Terraform bootstraps the management plane:

1. Build the CLI binary (`bin/pulumi-preview`).
2. Apply `ops/previews/projectinfra-sample.yaml` to the preview cluster.
3. Run the CLI against that resource; the job fails immediately if Pulumi returns a non-zero exit code.
4. Remove the temporary ProjectInfra during cleanup (`kubectl delete projectinfra …`).

You can inspect the streamed Pulumi diff in the "Run pulumi preview CLI" step of the workflow run. The manifest and CLI live in-repo, so updating them in tandem keeps the automation consistent.

Remember: only the preview command calls `stack.Preview`. The controller still executes `Refresh` + `Up`, so successful previews do not change the behaviour of live reconciliations.
