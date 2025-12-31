# Calcutta Analytics Database

Postgres database with medallion architecture for Calcutta tournament analytics.

## Architecture

**Medallion Layers:**
- **Bronze**: Raw data (tournaments, teams, simulations, entry bids)
- **Silver**: Cleaned/enriched data (ML predictions, market forecasts)
- **Gold**: Business metrics (entry performance, rankings, optimization runs)

**Design Principles:**
- Airflow-ready: Each pipeline stage is a discrete task
- Polyglot services: Go for compute-heavy tasks, Python for ML
- Direct database writes: No parquet intermediaries
- Single source of truth: Postgres is the canonical data store

## Table Write Responsibilities

**IMPORTANT**: Each service has exclusive write access to specific tables. This prevents conflicts and ensures clear ownership.

### Go Service Writes

**Bronze Layer:**
- `bronze_tournaments` - Tournament metadata
- `bronze_teams` - Team information with KenPom ratings
- `bronze_simulated_tournaments` - Simulation results (compute-intensive, parallel)
- `bronze_calcuttas` - Calcutta auction metadata
- `bronze_entry_bids` - Actual auction bids
- `bronze_payouts` - Prize structure

**Gold Layer:**
- `gold_entry_simulation_outcomes` - Per-simulation entry results (compute-intensive)

### Python Service Writes

**Silver Layer:**
- `silver_predicted_game_outcomes` - ML model predictions (sklearn)
- `silver_predicted_market_share` - Market prediction models (sklearn)
- `silver_team_tournament_value` - Expected points calculations

**Gold Layer:**
- `gold_optimization_runs` - Optimizer execution metadata
- `gold_recommended_entry_bids` - Portfolio optimizer outputs
- `gold_entry_performance` - Aggregated entry metrics
- `gold_detailed_investment_report` - Team-level ROI analysis

### Database Migrations

**Managed by Go**: All schema migrations are handled by the Go service using existing migration tooling. Both services read from the same schema but write to their designated tables only.

## Quick Start

### 1. Start Postgres

```bash
docker-compose up -d
```

### 2. Initialize Schema

Schema is managed by Go migrations. Run migrations from the Go service:

```bash
# From the Go service directory
make migrate-up
```

### 3. Load Data

Data is written directly to Postgres by the respective services:
- Go simulator writes tournament simulations
- Python ML service writes predictions
- Python optimizer writes portfolio recommendations

See the Airflow DAG section below for the full pipeline orchestration.

## Environment Variables

```bash
export CALCUTTA_ANALYTICS_DB_HOST=localhost
export CALCUTTA_ANALYTICS_DB_PORT=5432
export CALCUTTA_ANALYTICS_DB_NAME=calcutta_analytics
export CALCUTTA_ANALYTICS_DB_USER=postgres
export CALCUTTA_ANALYTICS_DB_PASSWORD=postgres
```

## Schema Overview

### Bronze Layer (Raw Data)
- `bronze_tournaments`: Tournament metadata
- `bronze_teams`: Team information with KenPom ratings
- `bronze_simulated_tournaments`: 5000+ simulations per tournament (compute-heavy, Go candidate)
- `bronze_entry_bids`: Actual auction bids from real entries
- `bronze_calcuttas`: Calcutta auction metadata
- `bronze_payouts`: Prize structure

### Silver Layer (ML Outputs)
- `silver_predicted_game_outcomes`: Game-by-game win probabilities (Python/sklearn)
- `silver_predicted_market_share`: Market prediction model outputs (Python/sklearn)
- `silver_team_tournament_value`: Expected points per team

### Gold Layer (Business Metrics)
- `gold_optimization_runs`: Tracks each strategy execution
- `gold_recommended_entry_bids`: Optimizer outputs (MINLP, greedy, etc.)
- `gold_entry_simulation_outcomes`: Per-simulation results for drill-down
- `gold_entry_performance`: Aggregated metrics (normalized payout, percentiles)
- `gold_detailed_investment_report`: Team-level ROI analysis

## Airflow DAG Structure

The analytics pipeline is orchestrated by Airflow using DockerOperators. Each stage runs in an isolated Docker container with the appropriate language/runtime.

**DockerOperator**: An Airflow task that runs a Docker container. Benefits include:
- Isolated execution environments
- Language-specific images (Go for compute, Python for ML)
- Easy horizontal scaling
- Version control via Docker tags

See `airflow/dags/calcutta_analytics_pipeline.py` for the full DAG implementation.

**Pipeline Flow:**
```
[Go: Simulate Tournaments]
    ↓
[Python: Predict Games] → [Python: Predict Market]
    ↓                           ↓
    └─────────→ [Python: Optimize Portfolio]
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
