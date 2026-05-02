#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SRC_DIR="$SCRIPT_DIR"
WORK_BASE="${HOME}/.codex-build"
WORK_DIR=""
NODE_BIN_DIR="${NODE_BIN_DIR:-/usr/local/node-v24.15.0-linux-x64/bin}"

if [[ -d "$NODE_BIN_DIR" ]]; then
  export PATH="$NODE_BIN_DIR:$PATH"
fi

if ! command -v node >/dev/null 2>&1; then
  echo "Error: node is not available on PATH." >&2
  exit 1
fi

if ! command -v npm >/dev/null 2>&1; then
  echo "Error: npm is not available on PATH." >&2
  exit 1
fi

mkdir -p "$WORK_BASE"
WORK_DIR="$(mktemp -d "${WORK_BASE}/devops-console-main-frontend.XXXXXX")"
cleanup() {
  if [[ -n "${WORK_DIR}" && -d "${WORK_DIR}" ]]; then
    rm -rf "${WORK_DIR}" 2>/dev/null || true
  fi
}
trap cleanup EXIT

mkdir -p "$WORK_DIR"
shopt -s dotglob
for item in "$SRC_DIR"/*; do
  name="$(basename "$item")"
  case "$name" in
    node_modules|dist)
      continue
      ;;
  esac
  cp -a "$item" "$WORK_DIR"/
done
shopt -u dotglob

cd "$WORK_DIR"
rm -rf node_modules dist

echo "Using Node: $(node -v)"
echo "Using npm: $(npm -v)"

npm ci
npm run build

rm -rf "$SRC_DIR/dist"
cp -r "$WORK_DIR/dist" "$SRC_DIR/dist"

echo "Build finished. dist has been synced back to:"
echo "  $SRC_DIR/dist"
