# Observability stack

The platform now installs a minimal metrics + logging + tracing stack into every provisioned AWS cluster. Everything stays inside the cluster and is meant to be consumed via backend proxies—not exposed publicly.

## Components
- Metrics: `kube-prometheus-stack` with small resource requests and 24h retention.
- Metrics: `metrics-server` (1 replica) scraped via ServiceMonitor.
- Logs: `loki` single-binary (filesystem storage, 7d retention, analytics disabled) in namespace `aegis-logging`, `ClusterIP` service only.
- Logs: `fluent-bit` DaemonSet shipping container logs and Kubernetes events into Loki with labels for `cluster`, `namespace`, `pod`, `container`, `app`, and event `reason/type`.
- Traces: `tempo` single-binary with local storage (48h retention), internal-only services, in namespace `aegis-tracing`.
- Traces: `opentelemetry-collector` deployment exposing OTLP gRPC/HTTP inside the cluster and exporting to Tempo.
- Chart values: `services/platform-api/config/observability/values-mvp.yaml`, `metrics-server-values.yaml`, `loki-values.yaml`, `fluent-bit-values.yaml`, `tempo-values.yaml`, and `otel-collector-values.yaml`.
- Namespaces: metrics in `aegis-observability`; logging in `aegis-logging`; tracing in `aegis-tracing`.

## Pulumi outputs (per cluster)
`ProjectInfra.status.outputs[].observability` contains:
- `namespace`: metrics namespace (`aegis-observability`)
- `prometheusService` / `prometheusPort`
- `alertmanagerService` / `alertmanagerPort`
- `alertmanagerConfigSecret`
- `metricsServerService` / `metricsServerPort`
- `lokiNamespace`: logging namespace (`aegis-logging`)
- `lokiService` / `lokiPort`: Loki HTTP service/port (single-binary gateway)
- `lokiAuthSecret`: empty when auth is disabled (default)
- `tracingNamespace`: tracing namespace (`aegis-tracing`)
- `tempoService` / `tempoPort`: Tempo query service/port (HTTP)
- `otelService` / `otelGrpcPort` / `otelHttpPort`: OTLP collector service + gRPC/HTTP ports

Backend/UI proxies should build URLs as `http://<service>.<namespace>.svc:<port>` using these values. No public exposure.

### Log query contract
The backend should proxy to Loki’s HTTP API using the outputs above.

Suggested request shape (backend endpoint `POST /logs/query`):

```
{
  "clusterId": "aegis-use1-dev",
  "namespace": "platform-api",
  "pod": "platform-api-7f9c5b9c8f-abcde",
  "substring": "error",
  "start": "2024-12-01T12:00:00Z",
  "end": "2024-12-01T12:10:00Z",
  "limit": 200,
  "cursor": ""
}
```

Translate to Loki `query_range`:

```
{cluster="aegis-use1-dev",namespace="platform-api",pod="platform-api-7f9c5b9c8f-abcde"} |= "error"
```

Sample response shape:

```
{
  "entries": [
    {
      "timestamp": "2024-12-01T12:03:21.123456Z",
      "namespace": "platform-api",
      "pod": "platform-api-7f9c5b9c8f-abcde",
      "container": "platform-api",
      "app": "platform-api",
      "eventReason": "",
      "eventType": "",
      "message": "handler failed: 500 ...",
      "labels": {
        "cluster": "aegis-use1-dev",
        "namespace": "platform-api",
        "pod": "platform-api-7f9c5b9c8f-abcde",
        "container": "platform-api",
        "app": "platform-api"
      }
    }
  ],
  "nextCursor": "<opaque Loki cursor>"
}
```

Pagination: propagate Loki `limit` and `cursor` (forward token from `query_range`). Supported filters: required `clusterId`; optional `namespace`, `pod`, `substring`; required `start`/`end` time window.

