# Hybrid Architecture Implementation Summary

## Architecture Overview

**Hybrid Approach**: Python writes directly to database, reads from Go API

```
┌─────────────────────────────────────────────────────────────┐
│                        Airflow DAG                          │
└─────────────────────────────────────────────────────────────┘
                              │
                ┌─────────────┴─────────────┐
                │                           │
         ┌──────▼──────┐           ┌───────▼────────┐
         │ Go Services │           │ Python Services│
         │  (Compute)  │           │   (ML/Optim)   │
         └──────┬──────┘           └───────┬────────┘
                │                           │
                │ WRITE                     │ WRITE
                │ (Bronze + Gold)           │ (Silver + Gold)
                │                           │
         ┌──────▼───────────────────────────▼──────┐
         │         Postgres (Analytics)            │
         │  Bronze / Silver / Gold Tables          │
         └──────┬──────────────────────────────────┘
                │ READ
         ┌──────▼──────┐
         │   Go API    │  (Read-only for analytics)
         │  (Serving)  │
         └──────┬──────┘
                │ READ
         ┌──────▼──────┐
         │  Frontend   │
         │   (React)   │
         └─────────────┘
```

## What's Been Implemented

### 1. Database Schema ✅
**Location**: `backend/migrations/schema/20251231000000_add_analytics_tables.{up,down}.sql`

- Bronze/Silver/Gold medallion architecture
- 15 tables with proper indexes and foreign keys
- 3 views for common queries
- 1 function for portfolio retrieval

### 2. Go API Layer ✅

**sqlc Queries**: `backend/internal/adapters/db/sqlc/queries/ml_analytics.sql`
- 11 type-safe SQL queries for reading analytics data
- Follows existing sqlc patterns

**Port Interfaces**: `backend/internal/ports/ml_analytics.go`
- `MLAnalyticsRepo` interface with 8 methods
- Well-defined domain models
- Follows Single Responsibility Principle

**API Endpoints**: `docs/api_endpoints.md`
- 8 RESTful endpoints specified
- Request/response formats defined
- Error handling patterns

### 3. Python Database Writers ✅

**Connection Module**: `moneyball/db/connection.py`
- Thread-safe connection pooling
- Context manager for safe connection handling
- Environment variable configuration

**Bronze Writers**: `moneyball/db/writers/bronze_writers.py`
- `write_tournaments()` - Tournament metadata
- `write_teams()` - Team information
- `write_simulated_tournaments()` - Simulation results (batch inserts)
- `write_calcuttas()` - Calcutta metadata
- `write_entry_bids()` - Actual auction bids
- `write_payouts()` - Prize structure

**Silver Writers**: `moneyball/db/writers/silver_writers.py`
- `write_predicted_game_outcomes()` - ML game predictions
- `write_predicted_market_share()` - Market predictions
- `write_team_tournament_value()` - Expected points

**Gold Writers**: `moneyball/db/writers/gold_writers.py`
- `write_optimization_run()` - Run metadata
- `write_recommended_entry_bids()` - Optimizer outputs
- `write_entry_simulation_outcomes()` - Per-sim results (batch inserts)
- `write_entry_performance()` - Aggregated metrics
- `write_detailed_investment_report()` - ROI analysis

### 4. Airflow Orchestration ✅

**DAG**: `airflow/dags/calcutta_analytics_pipeline.py`
- 7 tasks with clear dependencies
- DockerOperator for each service
- Configurable via DAG run parameters

**Docker Compose**: `docker-compose.airflow.yml`
- Complete Airflow stack (webserver, scheduler, worker, triggerer)
- Separate Postgres for Airflow metadata and analytics data
- Redis for Celery executor

**Documentation**: `docs/airflow_setup.md`
- Local development setup
- Production deployment options (AWS MWAA, K8s)
- Best practices and troubleshooting

## Write Responsibilities

### Go Writes
- `bronze_tournaments`
- `bronze_teams`
- `bronze_simulated_tournaments` (compute-intensive)
- `bronze_calcuttas`
- `bronze_entry_bids`
- `bronze_payouts`
- `gold_entry_simulation_outcomes` (compute-intensive)

### Python Writes
- `silver_predicted_game_outcomes` (sklearn)
- `silver_predicted_market_share` (sklearn)
- `silver_team_tournament_value`
- `gold_optimization_runs`
- `gold_recommended_entry_bids`
- `gold_entry_performance`
- `gold_detailed_investment_report`

## Performance Optimizations

### Batch Inserts
Both Go and Python use batch inserts for high-volume tables:
- Simulations: 10K rows per batch (320K total per year)
- Entry outcomes: 10K rows per batch (240K total per year)

### Connection Pooling
- Python: ThreadedConnectionPool (1-10 connections)
- Go: pgxpool (configured per service)

### Indexes
All foreign keys and query patterns have indexes:
- `(tournament_key, sim_id)` for simulations
- `(run_id, entry_key)` for entry outcomes
- `(run_id, percentile_rank)` for rankings

