# finish_analytics_pipeline

This checklist tracks what remains to fully finish the analytics + ML pipeline refactor and migration to lineage-native IDs.

Legend:
- [x] done
- [ ] remaining

## Schema + namespace refactor (core/lab/analytics)
- [x] Move legacy tables to schema-qualified layout (`core`, `lab_bronze`, `lab_silver`, `lab_gold`, `analytics`).
- [x] Restore compat medallion schemas/views temporarily to unblock lingering legacy references.
- [ ] Audit and eliminate any remaining unqualified schema references in SQL (no implicit `public`).
- [ ] Decide whether compat medallion views stay (temporary) or get removed (and enforce via CI).
- [ ] Document final schema conventions (where configs live, where derived outputs live, what is authoritative).

## Deprecate `lab_gold.optimization_runs` (migrate reads + remove dependency)
- [x] Migrate ML analytics endpoints to resolve `run_id` via `lab_gold.strategy_generation_runs.run_key`.
- [x] Remove remaining read-path references to `lab_gold.optimization_runs` in SQLC queries and repositories.
- [ ] Remove (or lock down) `lab_gold.optimization_runs` table once no longer needed by pipeline writes.
- [ ] Remove any remaining legacy drift / fallback code paths that depend on string `run_id` heuristics.

## Calcutta config: in-game budget points
- [x] Add `core.calcuttas.budget_points` (default 100).
- [x] Update queries to use `core.calcuttas.budget_points` instead of hardcoded `100`.
- [ ] Backfill/verify `budget_points` values for existing calcuttas as desired.

## Simulation run lineage (authoritative write path)
- [x] Add lineage tables:
  - `analytics.tournament_state_snapshots`
  - `analytics.tournament_simulation_batches`
  - `analytics.calcutta_evaluation_runs`
- [ ] Update pipeline to write `tournament_state_snapshots`.
- [ ] Update pipeline to write `tournament_simulation_batches`.
- [ ] Update pipeline to write `analytics.simulated_tournaments.tournament_simulation_batch_id` (make non-null once stable).
- [ ] Ensure pipeline writes `entry_simulation_outcomes.calcutta_evaluation_run_id`.
- [ ] Ensure pipeline writes `entry_performance.calcutta_evaluation_run_id`.
- [ ] Backfill lineage IDs for existing cached rows where possible.
- [ ] Tighten constraints (after backfill): make key lineage FKs NOT NULL for new data.

## Lineage-native read APIs (stop year-string fallbacks)
- [x] Add lineage-native selection for predicted investment / predicted returns (accept optional `strategy_generation_run_id`).
- [x] Add run lineage listing endpoints for selection UI.
- [x] Migrate tournament run listing + run endpoints (`our-entry`, `portfolio`, `rankings`, `simulations`) to strategy-run resolution.
- [ ] Update remaining analytics endpoints to accept lineage IDs (`tournament_simulation_batch_id`, `calcutta_evaluation_run_id`, `strategy_generation_run_id`).
- [ ] Replace remaining "latest" selection logic to use lineage tables (not run_id/year heuristics).

## Calcutta snapshots + Hall of Fame
- [x] Add `core.calcutta_snapshots` tables + nullable FK from `analytics.calcutta_evaluation_runs.calcutta_snapshot_id`.
- [ ] Update pipeline to write a calcutta snapshot for each evaluation context.
- [ ] Make `analytics.calcutta_evaluation_runs.calcutta_snapshot_id` NOT NULL once pipeline is writing it.
- [ ] Define realized outcomes snapshotting (when tournament completes).
- [ ] Implement hall-of-fame metrics models/tables/views.
- [ ] Add API endpoints for hall-of-fame views (per year, per calcutta, global).

## Strategy generation persistence (UUID-first)
- [ ] Make derived strategy outputs keyed by `strategy_generation_run_id` (UUID), not legacy string `run_id`.
- [ ] Key evaluation outputs by `calcutta_evaluation_run_id` wherever possible.
- [ ] Reduce/contain usage of legacy run string (`run_key`/`run_id`) to compatibility only.

## Strategies BYOM (bring your own model)
- [ ] Decide BYOM strategy submission format (params JSON vs constrained DSL vs code artifact).
- [ ] Add storage tables for user-submitted strategies (ownership, versioning, permissions).
- [ ] Integrate BYOM execution into lineage:
  - submission -> strategy_generation_run
  - evaluation -> calcutta_evaluation_run
  - outputs keyed by UUID lineage
- [ ] Add BYOM APIs (create/list/validate/run).
- [ ] Add UI flows to select BYOM strategy and view results.

## CI + guardrails
- [ ] Add CI guardrail: run `make sqlc-generate` and fail if `git diff` is non-empty.
- [ ] Add lint/format guardrails decision for data-science scripts (flake8 issues).

## Operational tooling
- [x] Ops-only reset command for ephemeral derived/cached tables.
- [ ] Add a repeatable "rebuild derived analytics" job for a given lineage run (batch/evaluation/strategy), if needed.
