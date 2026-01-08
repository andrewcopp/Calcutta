# Lab + Sandbox Refactor v2

## Goal
Finish consolidating Lab/Sandbox into a deterministic, reproducible evaluation pipeline with:
- clear domain nouns (Models, Runs, Artifacts, Cohorts, SyntheticCalcuttas, Candidates)
- reliable async execution
- strict provenance/lineage
- zero state mutation on read paths

This doc is a task checklist. Keep it updated as work lands.

## Guiding invariants (non-negotiable)
- Runs are immutable after completion.
- Read endpoints are read-only (no implicit “repair” and no implicit “seed”).
- Every completed Run produces a `metrics` artifact.
- Every evaluation result is traceable to explicit upstream artifacts and dataset refs.
- Sandbox never recomputes Lab artifacts implicitly.
- Simulations are seeded and reproducible.
- “Candidate identity” is stable and does not depend on display names.

## Current reality (symptoms)
- Some GET/read paths write state (AUTO cohort upserts; snapshot repair-on-read).
- Lab generation uses “latest run” heuristics (time-based selection) rather than pinned lineage.
- Candidate selection/metrics rely on string matching (e.g. “Our Strategy”).
- Workers and orchestration logic still live in transport layer.
- Artifacts are mostly `metrics` only; heavy outputs live in ad-hoc tables.

---

# Phase 0 — Safety and correctness hardening (highest priority)

## 0.1 Eliminate state mutation in read paths
- [x] Remove DB writes from `GET /api/lab/entries` coverage endpoint.
  - **Implementation idea**: move AUTO cohort creation to an explicit admin action or startup sync.
  - **Acceptance**:
    - `GET /api/lab/entries` performs no INSERT/UPDATE.
    - AUTO cohorts can still be created deterministically via a non-GET mechanism.

- [x] Remove/replace “repair snapshot on read” behavior in `handleListSyntheticEntries`.
  - **Acceptance**:
    - `GET /synthetic-calcuttas/{id}/synthetic-entries` performs no writes.
    - Missing snapshot becomes a clear 409 error OR is repaired only via an explicit endpoint/job.
  - **Notes**:
    - Enforced `derived.synthetic_calcuttas.calcutta_snapshot_id` invariant via `NOT NULL` migration.

- [x] Add regression tests (or at minimum integration assertions) that “read endpoints do not write”.
  - **Acceptance**: a test fails if any write query is executed during these read handlers.

## 0.2 Remove name-based correctness hacks
- [x] Replace “Our Strategy” string matching in simulation worker metrics selection.
  - **Implementation idea**: store a stable `focus_snapshot_entry_id` / `focus_candidate_id` on the SimulationRun.
  - **Implementation**:
    - Persist `derived.simulation_runs.focus_snapshot_entry_id` (FK to `core.calcutta_snapshot_entries`).
    - Simulation worker + read paths use `focus_snapshot_entry_id` (resolved to `display_name`) instead of brittle string matching.
  - **Acceptance**:
    - Simulation summary metrics for “focus” do not depend on entry_name.
    - Multiple candidates with arbitrary names work.

## 0.3 Make Lab/Sandbox determinism explicit
- [ ] Stop selecting upstream runs using “latest created_at” heuristics for correctness-critical operations.
  - **Acceptance**:
    - When generating entries or simulations, upstream artifact/run IDs are explicitly specified or deterministically resolved via pinned config.
    - Re-running the same request yields the same upstream references unless explicitly changed.

---

# Phase 1 — Provenance and lineage (make results trustworthy)

## 1.1 Pin upstream inputs for Lab Entry generation
- [x] Add explicit columns (or generated columns) on `derived.strategy_generation_runs` for:
  - `market_share_run_id`
  - `game_outcome_run_id`
  - `excluded_entry_name`
  - `starting_state_key` (if relevant)
  - **Acceptance**:
    - For every StrategyGenerationRun, upstream run IDs are queryable without parsing JSON.
    - Backfill is provided for existing rows where possible.

