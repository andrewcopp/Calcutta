# WS-C: Simulation / Evaluation runs + worker decomposition (v1)

## Status
Draft

## Goal
Make simulations and evaluations:
- **agnostic to why they were created** (Lab-driven vs Sandbox-driven)
- operationally reliable (progress, retries, idempotency)
- easier to understand and maintain (pure steps, explicit contracts)

This workstream defines:
- the relationship between **SimulationRun** and **CalcuttaEvaluationRun**
- the canonical evaluation outputs and how they’re stored
- how we refactor the worker into pure, testable steps

## Non-goals
- Sandbox persistence schema (WS-A), except as an input dependency.
- Lab Candidate generation model (WS-B), except as an input dependency.
- transport/httpserver package cleanup (tracked separately).

---

# Core definitions

## SimulationRun (user-facing)
A **SimulationRun** is the user-visible async run that answers:
> “How do entries perform under simulated tournament outcomes for this evaluation fixture?”

A SimulationRun should be defined by:
- **what fixture** it evaluates (SimulatedCalcutta)
- **what simulation parameters** were used (n_sims, seed, starting_state_key)
- **what upstream models** produced the simulated tournaments (advancement model/run)

SimulationRuns must not depend on Lab concepts; they may optionally record provenance if they were created from Lab artifacts.

## CalcuttaEvaluationRun (engine-room)
A **CalcuttaEvaluationRun** is the internal evaluation record that ties together:
- a set of tournament outcomes (simulated tournaments)
- a specific set of entries (from the fixture)
- computed outcome distributions and performance metrics

---

# Key design decision: do we keep both tables?
We need one clear user-facing contract.

## Option C1 (recommended v1): keep both, but enforce a strict 1:1 relationship
- `derived.simulation_runs` remains the user-facing run.
- `derived.calcutta_evaluation_runs` remains the internal evaluation record.
- Enforce and document:
  - a SimulationRun creates exactly one CalcuttaEvaluationRun
  - CalcuttaEvaluationRun is not created independently (except possibly for backfills/tests)

Practical implementation:
- Add `calcutta_evaluation_runs.simulation_run_id` (unique) OR enforce that `simulation_runs.calcutta_evaluation_run_id` is always non-null on success.

Benefits:
- minimal disruption to existing tables and query usage (`entry_performance` already keys off evaluation runs)
- clear layering: simulation orchestration vs evaluation compute

## Option C2: collapse into a single run table
- Remove `calcutta_evaluation_runs` and key all evaluation outputs directly on `simulation_runs.id`.

Benefits:
- fewer concepts

Costs:
- likely a large migration (many tables and queries reference `calcutta_evaluation_run_id`)

Recommendation:
- Start with **Option C1** to get clarity and incremental refactorability.

---

# Inputs and dependencies (post WS-A)
A SimulationRun evaluates exactly one SimulatedCalcutta.

### Required inputs
- `simulated_calcutta_id`
- `advancement_run_id` (or `game_outcome_run_id` until WS-D)
- `n_sims`
- `seed`
- `starting_state_key`

### Optional provenance inputs
- `source_kind` = `sandbox_manual` | `lab_candidate` | `other`
- `source_candidate_id` (if created from a Lab candidate)

---

# Evaluation fixture source of truth
After WS-A Phase A2 cutover:
- Payouts, scoring rules, and entries for evaluation are loaded from `derived.simulated_*`.
- Focus/highlight is by `highlighted_simulated_entry_id`.

No implicit snapshot creation or repair should occur as part of starting a SimulationRun.

---

# Outputs

## Persisted headline metrics (on SimulationRun)
Persist a small summary directly on `derived.simulation_runs` for UI list views:
- status
- our_rank / our_mean_normalized_payout / our_p_top1 / our_p_in_money
- total_simulations
- started_at / finished_at / error_message

## Canonical evaluation outputs (on CalcuttaEvaluationRun)
Keep existing DB-backed output tables as source of truth:
- `derived.entry_performance` (per entry)
- `derived.entry_simulation_outcomes` (if still needed)
- other per-team/per-entry breakdowns as needed

