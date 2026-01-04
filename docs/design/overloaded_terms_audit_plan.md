# Overloaded Terms Audit + Rename Plan

## Goal
Reduce ambiguity in **SQL columns** and **core data models** for overloaded terms like `roi`, `returns`, `investment`, `expected_value`, and `market`.

We can still present these as “ROI” / “Returns” in the UI, but the underlying fields should encode:
- **Units** (`_points`, `_cents`)
- **Denominator** (e.g. `..._vs_market`, `..._vs_bid`, `..._vs_market_plus_bid`)
- **Semantics** (expected vs realized, absolute vs normalized)

This doc inventories current usages and proposes a migration checklist.

## Naming rubric (proposed)

### Units
- Use `_points` for **in-game currency / auction bids / scoring**.
- Use `_cents` for **real money**.
- Use `_share` or `_fraction` for unitless proportions.

### Ratios
Avoid bare `roi`.

Prefer one of:
- `return_multiple` (dimensionless) = `value_received / cost`
- `return_ratio` (dimensionless) = `value_received / cost`

When the denominator is domain-specific, encode it:
- `return_multiple_vs_bid_points`
- `return_multiple_vs_market_points`
- `return_multiple_vs_market_plus_bid_points`

### Expected vs realized
- Prefix expected values with `expected_...`.
- Prefix realized/historical values with `actual_...`.

### “Normalized”
Only use `normalized_*` when the normalization scheme is explicit in code/docs.
If it’s “relative to baseline”, encode that:
- `performance_multiple_vs_baseline`


## Inventory: overloaded fields and current meanings

### A) Backend DB + SQLC queries (authoritative)

#### 1) “Analytics” (historical auction performance)
Source: `backend/internal/adapters/db/sqlc/queries/analytics.sql`

- `total_investment` (SQL alias)
  - Meaning: **sum of `bid_points`** allocated to a group (seed/region/team)
  - Units: points
  - Proposed rename (SQL alias + Go struct field): `total_bid_points`

- `baselineROI` (computed in `backend/internal/app/analytics/service.go`)
  - Current formula: `totalPoints / totalInvestment`
  - Meaning: **baseline points-per-bid-point** across the whole population
  - Proposed rename: `baseline_points_per_bid_point` (presentation can still label “Baseline ROI”)

- `ROI` in `SeedAnalyticsResult` / `RegionAnalyticsResult` / `TeamAnalyticsResult`
  - Current formula: `(group_points / group_bid_points) / (baseline_points / baseline_bid_points)`
  - Meaning: **relative performance multiple vs baseline**
  - Proposed rename in models: `performance_multiple_vs_baseline`
  - Keep JSON name `roi` if desired for UI continuity (map from new field).

#### 2) “Hall of fame” / best investments
Source: `backend/internal/adapters/db/sqlc/queries/analytics_hof.sql`

- `investment` (alias; derived from `bid_points`)
  - Units: points
  - Proposed rename: `bid_points`

- `raw_returns`
  - Current formula: `team_points * ownership_fraction`
  - Units: points
  - Proposed rename: `return_points`

- `normalized_returns`
  - Current formula rescales to a standard pool size:
    `raw_returns * (100 * total_participants) / calcutta_total_points`
  - Units: “points scaled to a standard pool” (still points-like)
  - Proposed rename: `return_points_normalized_to_standard_pool`

- `total_returns` in entry leaderboard
  - Current meaning: entry total points
  - Proposed rename: `total_points` (or `total_return_points`)

- `average_returns`
  - Current meaning: average points per entry
  - Proposed rename: `mean_points` (or `mean_return_points`)

- `raw_roi` / `normalized_roi` in best investments
  - `raw_roi = team_points / total_bid`
    - Meaning: **points per bid point** (realized)
    - Proposed rename: `points_per_bid_point`
  - `normalized_roi = raw_roi / (calcutta_total_points / calcutta_total_bid)`
    - Meaning: **relative performance multiple vs baseline** (same semantic as analytics ROI)
    - Proposed rename: `performance_multiple_vs_baseline`

#### 3) “Simulated entry” endpoint
Source: `backend/internal/transport/httpserver/handlers_simulated_entry.go`

- `ExpectedROI` JSON field (`expected_points / expected_market`)
  - Here ROI means: **points per market point**
- `OurROI` JSON field (`expected_points / (expected_market + our_bid)`)
  - Here ROI means: **points per (market+our_bid)**

These are **different denominators** than the historical ROI fields and also different from `derived.recommended_entry_bids.expected_roi` (see below).

Proposed rename:
- `expected_return_multiple_vs_market_points`
- `expected_return_multiple_vs_market_plus_bid_points`

