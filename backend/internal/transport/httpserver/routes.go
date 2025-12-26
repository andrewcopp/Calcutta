package httpserver

import "github.com/gorilla/mux"

// RegisterRoutes centralizes HTTP route registration
func (s *Server) RegisterRoutes(r *mux.Router) {
	s.registerBasicRoutes(r)
	s.registerAdminBundleRoutes(r)
	s.registerAdminAnalyticsRoutes(r)
	s.registerAuthRoutes(r)
	s.registerTournamentRoutes(r)
	s.registerBracketRoutes(r)
	s.registerPortfolioRoutes(r)
	s.registerCalcuttaRoutes(r)
	s.registerAnalyticsRoutes(r)
	s.registerHallOfFameRoutes(r)
}

func (s *Server) registerBasicRoutes(r *mux.Router) {
	// Health & basic
	r.HandleFunc("/api/health", s.healthHandler).Methods("GET")
	r.HandleFunc("/api/schools", s.schoolsHandler).Methods("GET")
}

func (s *Server) registerAuthRoutes(r *mux.Router) {
	// Auth
	r.HandleFunc("/api/auth/login", s.loginHandler).Methods("POST")
	r.HandleFunc("/api/auth/signup", s.signupHandler).Methods("POST")
	r.HandleFunc("/api/auth/refresh", s.refreshHandler).Methods("POST")
	r.HandleFunc("/api/auth/logout", s.logoutHandler).Methods("POST")
}

func (s *Server) registerTournamentRoutes(r *mux.Router) {
	// Tournaments
	r.HandleFunc("/api/tournaments", s.tournamentsHandler).Methods("GET")
	r.HandleFunc("/api/tournaments/{id}", s.tournamentHandler).Methods("GET")
	r.HandleFunc("/api/tournaments", s.createTournamentHandler).Methods("POST")
	r.HandleFunc("/api/tournaments/{id}", s.requirePermission("tournament.game.write", s.updateTournamentHandler)).Methods("PATCH")
	r.HandleFunc("/api/tournaments/{id}/teams", s.tournamentTeamsHandler).Methods("GET")
	r.HandleFunc("/api/tournaments/{id}/teams", s.createTournamentTeamHandler).Methods("POST")
	r.HandleFunc("/api/tournaments/{tournamentId}/teams/{teamId}", s.updateTeamHandler).Methods("PATCH", "OPTIONS")
	r.HandleFunc("/api/tournaments/{id}/recalculate-portfolios", s.recalculatePortfoliosHandler).Methods("POST")
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
	r.HandleFunc("/api/portfolios/{id}/calculate-scores", s.calculatePortfolioScoresHandler).Methods("POST")
	r.HandleFunc("/api/portfolios/{id}/teams/{teamId}/scores", s.updatePortfolioTeamScoresHandler).Methods("PUT")
	r.HandleFunc("/api/portfolios/{id}/maximum-score", s.updatePortfolioMaximumScoreHandler).Methods("PUT")
	r.HandleFunc("/api/portfolios/{id}/teams", s.portfolioTeamsHandler).Methods("GET")
}

func (s *Server) registerCalcuttaRoutes(r *mux.Router) {
	// Calcutta
	r.HandleFunc("/api/calcuttas", s.calcuttasHandler).Methods("GET", "POST")
	r.HandleFunc("/api/calcuttas/{id}", s.calcuttaHandler).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}/entries", s.calcuttaEntriesHandler).Methods("GET")
	r.HandleFunc("/api/calcuttas/{calcuttaId}/entries/{entryId}/teams", s.calcuttaEntryTeamsHandler).Methods("GET")
	r.HandleFunc("/api/entries/{id}/portfolios", s.portfoliosHandler).Methods("GET")
}

func (s *Server) registerAnalyticsRoutes(r *mux.Router) {
	// Analytics
	r.HandleFunc("/api/analytics", s.analyticsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/seeds", s.seedAnalyticsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/regions", s.regionAnalyticsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/teams", s.teamAnalyticsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/variance", s.seedVarianceAnalyticsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/seed-investment-distribution", s.seedInvestmentDistributionHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/analytics/best-investments", s.bestInvestmentsHandler).Methods("GET", "OPTIONS")
}

func (s *Server) registerHallOfFameRoutes(r *mux.Router) {
	// Hall of Fame
	r.HandleFunc("/api/hall-of-fame/best-teams", s.hofBestTeamsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/hall-of-fame/best-investments", s.hofBestInvestmentsHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/hall-of-fame/best-entries", s.hofBestEntriesHandler).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/hall-of-fame/best-careers", s.hofBestCareersHandler).Methods("GET", "OPTIONS")
}
