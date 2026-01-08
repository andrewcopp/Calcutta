package mlanalytics

import (
	"net/http"

	"github.com/gorilla/mux"
)

type Handlers struct {
	TournamentSimStats                 http.HandlerFunc
	TournamentSimStatsByID             http.HandlerFunc
	ListTournamentSimulationBatches    http.HandlerFunc
	ListAlgorithms                     http.HandlerFunc
	GetGameOutcomesAlgorithmCoverage   http.HandlerFunc
	GetMarketShareAlgorithmCoverage    http.HandlerFunc
	GetGameOutcomesCoverageDetail      http.HandlerFunc
	GetMarketShareCoverageDetail       http.HandlerFunc
	ListGameOutcomeRunsForTournament   http.HandlerFunc
	GetTournamentPredictedAdvancement  http.HandlerFunc
	GetCalcuttaPredictedReturns        http.HandlerFunc
	GetCalcuttaPredictedInvestment     http.HandlerFunc
	GetCalcuttaPredictedMarketShare    http.HandlerFunc
	ListMarketShareRunsForCalcutta     http.HandlerFunc
	GetLatestPredictionRunsForCalcutta http.HandlerFunc
	GetCalcuttaSimulatedEntry          http.HandlerFunc
	GetCalcuttaSimulatedCalcuttas      http.HandlerFunc
	ListCalcuttaEvaluationRuns         http.HandlerFunc
	ListStrategyGenerationRuns         http.HandlerFunc
	GetTeamPerformance                 http.HandlerFunc
	GetTeamPerformanceByCalcutta       http.HandlerFunc
	GetTeamPredictions                 http.HandlerFunc
	GetOptimizationRuns                http.HandlerFunc
	GetOurEntryDetails                 http.HandlerFunc
	GetEntryRankings                   http.HandlerFunc
	GetEntrySimulations                http.HandlerFunc
	GetEntryPortfolio                  http.HandlerFunc
}

func RegisterRoutes(r *mux.Router, h Handlers) {
	// ML Analytics (read-only)
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/simulations", h.TournamentSimStats).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/tournaments/{id}/simulations", h.TournamentSimStatsByID).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/tournaments/{id}/simulation-batches", h.ListTournamentSimulationBatches).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/algorithms", h.ListAlgorithms).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/coverage/game-outcomes", h.GetGameOutcomesAlgorithmCoverage).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/coverage/market-share", h.GetMarketShareAlgorithmCoverage).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/algorithms/{id}/coverage/game-outcomes", h.GetGameOutcomesCoverageDetail).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/algorithms/{id}/coverage/market-share", h.GetMarketShareCoverageDetail).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/tournaments/{id}/game-outcome-runs", h.ListGameOutcomeRunsForTournament).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/tournaments/{id}/predicted-advancement", h.GetTournamentPredictedAdvancement).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/predicted-returns", h.GetCalcuttaPredictedReturns).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/predicted-investment", h.GetCalcuttaPredictedInvestment).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/predicted-market-share", h.GetCalcuttaPredictedMarketShare).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/market-share-runs", h.ListMarketShareRunsForCalcutta).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/latest-prediction-runs", h.GetLatestPredictionRunsForCalcutta).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/simulated-entry", h.GetCalcuttaSimulatedEntry).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/simulated-calcuttas", h.GetCalcuttaSimulatedCalcuttas).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/evaluation-runs", h.ListCalcuttaEvaluationRuns).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/entry-runs", h.ListStrategyGenerationRuns).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/strategy-generation-runs", h.ListStrategyGenerationRuns).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/teams/{team_id}/performance", h.GetTeamPerformance).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/calcuttas/{calcutta_id}/teams/{team_id}/performance", h.GetTeamPerformanceByCalcutta).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/teams/predictions", h.GetTeamPredictions).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs", h.GetOptimizationRuns).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs/{run_id}/our-entry", h.GetOurEntryDetails).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs/{run_id}/rankings", h.GetEntryRankings).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs/{run_id}/entries/{entry_key}/simulations", h.GetEntrySimulations).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs/{run_id}/entries/{entry_key}/portfolio", h.GetEntryPortfolio).Methods("GET", "OPTIONS")
}
