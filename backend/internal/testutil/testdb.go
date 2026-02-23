//go:build integration

package testutil

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testDBName = "calcutta_test"
	testDBUser = "calcutta"
	testDBPass = "calcutta"
)

// StartPostgresContainer spins up a Postgres 16 container, runs all
// migrations, and returns a *pgxpool.Pool connected to it.
// Call once from TestMain. The returned cleanup function terminates
// the container and must be called via defer.
func StartPostgresContainer(ctx context.Context) (*pgxpool.Pool, func(), error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase(testDBName),
		postgres.WithUsername(testDBUser),
		postgres.WithPassword(testDBPass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("starting postgres container: %w", err)
	}

	cleanup := func() {
		_ = pgContainer.Terminate(ctx)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("getting connection string: %w", err)
	}

	if err := runMigrations(connStr); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("running migrations: %w", err)
	}

	pool, err := pgxpool.New(ctx, connStr)
	if err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("creating pgxpool: %w", err)
	}

	return pool, func() {
		pool.Close()
		cleanup()
	}, nil
}

func runMigrations(databaseURL string) error {
	migrationsDir := migrationsPath()
	sourceURL := fmt.Sprintf("file://%s", migrationsDir)

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations up: %w", err)
	}
	return nil
}

// migrationsPath resolves the absolute path to the migrations directory
// relative to this source file, so tests work regardless of working directory.
func migrationsPath() string {
	_, thisFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "schema")
}

// TruncateAll truncates every table across core, derived, and lab schemas.
// Call from t.Cleanup() for per-test isolation.
func TruncateAll(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		TRUNCATE
			-- lab
			lab.pipeline_calcutta_runs,
			lab.pipeline_runs,
			lab.evaluation_entry_results,
			lab.evaluations,
			lab.entries,
			lab.investment_models,
			-- compute
			compute.simulated_teams,
			compute.simulated_tournaments,
			compute.tournament_snapshot_teams,
			compute.tournament_snapshots,
			compute.predicted_team_values,
			compute.prediction_batches,
			compute.predicted_game_outcomes,
			compute.game_outcome_runs,
			-- derived
			derived.run_jobs,
			-- core
			core.entry_teams,
			core.entries,
			core.payouts,
			core.calcutta_scoring_rules,
			core.calcutta_invitations,
			core.calcuttas,
			core.team_kenpom_stats,
			core.teams,
			core.tournament_imports,
			core.tournaments,
			core.schools,
			core.seasons,
			core.competitions,
			core.grants,
			core.role_permissions,
			core.roles,
			core.permissions,
			core.idempotency_keys,
			core.api_keys,
			core.auth_sessions,
			core.users
		CASCADE
	`)
	return err
}
