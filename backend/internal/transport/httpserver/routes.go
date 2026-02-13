package httpserver

import (
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/basic"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/calcuttas"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/lab"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/tournaments"
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
}

func (s *Server) registerProtectedRoutes(r *mux.Router) {
	r.HandleFunc("/api/schools", s.schoolsHandler).Methods("GET")

	tHandler := tournaments.NewHandlerWithAuthUserID(s.app, authUserID)
	tournaments.RegisterRoutes(r, tournaments.Handlers{
		ListTournaments:      tHandler.HandleListTournaments,
		GetTournament:        tHandler.HandleGetTournament,
		CreateTournament:     s.requirePermission("tournament.game.write", tHandler.HandleCreateTournament),
		UpdateTournament:     s.requirePermission("tournament.game.write", tHandler.HandleUpdateTournament),
		ListTournamentTeams:  tHandler.HandleListTournamentTeams,
		CreateTournamentTeam: s.requirePermission("tournament.game.write", tHandler.HandleCreateTournamentTeam),
		UpdateTeam:           s.requirePermission("tournament.game.write", tHandler.HandleUpdateTeam),
	})

	s.registerBracketRoutes(r)
	s.registerPortfolioRoutes(r)

	cHandler := calcuttas.NewHandlerWithAuthUserID(s.app, s.authzRepo, authUserID)
	calcuttas.RegisterRoutes(r, calcuttas.Handlers{
		ListCalcuttas:       cHandler.HandleListCalcuttas,
		CreateCalcutta:      cHandler.HandleCreateCalcutta,
		GetCalcutta:         cHandler.HandleGetCalcutta,
		UpdateCalcutta:      cHandler.HandleUpdateCalcutta,
		ListCalcuttaEntries: cHandler.HandleListCalcuttaEntries,
		ListEntryTeams:      cHandler.HandleListEntryTeams,
		ListEntryPortfolios: cHandler.HandleListEntryPortfolios,
		UpdateEntry:         cHandler.HandleUpdateEntry,
	})

	// Lab endpoints (lab.* schema)
	labHandler := lab.NewHandlerWithAuthUserID(s.app, authUserID)
	lab.RegisterRoutes(r, lab.Handlers{
		ListModels:                 s.requirePermission("analytics.suites.read", labHandler.HandleListModels),
		GetModel:                   s.requirePermission("analytics.suites.read", labHandler.HandleGetModel),
		GetLeaderboard:             s.requirePermission("analytics.suites.read", labHandler.HandleGetLeaderboard),
		ListEntries:                s.requirePermission("analytics.suites.read", labHandler.HandleListEntries),
		GetEntry:                   s.requirePermission("analytics.suites.read", labHandler.HandleGetEntry),
		GetEntryByModelAndCalcutta: s.requirePermission("analytics.suites.read", labHandler.HandleGetEntryByModelAndCalcutta),
		ListEvaluations:            s.requirePermission("analytics.suites.read", labHandler.HandleListEvaluations),
		GetEvaluation:              s.requirePermission("analytics.suites.read", labHandler.HandleGetEvaluation),
	})

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
