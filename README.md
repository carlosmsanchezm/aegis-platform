# aegis

`aegis` is a multi-cluster workload scheduling system. It consists of a central `platform-api` and a Kubernetes `operator` that runs in each target cluster. Users submit workloads to the platform API, which then places them onto the best available cluster by creating an `AegisWorkload` custom resource. The operator in the target cluster reconciles this resource, creating Kubernetes Jobs and managing their lifecycle.
push
## Local Development Quickstart

This guide will get you running with your Docker Desktop Kubernetes cluster, the `platform-api`, and one `operator` instance.

### Prerequisites

  * Go (1.24+)
  * Docker Desktop (with Kubernetes enabled)
  * `kubectl`
  * `grpcurl`

### 1\. Prepare Your Environment

First, clean up any previous runs and prepare your cluster.

```bash
# Stop any lingering processes
make stop
pkill -f 'k8s-agent/cmd' || true

# Delete old test resources
kubectl delete jobs --all
kubectl delete aegisworkloads --all

# Apply the CRD and prepare the kubeconfig for the API
kubectl apply -f agents/k8s-agent/config/crd/bases/aegis.yourorg.dev_aegisworkloads.yaml
cp ~/.kube/config /tmp/dev-1.kubeconfig
```

-----

### 2\. Run the System

You'll need three terminal windows for this guide.

**Terminal A: Start the Platform API**

```bash
# Set up and run the API. Leave this running.
export KUBECONFIGS_DIR=/tmp
make run-api ALLOW_SOCKETS=1
```

**Terminal B: Start the Operator**

```bash
# This operator will run jobs immediately and advertise a CPU flavor.
# Leave this running throughout all tests.
AEGIS_DISABLE_KUEUE=1 \
AEGIS_CLUSTER_ID=dev-1 \
AEGIS_FLAVORS="cpu-small" \
HEALTH_PROBE_BIND_ADDRESS=:18081 \
make run-operator ALLOW_SOCKETS=1
```

**Terminal C: Your Command Terminal**
This is where you will interact with the system. Before submitting workloads, wait about 15 seconds for the operator to register with the API. You can confirm this by seeing a `"cluster registered"` log in Terminal A.

-----

## Interactive Workspace UI & VS Code Workflow

The Backstage plugin now lets you submit interactive workspaces and connect to
them from VS Code without hand-editing SSH configs. The same flow works for
local Docker Desktop clusters and remote GPU clusters.

1. **Build and publish the VS Code-ready image** (one time):

   ```bash
   docker build -t aegis-workspace:latest workspace-images/openssh-vscode
   docker tag aegis-workspace:latest carlosmsanchez/aegis-workspace-vscode:latest
   docker push carlosmsanchez/aegis-workspace-vscode:latest
   ```

2. **Run the platform API and operator** as described in the quickstart above
   (Terminals A and B). Wait for the operator to register before continuing.

3. **Start the Backstage UI** in a new terminal:

   ```bash
   cd aegis-platform
   yarn start
   ```

   This serves the app at <http://localhost:3000>.

4. **Launch a workspace from the UI**:

   - Open <http://localhost:3000/aegis/workspaces/launch>.
   - Fill in the form (example values):
     - Workload ID: `w-demo`
     - Project ID: `p-demo`
     - Queue: `default`
     - Flavor: `cpu-small`
     - Image: `carlosmsanchez/aegis-workspace-vscode:latest`
     - Ports: `2222`
     - Environment variables: leave empty unless you need overrides.
   - Submit the form. The page shows the created workload summary.

5. **Track the workspace**:

   - Navigate to <http://localhost:3000/aegis/workloads> to see the list.
   - Click the workload ID to open the details view. The live status, pod ID, and
     VS Code connection helpers appear once the pod is running.

6. **Connect from VS Code**:

   - In the workload details panel, click *Connect*.
   - The modal shows ready-to-run snippets:
     - `ssh` command with the `aegis-connect` proxy wrapper.
     - A downloadable `~/.ssh/config` fragment.
     - A “Open in VS Code” link (`vscode://` URI) that launches Remote-SSH.
   - Accept the password prompt (`aegis123` in the sample env) and the VS Code
     server installs automatically (the workspace image already bundles
     `libstdc++` and `libgcc`).

