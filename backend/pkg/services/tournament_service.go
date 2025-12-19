package services

import (
	"context"
	"log"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/google/uuid"
)

// TournamentService handles business logic for tournaments
type TournamentService struct {
	repo       *TournamentRepository
	schoolRepo *SchoolRepository
}

// NewTournamentService creates a new TournamentService
func NewTournamentService(repo *TournamentRepository, schoolRepo *SchoolRepository) *TournamentService {
	return &TournamentService{repo: repo, schoolRepo: schoolRepo}
}

// GetAllTournaments returns all tournaments
func (s *TournamentService) GetAllTournaments(ctx context.Context) ([]models.Tournament, error) {
	return s.repo.GetAll(ctx)
}

// GetTournamentWithWinner returns a tournament with its winning team and school
func (s *TournamentService) GetTournamentWithWinner(ctx context.Context, tournamentID string) (*models.Tournament, *models.TournamentTeam, *models.School, error) {
	// Get the tournament
	tournament, err := s.repo.GetByID(ctx, tournamentID)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get the winning team
	team, err := s.repo.GetWinningTeam(ctx, tournamentID)
	if err != nil {
		return tournament, nil, nil, err
	}

	if team == nil {
		return tournament, nil, nil, nil
	}

	// Get the school
	school, err := s.schoolRepo.GetByID(ctx, team.SchoolID)
	if err != nil {
		return tournament, team, nil, err
	}

	if school.ID == "" {
		return tournament, team, nil, nil
	}

	return tournament, team, &school, nil
}

// GetTournamentByID returns a tournament by ID
func (s *TournamentService) GetTournamentByID(ctx context.Context, id string) (*models.Tournament, error) {
	return s.repo.GetByID(ctx, id)
}

// GetWinningTeam returns the winning team for a tournament
func (s *TournamentService) GetWinningTeam(ctx context.Context, tournamentID string) (*models.TournamentTeam, error) {
	return s.repo.GetWinningTeam(ctx, tournamentID)
}

// GetRound determines the round number for a game by traversing the tournament bracket
// This is a more accurate implementation than what can be done in the model layer
func (s *TournamentService) GetRound(ctx context.Context, gameID string, totalRounds int) (int, error) {
	// Get all games for the tournament
	games, err := s.repo.GetGamesByTournamentID(ctx, gameID)
	if err != nil {
		return 0, err
	}

	// Create a map of games for quick lookup
	gamesMap := make(map[string]*models.TournamentGame)
	for _, game := range games {
		gamesMap[game.ID] = game
	}

	// Get the game from the map
	game, exists := gamesMap[gameID]
	if !exists {
		return 0, nil // Game not found
	}

	// Use the model's GetRound method to determine the round
	return game.GetRound(gamesMap, totalRounds), nil
}

// GetGame retrieves a game by ID
// This would be implemented in the actual service layer
func (s *TournamentService) GetGame(gameID string) (*models.TournamentGame, error) {
	// In a real implementation, we would fetch the game from the database
	// For now, we'll return a placeholder
	return &models.TournamentGame{}, nil
}

// CreateTournament creates a new tournament
func (s *TournamentService) CreateTournament(ctx context.Context, name string, rounds int) (*models.Tournament, error) {
	log.Printf("Creating tournament with name: %s, rounds: %d", name, rounds)

	tournament := &models.Tournament{
		ID:     uuid.New().String(),
		Name:   name,
		Rounds: rounds,
	}

	log.Printf("Generated tournament ID: %s", tournament.ID)

	if err := s.repo.Create(ctx, tournament); err != nil {
		log.Printf("Error creating tournament: %v", err)
		return nil, err
	}

	log.Printf("Successfully created tournament: %+v", tournament)
	return tournament, nil
}

func (s *TournamentService) GetTeams(ctx context.Context, tournamentID string) ([]*models.TournamentTeam, error) {
	return s.repo.GetTeams(ctx, tournamentID)
}

func (s *TournamentService) CreateTeam(ctx context.Context, team *models.TournamentTeam) error {
	return s.repo.CreateTeam(ctx, team)
}

func (s *TournamentService) UpdateTournamentTeam(ctx context.Context, team *models.TournamentTeam) error {
	return s.repo.UpdateTournamentTeam(ctx, team)
}
