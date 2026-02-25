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

// seedTestSchoolsAndTournament inserts schools, a competition, a season, a tournament,
// and teams directly via SQL so the calcutta importer can resolve FKs.
func seedTestSchoolsAndTournament(t *testing.T, ctx context.Context, pool *pgxpool.Pool, importKey string, teamCount int) {
	t.Helper()

	// Ensure competition exists
	_, err := pool.Exec(ctx, `
		INSERT INTO core.competitions (name) VALUES ('NCAA Tournament')
		ON CONFLICT (name) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("failed to seed competition: %v", err)
	}

	// Ensure season exists
	_, err = pool.Exec(ctx, `
		INSERT INTO core.seasons (year) VALUES (2026)
		ON CONFLICT (year) DO NOTHING
	`)
	if err != nil {
		t.Fatalf("failed to seed season: %v", err)
	}

	// Insert tournament
	_, err = pool.Exec(ctx, `
		INSERT INTO core.tournaments (competition_id, season_id, import_key, rounds,
			final_four_top_left, final_four_bottom_left, final_four_top_right, final_four_bottom_right)
		SELECT c.id, s.id, $1, 6, 'East', 'Midwest', 'South', 'West'
		FROM core.competitions c, core.seasons s
		WHERE c.name = 'NCAA Tournament' AND s.year = 2026
		ON CONFLICT (import_key) WHERE deleted_at IS NULL DO NOTHING
	`, importKey)
	if err != nil {
		t.Fatalf("failed to seed tournament: %v", err)
	}

	// Insert schools + teams
	regions := []string{"East", "Midwest", "South", "West"}
	schoolIdx := 1
	for _, region := range regions {
		for seed := 1; seed <= 16; seed++ {
			if schoolIdx > teamCount {
				return
			}
			slug := fmt.Sprintf("school-%d", schoolIdx)
			name := fmt.Sprintf("School %d", schoolIdx)

			_, err := pool.Exec(ctx, `
				INSERT INTO core.schools (slug, name) VALUES ($1, $2)
				ON CONFLICT (slug) WHERE deleted_at IS NULL DO NOTHING
			`, slug, name)
			if err != nil {
				t.Fatalf("failed to seed school %s: %v", slug, err)
			}

			_, err = pool.Exec(ctx, `
				INSERT INTO core.teams (tournament_id, school_id, seed, region, byes, wins, is_eliminated)
				SELECT t.id, s.id, $3, $4, 1, 0, false
				FROM core.tournaments t
				JOIN core.schools s ON s.slug = $2 AND s.deleted_at IS NULL
				WHERE t.import_key = $1 AND t.deleted_at IS NULL
				ON CONFLICT (tournament_id, school_id) WHERE deleted_at IS NULL DO NOTHING
			`, importKey, slug, seed, region)
			if err != nil {
				t.Fatalf("failed to seed team %s: %v", slug, err)
			}
			schoolIdx++
		}
	}

	// 4 First Four extras (to reach 68)
	firstFourExtras := []struct {
		seed   int
		region string
	}{
		{16, "East"},
		{16, "West"},
		{11, "South"},
		{11, "Midwest"},
	}
	for _, ff := range firstFourExtras {
		if schoolIdx > teamCount {
			return
		}
		slug := fmt.Sprintf("school-%d", schoolIdx)
		name := fmt.Sprintf("School %d", schoolIdx)

		_, err := pool.Exec(ctx, `
			INSERT INTO core.schools (slug, name) VALUES ($1, $2)
			ON CONFLICT (slug) WHERE deleted_at IS NULL DO NOTHING
		`, slug, name)
		if err != nil {
			t.Fatalf("failed to seed school %s: %v", slug, err)
		}

		_, err = pool.Exec(ctx, `
			INSERT INTO core.teams (tournament_id, school_id, seed, region, byes, wins, is_eliminated)
			SELECT t.id, s.id, $3, $4, 0, 0, false
			FROM core.tournaments t
			JOIN core.schools s ON s.slug = $2 AND s.deleted_at IS NULL
			WHERE t.import_key = $1 AND t.deleted_at IS NULL
			ON CONFLICT (tournament_id, school_id) WHERE deleted_at IS NULL DO NOTHING
		`, importKey, slug, ff.seed, ff.region)
		if err != nil {
			t.Fatalf("failed to seed team %s: %v", slug, err)
		}
		schoolIdx++
	}
}

// buildTestCalcuttaZIP creates a valid calcutta import ZIP containing only
// a calcutta JSON with default scoring rules (no schools or tournaments).
func buildTestCalcuttaZIP(t *testing.T, tournamentImportKey string) []byte {
	t.Helper()

	tmpDir, err := os.MkdirTemp("", "test-bundle-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(tmpDir) })

	ownerFirst := "Test"
	ownerLast := "Commissioner"
	calcuttaBundle := bundles.CalcuttaBundle{
		Version:     1,
		GeneratedAt: time.Now().UTC(),
		Tournament: bundles.TournamentRef{
			ImportKey: tournamentImportKey,
			Name:      "NCAA Tournament 2026",
		},
		Calcutta: bundles.CalcuttaRecord{
			Key:  "test-pool",
			Name: "Test Pool",
			Owner: &bundles.UserRef{
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
	if err := bundles.WriteJSON(filepath.Join(tmpDir, "calcuttas", tournamentImportKey, "test-pool.json"), calcuttaBundle); err != nil {
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
