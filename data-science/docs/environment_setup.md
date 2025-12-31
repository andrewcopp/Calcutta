# Environment Setup for ML Analytics

## Standardized Environment Variables

All services (Go backend, Python pipeline, Airflow) use the **same environment variables**:

```bash
# Database Connection
DB_HOST=localhost          # or 'db' in Docker
DB_PORT=5432
DB_NAME=calcutta
DB_USER=calcutta
DB_PASSWORD=calcutta

# Python Pipeline Control
CALCUTTA_WRITE_TO_DB=true  # Enable database writes
DB_MIN_CONN=1              # Connection pool min
DB_MAX_CONN=10             # Connection pool max
```

## No Manual .env Copying Required!

### For Airflow
Environment variables are **pre-configured** in `docker-compose.airflow.yml`. No manual setup needed.

### For Local Development
Copy the example file once:
```bash
cd data-science
cp .env.example .env
# Edit if needed (defaults work for local development)
```

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│  Main App (make up)                                     │
│  ├── Backend (Go)         → Reads from calcutta DB      │
│  ├── Frontend (React)                                   │
│  └── Postgres (calcutta)  → Single database             │
│      ├── Operational tables                             │
│      └── Analytics tables (added via migration)         │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  Python Pipeline                                        │
│  └── Writes to same calcutta DB                        │
└─────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────┐
│  Airflow (make airflow-up)                             │
│  ├── Shares network with main app                      │
│  ├── Env vars pre-configured                           │
│  └── Connects to calcutta DB via 'db' service          │
└─────────────────────────────────────────────────────────┘
```

## Usage Examples

### Local Python Development
```bash
# Start main app
make up

# Run pipeline (writes to DB)
cd data-science
export CALCUTTA_WRITE_TO_DB=true
python -m moneyball.cli report --year 2025

# Verify via API
curl http://localhost:8080/api/v1/analytics/tournaments/2025/runs
```

### Airflow Development
```bash
# Start main app
make up

# Start Airflow (env vars already configured)
make airflow-up

# Trigger DAG in UI (http://localhost:8080)
# No .env copying needed - it just works!
```

## Why This Design?

1. **Single Source of Truth**: One database for all data
2. **No Premature Optimization**: Keep it simple until we need to scale
3. **Consistent Naming**: Same env vars across all services
4. **Zero Manual Setup**: Airflow has env vars baked in
5. **Easy Testing**: Start main app, run pipeline, query API

## Migration Path (Future)

If we need to separate databases later:
1. Create new analytics DB
2. Update env vars to point to new DB
3. Run data migration script
4. No code changes needed!

The abstraction is already in place - just change the connection string.
