# transport/httpserver cleanup

## Goals
- [x] Define what belongs in `internal/transport/httpserver` vs `internal/adapters` vs `internal/app`
- [ ] Make HTTP concerns easy to find (routing, handlers, middleware, request/response DTOs, error mapping)
- [ ] Remove DB/repository and worker/background-job code from the transport layer (or clearly justify exceptions)
- [ ] Reduce filename-prefix sprawl (prefer package boundaries + smaller, consistent filenames)

## Progress so far (completed)
- [x] Decide worker runtime model: long-running workers deployed as a separate ECS service; UI actions enqueue jobs
- [x] Remove legacy wrapper packages by migrating all Go imports to `backend/internal/app/*`
- [x] Update `backend/internal/app/README.md` to reflect `internal/app/<feature>` as the canonical import path
- [x] Carve out ML analytics routes into `transport/httpserver/mlanalytics` feature package and compose via `routes.go`
- [x] Evict `transport/httpserver/*_repository.go` shims by wiring DB adapters directly and keeping handlers on `app` services/ports
- [x] Implement Lab Candidates vertical slice behind `internal/app/lab_candidates` + `internal/adapters/db` and expose via `transport/httpserver/labcandidates`
- [x] Delete empty placeholder files left behind by migrations (package-only `handlers_*` / `*_repository.go` stubs)

