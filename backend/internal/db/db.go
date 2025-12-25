package db

import (
	"context"
	"fmt"

	"github.com/andrewcopp/Calcutta/backend/internal/platform"
	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

// Initialize sets up the database connection pool
func Initialize(ctx context.Context) error {
	cfg, err := platform.LoadConfigFromEnv()
	if err != nil {
		return err
	}

	// Create connection pool
	config, err := pgxpool.ParseConfig(cfg.DatabaseURL)
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
