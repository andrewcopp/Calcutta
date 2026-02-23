//go:build integration

package workers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles"
	"github.com/andrewcopp/Calcutta/backend/internal/bundles/archive"
	"github.com/jackc/pgx/v5/pgxpool"
)

// buildTestTournamentZIP creates a valid tournament import ZIP containing
// schools.json, a tournament JSON with 68 teams (including KenPom data),
// and a calcutta JSON with default scoring rules.
func buildTestTournamentZIP(t *testing.T) []byte {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "test-bundle-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	// Build 68 schools
	var schools []bundles.SchoolEntry
	for i := 1; i <= 68; i++ {
		schools = append(schools, bundles.SchoolEntry{
			Slug: fmt.Sprintf("school-%d", i),
			Name: fmt.Sprintf("School %d", i),
		})
	}
	schoolsBundle := bundles.SchoolsBundle{
		Version:     1,
		GeneratedAt: time.Now().UTC(),
		Schools:     schools,
	}
	if err := bundles.WriteJSON(filepath.Join(tmpDir, "schools.json"), schoolsBundle); err != nil {
		t.Fatalf("failed to write schools.json: %v", err)
	}

	// Build 68 teams: 4 regions x 16 seeds + 4 First Four extras
	regions := []string{"East", "Midwest", "South", "West"}
	var teams []bundles.TeamRecord
	schoolIdx := 1
	for _, region := range regions {
		for seed := 1; seed <= 16; seed++ {
			kenpomNet := 30.0 - float64(seed)*2.5
			teams = append(teams, bundles.TeamRecord{
				SchoolSlug: fmt.Sprintf("school-%d", schoolIdx),
				SchoolName: fmt.Sprintf("School %d", schoolIdx),
				Seed:       seed,
				Region:     region,
				KenPom: &bundles.KenPomRecord{
					NetRTG: kenpomNet,
					ORTG:   kenpomNet + 100,
					DRTG:   100 - kenpomNet,
					AdjT:   68.0,
				},
			})
			schoolIdx++
		}
	}

	// 4 First Four teams: 2 extra 16-seeds, 2 extra 11-seeds
	firstFourExtras := []struct {
		seed   int
		region string
		net    float64
	}{
		{16, "East", -12.0},
		{16, "West", -11.0},
		{11, "South", 3.0},
		{11, "Midwest", 2.5},
	}
	for _, ff := range firstFourExtras {
		teams = append(teams, bundles.TeamRecord{
			SchoolSlug: fmt.Sprintf("school-%d", schoolIdx),
			SchoolName: fmt.Sprintf("School %d", schoolIdx),
			Seed:       ff.seed,
			Region:     ff.region,
			KenPom: &bundles.KenPomRecord{
				NetRTG: ff.net,
				ORTG:   ff.net + 100,
				DRTG:   100 - ff.net,
				AdjT:   68.0,
			},
		})
		schoolIdx++
	}

	tournamentBundle := bundles.TournamentBundle{
		Version:     1,
		GeneratedAt: time.Now().UTC(),
		Tournament: bundles.TournamentRecord{
			ImportKey:            "ncaa-tournament-2026",
			Name:                 "NCAA Tournament 2026",
			Rounds:               6,
			FinalFourTopLeft:     "East",
			FinalFourBottomLeft:  "Midwest",
			FinalFourTopRight:    "South",
			FinalFourBottomRight: "West",
		},
		Teams: teams,
	}
	if err := bundles.WriteJSON(filepath.Join(tmpDir, "tournaments", "ncaa-tournament-2026.json"), tournamentBundle); err != nil {
		t.Fatalf("failed to write tournament json: %v", err)
	}

	// Build calcutta with default scoring rules (no entries/bids needed)
	ownerEmail := "commissioner@test.com"
	ownerFirst := "Test"
	ownerLast := "Commissioner"
	calcuttaBundle := bundles.CalcuttaBundle{
		Version:     1,
		GeneratedAt: time.Now().UTC(),
		Tournament: bundles.TournamentRef{
			ImportKey: "ncaa-tournament-2026",
			Name:      "NCAA Tournament 2026",
		},
		Calcutta: bundles.CalcuttaRecord{
			Key:  "test-pool",
			Name: "Test Pool",
			Owner: &bundles.UserRef{
				Email:     &ownerEmail,
				FirstName: &ownerFirst,
				LastName:  &ownerLast,
			},
		},
		Rounds: []bundles.RoundRecord{
			{Round: 1, Points: 10},
			{Round: 2, Points: 20},
			{Round: 3, Points: 40},
			{Round: 4, Points: 80},
			{Round: 5, Points: 160},
			{Round: 6, Points: 320},
		},
	}
	if err := bundles.WriteJSON(filepath.Join(tmpDir, "calcuttas", "ncaa-tournament-2026", "test-pool.json"), calcuttaBundle); err != nil {
		t.Fatalf("failed to write calcutta json: %v", err)
	}

	zipBytes, err := archive.ZipDir(tmpDir)
	if err != nil {
		t.Fatalf("failed to create zip: %v", err)
	}
	return zipBytes
}

