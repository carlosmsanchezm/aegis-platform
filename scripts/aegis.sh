#!/usr/bin/env bash
set -euo pipefail

# Aegis Platform Deployment CLI
# Usage:
#   ./scripts/aegis.sh deploy                          # TLS, dev profile (default)
#   ./scripts/aegis.sh deploy --hardening standard     # TLS + enterprise hardening
#   ./scripts/aegis.sh status                          # Show pod/service status
#   ./scripts/aegis.sh sync-certs                      # Refresh CA bundles
#   ./scripts/aegis.sh clean                           # Uninstall all releases
#   ./scripts/aegis.sh port-forward                    # Start port-forwards

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

# Defaults
HARDENING="${HARDENING:-dev}"
NAMESPACE="${NAMESPACE:-aegis-system}"
KUBE_CONTEXT="${KUBE_CONTEXT:-docker-desktop}"
PLATFORM_API_IMAGE="${PLATFORM_API_IMAGE:-carlosmsanchez/aegis-platform-api:dev}"
K8S_AGENT_IMAGE="${K8S_AGENT_IMAGE:-carlosmsanchez/aegis-k8s-agent:dev}"
DEPLOY_SPOKE="${DEPLOY_SPOKE:-false}"
RHBK_USERNAME="${RHBK_USERNAME:-}"
RHBK_PASSWORD="${RHBK_PASSWORD:-}"
RHBK_EMAIL="${RHBK_EMAIL:-aegis@local.test}"

PF_PLATFORM_HTTP_PORT="${PF_PLATFORM_HTTP_PORT:-10080}"
PF_PLATFORM_GRPC_PORT="${PF_PLATFORM_GRPC_PORT:-10081}"
PF_PROXY_HTTP_PORT="${PF_PROXY_HTTP_PORT:-10085}"
PF_KEYCLOAK_HTTPS_PORT="${PF_KEYCLOAK_HTTPS_PORT:-10443}"

# ── Helpers ──────────────────────────────────────────────────────────────────

log_ok()   { echo "[OK]   $*"; }
log_fail() { echo "[FAIL] $*" >&2; }
log_info() { echo "[INFO] $*"; }

split_image_repo() { echo "${1%:*}"; }
split_image_tag()  {
  if [[ "$1" == *":"* ]]; then
    echo "${1##*:}"
  else
    echo "latest"
  fi
}

# ── deploy ───────────────────────────────────────────────────────────────────

