# Tech Debt Action Plan

Prioritized checklist of PR-sized work items surfaced by backend, frontend, data-science, database, and devops audits (2026-02-15). Each item is independently mergeable with no cross-PR dependencies within the same tier.

**Priority tiers:**
- **P0** — Bugs and runtime crashes (fix now)
- **P1** — Data integrity and UX-breaking issues (fix this week)
- **P2** — Dead code removal (reduce cognitive overhead)
- **P3** — Consistency (bring one-offs into established patterns)
- **P4** — Infrastructure hardening (Docker, CI, dev experience)
- **P5** — Schema cleanup and query optimization (opportunistic)

---

## P0 — Bugs and Runtime Crashes

### 1. Fix variable shadowing bug in calcutta evaluations batch ID
- [x] **Resolved** — `service.go` was restructured; batchID logic moved to `run_resolver.go` with no shadowing

### 2. Fix broken `minlp` strategy imports in Python
- [x] **Resolved** — Both `portfolio_strategies.py` and `recommended_entry_bids_db.py` were deleted in cleanup

### 3. Fix hardcoded year `2025` in predicted game outcomes
- [x] **Resolved** — `predicted_game_outcomes.py` was deleted in cleanup

### 4. Fix non-existent worker flags in local-prod compose
- [x] **Resolved** — `docker-compose.local-prod.yml` was deleted in cleanup

### 5. Fix setState during render in CalcuttaSettingsPage
- [x] **Resolved** — Both state initializations now use `useEffect` hooks with proper dependency arrays

---

## P1 — Data Integrity and UX-Breaking Issues

### 6. Fix dollar signs displayed on points values
- [ ] **`frontend/src/pages/CalcuttaTeamsPage.tsx:152`**
- `${team.totalInvestment.toFixed(2)}` renders a dollar sign on a points value, violating the domain unit convention (`_points` = in-game currency, never `$`)
- **Fix:** Replace `$` prefix with `pts` suffix or remove the currency symbol entirely

### 7. Fix N+1 waterfall in CalcuttaListPage
- [ ] **`frontend/src/pages/CalcuttaListPage.tsx:33-51`**
- Sequential `for` loop makes one HTTP request per calcutta to fetch entries
- **Fix:** Either add a batch endpoint that returns entries for all user calcuttas, or parallelize with `Promise.all`

### 8. Fix N+1 parallel requests in CalcuttaTeamsPage
- [ ] **`frontend/src/pages/CalcuttaTeamsPage.tsx:59-61`**
- `Promise.all(entries.map(e => getEntryTeams(e.id)))` fires one request per entry
- **Fix:** Add a batch endpoint (e.g., `GET /calcuttas/:id/entry-teams`) that returns all entry-team mappings in a single query

### 9. Fix analytics service swallowing 14 errors
- [ ] **`backend/internal/app/analytics/service.go`** (lines 234, 265, 288, 316, 344, 367, 413, 445, 480, 505, 615, 660, 703, 748)
- All 14 error paths call `log.Printf` and silently continue instead of returning the error
- **Fix:** Return errors to the caller. If partial results are acceptable, collect errors and return a multi-error at the end.

### 10. Fix race condition in concurrent evaluation result collection
- [ ] **`backend/internal/app/calcutta_evaluations/service.go:489-525`**
- Goroutine result collection drains all results before checking for errors, potentially processing empty/corrupt data
- **Fix:** Check errors before processing results, or use `errgroup` to fail fast on first error

### 11. Remove hardcoded pool size in recommended entry bids
- [ ] **`data-science/moneyball/models/recommended_entry_bids_db.py:65`**
- `total_pool = 47 * 100.0` assumes exactly 47 entries at 100 points each
- **Fix:** Accept pool size as a parameter derived from actual calcutta configuration

### 12. Remove hardcoded excluded entry name
- [ ] **`data-science/moneyball/pipeline/runner_db.py:87`**
- `EXCLUDED_ENTRY_NAME` defaults to `'Andrew Copp'`
- **Fix:** Make this a required pipeline parameter with no default, or remove the exclusion logic

---

## P2 — Dead Code Removal

