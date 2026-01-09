package simulate_tournaments

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"

	appbracket "github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_game_outcomes"
	tsim "github.com/andrewcopp/Calcutta/backend/internal/app/tournament_simulation"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	pool *pgxpool.Pool
}

type kenPomProvider struct {
	spec        *simulation_game_outcomes.Spec
	netByTeamID map[string]float64
	overrides   map[tsim.MatchupKey]float64
}

func (p kenPomProvider) Prob(gameID string, team1ID string, team2ID string) float64 {
	if p.overrides != nil {
		if v, ok := p.overrides[tsim.MatchupKey{GameID: gameID, Team1ID: team1ID, Team2ID: team2ID}]; ok {
			return v
		}
	}
	if p.spec == nil {
		return 0.5
	}
	n1, ok1 := p.netByTeamID[team1ID]
	n2, ok2 := p.netByTeamID[team2ID]
	if !ok1 || !ok2 {
		return 0.5
	}
	return p.spec.WinProb(n1, n2)
}

func (s *Service) loadKenPomNetByTeamID(ctx context.Context, coreTournamentID string) (map[string]float64, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.id, ks.net_rtg
		FROM core.teams t
		LEFT JOIN core.team_kenpom_stats ks
			ON ks.team_id = t.id
			AND ks.deleted_at IS NULL
		WHERE t.tournament_id = $1::uuid
			AND t.deleted_at IS NULL
	`, coreTournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]float64)
	for rows.Next() {
		var teamID string
		var net *float64
		if err := rows.Scan(&teamID, &net); err != nil {
			return nil, err
		}
		if net != nil {
			out[teamID] = *net
		}
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
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
	GameOutcomeRunID     *string
	GameOutcomeSpec      *simulation_game_outcomes.Spec
}

type RunResult struct {
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
	if p.Seed == 0 {
		return nil, errors.New("Seed must be non-zero")
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
	coreTournamentID, err := s.resolveCoreTournamentID(ctx, p.Season)
	if err != nil {
		return nil, err
	}

	ff, err := s.loadFinalFourConfig(ctx, coreTournamentID)
	if err != nil {
		return nil, err
	}

	teams, err := s.loadTeams(ctx, coreTournamentID)
	if err != nil {
		return nil, err
	}

	builder := appbracket.NewBracketBuilder()
	br, err := builder.BuildBracket(coreTournamentID, teams, ff)
	if err != nil {
		return nil, fmt.Errorf("failed to build bracket: %w", err)
	}

	var provider tsim.ProbabilityProvider
	var probs map[tsim.MatchupKey]float64
	if p.GameOutcomeSpec != nil {
		p.GameOutcomeSpec.Normalize()
		if err := p.GameOutcomeSpec.Validate(); err != nil {
			return nil, err
		}
		netByTeamID, err := s.loadKenPomNetByTeamID(ctx, coreTournamentID)
		if err != nil {
			return nil, err
		}
		if len(netByTeamID) == 0 {
			return nil, errors.New("no kenpom ratings available for tournament")
		}
		overrides := make(map[tsim.MatchupKey]float64)
		if p.StartingStateKey == "post_first_four" {
			if err := s.lockInFirstFourResults(ctx, br, overrides); err != nil {
				return nil, err
			}
		}
		provider = kenPomProvider{spec: p.GameOutcomeSpec, netByTeamID: netByTeamID, overrides: overrides}
	} else {
		selectedGameOutcomeRunID, loaded, nPredRows, err := s.loadPredictedGameOutcomesForTournament(ctx, coreTournamentID, p.GameOutcomeRunID)
		if err != nil {
			return nil, err
		}
		if nPredRows == 0 {
			if selectedGameOutcomeRunID != nil {
				return nil, fmt.Errorf("no predicted_game_outcomes found for run_id=%s", *selectedGameOutcomeRunID)
			}
			return nil, fmt.Errorf("no predicted_game_outcomes found for tournament_id=%s", coreTournamentID)
		}
		probs = loaded
		if p.StartingStateKey == "post_first_four" {
			if err := s.lockInFirstFourResults(ctx, br, probs); err != nil {
				return nil, err
			}
		}
		provider = nil
	}

	loadDur := time.Since(loadStart)

	snapshotID := ""
	if p.StartingStateKey == "post_first_four" {
		created, err := s.createTournamentStateSnapshotFromBracket(ctx, coreTournamentID, br, teams)
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
		var results []tsim.TeamSimulationResult
		if provider != nil {
			results, err = tsim.SimulateWithProvider(br, provider, n, batchSeed, tsim.Options{Workers: p.Workers})
		} else {
			results, err = tsim.Simulate(br, probs, n, batchSeed, tsim.Options{Workers: p.Workers})
		}
		if err != nil {
			return nil, err
		}

		inserted, err := s.copyInsertSimulatedTournaments(ctx, batchID, coreTournamentID, offset, results)
		if err != nil {
			return nil, err
		}
		rowsWritten += inserted
	}

	simDur := time.Since(simStart)
	overallDur := time.Since(overallStart)

	return &RunResult{
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
	br *models.BracketStructure,
	probs map[tsim.MatchupKey]float64,
) error {
	if br == nil {
		return errors.New("bracket must not be nil")
	}
	if probs == nil {
		return errors.New("probs must not be nil")
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

		wins1, elim1, err := s.loadCoreTeamWinsEliminated(ctx, team1)
		if err != nil {
			return err
		}
		wins2, elim2, err := s.loadCoreTeamWinsEliminated(ctx, team2)
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
	coreTournamentID string,
	br *models.BracketStructure,
	teams []*models.TournamentTeam,
) (string, error) {
	if br == nil {
		return "", errors.New("bracket must not be nil")
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

	for _, t := range teams {
		if t == nil || t.ID == "" {
			continue
		}
		coreTeamID := t.ID
		wins, byes, eliminated := models.CalculateWinsAndByes(coreTeamID, br)
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

func (s *Service) resolveCoreTournamentID(ctx context.Context, season int) (string, error) {
	var id string
	if err := s.pool.QueryRow(ctx, `
		SELECT t.id
		FROM core.tournaments t
		JOIN core.seasons s
			ON s.id = t.season_id
			AND s.deleted_at IS NULL
		WHERE s.year = $1::int
			AND t.deleted_at IS NULL
		ORDER BY t.created_at DESC
		LIMIT 1
	`, season).Scan(&id); err != nil {
		return "", err
	}
	return id, nil
}

func (s *Service) loadFinalFourConfig(ctx context.Context, coreTournamentID string) (*models.FinalFourConfig, error) {
	var tl, bl, tr, br *string
	err := s.pool.QueryRow(ctx, `
		SELECT final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right
		FROM core.tournaments
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, coreTournamentID).Scan(&tl, &bl, &tr, &br)
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