- [x] Update Lab generation code to populate these fields from the selected upstream runs.
  - **Acceptance**: no StrategyGenerationRun is created without explicit upstream IDs.

## 1.2 Strengthen run_artifacts lineage contract
- [x] Enforce that `strategy_generation` `metrics` artifact references exactly one:
  - `input_market_share_artifact_id`
  - `input_advancement_artifact_id` (or game outcome artifact, depending on the final naming)
  - **Acceptance**:
    - UI can traverse lineage from entry artifact -> upstream artifacts.

- [x] Add validation at artifact write time (worker/service) to prevent missing lineage.
  - **Acceptance**: artifact write fails fast (and marks run failed) if lineage is missing.

## 1.3 Dataset refs become first-class and consistent
- [ ] Normalize dataset ref schema in `derived.run_jobs.params_json` per `run_kind` (tournament_id, calcutta_id, snapshot ids, etc.).
- [ ] Ensure run rows also store dataset refs in columns where practical.
  - **Acceptance**:
    - Given any run_id, you can answer “what dataset was this run over?” with a single query.

---

# Phase 2 — Candidate identity (stop overloading snapshot rows)

## 2.1 Introduce a durable Candidate/SyntheticEntry concept
- [x] Create a first-class table for candidates, e.g. `derived.candidates` or `derived.synthetic_entries`.
   - Suggested fields:
     - `id`
     - `source_kind` (`manual` | `entry_artifact` | `other`)
     - `source_entry_artifact_id` (nullable)
     - `display_name`
     - `created_at`, `updated_at`, `deleted_at`
     - `metadata_json`
   - **Acceptance**:
     - Candidate has stable identity independent of any specific snapshot.

- [x] Create a table that materializes a candidate into a specific synthetic calcutta snapshot, e.g.:
   - `derived.synthetic_calcutta_candidates` with:
     - `synthetic_calcutta_id`
     - `candidate_id`
     - `snapshot_entry_id`
   - **Acceptance**:
     - Same candidate can be evaluated across multiple synthetic calcuttas without copying definition.

## 2.2 Update Sandbox endpoints to operate on Candidate IDs
- [x] Update “import synthetic entry” endpoint to create a Candidate (if not exists) and attach it.
- [x] Update manual entry creation to create a Candidate + attach it.
- [x] Update patch/delete to operate on candidate attachment (and snapshot entry materialization), not directly on snapshot entries.
   - **Acceptance**:
     - UI can list candidates for a synthetic calcutta.
     - Editing a candidate updates the materialized snapshot entry or regenerates it deterministically.

- [ ] Add explicit “candidate” naming in the API surface (alias routes), while keeping current endpoints for compatibility.

---

# Phase 3 — Worker architecture and contracts (make ops sane)

## 3.1 Move orchestration out of transport layer
- [ ] Move simulation worker logic from `transport/httpserver` into `internal/app` (or a dedicated worker package).
  - **Acceptance**:
    - HTTP layer only enqueues and reads.
    - Worker has explicit dependencies (DB repo interfaces, services).

## 3.2 Unify worker runtime patterns across Go and Python
- [ ] Replace “python worker shells out to another python script” with an in-process runner.
  - **Acceptance**:
    - Failures have structured exception info.
    - No stdout JSON parsing is required for success/failure.

- [ ] Standardize retry/backoff and dead-letter policy across run kinds.
  - **Acceptance**:
    - Poison jobs do not loop forever.
    - Operators can see attempt count, last error, last claimed_by.

## 3.3 Structured progress + logs
- [ ] Ensure all run kinds emit progress events in a consistent shape.
- [ ] Ensure progress is queryable by run_id.
  - **Acceptance**:
    - UI can show a timeline: queued -> running (phases) -> succeeded/failed.

---

# Phase 4 — Artifacts: from metrics-only to first-class datasets

## 4.1 Decide the “artifact slices” and storage strategy
- [ ] Decide which outputs should be artifacts vs “computed on read”.
  - Candidates:
    - simulation leaderboard
    - per-entry distribution summaries
    - per-team investment/ownership tables

