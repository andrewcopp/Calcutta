package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

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

// RunSeedMigrations runs the seed data migrations
func RunSeedMigrations(ctx context.Context) error {
	// Seed schools from JSON file
	if err := seedSchoolsFromJSON(ctx); err != nil {
		return fmt.Errorf("error seeding schools from JSON: %v", err)
	}

	// Note: Calcutta data is seeded separately using the seed-calcuttas command
	fmt.Println("Note: Calcutta data should be seeded separately using the seed-calcuttas command")

	return nil
}

// seedSchoolsFromJSON seeds the schools table from the schools.json file
func seedSchoolsFromJSON(ctx context.Context) error {
	// Get the database connection
	pool := GetPool()
	if pool == nil {
		return fmt.Errorf("database connection pool is nil")
	}

	// Read the schools.json file
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory: %v", err)
	}

	jsonPath := filepath.Join(workDir, "migrations", "seed", "schools", "schools.json")
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("error reading schools.json file: %v", err)
	}

	// Parse the JSON data
	var schoolsData struct {
		Schools []string `json:"schools"`
	}
	if err := json.Unmarshal(jsonData, &schoolsData); err != nil {
		return fmt.Errorf("error parsing schools.json: %v", err)
	}

	// Insert schools into the database
	for _, schoolName := range schoolsData.Schools {
		// Check if the school already exists
		var count int
		err := pool.QueryRow(ctx, "SELECT COUNT(*) FROM schools WHERE name = $1", schoolName).Scan(&count)
		if err != nil {
			return fmt.Errorf("error checking if school exists: %v", err)
		}

		// Insert the school if it doesn't exist
		if count == 0 {
			_, err := pool.Exec(ctx, "INSERT INTO schools (name) VALUES ($1)", schoolName)
			if err != nil {
				return fmt.Errorf("error inserting school %s: %v", schoolName, err)
			}
		}
	}

	return nil
}

// getMigrator returns a migrate.Migrate instance for the specified migration type
func getMigrator(migrationType string) (*migrate.Migrate, error) {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable not set")
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
	m, err := migrate.New(sourceURL, dbURL)
	if err != nil {
		return nil, fmt.Errorf("error creating migrator: %v", err)
	}

	return m, nil
}
