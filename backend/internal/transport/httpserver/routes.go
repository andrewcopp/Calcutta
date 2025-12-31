package httpserver

import (
	"net/http"

	"github.com/gorilla/mux"
)

// RegisterRoutes centralizes HTTP route registration
func (s *Server) RegisterRoutes(r *mux.Router) {
	s.registerBasicRoutes(r)
	s.registerAuthRoutes(r)

	protected := r.NewRoute().Subrouter()
	protected.Use(s.requireAuthMiddleware)
	s.registerAdminBundleRoutes(protected)
	s.registerAdminAPIKeyRoutes(protected)
	s.registerAdminAnalyticsRoutes(protected)
	s.registerProtectedRoutes(protected)
}

func (s *Server) registerBasicRoutes(r *mux.Router) {
	// Health & basic
	r.PathPrefix("/").Methods("OPTIONS").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
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
	s.registerAnalyticsRoutes(r)
	s.registerHallOfFameRoutes(r)
}

func (s *Server) registerAuthRoutes(r *mux.Router) {
	// Auth
	r.HandleFunc("/api/auth/login", s.loginHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/signup", s.signupHandler).Methods("POST", "OPTIONS")
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
	r.HandleFunc("/api/tournaments/{id}/recalculate-portfolios", s.requirePermission("tournament.game.write", s.recalculatePortfoliosHandler)).Methods("POST")
}

func (s *Server) registerBracketRoutes(r *mux.Router) {
	// Bracket management
	r.HandleFunc("/api/tournaments/{id}/bracket", s.getBracketHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/tournaments/{id}/bracket/validate", s.validateBracketSetupHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/tournaments/{tournamentId}/bracket/games/{gameId}/winner", s.requirePermission("tournament.game.write", s.selectWinnerHandler)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/tournaments/{tournamentId}/bracket/games/{gameId}/winner", s.requirePermission("tournament.game.write", s.unselectWinnerHandler)).Methods("DELETE", "OPTIONS")
}

func (s *Server) registerPortfolioRoutes(r *mux.Router) {
	// Portfolio scoring
	r.HandleFunc("/api/portfolios/{id}/calculate-scores", s.requirePermission("tournament.game.write", s.calculatePortfolioScoresHandler)).Methods("POST")
	r.HandleFunc("/api/portfolios/{id}/teams/{teamId}/scores", s.requirePermission("tournament.game.write", s.updatePortfolioTeamScoresHandler)).Methods("PUT")
	r.HandleFunc("/api/portfolios/{id}/maximum-score", s.requirePermission("tournament.game.write", s.updatePortfolioMaximumScoreHandler)).Methods("PUT")
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
	r.HandleFunc("/api/analytics/tournaments/{id}/predicted-returns", s.handleGetTournamentPredictedReturns).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/v1/analytics/tournaments/{year}/teams/{team_id}/performance", s.handleGetTeamPerformance).Methods("GET", "OPTIONS")
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
