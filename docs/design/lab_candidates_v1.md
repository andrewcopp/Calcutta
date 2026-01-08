# WS-B: Lab model — Candidates per real Calcutta (v1)

## Status
Draft

## Goal
Refactor Lab so it produces **Candidates** for **real calcuttas** without any dependency on Sandbox concepts.

Specifically:
- Lab produces one Candidate per real Calcutta per config (config is explicit and provenance-safe).
- Lab requires explicit upstream artifacts / run IDs (no “latest run” selection heuristics).
- Lab does not use cohorts, scenarios, synthetic calcuttas, or simulated calcuttas.

This workstream is about the **Lab candidate generation + API contract**.

## Non-goals
- Sandbox persistence model (WS-A).
- Worker decomposition / run schema changes beyond what’s required for candidate generation (WS-C).
- Naming migration work (WS-D).

---

# Core definitions

## Candidate
A **Candidate** is an immutable (or versioned) result of an entry optimization run.

Properties:
- A Candidate is a *thing worth evaluating*, not an evaluation itself.
- Candidate identity is stable and does not depend on display name.
- Candidates are produced for a specific:
  - `calcutta_id`
  - `tournament_id` (implicitly via calcutta)
  - optimization inputs (explicit upstream artifact/run IDs)
  - optimizer key (and any other algorithm/model keys)

Candidates may later be imported into the Sandbox as a `simulated_entry` (WS-A).

---

# Data model (current + proposed)

## Current reality (to remove)
Lab currently uses Sandbox-ish tables as its control plane:
- `derived.synthetic_calcutta_cohorts`
- `derived.synthetic_calcuttas`
- `focus_strategy_generation_run_id` on synthetic calcuttas as the “coverage” marker

This must be removed.

## Proposed v1 model
We have two viable approaches; pick one for implementation.

### Option B1 (minimal schema change): Candidate is `derived.candidates` + its source artifact
Use existing `derived.candidates` as the durable identity and point it at a run artifact.

Suggested `derived.candidates` fields (existing or to add as needed):
- `id`
- `source_kind` = `entry_artifact`
- `source_entry_artifact_id` (FK to `derived.run_artifacts.id`)
- `display_name`
- `metadata_json`
- timestamps

Then add a separate mapping table for “candidate per calcutta per config”.

### Option B2 (explicit mapping table first-class): `derived.lab_calcutta_candidates`
Introduce a dedicated table that records Lab’s canonical output per calcutta/config.

`derived.lab_calcutta_candidates` (proposed):
- `id uuid pk`
- `calcutta_id uuid not null` (FK to `core.calcuttas.id`)
- `tournament_id uuid not null` (denormalized for query convenience)
- `candidate_id uuid not null` (FK to `derived.candidates.id`)
- `strategy_generation_run_id uuid not null` (FK to `derived.strategy_generation_runs.id`)
- `market_share_run_id uuid not null`
- `market_share_artifact_id uuid not null` (FK to `derived.run_artifacts.id`)
- `advancement_run_id uuid not null` (or `game_outcome_run_id` until WS-D)
- `optimizer_key text not null`
- `starting_state_key text not null`
- `excluded_entry_name text null`
- `git_sha text null`
- `created_at/updated_at/deleted_at`

Uniqueness:
- Unique constraint on `(calcutta_id, optimizer_key, market_share_artifact_id, advancement_run_id, starting_state_key, COALESCE(excluded_entry_name,''))` where `deleted_at is null`.

Recommendation:
- Prefer **Option B2** because it makes the Lab contract explicit and queryable, and avoids overloading unrelated tables.

---

# Required provenance inputs (non-negotiable)
When generating a candidate for a calcutta, Lab must have explicit inputs.

For each candidate generation request we must specify (directly or deterministically via pinned config):
- `calcutta_id`
- `advancement_run_id` (or `game_outcome_run_id`)
- `market_share_artifact_id` (preferred) or `(market_share_run_id + metrics artifact id)`
- `optimizer_key`
- `starting_state_key`
- `excluded_entry_name` (optional)

No “latest run” selection for correctness-critical generation.

---

# API surface (v1)

