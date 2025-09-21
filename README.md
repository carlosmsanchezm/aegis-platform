# aegis

Local development quickstart lives in the make targets and scripts under
`scripts/dev`. For direct API pokes, see
`docs/grpcurl-cheatsheet.md` for step-by-step `grpcurl` examples compatible with
the current proto definitions.

> `make proto` expects `buf`, `protoc-gen-go`, and `protoc-gen-go-grpc` to be
> available in `./bin` (or on your `PATH`). The `bin/` directory is ignored in
> git so you can manage tool versions locally.

## Week 0 Validation Steps

These steps assume you are running locally (network-restricted sandboxes should
only execute the non-network commands).

1. **Regenerate protobuf stubs**
   ```bash
   PATH=$(pwd)/bin:$PATH make proto
   ```
2. **Tidy modules and build binaries**
   ```bash
   PATH=$(pwd)/bin:$PATH make tidy
   PATH=$(pwd)/bin:$PATH make build
   ```
3. **Run the control-plane API** (local shell only)
   ```bash
   PATH=$(pwd)/bin:$PATH make run-api
   ```
4. **Launch one or more agents** (local shell only)
   ```bash
   PATH=$(pwd)/bin:$PATH make run-agent
   # or customize
   PATH=$(pwd)/bin:$PATH make run-agent \
     AEGIS_CP_GRPC=localhost:8081 \
     AEGIS_CLUSTER_ID=dev-gke \
     AEGIS_REGION=us-central \
     AEGIS_PROVIDER=DEV
   ```
5. **Exercise the API** using the commands in
   `docs/grpcurl-cheatsheet.md` (create project, submit workload, list, etc.).

> **Note:** Do not invoke the `run-*` targets inside sandboxed CI/agent
> environments; they attempt to bind/listen on local ports. Use the `proto`,
> `tidy`, and `build` targets there instead.
