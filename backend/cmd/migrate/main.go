package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"calcutta/internal/db"
)

func main() {
	// Parse command line flags
	up := flag.Bool("up", false, "Run migrations up")
	down := flag.Bool("down", false, "Run migrations down")
	flag.Parse()

	// Check if at least one flag is set
	if !*up && !*down {
		fmt.Println("Please specify either -up or -down flag")
		flag.Usage()
		os.Exit(1)
	}

	// Initialize context
	ctx := context.Background()

	// Run migrations
	if *up {
		fmt.Println("Running migrations up...")
		if err := db.RunMigrations(ctx); err != nil {
			log.Fatalf("Error running migrations: %v", err)
		}
		fmt.Println("Migrations completed successfully")
	}

	if *down {
		fmt.Println("Rolling back migrations...")
		if err := db.RollbackMigrations(ctx); err != nil {
			log.Fatalf("Error rolling back migrations: %v", err)
		}
		fmt.Println("Migrations rolled back successfully")
	}
}
