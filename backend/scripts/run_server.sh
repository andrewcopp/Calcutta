#!/bin/bash

# Set environment variables
export DATABASE_URL="postgresql://calcutta:calcutta@localhost:5432/calcutta?sslmode=disable"
export PORT="8080"
export NODE_ENV="development"

# Change to the backend directory (in case script is run from elsewhere)
cd "$(dirname "$0")/.."

# Run the server
echo "Starting server on port $PORT..."
go run ./cmd/api