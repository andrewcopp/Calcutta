package mlanalytics

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleListTournamentSimulationBatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing tournament ID", "id")
		return
	}

	batches, err := h.app.MLAnalytics.ListTournamentSimulationBatchesByCoreTournamentID(ctx, tournamentID)
	if err != nil {
		log.Printf("Error listing tournament simulation batches: %v", err)
		httperr.WriteFromErr(w, r, err, h.authUserID)
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

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id": tournamentID,
		"batches":       data,
		"count":         len(data),
	})
}

func (h *Handler) HandleListCalcuttaEvaluationRuns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	runs, err := h.app.MLAnalytics.ListCalcuttaEvaluationRunsByCoreCalcuttaID(ctx, calcuttaID)
	if err != nil {
		log.Printf("Error listing calcutta evaluation runs: %v", err)
		httperr.WriteFromErr(w, r, err, h.authUserID)
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

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id": calcuttaID,
		"runs":        data,
		"count":       len(data),
	})
}

func (h *Handler) HandleListStrategyGenerationRuns(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta ID", "id")
		return
	}

	runs, err := h.app.MLAnalytics.ListStrategyGenerationRunsByCoreCalcuttaID(ctx, calcuttaID)
	if err != nil {
		log.Printf("Error listing strategy generation runs: %v", err)
		httperr.WriteFromErr(w, r, err, h.authUserID)
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

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"calcutta_id": calcuttaID,
		"runs":        data,
		"count":       len(data),
	})
}
