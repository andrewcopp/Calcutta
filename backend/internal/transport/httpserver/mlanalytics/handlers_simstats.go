package mlanalytics

import (
	"context"
	"log"
	"net/http"
	"strconv"

	"github.com/andrewcopp/Calcutta/backend/internal/app"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	app        *app.App
	pool       *pgxpool.Pool
	authUserID func(context.Context) string
}

func NewHandler(a *app.App) *Handler {
	return &Handler{app: a}
}

func NewHandlerWithAuthUserID(a *app.App, authUserID func(context.Context) string) *Handler {
	return &Handler{app: a, authUserID: authUserID}
}

func NewHandlerWithPoolAndAuthUserID(a *app.App, pool *pgxpool.Pool, authUserID func(context.Context) string) *Handler {
	return &Handler{app: a, pool: pool, authUserID: authUserID}
}

// HandleGetTournamentSimStats handles GET /tournaments/{year}/simulations
func (h *Handler) HandleGetTournamentSimStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}

	stats, err := h.app.MLAnalytics.GetTournamentSimStats(ctx, year)
	if err != nil {
		log.Printf("Error getting tournament sim stats: %v", err)
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	if stats == nil {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "Tournament simulations not found", "")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"tournament_id": stats.TournamentID,
		"season":        stats.Season,
		"n_sims":        stats.NSims,
		"n_teams":       stats.NTeams,
		"avg_progress":  stats.AvgProgress,
		"max_progress":  stats.MaxProgress,
	})
}

// HandleGetTeamPerformance handles GET /tournaments/{year}/teams/{team_id}/performance
func (h *Handler) HandleGetTeamPerformance(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}
	teamID := vars["team_id"]
	if teamID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing team_id parameter", "team_id")
		return
	}

	perf, err := h.app.MLAnalytics.GetTeamPerformance(ctx, year, teamID)
	if err != nil {
		log.Printf("Error getting team performance: %v", err)
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	if perf == nil {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "Team performance not found", "")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
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

// HandleGetTeamPerformanceByCalcutta handles GET /api/v1/analytics/calcuttas/{calcutta_id}/teams/{team_id}/performance
func (h *Handler) HandleGetTeamPerformanceByCalcutta(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	calcuttaID := vars["calcutta_id"]
	teamID := vars["team_id"]
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing calcutta_id parameter", "calcutta_id")
		return
	}
	if teamID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Missing team_id parameter", "team_id")
		return
	}

	perf, err := h.app.MLAnalytics.GetTeamPerformanceByCalcutta(ctx, calcuttaID, teamID)
	if err != nil {
		log.Printf("Error getting team performance by calcutta: %v", err)
		httperr.Write(w, r, http.StatusInternalServerError, "database_error", "Failed to get team performance", "")
		return
	}
	if perf == nil {
		httperr.Write(w, r, http.StatusNotFound, "not_found", "Team performance not found", "")
		return
	}

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
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

// HandleGetTeamPredictions handles GET /tournaments/{year}/teams/predictions
func (h *Handler) HandleGetTeamPredictions(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	year, err := strconv.Atoi(vars["year"])
	if err != nil {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Invalid year parameter", "year")
		return
	}

	// Optional run_id query parameter
	var runID *string
	if rid := r.URL.Query().Get("run_id"); rid != "" {
		runID = &rid
	}

	predictions, err := h.app.MLAnalytics.GetTeamPredictions(ctx, year, runID)
	if err != nil {
		log.Printf("Error getting team predictions: %v", err)
		httperr.WriteFromErr(w, r, err, h.authUserID)
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

	response.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"year":  year,
		"teams": teams,
	})
}
