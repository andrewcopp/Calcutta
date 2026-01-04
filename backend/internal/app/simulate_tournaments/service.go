package simulate_tournaments

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	tsim "github.com/andrewcopp/Calcutta/backend/internal/app/tournament_simulation"
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

type RunParams struct {
	Season               int
	NSims                int
	Seed                 int
	Workers              int
	BatchSize            int
	ProbabilitySourceKey string
	StartingStateKey     string
}

type RunResult struct {
	LabTournamentID             string
	CoreTournamentID            string
	TournamentStateSnapshotID   string
	TournamentSimulationBatchID string
	NSims                       int
	RowsWritten                 int64
	LoadDuration                time.Duration
	SimulateWriteDuration       time.Duration
	OverallDuration             time.Duration
}

func (s *Service) Run(ctx context.Context, p RunParams) (*RunResult, error) {
	if p.Season <= 0 {
		return nil, errors.New("Season must be positive")
	}
	if p.NSims <= 0 {
		return nil, errors.New("NSims must be positive")
	}
	if p.BatchSize <= 0 {
		return nil, errors.New("BatchSize must be positive")
	}
	if p.Workers <= 0 {
		p.Workers = runtime.GOMAXPROCS(0)
		if p.Workers <= 0 {
			p.Workers = 1
		}
	}
	if p.ProbabilitySourceKey == "" {
		p.ProbabilitySourceKey = "go_worker"
	}
	if strings.TrimSpace(p.StartingStateKey) == "" {
		p.StartingStateKey = "current"
	}
	if p.StartingStateKey != "current" && p.StartingStateKey != "post_first_four" {
		return nil, errors.New("StartingStateKey must be 'current' or 'post_first_four'")
	}

	overallStart := time.Now()

	loadStart := time.Now()
	labTournamentID, coreTournamentID, err := s.resolveTournamentIDs(ctx, p.Season)
	if err != nil {
		return nil, err
	}

	ff, err := s.loadFinalFourConfig(ctx, labTournamentID)
	if err != nil {
		return nil, err
	}

	teams, err := s.loadLabTeams(ctx, labTournamentID)
	if err != nil {
		return nil, err
	}

	builder := appbracket.NewBracketBuilder()
	br, err := builder.BuildBracket(labTournamentID, teams, ff)
	if err != nil {
		return nil, fmt.Errorf("failed to build bracket: %w", err)
	}

	probs, nPredRows, err := s.loadPredictedGameOutcomes(ctx, labTournamentID)
	if err != nil {
		return nil, err
	}
	if nPredRows == 0 {
		return nil, fmt.Errorf("no predicted_game_outcomes found for tournament_id=%s", labTournamentID)
	}

	if p.StartingStateKey == "post_first_four" {
		if err := s.lockInFirstFourResults(ctx, labTournamentID, coreTournamentID, br, probs); err != nil {
			return nil, err
		}
	}

	loadDur := time.Since(loadStart)

	snapshotID := ""
	if p.StartingStateKey == "post_first_four" {
		created, err := s.createTournamentStateSnapshotFromBracket(ctx, labTournamentID, coreTournamentID, br)
		if err != nil {
			return nil, err
		}
		snapshotID = created
	} else {
		created, err := s.createTournamentStateSnapshot(ctx, coreTournamentID)
		if err != nil {
			return nil, err
		}
		snapshotID = created
	}
	batchID, err := s.createTournamentSimulationBatch(ctx, coreTournamentID, snapshotID, p.NSims, p.Seed, p.ProbabilitySourceKey)
	if err != nil {
		return nil, err
	}

	simStart := time.Now()
	rowsWritten := int64(0)

	for offset := 0; offset < p.NSims; offset += p.BatchSize {
		n := p.BatchSize
		if offset+n > p.NSims {
			n = p.NSims - offset
		}

		batchSeed := int64(p.Seed) + int64(offset)*1_000_003
		results, err := tsim.Simulate(br, probs, n, batchSeed, tsim.Options{Workers: p.Workers})
		if err != nil {
			return nil, err
		}

		inserted, err := s.copyInsertSimulatedTournaments(ctx, batchID, labTournamentID, offset, results)
		if err != nil {
			return nil, err
		}
		rowsWritten += inserted
	}

	simDur := time.Since(simStart)
	overallDur := time.Since(overallStart)

	return &RunResult{
		LabTournamentID:             labTournamentID,
		CoreTournamentID:            coreTournamentID,
		TournamentStateSnapshotID:   snapshotID,
		TournamentSimulationBatchID: batchID,
		NSims:                       p.NSims,
		RowsWritten:                 rowsWritten,
		LoadDuration:                loadDur,
		SimulateWriteDuration:       simDur,
		OverallDuration:             overallDur,
	}, nil
}

