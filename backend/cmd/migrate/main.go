package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/andrewcopp/Calcutta/backend/internal/db"
)

func main() {
	// Parse command line flags
	up := flag.Bool("up", false, "Run migrations up")
	down := flag.Bool("down", false, "Run migrations down")
	seed := flag.Bool("seed", false, "Run seed data migrations")
	skipSeed := flag.Bool("skip-seed", false, "Skip seed data migrations when running up")
	flag.Parse()

	// Check if at least one flag is set
	if !*up && !*down && !*seed {
		fmt.Println("Please specify either -up, -down, or -seed flag")
		flag.Usage()
		os.Exit(1)
	}

	// Initialize context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Initialize database connection
	fmt.Println("Initializing database connection...")
	if err := db.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println("Database connection initialized successfully")

	// Run migrations
	if *up {
		fmt.Println("Running schema migrations up...")
		if err := db.RunSchemaMigrations(ctx); err != nil {
			log.Fatalf("Error running schema migrations: %v", err)
		}
		fmt.Println("Schema migrations completed successfully")

		// Run seed migrations unless explicitly skipped
		if !*skipSeed {
			fmt.Println("Running seed data migrations...")
			if err := db.RunSeedMigrations(ctx); err != nil {
				log.Fatalf("Error running seed migrations: %v", err)
			}
			fmt.Println("Seed migrations completed successfully")
		}
	}

	if *down {
		fmt.Println("Rolling back schema migrations...")
		if err := db.RollbackSchemaMigrations(ctx); err != nil {
			log.Fatalf("Error rolling back schema migrations: %v", err)
		}
		fmt.Println("Schema migrations rolled back successfully")
	}

	if *seed {
		fmt.Println("Running seed data migrations...")
		if err := db.RunSeedMigrations(ctx); err != nil {
			log.Fatalf("Error running seed migrations: %v", err)
		}
		fmt.Println("Seed migrations completed successfully")
	}

	// Close the database connection
	db.Close()
}
