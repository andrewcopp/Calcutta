package calcutta

import (
	"github.com/andrewcopp/Calcutta/backend/internal/app/prediction"
	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// FinalFourOutcome represents one of the 8 possible championship outcomes.
type FinalFourOutcome struct {
	Semifinal1Winner *models.BracketTeam
	Semifinal2Winner *models.BracketTeam
	Champion         *models.BracketTeam
	RunnerUp         *models.BracketTeam
	Standings        []*models.EntryStanding
}

// ComputeFinalFourOutcomes computes standings for all 8 possible championship
// outcomes. Returns nil if the Final Four field is not yet set (i.e. both
// semifinal games don't have both teams populated).
func ComputeFinalFourOutcomes(
	bracket *models.BracketStructure,
	entries []*models.CalcuttaEntry,
	portfolios []*models.CalcuttaPortfolio,
	portfolioTeams []*models.CalcuttaPortfolioTeam,
	tournamentTeams []*models.TournamentTeam,
	rounds []*models.CalcuttaRound,
	payouts []*models.CalcuttaPayout,
) []*FinalFourOutcome {
	if bracket == nil {
		return nil
	}

	semi1 := bracket.Games["final_four-1"]
	semi2 := bracket.Games["final_four-2"]
	if semi1 == nil || semi2 == nil {
		return nil
	}
	if semi1.Team1 == nil || semi1.Team2 == nil || semi2.Team1 == nil || semi2.Team2 == nil {
		return nil
	}

	semis := [2][2]*models.BracketTeam{
		{semi1.Team1, semi1.Team2},
		{semi2.Team1, semi2.Team2},
	}

	rules := make([]scoring.Rule, len(rounds))
	for i, rd := range rounds {
		rules[i] = scoring.Rule{WinIndex: rd.Round, PointsAwarded: rd.Points}
	}

	teamByID := make(map[string]*models.TournamentTeam, len(tournamentTeams))
	for _, tt := range tournamentTeams {
		teamByID[tt.ID] = tt
	}

	portfolioToEntry := prediction.BuildPortfolioToEntry(portfolios)

	var outcomes []*FinalFourOutcome

	for _, s1Winner := range semis[0] {
		for _, s2Winner := range semis[1] {
			for _, champion := range []*models.BracketTeam{s1Winner, s2Winner} {
				runnerUp := s2Winner
				if champion == s2Winner {
					runnerUp = s1Winner
				}

				hypotheticalWins := buildHypotheticalWins(semis, s1Winner, s2Winner, champion, teamByID)
				pointsByEntry := computePointsByEntry(portfolioTeams, teamByID, hypotheticalWins, rules, portfolioToEntry)
				standings := ComputeStandings(entries, pointsByEntry, payouts)

				outcomes = append(outcomes, &FinalFourOutcome{
					Semifinal1Winner: s1Winner,
					Semifinal2Winner: s2Winner,
					Champion:         champion,
					RunnerUp:         runnerUp,
					Standings:        standings,
				})
			}
		}
	}

	return outcomes
}

// buildHypotheticalWins returns a map of teamID -> additional wins beyond current state
// for the 3 remaining games in a given Final Four outcome.
func buildHypotheticalWins(
	semis [2][2]*models.BracketTeam,
	s1Winner, s2Winner, champion *models.BracketTeam,
	teamByID map[string]*models.TournamentTeam,
) map[string]int {
	extra := make(map[string]int)

	// Both semifinal winners get +1 win (for winning the semifinal)
	extra[s1Winner.TeamID]++
	extra[s2Winner.TeamID]++

	// Champion gets another +1 win (for winning the final)
	extra[champion.TeamID]++

	return extra
}

// eliteEightCap is the maximum progress (wins + byes) before the Final Four.
// We cap here so hypothetical wins are added on top of pre-Final-Four state,
// not on top of actual results that may already include FF/Championship wins.
const eliteEightCap = 5

// computePointsByEntry computes total points per entry for a hypothetical outcome.
func computePointsByEntry(
	portfolioTeams []*models.CalcuttaPortfolioTeam,
	teamByID map[string]*models.TournamentTeam,
	hypotheticalWins map[string]int,
	rules []scoring.Rule,
	portfolioToEntry map[string]string,
) map[string]float64 {
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

		capped := prediction.ProgressAtRound(team.Wins, team.Byes, eliteEightCap)
		wins := capped + hypotheticalWins[pt.TeamID]
		teamPoints := scoring.PointsForProgress(rules, wins, 0)
		pointsByEntry[entryID] += pt.OwnershipPercentage * float64(teamPoints)
	}

	return pointsByEntry
}
