# Go Database Models for Analytics

Go struct definitions for analytics tables. These should be added to the existing Go service's database models.

## Package Structure

```
internal/
  models/
    analytics/
      bronze.go      # Bronze layer models
      silver.go      # Silver layer models  
      gold.go        # Gold layer models
      queries.go     # Query helpers
```

## Bronze Layer Models

```go
// bronze.go
package analytics

import "time"

type Tournament struct {
    TournamentKey string    `db:"tournament_key" json:"tournament_key"`
    Season        int       `db:"season" json:"season"`
    TournamentName string   `db:"tournament_name" json:"tournament_name"`
    CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type Team struct {
    TeamKey      string     `db:"team_key" json:"team_key"`
    TournamentKey string    `db:"tournament_key" json:"tournament_key"`
    SchoolSlug   string     `db:"school_slug" json:"school_slug"`
    SchoolName   string     `db:"school_name" json:"school_name"`
    Seed         int        `db:"seed" json:"seed"`
    Region       string     `db:"region" json:"region"`
    Byes         int        `db:"byes" json:"byes"`
    KenpomNet    *float64   `db:"kenpom_net" json:"kenpom_net,omitempty"`
    KenpomO      *float64   `db:"kenpom_o" json:"kenpom_o,omitempty"`
    KenpomD      *float64   `db:"kenpom_d" json:"kenpom_d,omitempty"`
    KenpomAdjT   *float64   `db:"kenpom_adj_t" json:"kenpom_adj_t,omitempty"`
    CreatedAt    time.Time  `db:"created_at" json:"created_at"`
}

type SimulatedTournament struct {
    ID            int64     `db:"id" json:"id"`
    TournamentKey string    `db:"tournament_key" json:"tournament_key"`
    SimID         int       `db:"sim_id" json:"sim_id"`
    TeamKey       string    `db:"team_key" json:"team_key"`
    Wins          int       `db:"wins" json:"wins"`
    Byes          int       `db:"byes" json:"byes"`
    Eliminated    bool      `db:"eliminated" json:"eliminated"`
    CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type Calcutta struct {
    CalcuttaKey   string    `db:"calcutta_key" json:"calcutta_key"`
    TournamentKey string    `db:"tournament_key" json:"tournament_key"`
    CalcuttaName  string    `db:"calcutta_name" json:"calcutta_name"`
    BudgetPoints  int       `db:"budget_points" json:"budget_points"`
    CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type EntryBid struct {
    ID          int64     `db:"id" json:"id"`
    CalcuttaKey string    `db:"calcutta_key" json:"calcutta_key"`
    EntryKey    string    `db:"entry_key" json:"entry_key"`
    TeamKey     string    `db:"team_key" json:"team_key"`
    BidAmount   int       `db:"bid_amount" json:"bid_amount"`
    CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type Payout struct {
    ID          int       `db:"id" json:"id"`
    CalcuttaKey string    `db:"calcutta_key" json:"calcutta_key"`
    Position    int       `db:"position" json:"position"`
    AmountCents int       `db:"amount_cents" json:"amount_cents"`
    CreatedAt   time.Time `db:"created_at" json:"created_at"`
}
```

## Silver Layer Models

```go
// silver.go
package analytics

import "time"

type PredictedGameOutcome struct {
    ID            int64     `db:"id" json:"id"`
    TournamentKey string    `db:"tournament_key" json:"tournament_key"`
    GameID        string    `db:"game_id" json:"game_id"`
    Round         int       `db:"round" json:"round"`
    Team1Key      string    `db:"team1_key" json:"team1_key"`
    Team2Key      string    `db:"team2_key" json:"team2_key"`
    PTeam1Wins    float64   `db:"p_team1_wins" json:"p_team1_wins"`
    PMatchup      float64   `db:"p_matchup" json:"p_matchup"`
    ModelVersion  *string   `db:"model_version" json:"model_version,omitempty"`
    CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type PredictedMarketShare struct {
    ID                    int64     `db:"id" json:"id"`
    CalcuttaKey           string    `db:"calcutta_key" json:"calcutta_key"`
    TeamKey               string    `db:"team_key" json:"team_key"`
    PredictedShareOfPool  float64   `db:"predicted_share_of_pool" json:"predicted_share_of_pool"`
    ModelVersion          *string   `db:"model_version" json:"model_version,omitempty"`
    CreatedAt             time.Time `db:"created_at" json:"created_at"`
}

type TeamTournamentValue struct {
    ID             int64     `db:"id" json:"id"`
    TournamentKey  string    `db:"tournament_key" json:"tournament_key"`
    TeamKey        string    `db:"team_key" json:"team_key"`
    ExpectedPoints float64   `db:"expected_points" json:"expected_points"`
    PChampion      *float64  `db:"p_champion" json:"p_champion,omitempty"`
    PFinals        *float64  `db:"p_finals" json:"p_finals,omitempty"`
    PFinalFour     *float64  `db:"p_final_four" json:"p_final_four,omitempty"`
    PEliteEight    *float64  `db:"p_elite_eight" json:"p_elite_eight,omitempty"`
    PSweetSixteen  *float64  `db:"p_sweet_sixteen" json:"p_sweet_sixteen,omitempty"`
    PRound32       *float64  `db:"p_round_32" json:"p_round_32,omitempty"`
    CreatedAt      time.Time `db:"created_at" json:"created_at"`
}
```

