//go:build integration

package prediction

import (
	"context"
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
)

// --- helpers ---

func seedTournament(t *testing.T, ctx context.Context) string {
	t.Helper()
	var compID string
	err := pool.QueryRow(ctx, `
		INSERT INTO core.competitions (name)
		VALUES ('NCAA-' || gen_random_uuid()::text)
		RETURNING id::text
	`).Scan(&compID)
	if err != nil {
		t.Fatalf("seeding competition: %v", err)
	}
	var seasonID string
	err = pool.QueryRow(ctx, `
		INSERT INTO core.seasons (year) VALUES (2026)
		ON CONFLICT (year) DO UPDATE SET year = EXCLUDED.year
		RETURNING id::text
	`).Scan(&seasonID)
	if err != nil {
		t.Fatalf("seeding season: %v", err)
	}
	var id string
	err = pool.QueryRow(ctx, `
		INSERT INTO core.tournaments (competition_id, season_id, import_key, rounds)
		VALUES ($1::uuid, $2::uuid, 'test-' || gen_random_uuid()::text, 7)
		RETURNING id::text
	`, compID, seasonID).Scan(&id)
	if err != nil {
		t.Fatalf("seeding tournament: %v", err)
	}
	return id
}

func insertPredictionBatch(t *testing.T, ctx context.Context, tournamentID string, createdAt time.Time) string {
	t.Helper()
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO compute.prediction_batches (tournament_id, probability_source_key, game_outcome_spec_json, created_at)
		VALUES ($1::uuid, 'kenpom', '{}', $2)
		RETURNING id::text
	`, tournamentID, createdAt).Scan(&id)
	if err != nil {
		t.Fatalf("inserting prediction batch: %v", err)
	}
	return id
}

func insertPredictedTeamValue(t *testing.T, ctx context.Context, batchID, tournamentID string) {
	t.Helper()
	// Create a school and team to satisfy FKs.
	var schoolID string
	err := pool.QueryRow(ctx, `
		INSERT INTO core.schools (name, slug) VALUES ('School-' || gen_random_uuid()::text, 'slug-' || gen_random_uuid()::text)
		RETURNING id::text
	`).Scan(&schoolID)
	if err != nil {
		t.Fatalf("inserting school: %v", err)
	}

	var teamID string
	err = pool.QueryRow(ctx, `
		INSERT INTO core.teams (tournament_id, school_id, seed, region)
		VALUES ($1::uuid, $2::uuid, 1, 'East')
		RETURNING id::text
	`, tournamentID, schoolID).Scan(&teamID)
	if err != nil {
		t.Fatalf("inserting team: %v", err)
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO compute.predicted_team_values (
			prediction_batch_id, tournament_id, team_id, expected_points
		)
		VALUES ($1::uuid, $2::uuid, $3::uuid, 10.0)
	`, batchID, tournamentID, teamID)
	if err != nil {
		t.Fatalf("inserting predicted team value: %v", err)
	}
}

func countPredictionBatches(t *testing.T, ctx context.Context, tournamentID string) int {
	t.Helper()
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM compute.prediction_batches
		WHERE tournament_id = $1::uuid AND deleted_at IS NULL
	`, tournamentID).Scan(&count)
	if err != nil {
		t.Fatalf("counting prediction batches: %v", err)
	}
	return count
}

func countPredictedTeamValues(t *testing.T, ctx context.Context, batchID string) int {
	t.Helper()
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM compute.predicted_team_values
		WHERE prediction_batch_id = $1::uuid AND deleted_at IS NULL
	`, batchID).Scan(&count)
	if err != nil {
		t.Fatalf("counting predicted team values: %v", err)
	}
	return count
}

// --- tests ---

func TestThatPruneKeepsLatestThreePredictionBatches(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with 4 prediction batches
	tid := seedTournament(t, ctx)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	insertPredictionBatch(t, ctx, tid, base)
	insertPredictionBatch(t, ctx, tid, base.Add(1*time.Hour))
	insertPredictionBatch(t, ctx, tid, base.Add(2*time.Hour))
	insertPredictionBatch(t, ctx, tid, base.Add(3*time.Hour))

	// WHEN pruning with keepN=3
	svc := New(pool)
	svc.pruneOldBatches(ctx, tid, 3)

	// THEN 3 batches remain
	count := countPredictionBatches(t, ctx, tid)
	if count != 3 {
		t.Errorf("expected 3 prediction batches, got %d", count)
	}
}

func TestThatPruneCascadesDeleteToPredictedTeamValues(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with 2 batches, oldest has a child row
	tid := seedTournament(t, ctx)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	oldBatchID := insertPredictionBatch(t, ctx, tid, base)
	insertPredictedTeamValue(t, ctx, oldBatchID, tid)
	insertPredictionBatch(t, ctx, tid, base.Add(1*time.Hour))

	// WHEN pruning with keepN=1
	svc := New(pool)
	svc.pruneOldBatches(ctx, tid, 1)

	// THEN the child row of the old batch is also deleted
	count := countPredictedTeamValues(t, ctx, oldBatchID)
	if count != 0 {
		t.Errorf("expected 0 predicted team values for pruned batch, got %d", count)
	}
}

func TestThatPruneIsNoOpWhenFewerThanKeepNBatchesExist(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with only 2 batches
	tid := seedTournament(t, ctx)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	insertPredictionBatch(t, ctx, tid, base)
	insertPredictionBatch(t, ctx, tid, base.Add(1*time.Hour))

	// WHEN pruning with keepN=3
	svc := New(pool)
	svc.pruneOldBatches(ctx, tid, 3)

	// THEN both batches remain
	count := countPredictionBatches(t, ctx, tid)
	if count != 2 {
		t.Errorf("expected 2 prediction batches, got %d", count)
	}
}

func TestThatPredictionPruneOnlyAffectsSpecifiedTournament(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN two tournaments, each with 3 batches
	tidA := seedTournament(t, ctx)
	tidB := seedTournament(t, ctx)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		insertPredictionBatch(t, ctx, tidA, base.Add(time.Duration(i)*time.Hour))
		insertPredictionBatch(t, ctx, tidB, base.Add(time.Duration(i)*time.Hour))
	}

	// WHEN pruning tournament A with keepN=1
	svc := New(pool)
	svc.pruneOldBatches(ctx, tidA, 1)

	// THEN tournament B still has 3 batches
	count := countPredictionBatches(t, ctx, tidB)
	if count != 3 {
		t.Errorf("expected 3 prediction batches for tournament B, got %d", count)
	}
}
