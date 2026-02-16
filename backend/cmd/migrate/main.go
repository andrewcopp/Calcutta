package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	platform.InitLogger()
	if err := run(); err != nil {
		slog.Error("cmd_failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse command line flags
	up := flag.Bool("up", false, "Run migrations up")
	down := flag.Bool("down", false, "Run migrations down")
	force := flag.Int("force", 0, "Force schema migration version (clears dirty state)")
	flag.Parse()

	// Check if at least one flag is set
	if !*up && !*down && *force == 0 {
		fmt.Println("Please specify either -up or -down flag")
		flag.Usage()
		return fmt.Errorf("no migration action specified")
	}

	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	m, err := newMigrator(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}
	defer m.Close()

	// Run migrations
	if *force != 0 {
		fmt.Printf("Forcing schema migration version to %d (clearing dirty state)...\n", *force)
		if err := m.Force(*force); err != nil {
			return fmt.Errorf("error forcing schema migrations: %w", err)
		}
		fmt.Println("Schema migration version forced successfully")
	}

	if *up {
		fmt.Println("Running schema migrations up...")
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("error running schema migrations: %w", err)
		}
		fmt.Println("Schema migrations completed successfully")
	}

	if *down {
		fmt.Println("Rolling back schema migrations...")
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("error rolling back schema migrations: %w", err)
		}
		fmt.Println("Schema migrations rolled back successfully")
	}

	return nil
}

func newMigrator(databaseURL string) (*migrate.Migrate, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting working directory: %w", err)
	}

	migrationsDir := filepath.Join(workDir, "migrations", "schema")
	sourceURL := fmt.Sprintf("file://%s", migrationsDir)

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("error creating migrator: %w", err)
	}

	return m, nil
}
