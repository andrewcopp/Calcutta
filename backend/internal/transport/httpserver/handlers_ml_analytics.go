package httpserver

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
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

func (s *Server) handleGetGameOutcomesAlgorithmCoverage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := s.pool.Query(ctx, `
		WITH total AS (
			SELECT COUNT(*)::int AS total
			FROM core.tournaments
			WHERE deleted_at IS NULL
		), covered AS (
			SELECT algorithm_id, COUNT(DISTINCT tournament_id)::int AS covered
			FROM derived.game_outcome_runs
			WHERE deleted_at IS NULL
			GROUP BY algorithm_id
		)
		SELECT
			a.id::text,
			a.name,
			a.description,
			COALESCE(c.covered, 0)::int AS covered,
			t.total::int AS total
		FROM derived.algorithms a
		CROSS JOIN total t
		LEFT JOIN covered c ON c.algorithm_id = a.id
		WHERE a.kind = 'game_outcomes'
			AND a.deleted_at IS NULL
		ORDER BY a.created_at DESC
	`)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, name string
		var desc *string
		var covered, total int
		if err := rows.Scan(&id, &name, &desc, &covered, &total); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		items = append(items, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": desc,
			"covered":     covered,
			"total":       total,
		})
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}

func (s *Server) handleGetMarketShareAlgorithmCoverage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	rows, err := s.pool.Query(ctx, `
		WITH total AS (
			SELECT COUNT(*)::int AS total
			FROM core.calcuttas
			WHERE deleted_at IS NULL
		), covered AS (
			SELECT algorithm_id, COUNT(DISTINCT calcutta_id)::int AS covered
			FROM derived.market_share_runs
			WHERE deleted_at IS NULL
			GROUP BY algorithm_id
		)
		SELECT
			a.id::text,
			a.name,
			a.description,
			COALESCE(c.covered, 0)::int AS covered,
			t.total::int AS total
		FROM derived.algorithms a
		CROSS JOIN total t
		LEFT JOIN covered c ON c.algorithm_id = a.id
		WHERE a.kind = 'market_share'
			AND a.deleted_at IS NULL
		ORDER BY a.created_at DESC
	`)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	for rows.Next() {
		var id, name string
		var desc *string
		var covered, total int
		if err := rows.Scan(&id, &name, &desc, &covered, &total); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		items = append(items, map[string]interface{}{
			"id":          id,
			"name":        name,
			"description": desc,
			"covered":     covered,
			"total":       total,
		})
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"count": len(items),
	})
}

func (s *Server) handleGetGameOutcomesAlgorithmCoverageDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	algorithmID := vars["id"]
	if algorithmID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing algorithm ID", "id")
		return
	}

	var algID, algName string
	var algDesc *string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, name, description
		FROM derived.algorithms
		WHERE id = $1::uuid
			AND kind = 'game_outcomes'
			AND deleted_at IS NULL
		LIMIT 1
	`, algorithmID).Scan(&algID, &algName, &algDesc)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Algorithm not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			t.id::text,
			t.name,
			t.starting_at,
			MAX(r.created_at) AS last_run_at
		FROM core.tournaments t
		LEFT JOIN derived.game_outcome_runs r
			ON r.tournament_id = t.id
			AND r.algorithm_id = $1::uuid
			AND r.deleted_at IS NULL
		WHERE t.deleted_at IS NULL
		GROUP BY t.id, t.name, t.starting_at
		ORDER BY t.starting_at DESC NULLS LAST, t.name DESC
	`, algorithmID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	tournaments := make([]map[string]interface{}, 0)
	covered := 0
	total := 0
	for rows.Next() {
		total++
		var tid, name string
		var startingAt *time.Time
		var lastRunAt *time.Time
		if err := rows.Scan(&tid, &name, &startingAt, &lastRunAt); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		if lastRunAt != nil {
			covered++
		}
		tournaments = append(tournaments, map[string]interface{}{
			"tournament_id":   tid,
			"tournament_name": name,
			"starting_at":     startingAt,
			"last_run_at":     lastRunAt,
		})
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"algorithm": map[string]interface{}{
			"id":          algID,
			"name":        algName,
			"description": algDesc,
		},
		"covered": covered,
		"total":   total,
		"items":   tournaments,
		"count":   len(tournaments),
	})
}

