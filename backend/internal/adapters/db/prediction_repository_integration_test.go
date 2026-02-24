//go:build integration

package db_test

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"
	"time"

	db "github.com/andrewcopp/Calcutta/backend/internal/adapters/db"
	"github.com/andrewcopp/Calcutta/backend/internal/adapters/db/sqlc"
	"github.com/andrewcopp/Calcutta/backend/internal/app/prediction"
	"github.com/andrewcopp/Calcutta/backend/internal/models"
	"github.com/andrewcopp/Calcutta/backend/internal/testutil"
	"github.com/google/uuid"
)

// --- helpers ---

func seedTournamentForPredictions(t *testing.T, ctx context.Context) string {
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

func insertPredictionBatchRaw(t *testing.T, ctx context.Context, tournamentID string, throughRound int, createdAt time.Time) string {
	t.Helper()
	var id string
	err := pool.QueryRow(ctx, `
		INSERT INTO compute.prediction_batches (tournament_id, probability_source_key, game_outcome_spec_json, through_round, created_at)
		VALUES ($1::uuid, 'kenpom', '{}', $2, $3)
		RETURNING id::text
	`, tournamentID, throughRound, createdAt).Scan(&id)
	if err != nil {
		t.Fatalf("inserting prediction batch: %v", err)
	}
	return id
}

func insertPredictedTeamValueRaw(t *testing.T, ctx context.Context, batchID, tournamentID string) {
	t.Helper()
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

func countPredBatches(t *testing.T, ctx context.Context, tournamentID string) int {
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

func countPredBatchesForCheckpoint(t *testing.T, ctx context.Context, tournamentID string, throughRound int) int {
	t.Helper()
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM compute.prediction_batches
		WHERE tournament_id = $1::uuid AND through_round = $2 AND deleted_at IS NULL
	`, tournamentID, throughRound).Scan(&count)
	if err != nil {
		t.Fatalf("counting prediction batches for checkpoint: %v", err)
	}
	return count
}

func countPredTeamValues(t *testing.T, ctx context.Context, batchID string) int {
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

	// GIVEN a tournament with 4 prediction batches at throughRound=0
	tid := seedTournamentForPredictions(t, ctx)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	insertPredictionBatchRaw(t, ctx, tid, 0, base)
	insertPredictionBatchRaw(t, ctx, tid, 0, base.Add(1*time.Hour))
	insertPredictionBatchRaw(t, ctx, tid, 0, base.Add(2*time.Hour))
	insertPredictionBatchRaw(t, ctx, tid, 0, base.Add(3*time.Hour))

	// WHEN pruning with keepN=3
	repo := db.NewPredictionRepository(pool)
	_, _ = repo.PruneOldBatchesForCheckpoint(ctx, tid, 0, 3)

	// THEN 3 batches remain
	count := countPredBatches(t, ctx, tid)
	if count != 3 {
		t.Errorf("expected 3 prediction batches, got %d", count)
	}
}

func TestThatPruneCascadesDeleteToPredictedTeamValues(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with 2 batches at throughRound=0, oldest has a child row
	tid := seedTournamentForPredictions(t, ctx)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	oldBatchID := insertPredictionBatchRaw(t, ctx, tid, 0, base)
	insertPredictedTeamValueRaw(t, ctx, oldBatchID, tid)
	insertPredictionBatchRaw(t, ctx, tid, 0, base.Add(1*time.Hour))

	// WHEN pruning with keepN=1
	repo := db.NewPredictionRepository(pool)
	_, _ = repo.PruneOldBatchesForCheckpoint(ctx, tid, 0, 1)

	// THEN the child row of the old batch is also deleted
	count := countPredTeamValues(t, ctx, oldBatchID)
	if count != 0 {
		t.Errorf("expected 0 predicted team values for pruned batch, got %d", count)
	}
}

func TestThatPruneIsNoOpWhenFewerThanKeepNBatchesExist(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with only 2 batches at throughRound=0
	tid := seedTournamentForPredictions(t, ctx)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	insertPredictionBatchRaw(t, ctx, tid, 0, base)
	insertPredictionBatchRaw(t, ctx, tid, 0, base.Add(1*time.Hour))

	// WHEN pruning with keepN=3
	repo := db.NewPredictionRepository(pool)
	_, _ = repo.PruneOldBatchesForCheckpoint(ctx, tid, 0, 3)

	// THEN both batches remain
	count := countPredBatches(t, ctx, tid)
	if count != 2 {
		t.Errorf("expected 2 prediction batches, got %d", count)
	}
}

func TestThatPredictionPruneOnlyAffectsSpecifiedTournament(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN two tournaments, each with 3 batches at throughRound=0
	tidA := seedTournamentForPredictions(t, ctx)
	tidB := seedTournamentForPredictions(t, ctx)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		insertPredictionBatchRaw(t, ctx, tidA, 0, base.Add(time.Duration(i)*time.Hour))
		insertPredictionBatchRaw(t, ctx, tidB, 0, base.Add(time.Duration(i)*time.Hour))
	}

	// WHEN pruning tournament A with keepN=1
	repo := db.NewPredictionRepository(pool)
	_, _ = repo.PruneOldBatchesForCheckpoint(ctx, tidA, 0, 1)

	// THEN tournament B still has 3 batches
	count := countPredBatches(t, ctx, tidB)
	if count != 3 {
		t.Errorf("expected 3 prediction batches for tournament B, got %d", count)
	}
}

func TestThatPruneForCheckpointOnlyAffectsMatchingThroughRound(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with 3 batches at throughRound=0 and 2 batches at throughRound=3
	tid := seedTournamentForPredictions(t, ctx)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	insertPredictionBatchRaw(t, ctx, tid, 0, base)
	insertPredictionBatchRaw(t, ctx, tid, 0, base.Add(1*time.Hour))
	insertPredictionBatchRaw(t, ctx, tid, 0, base.Add(2*time.Hour))
	insertPredictionBatchRaw(t, ctx, tid, 3, base.Add(3*time.Hour))
	insertPredictionBatchRaw(t, ctx, tid, 3, base.Add(4*time.Hour))

	// WHEN pruning throughRound=0 with keepN=1
	repo := db.NewPredictionRepository(pool)
	_, _ = repo.PruneOldBatchesForCheckpoint(ctx, tid, 0, 1)

	// THEN throughRound=3 batches are unaffected
	count := countPredBatchesForCheckpoint(t, ctx, tid, 3)
	if count != 2 {
		t.Errorf("expected 2 prediction batches for throughRound=3, got %d", count)
	}
}

// --- new helpers ---

type predSeedResult struct {
	tournamentID string
	teamIDs      []string
}

// seedTournamentWithTeamsAndKenPom creates a tournament with 4 teams (one per region, seed 1)
// plus kenpom stats for each team.
func seedTournamentWithTeamsAndKenPom(t *testing.T, ctx context.Context) predSeedResult {
	t.Helper()

	base := mustSeedBase(t, ctx)

	regions := []string{"East", "West", "South", "Midwest"}
	netRtgs := []float64{25.0, 20.0, 15.0, 10.0}
	teamIDs := make([]string, 4)

	q := sqlc.New(pool)
	for i := 0; i < 4; i++ {
		school := &models.School{
			ID:   uuid.New().String(),
			Name: fmt.Sprintf("PredSchool %d %s", i+1, uuid.New().String()[:8]),
		}
		if err := base.schoolRepo.Create(ctx, school); err != nil {
			t.Fatalf("creating school %d: %v", i+1, err)
		}

		team := &models.TournamentTeam{
			ID:           uuid.New().String(),
			SchoolID:     school.ID,
			TournamentID: base.tournament.ID,
			Seed:         1,
			Region:       regions[i],
		}
		if err := base.tournamentRepo.CreateTeam(ctx, team); err != nil {
			t.Fatalf("creating team %d: %v", i+1, err)
		}
		teamIDs[i] = team.ID

		net := netRtgs[i]
		oRtg := net + 100.0
		dRtg := 100.0 - net
		adjT := 70.0
		if err := q.UpsertTeamKenPomStats(ctx, sqlc.UpsertTeamKenPomStatsParams{
			TeamID: team.ID,
			NetRtg: &net,
			ORtg:   &oRtg,
			DRtg:   &dRtg,
			AdjT:   &adjT,
		}); err != nil {
			t.Fatalf("upserting kenpom stats for team %d: %v", i+1, err)
		}
	}

	return predSeedResult{tournamentID: base.tournament.ID, teamIDs: teamIDs}
}

// seedCalcuttaWithScoringRules creates a calcutta for the given tournament
// with 6 standard scoring rules.
func seedCalcuttaWithScoringRules(t *testing.T, ctx context.Context, tournamentID string) {
	t.Helper()

	userRepo := db.NewUserRepository(pool)
	calcuttaRepo := db.NewCalcuttaRepository(pool)

	user := &models.User{
		ID:        uuid.New().String(),
		FirstName: "Pred",
		LastName:  "Owner",
		Status:    "active",
	}
	if err := userRepo.Create(ctx, user); err != nil {
		t.Fatalf("creating user for scoring rules: %v", err)
	}

	calcutta := &models.Calcutta{
		TournamentID: tournamentID,
		OwnerID:      user.ID,
		CreatedBy:    user.ID,
		Name:         "Pred Pool",
		BudgetPoints: 100,
		MinTeams:     3,
		MaxTeams:     10,
		MaxBidPoints: 50,
	}
	if err := calcuttaRepo.Create(ctx, calcutta); err != nil {
		t.Fatalf("creating calcutta for scoring rules: %v", err)
	}

	pointsPerRound := []int{0, 10, 20, 40, 80, 160}
	for i, pts := range pointsPerRound {
		round := &models.CalcuttaRound{
			CalcuttaID: calcutta.ID,
			Round:      i + 1,
			Points:     pts,
		}
		if err := calcuttaRepo.CreateRound(ctx, round); err != nil {
			t.Fatalf("creating scoring rule %d: %v", i+1, err)
		}
	}
}

func buildKnownTeamValues(teamIDs []string) []models.PredictedTeamValue {
	values := make([]models.PredictedTeamValue, len(teamIDs))
	for i, id := range teamIDs {
		f := float64(i + 1)
		values[i] = models.PredictedTeamValue{
			TeamID:               id,
			ExpectedPoints:       100.0 * f,
			VariancePoints:       50.0 * f,
			StdPoints:            7.0 * f,
			PRound1:              1.0,
			PRound2:              0.9 - 0.1*float64(i),
			PRound3:              0.7 - 0.1*float64(i),
			PRound4:              0.5 - 0.1*float64(i),
			PRound5:              0.3 - 0.05*float64(i),
			PRound6:              0.1 - 0.02*float64(i),
			PRound7:              0.05 - 0.01*float64(i),
			FavoritesTotalPoints: 500.0 + 10.0*f,
		}
	}
	return values
}

// --- new tests ---

func TestThatStorePredictionsAndGetTeamValuesRoundTripAllFields(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with 4 teams and known prediction values
	seed := seedTournamentWithTeamsAndKenPom(t, ctx)
	values := buildKnownTeamValues(seed.teamIDs)
	repo := db.NewPredictionRepository(pool)

	// WHEN storing predictions and retrieving them
	batchID, err := repo.StorePredictions(ctx, seed.tournamentID, "kenpom", []byte(`{"kind":"kenpom"}`), values, 0)
	if err != nil {
		t.Fatalf("storing predictions: %v", err)
	}
	got, err := repo.GetTeamValues(ctx, batchID)
	if err != nil {
		t.Fatalf("getting team values: %v", err)
	}

	// THEN all 4 teams are returned with correct field values
	sort.Slice(got, func(i, j int) bool { return got[i].TeamID < got[j].TeamID })
	sort.Slice(values, func(i, j int) bool { return values[i].TeamID < values[j].TeamID })

	if len(got) != 4 {
		t.Fatalf("expected 4 team values, got %d", len(got))
	}
	for i, want := range values {
		g := got[i]
		if g.TeamID != want.TeamID {
			t.Errorf("[%d] TeamID: want %s, got %s", i, want.TeamID, g.TeamID)
		}
		if math.Abs(g.ExpectedPoints-want.ExpectedPoints) > 0.001 {
			t.Errorf("[%d] ExpectedPoints: want %.2f, got %.2f", i, want.ExpectedPoints, g.ExpectedPoints)
		}
		if math.Abs(g.VariancePoints-want.VariancePoints) > 0.001 {
			t.Errorf("[%d] VariancePoints: want %.2f, got %.2f", i, want.VariancePoints, g.VariancePoints)
		}
		if math.Abs(g.StdPoints-want.StdPoints) > 0.001 {
			t.Errorf("[%d] StdPoints: want %.2f, got %.2f", i, want.StdPoints, g.StdPoints)
		}
		if math.Abs(g.PRound1-want.PRound1) > 0.001 {
			t.Errorf("[%d] PRound1: want %.3f, got %.3f", i, want.PRound1, g.PRound1)
		}
		if math.Abs(g.PRound2-want.PRound2) > 0.001 {
			t.Errorf("[%d] PRound2: want %.3f, got %.3f", i, want.PRound2, g.PRound2)
		}
		if math.Abs(g.PRound3-want.PRound3) > 0.001 {
			t.Errorf("[%d] PRound3: want %.3f, got %.3f", i, want.PRound3, g.PRound3)
		}
		if math.Abs(g.PRound4-want.PRound4) > 0.001 {
			t.Errorf("[%d] PRound4: want %.3f, got %.3f", i, want.PRound4, g.PRound4)
		}
		if math.Abs(g.PRound5-want.PRound5) > 0.001 {
			t.Errorf("[%d] PRound5: want %.3f, got %.3f", i, want.PRound5, g.PRound5)
		}
		if math.Abs(g.PRound6-want.PRound6) > 0.001 {
			t.Errorf("[%d] PRound6: want %.3f, got %.3f", i, want.PRound6, g.PRound6)
		}
		if math.Abs(g.PRound7-want.PRound7) > 0.001 {
			t.Errorf("[%d] PRound7: want %.3f, got %.3f", i, want.PRound7, g.PRound7)
		}
		if math.Abs(g.FavoritesTotalPoints-want.FavoritesTotalPoints) > 0.001 {
			t.Errorf("[%d] FavoritesTotalPoints: want %.2f, got %.2f", i, want.FavoritesTotalPoints, g.FavoritesTotalPoints)
		}
	}
}

func TestThatStorePredictionsCreatesRetrievableBatch(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with teams
	seed := seedTournamentWithTeamsAndKenPom(t, ctx)
	values := buildKnownTeamValues(seed.teamIDs)
	repo := db.NewPredictionRepository(pool)

	// WHEN storing predictions with throughRound=2
	batchID, err := repo.StorePredictions(ctx, seed.tournamentID, "kenpom-v2", []byte(`{}`), values, 2)
	if err != nil {
		t.Fatalf("storing predictions: %v", err)
	}

	// THEN GetLatestBatch returns a batch with matching source key and through round
	batch, found, err := repo.GetLatestBatch(ctx, seed.tournamentID)
	if err != nil {
		t.Fatalf("getting latest batch: %v", err)
	}
	if !found {
		t.Fatal("expected batch to be found")
	}
	if batch.ID != batchID {
		t.Errorf("expected batch ID %s, got %s", batchID, batch.ID)
	}
	if batch.ProbabilitySourceKey != "kenpom-v2" {
		t.Errorf("expected source key 'kenpom-v2', got '%s'", batch.ProbabilitySourceKey)
	}
	if batch.ThroughRound != 2 {
		t.Errorf("expected through round 2, got %d", batch.ThroughRound)
	}
}

func TestThatGetLatestBatchReturnsFalseForEmptyTournament(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with no prediction batches
	tid := seedTournamentForPredictions(t, ctx)
	repo := db.NewPredictionRepository(pool)

	// WHEN getting the latest batch
	batch, found, err := repo.GetLatestBatch(ctx, tid)

	// THEN found is false and no error occurs
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if found {
		t.Errorf("expected found=false, got true (batch=%+v)", batch)
	}
}

func TestThatListBatchesReturnsNewestFirst(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with 3 batches at different times
	tid := seedTournamentForPredictions(t, ctx)
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	id1 := insertPredictionBatchRaw(t, ctx, tid, 0, base)
	id2 := insertPredictionBatchRaw(t, ctx, tid, 0, base.Add(1*time.Hour))
	id3 := insertPredictionBatchRaw(t, ctx, tid, 0, base.Add(2*time.Hour))
	_ = id1

	repo := db.NewPredictionRepository(pool)

	// WHEN listing batches
	batches, err := repo.ListBatches(ctx, tid)
	if err != nil {
		t.Fatalf("listing batches: %v", err)
	}

	// THEN batches are in descending order (newest first)
	if len(batches) != 3 {
		t.Fatalf("expected 3 batches, got %d", len(batches))
	}
	if batches[0].ID != id3 {
		t.Errorf("expected newest batch first (id=%s), got %s", id3, batches[0].ID)
	}
	if batches[2].ID != id1 {
		t.Errorf("expected oldest batch last (id=%s), got %s", id1, batches[2].ID)
	}
	_ = id2
}

func TestThatLoadTeamsReturnsKenPomAndProgress(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with teams + kenpom stats, one team with 2 wins and 1 bye
	seed := seedTournamentWithTeamsAndKenPom(t, ctx)
	_, err := pool.Exec(ctx, `UPDATE core.teams SET wins = 2, byes = 1 WHERE id = $1::uuid`, seed.teamIDs[0])
	if err != nil {
		t.Fatalf("updating team wins: %v", err)
	}

	repo := db.NewPredictionRepository(pool)

	// WHEN loading teams
	teams, err := repo.LoadTeams(ctx, seed.tournamentID)
	if err != nil {
		t.Fatalf("loading teams: %v", err)
	}

	// THEN the updated team has correct wins, byes, and kenpom net rating
	if len(teams) != 4 {
		t.Fatalf("expected 4 teams, got %d", len(teams))
	}
	var found bool
	for _, tm := range teams {
		if tm.ID == seed.teamIDs[0] {
			found = true
			if tm.Wins != 2 {
				t.Errorf("expected wins=2, got %d", tm.Wins)
			}
			if tm.Byes != 1 {
				t.Errorf("expected byes=1, got %d", tm.Byes)
			}
			if math.Abs(tm.KenPomNet-25.0) > 0.001 {
				t.Errorf("expected KenPomNet=25.0, got %.2f", tm.KenPomNet)
			}
			if tm.Seed != 1 {
				t.Errorf("expected seed=1, got %d", tm.Seed)
			}
			if tm.Region != "East" {
				t.Errorf("expected region=East, got %s", tm.Region)
			}
		}
	}
	if !found {
		t.Error("expected to find the updated team in LoadTeams results")
	}
}

func TestThatLoadTeamsReturnsZeroKenPomWhenStatsAbsent(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with a team that has no kenpom stats
	base := mustSeedBase(t, ctx)
	school := &models.School{
		ID:   uuid.New().String(),
		Name: "NoKenPom School " + uuid.New().String()[:8],
	}
	if err := base.schoolRepo.Create(ctx, school); err != nil {
		t.Fatalf("creating school: %v", err)
	}
	team := &models.TournamentTeam{
		ID:           uuid.New().String(),
		SchoolID:     school.ID,
		TournamentID: base.tournament.ID,
		Seed:         16,
		Region:       "East",
	}
	if err := base.tournamentRepo.CreateTeam(ctx, team); err != nil {
		t.Fatalf("creating team: %v", err)
	}

	repo := db.NewPredictionRepository(pool)

	// WHEN loading teams
	teams, err := repo.LoadTeams(ctx, base.tournament.ID)
	if err != nil {
		t.Fatalf("loading teams: %v", err)
	}

	// THEN KenPomNet is 0.0 (COALESCE behavior)
	if len(teams) != 1 {
		t.Fatalf("expected 1 team, got %d", len(teams))
	}
	if teams[0].KenPomNet != 0.0 {
		t.Errorf("expected KenPomNet=0.0, got %.2f", teams[0].KenPomNet)
	}
}

func TestThatLoadScoringRulesReturnsRulesViaCalcuttaJoin(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with a calcutta and 6 scoring rules
	seed := seedTournamentWithTeamsAndKenPom(t, ctx)
	seedCalcuttaWithScoringRules(t, ctx, seed.tournamentID)

	repo := db.NewPredictionRepository(pool)

	// WHEN loading scoring rules
	rules, err := repo.LoadScoringRules(ctx, seed.tournamentID)
	if err != nil {
		t.Fatalf("loading scoring rules: %v", err)
	}

	// THEN 6 rules are returned with correct win indices and points
	if len(rules) != 6 {
		t.Fatalf("expected 6 scoring rules, got %d", len(rules))
	}
	expectedPoints := []int{0, 10, 20, 40, 80, 160}
	for i, r := range rules {
		if r.WinIndex != i+1 {
			t.Errorf("rule[%d] WinIndex: want %d, got %d", i, i+1, r.WinIndex)
		}
		if r.PointsAwarded != expectedPoints[i] {
			t.Errorf("rule[%d] PointsAwarded: want %d, got %d", i, expectedPoints[i], r.PointsAwarded)
		}
	}
}

func TestThatLoadScoringRulesReturnsEmptyWhenNoCalcuttaExists(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a tournament with no calcutta
	tid := seedTournamentForPredictions(t, ctx)
	repo := db.NewPredictionRepository(pool)

	// WHEN loading scoring rules
	rules, err := repo.LoadScoringRules(ctx, tid)

	// THEN an empty slice is returned with no error
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rules) != 0 {
		t.Errorf("expected 0 scoring rules, got %d", len(rules))
	}
}

func TestThatGetBatchSummaryReturnsErrorForNonexistentBatch(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN a random UUID that does not correspond to any batch
	repo := db.NewPredictionRepository(pool)

	// WHEN getting the batch summary
	_, err := repo.GetBatchSummary(ctx, uuid.New().String())

	// THEN an error containing "batch not found" is returned
	if err == nil {
		t.Fatal("expected error for nonexistent batch, got nil")
	}
	if !strings.Contains(err.Error(), "batch not found") {
		t.Errorf("expected error containing 'batch not found', got: %v", err)
	}
}

// --- end-to-end prediction run ---

// seedFullTournamentField creates a tournament with all 68 NCAA teams
// (4 regions * 16 seeds + 4 play-in teams at seed 16) with kenpom stats.
func seedFullTournamentField(t *testing.T, ctx context.Context) predSeedResult {
	t.Helper()

	base := mustSeedBase(t, ctx)
	q := sqlc.New(pool)

	regions := []string{"East", "West", "South", "Midwest"}
	var teamIDs []string

	for _, region := range regions {
		for seed := 1; seed <= 16; seed++ {
			school := &models.School{
				ID:   uuid.New().String(),
				Name: fmt.Sprintf("%s-%d %s", region, seed, uuid.New().String()[:6]),
			}
			if err := base.schoolRepo.Create(ctx, school); err != nil {
				t.Fatalf("creating school: %v", err)
			}
			team := &models.TournamentTeam{
				ID:           uuid.New().String(),
				SchoolID:     school.ID,
				TournamentID: base.tournament.ID,
				Seed:         seed,
				Region:       region,
			}
			if err := base.tournamentRepo.CreateTeam(ctx, team); err != nil {
				t.Fatalf("creating team: %v", err)
			}
			teamIDs = append(teamIDs, team.ID)

			net := 30.0 - float64(seed)*2.0
			oRtg := net + 100.0
			dRtg := 100.0 - net
			adjT := 70.0
			if err := q.UpsertTeamKenPomStats(ctx, sqlc.UpsertTeamKenPomStatsParams{
				TeamID: team.ID,
				NetRtg: &net,
				ORtg:   &oRtg,
				DRtg:   &dRtg,
				AdjT:   &adjT,
			}); err != nil {
				t.Fatalf("upserting kenpom: %v", err)
			}
		}
	}

	// Add 4 play-in teams (First Four: extra 16-seeds, one per region)
	for _, region := range regions {
		school := &models.School{
			ID:   uuid.New().String(),
			Name: fmt.Sprintf("%s-PlayIn %s", region, uuid.New().String()[:6]),
		}
		if err := base.schoolRepo.Create(ctx, school); err != nil {
			t.Fatalf("creating play-in school: %v", err)
		}
		team := &models.TournamentTeam{
			ID:           uuid.New().String(),
			SchoolID:     school.ID,
			TournamentID: base.tournament.ID,
			Seed:         16,
			Region:       region,
		}
		if err := base.tournamentRepo.CreateTeam(ctx, team); err != nil {
			t.Fatalf("creating play-in team: %v", err)
		}
		teamIDs = append(teamIDs, team.ID)

		net := -10.0
		oRtg := 90.0
		dRtg := 110.0
		adjT := 70.0
		if err := q.UpsertTeamKenPomStats(ctx, sqlc.UpsertTeamKenPomStatsParams{
			TeamID: team.ID,
			NetRtg: &net,
			ORtg:   &oRtg,
			DRtg:   &dRtg,
			AdjT:   &adjT,
		}); err != nil {
			t.Fatalf("upserting play-in kenpom: %v", err)
		}
	}

	return predSeedResult{tournamentID: base.tournament.ID, teamIDs: teamIDs}
}

func TestThatPredictionServiceRunCreatesTeamValuesForAllTeams(t *testing.T) {
	ctx := context.Background()
	t.Cleanup(func() { testutil.TruncateAll(ctx, pool) })

	// GIVEN 68 teams with kenpom stats and scoring rules
	seed := seedFullTournamentField(t, ctx)
	seedCalcuttaWithScoringRules(t, ctx, seed.tournamentID)

	repo := db.NewPredictionRepository(pool)
	svc := prediction.New(repo)

	// WHEN running predictions end-to-end
	result, err := svc.Run(ctx, prediction.RunParams{
		TournamentID:         seed.tournamentID,
		ProbabilitySourceKey: "kenpom",
	})
	if err != nil {
		t.Fatalf("running predictions: %v", err)
	}

	// THEN a batch is created with 68 team values, all with positive expected points
	if result.TeamCount != 68 {
		t.Errorf("expected TeamCount=68, got %d", result.TeamCount)
	}

	values, err := repo.GetTeamValues(ctx, result.BatchID)
	if err != nil {
		t.Fatalf("getting team values: %v", err)
	}
	if len(values) != 68 {
		t.Fatalf("expected 68 team values in DB, got %d", len(values))
	}
	for _, v := range values {
		if v.ExpectedPoints <= 0 {
			t.Errorf("team %s has non-positive expected points: %.2f", v.TeamID, v.ExpectedPoints)
		}
	}
}
