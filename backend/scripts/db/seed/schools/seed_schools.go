package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Schools represents the structure of our JSON file
type Schools struct {
	Schools []string `json:"schools"`
}

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
	defer db.Close()

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	// Load the schools file
	schoolsData, err := os.ReadFile("schools.json")
	if err != nil {
		log.Fatalf("Failed to read schools file: %v", err)
	}

	var schools Schools
	if err := json.Unmarshal(schoolsData, &schools); err != nil {
		log.Fatalf("Failed to parse schools file: %v", err)
	}

	// Begin a transaction
	tx, err := db.BeginTx(context.Background(), nil)
	if err != nil {
		log.Fatalf("Failed to begin transaction: %v", err)
	}

	// Clear existing schools
	_, err = tx.Exec("DELETE FROM schools")
	if err != nil {
		tx.Rollback()
		log.Fatalf("Failed to clear existing schools: %v", err)
	}

	// Insert schools
	now := time.Now()
	for _, schoolName := range schools.Schools {
		_, err = tx.Exec(`
			INSERT INTO schools (id, name, created_at, updated_at)
			VALUES ($1, $2, $3, $4)
		`, uuid.New().String(), schoolName, now, now)
		if err != nil {
			tx.Rollback()
			log.Fatalf("Failed to insert school %s: %v", schoolName, err)
		}
		fmt.Printf("Inserted school: %s\n", schoolName)
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Fatalf("Failed to commit transaction: %v", err)
	}

	fmt.Printf("Successfully seeded %d schools\n", len(schools.Schools))
}
