package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/andrewcopp/Calcutta/backend/pkg/common"
)

// TournamentTeam represents a team in the tournament from the JSON data
type TournamentTeam struct {
	Name   string `json:"name"`
	Seed   int    `json:"seed"`
	Region string `json:"region"`
}

// Track unmapped school names for future reference
var unmappedSchoolNames = make(map[string]bool)

func main() {
	// Get database connection string from environment
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to database
	db, err := sql.Open("pgx", connString)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Get the directory containing JSON files
	jsonDir := "../../../data/tournament_teams"
	files, err := os.ReadDir(jsonDir)
	if err != nil {
		log.Fatalf("Error reading directory: %v", err)
	}

	// Track statistics for reporting
	stats := struct {
		totalFiles     int
		processedFiles int
		failedFiles    int
		errors         []string
	}{
		totalFiles:     0,
		processedFiles: 0,
		failedFiles:    0,
		errors:         make([]string, 0),
	}

	// Process each JSON file
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		stats.totalFiles++

		// Extract year from filename (assuming format like "2021-ncaa-mens.json")
		year := strings.Split(file.Name(), "-")[0]
		log.Printf("Processing tournament data for year: %s", year)

		// Read and process the JSON file
		if err := processTournamentFile(db, filepath.Join(jsonDir, file.Name()), year); err != nil {
			stats.failedFiles++
			errorMsg := fmt.Sprintf("Error processing file %s: %v", file.Name(), err)
			stats.errors = append(stats.errors, errorMsg)
			log.Printf("%s", errorMsg)
			continue
		}

		stats.processedFiles++
	}

	// Print summary
	log.Println("\nProcessing Summary:")
	log.Printf("Total files: %d", stats.totalFiles)
	log.Printf("Successfully processed: %d", stats.processedFiles)
	log.Printf("Failed to process: %d", stats.failedFiles)

	if len(stats.errors) > 0 {
		log.Println("\nErrors encountered:")
		for _, err := range stats.errors {
			log.Printf("  - %s", err)
		}
	}

	// Print unmapped school names for future reference
	if len(unmappedSchoolNames) > 0 {
		log.Println("\nThe following school names were not mapped:")
		for name := range unmappedSchoolNames {
			log.Printf("  - %s", name)
		}
		log.Println("Consider adding these to the schoolNameMap.")
	}

	// Exit with error if any files failed to process
	if stats.failedFiles > 0 {
		os.Exit(1)
	}
}

