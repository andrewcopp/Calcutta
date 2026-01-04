# Codebase Hygiene Plan

Goal: keep the codebase modular and approachable for new engineers by splitting “god files”, isolating repeated logic, and clarifying boundaries. This plan intentionally favors small, low-risk refactors (mostly file splits and extraction of pure helpers) over new abstractions.

## Principles

- [ ] Prefer splitting by **domain boundary** (analytics, simulation, importer) over splitting by type (utils).
- [ ] Keep public APIs stable while refactoring internals.
- [ ] Avoid “junk drawer” utility packages.
- [ ] Prefer deterministic, unit-testable pure functions for business logic.

## Work Plan

### 1) Backend: Split SQLC analytics queries

Target: `backend/internal/adapters/db/sqlc/queries/analytics.sql`

Outcome: Smaller query files organized by endpoint/domain; reduced repeated CTE patterns; easier schema evolution.

Tasks:

- [x] Split `analytics.sql` into multiple files under `.../queries/` (no query behavior changes).
- [x] Ensure sqlc generation still works and query names remain stable (or update call sites if names must change).
- [ ] (Optional) Identify repeated `entry_bids` / `team_agg` patterns that should become DB functions/views later.

### 2) Backend: Decompose simulated calcutta service

Target: `backend/internal/app/simulated_calcutta/service.go`

Outcome: Clear separation between run resolution, simulation engine, and persistence.

Tasks:

- [x] Extract run/batch/snapshot resolution into `run_resolver.go` (same package).
- [x] Extract outcome calculation logic into `engine.go` (same package, pure helpers where possible).
- [x] Extract DB writes into `writer.go`.
- [x] Keep the exported `Service` API stable.
- [x] Make detailed per-simulation outcomes ephemeral by default; persist aggregates + snapshots; allow opt-in persistence via `CALCUTTA_PERSIST_SIMULATION_DETAILS=true`.
- [x] Update ML analytics queries/sqlc so `total_simulations` does not require persisted per-simulation entry outcomes.

### 3) Backend: Introduce internal/features wrappers (hybrid architecture)

Targets:

- `backend/internal/features/*`
- `backend/internal/app/app.go`
- `backend/internal/app/bootstrap/app.go`
- `backend/cmd/*`

Outcome: Feature-first import surface (`internal/features/*`) with wrapper-first migration; existing `internal/app/*` remains composition/wiring.

Tasks:

- [x] Add thin wrapper packages under `backend/internal/features/*` for existing services.
- [x] Update app wiring + CLI entrypoints to import from `backend/internal/features/*`.
- [x] Verify backend build/test remains green.
- [x] Document `internal/app` as wiring/composition only (`backend/internal/app/README.md`).

### 4) Backend: Reorganize ops scripts

Target: `backend/ops/`

Outcome: Operational SQL scripts are easier to find and run safely.

Tasks:

- [x] Reorganize `backend/ops/*.sql` into `backend/ops/db/{admin,audits,maintenance}/*.sql`.
- [x] Update references in `Makefile` and docs.
- [x] Add `backend/ops/README.md` describing script intent + how to run.

### 5) Backend: Reduce handler boilerplate for analytics

Target: `backend/internal/transport/httpserver/handlers_analytics.go`

Outcome: Handlers read as orchestration only; DTO mapping is centralized.

Tasks:

- [x] Extract analytics DTO mapping helpers to `backend/internal/transport/httpserver/dtos/mappers_analytics.go`.
- [x] Extract common query param parsing helpers (e.g., `limit`) into a shared `params.go`.

### 6) Backend: Split importer by bundle type

Target: `backend/internal/bundles/importer/importer.go`

Outcome: Bundle import logic is discoverable; each bundle type is isolated.

Tasks:

- [x] Move `importSchools` to `import_schools.go`.
- [x] Move `importTournaments` (+ helpers) to `import_tournaments.go`.
- [x] Move `importCalcuttas` (+ helpers) to `import_calcuttas.go`.
- [x] Keep `ImportFromDir` and transaction boundaries in `importer.go`.

### 7) Frontend: Split large analytics pages into components

Targets:

- `frontend/src/pages/AnalyticsPage.tsx`
- `frontend/src/pages/TournamentAnalyticsPage.tsx`

Outcome: Pages become composition roots; logic and views are reusable.

Tasks:

- [x] Extract major sections into `frontend/src/components/analytics/*`.
- [x] Consolidate analytics-related API calls into `frontend/src/services/analyticsService.ts`.
- [x] Keep routes and query keys stable.

## Execution Order

- [x] Start with (1) `analytics.sql` split (low risk, high readability win).
- [x] Then (2) `simulated_calcutta` decomposition.
- [x] Then (3) `internal/features` wrapper migration.
- [x] Then (4) ops script re-org.
- [x] Then (5) handler mapping extraction.
- [x] Then (6) importer split.
- [x] Then (7) frontend cleanup.
