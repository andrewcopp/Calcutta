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

	perf, err := s.app.MLAnalytics.GetTeamPerformanceByCalcutta(ctx, calcuttaID, teamID)
	if err != nil {
		log.Printf("Error getting team performance by calcutta: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to get team performance", "")
		return
	}
	if perf == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "Team performance not found", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":        calcuttaID,
		"team_id":            perf.TeamID,
		"school_name":        perf.SchoolName,
		"seed":               perf.Seed,
		"region":             perf.Region,
		"kenpom_net":         perf.KenpomNet,
		"total_sims":         perf.TotalSims,
		"avg_wins":           perf.AvgWins,
		"avg_points":         perf.AvgPoints,
		"round_distribution": perf.RoundDistribution,
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
			"team_id":      bid.TeamID,
			"school_name":  bid.SchoolName,
			"seed":         bid.Seed,
			"region":       bid.Region,
			"bid_points":   bid.BidPoints,
			"expected_roi": bid.ExpectedROI,
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
	limit := getLimit(r, 100)
	offset := getOffset(r, 0)

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
			"team_id":     team.TeamID,
			"school_name": team.SchoolName,
			"seed":        team.Seed,
			"region":      team.Region,
			"bid_points":  team.BidPoints,
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

func (s *Server) handleListTournamentSimulationBatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing tournament ID", "id")
		return
	}

	batches, err := s.app.MLAnalytics.ListTournamentSimulationBatchesByCoreTournamentID(ctx, tournamentID)
	if err != nil {
		log.Printf("Error listing tournament simulation batches: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	data := make([]map[string]interface{}, 0, len(batches))
	for _, b := range batches {
		data = append(data, map[string]interface{}{
			"id":                           b.ID,
			"tournament_id":                b.TournamentID,
			"tournament_state_snapshot_id": b.SimulationStateID,
			"n_sims":                       b.NSims,
			"seed":                         b.Seed,
			"probability_source_key":       b.ProbabilitySourceKey,
			"created_at":                   b.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id": tournamentID,
		"batches":       data,
		"count":         len(data),
	})
}

func (s *Server) handleListCalcuttaEvaluationRuns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	runs, err := s.app.MLAnalytics.ListCalcuttaEvaluationRunsByCoreCalcuttaID(ctx, calcuttaID)
	if err != nil {
		log.Printf("Error listing calcutta evaluation runs: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	data := make([]map[string]interface{}, 0, len(runs))
	for _, run := range runs {
		data = append(data, map[string]interface{}{
			"id":                             run.ID,
			"tournament_simulation_batch_id": run.SimulatedTournamentID,
			"calcutta_snapshot_id":           run.CalcuttaSnapshotID,
			"purpose":                        run.Purpose,
			"created_at":                     run.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id": calcuttaID,
		"runs":        data,
		"count":       len(data),
	})
}

func (s *Server) handleListStrategyGenerationRuns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	runs, err := s.app.MLAnalytics.ListStrategyGenerationRunsByCoreCalcuttaID(ctx, calcuttaID)
	if err != nil {
		log.Printf("Error listing strategy generation runs: %v", err)
		writeErrorFromErr(w, r, err)
		return
	}

	data := make([]map[string]interface{}, 0, len(runs))
	for _, run := range runs {
		var params interface{} = nil
		if len(run.ParamsJSON) > 0 {
			params = json.RawMessage(run.ParamsJSON)
		}

		data = append(data, map[string]interface{}{
			"id":                             run.ID,
			"run_key":                        run.RunKey,
			"tournament_simulation_batch_id": run.SimulatedTournamentID,
			"calcutta_id":                    run.CalcuttaID,
			"purpose":                        run.Purpose,
			"returns_model_key":              run.ReturnsModelKey,
			"investment_model_key":           run.InvestmentModelKey,
			"optimizer_key":                  run.OptimizerKey,
			"params_json":                    params,
			"git_sha":                        run.GitSHA,
			"created_at":                     run.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id": calcuttaID,
		"runs":        data,
		"count":       len(data),
	})
}