func processTournamentFile(db *sql.DB, filepath, year string) error {
	// Open JSON file
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Create JSON decoder
	decoder := json.NewDecoder(file)

	// Read the opening bracket
	_, err = decoder.Token()
	if err != nil {
		return fmt.Errorf("error reading JSON array start: %v", err)
	}

	// Create tournament
	tournamentID := uuid.New().String()
	now := time.Now()

	// Begin transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}

	// Create a function to handle transaction rollback
	rollback := func() {
		if err := tx.Rollback(); err != nil && err != sql.ErrTxDone {
			log.Printf("Error rolling back transaction: %v", err)
		}
	}

	// Insert tournament
	_, err = tx.Exec(`
		INSERT INTO tournaments (id, name, rounds, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`,
		tournamentID,
		fmt.Sprintf("NCAA Tournament %s", year),
		7, // NCAA tournament has 7 rounds (First Four, First Round, Round of 32, Sweet 16, Elite 8, Final Four, Championship)
		now,
		now,
	)
	if err != nil {
		rollback()
		return fmt.Errorf("error inserting tournament: %v", err)
	}

	// Track statistics for reporting
	stats := struct {
		totalTeams     int
		skippedTeams   int
		processedTeams int
		errors         int
	}{
		totalTeams:     0,
		skippedTeams:   0,
		processedTeams: 0,
		errors:         0,
	}

	// Process each team
	for decoder.More() {
		var team TournamentTeam
		if err := decoder.Decode(&team); err != nil {
			return fmt.Errorf("error decoding team: %v", err)
		}

		stats.totalTeams++

		// Skip empty team names
		if team.Name == "" || !common.IsValidTeamName(team.Name) {
			stats.skippedTeams++
			continue
		}

		// Skip "Play-in Losers" entries as they are not real teams
		if team.Name == "Play-in Losers" || team.Name == "Play-in losers" {
			stats.skippedTeams++
			continue
		}

		// Get standardized school name
		schoolName := common.GetStandardizedSchoolName(team.Name)
		if schoolName == "" {
			stats.skippedTeams++
			continue
		}

		// Find or create school
		schoolID, err := findSchool(context.Background(), tx, schoolName)
		if err != nil {
			// Log the error but continue processing other teams
			log.Printf("Warning: %v", err)
			stats.errors++
			continue
		}

		// For now, we'll set byes to 1 and wins to 0 since we don't have points data
		// This can be updated later if needed
		byes := 1
		wins := 0

		// If the team name had an asterisk, they played in the First Four (0 byes)
		if strings.HasSuffix(team.Name, "*") {
			byes = 0
			log.Printf("Team %s played in First Four for year %s", team.Name, year)
		}

		// Create tournament team
		teamID := uuid.New().String()
		_, err = tx.Exec(`
			INSERT INTO tournament_teams (
				id, tournament_id, school_id, seed, region, byes, wins, 
				created_at, updated_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`,
			teamID,
			tournamentID,
			schoolID,
			team.Seed,
			team.Region,
			byes,
			wins,
			now,
			now,
		)
		if err != nil {
			stats.errors++
			log.Printf("Error inserting team: %v", err)
			continue
		}

		stats.processedTeams++
	}

	// Read the closing bracket
	_, err = decoder.Token()
	if err != nil {
		return fmt.Errorf("error reading JSON array end: %v", err)
	}

	// Log statistics
	log.Printf("File processing statistics for %s:", year)
	log.Printf("  Total teams: %d", stats.totalTeams)
	log.Printf("  Skipped teams: %d", stats.skippedTeams)
	log.Printf("  Processed teams: %d", stats.processedTeams)
	log.Printf("  Errors: %d", stats.errors)

	// If there were any errors, rollback the transaction
	if stats.errors > 0 {
		rollback()
		return fmt.Errorf("errors occurred while processing file: %d errors", stats.errors)
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		rollback()
		return fmt.Errorf("error committing transaction: %v", err)
	}

	return nil
}

func findSchool(ctx context.Context, tx *sql.Tx, schoolName string) (string, error) {
	// Log the exact characters in the school name for debugging
	if strings.Contains(schoolName, "St.") {
		log.Printf("Debug - School name before processing: %q", schoolName)
		for i, r := range schoolName {
			log.Printf("Debug - Character at position %d: %q (Unicode: %U)", i, r, r)
		}
	}

	// Replace different types of apostrophes with a standard one
	schoolName = strings.ReplaceAll(schoolName, "'", "'")
	schoolName = strings.ReplaceAll(schoolName, "'", "'")

	// Remove asterisk if present (do this BEFORE looking up in the map)
	schoolName = strings.TrimSuffix(schoolName, "*")

	// Trim whitespace from the school name
	schoolName = strings.TrimSpace(schoolName)

	// Log the school name after processing for debugging
	if strings.Contains(schoolName, "St.") {
		log.Printf("Debug - School name after processing: %q", schoolName)
	}

	// Check if we have a mapping for this school name
	standardizedName, hasMapping := common.SchoolNameMap[schoolName]
	if hasMapping {
		log.Printf("Mapped school name: %s -> %s", schoolName, standardizedName)
	} else {
		// If no mapping exists, try to find the school directly
		standardizedName = schoolName
	}

	// Find the school in the database
	var schoolID string
	err := tx.QueryRowContext(ctx, `
		SELECT id FROM schools WHERE name = $1
	`, standardizedName).Scan(&schoolID)

	if err != nil {
		if err == sql.ErrNoRows {
			// School not found - track it for future mapping
			unmappedSchoolNames[schoolName] = true
			return "", fmt.Errorf("school not found: %s (standardized: %s)", schoolName, standardizedName)
		}
		return "", fmt.Errorf("error finding school: %v", err)
	}

	return schoolID, nil
}
