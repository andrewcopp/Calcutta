package httpserver

import (
	"net/http"

	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/mlanalytics"
)

func (s *Server) handleGetTournamentSimStats(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetTournamentSimStats(w, r)
}

func (s *Server) handleGetTournamentSimStatsByID(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetTournamentSimStatsByID(w, r)
}

func (s *Server) handleListTournamentSimulationBatches(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleListTournamentSimulationBatches(w, r)
}

func (s *Server) handleListAlgorithms(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleListAlgorithms(w, r)
}

func (s *Server) handleGetGameOutcomesAlgorithmCoverage(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithPoolAndAuthUserID(s.app, s.pool, authUserID).HandleGetGameOutcomesAlgorithmCoverage(w, r)
}

func (s *Server) handleGetMarketShareAlgorithmCoverage(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithPoolAndAuthUserID(s.app, s.pool, authUserID).HandleGetMarketShareAlgorithmCoverage(w, r)
}

func (s *Server) handleGetGameOutcomesAlgorithmCoverageDetail(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithPoolAndAuthUserID(s.app, s.pool, authUserID).HandleGetGameOutcomesAlgorithmCoverageDetail(w, r)
}

func (s *Server) handleGetMarketShareAlgorithmCoverageDetail(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithPoolAndAuthUserID(s.app, s.pool, authUserID).HandleGetMarketShareAlgorithmCoverageDetail(w, r)
}

func (s *Server) handleListGameOutcomeRunsForTournament(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleListGameOutcomeRunsForTournament(w, r)
}

func (s *Server) handleGetTournamentPredictedAdvancement(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetTournamentPredictedAdvancement(w, r)
}

func (s *Server) handleGetCalcuttaPredictedReturns(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetCalcuttaPredictedReturns(w, r)
}

func (s *Server) handleGetCalcuttaPredictedInvestment(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetCalcuttaPredictedInvestment(w, r)
}

func (s *Server) handleGetCalcuttaPredictedMarketShare(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetCalcuttaPredictedMarketShare(w, r)
}

func (s *Server) handleListMarketShareRunsForCalcutta(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleListMarketShareRunsForCalcutta(w, r)
}

func (s *Server) handleGetLatestPredictionRunsForCalcutta(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetLatestPredictionRunsForCalcutta(w, r)
}

func (s *Server) handleGetCalcuttaSimulatedEntry(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetCalcuttaSimulatedEntry(w, r)
}

func (s *Server) handleGetCalcuttaSimulatedCalcuttas(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithPoolAndAuthUserID(s.app, s.pool, authUserID).HandleGetCalcuttaSimulatedCalcuttas(w, r)
}

func (s *Server) handleListCalcuttaEvaluationRuns(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleListCalcuttaEvaluationRuns(w, r)
}

func (s *Server) handleListStrategyGenerationRuns(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleListStrategyGenerationRuns(w, r)
}

func (s *Server) handleGetTeamPerformance(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetTeamPerformance(w, r)
}

func (s *Server) handleGetTeamPerformanceByCalcutta(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetTeamPerformanceByCalcutta(w, r)
}

func (s *Server) handleGetTeamPredictions(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetTeamPredictions(w, r)
}

func (s *Server) handleGetOptimizationRuns(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetOptimizationRuns(w, r)
}

func (s *Server) handleGetOurEntryDetails(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetOurEntryDetails(w, r)
}

func (s *Server) handleGetEntryRankings(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetEntryRankings(w, r)
}

func (s *Server) handleGetEntrySimulations(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetEntrySimulations(w, r)
}

func (s *Server) handleGetEntryPortfolio(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithAuthUserID(s.app, authUserID).HandleGetEntryPortfolio(w, r)
}

func (s *Server) handleBulkCreateGameOutcomeRunsForAlgorithm(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithPoolAndAuthUserID(s.app, s.pool, authUserID).HandleBulkCreateGameOutcomeRunsForAlgorithm(w, r)
}

func (s *Server) handleBulkCreateMarketShareRunsForAlgorithm(w http.ResponseWriter, r *http.Request) {
	mlanalytics.NewHandlerWithPoolAndAuthUserID(s.app, s.pool, authUserID).HandleBulkCreateMarketShareRunsForAlgorithm(w, r)
}
