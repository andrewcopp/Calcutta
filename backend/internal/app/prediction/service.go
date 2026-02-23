package prediction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/app/scoring"
	"github.com/andrewcopp/Calcutta/backend/internal/app/simulation_game_outcomes"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Service handles prediction generation and storage.
type Service struct {
	pool *pgxpool.Pool
}

// New creates a new prediction service.
func New(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool}
}

// RunParams configures a prediction run.
type RunParams struct {
	TournamentID         string
	ProbabilitySourceKey string                        // e.g., "kenpom"
	GameOutcomeSpec      *simulation_game_outcomes.Spec // KenPom parameters
}

// RunResult holds the output of a prediction run.
type RunResult struct {
	BatchID              string
	TournamentID         string
	ProbabilitySourceKey string
	TeamCount            int
	Duration             time.Duration
}

// Run generates predictions for a tournament and stores them in the database.
func (s *Service) Run(ctx context.Context, p RunParams) (*RunResult, error) {
	start := time.Now()

	if p.TournamentID == "" {
		return nil, errors.New("TournamentID is required")
	}
	if p.ProbabilitySourceKey == "" {
		p.ProbabilitySourceKey = "kenpom"
	}
	if p.GameOutcomeSpec == nil {
		p.GameOutcomeSpec = &simulation_game_outcomes.Spec{Kind: "kenpom", Sigma: 10.0}
	}
	p.GameOutcomeSpec.Normalize()

	// Load teams from database
	teams, err := s.loadTeams(ctx, p.TournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load teams: %w", err)
	}
	if len(teams) != 68 {
		return nil, fmt.Errorf("expected 68 teams, got %d", len(teams))
	}

	// Load scoring rules from database (use default calcutta for tournament)
	rules, err := s.loadScoringRules(ctx, p.TournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to load scoring rules: %w", err)
	}
	if len(rules) == 0 {
		rules = DefaultScoringRules()
	}

	// Generate all theoretical matchups
	matchups, err := GenerateAllTheoreticalMatchups(teams, p.GameOutcomeSpec)
	if err != nil {
		return nil, fmt.Errorf("failed to generate matchups: %w", err)
	}

	// Calculate tournament values
	teamValues := GenerateTournamentValues(matchups, rules)

	// Store results
	specJSON, _ := json.Marshal(p.GameOutcomeSpec)
	batchID, err := s.storePredictions(ctx, p.TournamentID, p.ProbabilitySourceKey, specJSON, teamValues)
	if err != nil {
		return nil, fmt.Errorf("failed to store predictions: %w", err)
	}

	s.pruneOldBatches(ctx, p.TournamentID, 5)

	return &RunResult{
		BatchID:              batchID,
		TournamentID:         p.TournamentID,
		ProbabilitySourceKey: p.ProbabilitySourceKey,
		TeamCount:            len(teamValues),
		Duration:             time.Since(start),
	}, nil
}

// pruneOldBatches deletes prediction batches older than the latest keepN
// for a tournament. CASCADE on the FK auto-deletes predicted_team_values.
// Best-effort: logs and continues on error.
func (s *Service) pruneOldBatches(ctx context.Context, tournamentID string, keepN int) {
	result, err := s.pool.Exec(ctx, `
		DELETE FROM compute.prediction_batches
		WHERE tournament_id = $1::uuid
			AND deleted_at IS NULL
			AND id NOT IN (
				SELECT id FROM compute.prediction_batches
				WHERE tournament_id = $1::uuid
					AND deleted_at IS NULL
				ORDER BY created_at DESC
				LIMIT $2
			)
	`, tournamentID, keepN)
	if err != nil {
		slog.Warn("prediction_prune_failed", "tournament_id", tournamentID, "error", err)
		return
	}
	if n := result.RowsAffected(); n > 0 {
		slog.Info("prediction_prune_succeeded", "tournament_id", tournamentID, "batches_deleted", n)
	}
}

