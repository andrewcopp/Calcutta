package httpserver

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// handleGetTournamentSimStats handles GET /tournaments/{year}/simulations
func (s *Server) handleGetTournamentSimStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}

	stats, err := s.app.MLAnalytics.GetTournamentSimStats(ctx, year)
	if err != nil {
		log.Printf("Error getting tournament sim stats: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	if stats == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Tournament simulations not found", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id": stats.TournamentID,
		"season":        stats.Season,
		"n_sims":        stats.NSims,
		"n_teams":       stats.NTeams,
		"avg_progress":  stats.AvgProgress,
		"max_progress":  stats.MaxProgress,
	})
}

// handleGetTeamPerformance handles GET /tournaments/{year}/teams/{team_id}/performance
func (s *Server) handleGetTeamPerformance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}
	teamID := vars["team_id"]
	if teamID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing team_id parameter", "team_id")
		return
	}

	perf, err := s.app.MLAnalytics.GetTeamPerformance(ctx, year, teamID)
	if err != nil {
		log.Printf("Error getting team performance: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	if perf == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Team performance not found", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"team_id":            perf.TeamID,
		"school_name":        perf.SchoolName,
		"seed":               perf.Seed,
		"region":             perf.Region,
		"kenpom_net":         perf.KenpomNet,
		"total_sims":         perf.TotalSims,
		"avg_wins":           perf.AvgWins,
		"round_distribution": perf.RoundDistribution,
	})
}

// handleGetTeamPerformanceByCalcutta handles GET /api/v1/analytics/calcuttas/{calcutta_id}/teams/{team_id}/performance
func (s *Server) handleGetTeamPerformanceByCalcutta(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["calcutta_id"]
	teamID := vars["team_id"]
	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta_id parameter", "calcutta_id")
		return
	}
	if teamID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing team_id parameter", "team_id")
		return
	}

	query := `
		WITH calcutta AS (
			SELECT
				c.id AS calcutta_id,
				c.tournament_id,
				s.year AS season
			FROM core.calcuttas c
			JOIN core.tournaments tr ON tr.id = c.tournament_id AND tr.deleted_at IS NULL
			JOIN core.seasons s ON s.id = tr.season_id AND s.deleted_at IS NULL
			WHERE c.id = $1
			  AND c.deleted_at IS NULL
			LIMIT 1
		),
		team_ctx AS (
			SELECT
				t.id AS team_id,
				bt.core_tournament_id,
				bt.season
			FROM bronze.teams t
			JOIN bronze.tournaments bt ON bt.id = t.tournament_id
			WHERE t.id = $2::uuid
			LIMIT 1
		),
		valid AS (
			SELECT 1
			FROM calcutta c
			JOIN team_ctx tc ON (tc.core_tournament_id = c.tournament_id OR tc.season = c.season)
			LIMIT 1
		),
		round_distribution AS (
			SELECT
				st.team_id,
				CASE (st.wins + st.byes)
					WHEN 0 THEN 'R64'
					WHEN 1 THEN 'R64'
					WHEN 2 THEN 'R32'
					WHEN 3 THEN 'S16'
					WHEN 4 THEN 'E8'
					WHEN 5 THEN 'F4'
					WHEN 6 THEN 'Finals'
					WHEN 7 THEN 'Champion'
					ELSE 'Unknown'
				END as round_name,
				COUNT(*)::int as count
			FROM silver.simulated_tournaments st
			JOIN bronze.teams t ON t.id = st.team_id
			WHERE st.team_id = $2::uuid
			GROUP BY st.team_id, round_name
		)
		SELECT
			t.id as team_id,
			t.school_name,
			t.seed,
			t.region,
			t.kenpom_net,
			COUNT(DISTINCT st.sim_id)::int as total_sims,
			AVG(st.wins)::float as avg_wins,
			AVG(core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta), st.wins, st.byes))::float as avg_points,
			jsonb_object_agg(rd.round_name, rd.count) as round_distribution
		FROM bronze.teams t
		JOIN valid v ON true
		JOIN silver.simulated_tournaments st ON st.team_id = t.id
		LEFT JOIN round_distribution rd ON rd.team_id = t.id
		WHERE t.id = $2::uuid
		GROUP BY t.id, t.school_name, t.seed, t.region, t.kenpom_net;
	`

	var resp struct {
		TeamID            string
		SchoolName        string
		Seed              *int32
		Region            *string
		KenpomNet         *float64
		TotalSims         int32
		AvgWins           float64
		AvgPoints         float64
		RoundDistribution []byte
	}

	err := s.pool.QueryRow(ctx, query, calcuttaID, teamID).Scan(
		&resp.TeamID,
		&resp.SchoolName,
		&resp.Seed,
		&resp.Region,
		&resp.KenpomNet,
		&resp.TotalSims,
		&resp.AvgWins,
		&resp.AvgPoints,
		&resp.RoundDistribution,
	)
	if err != nil {
		log.Printf("Error getting team performance by calcutta: %v", err)
		writeError(w, r, http.StatusNotFound, "not_found", "Team performance not found", "")
		return
	}

	var roundDist map[string]int
	if err := json.Unmarshal(resp.RoundDistribution, &roundDist); err != nil {
		log.Printf("Error parsing round distribution: %v", err)
		writeError(w, r, http.StatusInternalServerError, "internal_error", "Failed to parse round distribution", "")
		return
	}

	seed := 0
	if resp.Seed != nil {
		seed = int(*resp.Seed)
	}
	region := ""
	if resp.Region != nil {
		region = *resp.Region
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":        calcuttaID,
		"team_id":            resp.TeamID,
		"school_name":        resp.SchoolName,
		"seed":               seed,
		"region":             region,
		"kenpom_net":         resp.KenpomNet,
		"total_sims":         int(resp.TotalSims),
		"avg_wins":           resp.AvgWins,
		"avg_points":         resp.AvgPoints,
		"round_distribution": roundDist,
	})
}

