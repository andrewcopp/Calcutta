package bracket

import appbracket "github.com/andrewcopp/Calcutta/backend/internal/app/bracket"

type Service = appbracket.Service

type TournamentRepo = appbracket.TournamentRepo

type BracketBuilder = appbracket.BracketBuilder

func New(tournamentRepo TournamentRepo) *Service {
	return appbracket.New(tournamentRepo)
}

func NewBracketBuilder() *BracketBuilder {
	return appbracket.NewBracketBuilder()
}
