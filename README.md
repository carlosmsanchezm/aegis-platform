# aegis

Local development quickstart lives in the make targets and scripts under
`scripts/dev`. For direct API pokes, see
`docs/grpcurl-cheatsheet.md` for step-by-step `grpcurl` examples compatible with
the current proto definitions.

> `make proto` expects `buf`, `protoc-gen-go`, and `protoc-gen-go-grpc` to be
> available in `./bin` (or on your `PATH`). The `bin/` directory is ignored in
> git so you can manage tool versions locally.

## Prompt 1 Validation Steps

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


---

---

## Week 1 Validation Steps — Placement + Lease/Ack (Prompt 2)

Goal: Control plane **places** workloads to a cluster based on **regions + flavor + lowest TTFG**; agent **leases** and **acks** to **SUCCEEDED**.

> Prereqs: use the make guards (notes at bottom). Memstore resets on API restart.

### 0) Clean slate (optional)
```bash
pkill -f services/platform-api || true
pkill -f k8s-agent/cmd/agent   || true
lsof -nP -iTCP:8081 -sTCP:LISTEN
````

### 1) Generate + build + basic static checks

```bash
PATH="$(pwd)/bin:$PATH" BUF_CACHE_DIR="$(pwd)/.bufcache" make proto
make tidy  ALLOW_NET=1
make build ALLOW_NET=1
make test  ALLOW_NET=1
(cd services/platform-api && go list ./...)
(cd agents/k8s-agent    && go list ./...)
```

### 2) Run the API (Terminal A)

```bash
make run-api ALLOW_SOCKETS=1
```

### 3) Quick path (single agent, DRY-RUN executor) — works without a GPU

Terminal B:

```bash
AEGIS_EXECUTOR=job \
AEGIS_DRY_RUN=1 \
AEGIS_NAMESPACE=default \
AEGIS_TTFG_P50=15 \
make run-agent ALLOW_SOCKETS=1 \
  AEGIS_CLUSTER_ID=dev-gke AEGIS_REGION=us-central
```

Create project + submit (Terminal C):

```bash
grpcurl -plaintext -d '{
  "project":{"id":"p-demo","displayName":"Demo","policy":{"regions":["us-central"]}}
}' localhost:8081 aegis.v1.AegisPlatform/CreateProject

# Use a flavor the agent advertises (see API heartbeats); default static set:
#   ["a10-mig-1g","a100-8x"]
grpcurl -plaintext -d '{
  "workload":{"projectId":"p-demo","queue":"default",
              "workspace":{"flavor":"a10-mig-1g","image":"alpine:3.19"}}
}' localhost:8081 aegis.v1.AegisPlatform/SubmitWorkload
sleep 8
grpcurl -plaintext -d '{"projectId":"p-demo"}' \
  localhost:8081 aegis.v1.AegisPlatform/ListWorkloads
# Expect: "status":"SUCCEEDED", "clusterId":"dev-gke"
```

### 3b) (Optional) Two-agent demo — placement prefers lower TTFG

Terminal B:

```bash
AEGIS_TTFG_P50=120 make run-agent ALLOW_SOCKETS=1 \
  AEGIS_CLUSTER_ID=dev-eks AEGIS_REGION=us-east
```

Terminal C:

```bash
AEGIS_TTFG_P50=30  make run-agent ALLOW_SOCKETS=1 \
  AEGIS_CLUSTER_ID=dev-gke AEGIS_REGION=us-central
