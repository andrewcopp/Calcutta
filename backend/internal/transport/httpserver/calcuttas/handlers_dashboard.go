package calcuttas

import (
	"net/http"
	"time"

	calcuttaapp "github.com/andrewcopp/Calcutta/backend/internal/app/calcutta"
	"github.com/andrewcopp/Calcutta/backend/internal/app/prediction"
	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
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

	entries, standings, err := h.app.Calcutta.GetEntries(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	standingsByID := make(map[string]*models.EntryStanding, len(standings))
	for _, s := range standings {
		standingsByID[s.EntryID] = s
	}

	tournament, err := h.app.Tournament.GetByID(r.Context(), calcutta.TournamentID)
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

	tournamentTeamResponses := make([]*dtos.TournamentTeamResponse, 0, len(tournamentTeams))
	for _, team := range tournamentTeams {
		tournamentTeamResponses = append(tournamentTeamResponses, dtos.NewTournamentTeamResponse(team, team.School))
	}

	rounds, err := h.app.Calcutta.GetRounds(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	payouts, err := h.app.Calcutta.GetPayouts(r.Context(), calcuttaID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	biddingOpen := !tournament.HasStarted(time.Now())

	var currentUserEntry *models.CalcuttaEntry
	for _, entry := range entries {
		if entry.UserID != nil && *entry.UserID == userID {
			currentUserEntry = entry
			break
		}
	}

	resp := &dtos.CalcuttaDashboardResponse{
		Calcutta:             dtos.NewCalcuttaResponse(calcutta),
		TournamentStartingAt: tournament.StartingAt,
		BiddingOpen:          biddingOpen,
		TotalEntries:         len(entries),
		Abilities:            computeAbilities(r.Context(), h.authz, userID, calcutta),
		Schools:              dtos.NewSchoolListResponse(schools),
		TournamentTeams:      tournamentTeamResponses,
		RoundStandings:       []*dtos.RoundStandingGroup{},
	}

	if currentUserEntry != nil {
		userTeams, err := h.app.Calcutta.GetEntryTeams(r.Context(), currentUserEntry.ID)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}
		currentUserEntry.Status = calcuttaapp.DeriveEntryStatus(userTeams)
		resp.CurrentUserEntry = dtos.NewEntryResponse(currentUserEntry, standingsByID[currentUserEntry.ID])
	}

	if biddingOpen {
		resp.Entries = []*dtos.EntryResponse{}
		resp.EntryTeams = []*dtos.EntryTeamResponse{}
		resp.Portfolios = []*dtos.PortfolioResponse{}
		resp.PortfolioTeams = []*dtos.PortfolioTeamResponse{}
	} else {
		entryIDs := make([]string, 0, len(entries))
		for _, entry := range entries {
			entryIDs = append(entryIDs, entry.ID)
		}

		entryTeamsByEntry, err := h.app.Calcutta.GetEntryTeamsByEntryIDs(r.Context(), entryIDs)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}

		portfoliosByEntry, err := h.app.Calcutta.GetPortfoliosByEntryIDs(r.Context(), entryIDs)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}

		var allEntryTeams []*models.CalcuttaEntryTeam
		for _, teams := range entryTeamsByEntry {
			allEntryTeams = append(allEntryTeams, teams...)
		}

		var allPortfolios []*models.CalcuttaPortfolio
		portfolioIDs := make([]string, 0)
		for _, portfolios := range portfoliosByEntry {
			for _, p := range portfolios {
				allPortfolios = append(allPortfolios, p)
				portfolioIDs = append(portfolioIDs, p.ID)
			}
		}

		portfolioTeamsByPortfolio, err := h.app.Calcutta.GetPortfolioTeamsByPortfolioIDs(r.Context(), portfolioIDs)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}

		var allPortfolioTeams []*models.CalcuttaPortfolioTeam
		for _, teams := range portfolioTeamsByPortfolio {
			allPortfolioTeams = append(allPortfolioTeams, teams...)
		}

		for _, entry := range entries {
			entry.Status = calcuttaapp.DeriveEntryStatus(entryTeamsByEntry[entry.ID])
		}

		// Best-effort prediction loading for projected EV
		var ptvByTeam map[string]prediction.PredictedTeamValue
		batchID, found, batchErr := h.app.Prediction.GetLatestBatchID(r.Context(), calcutta.TournamentID)
		if batchErr == nil && found {
			teamValues, tvErr := h.app.Prediction.GetTeamValues(r.Context(), batchID)
			if tvErr == nil {
				ptvByTeam = make(map[string]prediction.PredictedTeamValue, len(teamValues))
				for _, tv := range teamValues {
					ptvByTeam[tv.TeamID] = tv
				}
			}
		}

		scoringRules := make([]scoring.Rule, len(rounds))
		for i, rd := range rounds {
			scoringRules[i] = scoring.Rule{WinIndex: rd.Round, PointsAwarded: rd.Points}
		}

		attachProjectedEV(ptvByTeam, scoringRules, standingsByID, allPortfolios, allPortfolioTeams, tournamentTeams)

		resp.Entries = dtos.NewEntryListResponse(entries, standingsByID)
		resp.EntryTeams = dtos.NewEntryTeamListResponse(allEntryTeams)
		resp.Portfolios = dtos.NewPortfolioListResponse(allPortfolios)
		resp.PortfolioTeams = dtos.NewPortfolioTeamListResponse(allPortfolioTeams)
		resp.RoundStandings = computeRoundStandings(entries, allPortfolios, allPortfolioTeams, tournamentTeams, rounds, payouts, ptvByTeam)
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

	tournaments, err := h.app.Tournament.List(r.Context())
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	tournamentByID := make(map[string]*models.Tournament, len(tournaments))
	for i := range tournaments {
		tournamentByID[tournaments[i].ID] = &tournaments[i]
	}

	results := make([]*dtos.CalcuttaWithRankingResponse, 0, len(calcuttas))
	for _, calcutta := range calcuttas {
		item := &dtos.CalcuttaWithRankingResponse{
			CalcuttaResponse: dtos.NewCalcuttaResponse(calcutta),
		}

		if tournament, ok := tournamentByID[calcutta.TournamentID]; ok {
			item.TournamentStartingAt = tournament.StartingAt
		}

		entries, standings, err := h.app.Calcutta.GetEntries(r.Context(), calcutta.ID)
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

			// standings are already sorted by points desc
			rank := 1
			var userPoints float64
			for i, s := range standings {
				if s.EntryID == userEntry.ID {
					rank = i + 1
					userPoints = s.TotalPoints
					break
				}
			}

			item.Ranking = &dtos.CalcuttaRankingResponse{
				Rank:         rank,
				TotalEntries: len(entries),
				Points:       userPoints,
			}
		}

		results = append(results, item)
	}

	response.WriteJSON(w, http.StatusOK, results)
}

