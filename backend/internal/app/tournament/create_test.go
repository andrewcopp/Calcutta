package tournament

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

// mockRepo implements TournamentRepo with only Create wired up.
// All other methods panic to catch unintended calls.
type mockRepo struct {
	createFn            func(ctx context.Context, t *models.Tournament, competitionName string, year int) error
	lastCreate          *models.Tournament
	lastCompetitionName string
	lastYear            int
}

func (m *mockRepo) Create(ctx context.Context, t *models.Tournament, competitionName string, year int) error {
	m.lastCreate = t
	m.lastCompetitionName = competitionName
	m.lastYear = year
	if m.createFn != nil {
		return m.createFn(ctx, t, competitionName, year)
	}
	return nil
}

func (m *mockRepo) GetAll(context.Context) ([]models.Tournament, error) {
	panic("GetAll not implemented")
}

func (m *mockRepo) GetByID(context.Context, string) (*models.Tournament, error) {
	panic("GetByID not implemented")
}

func (m *mockRepo) UpdateStartingAt(context.Context, string, *time.Time) error {
	panic("UpdateStartingAt not implemented")
}

func (m *mockRepo) GetTeams(context.Context, string) ([]*models.TournamentTeam, error) {
	panic("GetTeams not implemented")
}

func (m *mockRepo) GetTournamentTeam(context.Context, string) (*models.TournamentTeam, error) {
	panic("GetTournamentTeam not implemented")
}

func (m *mockRepo) CreateTeam(context.Context, *models.TournamentTeam) error {
	panic("CreateTeam not implemented")
}

func (m *mockRepo) UpdateTournamentTeam(context.Context, *models.TournamentTeam) error {
	panic("UpdateTournamentTeam not implemented")
}

func (m *mockRepo) GetWinningTeam(context.Context, string) (*models.TournamentTeam, error) {
	panic("GetWinningTeam not implemented")
}

func (m *mockRepo) ListWinningTeams(context.Context) (map[string]*models.TournamentTeam, error) {
	panic("ListWinningTeams not implemented")
}

func (m *mockRepo) GetCompetitions(context.Context) ([]models.Competition, error) {
	panic("GetCompetitions not implemented")
}

func (m *mockRepo) GetSeasons(context.Context) ([]models.Season, error) {
	panic("GetSeasons not implemented")
}

func (m *mockRepo) ReplaceTeams(context.Context, string, []*models.TournamentTeam) error {
	panic("ReplaceTeams not implemented")
}

func (m *mockRepo) UpdateFinalFour(context.Context, string, string, string, string, string) error {
	panic("UpdateFinalFour not implemented")
}

func (m *mockRepo) BulkUpsertKenPomStats(context.Context, []models.TeamKenPomUpdate) error {
	panic("BulkUpsertKenPomStats not implemented")
}

func TestThatCreateReturnsNewTournamentWithCorrectName(t *testing.T) {
	// GIVEN a service with a mock repo
	repo := &mockRepo{}
	svc := New(repo)

	// WHEN creating a tournament with competition "NCAA Tournament" and year 2026
	result, err := svc.Create(context.Background(), "NCAA Tournament", 2026, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the returned tournament has the derived name
	if result.Name != "NCAA Tournament (2026)" {
		t.Errorf("expected name %q, got %q", "NCAA Tournament (2026)", result.Name)
	}
}

func TestThatCreateReturnsNewTournamentWithCorrectRounds(t *testing.T) {
	// GIVEN a service with a mock repo
	repo := &mockRepo{}
	svc := New(repo)

	// WHEN creating a tournament with 6 rounds
	result, err := svc.Create(context.Background(), "NCAA Tournament", 2026, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the returned tournament has 6 rounds
	if result.Rounds != 6 {
		t.Errorf("expected rounds=6, got %d", result.Rounds)
	}
}

func TestThatCreateReturnsNewTournamentWithNonEmptyID(t *testing.T) {
	// GIVEN a service with a mock repo
	repo := &mockRepo{}
	svc := New(repo)

	// WHEN creating a tournament
	result, err := svc.Create(context.Background(), "NCAA Tournament", 2026, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the returned tournament has a non-empty ID
	if result.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestThatCreateReturnsErrorWhenRepoFails(t *testing.T) {
	// GIVEN a repo that returns an error on Create
	repo := &mockRepo{
		createFn: func(_ context.Context, _ *models.Tournament, _ string, _ int) error {
			return errors.New("database connection failed")
		},
	}
	svc := New(repo)

	// WHEN creating a tournament
	_, err := svc.Create(context.Background(), "NCAA Tournament", 2026, 6)

	// THEN the error is propagated
	if err == nil {
		t.Error("expected error when repo fails")
	}
}

func TestThatCreatePassesTournamentToRepo(t *testing.T) {
	// GIVEN a service with a mock repo that captures the tournament
	repo := &mockRepo{}
	svc := New(repo)

	// WHEN creating a tournament with competition "Big East" and year 2025
	_, err := svc.Create(context.Background(), "Big East", 2025, 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// THEN the repo received a tournament with the correct derived name
	if repo.lastCreate == nil {
		t.Fatal("expected repo.Create to be called")
	}
	if repo.lastCreate.Name != "Big East (2025)" {
		t.Errorf("expected repo to receive name %q, got %q", "Big East (2025)", repo.lastCreate.Name)
	}
}