To double-check connectivity, `kubectl exec` into the workspace pod and confirm
`/config/sshd/sshd_config` contains `AllowTcpForwarding yes`. You can also list
open sessions with `kubectl logs` on the proxy to watch the one-time JWTs being
consumed.

-----

## Validation Scenarios

The following scenarios validate the full functionality of the refactored system, mapped to the original project prompts.

### Prompt 1, 4, & 5: Full Lifecycle Validation

**Goal**: Prove the core API works, jobs run to completion, the Start/Ack bridge fires, metrics are recorded, and CPU flavors are handled correctly.

1.  **Seed the API (Terminal C)**

    ```bash
    grpcurl -plaintext -d '{"project":{"id":"p-demo"}}' localhost:8081 aegis.v1.AegisPlatform/CreateProject
    grpcurl -plaintext -d '{"flavor":{"name":"cpu-small","gpuCount":0}}' localhost:8081 aegis.v1.AegisPlatform/UpsertFlavor
    grpcurl -plaintext -d '{"queue":{"name":"default","projectId":"p-demo","allowedFlavors":["cpu-small"]}}' localhost:8081 aegis.v1.AegisPlatform/UpsertQueue
    ```

2.  **Submit Workloads (Terminal C)**

    ```bash
    # Submit one workload that will SUCCEED
    grpcurl -plaintext -d '{
      "workload":{"projectId":"p-demo","queue":"default","workspace":{"flavor":"cpu-small","image":"alpine:3.19","command":["sh","-c","echo Success!; sleep 2"]}}
    }' localhost:8081 aegis.v1.AegisPlatform/SubmitWorkload

    # Submit one workload that will FAIL
    grpcurl -plaintext -d '{
      "workload":{"projectId":"p-demo","queue":"default","workspace":{"flavor":"cpu-small","image":"alpine:3.19","command":["sh","-c","echo I will fail; exit 1"]}}
    }' localhost:8081 aegis.v1.AegisPlatform/SubmitWorkload
    ```

3.  **Wait and Verify (Terminal C)**

    ```bash
    # Give the jobs time to finish and be acknowledged
    echo "Waiting 20 seconds for jobs to complete..."
    sleep 20

    # Verify the final state
    echo "--- Final AegisWorkload Status ---"
    kubectl get aegisworkloads
    # ==> Expect one 'Succeeded' and one 'Failed' in the PHASE column.

    echo -e "\n--- Final Metrics ---"
    curl -s localhost:8080/metrics | grep "aegis_workload_acked_total"
    # ==> Expect counters for both succeeded and failed statuses.
    ```

      * **Verify API Logs (Terminal A)**: The logs will show the full `placed` -\> `start acknowledged` -\> `workload acknowledged` sequence for both the `SUCCEEDED` and `FAILED` jobs. 
      * You can also place for more metrics
      
      ```bash
      curl -s localhost:8080/metrics | grep -A1 '^# HELP aegis_workload_acked_total'
      curl -s localhost:8080/metrics | egrep 'aegis_workload_(placed|leased|queue_wait)'
      curl -si localhost:8080/metrics | head
      ```
      


-----

### Prompt 3: Stale Heartbeats

**Goal**: Prove the API will not place workloads on clusters that have stopped sending heartbeats.

1.  Start the Foundational Setup.
2.  Stop the operator in Terminal B with `Ctrl+C`.
3.  In Terminal C, wait for the TTL to expire, then try to submit a workload.
    ```bash
    # Wait 50 seconds (TTL is 45s)
    sleep 50
    grpcurl -plaintext -d '{"workload":{"projectId":"p-demo","queue":"default","workspace":{"flavor":"cpu-small","image":"alpine:3.19"}}}' localhost:8081 aegis.v1.AegisPlatform/SubmitWorkload
    ```
4.  **Result**: The command will fail with a `FailedPrecondition` error, and the API logs in Terminal A will show `"placement failed"` with the error `"no eligible cluster"`.

-----

### Prompt 2 & 5: Multi-Cluster Placement & Kueue

**Goal**: Prove placement logic prefers the best cluster and that Kueue integration is present.

