#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PARENT_DIR="$(cd "$ROOT_DIR/.." && pwd)"
PROJECT_NAME="${COMPOSE_PROJECT_NAME_OVERRIDE:-kafka-console}"
RUNTIME_DIR="${KAFKA_CONSOLE_RUNTIME_DIR:-$PARENT_DIR/.${PROJECT_NAME}-runtime}"
RUNTIME_ENV="$RUNTIME_DIR/.env"
CURRENT_ENV="$ROOT_DIR/.env"
ACTION="${1:-install}"
if [[ $# -gt 0 ]]; then
  shift
fi

log() {
  printf '[release] %s\n' "$*"
}

warn() {
  printf '[release][warn] %s\n' "$*" >&2
}

die() {
  printf '[release][error] %s\n' "$*" >&2
  exit 1
}

require_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "缺少命令: $1"
}

random_secret() {
  local length="${1:-32}"
  if command -v openssl >/dev/null 2>&1; then
    openssl rand -hex "$length"
  else
    head -c "$length" /dev/urandom | od -An -tx1 | tr -d ' \n'
  fi
}

ensure_runtime_layout() {
  mkdir -p "$RUNTIME_DIR/data/mysql" "$RUNTIME_DIR/data/redis"
}

generate_runtime_env() {
  [[ -f "$ROOT_DIR/.env.example" ]] || die "缺少 .env.example"
  cp "$ROOT_DIR/.env.example" "$RUNTIME_ENV"

  local mysql_password redis_password jwt_secret
  mysql_password="$(random_secret 16)"
  redis_password="$(random_secret 16)"
  jwt_secret="$(random_secret 32)"

  sed -i "s/^MYSQL_ROOT_PASSWORD=.*/MYSQL_ROOT_PASSWORD=${mysql_password}/" "$RUNTIME_ENV"
  sed -i "s/^REDIS_PASSWORD=.*/REDIS_PASSWORD=${redis_password}/" "$RUNTIME_ENV"
  sed -i "s/^JWT_SECRET=.*/JWT_SECRET=${jwt_secret}/" "$RUNTIME_ENV"

  log "首次生成运行时配置: $RUNTIME_ENV"
}

ensure_runtime_env() {
  if [[ -f "$RUNTIME_ENV" ]]; then
    return 0
  fi

  if [[ -f "$CURRENT_ENV" ]]; then
    cp "$CURRENT_ENV" "$RUNTIME_ENV"
    log "已从当前发布目录复制 .env 到运行时目录"
    return 0
  fi

  generate_runtime_env
}

sync_env_to_release() {
  [[ -f "$RUNTIME_ENV" ]] || die "运行时配置不存在: $RUNTIME_ENV"
  cp "$RUNTIME_ENV" "$CURRENT_ENV"
}

export_runtime_envs() {
  export COMPOSE_PROJECT_NAME="$PROJECT_NAME"
  export MYSQL_DATA_DIR="$RUNTIME_DIR/data/mysql"
  export REDIS_DATA_DIR="$RUNTIME_DIR/data/redis"
}

verify_prebuilt_assets() {
  [[ -f "$ROOT_DIR/backend/devops" ]] || die "缺少预构建后端二进制: $ROOT_DIR/backend/devops"
  [[ -f "$ROOT_DIR/backend/Dockerfile.prebuilt" ]] || die "缺少 backend/Dockerfile.prebuilt"
  [[ -f "$ROOT_DIR/frontend/dist/index.html" ]] || die "缺少前端构建产物: $ROOT_DIR/frontend/dist/index.html"
  [[ -f "$ROOT_DIR/frontend/Dockerfile.prebuilt" ]] || die "缺少 frontend/Dockerfile.prebuilt"
}

install_release() {
  require_cmd docker
  docker info >/dev/null 2>&1 || die "Docker 服务不可用，请先启动 Docker"
  ensure_runtime_layout
  ensure_runtime_env
  sync_env_to_release
  export_runtime_envs
  verify_prebuilt_assets
  chmod +x "$ROOT_DIR/deploy.sh" "$ROOT_DIR/deploy-prebuilt.sh"
  log "使用固定项目名部署: $COMPOSE_PROJECT_NAME"
  log "运行时目录: $RUNTIME_DIR"
  bash "$ROOT_DIR/deploy-prebuilt.sh" up "$@"
}

uninstall_release() {
  require_cmd docker
  ensure_runtime_layout
  ensure_runtime_env
  sync_env_to_release
  export_runtime_envs
  chmod +x "$ROOT_DIR/deploy.sh" "$ROOT_DIR/deploy-prebuilt.sh"
  log "停止当前发布: $ROOT_DIR"
  bash "$ROOT_DIR/deploy-prebuilt.sh" down || true

  if [[ "${SELF_DELETE_RELEASE:-true}" == "true" ]]; then
    log "准备删除当前发布目录，运行时目录会保留: $RUNTIME_DIR"
    (
      sleep 1
      rm -rf "$ROOT_DIR"
    ) >/dev/null 2>&1 &
  else
    log "已停止服务，按配置保留当前发布目录"
  fi
}

show_runtime_info() {
  cat <<EOF
release.sh 用法:
  ./release.sh              默认执行 install
  ./release.sh install      安装/升级当前解压包
  ./release.sh uninstall    停止服务并删除当前解压目录
  ./release.sh runtime      查看运行时目录

固定运行时目录:
  $RUNTIME_DIR

持久化内容:
  $RUNTIME_DIR/.env
  $RUNTIME_DIR/data/mysql
  $RUNTIME_DIR/data/redis
EOF
}

case "$ACTION" in
  install|up)
    install_release "$@"
    ;;
  uninstall|down|remove)
    uninstall_release "$@"
    ;;
  runtime|info)
    show_runtime_info
    ;;
  *)
    die "不支持的动作: $ACTION"
    ;;
esac
