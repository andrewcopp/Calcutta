# Schema Migration Plan (Core + Medallion Tiers)

## Purpose
This document is an executable checklist for migrating the database to:

- `core` schema for the playable product
- `bronze` / `silver` / `gold` schemas for the lab (medallion tiers)
- Explicit tournament identity (`seasons`, `competitions`, `tournaments`) and removal of brittle string joins
- Configurable, centralized scoring via `core.calcutta_scoring_rules` (incremental-per-win)

This plan is designed to be safe with respect to historical cleaned data. When in doubt, preserve and backfill first, prune later.

---

## Locked decisions
- **Schemas**: use `core`, `bronze`, `silver`, `gold`
- **SQL style**: always schema-qualify (`core.tournaments`, `gold.optimization_runs`, etc.)
- **Core naming**: short, plural nouns (`core.entries`, `core.entry_teams`, `core.teams`)
- **Identity**: lab tiers reference `core.tournaments.id`
- **Audit columns**: mandatory `created_at`, `updated_at`, nullable `deleted_at` everywhere + `updated_at` triggers
- **Portfolios**: treated as derived/recomputable; do not preserve/backfill

---

## Safety rails (must do)
- [x] **DB snapshot** taken (before any migration/backfill work)
- [x] **Freeze invariants** (row counts + relational invariants) for protected data tables
- [x] **Inventory** current tables + mark each as: protected / derived / deprecated / unknown

### Inventory (repo-derived; verified against live DB 2026-01-02)

#### Protected (must preserve + verify)

##### Public schema (auth + ops)
- `public.users`
- `public.auth_sessions`
- `public.permissions`
- `public.labels`
- `public.label_permissions`
- `public.grants`
- `public.api_keys`
- `public.bundle_uploads`
- `public.schema_migrations`

##### Core schema (product)
- `core.seasons`
- `core.competitions`
- `core.tournaments`
- `core.schools`
- `core.teams`
- `core.team_kenpom_stats`
- `core.calcuttas`
- `core.calcutta_scoring_rules`
- `core.entries`
- `core.entry_teams`
- `core.payouts`

#### Derived / recomputable (safe to rebuild; keep for performance)

##### Core derived views
- `core.derived_portfolios`
- `core.derived_portfolio_teams`

##### Lab schemas (medallion tiers)
- `bronze.tournaments`
- `bronze.teams`
- `bronze.calcuttas`
- `bronze.entry_bids`
- `bronze.payouts`
- `silver.predicted_game_outcomes`
- `silver.simulated_tournaments`
- `silver.predicted_market_share`
- `gold.optimization_runs`
- `gold.recommended_entry_bids`
- `gold.entry_simulation_outcomes`
- `gold.entry_performance`
- `gold.detailed_investment_report`

#### Deprecated / transitional (remove after verification window)

##### Public compatibility / core-context views
- (dropped)

#### Unknown / needs decision
- (none listed yet)

### Orphan checks (baseline)
- [x] `core.teams.tournament_id -> core.tournaments.id`
- [x] `core.team_kenpom_stats.team_id -> core.teams.id`
- [x] `tournament_games` checks: derived (not stored in DB)
- [x] `core.calcuttas.tournament_id -> core.tournaments.id`
- [x] `core.calcuttas.owner_id -> public.users.id`
- [x] `core.entries.calcutta_id -> core.calcuttas.id`
- [x] `core.entries.user_id -> public.users.id` (nullable)
- [x] `core.entry_teams.entry_id -> core.entries.id`
- [x] `core.entry_teams.team_id -> core.teams.id`
- [x] `core.payouts.calcutta_id -> core.calcuttas.id`
- [x] `core.calcutta_scoring_rules.calcutta_id -> core.calcuttas.id`

---

## Protected data (must preserve + verify)
This list should be kept short and explicit.

### Core gameplay data
- [x] Tournament identity and structure data (existing)
- [x] Calcutta definitions
- [x] Entries and their bids
- [x] Payout structures

### KenPom cleaned data
- [x] `core.team_kenpom_stats`

### Scoring rules
- [x] `core.calcutta_scoring_rules`

