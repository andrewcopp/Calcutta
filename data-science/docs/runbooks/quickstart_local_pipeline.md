# Quick Start: Local Pipeline with Database

## Prerequisites

1. **Database running** (from main Calcutta project):
   ```bash
   cd /Users/andrewcopp/Developer/Calcutta
   make up  # Starts postgres on port 5432
   ```

2. **Analytics database** (separate):
   ```bash
   cd /Users/andrewcopp/Developer/Calcutta/data-science
   docker-compose -f docker-compose.airflow.yml up postgres-analytics -d
   # Runs on port 5433
   ```

3. **Migrations applied**:
   ```bash
   cd /Users/andrewcopp/Developer/Calcutta/backend
   make migrate-up
   ```

## Option 1: Run Pipeline Manually (Fastest for Testing)

### Step 1: Configure Environment

```bash
cd /Users/andrewcopp/Developer/Calcutta/data-science

# Copy example env file
cp .env.example .env

# Edit .env and set:
# CALCUTTA_WRITE_TO_DB=true
# CALCUTTA_ANALYTICS_DB_HOST=localhost
# CALCUTTA_ANALYTICS_DB_PORT=5433
```

### Step 2: Run Pipeline

```bash
# Run full pipeline for 2025
python -m moneyball.cli report --year 2025

# Or with custom parameters
python -m moneyball.cli report \
    --year 2025 \
    --n-sims 5000 \
    --seed 42 \
    --strategy minlp
```

### Step 3: Verify Data

```bash
# Check parquet files (existing behavior)
ls -lh out/2025/derived/calcutta/*/

# Check database (new behavior)
psql -h localhost -p 5433 -U postgres -d calcutta_analytics

# In psql:
SELECT COUNT(*) FROM bronze_simulated_tournaments;
SELECT COUNT(*) FROM silver_predicted_game_outcomes;
SELECT COUNT(*) FROM gold_optimization_runs;
```

### Step 4: Query via Go API

```bash
# API is already running from `make up`
curl http://localhost:8080/api/v1/analytics/tournaments/2025/runs

# Get our strategy details
curl http://localhost:8080/api/v1/analytics/tournaments/2025/runs/{run_id}/our-entry
```

## Option 2: Run with Airflow (Production-like)

### Step 1: Start Airflow

```bash
cd /Users/andrewcopp/Developer/Calcutta/data-science

# Set Airflow UID (first time only)
echo -e "AIRFLOW_UID=$(id -u)" > .env.airflow

# Start all Airflow services
docker-compose -f docker-compose.airflow.yml up -d

# Wait for services to be healthy (~30 seconds)
docker-compose -f docker-compose.airflow.yml ps
```

### Step 2: Access Airflow UI

```bash
# Open browser to http://localhost:8080
# Login: airflow / airflow

# You'll see two DAGs:
# - calcutta_analytics_pipeline (production, uses Docker images)
# - calcutta_pipeline_local (development, uses local Python)
```

### Step 3: Trigger DAG

```bash
# Via UI: Click on "calcutta_pipeline_local" > Trigger DAG

# Or via CLI:
docker exec airflow-scheduler airflow dags trigger calcutta_pipeline_local \
    --conf '{"year": "2025", "n_sims": 5000, "seed": 42, "strategy": "minlp"}'
```

### Step 4: Monitor Progress

```bash
# Watch logs
docker-compose -f docker-compose.airflow.yml logs -f airflow-worker

# Or in UI: Click on DAG run > Graph view
```

## Troubleshooting

### Database Connection Fails

```bash
# Check analytics DB is running
docker ps | grep calcutta-analytics-db

# Test connection
psql -h localhost -p 5433 -U postgres -d calcutta_analytics -c "SELECT 1"

# If fails, restart:
docker-compose -f docker-compose.airflow.yml restart postgres-analytics
```

### Pipeline Runs But No Database Writes

```bash
# Check environment variable
echo $CALCUTTA_WRITE_TO_DB  # Should be "true"

# Check logs for warnings
python -m moneyball.cli report --year 2025 2>&1 | grep "Database"

# Should see:
# ✓ Wrote bronze layer to database
# ✓ Wrote 340000 simulation records to database
# ✓ Wrote 67 game predictions to database
```

### Airflow DAG Not Appearing

```bash
# Check DAG file syntax
docker exec airflow-scheduler python -m py_compile \
    /opt/airflow/dags/calcutta_pipeline_local.py

# Refresh DAGs
docker exec airflow-scheduler airflow dags list-import-errors
```

## Next Steps

1. **Integrate database writes into pipeline stages**
   - See `docs/database_integration_guide.md`
   - Add `get_db_writer()` calls after parquet writes

2. **Build frontend dashboard**
   - Consume Go API endpoints
   - Display optimization runs, entry rankings, etc.

3. **Deploy to production**
   - Use production Airflow (AWS MWAA, Kubernetes, etc.)
   - Set `CALCUTTA_WRITE_TO_DB=true` in production env
   - Schedule DAG to run daily/weekly

## Architecture Summary

```
┌─────────────────────────────────────────────────────────┐
│                   Airflow Scheduler                      │
│              (Orchestrates pipeline runs)                │
└─────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        │                   │                   │
   ┌────▼────┐         ┌────▼────┐        ┌────▼────┐
   │ Python  │         │ Python  │        │ Python  │
   │  CLI    │         │  CLI    │        │  CLI    │
   │ (Stage) │         │ (Stage) │        │ (Stage) │
   └────┬────┘         └────┬────┘        └────┬────┘
        │                   │                   │
        │ Writes to:        │                   │
        │ 1. Parquet        │                   │
        │ 2. Database       │                   │
        │                   │                   │
        └───────────────────┴───────────────────┘
                            │
                ┌───────────┴───────────┐
                │                       │
         ┌──────▼──────┐        ┌──────▼──────┐
         │  Parquet    │        │  Postgres   │
         │   Files     │        │  Analytics  │
         │ (Debugging) │        │     DB      │
         └─────────────┘        └──────┬──────┘
                                       │
                                ┌──────▼──────┐
                                │   Go API    │
                                │  (Serving)  │
                                └──────┬──────┘
                                       │
                                ┌──────▼──────┐
                                │  Frontend   │
                                │  Dashboard  │
                                └─────────────┘
```

## Performance Notes

- **Parquet writes**: ~1-2 seconds per stage
- **Database writes**: ~2-5 seconds per stage (batch inserts)
- **Total overhead**: ~10-20 seconds for full pipeline
- **Acceptable** for batch processing, can disable for rapid iteration
