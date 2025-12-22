#!/bin/bash

set -e

if [ -f .env ]; then
    source .env
fi

DB_PORT=${DB_PORT:-5432}
DB_NAME=${DB_NAME:-calcutta}
DB_USER=${DB_USER:-calcutta}
DB_PASSWORD=${DB_PASSWORD:-calcutta}
DB_HOST=localhost

if [ -z "$DATABASE_URL" ]; then
    export DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
fi

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
DATA_DIR="$ROOT_DIR/data/kenpom"

if [ ! -d "$BACKEND_DIR" ]; then
    echo "Error: backend directory not found at $BACKEND_DIR"
    exit 1
fi

HAS_DATA_DIR=false
for arg in "$@"; do
    if [ "$arg" = "--data-dir" ] || [[ "$arg" == --data-dir=* ]]; then
        HAS_DATA_DIR=true
        break
    fi
done

if [ "$HAS_DATA_DIR" = true ]; then
    (cd "$BACKEND_DIR" && go run ./cmd/import-kenpom "$@")
else
    (cd "$BACKEND_DIR" && go run ./cmd/import-kenpom --data-dir "$DATA_DIR" "$@")
fi
