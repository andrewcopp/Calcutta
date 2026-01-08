package mlanalytics

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListAlgorithms(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var kind *string
	if v := r.URL.Query().Get("kind"); v != "" {
		kind = &v
	}

	items, err := h.app.Analytics.ListAlgorithms(ctx, kind)
	if err != nil {
		log.Printf("Error listing algorithms: %v", err)
		httperr.Write(w, r, http.StatusInternalServerError, "database_error", "Failed to list algorithms", "")
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

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"kind":  kind,
		"items": data,
		"count": len(data),
	})
}

func (h *Handler) HandleListGameOutcomeRunsForTournament(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing tournament ID", "id")
		return
	}

	runs, err := h.app.Analytics.ListGameOutcomeRunsByTournamentID(ctx, tournamentID)
	if err != nil {
		log.Printf("Error listing game outcome runs: %v", err)
		httperr.Write(w, r, http.StatusInternalServerError, "database_error", "Failed to list game outcome runs", "")
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

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id": tournamentID,
		"runs":          data,
		"count":         len(data),
	})
}

func (h *Handler) HandleListMarketShareRunsForCalcutta(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	runs, err := h.app.Analytics.ListMarketShareRunsByCalcuttaID(ctx, calcuttaID)
	if err != nil {
		log.Printf("Error listing market share runs: %v", err)
		httperr.Write(w, r, http.StatusInternalServerError, "database_error", "Failed to list market share runs", "")
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

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id": calcuttaID,
		"runs":        data,
		"count":       len(data),
	})
}

func (h *Handler) HandleGetLatestPredictionRunsForCalcutta(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	latest, err := h.app.Analytics.GetLatestPredictionRunsForCalcutta(ctx, calcuttaID)
	if err != nil {
		log.Printf("Error getting latest prediction runs: %v", err)
		httperr.Write(w, r, http.StatusInternalServerError, "database_error", "Failed to get latest prediction runs", "")
		return
	}
	if latest == nil {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "No prediction run metadata found for calcutta", "")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":         calcuttaID,
		"tournament_id":       latest.TournamentID,
		"game_outcome_run_id": latest.GameOutcomeRunID,
		"market_share_run_id": latest.MarketShareRunID,
	})
}

func (h *Handler) HandleGetTournamentPredictedAdvancement(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing tournament ID", "id")
		return
	}

	var gameOutcomeRunID *string
	if v := r.URL.Query().Get("game_outcome_run_id"); v != "" {
		gameOutcomeRunID = &v
	}
	if gameOutcomeRunID == nil || *gameOutcomeRunID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "game_outcome_run_id is required", "game_outcome_run_id")
		return
	}

	selectedRunID, data, err := h.app.Analytics.GetTournamentPredictedAdvancement(ctx, tournamentID, gameOutcomeRunID)
	if err != nil {
		log.Printf("Error querying tournament predicted advancement: %v", err)
		httperr.Write(w, r, http.StatusInternalServerError, "database_error", "Failed to query predicted advancement", "")
		return
	}
	if len(data) == 0 {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "No predicted advancement found for tournament", "")
		return
	}

	results := make([]map[string]interface{}, 0, len(data))
	for _, d := range data {
		results = append(results, map[string]interface{}{
			"team_id":     d.TeamID,
			"school_name": d.SchoolName,
			"seed":        d.Seed,
			"region":      d.Region,
			"prob_pi":     d.ProbPI,
			"reach_r64":   d.ReachR64,
			"reach_r32":   d.ReachR32,
			"reach_s16":   d.ReachS16,
			"reach_e8":    d.ReachE8,
			"reach_ff":    d.ReachFF,
			"reach_champ": d.ReachChamp,
			"win_champ":   d.WinChamp,
		})
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id":       tournamentID,
		"game_outcome_run_id": selectedRunID,
		"teams":               results,
		"count":               len(results),
	})
}

