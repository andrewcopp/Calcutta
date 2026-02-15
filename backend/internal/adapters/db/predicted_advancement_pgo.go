package db

import (
	"context"
	"errors"
	"fmt"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/jackc/pgx/v5/pgxpool"
)

func computeTournamentPredictedAdvancementFromPGO(ctx context.Context, pool *pgxpool.Pool, bracketBuilder ports.BracketBuilder, tournamentID string, gameOutcomeRunID *string) (*string, []ports.TournamentPredictedAdvancementData, error) {
	if tournamentID == "" {
		return nil, nil, errors.New("tournamentID is required")
	}

	kenpomScale := 10.0

	finalFour, err := loadFinalFourForTournament(ctx, pool, tournamentID)
	if err != nil {
		return nil, nil, err
	}

	teamsMeta, teams, netByTeamID, err := loadTeams(ctx, pool, tournamentID)
	if err != nil {
		return nil, nil, err
	}

	br, err := bracketBuilder.BuildBracket(tournamentID, teams, finalFour)
	if err != nil {
		return nil, nil, err
	}

	selectedGameOutcomeRunID, probs, nPred, err := loadPredictedGameOutcomesForTournament(ctx, pool, tournamentID, gameOutcomeRunID)
	if err != nil {
		return nil, nil, err
	}
	if nPred == 0 {
		return selectedGameOutcomeRunID, nil, fmt.Errorf("no predicted_game_outcomes found for tournament_id=%s", tournamentID)
	}

	// We only need round reach probabilities here, but reuse the DP implementation.
	_, reachByTeam, err := computeExpectedValueFromPGO(br, probs, []scoringRule{{WinIndex: 0, PointsAwarded: 0}}, netByTeamID, kenpomScale)
	if err != nil {
		return selectedGameOutcomeRunID, nil, err
	}

	out := make([]ports.TournamentPredictedAdvancementData, 0, len(teamsMeta))
	for _, tm := range teamsMeta {
		rr := reachByTeam[tm.TeamID]

		probPI := 0.0
		if rr.ReachFirstFour > 0 {
			probPI = rr.ReachR64
		}

		out = append(out, ports.TournamentPredictedAdvancementData{
			TeamID:     tm.TeamID,
			SchoolName: tm.SchoolName,
			Seed:       tm.Seed,
			Region:     tm.Region,
			ProbPI:     probPI,
			ReachR64:   rr.ReachR64,
			ReachR32:   rr.ReachR32,
			ReachS16:   rr.ReachS16,
			ReachE8:    rr.ReachE8,
			ReachFF:    rr.ReachFF,
			ReachChamp: rr.ReachChamp,
			WinChamp:   rr.WinChamp,
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].WinChamp != out[j].WinChamp {
			return out[i].WinChamp > out[j].WinChamp
		}
		if out[i].ReachChamp != out[j].ReachChamp {
			return out[i].ReachChamp > out[j].ReachChamp
		}
		if out[i].Seed != out[j].Seed {
			return out[i].Seed < out[j].Seed
		}
		return out[i].SchoolName < out[j].SchoolName
	})

	return selectedGameOutcomeRunID, out, nil
}

func loadFinalFourForTournament(ctx context.Context, pool *pgxpool.Pool, tournamentID string) (*models.FinalFourConfig, error) {
	var tl, bl, tr, br *string
	if err := pool.QueryRow(ctx, `
		SELECT final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right
		FROM core.tournaments
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, tournamentID).Scan(&tl, &bl, &tr, &br); err != nil {
		return nil, err
	}

	cfg := &models.FinalFourConfig{}
	if tl != nil {
		cfg.TopLeftRegion = *tl
	}
	if bl != nil {
		cfg.BottomLeftRegion = *bl
	}
	if tr != nil {
		cfg.TopRightRegion = *tr
	}
	if br != nil {
		cfg.BottomRightRegion = *br
	}

	if cfg.TopLeftRegion == "" {
		cfg.TopLeftRegion = "East"
	}
	if cfg.BottomLeftRegion == "" {
		cfg.BottomLeftRegion = "West"
	}
	if cfg.TopRightRegion == "" {
		cfg.TopRightRegion = "South"
	}
	if cfg.BottomRightRegion == "" {
		cfg.BottomRightRegion = "Midwest"
	}

	return cfg, nil
}
