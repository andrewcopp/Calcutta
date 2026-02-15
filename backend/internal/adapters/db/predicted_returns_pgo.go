package db

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sort"

	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type matchupKey struct {
	GameID  string
	Team1ID string
	Team2ID string
}

type scoringRule struct {
	WinIndex      int
	PointsAwarded int
}

type roundReach struct {
	ReachFirstFour float64
	ReachR64       float64
	ReachR32       float64
	ReachS16       float64
	ReachE8        float64
	ReachFF        float64
	ReachChamp     float64
	WinChamp       float64
}

type teamMeta struct {
	TeamID     string
	SchoolName string
	Seed       int
	Region     string
}

func computeCalcuttaPredictedReturnsFromPGO(ctx context.Context, pool *pgxpool.Pool, bracketBuilder ports.BracketBuilder, calcuttaID string, gameOutcomeRunID *string) (*string, []ports.CalcuttaPredictedReturnsData, error) {
	if calcuttaID == "" {
		return nil, nil, errors.New("calcuttaID is required")
	}

	kenpomScale := 10.0

	coreTournamentID, finalFour, err := loadTournamentForCalcutta(ctx, pool, calcuttaID)
	if err != nil {
		return nil, nil, err
	}

	scoringRules, err := loadScoringRules(ctx, pool, calcuttaID)
	if err != nil {
		return nil, nil, err
	}
	if len(scoringRules) == 0 {
		return nil, nil, errors.New("no calcutta scoring rules found")
	}

	teamsMeta, teams, netByTeamID, err := loadTeams(ctx, pool, coreTournamentID)
	if err != nil {
		return nil, nil, err
	}

	br, err := bracketBuilder.BuildBracket(coreTournamentID, teams, finalFour)
	if err != nil {
		return nil, nil, err
	}

	selectedGameOutcomeRunID, probs, nPred, err := loadPredictedGameOutcomesForTournament(ctx, pool, coreTournamentID, gameOutcomeRunID)
	if err != nil {
		return nil, nil, err
	}
	if nPred == 0 {
		return selectedGameOutcomeRunID, nil, fmt.Errorf("no predicted_game_outcomes found for tournament_id=%s", coreTournamentID)
	}

	evByTeam, reachByTeam, err := computeExpectedValueFromPGO(br, probs, scoringRules, netByTeamID, kenpomScale)
	if err != nil {
		return selectedGameOutcomeRunID, nil, err
	}

	out := make([]ports.CalcuttaPredictedReturnsData, 0, len(teamsMeta))
	for _, tm := range teamsMeta {
		rr := reachByTeam[tm.TeamID]

		probPI := 0.0
		if rr.ReachFirstFour > 0 {
			probPI = rr.ReachR64
		}

		out = append(out, ports.CalcuttaPredictedReturnsData{
			TeamID:        tm.TeamID,
			SchoolName:    tm.SchoolName,
			Seed:          tm.Seed,
			Region:        tm.Region,
			ProbPI:        probPI,
			ProbR64:       rr.ReachR32,
			ProbR32:       rr.ReachS16,
			ProbS16:       rr.ReachE8,
			ProbE8:        rr.ReachFF,
			ProbFF:        rr.ReachChamp,
			ProbChamp:     rr.WinChamp,
			ExpectedValue: evByTeam[tm.TeamID],
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].ExpectedValue != out[j].ExpectedValue {
			return out[i].ExpectedValue > out[j].ExpectedValue
		}
		return out[i].Seed < out[j].Seed
	})

	return selectedGameOutcomeRunID, out, nil
}

