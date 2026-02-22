package tournament

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type kenPomMockRepo struct {
	mockRepo
	teams              []*models.TournamentTeam
	bulkUpsertCalled   bool
	bulkUpsertUpdates  []models.TeamKenPomUpdate
}

func (m *kenPomMockRepo) GetTeams(_ context.Context, _ string) ([]*models.TournamentTeam, error) {
	return m.teams, nil
}

func (m *kenPomMockRepo) BulkUpsertKenPomStats(_ context.Context, updates []models.TeamKenPomUpdate) error {
	m.bulkUpsertCalled = true
	m.bulkUpsertUpdates = updates
	return nil
}

func TestThatUpdateKenPomStatsRejectsTeamNotInTournament(t *testing.T) {
	// GIVEN a tournament with one team
	repo := &kenPomMockRepo{
		teams: []*models.TournamentTeam{
			{ID: "team-1", TournamentID: "tourney-1"},
		},
	}
	svc := New(repo)

	// WHEN updating KenPom stats with a team ID not in the tournament
	err := svc.UpdateKenPomStats(context.Background(), "tourney-1", []KenPomUpdateInput{
		{TeamID: "team-unknown", NetRtg: 10.0, ORtg: 110.0, DRtg: 100.0, AdjT: 68.0},
	})

	// THEN an error is returned
	if err == nil {
		t.Error("expected error for team not in tournament")
	}
}

func TestThatUpdateKenPomStatsPassesValidUpdatesToRepo(t *testing.T) {
	// GIVEN a tournament with two teams
	repo := &kenPomMockRepo{
		teams: []*models.TournamentTeam{
			{ID: "team-1", TournamentID: "tourney-1"},
			{ID: "team-2", TournamentID: "tourney-1"},
		},
	}
	svc := New(repo)

	// WHEN updating KenPom stats for both teams
	err := svc.UpdateKenPomStats(context.Background(), "tourney-1", []KenPomUpdateInput{
		{TeamID: "team-1", NetRtg: 10.0, ORtg: 110.0, DRtg: 100.0, AdjT: 68.0},
		{TeamID: "team-2", NetRtg: 5.0, ORtg: 105.0, DRtg: 100.0, AdjT: 70.0},
	})

	// THEN no error is returned and repo was called
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !repo.bulkUpsertCalled {
		t.Error("expected BulkUpsertKenPomStats to be called")
	}
}

func TestThatUpdateKenPomStatsPassesCorrectUpdateCount(t *testing.T) {
	// GIVEN a tournament with two teams
	repo := &kenPomMockRepo{
		teams: []*models.TournamentTeam{
			{ID: "team-1", TournamentID: "tourney-1"},
			{ID: "team-2", TournamentID: "tourney-1"},
		},
	}
	svc := New(repo)

	// WHEN updating KenPom stats for both teams
	_ = svc.UpdateKenPomStats(context.Background(), "tourney-1", []KenPomUpdateInput{
		{TeamID: "team-1", NetRtg: 10.0, ORtg: 110.0, DRtg: 100.0, AdjT: 68.0},
		{TeamID: "team-2", NetRtg: 5.0, ORtg: 105.0, DRtg: 100.0, AdjT: 70.0},
	})

	// THEN the repo received exactly 2 updates
	if len(repo.bulkUpsertUpdates) != 2 {
		t.Errorf("expected 2 updates, got %d", len(repo.bulkUpsertUpdates))
	}
}
