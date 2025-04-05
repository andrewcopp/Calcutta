package services

import (
	"calcutta/internal/models"
	"testing"
)

func TestValidateEntry(t *testing.T) {
	service := NewCalcuttaService()
	entry := &models.CalcuttaEntry{
		ID:         "entry1",
		UserID:     "user1",
		CalcuttaID: "calcutta1",
	}

	// Test case 1: Valid entry (three bids)
	bids := []*models.CalcuttaEntryBid{
		{
			ID:      "bid1",
			EntryID: "entry1",
			TeamID:  "team1",
			Amount:  20,
		},
		{
			ID:      "bid2",
			EntryID: "entry1",
			TeamID:  "team2",
			Amount:  30,
		},
		{
			ID:      "bid3",
			EntryID: "entry1",
			TeamID:  "team3",
			Amount:  40,
		},
	}
	err := service.ValidateEntry(entry, bids)
	if err != nil {
		t.Errorf("Expected no error for valid entry, got: %v", err)
	}

	// Test case 2: Invalid entry (less than 3 teams)
	bids = []*models.CalcuttaEntryBid{
		{
			ID:      "bid1",
			EntryID: "entry1",
			TeamID:  "team1",
			Amount:  20,
		},
		{
			ID:      "bid2",
			EntryID: "entry1",
			TeamID:  "team2",
			Amount:  30,
		},
	}
	err = service.ValidateEntry(entry, bids)
	if err == nil {
		t.Error("Expected error for less than 3 teams, got nil")
	}

	// Test case 3: Invalid entry (more than 10 teams)
	bids = make([]*models.CalcuttaEntryBid, 11)
	for i := 0; i < 11; i++ {
		bids[i] = &models.CalcuttaEntryBid{
			ID:      "bid" + string(rune('1'+i)),
			EntryID: "entry1",
			TeamID:  "team" + string(rune('1'+i)),
			Amount:  10,
		}
	}
	err = service.ValidateEntry(entry, bids)
	if err == nil {
		t.Error("Expected error for more than 10 teams, got nil")
	}

	// Test case 4: Invalid entry (more than $50 on a single team)
	bids = []*models.CalcuttaEntryBid{
		{
			ID:      "bid1",
			EntryID: "entry1",
			TeamID:  "team1",
			Amount:  51,
		},
		{
			ID:      "bid2",
			EntryID: "entry1",
			TeamID:  "team2",
			Amount:  20,
		},
		{
			ID:      "bid3",
			EntryID: "entry1",
			TeamID:  "team3",
			Amount:  20,
		},
	}
	err = service.ValidateEntry(entry, bids)
	if err == nil {
		t.Error("Expected error for more than $50 on a single team, got nil")
	}

	// Test case 5: Invalid entry (total bids exceed $100)
	bids = []*models.CalcuttaEntryBid{
		{
			ID:      "bid1",
			EntryID: "entry1",
			TeamID:  "team1",
			Amount:  50,
		},
		{
			ID:      "bid2",
			EntryID: "entry1",
			TeamID:  "team2",
			Amount:  51,
		},
		{
			ID:      "bid3",
			EntryID: "entry1",
			TeamID:  "team3",
			Amount:  1,
		},
	}
	err = service.ValidateEntry(entry, bids)
	if err == nil {
		t.Error("Expected error for total bids exceeding $100, got nil")
	}

	// Test case 6: Invalid entry (less than $1 on a team)
	bids = []*models.CalcuttaEntryBid{
		{
			ID:      "bid1",
			EntryID: "entry1",
			TeamID:  "team1",
			Amount:  20,
		},
		{
			ID:      "bid2",
			EntryID: "entry1",
			TeamID:  "team2",
			Amount:  30,
		},
		{
			ID:      "bid3",
			EntryID: "entry1",
			TeamID:  "team3",
			Amount:  0,
		},
	}
	err = service.ValidateEntry(entry, bids)
	if err == nil {
		t.Error("Expected error for less than $1 on a team, got nil")
	}

	// Test case 7: Invalid entry (bidding on the same team multiple times)
	bids = []*models.CalcuttaEntryBid{
		{
			ID:      "bid1",
			EntryID: "entry1",
			TeamID:  "team1",
			Amount:  20,
		},
		{
			ID:      "bid2",
			EntryID: "entry1",
			TeamID:  "team2",
			Amount:  30,
		},
		{
			ID:      "bid3",
			EntryID: "entry1",
			TeamID:  "team1", // Same team as bid1
			Amount:  10,
		},
	}
	err = service.ValidateEntry(entry, bids)
	if err == nil {
		t.Error("Expected error for bidding on the same team multiple times, got nil")
	}
}

