# Lab / Sandbox Separation — Master Refactor Plan

## Goal
Converge on a clean separation between:
- **Lab**: produces immutable **Candidates** (optimized entry artifacts) for each real Calcutta.
- **Sandbox**: owns **SimulatedCalcuttas** (self-contained evaluation fixtures) and **SimulatedEntries** (editable entries inside a simulated calcutta), and runs simulations/evaluations against them.

This plan is deliberately split into multiple independent workstreams (subdocs) so we can execute in parallel and avoid half-finished migrations.

## Target end-state (authoritative nouns)
- **Real world**
  - `core.calcuttas` (source of truth)
  - `core.entries` / `core.entry_teams`
  - `core.payouts` / `core.calcutta_scoring_rules`
- **Lab**
  - `derived.candidates` represent “this entry is a candidate for evaluation” and is the primary output of entry optimization.
  - Candidates are immutable/versioned (identity does not depend on display name).
- **Sandbox**
  - `derived.simulated_calcuttas` represent evaluation fixtures.
    - Always tied to a tournament.
    - Can be created from scratch, or instantiated from a real calcutta.
    - After creation, can diverge arbitrarily from the seed.
  - `derived.simulated_entries` represent entries inside a simulated calcutta.
    - Real entries copied into a simulated calcutta are editable.
    - Candidates imported into Sandbox become simulated entries.
  - “Focus/highlight” is by `simulated_entry_id`.

## Explicit create flows (Sandbox)
- **Create from scratch**: create a simulated calcutta with rules + payouts only; entries can be added later.
- **Instantiate from real calcutta**: copy rules + payouts + entries as the starting point; entries remain editable.

## Non-goals
- Perfect long-term platform architecture.
- A single mega-plan that requires serial execution.

## Current problems this plan resolves
- Lab incorrectly depends on Sandbox concepts (synthetic calcuttas/cohorts/scenarios).
- Multiple overlapping concepts (synthetic calcuttas, snapshots, scenarios) create confusion.
- Implicit “repair/create missing state” behavior exists in some write paths.

---

# Workstreams (subdocs)
Each workstream has its own phases, acceptance criteria, and cutover plan.

## WS-A — Sandbox model: SimulatedCalcuttas + SimulatedEntries
**Doc**: `docs/design/simulated_calcuttas_v1.md`

**Outcome**:
- Introduce `derived.simulated_calcuttas` + `derived.simulated_entries` (and supporting tables).
- Provide explicit create flows (from scratch / from real calcutta).
- Remove reliance on `core.calcutta_snapshots*` and `derived.synthetic_calcuttas*`.

**Parallelization**:
- Can be developed largely in parallel with WS-B and WS-C.

**Prerequisites**:
- None.

**Blocks**:
- Blocks WS-C cutover if simulations need the new simulated tables.

---

## WS-B — Lab model: Candidates per real Calcutta (no cohorts)
**Doc**: `docs/design/lab_candidates_v1.md`

**Outcome**:
- Lab produces one Candidate per real calcutta per config.
- Lab has explicit upstream artifacts/run IDs.
- Lab does not create or depend on any simulated/synthetic calcutta concept.

**Parallelization**:
- Can be developed in parallel with WS-A.

**Prerequisites**:
- None.

**Blocks**:
- Unblocks WS-A “import candidate into simulated entry” integration once Candidate identity/contract is stable.

---

## WS-C — Simulation / evaluation runs and worker decomposition
**Doc**: `docs/design/simulation_evaluation_runs_v1.md`

**Outcome**:
- Clarify whether we keep both `simulation_runs` and `calcutta_evaluation_runs` or collapse to a single user-facing run.
- Refactor worker into pure, testable steps.
- Ensure simulation inputs are agnostic to why the run exists (Lab-origin vs Sandbox-origin).

**Parallelization**:
- Pure-function refactor can start immediately.

**Prerequisites**:
- For a full cutover to simulated tables, this depends on WS-A being implemented.

---

## WS-D — Naming + API migration (synthetic → simulated, suite → cohort, game_outcomes → advancement)
**Doc**: `docs/design/naming_and_api_migration_v1.md`

**Outcome**:
- Finish the naming consolidation without half-done shims.
- Provide clear deprecation windows and cutover points.

**Parallelization**:
- Can proceed in parallel, but should avoid renaming tables that are about to be dropped.

**Prerequisites**:
- Prefer to land after WS-A/WS-B schemas are stable.

---

## WS-E — transport/httpserver cleanup (ongoing)
**Doc**: `docs/design/transport_httpserver_cleanup.md` (existing)

**Outcome**:
- Continue feature-first package split.
- Reduce coupling (handlers thin, no workers/repo shims in transport).

**Parallelization**:
- Can be executed independently and continuously.

---

# Cutover principles (avoid half-finished migrations)
- Prefer explicit phases: backfill -> dual-read -> cutover writes -> worker cutover -> delete old.
- Once cutover is complete, freeze old write paths and remove shims quickly.
- No implicit repair behavior in read paths; “repair” must be explicit (admin endpoint/job).

# Links
- `docs/design/lab_and_sandbox_refactor_v2.md`
- `docs/design/lab_sandbox_separation_master_plan.md`
- `docs/design/simulated_calcuttas_v1.md`
- `docs/design/lab_candidates_v1.md`
- `docs/design/simulation_evaluation_runs_v1.md`
- `docs/design/naming_and_api_migration_v1.md`
- `docs/design/transport_httpserver_cleanup.md`