Keep UI labels “Expected ROI” / “Our ROI” if desired.

#### 4) Derived artifacts: strategy generation / bids / detailed report
Schema lineage:
- Initially created in `backend/migrations/schema/20251231050000_create_uuid_analytics_schema.up.sql` as `gold_*`.
- Later moved to `gold.*`, then to `lab_gold.*`, then collapsed to `derived.*`.

Key columns currently in play:

- `derived.recommended_entry_bids.expected_roi`
  - Observed producer: `backend/internal/app/recommended_entry_bids/service.go`
  - Current formula in Go write path: `expected_points / bid_points`
  - Meaning: **expected points per bid point** (ignores market and ownership)
  - Problem: conflicts with other “expected_roi” usages.
  - Proposed rename: `expected_points_per_bid_point`

- `derived.detailed_investment_report.expected_roi` + `our_roi`
  - Created in `20251231050000_create_uuid_analytics_schema.up.sql` with:
    - `expected_points`
    - `predicted_market_points`
    - `expected_roi`
    - `our_roi`
  - Intended meaning (per Python report + HTTP simulated entry):
    - `expected_roi ≈ expected_points / predicted_market_points`
    - `our_roi ≈ expected_points / (predicted_market_points + our_bid_points)`
  - Proposed rename:
    - `expected_return_multiple_vs_market_points`
    - `expected_return_multiple_vs_market_plus_bid_points`

#### 5) Predicted returns endpoint
Source: `backend/internal/adapters/db/sqlc/queries/analytics_calcutta_predictions.sql`

- `expected_value`
  - Current meaning: **expected points** from simulation (`calcutta_points_for_progress`)
  - Proposed rename: `expected_points`

- `expected_market`
  - Current meaning: **predicted market bid points** for that team
  - Proposed rename: `predicted_market_points` (or `predicted_market_bid_points`)

- `rational` / `predicted` / `delta` in predicted investment
  - `rational`: “fair” market points proportional to expected points
  - `predicted`: predicted market points from model
  - `delta`: percent over/under
  - Proposed renames:
    - `fair_market_points`
    - `predicted_market_points`
    - `market_overpricing_pct`


### B) Go ports / DTOs / JSON contracts

#### 1) `backend/internal/ports/analytics.go`
- `TotalInvestment` → `TotalBidPoints`
- `Investment` → `BidPoints`
- `RawReturns` → `ReturnPoints`
- `NormalizedReturns` → `ReturnPointsNormalizedToStandardPool`
- `TotalReturns` / `AverageReturns` → `TotalPoints` / `MeanPoints`
- `RawROI` → `PointsPerBidPoint`
- `NormalizedROI` → `PerformanceMultipleVsBaseline`

#### 2) `backend/internal/ports/ml_analytics.go`
- `OurEntryBid.ExpectedROI`
  - Semantics depend on DB field `derived.recommended_entry_bids.expected_roi`.
  - Proposed rename: `ExpectedPointsPerBidPoint`.

#### 3) HTTP handlers that currently compute ROI
- `handlers_simulated_entry.go`: rename JSON keys (or keep keys and rename internal fields)


### C) Frontend types/services

#### 1) `frontend/src/types/analytics.ts`
- `roi` fields (seed/region/team analytics) are **performance multiple vs baseline**, not a raw ratio.
  - Proposed internal rename: `performanceMultipleVsBaseline` (keep `roi` in API if desired).

- `BestInvestment.rawROI`/`normalizedROI`
  - Same rename as backend.

#### 2) `frontend/src/types/hallOfFame.ts`
- `investment` should become `bidPoints` (units)
- `rawReturns` / `normalizedReturns` should become `returnPoints` / `returnPointsNormalizedToStandardPool`
- `totalReturns` / `averageReturns` should become `totalPoints` / `meanPoints`

#### 3) `frontend/src/services/mlAnalyticsService.ts`
- `expected_roi` in `OurEntryDetailsResponse.portfolio[]`
  - Proposed: `expected_points_per_bid_point`

- `expected_value` in predicted returns response
  - Proposed: `expected_points`


### D) Data-science (Moneyball Python)

#### 1) `data-science/moneyball/models/detailed_investment_report.py`
- Uses:
  - `expected_roi = expected_points / expected_market`
  - `our_roi = expected_points / (expected_market + our_bid)`
  - `actual_roi = expected_points / actual_market`

Proposed renames:
- `expected_return_multiple_vs_market_points`
- `expected_return_multiple_vs_market_plus_bid_points`
- `actual_return_multiple_vs_market_points`