func TestCalculateOwnershipPercentage(t *testing.T) {
	service := NewCalcuttaService()

	// Example 1 from rules.md
	bid := &models.CalcuttaEntryBid{
		ID:      "bid1",
		EntryID: "entry1",
		TeamID:  "teamA",
		Amount:  20,
	}
	allBids := []*models.CalcuttaEntryBid{
		{
			ID:      "bid1",
			EntryID: "entry1",
			TeamID:  "teamA",
			Amount:  20,
		},
		{
			ID:      "bid2",
			EntryID: "entry2",
			TeamID:  "teamA",
			Amount:  30,
		},
		{
			ID:      "bid3",
			EntryID: "entry3",
			TeamID:  "teamA",
			Amount:  50,
		},
	}
	percentage := service.CalculateOwnershipPercentage(bid, allBids)
	expectedPercentage := 0.2 // 20/100
	if percentage != expectedPercentage {
		t.Errorf("Expected ownership percentage of %v, got %v", expectedPercentage, percentage)
	}

	// Example 2 from rules.md
	bid = &models.CalcuttaEntryBid{
		ID:      "bid1",
		EntryID: "entry1",
		TeamID:  "teamA",
		Amount:  40,
	}
	allBids = []*models.CalcuttaEntryBid{
		{
			ID:      "bid1",
			EntryID: "entry1",
			TeamID:  "teamA",
			Amount:  40,
		},
		{
			ID:      "bid2",
			EntryID: "entry2",
			TeamID:  "teamA",
			Amount:  60,
		},
	}
	percentage = service.CalculateOwnershipPercentage(bid, allBids)
	expectedPercentage = 0.4 // 40/100
	if percentage != expectedPercentage {
		t.Errorf("Expected ownership percentage of %v, got %v", expectedPercentage, percentage)
	}
}