func (h *Handler) HandleGetCalcuttaPredictedReturns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]

	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	var entryRunID *string
	if v := r.URL.Query().Get("entry_run_id"); v != "" {
		entryRunID = &v
	} else if v := r.URL.Query().Get("strategy_generation_run_id"); v != "" {
		// Backward compat.
		entryRunID = &v
	}

	var gameOutcomeRunID *string
	if v := r.URL.Query().Get("game_outcome_run_id"); v != "" {
		gameOutcomeRunID = &v
	}
	if gameOutcomeRunID == nil || *gameOutcomeRunID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "game_outcome_run_id is required", "game_outcome_run_id")
		return
	}

	selectedID, gameOutcomeSelectedID, data, err := h.app.Analytics.GetCalcuttaPredictedReturns(ctx, calcuttaID, entryRunID, gameOutcomeRunID)
	if err != nil {
		log.Printf("Error querying predicted returns: %v", err)
		httperr.Write(w, r, http.StatusInternalServerError, "database_error", "Failed to query predicted returns", "")
		return
	}

	if len(data) == 0 {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "No predicted returns found for calcutta", "")
		return
	}

	results := make([]map[string]interface{}, 0, len(data))
	for _, d := range data {
		results = append(results, map[string]interface{}{
			"team_id":        d.TeamID,
			"school_name":    d.SchoolName,
			"seed":           d.Seed,
			"region":         d.Region,
			"prob_pi":        d.ProbPI,
			"prob_r64":       d.ProbR64,
			"prob_r32":       d.ProbR32,
			"prob_s16":       d.ProbS16,
			"prob_e8":        d.ProbE8,
			"prob_ff":        d.ProbFF,
			"prob_champ":     d.ProbChamp,
			"expected_value": d.ExpectedValue,
		})
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":                calcuttaID,
		"entry_run_id":               selectedID,
		"strategy_generation_run_id": selectedID,
		"game_outcome_run_id":        gameOutcomeSelectedID,
		"teams":                      results,
		"count":                      len(results),
	})
}

func (h *Handler) HandleGetCalcuttaPredictedInvestment(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]

	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	var entryRunID *string
	if v := r.URL.Query().Get("entry_run_id"); v != "" {
		entryRunID = &v
	} else if v := r.URL.Query().Get("strategy_generation_run_id"); v != "" {
		// Backward compat.
		entryRunID = &v
	}

	var marketShareRunID *string
	if v := r.URL.Query().Get("market_share_run_id"); v != "" {
		marketShareRunID = &v
	}
	if marketShareRunID == nil || *marketShareRunID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "market_share_run_id is required", "market_share_run_id")
		return
	}

	var gameOutcomeRunID *string
	if v := r.URL.Query().Get("game_outcome_run_id"); v != "" {
		gameOutcomeRunID = &v
	}
	if gameOutcomeRunID == nil || *gameOutcomeRunID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "game_outcome_run_id is required", "game_outcome_run_id")
		return
	}

	selectedID, marketShareSelectedID, data, err := h.app.Analytics.GetCalcuttaPredictedInvestment(ctx, calcuttaID, entryRunID, marketShareRunID, gameOutcomeRunID)
	if err != nil {
		log.Printf("Error querying predicted investment: %v", err)
		httperr.Write(w, r, http.StatusInternalServerError, "database_error", "Failed to query predicted investment", "")
		return
	}

	if len(data) == 0 {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "No predicted investment found for calcutta", "")
		return
	}

	results := make([]map[string]interface{}, 0, len(data))
	for _, d := range data {
		results = append(results, map[string]interface{}{
			"team_id":     d.TeamID,
			"school_name": d.SchoolName,
			"seed":        d.Seed,
			"region":      d.Region,
			"rational":    d.Rational,
			"predicted":   d.Predicted,
			"delta":       d.Delta,
		})
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":                calcuttaID,
		"entry_run_id":               selectedID,
		"strategy_generation_run_id": selectedID,
		"market_share_run_id":        marketShareSelectedID,
		"game_outcome_run_id":        *gameOutcomeRunID,
		"teams":                      results,
		"count":                      len(results),
	})
}