// attachProjectedEV computes and attaches projected EV to each standing using prediction data.
// If ptvByTeam is nil, this is a no-op.
func attachProjectedEV(
	ptvByTeam map[string]prediction.PredictedTeamValue,
	rules []scoring.Rule,
	standingsByID map[string]*models.EntryStanding,
	portfolios []*models.CalcuttaPortfolio,
	portfolioTeams []*models.CalcuttaPortfolioTeam,
	tournamentTeams []*models.TournamentTeam,
) {
	if ptvByTeam == nil {
		return
	}

	ttByID := make(map[string]*models.TournamentTeam, len(tournamentTeams))
	for _, tt := range tournamentTeams {
		ttByID[tt.ID] = tt
	}

	// Accumulate projected EV per portfolio from portfolio teams
	evByPortfolio := make(map[string]float64)
	for _, pt := range portfolioTeams {
		ptv, ok := ptvByTeam[pt.TeamID]
		if !ok {
			continue
		}
		tt := ttByID[pt.TeamID]
		if tt == nil {
			continue
		}
		tp := prediction.TeamProgress{
			Wins:         tt.Wins,
			Byes:         tt.Byes,
			IsEliminated: tt.IsEliminated,
		}
		teamEV := prediction.ProjectedTeamEV(ptv, rules, tp, 0)
		evByPortfolio[pt.PortfolioID] += pt.OwnershipPercentage * teamEV
	}

	// Map portfolioID -> entryID using already-loaded portfolios
	portfolioToEntry := make(map[string]string, len(portfolios))
	for _, p := range portfolios {
		portfolioToEntry[p.ID] = p.EntryID
	}

	entryEV := make(map[string]float64)
	for portfolioID, ev := range evByPortfolio {
		if entryID, ok := portfolioToEntry[portfolioID]; ok {
			entryEV[entryID] += ev
		}
	}

	for entryID, ev := range entryEV {
		if s, ok := standingsByID[entryID]; ok {
			v := ev
			s.ProjectedEV = &v
		}
	}
}

