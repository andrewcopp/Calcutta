//go:build integration

package simulation

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

func insertSnapshot(t *testing.T, ctx context.Context, tournamentID string) string {
	t.Helper()
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO compute.tournament_snapshots (tournament_id, source)
		VALUES ($1::uuid, 'test')
		RETURNING id::text
	`, tournamentID).Scan(&id)
	if err != nil {
		t.Fatalf("inserting snapshot: %v", err)
	}
	return id
}

func insertSnapshotTeam(t *testing.T, ctx context.Context, snapshotID, tournamentID string) {
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
		INSERT INTO compute.tournament_snapshot_teams (tournament_snapshot_id, team_id, wins, byes)
		VALUES ($1::uuid, $2::uuid, 0, 0)
	`, snapshotID, teamID)
	if err != nil {
		t.Fatalf("inserting snapshot team: %v", err)
	}
}

var simBatchSeq int

func insertSimBatch(t *testing.T, ctx context.Context, tournamentID, snapshotID string, createdAt time.Time) string {
	t.Helper()
	simBatchSeq++
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO compute.simulated_tournaments (tournament_id, tournament_snapshot_id, n_sims, seed, probability_source_key, created_at)
		VALUES ($1::uuid, $2::uuid, 100, $3, 'test', $4)
		RETURNING id::text
	`, tournamentID, snapshotID, simBatchSeq, createdAt).Scan(&id)
	if err != nil {
		t.Fatalf("inserting simulation batch: %v", err)
	}
	return id
}

func insertSimTeam(t *testing.T, ctx context.Context, simBatchID, tournamentID string) {
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
		INSERT INTO compute.simulated_teams (simulated_tournament_id, tournament_id, sim_id, team_id, wins, byes)
		VALUES ($1::uuid, $2::uuid, 0, $3::uuid, 3, 0)
	`, simBatchID, tournamentID, teamID)
	if err != nil {
		t.Fatalf("inserting simulated team: %v", err)
	}
}

func countSimBatches(t *testing.T, ctx context.Context, tournamentID string) int {
	t.Helper()
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM compute.simulated_tournaments
		WHERE tournament_id = $1::uuid AND deleted_at IS NULL
	`, tournamentID).Scan(&count)
	if err != nil {
		t.Fatalf("counting sim batches: %v", err)
	}
	return count
}

func countSimTeams(t *testing.T, ctx context.Context, simBatchID string) int {
	t.Helper()
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM compute.simulated_teams
		WHERE simulated_tournament_id = $1::uuid AND deleted_at IS NULL
	`, simBatchID).Scan(&count)
	if err != nil {
		t.Fatalf("counting sim teams: %v", err)
	}
	return count
}

func countSnapshots(t *testing.T, ctx context.Context, tournamentID string) int {
	t.Helper()
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM compute.tournament_snapshots
		WHERE tournament_id = $1::uuid
	`, tournamentID).Scan(&count)
	if err != nil {
		t.Fatalf("counting snapshots: %v", err)
	}
	return count
}

func countSnapshotTeams(t *testing.T, ctx context.Context, snapshotID string) int {
	t.Helper()
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM compute.tournament_snapshot_teams
		WHERE tournament_snapshot_id = $1::uuid
	`, snapshotID).Scan(&count)
	if err != nil {
		t.Fatalf("counting snapshot teams: %v", err)
	}
	return count
}

// --- tests ---

func TestThatSimulationPruneKeepsLatestOneBatch(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with 3 simulation batches
	tid := seedTournament(t, ctx)
	snapID := insertSnapshot(t, ctx, tid)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	insertSimBatch(t, ctx, tid, snapID, base)
	insertSimBatch(t, ctx, tid, snapID, base.Add(1*time.Hour))
	insertSimBatch(t, ctx, tid, snapID, base.Add(2*time.Hour))

	// WHEN pruning with keepN=1
	svc := New(pool)
	svc.pruneOldBatches(ctx, tid, 1)

	// THEN 1 batch remains
	count := countSimBatches(t, ctx, tid)
	if count != 1 {
		t.Errorf("expected 1 simulation batch, got %d", count)
	}
}

