package httpserver

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes centralizes HTTP route registration
func (s *Server) RegisterRoutes(r *mux.Router) {
	s.registerBasicRoutes(r)
	if s.cfg.AuthMode != "cognito" {
		s.registerAuthRoutes(r)
	}

	protected := r.NewRoute().Subrouter()
	protected.Use(s.requireAuthMiddleware)
	s.registerAdminBundleRoutes(protected)
	s.registerAdminAPIKeyRoutes(protected)
	s.registerAdminAnalyticsRoutes(protected)
	s.registerAdminUsersRoutes(protected)
	s.registerProtectedRoutes(protected)
}

func (s *Server) registerBasicRoutes(r *mux.Router) {
	// Health & basic
	r.PathPrefix("/").Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	r.HandleFunc("/healthz", s.healthHandler).Methods("GET")
	r.HandleFunc("/readyz", s.readyHandler).Methods("GET")
	r.HandleFunc("/health/live", s.healthHandler).Methods("GET")
	r.HandleFunc("/health/ready", s.readyHandler).Methods("GET")
	if s.cfg.MetricsEnabled {
		r.HandleFunc("/metrics", s.metricsHandler).Methods("GET")
	}
	r.HandleFunc("/api/health", s.healthHandler).Methods("GET")
	r.HandleFunc("/api/ready", s.readyHandler).Methods("GET")

	// ML Analytics (public read-only endpoints)
	s.registerMLAnalyticsRoutes(r)
}

func (s *Server) registerProtectedRoutes(r *mux.Router) {
	r.HandleFunc("/api/schools", s.schoolsHandler).Methods("GET")
	s.registerTournamentRoutes(r)
	s.registerBracketRoutes(r)
	s.registerPortfolioRoutes(r)
	s.registerCalcuttaRoutes(r)
	s.registerEntryEvaluationRequestRoutes(r)
	s.registerSuiteRoutes(r)
	s.registerSuiteScenarioRoutes(r)
	s.registerStrategyGenerationRunRoutes(r)
	s.registerLabEntriesRoutes(r)
	s.registerModelCatalogRoutes(r)
	s.registerSuiteCalcuttaEvaluationRoutes(r)
	s.registerSuiteExecutionRoutes(r)
	s.registerAnalyticsRoutes(r)
	s.registerHallOfFameRoutes(r)
}

func (s *Server) registerAuthRoutes(r *mux.Router) {
	// Auth
	r.HandleFunc("/api/auth/login", s.loginHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/signup", s.signupHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/invite/accept", s.acceptInviteHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/refresh", s.refreshHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/logout", s.logoutHandler).Methods("POST", "OPTIONS")
}

func (s *Server) registerTournamentRoutes(r *mux.Router) {
	// Tournaments
	r.HandleFunc("/api/tournaments", s.tournamentsHandler).Methods("GET")
	r.HandleFunc("/api/tournaments/{id}", s.tournamentHandler).Methods("GET")
	r.HandleFunc("/api/tournaments", s.requirePermission("tournament.game.write", s.createTournamentHandler)).Methods("POST")
	r.HandleFunc("/api/tournaments/{id}", s.requirePermission("tournament.game.write", s.updateTournamentHandler)).Methods("PATCH")
	r.HandleFunc("/api/tournaments/{id}/teams", s.tournamentTeamsHandler).Methods("GET")
	r.HandleFunc("/api/tournaments/{id}/teams", s.requirePermission("tournament.game.write", s.createTournamentTeamHandler)).Methods("POST")
	r.HandleFunc("/api/tournaments/{tournamentId}/teams/{teamId}", s.requirePermission("tournament.game.write", s.updateTeamHandler)).Methods("PATCH", "OPTIONS")
}

func (s *Server) registerBracketRoutes(r *mux.Router) {
	// Bracket management
	r.HandleFunc("/api/tournaments/{id}/bracket", s.getBracketHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/tournaments/{id}/bracket/validate", s.validateBracketSetupHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/tournaments/{tournamentId}/bracket/games/{gameId}/winner", s.requirePermission("tournament.game.write", s.selectWinnerHandler)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/tournaments/{tournamentId}/bracket/games/{gameId}/winner", s.requirePermission("tournament.game.write", s.unselectWinnerHandler)).Methods("DELETE", "OPTIONS")
}

func (s *Server) registerPortfolioRoutes(r *mux.Router) {
	r.HandleFunc("/api/portfolios/{id}/teams", s.portfolioTeamsHandler).Methods("GET")
}

func (s *Server) registerCalcuttaRoutes(r *mux.Router) {
	// Calcutta
	r.HandleFunc("/api/calcuttas", s.calcuttasHandler).Methods("GET")
	r.HandleFunc("/api/calcuttas", s.calcuttasHandler).Methods("POST")
	r.HandleFunc("/api/calcuttas/{id}", s.calcuttaHandler).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}", s.updateCalcuttaHandler).Methods("PATCH")
	r.HandleFunc("/api/calcuttas/{id}/entries", s.calcuttaEntriesHandler).Methods("GET")
	r.HandleFunc("/api/calcuttas/{calcuttaId}/entries/{entryId}/teams", s.calcuttaEntryTeamsHandler).Methods("GET")
	r.HandleFunc("/api/entries/{id}/portfolios", s.portfoliosHandler).Methods("GET")
	r.HandleFunc("/api/entries/{id}", s.updateEntryHandler).Methods("PATCH")
}

