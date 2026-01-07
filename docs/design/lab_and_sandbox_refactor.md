# Lab + Sandbox Refactor Plan

## Goal
Consolidate and “lock in” the Lab/Sandbox architecture so we can iterate quickly without accumulating naming, schema, URL, and handler sprawl.

This doc is a task checklist. Keep it updated as we execute the refactor.

## Guiding invariants
- Runs are immutable after completion.
- Runs produce artifacts.
- Evaluations measure how good a candidate/entry was under a scenario.
- Sandbox never recomputes Lab.
- Simulations are seeded and reproducible.
- Simulation artifacts include outcomes for *all entries*, not only a focus entry.
- Batched operations create one Run (one worker job) per item in the batch.
- Synthetic objects are the ground-truth inputs to simulation; real/historical calcuttas can be copied into synthetic calcuttas.

## Durable domain language (proposed)
### Prefixes (what kind of work)
- Advancement
- MarketShare
- Entry
- Simulation
- Synthetic

### Suffixes (what it is)
- Model (implementation of an algorithm)
- Run (an execution)
- Artifact (output dataset)
- Candidate (an entry definition to simulate)
- SyntheticEntry (an entry definition to simulate, with synthetic-only metadata)
- SyntheticCalcutta (a simulation-ready calcutta “world”; may be created from scratch or copied from a real/historical calcutta)
- SyntheticCalcuttaCohort (a user-managed collection of synthetic calcuttas)
- Evaluation (a computed score/metrics for a SyntheticEntry under a SyntheticCalcutta)

### Core concepts
- **Model**: compiled-in algorithm implementation. Discoverable via a code registry.
- **Run**: an execution of a model over a dataset (tournament, calcutta, scenario).
- **Artifact**: typed output from a run.
- **SyntheticEntry**: a candidate entry to be simulated. Often produced by an Entry run (Lab), but can be hand-authored (Sandbox).
- **SyntheticCalcuttaCohort**: user-managed collection of synthetic calcuttas.
- **SyntheticCalcutta**: a calcutta “world” suitable for simulation; may be copied from a real/historical calcutta or created from scratch. Includes a `HighlightedEntry` for cohort summaries.
- **SimulationRun**: execution of simulations for one SyntheticCalcutta and one or more SyntheticEntries.
- **SimulationArtifact**: results of a SimulationRun (full entry leaderboard/outcomes + per-candidate metrics).
- **SimulationEvaluation**: derived metrics/summary for a SyntheticEntry within a SyntheticCalcutta.

## Mapping from current terms (to be filled as we refactor)
- `SuiteCalcuttaEvaluation` -> `SyntheticCalcuttaCohort` (recommended)
- `SuiteCalcuttaEvaluationEntry` -> `SimulationEvaluation` (per SyntheticEntry)
- `StrategyGenerationRun` -> `EntryRun` (if it produces candidates) OR `EntryOptimizationRun` (if it produces optimized entries)

- [ ] Audit existing names in:
  - `backend/internal/transport/httpserver/handlers_suite_calcutta_evaluations.go`
  - `frontend/src/services/suiteCalcuttaEvaluationsService.ts`
  - `backend/internal/transport/httpserver/handlers_strategy_generation_runs.go`
  - and document exact mapping in this section

## Model registries (code, not DB)
We treat the set of available models as part of the codebase.

- [ ] Define stable interfaces for:
  - AdvancementModelInterface
  - MarketShareModelInterface
  - EntryOptimizerInterface
  - SimulationModelInterface (if variants exist; otherwise a single implementation)

- [x] Expose model catalog endpoints for UI discoverability:
  - `GET /advancement-models`
  - `GET /market-share-models`
  - `GET /entry-optimizers`

- [x] Create registries with stable IDs:
  - `AvailableAdvancementModels` (e.g. `kenpom_ratings_v1`, `kenpom_ratings_v2`)
  - `AvailableMarketShareModels`
  - `AvailableEntryOptimizers`

- [x] Ensure every model registry entry includes:
  - `id` (stable string)
  - `display_name`
  - `schema_version` for produced artifacts
  - optional `deprecated` flag

## Runs and artifacts
### Run lifecycle
- [x] Standardize run status fields (`queued`, `running`, `succeeded`, `failed`) via a shared `derived.run_jobs` envelope queue (per-kind workers claim by `run_kind`).
- [x] Standardize timestamps (`created_at`, `started_at`, `finished_at`) on `derived.run_jobs`.
- [ ] Standardize run parameters serialization (`params_json`) and include `seed` where applicable.

