package pool

import (
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// Helper to build a minimal BracketStructure with Final Four games.
func buildFinalFourBracket(semi1Team1, semi1Team2, semi2Team1, semi2Team2 *models.BracketTeam) *models.BracketStructure {
	return &models.BracketStructure{
		TournamentID: "tournament-1",
		Games: map[string]*models.BracketGame{
			"final_four-1": {
				GameID: "final_four-1",
				Round:  models.RoundFinalFour,
				Team1:  semi1Team1,
				Team2:  semi1Team2,
			},
			"final_four-2": {
				GameID: "final_four-2",
				Round:  models.RoundFinalFour,
				Team1:  semi2Team1,
				Team2:  semi2Team2,
			},
		},
	}
}

func bracketTeam(id, schoolID string, seed int, region string) *models.BracketTeam {
	return &models.BracketTeam{
		TeamID:   id,
		SchoolID: schoolID,
		Name:     id,
		Seed:     seed,
		Region:   region,
	}
}

func tournamentTeam(id string, wins, byes int, eliminated bool) *models.TournamentTeam {
	return &models.TournamentTeam{
		ID:           id,
		Wins:         wins,
		Byes:         byes,
		IsEliminated: eliminated,
	}
}

func testPortfolio(id string) *models.Portfolio {
	return &models.Portfolio{
		ID:        id,
		Name:      id,
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func ownershipSummary(id, portfolioID string) *models.OwnershipSummary {
	return &models.OwnershipSummary{ID: id, PortfolioID: portfolioID}
}

func ownershipDetail(portfolioID, teamID string, ownership float64) *models.OwnershipDetail {
	return &models.OwnershipDetail{
		PortfolioID:         portfolioID,
		TeamID:              teamID,
		OwnershipPercentage: ownership,
	}
}

func scoringRule(winIndex, pointsAwarded int) *models.ScoringRule {
	return &models.ScoringRule{WinIndex: winIndex, PointsAwarded: pointsAwarded}
}

func poolPayout(position, amountCents int) *models.PoolPayout {
	return &models.PoolPayout{Position: position, AmountCents: amountCents}
}

func TestThatNilReturnedWhenBracketIsNil(t *testing.T) {
	// GIVEN a nil bracket
	// WHEN computing Final Four outcomes
	result := ComputeFinalFourOutcomes(nil, nil, nil, nil, nil, nil, nil)

	// THEN nil is returned
	if result != nil {
		t.Error("expected nil when bracket is nil")
	}
}

func TestThatNilReturnedWhenSemifinalGameMissing(t *testing.T) {
	// GIVEN a bracket without final_four-2
	bracket := &models.BracketStructure{
		Games: map[string]*models.BracketGame{
			"final_four-1": {
				GameID: "final_four-1",
				Team1:  bracketTeam("A", "sa", 1, "East"),
				Team2:  bracketTeam("B", "sb", 2, "East"),
			},
		},
	}

	// WHEN computing Final Four outcomes
	result := ComputeFinalFourOutcomes(bracket, nil, nil, nil, nil, nil, nil)

	// THEN nil is returned
	if result != nil {
		t.Error("expected nil when semifinal game is missing")
	}
}

func TestThatNilReturnedWhenSemifinalTeamNotSet(t *testing.T) {
	// GIVEN a bracket where one semifinal team is nil
	bracket := buildFinalFourBracket(
		bracketTeam("A", "sa", 1, "East"),
		bracketTeam("B", "sb", 2, "West"),
		bracketTeam("C", "sc", 1, "South"),
		nil, // team not yet determined
	)

	// WHEN computing Final Four outcomes
	result := ComputeFinalFourOutcomes(bracket, nil, nil, nil, nil, nil, nil)

	// THEN nil is returned
	if result != nil {
		t.Error("expected nil when semifinal team is not set")
	}
}

func TestThatEightOutcomesReturnedForCompleteFinalFour(t *testing.T) {
	// GIVEN a complete Final Four bracket
	bracket := buildFinalFourBracket(
		bracketTeam("A", "sa", 1, "East"),
		bracketTeam("B", "sb", 2, "West"),
		bracketTeam("C", "sc", 1, "South"),
		bracketTeam("D", "sd", 2, "Midwest"),
	)

	portfolios := []*models.Portfolio{testPortfolio("p1")}
	summaries := []*models.OwnershipSummary{ownershipSummary("os1", "p1")}
	details := []*models.OwnershipDetail{ownershipDetail("os1", "A", 1.0)}
	tts := []*models.TournamentTeam{
		tournamentTeam("A", 5, 1, false),
		tournamentTeam("B", 5, 1, false),
		tournamentTeam("C", 5, 1, false),
		tournamentTeam("D", 5, 1, false),
	}
	rounds := []*models.ScoringRule{scoringRule(1, 1)}
	payouts := []*models.PoolPayout{}

	// WHEN computing Final Four outcomes
	result := ComputeFinalFourOutcomes(bracket, portfolios, summaries, details, tts, rounds, payouts)

	// THEN 8 outcomes are returned
	if len(result) != 8 {
		t.Errorf("expected 8 outcomes, got %d", len(result))
	}
}

func TestThatChampionGetsExtraWinsInScoring(t *testing.T) {
	// GIVEN a bracket with 4 FF teams, one portfolio owns team A at 100%
	bracket := buildFinalFourBracket(
		bracketTeam("A", "sa", 1, "East"),
		bracketTeam("B", "sb", 2, "West"),
		bracketTeam("C", "sc", 1, "South"),
		bracketTeam("D", "sd", 2, "Midwest"),
	)

	portfolios := []*models.Portfolio{testPortfolio("p1")}
	summaries := []*models.OwnershipSummary{ownershipSummary("os1", "p1")}
	details := []*models.OwnershipDetail{ownershipDetail("os1", "A", 1.0)}
	// All teams have 4 wins + 1 bye = 5 progress (through Elite 8)
	tts := []*models.TournamentTeam{
		tournamentTeam("A", 4, 1, false),
		tournamentTeam("B", 4, 1, false),
		tournamentTeam("C", 4, 1, false),
		tournamentTeam("D", 4, 1, false),
	}
	// Each round of progress = 10 points
	rounds := []*models.ScoringRule{
		scoringRule(1, 10), scoringRule(2, 10), scoringRule(3, 10), scoringRule(4, 10),
		scoringRule(5, 10), scoringRule(6, 10), scoringRule(7, 10),
	}
	payouts := []*models.PoolPayout{}

	result := ComputeFinalFourOutcomes(bracket, portfolios, summaries, details, tts, rounds, payouts)

	// Find outcomes where A is champion vs where A only wins semifinal
	var championReturns, semiOnlyReturns float64
	for _, o := range result {
		if o.Champion.TeamID == "A" {
			championReturns = o.Standings[0].TotalReturns
			break
		}
	}
	for _, o := range result {
		if o.Semifinal1Winner.TeamID == "A" && o.Champion.TeamID != "A" {
			semiOnlyReturns = o.Standings[0].TotalReturns
			break
		}
	}

	// THEN champion outcome gives more returns than semifinal-only outcome
	if championReturns <= semiOnlyReturns {
		t.Errorf("expected champion returns (%.2f) > semifinal-only returns (%.2f)", championReturns, semiOnlyReturns)
	}
}

func TestThatCompletedTournamentProducesSameResultsAsPreFinalFour(t *testing.T) {
	// GIVEN two setups: one with pre-Final-Four progress, one with completed tournament
	bracket := buildFinalFourBracket(
		bracketTeam("A", "sa", 1, "East"),
		bracketTeam("B", "sb", 2, "West"),
		bracketTeam("C", "sc", 1, "South"),
		bracketTeam("D", "sd", 2, "Midwest"),
	)

	portfolios := []*models.Portfolio{testPortfolio("p1")}
	summaries := []*models.OwnershipSummary{ownershipSummary("os1", "p1")}
	details := []*models.OwnershipDetail{ownershipDetail("os1", "A", 1.0)}
	rounds := []*models.ScoringRule{
		scoringRule(1, 10), scoringRule(2, 10), scoringRule(3, 10), scoringRule(4, 10),
		scoringRule(5, 10), scoringRule(6, 10), scoringRule(7, 10),
	}
	payouts := []*models.PoolPayout{}

	// Pre-Final-Four: all teams at 4 wins + 1 bye = 5 progress
	preTTs := []*models.TournamentTeam{
		tournamentTeam("A", 4, 1, false),
		tournamentTeam("B", 4, 1, false),
		tournamentTeam("C", 4, 1, false),
		tournamentTeam("D", 4, 1, false),
	}

	// Completed: A won championship (6 wins), C lost in final (5 wins),
	// B and D lost in semis (4 wins each)
	completedTTs := []*models.TournamentTeam{
		tournamentTeam("A", 6, 1, false),
		tournamentTeam("B", 4, 1, true),
		tournamentTeam("C", 5, 1, true),
		tournamentTeam("D", 4, 1, true),
	}

	preResult := ComputeFinalFourOutcomes(bracket, portfolios, summaries, details, preTTs, rounds, payouts)
	completedResult := ComputeFinalFourOutcomes(bracket, portfolios, summaries, details, completedTTs, rounds, payouts)

	// THEN both produce identical standings for each outcome
	for i := range preResult {
		preReturns := preResult[i].Standings[0].TotalReturns
		completedReturns := completedResult[i].Standings[0].TotalReturns
		if preReturns != completedReturns {
			t.Errorf("outcome %d: pre-FF returns (%.2f) != completed returns (%.2f) for champion %s",
				i, preReturns, completedReturns, preResult[i].Champion.TeamID)
		}
	}
}

func TestThatStandingsIncludePayouts(t *testing.T) {
	// GIVEN two portfolios with payouts configured for 1st place
	bracket := buildFinalFourBracket(
		bracketTeam("A", "sa", 1, "East"),
		bracketTeam("B", "sb", 2, "West"),
		bracketTeam("C", "sc", 1, "South"),
		bracketTeam("D", "sd", 2, "Midwest"),
	)

	portfolios := []*models.Portfolio{testPortfolio("p1"), testPortfolio("p2")}
	summaries := []*models.OwnershipSummary{ownershipSummary("os1", "p1"), ownershipSummary("os2", "p2")}
	details := []*models.OwnershipDetail{
		ownershipDetail("os1", "A", 1.0),
		ownershipDetail("os2", "B", 1.0),
	}
	tts := []*models.TournamentTeam{
		tournamentTeam("A", 5, 1, false),
		tournamentTeam("B", 5, 1, false),
		tournamentTeam("C", 5, 1, false),
		tournamentTeam("D", 5, 1, false),
	}
	rounds := []*models.ScoringRule{
		scoringRule(1, 10), scoringRule(2, 10), scoringRule(3, 10), scoringRule(4, 10),
		scoringRule(5, 10), scoringRule(6, 10), scoringRule(7, 10),
	}
	payoutsSlice := []*models.PoolPayout{poolPayout(1, 10000)}

	result := ComputeFinalFourOutcomes(bracket, portfolios, summaries, details, tts, rounds, payoutsSlice)

	// THEN first place in each outcome gets the payout
	for _, o := range result {
		if o.Standings[0].PayoutCents != 10000 {
			t.Errorf("expected 1st place payout of 10000, got %d for champion %s", o.Standings[0].PayoutCents, o.Champion.TeamID)
		}
	}
}

func TestThatAllFourTeamsAppearAsChampion(t *testing.T) {
	// GIVEN a complete Final Four bracket
	bracket := buildFinalFourBracket(
		bracketTeam("A", "sa", 1, "East"),
		bracketTeam("B", "sb", 2, "West"),
		bracketTeam("C", "sc", 1, "South"),
		bracketTeam("D", "sd", 2, "Midwest"),
	)

	portfolios := []*models.Portfolio{testPortfolio("p1")}
	summaries := []*models.OwnershipSummary{ownershipSummary("os1", "p1")}
	details := []*models.OwnershipDetail{ownershipDetail("os1", "A", 1.0)}
	tts := []*models.TournamentTeam{
		tournamentTeam("A", 5, 1, false),
		tournamentTeam("B", 5, 1, false),
		tournamentTeam("C", 5, 1, false),
		tournamentTeam("D", 5, 1, false),
	}
	rounds := []*models.ScoringRule{scoringRule(1, 1)}
	payouts := []*models.PoolPayout{}

	result := ComputeFinalFourOutcomes(bracket, portfolios, summaries, details, tts, rounds, payouts)

	// THEN each of the 4 teams appears as champion exactly twice
	championCounts := map[string]int{}
	for _, o := range result {
		championCounts[o.Champion.TeamID]++
	}
	for _, teamID := range []string{"A", "B", "C", "D"} {
		if championCounts[teamID] != 2 {
			t.Errorf("expected team %s as champion 2 times, got %d", teamID, championCounts[teamID])
		}
	}
}

func TestThatRunnerUpIsOpposingSemifinalWinner(t *testing.T) {
	// GIVEN a complete Final Four bracket
	bracket := buildFinalFourBracket(
		bracketTeam("A", "sa", 1, "East"),
		bracketTeam("B", "sb", 2, "West"),
		bracketTeam("C", "sc", 1, "South"),
		bracketTeam("D", "sd", 2, "Midwest"),
	)

	portfolios := []*models.Portfolio{testPortfolio("p1")}
	summaries := []*models.OwnershipSummary{ownershipSummary("os1", "p1")}
	details := []*models.OwnershipDetail{ownershipDetail("os1", "A", 1.0)}
	tts := []*models.TournamentTeam{
		tournamentTeam("A", 5, 1, false),
		tournamentTeam("B", 5, 1, false),
		tournamentTeam("C", 5, 1, false),
		tournamentTeam("D", 5, 1, false),
	}
	rounds := []*models.ScoringRule{scoringRule(1, 1)}
	payouts := []*models.PoolPayout{}

	result := ComputeFinalFourOutcomes(bracket, portfolios, summaries, details, tts, rounds, payouts)

	// THEN in every outcome, the runner-up is the opposing semifinal winner
	for _, o := range result {
		if o.Champion.TeamID == o.Semifinal1Winner.TeamID {
			if o.RunnerUp.TeamID != o.Semifinal2Winner.TeamID {
				t.Errorf("expected runner-up %s but got %s when champion is %s",
					o.Semifinal2Winner.TeamID, o.RunnerUp.TeamID, o.Champion.TeamID)
			}
		} else {
			if o.RunnerUp.TeamID != o.Semifinal1Winner.TeamID {
				t.Errorf("expected runner-up %s but got %s when champion is %s",
					o.Semifinal1Winner.TeamID, o.RunnerUp.TeamID, o.Champion.TeamID)
			}
		}
	}
}