## Usage Examples

### Python: Writing Simulations

```python
from moneyball.db.writers import write_simulated_tournaments
import pandas as pd

# Simulate tournaments (your existing code)
simulations_df = simulate_tournaments(year=2025, n_sims=5000)

# Write directly to Postgres
write_simulated_tournaments(
    tournament_key="ncaa-tournament-2025",
    simulations_df=simulations_df
)
```

### Python: Writing Optimizer Results

```python
from moneyball.db.writers import (
    write_optimization_run,
    write_recommended_entry_bids,
    write_detailed_investment_report
)

# Run optimizer (your existing code)
run_id = "20251231T120000Z"
recommended_bids = optimize_portfolio(strategy="minlp")

# Write to Postgres
write_optimization_run(
    run_id=run_id,
    calcutta_key="calcutta-2025",
    strategy="minlp",
    n_sims=5000,
    seed=42
)

write_recommended_entry_bids(run_id, recommended_bids)
write_detailed_investment_report(run_id, investment_report)
```

### Go: Reading from API (Frontend)

```javascript
// Fetch entry rankings
const response = await fetch(
  `/api/v1/analytics/tournaments/2025/runs/${runId}/rankings`
);
const data = await response.json();

// Display rankings
data.entries.forEach(entry => {
  console.log(`${entry.rank}. ${entry.entry_key}: ${entry.mean_normalized_payout}`);
});
```

## Next Steps

### Immediate (Go Implementation)
1. Run sqlc to generate Go code: `cd backend && sqlc generate`
2. Implement `MLAnalyticsRepository` in `backend/internal/adapters/db/ml_analytics_repository.go`
3. Write repository tests following GIVEN/WHEN/THEN structure
4. Implement service layer in `backend/internal/app/ml_analytics_service.go`
5. Implement HTTP handlers in `backend/internal/transport/httpserver/ml_analytics_handlers.go`
6. Add routes to `backend/internal/transport/httpserver/router.go`

### Python Integration
1. Update existing Python scripts to use new database writers
2. Remove parquet file generation (or make optional for backup)
3. Test end-to-end: Python writes → Postgres → Go API reads

### Frontend
1. Create React components for 6 analytics views
2. Integrate with Go API endpoints
3. Add to existing frontend navigation

### Deployment
1. Build Docker images for each service
2. Test Airflow DAG locally
3. Deploy to staging environment
4. Configure production monitoring

## Testing Strategy

### Go Tests (Unit Only)
```go
func TestThatEntryRankingsAreSortedByNormalizedPayout(t *testing.T) {
    // GIVEN
    mockRepo := &mockMLAnalyticsRepo{
        rankings: []ports.EntryRanking{
            {Rank: 1, MeanNormalizedPayout: 0.578},
            {Rank: 2, MeanNormalizedPayout: 0.529},
        },
    }
    service := NewMLAnalyticsService(mockRepo)
    
    // WHEN
    rankings, err := service.GetEntryRankings(ctx, 2025, runID, 10, 0)
    
    // THEN
    assert.NoError(t, err)
    assert.Equal(t, 2, len(rankings))
    assert.Greater(t, rankings[0].MeanNormalizedPayout, 
                   rankings[1].MeanNormalizedPayout)
}
```

### Python Tests
```python
def test_write_simulated_tournaments_batch_inserts():
    # GIVEN
    simulations_df = pd.DataFrame({
        'sim_id': range(5000),
        'team_key': ['team-1'] * 5000,
        'wins': [3] * 5000,
        'byes': [0] * 5000,
        'eliminated': [False] * 5000,
    })
    
    # WHEN
    count = write_simulated_tournaments('tournament-1', simulations_df)
    
    # THEN
    assert count == 5000
```

## Environment Variables

```bash
# Analytics Database
export CALCUTTA_ANALYTICS_DB_HOST=localhost
export CALCUTTA_ANALYTICS_DB_PORT=5433
export CALCUTTA_ANALYTICS_DB_NAME=calcutta_analytics
export CALCUTTA_ANALYTICS_DB_USER=postgres
export CALCUTTA_ANALYTICS_DB_PASSWORD=postgres
```

## Benefits of Hybrid Approach

1. **Performance**: Python bulk writes are fast (no HTTP overhead)
2. **Simplicity**: Each service is self-contained
3. **Separation**: Data generation (Python) vs serving (Go)
4. **Scalability**: Services can scale independently
5. **Type Safety**: Go API uses sqlc for compile-time SQL validation
6. **Flexibility**: Python can use any DB library (psycopg2, SQLAlchemy)

## Migration from Parquet

Existing parquet files can be:
1. **One-time migrated** to Postgres for historical data
2. **Kept as backups** for debugging/archival
3. **Phased out** as new data flows through Postgres

No need to maintain parquet ETL pipeline going forward.
