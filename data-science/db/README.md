# Calcutta Analytics Database

Postgres database backing the Calcutta app and analytics workflows.

## Architecture

This repo is converging on a schema-qualified layout:

- `core.*` for canonical entities and invariants.
- `derived.*` for computed/derived artifacts.

Some legacy/compat schemas may exist (e.g. `lab_*`) to support incremental migration.

Data science workflows in `data-science/` commonly produce Parquet artifacts on disk. Some stages also write to Postgres when enabled.

## Table Write Responsibilities

**IMPORTANT**: Prefer clear ownership by layer.

### Python writes (ML outputs)

- `lab_silver.predicted_game_outcomes`
- `lab_silver.predicted_market_share`

Python should stay focused on training/inference and writing the prediction artifacts.

### Go writes (simulation/evaluation/optimization)

Go binaries under `backend/cmd/*` are the preferred path for simulation/evaluation/optimization and writing derived artifacts.

### Database Migrations

**Managed by Go**: All schema migrations are handled by the Go service using existing migration tooling. Both services read from the same schema but write to their designated tables only.

## Quick Start

### 1. Start Postgres

```bash
make up
```

### 2. Initialize Schema

Schema is managed by Go migrations.

```bash
make ops-migrate
```

### 3. Load Data

See `data-science/README.md` for running snapshot-based pipelines.

## Environment Variables

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=calcutta
export DB_USER=calcutta
export DB_PASSWORD=...
```

## Schema Overview

The canonical schemas are `core.*` and `derived.*`.

ML prediction artifacts are written under `lab_silver.*`.

## Airflow DAG Structure

The analytics pipeline is orchestrated by Airflow using DockerOperators. Each stage runs in an isolated Docker container with the appropriate language/runtime.

**DockerOperator**: An Airflow task that runs a Docker container. Benefits include:
- Isolated execution environments
- Language-specific images (Go for compute, Python for ML)
- Easy horizontal scaling
- Version control via Docker tags

See `airflow/dags/calcutta_analytics_pipeline.py` for the full DAG implementation.

**Pipeline Flow (high-level):**
```
[Go: Simulate Tournaments]
    ↓
[Python: Predict Games] → [Python: Predict Market]
    ↓                           ↓
    └─────────→ [Go: Optimize Portfolio]
                        ↓
                [Go: Evaluate All Entries]
```

## Queries

### Get entry rankings
```sql
SELECT * FROM view_entry_rankings
WHERE run_id = '20251230T215837Z'
ORDER BY rank;
```

### Get entry portfolio
```sql
SELECT * FROM get_entry_portfolio('20251230T215837Z', 'entry-andrew-copp');
```

### Tournament simulation stats
```sql
SELECT * FROM view_tournament_sim_stats
WHERE season = 2025;
```

## Performance Considerations

**Large Tables:**
- `bronze_simulated_tournaments`: ~320K rows per year (64 teams × 5000 sims)
- `gold_entry_simulation_outcomes`: ~240K rows per year (48 entries × 5000 sims)

**Optimization:**
- Batch inserts for simulations (10K rows at a time)
- Indexes on foreign keys and query patterns
- Partitioning by year (future enhancement)

**Go Candidates (for speed):**
- Tournament simulation (math-heavy, embarrassingly parallel)
- Entry evaluation across all simulations (compute-intensive)
- Market simulation (if needed)
