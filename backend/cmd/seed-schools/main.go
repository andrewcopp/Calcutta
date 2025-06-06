package main

import (
	"context"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// Load database connection string from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to the database
	db, err := sql.Open("pgx", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
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

	// Load the schools file
	csvPath := filepath.Join("migrations", "seed", "schools", "active_d1_teams.csv")
	file, err := os.Open(csvPath)
	if err != nil {
		log.Fatalf("Failed to open schools file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Skip header row
	if _, err := reader.Read(); err != nil {
		log.Fatalf("Failed to read CSV header: %v", err)
	}

	// Read all records
	records, err := reader.ReadAll()
	if err != nil {
		log.Fatalf("Failed to read CSV records: %v", err)
	}

	// Begin a transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	// Clear existing schools
	_, err = tx.Exec("DELETE FROM schools")
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			log.Printf("Error rolling back transaction: %v", rbErr)
		}
		log.Fatalf("Failed to clear existing schools: %v", err)
	}

	// Insert schools
	now := time.Now()
	for _, record := range records {
		if len(record) == 0 {
			continue
		}
		schoolName := record[0]
		_, err = tx.Exec(`
			INSERT INTO schools (id, name, created_at, updated_at)
			VALUES ($1, $2, $3, $4)
		`, uuid.New().String(), schoolName, now, now)
		if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				log.Printf("Error rolling back transaction: %v", rbErr)
			}
			log.Fatalf("Failed to insert school %s: %v", schoolName, err)
		}
		fmt.Printf("Inserted school: %s\n", schoolName)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Printf("Successfully seeded %d schools\n", len(records))
}
