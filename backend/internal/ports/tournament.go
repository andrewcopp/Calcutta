package ports

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
)

type TournamentReader interface {
	GetAll(ctx context.Context) ([]models.Tournament, error)
	GetByID(ctx context.Context, id string) (*models.Tournament, error)
	GetWinningTeam(ctx context.Context, tournamentID string) (*models.TournamentTeam, error)
	GetGamesByTournamentID(ctx context.Context, gameID string) ([]*models.TournamentGame, error)
}

type TournamentWriter interface {
	Create(ctx context.Context, tournament *models.Tournament) error
	Update(ctx context.Context, tournament *models.Tournament) error
	Delete(ctx context.Context, id string) error
}

type TeamReader interface {
	GetTeams(ctx context.Context, tournamentID string) ([]*models.TournamentTeam, error)
	GetTournamentTeam(ctx context.Context, id string) (*models.TournamentTeam, error)
}

type TeamWriter interface {
	CreateTeam(ctx context.Context, team *models.TournamentTeam) error
	UpdateTournamentTeam(ctx context.Context, team *models.TournamentTeam) error
}
