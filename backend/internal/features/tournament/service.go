package tournament

import (
	dbadapters "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	apptournament "github.com/andrewcopp/Calcutta/backend/internal/app/tournament"
)

type Service = apptournament.Service

func New(repo *dbadapters.TournamentRepository) *Service {
	return apptournament.New(repo)
}