func (s *Server) handleGetMarketShareAlgorithmCoverageDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	algorithmID := vars["id"]
	if algorithmID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing algorithm ID", "id")
		return
	}

	var algID, algName string
	var algDesc *string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text, name, description
		FROM derived.algorithms
		WHERE id = $1::uuid
			AND kind = 'market_share'
			AND deleted_at IS NULL
		LIMIT 1
	`, algorithmID).Scan(&algID, &algName, &algDesc)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, r, http.StatusNotFound, "not_found", "Algorithm not found", "id")
			return
		}
		writeErrorFromErr(w, r, err)
		return
	}

	rows, err := s.pool.Query(ctx, `
		SELECT
			c.id::text,
			c.name,
			c.tournament_id::text,
			t.name,
			t.starting_at,
			MAX(r.created_at) AS last_run_at
		FROM core.calcuttas c
		JOIN core.tournaments t
			ON t.id = c.tournament_id
			AND t.deleted_at IS NULL
		LEFT JOIN derived.market_share_runs r
			ON r.calcutta_id = c.id
			AND r.algorithm_id = $1::uuid
			AND r.deleted_at IS NULL
		WHERE c.deleted_at IS NULL
		GROUP BY c.id, c.name, c.tournament_id, t.name, t.starting_at
		ORDER BY t.starting_at DESC NULLS LAST, c.created_at DESC
	`, algorithmID)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]map[string]interface{}, 0)
	covered := 0
	total := 0
	for rows.Next() {
		total++
		var calcuttaID, calcuttaName, tournamentID, tournamentName string
		var startingAt *time.Time
		var lastRunAt *time.Time
		if err := rows.Scan(&calcuttaID, &calcuttaName, &tournamentID, &tournamentName, &startingAt, &lastRunAt); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}
		if lastRunAt != nil {
			covered++
		}
		items = append(items, map[string]interface{}{
			"calcutta_id":     calcuttaID,
			"calcutta_name":   calcuttaName,
			"tournament_id":   tournamentID,
			"tournament_name": tournamentName,
			"starting_at":     startingAt,
			"last_run_at":     lastRunAt,
		})
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"algorithm": map[string]interface{}{
			"id":          algID,
			"name":        algName,
			"description": algDesc,
		},
		"covered": covered,
		"total":   total,
		"items":   items,
		"count":   len(items),
	})
}

func (s *Server) handleListAlgorithms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var kind *string
	if v := r.URL.Query().Get("kind"); v != "" {
		kind = &v
	}

	items, err := s.app.Analytics.ListAlgorithms(ctx, kind)
	if err != nil {
		log.Printf("Error listing algorithms: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to list algorithms", "")
		return
	}

	data := make([]map[string]interface{}, 0, len(items))
	for _, a := range items {
		var params interface{} = nil
		if len(a.ParamsJSON) > 0 {
			params = json.RawMessage(a.ParamsJSON)
		}
		data = append(data, map[string]interface{}{
			"id":          a.ID,
			"kind":        a.Kind,
			"name":        a.Name,
			"description": a.Description,
			"params_json": params,
			"created_at":  a.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"kind":  kind,
		"items": data,
		"count": len(data),
	})
}

func (s *Server) handleListGameOutcomeRunsForTournament(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing tournament ID", "id")
		return
	}

	runs, err := s.app.Analytics.ListGameOutcomeRunsByTournamentID(ctx, tournamentID)
	if err != nil {
		log.Printf("Error listing game outcome runs: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to list game outcome runs", "")
		return
	}

	data := make([]map[string]interface{}, 0, len(runs))
	for _, run := range runs {
		var params interface{} = nil
		if len(run.ParamsJSON) > 0 {
			params = json.RawMessage(run.ParamsJSON)
		}
		data = append(data, map[string]interface{}{
			"id":            run.ID,
			"algorithm_id":  run.AlgorithmID,
			"tournament_id": run.TournamentID,
			"params_json":   params,
			"git_sha":       run.GitSHA,
			"created_at":    run.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id": tournamentID,
		"runs":          data,
		"count":         len(data),
	})
}

func (s *Server) handleListMarketShareRunsForCalcutta(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	runs, err := s.app.Analytics.ListMarketShareRunsByCalcuttaID(ctx, calcuttaID)
	if err != nil {
		log.Printf("Error listing market share runs: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to list market share runs", "")
		return
	}

	data := make([]map[string]interface{}, 0, len(runs))
	for _, run := range runs {
		var params interface{} = nil
		if len(run.ParamsJSON) > 0 {
			params = json.RawMessage(run.ParamsJSON)
		}
		data = append(data, map[string]interface{}{
			"id":           run.ID,
			"algorithm_id": run.AlgorithmID,
			"calcutta_id":  run.CalcuttaID,
			"params_json":  params,
			"git_sha":      run.GitSHA,
			"created_at":   run.CreatedAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id": calcuttaID,
		"runs":        data,
		"count":       len(data),
	})
}

func (s *Server) handleGetLatestPredictionRunsForCalcutta(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		writeError(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	latest, err := s.app.Analytics.GetLatestPredictionRunsForCalcutta(ctx, calcuttaID)
	if err != nil {
		log.Printf("Error getting latest prediction runs: %v", err)
		writeError(w, r, http.StatusInternalServerError, "database_error", "Failed to get latest prediction runs", "")
		return
	}
	if latest == nil {
		writeError(w, r, http.StatusNotFound, "not_found", "No prediction run metadata found for calcutta", "")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":         calcuttaID,
		"tournament_id":       latest.TournamentID,
		"game_outcome_run_id": latest.GameOutcomeRunID,
		"market_share_run_id": latest.MarketShareRunID,
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