Artifacts:
- A `metrics` artifact must exist for each completed SimulationRun.
- Additional artifacts are optional; DB remains the source of truth (no `storage_uri` required).

---

# Worker decomposition
The existing worker currently does orchestration + compute in one method.

## Target internal call graph
Refactor into pure steps (functions/services) with explicit inputs/outputs:

1) `ResolveSimulationInputs(runID) -> ResolvedInputs`
- load run row
- validate required fields
- load simulated calcutta fixture (rules/payouts/entries)
- validate invariants

2) `SimulateTournaments(resolved) -> TournamentSimResult`
- run tournament simulation using `advancement_run_id`, `n_sims`, `seed`, `starting_state_key`

3) `EvaluateCalcutta(resolved, tournamentSimResult) -> evaluationRunID + summary`
- create a `calcutta_evaluation_runs` row
- write `entry_performance` etc.

4) `PersistSimulationSummary(runID, summary)`
- update headline fields on `simulation_runs`

5) `WriteArtifacts(runID, summary)`
- write `run_artifacts` (metrics)

6) `EmitProgress(runID, phase, percent, message)`
- consistent, structured progress events

### Idempotency rules
Each step must be safe under retry:
- if a step created an evaluation run already, reuse it or fail fast with a clear error
- use unique constraints / status transitions to prevent duplicate evaluation runs

---

# Job/worker runtime
- Continue using `derived.run_jobs` as the queue.
- Continue writing progress to `derived.run_progress_events`.
- Standardize status transitions and attempt limits (already mostly present).

---

# Migration plan

## Phase C0 — Document and enforce the relationship
- Choose Option C1 vs C2 (recommend C1).
- Add DB constraints if needed to make 1:1 relationship explicit.

Acceptance:
- It is impossible (or at least strongly prevented) for one SimulationRun to fan out into multiple evaluation runs.

## Phase C1 — Pure-function refactor (no behavior change)
- Refactor worker into the target internal call graph.

Acceptance:
- Output tables and run summary fields match current behavior.

## Phase C2 — Switch simulation inputs to SimulatedCalcuttas
- After WS-A Phase A2, update worker to load fixture from `derived.simulated_*`.

Acceptance:
- Simulation runs can be created and executed without reading synthetic/snapshot tables.

## Phase C3 — Normalize focus/highlight semantics
- Ensure the run summary and UI focus flows use `highlighted_simulated_entry_id` rather than name-matching or snapshot-entry matching.

Acceptance:
- Highlighting is stable and does not depend on display names.

---

# Parallelization and prerequisites
- Phase C1 can be done immediately.
- Phase C2 depends on WS-A Phase A2 (worker can’t read simulated tables until they exist and are populated).
- WS-B can proceed independently; simulations should be able to run without Lab candidate provenance.

---

# Open questions
- Do we want SimulationRuns to support evaluating only a subset of entries, or always “all entries in the fixture”?
- Do we need a first-class concept for “tournament simulation result set” separate from calcutta evaluation, or is determinism by seed sufficient?
- Are there any evaluation outputs currently too large or slow in DB that would justify reintroducing `storage_uri` later?

# Open tasks
- [ ] Choose and record the table relationship decision (Option C1 vs C2).
- [ ] Add/adjust DB constraints to enforce the chosen relationship.
- [ ] Refactor `SimulationWorker` into pure internal steps (ResolveInputs / SimulateTournaments / EvaluateCalcutta / PersistSummary / WriteArtifacts).
- [ ] Add retry/idempotency guardrails to prevent duplicate evaluation runs on worker retry.
- [ ] Standardize progress phases/messages emitted by simulation execution.
- [ ] Update simulation run creation endpoints to require `simulated_calcutta_id` and explicit upstream run IDs.
- [ ] After WS-A Phase A2, switch fixture loading to `derived.simulated_*`.
- [ ] Update UI drill-down pages/links to pivot around SimulationRun (user-facing) and hide CalcuttaEvaluationRun unless needed for debugging.
- [ ] Add a regression test: simulation execution should not create/repair fixture state implicitly.