cmd_deploy() {
  log_info "Switching to $KUBE_CONTEXT context..."
  kubectl config use-context "$KUBE_CONTEXT" >/dev/null

  local platform_repo platform_tag agent_repo agent_tag
  platform_repo="$(split_image_repo "$PLATFORM_API_IMAGE")"
  platform_tag="$(split_image_tag "$PLATFORM_API_IMAGE")"
  agent_repo="$(split_image_repo "$K8S_AGENT_IMAGE")"
  agent_tag="$(split_image_tag "$K8S_AGENT_IMAGE")"

  log_info "Platform API image: $platform_repo:$platform_tag"
  log_info "K8s Agent image:    $agent_repo:$agent_tag"
  log_info "Hardening profile:  $HARDENING"

  # Clean up stale ingress webhook
  if kubectl get validatingwebhookconfiguration ingress-nginx-admission >/dev/null 2>&1; then
    log_info "Cleaning up stale ingress webhook..."
    kubectl delete validatingwebhookconfiguration ingress-nginx-admission --ignore-not-found >/dev/null 2>&1 || true
    sleep 2
  fi

  # Ensure AWS dev-relay infrastructure exists (EC2 relay, NLB, default VPC)
  # The relay enables local hub ↔ remote EKS spoke connectivity via SSH tunnel.
  # The default VPC is required by the Pulumi provisioner for EKS cluster creation.
  local relay_tf_dir="$ROOT_DIR/terraform/dev-relay"
  if [[ -d "$relay_tf_dir" ]]; then
    log_info "Ensuring AWS dev-relay infrastructure..."
    # Ensure default VPC exists (Pulumi provisioner requires it)
    if ! aws ec2 describe-vpcs --region us-east-1 --filters "Name=isDefault,Values=true" \
         --query 'Vpcs[0].VpcId' --output text 2>/dev/null | grep -q "vpc-"; then
      log_info "Creating default VPC (required by Pulumi provisioner)..."
      aws ec2 create-default-vpc --region us-east-1 >/dev/null 2>&1 || true
    fi
    (cd "$relay_tf_dir" && terraform init -input=false >/dev/null 2>&1 && \
     terraform apply -auto-approve -input=false) && \
      log_ok "AWS dev-relay infrastructure ready" || \
      log_fail "AWS dev-relay terraform failed (continuing without relay)"

    # Start the SSH tunnel in the background so the NLB can route spoke traffic
    # to local services. The tunnel must be running BEFORE any cluster provisioning.
    # The script runs in foreground mode by default, so we background it and wait
    # briefly for the tunnel to establish.
    if [[ -x "$SCRIPT_DIR/start-aws-tunnel.sh" ]]; then
      if pgrep -f "ssh.*aegis-relay.*-R.*8081" >/dev/null 2>&1; then
        log_ok "AWS relay SSH tunnel already running"
      else
        log_info "Starting AWS relay SSH tunnel..."
        nohup "$SCRIPT_DIR/start-aws-tunnel.sh" >/dev/null 2>&1 &
        sleep 5
        if pgrep -f "ssh.*aegis-relay.*-R.*8081" >/dev/null 2>&1; then
          log_ok "AWS relay SSH tunnel started (background)"
        else
          log_fail "AWS relay SSH tunnel failed (spoke connectivity may not work)"
        fi
      fi
    fi
  fi

  # RHBK pull secret
  if [[ -n "$RHBK_USERNAME" && -n "$RHBK_PASSWORD" ]]; then
    log_info "Configuring registry.redhat.io pull secret in keycloak namespace..."
    if kubectl get namespace keycloak >/dev/null 2>&1; then
      local phase
      phase=$(kubectl get namespace keycloak -o jsonpath='{.status.phase}')
      if [[ "$phase" == "Terminating" ]]; then
        log_info "Waiting for keycloak namespace to finish terminating..."
        kubectl wait --for=delete namespace/keycloak --timeout=120s >/dev/null 2>&1 || true
      fi
    fi
    kubectl create namespace keycloak --dry-run=client -o yaml | kubectl apply -f - >/dev/null
    kubectl delete secret redhat-pull-secret -n keycloak --ignore-not-found >/dev/null
    kubectl create secret docker-registry redhat-pull-secret \
      --namespace keycloak \
      --docker-server=registry.redhat.io \
      --docker-username="$RHBK_USERNAME" \
      --docker-password="$RHBK_PASSWORD" \
      --docker-email="$RHBK_EMAIL" >/dev/null
    log_ok "RHBK pull secret created"
  fi

  # Install internal PKI
  log_info "Installing internal PKI (cert-manager + step-ca + step-issuer)..."
  local step_ca_dns_file="$ROOT_DIR/.step-ca-external-dns"
  local step_ca_external_dns=""
  if [[ -f "$step_ca_dns_file" ]]; then
    step_ca_external_dns="$(cat "$step_ca_dns_file" | xargs)"
    log_info "Using step-ca external DNS: $step_ca_external_dns"
  fi

  PKI_NAMESPACE=aegis-pki \
  CERT_MANAGER_NAMESPACE=cert-manager \
  TRUST_BUNDLE_NAMESPACES=aegis-system,keycloak \
  STEP_CA_EXTERNAL_DNS_NAMES="$step_ca_external_dns" \
    "$SCRIPT_DIR/install-internal-pki.sh"

  log_info "Waiting for cert-manager webhook..."
  kubectl rollout status deployment/cert-manager-webhook -n cert-manager --timeout=180s >/dev/null
  kubectl wait --for=condition=Ready pod -l app.kubernetes.io/component=webhook -n cert-manager --timeout=180s >/dev/null

  # Relax cert-manager webhook for local dev
  log_info "Applying local cert-manager webhook fallback policy..."
  kubectl patch validatingwebhookconfiguration cert-manager-webhook --type='json' \
    -p='[{"op":"replace","path":"/webhooks/0/failurePolicy","value":"Ignore"},{"op":"replace","path":"/webhooks/0/timeoutSeconds","value":2}]' >/dev/null 2>&1 || true
  kubectl patch mutatingwebhookconfiguration cert-manager-webhook --type='json' \
    -p='[{"op":"replace","path":"/webhooks/0/failurePolicy","value":"Ignore"},{"op":"replace","path":"/webhooks/0/timeoutSeconds","value":2}]' >/dev/null 2>&1 || true

  # PKI secrets (step-ca-credentials, spoke-proxy-ca, OIDC CA) are now
  # auto-created by the pki-secrets-init Helm post-install hook.

  # Update chart dependencies
  log_info "Updating chart dependencies..."
  helm dependency update "$ROOT_DIR/charts/aegis-services" >/dev/null

  # Deploy aegis-services with retry
  log_info "Deploying aegis-services..."
  # Auto-layer relay overlay if present (generated by terraform/dev-relay).
  local relay_overlay="$ROOT_DIR/charts/aegis-services/values/local-relay.yaml"
  local relay_flag=""
  if [[ -f "$relay_overlay" ]]; then
    relay_flag="-f $relay_overlay"
    log_info "Using relay overlay: $relay_overlay"
  fi

  # Extract spoke-agent OIDC client secret from the Keycloak realm JSON.
  # This is the authoritative source — the realm import creates the client with this secret.
  local spoke_oidc_flag=""
  local spoke_secret
  spoke_secret=$(python3 -c "
import json
with open('$ROOT_DIR/charts/aegis-services/files/keycloak/aegis-realm.json') as f:
    realm = json.load(f)
for client in realm.get('clients', []):
    if client.get('clientId') == 'spoke-agent':
        print(client.get('secret', ''))
        break
" 2>/dev/null)
  if [[ -n "$spoke_secret" ]]; then
    spoke_oidc_flag="--set platformApi.env.AEGIS_SPOKE_OIDC_CLIENT_SECRET=$spoke_secret"
    log_ok "Extracted spoke-agent OIDC client secret from realm JSON"
  else
    log_fail "Could not extract spoke-agent client secret from realm JSON (provisioning will fail)"
  fi

  # Inject AWS credentials from local profile for Pulumi provisioner.
  # Credentials go into the K8s secret only — never committed to git.
  local aws_flags=""
  local aws_key_id aws_secret_key aws_account_id
  aws_key_id=$(aws configure get aws_access_key_id --profile "${AWS_PROFILE:-aegis-new}" 2>/dev/null)
  aws_secret_key=$(aws configure get aws_secret_access_key --profile "${AWS_PROFILE:-aegis-new}" 2>/dev/null)
  if [[ -n "$aws_key_id" && -n "$aws_secret_key" ]]; then
    aws_account_id=$(aws sts get-caller-identity --profile "${AWS_PROFILE:-aegis-new}" --query Account --output text 2>/dev/null || echo "")
    aws_flags="--set-string platformApi.secrets.aws-access-key-id=$aws_key_id"
    aws_flags="$aws_flags --set-string platformApi.secrets.aws-secret-access-key=$aws_secret_key"
    if [[ -n "$aws_account_id" ]]; then
      aws_flags="$aws_flags --set-string platformApi.secrets.AEGIS_DEFAULT_AWS_ACCOUNT_ID=$aws_account_id"
    fi
    log_ok "Injected AWS credentials from profile ${AWS_PROFILE:-aegis-new}"
  else
    log_fail "Could not read AWS credentials from profile ${AWS_PROFILE:-aegis-new} (Pulumi provisioning will fail)"
  fi

  local success=0
  for attempt in 1 2 3; do
    if helm upgrade --install aegis-services "$ROOT_DIR/charts/aegis-services" \
      -f "$ROOT_DIR/charts/aegis-services/values/common.yaml" \
      -f "$ROOT_DIR/charts/aegis-services/values/local.yaml" \
      -f "$ROOT_DIR/charts/aegis-services/values/local-tls.yaml" \
      -f "$ROOT_DIR/charts/aegis-services/values/local-pki.yaml" \
      $relay_flag \
      --set platformApi.image.repository="$platform_repo" \
      --set platformApi.image.tag="$platform_tag" \
      --set hardeningProfile="$HARDENING" \
      $spoke_oidc_flag \
      $aws_flags \
      --namespace "$NAMESPACE" --create-namespace \
      --wait --timeout 5m; then
        success=1
        break
    fi
    log_info "cert-manager webhook not ready (attempt $attempt/3); retrying in 15s..."
    sleep 15
  done
  if [[ $success -ne 1 ]]; then
    log_fail "Failed to deploy aegis-services after 3 attempts"
    return 1
  fi
  log_ok "aegis-services deployed"

  # If backstage is enabled, resolve ingress ClusterIP and set hostAliases
  # so the backstage pod can reach keycloak.localtest.me (which resolves to
  # 127.0.0.1 by default — useless inside a pod).
  local bs_enabled
  bs_enabled=$(helm get values aegis-services -n "$NAMESPACE" -o json 2>/dev/null \
    | python3 -c "import sys,json; print(json.load(sys.stdin).get('backstage',{}).get('enabled',False))" 2>/dev/null || echo "False")
  if [[ "$bs_enabled" == "True" ]]; then
    # Copy keycloak client secret from keycloak namespace if not present
    if ! kubectl get secret keycloak-backstage-client-secret -n "$NAMESPACE" >/dev/null 2>&1; then
      log_info "Copying keycloak-backstage-client-secret to $NAMESPACE..."
      kubectl get secret keycloak-backstage-client-secret -n keycloak -o json \
        | python3 -c "import json,sys; s=json.load(sys.stdin); s['metadata']={'name':s['metadata']['name'],'namespace':'$NAMESPACE'}; print(json.dumps(s))" \
        | kubectl apply -f - >/dev/null 2>&1 || true
    fi
    local ingress_ip
    ingress_ip=$(kubectl get svc aegis-services-ingress-nginx-controller -n "$NAMESPACE" -o jsonpath='{.spec.clusterIP}' 2>/dev/null || echo "")
    if [[ -n "$ingress_ip" && "$ingress_ip" != "None" ]]; then
      log_info "Setting backstage hostAliases: keycloak.localtest.me → $ingress_ip"
      helm upgrade aegis-services "$ROOT_DIR/charts/aegis-services" \
        -n "$NAMESPACE" --reuse-values \
        --set "backstage.hostAliases[0].ip=$ingress_ip" \
        --set "backstage.hostAliases[0].hostnames[0]=keycloak.localtest.me" \
        --set "backstage.hostAliases[0].hostnames[1]=backstage.localtest.me" \
        --set "backstage.hostAliases[0].hostnames[2]=platform-api.localtest.me" \
        --wait --timeout 3m >/dev/null 2>&1 || true
    fi
  fi

  # Patch Keycloak ingress TLS
  local kc_host_raw kc_host kc_tls_secret
  kc_host_raw=$(kubectl get keycloak aegis-services-keycloak -n keycloak -o jsonpath='{.spec.hostname.hostname}' 2>/dev/null || echo "keycloak.localtest.me")
  kc_host="${kc_host_raw#https://}"
  kc_host="${kc_host#http://}"
  kc_host="${kc_host%/}"
  kc_tls_secret=$(kubectl get keycloak aegis-services-keycloak -n keycloak -o jsonpath='{.spec.http.tlsSecret}' 2>/dev/null || echo "keycloak-tls")
  if [[ -n "$kc_host" && -n "$kc_tls_secret" ]]; then
    log_info "Patching Keycloak ingress TLS for $kc_host..."
    kubectl patch ingress aegis-services-keycloak-ingress -n keycloak --type merge \
      -p "{\"spec\":{\"tls\":[{\"hosts\":[\"$kc_host\"],\"secretName\":\"$kc_tls_secret\"}]}}" >/dev/null 2>&1 || true
  fi

  # Relax ingress-nginx webhook for local dev
  kubectl patch validatingwebhookconfiguration aegis-services-ingress-nginx-admission --type='json' \
    -p='[{"op":"replace","path":"/webhooks/0/failurePolicy","value":"Ignore"},{"op":"replace","path":"/webhooks/0/timeoutSeconds","value":2}]' >/dev/null 2>&1 || true

  # Deploy aegis-spoke (opt-in via DEPLOY_SPOKE=true or --spoke flag)
  # Local dev: IfNotPresent because Docker Desktop shares daemon with k8s.
  # For cloud spoke (EKS), use values-cloud-remote.yaml which sets Always.
  if [[ "$DEPLOY_SPOKE" == "true" ]]; then
    log_info "Deploying aegis-spoke..."
    success=0
    for attempt in 1 2 3; do
      if helm upgrade --install aegis-spoke "$ROOT_DIR/charts/aegis-spoke" \
        -f "$ROOT_DIR/charts/aegis-spoke/values.yaml" \
        -f "$ROOT_DIR/charts/aegis-spoke/values-local.yaml" \
        -f "$ROOT_DIR/charts/aegis-spoke/values-local-tls.yaml" \
        -f "$ROOT_DIR/charts/aegis-spoke/values-local-pki.yaml" \
        --set k8sAgent.image.repository="$agent_repo" \
        --set k8sAgent.image.tag="$agent_tag" \
        --set k8sAgent.image.pullPolicy=IfNotPresent \
        --set hardeningProfile="$HARDENING" \
        --namespace "$NAMESPACE" --create-namespace; then
          success=1
          break
      fi
      log_info "Ingress webhook not ready (attempt $attempt/3); retrying in 15s..."
      sleep 15
    done
    if [[ $success -ne 1 ]]; then
      log_fail "Failed to deploy aegis-spoke after 3 attempts"
      return 1
    fi
    log_ok "aegis-spoke deployed"
  else
    log_info "Skipping aegis-spoke (DEPLOY_SPOKE=false). Use --spoke or DEPLOY_SPOKE=true to include it."
  fi

  # Sync certs
  cmd_sync_certs

  # Bootstrap .env if needed
  if [[ -f "$ROOT_DIR/.env.development" && ! -f "$ROOT_DIR/.env" ]]; then
    cp "$ROOT_DIR/.env.development" "$ROOT_DIR/.env"
    log_info "Created .env from .env.development"
  fi

  log_ok "Deployment complete (hardeningProfile=$HARDENING)"
  echo ""
  echo "   Platform API gRPC: platform-api-grpc.localtest.me:443"
  echo "   Platform API HTTPS: https://platform-api.localtest.me"
  echo "   Proxy: https://proxy.localtest.me"
  echo ""
  echo "   For E2E tests with TLS:"
  echo "   export GRPC_TLS=1"
  echo "   export GRPC_TLS_SKIP_VERIFY=0"
  echo "   ./scripts/e2e-platform-api.sh"
}

# ── status ───────────────────────────────────────────────────────────────────

cmd_status() {
  log_info "Pods in $NAMESPACE:"
  kubectl get pods -n "$NAMESPACE" -o wide 2>/dev/null || log_fail "Could not list pods"
  echo ""

  log_info "Services in $NAMESPACE:"
  kubectl get svc -n "$NAMESPACE" 2>/dev/null || log_fail "Could not list services"
  echo ""

  log_info "NetworkPolicies in $NAMESPACE:"
  kubectl get networkpolicy -n "$NAMESPACE" 2>/dev/null || log_fail "Could not list network policies"
  echo ""

  log_info "Keycloak namespace pods:"
  kubectl get pods -n keycloak 2>/dev/null || log_info "No keycloak namespace"
  echo ""

  # Check hardening profile from deployed release
  local profile
  profile=$(helm get values aegis-services -n "$NAMESPACE" -o json 2>/dev/null | python3 -c "import sys,json; print(json.load(sys.stdin).get('hardeningProfile','unknown'))" 2>/dev/null || echo "unknown")
  log_info "Active hardeningProfile: $profile"
}

# ── sync-certs ───────────────────────────────────────────────────────────────

cmd_sync_certs() {
  log_info "Syncing CA bundles..."

  kubectl get secret aegis-trust-bundle -n "$NAMESPACE" -o "jsonpath={.data.ca\.crt}" | base64 --decode > "$HOME/aegis-platform-api-ca.crt"
  chmod 0644 "$HOME/aegis-platform-api-ca.crt"

  kubectl get secret aegis-trust-bundle -n keycloak -o "jsonpath={.data.ca\.crt}" | base64 --decode > "$HOME/keycloak.localtest.me.crt" 2>/dev/null || true
  chmod 0644 "$HOME/keycloak.localtest.me.crt" 2>/dev/null || true

  cp "$HOME/aegis-platform-api-ca.crt" "$HOME/aegis-local-trust.pem"
  chmod 0644 "$HOME/aegis-local-trust.pem"

  log_ok "CA bundles refreshed"
  echo "   Combined trust store: $HOME/aegis-local-trust.pem"
  echo "   Launch VS Code with TLS trust:"
  echo "     NODE_EXTRA_CA_CERTS=$HOME/aegis-local-trust.pem /Applications/Visual\\ Studio\\ Code.app/Contents/MacOS/Electron --enable-proposed-api aegis.aegis-remote \$PWD"
}

# ── clean ────────────────────────────────────────────────────────────────────

cmd_clean() {
  for release in aegis-services aegis-spoke; do
    if helm status "$release" -n "$NAMESPACE" >/dev/null 2>&1; then
      log_info "Uninstalling $release..."
      helm uninstall "$release" -n "$NAMESPACE" >/dev/null
      log_ok "$release uninstalled"
    else
      log_info "Skipping $release (not installed)"
    fi
  done

  log_info "Cleaning up ingress webhook..."
  kubectl delete validatingwebhookconfiguration ingress-nginx-admission --ignore-not-found >/dev/null 2>&1 || true
  kubectl delete namespace keycloak --ignore-not-found >/dev/null 2>&1 || true
  kubectl wait --for=delete namespace/keycloak --timeout=120s >/dev/null 2>&1 || true
  log_ok "Cleanup complete"
}

# ── port-forward ─────────────────────────────────────────────────────────────

cmd_port_forward() {
  log_info "Stopping existing port-forwards..."
  pkill -f "kubectl.*port-forward" 2>/dev/null || true
  sleep 2

  log_info "Starting port-forwards..."
  nohup kubectl --context "$KUBE_CONTEXT" -n "$NAMESPACE" port-forward svc/aegis-services-platform-api "$PF_PLATFORM_HTTP_PORT":8080 "$PF_PLATFORM_GRPC_PORT":8081 >/dev/null 2>&1 &
  nohup kubectl --context "$KUBE_CONTEXT" -n "$NAMESPACE" port-forward svc/aegis-services-proxy "$PF_PROXY_HTTP_PORT":8085 >/dev/null 2>&1 &
  nohup kubectl --context "$KUBE_CONTEXT" -n keycloak port-forward svc/aegis-services-keycloak-service "$PF_KEYCLOAK_HTTPS_PORT":8443 >/dev/null 2>&1 &
  nohup kubectl --context "$KUBE_CONTEXT" -n "$NAMESPACE" port-forward svc/aegis-services-platform-api 8081:8081 >/dev/null 2>&1 &
  nohup kubectl --context "$KUBE_CONTEXT" -n keycloak port-forward svc/aegis-services-keycloak-service 8443:8443 >/dev/null 2>&1 &
  sleep 3

  if ! pgrep -f "kubectl.*port-forward.*aegis-services-platform-api.*$PF_PLATFORM_HTTP_PORT" >/dev/null; then
    log_fail "Platform API port-forward failed. Check: kubectl -n $NAMESPACE get pods"
    return 1
  fi

  log_ok "Port-forwards started:"
  echo "  Platform API: http://localhost:$PF_PLATFORM_HTTP_PORT (HTTP) / localhost:$PF_PLATFORM_GRPC_PORT (gRPC)"
  echo "  Proxy: http://localhost:$PF_PROXY_HTTP_PORT"
  echo "  Keycloak: https://localhost:$PF_KEYCLOAK_HTTPS_PORT"
  echo ""

  if curl -s --max-time 3 "http://localhost:$PF_PLATFORM_HTTP_PORT/healthz" >/dev/null 2>&1; then
    log_ok "Platform API is reachable"
  else
    log_info "Platform API health check failed (may still be starting)"
  fi
}

# ── Main ─────────────────────────────────────────────────────────────────────

usage() {
  cat <<'EOF'
Aegis Platform Deployment CLI

Usage: ./scripts/aegis.sh <command> [options]

Commands:
  deploy           Deploy aegis-services + aegis-spoke with TLS
  status           Show pod/service/networkpolicy status
  sync-certs       Refresh CA bundles to ~/aegis-*.pem
  clean            Uninstall all Helm releases
  port-forward     Start port-forwards to local services

Options (deploy only):
  --hardening <profile>   Set hardening profile: dev (default) | standard
  --spoke                 Also deploy aegis-spoke (k8s-agent) locally (default: skip)

Environment Variables:
  HARDENING               Same as --hardening flag
  DEPLOY_SPOKE            Set to "true" to deploy aegis-spoke locally (default: false)
  PLATFORM_API_IMAGE      Platform API image (default: carlosmsanchez/aegis-platform-api:dev)
  K8S_AGENT_IMAGE         K8s Agent image (default: carlosmsanchez/aegis-k8s-agent:dev)
  NAMESPACE               Target namespace (default: aegis-system)
  KUBE_CONTEXT            kubectl context (default: docker-desktop)

Examples:
  ./scripts/aegis.sh deploy                           # Hub only (default)
  ./scripts/aegis.sh deploy --spoke                   # Hub + spoke
  ./scripts/aegis.sh deploy --hardening standard      # Enterprise hardening
  DEPLOY_SPOKE=true ./scripts/aegis.sh deploy         # Hub + spoke via env var
  ./scripts/aegis.sh status                           # Check deployment
  ./scripts/aegis.sh clean                            # Tear down
EOF
}

main() {
  if [[ $# -lt 1 ]]; then
    usage
    exit 1
  fi

  local cmd="$1"
  shift

  # Parse flags
  while [[ $# -gt 0 ]]; do
    case "$1" in
      --hardening)
        HARDENING="${2:-dev}"
        shift 2
        ;;
      --namespace)
        NAMESPACE="${2:-aegis-system}"
        shift 2
        ;;
      --spoke)
        DEPLOY_SPOKE="true"
        shift
        ;;
      --context)
        KUBE_CONTEXT="${2:-docker-desktop}"
        shift 2
        ;;
      --help|-h)
        usage
        exit 0
        ;;
      *)
        log_fail "Unknown option: $1"
        usage
        exit 1
        ;;
    esac
  done

  # Validate hardening profile
  case "$HARDENING" in
    dev|standard) ;;
    *)
      log_fail "Invalid hardening profile: $HARDENING (must be dev or standard)"
      exit 1
      ;;
  esac

  case "$cmd" in
    deploy)       cmd_deploy ;;
    status)       cmd_status ;;
    sync-certs)   cmd_sync_certs ;;
    clean)        cmd_clean ;;
    port-forward) cmd_port_forward ;;
    help|--help|-h) usage ;;
    *)
      log_fail "Unknown command: $cmd"
      usage
      exit 1
      ;;
  esac
}

main "$@"