### 13. Delete dead Go adapter and domain packages
- [ ] **`backend/internal/adapters/httpapi/`** — 3 files (`response.go`, `context_keys.go`, `router.go`), zero imports
- [ ] **`backend/internal/domain/auth/types.go`** — zero imports
- [ ] **`backend/internal/domain/tournament/types.go`** — zero imports
- [ ] **`backend/internal/domain/permissions/types.go`** — zero imports
- [ ] **`backend/internal/adapters/postgres/`** — empty directory
- [ ] **`backend/internal/core/`** — empty directory
- [ ] **`backend/internal/repository/`** — empty directory
- **Fix:** Delete all listed files and directories

### 14. Delete dead frontend components
- [ ] `frontend/src/components/RunViewerHeader.tsx` — zero imports
- [ ] `frontend/src/components/TabsNav.tsx` — zero imports
- [ ] `frontend/src/components/Lab/ModelPipelineCard.tsx` — zero imports
- [ ] `frontend/src/components/ui/StatusBadge.tsx` — zero imports
- [ ] `frontend/src/components/ui/ProgressBar.tsx` — zero imports
- [ ] `frontend/src/pages/Lab/EntriesTab.tsx` — zero imports (top-level; the detail sub-tab at `Lab/EntryDetail/` is still used)
- [ ] `frontend/src/pages/Lab/EvaluationsTab.tsx` — zero imports (top-level; the detail sub-tab is still used)
- [ ] `frontend/src/types/tournament.ts` — empty interface, `Tournament` lives in `types/calcutta.ts`
- [ ] `frontend/src/types/lodash.d.ts` — no lodash usage in codebase
- **Fix:** Delete all 9 files

### 15. Delete dead frontend service methods
- [ ] **`frontend/src/services/adminService.ts`** — remove `fetchTournamentTeams()`, `updateTournamentTeam()`, `recalculatePortfolios()`, `createTournamentTeam()` (duplicates of `tournamentService` methods; keep `adminService.getAllSchools()`)
- [ ] **`frontend/src/services/analyticsService.ts`** + **`frontend/src/types/analytics.ts`** — zombie service with zero external consumers (only references itself)
- **Fix:** Delete the dead functions and the zombie service/type pair

### 16. Delete dead Python files with RuntimeError guards
- [ ] **`data-science/moneyball/models/entry_percentile_analysis.py`** — 162 lines, raises `RuntimeError("deprecated")` before any logic
- [ ] **`data-science/moneyball/models/simulated_entry_outcomes.py`** — 377 lines, raises `RuntimeError("deprecated")` before any logic
- **Fix:** Delete both files. Remove any references in `__init__.py` or imports.

### 17. Delete dead Python pipeline stages
- [ ] **`data-science/moneyball/pipeline/runner.py:474-498`** — `_stage_simulated_entry_outcomes` and `_stage_investment_report` both raise deprecation errors
- **Fix:** Remove both dead stage methods and their registrations (lines 561-567)

### 18. Delete unused Airflow infrastructure
- [ ] **`data-science/airflow/`** — entire directory (`dags/`, `logs/`, `plugins/`)
- [ ] **`data-science/docker-compose.airflow.yml`** — 9 containers referencing nonexistent Docker images and modules
- **Fix:** Delete the directory and compose file. If Airflow is needed later, recreate from scratch.

### 19. Delete orphan Python database files
- [ ] **`data-science/moneyball/db/etl_parquet_to_postgres.py`** — completely empty file (0 lines)
- [ ] **`data-science/moneyball/db/init_db.py`** — orphan file with its own `get_db_connection()` using different env vars (`CALCUTTA_ANALYTICS_DB_*`), not imported anywhere
- **Fix:** Delete both files

### 20. Drop dead database views
- [x] `derived.calcuttas` — zero references in app code
- [x] `derived.teams` — zero references in app code
- [x] `derived.tournaments` — zero references in app code
- [x] `derived.v_algorithms` — zero references in app code
- [x] `derived.v_strategy_generation_run_bids` — zero references in app code
- **Fix:** ~~Create a migration that drops all 5 views~~ Done in migrations `20260216000007` and `20260216000008`

