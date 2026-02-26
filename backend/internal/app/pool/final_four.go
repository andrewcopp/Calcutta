package pool

import (
	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// FinalFourOutcome represents one of the 8 possible championship outcomes.
type FinalFourOutcome struct {
	Semifinal1Winner *models.BracketTeam
	Semifinal2Winner *models.BracketTeam
	Champion         *models.BracketTeam
	RunnerUp         *models.BracketTeam
	Standings        []*models.PortfolioStanding
}

// ComputeFinalFourOutcomes computes standings for all 8 possible championship
// outcomes. Returns nil if the Final Four field is not yet set (i.e. both
// semifinal games don't have both teams populated).
func ComputeFinalFourOutcomes(
	bracket *models.BracketStructure,
	portfolios []*models.Portfolio,
	ownershipSummaries []*models.OwnershipSummary,
	ownershipDetails []*models.OwnershipDetail,
	tournamentTeams []*models.TournamentTeam,
	scoringRules []*models.ScoringRule,
	payouts []*models.PoolPayout,
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

	rules := make([]scoring.Rule, len(scoringRules))
	for i, sr := range scoringRules {
		rules[i] = scoring.Rule{WinIndex: sr.WinIndex, PointsAwarded: sr.PointsAwarded}
	}

	teamByID := make(map[string]*models.TournamentTeam, len(tournamentTeams))
	for _, tt := range tournamentTeams {
		teamByID[tt.ID] = tt
	}

	summaryToPortfolio := buildSummaryToPortfolioMap(ownershipSummaries)

	var outcomes []*FinalFourOutcome

	for _, s1Winner := range semis[0] {
		for _, s2Winner := range semis[1] {
			for _, champion := range []*models.BracketTeam{s1Winner, s2Winner} {
				runnerUp := s2Winner
				if champion == s2Winner {
					runnerUp = s1Winner
				}

				hypotheticalWins := buildHypotheticalWins(semis, s1Winner, s2Winner, champion, teamByID)
				returnsByPortfolio := computeReturnsByPortfolio(ownershipDetails, teamByID, hypotheticalWins, rules, summaryToPortfolio)
				standings := ComputeStandings(portfolios, returnsByPortfolio, payouts)

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

// buildSummaryToPortfolioMap builds an ownership summary ID -> portfolio ID lookup map.
func buildSummaryToPortfolioMap(summaries []*models.OwnershipSummary) map[string]string {
	m := make(map[string]string, len(summaries))
	for _, s := range summaries {
		m[s.ID] = s.PortfolioID
	}
	return m
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

// progressAtRound caps progress (wins + byes) at a given round threshold.
func progressAtRound(wins, byes, round int) int {
	total := wins + byes
	if total > round {
		return round
	}
	return total
}

// computeReturnsByPortfolio computes total returns per portfolio for a hypothetical outcome.
func computeReturnsByPortfolio(
	ownershipDetails []*models.OwnershipDetail,
	teamByID map[string]*models.TournamentTeam,
	hypotheticalWins map[string]int,
	rules []scoring.Rule,
	summaryToPortfolio map[string]string,
) map[string]float64 {
	returnsByPortfolio := make(map[string]float64)

	for _, od := range ownershipDetails {
		team := teamByID[od.TeamID]
		if team == nil {
			continue
		}
		portfolioID := summaryToPortfolio[od.PortfolioID]
		if portfolioID == "" {
			continue
		}

		capped := progressAtRound(team.Wins, team.Byes, eliteEightCap)
		wins := capped + hypotheticalWins[od.TeamID]
		teamReturns := scoring.PointsForProgress(rules, wins, 0)
		returnsByPortfolio[portfolioID] += od.OwnershipPercentage * float64(teamReturns)
	}

	return returnsByPortfolio
}
