//go:build integration

package workers

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
)

func TestThatSuccessfulImportEnqueuesPredictionJob(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a pending tournament import with a valid 68-team ZIP bundle
	zipBytes := buildTestTournamentZIP(t)
	uploadID := insertPendingImport(t, ctx, pool, zipBytes)

	// WHEN processTournamentImport runs
	w := NewTournamentImportWorker(pool)
	w.processTournamentImport(ctx, uploadID)

	// THEN derived.run_jobs has a queued refresh_predictions job
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM derived.run_jobs
		WHERE run_kind = 'refresh_predictions' AND status = 'queued'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count prediction jobs: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 queued prediction job, got %d", count)
	}
}

func TestThatSuccessfulImportDeduplicatesPredictionJob(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a pending tournament import with a valid 68-team ZIP bundle
	zipBytes := buildTestTournamentZIP(t)
	uploadID := insertPendingImport(t, ctx, pool, zipBytes)

	// WHEN processTournamentImport runs twice for the same tournament
	w := NewTournamentImportWorker(pool)
	w.processTournamentImport(ctx, uploadID)

	// Re-insert so we can import again
	uploadID2 := insertPendingImport(t, ctx, pool, zipBytes)
	w.processTournamentImport(ctx, uploadID2)

	// THEN derived.run_jobs still has exactly 1 queued refresh_predictions job (deduplicated)
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM derived.run_jobs
		WHERE run_kind = 'refresh_predictions' AND status = 'queued'
	`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count prediction jobs: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 queued prediction job (deduplicated), got %d", count)
	}
}

func TestThatSuccessfulImportMarksStatusSucceeded(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a pending tournament import with a valid 68-team ZIP bundle
	zipBytes := buildTestTournamentZIP(t)
	uploadID := insertPendingImport(t, ctx, pool, zipBytes)

	// WHEN processTournamentImport runs
	w := NewTournamentImportWorker(pool)
	w.processTournamentImport(ctx, uploadID)

	// THEN the import status is 'succeeded'
	status := getImportStatus(t, ctx, pool, uploadID)
	if status != "succeeded" {
		t.Errorf("expected status 'succeeded', got %q", status)
	}
}

func TestThatFailedImportDoesNotCreatePredictions(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a pending tournament import with corrupt/invalid ZIP bytes
	corruptBytes := []byte("not a zip")
	uploadID := insertPendingImport(t, ctx, pool, corruptBytes)

	// WHEN processTournamentImport runs
	w := NewTournamentImportWorker(pool)
	w.processTournamentImport(ctx, uploadID)

	// THEN compute.prediction_batches has 0 rows
	var count int
	err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM compute.prediction_batches WHERE deleted_at IS NULL`).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count prediction batches: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 prediction batches, got %d", count)
	}
}

func TestThatSuccessfulImportDerivesIsEliminatedForCompletedTournament(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a pending tournament import with a completed tournament (champion has 6 wins)
	zipBytes := buildCompletedTournamentZIP(t)
	uploadID := insertPendingImport(t, ctx, pool, zipBytes)

	// WHEN processTournamentImport runs
	w := NewTournamentImportWorker(pool)
	w.processTournamentImport(ctx, uploadID)

	// THEN 67 teams are eliminated and 1 team (the champion) is alive
	tournamentID := getTournamentIDByImportKey(t, ctx, pool, "ncaa-tournament-2025")
	eliminated, alive := getTeamEliminationCounts(t, ctx, pool, tournamentID)
	if eliminated != 67 {
		t.Errorf("expected 67 eliminated teams, got %d", eliminated)
	}
	if alive != 1 {
		t.Errorf("expected 1 alive team (champion), got %d", alive)
	}
}

func TestThatFailedImportMarksStatusFailed(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a pending tournament import with corrupt/invalid ZIP bytes
	corruptBytes := []byte("not a zip")
	uploadID := insertPendingImport(t, ctx, pool, corruptBytes)

	// WHEN processTournamentImport runs
	w := NewTournamentImportWorker(pool)
	w.processTournamentImport(ctx, uploadID)

	// THEN the import status is 'failed'
	status := getImportStatus(t, ctx, pool, uploadID)
	if status != "failed" {
		t.Errorf("expected status 'failed', got %q", status)
	}
}