func (s *Service) loadTeams(ctx context.Context, coreTournamentID string) ([]*models.TournamentTeam, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			t.id,
			t.seed,
			t.region,
			s.name
		FROM core.teams t
		JOIN core.schools s
			ON s.id = t.school_id
			AND s.deleted_at IS NULL
		WHERE t.tournament_id = $1::uuid
			AND t.deleted_at IS NULL
		ORDER BY t.seed ASC, s.name ASC
	`, coreTournamentID)
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

func (s *Service) loadPredictedGameOutcomesForTournament(ctx context.Context, tournamentID string, gameOutcomeRunID *string) (*string, map[tsim.MatchupKey]float64, int, error) {
	if gameOutcomeRunID != nil && *gameOutcomeRunID != "" {
		out, n, err := s.loadPredictedGameOutcomesByRunID(ctx, *gameOutcomeRunID)
		if err != nil {
			return nil, nil, 0, err
		}
		if n == 0 {
			return nil, nil, 0, fmt.Errorf("no predicted_game_outcomes found for run_id=%s", *gameOutcomeRunID)
		}
		return gameOutcomeRunID, out, n, nil
	}

	var latestRunID string
	if err := s.pool.QueryRow(ctx, `
		SELECT id
		FROM derived.game_outcome_runs
		WHERE tournament_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, tournamentID).Scan(&latestRunID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, 0, fmt.Errorf("no game_outcome_runs found for tournament_id=%s", tournamentID)
		}
		return nil, nil, 0, err
	}

	ptr := &latestRunID
	out, n, err := s.loadPredictedGameOutcomesByRunID(ctx, latestRunID)
	if err != nil {
		return nil, nil, 0, err
	}
	if n == 0 {
		return nil, nil, 0, fmt.Errorf("no predicted_game_outcomes found for run_id=%s", latestRunID)
	}
	return ptr, out, n, nil
}

func (s *Service) loadPredictedGameOutcomesByRunID(ctx context.Context, runID string) (map[tsim.MatchupKey]float64, int, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT game_id, team1_id, team2_id, p_team1_wins
		FROM derived.predicted_game_outcomes
		WHERE run_id = $1::uuid
			AND deleted_at IS NULL
	`, runID)
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
	tournamentID string,
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
		tournamentID: tournamentID,
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
