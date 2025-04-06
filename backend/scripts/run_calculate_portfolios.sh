#!/bin/bash

# Set environment variables
export DATABASE_URL="postgresql://calcutta:calcutta@localhost:5432/calcutta?sslmode=disable"

# Change to the backend directory (in case script is run from elsewhere)
cd "$(dirname "$0")/.."

# Check if calcutta ID is provided
if [ -z "$1" ]; then
  echo "Error: Calcutta ID is required"
  echo "Usage: $0 <calcutta_id>"
  exit 1
fi

# Run the calculate-portfolios command
echo "Calculating portfolios for Calcutta $1..."
go run cmd/calculate-portfolios/main.go -calcutta "$1" 