### Explicitly NOT protected
- [ ] `calcutta_portfolios`
- [ ] `calcutta_portfolio_teams`

---

## Target schemas
### 1) Create schemas
- [x] Create schema `core`
- [x] Create schema `bronze`
- [x] Create schema `silver`
- [x] Create schema `gold`

### 2) Create DB roles + permissions
- [x] Run ops script: `backend/ops/db/admin/db_roles_grants.sql`
- [x] Verify roles/grants: `backend/ops/db/admin/verify_db_roles_grants.sql` (output: `verify_db_roles_grants_output.txt`)
- [x] Role: `app_writer`
  - [x] Read/write on `core.*`
  - [x] Read-only on `bronze/silver/gold.*` (optional)
- [x] Role: `lab_pipeline`
  - [x] Read-only on `core.*`
  - [x] Read/write on `bronze/silver/gold.*`

### 3) Standard triggers
- [x] Create a standard `updated_at` trigger function
- [x] Add `updated_at` triggers for all new tables

### Current trigger strategy
- **Core**: `core.set_updated_at()` on `core.*`
- **Lab tiers + public ops tables**: `public.set_updated_at()` on `bronze/silver/gold/public.*`

---

## Core identity model
### Tables
- [x] `core.seasons` (e.g. `year`)
- [x] `core.competitions` (e.g. "NCAA Men's")
- [x] `core.tournaments` (`competition_id`, `season_id`, etc.)

### Backfill
- [x] Backfill `core.seasons`
- [x] Backfill `core.competitions`
- [x] Backfill `core.tournaments`

### Verification
- [x] Every old tournament maps to exactly one `core.tournaments.id`
- [x] No string parsing joins remain in application queries

---

## Core gameplay tables
### Tables (short names)
- [x] `core.schools`
- [x] `core.teams`
- [x] `core.calcuttas`
- [x] `core.entries`
- [x] `core.entry_teams`
- [x] `core.payouts`

### Backfill order
- [x] `core.schools`
- [x] `core.teams`
- [x] `core.calcuttas`
- [x] `core.entries`
- [x] `core.entry_teams`
- [x] `core.payouts`

### Verification
- [x] Row counts match expected
- [x] FKs valid (no orphaned entries, teams, entry_teams)

---

## Scoring migration (incremental-per-win)
### New rule table
- [x] Create `core.calcutta_scoring_rules`:
  - `calcutta_id`
  - `win_index`
  - `points_awarded`
  - audit columns

### Backfill
- [x] Backfill from `calcutta_rounds` ordered by `round ASC`:
  - map to `win_index` sequentially

### Central scoring function/view
- [x] Implement a canonical SQL function or view to compute points:
  - input: `calcutta_id`, `wins`
  - output: total points = sum(points_awarded for win_index <= wins)

### Current canonical scoring API
- [x] `core.calcutta_points_for_progress(calcutta_id, wins, byes)`

### Replace cumulative scoring assumptions
- [x] Update backend SQL that hardcodes `50/150/300/500/750/1050` to use scoring rules
- [x] Update analytics SQL that uses `(wins + byes)` mapping
- [x] Update Python readers/logic that computes points from hardcoded thresholds

---

## KenPom stats migration
### Target table
- [x] Create `core.team_kenpom_stats`:
  - `team_id` (FK to `core.teams`)
  - `net_rtg`, `o_rtg`, `d_rtg`, `adj_t`
  - audit columns

### Backfill
- [x] Backfill from `tournament_team_kenpom_stats` into `core.team_kenpom_stats`

### Verification
- [x] Per-team stats row counts match (excluding soft-deleted)
- [x] Spot-check known teams across historical tournaments

---

## Lab (medallion) schemas
### Design principles
- **No redundant tournament identity tables** (e.g. remove `bronze_tournaments` in end state)
- All tier tables reference `core.tournaments.id` where applicable
- Names are semantic (not mirrored copies of core)

### Implementation
- [x] Create `bronze.*` tables for ingestion snapshots/raw-ish facts as needed
- [x] Create `silver.*` tables for cleaned/normalized outputs (predictions, simulations)
- [x] Create `gold.*` tables for evaluation/summary/recommendations (runs, performance)

