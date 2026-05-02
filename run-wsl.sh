#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="${ROOT_DIR}/backend"
FRONTEND_DIR="${ROOT_DIR}/frontend"
NGINX_CONF="${FRONTEND_DIR}/nginx.wsl.conf"
RUNTIME_DIR="${HOME}/.codex-run/devops-console-main"
BACKEND_LOG="${RUNTIME_DIR}/backend.wsl.log"
BACKEND_BIN="${RUNTIME_DIR}/server"
BACKEND_PID_FILE="${RUNTIME_DIR}/backend.wsl.pid"

mkdir -p "${RUNTIME_DIR}"

echo "[1/4] checking mysql and redis..."
systemctl is-active --quiet mysql || {
  echo "mysql is not running" >&2
  exit 1
}
systemctl is-active --quiet redis-server || {
  echo "redis-server is not running" >&2
  exit 1
}

echo "[2/4] building frontend..."
cd "${FRONTEND_DIR}"
bash ./build-wsl.sh

echo "[3/4] starting backend..."
pkill -f "devops-console-backend/cmd/server" 2>/dev/null || true
pkill -f "${BACKEND_BIN}" 2>/dev/null || true
if [[ -f "${BACKEND_PID_FILE}" ]]; then
  OLD_BACKEND_PID="$(cat "${BACKEND_PID_FILE}")"
  kill "${OLD_BACKEND_PID}" 2>/dev/null || true
  rm -f "${BACKEND_PID_FILE}"
fi
OLD_BACKEND_PID="$(ss -tnlp | awk '/:18081/ { if (match($0, /pid=[0-9]+/)) { print substr($0, RSTART + 4, RLENGTH - 4); exit } }')"
if [[ -n "${OLD_BACKEND_PID}" ]]; then
  kill "${OLD_BACKEND_PID}" 2>/dev/null || true
  sleep 1
fi
cd "${BACKEND_DIR}"
go build -o "${BACKEND_BIN}" ./cmd/server
nohup env \
  DB_HOST=127.0.0.1 \
  DB_PORT=3306 \
  DB_USER=devops \
  DB_PASSWORD=devops123456 \
  DB_NAME=devops_console \
  REDIS_HOST=127.0.0.1 \
  REDIS_PORT=6379 \
  REDIS_PASSWORD= \
  "${BACKEND_BIN}" > "${BACKEND_LOG}" 2>&1 < /dev/null &
echo $! > "${BACKEND_PID_FILE}"

BACKEND_READY=0
for _ in $(seq 1 30); do
  if curl --noproxy "*" -fsS http://127.0.0.1:18081/health >/dev/null; then
    BACKEND_READY=1
    break
  fi
  sleep 1
done

if [[ "${BACKEND_READY}" -ne 1 ]]; then
  echo "backend failed to start, see ${BACKEND_LOG}" >&2
  exit 1
fi

echo "[4/4] starting nginx..."
nginx -s stop -c "${NGINX_CONF}" 2>/dev/null || true
nginx -c "${NGINX_CONF}"

echo
echo "DevOps Console main is running:"
echo "  frontend: http://127.0.0.1:18088"
echo "  backend : http://127.0.0.1:18081/health"
echo "  backend log: ${BACKEND_LOG}"
