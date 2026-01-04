# Core vs Lab Architecture

## Update
- Completed: consolidated runtime schemas to `core` + `derived` via migration `20260104211500_collapse_lab_schemas_into_derived`.

## Goals
- Separate **product concerns** (Core) from **experimentation / ML / sandbox** (Lab).
- Keep a clean path to the long-term vision: **two separate databases**.
- Avoid mixing two axes:
  - **Product boundary**: core vs lab
  - **Medallion tiers**: bronze/silver/gold

## Terminology
- **Core**: the product.
  - Canonical business data (entries, payouts, scoring rules, etc.)
  - Product analytics and derived materializations that the product serves to users (leaderboards, simulation outputs, hall-of-fame stats)
- **Lab (aka Sandbox)**: experimentation surface.
  - ML training, backtests, offline analysis, strategies, ad-hoc experiments
- **Medallion tiers**: bronze/silver/gold are *pipeline tiers*, not product boundaries.

## Long-term target (two DBs)
- **Core DB**
  - Schemas:
    - `core`: canonical business tables
    - `analytics`: product-derived materializations (simulation + evaluation + hall-of-fame)
- **Lab DB**
  - Schemas:
    - `bronze`, `silver`, `gold`: medallion layers for pipelines and experiments
  - Optionally `core_mirror` (read-only replication / periodic snapshot of the minimal core data required for training)

## Short-term target (single DB, but clean boundary)
We keep one physical database for now, but we make the boundary explicit with schemas.

### Proposed schema layout (single DB)
- `core`
  - canonical tables
- `analytics`
  - product-facing derived tables (not “lab”) used by the core app
- `lab_bronze`, `lab_silver`, `lab_gold`
  - lab-only medallion layers

Rationale:
- Frees up `bronze/silver/gold` to be used *unprefixed* in the eventual Lab DB.
- Keeps product-facing derived tables out of `core` while still “owned by core”.
- Makes it obvious which objects are safe to drop/refresh (analytics/lab) vs must be preserved (core).

## Units and naming conventions
### Data layer (DB + APIs)
- **In-game currency**: `*_points`
  - `budget_points`, `bid_points`, `payout_points` (if using points payouts)
- **Real-world currency**: `*_cents` (or `*_dollars` if needed)
  - `entry_fee_cents`, `prize_amount_cents`
- **Normalized metrics**: unitless
  - `normalized_payout` (dimensionless)

### Presentation layer
We may display points using “$” metaphors (e.g., "$100 bankroll") for UX/investment framing, but:
- storage and APIs must remain **unambiguous** (`*_points` vs `*_cents`)
- any “$” shown for points is a **display label**, not a true currency value

## Ownership rules
- **Core app** owns:
  - `core.*`
  - `analytics.*`
- **Lab workflows** (Airflow + ML) own:
  - `lab_*.*`

## What moves where (high-level)
### Product-facing analytics (belongs in `analytics`)
- tournament state snapshots
- simulation batches
- evaluation runs
- entry simulation outcomes
- entry performance summaries
- hall-of-fame aggregates (pool-scoped)

### Lab-only ML artifacts (belongs in `lab_*`)
- predicted returns/investment models
- predicted market share
- optimizer outputs (unless explicitly promoted into a product strategy)

## Migration strategy (schema rename, short-term)
- Create `analytics`, `lab_bronze`, `lab_silver`, `lab_gold` schemas.
- Move/rename tables in phases with compatibility views (if needed).
- Update Go + SQLC queries to reference new schemas.
- Only after code is migrated and data is validated, drop old schemas.