Also rename:
- `expected_market` → `predicted_market_points`
- `our_bid` → `recommended_bid_points`

#### 2) `data-science/moneyball/models/portfolio_construction.py`
- debug artifact includes `roi = expected_return_points / bid_amount_points`
  - Proposed rename: `expected_points_per_bid_point` or `expected_return_multiple_vs_bid_points`


## High-risk / compatibility concerns
- Some tables previously had FKs that prevented turning them into views. (We’ve already had to drop/migrate some FKs historically.)
- Frontend contracts currently assume `roi`, `rawROI`, `normalizedROI`, `expected_value`, `expected_roi`, etc.
- Strategy-generation artifacts are treated as regenerable; still, API consumers will break on rename unless we provide compat fields.


## Execution plan (checklist)

### Phase 0: Agree on conventions
- [ ] Decide canonical ratio term: `return_multiple` vs `points_per_bid_point` for points/bid ratios.
- [ ] Decide which API fields stay as legacy presentation names (`roi`, `rawROI`, etc.) vs fully renamed.

### Phase 1: SQL schema + compat views
- [ ] Add new columns (or generated columns) to `derived.recommended_entry_bids`:
  - [ ] `expected_points_per_bid_point` (backfill from `expected_roi`)
  - [ ] Keep `expected_roi` temporarily (compat).
- [ ] Add new columns to `derived.detailed_investment_report`:
  - [ ] `expected_return_multiple_vs_market_points` (from `expected_roi`)
  - [ ] `expected_return_multiple_vs_market_plus_bid_points` (from `our_roi`)
  - [ ] Keep `expected_roi` / `our_roi` temporarily.
- [ ] For historical analytics queries, standardize aliases:
  - [ ] Replace `total_investment` alias with `total_bid_points` in `analytics.sql` (and update sqlc + Go mapping).
- [ ] For HOF queries, standardize alias names:
  - [ ] `investment` → `bid_points`
  - [ ] `raw_returns` → `return_points`
  - [ ] `normalized_returns` → `return_points_normalized_to_standard_pool`
  - [ ] `raw_roi` → `points_per_bid_point`
  - [ ] `normalized_roi` → `performance_multiple_vs_baseline`
- [ ] For predicted returns/investment queries:
  - [ ] `expected_value` → `expected_points`
  - [ ] `expected_market` → `predicted_market_points`
  - [ ] `rational` → `fair_market_points`
  - [ ] `predicted` → `predicted_market_points`
  - [ ] `delta` → `market_overpricing_pct`

### Phase 2: Go backend models + JSON
- [ ] Update `ports/analytics.go` field names to match new aliases/semantics.
- [ ] Update `analytics/service.go` to compute:
  - [ ] `baseline_points_per_bid_point`
  - [ ] `performance_multiple_vs_baseline`
- [ ] Update `ports/ml_analytics.go`:
  - [ ] `ExpectedROI` → `ExpectedPointsPerBidPoint`
- [ ] Decide whether to keep HTTP JSON keys as legacy (`roi`, `expected_roi`, etc.):
  - [ ] If keeping legacy keys, create mapping structs so internal names are precise.
  - [ ] If renaming JSON keys, version the API or provide parallel fields during migration.

### Phase 3: Frontend types + usage
- [ ] Update TS types:
  - [ ] `analytics.ts` (seed/region/team analytics) to use `performanceMultipleVsBaseline` (or keep `roi` if API unchanged).
  - [ ] `hallOfFame.ts` rename investment/returns/ROI fields.
- [ ] Update frontend service response types in `mlAnalyticsService.ts`:
  - [ ] `expected_roi` → `expected_points_per_bid_point`
  - [ ] `expected_value` → `expected_points`

### Phase 4: Data-science artifacts (if still consumed)
- [ ] Rename columns in Python-generated artifacts:
  - [ ] `detailed_investment_report.py` rename ROI columns to `*_return_multiple_*` names.
  - [ ] Update downstream readers/writers/tests accordingly.

### Phase 5: Deprecation cleanup
- [ ] After all consumers migrate, drop/stop emitting:
  - [ ] `expected_roi` (ambiguous)
  - [ ] `our_roi` (ambiguous)
  - [ ] `raw_roi`/`normalized_roi` (replace with explicit names)
  - [ ] `expected_value` (replace with expected_points)


## Notes / open questions
- Do we want **one** canonical name for “points per bid point” across the whole codebase (`points_per_bid_point`), even when it’s “expected”?
- For simulated-entry vs detailed-investment-report, do we want to keep market-based ratios as `..._vs_market_points` even though the numerator is expected points (not payouts)?