### 21. Drop orphan trigger functions
- [x] `derived.enqueue_run_job_for_entry_evaluation_request()` — no `CREATE TRIGGER` references it
- [x] `derived.enqueue_run_job_for_market_share_run()` — no `CREATE TRIGGER` references it
- [x] `derived.enqueue_run_job_for_strategy_generation_run()` — no `CREATE TRIGGER` references it
- **Fix:** ~~Create a migration that drops all 3 functions~~ Done in migration `20260216000007`

### 22. Drop orphan column on simulated_entries
- [x] **`derived.simulated_entries.source_candidate_id`** — references the dropped `candidates` table
- **Fix:** ~~Create a migration~~ Done in migration `20260216000008` (also updated `source_kind` check constraint)

---

## P3 — Consistency

### 23. Extract duplicate sigmoid/winProb into shared Go package
- [ ] **`backend/internal/adapters/db/predicted_returns_pgo.go:518`**
- [ ] **`backend/internal/app/simulation_game_outcomes/spec.go:53`**
- [ ] **`backend/internal/app/predicted_game_outcomes/service.go:740`**
- Identical implementations copy-pasted in 3 files
- **Fix:** Create a shared `mathutil` package (e.g., `backend/internal/mathutil/probability.go`) and replace all 3 call sites

### 24. Deduplicate nullUUIDParam helper
- [ ] **`backend/internal/app/workers/util.go:12`**
- [ ] **`backend/internal/transport/httpserver/sql_params.go:5`**
- Identical `nullUUIDParam()` and `nullUUIDParamPtr()` in both files
- **Fix:** Keep the one in `httpserver/sql_params.go` (closest to query layer) and import it from `workers/`

### 25. Consolidate Python get_db_connection implementations
- [ ] **`data-science/moneyball/db/connection.py:53`** — pooled, uses `DB_*` env vars (canonical)
- [ ] **`data-science/moneyball/db/readers.py:18`** — raw per-call, uses `DB_*` env vars
- [ ] **`data-science/moneyball/db/init_db.py:9`** — raw, uses `CALCUTTA_ANALYTICS_DB_*` env vars (deleted in item 19)
- **Fix:** After deleting `init_db.py` (item 19), update `readers.py` to use the canonical pooled connection from `connection.py`

### 26. Deduplicate frontend School type
- [ ] **`frontend/src/types/calcutta.ts:30-35`** — defines `School` interface
- [ ] **`frontend/src/types/school.ts:1-6`** — identical `School` interface
- **Fix:** Keep one (prefer `school.ts` as the canonical source), re-export from `calcutta.ts` if needed, update all imports

### 27. Migrate admin pages to React Query
- [ ] **`frontend/src/pages/AdminApiKeysPage.tsx`** — uses `useEffect` + `useState` for data fetching
- [ ] **`frontend/src/pages/AdminUsersPage.tsx`** — uses `useEffect` + `useState` for data fetching
- **Fix:** Replace manual fetch patterns with `useQuery` / `useMutation` hooks from React Query, matching the pattern used in other pages

### 28. Create hallOfFameService and wire up HallOfFamePage
- [ ] **`frontend/src/pages/HallOfFamePage.tsx:132-150`**
- Direct `apiClient.get()` calls bypass the service layer convention
- **Fix:** Create `frontend/src/services/hallOfFameService.ts` with typed methods, update the page to call through the service

### 29. Align Lab frontend pages with established patterns
- [ ] Lab pages use raw `<div>` instead of `PageContainer`
- [ ] Lab pages use inline query keys instead of centralized `queryKeys.ts`
- **Fix:** Wrap Lab pages in `PageContainer`, move query keys to `queryKeys.ts`

### 30. Replace Python print-with-emoji logging with proper logging
- [ ] **`data-science/moneyball/pipeline/runner.py`** — uses `print()` with emoji characters
- [ ] **`data-science/moneyball/pipeline/runner_db.py`** — uses `print()` with emoji characters
- [ ] **`data-science/moneyball/pipeline/db_writer.py`** — uses `print()`
- [ ] **`data-science/moneyball/models/recommended_entry_bids_db.py`** — uses `print()`
- **Fix:** Replace all `print()` calls with `logging.getLogger(__name__)` and appropriate log levels

