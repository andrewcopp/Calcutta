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
make test                                            # All tests (backend + frontend + data science)
make backend-test                                    # All Go unit tests (fast, no DB)
make backend-integration-test                        # Go integration tests (requires Docker)
make frontend-test                                   # All frontend tests (vitest)
make ds-test                                         # All Python tests (pytest)
go -C backend test ./...                             # Alternative (unit only)
go -C backend test -v ./internal/app/bracket/...    # Specific package
go -C backend test -v ./internal/app/bracket/ -run TestThatBracketBuilderGeneratesSameGameIDs  # Single test

cd data-science && source .venv/bin/activate && pytest  # Python tests
cd frontend && npm test                                  # Frontend tests
cd frontend && npm run lint                              # Frontend linting
```

### Code Generation (SQLC)
```bash
make sqlc-generate      # Regenerate SQL type-safe wrappers from queries
```

SQLC pipeline: write SQL in `backend/internal/adapters/db/sqlc/queries/*.sql` → run `make sqlc-generate` → generated Go in `backend/internal/adapters/db/sqlc/*.sql.go`. Query annotations (`-- name: GetCalcutta :one`) control generated method signatures. Config in `backend/sqlc.yaml`.

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
- `backend/internal/transport/httpserver/` - HTTP handlers (`handlers_*.go`), DTOs (`dtos/`), middleware (`middleware/`)
- `backend/internal/ports/` - Interface contracts (CalcuttaRepository, EntryRepository, etc.)
- `backend/internal/models/` - Domain structs mirroring database tables
- `backend/internal/auth/` - Token manager, session/API key/chain authenticators
- `backend/internal/platform/` - Config loading, logger, pgx pool, mailer
- `backend/migrations/schema/` - Versioned SQL migrations (YYYYMMDDHHMMSS format)
- `frontend/src/` - React components, pages, hooks, services
- `data-science/moneyball/` - db, lab, models, utils
- `data-science/scripts/` - Standalone Python scripts

### Layering (dependency direction)
```
cmd/* (handlers/CLIs) → internal/app/* (services) → internal/ports (interfaces) → internal/adapters/* → internal/models
```

Services depend on small interfaces in `internal/ports`, not concrete implementations. Lower layers never import higher layers.

### Wiring & Dependency Injection

`backend/internal/app/bootstrap/app.go` creates the `App` struct with all services. Each service takes a `Ports` struct of interfaces:

```go
// Example: calcutta service depends on port interfaces, not concrete repos
type Ports struct {
    Calcuttas       ports.CalcuttaRepository
    Entries         ports.EntryRepository
    Payouts         ports.PayoutRepository
    PortfolioReader ports.PortfolioReader
}
```

Adapters in `internal/adapters/db/` implement these interfaces, wrapping SQLC-generated queries and mapping to domain models.

### Code Patterns
- **Handlers:** parse input → call service → `response.WriteJSON(w, status, dto)` or `httperr.WriteFromErr(w, r, err, userID)`
- **Services:** business logic only, no logging, return errors
- **Adapters:** wrap SQLC queries, convert between SQLC types and domain models
- **DTOs:** `transport/httpserver/dtos/` with `NewXxxResponse(model)` mapper constructors
- **File size:** <200-400 LOC, split by responsibility
- **Function size:** 20-60 LOC (orchestration may be longer)

### Database Schemas
- **`core.*`** - Production tables (calcuttas, entries, teams, tournaments, users)
- **`derived.*`** - Simulation/sandbox infrastructure (simulation_runs, simulated_calcuttas)
- **`lab.*`** - R&D schema (investment_models, entries, evaluations)
- **`archive.*`** - Deprecated tables

### Workers

`backend/cmd/workers/main.go` runs background jobs. Two workers:
- **TournamentImportWorker** - Polls DB for import jobs, executes bracket imports
- **LabPipelineWorker** - Polls DB for evaluation jobs, orchestrates Python ML scripts

Worker container mounts both `backend/` and `data-science/`, installing Python deps at startup. Workers report progress via `DBProgressWriter` and shut down gracefully on SIGTERM/SIGINT.

### Frontend Stack
- **Vite** + **TypeScript** + **React Router** v6
- **React Query** (`@tanstack/react-query`) for server state with `queryKeys.ts` cache key registry
- **Tailwind CSS** + **Radix UI** primitives for styling and accessibility
- **Recharts** for data visualization
- API client in `frontend/src/api/apiClient.ts` with auto-refresh on 401 (Bearer token + HTTP-only refresh cookie)
- Domain services in `frontend/src/services/` call apiClient methods

### Auth
Two modes controlled by `AUTH_MODE` env var:
- **`legacy`** (default, development) - Local email/password with JWT tokens
- **`cognito`** - AWS Cognito integration

Auth chain in `internal/auth/`: `ChainAuthenticator` tries `SessionAuthenticator` (cookie) then `APIKeyAuthenticator`. Middleware extracts user from token and injects into request context.

### API Error Envelope
```json
{"error": {"code": "validation_error", "message": "...", "field": "...", "requestId": "..."}}
```

Status codes: 200, 201, 204, 400, 401, 403, 404, 409, 423 (business rule lock), 500.

## Testing Conventions (Strictly Enforced)

**These are non-negotiable and will cause code review rejection if violated:**

1. **One assertion per test** - Each test has exactly one reason to fail
2. **`TestThat{Scenario}` naming** - Descriptive behavior-focused names
3. **GIVEN/WHEN/THEN structure** - Clear setup, action, assertion
4. **Deterministic tests** - Fix time/randomness, sort before comparing

### When to Unit Test (Default)
- Pure functions (validation, computation, mapping, status derivation)
- Business logic that takes domain models and returns domain models
- Anything testable without I/O
- No build tag needed — runs with `make backend-test`

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

### When to Integration Test
- Adapter methods that execute SQL (CRUD, transactions, complex queries)
- Constraint enforcement (unique violations, FK checks)
- Multi-table joins and CTEs where SQL correctness is the concern
- Transaction atomicity (ReplaceEntryTeams, ReplacePayouts)

**Integration test conventions:**
- Build tag: `//go:build integration` — runs only with `make backend-integration-test`
- Same naming: `TestThat{Scenario}`, GIVEN/WHEN/THEN, one assertion per test
- `TestMain` per package for shared container (via `testutil.StartPostgresContainer`)
- `t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })` for isolation
- Seed prerequisite data using repository methods, not raw SQL
- File naming: `*_integration_test.go`

### Frontend Tests
Frontend tests follow the same conventions using Vitest:
```typescript
describe('createEmptySlot', () => {
  it('returns empty teamId', () => {
    const slot = createEmptySlot();
    expect(slot.teamId).toBe('');
  });
});
```

## Database Rules

- **All schema changes via versioned migrations** - No ad-hoc DDL on shared databases
- **Use `/migrate` skill** to create new migration file pairs (generates timestamped `.up.sql` and `.down.sql`)
- **Migration filenames:** `YYYYMMDDHHMMSS_description.up.sql` and `.down.sql`
- **Never modify existing migrations** - Create new ones
- **Soft deletes:** Use `deleted_at` timestamps
- **DO NOT wrap migrations in BEGIN/COMMIT** - golang-migrate auto-wraps each file in a transaction
- **Always schema-qualify** table/index/function names (e.g., `core.users`, not `users`)
- **Use `public.uuid_generate_v4()`** (fully qualified) since migrations set `search_path = ''`

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
- `docs/standards/api_conventions.md` - HTTP API standards and error envelope
- `docs/reference/rules.md` - Business logic and game rules
- `docs/runbooks/moneyball_pipeline_usage.md` - Data science workflow