## Gold Layer Models

```go
// gold.go
package analytics

import "time"

type OptimizationRun struct {
    RunID         string    `db:"run_id" json:"run_id"`
    CalcuttaKey   string    `db:"calcutta_key" json:"calcutta_key"`
    Strategy      string    `db:"strategy" json:"strategy"`
    NSims         int       `db:"n_sims" json:"n_sims"`
    Seed          int       `db:"seed" json:"seed"`
    BudgetPoints  int       `db:"budget_points" json:"budget_points"`
    RunTimestamp  time.Time `db:"run_timestamp" json:"run_timestamp"`
    CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type RecommendedEntryBid struct {
    ID              int64     `db:"id" json:"id"`
    RunID           string    `db:"run_id" json:"run_id"`
    TeamKey         string    `db:"team_key" json:"team_key"`
    BidAmountPoints int       `db:"bid_amount_points" json:"bid_amount_points"`
    CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

type EntrySimulationOutcome struct {
    ID                 int64     `db:"id" json:"id"`
    RunID              string    `db:"run_id" json:"run_id"`
    EntryKey           string    `db:"entry_key" json:"entry_key"`
    SimID              int       `db:"sim_id" json:"sim_id"`
    PayoutCents        int       `db:"payout_cents" json:"payout_cents"`
    TotalPoints        float64   `db:"total_points" json:"total_points"`
    FinishPosition     int       `db:"finish_position" json:"finish_position"`
    IsTied             bool      `db:"is_tied" json:"is_tied"`
    NEntries           int       `db:"n_entries" json:"n_entries"`
    NormalizedPayout   float64   `db:"normalized_payout" json:"normalized_payout"`
    CreatedAt          time.Time `db:"created_at" json:"created_at"`
}

type EntryPerformance struct {
    ID                     int64     `db:"id" json:"id"`
    RunID                  string    `db:"run_id" json:"run_id"`
    EntryKey               string    `db:"entry_key" json:"entry_key"`
    IsOurStrategy          bool      `db:"is_our_strategy" json:"is_our_strategy"`
    NTeams                 int       `db:"n_teams" json:"n_teams"`
    TotalBidPoints         int       `db:"total_bid_points" json:"total_bid_points"`
    MeanPayoutCents        float64   `db:"mean_payout_cents" json:"mean_payout_cents"`
    MeanPoints             float64   `db:"mean_points" json:"mean_points"`
    MeanNormalizedPayout   float64   `db:"mean_normalized_payout" json:"mean_normalized_payout"`
    P50NormalizedPayout    float64   `db:"p50_normalized_payout" json:"p50_normalized_payout"`
    P90NormalizedPayout    float64   `db:"p90_normalized_payout" json:"p90_normalized_payout"`
    PTop1                  float64   `db:"p_top1" json:"p_top1"`
    PInMoney               float64   `db:"p_in_money" json:"p_in_money"`
    PercentileRank         *float64  `db:"percentile_rank" json:"percentile_rank,omitempty"`
    CreatedAt              time.Time `db:"created_at" json:"created_at"`
}

type DetailedInvestmentReport struct {
    ID                    int64     `db:"id" json:"id"`
    RunID                 string    `db:"run_id" json:"run_id"`
    TeamKey               string    `db:"team_key" json:"team_key"`
    OurBidPoints          int       `db:"our_bid_points" json:"our_bid_points"`
    ExpectedPoints        float64   `db:"expected_points" json:"expected_points"`
    PredictedMarketPoints float64   `db:"predicted_market_points" json:"predicted_market_points"`
    ActualMarketPoints    float64   `db:"actual_market_points" json:"actual_market_points"`
    OurOwnership          float64   `db:"our_ownership" json:"our_ownership"`
    ExpectedROI           float64   `db:"expected_roi" json:"expected_roi"`
    OurROI                float64   `db:"our_roi" json:"our_roi"`
    ROIDegradation        float64   `db:"roi_degradation" json:"roi_degradation"`
    CreatedAt             time.Time `db:"created_at" json:"created_at"`
}
```

## Query Helpers

