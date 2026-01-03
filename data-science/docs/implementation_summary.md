# Analytics Infrastructure Implementation Summary

## What We've Built

### 1. Database Schema (Medallion Architecture)
**Location**: `backend/migrations/schema/20251231000000_add_analytics_tables.{up,down}.sql`

**Bronze Layer** (Raw Data):
- `bronze_tournaments` - Tournament metadata
- `bronze_teams` - Team information with KenPom ratings
- `bronze_simulated_tournaments` - 5000+ simulations per tournament
- `bronze_calcuttas` - Calcutta auction metadata
- `bronze_entry_bids` - Actual auction bids
- `bronze_payouts` - Prize structure

**Silver Layer** (ML Predictions):
- `silver_predicted_game_outcomes` - Game-by-game win probabilities
- `silver_predicted_market_share` - Market prediction model outputs
- `silver_team_tournament_value` - Expected points per team

**Gold Layer** (Business Metrics):
- `gold_optimization_runs` - Strategy execution tracking
- `gold_recommended_entry_bids` - Portfolio optimizer outputs
- `gold_entry_simulation_outcomes` - Per-simulation results (for drill-down)
- `gold_entry_performance` - Aggregated entry metrics
- `gold_detailed_investment_report` - Team-level ROI analysis

**Views & Functions**:
- `view_latest_optimization_runs` - Latest run per strategy
- `view_entry_rankings` - Entry rankings with percentiles
- `view_tournament_sim_stats` - Tournament simulation statistics
- `get_entry_portfolio()` - Get portfolio for any entry

### 2. API Endpoints Specification
**Location**: `docs/reference/api_endpoints.md`

8 RESTful endpoints designed:
1. `GET /tournaments/{year}/simulations` - Tournament overview
2. `GET /tournaments/{year}/teams/{team_key}/performance` - Team performance
3. `GET /tournaments/{year}/teams/predictions` - Team predictions
4. `GET /tournaments/{year}/runs/{run_id}/our-entry` - Our entry details
5. `GET /tournaments/{year}/runs/{run_id}/rankings` - Entry rankings
6. `GET /tournaments/{year}/runs/{run_id}/entries/{entry_key}/simulations` - Drill-down
7. `GET /tournaments/{year}/runs/{run_id}/entries/{entry_key}/portfolio` - Portfolio
8. `GET /tournaments/{year}/runs` - Available optimization runs

### 3. Go Port Interfaces
**Location**: `backend/internal/ports/ml_analytics.go`

Defined `MLAnalyticsRepo` interface with 8 methods following Single Responsibility Principle:
- Each method has one clear purpose
- Return types are well-defined domain models
- Follows existing port patterns in the codebase

### 4. Airflow Orchestration
**Location**: `airflow/dags/calcutta_analytics_pipeline.py`

Complete DAG with 7 tasks:
1. Load tournament data (Go)
2. Simulate tournaments (Go - compute-intensive)
3. Predict game outcomes (Python - ML)
4. Predict market share (Python - ML)
5. Calculate team value (Python)
6. Optimize portfolio (Python - MINLP)
7. Evaluate all entries (Go - compute-intensive)

### 5. Documentation

**Airflow Setup** (`data-science/docs/airflow_setup.md`):
- Local development with docker-compose
- Production deployment options (AWS MWAA, K8s, self-hosted)
- Best practices and troubleshooting

**Database README** (`db/README.md`):
- Table write responsibilities (Go vs Python)
- Architecture overview
- Query examples

**Go Models** (`data-science/docs/go_models.md`):
- Complete struct definitions for all analytics tables
- Query helper functions
- Usage examples

### 6. Infrastructure

**Docker Compose** (`docker-compose.airflow.yml`):
- Airflow webserver, scheduler, worker, triggerer
- Postgres for Airflow metadata (port 5434)
- Postgres for analytics data (port 5433)
- Redis for Celery executor

**Database Tools** (`db/init_db.py`):
- Initialize/reset database schema
- Drop all tables for development

## Architecture Decisions

### Polyglot Services
- **Go**: Compute-heavy tasks (simulation, evaluation), API layer
- **Python**: ML tasks (sklearn predictions, MINLP optimization)

### Table Write Responsibilities
**Go writes**:
- All bronze layer tables (raw data)
- `gold_entry_simulation_outcomes` (compute-intensive)

**Python writes**:
- All silver layer tables (ML predictions)
- Most gold layer tables (optimizer outputs, aggregations)

### Database Migrations
- Managed by Go service
- Both services read from same schema
- Clear separation of write responsibilities