func TestCalculatePoints(t *testing.T) {
	service := NewCalcuttaService()
	rounds := []*models.CalcuttaRound{
		{
			ID:         "round1",
			CalcuttaID: "calcutta1",
			Round:      1,
			Points:     0, // First Four
		},
		{
			ID:         "round2",
			CalcuttaID: "calcutta1",
			Round:      2,
			Points:     50, // First Round
		},
		{
			ID:         "round3",
			CalcuttaID: "calcutta1",
			Round:      3,
			Points:     100, // Sweet 16
		},
		{
			ID:         "round4",
			CalcuttaID: "calcutta1",
			Round:      4,
			Points:     150, // Elite 8
		},
		{
			ID:         "round5",
			CalcuttaID: "calcutta1",
			Round:      5,
			Points:     200, // Final Four
		},
		{
			ID:         "round6",
			CalcuttaID: "calcutta1",
			Round:      6,
			Points:     250, // Championship Game
		},
		{
			ID:         "round7",
			CalcuttaID: "calcutta1",
			Round:      7,
			Points:     300, // Tournament Winner
		},
	}

	// Test case 1: Team eliminated in First Four (0 wins, 0 byes)
	team := &models.TournamentTeam{
		ID:           "team1",
		SchoolID:     "school1",
		TournamentID: "tournament1",
		Byes:         0,
		Wins:         0,
	}
	points := service.CalculatePoints(team, rounds)
	expectedPoints := 0
	if points != expectedPoints {
		t.Errorf("Expected %v points for team eliminated in First Four, got %v", expectedPoints, points)
	}

	// Test case 2: Team wins First Four but loses in First Round (1 win)
	team = &models.TournamentTeam{
		ID:           "team2",
		SchoolID:     "school2",
		TournamentID: "tournament1",
		Byes:         0,
		Wins:         1,
	}
	points = service.CalculatePoints(team, rounds)
	expectedPoints = 0 // First Four Win (no points)
	if points != expectedPoints {
		t.Errorf("Expected %v points for team winning First Four but losing in First Round, got %v", expectedPoints, points)
	}

	// Test case 3: Team with bye loses in First Round (1 bye)
	team = &models.TournamentTeam{
		ID:           "team3",
		SchoolID:     "school3",
		TournamentID: "tournament1",
		Byes:         1,
		Wins:         0,
	}
	points = service.CalculatePoints(team, rounds)
	expectedPoints = 0 // First Round Loss with bye (no points)
	if points != expectedPoints {
		t.Errorf("Expected %v points for team with bye losing in First Round, got %v", expectedPoints, points)
	}

	// Test case 4: Team makes the Round of 32 (2 wins or 1 win + 1 bye)
	team = &models.TournamentTeam{
		ID:           "team4",
		SchoolID:     "school4",
		TournamentID: "tournament1",
		Byes:         0,
		Wins:         2,
	}
	points = service.CalculatePoints(team, rounds)
	expectedPoints = 50 // First Round Winner
	if points != expectedPoints {
		t.Errorf("Expected %v points for team making Sweet 16, got %v", expectedPoints, points)
	}

	// Test case 5: Team makes Sweet 16 (3 wins or 2 wins + 1 bye)
	team = &models.TournamentTeam{
		ID:           "team5",
		SchoolID:     "school5",
		TournamentID: "tournament1",
		Byes:         0,
		Wins:         3,
	}
	points = service.CalculatePoints(team, rounds)
	expectedPoints = 150 // Round of 32 Winner
	if points != expectedPoints {
		t.Errorf("Expected %v points for team making Elite 8, got %v", expectedPoints, points)
	}

	// Test case 6: Team makes Elite 8 (4 wins or 3 wins + 1 bye)
	team = &models.TournamentTeam{
		ID:           "team6",
		SchoolID:     "school6",
		TournamentID: "tournament1",
		Byes:         0,
		Wins:         4,
	}
	points = service.CalculatePoints(team, rounds)
	expectedPoints = 300 // Sweet 16 Winner
	if points != expectedPoints {
		t.Errorf("Expected %v points for team making Final Four, got %v", expectedPoints, points)
	}

	// Test case 7: Team makes Final Foure (5 wins or 4 wins + 1 bye)
	team = &models.TournamentTeam{
		ID:           "team7",
		SchoolID:     "school7",
		TournamentID: "tournament1",
		Byes:         0,
		Wins:         5,
	}
	points = service.CalculatePoints(team, rounds)
	expectedPoints = 500 // Elite 8 Winner
	if points != expectedPoints {
		t.Errorf("Expected %v points for team making Championship Game, got %v", expectedPoints, points)
	}

	// Test case 8: Team wins tournament (7 wins for First Four team or 6 wins for team with bye)
	team = &models.TournamentTeam{
		ID:           "team8",
		SchoolID:     "school8",
		TournamentID: "tournament1",
		Byes:         0,
		Wins:         7, // First Four team needs 7 wins to win the tournament
	}
	points = service.CalculatePoints(team, rounds)
	expectedPoints = 1050
	if points != expectedPoints {
		t.Errorf("Expected %v points for team winning tournament, got %v", expectedPoints, points)
	}

	// Test case 9: Team with bye makes Sweet 16 (1 bye, 2 wins)
	team = &models.TournamentTeam{
		ID:           "team9",
		SchoolID:     "school9",
		TournamentID: "tournament1",
		Byes:         1,
		Wins:         2,
	}
	points = service.CalculatePoints(team, rounds)
	expectedPoints = 150 // Round of 32 Winner (1 bye + 2 wins = 3 total)
	if points != expectedPoints {
		t.Errorf("Expected %v points for team with bye making Elite 8, got %v", expectedPoints, points)
	}

	// Test case 10: Team with bye wins tournament (1 bye, 6 wins)
	team = &models.TournamentTeam{
		ID:           "team10",
		SchoolID:     "school10",
		TournamentID: "tournament1",
		Byes:         1,
		Wins:         6, // Team with bye needs 6 wins to win the tournament
	}
	points = service.CalculatePoints(team, rounds)
	expectedPoints = 1050 // Tournament Winner
	if points != expectedPoints {
		t.Errorf("Expected %v points for team with bye winning tournament, got %v", expectedPoints, points)
	}
}

func TestCalculatePlayerPoints(t *testing.T) {
	service := NewCalcuttaService()
	portfolio := &models.CalcuttaPortfolio{
		ID:      "portfolio1",
		EntryID: "entry1",
	}
	portfolioTeams := []*models.CalcuttaPortfolioTeam{
		{
			ID:                  "portfolioTeam1",
			PortfolioID:         "portfolio1",
			TeamID:              "team1",
			OwnershipPercentage: 0.25,
			PointsEarned:        125,
		},
		{
			ID:                  "portfolioTeam2",
			PortfolioID:         "portfolio1",
			TeamID:              "team2",
			OwnershipPercentage: 0.15,
			PointsEarned:        0,
		},
		{
			ID:                  "portfolioTeam3",
			PortfolioID:         "portfolio1",
			TeamID:              "team3",
			OwnershipPercentage: 0.1,
			PointsEarned:        15,
		},
		{
			ID:                  "portfolioTeam4",
			PortfolioID:         "portfolio2", // Different portfolio
			TeamID:              "team4",
			OwnershipPercentage: 0.5,
			PointsEarned:        500,
		},
	}

	// Example 2 from rules.md
	points := service.CalculatePlayerPoints(portfolio, portfolioTeams)
	expectedPoints := 140 // 125 + 0 + 15
	if points != expectedPoints {
		t.Errorf("Expected %v points for player, got %v", expectedPoints, points)
	}
}
