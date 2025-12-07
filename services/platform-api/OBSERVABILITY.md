# Observability stack

The platform now installs a minimal metrics + logging stack into every provisioned AWS cluster. Everything stays inside the cluster and is meant to be consumed via backend proxies—not exposed publicly.

## Components
- Metrics: `kube-prometheus-stack` with small resource requests and 24h retention.
- Metrics: `metrics-server` (1 replica) scraped via ServiceMonitor.
- Logs: `loki` single-binary (filesystem storage, 7d retention, analytics disabled) in namespace `aegis-logging`, `ClusterIP` service only.
- Logs: `fluent-bit` DaemonSet shipping container logs and Kubernetes events into Loki with labels for `cluster`, `namespace`, `pod`, `container`, `app`, and event `reason/type`.
- Chart values: `services/platform-api/config/observability/values-mvp.yaml`, `metrics-server-values.yaml`, `loki-values.yaml`, and `fluent-bit-values.yaml`.
- Namespaces: metrics in `aegis-observability`; logging in `aegis-logging`.

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

Backend/UI proxies should build URLs as `http://<service>.<namespace>.svc:<port>` using these values. No public exposure.

## Platform API observability endpoints
- `POST /api/metrics/query` – proxy to Prometheus `query_range`. Body:
  ```json
  {
    "projectId": "demo",
    "clusterId": "aegis-use1-dev",
    "query": "rate(container_cpu_usage_seconds_total[5m])",
    "start": "2024-12-01T12:00:00Z",
    "end": "2024-12-01T12:10:00Z",
    "stepSeconds": 30
  }
  ```
  Response:
  ```json
  {
    "series": [
      {
        "labels": {"pod": "platform-api-123"},
        "samples": [
          {"timestamp": "2024-12-01T12:00:00Z", "value": 0.12},
          {"timestamp": "2024-12-01T12:00:30Z", "value": 0.10}
        ]
      }
    ]
  }
  ```

- `POST /api/logs/query` – proxy to Loki `query_range`. Body:
  ```json
  {
    "projectId": "demo",
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
  Response:
  ```json
  {
    "entries": [
      {
        "timestamp": "2024-12-01T12:03:21.123456Z",
        "namespace": "platform-api",
        "pod": "platform-api-7f9c5b9c8f-abcde",
        "container": "platform-api",
        "message": "handler failed: 500 ...",
        "labels": {
          "cluster": "aegis-use1-dev",
          "namespace": "platform-api",
          "pod": "platform-api-7f9c5b9c8f-abcde",
          "container": "platform-api"
        }
      }
    ],
    "nextCursor": "2024-12-01T12:03:21.123456Z"
  }
  ```

- `GET /api/traces/{traceId}?projectId=<pid>&clusterId=<cid>` – proxy to Tempo trace lookup. Responds with the Tempo JSON payload.

- `GET /api/alerts?projectId=<pid>&clusterId=<cid>` – proxy to Alertmanager `api/v2/alerts`. Responds with normalized alert list:
  ```json
  {
    "alerts": [
      {
        "state": "firing",
        "labels": {"alertname": "KubePodCrashLooping"},
        "annotations": {"summary": "Pod crash looping"},
        "startsAt": "2024-12-01T12:05:00Z"
      }
    ]
  }
  ```

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
4. Checks:
   - `kubectl get pods,svc -n aegis-observability`
   - `kubectl get pods,svc -n aegis-logging`
   - Port-forward Loki: `kubectl -n aegis-logging port-forward svc/aegis-logging-loki 3100:3100` and query logs: `curl -G "http://127.0.0.1:3100/loki/api/v1/query" --data-urlencode 'query={cluster="local-demo"} |= "fluent-bit"'`
   - Confirm fluent-bit DaemonSet status: `kubectl -n aegis-logging get ds aegis-logging-fluentbit`

## How to validate in AWS (EKS)
1. Point kube context to EKS with the provisioning role (example):
   ```bash
   aws eks update-kubeconfig --name <cluster> --region <region> --role-arn arn:aws:iam::567751785679:role/aegis-platform --profile aegis
   ```
2. Provision via ProjectInfra with `addons.observability: true` (default). The AWS runner installs metrics + logging to `aegis-observability` and `aegis-logging`.
3. Validate:
   - `kubectl get pods,svc -n aegis-observability`
   - `kubectl get pods,svc -n aegis-logging`
   - Port-forward Loki and run a query as in the local section to confirm labels (`cluster`, `namespace`, `pod`, `event_reason`, `event_type`).


Outputs for backend/UI remain the same (namespace + service names/ports + alertmanager secret) and are exported via Pulumi stack outputs and `ClusterOutput.Observability`.

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
- Values files: packaged under `/services/platform-api/config/observability` in the image; override via `AEGIS_OBSERVABILITY_VALUES_FILE` and `AEGIS_METRICS_SERVER_VALUES_FILE` if needed.
