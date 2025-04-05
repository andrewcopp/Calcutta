package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"calcutta/internal/db"
)

func main() {
	// Set up context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Initialize database connection
	fmt.Println("Initializing database connection...")
	if err := db.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	fmt.Println("Database connection initialized successfully")

	// Get the connection pool
	pool := db.GetPool()
	if pool == nil {
		log.Fatal("Connection pool is nil")
	}

	// Test the connection with a simple query
	fmt.Println("Testing connection with a simple query...")
	var result string
	err := pool.QueryRow(ctx, "SELECT current_database()").Scan(&result)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	fmt.Printf("Connected to database: %s\n", result)

	// Print connection stats
	stats := pool.Stat()
	fmt.Printf("Connection pool stats:\n")
	fmt.Printf("  Total connections: %d\n", stats.TotalConns())
	fmt.Printf("  Idle connections: %d\n", stats.IdleConns())
	fmt.Printf("  Max connections: %d\n", stats.MaxConns())

	// Close the connection
	fmt.Println("Closing database connection...")
	db.Close()
	fmt.Println("Database connection closed")
}
