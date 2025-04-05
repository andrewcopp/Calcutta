package db

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations runs all pending migrations
func RunMigrations(ctx context.Context) error {
	// Get connection string from environment
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	// Get migrations directory
	migrationsDir := filepath.Join("internal", "db", "migrations")
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("error getting absolute path for migrations: %v", err)
	}

	// Create migration instance
	m, err := migrate.New(
		fmt.Sprintf("file://%s", absPath),
		connString,
	)
	if err != nil {
		return fmt.Errorf("error creating migration instance: %v", err)
	}
	defer m.Close()

	// Run migrations
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error running migrations: %v", err)
	}

	return nil
}

// RollbackMigrations rolls back the last migration
func RollbackMigrations(ctx context.Context) error {
	// Get connection string from environment
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	// Get migrations directory
	migrationsDir := filepath.Join("internal", "db", "migrations")
	absPath, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("error getting absolute path for migrations: %v", err)
	}

	// Create migration instance
	m, err := migrate.New(
		fmt.Sprintf("file://%s", absPath),
		connString,
	)
	if err != nil {
		return fmt.Errorf("error creating migration instance: %v", err)
	}
	defer m.Close()

	// Rollback migration
	if err := m.Steps(-1); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("error rolling back migration: %v", err)
	}

	return nil
}
