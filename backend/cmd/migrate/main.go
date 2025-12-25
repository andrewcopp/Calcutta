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
	force := flag.Int("force", 0, "Force schema migration version (clears dirty state)")
	flag.Parse()

	// Check if at least one flag is set
	if !*up && !*down && *force == 0 {
		fmt.Println("Please specify either -up or -down flag")
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
	if *force != 0 {
		fmt.Printf("Forcing schema migration version to %d (clearing dirty state)...\n", *force)
		if err := db.ForceSchemaMigrations(ctx, *force); err != nil {
			log.Fatalf("Error forcing schema migrations: %v", err)
		}
		fmt.Println("Schema migration version forced successfully")
	}

	if *up {
		fmt.Println("Running schema migrations up...")
		if err := db.RunSchemaMigrations(ctx); err != nil {
			log.Fatalf("Error running schema migrations: %v", err)
		}
		fmt.Println("Schema migrations completed successfully")
	}

	if *down {
		fmt.Println("Rolling back schema migrations...")
		if err := db.RollbackSchemaMigrations(ctx); err != nil {
			log.Fatalf("Error rolling back schema migrations: %v", err)
		}
		fmt.Println("Schema migrations rolled back successfully")
	}

	// Close the database connection
	db.Close()
}
