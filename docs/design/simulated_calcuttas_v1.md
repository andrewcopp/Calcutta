# WS-A: Sandbox model — SimulatedCalcuttas + SimulatedEntries (v1)

## Status
Implemented (v1)

Phase A4 complete:
- Sandbox UI uses simulated endpoints only.
- Legacy synthetic Sandbox UI (pages/services) removed.
- Legacy synthetic write endpoints return 410 Gone.

Phase A5 (minimal) complete:
- Dropped legacy synthetic tables:
  - `derived.synthetic_calcuttas`
  - `derived.synthetic_calcutta_candidates`
- Removed `synthetic_calcutta_id` from `derived.simulation_runs` (+ updated run job enqueue trigger).
- Removed legacy synthetic scenario production code paths (no route registrations; handlers are inert).

## Goal
Replace the current Sandbox persistence model (synthetic calcuttas + snapshots + candidate attachments) with a clean, self-contained Sandbox model:
- **SimulatedCalcuttas**: evaluation fixtures (rules + payouts + entries)
- **SimulatedEntries**: editable entries inside a simulated calcutta

This workstream is specifically about the **Sandbox data model + API**.

## Non-goals
- Lab Candidate model changes (see WS-B).
- Worker decomposition or run schema changes (see WS-C), beyond updating references.
- transport/httpserver package cleanup (tracked separately).

Note:
- Simulation runs can now be driven directly from a `simulated_calcutta_id` without requiring a lab `game_outcome_run_id`.

---

# Core definitions

## SimulatedCalcutta
A **SimulatedCalcutta** is a Sandbox-owned evaluation fixture.

Properties:
- Always tied to a `tournament_id`.
- May optionally record a `base_calcutta_id` purely as provenance (“seed / starting point”).
- Is **self-contained**: it stores its own payouts, scoring rules, and entries.
- Can diverge arbitrarily from its seed; it is not required to remain “similar” to the real calcutta.

## SimulatedEntry
A **SimulatedEntry** is an editable entry belonging to a SimulatedCalcutta.

Properties:
- Can originate from:
  - manual entry creation
  - copying a real entry from a real calcutta (instantiation path)
  - importing a Lab Candidate
- Remains editable regardless of origin.

## Focus/highlight
Focus/highlight is by `simulated_entry_id`.

---

# Data model (Postgres)
All new tables live under `derived.*`.

## `derived.simulated_calcuttas`
Fields:
- `id uuid primary key`
- `name text not null`
- `description text null`
- `tournament_id uuid not null` (FK to `core.tournaments.id`)
- `base_calcutta_id uuid null` (FK to `core.calcuttas.id`; optional)
- `starting_state_key text not null default 'post_first_four'`
- `excluded_entry_name text null`
- `highlighted_simulated_entry_id uuid null` (FK to `derived.simulated_entries.id`)
- `metadata_json jsonb not null default '{}'::jsonb`
- `created_at timestamptz not null default now()`
- `updated_at timestamptz not null default now()`
- `deleted_at timestamptz null`

Notes:
- We keep `highlighted_simulated_entry_id` on the simulated calcutta for UX convenience.

## `derived.simulated_calcutta_payouts`
Fields:
- `id uuid primary key`
- `simulated_calcutta_id uuid not null` (FK)
- `position int not null`
- `amount_cents int not null`
- `created_at/updated_at/deleted_at`

Constraints:
- Unique `(simulated_calcutta_id, position)` where `deleted_at is null`.

## `derived.simulated_calcutta_scoring_rules`
Fields:
- `id uuid primary key`
- `simulated_calcutta_id uuid not null` (FK)
- `win_index int not null`
- `points_awarded int not null`
- `created_at/updated_at/deleted_at`

Constraints:
- Unique `(simulated_calcutta_id, win_index)` where `deleted_at is null`.

## `derived.simulated_entries`
Fields:
- `id uuid primary key`
- `simulated_calcutta_id uuid not null` (FK)
- `display_name text not null`
- `source_kind text not null` (enum-ish: `manual` | `from_real_entry` | `from_candidate`)
- `source_entry_id uuid null` (FK to `core.entries.id`)
- `source_candidate_id uuid null` (FK to `derived.candidates.id`)
- `created_at/updated_at/deleted_at`

Notes:
- Display name is not used as identity.

## `derived.simulated_entry_teams`
Fields:
- `id uuid primary key`
- `simulated_entry_id uuid not null` (FK)
- `team_id uuid not null` (FK to `core.teams.id`)
- `bid_points int not null`
- `created_at/updated_at/deleted_at`

Constraints:
- Unique `(simulated_entry_id, team_id)` where `deleted_at is null`.

---

# API surface (v1)
Routes are under `/api/*`.

## Create SimulatedCalcutta (from scratch)
`POST /api/simulated-calcuttas`