1.  **Start a Second "Slow" Operator (Terminal D)**:
    ```bash
    # Terminal D
    AEGIS_DISABLE_KUEUE=1 \
    AEGIS_CLUSTER_ID=slow-cluster \
    AEGIS_FLAVORS="cpu-small" \
    AEGIS_TTFG_P50=120 \
    HEALTH_PROBE_BIND_ADDRESS=:18082 \
    make run-operator ALLOW_SOCKETS=1
    ```
2.  **Verify Placement**: Submit a workload from Terminal C. Check the API logs in Terminal A. You will see heartbeats from both operators, but the `"workload placed"` log will show `"cluster_id":"dev-1"`, proving the system chose the faster operator.
3.  **Verify Kueue Integration (Optional)**: Stop the operators. Restart one operator **without** `AEGIS_DISABLE_KUEUE=1`.
    ```bash
    # Restart Terminal B without the disable flag
    AEGIS_CLUSTER_ID=dev-1 \
    AEGIS_FLAVORS="cpu-small" \
    HEALTH_PROBE_BIND_ADDRESS=:18081 \
    make run-operator ALLOW_SOCKETS=1
    ```
    Submit a workload. The created Kubernetes Job will now be in the `Suspended` state, proving the Kueue integration is active by default.
    ```bash
    kubectl get jobs
    # NAME                 STATUS      COMPLETIONS   DURATION   AGE
    # aegis-w-xxxxxxxx   Suspended   0/1                      5s
    ```

### Prompt 6: Budget Guardrail

**Goal**: Prove the `platform-api` can enforce usage budgets, blocking workloads that would exceed their limit under a "HARD" policy and allowing them under a "SOFT" policy.

1.  **Set Up Environment (Terminal A & B)**: Ensure the `platform-api` and `operator` are running as described in the "Run the System" section.

2.  **Seed the API and Configure Budget (Terminal C)**:

    ```bash
    # Create the project
    grpcurl -plaintext -d '{"project":{"id":"p-demo"}}' localhost:8081 aegis.v1.AegisPlatform/CreateProject

    # Create a GPU flavor and assign a price per hour for it
    grpcurl -plaintext -d '{
      "flavor":{"name":"a10-mig-1g","gpuCount":1,"resourceName":"nvidia.com/mig-1g.10gb","priceUsdPerGpuHour":1.50}
    }' localhost:8081 aegis.v1.AegisPlatform/UpsertFlavor

    # Create a queue that allows the new flavor
    grpcurl -plaintext -d '{
      "queue":{"name":"default","projectId":"p-demo","allowedFlavors":["a10-mig-1g"],"defaultMaxDurationSeconds":600}
    }' localhost:8081 aegis.v1.AegisPlatform/UpsertQueue

    # Set a HARD budget of $0.50 for the queue
    grpcurl -plaintext -d '{
      "budget":{"projectId":"p-demo","queue":"default","limitUsd":0.50,"policyMode":"HARD"}
    }' localhost:8081 aegis.v1.AegisPlatform/UpsertBudget
    ```

3.  **Test "HARD" Policy (Terminal C)**: Submit a workload that is estimated to cost more than the budget.

    ```bash
    # This job's max duration (3600s) makes its estimated cost (~$1.50) exceed the $0.50 budget.
    grpcurl -plaintext -d '{
      "workload":{"projectId":"p-demo","queue":"default",
        "workspace":{"flavor":"a10-mig-1g","image":"alpine:3.19",
                     "maxDurationSeconds":3600,"command":["sh","-c","echo BIG; sleep 1"]}}
    }' localhost:8081 aegis.v1.AegisPlatform/SubmitWorkload
    ```

4.  **Verify Denial and Check Budget (Terminal C)**:

    ```bash
    # The previous command should fail as expected. Now check the metrics and budget status.
    echo "--- Metrics ---"
    curl -s localhost:8080/metrics | egrep 'aegis_budget_denied_total|aegis_workload_estimated_cost_usd'

    echo "--- GetBudget ---"
    grpcurl -plaintext -d '{"projectId":"p-demo","queue":"default"}' localhost:8081 aegis.v1.AegisPlatform/GetBudget
    ```

