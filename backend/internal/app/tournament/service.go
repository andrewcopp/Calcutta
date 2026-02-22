package tournament

import (
	"context"
	"fmt"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/google/uuid"
)

// TournamentRepo defines the repository methods used by the tournament service.
type TournamentRepo interface {
	GetAll(ctx context.Context) ([]models.Tournament, error)
	GetByID(ctx context.Context, id string) (*models.Tournament, error)
	Create(ctx context.Context, tournament *models.Tournament, competitionName string, year int) error
	UpdateStartingAt(ctx context.Context, tournamentID string, startingAt *time.Time) error
	UpdateFinalFour(ctx context.Context, tournamentID, topLeft, bottomLeft, topRight, bottomRight string) error
	GetTeams(ctx context.Context, tournamentID string) ([]*models.TournamentTeam, error)
	GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error)
	CreateTeam(ctx context.Context, team *models.TournamentTeam) error
	UpdateTournamentTeam(ctx context.Context, team *models.TournamentTeam) error
	GetWinningTeam(ctx context.Context, tournamentID string) (*models.TournamentTeam, error)
	ListWinningTeams(ctx context.Context) (map[string]*models.TournamentTeam, error)
	GetCompetitions(ctx context.Context) ([]models.Competition, error)
	GetSeasons(ctx context.Context) ([]models.Season, error)
	ReplaceTeams(ctx context.Context, tournamentID string, teams []*models.TournamentTeam) error
	BulkUpsertKenPomStats(ctx context.Context, updates []models.TeamKenPomUpdate) error
}

type Service struct {
	repo TournamentRepo
}

func New(repo TournamentRepo) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context) ([]models.Tournament, error) {
	return s.repo.GetAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id string) (*models.Tournament, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Create(ctx context.Context, competitionName string, year int, rounds int) (*models.Tournament, error) {
	name := fmt.Sprintf("%s (%d)", competitionName, year)
	t := &models.Tournament{ID: uuid.New().String(), Name: name, Rounds: rounds}
	if err := s.repo.Create(ctx, t, competitionName, year); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) GetTeams(ctx context.Context, tournamentID string) ([]*models.TournamentTeam, error) {
	return s.repo.GetTeams(ctx, tournamentID)
}

func (s *Service) CreateTeam(ctx context.Context, team *models.TournamentTeam) error {
	return s.repo.CreateTeam(ctx, team)
}

func (s *Service) UpdateTournamentTeam(ctx context.Context, team *models.TournamentTeam) error {
	return s.repo.UpdateTournamentTeam(ctx, team)
}

func (s *Service) GetWinningTeam(ctx context.Context, tournamentID string) (*models.TournamentTeam, error) {
	return s.repo.GetWinningTeam(ctx, tournamentID)
}

func (s *Service) ListWinningTeams(ctx context.Context) (map[string]*models.TournamentTeam, error) {
	return s.repo.ListWinningTeams(ctx)
}

func (s *Service) UpdateStartingAt(ctx context.Context, tournamentID string, startingAt *time.Time) error {
	return s.repo.UpdateStartingAt(ctx, tournamentID, startingAt)
}

func (s *Service) UpdateFinalFour(ctx context.Context, tournamentID, topLeft, bottomLeft, topRight, bottomRight string) error {
	return s.repo.UpdateFinalFour(ctx, tournamentID, topLeft, bottomLeft, topRight, bottomRight)
}

func (s *Service) ListCompetitions(ctx context.Context) ([]models.Competition, error) {
	return s.repo.GetCompetitions(ctx)
}

func (s *Service) ListSeasons(ctx context.Context) ([]models.Season, error) {
	return s.repo.GetSeasons(ctx)
}

// ReplaceTeamsInput represents a single team entry for the bulk replace operation.
type ReplaceTeamsInput struct {
	SchoolID string
	Seed     int
	Region   string
}

// ReplaceTeams validates and replaces all teams in a tournament.
func (s *Service) ReplaceTeams(ctx context.Context, tournamentID string, inputs []ReplaceTeamsInput) ([]*models.TournamentTeam, error) {
	teams := buildTeamsFromInputs(tournamentID, inputs)

	if errs := ValidateBracketSetup(teams); len(errs) > 0 {
		return nil, &BracketValidationError{Errors: errs}
	}

	if err := s.repo.ReplaceTeams(ctx, tournamentID, teams); err != nil {
		return nil, fmt.Errorf("failed to replace teams: %w", err)
	}

	return s.repo.GetTeams(ctx, tournamentID)
}

// KenPomUpdateInput represents a single team's KenPom rating update.
type KenPomUpdateInput struct {
	TeamID string
	NetRtg float64
	ORtg   float64
	DRtg   float64
	AdjT   float64
}

// UpdateKenPomStats validates that all team IDs belong to the tournament, then upserts KenPom ratings.
func (s *Service) UpdateKenPomStats(ctx context.Context, tournamentID string, inputs []KenPomUpdateInput) error {
	teams, err := s.repo.GetTeams(ctx, tournamentID)
	if err != nil {
		return fmt.Errorf("failed to fetch teams: %w", err)
	}

	validIDs := make(map[string]bool, len(teams))
	for _, t := range teams {
		validIDs[t.ID] = true
	}

	for _, input := range inputs {
		if !validIDs[input.TeamID] {
			return fmt.Errorf("team %s does not belong to tournament %s", input.TeamID, tournamentID)
		}
	}

	updates := make([]models.TeamKenPomUpdate, 0, len(inputs))
	for _, input := range inputs {
		updates = append(updates, models.TeamKenPomUpdate{
			TeamID: input.TeamID,
			NetRtg: input.NetRtg,
			ORtg:   input.ORtg,
			DRtg:   input.DRtg,
			AdjT:   input.AdjT,
		})
	}

	return s.repo.BulkUpsertKenPomStats(ctx, updates)
}

// buildTeamsFromInputs converts ReplaceTeamsInput entries into TournamentTeam models.
// It auto-computes byes: if two teams share the same region+seed, both get byes=0 (play-in);
// otherwise the team gets byes=1 (first-round bye).
func buildTeamsFromInputs(tournamentID string, inputs []ReplaceTeamsInput) []*models.TournamentTeam {
	type regionSeed struct {
		region string
		seed   int
	}
	counts := make(map[regionSeed]int)
	for _, input := range inputs {
		counts[regionSeed{region: input.Region, seed: input.Seed}]++
	}

	teams := make([]*models.TournamentTeam, 0, len(inputs))
	for _, input := range inputs {
		key := regionSeed{region: input.Region, seed: input.Seed}
		byes := 1
		if counts[key] == 2 {
			byes = 0
		}
		teams = append(teams, &models.TournamentTeam{
			ID:           uuid.New().String(),
			TournamentID: tournamentID,
			SchoolID:     input.SchoolID,
			Seed:         input.Seed,
			Region:       input.Region,
			Byes:         byes,
			Wins:         0,
			IsEliminated: false,
		})
	}
	return teams
}