### 31. Extract generateInviteToken business logic from handler
- [ ] **`backend/internal/transport/httpserver/handlers_admin_users.go:230-321`**
- Handler contains database transactions, token generation, and retry logic
- **Fix:** Move token generation logic into an auth or invite service, handler should only parse request and call service

### 32. Standardize Go logging on slog
- [ ] 18 files still use `log.Printf` instead of `slog`
- Key offenders: all 6 worker files, `analytics/service.go`, `calcutta_evaluations/service.go`, handler files
- **Fix:** Replace `log.Printf` with structured `slog.Info`/`slog.Error` calls. Remove `log` imports.

### 33. Encapsulate global mutable database pool
- [ ] **`backend/internal/db/db.go:12`**
- Package-level `var pool *pgxpool.Pool` is globally mutable — any package can import and use it, bypassing DI and making testing difficult
- **Fix:** Remove the global variable. Pass `*pgxpool.Pool` through the bootstrap layer into services that need it (most already receive it via constructor).

### 34. Split oversized Go files exceeding 400 LOC guideline
- [ ] 17+ files exceed the 400 LOC guideline, top offenders:
  - `internal/adapters/db/lab_repository.go` (910 LOC)
  - `internal/app/workers/lab_pipeline_worker.go` (906 LOC)
  - `internal/app/recommended_entry_bids/service.go` (850 LOC)
  - `internal/app/calcutta_evaluations/run_resolver.go` (841 LOC)
  - `internal/adapters/db/lab_pipeline_repository.go` (835 LOC)
  - `internal/app/analytics/service.go` (817 LOC)
  - `internal/transport/httpserver/calcuttas/handlers.go` (804 LOC)
  - `internal/app/predicted_game_outcomes/service.go` (791 LOC)
- **Fix:** Split each file by responsibility into 2-3 smaller files (e.g., separate read vs write methods, or split by sub-domain). Target <400 LOC per file.

### 35. Migrate services off direct pgxpool.Pool to port/adapter pattern
- [ ] 14 files in `backend/internal/app/` import `pgxpool` directly instead of depending on port interfaces
- Key offenders: all 6 workers, `recommended_entry_bids/service.go`, `predicted_game_outcomes/service.go`, `simulate_tournaments/service.go`, `calcutta_evaluations/service.go`, `simulation_artifacts/service.go`
- **Fix:** Define port interfaces for each service's data needs, create adapter implementations, inject via constructor. `bootstrap/app.go` (wiring layer) is the only acceptable place to reference `pgxpool` directly.

### 36. Stop Python gold writers from writing to deprecated derived tables
- [ ] **`data-science/moneyball/db/writers/gold_writers.py:39, 97`** — writes to `derived.strategy_generation_runs`
- [ ] **`data-science/moneyball/pipeline/runner_db.py:401`** — references same table
- **Fix:** Update writers to target `lab.*` tables. Remove references to `derived.strategy_generation_runs`.

---

## P4 — Infrastructure Hardening

### 37. Add health checks and restart policies to Docker services
- [ ] **`docker-compose.yml`** — backend, worker, and frontend services lack health checks
- [ ] **`docker-compose.local-prod.yml`** — same
- **Fix:** Add `healthcheck` blocks (HTTP check for backend/frontend, process check for workers) and `restart: unless-stopped` to all services

### 38. Remove meaningless NODE_ENV from Go services
- [ ] **`docker-compose.yml:15, 28, 62, 86`**
- [ ] **`docker-compose.local-prod.yml:21, 33`**
- `NODE_ENV` is passed to Go backend and worker services where it has no effect
- **Fix:** Remove `NODE_ENV` from Go service environment blocks. Keep it only on the frontend service.

### 39. Bake pip install into Docker image instead of running at startup
- [ ] **`docker-compose.yml:52, 73`**
- `pip install --break-system-packages` runs on every container start, adding latency and fragility
- **Fix:** Move the pip install into the Dockerfile `RUN` layer so dependencies are baked into the image