Request:
- `name: string`
- `description?: string`
- `tournamentId: string`
- `startingStateKey?: 'post_first_four' | 'current'`
- `excludedEntryName?: string`
- `payouts: Array<{ position: number, amountCents: number }>`
- `scoringRules: Array<{ winIndex: number, pointsAwarded: number }>`
- `metadata?: object`

Response:
- `{ id: string }`

Contract:
- Creates a fully-formed simulated calcutta with rules/payouts.
- Creates **no entries**.

## Create SimulatedCalcutta (instantiate from real calcutta)
`POST /api/simulated-calcuttas/from-calcutta`

Request:
- `calcuttaId: string`
- `name?: string`
- `description?: string`
- `startingStateKey?: 'post_first_four' | 'current'`
- `excludedEntryName?: string`
- `metadata?: object`

Response:
- `{ id: string, copiedEntries: number }`

Contract:
- Copies:
  - tournament_id (from the real calcutta)
  - payouts
  - scoring rules
  - entries + bids
- Sets `base_calcutta_id = calcuttaId`.
- Copied entries are normal `simulated_entries` and remain editable.

Implementation notes:
- Copying bids from real entries de-dupes duplicate `(entry_id, team_id)` rows by taking `MAX(bid_points)` per `team_id`.
- Copy implementation avoids nested queries while streaming `Rows` on the same connection (prevents pgx `conn busy`).

## Read/list simulated calcuttas
- `GET /api/simulated-calcuttas` (paginated)
  - Optional filter: `?tournament_id={uuid}`
- `GET /api/simulated-calcuttas/{id}`

Notes:
- `metadata` is required to be a JSON object (the server normalizes missing/empty to `{}`).

## Edit simulated calcutta
`PATCH /api/simulated-calcuttas/{id}`
- Update name/description/metadata
- Update `highlightedSimulatedEntryId`

## Replace rules/payouts
Two acceptable designs (pick one during implementation):
- **Option 1**: `PUT /api/simulated-calcuttas/{id}/rules` with full replace payload
- **Option 2**: granular CRUD endpoints for payouts/rules

v1 recommendation: **full replace** (simpler, fewer partial states).
 
Implemented (v1): **Option 1**.

## SimulatedEntries
- `GET /api/simulated-calcuttas/{id}/entries`
- `POST /api/simulated-calcuttas/{id}/entries` (manual create)
- `PATCH /api/simulated-calcuttas/{simulatedCalcuttaId}/entries/{entryId}` (rename / replace teams)
- `DELETE /api/simulated-calcuttas/{simulatedCalcuttaId}/entries/{entryId}`

Notes:
- Entry `teams` use snake_case fields: `{ team_id: string, bid_points: number }`.
- Entry payload fields use camelCase at the top-level (ex: `displayName`).

---

## Cohort simulations (evaluate-only)
Simulations are created under a Cohort and may target either a real Calcutta or a SimulatedCalcutta.

`POST /api/cohorts/{cohortId}/simulations`

Request (relevant fields):
- `calcuttaId?: string`
- `simulatedCalcuttaId?: string`
- `gameOutcomeRunId?: string` (optional; legacy/lab compatibility)
- `gameOutcomeSpec?: { kind: 'kenpom', sigma: number }` (preferred for sandbox)
- `nSims: number`
- `seed: number`
- `startingStateKey: 'post_first_four' | 'current'`

Contract:
- If `gameOutcomeSpec` is provided, simulations generate matchup probabilities on-the-fly using KenPom net ratings and the provided sigma.
- If `gameOutcomeSpec` is omitted, the system falls back to legacy behavior using `gameOutcomeRunId` (or latest run for the tournament).

### Import Candidate as SimulatedEntry
`POST /api/simulated-calcuttas/{id}/entries/import-candidate`

Request:
- `candidateId: string`
- `displayName?: string`

Response:
- `{ simulatedEntryId: string, nTeams: number }`

Contract:
- Creates a new simulated entry with `source_kind='from_candidate'` and copies the bid set.

---

# Invariants (non-negotiable)
- Read endpoints perform no inserts/updates (“no writes on GET”).
- A simulated calcutta is always created with complete rules/payouts.
- A simulated calcutta can be evaluated without any implicit “repair/create missing snapshot state” behavior.
- Entries are editable regardless of origin.

---

# Migration plan (avoid half-finished state)
This plan explicitly avoids long-lived dual-write states.

## Phase A0 — Add new schema
- Add `derived.simulated_*` tables and indexes.
- Add minimal adapter/service layer for creating and reading simulated calcuttas.

Acceptance:
- New schema exists.
- No existing behavior changes.

## Phase A1 — Add new endpoints + UI behind a feature flag
- Implement `POST /simulated-calcuttas` and `/from-calcutta`.
- Implement entry list/create/edit/delete.

