package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"runtime"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/bracket"
	tsim "github.com/andrewcopp/Calcutta/backend/internal/app/tournament_simulation"
	"github.com/andrewcopp/Calcutta/backend/internal/db"
	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type tournamentIDs struct {
	labTournamentID  string
	coreTournamentID string
}

type finalFourConfig struct {
	topLeft     string
	bottomLeft  string
	topRight    string
	bottomRight string
}

func main() {
	platform.InitLogger()
	if err := run(); err != nil {
		slog.Error("cmd_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	var season int
	var nSims int
	var seed int
	var workers int
	var batchSize int
	var probabilitySourceKey string

	flag.IntVar(&season, "season", 0, "Tournament season/year (e.g. 2025)")
	flag.IntVar(&nSims, "n-sims", 100000, "Number of Monte Carlo simulations")
	flag.IntVar(&seed, "seed", 42, "Base RNG seed")
	flag.IntVar(&workers, "workers", runtime.GOMAXPROCS(0), "Number of simulation workers")
	flag.IntVar(&batchSize, "batch-size", 1000, "Number of simulations per DB COPY batch")
	flag.StringVar(&probabilitySourceKey, "probability-source-key", "go_benchmark", "Stored in derived.simulated_tournaments.probability_source_key")
	flag.Parse()

	if season <= 0 {
		flag.Usage()
		return fmt.Errorf("--season is required")
	}
	if nSims <= 0 {
		flag.Usage()
		return fmt.Errorf("--n-sims must be positive")
	}
	if batchSize <= 0 {
		flag.Usage()
		return fmt.Errorf("--batch-size must be positive")
	}
	if workers <= 0 {
		workers = 1
	}

	ctx := context.Background()
	if err := db.Initialize(ctx); err != nil {
		return fmt.Errorf("failed to init db: %w", err)
	}
	defer db.Close()

	pool := db.GetPool()
	if pool == nil {
		return fmt.Errorf("db pool is nil")
	}

	overallStart := time.Now()

	// 1) Pull tournament shape from DB
	loadStart := time.Now()
	ids, err := resolveTournamentIDs(ctx, pool, season)
	if err != nil {
		return fmt.Errorf("failed to resolve tournament ids: %w", err)
	}

	ff, err := loadFinalFourConfig(ctx, pool, ids.coreTournamentID)
	if err != nil {
		return fmt.Errorf("failed to load final four config: %w", err)
	}

	teams, err := loadLabTeams(ctx, pool, ids.labTournamentID)
	if err != nil {
		return fmt.Errorf("failed to load teams: %w", err)
	}

	builder := bracket.NewBracketBuilder()
	br, err := builder.BuildBracket(ids.labTournamentID, teams, &models.FinalFourConfig{
		TopLeftRegion:     ff.topLeft,
		BottomLeftRegion:  ff.bottomLeft,
		TopRightRegion:    ff.topRight,
		BottomRightRegion: ff.bottomRight,
	})
	if err != nil {
		return fmt.Errorf("failed to build bracket: %w", err)
	}

	probs, err := loadPredictedGameOutcomes(ctx, pool, ids.labTournamentID)
	if err != nil {
		return fmt.Errorf("failed to load predicted game outcomes: %w", err)
	}

	loadDur := time.Since(loadStart)
	log.Printf("Loaded tournament season=%d lab_tournament_id=%s core_tournament_id=%s teams=%d probs=%d in %s",
		season,
		ids.labTournamentID,
		ids.coreTournamentID,
		len(teams),
		len(probs),
		loadDur,
	)

	// 2) Create snapshot + batch
	batchStart := time.Now()
	snapshotID, err := createTournamentStateSnapshot(ctx, pool, ids.coreTournamentID)
	if err != nil {
		return fmt.Errorf("failed to create tournament state snapshot: %w", err)
	}
	batchID, err := createTournamentSimulationBatch(ctx, pool, ids.coreTournamentID, snapshotID, nSims, seed, probabilitySourceKey)
	if err != nil {
		return fmt.Errorf("failed to create tournament simulation batch: %w", err)
	}
	log.Printf("Created snapshot_id=%s batch_id=%s in %s", snapshotID, batchID, time.Since(batchStart))

	// 3) Run simulations + write back
	simStart := time.Now()
	rowsWritten := int64(0)
	nBatches := 0

	for offset := 0; offset < nSims; offset += batchSize {
		n := batchSize
		if offset+n > nSims {
			n = nSims - offset
		}

		batchSeed := int64(seed) + int64(offset)*1_000_003
		results, err := tsim.Simulate(br, probs, n, batchSeed, tsim.Options{Workers: workers})
		if err != nil {
			return fmt.Errorf("simulation failed at offset=%d: %w", offset, err)
		}

		startWrite := time.Now()
		inserted, err := copyInsertSimulatedTournaments(ctx, pool, batchID, ids.labTournamentID, offset, results)
		if err != nil {
			return fmt.Errorf("failed writing results at offset=%d: %w", offset, err)
		}
		rowsWritten += inserted
		nBatches++

		log.Printf(
			"Batch %d offset=%d sims=%d inserted_rows=%d elapsed_sim=%s elapsed_write=%s",
			nBatches,
			offset,
			n,
			inserted,
			time.Since(simStart),
			time.Since(startWrite),
		)
	}

	simDur := time.Since(simStart)
	overallDur := time.Since(overallStart)

	log.Printf("DONE season=%d n_sims=%d workers=%d batch_size=%d rows=%d",
		season,
		nSims,
		workers,
		batchSize,
		rowsWritten,
	)
	log.Printf("Timing: load=%s simulate+write=%s overall=%s", loadDur, simDur, overallDur)
	return nil
}

func resolveTournamentIDs(ctx context.Context, pool *pgxpool.Pool, season int) (*tournamentIDs, error) {
	var labID string
	var coreID *string
	if err := pool.QueryRow(ctx, `
		SELECT t.id, t.core_tournament_id
		FROM derived.tournaments t
		WHERE t.season = $1::int
		  AND t.deleted_at IS NULL
		ORDER BY t.created_at DESC
		LIMIT 1
	`, season).Scan(&labID, &coreID); err != nil {
		return nil, err
	}
	if coreID == nil || *coreID == "" {
		return nil, fmt.Errorf("derived.tournaments.core_tournament_id is NULL for season=%d", season)
	}
	return &tournamentIDs{labTournamentID: labID, coreTournamentID: *coreID}, nil
}

func loadFinalFourConfig(ctx context.Context, pool *pgxpool.Pool, coreTournamentID string) (*finalFourConfig, error) {
	var tl, bl, tr, br *string
	if err := pool.QueryRow(ctx, `
		SELECT final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right
		FROM core.tournaments
		WHERE id = $1::uuid
		  AND deleted_at IS NULL
		LIMIT 1
	`, coreTournamentID).Scan(&tl, &bl, &tr, &br); err != nil {
		return nil, err
	}

	cfg := &finalFourConfig{}
	if tl != nil {
		cfg.topLeft = *tl
	}
	if bl != nil {
		cfg.bottomLeft = *bl
	}
	if tr != nil {
		cfg.topRight = *tr
	}
	if br != nil {
		cfg.bottomRight = *br
	}

	if cfg.topLeft == "" {
		cfg.topLeft = "East"
	}
	if cfg.bottomLeft == "" {
		cfg.bottomLeft = "West"
	}
	if cfg.topRight == "" {
		cfg.topRight = "South"
	}
	if cfg.bottomRight == "" {
		cfg.bottomRight = "Midwest"
	}

	return cfg, nil
}

func loadLabTeams(ctx context.Context, pool *pgxpool.Pool, labTournamentID string) ([]*models.TournamentTeam, error) {
	rows, err := pool.Query(ctx, `
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

	teams := make([]*models.TournamentTeam, 0)
	for rows.Next() {
		var id string
		var seed int
		var region string
		var schoolName string
		if err := rows.Scan(&id, &seed, &region, &schoolName); err != nil {
			return nil, err
		}
		teams = append(teams, &models.TournamentTeam{
			ID:     id,
			Seed:   seed,
			Region: region,
			School: &models.School{Name: schoolName},
		})
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	if len(teams) != 68 {
		return nil, fmt.Errorf("expected 68 teams, got %d", len(teams))
	}
	return teams, nil
}

func loadPredictedGameOutcomes(ctx context.Context, pool *pgxpool.Pool, labTournamentID string) (map[tsim.MatchupKey]float64, error) {
	rows, err := pool.Query(ctx, `
		SELECT game_id, team1_id, team2_id, p_team1_wins
		FROM derived.predicted_game_outcomes
		WHERE tournament_id = $1
		  AND deleted_at IS NULL
	`, labTournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[tsim.MatchupKey]float64)
	for rows.Next() {
		var gameID string
		var t1 string
		var t2 string
		var p float64
		if err := rows.Scan(&gameID, &t1, &t2, &p); err != nil {
			return nil, err
		}

		out[tsim.MatchupKey{GameID: gameID, Team1ID: t1, Team2ID: t2}] = p
		out[tsim.MatchupKey{GameID: gameID, Team1ID: t2, Team2ID: t1}] = 1.0 - p
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func createTournamentStateSnapshot(ctx context.Context, pool *pgxpool.Pool, coreTournamentID string) (string, error) {
	var snapshotID string
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := tx.QueryRow(ctx, `
		INSERT INTO derived.simulation_states (tournament_id, source, description)
		VALUES ($1, 'go_benchmark', 'Autogenerated snapshot for end-to-end simulation benchmark')
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

func createTournamentSimulationBatch(ctx context.Context, pool *pgxpool.Pool, coreTournamentID string, snapshotID string, nSims int, seed int, probabilitySourceKey string) (string, error) {
	var batchID string
	if err := pool.QueryRow(ctx, `
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

type resultsSource struct {
	batchID      string
	tournamentID string
	simOffset    int
	results      []tsim.TeamSimulationResult
	idx          int
	rowsInserted int64
	err          error
}

func (s *resultsSource) Next() bool {
	if s.err != nil {
		return false
	}
	if s.idx >= len(s.results) {
		return false
	}
	return true
}

func (s *resultsSource) Values() ([]any, error) {
	if s.idx >= len(s.results) {
		return nil, fmt.Errorf("out of bounds")
	}
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

func (s *resultsSource) Err() error {
	return s.err
}

func copyInsertSimulatedTournaments(
	ctx context.Context,
	pool *pgxpool.Pool,
	batchID string,
	tournamentID string,
	simOffset int,
	results []tsim.TeamSimulationResult,
) (int64, error) {
	conn, err := pool.Acquire(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Release()

	src := &resultsSource{
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