### Trace query contract
The backend should proxy to Tempo’s HTTP API (`tempoService`/`tempoPort`) while accepting OTLP writes via the OTEL collector (`otelService`/`otelGrpcPort` or `otelHttpPort`).

Suggested request shape (backend endpoint `POST /traces/query`):

```
{
  "clusterId": "aegis-use1-dev",
  "traceId": "3e5d4c8b9d12c4d7",
  "service": "platform-api",
  "operation": "POST /v1/jobs",
  "start": "2024-12-01T12:00:00Z",
  "end": "2024-12-01T12:10:00Z",
  "limit": 50,
  "minDurationMs": 0,
  "maxDurationMs": 0
}
```

Backend behavior:
- If `traceId` is present, fetch `GET /api/traces/{traceId}` from Tempo.
- Otherwise, call `POST /api/search` with filters for `service`, `operation`, `start`/`end`, optional duration bounds, and `limit`.

Sample response shape:

```
{
  "traces": [
    {
      "traceId": "3e5d4c8b9d12c4d7",
      "rootService": "platform-api",
      "rootOperation": "POST /v1/jobs",
      "start": "2024-12-01T12:03:21.100Z",
      "durationMs": 142,
      "spanCount": 8,
      "spans": [
        {
          "spanId": "fb2c4c9e0c0ba1b6",
          "parentSpanId": "",
          "service": "platform-api",
          "operation": "POST /v1/jobs",
          "start": "2024-12-01T12:03:21.100Z",
          "durationMs": 142,
          "attributes": {
            "http.status_code": 200,
            "cluster": "aegis-use1-dev"
          }
        }
      ]
    }
  ],
  "nextPageToken": ""
}
```

Pagination: honor Tempo search pagination token; surface it as `nextPageToken`.

## Alert rules
Baseline alerts are installed for:
- Node not ready
- CrashLooping pods
- Unschedulable pods
- API server latency (p99 > 1s)
- API server 5xx rate

These run alongside kube-prometheus defaults; heavy control-plane-only rules are disabled to avoid noise on managed EKS.

## How to validate locally (docker-desktop/kind)
1. Ensure your kube context points at the local cluster (`kubectl config current-context`).
2. Install metrics stack:
   ```bash
   helm upgrade --install aegis-obsv kube-prometheus-stack \
     --repo https://prometheus-community.github.io/helm-charts \
     -f services/platform-api/config/observability/values-mvp.yaml \
     --namespace aegis-observability --create-namespace

   helm upgrade --install aegis-metrics metrics-server \
     --repo https://kubernetes-sigs.github.io/metrics-server/ \
     -f services/platform-api/config/observability/metrics-server-values.yaml \
     --namespace aegis-observability --create-namespace
   ```
3. Install logging stack (adjust `CLUSTER_ID` if you want namespaced labels):
   ```bash
   helm upgrade --install aegis-logging-loki loki \
     --repo https://grafana.github.io/helm-charts \
     -f services/platform-api/config/observability/loki-values.yaml \
     --namespace aegis-logging --create-namespace

   helm upgrade --install aegis-logging-fluentbit fluent-bit \
     --repo https://fluent.github.io/helm-charts \
     -f services/platform-api/config/observability/fluent-bit-values.yaml \
     --namespace aegis-logging --create-namespace \
     --set env[0].value=local-demo \
     --set env[1].value=aegis-logging-loki \
     --set env[2].value=3100
   ```
4. Install tracing stack:
   ```bash
   helm upgrade --install aegis-tracing-tempo tempo \
     --repo https://grafana.github.io/helm-charts \
     -f services/platform-api/config/observability/tempo-values.yaml \
     --namespace aegis-tracing --create-namespace

   helm upgrade --install aegis-tracing-otel opentelemetry-collector \
     --repo https://open-telemetry.github.io/opentelemetry-helm-charts \
     -f services/platform-api/config/observability/otel-collector-values.yaml \
     --namespace aegis-tracing --create-namespace
   ```
