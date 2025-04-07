package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/andrewcopp/Calcutta/backend/pkg/common"
)

// CalcuttaData represents the data from a CSV file
type CalcuttaData struct {
	Year     int
	Teams    []TeamData
	Entries  []EntryData
	Filename string
}

// TeamData represents a team in the tournament
type TeamData struct {
	Name       string
	Seed       int
	Region     string
	Points     int
	Investment int
}

// EntryData represents a player's entry in the Calcutta
type EntryData struct {
	Name  string
	Teams []EntryTeamData
}

// EntryTeamData represents a team in a player's entry
type EntryTeamData struct {
	TeamName string
	Bid      float64
	Points   int
}

// main function to seed the database with Calcutta data
func main() {
	// Connect to the database
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		fmt.Println("DATABASE_URL environment variable is not set")
		os.Exit(1)
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		fmt.Printf("Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	// Get the data directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Printf("Error getting working directory: %v\n", err)
		os.Exit(1)
	}
	// Go up two directories from cmd/seed-calcuttas to backend
	backendDir := filepath.Join(wd, "..", "..")
	dataDir := filepath.Join(backendDir, "migrations", "seed", "calcuttas")
	fmt.Printf("Looking for data in: %s\n", dataDir)

	// Process each CSV file in the data directory
	files, err := os.ReadDir(dataDir)
	if err != nil {
		fmt.Printf("Error reading data directory: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".csv") {
			continue
		}

		// Extract year from filename
		yearStr := strings.TrimSuffix(file.Name(), ".csv")
		year, err := strconv.Atoi(yearStr)
		if err != nil {
			fmt.Printf("Error parsing year from filename %s: %v\n", file.Name(), err)
			continue
		}

		// Process the CSV file
		calcuttaData, err := processCSVFile(filepath.Join(dataDir, file.Name()), year)
		if err != nil {
			fmt.Printf("Error processing file %s: %v\n", file.Name(), err)
			continue
		}

		// Seed the database with the Calcutta data
		err = seedCalcutta(database, calcuttaData)
		if err != nil {
			fmt.Printf("Error seeding Calcutta for year %d: %v\n", year, err)
			continue
		}

		fmt.Printf("Successfully seeded Calcutta for year %d\n", year)
	}
}

// processCSVFile reads a CSV file and returns the Calcutta data
func processCSVFile(filepath string, year int) (*CalcuttaData, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header row
	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading header: %v", err)
	}

	// Read subheader row
	_, err = reader.Read()
	if err != nil {
		return nil, fmt.Errorf("error reading subheader: %v", err)
	}

	// Extract player names from header
	players := make([]string, 0)
	for i := 5; i < len(header); i += 3 {
		if header[i] != "" {
			players = append(players, header[i])
		}
	}

	// Create Calcutta data
	calcuttaData := &CalcuttaData{
		Year:     year,
		Teams:    make([]TeamData, 0),
		Entries:  make([]EntryData, 0),
		Filename: filepath,
	}

	// Initialize entries
	for _, player := range players {
		calcuttaData.Entries = append(calcuttaData.Entries, EntryData{
			Name:  player,
			Teams: make([]EntryTeamData, 0),
		})
	}

	// Read data rows
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error reading row: %v", err)
		}

		// Skip empty rows
		if len(row) < 5 || row[0] == "" {
			continue
		}

		// Parse team data
		seed, _ := strconv.Atoi(row[1])
		points, _ := strconv.Atoi(row[3])
		investment, _ := strconv.Atoi(row[4])

		team := TeamData{
			Name:       row[0],
			Seed:       seed,
			Region:     row[2],
			Points:     points,
			Investment: investment,
		}
		calcuttaData.Teams = append(calcuttaData.Teams, team)

		// Parse entry data
		for i := range players {
			bidIndex := 5 + (i * 3)
			if bidIndex+1 >= len(row) {
				continue
			}

			bid, err := strconv.ParseFloat(row[bidIndex], 64)
			if err != nil || bid == 0 {
				continue
			}

			points, _ := strconv.Atoi(row[bidIndex+2])

			entryTeam := EntryTeamData{
				TeamName: row[0],
				Bid:      bid,
				Points:   points,
			}

			calcuttaData.Entries[i].Teams = append(calcuttaData.Entries[i].Teams, entryTeam)
		}
	}

	return calcuttaData, nil
}

