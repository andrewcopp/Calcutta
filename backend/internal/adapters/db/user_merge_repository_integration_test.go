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

func TestThatMergeMovesPortfoliosFromSourceToTarget(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a stub user (source) with a portfolio and an active user (target)
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	source := seedUser(t, ctx, userRepo, "John", "Doe", "stub")
	target := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	admin := seedUser(t, ctx, userRepo, "Admin", "User", "active")

	tournamentID := seedTournament(t, ctx)
	poolID := seedPool(t, ctx, tournamentID, source.ID)
	seedPortfolio(t, ctx, poolID, source.ID, "John's Portfolio")

	// WHEN merging source into target
	merge, err := mergeRepo.MergeUsers(ctx, source.ID, target.ID, admin.ID)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// THEN the merge moved 1 portfolio
	if merge.EntriesMoved != 1 {
		t.Errorf("expected 1 entry moved, got %d", merge.EntriesMoved)
	}
}

func TestThatMergeSkipsPortfolioWhenTargetAlreadyHasPortfolioInSamePool(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN both source and target have portfolios in the same pool
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	source := seedUser(t, ctx, userRepo, "John", "Doe", "stub")
	target := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	admin := seedUser(t, ctx, userRepo, "Admin", "User", "active")

	tournamentID := seedTournament(t, ctx)
	poolID := seedPool(t, ctx, tournamentID, source.ID)
	seedPortfolio(t, ctx, poolID, source.ID, "Source Portfolio")
	seedPortfolio(t, ctx, poolID, target.ID, "Target Portfolio")

	// WHEN merging source into target
	merge, err := mergeRepo.MergeUsers(ctx, source.ID, target.ID, admin.ID)
	if err != nil {
		t.Fatalf("merge failed: %v", err)
	}

	// THEN the merge moved 0 portfolios (target already has one)
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

	// GIVEN a stub user with a portfolio
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	source := seedUser(t, ctx, userRepo, "John", "Doe", "stub")
	target := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	admin := seedUser(t, ctx, userRepo, "Admin", "User", "active")

	tournamentID := seedTournament(t, ctx)
	poolID := seedPool(t, ctx, tournamentID, source.ID)
	seedPortfolio(t, ctx, poolID, source.ID, "Portfolio 1")

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

func TestThatMergeSucceedsWhenSourceIsActiveUser(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN an active user (not a stub) as source
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	source := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	target := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	admin := seedUser(t, ctx, userRepo, "Admin", "User", "active")

	// WHEN merging an active source into target
	_, err := mergeRepo.MergeUsers(ctx, source.ID, target.ID, admin.ID)

	// THEN it succeeds (no stub constraint)
	if err != nil {
		t.Fatalf("expected merge to succeed for active source, got: %v", err)
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

// ─── Batch Merge Tests ──────────────────────────────────────────────────────

func TestThatBatchMergeMovesPortfoliosFromAllSourcesToTarget(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN two source users each with a portfolio in different pools
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	source1 := seedUser(t, ctx, userRepo, "John", "Doe", "stub")
	source2 := seedUser(t, ctx, userRepo, "Johnny", "Doe", "stub")
	target := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	admin := seedUser(t, ctx, userRepo, "Admin", "User", "active")

	tournamentID := seedTournament(t, ctx)
	pool1 := seedPool(t, ctx, tournamentID, admin.ID)
	pool2 := seedPool(t, ctx, tournamentID, admin.ID)
	seedPortfolio(t, ctx, pool1, source1.ID, "Source1 Portfolio")
	seedPortfolio(t, ctx, pool2, source2.ID, "Source2 Portfolio")

	// WHEN batch merging both sources into target
	merges, err := mergeRepo.BatchMergeUsers(ctx, []string{source1.ID, source2.ID}, target.ID, admin.ID)
	if err != nil {
		t.Fatalf("batch merge failed: %v", err)
	}

	// THEN both merges moved 1 portfolio each
	totalMoved := 0
	for _, m := range merges {
		totalMoved += m.EntriesMoved
	}
	if totalMoved != 2 {
		t.Errorf("expected 2 total entries moved, got %d", totalMoved)
	}
}

func TestThatBatchMergeSoftDeletesAllSourceUsers(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN two source users
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	source1 := seedUser(t, ctx, userRepo, "John", "Doe", "stub")
	source2 := seedUser(t, ctx, userRepo, "Johnny", "Doe", "active")
	target := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	admin := seedUser(t, ctx, userRepo, "Admin", "User", "active")

	// WHEN batch merging
	_, err := mergeRepo.BatchMergeUsers(ctx, []string{source1.ID, source2.ID}, target.ID, admin.ID)
	if err != nil {
		t.Fatalf("batch merge failed: %v", err)
	}

	// THEN both source users are soft-deleted
	u1, _ := userRepo.GetByID(ctx, source1.ID)
	if u1 != nil {
		t.Error("expected source1 to be soft-deleted")
	}
	u2, _ := userRepo.GetByID(ctx, source2.ID)
	if u2 != nil {
		t.Error("expected source2 to be soft-deleted")
	}
}

func TestThatBatchMergeRecordsAuditTrailPerSourceUser(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN two source users
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	source1 := seedUser(t, ctx, userRepo, "John", "Doe", "stub")
	source2 := seedUser(t, ctx, userRepo, "Johnny", "Doe", "stub")
	target := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	admin := seedUser(t, ctx, userRepo, "Admin", "User", "active")

	// WHEN batch merging
	merges, err := mergeRepo.BatchMergeUsers(ctx, []string{source1.ID, source2.ID}, target.ID, admin.ID)
	if err != nil {
		t.Fatalf("batch merge failed: %v", err)
	}

	// THEN two audit records are returned
	if len(merges) != 2 {
		t.Errorf("expected 2 merge records, got %d", len(merges))
	}
}

func TestThatBatchMergeFailsWhenTargetIsInSourceList(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a user who appears in both source and target
	userRepo := db.NewUserRepository(pool)
	mergeRepo := db.NewUserMergeRepository(pool)

	user1 := seedUser(t, ctx, userRepo, "John", "Doe", "active")
	user2 := seedUser(t, ctx, userRepo, "Johnny", "Doe", "active")

	admin := seedUser(t, ctx, userRepo, "Admin", "User", "active")

	// WHEN batch merging with target in source list
	_, err := mergeRepo.BatchMergeUsers(ctx, []string{user1.ID, user2.ID}, user1.ID, admin.ID)

	// THEN it fails
	if err == nil {
		t.Fatal("expected error when target is in source list")
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

func seedPool(t *testing.T, ctx context.Context, tournamentID, ownerID string) string {
	t.Helper()
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO core.pools (tournament_id, owner_id, created_by, name)
		VALUES ($1::uuid, $2::uuid, $2::uuid, $3)
		RETURNING id
	`, tournamentID, ownerID, "Test Pool "+uuid.New().String()[:8]).Scan(&id)
	if err != nil {
		t.Fatalf("creating pool: %v", err)
	}
	return id
}

func seedPortfolio(t *testing.T, ctx context.Context, poolID, userID, name string) string {
	t.Helper()
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO core.portfolios (pool_id, user_id, name)
		VALUES ($1::uuid, $2::uuid, $3)
		RETURNING id
	`, poolID, userID, name).Scan(&id)
	if err != nil {
		t.Fatalf("creating portfolio: %v", err)
	}
	return id
}
