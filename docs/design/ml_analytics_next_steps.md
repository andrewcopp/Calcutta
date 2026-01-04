# ML Analytics - Next Steps

## ✅ Completed
- Database migrations created and applied
- sqlc queries written and generated
- Repository layer implemented with type conversions
- Service layer implemented
- HTTP handlers implemented (using gorilla/mux)
- Python database writers created (bronze/silver/gold layers)

## 🔧 Remaining Integration Steps

### 1. Wire Up Service in App

**File**: `internal/app/app.go`

Add the MLAnalytics service to the App struct:

```go
type App struct {
    // ... existing services
    MLAnalytics *ml_analytics.Service
}
```

**File**: `internal/app/bootstrap/bootstrap.go`

Initialize the service in the bootstrap function:

```go
import (
    "github.com/andrewcopp/Calcutta/backend/internal/app/ml_analytics"
    // ... other imports
)

func Bootstrap(pool *pgxpool.Pool) *app.App {
    // ... existing code
    
    // ML Analytics
    mlAnalyticsRepo := db.NewMLAnalyticsRepository(pool)
    mlAnalyticsService := ml_analytics.New(mlAnalyticsRepo)
    
    return &app.App{
        // ... existing services
        MLAnalytics: mlAnalyticsService,
    }
}
```

### 2. Add Routes

**File**: `internal/transport/httpserver/router.go`

Add the analytics routes (adjust auth middleware as needed):

```go
// ML Analytics routes (read-only)
r.HandleFunc("/api/v1/analytics/tournaments/{year}/simulations", s.handleGetTournamentSimStats).Methods("GET")
r.HandleFunc("/api/v1/analytics/tournaments/{year}/teams/{team_key}/performance", s.handleGetTeamPerformance).Methods("GET")
r.HandleFunc("/api/v1/analytics/tournaments/{year}/teams/predictions", s.handleGetTeamPredictions).Methods("GET")
r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs", s.handleGetOptimizationRuns).Methods("GET")
r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs/{run_id}/our-entry", s.handleGetOurEntryDetails).Methods("GET")
r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs/{run_id}/rankings", s.handleGetEntryRankings).Methods("GET")
r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs/{run_id}/entries/{entry_key}/simulations", s.handleGetEntrySimulations).Methods("GET")
r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs/{run_id}/entries/{entry_key}/portfolio", s.handleGetEntryPortfolio).Methods("GET")
```

### 3. Test Compilation

```bash
cd backend
go build -o /tmp/test-build ./cmd/api
```

If successful, test with:
```bash
make up
```

## 📝 Notes

### sqlc Parameter Names
Some queries use `Column1`, `Column2`, etc. instead of named parameters because we used `$1`, `$2` syntax. This is fine and works correctly.

### Type Conversions
The repository includes helper functions to convert `pgtype.Numeric` to `float64`:
- `floatFromPgNumeric()` - for non-nullable fields
- `floatPtrFromPgNumeric()` - for nullable fields

### Portfolio Query
The `GetEntryPortfolio` method handles two cases:
- `entryKey == "our_strategy"` - uses `GetEntryPortfolio` query
- Other entries - uses `GetActualEntryPortfolio` query

## 🐍 Python Integration

Once the Go API is working, update your Python scripts to use the database writers:

```python
from moneyball.db.writers import (
    write_simulated_tournaments,
    write_predicted_game_outcomes,
    write_optimization_run,
    write_recommended_entry_bids,
)

# After running simulations
write_simulated_tournaments(tournament_key, simulations_df)

# After ML predictions
write_predicted_game_outcomes(tournament_key, predictions_df)

# After optimization
write_optimization_run(run_id, calcutta_key, strategy, n_sims, seed)
write_recommended_entry_bids(run_id, bids_df)
```

## 🧪 Testing

Once integrated, test the API endpoints:

```bash
# Get tournament simulations
curl http://localhost:8080/api/v1/analytics/tournaments/2025/simulations

# Get team performance
curl http://localhost:8080/api/v1/analytics/tournaments/2025/teams/ncaa-tournament-2025:duke/performance

# Get optimization runs
curl http://localhost:8080/api/v1/analytics/tournaments/2025/runs
```

## 🎯 Success Criteria

- [ ] App compiles without errors
- [ ] `make up` starts successfully
- [ ] API endpoints return 404 (no data yet) or valid JSON
- [ ] Python writers successfully insert data
- [ ] API endpoints return populated data after Python writes