```go
// queries.go
package analytics

import (
    "context"
    "database/sql"
)

// TournamentSimStats represents aggregated simulation statistics
type TournamentSimStats struct {
    TournamentKey string  `db:"tournament_key" json:"tournament_key"`
    Season        int     `db:"season" json:"season"`
    NSims         int     `db:"n_sims" json:"n_sims"`
    NTeams        int     `db:"n_teams" json:"n_teams"`
    AvgProgress   float64 `db:"avg_progress" json:"avg_progress"`
    MaxProgress   int     `db:"max_progress" json:"max_progress"`
}

// EntryRanking represents an entry's ranking
type EntryRanking struct {
    RunID                string  `db:"run_id" json:"run_id"`
    EntryKey             string  `db:"entry_key" json:"entry_key"`
    IsOurStrategy        bool    `db:"is_our_strategy" json:"is_our_strategy"`
    NTeams               int     `db:"n_teams" json:"n_teams"`
    TotalBidPoints       int     `db:"total_bid_points" json:"total_bid_points"`
    MeanNormalizedPayout float64 `db:"mean_normalized_payout" json:"mean_normalized_payout"`
    PercentileRank       float64 `db:"percentile_rank" json:"percentile_rank"`
    PTop1                float64 `db:"p_top1" json:"p_top1"`
    PInMoney             float64 `db:"p_in_money" json:"p_in_money"`
    Rank                 int     `db:"rank" json:"rank"`
    TotalEntries         int     `db:"total_entries" json:"total_entries"`
}

// GetTournamentSimStats retrieves simulation statistics for a tournament
func GetTournamentSimStats(ctx context.Context, db *sql.DB, year int) (*TournamentSimStats, error) {
    var stats TournamentSimStats
    err := db.QueryRowContext(ctx, `
        SELECT * FROM view_tournament_sim_stats
        WHERE season = $1
    `, year).Scan(
        &stats.TournamentKey,
        &stats.Season,
        &stats.NSims,
        &stats.NTeams,
        &stats.AvgProgress,
        &stats.MaxProgress,
    )
    return &stats, err
}

// GetEntryRankings retrieves entry rankings for a run
func GetEntryRankings(ctx context.Context, db *sql.DB, runID string, limit, offset int) ([]EntryRanking, error) {
    rows, err := db.QueryContext(ctx, `
        SELECT * FROM view_entry_rankings
        WHERE run_id = $1
        ORDER BY rank
        LIMIT $2 OFFSET $3
    `, runID, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var rankings []EntryRanking
    for rows.Next() {
        var r EntryRanking
        err := rows.Scan(
            &r.RunID,
            &r.EntryKey,
            &r.IsOurStrategy,
            &r.NTeams,
            &r.TotalBidPoints,
            &r.MeanNormalizedPayout,
            &r.PercentileRank,
            &r.PTop1,
            &r.PInMoney,
            &r.Rank,
            &r.TotalEntries,
        )
        if err != nil {
            return nil, err
        }
        rankings = append(rankings, r)
    }
    return rankings, rows.Err()
}

// GetEntryPortfolio retrieves portfolio for an entry
func GetEntryPortfolio(ctx context.Context, db *sql.DB, runID, entryKey string) ([]struct {
    TeamKey    string `db:"team_key" json:"team_key"`
    SchoolName string `db:"school_name" json:"school_name"`
    Seed       int    `db:"seed" json:"seed"`
    Region     string `db:"region" json:"region"`
    BidAmount  int    `db:"bid_amount" json:"bid_amount"`
}, error) {
    rows, err := db.QueryContext(ctx, `
        SELECT * FROM get_entry_portfolio($1, $2)
    `, runID, entryKey)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var portfolio []struct {
        TeamKey    string `db:"team_key" json:"team_key"`
        SchoolName string `db:"school_name" json:"school_name"`
        Seed       int    `db:"seed" json:"seed"`
        Region     string `db:"region" json:"region"`
        BidAmount  int    `db:"bid_amount" json:"bid_amount"`
    }
    
    for rows.Next() {
        var p struct {
            TeamKey    string `db:"team_key" json:"team_key"`
            SchoolName string `db:"school_name" json:"school_name"`
            Seed       int    `db:"seed" json:"seed"`
            Region     string `db:"region" json:"region"`
            BidAmount  int    `db:"bid_amount" json:"bid_amount"`
        }
        err := rows.Scan(&p.TeamKey, &p.SchoolName, &p.Seed, &p.Region, &p.BidAmount)
        if err != nil {
            return nil, err
        }
        portfolio = append(portfolio, p)
    }
    return portfolio, rows.Err()
}
```

## Usage in Handlers

```go
// Example handler
func (h *Handler) GetEntryRankings(w http.ResponseWriter, r *http.Request) {
    runID := chi.URLParam(r, "runID")
    
    limit := 100
    offset := 0
    // Parse query params...
    
    rankings, err := analytics.GetEntryRankings(r.Context(), h.db, runID, limit, offset)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    json.NewEncoder(w).Encode(map[string]interface{}{
        "run_id": runID,
        "total_entries": len(rankings),
        "entries": rankings,
    })
}
```

## Migration Notes

These models should be added to the existing Go service alongside the current models. The database migrations will create these tables, and the Go service will handle both reading and writing to the analytics tables as specified in the table write responsibilities.