func loadTournamentForCalcutta(ctx context.Context, pool *pgxpool.Pool, calcuttaID string) (string, *models.FinalFourConfig, error) {
	var tournamentID string
	var tl, bl, tr, br *string
	if err := pool.QueryRow(ctx, `
		SELECT t.id, t.final_four_top_left, t.final_four_bottom_left, t.final_four_top_right, t.final_four_bottom_right
		FROM core.calcuttas c
		JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
		WHERE c.id = $1::uuid
			AND c.deleted_at IS NULL
		LIMIT 1
	`, calcuttaID).Scan(&tournamentID, &tl, &bl, &tr, &br); err != nil {
		return "", nil, err
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

	return tournamentID, cfg, nil
}

func loadScoringRules(ctx context.Context, pool *pgxpool.Pool, calcuttaID string) ([]scoringRule, error) {
	rows, err := pool.Query(ctx, `
		SELECT win_index, points_awarded
		FROM core.calcutta_scoring_rules
		WHERE calcutta_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY win_index ASC
	`, calcuttaID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]scoringRule, 0)
	for rows.Next() {
		var r scoringRule
		if err := rows.Scan(&r.WinIndex, &r.PointsAwarded); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func loadTeams(ctx context.Context, pool *pgxpool.Pool, tournamentID string) ([]teamMeta, []*models.TournamentTeam, map[string]float64, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			t.id,
			COALESCE(t.seed, 0)::int,
			COALESCE(t.region, '')::text,
			s.name,
			s.id,
			ks.net_rtg
		FROM core.teams t
		JOIN core.schools s
			ON s.id = t.school_id
			AND s.deleted_at IS NULL
		LEFT JOIN core.team_kenpom_stats ks
			ON ks.team_id = t.id
			AND ks.deleted_at IS NULL
		WHERE t.tournament_id = $1::uuid
			AND t.deleted_at IS NULL
		ORDER BY t.seed ASC, s.name ASC
	`, tournamentID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	metas := make([]teamMeta, 0)
	teams := make([]*models.TournamentTeam, 0)
	netByID := make(map[string]float64)
	for rows.Next() {
		var id string
		var seed int
		var region string
		var schoolName string
		var schoolID string
		var kenpomNet *float64
		if err := rows.Scan(&id, &seed, &region, &schoolName, &schoolID, &kenpomNet); err != nil {
			return nil, nil, nil, err
		}

		metas = append(metas, teamMeta{TeamID: id, SchoolName: schoolName, Seed: seed, Region: region})
		teams = append(teams, &models.TournamentTeam{ID: id, Seed: seed, Region: region, School: &models.School{ID: schoolID, Name: schoolName}})
		if kenpomNet != nil {
			netByID[id] = *kenpomNet
		}
	}
	if rows.Err() != nil {
		return nil, nil, nil, rows.Err()
	}
	if len(teams) != 68 {
		return nil, nil, nil, fmt.Errorf("expected 68 teams, got %d", len(teams))
	}
	return metas, teams, netByID, nil
}

func loadPredictedGameOutcomesForTournament(ctx context.Context, pool *pgxpool.Pool, tournamentID string, gameOutcomeRunID *string) (*string, map[matchupKey]float64, int, error) {
	if gameOutcomeRunID != nil {
		out, n, err := loadPredictedGameOutcomesByRunID(ctx, pool, *gameOutcomeRunID)
		if err != nil {
			return nil, nil, 0, err
		}
		if n == 0 {
			return nil, nil, 0, fmt.Errorf("no predicted_game_outcomes found for run_id=%s", *gameOutcomeRunID)
		}
		return gameOutcomeRunID, out, n, nil
	}

	var latestRunID string
	if err := pool.QueryRow(ctx, `
		SELECT id
		FROM derived.game_outcome_runs
		WHERE tournament_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, tournamentID).Scan(&latestRunID); err == nil {
		latestRunIDPtr := &latestRunID
		out, n, err := loadPredictedGameOutcomesByRunID(ctx, pool, latestRunID)
		if err != nil {
			return nil, nil, 0, err
		}
		if n == 0 {
			return nil, nil, 0, fmt.Errorf("no predicted_game_outcomes found for run_id=%s", latestRunID)
		}
		return latestRunIDPtr, out, n, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, nil, 0, err
	}

	return nil, nil, 0, fmt.Errorf("no game_outcome_runs found for tournament_id=%s", tournamentID)
}

func loadPredictedGameOutcomesByRunID(ctx context.Context, pool *pgxpool.Pool, runID string) (map[matchupKey]float64, int, error) {
	rows, err := pool.Query(ctx, `
		SELECT game_id, team1_id, team2_id, p_team1_wins
		FROM derived.predicted_game_outcomes
		WHERE run_id = $1::uuid
			AND deleted_at IS NULL
	`, runID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make(map[matchupKey]float64)
	n := 0
	for rows.Next() {
		var gameID, t1, t2 string
		var p float64
		if err := rows.Scan(&gameID, &t1, &t2, &p); err != nil {
			return nil, 0, err
		}
		n++
		out[matchupKey{GameID: gameID, Team1ID: t1, Team2ID: t2}] = p
		out[matchupKey{GameID: gameID, Team1ID: t2, Team2ID: t1}] = 1.0 - p
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}
	return out, n, nil
}

func computeExpectedValueFromPGO(
	br *models.BracketStructure,
	probs map[matchupKey]float64,
	scoringRules []scoringRule,
	netByTeamID map[string]float64,
	kenpomScale float64,
) (map[string]float64, map[string]roundReach, error) {
	if br == nil {
		return nil, nil, errors.New("bracket must not be nil")
	}
	if len(br.Games) == 0 {
		return nil, nil, errors.New("bracket must have games")
	}
	if len(scoringRules) == 0 {
		return nil, nil, errors.New("scoringRules must not be empty")
	}

	games := make([]*models.BracketGame, 0, len(br.Games))
	prevByNext := make(map[string]map[int]string)
	for _, g := range br.Games {
		if g == nil {
			continue
		}
		games = append(games, g)
		if g.NextGameID != "" && (g.NextGameSlot == 1 || g.NextGameSlot == 2) {
			slots := prevByNext[g.NextGameID]
			if slots == nil {
				slots = make(map[int]string)
				prevByNext[g.NextGameID] = slots
			}
			slots[g.NextGameSlot] = g.GameID
		}
	}

	sort.Slice(games, func(i, j int) bool {
		gi := games[i]
		gj := games[j]
		ri := gi.Round.Order()
		rj := gj.Round.Order()
		if ri != rj {
			return ri < rj
		}
		if gi.SortOrder != gj.SortOrder {
			return gi.SortOrder < gj.SortOrder
		}
		return gi.GameID < gj.GameID
	})

	winnerDistByGame := make(map[string]map[string]float64)
	reach := make(map[string]roundReach)
	champGameID := ""

	for _, g := range games {
		if g == nil || g.GameID == "" {
			continue
		}
		if g.Round == models.RoundChampionship {
			champGameID = g.GameID
		}

		slot1 := make(map[string]float64)
		slot2 := make(map[string]float64)

		if g.Team1 != nil && g.Team1.TeamID != "" {
			slot1[g.Team1.TeamID] = 1.0
		} else {
			prev := prevByNext[g.GameID][1]
			if prev == "" {
				return nil, nil, fmt.Errorf("game %s missing Team1 and missing prev slot 1", g.GameID)
			}
			wd := winnerDistByGame[prev]
			if wd == nil {
				return nil, nil, fmt.Errorf("game %s slot 1 depends on %s but winner distribution not computed", g.GameID, prev)
			}
			for tid, p := range wd {
				slot1[tid] = p
			}
		}

		if g.Team2 != nil && g.Team2.TeamID != "" {
			slot2[g.Team2.TeamID] = 1.0
		} else {
			prev := prevByNext[g.GameID][2]
			if prev == "" {
				return nil, nil, fmt.Errorf("game %s missing Team2 and missing prev slot 2", g.GameID)
			}
			wd := winnerDistByGame[prev]
			if wd == nil {
				return nil, nil, fmt.Errorf("game %s slot 2 depends on %s but winner distribution not computed", g.GameID, prev)
			}
			for tid, p := range wd {
				slot2[tid] = p
			}
		}

		for tid, p := range slot1 {
			r := reach[tid]
			applyReach(&r, g.Round, p)
			reach[tid] = r
		}
		for tid, p := range slot2 {
			r := reach[tid]
			applyReach(&r, g.Round, p)
			reach[tid] = r
		}

		winners := make(map[string]float64)
		for t1, pSlot1 := range slot1 {
			for t2, pSlot2 := range slot2 {
				pMatch := pSlot1 * pSlot2
				if pMatch == 0 {
					continue
				}
				p1, ok := probs[matchupKey{GameID: g.GameID, Team1ID: t1, Team2ID: t2}]
				if !ok {
					n1, ok1 := netByTeamID[t1]
					n2, ok2 := netByTeamID[t2]
					if ok1 && ok2 {
						p1 = winProb(n1, n2, kenpomScale)
					} else {
						p1 = 0.5
					}
				}
				winners[t1] += pMatch * p1
				winners[t2] += pMatch * (1.0 - p1)
			}
		}

		winnerDistByGame[g.GameID] = winners
	}

	if champGameID == "" {
		return nil, nil, errors.New("championship game not found in bracket")
	}

	champWinners := winnerDistByGame[champGameID]
	for tid, p := range champWinners {
		r := reach[tid]
		r.WinChamp = p
		reach[tid] = r
	}

	ev := make(map[string]float64)
	for tid, rr := range reach {
		pAtLeast := func(winIndex int) float64 {
			switch winIndex {
			case 0:
				return 1.0
			case 1:
				return rr.ReachR64
			case 2:
				return rr.ReachR32
			case 3:
				return rr.ReachS16
			case 4:
				return rr.ReachE8
			case 5:
				return rr.ReachFF
			case 6:
				return rr.ReachChamp
			case 7:
				return rr.WinChamp
			default:
				return 0.0
			}
		}

		val := 0.0
		for _, sr := range scoringRules {
			val += float64(sr.PointsAwarded) * pAtLeast(sr.WinIndex)
		}
		ev[tid] = val
	}

	return ev, reach, nil
}

func applyReach(r *roundReach, round models.BracketRound, p float64) {
	switch round {
	case models.RoundFirstFour:
		r.ReachFirstFour += p
	case models.RoundOf64:
		r.ReachR64 += p
	case models.RoundOf32:
		r.ReachR32 += p
	case models.RoundSweet16:
		r.ReachS16 += p
	case models.RoundElite8:
		r.ReachE8 += p
	case models.RoundFinalFour:
		r.ReachFF += p
	case models.RoundChampionship:
		r.ReachChamp += p
	}
}

func winProb(net1 float64, net2 float64, scale float64) float64 {
	if scale <= 0 {
		return 0.5
	}
	return sigmoid((net1 - net2) / scale)
}

func sigmoid(x float64) float64 {
	if x >= 0 {
		z := math.Exp(-x)
		return 1.0 / (1.0 + z)
	}
	z := math.Exp(x)
	return z / (1.0 + z)
}