### Data Flow
```
Parquet files (legacy) → One-time migration → Postgres
                                                  ↓
Future: Services write directly to Postgres (no parquet intermediaries)
```

## Next Steps

### Phase 1: Repository Implementation (Following Go Best Practices)
1. **Create sqlc queries** (`backend/internal/adapters/db/queries/ml_analytics.sql`)
   - Write SQL queries for each port method
   - Use sqlc to generate type-safe Go code
   
2. **Implement repository** (`backend/internal/adapters/db/ml_analytics_repository.go`)
   - Implement `MLAnalyticsRepo` interface
   - Use generated sqlc code
   - Follow existing repository patterns
   - **SRP**: Each method does one thing
   - **OCP**: Closed for modification, open for extension

3. **Write repository tests** (`backend/internal/adapters/db/ml_analytics_repository_test.go`)
   - GIVEN/WHEN/THEN structure
   - One assertion per test
   - Deterministic tests (fixed time, sorted results)
   - Test naming: `TestThat{Scenario}`

### Phase 2: Service Layer
1. **Create service** (`backend/internal/app/ml_analytics_service.go`)
   - Business logic layer
   - Orchestrates repository calls
   - Handles data transformations
   - **SRP**: Each method handles one business operation

2. **Write service tests** (`backend/internal/app/ml_analytics_service_test.go`)
   - Mock repository interface
   - Test business logic in isolation
   - Follow testing guidelines

### Phase 3: HTTP Handlers
1. **Create handlers** (`backend/internal/transport/httpserver/ml_analytics_handlers.go`)
   - Implement 8 API endpoints
   - Request validation
   - Response formatting
   - Error handling
   - **SRP**: Each handler handles one endpoint

2. **Write handler tests** (`backend/internal/transport/httpserver/ml_analytics_handlers_test.go`)
   - Test HTTP layer in isolation
   - Mock service interface
   - Test status codes, response formats

3. **Add routes** (`backend/internal/transport/httpserver/router.go`)
   - Register analytics endpoints
   - Apply middleware (auth, logging)
   - Group under `/api/v1/analytics`

### Phase 4: Python Services
1. **Create Python modules** for writing to silver/gold tables
2. **Implement direct DB writes** (no parquet intermediaries)
3. **Create Docker images** for each Python service

### Phase 5: Frontend Dashboard
1. **Create React components** for 6 views
2. **Integrate with Go API**
3. **Add to existing frontend**

### Phase 6: Airflow Deployment
1. **Build Docker images** for all services
2. **Test DAG locally**
3. **Deploy to staging/production**

## Testing Strategy

Following `docs/standards/engineering.md` and `docs/standards/bracket_testing_guidelines.md`:

### Unit Tests Only (For Now)
- No DB/integration tests initially
- Focus on business logic

### Test Structure
```go
func TestThatTeamPerformanceIsCalculatedCorrectly(t *testing.T) {
    // GIVEN
    repo := &mockMLAnalyticsRepo{...}
    service := NewMLAnalyticsService(repo)
    
    // WHEN
    result, err := service.GetTeamPerformance(ctx, 2025, "ncaa-tournament-2025:duke")
    
    // THEN
    assert.NoError(t, err)
    assert.Equal(t, expectedPerformance, result)
}
```

### Test Naming
- `TestThat{Scenario}` - Descriptive, behavior-focused
- Example: `TestThatEntryRankingsAreSortedByNormalizedPayout`

### Deterministic Tests
- Fix time/randomness
- Sort before comparing
- One reason to fail per test

## Running Locally

### 1. Start Databases
```bash
cd data-science
docker-compose -f docker-compose.airflow.yml up -d postgres-analytics
```

### 2. Run Migrations
```bash
cd backend
make migrate-up
```

### 3. Start Airflow (Optional)
```bash
cd data-science
docker-compose -f docker-compose.airflow.yml up -d
# Access UI: http://localhost:8080 (airflow/airflow)
```

### 4. Start Go API
```bash
cd backend
make run
```

## Production Considerations

### Performance
- Batch inserts for simulations (10K rows at a time)
- Indexes on all foreign keys and query patterns
- Read replicas for analytics queries
- Materialized views for expensive aggregations

### Monitoring
- Track DAG run duration
- Monitor task failure rates
- Alert on SLA misses
- Data quality checks

### Security
- Use Airflow Connections/Variables for secrets
- Network isolation for tasks
- VPC peering for database access

### Scalability
- Partition large tables by year (future)
- Horizontal scaling with K8s
- Caching for static data (tournament stats, rankings)
