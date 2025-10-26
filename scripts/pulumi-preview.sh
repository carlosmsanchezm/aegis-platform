#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SERVICE_DIR="${ROOT_DIR}/services/platform-api"
DEFAULT_MANIFEST="${ROOT_DIR}/ops/previews/projectinfra-sample.yaml"
PLATFORM_NAMESPACE="${PLATFORM_NAMESPACE:-aegis-system}"
PLATFORM_DEPLOYMENT="${PLATFORM_DEPLOYMENT:-platform-api}"

BUILD_IMAGE=false
REFRESH_DEPLOYMENT=true
KEEP_RESOURCE=false
MANIFEST_PATH=""
TARGET_NAME=""
TARGET_NAMESPACE=""

usage() {
	cat <<'EOF'
Run a Pulumi preview against a ProjectInfra specification.

Usage: scripts/pulumi-preview.sh [options]

Options:
  --file PATH           Preview the specified ProjectInfra manifest from disk.
  --namespace NAME      Namespace of the ProjectInfra resource to preview.
  --name NAME           Name of the ProjectInfra resource to preview.
  --build-image         Build and push the latest platform-api image before preview.
  --no-deploy           Skip rolling the platform-api deployment before preview.
  --keep                Keep any applied ProjectInfra resource instead of deleting it.
  -h, --help            Show this help message and exit.

Environment variables:
  PLATFORM_NAMESPACE    Namespace containing the platform-api deployment (default: aegis-system)
  PLATFORM_DEPLOYMENT   Deployment name for platform-api (default: platform-api)
  PULUMI_PREVIEW_ARGS   Additional arguments forwarded to the CLI via make (advanced).
EOF
}

while [[ $# -gt 0 ]]; do
	case "$1" in
		--file)
			MANIFEST_PATH="$2"
			shift 2
			;;
		--namespace)
			TARGET_NAMESPACE="$2"
			shift 2
			;;
		--name)
			TARGET_NAME="$2"
			shift 2
			;;
		--build-image)
			BUILD_IMAGE=true
			shift
			;;
		--no-deploy)
			REFRESH_DEPLOYMENT=false
			shift
			;;
		--keep)
			KEEP_RESOURCE=true
			shift
			;;
		-h|--help)
			usage
			exit 0
			;;
		*)
			echo "Unknown option: $1" >&2
			usage >&2
			exit 2
			;;
	esac
done

if [[ -z "$MANIFEST_PATH" && ( -z "$TARGET_NAME" || -z "$TARGET_NAMESPACE" ) ]]; then
	echo "Either --file or both --namespace/--name must be provided." >&2
	exit 2
fi

if [[ -n "$MANIFEST_PATH" && ( -n "$TARGET_NAME" || -n "$TARGET_NAMESPACE" ) ]]; then
	echo "Ignoring --namespace/--name because --file was supplied." >&2
fi

if [[ "$BUILD_IMAGE" == true ]]; then
	echo "[pulumi-preview] Building and pushing platform-api image..."
	make -C "$SERVICE_DIR" docker-build
fi

if [[ "$REFRESH_DEPLOYMENT" == true ]]; then
	echo "[pulumi-preview] Restarting platform-api deployment ${PLATFORM_NAMESPACE}/${PLATFORM_DEPLOYMENT}"
	kubectl -n "$PLATFORM_NAMESPACE" rollout restart deployment "$PLATFORM_DEPLOYMENT"
	echo "[pulumi-preview] Waiting for platform-api rollout to complete"
	kubectl -n "$PLATFORM_NAMESPACE" rollout status deployment "$PLATFORM_DEPLOYMENT"
fi

cleanup() {
	if [[ -n "$MANIFEST_PATH" ]]; then
		return
	fi
	if [[ "$KEEP_RESOURCE" == true ]]; then
		echo "[pulumi-preview] Keeping ProjectInfra ${TARGET_NAMESPACE}/${TARGET_NAME} as requested"
		return
	fi
	if [[ -z "$TARGET_NAME" || -z "$TARGET_NAMESPACE" ]]; then
		return
	fi
	echo "[pulumi-preview] Cleaning up ProjectInfra ${TARGET_NAMESPACE}/${TARGET_NAME}"
	kubectl delete projectinfra "$TARGET_NAME" -n "$TARGET_NAMESPACE" --ignore-not-found=true
}

trap cleanup EXIT

run_preview_from_file() {
	echo "[pulumi-preview] Running preview from manifest ${MANIFEST_PATH}"
	make -C "$SERVICE_DIR" pulumi-preview PULUMI_PREVIEW_ARGS="--file ${MANIFEST_PATH} ${PULUMI_PREVIEW_ARGS:-}"
}

run_preview_from_cluster() {
	local manifest="${MANIFEST_PATH:-$DEFAULT_MANIFEST}"
	if [[ ! -f "$manifest" ]]; then
		echo "Sample manifest not found at $manifest" >&2
		exit 1
	fi
	echo "[pulumi-preview] Applying ProjectInfra manifest ${manifest}"
	kubectl apply -f "$manifest"

	TARGET_NAME="$(kubectl get -f "$manifest" -o jsonpath='{.metadata.name}')"
	TARGET_NAMESPACE="$(kubectl get -f "$manifest" -o jsonpath='{.metadata.namespace}')"
	if [[ -z "$TARGET_NAMESPACE" ]]; then
		TARGET_NAMESPACE="default"
	fi
	echo "[pulumi-preview] Running preview for ${TARGET_NAMESPACE}/${TARGET_NAME}"
	make -C "$SERVICE_DIR" pulumi-preview PULUMI_PREVIEW_ARGS="--namespace ${TARGET_NAMESPACE} --name ${TARGET_NAME} ${PULUMI_PREVIEW_ARGS:-}"
}

if [[ -n "$MANIFEST_PATH" ]]; then
	run_preview_from_file
else
	run_preview_from_cluster
fi
