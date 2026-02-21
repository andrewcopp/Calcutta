package ports

import (
	"context"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
)

type TournamentResolver interface {
	ResolveCoreTournamentID(ctx context.Context, season int) (string, error)
	ResolveSeasonFromTournamentID(ctx context.Context, tournamentID string) (int, error)
	LoadFinalFourConfig(ctx context.Context, coreTournamentID string) (*models.FinalFourConfig, error)
}
