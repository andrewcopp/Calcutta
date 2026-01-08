package httpserver

import (
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/basic"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/mlanalytics"
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
	basic.RegisterRoutes(r,
		basic.Options{MetricsEnabled: s.cfg.MetricsEnabled},
		basic.Handlers{Health: s.healthHandler, Ready: s.readyHandler, Metrics: s.metricsHandler},
	)

	// ML Analytics (public read-only endpoints)
	mlanalytics.RegisterRoutes(r, mlanalytics.Handlers{
		TournamentSimStats:                 s.handleGetTournamentSimStats,
		TournamentSimStatsByID:             s.handleGetTournamentSimStatsByID,
		ListTournamentSimulationBatches:    s.handleListTournamentSimulationBatches,
		ListAlgorithms:                     s.handleListAlgorithms,
		GetGameOutcomesAlgorithmCoverage:   s.handleGetGameOutcomesAlgorithmCoverage,
		GetMarketShareAlgorithmCoverage:    s.handleGetMarketShareAlgorithmCoverage,
		GetGameOutcomesCoverageDetail:      s.handleGetGameOutcomesAlgorithmCoverageDetail,
		GetMarketShareCoverageDetail:       s.handleGetMarketShareAlgorithmCoverageDetail,
		ListGameOutcomeRunsForTournament:   s.handleListGameOutcomeRunsForTournament,
		GetTournamentPredictedAdvancement:  s.handleGetTournamentPredictedAdvancement,
		GetCalcuttaPredictedReturns:        s.handleGetCalcuttaPredictedReturns,
		GetCalcuttaPredictedInvestment:     s.handleGetCalcuttaPredictedInvestment,
		GetCalcuttaPredictedMarketShare:    s.handleGetCalcuttaPredictedMarketShare,
		ListMarketShareRunsForCalcutta:     s.handleListMarketShareRunsForCalcutta,
		GetLatestPredictionRunsForCalcutta: s.handleGetLatestPredictionRunsForCalcutta,
		GetCalcuttaSimulatedEntry:          s.handleGetCalcuttaSimulatedEntry,
		GetCalcuttaSimulatedCalcuttas:      s.handleGetCalcuttaSimulatedCalcuttas,
		ListCalcuttaEvaluationRuns:         s.handleListCalcuttaEvaluationRuns,
		ListStrategyGenerationRuns:         s.handleListStrategyGenerationRuns,
		GetTeamPerformance:                 s.handleGetTeamPerformance,
		GetTeamPerformanceByCalcutta:       s.handleGetTeamPerformanceByCalcutta,
		GetTeamPredictions:                 s.handleGetTeamPredictions,
		GetOptimizationRuns:                s.handleGetOptimizationRuns,
		GetOurEntryDetails:                 s.handleGetOurEntryDetails,
		GetEntryRankings:                   s.handleGetEntryRankings,
		GetEntrySimulations:                s.handleGetEntrySimulations,
		GetEntryPortfolio:                  s.handleGetEntryPortfolio,
	})
}

func (s *Server) registerProtectedRoutes(r *mux.Router) {
	r.HandleFunc("/api/schools", s.schoolsHandler).Methods("GET")
	s.registerTournamentRoutes(r)
	s.registerBracketRoutes(r)
	s.registerPortfolioRoutes(r)
	s.registerCalcuttaRoutes(r)
	s.registerEntryEvaluationRequestRoutes(r)
	s.registerRunProgressRoutes(r)
	s.registerEntryRunRoutes(r)
	s.registerEntryArtifactRoutes(r)
	s.registerStrategyGenerationRunRoutes(r)
	s.registerLabEntriesRoutes(r)
	s.registerLabCandidatesRoutes(r)
	s.registerModelCatalogRoutes(r)
	s.registerSyntheticCalcuttaCohortRoutes(r)
	s.registerSyntheticCalcuttaRoutes(r)
	s.registerSyntheticEntryRoutes(r)
	s.registerCohortSimulationRoutes(r)
	s.registerCohortSimulationBatchRoutes(r)
	s.registerAnalyticsRoutes(r)
	r.HandleFunc(
		"/api/analytics/algorithms/{id}/game-outcome-runs/bulk",
		s.requirePermission("analytics.strategy_generation_runs.write", s.handleBulkCreateGameOutcomeRunsForAlgorithm),
	).Methods("POST", "OPTIONS")
	r.HandleFunc(
		"/api/analytics/algorithms/{id}/market-share-runs/bulk",
		s.requirePermission("analytics.strategy_generation_runs.write", s.handleBulkCreateMarketShareRunsForAlgorithm),
	).Methods("POST", "OPTIONS")
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

func (s *Server) registerHallOfFameRoutes(r *mux.Router) {
	// Hall of Fame
	r.HandleFunc("/api/hall-of-fame/best-teams", s.requirePermission("admin.hof.read", s.hofBestTeamsHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/hall-of-fame/best-investments", s.requirePermission("admin.hof.read", s.hofBestInvestmentsHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/hall-of-fame/best-entries", s.requirePermission("admin.hof.read", s.hofBestEntriesHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/hall-of-fame/best-careers", s.requirePermission("admin.hof.read", s.hofBestCareersHandler)).Methods("GET", "OPTIONS")
}