// GetLatestBatchID returns the most recent prediction batch for a tournament.
func (s *Service) GetLatestBatchID(ctx context.Context, tournamentID string) (string, bool, error) {
	var batchID string
	err := s.pool.QueryRow(ctx, `
		SELECT id::text
		FROM compute.prediction_batches
		WHERE tournament_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, tournamentID).Scan(&batchID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", false, nil
		}
		return "", false, fmt.Errorf("getting latest prediction batch: %w", err)
	}
	return batchID, true, nil
}

// GetTeamValues returns predicted team values for a batch.
func (s *Service) GetTeamValues(ctx context.Context, batchID string) ([]PredictedTeamValue, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			team_id::text,
			expected_points,
			COALESCE(variance_points, 0),
			COALESCE(std_points, 0),
			COALESCE(p_round_1, 0),
			COALESCE(p_round_2, 0),
			COALESCE(p_round_3, 0),
			COALESCE(p_round_4, 0),
			COALESCE(p_round_5, 0),
			COALESCE(p_round_6, 0),
			COALESCE(p_round_7, 0)
		FROM compute.predicted_team_values
		WHERE prediction_batch_id = $1::uuid
			AND deleted_at IS NULL
	`, batchID)
	if err != nil {
		return nil, fmt.Errorf("querying team values: %w", err)
	}
	defer rows.Close()

	var results []PredictedTeamValue
	for rows.Next() {
		var v PredictedTeamValue
		if err := rows.Scan(
			&v.TeamID,
			&v.ExpectedPoints,
			&v.VariancePoints,
			&v.StdPoints,
			&v.PRound1,
			&v.PRound2,
			&v.PRound3,
			&v.PRound4,
			&v.PRound5,
			&v.PRound6,
			&v.PRound7,
		); err != nil {
			return nil, fmt.Errorf("scanning team value: %w", err)
		}
		results = append(results, v)
	}
	return results, rows.Err()
}

