# Analytics API Endpoints

Go API endpoints for the Calcutta analytics dashboard.

## Base URL
```
/api/v1/analytics
```

## Endpoints

### 1. Tournament Simulations Overview

**GET** `/tournaments/{year}/simulations`

Get statistics about tournament simulations for a given year.

**Response:**
```json
{
  "tournament_key": "ncaa-tournament-2025",
  "season": 2025,
  "n_sims": 5000,
  "n_teams": 64,
  "avg_progress": 2.5,
  "max_progress": 6,
  "completion_rate": 1.0
}
```

---

### 2. Team Performance in Simulations

**GET** `/tournaments/{year}/teams/{team_key}/performance`

Get detailed performance statistics for a specific team across all simulations.

**Response:**
```json
{
  "team_key": "ncaa-tournament-2025:duke",
  "school_name": "Duke",
  "seed": 1,
  "region": "East",
  "kenpom_net": 38.15,
  "simulations": {
    "total": 5000,
    "avg_wins": 3.2,
    "avg_points": 425.5,
    "round_distribution": {
      "R64": 0,
      "R32": 150,
      "S16": 800,
      "E8": 1200,
      "F4": 1500,
      "Finals": 900,
      "Champion": 450
    }
  },
  "probabilities": {
    "p_champion": 0.090,
    "p_finals": 0.180,
    "p_final_four": 0.300,
    "p_elite_eight": 0.540,
    "p_sweet_sixteen": 0.760,
    "p_round_32": 0.970
  }
}
```

---

### 3. Team Predictions and Investment

**GET** `/tournaments/{year}/teams/predictions`

Get predicted performance and investment metrics for all teams.

**Query Parameters:**
- `run_id` (optional): Specific optimization run ID

**Response:**
```json
{
  "teams": [
    {
      "team_key": "ncaa-tournament-2025:duke",
      "school_name": "Duke",
      "seed": 1,
      "region": "East",
      "expected_points": 425.5,
      "predicted_market_share": 0.085,
      "predicted_market_points": 425.0,
      "p_champion": 0.090,
      "kenpom_net": 38.15
    }
  ]
}
```

---

### 4. Our Entry Details

**GET** `/tournaments/{year}/runs/{run_id}/our-entry`

Get detailed information about our optimized entry.

**Response:**
```json
{
  "run_id": "20251230T215837Z",
  "strategy": "minlp",
  "n_sims": 5000,
  "seed": 42,
  "budget_points": 100,
  "run_timestamp": "2025-12-30T21:58:37Z",
  "portfolio": [
    {
      "team_key": "ncaa-tournament-2025:duke",
      "school_name": "Duke",
      "seed": 1,
      "region": "East",
      "bid_amount_points": 38,
      "expected_points": 425.5,
      "predicted_market_points": 425.0,
      "actual_market_points": 450.0,
      "our_ownership": 0.078,
      "expected_roi": 1.001,
      "our_roi": 0.947,
      "roi_degradation": -0.054
    }
  ],
  "summary": {
    "total_teams": 10,
    "total_bid_points": 100,
    "mean_normalized_payout": 0.033,
    "p_top1": 0.002,
    "p_in_money": 0.125,
    "percentile_rank": 2.1
  }
}
```

---

### 5. Entry Rankings

**GET** `/tournaments/{year}/runs/{run_id}/rankings`

Get all entries ranked by normalized payout.

**Query Parameters:**
- `limit` (optional): Number of entries to return (default: all)
- `offset` (optional): Pagination offset

**Response:**
```json
{
  "run_id": "20251230T215837Z",
  "total_entries": 48,
  "entries": [
    {
      "rank": 1,
      "entry_key": "calcutta-2025:entry-andrew-copp",
      "is_our_strategy": false,
      "n_teams": 10,
      "total_bid_points": 100,
      "mean_normalized_payout": 0.578,
      "percentile_rank": 100.0,
      "p_top1": 0.234,
      "p_in_money": 0.876
    }
  ]
}
```

---

### 6. Entry Simulation Drill-Down

**GET** `/tournaments/{year}/runs/{run_id}/entries/{entry_key}/simulations`

Get all simulation outcomes for a specific entry, sorted by payout.

**Query Parameters:**
- `limit` (optional): Number of simulations to return (default: 100)
- `offset` (optional): Pagination offset

**Response:**
```json
{
  "entry_key": "calcutta-2025:entry-andrew-copp",
  "run_id": "20251230T215837Z",
  "total_simulations": 5000,
  "simulations": [
    {
      "sim_id": 1234,
      "payout_cents": 60000,
      "total_points": 1050,
      "finish_position": 1,
      "is_tied": false,
      "normalized_payout": 1.0,
      "n_entries": 48
    },
    {
      "sim_id": 5678,
      "payout_cents": 30000,
      "total_points": 875,
      "finish_position": 2,
      "is_tied": false,
      "normalized_payout": 0.5,
      "n_entries": 48
    }
  ],
  "summary": {
    "mean_payout_cents": 20060,
    "mean_points": 425.5,
    "mean_normalized_payout": 0.578,
    "p50_payout_cents": 15000,
    "p90_payout_cents": 45000
  }
}
```

---

### 7. Entry Portfolio

**GET** `/tournaments/{year}/runs/{run_id}/entries/{entry_key}/portfolio`

Get the team composition for any entry (including our strategy).

**Response:**
```json
{
  "entry_key": "calcutta-2025:entry-andrew-copp",
  "teams": [
    {
      "team_key": "ncaa-tournament-2025:auburn",
      "school_name": "Auburn",
      "seed": 1,
      "region": "South",
      "bid_amount": 26
    },
    {
      "team_key": "ncaa-tournament-2025:houston",
      "school_name": "Houston",
      "seed": 1,
      "region": "Midwest",
      "bid_amount": 26
    }
  ],
  "total_bid": 100,
  "n_teams": 10
}
```

---

### 8. Available Optimization Runs

**GET** `/tournaments/{year}/runs`

Get all optimization runs for a given year.

**Response:**
```json
{
  "runs": [
    {
      "run_id": "20251230T215837Z",
      "strategy": "minlp",
      "n_sims": 5000,
      "seed": 42,
      "run_timestamp": "2025-12-30T21:58:37Z"
    },
    {
      "run_id": "20251230T193745Z",
      "strategy": "greedy",
      "n_sims": 5000,
      "seed": 42,
      "run_timestamp": "2025-12-30T19:37:45Z"
    }
  ]
}
```

---

## Error Responses

All endpoints return standard error responses:

```json
{
  "error": "tournament not found",
  "code": "NOT_FOUND",
  "status": 404
}
```

**Common Error Codes:**
- `400` - Bad Request (invalid parameters)
- `404` - Not Found (tournament/entry/run doesn't exist)
- `500` - Internal Server Error

---

## Implementation Notes

### Database Queries

Most endpoints will use the views and functions defined in the schema:
- `view_entry_rankings` - For rankings endpoint
- `view_tournament_sim_stats` - For tournament overview
- `get_entry_portfolio()` - For portfolio endpoint

### Performance Considerations

**Pagination**: All list endpoints support `limit` and `offset` for pagination.

**Caching**: Consider caching expensive aggregations:
- Tournament simulation stats (rarely changes)
- Entry rankings (static per run)
- Team performance distributions

**Indexes**: Ensure indexes exist on:
- `(run_id, entry_key)` for drill-down queries
- `(tournament_key, team_key)` for team lookups
- `(run_id, percentile_rank)` for rankings

### Read Replicas

For production, consider routing analytics queries to read replicas to avoid impacting the main application database.
