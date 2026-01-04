package platform

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PGXPoolOptions struct {
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	HealthCheckPeriod time.Duration
}

func OpenPGXPool(ctx context.Context, cfg Config, opts *PGXPoolOptions) (*pgxpool.Pool, error) {
	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL environment variable is not set")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	pcfg, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	if cfg.MetricsEnabled {
		pcfg.ConnConfig.Tracer = newDBQueryTracer()
	}

	if opts != nil {
		if opts.MaxConns > 0 {
			pcfg.MaxConns = opts.MaxConns
		}
		if opts.MinConns > 0 {
			pcfg.MinConns = opts.MinConns
		}
		if opts.MaxConnLifetime > 0 {
			pcfg.MaxConnLifetime = opts.MaxConnLifetime
		}
		if opts.HealthCheckPeriod > 0 {
			pcfg.HealthCheckPeriod = opts.HealthCheckPeriod
		}
	}

	pool, err := pgxpool.NewWithConfig(ctx, pcfg)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
