package services

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

// Tests for eliminated status tracking when selecting/unselecting winners
// These tests verify the database updates for eliminated status

func TestThatSelectingWinnerMarksLosingTeamAsEliminated(t *testing.T) {
	// GIVEN a mock repository with two teams
	ctx := context.Background()
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()

	// Use the fixture teams to avoid accidentally creating duplicate seeds
	var team1 *models.TournamentTeam
	var team16 *models.TournamentTeam
	for _, tm := range teams {
		if tm.Region == "East" && tm.Seed == 1 {
			team1 = tm
		}
		if tm.Region == "East" && tm.Seed == 16 {
			team16 = tm
		}
		// Default non-play-in teams have a bye into the Round of 64
		tm.Byes = 1
		tm.Wins = 0
		tm.Eliminated = false
	}
	if team1 == nil || team16 == nil {
		t.Fatal("failed to find required East seed 1/16 teams in fixture")
	}

	mockRepo := &MockTournamentRepo{
		teams:        make(map[string]*models.TournamentTeam),
		allTeams:     teams,
		tournamentID: "tournament-1",
	}
	mockRepo.teams[team1.ID] = team1
	mockRepo.teams[team16.ID] = team16

	service := &BracketService{
		tournamentRepo: mockRepo,
		builder:        NewBracketBuilder(),
		validator:      models.NewBracketValidator(),
	}

	// WHEN selecting team1 as winner
	_, err := service.SelectWinner(ctx, "tournament-1", "East-round_of_64-1", team1.ID)
	if err != nil {
		t.Fatalf("Failed to select winner: %v", err)
	}

	// THEN team16 is marked as eliminated
	team16Updated := mockRepo.teams[team16.ID]
	if !team16Updated.Eliminated {
		t.Error("Expected losing team (team-16) to be marked as eliminated")
	}
}

func TestThatSelectingWinnerDoesNotMarkWinnerAsEliminated(t *testing.T) {
	// GIVEN a Round of 64 game with team1 and team16
	ctx := context.Background()
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()

	var team1 *models.TournamentTeam
	var team16 *models.TournamentTeam
	for _, tm := range teams {
		if tm.Region == "East" && tm.Seed == 1 {
			team1 = tm
		}
		if tm.Region == "East" && tm.Seed == 16 {
			team16 = tm
		}
		tm.Byes = 1
		tm.Wins = 0
		tm.Eliminated = false
	}
	if team1 == nil || team16 == nil {
		t.Fatal("failed to find required East seed 1/16 teams in fixture")
	}

	mockRepo := &MockTournamentRepo{
		teams:    make(map[string]*models.TournamentTeam),
		allTeams: teams,
	}
	mockRepo.teams[team1.ID] = team1
	mockRepo.teams[team16.ID] = team16

	service := &BracketService{
		tournamentRepo: mockRepo,
		builder:        NewBracketBuilder(),
		validator:      models.NewBracketValidator(),
	}

	// WHEN selecting team1 as winner
	_, err := service.SelectWinner(ctx, "tournament-1", "East-round_of_64-1", team1.ID)
	if err != nil {
		t.Fatalf("Failed to select winner: %v", err)
	}

	// THEN team1 is NOT marked as eliminated
	team1Updated := mockRepo.teams[team1.ID]
	if team1Updated.Eliminated {
		t.Error("Expected winning team (team-1) to NOT be marked as eliminated")
	}
}

func TestThatUnselectingWinnerReactivatesLosingTeam(t *testing.T) {
	// GIVEN a Round of 64 game where team1 won and team16 is eliminated
	ctx := context.Background()
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()

	var team1 *models.TournamentTeam
	var team16 *models.TournamentTeam
	for _, tm := range teams {
		if tm.Region == "East" && tm.Seed == 1 {
			team1 = tm
		}
		if tm.Region == "East" && tm.Seed == 16 {
			team16 = tm
		}
		tm.Byes = 1
		tm.Wins = 0
		tm.Eliminated = false
	}
	if team1 == nil || team16 == nil {
		t.Fatal("failed to find required East seed 1/16 teams in fixture")
	}
	team1.Wins = 1
	team16.Eliminated = true

	mockRepo := &MockTournamentRepo{
		teams:    make(map[string]*models.TournamentTeam),
		allTeams: teams,
	}
	mockRepo.teams[team1.ID] = team1
	mockRepo.teams[team16.ID] = team16

	service := &BracketService{
		tournamentRepo: mockRepo,
		builder:        NewBracketBuilder(),
		validator:      models.NewBracketValidator(),
	}

	// WHEN unselecting the winner
	_, err := service.UnselectWinner(ctx, "tournament-1", "East-round_of_64-1")
	if err != nil {
		t.Fatalf("Failed to unselect winner: %v", err)
	}

	// THEN team16 is reactivated (no longer eliminated)
	team16Updated := mockRepo.teams[team16.ID]
	if team16Updated.Eliminated {
		t.Error("Expected losing team (team-16) to be reactivated after unselecting winner")
	}
}