// GetExpectedPointsMap returns a map of team_id -> expected_points for a tournament.
// Uses the latest prediction batch.
func (s *Service) GetExpectedPointsMap(ctx context.Context, tournamentID string) (map[string]float64, error) {
	batchID, found, err := s.GetLatestBatchID(ctx, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("getting latest batch id: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("no prediction batch found for tournament %s", tournamentID)
	}

	values, err := s.GetTeamValues(ctx, batchID)
	if err != nil {
		return nil, fmt.Errorf("getting team values: %w", err)
	}

	result := make(map[string]float64, len(values))
	for _, v := range values {
		result[v.TeamID] = v.ExpectedPoints
	}
	return result, nil
}

func (s *Service) loadTeams(ctx context.Context, tournamentID string) ([]TeamInput, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			t.id::text,
			t.seed,
			t.region,
			COALESCE(ks.net_rtg, 0) AS kenpom_net
		FROM core.teams t
		LEFT JOIN core.team_kenpom_stats ks
			ON ks.team_id = t.id
			AND ks.deleted_at IS NULL
		WHERE t.tournament_id = $1::uuid
			AND t.deleted_at IS NULL
		ORDER BY t.region, t.seed
	`, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("querying teams: %w", err)
	}
	defer rows.Close()

	var teams []TeamInput
	for rows.Next() {
		var t TeamInput
		if err := rows.Scan(&t.ID, &t.Seed, &t.Region, &t.KenPomNet); err != nil {
			return nil, fmt.Errorf("scanning team: %w", err)
		}
		teams = append(teams, t)
	}
	return teams, rows.Err()
}

func (s *Service) loadScoringRules(ctx context.Context, tournamentID string) ([]scoring.Rule, error) {
	// Get the first calcutta for this tournament to use its scoring rules
	rows, err := s.pool.Query(ctx, `
		SELECT csr.win_index::int, csr.points_awarded::int
		FROM core.calcutta_scoring_rules csr
		JOIN core.calcuttas c ON c.id = csr.calcutta_id AND c.deleted_at IS NULL
		WHERE c.tournament_id = $1::uuid
			AND csr.deleted_at IS NULL
		ORDER BY csr.win_index ASC
		LIMIT 10
	`, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("querying scoring rules: %w", err)
	}
	defer rows.Close()

	var rules []scoring.Rule
	for rows.Next() {
		var r scoring.Rule
		if err := rows.Scan(&r.WinIndex, &r.PointsAwarded); err != nil {
			return nil, fmt.Errorf("scanning scoring rule: %w", err)
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

func (s *Service) storePredictions(
	ctx context.Context,
	tournamentID string,
	probabilitySourceKey string,
	specJSON []byte,
	values []PredictedTeamValue,
) (string, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("beginning transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create batch record
	var batchID string
	err = tx.QueryRow(ctx, `
		INSERT INTO compute.prediction_batches (
			tournament_id,
			probability_source_key,
			game_outcome_spec_json
		)
		VALUES ($1::uuid, $2, $3::jsonb)
		RETURNING id::text
	`, tournamentID, probabilitySourceKey, specJSON).Scan(&batchID)
	if err != nil {
		return "", fmt.Errorf("failed to create batch: %w", err)
	}

	// Insert team values
	for _, v := range values {
		_, err = tx.Exec(ctx, `
			INSERT INTO compute.predicted_team_values (
				prediction_batch_id,
				tournament_id,
				team_id,
				expected_points,
				variance_points,
				std_points,
				p_round_1,
				p_round_2,
				p_round_3,
				p_round_4,
				p_round_5,
				p_round_6,
				p_round_7
			)
			VALUES ($1::uuid, $2::uuid, $3::uuid, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`,
			batchID,
			tournamentID,
			v.TeamID,
			v.ExpectedPoints,
			v.VariancePoints,
			v.StdPoints,
			v.PRound1,
			v.PRound2,
			v.PRound3,
			v.PRound4,
			v.PRound5,
			v.PRound6,
			v.PRound7,
		)
		if err != nil {
			return "", fmt.Errorf("failed to insert team value for %s: %w", v.TeamID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("committing predictions: %w", err)
	}

	return batchID, nil
}

// BackfillMissing generates predictions for any tournament that has 68 teams
// with KenPom data and scoring rules but no prediction batch. Returns the
// number of tournaments backfilled.
func (s *Service) BackfillMissing(ctx context.Context) int {
	rows, err := s.pool.Query(ctx, `
		SELECT t.id::text
		FROM core.tournaments t
		WHERE t.deleted_at IS NULL
			AND NOT EXISTS (
				SELECT 1 FROM compute.prediction_batches pb
				WHERE pb.tournament_id = t.id AND pb.deleted_at IS NULL
			)
			AND (
				SELECT COUNT(*) FROM core.teams tt
				WHERE tt.tournament_id = t.id AND tt.deleted_at IS NULL
			) = 68
			AND EXISTS (
				SELECT 1 FROM core.team_kenpom_stats ks
				JOIN core.teams tt ON tt.id = ks.team_id AND tt.deleted_at IS NULL
				WHERE tt.tournament_id = t.id AND ks.deleted_at IS NULL
			)
			AND EXISTS (
				SELECT 1 FROM core.calcutta_scoring_rules csr
				JOIN core.calcuttas c ON c.id = csr.calcutta_id AND c.deleted_at IS NULL
				WHERE c.tournament_id = t.id AND csr.deleted_at IS NULL
			)
	`)
	if err != nil {
		slog.Warn("prediction_backfill_query_failed", "error", err)
		return 0
	}
	defer rows.Close()

	var tournamentIDs []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			slog.Warn("prediction_backfill_scan_failed", "error", err)
			return 0
		}
		tournamentIDs = append(tournamentIDs, id)
	}
	if err := rows.Err(); err != nil {
		slog.Warn("prediction_backfill_rows_failed", "error", err)
		return 0
	}

	if len(tournamentIDs) == 0 {
		return 0
	}

	count := 0
	for _, tid := range tournamentIDs {
		result, err := s.Run(ctx, RunParams{
			TournamentID:         tid,
			ProbabilitySourceKey: "kenpom",
		})
		if err != nil {
			slog.Warn("prediction_backfill_failed", "tournament_id", tid, "error", err)
			continue
		}
		slog.Info("prediction_backfill_succeeded",
			"tournament_id", tid, "batch_id", result.BatchID,
			"team_count", result.TeamCount, "duration_ms", result.Duration.Milliseconds())
		count++
	}
	return count
}