---

## API contract changes
### Tournament vs calcutta scoping
- [x] Identify endpoints that must accept `calcutta_id` (scoring-dependent)
- [ ] Keep tournament-based endpoints only where scoring is irrelevant

### Notes
A calcutta always has a tournament, so any endpoint can join tournament data as needed.

---

## Cutover plan
### Backend
- [x] Update sqlc query files to schema-qualify and target new tables
- [x] Regenerate sqlc
- [ ] Refactor endpoints to consistent layering:
  - handlers -> service -> sqlc
- [ ] Remove direct SQL in handlers (where practical)
  - [x] Calcutta analytics endpoints (predicted investment/returns/simulated entry) refactored to handlers -> service -> sqlc
  - [x] ML analytics endpoints (simulated calcuttas, team performance-by-calcutta) refactored to handlers -> service -> sqlc
  - [x] Bundle upload admin endpoints + bundle import worker refactored to sqlc
  - [x] Tournament sim-stats-by-core-tournament-id endpoint refactored to MLAnalytics service -> sqlc
  - [ ] Continue inventory + refactor remaining handler inline SQL

### Airflow
- [ ] Use `lab_pipeline` role
- [ ] Read from `core.*`
- [ ] Write to `bronze/silver/gold.*`
- [ ] Ensure each task is idempotent and keyed by explicit run IDs

---

## Verification checklist (post-cutover)
- [x] App can list tournaments/teams/calcuttas/entries
- [x] Entry bids and payouts match historical truth
- [x] Scoring results match expected (incremental-per-win)
- [x] KenPom stats visible and correct
- [x] Analytics endpoints no longer rely on brittle joins

---

## Deprecation + cleanup
- [x] Move old tables into `legacy` schema (temporary)
  - Migration: `20260101123000_move_legacy_tables_to_legacy_schema`
  - Verified:
    - bundle exporter maintains data
    - frontend still loads correctly
    - `backend/ops/db/audits/core_sanity_checks.sql` passes

### Drop public bronze core-context views (after compat window)
- [x] Migration: `20260102095000_drop_public_bronze_core_ctx_views`
- [x] Verified: runtime `sqlc` no longer depends on `public.bronze_*_core_ctx`

### Drop unused public legacy functions
- [x] Migration: `20260102102200_drop_unused_public_functions`
- [x] Dropped:
  - `public.get_entry_portfolio`
  - `public.set_schools_slug`
  - `public.set_tournaments_import_key`

### Audit columns everywhere
- [x] Migration: `20260102103000_add_audit_columns_to_lab_and_public_tables`
- [x] Added `updated_at` + nullable `deleted_at` to all `bronze/silver/gold` tables and remaining public tables missing them
- [x] Ensured `set_updated_at` triggers exist for tables with `updated_at` in `bronze/silver/gold/public`

### Core owns core triggers
- [x] Migration: `20260102104500_standardize_core_updated_at_triggers`
- [x] Ensured `core.*` uses `core.set_updated_at()` (removed legacy core triggers that called `public.set_updated_at()`)

### Drop legacy schema (after verification window)
- [x] **Go/no-go checklist**:
  - [x] Bundle tooling reads/writes `core.*` only (no `legacy.*` dependencies)
  - [x] Analytics snapshot exporter reads `core.*` only (no `legacy.*` dependencies)
  - [x] Repo-wide search shows no runtime/tooling `legacy.*` usage
  - [x] `backend/ops/db/audits/core_sanity_checks.sql` passes (core-only)
  - [x] `backend/ops/db/maintenance/freeze_invariants.sql` passes (core-only)
  - [x] DB snapshot taken
- [x] Migration: `20260101128000_drop_legacy_schema`
- [ ] Remove unused tables after verification period
- [ ] Remove dead sqlc queries
- [ ] Audit/clean migrations once new baseline is stable

---

## Progress log
Add dated notes here as work is completed.

