#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$SCRIPT_DIR/backend"

APP_HOST="${APP_HOST:-0.0.0.0}"
APP_PORT="${APP_PORT:-8080}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-1433}"
DB_NAME="${DB_NAME:-PrintSecWarRoom}"
DB_USER="${DB_USER:-sa}"
DB_PASSWORD="${DB_PASSWORD:-YourStrong!Passw0rd}"
AUTO_MIGRATE="${AUTO_MIGRATE:-true}"
UPLOAD_DIR="${UPLOAD_DIR:-uploads}"
ALLOWED_ORIGIN="${ALLOWED_ORIGIN:-*}"

DATABASE_URL="${DATABASE_URL:-sqlserver://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}?database=${DB_NAME}&encrypt=disable}"

show_help() {
  cat <<'EOF'
Usage: ./debug-go.sh

Starts the Go backend in local debug mode with sensible defaults.

Optional environment overrides:
  APP_HOST
  APP_PORT
  DB_HOST
  DB_PORT
  DB_NAME
  DB_USER
  DB_PASSWORD
  DATABASE_URL
  AUTO_MIGRATE
  UPLOAD_DIR
  ALLOWED_ORIGIN

Example:
  APP_PORT=8081 ./debug-go.sh
EOF
}

port_in_use() {
  if ! command -v ss >/dev/null 2>&1; then
    return 1
  fi

  ss -lHtn | awk '{print $4}' | grep -Eq "[:.]${APP_PORT}$"
}

if [[ "${1:-}" == "-h" || "${1:-}" == "--help" ]]; then
  show_help
  exit 0
fi

if [[ ! -d "$BACKEND_DIR" ]]; then
  echo "Cannot find backend directory: $BACKEND_DIR" >&2
  exit 1
fi

if port_in_use; then
  echo "Port ${APP_PORT} is already in use." >&2
  echo "If Docker backend is running, stop it and keep only mssql for Go debugging." >&2
  echo "Or run with a different port, for example: APP_PORT=8081 ./debug-go.sh" >&2
  exit 1
fi

echo "Starting Go backend debug server..."
echo "Backend: $BACKEND_DIR"
echo "Listen: http://${APP_HOST}:${APP_PORT}"
echo "Database: ${DB_HOST}:${DB_PORT}/${DB_NAME}"

cd "$BACKEND_DIR"
exec env \
  APP_HOST="$APP_HOST" \
  APP_PORT="$APP_PORT" \
  DATABASE_URL="$DATABASE_URL" \
  AUTO_MIGRATE="$AUTO_MIGRATE" \
  UPLOAD_DIR="$UPLOAD_DIR" \
  ALLOWED_ORIGIN="$ALLOWED_ORIGIN" \
  go run .
