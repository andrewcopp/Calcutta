package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/andrewcopp/Calcutta/backend/pkg/common"
)

// TournamentData represents tournament data extracted from a Calcutta CSV file
type TournamentData struct {
	SchoolName      string
	Seed            int
	Region          string
	PointsScored    int
	TotalInvestment int
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

	// Get the directory containing CSV files
	csvDir := "../../migrations/seed/calcuttas"
	files, err := os.ReadDir(csvDir)
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

	// Process each CSV file
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".csv") {
			continue
		}

		stats.totalFiles++

		// Extract year from filename (assuming format like "2023.csv")
		year := strings.TrimSuffix(file.Name(), ".csv")
		log.Printf("Processing Calcutta data for year: %s", year)

		// Read and process the CSV file
		if err := processCalcuttaFile(db, filepath.Join(csvDir, file.Name()), year); err != nil {
			stats.failedFiles++
			errorMsg := fmt.Sprintf("Error processing file %s: %v", file.Name(), err)
			stats.errors = append(stats.errors, errorMsg)
			log.Printf(errorMsg)
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

func processCalcuttaFile(db *sql.DB, filepath, year string) error {
	// Open CSV file
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	// Create CSV reader
	reader := csv.NewReader(file)

	// Read header
	_, err = reader.Read()
	if err != nil {
		return fmt.Errorf("error reading header: %v", err)
	}

	// Read sub-header (player data headers)
	_, err = reader.Read()
	if err != nil {
		return fmt.Errorf("error reading sub-header: %v", err)
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
		totalRows     int
		skippedRows   int
		processedRows int
		errors        int
	}{
		totalRows:     0,
		skippedRows:   0,
		processedRows: 0,
		errors:        0,
	}

	// Process each row
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading row: %v", err)
		}

		// Skip empty rows and summary rows
		if len(row) < 5 || row[0] == "" || !common.IsValidTeamName(row[0]) {
			continue
		}

		// Get standardized school name
		schoolName := common.GetStandardizedSchoolName(row[0])
		if schoolName == "" {
			continue
		}

		// Parse data
		seed, _ := strconv.Atoi(row[1])
		points, _ := strconv.Atoi(row[3])
		investment, _ := strconv.Atoi(row[4])

		data := TournamentData{
			SchoolName:      schoolName,
			Seed:            seed,
			Region:          row[2],
			PointsScored:    points,
			TotalInvestment: investment,
		}

		// Find or create school
		schoolID, err := findSchool(context.Background(), tx, data.SchoolName)
		if err != nil {
			stats.errors++
			log.Printf("Error finding/creating school: %v", err)
			continue
		}

		// Calculate byes and wins based on points scored
		byes, wins := calculateByesAndWins(data.PointsScored)

		// If the team name had an asterisk, they played in the First Four (0 byes)
		if strings.HasSuffix(data.SchoolName, "*") {
			byes = 0
			log.Printf("Team %s played in First Four for year %s", data.SchoolName, year)
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
			data.Seed,
			data.Region,
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

		stats.processedRows++
	}

	// Log statistics
	log.Printf("File processing statistics for %s:", year)
	log.Printf("  Total rows: %d", stats.totalRows)
	log.Printf("  Skipped rows: %d", stats.skippedRows)
	log.Printf("  Processed rows: %d", stats.processedRows)
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

func parseTournamentTeamRow(row []string) (*TournamentData, error) {
	if len(row) < 5 {
		return nil, fmt.Errorf("row has insufficient columns: %d", len(row))
	}

	// Skip rows with "Play-in losers"
	if row[0] == "Play-in losers" {
		return nil, fmt.Errorf("skipping Play-in losers row")
	}

	// Skip rows with empty points values
	if row[3] == "" {
		return nil, fmt.Errorf("skipping row with empty points value")
	}

	// Skip rows that represent totals
	if strings.Contains(strings.ToUpper(row[0]), "TOTAL") {
		return nil, fmt.Errorf("skipping total row")
	}

	// Parse investment to check if it's a total row
	investment, err := strconv.Atoi(row[4])
	if err != nil {
		return nil, fmt.Errorf("invalid investment value: %v", err)
	}

	data := &TournamentData{
		SchoolName: row[0],
	}

	// Parse seed - handle empty string case
	if row[1] == "" {
		data.Seed = -1 // Use -1 to indicate First Four team with unknown seed
	} else {
		seed, err := strconv.Atoi(row[1])
		if err != nil {
			return nil, fmt.Errorf("invalid seed value: %v", err)
		}
		data.Seed = seed
	}

	// Parse region
	data.Region = row[2]

	// Parse points scored
	points, err := strconv.Atoi(row[3])
	if err != nil {
		return nil, fmt.Errorf("invalid points value: %v", err)
	}
	data.PointsScored = points

	// Set the total investment
	data.TotalInvestment = investment

	return data, nil
}

// calculateByesAndWins determine
// According to the rules:
// First Round Win: 50 points
// Sweet 16 Appearance: +100 points
// Elite 8 Appearance: +150 points
// Final Four Appearance: +200 points
// Championship Game Appearance: +250 points
// Tournament Winner: +300 points
func calculateByesAndWins(points int) (byes, wins int) {
	// Validate points value
	validPoints := map[int]bool{
		0:    true, // First Four Loss
		50:   true, // First Round Win
		150:  true, // Round of 32 Win
		300:  true, // Sweet 16 Win
		500:  true, // Elite 8 Win
		750:  true, // Final Four Win
		1050: true, // Tournament Winner
	}

	if !validPoints[points] {
		log.Printf("Warning: Unexpected points value %d. This may indicate data inconsistency.", points)
		// Estimate wins based on points, but log a warning
		wins = points / 150
		if wins > 6 {
			wins = 6 // Cap at 6 wins
		}
		byes = 1 // Default to 1 bye since most teams receive a bye to the First Round
		return byes, wins
	}

	// Calculate wins based on points
	switch points {
	case 0:
		wins = 0
	case 50:
		wins = 1
	case 150:
		wins = 2
	case 300:
		wins = 3
	case 500:
		wins = 4
	case 750:
		wins = 5
	case 1050:
		wins = 6
	}

	// Determine byes based on seed (simplified)
	// In reality, this would depend on the tournament structure
	// For now, we'll assume teams with seeds 1-4 get byes
	// This is a simplification and may need adjustment
	byes = 1 // Default to 1 bye since most teams receive a bye to the First Round

	return byes, wins
}

// findSchool finds a school by name, using the schoolNameMap to standardize the name
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
			// School not found - create it
			schoolID = uuid.New().String()
			now := time.Now()
			_, err = tx.ExecContext(ctx, `
				INSERT INTO schools (id, name, created_at, updated_at)
				VALUES ($1, $2, $3, $4)
			`, schoolID, standardizedName, now, now)
			if err != nil {
				return "", fmt.Errorf("error creating school: %v", err)
			}
			log.Printf("Created new school: %s", standardizedName)

			// Only track schools that don't map to any existing school
			// This is the only case where we need to track for future mapping
			unmappedSchoolNames[schoolName] = true

			return schoolID, nil
		}
		return "", fmt.Errorf("error finding school: %v", err)
	}

	return schoolID, nil
}
