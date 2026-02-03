#!/usr/bin/env bash
# Deploys recsys-service to a local Kubernetes cluster (kind or minikube).
# - Builds a local Docker image (optionally Dockerfile.dev with --dev)
# - Loads it into the cluster (kind load or minikube docker-env)
# - Installs/updates the Helm chart with values.local.yaml
# - Waits for the API deployment to be ready and prints port-forward command
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
CHART_DIR="${ROOT}/charts/recsys"
VALUES_FILE="${CHART_DIR}/values.local.yaml"
NAMESPACE="${NAMESPACE:-recsys}"
RELEASE="${RELEASE:-recsys}"
IMAGE_REPO="${IMAGE_REPO:-recsys-svc}"
IMAGE_TAG="${IMAGE_TAG:-local}"
DOCKERFILE="${DOCKERFILE:-${ROOT}/api/Dockerfile}"
CONTEXT_DIR="${CONTEXT_DIR:-${ROOT}/api}"

usage() {
  echo "Usage: $(basename "$0") [--kind|--minikube] [--dev]" >&2
  echo "Environment overrides: NAMESPACE, RELEASE, IMAGE_REPO, IMAGE_TAG, DOCKERFILE, CONTEXT_DIR" >&2
}

pick_cluster() {
  if command -v kind >/dev/null 2>&1; then
    if kind get clusters >/dev/null 2>&1 && [ -n "$(kind get clusters)" ]; then
      echo "kind"
      return
    fi
  fi
  if command -v minikube >/dev/null 2>&1; then
    if minikube status >/dev/null 2>&1; then
      echo "minikube"
      return
    fi
  fi
  echo ""  # none found
}

mode="${1:-}"
case "${mode}" in
  --kind)
    CLUSTER_MODE="kind"
    ;;
  --minikube)
    CLUSTER_MODE="minikube"
    ;;
  --dev)
    DOCKERFILE="${ROOT}/api/Dockerfile.dev"
    CLUSTER_MODE="$(pick_cluster)"
    ;;
  "")
    CLUSTER_MODE="$(pick_cluster)"
    ;;
  *)
    usage
    exit 2
    ;;
esac

if [ -z "${CLUSTER_MODE}" ]; then
  echo "No Kubernetes cluster detected. Start kind or minikube first." >&2
  exit 1
fi

build_image() {
  echo "Building ${IMAGE_REPO}:${IMAGE_TAG} from ${DOCKERFILE}..."
  docker build -t "${IMAGE_REPO}:${IMAGE_TAG}" -f "${DOCKERFILE}" "${CONTEXT_DIR}"
}

if [ "${CLUSTER_MODE}" = "minikube" ]; then
  echo "Using minikube Docker daemon"
  eval "$(minikube docker-env)"
  build_image
else
  echo "Using kind cluster"
  build_image
  KIND_CLUSTER="${KIND_CLUSTER:-$(kind get clusters | head -n 1)}"
  kind load docker-image "${IMAGE_REPO}:${IMAGE_TAG}" --name "${KIND_CLUSTER}"
fi

helm upgrade --install "${RELEASE}" "${CHART_DIR}" \
  --namespace "${NAMESPACE}" \
  --create-namespace \
  -f "${VALUES_FILE}" \
  --set api.image.repository="${IMAGE_REPO}" \
  --set api.image.tag="${IMAGE_TAG}" \
  --set api.image.pullPolicy=IfNotPresent

kubectl rollout status "deploy/${RELEASE}-api" -n "${NAMESPACE}"

echo "API service is up. Port-forward with:"
echo "  kubectl port-forward svc/${RELEASE}-api 8000:8000 -n ${NAMESPACE}"