- [2026-01-01] Created DB snapshot: `db_snapshot_pre_schema_migration.dump`
- [2026-01-01] Captured baseline invariants: `invariants_pre_schema_migration.txt`
  - tournaments=9
  - tournament_teams=612
  - tournament_team_kenpom_stats=544
  - calcuttas=8
  - calcutta_entries=334
  - calcutta_entry_teams=2911

- [2026-01-01] Applied Phase 1/2 migrations (create schemas + core tables)
- [2026-01-01] Captured post Phase 1/2 invariants: `invariants_post_phase1_phase2.txt`

- [2026-01-01] Applied Phase 3 migration (backfill public -> core): `20260101102000_backfill_core_from_public`
- [2026-01-01] Captured post Phase 3 invariants: `invariants_post_phase3_backfill.txt`
- [2026-01-01] Captured post Phase 3 core sanity report: `core_sanity_post_phase3.txt`

- [2026-01-01] Applied roles/grants ops script: `backend/ops/db/admin/db_roles_grants.sql`
- [2026-01-01] Verified roles/grants: `verify_db_roles_grants_output.txt`
- [2026-01-01] Cutover completed: app + tools read/write core gameplay data in `core.*`

- [2026-01-01] Scoring rules cutover: app + tools read canonical scoring rules from `core.calcutta_scoring_rules`
  - [2026-01-01] Added canonical scoring function: `core.calcutta_points_for_progress(calcutta_id, wins, byes)`
  - [2026-01-01] Analytics SQL updated to use scoring function (removed hardcoded `50/150/300/500/750/1050`)

- [2026-01-01] Analytics query cutover: moved remaining `sqlc` analytics queries to use `core.*` tables for tournaments/calcuttas/entries/bids/payouts

- [2026-01-01] Portfolios fully derived: added canonical views `core.derived_portfolios` and `core.derived_portfolio_teams` (ownership + points derived from `core.entry_teams` + `core.calcutta_points_for_progress`)
- [2026-01-01] Portfolio read endpoints now use derived views via `sqlc` queries; portfolio mutation/recalc endpoints disabled and all recalculation side effects removed
   - Note: legacy tables `calcutta_portfolios` and `calcutta_portfolio_teams` are now unused by the app and can be deprecated/dropped after a cleanup window

 - [2026-01-01] Lab tiers cutover: moved analytics tables into schema-qualified `bronze.*` / `silver.*` / `gold.*`, updated app + pipeline queries to schema-qualified references, and dropped temporary public compatibility views

- [2026-01-01] Bundle tooling refactored to core-only (exporter/importer/verifier)
- [2026-01-01] Analytics snapshot exporter refactored to core-only
- [2026-01-01] Dropped legacy schema: `20260101128000_drop_legacy_schema`
- [2026-01-01] Ops verification scripts updated to core-only: `backend/ops/db/audits/core_sanity_checks.sql`, `backend/ops/db/maintenance/freeze_invariants.sql`

- [2026-01-01] Standard `updated_at` trigger function + triggers added for core + medallion tables
- [2026-01-01] Core identity verification: core/bronze ID linkage validated; removed remaining join-relevant string parsing from runtime analytics queries

- [2026-01-02] Backend cleanup: removed inline SQL from additional endpoints by refactoring to handlers -> service/repo -> sqlc
  - Calcutta analytics: predicted investment/returns/simulated entry
  - ML analytics: simulated calcuttas, team performance-by-calcutta
  - Admin bundles + bundle import worker: bundle upload lifecycle via sqlc queries
  - Tournament analytics: sim stats by `core.tournaments.id` via MLAnalytics/sqlc

- [2026-01-02] Dropped deprecated compatibility views `public.bronze_*_core_ctx` via migration: `20260102095000_drop_public_bronze_core_ctx_views`

- [2026-01-02] Dropped unused public legacy functions via migration: `20260102102200_drop_unused_public_functions`
- [2026-01-02] Added audit columns + updated_at triggers for lab tiers and remaining public tables via migration: `20260102103000_add_audit_columns_to_lab_and_public_tables`
- [2026-01-02] Standardized core updated_at triggers to `core.set_updated_at()` via migration: `20260102104500_standardize_core_updated_at_triggers`