// seedCalcutta seeds the database with Calcutta data
func seedCalcutta(database *gorm.DB, data *CalcuttaData) error {
	// Start a transaction
	tx := database.Begin()
	if tx.Error != nil {
		return fmt.Errorf("error starting transaction: %v", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find the tournament for this year
	var tournament struct {
		ID string
	}
	err := tx.Raw("SELECT id FROM tournaments WHERE name LIKE ?", fmt.Sprintf("%%%d%%", data.Year)).Scan(&tournament).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error finding tournament for year %d: %v", data.Year, err)
	}

	if tournament.ID == "" {
		tx.Rollback()
		return fmt.Errorf("no tournament found for year %d", data.Year)
	}

	// Find or create the admin user for all Calcuttas
	var adminUser struct {
		ID string
	}
	err = tx.Raw(`SELECT id FROM users WHERE email = ?`, "admin@calcutta.com").Scan(&adminUser).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return fmt.Errorf("error checking for admin user: %v", err)
	}

	if adminUser.ID == "" {
		// Create the admin user if it doesn't exist
		adminUser.ID = uuid.New().String()
		err = tx.Exec(`
			INSERT INTO users (id, email, first_name, last_name, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, adminUser.ID, "admin@calcutta.com", "Calcutta", "Admin", time.Now(), time.Now()).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error creating admin user: %v", err)
		}
	}

	// Create the Calcutta
	calcuttaID := uuid.New().String()
	err = tx.Exec(`
		INSERT INTO calcuttas (id, tournament_id, owner_id, name, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, calcuttaID, tournament.ID, adminUser.ID, fmt.Sprintf("%d Calcutta", data.Year), time.Now(), time.Now()).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("error creating Calcutta: %v", err)
	}

	// Create Calcutta rounds
	rounds := []struct {
		round  int
		points int
	}{
		{1, 50},  // Round of 64
		{2, 100}, // Round of 32
		{3, 150}, // Sweet 16
		{4, 200}, // Elite 8
		{5, 250}, // Final 4
		{6, 300}, // Championship
	}

	for _, r := range rounds {
		roundID := uuid.New().String()
		err = tx.Exec(`
			INSERT INTO calcutta_rounds (id, calcutta_id, round, points, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, roundID, calcuttaID, r.round, r.points, time.Now(), time.Now()).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error creating Calcutta round: %v", err)
		}
	}

	// Create a map of team name to seed for quick lookup
	teamSeedMap := make(map[string]int)
	for _, team := range data.Teams {
		standardizedName := common.GetStandardizedSchoolName(team.Name)
		teamSeedMap[standardizedName] = team.Seed
	}

	// Create entries for each player
	for _, entryData := range data.Entries {
		// Create the entry
		entryID := uuid.New().String()
		err = tx.Exec(`
			INSERT INTO calcutta_entries (id, calcutta_id, user_id, name, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
		`, entryID, calcuttaID, adminUser.ID, entryData.Name, time.Now(), time.Now()).Error
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("error creating entry: %v", err)
		}

		// Add teams to the entry
		for _, teamData := range entryData.Teams {
			standardizedName := common.GetStandardizedSchoolName(teamData.TeamName)
			seed := teamSeedMap[standardizedName]

			// Find the tournament team by seed and standardized school name
			var tournamentTeam struct {
				ID string
			}
			err = tx.Raw(`
				SELECT tt.id 
				FROM tournament_teams tt
				JOIN schools s ON tt.school_id = s.id
				WHERE tt.tournament_id = ? AND tt.seed = ? AND s.name = ?
			`, tournament.ID, seed, standardizedName).Scan(&tournamentTeam).Error
			if err != nil {
				fmt.Printf("Warning: Could not find team %s (standardized: %s) for tournament %d\n", teamData.TeamName, standardizedName, data.Year)
				continue
			}

			if tournamentTeam.ID == "" {
				fmt.Printf("Warning: Found no tournament team for %s (standardized: %s) with seed %d in tournament %d\n", teamData.TeamName, standardizedName, seed, data.Year)
				continue
			}

			fmt.Printf("Found tournament team %s for %s (standardized: %s) with seed %d in tournament %d\n", tournamentTeam.ID, teamData.TeamName, standardizedName, seed, data.Year)

			// Create entry team
			entryTeamID := uuid.New().String()
			err = tx.Exec(`
				INSERT INTO calcutta_entry_teams (id, entry_id, team_id, bid, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?)
			`, entryTeamID, entryID, tournamentTeam.ID, teamData.Bid, time.Now(), time.Now()).Error
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("error creating entry team: %v", err)
			}
		}
	}

	return tx.Commit().Error
}
