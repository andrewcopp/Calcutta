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

Then add the Lab provenance fields to `derived.candidates` so we can enforce “one candidate per calcutta per config” without another table.

Suggested additional fields (to add as needed):
- `calcutta_id`
- `tournament_id`
- `strategy_generation_run_id`
- `market_share_run_id`
- `market_share_artifact_id`
- `advancement_run_id`
- `optimizer_key`
- `starting_state_key`
- `excluded_entry_name`
- `git_sha`

Decision (v1):
- Use **Option B1** (store Lab provenance on `derived.candidates`).

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

## List candidates
`GET /api/lab/candidates`

Notes:
- Prefer list+filter over specialized “coverage” endpoints.
- v1 should support filters + pagination.

## Create candidates
`POST /api/lab/candidates`

Notes:
- v1 should support creating a single candidate per request; callers can loop per calcutta.
- If we support bulk, it should be transactional.

## Candidate detail
`GET /api/lab/candidates/{candidateId}`

Response:
- provenance: calcutta_id, upstream run/artifact ids, optimizer key, params
- output: link to the entry bid set artifact / rows

## Delete candidate
`DELETE /api/lab/candidates/{candidateId}`

Notes:
- Used for client-driven reruns: client deletes the old candidate then creates a new one.
- Server performs a soft delete (`deleted_at`) for both candidate + bids.

---

# Execution model
Candidate generation is async:
- Enqueue a run job (or per-calcutta jobs) and write progress events.
- The worker produces:
  - `strategy_generation_runs` row
  - a `metrics` artifact
  - the bid set (in DB tables)
  - a `candidate` identity pointing to the artifact
  - a `derived.candidates` row (stable identity)

---

# Migration plan

## Phase B0 — Schema
- Add provenance fields to `derived.candidates`.

Acceptance:
- Schema exists; no behavior change.

## Phase B1 — New Lab endpoints (read-only)
- Add `/api/lab/candidates` and `/api/lab/candidates/{id}`.
- Implement using `derived.candidates`.

Acceptance:
- Lab list+filter works without referencing synthetic/snapshot/sandbox tables.

## Phase B2 — Creation (write path)
- Implement `/api/lab/candidates`.
- Require explicit upstream IDs.

Acceptance:
- Creating candidates produces the requested candidate(s) (or fails with explicit missing-input errors).

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

Decisions (v1):
- Candidate owns bids via `derived.candidate_bids` (not a first-class “bids artifact”).
- Versioning/reruns are client-driven: `DELETE` then `POST` (server uses soft delete).
- Listing remains global scope for now.

# Open tasks
- [x] Choose the implementation option for v1 (Option B1 vs Option B2) and record the decision here.
- [x] Add provenance fields + constraints/indexes to `derived.candidates`.
- [x] Implement `GET /api/lab/candidates` backed by the new schema.
- [x] Implement `GET /api/lab/candidates/{candidateId}` backed by the new schema.
- [x] Implement `POST /api/lab/candidates` with explicit upstream inputs required.
- [x] Ensure candidate generation creates/links:
  - [x] `strategy_generation_runs`
  - [x] `run_artifacts` (metrics)
  - [x] `derived.candidates` row (stable identity)
  - [x] `derived.candidates.source_entry_artifact_id` points at the `strategy_generation` metrics artifact
- [x] Decide against adding the invariant test/assertion (removed per repo testing rules).
- [x] Delete/cut over legacy Lab-over-Sandbox logic in `handlers_lab_entries.go`.
- [x] Confirm UI no longer calls `/api/lab/entries*` endpoints once cutover is complete.
