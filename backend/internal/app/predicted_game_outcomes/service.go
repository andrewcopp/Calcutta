package predicted_game_outcomes

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"strings"

	dbadapter "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	appbracket "github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	"github.com/andrewcopp/Calcutta/backend/internal/mathutil"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func defaultKenPomScaleForModelVersion(modelVersion string) float64 {
	mv := strings.TrimSpace(strings.ToLower(modelVersion))
	if mv == "kenpom-v1-sigma11-go" {
		return 11.0
	}
	return 10.0
}

func normalizeKenPomScale(scale float64, modelVersion string) float64 {
	if scale <= 0 {
		return defaultKenPomScaleForModelVersion(modelVersion)
	}

	mv := strings.TrimSpace(strings.ToLower(modelVersion))
	if mv == "kenpom-v1-sigma11-go" && scale == 10.0 {
		return 11.0
	}
	return scale
}

type Service struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

type GenerateParams struct {
	Season       int
	KenPomScale  float64
	NSims        int
	Seed         int
	ModelVersion string
}

type persistedParams struct {
	Season       int     `json:"season"`
	KenPomScale  float64 `json:"kenpom_scale"`
	NSims        int     `json:"n_sims"`
	Seed         int     `json:"seed"`
	ModelVersion string  `json:"model_version"`
}

type matchupKey struct {
	gameID  string
	team1ID string
	team2ID string
}

type gameMeta struct {
	gameID string
	round  models.BracketRound
	order  int
	sort   int
}

type predictionRow struct {
	tournamentID string
	gameID       string
	roundInt     int
	team1ID      string
	team2ID      string
	pTeam1Wins   float64
	pMatchup     float64
	modelVersion *string
}