// computeRoundStandings computes standings at each round cap (0 through maxRound).
// This lets the frontend show "as of" standings for any point in the tournament.
// If ptvByTeam is non-nil, projected EV is computed at each cap using capped progress.
func computeRoundStandings(
	entries []*models.CalcuttaEntry,
	portfolios []*models.CalcuttaPortfolio,
	portfolioTeams []*models.CalcuttaPortfolioTeam,
	tournamentTeams []*models.TournamentTeam,
	rounds []*models.CalcuttaRound,
	payouts []*models.CalcuttaPayout,
	ptvByTeam map[string]prediction.PredictedTeamValue,
) []*dtos.RoundStandingGroup {
	if len(rounds) == 0 {
		return []*dtos.RoundStandingGroup{}
	}

	rules := make([]scoring.Rule, len(rounds))
	maxRound := 0
	for i, rd := range rounds {
		rules[i] = scoring.Rule{WinIndex: rd.Round, PointsAwarded: rd.Points}
		if rd.Round > maxRound {
			maxRound = rd.Round
		}
	}

	teamByID := make(map[string]*models.TournamentTeam, len(tournamentTeams))
	for _, tt := range tournamentTeams {
		teamByID[tt.ID] = tt
	}

	portfolioToEntry := make(map[string]string, len(portfolios))
	for _, p := range portfolios {
		portfolioToEntry[p.ID] = p.EntryID
	}

	groups := make([]*dtos.RoundStandingGroup, 0, maxRound+1)
	for cap := 0; cap <= maxRound; cap++ {
		pointsByEntry := make(map[string]float64)
		var evByEntry map[string]float64
		if ptvByTeam != nil {
			evByEntry = make(map[string]float64)
		}

		for _, pt := range portfolioTeams {
			team := teamByID[pt.TeamID]
			if team == nil {
				continue
			}
			entryID := portfolioToEntry[pt.PortfolioID]
			if entryID == "" {
				continue
			}
			progress := team.Wins + team.Byes
			capped := progress
			if capped > cap {
				capped = cap
			}
			teamPoints := scoring.PointsForProgress(rules, capped, 0)
			pointsByEntry[entryID] += pt.OwnershipPercentage * float64(teamPoints)

			if evByEntry != nil {
				ptv, ok := ptvByTeam[pt.TeamID]
				if ok {
					isEliminated := team.IsEliminated && progress <= cap
					tp := prediction.TeamProgress{Wins: capped, Byes: 0, IsEliminated: isEliminated}
					evByEntry[entryID] += pt.OwnershipPercentage * prediction.ProjectedTeamEV(ptv, rules, tp, 0)
				}
			}
		}

		standings := calcuttaapp.ComputeStandings(entries, pointsByEntry, payouts)
		standingEntries := make([]*dtos.RoundStandingEntry, len(standings))
		for i, s := range standings {
			entry := &dtos.RoundStandingEntry{
				EntryID:        s.EntryID,
				TotalPoints:    s.TotalPoints,
				FinishPosition: s.FinishPosition,
				IsTied:         s.IsTied,
				PayoutCents:    s.PayoutCents,
				InTheMoney:     s.InTheMoney,
			}
			if evByEntry != nil {
				ev := evByEntry[s.EntryID]
				entry.ProjectedEV = &ev
			}
			standingEntries[i] = entry
		}

		groups = append(groups, &dtos.RoundStandingGroup{
			Round:   cap,
			Entries: standingEntries,
		})
	}

	return groups
}