func TestThatFirstFourLoserIsMarkedAsEliminated(t *testing.T) {
	// GIVEN a First Four game with two 11-seeds
	ctx := context.Background()
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()

	var team11a *models.TournamentTeam
	var team11b *models.TournamentTeam
	for _, tm := range teams {
		tm.Byes = 1
		tm.Wins = 0
		tm.Eliminated = false
	}
	for _, tm := range teams {
		if tm.Region == "East" && tm.Seed == 11 {
			if team11a == nil {
				team11a = tm
			} else if team11b == nil {
				team11b = tm
			}
		}
	}
	if team11a == nil || team11b == nil {
		t.Fatal("failed to find required East seed 11 play-in teams in fixture")
	}
	team11a.Byes = 0
	team11b.Byes = 0

	mockRepo := &MockTournamentRepo{
		teams:    make(map[string]*models.TournamentTeam),
		allTeams: teams,
	}
	mockRepo.teams[team11a.ID] = team11a
	mockRepo.teams[team11b.ID] = team11b

	service := &BracketService{
		tournamentRepo: mockRepo,
		builder:        NewBracketBuilder(),
		validator:      models.NewBracketValidator(),
	}

	// WHEN selecting team-11a as winner
	_, err := service.SelectWinner(ctx, "tournament-1", "East-first_four-11", team11a.ID)
	if err != nil {
		t.Fatalf("Failed to select winner: %v", err)
	}

	// THEN team-11b is marked as eliminated
	team11bUpdated := mockRepo.teams[team11b.ID]
	if !team11bUpdated.Eliminated {
		t.Error("Expected First Four loser (team-11b) to be marked as eliminated")
	}
}

func TestThatUnselectingFirstFourWinnerReactivatesLoser(t *testing.T) {
	// GIVEN a First Four game where team-11a won and team-11b is eliminated
	ctx := context.Background()
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()

	var team11a *models.TournamentTeam
	var team11b *models.TournamentTeam
	for _, tm := range teams {
		tm.Byes = 1
		tm.Wins = 0
		tm.Eliminated = false
	}
	for _, tm := range teams {
		if tm.Region == "East" && tm.Seed == 11 {
			if team11a == nil {
				team11a = tm
			} else if team11b == nil {
				team11b = tm
			}
		}
	}
	if team11a == nil || team11b == nil {
		t.Fatal("failed to find required East seed 11 play-in teams in fixture")
	}
	team11a.Byes = 0
	team11b.Byes = 0
	team11a.Wins = 1
	team11b.Eliminated = true

	mockRepo := &MockTournamentRepo{
		teams:    make(map[string]*models.TournamentTeam),
		allTeams: teams,
	}
	mockRepo.teams[team11a.ID] = team11a
	mockRepo.teams[team11b.ID] = team11b

	service := &BracketService{
		tournamentRepo: mockRepo,
		builder:        NewBracketBuilder(),
		validator:      models.NewBracketValidator(),
	}

	// WHEN unselecting the winner
	_, err := service.UnselectWinner(ctx, "tournament-1", "East-first_four-11")
	if err != nil {
		t.Fatalf("Failed to unselect winner: %v", err)
	}

	// THEN team-11b is reactivated
	team11bUpdated := mockRepo.teams[team11b.ID]
	if team11bUpdated.Eliminated {
		t.Error("Expected First Four loser (team-11b) to be reactivated after unselecting winner")
	}
}

func TestThatChampionshipLoserIsMarkedAsEliminated(t *testing.T) {
	// GIVEN a Championship game
	ctx := context.Background()
	helper := NewBracketTestHelper()
	teams := helper.CreateTournament68Teams()

	// Make the bracket fully resolved so the championship game has both teams.
	for _, tm := range teams {
		tm.Byes = 1
		tm.Wins = 6
		tm.Eliminated = false
	}

	var champion *models.TournamentTeam
	var runnerUp *models.TournamentTeam
	for _, tm := range teams {
		if tm.Region == "East" && tm.Seed == 1 {
			champion = tm
		}
		if tm.Region == "South" && tm.Seed == 1 {
			runnerUp = tm
		}
	}
	if champion == nil || runnerUp == nil {
		t.Fatal("failed to find required seed 1 teams in fixture")
	}

	mockRepo := &MockTournamentRepo{
		teams:    make(map[string]*models.TournamentTeam),
		allTeams: teams,
	}
	mockRepo.teams[champion.ID] = champion
	mockRepo.teams[runnerUp.ID] = runnerUp

	service := &BracketService{
		tournamentRepo: mockRepo,
		builder:        NewBracketBuilder(),
		validator:      models.NewBracketValidator(),
	}

	// WHEN selecting team-champion as winner
	_, err := service.SelectWinner(ctx, "tournament-1", "championship", champion.ID)
	if err != nil {
		t.Fatalf("Failed to select winner: %v", err)
	}

	// THEN team-runner-up is marked as eliminated
	runnerUpUpdated := mockRepo.teams[runnerUp.ID]
	if !runnerUpUpdated.Eliminated {
		t.Error("Expected championship loser (team-runner-up) to be marked as eliminated")
	}
}

// MockTournamentRepo for testing eliminated status
type MockTournamentRepo struct {
	teams        map[string]*models.TournamentTeam
	allTeams     []*models.TournamentTeam
	tournamentID string
}

func (m *MockTournamentRepo) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	team, exists := m.teams[id]
	if !exists {
		return nil, nil
	}
	// Return a copy to simulate database behavior
	teamCopy := *team
	return &teamCopy, nil
}

func (m *MockTournamentRepo) UpdateTournamentTeam(ctx context.Context, team *models.TournamentTeam) error {
	// Update the mock storage
	m.teams[team.ID] = team
	return nil
}

func (m *MockTournamentRepo) GetByID(ctx context.Context, id string) (*models.Tournament, error) {
	return &models.Tournament{
		ID:                   id,
		FinalFourTopLeft:     "East",
		FinalFourBottomLeft:  "West",
		FinalFourTopRight:    "South",
		FinalFourBottomRight: "Midwest",
	}, nil
}

func (m *MockTournamentRepo) GetTeams(ctx context.Context, tournamentID string) ([]*models.TournamentTeam, error) {
	if m.allTeams != nil {
		return m.allTeams, nil
	}
	teams := make([]*models.TournamentTeam, 0, len(m.teams))
	for _, team := range m.teams {
		teamCopy := *team
		teams = append(teams, &teamCopy)
	}
	return teams, nil
}