func (s *Service) GenerateAndWrite(ctx context.Context, p GenerateParams) (string, int, error) {
	if p.Season <= 0 {
		return "", 0, errors.New("Season must be positive")
	}
	if p.NSims <= 0 {
		return "", 0, errors.New("NSims must be positive")
	}

	p.KenPomScale = normalizeKenPomScale(p.KenPomScale, p.ModelVersion)
	if p.KenPomScale <= 0 {
		return "", 0, errors.New("KenPomScale must be positive")
	}

	coreTournamentID, err := dbadapter.ResolveCoreTournamentID(ctx, s.pool, p.Season)
	if err != nil {
		return "", 0, err
	}

	ff, err := dbadapter.LoadFinalFourConfig(ctx, s.pool, coreTournamentID)
	if err != nil {
		return "", 0, err
	}

	teams, netByTeamID, err := s.loadTeams(ctx, coreTournamentID)
	if err != nil {
		return "", 0, err
	}

	builder := appbracket.NewBracketBuilder()
	br, err := builder.BuildBracket(coreTournamentID, teams, ff)
	if err != nil {
		return "", 0, fmt.Errorf("failed to build bracket: %w", err)
	}

	games, prevByNext, metaByGame := prepareGames(br)

	matchupCounts := make(map[matchupKey]int)
	team1WinCounts := make(map[matchupKey]int)

	rng := rand.New(rand.NewSource(int64(p.Seed)))
	for i := 0; i < p.NSims; i++ {
		winnersByGame := make(map[string]string, len(games))

		for _, g := range games {
			if g == nil || g.GameID == "" {
				continue
			}
			gid := g.GameID

			t1 := ""
			t2 := ""
			if g.Team1 != nil {
				t1 = g.Team1.TeamID
			}
			if g.Team2 != nil {
				t2 = g.Team2.TeamID
			}

			if t1 == "" {
				slots := prevByNext[gid]
				if slots != nil {
					if prev := slots[1]; prev != "" {
						t1 = winnersByGame[prev]
					}
				}
			}
			if t2 == "" {
				slots := prevByNext[gid]
				if slots != nil {
					if prev := slots[2]; prev != "" {
						t2 = winnersByGame[prev]
					}
				}
			}

			if t1 == "" || t2 == "" {
				continue
			}

			net1, ok1 := netByTeamID[t1]
			net2, ok2 := netByTeamID[t2]
			if !ok1 || !ok2 {
				continue
			}

			p1 := mathutil.WinProb(net1, net2, p.KenPomScale)
			winner := t2
			if rng.Float64() < p1 {
				winner = t1
			}
			winnersByGame[gid] = winner

			k := matchupKey{gameID: gid, team1ID: t1, team2ID: t2}
			matchupCounts[k] = matchupCounts[k] + 1
			if winner == t1 {
				team1WinCounts[k] = team1WinCounts[k] + 1
			}
		}
	}

	nSimsF := float64(p.NSims)
	rows := make([]predictionRow, 0, len(matchupCounts))
	for k, c := range matchupCounts {
		if c <= 0 {
			continue
		}

		meta := metaByGame[k.gameID]
		pMatchup := 0.0
		if nSimsF > 0 {
			pMatchup = float64(c) / nSimsF
		}

		w1 := team1WinCounts[k]
		pTeam1 := float64(w1) / float64(c)

		roundInt := meta.round.StorageInt()
		mv := (*string)(nil)
		if p.ModelVersion != "" {
			mv = &p.ModelVersion
		}

		rows = append(rows, predictionRow{
			tournamentID: coreTournamentID,
			gameID:       meta.gameID,
			roundInt:     roundInt,
			team1ID:      k.team1ID,
			team2ID:      k.team2ID,
			pTeam1Wins:   pTeam1,
			pMatchup:     pMatchup,
			modelVersion: mv,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		a := rows[i]
		b := rows[j]
		ma := metaByGame[a.gameID]
		mb := metaByGame[b.gameID]
		if ma.order != mb.order {
			return ma.order < mb.order
		}
		if ma.sort != mb.sort {
			return ma.sort < mb.sort
		}
		if a.gameID != b.gameID {
			return a.gameID < b.gameID
		}
		if a.pMatchup != b.pMatchup {
			return a.pMatchup > b.pMatchup
		}
		if a.team1ID != b.team1ID {
			return a.team1ID < b.team1ID
		}
		return a.team2ID < b.team2ID
	})

	if err := s.writePredictedGameOutcomes(ctx, coreTournamentID, p, rows); err != nil {
		return "", 0, err
	}

	return coreTournamentID, len(rows), nil
}

func (s *Service) GenerateAndWriteToExistingRun(ctx context.Context, runID string) (string, int, error) {
	if runID == "" {
		return "", 0, errors.New("runID is required")
	}

	var coreTournamentID string
	var paramsRaw []byte
	var modelVersion *string
	if err := s.pool.QueryRow(ctx, `
		SELECT
			r.tournament_id::text,
			r.params_json,
			COALESCE(NULLIF(pm.name, ''), NULL) AS model_version
		FROM derived.game_outcome_runs r
		LEFT JOIN derived.prediction_models pm ON pm.id = COALESCE(r.prediction_model_id, r.algorithm_id) AND pm.deleted_at IS NULL
		WHERE r.id = $1::uuid
			AND r.deleted_at IS NULL
		LIMIT 1
	`, runID).Scan(&coreTournamentID, &paramsRaw, &modelVersion); err != nil {
		return "", 0, err
	}
	if coreTournamentID == "" {
		return "", 0, errors.New("game_outcome_run missing tournament_id")
	}

	pp := persistedParams{}
	if len(paramsRaw) > 0 {
		_ = json.Unmarshal(paramsRaw, &pp)
	}
	if modelVersion != nil && *modelVersion != "" {
		pp.ModelVersion = *modelVersion
	}

	pp.KenPomScale = normalizeKenPomScale(pp.KenPomScale, pp.ModelVersion)
	if pp.NSims <= 0 {
		pp.NSims = 5000
	}
	if pp.Seed == 0 {
		pp.Seed = 42
	}

	ff, err := dbadapter.LoadFinalFourConfig(ctx, s.pool, coreTournamentID)
	if err != nil {
		return "", 0, err
	}

	teams, netByTeamID, err := s.loadTeams(ctx, coreTournamentID)
	if err != nil {
		return "", 0, err
	}

	builder := appbracket.NewBracketBuilder()
	br, err := builder.BuildBracket(coreTournamentID, teams, ff)
	if err != nil {
		return "", 0, fmt.Errorf("failed to build bracket: %w", err)
	}

	games, prevByNext, metaByGame := prepareGames(br)

	matchupCounts := make(map[matchupKey]int)
	team1WinCounts := make(map[matchupKey]int)

	rng := rand.New(rand.NewSource(int64(pp.Seed)))
	for i := 0; i < pp.NSims; i++ {
		winnersByGame := make(map[string]string, len(games))

		for _, g := range games {
			if g == nil || g.GameID == "" {
				continue
			}
			gid := g.GameID

			t1 := ""
			t2 := ""
			if g.Team1 != nil {
				t1 = g.Team1.TeamID
			}
			if g.Team2 != nil {
				t2 = g.Team2.TeamID
			}

			if t1 == "" {
				slots := prevByNext[gid]
				if slots != nil {
					if prev := slots[1]; prev != "" {
						t1 = winnersByGame[prev]
					}
				}
			}
			if t2 == "" {
				slots := prevByNext[gid]
				if slots != nil {
					if prev := slots[2]; prev != "" {
						t2 = winnersByGame[prev]
					}
				}
			}

			if t1 == "" || t2 == "" {
				continue
			}

			net1, ok1 := netByTeamID[t1]
			net2, ok2 := netByTeamID[t2]
			if !ok1 || !ok2 {
				continue
			}

			p1 := mathutil.WinProb(net1, net2, pp.KenPomScale)
			winner := t2
			if rng.Float64() < p1 {
				winner = t1
			}
			winnersByGame[gid] = winner

			k := matchupKey{gameID: gid, team1ID: t1, team2ID: t2}
			matchupCounts[k] = matchupCounts[k] + 1
			if winner == t1 {
				team1WinCounts[k] = team1WinCounts[k] + 1
			}
		}
	}

	nSimsF := float64(pp.NSims)
	rows := make([]predictionRow, 0, len(matchupCounts))
	for k, c := range matchupCounts {
		if c <= 0 {
			continue
		}

		meta := metaByGame[k.gameID]
		pMatchup := 0.0
		if nSimsF > 0 {
			pMatchup = float64(c) / nSimsF
		}

		w1 := team1WinCounts[k]
		pTeam1 := float64(w1) / float64(c)

		roundInt := meta.round.StorageInt()
		mv := (*string)(nil)
		if pp.ModelVersion != "" {
			mv = &pp.ModelVersion
		}

		rows = append(rows, predictionRow{
			tournamentID: coreTournamentID,
			gameID:       meta.gameID,
			roundInt:     roundInt,
			team1ID:      k.team1ID,
			team2ID:      k.team2ID,
			pTeam1Wins:   pTeam1,
			pMatchup:     pMatchup,
			modelVersion: mv,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		a := rows[i]
		b := rows[j]
		ma := metaByGame[a.gameID]
		mb := metaByGame[b.gameID]
		if ma.order != mb.order {
			return ma.order < mb.order
		}
		if ma.sort != mb.sort {
			return ma.sort < mb.sort
		}
		if a.gameID != b.gameID {
			return a.gameID < b.gameID
		}
		if a.pMatchup != b.pMatchup {
			return a.pMatchup > b.pMatchup
		}
		if a.team1ID != b.team1ID {
			return a.team1ID < b.team1ID
		}
		return a.team2ID < b.team2ID
	})

	gp := GenerateParams{Season: pp.Season, KenPomScale: pp.KenPomScale, NSims: pp.NSims, Seed: pp.Seed, ModelVersion: pp.ModelVersion}
	if err := s.writePredictedGameOutcomesForRun(ctx, coreTournamentID, gp, rows, runID); err != nil {
		return "", 0, err
	}

	return coreTournamentID, len(rows), nil
}

func (s *Service) loadTeams(
	ctx context.Context,
	coreTournamentID string,
) ([]*models.TournamentTeam, map[string]float64, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			t.id,
			t.seed,
			t.region,
			s.name,
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
	`, coreTournamentID)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	teams := make([]*models.TournamentTeam, 0)
	netByID := make(map[string]float64)

	for rows.Next() {
		var id string
		var seed *int
		var region *string
		var schoolName string
		var kenpomNet *float64
		if err := rows.Scan(&id, &seed, &region, &schoolName, &kenpomNet); err != nil {
			return nil, nil, err
		}

		seedVal := 0
		if seed != nil {
			seedVal = *seed
		}
		regionVal := ""
		if region != nil {
			regionVal = *region
		}

		teams = append(teams, &models.TournamentTeam{
			ID:     id,
			Seed:   seedVal,
			Region: regionVal,
			School: &models.School{Name: schoolName},
		})
		if kenpomNet != nil {
			netByID[id] = *kenpomNet
		}
	}
	if rows.Err() != nil {
		return nil, nil, rows.Err()
	}

	if len(teams) == 0 {
		return nil, nil, errors.New("no teams found")
	}

	return teams, netByID, nil
}

func (s *Service) writePredictedGameOutcomes(
	ctx context.Context,
	labTournamentID string,
	p GenerateParams,
	rows []predictionRow,
) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	params := map[string]any{
		"season":        p.Season,
		"kenpom_scale":  p.KenPomScale,
		"n_sims":        p.NSims,
		"seed":          p.Seed,
		"model_version": p.ModelVersion,
	}
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return err
	}

	predictionModelName := p.ModelVersion
	if predictionModelName == "" {
		predictionModelName = "kenpom"
	}

	var predictionModelID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.prediction_models (kind, name, params_json)
		VALUES ('game_outcomes', $1, $2::jsonb)
		ON CONFLICT (kind, name) WHERE deleted_at IS NULL
		DO UPDATE SET
			params_json = EXCLUDED.params_json,
			updated_at = NOW()
		RETURNING id
	`, predictionModelName, string(paramsJSON)).Scan(&predictionModelID); err != nil {
		return err
	}

	var runID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.game_outcome_runs (algorithm_id, prediction_model_id, tournament_id, params_json)
		VALUES ($1::uuid, $1::uuid, $2::uuid, $3::jsonb)
		RETURNING id
	`, predictionModelID, labTournamentID, string(paramsJSON)).Scan(&runID); err != nil {
		return err
	}

	// Legacy cleanup (temporary): remove old tournament-scoped rows so consumers don't
	// accidentally read stale data.
	_, err = tx.Exec(ctx, `
		DELETE FROM derived.predicted_game_outcomes
		WHERE tournament_id = $1::uuid
			AND run_id IS NULL
	`, labTournamentID)
	if err != nil {
		return err
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"derived", "predicted_game_outcomes"},
		[]string{"tournament_id", "game_id", "round", "team1_id", "team2_id", "p_team1_wins", "p_matchup", "model_version", "run_id"},
		&predictionSource{rows: rows, idx: 0, runID: runID},
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

func (s *Service) writePredictedGameOutcomesForRun(
	ctx context.Context,
	coreTournamentID string,
	p GenerateParams,
	rows []predictionRow,
	runID string,
) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	// Idempotency: clear any previously written rows for this run.
	_, err = tx.Exec(ctx, `
		DELETE FROM derived.predicted_game_outcomes
		WHERE run_id = $1::uuid
	`, runID)
	if err != nil {
		return err
	}

	// Legacy cleanup (temporary): remove old tournament-scoped rows so consumers don't
	// accidentally read stale data.
	_, err = tx.Exec(ctx, `
		DELETE FROM derived.predicted_game_outcomes
		WHERE tournament_id = $1::uuid
			AND run_id IS NULL
	`, coreTournamentID)
	if err != nil {
		return err
	}

	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"derived", "predicted_game_outcomes"},
		[]string{"tournament_id", "game_id", "round", "team1_id", "team2_id", "p_team1_wins", "p_matchup", "model_version", "run_id"},
		&predictionSource{rows: rows, idx: 0, runID: runID},
	)
	if err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}
	return nil
}

type predictionSource struct {
	rows  []predictionRow
	idx   int
	runID string
}

func (s *predictionSource) Next() bool {
	return s.idx < len(s.rows)
}

func (s *predictionSource) Values() ([]any, error) {
	r := s.rows[s.idx]
	s.idx++
	return []any{
		r.tournamentID,
		r.gameID,
		r.roundInt,
		r.team1ID,
		r.team2ID,
		r.pTeam1Wins,
		r.pMatchup,
		r.modelVersion,
		s.runID,
	}, nil
}

func (s *predictionSource) Err() error { return nil }

func prepareGames(bracket *models.BracketStructure) ([]*models.BracketGame, map[string]map[int]string, map[string]gameMeta) {
	games := make([]*models.BracketGame, 0, len(bracket.Games))
	prevByNext := make(map[string]map[int]string)
	metaByGame := make(map[string]gameMeta)

	for _, g := range bracket.Games {
		if g == nil {
			continue
		}
		games = append(games, g)
		metaByGame[g.GameID] = gameMeta{
			gameID: g.GameID,
			round:  g.Round,
			order:  g.Round.Order(),
			sort:   g.SortOrder,
		}
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

	return games, prevByNext, metaByGame
}
