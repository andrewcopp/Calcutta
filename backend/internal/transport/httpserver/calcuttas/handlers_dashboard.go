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

	scoringRules, err := h.app.Calcutta.GetScoringRules(r.Context(), calcuttaID)
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

		// Best-effort prediction loading: load all checkpoint batches
		checkpoints := h.app.Prediction.LoadCheckpointPredictions(r.Context(), calcutta.TournamentID)

		rules := make([]scoring.Rule, len(scoringRules))
		for i, sr := range scoringRules {
			rules[i] = scoring.Rule{WinIndex: sr.WinIndex, PointsAwarded: sr.PointsAwarded}
		}

		portfolioToEntry := prediction.BuildPortfolioToEntry(allPortfolios)
		ptInputs := prediction.ToPortfolioTeamInputs(allPortfolioTeams)
		ttInputs := prediction.ToTournamentTeamInputs(tournamentTeams)

		if proj := prediction.ComputeEntryProjections(checkpoints, rules, portfolioToEntry, ptInputs, ttInputs); proj != nil {
			for entryID, ev := range proj.EV {
				if s, ok := standingsByID[entryID]; ok {
					v := ev
					s.ExpectedValue = &v
				}
			}
			for entryID, fav := range proj.Favorites {
				if s, ok := standingsByID[entryID]; ok {
					v := fav
					s.ProjectedFavorites = &v
				}
			}
		}

		resp.Entries = dtos.NewEntryListResponse(entries, standingsByID)
		resp.EntryTeams = dtos.NewEntryTeamListResponse(allEntryTeams)
		resp.Portfolios = dtos.NewPortfolioListResponse(allPortfolios)
		resp.PortfolioTeams = dtos.NewPortfolioTeamListResponse(allPortfolioTeams)
		resp.RoundStandings = computeRoundStandings(entries, allPortfolios, allPortfolioTeams, tournamentTeams, scoringRules, payouts, checkpoints)

		bracket, err := h.app.Bracket.GetBracket(r.Context(), calcutta.TournamentID)
		if err == nil && bracket != nil {
			if ffOutcomes := calcuttaapp.ComputeFinalFourOutcomes(bracket, entries, allPortfolios, allPortfolioTeams, tournamentTeams, scoringRules, payouts); ffOutcomes != nil {
				ffResponses := make([]*dtos.FinalFourOutcomeResponse, len(ffOutcomes))
				for i, o := range ffOutcomes {
					standingEntries := make([]*dtos.RoundStandingEntry, len(o.Standings))
					for j, s := range o.Standings {
						standingEntries[j] = &dtos.RoundStandingEntry{
							EntryID:        s.EntryID,
							TotalPoints:    s.TotalPoints,
							FinishPosition: s.FinishPosition,
							IsTied:         s.IsTied,
							PayoutCents:    s.PayoutCents,
							InTheMoney:     s.InTheMoney,
						}
					}
					ffResponses[i] = &dtos.FinalFourOutcomeResponse{
						Semifinal1Winner: dtos.NewFinalFourTeam(o.Semifinal1Winner),
						Semifinal2Winner: dtos.NewFinalFourTeam(o.Semifinal2Winner),
						Champion:         dtos.NewFinalFourTeam(o.Champion),
						RunnerUp:         dtos.NewFinalFourTeam(o.RunnerUp),
						Entries:          standingEntries,
					}
				}
				resp.FinalFourOutcomes = ffResponses
			}
		}
	}

	response.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) listCalcuttasWithRankings(w http.ResponseWriter, r *http.Request, userID string, calcuttas []*models.Calcutta) {
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

	response.WriteJSON(w, http.StatusOK, map[string]any{"items": results})
}

// computeRoundStandings computes standings at each round cap (0 through maxRound).
// This lets the frontend show "as of" standings for any point in the tournament.
// Each cap uses the checkpoint batch with the highest throughRound <= cap, so
// projections change per round when multiple checkpoint batches exist.
func computeRoundStandings(
	entries []*models.CalcuttaEntry,
	portfolios []*models.CalcuttaPortfolio,
	portfolioTeams []*models.CalcuttaPortfolioTeam,
	tournamentTeams []*models.TournamentTeam,
	scoringRules []*models.ScoringRule,
	payouts []*models.CalcuttaPayout,
	checkpoints []prediction.CheckpointData,
) []*dtos.RoundStandingGroup {
	if len(scoringRules) == 0 {
		return []*dtos.RoundStandingGroup{}
	}

	rules := make([]scoring.Rule, len(scoringRules))
	maxRound := 0
	for i, sr := range scoringRules {
		rules[i] = scoring.Rule{WinIndex: sr.WinIndex, PointsAwarded: sr.PointsAwarded}
		if sr.WinIndex > maxRound {
			maxRound = sr.WinIndex
		}
	}

	teamByID := make(map[string]*models.TournamentTeam, len(tournamentTeams))
	for _, tt := range tournamentTeams {
		teamByID[tt.ID] = tt
	}

	portfolioToEntry := prediction.BuildPortfolioToEntry(portfolios)
	ptInputs := prediction.ToPortfolioTeamInputs(portfolioTeams)
	ttInputs := prediction.ToTournamentTeamInputs(tournamentTeams)

	groups := make([]*dtos.RoundStandingGroup, 0, maxRound+1)
	for cap := 0; cap <= maxRound; cap++ {
		pointsByEntry := make(map[string]float64)
		for _, pt := range portfolioTeams {
			team := teamByID[pt.TeamID]
			if team == nil {
				continue
			}
			entryID := portfolioToEntry[pt.PortfolioID]
			if entryID == "" {
				continue
			}
			capped := prediction.ProgressAtRound(team.Wins, team.Byes, cap)
			teamPoints := scoring.PointsForProgress(rules, capped, 0)
			pointsByEntry[entryID] += pt.OwnershipPercentage * float64(teamPoints)
		}

		standings := calcuttaapp.ComputeStandings(entries, pointsByEntry, payouts)
		proj := prediction.ComputeRoundProjections(checkpoints, rules, portfolioToEntry, ptInputs, ttInputs, cap)

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
			if proj != nil {
				ev := proj.EV[s.EntryID]
				entry.ExpectedValue = &ev
				fav := proj.Favorites[s.EntryID]
				entry.ProjectedFavorites = &fav
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

