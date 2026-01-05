package db

import (
	"context"
	"fmt"
	"time"

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

	pool, err = platform.OpenPGXPool(ctx, cfg, &platform.PGXPoolOptions{
		MaxConns:          cfg.PGXPoolMaxConns,
		MinConns:          cfg.PGXPoolMinConns,
		MaxConnLifetime:   time.Duration(cfg.PGXPoolMaxConnLifetimeSeconds) * time.Second,
		HealthCheckPeriod: time.Duration(cfg.PGXPoolHealthCheckPeriodSeconds) * time.Second,
	})
	if err != nil {
		return fmt.Errorf("db_connect_failed: %w", err)
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