func TestThatSimulationPruneCascadesDeleteToSimulatedTeams(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with 2 batches, oldest has a child simulated_teams row
	tid := seedTournament(t, ctx)
	snapID := insertSnapshot(t, ctx, tid)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	oldBatchID := insertSimBatch(t, ctx, tid, snapID, base)
	insertSimTeam(t, ctx, oldBatchID, tid)
	insertSimBatch(t, ctx, tid, snapID, base.Add(1*time.Hour))

	// WHEN pruning with keepN=1
	svc := New(pool)
	svc.pruneOldBatches(ctx, tid, 1)

	// THEN child simulated_teams row of old batch is also deleted
	count := countSimTeams(t, ctx, oldBatchID)
	if count != 0 {
		t.Errorf("expected 0 simulated teams for pruned batch, got %d", count)
	}
}

func TestThatOrphanedSnapshotsAreCleanedUp(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with 2 batches on separate snapshots
	tid := seedTournament(t, ctx)
	orphanSnapID := insertSnapshot(t, ctx, tid)
	insertSnapshotTeam(t, ctx, orphanSnapID, tid)
	survivorSnapID := insertSnapshot(t, ctx, tid)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	insertSimBatch(t, ctx, tid, orphanSnapID, base)
	insertSimBatch(t, ctx, tid, survivorSnapID, base.Add(1*time.Hour))

	// WHEN pruning with keepN=1 (removes the older batch referencing orphanSnap)
	svc := New(pool)
	svc.pruneOldBatches(ctx, tid, 1)

	// THEN the orphaned snapshot and its child teams are deleted
	snapCount := countSnapshots(t, ctx, tid)
	if snapCount != 1 {
		t.Errorf("expected 1 snapshot, got %d", snapCount)
	}
	teamCount := countSnapshotTeams(t, ctx, orphanSnapID)
	if teamCount != 0 {
		t.Errorf("expected 0 snapshot teams for orphaned snapshot, got %d", teamCount)
	}
}

func TestThatSharedSnapshotSurvivesWhenOneBatchDeleted(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN two batches sharing the same snapshot
	tid := seedTournament(t, ctx)
	sharedSnapID := insertSnapshot(t, ctx, tid)
	insertSnapshotTeam(t, ctx, sharedSnapID, tid)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	insertSimBatch(t, ctx, tid, sharedSnapID, base)
	insertSimBatch(t, ctx, tid, sharedSnapID, base.Add(1*time.Hour))

	// WHEN pruning with keepN=1
	svc := New(pool)
	svc.pruneOldBatches(ctx, tid, 1)

	// THEN the shared snapshot survives because the remaining batch still references it
	snapCount := countSnapshots(t, ctx, tid)
	if snapCount != 1 {
		t.Errorf("expected 1 snapshot (shared, still referenced), got %d", snapCount)
	}
}

func TestThatSimulationPruneOnlyAffectsSpecifiedTournament(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN two tournaments, each with 3 batches
	tidA := seedTournament(t, ctx)
	tidB := seedTournament(t, ctx)
	snapA := insertSnapshot(t, ctx, tidA)
	snapB := insertSnapshot(t, ctx, tidB)

	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		insertSimBatch(t, ctx, tidA, snapA, base.Add(time.Duration(i)*time.Hour))
		insertSimBatch(t, ctx, tidB, snapB, base.Add(time.Duration(i)*time.Hour))
	}

	// WHEN pruning tournament A with keepN=1
	svc := New(pool)
	svc.pruneOldBatches(ctx, tidA, 1)

	// THEN tournament B still has 3 batches
	count := countSimBatches(t, ctx, tidB)
	if count != 3 {
		t.Errorf("expected 3 simulation batches for tournament B, got %d", count)
	}
}