5.  **Expected Outcome**:

      * The second `SubmitWorkload` command will be **denied** with a `FailedPrecondition` error message: `budget exceeded: insufficient_funds`.
      * The metrics will show that the budget denial was recorded: `aegis_budget_denied_total{...} 1`.
      * Querying the budget will show the full budget remains, as the workload was never run:
        ```json
        {
          "budget": {
            "projectId": "p-demo",
            "queue": "default",
            "limitUsd": 0.5,
            "policyMode": "HARD"
          },
          "usage": {
        "remainingUsd": 0.5,
        "periodStartUtc": "2025-09-01",
        "periodEndUtc": "2025-10-01"
      }
    }
    ```

## Backstage Workspace MVP (UI → API → Operator)

We now ship a Backstage plugin that lets anyone submit a Workspace workload through the web app and watch it complete end-to-end.

### What was implemented

- **Proxy wiring** – Backstage backend exposes `/api/proxy/aegis/*` and forwards to the platform API on `http://localhost:8080` (`app-config.yaml` moved under `proxy.endpoints`).
- **Aegis plugin** – Route `/aegis` renders a form with `projectId`, `queue`, `flavor`, `image`, `command`, and optional `maxDurationSeconds` fields. Submissions call the REST gateway (`SubmitWorkload`) through the proxy.
- **Auto status polling** – After submit, the page polls `GetWorkload` until it reaches `SUCCEEDED` or `FAILED`, updating the UI card without a manual refresh.

### Runbook

1. **Start Backstage** (from `aegis-platform`):
   ```bash
   yarn start
   ```
2. **Load the page** – Visit `http://localhost:3000/aegis` and keep the default sample values (`echo Hello from Aegis; sleep 2`).
3. **Submit** – Click **Submit** and watch the card show the new workload ID. Within a few seconds the status will advance to `SUCCEEDED` (or `FAILED` if you pick a shorter deadline).
4. **Verify backend traffic** – Platform API logs will show `workload placed`, `workload start acknowledged`, and finally `workload acknowledged` when the Job finishes.

If the command never finishes, check the Kueue instructions below (Kueue may still have the Job queued).

## Operator + Kueue Integration

The operator has been updated to respect Kueue’s admission flow instead of fighting it. Jobs carry `kueue.x-k8s.io/queue-name`, start suspended, and Kueue unsuspends them when resources are available. The operator no longer strips the label or force-unsuspends when Kueue is enabled.

### Configuration knobs

- **Enable Kueue integration** when you want Kueue to queue workloads:
  ```bash
  export AEGIS_KUEUE_ENABLED=1
  export AEGIS_DISABLE_KUEUE=0
  HEALTH_PROBE_BIND_ADDRESS=:8082 make run-operator ALLOW_SOCKETS=1
  ```
  If your cluster uses a custom LocalQueue name, set `AEGIS_KUEUE_QUEUE=<queue>`; otherwise the operator reuses `AegisWorkload.Spec.Queue` (e.g., `default`).

- **Disable Kueue integration** for clusters without Kueue:
  ```bash
  export AEGIS_KUEUE_ENABLED=0
  export AEGIS_DISABLE_KUEUE=1   # default from the Makefile
  HEALTH_PROBE_BIND_ADDRESS=:8082 make run-operator ALLOW_SOCKETS=1
  ```
  In this mode Jobs are created unsuspended, queue labels are removed, and behavior matches the pre-Kueue flow.

### Testing with Kueue

1. Ensure your namespace is bound to a Kueue LocalQueue and any previous `aegis-w-*` Jobs/Workloads are deleted.
2. Start the operator with `AEGIS_KUEUE_ENABLED=1`, as shown above.
3. Submit a workload (`grpcurl`, Backstage UI, or other client).
4. Watch the Job lifecycle:
   ```bash
   kubectl get jobs -w
   ```
   - Initially `STATUS` is `Suspended 0/1`; the operator records `QueuedByKueue`.
   - When Kueue admits the workload, `Suspend` flips to `false`, Kubernetes creates pods, and the operator logs `AdmittedByKueue` and transitions the AegisWorkload to `Running`.
   - Once pods exit, the Job becomes `Completed 1/1` and the operator sends the `AckWorkload` bridge (`SUCCEEDED` or `FAILED`).

If Jobs stay suspended forever, confirm Kueue has available quota or temporarily disable Kueue (set `AEGIS_DISABLE_KUEUE=1`) to revert to the legacy behavior.