- [ ] Add `storage_uri` support for non-trivial outputs (e.g. parquet/JSONL) and keep `summary_json` small.
  - **Acceptance**:
    - Artifacts can be fetched without custom queries per run kind.

## 4.2 Cleanup and retention
- [ ] Implement retention policy consistently:
  - delete old simulation artifacts/runs after N days, but keep aggregate metrics if desired.
  - **Acceptance**:
    - Cleanup is an explicit job/command, not implicit.

---

# Phase 5 — API/UX consolidation (finish the refactor)

## 5.1 Remove remaining legacy “suite*” naming
- [ ] Remove remaining suite terminology from handler names, DTOs, routes, and frontend service naming.
  - **Acceptance**:
    - No `suite_` endpoints remain as public API.
    - Compatibility shims are deleted.

## 5.2 Navigation and drill-down pages
- [ ] Define and implement canonical drill-down pages:
  - RunDetail
  - ArtifactDetail
  - CohortDetail
  - SyntheticCalcuttaDetail
  - SimulationRunDetail

- [ ] Ensure UI pivots are clear:
  - dataset (tournament/calcutta)
  - run (timeline + artifacts)
  - evaluation (scenario + candidate comparison)

---

 # Appendix — Suggested PR sequence (small, safe increments)

- [x] PR A: Remove AUTO cohort upsert from GET; add explicit creation path.
- [x] PR B: Remove snapshot repair-on-read; enforce invariant.
- [x] PR C: Replace “Our Strategy” string matching with stable IDs.
- [x] PR D: Add explicit upstream run columns + backfill.
- [x] PR E: Introduce Candidate tables + migrate sandbox endpoints.
- [x] PR F: Move worker logic out of transport layer.
   - Implemented `backend/internal/app/workers` with explicit dependencies.
   - Updated `backend/cmd/workers` to run new workers directly.
   - Replaced `transport/httpserver/*_worker.go` implementations with thin wrappers.
   - Deleted unused `transport/httpserver/run_progress.go`.
   - Deferred: `transport/httpserver/sql_params.go` is still used by HTTP handlers.
- [x] PR G: Artifact storage expansion + retention job.
   - Added `ARTIFACTS_DIR` config and plumbed into `SimulationWorker`.
   - Exported heavy outputs (`derived.entry_performance`, `derived.entry_simulation_outcomes`) to JSONL and stored `storage_uri` in `derived.run_artifacts`.
   - Added `backend/cmd/tools/retain-simulation-runs` retention sweeper (dry-run by default).
- [ ] PR H: Delete legacy suite shims/endpoints; final naming cleanup.
- [ ] PR I: Eliminate remaining “latest-run” heuristics in correctness-critical paths.
- [x] PR J: Enforce `strategy_generation` artifact lineage (`input_*_artifact_id`) + validation.
   - Enforced `marketShareArtifactId` on entry/strategy run creation and persisted `market_share_artifact_id` into `run_jobs.params_json`.
   - Worker now requires and validates upstream lineage before writing the `strategy_generation` metrics artifact.
- [x] PR K: Standardize dataset refs across `run_jobs.params_json` and run tables.
   - Restored `dataset_refs` in the `strategy_generation` `run_jobs` enqueue contract + backfilled missing `dataset_refs` rows.
- [ ] PR L: Add explicit “candidate” naming to API surface (alias routes) while keeping compatibility.
- [ ] PR M: Python worker: replace subprocess runner with in-process execution + structured failures.
- [ ] PR N: Standardize retry/backoff + dead-letter policy across run kinds.
- [ ] PR O: Standardize run progress events and make progress queryable by run.
- [ ] PR P: Artifact slices decision + implement `storage_uri` for at least one non-trivial artifact.
- [ ] PR Q: Implement explicit retention/cleanup job for runs/artifacts.
- [ ] PR R: UI drill-down pages + navigation pivots (runs/artifacts/evaluations) to match canonical nouns.
