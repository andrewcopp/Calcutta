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
	msRidgeV1Params, _ := json.Marshal(map[string]any{
		"feature_set": "optimal",
	})
	msRidgeV1RecentParams, _ := json.Marshal(map[string]any{
		"feature_set":        "optimal",
		"train_years_window": 3,
	})
	msRidgeV2Params, _ := json.Marshal(map[string]any{
		"feature_set": "optimal_v2",
	})
	msRidgeV2ShrunkParams, _ := json.Marshal(map[string]any{
		"feature_set":         "optimal_v2",
		"seed_prior_monotone": true,
		"seed_prior_k":        20.0,
		"program_prior_k":     50.0,
	})
	msRidgeV2Underbid1SigmaParams, _ := json.Marshal(map[string]any{
		"feature_set": "optimal_v2",
	})
	msRidgeV2LogParams, _ := json.Marshal(map[string]any{
		"feature_set":         "optimal_v2",
		"target_transform":    "log",
		"seed_prior_monotone": true,
	})
	msNaiveEVParams, _ := json.Marshal(map[string]any{})
	msOracleParams, _ := json.Marshal(map[string]any{})

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
			Name:        "ridge-v1",
			Description: "Ridge Regression V1 (optimal)",
			ParamsJSON:  msRidgeV1Params,
		},
		{
			Kind:        "market_share",
			Name:        "ridge-v1-recent",
			Description: "Ridge Regression V1 (recent training window)",
			ParamsJSON:  msRidgeV1RecentParams,
		},
		{
			Kind:        "market_share",
			Name:        "ridge-v2",
			Description: "Ridge Regression V2 (optimal_v2)",
			ParamsJSON:  msRidgeV2Params,
		},
		{
			Kind:        "market_share",
			Name:        "ridge-v2-shrunk",
			Description: "Ridge Regression V2 (shrunk priors)",
			ParamsJSON:  msRidgeV2ShrunkParams,
		},
		{
			Kind:        "market_share",
			Name:        "ridge-v2-underbid-1sigma",
			Description: "Ridge Regression V2 (underbid 1-sigma, global)",
			ParamsJSON:  msRidgeV2Underbid1SigmaParams,
		},
		{
			Kind:        "market_share",
			Name:        "ridge-v2-log",
			Description: "Ridge Regression V2 (log target)",
			ParamsJSON:  msRidgeV2LogParams,
		},
		{
			Kind:        "market_share",
			Name:        "naive-ev-baseline",
			Description: "Naive EV Baseline",
			ParamsJSON:  msNaiveEVParams,
		},
		{
			Kind:        "market_share",
			Name:        "oracle_actual_market",
			Description: "Oracle Actual Market (Python runner)",
			ParamsJSON:  msOracleParams,
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