### Artifacts
We can keep separate artifact tables/types, but they should share a common contract:
- `run_id`
- `schema_version`
- `storage_uri` / payload pointer
- `summary_json` (small UI-friendly preview)

- [ ] Decide on the final approach:
  - [ ] Option: Typed artifact tables (AdvancementArtifacts, MarketShareArtifacts, EntryArtifacts, SimulationArtifacts) + a shared contract (`run_id`, `schema_version`, `storage_uri`, `summary_json`)

- [ ] Ensure EntryArtifacts explicitly reference exactly one:
  - AdvancementArtifact
  - MarketShareArtifact

## Sandbox durability requirements
Sandbox should support:
- Bulk evaluation of many candidates (Lab output cohort)
- User-managed iterative evaluation (create group, add scenario, evaluate one-by-one)
- Reuse for real-calcuttas forecasting (Day 1 projections)

Decisions:
- Cohort summary metrics come from the most recent completed SimulationRun per SyntheticCalcutta.
- Delete SimulationRuns (and their SimulationArtifacts) older than 30 days.
- Forecast is a UI label for simulation outputs derived from real calcuttas; domain nouns remain `SimulationRun`/`SimulationArtifact`.

### SyntheticCalcuttaCohorts / SyntheticCalcuttas
- [x] Introduce SyntheticCalcuttaCohorts as durable entities (`derived.synthetic_calcutta_cohorts`), backed by `derived.suites` during migration.
- [x] Introduce SyntheticCalcuttas as durable entities (`derived.synthetic_calcuttas`), with backfill from `derived.suite_scenarios` (preserve IDs for compatibility).
- [ ] SyntheticCalcutta should support:
  - optional source calcutta reference (historical or real) for copy/import convenience
  - full set of synthetic entries (all entries are SyntheticEntries)
  - add/edit/delete synthetic entries (not limited to swapping a single entry)
  - HighlightedEntry (the synthetic entry to show in cohort summaries)
  - optional notes/metadata

- [x] SyntheticEntries CRUD (snapshot-backed) endpoints:
  - `GET /synthetic-calcuttas/{synthetic_calcutta_id}/synthetic-entries`
  - `POST /synthetic-calcuttas/{synthetic_calcutta_id}/synthetic-entries`
  - `PATCH /synthetic-entries/{synthetic_entry_id}`
  - `DELETE /synthetic-entries/{synthetic_entry_id}`

### Candidate inputs
- [ ] Support two candidate sources:
  - from EntryArtifacts (Lab) by importing/copying them into SyntheticEntries scoped to a SyntheticCalcutta
  - hand-authored SyntheticEntries (Sandbox)

## API + URL cleanup
Resource-oriented URLs with stable IDs.

- [ ] Define canonical resources:
  - `/advancement-runs`, `/advancement-artifacts`
  - `/market-share-runs`, `/market-share-artifacts`
  - `/entry-runs`, `/entry-artifacts`
  - `/synthetic-calcutta-cohorts`, `/synthetic-calcuttas`, `/synthetic-entries`
  - `/simulation-runs`, `/simulation-artifacts`

- [ ] Ensure navigation uses nested routes only for convenience, not identity.

- [x] Frontend: canonical Sandbox routes use `/sandbox/cohorts` (keep `/sandbox/suites` as an alias during migration).

- [x] Frontend: rename Sandbox pages/services to cohort/simulation naming (keep old filenames/exports as shims during migration).

- [x] Backend: accept `cohortId`/`cohort_id` and `simulationRunBatchId`/`simulation_run_batch_id` aliases for legacy `suite*` params.

- [x] Add compatibility endpoints that proxy/redirect to new resources during migration.

- [x] Implement compatibility-first Synthetic/Simulation endpoints in backend (new URLs, old behavior).

- [x] Cut over frontend services to compatibility endpoints:
  - `suiteCalcuttaEvaluationsService` -> `/simulation-runs`
  - `suitesService` -> `/synthetic-calcutta-cohorts`
  - `syntheticCalcuttasService` -> `/synthetic-calcuttas`
  - `suiteExecutionsService` -> `/simulation-run-batches`

- [x] Smoke test: `/synthetic-calcutta-cohorts`, `/synthetic-calcuttas`, `/synthetic-entries` (create calcutta snapshot, list entries)

