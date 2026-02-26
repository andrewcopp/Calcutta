//go:build integration

package workers

import (
	"context"
	"testing"

	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
)

func TestThatSuccessfulImportMarksStatusSucceeded(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a pending tournament import with a valid pool ZIP bundle
	seedTestSchoolsAndTournament(t, ctx, pool, "ncaa-tournament-2026", 68)
	zipBytes := buildTestPoolZIP(t, "ncaa-tournament-2026")
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
