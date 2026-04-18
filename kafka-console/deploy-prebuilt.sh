#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ACTION="${1:-up}"
if [[ $# -gt 0 ]]; then
  shift
fi

require_prebuilt_assets() {
  [[ -f "$ROOT_DIR/backend/devops" ]] || {
    printf '[deploy-prebuilt][error] 缺少预构建后端二进制: %s\n' "$ROOT_DIR/backend/devops" >&2
    exit 1
  }
  [[ -f "$ROOT_DIR/backend/Dockerfile.prebuilt" ]] || {
    printf '[deploy-prebuilt][error] 缺少 backend/Dockerfile.prebuilt\n' >&2
    exit 1
  }
  [[ -f "$ROOT_DIR/frontend/dist/index.html" ]] || {
    printf '[deploy-prebuilt][error] 缺少前端构建产物: %s\n' "$ROOT_DIR/frontend/dist/index.html" >&2
    exit 1
  }
  [[ -f "$ROOT_DIR/frontend/Dockerfile.prebuilt" ]] || {
    printf '[deploy-prebuilt][error] 缺少 frontend/Dockerfile.prebuilt\n' >&2
    exit 1
  }
}

if [[ "$ACTION" == "up" || "$ACTION" == "install" ]]; then
  require_prebuilt_assets
fi

COMPOSE_FILE="$ROOT_DIR/docker-compose.prebuilt.yml" bash "$ROOT_DIR/deploy.sh" "$ACTION" "$@"
