package httpserver

import (
	"net/http"
	"strings"

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

	// API v1 path rewrite: /api/v1/* → /api/* (both paths work identically)
	r.PathPrefix("/api/v1/").HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		req.URL.Path = "/api/" + strings.TrimPrefix(req.URL.Path, "/api/v1/")
		req.RequestURI = req.URL.RequestURI()
		r.ServeHTTP(w, req)
	})

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
	r.HandleFunc("/api/me/permissions", s.mePermissionsHandler).Methods("GET", "OPTIONS")
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
		ListCompetitions:     tHandler.HandleListCompetitions,
		ListSeasons:          tHandler.HandleListSeasons,
		ReplaceTeams:         s.requirePermission("tournament.game.write", tHandler.HandleReplaceTeams),
	})

	s.registerBracketRoutes(r)
	s.registerPortfolioRoutes(r)

	cHandler := calcuttas.NewHandlerWithAuthUserID(s.app, s.authzRepo, s.authzRepo, authUserID)
	calcuttas.RegisterRoutes(r, calcuttas.Handlers{
		ListCalcuttas:             cHandler.HandleListCalcuttas,
		ListCalcuttasWithRankings: cHandler.HandleListCalcuttasWithRankings,
		CreateCalcutta:            cHandler.HandleCreateCalcutta,
		GetCalcutta:               cHandler.HandleGetCalcutta,
		GetDashboard:              cHandler.HandleGetDashboard,
		UpdateCalcutta:            cHandler.HandleUpdateCalcutta,
		ListCalcuttaEntries:       cHandler.HandleListCalcuttaEntries,
		CreateEntry:               cHandler.HandleCreateEntry,
		CreateInvitation:          cHandler.HandleCreateInvitation,
		ListInvitations:           cHandler.HandleListInvitations,
		AcceptInvitation:          cHandler.HandleAcceptInvitation,
		RevokeInvitation:          cHandler.HandleRevokeInvitation,
		ListMyInvitations:         cHandler.HandleListMyInvitations,
		ListEntryTeams:            cHandler.HandleListEntryTeams,
		ListEntryPortfolios:       cHandler.HandleListEntryPortfolios,
		UpdateEntry:               idempotencyMiddleware(s.idempotencyRepo, cHandler.HandleUpdateEntry),
		Reinvite:                  cHandler.HandleReinvite,
		ListPayouts:               cHandler.HandleListPayouts,
		ReplacePayouts:            cHandler.HandleReplacePayouts,
	})

	// Lab endpoints (lab.* schema)
	labHandler := lab.NewHandlerWithAuthUserID(s.app, authUserID)
	lab.RegisterRoutes(r, lab.Handlers{
		ListModels:                 s.requirePermission("lab.read", labHandler.HandleListModels),
		GetModel:                   s.requirePermission("lab.read", labHandler.HandleGetModel),
		GetLeaderboard:             s.requirePermission("lab.read", labHandler.HandleGetLeaderboard),
		StartPipeline:              s.requirePermission("lab.write", labHandler.HandleStartPipeline),
		GetModelPipelineProgress:   s.requirePermission("lab.read", labHandler.HandleGetModelPipelineProgress),
		GetPipelineRun:             s.requirePermission("lab.read", labHandler.HandleGetPipelineRun),
		CancelPipeline:             s.requirePermission("lab.write", labHandler.HandleCancelPipeline),
		ListEntries:                s.requirePermission("lab.read", labHandler.HandleListEntries),
		GetEntry:                   s.requirePermission("lab.read", labHandler.HandleGetEntry),
		GetEntryByModelAndCalcutta: s.requirePermission("lab.read", labHandler.HandleGetEntryByModelAndCalcutta),
		ListEvaluations:            s.requirePermission("lab.read", labHandler.HandleListEvaluations),
		GetEvaluation:              s.requirePermission("lab.read", labHandler.HandleGetEvaluation),
		GetEvaluationEntryResults:  s.requirePermission("lab.read", labHandler.HandleGetEvaluationEntryResults),
		GetEvaluationEntryProfile:  s.requirePermission("lab.read", labHandler.HandleGetEvaluationEntryProfile),
	})

	s.registerCalcuttaCoManagerRoutes(r)
	s.registerTournamentModeratorRoutes(r)
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
	r.HandleFunc("/api/tournaments/{tournamentId}/bracket/games/{gameId}/winner", s.requirePermissionWithScope("tournament.game.write", "tournament", "tournamentId", s.selectWinnerHandler)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/tournaments/{tournamentId}/bracket/games/{gameId}/winner", s.requirePermissionWithScope("tournament.game.write", "tournament", "tournamentId", s.unselectWinnerHandler)).Methods("DELETE", "OPTIONS")
}

func (s *Server) registerPortfolioRoutes(r *mux.Router) {
	r.HandleFunc("/api/portfolios/{id}/teams", s.portfolioTeamsHandler).Methods("GET")
}

func (s *Server) registerCalcuttaCoManagerRoutes(r *mux.Router) {
	r.HandleFunc("/api/calcuttas/{id}/co-managers", s.requirePermissionWithScope("calcutta.config.write", "calcutta", "id", s.listCalcuttaCoManagersHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/calcuttas/{id}/co-managers", s.requirePermissionWithScope("calcutta.config.write", "calcutta", "id", s.grantCalcuttaCoManagerHandler)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/calcuttas/{id}/co-managers/{userId}", s.requirePermissionWithScope("calcutta.config.write", "calcutta", "id", s.revokeCalcuttaCoManagerHandler)).Methods("DELETE", "OPTIONS")
}

func (s *Server) registerTournamentModeratorRoutes(r *mux.Router) {
	r.HandleFunc("/api/tournaments/{id}/moderators", s.requirePermission("tournament.game.write", s.listTournamentModeratorsHandler)).Methods("GET", "OPTIONS")
	r.HandleFunc("/api/tournaments/{id}/moderators", s.requirePermission("tournament.game.write", s.grantTournamentModeratorHandler)).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/tournaments/{id}/moderators/{userId}", s.requirePermission("tournament.game.write", s.revokeTournamentModeratorHandler)).Methods("DELETE", "OPTIONS")
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
