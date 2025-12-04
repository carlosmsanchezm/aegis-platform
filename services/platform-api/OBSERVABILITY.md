# Observability stack

This repository now installs a minimal in-cluster metrics stack for every provisioned AWS cluster. The stack lives entirely inside the cluster and is intended to be consumed via backend proxies rather than exposed publicly.

## Components
- `kube-prometheus-stack` (Prometheus, Alertmanager, kube-state-metrics, node-exporter) with small resource requests and 24h retention
- `metrics-server` (one replica) with a ServiceMonitor so Prometheus can scrape it
- Namespace: `aegis-observability` (created automatically)
- Services are `ClusterIP` only (no ingresses/load balancers)
- Chart values: `services/platform-api/config/observability/values-mvp.yaml` and `services/platform-api/config/observability/metrics-server-values.yaml`

## Pulumi outputs (per cluster)
`ProjectInfra.status.outputs[].observability` contains:
- `namespace`: namespace for all observability workloads (`aegis-observability`)
- `prometheusService` / `prometheusPort`: service/port for Prometheus (e.g., `aegis-obsv-<cluster>-prometheus:9090`)
- `alertmanagerService` / `alertmanagerPort`: service/port for Alertmanager (e.g., `aegis-obsv-<cluster>-alertmanager:9093`)
- `alertmanagerConfigSecret`: name of the Alertmanager config secret (for later wiring to routes/receivers)
- `metricsServerService` / `metricsServerPort`: service/port for metrics-server

Backend/UI proxy contract: build cluster-internal URLs as `http://<service>.<namespace>.svc:<port>` using the values above. No public exposure should be configured; proxies must run in-cluster.

## Alert rules
Baseline alerts are installed for:
- Node not ready
- CrashLooping pods
- Unschedulable pods
- API server latency (p99 > 1s)
- API server 5xx rate

These run alongside the kube-prometheus default rules, with heavy control-plane-only rules (scheduler/controller-manager/etcd) disabled to avoid noise on managed EKS.

## How to validate locally (docker-desktop/kind)
The observability installer is now a shared helper (`internal/provisioning/observability/stack.go`). To verify it on your local cluster without touching Helm values:

1. Ensure your kube context points at the local cluster: `kubectl config current-context` → `docker-desktop`.
2. Install using the same charts/values the Pulumi installer uses:
   ```bash
   helm upgrade --install aegis-obsv charts/kube-prometheus-stack \
     --repo https://prometheus-community.github.io/helm-charts \
     -f services/platform-api/config/observability/values-mvp.yaml \
     --namespace aegis-observability --create-namespace

   helm upgrade --install aegis-metrics metrics-server \
     --repo https://kubernetes-sigs.github.io/metrics-server/ \
     -f services/platform-api/config/observability/metrics-server-values.yaml \
     --namespace aegis-observability --create-namespace
   ```
3. Check health:
   - `kubectl get pods,svc -n aegis-observability`
   - `kubectl -n aegis-observability port-forward svc/aegis-obsv-prometheus 9090:9090` and `curl http://127.0.0.1:9090/-/healthy`
   - Alertmanager secret: `kubectl get secret alertmanager-aegis-obsv-alertmanager -n aegis-observability`

This mirrors what Pulumi will do during provisioning; it’s a quick smoke test of charts/values.

## How to validate in AWS (EKS)
1. Point kube context to EKS with the provisioning role (example):
   ```bash
   aws eks update-kubeconfig --name <cluster> --region <region> --role-arn arn:aws:iam::567751785679:role/aegis-platform --profile aegis
   ```
2. Provision via ProjectInfra with `addons.observability: true` (current default). The AWS runner calls the shared installer and deploys to `aegis-observability`.
3. Validate in the same way as local:
   - `kubectl get pods,svc -n aegis-observability`
   - Port-forward Prometheus/Alertmanager for health checks.

Outputs for backend/UI remain the same (namespace + service names/ports + alertmanager secret) and are exported via Pulumi stack outputs and `ClusterOutput.Observability`.
