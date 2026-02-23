//go:build integration

package workers

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
)

func TestThatSuccessfulImportCreatesPredictionBatch(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a pending tournament import with a valid 68-team ZIP bundle
	zipBytes := buildTestTournamentZIP(t)
	uploadID := insertPendingImport(t, ctx, pool, zipBytes)

	// WHEN processTournamentImport runs
	w := NewTournamentImportWorker(pool)
	w.processTournamentImport(ctx, uploadID)

	// THEN compute.prediction_batches has exactly 1 row for the tournament
	tournamentID := getTournamentIDByImportKey(t, ctx, pool, "ncaa-tournament-2026")
	count := countPredictionBatches(t, ctx, pool, tournamentID)
	if count != 1 {
		t.Errorf("expected 1 prediction batch, got %d", count)
	}
}

func TestThatSuccessfulImportCreates68PredictedTeamValues(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a pending tournament import with a valid 68-team ZIP bundle
	zipBytes := buildTestTournamentZIP(t)
	uploadID := insertPendingImport(t, ctx, pool, zipBytes)

	// WHEN processTournamentImport runs
	w := NewTournamentImportWorker(pool)
	w.processTournamentImport(ctx, uploadID)

	// THEN compute.predicted_team_values has 68 rows in the batch
	tournamentID := getTournamentIDByImportKey(t, ctx, pool, "ncaa-tournament-2026")
	batchID := getLatestBatchID(t, ctx, pool, tournamentID)
	count := countPredictedTeamValues(t, ctx, pool, batchID)
	if count != 68 {
		t.Errorf("expected 68 predicted team values, got %d", count)
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
