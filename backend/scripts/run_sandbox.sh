#!/bin/bash

set -e

cd "$(dirname "$0")/.."

if [ -f .env ]; then
  export $(grep -v '^#' .env | xargs)
fi

if [ -f ../.env ]; then
  export $(grep -v '^#' ../.env | xargs)
fi

if [ -z "$DATABASE_URL" ]; then
  if [ -z "$DB_HOST" ] || [ "$DB_HOST" = "db" ]; then
    DB_HOST=localhost
  fi
fi

if [ -z "$DB_PORT" ]; then
  DB_PORT=5432
fi

if [ -z "$DB_NAME" ]; then
  DB_NAME=calcutta
fi

if [ -z "$DB_USER" ]; then
  DB_USER=calcutta
fi

if [ -z "$DB_PASSWORD" ]; then
  DB_PASSWORD=calcutta
fi

if [ -z "$DATABASE_URL" ]; then
  export DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
fi

go run ./cmd/sandbox "$@"
