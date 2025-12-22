package main

import (
	"context"
	"database/sql"

	"github.com/andrewcopp/Calcutta/backend/pkg/services"
)

func runKenPomReturns(ctx context.Context, db *sql.DB, calcuttaID string, sigma float64) ([]KenPomReturnsRow, error) {
	meta, err := tournamentMetaForCalcuttaID(ctx, db, calcuttaID)
	if err != nil {
		return nil, err
	}

	repo := services.NewTournamentRepository(db)
	bracketService := services.NewBracketService(repo)
	returnsService := services.NewKenPomPredictedReturnsService(bracketService).WithSigma(sigma)

	pred, err := returnsService.GetPredictedReturnsPreTournament(ctx, meta.TournamentID)
	if err != nil {
		return nil, err
	}

	rows := make([]KenPomReturnsRow, 0, len(pred))
	for _, p := range pred {
		rows = append(rows, KenPomReturnsRow{
			TournamentName: meta.TournamentName,
			TournamentYear: meta.TournamentYear,
			CalcuttaID:     calcuttaID,
			TeamID:         p.TeamID,
			SchoolName:     p.SchoolName,
			Seed:           p.Seed,
			Region:         p.Region,
			Sigma:          sigma,
			PredPoints:     p.ExpectedPoints,
		})
	}

	return rows, nil
}
