package bracket

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type stubTournamentRepo struct {
	teams []*models.TournamentTeam
}

func (r *stubTournamentRepo) GetByID(ctx context.Context, id string) (*models.Tournament, error) {
	return nil, nil
}

func (r *stubTournamentRepo) GetTeams(ctx context.Context, tournamentID string) ([]*models.TournamentTeam, error) {
	return r.teams, nil
}

func (r *stubTournamentRepo) GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error) {
	return nil, nil
}

func (r *stubTournamentRepo) UpdateTournamentTeam(ctx context.Context, team *models.TournamentTeam) error {
	return nil
}

func TestThatValidateBracketSetupReturnsErrorWhenPlayInTeamHasBye(t *testing.T) {
	// GIVEN a tournament with a duplicated (region, seed) but a play-in team has byes=1
	teams := createFullTournamentTeams("t")
	for _, team := range teams {
		if team != nil && team.Region == "East" && team.Seed == 11 {
			team.Byes = 1
			break
		}
	}

	svc := New(&stubTournamentRepo{teams: teams})

	// WHEN validating bracket setup
	err := svc.ValidateBracketSetup(context.Background(), "t")

	// THEN validation fails
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestThatValidateBracketSetupReturnsErrorWhenNonPlayInTeamHasNoBye(t *testing.T) {
	// GIVEN a tournament with a non-play-in team but byes=0
	teams := createFullTournamentTeams("t")
	for _, team := range teams {
		if team != nil && team.Region == "West" && team.Seed == 1 {
			team.Byes = 0
			break
		}
	}

	svc := New(&stubTournamentRepo{teams: teams})

	// WHEN validating bracket setup
	err := svc.ValidateBracketSetup(context.Background(), "t")

	// THEN validation fails
	if err == nil {
		t.Errorf("expected error")
	}
}