// insertPendingImport inserts a tournament import in 'pending' status and returns its upload ID.
func insertPendingImport(t *testing.T, ctx context.Context, pool *pgxpool.Pool, zipBytes []byte) string {
	t.Helper()

	h := sha256.Sum256(zipBytes)
	sha := fmt.Sprintf("%x", h[:])

	q := sqlc.New(pool)
	uploadID, err := q.UpsertTournamentImport(ctx, sqlc.UpsertTournamentImportParams{
		Filename:  "test-bundle.zip",
		Sha256:    sha,
		SizeBytes: int64(len(zipBytes)),
		Archive:   zipBytes,
	})
	if err != nil {
		t.Fatalf("failed to insert tournament import: %v", err)
	}
	return uploadID
}

// getImportStatus returns the status string for a tournament import.
func getImportStatus(t *testing.T, ctx context.Context, pool *pgxpool.Pool, uploadID string) string {
	t.Helper()

	q := sqlc.New(pool)
	row, err := q.GetTournamentImportStatus(ctx, uploadID)
	if err != nil {
		t.Fatalf("failed to get import status: %v", err)
	}
	return row.Status
}

// countPredictionBatches counts prediction batches for a tournament.
func countPredictionBatches(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tournamentID string) int {
	t.Helper()

	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM compute.prediction_batches
		WHERE tournament_id = $1::uuid AND deleted_at IS NULL
	`, tournamentID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count prediction batches: %v", err)
	}
	return count
}

// countPredictedTeamValues counts predicted team values for a batch.
func countPredictedTeamValues(t *testing.T, ctx context.Context, pool *pgxpool.Pool, batchID string) int {
	t.Helper()

	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM compute.predicted_team_values
		WHERE prediction_batch_id = $1::uuid AND deleted_at IS NULL
	`, batchID).Scan(&count)
	if err != nil {
		t.Fatalf("failed to count predicted team values: %v", err)
	}
	return count
}

// getLatestBatchID returns the latest prediction batch ID for a tournament.
func getLatestBatchID(t *testing.T, ctx context.Context, pool *pgxpool.Pool, tournamentID string) string {
	t.Helper()

	var batchID string
	err := pool.QueryRow(ctx, `
		SELECT id::text
		FROM compute.prediction_batches
		WHERE tournament_id = $1::uuid AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, tournamentID).Scan(&batchID)
	if err != nil {
		t.Fatalf("failed to get latest batch ID: %v", err)
	}
	return batchID
}

// getTournamentIDByImportKey looks up the tournament ID by import key.
func getTournamentIDByImportKey(t *testing.T, ctx context.Context, pool *pgxpool.Pool, importKey string) string {
	t.Helper()

	var tournamentID string
	err := pool.QueryRow(ctx, `
		SELECT id::text
		FROM core.tournaments
		WHERE import_key = $1 AND deleted_at IS NULL
	`, importKey).Scan(&tournamentID)
	if err != nil {
		t.Fatalf("failed to get tournament ID for import key %s: %v", importKey, err)
	}
	return tournamentID
}