func (s *Service) lockInFirstFourResults(
	ctx context.Context,
	labTournamentID string,
	coreTournamentID string,
	br *models.BracketStructure,
	probs map[tsim.MatchupKey]float64,
) error {
	if br == nil {
		return errors.New("bracket must not be nil")
	}
	if probs == nil {
		return errors.New("probs must not be nil")
	}

	coreByLab, err := s.mapLabTeamIDToCoreTeamID(ctx, labTournamentID, coreTournamentID)
	if err != nil {
		return err
	}

	for _, g := range br.Games {
		if g == nil {
			continue
		}
		if g.Round != models.RoundFirstFour {
			continue
		}
		if g.Team1 == nil || g.Team2 == nil {
			continue
		}
		team1 := g.Team1.TeamID
		team2 := g.Team2.TeamID
		if team1 == "" || team2 == "" {
			continue
		}

		core1, ok1 := coreByLab[team1]
		core2, ok2 := coreByLab[team2]
		if !ok1 || !ok2 {
			return fmt.Errorf("failed to map first four teams to core: team1=%s ok=%v team2=%s ok=%v", team1, ok1, team2, ok2)
		}

		wins1, elim1, err := s.loadCoreTeamWinsEliminated(ctx, core1)
		if err != nil {
			return err
		}
		wins2, elim2, err := s.loadCoreTeamWinsEliminated(ctx, core2)
		if err != nil {
			return err
		}

		winner := ""
		if elim1 && !elim2 {
			winner = team2
		} else if elim2 && !elim1 {
			winner = team1
		} else if wins1 > wins2 {
			winner = team1
		} else if wins2 > wins1 {
			winner = team2
		} else {
			return fmt.Errorf("post_first_four requested but first four game not resolved for game_id=%s", g.GameID)
		}

		p1 := 0.0
		if winner == team1 {
			p1 = 1.0
			g.Winner = g.Team1
		} else {
			p1 = 0.0
			g.Winner = g.Team2
		}

		probs[tsim.MatchupKey{GameID: g.GameID, Team1ID: team1, Team2ID: team2}] = p1
		probs[tsim.MatchupKey{GameID: g.GameID, Team1ID: team2, Team2ID: team1}] = 1.0 - p1
	}

	return nil
}

func (s *Service) mapLabTeamIDToCoreTeamID(ctx context.Context, labTournamentID string, coreTournamentID string) (map[string]string, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT bt.id, tt.id
		FROM derived.teams bt
		JOIN core.schools s
			ON s.name = bt.school_name
			AND s.deleted_at IS NULL
		JOIN core.teams tt
			ON tt.school_id = s.id
			AND tt.tournament_id = $2
			AND tt.deleted_at IS NULL
		WHERE bt.tournament_id = $1
			AND bt.deleted_at IS NULL
	`, labTournamentID, coreTournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]string)
	for rows.Next() {
		var labID string
		var coreID string
		if err := rows.Scan(&labID, &coreID); err != nil {
			return nil, err
		}
		out[labID] = coreID
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *Service) loadCoreTeamWinsEliminated(ctx context.Context, coreTeamID string) (int, bool, error) {
	var wins int
	var eliminated bool
	err := s.pool.QueryRow(ctx, `
		SELECT wins, eliminated
		FROM core.teams
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, coreTeamID).Scan(&wins, &eliminated)
	if err != nil {
		return 0, false, err
	}
	return wins, eliminated, nil
}

func (s *Service) createTournamentStateSnapshotFromBracket(
	ctx context.Context,
	labTournamentID string,
	coreTournamentID string,
	br *models.BracketStructure,
) (string, error) {
	if br == nil {
		return "", errors.New("bracket must not be nil")
	}

	coreByLab, err := s.mapLabTeamIDToCoreTeamID(ctx, labTournamentID, coreTournamentID)
	if err != nil {
		return "", err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var snapshotID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.simulation_states (tournament_id, source, description)
		VALUES ($1, 'go_simulate_tournaments', 'Snapshot from bracket state (post_first_four)')
		RETURNING id
	`, coreTournamentID).Scan(&snapshotID); err != nil {
		return "", err
	}

	for labTeamID, coreTeamID := range coreByLab {
		wins, byes, eliminated := models.CalculateWinsAndByes(labTeamID, br)
		_, err := tx.Exec(ctx, `
			INSERT INTO derived.simulation_state_teams (
				simulation_state_id,
				team_id,
				wins,
				byes,
				eliminated
			)
			VALUES ($1, $2::uuid, $3, $4, $5)
			ON CONFLICT (simulation_state_id, team_id) DO NOTHING
		`, snapshotID, coreTeamID, wins, byes, eliminated)
		if err != nil {
			return "", err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return snapshotID, nil
}

func (s *Service) resolveTournamentIDs(ctx context.Context, season int) (string, string, error) {
	var labID string
	var coreID *string
	if err := s.pool.QueryRow(ctx, `
		SELECT id, core_tournament_id
		FROM derived.tournaments
		WHERE season = $1::int
		  AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, season).Scan(&labID, &coreID); err != nil {
		return "", "", err
	}
	if coreID == nil || *coreID == "" {
		return "", "", fmt.Errorf("core_tournament_id is NULL for season=%d", season)
	}
	return labID, *coreID, nil
}