## Plan of attack (milestones)
- [ ] Milestone 0: hygiene + guardrails
  - [x] Ensure the doc’s dependency rules are the reference for the refactor (see: [Target architecture (DBClient / HTTPClient / models / adapters)](#target-architecture-dbclient--httpclient--models--adapters))
  - [x] Confirm/remove any remaining references to the deleted legacy wrapper layer (Go, docs, scripts)
  - [x] Add a CI guardrail to prevent the legacy wrapper layer from reappearing
  - [x] Decide whether to introduce temporary allowlists for import-boundary checks during migration (decision: defer until Milestone 7; only legacy-wrapper guardrail is enforced for now)
- [x] Milestone 1: map the current state
  - [x] Complete the inventory and classification for `transport/httpserver` (see: [Inventory + classification](#inventory--classification))
  - [x] Fill in the move map until every file has a target package (see: [Move map](#move-map-fill-in-during-inventory))
- [x] Milestone 2: low-risk package reshuffle (no behavior change)
  - [x] Move middleware + error helpers into subpackages (see: [Target structure (package layout)](#target-structure-package-layout))
    - [x] Move CORS + max-body-bytes middleware into `httpserver/middleware`
    - [x] Move logging middleware into `httpserver/middleware` (keep root wrapper)
    - [x] Move rate limiting middleware into `httpserver/middleware` (keep root wrapper)
    - [x] Move HTTP error helpers into `httpserver/httperr` (keep root wrapper)
  - [x] Introduce first feature subpackage (`modelcatalogs`) and migrate it without behavior changes
  - [x] Extract shared response helper (`httpserver/response.WriteJSON`) to avoid feature-package import cycles
  - [x] Extract request context helpers (`httpserver/requestctx`) and update call sites/tests
  - [x] Introduce additional feature subpackages under `httpserver/` and migrate one feature end-to-end without changing logic (see: [Proposed `httpserver/` layout (feature-first)](#proposed-httpserver-layout-feature-first))
- [x] Milestone 3: handler cleanup (feature-first)
  - [x] Carve out 1 feature (`mlanalytics`) into `transport/httpserver/<feature>` with `routes.go`, `handlers.go`, `types.go`, `params.go`
  - [x] Update top-level routing to compose feature route registrars (see: [Routing + wiring](#routing--wiring))
 - [ ] Milestone 4: repositories eviction
  - [x] Remove repository shims and ensure handlers call `app` services/ports (see: [Repositories: eviction plan](#repositories-eviction-plan))
  - [ ] Ensure DB access lives behind `internal/app` services + `internal/adapters/db` repositories (no handler SQL)
    - [x] Simulated calcuttas + entries handlers: SQL evicted into `internal/app/suite_scenarios` + `internal/adapters/db/suite_scenarios_repository.go`
    - [x] Entry runs + strategy generation runs handlers: SQL evicted into `internal/app/strategy_runs` + `internal/adapters/db/strategy_runs_repository.go`
    - [x] Synthetic calcuttas + entries handlers: SQL evicted into `internal/app/synthetic_scenarios` + `internal/adapters/db/synthetic_scenarios_repository.go`
    - [x] Suite calcutta evaluations handlers + domain helper: SQL evicted into `internal/app/suite_evaluations` + `internal/adapters/db/suite_evaluations_repository.go`
    - [ ] Remaining handler SQL in `transport/httpserver` still needs eviction (examples: `handlers_run_progress.go`, `handlers_entry_evaluation_requests.go`, `handlers_admin_users.go`, `handlers_synthetic_cohorts.go`, `handlers_suite_executions.go`, `handlers_auth_invite.go`)
- [x] Milestone 5: workers eviction
  - [x] Move `*_worker.go` out of `transport/httpserver` into a worker layer and make dependencies explicit (see: [Workers: eviction plan](#workers-eviction-plan))
  - [x] Align the worker entrypoint with the ECS service model (see: [Worker runtime model (app-triggered, not CLI)](#worker-runtime-model-app-triggered-not-cli))
- [ ] Milestone 6: naming + cleanup
  - [ ] Remove `handlers_*` filename prefixes once feature packages exist (see: [Naming + file cleanup](#naming--file-cleanup))
  - [x] Tournaments feature package created and migrated
  - [x] Calcuttas feature package created and migrated
  - [x] Model catalogs feature package created and migrated
  - [ ] Align backend resource naming (remove “suite” where it no longer matches API resources)
- [ ] Milestone 7: enforce the boundaries
  - [ ] Add CI checks for forbidden imports and anti-regression checks for new `*_repository.go` / `*_worker.go` under `transport/httpserver` (see: [Best-practice guardrails](#best-practice-guardrails-avoid-the-sideways-look))

## Next steps (near-term)
- [x] Delete legacy wrapper directory
- [x] Run backend tests to confirm: `go test ./...` (from `backend/`)
- [x] Re-run grep to confirm no remaining references to the deleted legacy wrapper layer (Go, docs, scripts)

## Target architecture (DBClient / HTTPClient / models / adapters)
- [ ] Architecture note (keep this true as the refactor progresses):
  - [x] `internal/app/...` is the canonical home for feature logic (use-cases/services) and the port interfaces they depend on
  - [x] `internal/adapters/...` contains concrete IO implementations (DB repositories, external clients)
  - [x] `internal/transport/httpserver/...` is an inbound HTTP adapter: routing, middleware, request/response DTOs, and error mapping
  - [ ] Dependency direction:
    - [x] `transport/httpserver` -> `app` (calls use-cases)
    - [x] `adapters/*` -> `app` (implements ports)
    - [x] `app` must not import `transport` or concrete `adapters`
  - [ ] DTO/type boundaries:
    - [x] HTTP DTOs stay in `transport/httpserver`
    - [x] DB row/query types stay in `adapters/db`
    - [x] Mapping between layers happens at the edges
- [ ] Agree on core conceptual buckets:
  - [ ] **Models (core/app)**: domain types + use-cases/services
  - [ ] **DBClient**: the database handle/transaction abstraction (sql.DB/sqlx/pgx/etc.)
  - [ ] **HTTPClient**: outbound HTTP client(s) for external APIs
  - [ ] **Model↔DB adapters**: repository implementations mapping DB <-> model
  - [ ] **Model↔HTTP adapters**:
    - [ ] inbound: HTTP server handlers mapping HTTP <-> model/service calls
    - [ ] outbound: HTTP clients mapping model/service calls <-> external HTTP
- [ ] Decide canonical package roots (example; adjust to match existing conventions):
  - [x] `backend/internal/app/...` for use-cases/services + port interfaces
  - [x] `backend/internal/adapters/db/...` for DB repositories (Model↔DB adapters)
  - [ ] `backend/internal/adapters/httpclient/...` for outbound HTTP client adapters
  - [x] `backend/internal/transport/httpserver/...` for inbound HTTP server adapter
- [ ] Write down dependency rules and enforce them during the refactor:
  - [ ] `transport/httpserver` depends on `app` (interfaces/services), not on `adapters/db`
  - [ ] `adapters/*` depends on `app` (implements ports), not on `transport/httpserver`
  - [ ] `app` depends on neither `transport` nor concrete `adapters`
  - [ ] HTTP DTOs stay in `transport/httpserver` (do not leak into `app`)
  - [ ] DB row types stay in `adapters/db` (do not leak into `app`)
- [ ] Decide where the “ports” (interfaces) live:
  - [x] `app/<feature>` defines repository/service interfaces it needs
  - [x] `adapters/db/<feature>` implements those interfaces
  - [x] `transport/httpserver/<feature>` depends on app services (or interfaces)

## Inventory + classification
- [x] List all files in `backend/internal/transport/httpserver/` and classify each as:
  - [x] HTTP transport (handlers, routing, request parsing, response encoding)
  - [x] Middleware
  - [x] HTTP error mapping/utilities
  - [x] DB/repository/data access
  - [x] Worker/background job/orchestration
  - [x] Shared types/utilities (decide destination)
- [x] Identify `*_repository.go` files and answer for each:
  - [x] Is it truly HTTP-specific (e.g., translating HTTP -> app/service call), or is it a data-access adapter?
  - [x] What does it depend on (sql/db, external clients, app services)?
  - [x] Who calls it (handlers only, or broader usage)?
- [x] Identify `*_worker.go` files and answer for each:
  - [x] Is it triggered by HTTP requests or is it a standalone/background process?
  - [x] What runtime owns it today (same API binary, separate cmd, cron, queue consumer)?
  - [x] What state/artifacts does it create/update?

## Inventory findings (captured)
- [ ] Confirm/record these observed facts:
  - [x] `transport/httpserver/*_repository.go` are currently thin re-exports of `internal/adapters/db/*_repository.go` (aliases + `New*Repository(pool)` helpers)
  - [x] `internal/adapters/db/` already contains `api_keys_repository.go`, `auth_repository.go`, `authz_repository.go`, `user_repository.go` (i.e., canonical DB adapter implementations already exist)
  - [x] `transport/httpserver/*_worker.go` implement long-running poller workers directly as `Server` methods and directly manage DB transactions via `s.pool`
  - [x] There is an existing `internal/adapters/httpapi/` package that likely wants to absorb some shared HTTP server primitives (router/response helpers)

## Milestone 1 inventory snapshot (current `transport/httpserver`)
- [ ] Classify these buckets and keep them up to date as files move:
  - [ ] **HTTP server wiring (keep in root `httpserver` package, but may be slimmed)**
    - [ ] `main.go` (contains `Run()`)
    - [ ] `server.go` (constructs server, holds dependencies)
    - [ ] `routes.go` (top-level route composition)
  - [ ] **Transport infra (move into subpackages under `httpserver/` in Milestone 2)**
    - [x] `middleware.go`, `middleware_auth.go`
    - [x] `http_helpers.go`
    - [ ] `params.go`, `sql_params.go`
    - [ ] `metrics.go`
    - [ ] `cognito_jwt_verifier.go`
  - [ ] **DTOs (currently centralized under `httpserver/dtos/`)**
    - [ ] `dtos/*.go` (14 files)
  - [x] **Repository shims (deleted in Milestone 4; redundant wrappers over `adapters/db`)**
    - [x] `api_keys_repository.go`, `auth_repository.go`, `authz_repository.go`, `user_repository.go`
  - [x] **Worker shims (deleted in Milestone 5; workers already exist under `internal/app/workers`)**
    - [x] `bundle_import_worker.go`, `entry_evaluation_worker.go`, `market_share_worker.go`, `game_outcome_worker.go`, `strategy_generation_worker.go`, `calcutta_evaluation_worker.go`
  - [ ] **Feature handlers (largest set; to be moved into `httpserver/<feature>/...`)**
    - [ ] Admin: `handlers_admin_*`
    - [ ] Auth: `handlers_auth*`
    - [ ] Analytics read endpoints: `handlers_analytics.go`, `handlers_hall_of_fame*` (if present)
    - [x] ML analytics endpoints: `httpserver/mlanalytics/*` (feature package)
    - [x] Tournaments: `httpserver/tournaments/*` (feature package)
    - [ ] Bracket: `handlers_bracket.go`
    - [x] Calcuttas/entries: `httpserver/calcuttas/*` (feature package)
    - [ ] Portfolios: remaining handlers still in root `httpserver` (partial work done; see `httpserver/calcuttas`)
    - [ ] Runs/progress/artifacts: DB logic evicted for `handlers_entry_runs.go` + `handlers_strategy_generation_runs.go`, but feature packages not created and `handlers_run_progress.go` still has SQL
    - [ ] Entry eval requests: `handlers_entry_evaluation_requests.go`
    - [ ] Strategy generation runs: DB logic evicted, feature package not created
    - [ ] Lab + synthetic data: `handlers_lab_entries*`, `handlers_synthetic_*`, `httpserver/labcandidates/*` (feature package)
    - [ ] Cohorts/simulations: DB logic evicted, feature package not created (still named `handlers_suite_calcutta_evaluations*`)
    - [x] Model catalogs: `httpserver/modelcatalogs/*` (feature package)
  - [ ] **Tests to relocate alongside moved packages**
    - [ ] `http_helpers_test.go`
    - [ ] `handlers_*_test.go` (various)
    - [ ] `dtos/mappers_analytics_test.go`

## Proposed `httpserver/` layout (feature-first)
- [ ] Create feature subpackages under `httpserver/` so names don’t encode architecture:
  - [x] `httpserver/mlanalytics/`
  - [ ] `httpserver/calcuttaevaluations/`
  - [ ] (add others as discovered in inventory)
- [ ] Introduce support subpackages:
  - [x] `httpserver/middleware/` (auth, logging, request-id, etc.)
  - [x] `httpserver/httperr/` (error mapping/encoding)
  - [ ] `httpserver/params/` if shared param parsing helpers are genuinely cross-feature
- [ ] Within each feature subpackage, standardize filenames:
  - [ ] `routes.go` (route registration)
  - [ ] `handlers.go` (HTTP handler methods)
  - [ ] `types.go` (request/response DTOs)
  - [ ] `params.go` (query/path parsing + validation)

## Target structure (package layout)
- [ ] Decide on feature subpackages under `httpserver/` (e.g., `calcuttaevaluations/`, `mlanalytics/`, etc.)
- [x] Create `httpserver/middleware/` package (move `middleware*.go` there)
- [x] Create `httpserver/httperr/` (or `apierr/`) package for error encoding/mapping
- [ ] Establish a consistent pattern per feature package:
  - [ ] `routes.go` with `RegisterRoutes(...)`
  - [ ] `handlers.go` for handler methods
  - [ ] `types.go` for HTTP DTOs
  - [ ] `params.go` for query/path parsing + validation

## Repositories: eviction plan
- [ ] Decide the canonical location for DB/external IO adapters currently in `httpserver/`:
  - [ ] `internal/adapters/repository/...` (DB)
  - [ ] `internal/adapters/httpclient/...` (external HTTP)
  - [ ] `internal/adapters/query/...` (read-model/query services)
- [ ] For each `*_repository.go`:
  - [ ] Move the concrete implementation out of `httpserver/`
  - [ ] Define/confirm the interface it should satisfy (ideally in `internal/app/...`)
  - [ ] Update HTTP handlers to depend on interfaces (constructor injection)
  - [ ] Update wiring/initialization (where the server is assembled)

## Model↔DB adapters: concrete refactor steps
- [ ] For each DB-backed feature currently reaching into DB from `httpserver/`:
  - [ ] Create/confirm an `app/<feature>` service (use-case) boundary that the handler calls
  - [ ] Put the required repository interface(s) in `app/<feature>` (ports)
  - [ ] Move SQL/query code into `adapters/db/<feature>` implementing those ports
  - [ ] Ensure transaction/DBClient management is owned by the DB adapter layer (or a shared DB adapter pkg)
  - [ ] Ensure handler tests can mock the app service without DB

## Model↔HTTP adapters: inbound vs outbound
- [ ] Inbound HTTP server adapter (`transport/httpserver`):
  - [ ] Keep HTTP DTOs, routing, middleware, and error mapping here
  - [ ] Ensure handlers call `app` services, not repositories
- [ ] Outbound HTTP client adapters (`adapters/httpclient/...`):
  - [ ] For each external API dependency, define an interface in `app/<feature>`
  - [ ] Implement it in `adapters/httpclient/<external>`
  - [ ] Keep request/response mapping for the external API in the adapter

## Workers: eviction plan
- [ ] Decide where worker/orchestration code should live:
  - [ ] `internal/app/workers/...` (app-level jobs)
  - [ ] `internal/adapters/worker/...` (queue/runner-specific implementations)
  - [ ] `cmd/<job>/...` if it should be a separate binary
- [ ] For each `*_worker.go`:
  - [ ] Move the worker out of `httpserver/`
  - [ ] Ensure it has a clear entrypoint (HTTP-triggered vs background)
  - [ ] Ensure dependencies are injected (no hidden coupling to httpserver globals)

## Worker-specific decisions
- [ ] Decide “trigger vs execute” split:
  - [ ] HTTP handler may enqueue/start a job (trigger)
  - [ ] Worker package executes the job (implementation)
- [ ] Decide where job definitions live:
  - [ ] `app/<feature>` defines the job interface/inputs/outputs
  - [ ] `adapters/worker/<runner>` implements execution (if runner-specific)
  - [ ] `cmd/<job>` provides a standalone entrypoint when needed

## Worker runtime model (app-triggered, not CLI)
- [ ] Confirm operating model: workers are not primarily started manually from the CLI; UI actions enqueue jobs that workers pick up
- [ ] Decide default execution approach:
  - [ ] Option A: long-running worker service(s) polling/consuming jobs (chosen: separate ECS service)
  - [ ] Option B: on-demand workers spun up per job batch (ECS task per run, Lambda per job, etc.)
- [ ] Decide where the queue lives:
  - [ ] DB-backed queue via `derived.run_jobs` (current pattern in worker code)
  - [ ] External queue (SQS, Redis, etc.) with DB as source-of-truth
- [ ] Define the “trigger” contract for HTTP:
  - [ ] Handler creates a `run_jobs` record (or equivalent) and returns a run id
  - [ ] Handler does not perform the heavy work inline
- [ ] Define worker scaling + deployment expectation:
  - [ ] Start with a long-running worker service colocated with API deployment (same cluster) but separate process
  - [ ] Allow later migration to separate ECS service / task-per-run without changing the app interface
- [ ] Ensure workers are safe for multiple replicas:
  - [ ] Claim/lease semantics are correct (at-most-once or at-least-once, with idempotency)
  - [ ] Jobs are idempotent or protected by unique constraints / status transitions
  - [ ] Workers emit progress/status into DB so UI can show progress

## Consolidate `internal/app` vs legacy wrapper layer
- [x] Confirm current state: the legacy wrapper layer duplicated feature directories already present under `internal/app/*`
- [ ] Decide canonical convention going forward:
  - [ ] Use `internal/app/<feature>` as the only home for feature use-cases/services (recommended)
  - [x] Remove the legacy wrapper layer entirely (chosen: remove immediately)
- [ ] Define “what goes where” rules:
  - [ ] `internal/app/<feature>`: use-cases/services, ports (interfaces), orchestration logic
  - [ ] `internal/adapters/db/<feature>`: repository implementations (DB adapter)
  - [ ] `internal/adapters/httpclient/<external>`: outbound HTTP client adapters
  - [ ] `internal/transport/httpserver/<feature>`: inbound HTTP handlers + DTOs + routing
- [ ] Migration tasks:
  - [x] Update imports throughout the repo to use `internal/app/<feature>`
  - [x] Update `internal/app/README.md` describing the canonical structure
  - [x] Delete the legacy wrapper directory

## Naming + file cleanup
- [ ] Remove `handlers_*` filename prefixes after feature subpackages exist
- [ ] Replace “split-by-name only” files like `*_domain.go` vs `*_types.go` with clearer boundaries:
  - [ ] If HTTP-specific: keep in feature package as DTOs/helpers
  - [ ] If domain/app: move to `internal/app/...` (or appropriate package)
- [ ] Standardize `params` parsing patterns (where validation lives, how errors are returned)

## Acceptance criteria (definition of done)
- [ ] `internal/transport/httpserver` contains only HTTP server concerns (no DB queries, no repository impls)
- [x] No `*_repository.go` remains under `transport/httpserver` (unless explicitly justified in this doc)
- [x] No `*_worker.go` remains under `transport/httpserver` (unless explicitly justified in this doc)
- [ ] Each feature has a clear `app/<feature>` boundary used by HTTP handlers
- [ ] Build/test passes after each phase; changes are batched to keep diffs reviewable

## Best-practice guardrails (avoid the “sideways look”)
- [ ] Keep layering pragmatic (avoid ceremony):
  - [ ] Do not introduce new interfaces unless they buy something concrete (testing seam, multiple implementations, or stable boundary)
  - [ ] Do not create “helper” packages by default; prefer feature-local helpers unless truly shared
  - [ ] Keep handlers thin, but don’t force a "service" layer for trivial pass-through endpoints
- [ ] Clarify package semantics to reduce confusion:
  - [x] Decide what `internal/app` means vs the legacy wrapper layer (pick one of: consolidate, or define the rule for each)
  - [ ] Add a short architecture note in this doc describing the chosen rule (2-10 bullets)
- [ ] Enforce DTO/type boundaries:
  - [ ] HTTP request/response DTOs live only in `transport/httpserver/...` packages
  - [ ] DB row/query structs live only in `adapters/db/...` (e.g., sqlc types stay there)
  - [ ] `app` types are transport-agnostic and persistence-agnostic
  - [ ] Mapping between these layers happens at the edges (handlers/adapters), not in core
- [ ] Reduce “shim” indirection:
  - [x] Remove `httpserver/*_repository.go` alias shims (they hide the real dependency direction)
  - [ ] Prefer explicit dependency injection from server wiring into handlers/services
- [ ] Make worker ownership explicit:
  - [ ] Workers should not be methods on `httpserver.Server` (avoid coupling to HTTP runtime)
  - [ ] Define a clear entrypoint per worker:
    - [ ] Either a separate `cmd/<worker>` binary
    - [ ] Or a worker runner started by the API binary (but still in a non-transport package)
  - [ ] Ensure workers have explicit deps (`pool`, repositories, services) passed in
- [ ] “One obvious place” rules:
  - [ ] One feature = one `httpserver/<feature>` package with `routes.go`, `handlers.go`, `types.go`, `params.go`
  - [ ] One feature = one `app/<feature>` package defining its ports + use-cases
  - [ ] One feature = one `adapters/db/<feature>` package implementing DB ports (avoid cross-feature repos)
- [ ] Lightweight enforcement (so future changes don’t regress):
  - [ ] Add a CI check that prevents `transport/httpserver` from importing `internal/adapters/db` (except temporary allowlist during migration)
  - [ ] Add a CI check that prevents `internal/app` from importing `transport/httpserver` or concrete `adapters/*`
  - [x] Add a grep-based CI check to fail on new `*_repository.go` files under `transport/httpserver`
  - [x] Add a grep-based CI check to fail on new `*_worker.go` files under `transport/httpserver`

## Routing + wiring
- [ ] Establish a single top-level `router.go` (or equivalent) that composes feature `RegisterRoutes` calls
- [ ] Ensure each feature package exports only what the router needs (minimize cross-feature coupling)
- [ ] Confirm middleware ordering and ownership (global vs per-route)

## Tests + safety
- [ ] Add/strengthen handler tests around request parsing + error mapping for moved endpoints
- [ ] Ensure any moved repository/worker code has unit tests (or at least integration coverage where appropriate)
- [ ] Run API build/test locally after each move step (small batches)

## Rollout plan
- [x] Phase 1: move middleware + error helpers into subpackages (no behavior changes)
- [x] Phase 2: carve out 1 feature (e.g., `mlanalytics`) into its own package and update router
- [x] Phase 3: evict `*_repository.go` implementations out of `httpserver/`
- [x] Phase 4: evict `*_worker.go` out of `httpserver/`
- [ ] Phase 5: rinse/repeat for remaining features, then remove legacy handler files

## Move map (fill in during inventory)
- [ ] `backend/internal/transport/httpserver/<file>.go` -> `<target package/path>`
- [ ] `backend/internal/transport/httpserver/<file>.go` -> `<target package/path>`

## Move map (grouped; every current file should fit one bucket)
- [ ] **Keep (root wiring)**
  - [ ] `backend/internal/transport/httpserver/main.go` -> keep (may later move to `cmd/api` wiring)
  - [ ] `backend/internal/transport/httpserver/server.go` -> keep (but remove DB/worker shims over time)
  - [ ] `backend/internal/transport/httpserver/routes.go` -> keep until feature routers exist; later becomes a thin composition root
- [ ] **Move (shared transport infra)**
  - [x] `backend/internal/transport/httpserver/middleware.go` -> `backend/internal/transport/httpserver/middleware/middleware.go`
  - [x] `backend/internal/transport/httpserver/middleware_auth.go` -> `backend/internal/transport/httpserver/middleware/auth.go`
  - [x] `backend/internal/transport/httpserver/http_helpers.go` -> `backend/internal/transport/httpserver/httperr/encode.go` (or similar)
  - [ ] `backend/internal/transport/httpserver/params.go` -> `backend/internal/transport/httpserver/params/params.go` (or feature-local)
  - [ ] `backend/internal/transport/httpserver/sql_params.go` -> `backend/internal/transport/httpserver/sqlparams/sqlparams.go` (or delete if unused after repo moves)
  - [ ] `backend/internal/transport/httpserver/metrics.go` -> `backend/internal/transport/httpserver/metrics/metrics.go`
  - [ ] `backend/internal/transport/httpserver/cognito_jwt_verifier.go` -> `backend/internal/transport/httpserver/authn/cognito_jwt_verifier.go`
- [ ] **Keep (DTOs for now; split later by feature)**
  - [ ] `backend/internal/transport/httpserver/dtos/*.go` -> keep in `backend/internal/transport/httpserver/dtos/` until feature packages stabilize
- [ ] **Delete (repository shims; replace with `app` ports/services)**
  - [x] `backend/internal/transport/httpserver/api_keys_repository.go` -> delete
  - [x] `backend/internal/transport/httpserver/auth_repository.go` -> delete
  - [x] `backend/internal/transport/httpserver/authz_repository.go` -> delete
  - [x] `backend/internal/transport/httpserver/user_repository.go` -> delete
- [ ] **Delete (worker shims; keep `cmd/workers` + `internal/app/workers`)**
  - [x] `backend/internal/transport/httpserver/bundle_import_worker.go` -> delete
  - [x] `backend/internal/transport/httpserver/entry_evaluation_worker.go` -> delete
  - [x] `backend/internal/transport/httpserver/market_share_worker.go` -> delete
  - [x] `backend/internal/transport/httpserver/game_outcome_worker.go` -> delete
  - [x] `backend/internal/transport/httpserver/strategy_generation_worker.go` -> delete
  - [x] `backend/internal/transport/httpserver/calcutta_evaluation_worker.go` -> delete
- [ ] **Move (feature handler clusters; create feature subpackages)**
  - [ ] `backend/internal/transport/httpserver/handlers_admin_*.go` -> `backend/internal/transport/httpserver/admin/*`
  - [ ] `backend/internal/transport/httpserver/handlers_auth*.go` -> `backend/internal/transport/httpserver/auth/*`
  - [ ] `backend/internal/transport/httpserver/handlers_analytics.go` -> `backend/internal/transport/httpserver/analytics/*`
  - [x] `backend/internal/transport/httpserver/handlers_ml_analytics*.go` -> `backend/internal/transport/httpserver/mlanalytics/*`
  - [x] `backend/internal/transport/httpserver/handlers_model_catalogs.go` -> `backend/internal/transport/httpserver/modelcatalogs/*`
  - [x] `backend/internal/transport/httpserver/handlers_tournaments*.go` + `handlers_tournament_analytics.go` -> `backend/internal/transport/httpserver/tournaments/*`
  - [ ] `backend/internal/transport/httpserver/handlers_bracket.go` -> `backend/internal/transport/httpserver/bracket/*`
  - [x] `backend/internal/transport/httpserver/handlers_calcuttas*.go` -> `backend/internal/transport/httpserver/calcuttas/*`
  - [ ] `backend/internal/transport/httpserver/handlers_portfolios*.go` -> `backend/internal/transport/httpserver/portfolios/*`
  - [ ] `backend/internal/transport/httpserver/handlers_entry_evaluation_requests.go` -> `backend/internal/transport/httpserver/entryevaluationrequests/*`
  - [ ] `backend/internal/transport/httpserver/handlers_run_progress.go` -> `backend/internal/transport/httpserver/runprogress/*`
  - [ ] `backend/internal/transport/httpserver/handlers_entry_runs.go` -> `backend/internal/transport/httpserver/entryruns/*` (DB logic evicted into `internal/app/strategy_runs`)
  - [ ] `backend/internal/transport/httpserver/handlers_strategy_generation_runs.go` -> `backend/internal/transport/httpserver/strategygenerationruns/*` (DB logic evicted into `internal/app/strategy_runs`)
  - [ ] `backend/internal/transport/httpserver/handlers_lab_entries*.go` -> `backend/internal/transport/httpserver/labentries/*`
  - [ ] `backend/internal/transport/httpserver/handlers_synthetic_*.go` -> `backend/internal/transport/httpserver/synthetic/*` (DB logic evicted into `internal/app/synthetic_scenarios`)
  - [ ] `backend/internal/transport/httpserver/handlers_suite_calcutta_evaluations*.go` -> `backend/internal/transport/httpserver/simulations/*` (DB logic evicted into `internal/app/suite_evaluations`; rename away from “suite” still pending)
- [ ] **Move tests with their packages**
  - [ ] `backend/internal/transport/httpserver/http_helpers_test.go` -> move with `httperr/*` package
  - [ ] `backend/internal/transport/httpserver/handlers_*_test.go` -> move with feature package

## Move map (seeded from initial inventory)
- [x] `backend/internal/transport/httpserver/api_keys_repository.go` -> delete (replace uses by injecting an `app` service/port; construct `adapters/db` repo in wiring outside `transport/httpserver`)
- [x] `backend/internal/transport/httpserver/auth_repository.go` -> delete (replace uses by injecting an `app` service/port; construct `adapters/db` repo in wiring outside `transport/httpserver`)
- [x] `backend/internal/transport/httpserver/authz_repository.go` -> delete (replace uses by injecting an `app` service/port; construct `adapters/db` repo in wiring outside `transport/httpserver`)
- [x] `backend/internal/transport/httpserver/user_repository.go` -> delete (replace uses by injecting an `app` service/port; construct `adapters/db` repo in wiring outside `transport/httpserver`)
- [x] `backend/internal/transport/httpserver/middleware.go` -> `backend/internal/transport/httpserver/middleware/middleware.go`
- [x] `backend/internal/transport/httpserver/middleware_auth.go` -> `backend/internal/transport/httpserver/middleware/auth.go`
- [x] `backend/internal/transport/httpserver/entry_evaluation_worker.go` -> delete (moved out of `transport/httpserver`)
- [x] `backend/internal/transport/httpserver/suite_calcutta_evaluation_worker.go` -> delete (moved out of `transport/httpserver`)
- [x] `backend/internal/transport/httpserver/bundle_import_worker.go` -> delete (moved out of `transport/httpserver`)
- [x] `backend/internal/transport/httpserver/calcutta_evaluation_worker.go` -> delete (moved out of `transport/httpserver`)
- [x] `backend/internal/transport/httpserver/game_outcome_worker.go` -> delete (moved out of `transport/httpserver`)
- [x] `backend/internal/transport/httpserver/market_share_worker.go` -> delete (moved out of `transport/httpserver`)
- [x] `backend/internal/transport/httpserver/strategy_generation_worker.go` -> delete (moved out of `transport/httpserver`)