// handleGetTeamPredictions handles GET /tournaments/{year}/teams/predictions
func (s *Server) handleGetTeamPredictions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}

	// Optional run_id query parameter
	var runID *string
	if rid := r.URL.Query().Get("run_id"); rid != "" {
		runID = &rid
	}

	predictions, err := s.app.MLAnalytics.GetTeamPredictions(ctx, year, runID)
	if err != nil {
		log.Printf("Error getting team predictions: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	teams := make([]map[string]interface{}, len(predictions))
	for i, pred := range predictions {
		teams[i] = map[string]interface{}{
			"team_id":     pred.TeamID,
			"school_name": pred.SchoolName,
			"seed":        pred.Seed,
			"region":      pred.Region,
			"kenpom_net":  pred.KenpomNet,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"year":  year,
		"teams": teams,
	})
}

// handleGetOurEntryDetails handles GET /tournaments/{year}/runs/{run_id}/our-entry
func (s *Server) handleGetOurEntryDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}
	runID := vars["run_id"]

	details, err := s.app.MLAnalytics.GetOurEntryDetails(ctx, year, runID)
	if err != nil {
		log.Printf("Error getting our entry details: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	if details == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Entry details not found", "")
		return
	}

	portfolio := make([]map[string]interface{}, len(details.Portfolio))
	for i, bid := range details.Portfolio {
		portfolio[i] = map[string]interface{}{
			"team_id":                bid.TeamID,
			"school_name":            bid.SchoolName,
			"seed":                   bid.Seed,
			"region":                 bid.Region,
			"recommended_bid_points": bid.RecommendedBidPoints,
			"expected_roi":           bid.ExpectedROI,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"run": map[string]interface{}{
			"run_id":        details.Run.RunID,
			"calcutta_id":   details.Run.CalcuttaID,
			"strategy":      details.Run.Strategy,
			"n_sims":        details.Run.NSims,
			"seed":          details.Run.Seed,
			"budget_points": details.Run.BudgetPoints,
			"created_at":    details.Run.CreatedAt,
		},
		"portfolio": portfolio,
		"summary": map[string]interface{}{
			"mean_normalized_payout": details.Summary.MeanNormalizedPayout,
			"p_top1":                 details.Summary.PTop1,
			"p_in_money":             details.Summary.PInMoney,
			"percentile_rank":        details.Summary.PercentileRank,
		},
	})
}