func (h *Handler) HandleGetCalcuttaPredictedMarketShare(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	var marketShareRunID *string
	if v := r.URL.Query().Get("market_share_run_id"); v != "" {
		marketShareRunID = &v
	}
	if marketShareRunID == nil || *marketShareRunID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "market_share_run_id is required", "market_share_run_id")
		return
	}

	var gameOutcomeRunID *string
	if v := r.URL.Query().Get("game_outcome_run_id"); v != "" {
		gameOutcomeRunID = &v
	}
	if gameOutcomeRunID == nil || *gameOutcomeRunID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "game_outcome_run_id is required", "game_outcome_run_id")
		return
	}

	marketShareSelectedID, gameOutcomeSelectedID, data, err := h.app.Analytics.GetCalcuttaPredictedMarketShare(ctx, calcuttaID, marketShareRunID, gameOutcomeRunID)
	if err != nil {
		log.Printf("Error querying predicted market share: %v", err)
		httperr.Write(w, r, http.StatusInternalServerError, "database_error", "Failed to query predicted market share", "")
		return
	}
	if len(data) == 0 {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "No predicted market share found for calcutta", "")
		return
	}

	results := make([]map[string]interface{}, 0, len(data))
	for _, d := range data {
		results = append(results, map[string]interface{}{
			"team_id":         d.TeamID,
			"school_name":     d.SchoolName,
			"seed":            d.Seed,
			"region":          d.Region,
			"rational_share":  d.RationalShare,
			"predicted_share": d.PredictedShare,
			"delta_percent":   d.DeltaPercent,
		})
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":         calcuttaID,
		"market_share_run_id": marketShareSelectedID,
		"game_outcome_run_id": gameOutcomeSelectedID,
		"teams":               results,
		"count":               len(results),
	})
}

func (h *Handler) HandleGetCalcuttaSimulatedEntry(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]

	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	var entryRunID *string
	if v := r.URL.Query().Get("entry_run_id"); v != "" {
		entryRunID = &v
	} else if v := r.URL.Query().Get("strategy_generation_run_id"); v != "" {
		// Backward compat.
		entryRunID = &v
	}
	if entryRunID == nil || *entryRunID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "entry_run_id is required", "entry_run_id")
		return
	}

	selectedID, data, err := h.app.Analytics.GetCalcuttaSimulatedEntry(ctx, calcuttaID, entryRunID)
	if err != nil {
		log.Printf("Error querying simulated entry: %v", err)
		httperr.Write(w, r, http.StatusInternalServerError, "database_error", "Failed to query simulated entry", "")
		return
	}

	if len(data) == 0 {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "No simulated entry found for calcutta", "")
		return
	}

	results := make([]map[string]interface{}, 0, len(data))
	for _, d := range data {
		se := map[string]interface{}{
			"team_id":         d.TeamID,
			"school_name":     d.SchoolName,
			"seed":            d.Seed,
			"region":          d.Region,
			"expected_points": d.ExpectedPoints,
			"expected_market": d.ExpectedMarket,
			"our_bid":         d.OurBid,
		}

		expectedROI := 0.0
		if d.ExpectedMarket > 0 {
			expectedROI = d.ExpectedPoints / d.ExpectedMarket
		}
		ourROI := 0.0
		totalMarket := d.ExpectedMarket + d.OurBid
		if totalMarket > 0 {
			ourROI = d.ExpectedPoints / totalMarket
		}
		se["expected_roi"] = expectedROI
		se["our_roi"] = ourROI

		results = append(results, se)
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id":                calcuttaID,
		"entry_run_id":               selectedID,
		"strategy_generation_run_id": selectedID,
		"teams":                      results,
		"count":                      len(results),
	})
}
