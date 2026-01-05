# Database Integration Guide

## Overview

This guide shows how to integrate database writes into your existing pipeline stages while maintaining parquet files for backward compatibility.

## Architecture

```
Pipeline Stage
    ↓
Write to Parquet (existing)
    ↓
Write to Database (new) ← Optional, controlled by env var
    ↓
Return parquet path (existing)
```

## Environment Variables

Set these to enable database writes:

```bash
# Enable database writes
export CALCUTTA_WRITE_TO_DB=true

# Database connection (same as Go backend)
export CALCUTTA_ANALYTICS_DB_HOST=localhost
export CALCUTTA_ANALYTICS_DB_PORT=5432
export CALCUTTA_ANALYTICS_DB_NAME=calcutta_analytics
export CALCUTTA_ANALYTICS_DB_USER=postgres
export CALCUTTA_ANALYTICS_DB_PASSWORD=...
```

## Integration Pattern

### Step 1: Import the database writer

```python
from moneyball.pipeline.db_writer import get_db_writer
```

### Step 2: Add database writes after parquet writes

```python
def _stage_simulated_tournaments(...):
    # ... existing code ...
    
    # Write to parquet (existing)
    result.to_parquet(out_path, index=False)
    
    # Write to database (new - optional based on env var)
    db_writer = get_db_writer()
    db_writer.write_simulated_tournaments(
        tournament_key=tournament_key,
        simulations_df=result,
    )
    
    # ... rest of existing code ...
    return out_path, manifest
```

## Stage-by-Stage Integration

### 1. Bronze Layer (Snapshot Data)

**When:** At the start of a pipeline run  
**Where:** Add to `run()` function or create a new `_stage_bronze_layer()`

```python
def run(...):
    # ... existing setup ...
    
    # Write bronze layer to database
    db_writer = get_db_writer()
    db_writer.write_bronze_layer(
        snapshot_dir=sd,
        tournament_key=f"ncaa-tournament-{sname}",
        calcutta_key=calcutta_key,
    )
    
    # ... continue with existing stages ...
```

### 2. Predicted Game Outcomes (Silver Layer)

**File:** `_stage_predicted_game_outcomes()`  
**After:** `df.to_parquet(out_path, index=False)`

```python
# Write to database
db_writer = get_db_writer()
db_writer.write_predicted_game_outcomes(
    tournament_key=f"ncaa-tournament-{snapshot_dir.name}",
    predictions_df=df,
    model_version="kenpom-v1",
)
```

### 3. Simulated Tournaments (Bronze Layer)

**File:** `_stage_simulated_tournaments()`  
**After:** `result.to_parquet(out_path, index=False)`

```python
# Write to database
db_writer = get_db_writer()
db_writer.write_simulated_tournaments(
    tournament_key=f"ncaa-tournament-{snapshot_dir.name}",
    simulations_df=result,
)
```

### 4. Predicted Market Share (Silver Layer)

**File:** `_stage_predicted_auction_share_of_pool()`  
**After:** `df.to_parquet(out_path, index=False)`

```python
# Write to database
db_writer = get_db_writer()
db_writer.write_predicted_market_share(
    calcutta_key=calcutta_key,
    predictions_df=df,
    model_version="ridge-v1",
)
```

### 5. Recommended Entry Bids (Gold Layer)

**File:** `_stage_recommended_entry_bids()`  
**After:** `df.to_parquet(out_path, index=False)`

```python
# Write to database
db_writer = get_db_writer()
db_writer.write_optimization_results(
    run_id=store.run_id,  # Use orchestrator's run_id
    calcutta_key=calcutta_key,
    strategy=strategy,
    n_sims=n_sims,
    seed=seed,
    budget_points=budget_points,
    recommended_bids_df=df,
)
```

### 6. Simulated Entry Outcomes (Gold Layer)

**File:** `_stage_simulated_entry_outcomes()`  
**After:** `summary_df.to_parquet(out_summary_path, index=False)`

```python
# Write to database
db_writer = get_db_writer()
db_writer.write_entry_outcomes(
    run_id=store.run_id,
    outcomes_df=sims_df if keep_sims else pd.DataFrame(),
    summary_df=summary_df,
)
```

## Testing

### Test with Database Disabled (Default)

```bash
# Database writes are disabled by default
python -m moneyball.cli report --year 2025
```

### Test with Database Enabled

```bash
# Enable database writes
export CALCUTTA_WRITE_TO_DB=true
export CALCUTTA_ANALYTICS_DB_HOST=localhost
export CALCUTTA_ANALYTICS_DB_NAME=calcutta_analytics

# Run pipeline
python -m moneyball.cli report --year 2025
```

### Verify Data in Database

```bash
# Check data was written
psql -h localhost -U postgres calcutta_analytics

# Query examples
SELECT COUNT(*) FROM bronze_simulated_tournaments;
SELECT COUNT(*) FROM silver_predicted_game_outcomes;
SELECT COUNT(*) FROM gold_optimization_runs;
```

## Gradual Migration Strategy

1. **Phase 1 (Current):** Write to both parquet and database
   - Parquet remains source of truth
   - Database is for API serving only
   - Easy rollback if issues

2. **Phase 2 (Future):** Read from database in pipeline
   - Keep writing to both
   - Start reading from database for downstream stages
   - Validate consistency

3. **Phase 3 (Final):** Database-first
   - Stop writing parquet files
   - Database is source of truth
   - Parquet only for exports/backups

## Troubleshooting

### Database Connection Fails

The database writer will gracefully disable itself and log a warning:
```
⚠ Database writer disabled: missing env vars ['CALCUTTA_ANALYTICS_DB_HOST']
```

Pipeline continues normally, only writing to parquet.

### Partial Writes

If a database write fails mid-stage, it logs an error but doesn't stop the pipeline:
```
⚠ Failed to write simulations: connection timeout
```

The parquet file is still written successfully.

### Performance

Database writes add ~2-5 seconds per stage for typical datasets:
- Simulated tournaments (5000 sims × 68 teams): ~3 seconds
- Entry outcomes (100 entries × 5000 sims): ~4 seconds

This is acceptable for batch processing but can be disabled for rapid iteration.

## Next Steps

1. Add database writes to your pipeline stages (see examples above)
2. Test locally with database enabled
3. Set up Airflow to orchestrate the full pipeline
4. Deploy to production with database writes enabled