// handleGetEntryRankings handles GET /tournaments/{year}/runs/{run_id}/rankings
func (s *Server) handleGetEntryRankings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}
	runID := vars["run_id"]

	// Parse pagination parameters
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	rankings, err := s.app.MLAnalytics.GetEntryRankings(ctx, year, runID, limit, offset)
	if err != nil {
		log.Printf("Error getting entry rankings: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	entries := make([]map[string]interface{}, len(rankings))
	for i, rank := range rankings {
		entries[i] = map[string]interface{}{
			"rank":                   rank.Rank,
			"entry_key":              rank.EntryKey,
			"is_our_strategy":        rank.IsOurStrategy,
			"n_teams":                rank.NTeams,
			"total_bid_points":       rank.TotalBidPoints,
			"mean_normalized_payout": rank.MeanNormalizedPayout,
			"percentile_rank":        rank.PercentileRank,
			"p_top1":                 rank.PTop1,
			"p_in_money":             rank.PInMoney,
		}
	}

	var totalEntries int
	if len(rankings) > 0 {
		totalEntries = rankings[0].TotalEntries
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"run_id":        runID,
		"total_entries": totalEntries,
		"limit":         limit,
		"offset":        offset,
		"entries":       entries,
	})
}

// handleGetEntrySimulations handles GET /tournaments/{year}/runs/{run_id}/entries/{entry_key}/simulations
func (s *Server) handleGetEntrySimulations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}
	runID := vars["run_id"]
	entryKey := vars["entry_key"]

	// Parse pagination parameters
	limit := 100
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	offset := 0
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil {
			offset = parsed
		}
	}

	drillDown, err := s.app.MLAnalytics.GetEntrySimulations(ctx, year, runID, entryKey, limit, offset)
	if err != nil {
		log.Printf("Error getting entry simulations: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	if drillDown == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Entry simulations not found", "")
		return
	}

	simulations := make([]map[string]interface{}, len(drillDown.Simulations))
	for i, sim := range drillDown.Simulations {
		simulations[i] = map[string]interface{}{
			"sim_id":            sim.SimID,
			"payout_cents":      sim.PayoutCents,
			"total_points":      sim.TotalPoints,
			"finish_position":   sim.FinishPosition,
			"is_tied":           sim.IsTied,
			"normalized_payout": sim.NormalizedPayout,
			"n_entries":         sim.NEntries,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"entry_key": entryKey,
		"run_id":    runID,
		"summary": map[string]interface{}{
			"total_simulations":      drillDown.Summary.TotalSimulations,
			"mean_payout_cents":      drillDown.Summary.MeanPayoutCents,
			"mean_points":            drillDown.Summary.MeanPoints,
			"mean_normalized_payout": drillDown.Summary.MeanNormalizedPayout,
			"p50_payout_cents":       drillDown.Summary.P50PayoutCents,
			"p90_payout_cents":       drillDown.Summary.P90PayoutCents,
		},
		"limit":       limit,
		"offset":      offset,
		"simulations": simulations,
	})
}

// handleGetEntryPortfolio handles GET /tournaments/{year}/runs/{run_id}/entries/{entry_key}/portfolio
func (s *Server) handleGetEntryPortfolio(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}
	runID := vars["run_id"]
	entryKey := vars["entry_key"]

	portfolio, err := s.app.MLAnalytics.GetEntryPortfolio(ctx, year, runID, entryKey)
	if err != nil {
		log.Printf("Error getting entry portfolio: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	if portfolio == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Entry portfolio not found", "")
		return
	}

	teams := make([]map[string]interface{}, len(portfolio.Teams))
	for i, team := range portfolio.Teams {
		teams[i] = map[string]interface{}{
			"team_id":           team.TeamID,
			"school_name":       team.SchoolName,
			"seed":              team.Seed,
			"region":            team.Region,
			"bid_amount_points": team.BidAmountPoints,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"entry_key": portfolio.EntryKey,
		"teams":     teams,
		"total_bid": portfolio.TotalBid,
		"n_teams":   portfolio.NTeams,
	})
}

// handleGetOptimizationRuns handles GET /tournaments/{year}/runs
func (s *Server) handleGetOptimizationRuns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}

	runs, err := s.app.MLAnalytics.GetOptimizationRuns(ctx, year)
	if err != nil {
		log.Printf("Error getting optimization runs: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	runsData := make([]map[string]interface{}, len(runs))
	for i, run := range runs {
		runsData[i] = map[string]interface{}{
			"run_id":        run.RunID,
			"calcutta_id":   run.CalcuttaID,
			"strategy":      run.Strategy,
			"n_sims":        run.NSims,
			"seed":          run.Seed,
			"budget_points": run.BudgetPoints,
			"created_at":    run.CreatedAt,
		}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"year": year,
		"runs": runsData,
	})
}