Acceptance:
- A user can create a simulated calcutta from scratch and from a real calcutta.
- A user can add/edit/delete entries.

## Phase A2 — Simulation integration uses `simulated_calcuttas`
- Update new simulation creation flow to point at `simulated_calcutta_id`.
- Worker reads rules/payouts/entries from simulated tables.
- Sandbox simulations may omit `game_outcome_run_id` by providing a sandbox-owned `gameOutcomeSpec`.
- Persist the sandbox spec on `derived.simulation_runs.game_outcome_spec_json` and include it in the run job trigger payload.

Acceptance:
- Simulations can run against simulated calcuttas without referencing synthetic/snapshot tables.
- Simulations can run without a lab `game_outcome_run_id`.

## Phase A3 — Backfill legacy data (one-way)
Skipped.

Rationale:
- Sandbox simulations and scenarios are ephemeral; we do not need to preserve historical synthetic/snapshot scenarios.

Acceptance:
- No backfill is required.

## Phase A4 — Cutover + freeze legacy writes
- UI uses only simulated endpoints.
- Legacy synthetic Sandbox UI removed.
- Legacy synthetic/snapshot endpoints become read-only or return 410.
- Any remaining write paths to legacy synthetic/snapshot tables are blocked.

Acceptance:
- No new writes happen to synthetic/snapshot tables.

## Phase A5 — Drop old tables and code
- Phase A5 (minimal) drop:
  - `derived.synthetic_calcuttas`
  - `derived.synthetic_calcutta_candidates`
  - `derived.simulation_runs.synthetic_calcutta_id`
- Update the `derived.simulation_runs` run-job enqueue trigger to remove `synthetic_calcutta_id`.
- Remove legacy code paths.

Explicitly deferred:
- Dropping `core.calcutta_snapshots*` (handled elsewhere).

Acceptance:
- No production code references old tables.
- DB schema no longer contains legacy synthetic tables.

---

# Parallelization and prerequisites
This workstream is designed to be parallelizable.

- Can proceed in parallel with WS-B (Lab Candidates) because the only coupling is the optional import endpoint.
- WS-C (worker/run schema) can begin immediately via pure-function refactors, but the final cutover to simulated tables depends on Phase A2.

---

# Open questions
- Do we allow editing scoring rules/payouts after simulations have been run? If yes, do we version simulated calcuttas or treat simulations as referencing a specific immutable revision?
- Do we want sandbox cohorts as a separate feature (join table), or keep grouping external to the DB?

# Open tasks
- [ ] Confirm whether scoring rules/payouts are editable after evaluations exist; if yes, decide on revisioning/versioning approach.
- [x] Add `derived.simulated_*` schema (tables + indexes + constraints).
- [x] Implement `POST /api/simulated-calcuttas` (create from scratch: rules+payouts only).
- [x] Implement `POST /api/simulated-calcuttas/from-calcutta` (instantiate from real calcutta).
- [x] Implement read endpoints:
  - [x] `GET /api/simulated-calcuttas`
  - [x] `GET /api/simulated-calcuttas/{id}`
- [x] Implement simulated calcutta updates:
  - [x] `PATCH /api/simulated-calcuttas/{id}` (name/description/metadata/highlight)
  - [x] Rules/payouts update endpoint (full replace: `PUT /api/simulated-calcuttas/{id}/rules`)
- [x] Implement SimulatedEntry endpoints:
  - [x] `GET /api/simulated-calcuttas/{id}/entries`
  - [x] `POST /api/simulated-calcuttas/{id}/entries` (manual create)
  - [x] `PATCH /api/simulated-calcuttas/{simulatedCalcuttaId}/entries/{entryId}` (rename/replace teams)
  - [x] `DELETE /api/simulated-calcuttas/{simulatedCalcuttaId}/entries/{entryId}`
- [x] Implement Candidate import endpoint:
  - [x] `POST /api/simulated-calcuttas/{id}/entries/import-candidate`
- [ ] Add invariant test/assertion: SimulatedCalcutta read endpoints must not write (no implicit repair/create).
- [x] Update simulation creation + worker read paths to use `derived.simulated_*` as the source of truth (Phase A2).
- [x] Allow sandbox simulations to omit `game_outcome_run_id` by using sandbox-owned `gameOutcomeSpec`.
- [x] Phase A4: Cut over UI to simulated endpoints only.
- [x] Phase A4: Make legacy synthetic/snapshot endpoints read-only or return 410.
- [x] Phase A4: Remove/disable any remaining writes to legacy synthetic/snapshot tables.
- [x] Phase A5 (minimal): Drop legacy synthetic tables and remove `synthetic_calcutta_id`.
- [x] Phase A5 (minimal): Remove legacy endpoints and any production code references.
