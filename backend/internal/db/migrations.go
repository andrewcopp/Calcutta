package db

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func stripPsqlMetaCommands(sql string) string {
	lines := strings.Split(sql, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "\\") {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

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
// Prefers SQL dumps if available, falls back to JSON/CSV-based seeding
func RunSeedMigrations(ctx context.Context) error {
	// Try SQL dump-based seeding first (more robust)
	if err := seedFromSQLDumps(ctx); err == nil {
		fmt.Println("✅ Seed data loaded from SQL dumps")
		return nil
	} else {
		fmt.Printf("SQL dumps not available or failed (%v), falling back to JSON-based seeding\n", err)
	}

	// Fallback to JSON-based seeding
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

// seedFromSQLDumps loads seed data from SQL dump files
func seedFromSQLDumps(ctx context.Context) error {
	// Get the database connection
	pool := GetPool()
	if pool == nil {
		return fmt.Errorf("database connection pool is nil")
	}

	// Get the SQL dumps directory
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("error getting working directory: %v", err)
	}

	dumpsDir := filepath.Join(workDir, "migrations", "seed", "sql-dumps")

	// Check if dumps directory exists
	if _, err := os.Stat(dumpsDir); os.IsNotExist(err) {
		return fmt.Errorf("sql-dumps directory does not exist")
	}

	// Define the order of SQL files to execute (respecting foreign key dependencies)
	sqlFiles := []string{
		"schools.sql",
		"tournaments.sql",
		"tournament_teams.sql",
		"tournament_team_kenpom_stats.sql",
		"users.sql",
		"calcuttas.sql",
		"calcutta_payouts.sql",
		"calcutta_rounds.sql",
		"calcutta_entries.sql",
		"calcutta_entry_teams.sql",
	}

	// Check if seed data already exists (idempotency check)
	// We check multiple tables to ensure complete seed data is present
	var schoolCount, tournamentCount, calcuttaCount int

	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM schools").Scan(&schoolCount)
	if err != nil {
		return fmt.Errorf("error checking schools: %v", err)
	}

	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM tournaments").Scan(&tournamentCount)
	if err != nil {
		return fmt.Errorf("error checking tournaments: %v", err)
	}

	err = pool.QueryRow(ctx, "SELECT COUNT(*) FROM calcuttas").Scan(&calcuttaCount)
	if err != nil {
		return fmt.Errorf("error checking calcuttas: %v", err)
	}

	// Only skip if we have data in all key tables
	if schoolCount > 0 && tournamentCount > 0 && calcuttaCount > 0 {
		fmt.Printf("Seed data already exists (schools: %d, tournaments: %d, calcuttas: %d), skipping SQL dump import\n",
			schoolCount, tournamentCount, calcuttaCount)
		return nil
	}

	// If we have partial data, warn the user
	if schoolCount > 0 || tournamentCount > 0 || calcuttaCount > 0 {
		fmt.Printf("Warning: Partial seed data detected (schools: %d, tournaments: %d, calcuttas: %d)\n",
			schoolCount, tournamentCount, calcuttaCount)
		fmt.Println("Proceeding with SQL dump import - this may create duplicate data if not careful")
	}

	fmt.Println("Loading seed data from SQL dumps...")

	// Execute each SQL file in order
	for _, sqlFile := range sqlFiles {
		filePath := filepath.Join(dumpsDir, sqlFile)

		// Check if file exists
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			fmt.Printf("Warning: SQL dump file not found: %s (skipping)\n", sqlFile)
			continue
		}

		// Read the SQL file
		sqlContent, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("error reading SQL file %s: %v", sqlFile, err)
		}

		sanitizedSQL := stripPsqlMetaCommands(string(sqlContent))

		// Skip empty files
		if len(strings.TrimSpace(sanitizedSQL)) == 0 {
			fmt.Printf("Skipping empty file: %s\n", sqlFile)
			continue
		}

		fmt.Printf("  Importing %s...\n", sqlFile)

		// Execute the SQL
		_, err = pool.Exec(ctx, sanitizedSQL)
		if err != nil {
			return fmt.Errorf("error executing SQL file %s: %v", sqlFile, err)
		}
	}

	return nil
}
