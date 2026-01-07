package httpserver

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

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

	percentileRank := 0.0
	if details.Summary.PercentileRank != nil {
		percentileRank = *details.Summary.PercentileRank
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"run": map[string]interface{}{
			"run_id":        details.Run.RunID,
			"name":          details.Run.Name,
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
			"percentile_rank":        percentileRank,
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
			"name":          run.Name,
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
