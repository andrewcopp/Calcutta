package pools

import (
	"net/http"
	"time"

	poolapp "github.com/andrewcopp/Calcutta/backend/internal/app/pool"
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
	poolID := vars["id"]
	if poolID == "" {
		httperr.Write(w, r, http.StatusBadRequest, "validation_error", "Pool ID is required", "id")
		return
	}

	pool, err := h.app.Pool.GetPoolByID(r.Context(), poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	userID := ""
	if h.authUserID != nil {
		userID = h.authUserID(r.Context())
	}

	participantIDs, err := h.app.Pool.GetDistinctUserIDsByPool(r.Context(), poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	decision, err := policy.CanViewPool(r.Context(), h.authz, userID, pool, participantIDs)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	if !decision.Allowed {
		httperr.Write(w, r, decision.Status, decision.Code, decision.Message, "")
		return
	}

	portfolios, standings, err := h.app.Pool.GetPortfolios(r.Context(), poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	standingsByID := make(map[string]*models.PortfolioStanding, len(standings))
	for _, s := range standings {
		standingsByID[s.PortfolioID] = s
	}

	tournament, err := h.app.Tournament.GetByID(r.Context(), pool.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	schools, err := h.app.School.List(r.Context())
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	tournamentTeams, err := h.app.Tournament.GetTeams(r.Context(), pool.TournamentID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	tournamentTeamResponses := make([]*dtos.TournamentTeamResponse, 0, len(tournamentTeams))
	for _, team := range tournamentTeams {
		tournamentTeamResponses = append(tournamentTeamResponses, dtos.NewTournamentTeamResponse(team, team.School))
	}

	scoringRules, err := h.app.Pool.GetScoringRules(r.Context(), poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	payouts, err := h.app.Pool.GetPayouts(r.Context(), poolID)
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}

	investingOpen := !tournament.HasStarted(time.Now())

	var currentUserPortfolio *models.Portfolio
	for _, portfolio := range portfolios {
		if portfolio.UserID != nil && *portfolio.UserID == userID {
			currentUserPortfolio = portfolio
			break
		}
	}

	resp := &dtos.PoolDashboardResponse{
		Pool:                 dtos.NewPoolResponse(pool),
		TournamentStartingAt: tournament.StartingAt,
		InvestingOpen:        investingOpen,
		TotalPortfolios:      len(portfolios),
		Abilities:            computeAbilities(r.Context(), h.authz, userID, pool),
		ScoringRules:         dtos.NewScoringRuleListResponse(scoringRules),
		Schools:              dtos.NewSchoolListResponse(schools),
		TournamentTeams:      tournamentTeamResponses,
		RoundStandings:       []*dtos.RoundStandingGroup{},
	}

	if currentUserPortfolio != nil {
		resp.CurrentUserPortfolio = dtos.NewPortfolioResponse(currentUserPortfolio, standingsByID[currentUserPortfolio.ID])
	}

	if investingOpen {
		resp.Portfolios = []*dtos.PortfolioResponse{}
		resp.Investments = []*dtos.InvestmentResponse{}
		resp.OwnershipSummaries = []*dtos.OwnershipSummaryResponse{}
		resp.OwnershipDetails = []*dtos.OwnershipDetailResponse{}
	} else {
		portfolioIDs := make([]string, 0, len(portfolios))
		for _, portfolio := range portfolios {
			portfolioIDs = append(portfolioIDs, portfolio.ID)
		}

		investmentsByPortfolio, err := h.app.Pool.GetInvestmentsByPortfolioIDs(r.Context(), portfolioIDs)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}

		ownershipByPortfolio, err := h.app.Pool.GetOwnershipSummariesByPortfolioIDs(r.Context(), portfolioIDs)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}

		var allInvestments []*models.Investment
		for _, investments := range investmentsByPortfolio {
			allInvestments = append(allInvestments, investments...)
		}

		var allOwnershipSummaries []*models.OwnershipSummary
		for _, summaries := range ownershipByPortfolio {
			allOwnershipSummaries = append(allOwnershipSummaries, summaries...)
		}

		ownershipDetailsByPortfolio, err := h.app.Pool.GetOwnershipDetailsByPortfolioIDs(r.Context(), portfolioIDs)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}

		var allOwnershipDetails []*models.OwnershipDetail
		for _, details := range ownershipDetailsByPortfolio {
			allOwnershipDetails = append(allOwnershipDetails, details...)
		}

		// Best-effort prediction loading: load all checkpoint batches
		checkpoints := h.app.Prediction.LoadCheckpointPredictions(r.Context(), pool.TournamentID)

		rules := make([]scoring.Rule, len(scoringRules))
		for i, sr := range scoringRules {
			rules[i] = scoring.Rule{WinIndex: sr.WinIndex, PointsAwarded: sr.PointsAwarded}
		}

		ownershipToPortfolio := prediction.BuildSummaryToPortfolioMap(allOwnershipSummaries)
		odInputs := prediction.ToPortfolioTeamInputs(allOwnershipDetails)
		ttInputs := prediction.ToTournamentTeamInputs(tournamentTeams)

		if proj := prediction.ComputeEntryProjections(checkpoints, rules, ownershipToPortfolio, odInputs, ttInputs); proj != nil {
			for portfolioID, ev := range proj.EV {
				if s, ok := standingsByID[portfolioID]; ok {
					v := ev
					s.ExpectedValue = &v
				}
			}
			for portfolioID, fav := range proj.Favorites {
				if s, ok := standingsByID[portfolioID]; ok {
					v := fav
					s.ProjectedFavorites = &v
				}
			}
		}

		resp.Portfolios = dtos.NewPortfolioListResponse(portfolios, standingsByID)
		resp.Investments = dtos.NewInvestmentListResponse(allInvestments)
		resp.OwnershipSummaries = dtos.NewOwnershipSummaryListResponse(allOwnershipSummaries)
		resp.OwnershipDetails = dtos.NewOwnershipDetailListResponse(allOwnershipDetails)
		resp.RoundStandings = computeRoundStandings(portfolios, allOwnershipSummaries, allOwnershipDetails, tournamentTeams, scoringRules, payouts, checkpoints)

		bracket, err := h.app.Bracket.GetBracket(r.Context(), pool.TournamentID)
		if err == nil && bracket != nil {
			if ffOutcomes := poolapp.ComputeFinalFourOutcomes(bracket, portfolios, allOwnershipSummaries, allOwnershipDetails, tournamentTeams, scoringRules, payouts); ffOutcomes != nil {
				ffResponses := make([]*dtos.FinalFourOutcomeResponse, len(ffOutcomes))
				for i, o := range ffOutcomes {
					standingEntries := make([]*dtos.RoundStandingEntry, len(o.Standings))
					for j, s := range o.Standings {
						standingEntries[j] = &dtos.RoundStandingEntry{
							PortfolioID:    s.PortfolioID,
							TotalReturns:   s.TotalReturns,
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

func (h *Handler) listPoolsWithRankings(w http.ResponseWriter, r *http.Request, userID string, pools []*models.Pool) {
	tournaments, err := h.app.Tournament.List(r.Context())
	if err != nil {
		httperr.WriteFromErr(w, r, err, h.authUserID)
		return
	}
	tournamentByID := make(map[string]*models.Tournament, len(tournaments))
	for i := range tournaments {
		tournamentByID[tournaments[i].ID] = &tournaments[i]
	}

	results := make([]*dtos.PoolWithRankingResponse, 0, len(pools))
	for _, pool := range pools {
		item := &dtos.PoolWithRankingResponse{
			PoolResponse: dtos.NewPoolResponse(pool),
		}

		if tournament, ok := tournamentByID[pool.TournamentID]; ok {
			item.TournamentStartingAt = tournament.StartingAt
		}

		portfolios, standings, err := h.app.Pool.GetPortfolios(r.Context(), pool.ID)
		if err != nil {
			httperr.WriteFromErr(w, r, err, h.authUserID)
			return
		}

		var userPortfolio *models.Portfolio
		for _, portfolio := range portfolios {
			if portfolio.UserID != nil && *portfolio.UserID == userID {
				userPortfolio = portfolio
				break
			}
		}

		if userPortfolio != nil {
			item.HasPortfolio = true

			// standings are already sorted by returns desc
			rank := 1
			var userReturns float64
			for i, s := range standings {
				if s.PortfolioID == userPortfolio.ID {
					rank = i + 1
					userReturns = s.TotalReturns
					break
				}
			}

			item.Ranking = &dtos.PoolRankingResponse{
				Rank:            rank,
				TotalPortfolios: len(portfolios),
				Returns:         userReturns,
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
	portfolios []*models.Portfolio,
	ownershipSummaries []*models.OwnershipSummary,
	ownershipDetails []*models.OwnershipDetail,
	tournamentTeams []*models.TournamentTeam,
	scoringRules []*models.ScoringRule,
	payouts []*models.PoolPayout,
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

	ownershipToPortfolio := prediction.BuildSummaryToPortfolioMap(ownershipSummaries)
	odInputs := prediction.ToPortfolioTeamInputs(ownershipDetails)
	ttInputs := prediction.ToTournamentTeamInputs(tournamentTeams)

	groups := make([]*dtos.RoundStandingGroup, 0, maxRound+1)
	for cap := 0; cap <= maxRound; cap++ {
		returnsByPortfolio := make(map[string]float64)
		for _, od := range ownershipDetails {
			team := teamByID[od.TeamID]
			if team == nil {
				continue
			}
			portfolioID := ownershipToPortfolio[od.PortfolioID]
			if portfolioID == "" {
				continue
			}
			capped := prediction.ProgressAtRound(team.Wins, team.Byes, cap)
			teamPoints := scoring.PointsForProgress(rules, capped, 0)
			returnsByPortfolio[portfolioID] += od.OwnershipPercentage * float64(teamPoints)
		}

		standings := poolapp.ComputeStandings(portfolios, returnsByPortfolio, payouts)
		proj := prediction.ComputeRoundProjections(checkpoints, rules, ownershipToPortfolio, odInputs, ttInputs, cap)

		standingEntries := make([]*dtos.RoundStandingEntry, len(standings))
		for i, s := range standings {
			entry := &dtos.RoundStandingEntry{
				PortfolioID:    s.PortfolioID,
				TotalReturns:   s.TotalReturns,
				FinishPosition: s.FinishPosition,
				IsTied:         s.IsTied,
				PayoutCents:    s.PayoutCents,
				InTheMoney:     s.InTheMoney,
			}
			if proj != nil {
				ev := proj.EV[s.PortfolioID]
				entry.ExpectedValue = &ev
				fav := proj.Favorites[s.PortfolioID]
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
