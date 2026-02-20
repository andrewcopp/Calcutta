package calcuttas

import (
	"net/http"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/policy"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/dtos"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/httperr"
	"github.com/andrewcopp/Calcutta/backend/internal/transport/httpserver/response"
	"github.com/gorilla/mux"
)

func (h *Handler) HandleGetDashboard(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	if calcuttaID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Calcutta ID is required", "id")
		return
	}

	calcutta, err := h.app.Calcutta.GetCalcuttaByID(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	userID := ""
	if h.authUserID != nil {
		userID = h.authUserID(r.Context())
	}

	participantIDs, err := h.app.Calcutta.GetDistinctUserIDsByCalcutta(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanViewCalcutta(r.Context(), h.authz, userID, calcutta, participantIDs)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	entries, err := h.app.Calcutta.GetEntries(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	schools, err := h.app.School.List(r.Context())
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	tournamentTeams, err := h.app.Tournament.GetTeams(r.Context(), calcutta.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	var allEntryTeams []*models.CalcuttaEntryTeam
	var allPortfolios []*models.CalcuttaPortfolio
	for _, entry := range entries {
		entryTeams, err := h.app.Calcutta.GetEntryTeams(r.Context(), entry.ID)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}
		allEntryTeams = append(allEntryTeams, entryTeams...)

		portfolios, err := h.app.Calcutta.GetPortfoliosByEntry(r.Context(), entry.ID)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}
		allPortfolios = append(allPortfolios, portfolios...)
	}

	var allPortfolioTeams []*models.CalcuttaPortfolioTeam
	for _, portfolio := range allPortfolios {
		portfolioTeams, err := h.app.Calcutta.GetPortfolioTeams(r.Context(), portfolio.ID)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}
		allPortfolioTeams = append(allPortfolioTeams, portfolioTeams...)
	}

	tournamentTeamResponses := make([]*dtos.TournamentTeamResponse, 0, len(tournamentTeams))
	for _, team := range tournamentTeams {
		tournamentTeamResponses = append(tournamentTeamResponses, dtos.NewTournamentTeamResponse(team, team.School))
	}

	tournament, err := h.app.Tournament.GetByID(r.Context(), calcutta.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	resp := &dtos.CalcuttaDashboardResponse{
		Calcutta:             dtos.NewCalcuttaResponse(calcutta),
		TournamentStartingAt: tournament.StartingAt,
		Abilities:            computeAbilities(r.Context(), h.authz, userID, calcutta),
		Entries:              dtos.NewEntryListResponse(entries),
		EntryTeams:           dtos.NewEntryTeamListResponse(allEntryTeams),
		Portfolios:           dtos.NewPortfolioListResponse(allPortfolios),
		PortfolioTeams:       dtos.NewPortfolioTeamListResponse(allPortfolioTeams),
		Schools:              dtos.NewSchoolListResponse(schools),
		TournamentTeams:      tournamentTeamResponses,
	}
	response.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) HandleListCalcuttasWithRankings(w http.ResponseWriter, r *http.Request) {
	userID := ""
	if h.authUserID != nil {
		userID = h.authUserID(r.Context())
	}
	if userID == "" {
		httperr.Write(w, r, http.StatusUnauthorized, "unauthorized", "Authentication required", "")
		return
	}

	isAdmin, err := h.authz.HasPermission(r.Context(), userID, "global", "", "calcutta.config.write")
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	var calcuttas []*models.Calcutta
	if isAdmin {
		calcuttas, err = h.app.Calcutta.GetAllCalcuttas(r.Context())
	} else {
		calcuttas, err = h.app.Calcutta.GetCalcuttasByUser(r.Context(), userID)
	}
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	results := make([]*dtos.CalcuttaWithRankingResponse, 0, len(calcuttas))
	for _, calcutta := range calcuttas {
		item := &dtos.CalcuttaWithRankingResponse{
			CalcuttaResponse: dtos.NewCalcuttaResponse(calcutta),
		}

		tournament, err := h.app.Tournament.GetByID(r.Context(), calcutta.TournamentID)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}
		item.TournamentStartingAt = tournament.StartingAt

		entries, err := h.app.Calcutta.GetEntries(r.Context(), calcutta.ID)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}

		var userEntry *models.CalcuttaEntry
		for _, entry := range entries {
			if entry.UserID != nil && *entry.UserID == userID {
				userEntry = entry
				break
			}
		}

		if userEntry != nil {
			item.HasEntry = true

			sorted := make([]*models.CalcuttaEntry, len(entries))
			copy(sorted, entries)
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].TotalPoints > sorted[j].TotalPoints
			})

			rank := 1
			for i, e := range sorted {
				if e.ID == userEntry.ID {
					rank = i + 1
					break
				}
			}

			item.Ranking = &dtos.CalcuttaRankingResponse{
				Rank:         rank,
				TotalEntries: len(entries),
				Points:       userEntry.TotalPoints,
			}
		}

		results = append(results, item)
	}

	response.WriteJSON(w, http.StatusOK, results)
}
