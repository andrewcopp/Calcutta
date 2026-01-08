# WS-D: Naming consolidation + API migration plan (v1)

## Status
Draft

## Goal
Consolidate naming across the system without leaving long-lived half-migrations.

This workstream defines:
- the **canonical vocabulary** for public API + UI
- the **compatibility/deprecation plan** for legacy routes and terms
- guidance on **when to rename DB objects** vs when to introduce new tables and delete old ones

## Non-goals
- Implementing the Sandbox model (WS-A) or Lab model (WS-B).
- Worker refactors and table relationship decisions (WS-C).
- transport/httpserver package layout cleanup (tracked separately).

---

# Guiding principles
- Prefer **new canonical resources** + **explicit cutovers** over perpetual aliasing.
- Avoid renaming DB tables that are about to be dropped.
- When a rename is unavoidable, do it as an isolated phase with clear acceptance criteria.
- Keep semantics stable: naming changes must not change business meaning.

---

# Canonical vocabulary (target)

## Sandbox
- **SimulatedCalcutta** (replaces “synthetic calcutta”, “scenario”, and snapshot-based mental models)
- **SimulatedEntry** (replaces “synthetic entry” and snapshot-entry-based naming)
- Optional grouping: **Cohort** (Sandbox-only grouping for investigation)

## Lab
- **Candidate** (output of entry optimization; immutable/versioned identity)

## Models
- **Advancement** (replaces “game outcomes” / “predicted_game_outcomes” naming in user-facing surfaces)
- **Market share** (already consolidated)

## Runs
- **SimulationRun** (user-facing)
- **CalcuttaEvaluationRun** (internal, subordinate to SimulationRun; kept or collapsed per WS-C)

---

# Legacy → canonical mapping

## Resource nouns
- `synthetic_calcuttas` → `simulated_calcuttas`
- `synthetic_entries` → `simulated_entries`
- `suite_executions` / `simulation_run_batches` → `simulation_batches` (or keep `simulation_run_batches` as internal)

## Model/run kinds
- `game_outcome` (run kind / tables) → `advancement`

Notes:
- If DB tables are later renamed, the canonical term should already be in API/UI first.

---

# API migration strategy

## Phase D0 — Introduce canonical routes (additive)
- Add new canonical routes and DTOs that use the new vocabulary.
- Keep existing routes operational.

Example route migrations:
- `/api/synthetic-calcuttas` → `/api/simulated-calcuttas`
- `/api/synthetic-calcuttas/{id}/synthetic-entries` → `/api/simulated-calcuttas/{id}/entries`

## Phase D1 — Compatibility shims (short-lived)
- Keep legacy routes as thin aliases that call canonical handlers/services.
- Avoid alias handlers re-implementing logic.

Rules:
- Legacy routes must not be the only way to access new data.
- Legacy routes should emit a clear deprecation signal (response header, logs, and/or UI banner).

## Phase D2 — Cutover UI
- Update frontend to call canonical routes only.

## Phase D3 — Remove legacy routes
- Delete legacy routes and DTOs.
- Delete old handler files if they only supported legacy routes.

---

# DB naming strategy

## Prefer “new tables + drop old” over “rename in place”
This is especially important for:
- snapshots/synthetic/scenario concepts that are being eliminated.

Concretely:
- WS-A creates new `derived.simulated_*` tables and later drops snapshot/synthetic tables.
- WS-B creates new Lab-owned mapping tables and deletes Lab-over-Sandbox logic.

## When to rename DB tables
Only rename DB tables/columns when:
- the table is not being replaced/dropped soon, and
- the renaming reduces long-term confusion more than it costs.

Given current direction:
- Renaming *snapshot/synthetic* DB tables is not recommended; they should be dropped.
- Renaming *game_outcome* → *advancement* may be valuable, but can be staged:
  - API/UI uses “advancement” first
  - DB rename later if desired, or keep DB run_kind as `game_outcome` while presenting “advancement” externally

---

# Migration notes for “game_outcomes” → “advancement”

## Option D-Adv-1 (lowest risk): rename only in API/UI
- Keep DB run kind/table names as-is (`game_outcome_runs`).
- Expose them as “advancement” in API DTOs and UI.

## Option D-Adv-2: rename DB run kind/table names
- Requires coordinated migrations across:
  - `derived.game_outcome_runs` table
  - `run_jobs.run_kind` values
  - `run_artifacts.run_kind` values
  - any foreign keys / code references

Recommendation:
- Start with Option D-Adv-1.
- Decide on Option D-Adv-2 only after WS-A/WS-B cutovers reduce churn.

---

# Parallelization and prerequisites
- This workstream can proceed in parallel with WS-A/WS-B/WS-C if it focuses on **API/UI naming and aliasing**.
- Avoid renaming or migrating DB objects that WS-A/WS-B plan to delete.
- The final removal of legacy routes should be scheduled after WS-A and WS-B cutovers.

---

# Open tasks
- [ ] Enumerate all legacy endpoints that expose synthetic/snapshot/suite nouns and map them to canonical routes.
- [ ] Decide canonical route structure for simulations/batches (keep existing naming vs introduce new public resource name).
- [ ] Implement canonical Sandbox routes for simulated calcuttas/entries (WS-A) and add short-lived legacy aliases.
- [ ] Implement canonical Lab routes for candidates (WS-B) and deprecate `/api/lab/entries*` legacy routes.
- [ ] Add a deprecation mechanism (pick one):
  - [ ] response header (`Deprecation`, `Sunset`, or custom)
  - [ ] structured log line on legacy route access
  - [ ] UI banner if legacy routes are still in use
- [ ] Update frontend to use canonical routes only.
- [ ] Add a CI guardrail to prevent adding new routes with legacy nouns (`synthetic`, `suite`, `snapshot`) except in legacy alias packages.
- [ ] Schedule and execute removal of legacy routes after cutover.
- [ ] Decide approach for “game_outcomes” → “advancement” (Option D-Adv-1 vs D-Adv-2) and record decision.
