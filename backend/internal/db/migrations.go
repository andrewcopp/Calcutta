package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunSchemaMigrations runs the schema migrations
func RunSchemaMigrations(ctx context.Context) error {
	m, err := getMigrator("schema")
	if err != nil {
		return fmt.Errorf("error getting schema migrator: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error running schema migrations: %v", err)
	}
	return nil
}

// ForceSchemaMigrations sets the schema migration version and clears any dirty state.
// This should typically only be used as a recovery tool for development databases.
func ForceSchemaMigrations(ctx context.Context, version int) error {
	m, err := getMigrator("schema")
	if err != nil {
		return fmt.Errorf("error getting schema migrator: %v", err)
	}
	defer m.Close()

	if err := m.Force(version); err != nil {
		return fmt.Errorf("error forcing schema migrations: %v", err)
	}
	return nil
}

// RollbackSchemaMigrations rolls back the schema migrations
func RollbackSchemaMigrations(ctx context.Context) error {
	m, err := getMigrator("schema")
	if err != nil {
		return fmt.Errorf("error getting schema migrator: %v", err)
	}
	defer m.Close()

	if err := m.Down(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error rolling back schema migrations: %v", err)
	}
	return nil
}

// getMigrator returns a migrate.Migrate instance for the specified migration type
func getMigrator(migrationType string) (*migrate.Migrate, error) {
	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return nil, err
	}

	// Get migrations directory
	// Use absolute path to migrations directory
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting working directory: %v", err)
	}

	migrationsDir := filepath.Join(workDir, "migrations", migrationType)
	sourceURL := fmt.Sprintf("file://%s", migrationsDir)

	// Create migrator
	m, err := migrate.New(sourceURL, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("error creating migrator: %v", err)
	}

	return m, nil
}
