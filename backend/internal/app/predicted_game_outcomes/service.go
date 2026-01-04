package predicted_game_outcomes

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"sort"

	appbracket "github.com/andrewcopp/Calcutta/backend/internal/features/bracket"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
	if p.KenPomScale <= 0 {
		return "", 0, errors.New("KenPomScale must be positive")
	}

	labTournamentID, err := s.resolveLabTournamentID(ctx, p.Season)
	if err != nil {
		return "", 0, err
	}

	ff, err := s.loadFinalFourConfig(ctx, labTournamentID)
	if err != nil {
		return "", 0, err
	}

	teams, netByTeamID, err := s.loadTeams(ctx, labTournamentID)
	if err != nil {
		return "", 0, err
	}

	builder := appbracket.NewBracketBuilder()
	br, err := builder.BuildBracket(labTournamentID, teams, ff)
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

			p1 := winProb(net1, net2, p.KenPomScale)
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
			tournamentID: labTournamentID,
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

	if err := s.writePredictedGameOutcomes(ctx, labTournamentID, rows); err != nil {
		return "", 0, err
	}

	return labTournamentID, len(rows), nil
}

func (s *Service) resolveLabTournamentID(ctx context.Context, season int) (string, error) {
	var id string
	if err := s.pool.QueryRow(ctx, `
		SELECT id
		FROM lab_bronze.tournaments
		WHERE season = $1::int
		  AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, season).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) loadTeams(
	ctx context.Context,
	labTournamentID string,
) ([]*models.TournamentTeam, map[string]float64, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, seed, region, school_name, kenpom_net
		FROM lab_bronze.teams
		WHERE tournament_id = $1
		  AND deleted_at IS NULL
		ORDER BY seed ASC
	`, labTournamentID)
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

func (s *Service) loadFinalFourConfig(ctx context.Context, labTournamentID string) (*models.FinalFourConfig, error) {
	var tl, bl, tr, br *string
	err := s.pool.QueryRow(ctx, `
		SELECT ct.final_four_top_left,
		       ct.final_four_bottom_left,
		       ct.final_four_top_right,
		       ct.final_four_bottom_right
		FROM lab_bronze.tournaments bt
		LEFT JOIN core.tournaments ct
		  ON ct.id = bt.core_tournament_id
		 AND ct.deleted_at IS NULL
		WHERE bt.id = $1
		  AND bt.deleted_at IS NULL
		LIMIT 1
	`, labTournamentID).Scan(&tl, &bl, &tr, &br)
	if err != nil {
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

func (s *Service) writePredictedGameOutcomes(
	ctx context.Context,
	labTournamentID string,
	rows []predictionRow,
) error {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	_, err = tx.Exec(ctx, `
		DELETE FROM lab_silver.predicted_game_outcomes
		WHERE tournament_id = $1
	`, labTournamentID)
	if err != nil {
		return err
	}

	src := &predictionSource{rows: rows, idx: 0}
	_, err = tx.CopyFrom(
		ctx,
		pgx.Identifier{"lab_silver", "predicted_game_outcomes"},
		[]string{"tournament_id", "game_id", "round", "team1_id", "team2_id", "p_team1_wins", "p_matchup", "model_version"},
		src,
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
	rows []predictionRow
	idx  int
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
	}, nil
}

func (s *predictionSource) Err() error { return nil }

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
