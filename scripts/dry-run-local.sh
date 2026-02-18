#!/usr/bin/env bash
#
# Local Dry Run - Complete go-live lifecycle rehearsal
#
# This script simulates the complete deployment and setup flow locally
# using docker-compose.local-prod.yml (production builds, no source mounts).
#
# Flow:
# 1. Create a new environment (fresh containers + DB)
# 2. Get admin access (via make create-admin)
# 3. Seed the database (schools, tournaments, calcuttas)
# 4. Invite users to join (email simulation via Mailpit)
# 5. Create tournament + configure teams + bracket
# 6. Create calcutta + invite entries
# 7. Lock tournament + simulate gameplay
#
# Usage:
#   ./scripts/dry-run-local.sh

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project root
REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_ROOT"

# Load .env for DB credentials
if [ ! -f .env ]; then
  echo -e "${RED}ERROR: .env file not found. Run 'make env-init' first.${NC}"
  exit 1
fi

set -a
source .env
set +a

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Calcutta Local Dry Run${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

# Step 1: Create a new environment
echo -e "${YELLOW}[1/7] Creating new environment (fresh containers + DB)${NC}"
echo "Stopping existing containers and removing volumes..."
make prod-reset

echo "Starting production containers..."
make prod-up

# Wait for DB to be ready
echo "Waiting for database to be ready..."
sleep 5
docker compose -f docker-compose.local-prod.yml exec db pg_isready -U "$DB_USER" || {
  echo -e "${RED}Database not ready after 5 seconds. Check 'docker compose logs db'${NC}"
  exit 1
}

echo -e "${GREEN}✓ Environment created${NC}"
echo ""

# Step 2: Run migrations
echo -e "${YELLOW}[2/7] Running database migrations${NC}"
make prod-ops-migrate
echo -e "${GREEN}✓ Migrations complete${NC}"
echo ""

# Step 3: Seed the database
echo -e "${YELLOW}[3/7] Seeding database with historical data${NC}"

if [ ! -d "backend/exports/bundles" ]; then
  echo -e "${RED}ERROR: Bundle directory not found at backend/exports/bundles${NC}"
  echo "Expected structure:"
  echo "  backend/exports/bundles/"
  echo "    ├── schools.json"
  echo "    ├── tournaments/"
  echo "    └── calcuttas/"
  exit 1
fi

echo "Importing bundles (schools, tournaments, calcuttas)..."
cd backend
go run ./cmd/tools/import-bundles -in ./exports/bundles -dry-run=false
cd ..

echo -e "${GREEN}✓ Database seeded${NC}"
echo ""

# Step 4: Verify environment
echo -e "${YELLOW}[4/7] Verifying environment health${NC}"

echo "Backend health check..."
HEALTH_RESPONSE=$(curl -s http://localhost:8080/api/health || echo "FAILED")
if [[ "$HEALTH_RESPONSE" == *"ok"* ]] || [[ "$HEALTH_RESPONSE" == *"healthy"* ]]; then
  echo -e "${GREEN}✓ Backend healthy${NC}"
else
  echo -e "${RED}✗ Backend health check failed${NC}"
  echo "Response: $HEALTH_RESPONSE"
  echo "Check logs: make logs-backend"
  exit 1
fi

echo "Database row counts..."
PROD_PSQL="docker compose -f docker-compose.local-prod.yml exec -T db psql -U ${DB_USER} -d ${DB_NAME} -tAc"
SCHOOLS_COUNT=$($PROD_PSQL "SELECT COUNT(*) FROM core.schools" 2>/dev/null | tr -d '[:space:]')
TOURNAMENTS_COUNT=$($PROD_PSQL "SELECT COUNT(*) FROM core.tournaments" 2>/dev/null | tr -d '[:space:]')
CALCUTTAS_COUNT=$($PROD_PSQL "SELECT COUNT(*) FROM core.calcuttas" 2>/dev/null | tr -d '[:space:]')

echo "  Schools: $SCHOOLS_COUNT"
echo "  Tournaments: $TOURNAMENTS_COUNT"
echo "  Calcuttas: $CALCUTTAS_COUNT"

if [ "$SCHOOLS_COUNT" -gt 0 ] && [ "$TOURNAMENTS_COUNT" -gt 0 ]; then
  echo -e "${GREEN}✓ Database seeded successfully${NC}"
else
  echo -e "${RED}✗ Database seeding incomplete${NC}"
  exit 1
fi

echo ""

# Step 5: Manual testing instructions
echo -e "${YELLOW}[5/7] Manual testing (interactive)${NC}"
echo ""
echo "The environment is now ready for manual testing."
echo ""
echo "Test the following workflows:"
echo ""
echo "  a) Invite users:"
echo "     - Mailpit UI: ${BLUE}http://localhost:8025${NC}"
echo "     - Send invite: curl -X POST http://localhost:8080/api/admin/invites -d '{\"email\":\"user@test.com\"}'"
echo ""
echo "  b) Create tournament:"
echo "     - POST /api/tournaments"
echo "     - Configure teams and bracket"
echo ""
echo "  c) Create calcutta:"
echo "     - POST /api/calcuttas"
echo "     - Invite entries"
echo ""
echo "  d) Lock tournament and simulate:"
echo "     - PUT /api/tournaments/{id}/lock"
echo "     - Run simulation worker"
echo ""
echo "  e) Frontend:"
echo "     - URL: ${BLUE}http://localhost:3000${NC}"
echo "     - Test tournament creation, calcutta management"
echo ""
echo "Useful commands:"
echo "  make logs-backend        # Tail backend logs"
echo "  make logs-worker         # Tail worker logs"
echo "  make db                  # Interactive psql shell"
echo "  make api-test ENDPOINT=/api/tournaments  # Test API"
echo ""
echo -e "${YELLOW}Press ENTER when you're done testing...${NC}"
read -r

echo ""

# Step 6: Teardown
echo -e "${YELLOW}[6/7] Tearing down environment${NC}"
echo "Stopping containers..."
make prod-down
echo -e "${GREEN}✓ Containers stopped${NC}"
echo ""

# Step 7: Summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Dry Run Complete${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo "Summary:"
echo "  - Environment created and torn down successfully"
echo "  - Database seeded with $SCHOOLS_COUNT schools, $TOURNAMENTS_COUNT tournaments, $CALCUTTAS_COUNT calcuttas"
echo ""
echo "Next steps:"
echo "  1. Review any issues encountered during testing"
echo "  2. Build missing CLI tools (create-admin, create-tournament, etc.)"
echo "  3. Repeat this dry-run until the flow is smooth"
echo "  4. Decide on deployment platform (AWS vs Render/Fly)"
echo "  5. Run this same flow in staging before prod launch"
echo ""
echo -e "${GREEN}Dry run completed successfully!${NC}"
