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

- [ ] Split `analytics.sql` into multiple files under `.../queries/` (no query behavior changes).
- [ ] Ensure sqlc generation still works and query names remain stable (or update call sites if names must change).
- [ ] (Optional) Identify repeated `entry_bids` / `team_agg` patterns that should become DB functions/views later.

### 2) Backend: Decompose simulated calcutta service

Target: `backend/internal/app/simulated_calcutta/service.go`

Outcome: Clear separation between run resolution, simulation engine, and persistence.

Tasks:

- [ ] Extract run/batch/snapshot resolution into `run_resolver.go` (same package).
- [ ] Extract outcome calculation logic into `engine.go` (same package, pure helpers where possible).
- [ ] Extract DB writes into `writer.go`.
- [ ] Keep the exported `Service` API stable.

### 3) Backend: Reduce handler boilerplate for analytics

Target: `backend/internal/transport/httpserver/handlers_analytics.go`

Outcome: Handlers read as orchestration only; DTO mapping is centralized.

Tasks:

- [ ] Extract analytics DTO mapping helpers to `backend/internal/transport/httpserver/dtos/mappers_analytics.go`.
- [ ] Extract common query param parsing helpers (e.g., `limit`) into a shared `params.go`.

### 4) Backend: Split importer by bundle type

Target: `backend/internal/bundles/importer/importer.go`

Outcome: Bundle import logic is discoverable; each bundle type is isolated.

Tasks:

- [ ] Move `importSchools` to `import_schools.go`.
- [ ] Move `importTournaments` (+ helpers) to `import_tournaments.go`.
- [ ] Move `importCalcuttas` (+ helpers) to `import_calcuttas.go`.
- [ ] Keep `ImportFromDir` and transaction boundaries in `importer.go`.

### 5) Frontend: Split large analytics pages into components

Targets:

- `frontend/src/pages/AnalyticsPage.tsx`
- `frontend/src/pages/TournamentAnalyticsPage.tsx`

Outcome: Pages become composition roots; logic and views are reusable.

Tasks:

- [ ] Extract major sections into `frontend/src/components/analytics/*`.
- [ ] Consolidate analytics-related API calls into `frontend/src/services/analyticsService.ts`.
- [ ] Keep routes and query keys stable.

## Execution Order

- [ ] Start with (1) `analytics.sql` split (low risk, high readability win).
- [ ] Then (2) `simulated_calcutta` decomposition.
- [ ] Then (3) handler mapping extraction.
- [ ] Then (4) importer split.
- [ ] Then (5) frontend cleanup.