- [ ] SyntheticCalcutta creation:
  - [x] Backfill/repair legacy synthetic calcuttas to ensure `calcutta_snapshot_id` is always present
  - [x] `POST /synthetic-calcuttas` creates a snapshot when `calcuttaSnapshotId` is not provided
  - [x] `POST /synthetic-calcuttas` supports optional `source_calcutta_id` to copy/import from a real/historical calcutta
  - [x] Sandbox UI supports creating synthetic calcuttas from a selected source calcutta

### Proposed REST surface (initial)
Catalogs (compiled-in registries):
- `GET /advancement-models`
- `GET /market-share-models`
- `GET /entry-optimizers`

Lab runs/artifacts:
- `POST /advancement-runs` (params include `advancement_model_id`, tournament refs)
- `GET /advancement-runs`
- `GET /advancement-runs/{run_id}`
- `GET /advancement-artifacts/{artifact_id}`

- `POST /market-share-runs` (params include `market_share_model_id`, calcutta refs)
- `GET /market-share-runs`
- `GET /market-share-runs/{run_id}`
- `GET /market-share-artifacts/{artifact_id}`

- `POST /entry-runs` (params include `entry_optimizer_id`, `advancement_artifact_id`, `market_share_artifact_id`)
- `GET /entry-runs`
- `GET /entry-runs/{run_id}`
- `GET /entry-artifacts/{artifact_id}`

Sandbox authoring:
- `POST /synthetic-calcutta-cohorts`
- `GET /synthetic-calcutta-cohorts`
- `GET /synthetic-calcutta-cohorts/{cohort_id}`

- `POST /synthetic-calcuttas` (body supports `cohort_id` and optional `source_calcutta_id`)
- `GET /synthetic-calcuttas/{synthetic_calcutta_id}`
- `PATCH /synthetic-calcuttas/{synthetic_calcutta_id}` (e.g. rename, set `highlighted_entry_id`)

- `GET /synthetic-calcuttas/{synthetic_calcutta_id}/synthetic-entries`
- `POST /synthetic-calcuttas/{synthetic_calcutta_id}/synthetic-entries`
- `PATCH /synthetic-entries/{synthetic_entry_id}`
- `DELETE /synthetic-entries/{synthetic_entry_id}`

Sandbox simulation:
- `POST /synthetic-calcuttas/{synthetic_calcutta_id}/simulation-runs` (params include `simulation_model_id`, `seed`, `n_sims`)
- `GET /simulation-runs/{simulation_run_id}`
- `GET /simulation-runs/{simulation_run_id}/evaluations` (per SyntheticEntry metrics)
- `GET /simulation-artifacts/{artifact_id}`

Batch orchestration (optional grouping record):
- `POST /simulation-run-batches` (body includes `synthetic_calcutta_ids` and run params; creates one SimulationRun per SyntheticCalcutta)
- `GET /simulation-run-batches/{batch_id}` (returns created run ids + status summary)

## Worker contract (Go <-> Python and async jobs)
- [ ] Standardize job submission payload (run kind + params + dataset refs + seed).
- [ ] Standardize progress events (percent + phase + message).
- [ ] Standardize completion contract:
  - run status transitions
  - produced artifacts registration
  - metrics artifact always produced

- [ ] Require seeded simulations:
  - seed is a required run param
  - any RNG usage is derived from the seed

## Backend code refactor (reduce “big file” handlers)
Target: smaller single-responsibility packages/files.

- [ ] Split `handlers_suite_calcutta_evaluations.go` into:
  - routes/wiring
  - request/response DTOs
  - query/service layer
  - simulation domain logic (if any)

- [ ] Split `handlers_ml_analytics*` similarly.

- [ ] Align handler names to the new resources (no “suite” in backend resource names).

## Data migration and cleanup
- [ ] Implement compatibility-first migration:
  - add new tables/resources
  - dual-write from new runs
  - backfill history
  - cut over UI
  - delete/retire deprecated tables and endpoints

- [ ] Create a concrete “final cleanup” checklist:
  - tables to drop
  - endpoints to remove
  - frontend routes to remove
  - docs to update

## UI/UX follow-ups
- [ ] Ensure navigation pivots:
  - dataset (tournament/calcutta)
  - run (timeline + artifacts)
  - evaluation (scenario group + candidate comparison)

- [ ] Define the drill-down pages:
  - RunDetail
  - ArtifactDetail
  - SyntheticCalcuttaCohortDetail
  - SyntheticCalcuttaDetail
  - SimulationRunDetail

## Open questions
- [ ] Decide which simulation result slices are “first-class artifacts” vs computed-on-read.
