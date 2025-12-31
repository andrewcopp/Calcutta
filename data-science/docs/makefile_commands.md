# Makefile Commands for ML Analytics

## Main Application (Always Running)

```bash
# Start main Calcutta app (backend + frontend + postgres)
make up

# Stop main app
make down

# Reset main app (removes volumes)
make reset

# Run migrations
make ops-migrate
```

## Analytics Database (Lightweight - Start When Needed)

```bash
# Start just the analytics database (for Python pipeline)
make analytics-db-up

# Stop analytics database
make analytics-db-down
```

**Use case:** When you want to run Python pipeline locally and write to database, but don't need Airflow orchestration.

**Example workflow:**
```bash
# Terminal 1: Start main app
make up

# Terminal 2: Start analytics DB
make analytics-db-up

# Terminal 3: Run pipeline with database writes
cd data-science
export CALCUTTA_WRITE_TO_DB=true
python -m moneyball.cli report --year 2025

# Verify data
curl http://localhost:8080/api/v1/analytics/tournaments/2025/runs
```

## Airflow (Heavyweight - Start Only for Orchestration Testing)

```bash
# Start full Airflow stack (scheduler, webserver, worker, redis, postgres)
make airflow-up

# Stop Airflow
make airflow-down

# Watch Airflow worker logs
make airflow-logs

# Reset Airflow (removes all data/volumes)
make airflow-reset
```

**Use case:** When you want to test the full orchestrated pipeline with Airflow DAGs.

**Example workflow:**
```bash
# Terminal 1: Start main app
make up

# Terminal 2: Start Airflow (includes analytics DB)
make airflow-up

# Wait ~30 seconds for Airflow to initialize
# Open browser: http://localhost:8080 (airflow/airflow)
# Trigger DAG: calcutta_pipeline_local

# Watch logs
make airflow-logs
```

## Architecture Overview

```
┌─────────────────────────────────────────────────────┐
│  make up                                            │
│  ├── Backend API (Go)          :8080                │
│  ├── Frontend (React)          :3000                │
│  └── Main Postgres             :5432                │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│  make analytics-db-up                               │
│  └── Analytics Postgres        :5433                │
└─────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────┐
│  make airflow-up                                    │
│  ├── Airflow Webserver         :8080                │
│  ├── Airflow Scheduler                              │
│  ├── Airflow Worker                                 │
│  ├── Redis                     :6379                │
│  ├── Airflow Postgres          :5434                │
│  └── Analytics Postgres        :5433                │
└─────────────────────────────────────────────────────┘
```

## Typical Development Workflows

### 1. Backend API Development
```bash
make up                    # Just the main app
# Develop backend, test API endpoints
```

### 2. Python Pipeline Development (with Database)
```bash
make up                    # Main app (for API testing)
make analytics-db-up       # Analytics DB
# Run Python pipeline, verify via API
```

### 3. Airflow DAG Development
```bash
make up                    # Main app
make airflow-up            # Full Airflow stack
# Test DAGs in Airflow UI
make airflow-logs          # Debug issues
```

### 4. Full Integration Testing
```bash
make up                    # Main app
make airflow-up            # Airflow
# Trigger DAG, verify data flows through entire system
```

## Resource Usage

| Command | Containers | Memory | Startup Time |
|---------|-----------|--------|--------------|
| `make up` | 3 | ~500MB | 5-10s |
| `make analytics-db-up` | 1 | ~50MB | 2-3s |
| `make airflow-up` | 6 | ~1.5GB | 30-60s |

## Troubleshooting

### Analytics DB won't start
```bash
# Check if port 5433 is in use
lsof -i :5433

# Reset and restart
make analytics-db-down
make analytics-db-up
```

### Airflow stuck initializing
```bash
# Check logs
make airflow-logs

# Common issue: Need to initialize DB
make airflow-reset
make airflow-up
```

### Port conflicts
- Main app uses: 8080 (API), 3000 (frontend), 5432 (postgres)
- Analytics DB uses: 5433
- Airflow uses: 8080 (UI), 5434 (postgres), 6379 (redis)

**Note:** Airflow UI conflicts with main API on port 8080. If both are running, Airflow UI will be inaccessible. This is intentional - you typically run them separately.
