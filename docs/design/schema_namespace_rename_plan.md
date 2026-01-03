# Schema Namespace Rename Plan (Single DB → Two DBs)

## Problem
We conflated:
- product boundary (core vs lab)
- medallion tiers (bronze/silver/gold)

We also want the long-term vision:
- Core and Lab are separate products and databases
- `bronze/silver/gold` namespaces should be available for Lab DB

## Short-term target (single DB)
Introduce explicit namespaces now:
- `core`: canonical product data
- `analytics`: product-derived materializations
- `lab_bronze`, `lab_silver`, `lab_gold`: lab-only medallion tiers

## Which tables belong where (initial classification)
### `analytics` (product-facing derived)
- tournament_state_snapshots (+ game locks)
- simulation_batches
- simulated_tournaments (currently `silver.simulated_tournaments`)
- evaluation_runs
- entry_simulation_outcomes (currently `gold.entry_simulation_outcomes`)
- entry_performance (currently `gold.entry_performance`)

### `lab_*` (experimental / ML)
- predicted_* tables
- optimizer outputs (unless explicitly promoted into strategies)
- ML runs metadata

## Migration approach (phased)
### Phase 0: Create new schemas
- Create schemas: `analytics`, `lab_bronze`, `lab_silver`, `lab_gold`.

### Phase 1: Move product analytics tables
- Move/rename:
  - `silver.simulated_tournaments` → `analytics.simulated_tournaments`
  - `gold.entry_simulation_outcomes` → `analytics.entry_simulation_outcomes`
  - `gold.entry_performance` → `analytics.entry_performance`

Compatibility:
- Create temporary views in the old schemas to avoid breaking older queries during migration.

### Phase 2: Update application code
- Update Go services and HTTP handlers to query `analytics.*`.
- Update SQLC query files accordingly.

### Phase 3: Clean up
- When no code references old schemas:
  - drop compatibility views
  - optionally drop old schemas if they are fully replaced

## Long-term (two DBs)
### Core DB
- `core`, `analytics`

### Lab DB
- `bronze`, `silver`, `gold`

Data movement options:
- replicate selected `core` tables to lab
- export/import snapshots

## Non-goals (for this plan)
- This plan does not decide ML model architecture.
- This plan does not decide UI naming; it only enforces unit clarity at the storage/API layer.
