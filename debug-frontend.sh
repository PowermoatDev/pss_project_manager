#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FRONTEND_DIR="$SCRIPT_DIR/frontend"

FRONTEND_HOST="${FRONTEND_HOST:-0.0.0.0}"
FRONTEND_PORT="${FRONTEND_PORT:-5173}"
API_HOST="${API_HOST:-127.0.0.1}"
API_PORT="${API_PORT:-8080}"

show_help() {
  cat <<'EOF'
Usage: ./debug-frontend.sh

Starts the Vue/Vite frontend in local debug mode.

Defaults:
  FRONTEND_HOST=0.0.0.0
  FRONTEND_PORT=5173
  API_HOST=127.0.0.1
  API_PORT=8080

Notes:
  - Current Vite config proxies /api and /uploads to 127.0.0.1:8080.
  - Start the Go backend first, or update frontend/vite.config.js if you need another API port.

Example:
  ./debug-frontend.sh
EOF
}

port_in_use() {
  if ! command -v ss >/dev/null 2>&1; then
    return 1
  fi

  ss -lHtn | awk '{print $4}' | grep -Eq "[:.]${FRONTEND_PORT}$"
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  show_help
  exit 0
fi

if [[ ! -d "$FRONTEND_DIR" ]]; then
  echo "Cannot find frontend directory: $FRONTEND_DIR" >&2
  exit 1
fi

if port_in_use; then
  echo "Port ${FRONTEND_PORT} is already in use." >&2
  echo "Stop the existing frontend dev server or change the port in frontend/package.json." >&2
  exit 1
fi

if [[ "$API_HOST" != "127.0.0.1" || "$API_PORT" != "8080" ]]; then
  echo "This script does not rewrite Vite proxy settings." >&2
  echo "Current frontend proxy still points to 127.0.0.1:8080 in frontend/vite.config.js." >&2
fi

echo "Starting frontend debug server..."
echo "Frontend: $FRONTEND_DIR"
echo "Open: http://${FRONTEND_HOST}:${FRONTEND_PORT}"
echo "Proxy API target: http://${API_HOST}:${API_PORT}"

cd "$FRONTEND_DIR"
exec npm run dev