## List Lab coverage
`GET /api/lab/candidates/coverage`

Response:
- For the active config(s), how many calcuttas have candidates.

Notes:
- Lab currently assumes “all calcuttas”; this endpoint should reflect that.

## Generate candidates (bulk)
`POST /api/lab/candidates/generate`

Request:
- `optimizerKey: string`
- `startingStateKey?: string`
- `excludedEntryName?: string`
- `advancementRunId: string` OR a pinned reference describing which advancement model run set to use
- `marketShareArtifactIdByCalcutta?: Record<calcuttaId, artifactId>` OR a pinned selection rule that is explicit and stable

Response:
- counts created/skipped/failed + failure list

Notes:
- v1 can be “generate for all calcuttas” only.
- Later we can add filtering (years, tournaments) if needed.

## Candidate detail
`GET /api/lab/candidates/{candidateId}`

Response:
- provenance: calcutta_id, upstream run/artifact ids, optimizer key, params
- output: link to the entry bid set artifact / rows

---

# Execution model
Candidate generation is async:
- Enqueue a run job (or per-calcutta jobs) and write progress events.
- The worker produces:
  - `strategy_generation_runs` row
  - a `metrics` artifact
  - the bid set (in DB tables)
  - a `candidate` identity pointing to the artifact
  - a `lab_calcutta_candidates` mapping row

---

# Migration plan

## Phase B0 — Schema
- Add `derived.lab_calcutta_candidates` (if using Option B2).

Acceptance:
- Schema exists; no behavior change.

## Phase B1 — New Lab endpoints (read-only)
- Add `/api/lab/candidates/coverage` and `/api/lab/candidates/{id}`.
- Implement using the new table.

Acceptance:
- Lab coverage works without referencing synthetic/snapshot/sandbox tables.

## Phase B2 — Generation (write path)
- Implement `/api/lab/candidates/generate`.
- Require explicit upstream IDs.

Acceptance:
- Running generation produces candidates for all calcuttas (or fails with explicit missing-input errors).

## Phase B3 — Cutover UI and delete legacy Lab-over-Sandbox logic
- Remove/disable:
  - `handlers_lab_entries.go` flows that write `derived.synthetic_calcuttas`.
  - any Lab reliance on cohorts/scenarios.

Acceptance:
- Lab no longer reads/writes synthetic/snapshot/sandbox tables.

---

# Parallelization and prerequisites
- Can proceed in parallel with WS-A.
- Importing candidates into Sandbox is an integration point but does not block WS-B core.
- Worker refactors (WS-C) can proceed in parallel; WS-B only needs a stable “create strategy generation run + artifacts + candidate row” contract.

---

# Open questions
- Is a Candidate just a pointer to a `strategy_generation` metrics artifact, or does it need its own “bids artifact” as first-class?
- Should we allow multiple candidates per calcutta for the same config (versioning), or enforce a strict unique mapping?
- Do we want “coverage” to be per tournament year (subset) later, or stay global?

# Open tasks
- [ ] Choose the implementation option for v1 (Option B1 vs Option B2) and record the decision here.
- [ ] Add `derived.lab_calcutta_candidates` schema (if using Option B2) and associated indexes/constraints.
- [ ] Implement `GET /api/lab/candidates/coverage` backed by the new schema.
- [ ] Implement `GET /api/lab/candidates/{candidateId}` backed by the new schema.
- [ ] Implement `POST /api/lab/candidates/generate` (generate for all calcuttas) with explicit upstream inputs required.
- [ ] Ensure candidate generation creates/links:
  - [ ] `strategy_generation_runs`
  - [ ] `run_artifacts` (metrics)
  - [ ] `derived.candidates` row (stable identity)
  - [ ] `derived.lab_calcutta_candidates` mapping row (if using Option B2)
- [ ] Add an invariant test/assertion: Lab endpoints must not read/write Sandbox tables (`synthetic_*`, `simulated_*`, snapshots).
- [ ] Delete/cut over legacy Lab-over-Sandbox logic in `handlers_lab_entries.go`.
- [ ] Confirm UI no longer calls `/api/lab/entries*` endpoints once cutover is complete.
