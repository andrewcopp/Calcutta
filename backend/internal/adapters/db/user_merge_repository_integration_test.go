//go:build integration

package db_test

import (
	"context"
	"testing"

	db "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
	"github.com/google/uuid"
)

func TestThatMergeMovesEntriesFromSourceToTarget(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a stub user (source) with an entry and an active user (target)
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	source := seedUser(t, ctx, userRepo, "John", "Doe", "stub")
	target := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	admin := seedUser(t, ctx, userRepo, "Admin", "User", "active")

	tournamentID := seedTournament(t, ctx)
	calcuttaID := seedCalcutta(t, ctx, tournamentID, source.ID)
	seedEntry(t, ctx, calcuttaID, source.ID, "John's Entry")

	// WHEN merging source into target
	merge, err := mergeRepo.MergeUsers(ctx, source.ID, target.ID, admin.ID)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// THEN the merge moved 1 entry
	if merge.EntriesMoved != 1 {
		t.Errorf("expected 1 entry moved, got %d", merge.EntriesMoved)
	}
}

func TestThatMergeSkipsEntryWhenTargetAlreadyHasEntryInSameCalcutta(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN both source and target have entries in the same calcutta
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	source := seedUser(t, ctx, userRepo, "John", "Doe", "stub")
	target := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	admin := seedUser(t, ctx, userRepo, "Admin", "User", "active")

	tournamentID := seedTournament(t, ctx)
	calcuttaID := seedCalcutta(t, ctx, tournamentID, source.ID)
	seedEntry(t, ctx, calcuttaID, source.ID, "Source Entry")
	seedEntry(t, ctx, calcuttaID, target.ID, "Target Entry")

	// WHEN merging source into target
	merge, err := mergeRepo.MergeUsers(ctx, source.ID, target.ID, admin.ID)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// THEN the merge moved 0 entries (target already has an entry)
	if merge.EntriesMoved != 0 {
		t.Errorf("expected 0 entries moved, got %d", merge.EntriesMoved)
	}
}

func TestThatMergeSoftDeletesSourceUser(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a stub user (source) and an active user (target)
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	source := seedUser(t, ctx, userRepo, "John", "Doe", "stub")
	target := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	admin := seedUser(t, ctx, userRepo, "Admin", "User", "active")

	// WHEN merging source into target
	_, err := mergeRepo.MergeUsers(ctx, source.ID, target.ID, admin.ID)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// THEN the source user is soft-deleted (GetByID returns nil for deleted users)
	deleted, err := userRepo.GetByID(ctx, source.ID)
	if err != nil {
		t.Fatalf("getting source user: %v", err)
	}
	if deleted != nil {
		t.Error("expected source user to be soft-deleted (GetByID should return nil)")
	}
}

func TestThatMergeRecordsAuditTrailWithCorrectCounts(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a stub user with an entry
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	source := seedUser(t, ctx, userRepo, "John", "Doe", "stub")
	target := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	admin := seedUser(t, ctx, userRepo, "Admin", "User", "active")

	tournamentID := seedTournament(t, ctx)
	calcuttaID := seedCalcutta(t, ctx, tournamentID, source.ID)
	seedEntry(t, ctx, calcuttaID, source.ID, "Entry 1")

	// WHEN merging
	merge, err := mergeRepo.MergeUsers(ctx, source.ID, target.ID, admin.ID)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// THEN the audit trail has correct data
	history, err := mergeRepo.ListMergeHistory(ctx, target.ID)
	if err != nil {
		t.Fatalf("list merge history: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 merge record, got %d", len(history))
	}
	if history[0].ID != merge.ID {
		t.Errorf("expected merge ID %s, got %s", merge.ID, history[0].ID)
	}
}

func TestThatMergeFailsWhenSourceIsNotStub(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN an active user (not a stub) as source
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	source := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	target := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	admin := seedUser(t, ctx, userRepo, "Admin", "User", "active")

	// WHEN attempting to merge
	_, err := mergeRepo.MergeUsers(ctx, source.ID, target.ID, admin.ID)

	// THEN it fails with an error
	if err == nil {
		t.Fatal("expected error when merging non-stub source")
	}
}

func TestThatFindMergeCandidatesReturnsSameNameUsers(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN three users: two with the same name, one different
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	user1 := seedUser(t, ctx, userRepo, "John", "Doe", "stub")
	seedUser(t, ctx, userRepo, "John", "Doe", "active")
	seedUser(t, ctx, userRepo, "Jane", "Smith", "stub")

	// WHEN finding merge candidates for user1
	candidates, err := mergeRepo.FindMergeCandidates(ctx, user1.ID)
	if err != nil {
		t.Fatalf("find merge candidates: %v", err)
	}

	// THEN only the other "John Doe" is returned
	if len(candidates) != 1 {
		t.Errorf("expected 1 merge candidate, got %d", len(candidates))
	}
}

// ─── Test Helpers ────────────────────────────────────────────────────────────

func seedUser(t *testing.T, ctx context.Context, repo *db.UserRepository, first, last, status string) *models.User {
	t.Helper()
	u := &models.User{
		ID:        uuid.New().String(),
		FirstName: first,
		LastName:  last,
		Status:    status,
	}
	if err := repo.Create(ctx, u); err != nil {
		t.Fatalf("creating user: %v", err)
	}
	return u
}

func seedTournament(t *testing.T, ctx context.Context) string {
	t.Helper()
	id := uuid.New().String()

	_, err := pool.Exec(ctx, `INSERT INTO core.competitions (name) VALUES ('NCAA Tournament') ON CONFLICT (name) DO NOTHING`)
	if err != nil {
		t.Fatalf("creating competition: %v", err)
	}
	_, err = pool.Exec(ctx, `INSERT INTO core.seasons (year) VALUES (2026) ON CONFLICT (year) DO NOTHING`)
	if err != nil {
		t.Fatalf("creating season: %v", err)
	}
	_, err = pool.Exec(ctx, `
		INSERT INTO core.tournaments (id, competition_id, season_id, import_key, rounds)
		SELECT $1::uuid, c.id, s.id, $2, 6
		FROM core.competitions c, core.seasons s
		WHERE c.name = 'NCAA Tournament' AND s.year = 2026
	`, id, "test-tournament-"+id)
	if err != nil {
		t.Fatalf("creating tournament: %v", err)
	}
	return id
}

func seedCalcutta(t *testing.T, ctx context.Context, tournamentID, ownerID string) string {
	t.Helper()
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO core.calcuttas (tournament_id, owner_id, created_by, name)
		VALUES ($1::uuid, $2::uuid, $2::uuid, $3)
		RETURNING id
	`, tournamentID, ownerID, "Test Pool "+uuid.New().String()[:8]).Scan(&id)
	if err != nil {
		t.Fatalf("creating calcutta: %v", err)
	}
	return id
}

func seedEntry(t *testing.T, ctx context.Context, calcuttaID, userID, name string) string {
	t.Helper()
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO core.entries (calcutta_id, user_id, name)
		VALUES ($1::uuid, $2::uuid, $3)
		RETURNING id
	`, calcuttaID, userID, name).Scan(&id)
	if err != nil {
		t.Fatalf("creating entry: %v", err)
	}
	return id
}
