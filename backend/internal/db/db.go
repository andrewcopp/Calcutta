package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

// Initialize sets up the database connection pool
func Initialize(ctx context.Context) error {
	// Get connection string from environment
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	// Create connection pool
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return fmt.Errorf("error parsing connection string: %v", err)
	}

	// Set pool configuration
	config.MaxConns = 10

	// Create the pool
	pool, err = pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return fmt.Errorf("error creating connection pool: %v", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("error connecting to the database: %v", err)
	}

	return nil
}

// GetPool returns the database connection pool
func GetPool() *pgxpool.Pool {
	return pool
}

// Close closes the database connection pool
func Close() {
	if pool != nil {
		pool.Close()
	}
}