5. Checks:
   - `kubectl get pods,svc -n aegis-observability`
   - `kubectl get pods,svc -n aegis-logging`
   - `kubectl get pods,svc -n aegis-tracing`
   - Port-forward Loki: `kubectl -n aegis-logging port-forward svc/aegis-logging-loki 3100:3100` and query logs: `curl -G "http://127.0.0.1:3100/loki/api/v1/query" --data-urlencode 'query={cluster="local-demo"} |= "fluent-bit"'`
   - Port-forward Tempo: `kubectl -n aegis-tracing port-forward svc/aegis-tracing-tempo 3200:3200` and fetch a trace once a trace ID is known: `curl http://127.0.0.1:3200/api/traces/<traceId>`
   - Send a sample trace (emits IDs in stdout): `kubectl -n aegis-tracing run telemetrygen --rm -it --image=ghcr.io/open-telemetry/telemetrygen:<tag> -- --otlp-endpoint=aegis-tracing-otel.aegis-tracing.svc:4317 --otlp-insecure --duration=20s --rate=5`
   - Confirm fluent-bit DaemonSet status: `kubectl -n aegis-logging get ds aegis-logging-fluentbit`
   - Confirm OTEL collector deployment: `kubectl -n aegis-tracing get deploy aegis-tracing-otel`

## How to validate in AWS (EKS)
1. Point kube context to EKS with the provisioning role (example):
   ```bash
   aws eks update-kubeconfig --name <cluster> --region <region> --role-arn arn:aws:iam::567751785679:role/aegis-platform --profile aegis
   ```
2. Provision via ProjectInfra with `addons.observability: true` (default). The AWS runner installs metrics + logging + tracing to `aegis-observability`, `aegis-logging`, and `aegis-tracing`.
3. Validate:
   - `kubectl get pods,svc -n aegis-observability`
   - `kubectl get pods,svc -n aegis-logging`
   - `kubectl get pods,svc -n aegis-tracing`
   - Port-forward Loki and run a query as in the local section to confirm labels (`cluster`, `namespace`, `pod`, `event_reason`, `event_type`).
   - Port-forward Tempo: `kubectl -n aegis-tracing port-forward svc/aegis-tracing-tempo 3200:3200` and fetch a known trace ID: `curl http://127.0.0.1:3200/api/traces/<traceId>`.
   - Optional: generate a trace via telemetrygen in-cluster using the OTLP collector service endpoint and confirm it appears via Tempo.

Outputs for backend/UI include the namespace + service names/ports + alertmanager secret + tracing endpoints, exported via Pulumi stack outputs and `ClusterOutput.Observability`.

## CRD / schema sync checklist
When you add fields to `ProjectInfra` or its status (e.g., observability outputs):
- Update the Go types in `api/v1alpha1` (spec/status structs).
- Regenerate or update the CRD schema in `config/crd/bases/...projectinfras.yaml` so the API server accepts the new fields.
- Apply the updated CRD to your cluster (even for dev/test) so status updates are not rejected.
- Verify the controller writes the new fields and handles nil/empty values.
- Optional: add a quick diff/lint to catch drift between Go types and the checked-in CRD.

## Enabling observability in provisioning
- Runtime flag: set `AEGIS_OBSERVABILITY_ENABLED=true` in the platform-api environment to allow installs (default is off).
- ProjectInfra addon: set `spec.addons["observability"]=true` on the ProjectInfra request. Both the env flag **and** the addon must be true for the AWS runner to install the stack.
- Values files: packaged under `/services/platform-api/config/observability` in the image; override via `AEGIS_OBSERVABILITY_VALUES_FILE`, `AEGIS_METRICS_SERVER_VALUES_FILE`, `AEGIS_LOGGING_LOKI_VALUES_FILE`, `AEGIS_LOGGING_FLUENT_BIT_VALUES_FILE`, `AEGIS_TRACING_TEMPO_VALUES_FILE`, and `AEGIS_TRACING_COLLECTOR_VALUES_FILE` if needed.