### 40. Use create-admin CLI tool in dry-run script
- [ ] **`scripts/dry-run-local.sh:96-109`**
- Raw SQL `INSERT` to create admin user; comment says `TODO: Build create-admin CLI tool` but the tool already exists at `backend/cmd/tools/create-admin/main.go`
- **Fix:** Replace the raw SQL block with a call to the `create-admin` tool

### 41. Add resource limits to Docker containers
- [ ] **`docker-compose.yml`** — no `mem_limit` or `cpus` on any service
- [ ] **`docker-compose.local-prod.yml`** — same
- **Fix:** Add reasonable resource limits (`mem_limit`, `cpus`) to prevent any single service from consuming all host resources

---

## P5 — Schema Cleanup and Query Optimization

### 42. Add missing FK constraints and indexes on derived tables
- [ ] `derived.game_outcome_runs.algorithm_id` — no FK to `derived.algorithms`
- [ ] `derived.simulation_cohorts` algorithm columns — no FK constraints, no indexes
- **Fix:** Create a migration adding `FOREIGN KEY` constraints and `CREATE INDEX` on the algorithm columns

### 43. Rename legacy medallion-era index and constraint names
- [ ] PK/index names still use `bronze_*`, `gold_*`, `silver_*`, `analytics_*` prefixes on tables that have been renamed
- **Fix:** Create a migration that renames indexes and constraints to match current table names (use `ALTER INDEX ... RENAME TO ...`)

### 44. Add missing deleted_at columns to lab tables
- [ ] `lab.evaluation_entry_results` — has `updated_at` but no `deleted_at`
- [ ] `lab.pipeline_runs` — missing `deleted_at`
- [ ] `lab.pipeline_calcutta_runs` — missing `deleted_at`
- **Fix:** Create a migration adding `deleted_at TIMESTAMPTZ` to all three tables

### 45. Extract entry_bids CTE into a reusable SQL view
- [ ] Duplicated in 5+ files: `lab_repository.go`, `sqlc/calcutta_entries.sql.go`, `sqlc/queries/calcutta_entries.sql` (3x), `sqlc/queries/analytics.sql` (4x)
- **Fix:** Create a `core.v_entry_bids` view encapsulating the CTE logic, then replace all inline CTEs with `SELECT * FROM core.v_entry_bids WHERE ...`

### 46. Fix N+1 INSERT loop in calcutta evaluations writer
- [ ] **`backend/internal/app/calcutta_evaluations/writer.go`** — `writePerformanceMetrics()` uses individual INSERTs in a loop
- The same file uses `pgx.Batch{}` for `writeSimulationOutcomes()` 20 lines above
- **Fix:** Rewrite `writePerformanceMetrics()` to use `pgx.Batch{}` matching the existing pattern

### 47. Optimize correlated subqueries in simulation entry query
- [ ] `GetEntrySimulationsByRunKeyAndEntryName` — 5 correlated subqueries per row
- **Fix:** Rewrite as JOINs or window functions to eliminate per-row subquery execution

### 48. Migrate os.Getenv calls out of Go service layer
- [ ] 9 files in `backend/internal/app/` call `os.Getenv()` directly instead of receiving config via dependency injection
- Key files: all 6 workers, `calcutta_evaluations/service.go`, `lab/service.go`, `lab/python_runner.go`
- **Fix:** Add config fields to service structs, inject values from `cmd/` layer at startup. Remove `os` imports from service files.

### 49. Migrate raw SQL calls to sqlc for type safety
- [ ] 33 Go files use `pool.Query`, `pool.QueryRow`, or `pool.Exec` directly, bypassing sqlc type safety
- These 100+ raw SQL statements are invisible to `make sqlc-generate` and will silently break on column renames
- **Fix:** Incrementally migrate the highest-risk raw queries (writes, joins) into `sqlc/queries/*.sql` files. Prioritize files that also violate the port/adapter pattern (item 35).

---

## Summary

| Tier | Items | Theme |
|------|-------|-------|
| P0   | 5     | Crashes and broken code paths |
| P1   | 7     | Silent data corruption and UX bugs |
| P2   | 10    | Dead code across all 5 layers |
| P3   | 14    | Pattern violations and duplication |
| P4   | 5     | Docker and deployment hygiene |
| P5   | 8     | Schema and query improvements |
| **Total** | **49** | |