func (s *Service) loadFinalFourConfig(ctx context.Context, labTournamentID string) (*models.FinalFourConfig, error) {
	var tl, bl, tr, br *string
	err := s.pool.QueryRow(ctx, `
		SELECT ct.final_four_top_left,
		       ct.final_four_bottom_left,
		       ct.final_four_top_right,
		       ct.final_four_bottom_right
		FROM derived.tournaments bt
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

func (s *Service) loadLabTeams(ctx context.Context, labTournamentID string) ([]*models.TournamentTeam, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, seed, region, school_name
		FROM derived.teams
		WHERE tournament_id = $1
		  AND deleted_at IS NULL
		ORDER BY seed ASC
	`, labTournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*models.TournamentTeam, 0)
	for rows.Next() {
		var id string
		var seed *int
		var region *string
		var schoolName string
		if err := rows.Scan(&id, &seed, &region, &schoolName); err != nil {
			return nil, err
		}

		seedVal := 0
		if seed != nil {
			seedVal = *seed
		}
		regionVal := ""
		if region != nil {
			regionVal = *region
		}

		out = append(out, &models.TournamentTeam{
			ID:     id,
			Seed:   seedVal,
			Region: regionVal,
			School: &models.School{Name: schoolName},
		})
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if len(out) != 68 {
		return nil, fmt.Errorf("expected 68 teams, got %d", len(out))
	}
	return out, nil
}

func (s *Service) loadPredictedGameOutcomes(ctx context.Context, labTournamentID string) (map[tsim.MatchupKey]float64, int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT game_id, team1_id, team2_id, p_team1_wins
		FROM derived.predicted_game_outcomes
		WHERE tournament_id = $1
		  AND deleted_at IS NULL
	`, labTournamentID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make(map[tsim.MatchupKey]float64)
	n := 0
	for rows.Next() {
		var gameID string
		var t1 string
		var t2 string
		var p float64
		if err := rows.Scan(&gameID, &t1, &t2, &p); err != nil {
			return nil, 0, err
		}
		n++
		out[tsim.MatchupKey{GameID: gameID, Team1ID: t1, Team2ID: t2}] = p
		out[tsim.MatchupKey{GameID: gameID, Team1ID: t2, Team2ID: t1}] = 1.0 - p
	}
	if rows.Err() != nil {
		return nil, 0, rows.Err()
	}
	return out, n, nil
}

func (s *Service) createTournamentStateSnapshot(ctx context.Context, coreTournamentID string) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var snapshotID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.simulation_states (tournament_id, source, description)
		VALUES ($1, 'go_simulate_tournaments', 'Autogenerated snapshot for tournament simulation batch')
		RETURNING id
	`, coreTournamentID).Scan(&snapshotID); err != nil {
		return "", err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO derived.simulation_state_teams (
			simulation_state_id,
			team_id,
			wins,
			byes,
			eliminated
		)
		SELECT
			$1,
			ct.id,
			ct.wins,
			ct.byes,
			ct.eliminated
		FROM core.teams ct
		WHERE ct.tournament_id = $2
		  AND ct.deleted_at IS NULL
		ON CONFLICT (simulation_state_id, team_id) DO NOTHING
	`, snapshotID, coreTournamentID)
	if err != nil {
		return "", err
	}

	if err := tx.Commit(ctx); err != nil {
		return "", err
	}

	return snapshotID, nil
}

func (s *Service) createTournamentSimulationBatch(
	ctx context.Context,
	coreTournamentID string,
	snapshotID string,
	nSims int,
	seed int,
	probabilitySourceKey string,
) (string, error) {
	var batchID string
	if err := s.pool.QueryRow(ctx, `
		INSERT INTO derived.simulated_tournaments (
			tournament_id,
			simulation_state_id,
			n_sims,
			seed,
			probability_source_key
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`, coreTournamentID, snapshotID, nSims, seed, probabilitySourceKey).Scan(&batchID); err != nil {
		return "", err
	}
	return batchID, nil
}

type simResultsSource struct {
	batchID      string
	tournamentID string
	simOffset    int
	results      []tsim.TeamSimulationResult
	idx          int
}

func (s *simResultsSource) Next() bool {
	return s.idx < len(s.results)
}

func (s *simResultsSource) Values() ([]any, error) {
	r := s.results[s.idx]
	s.idx++
	return []any{
		s.batchID,
		s.tournamentID,
		r.SimID + s.simOffset,
		r.TeamID,
		r.Wins,
		r.Byes,
		r.Eliminated,
	}, nil
}

func (s *simResultsSource) Err() error { return nil }

func (s *Service) copyInsertSimulatedTournaments(
	ctx context.Context,
	batchID string,
	labTournamentID string,
	simOffset int,
	results []tsim.TeamSimulationResult,
) (int64, error) {
	conn, err := s.pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	src := &simResultsSource{
		batchID:      batchID,
		tournamentID: labTournamentID,
		simOffset:    simOffset,
		results:      results,
		idx:          0,
	}

	inserted, err := conn.Conn().CopyFrom(
		ctx,
		pgx.Identifier{"derived", "simulated_teams"},
		[]string{"simulated_tournament_id", "tournament_id", "sim_id", "team_id", "wins", "byes", "eliminated"},
		src,
	)
	if err != nil {
		return 0, err
	}
	return inserted, nil
}
