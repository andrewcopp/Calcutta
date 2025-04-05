#!/bin/bash

# Set environment variables
export DB_USER=calcutta
export DB_PASSWORD=calcutta
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=calcutta

# Set the DATABASE_URL environment variable
export DATABASE_URL="postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"

# Run the Go script
cd "$(dirname "$0")"
go run seed_schools.go 