func (s *Server) registerAnalyticsRoutes(r *mux.Router) {
	// Analytics
	r.HandleFunc("/api/analytics", s.requirePermission("admin.analytics.read", s.analyticsHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/seeds", s.requirePermission("admin.analytics.read", s.seedAnalyticsHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/regions", s.requirePermission("admin.analytics.read", s.regionAnalyticsHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/teams", s.requirePermission("admin.analytics.read", s.teamAnalyticsHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/variance", s.requirePermission("admin.analytics.read", s.seedVarianceAnalyticsHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/seed-investment-distribution", s.requirePermission("admin.analytics.read", s.seedInvestmentDistributionHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/best-investments", s.requirePermission("admin.analytics.read", s.bestInvestmentsHandler)).Methods("GET", "OPTIONS")
}

func (s *Server) registerMLAnalyticsRoutes(r *mux.Router) {
	// ML Analytics (read-only)
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/simulations", s.handleGetTournamentSimStats).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/tournaments/{id}/simulations", s.handleGetTournamentSimStatsByID).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/tournaments/{id}/simulation-batches", s.handleListTournamentSimulationBatches).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/algorithms", s.handleListAlgorithms).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/coverage/game-outcomes", s.handleGetGameOutcomesAlgorithmCoverage).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/coverage/market-share", s.handleGetMarketShareAlgorithmCoverage).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/algorithms/{id}/coverage/game-outcomes", s.handleGetGameOutcomesAlgorithmCoverageDetail).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/algorithms/{id}/coverage/market-share", s.handleGetMarketShareAlgorithmCoverageDetail).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/tournaments/{id}/game-outcome-runs", s.handleListGameOutcomeRunsForTournament).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/tournaments/{id}/predicted-advancement", s.handleGetTournamentPredictedAdvancement).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/predicted-returns", s.handleGetCalcuttaPredictedReturns).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/predicted-investment", s.handleGetCalcuttaPredictedInvestment).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/predicted-market-share", s.handleGetCalcuttaPredictedMarketShare).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/market-share-runs", s.handleListMarketShareRunsForCalcutta).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/latest-prediction-runs", s.handleGetLatestPredictionRunsForCalcutta).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/simulated-entry", s.handleGetCalcuttaSimulatedEntry).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/simulated-calcuttas", s.handleGetCalcuttaSimulatedCalcuttas).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/evaluation-runs", s.handleListCalcuttaEvaluationRuns).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/calcuttas/{id}/strategy-generation-runs", s.handleListStrategyGenerationRuns).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/teams/{team_id}/performance", s.handleGetTeamPerformance).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/calcuttas/{calcutta_id}/teams/{team_id}/performance", s.handleGetTeamPerformanceByCalcutta).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/teams/predictions", s.handleGetTeamPredictions).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs", s.handleGetOptimizationRuns).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs/{run_id}/our-entry", s.handleGetOurEntryDetails).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs/{run_id}/rankings", s.handleGetEntryRankings).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs/{run_id}/entries/{entry_key}/simulations", s.handleGetEntrySimulations).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/runs/{run_id}/entries/{entry_key}/portfolio", s.handleGetEntryPortfolio).Methods("GET", "OPTIONS")
}

func (s *Server) registerHallOfFameRoutes(r *mux.Router) {
	// Hall of Fame
	r.HandleFunc("/api/hall-of-fame/best-teams", s.requirePermission("admin.hof.read", s.hofBestTeamsHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/hall-of-fame/best-investments", s.requirePermission("admin.hof.read", s.hofBestInvestmentsHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/hall-of-fame/best-entries", s.requirePermission("admin.hof.read", s.hofBestEntriesHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/hall-of-fame/best-careers", s.requirePermission("admin.hof.read", s.hofBestCareersHandler)).Methods("GET", "OPTIONS")
}
