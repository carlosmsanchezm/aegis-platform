# gRPCURL Quickstart

These examples assume the control-plane API is running on `localhost:8081` and
reflection is enabled (already true for the current build).

## Inspect message shapes

```bash
grpcurl -plaintext localhost:8081 describe aegis.v1.Project
grpcurl -plaintext localhost:8081 describe aegis.v1.Workload
grpcurl -plaintext localhost:8081 describe aegis.v1.ListWorkloadsResponse
```

`describe` is the easiest way to confirm exact field names before crafting
requests. Remember: use the proto-cased field names such as `projectId` rather
than `project_id`.

## Minimal happy-path sequence

1. **Start the platform API**

   ```bash
   PATH=$(pwd)/bin:$PATH go run ./services/platform-api
   ```

   The process logs both the gRPC and HTTP listener addresses at startup.

2. **Create a project**

   ```bash
   grpcurl -plaintext \
     -d '{"project":{"id":"p-demo","displayName":"Demo"}}' \
     localhost:8081 aegis.v1.AegisPlatform/CreateProject
   ```

   Only the fields declared in the proto are accepted. `Project` uses `id`,
   `displayName`, `ownerGroup`, and `policy`—no `name` field exists.

3. **Register a cluster**

   ```bash
   grpcurl -plaintext \
     -d '{"clusterId":"dev-k3d","provider":"DEV","region":"us-local","labels":{"gpu.chip":"A10"}}' \
     localhost:8081 aegis.v1.AegisPlatform/RegisterCluster
   ```

4. **Submit a workload**

   ```bash
   grpcurl -plaintext \
     -d '{"workload":{"projectId":"p-demo","queue":"default"}}' \
     localhost:8081 aegis.v1.AegisPlatform/SubmitWorkload
   ```

   The server assigns a workload ID if omitted and returns the full workload
   record with status `PENDING`.

5. **Fetch a workload by ID**

   ```bash
   grpcurl -plaintext \
     -d '{"id":"<PASTE_ID_FROM_SUBMIT>"}' \
     localhost:8081 aegis.v1.AegisPlatform/GetWorkload
   ```

6. **List workloads for the project**

   ```bash
   grpcurl -plaintext \
     -d '{"projectId":"p-demo"}' \
     localhost:8081 aegis.v1.AegisPlatform/ListWorkloads
   ```

   If the list is empty right after starting the API, double-check that you
   passed `projectId` on submit and that the API did not restart (the in-memory
   store resets on restart).

## Tips

- Use `grpcurl -plaintext localhost:8081 list` to enumerate services, then
  `describe` for payload details.
- All fields are camelCase per proto definitions; JSON payloads mirror that.
- Zap logs now surface invalid payloads with `codes.InvalidArgument` and include
  identifying metadata (project IDs, cluster IDs, workload IDs).
- Restarting the API clears the in-memory store; re-run the sequence after a
  restart.
