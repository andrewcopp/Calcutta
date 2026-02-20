# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Calcutta is a March Madness investment pool platform where participants bid on teams in blind auction format. Monorepo with three services: Go backend, React frontend, Python data science.

## Commands

### Development
```bash
make bootstrap          # First-time setup (copies .env, starts containers, runs migrations)
make dev                # Start containers + run migrations
make up-d               # Start containers (detached)
make down               # Stop containers
make reset              # Stop and remove volumes (fresh start)
make ops-migrate        # Run database migrations
```

### Testing
```bash
make backend-test                                    # All Go tests
go -C backend test ./...                             # Alternative
go -C backend test -v ./internal/app/bracket/...    # Specific package
go -C backend test -v ./internal/app/bracket/ -run TestThatBracketBuilderGeneratesSameGameIDs  # Single test

cd data-science && source .venv/bin/activate && pytest  # Python tests
cd frontend && npm run lint                              # Frontend linting
```

### Code Generation
```bash
make sqlc-generate      # Regenerate SQL type-safe wrappers from queries
```

### Data Science
```bash
source data-science/.venv/bin/activate
python scripts/register_investment_models.py   # Register models in lab schema
python scripts/generate_lab_predictions.py     # Generate market predictions for a model
```

### Developer Productivity
Use these instead of raw docker/psql/curl commands:
```bash
# Database
make db                              # Interactive psql shell
make query SQL="SELECT ..."          # Run SQL query
make query-file FILE="path/to.sql"   # Run SQL from file
make query-csv SQL="SELECT ..."      # Export query as CSV
make db-sizes                        # Show table sizes
make db-activity                     # Show active queries

# Logs
make logs-backend                    # Tail backend logs
make logs-worker                     # Tail worker logs
make logs-search PATTERN="error"     # Search logs for pattern
make logs-tail LINES=100             # Recent logs (all services)

# API
make api-health                      # Health check
make api-test ENDPOINT="/api/..."    # Test endpoint (supports METHOD, DATA)
```

## Architecture

### Directory Structure
- `backend/cmd/` - Runnable binaries (api, migrate, workers, tools)
- `backend/internal/app/` - Feature services (~15 domains: bracket, calcutta, scoring, analytics, etc.)
- `backend/internal/adapters/` - Database and external API implementations
- `backend/internal/transport/httpserver/` - HTTP handlers (`handlers_*.go`)
- `backend/migrations/schema/` - Versioned SQL migrations (YYYYMMDDHHMMSS format)
- `frontend/src/` - React components, pages, hooks, services
- `data-science/moneyball/` - db, lab, models, utils
- `data-science/scripts/` - Standalone Python scripts

### Layering (dependency direction)
```
cmd/* (handlers/CLIs) → internal/app/* (services) → internal/ports (interfaces) → internal/adapters/* → internal/models
```

Services depend on small interfaces in `internal/ports`, not concrete implementations. Lower layers never import higher layers.

### Code Patterns
- **Handlers:** parse input → call service → return response DTO
- **Services:** business logic only, no logging, return errors
- **File size:** <200-400 LOC, split by responsibility
- **Function size:** 20-60 LOC (orchestration may be longer)

## Testing Conventions (Strictly Enforced)

**These are non-negotiable and will cause code review rejection if violated:**

1. **One assertion per test** - Each test has exactly one reason to fail
2. **`TestThat{Scenario}` naming** - Descriptive behavior-focused names
3. **GIVEN/WHEN/THEN structure** - Clear setup, action, assertion
4. **Deterministic tests** - Fix time/randomness, sort before comparing
5. **Unit tests only** - No DB/integration tests (pure logic only)

```go
func TestThatFirstFourGameForElevenSeedHasDeterministicID(t *testing.T) {
    // GIVEN a region with two 11-seeds
    teams := createRegionWithDuplicateSeeds("East", 11, 16)
    builder := NewBracketBuilder()

    // WHEN building the regional bracket
    _, err := builder.buildRegionalBracket(bracket, "East", teams)

    // THEN First Four game for 11-seeds has ID 'East-first_four-11'
    if bracket.Games["East-first_four-11"] == nil {
        t.Error("Expected First Four game with ID 'East-first_four-11'")
    }
}
```

## Database Rules

- **All schema changes via versioned migrations** - No ad-hoc DDL on shared databases
- **Migration filenames:** `YYYYMMDDHHMMSS_description.up.sql` and `.down.sql`
- **Never modify existing migrations** - Create new ones
- **Soft deletes:** Use `deleted_at` timestamps

## Domain Units

Explicit naming prevents mixing in-game currency with real money:
- **Points** (in-game): `budget_points`, `bid_amount_points`, `payout_points`
- **Cents/Dollars** (real money): `entry_fee_cents`, `prize_amount_cents`

## Naming Conventions

Use prefixes to distinguish data types:
- `observed_` - Historical/actual data
- `predicted_` - ML model outputs
- `simulated_` - Simulation results
- `recommended_` - Strategy recommendations

## Architecture Direction

- **Python:** ML training/inference, writing predicted market artifacts
- **Go + SQL:** Simulation, evaluation, ownership math, derived reporting
- **Frontend:** Read-only viewer of pipeline runs and results

## Key Documentation

- `docs/standards/engineering.md` - Architecture principles
- `docs/standards/bracket_testing_guidelines.md` - Testing conventions (required reading)
- `docs/reference/rules.md` - Business logic and game rules
- `docs/runbooks/moneyball_pipeline_usage.md` - Data science workflow
