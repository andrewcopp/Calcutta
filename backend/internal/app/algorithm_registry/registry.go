package algorithm_registry

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Algorithm struct {
	Kind        string
	Name        string
	Description string
	ParamsJSON  json.RawMessage
}

func RegisteredAlgorithms() []Algorithm {
	params1, _ := json.Marshal(map[string]any{
		"model_version": "kenpom-v1-go",
		"kenpom_scale":  10.0,
		"n_sims":        5000,
		"seed":          42,
	})
	params2, _ := json.Marshal(map[string]any{
		"model_version": "kenpom-v1-sigma11-go",
		"kenpom_scale":  11.0,
		"n_sims":        5000,
		"seed":          42,
	})

	msRidgeParams, _ := json.Marshal(map[string]any{})
	msNaiveEVParams, _ := json.Marshal(map[string]any{})

	return []Algorithm{
		{
			Kind:        "game_outcomes",
			Name:        "kenpom-v1-go",
			Description: "KenPom V1 (Go), sigma=10",
			ParamsJSON:  params1,
		},
		{
			Kind:        "game_outcomes",
			Name:        "kenpom-v1-sigma11-go",
			Description: "KenPom V1 (Go), sigma=11",
			ParamsJSON:  params2,
		},
		{
			Kind:        "market_share",
			Name:        "ridge",
			Description: "Ridge Regression (Python runner)",
			ParamsJSON:  msRidgeParams,
		},
		{
			Kind:        "market_share",
			Name:        "naive-ev-baseline",
			Description: "Naive EV Baseline",
			ParamsJSON:  msNaiveEVParams,
		},
	}
}

func SyncToDatabase(ctx context.Context, pool *pgxpool.Pool, algorithms []Algorithm) error {
	if pool == nil {
		return fmt.Errorf("pool is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	if _, err := pool.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS derived`); err != nil {
		return err
	}

	for _, a := range algorithms {
		if a.Kind == "" {
			return fmt.Errorf("algorithm kind is required")
		}
		if a.Name == "" {
			return fmt.Errorf("algorithm name is required")
		}

		params := a.ParamsJSON
		if len(params) == 0 {
			params = json.RawMessage([]byte(`{}`))
		}

		_, err := pool.Exec(ctx, `
			INSERT INTO derived.algorithms (kind, name, description, params_json)
			VALUES ($1, $2, $3, $4::jsonb)
			ON CONFLICT (kind, name) WHERE deleted_at IS NULL
			DO UPDATE SET
				description = EXCLUDED.description,
				params_json = EXCLUDED.params_json,
				updated_at = NOW(),
				deleted_at = NULL
		`, a.Kind, a.Name, a.Description, string(params))
		if err != nil {
			return err
		}
	}

	_, err := pool.Exec(ctx, `
		UPDATE derived.algorithms
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE kind = 'game_outcomes'
			AND name = ANY($1::text[])
			AND deleted_at IS NULL
	`, []string{"smoke_go", "smoke-go"})
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		UPDATE derived.algorithms
		SET deleted_at = NOW(), updated_at = NOW()
		WHERE kind = 'market_share'
			AND name = ANY($1::text[])
			AND deleted_at IS NULL
	`, []string{"smoke_ms"})
	if err != nil {
		return err
	}

	return nil
}