```

Then create project (regions `us-east/us-central`) and submit as above; expect `dev-gke` to be chosen.

### Negative checks (optional)

Missing flavor → `InvalidArgument`:

```bash
grpcurl -plaintext -d '{
  "workload":{"projectId":"p-demo","queue":"default","workspace":{}}
}' localhost:8081 aegis.v1.AegisPlatform/SubmitWorkload
```

Unsupported flavor → `FailedPrecondition`:

```bash
grpcurl -plaintext -d '{
  "workload":{"projectId":"p-demo","queue":"default","workspace":{"flavor":"h100-80gb"}}
}' localhost:8081 aegis.v1.AegisPlatform/SubmitWorkload
```

### Notes

* Memstore resets when the API restarts → re-create the project.
* Use make guards:

  * `ALLOW_NET=1` for `tidy/build/test`
  * `ALLOW_SOCKETS=1` for `run-*`
* Agent flavor matching: submit a flavor the agent advertises (see API heartbeats).
* If you see a stray `cluster_id:"dev-1"` heartbeat, kill old agents:

  ```bash
  pkill -f 'k8s-agent.*AEGIS_CLUSTER_ID=dev-1' || true
  ```

---

## Prompt 2 Validation — Real K8s executor (DRY-RUN)

Use a local Kubernetes cluster (Docker Desktop works). `AEGIS_DRY_RUN=1` omits GPU requests so no GPU is required.

- **Terminal A**
  ```bash
  make run-api ALLOW_SOCKETS=1
  ```
- **Terminal B**
  ```bash
  AEGIS_EXECUTOR=job \
  AEGIS_DRY_RUN=1 \
  AEGIS_NAMESPACE=default \
  AEGIS_TTFG_P50=15 \
  make run-agent ALLOW_SOCKETS=1 \
    AEGIS_CLUSTER_ID=dev-gke \
    AEGIS_REGION=us-central
  ```
- **Terminal C**
  ```bash
  grpcurl -plaintext -d '{
    "project": {
      "id": "p-demo",
      "displayName": "Demo",
      "policy": {"regions": ["us-central"]}
    }
  }' localhost:8081 aegis.v1.AegisPlatform/CreateProject

  grpcurl -plaintext -d '{
    "workload": {
      "projectId": "p-demo",
      "queue": "default",
      "workspace": {
        "flavor": "a10-mig-1g",
        "image": "alpine:3.19"
      }
    }
  }' localhost:8081 aegis.v1.AegisPlatform/SubmitWorkload
  ```
- **Verify** the workload reaches `SUCCEEDED` via `ListWorkloads`.
- **Optional logs**
  ```bash
  kubectl logs -f -l aegis.workload/id=<WORKLOAD_ID> -n default --timestamps
  ```
  The Job prologue prints workload/pod/namespace/node for quick correlation.


---

## What’s new in Prompt 3 vs Prompt 2

**Prompt 2 ** proved the control-plane/agent contract:
- Placement chose a cluster by **regions + flavor + lowest TTFG**.
- Agent **leased** and immediately **acked** (simulated completion). No Kubernetes integration.

**Prompt 3 ** turns that into **real execution on Kubernetes**:
- Agent creates a **Kubernetes Job** for `workspace` workloads, **watches** it to completion, then **acks**.
- Pods print a **prologue** (workload id, pod, namespace, node, image) for easy correlation.
- Control plane **skips stale clusters** (no heartbeat within TTL) to avoid black-hole placements.
- **DRY-RUN** (`AEGIS_DRY_RUN=1`) omits GPU requests, so the same code path runs on Docker Desktop/Kind. In production, unset DRY-RUN and optionally set `AEGIS_GPU_RESOURCE_NAME` (e.g., `nvidia.com/mig-1g.10gb`).

### Demo: Stale heartbeat filter (place when fresh, refuse when stale)
> Shows that CP won’t place onto clusters that aren’t heartbeating.

- **Terminal A (CP)**
  ```bash
  make run-api ALLOW_SOCKETS=1
  ```

- **Terminal B (agent registers once, then stop)**
  ```bash
  AEGIS_EXECUTOR=none AEGIS_TTFG_P50=15 \
  make run-agent ALLOW_SOCKETS=1 \
    AEGIS_CLUSTER_ID=dev-gke AEGIS_REGION=us-central
  # Ctrl-C after you see "cluster registered" in Terminal A
  ```

- **Terminal C (create project, wait past TTL, submit)**
  ```bash
  grpcurl -plaintext -d '{
    "project":{"id":"p-demo","displayName":"Demo",
               "policy":{"regions":["us-central"]}}
  }' localhost:8081 aegis.v1.AegisPlatform/CreateProject

  sleep 50   # TTL is 45s
  grpcurl -plaintext -d '{
    "workload":{"projectId":"p-demo","queue":"default",
                "workspace":{"flavor":"a10-mig-1g","image":"alpine:3.19"}}
  }' localhost:8081 aegis.v1.AegisPlatform/SubmitWorkload
  # Expect: FailedPrecondition; CP logs "skipping stale cluster ..."
  ```

### Demo: Failure path (Job fails → FAILED)
> Shows Job executor drives the final status based on Pod outcome.

- **Terminal A (CP)**
  ```bash
  make run-api ALLOW_SOCKETS=1
  ```

- **Terminal B (agent, failing command)**
  ```bash
  AEGIS_EXECUTOR=job AEGIS_DRY_RUN=1 AEGIS_NAMESPACE=default AEGIS_TTFG_P50=15 \
  AEGIS_WORKSPACE_COMMAND='echo "[AEGIS] failing"; exit 1' \
  make run-agent ALLOW_SOCKETS=1 \
    AEGIS_CLUSTER_ID=dev-gke AEGIS_REGION=us-central
  ```

- **Terminal C (create + submit)**
  ```bash
  grpcurl -plaintext -d '{
    "project":{"id":"p-demo","displayName":"Demo",
               "policy":{"regions":["us-central"]}}
  }' localhost:8081 aegis.v1.AegisPlatform/CreateProject

  grpcurl -plaintext -d '{
    "workload":{"projectId":"p-demo","queue":"default",
                "workspace":{"flavor":"a10-mig-1g","image":"alpine:3.19"}}
  }' localhost:8081 aegis.v1.AegisPlatform/SubmitWorkload

  sleep 8
  grpcurl -plaintext -d '{"projectId":"p-demo"}' \
    localhost:8081 aegis.v1.AegisPlatform/ListWorkloads
  # Expect: "status":"FAILED"
  ```

- **Optional: Pod logs**
  ```bash
  kubectl logs -f -l aegis.workload/id=<WORKLOAD_ID> -n default --timestamps
  ```

> **Tip:** If you see a stray `cluster_id:"dev-1"` heartbeat, kill leftover agents:
>
> ```bash
> pkill -f 'k8s-agent.*AEGIS_CLUSTER_ID=dev-1' || true
> ```

