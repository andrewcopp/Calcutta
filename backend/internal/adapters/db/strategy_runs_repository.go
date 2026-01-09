package db

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/internal/app/strategy_runs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func uuidParamOrNil(id *string) any {
	if id == nil {
		return nil
	}
	v := strings.TrimSpace(*id)
	if v == "" {
		return nil
	}
	return v
}

type StrategyRunsRepository struct {
	pool *pgxpool.Pool
}

func NewStrategyRunsRepository(pool *pgxpool.Pool) *StrategyRunsRepository {
	return &StrategyRunsRepository{pool: pool}
}

func (r *StrategyRunsRepository) CreateRun(ctx context.Context, p strategy_runs.CreateRunParams) (*strategy_runs.CreateRunResult, error) {
	// Resolve market_share_run_id and input_market_share_artifact_id per legacy handler behavior.
	marketShareRunID := (*string)(nil)
	inputMarketShareArtifactID := (*string)(nil)
	if p.MarketShareArtifactID != nil && strings.TrimSpace(*p.MarketShareArtifactID) != "" {
		artifactID := strings.TrimSpace(*p.MarketShareArtifactID)
		var runIDFromArtifact string
		if err := r.pool.QueryRow(ctx, `
			SELECT r.id::text
			FROM derived.run_artifacts a
			JOIN derived.market_share_runs r
				ON r.id = a.run_id
				AND r.deleted_at IS NULL
			WHERE a.id = $1::uuid
				AND a.run_kind = 'market_share'
				AND a.run_id = r.id
				AND r.calcutta_id = $2::uuid
				AND a.artifact_kind = 'metrics'
				AND a.deleted_at IS NULL
			LIMIT 1
		`, artifactID, p.CalcuttaID).Scan(&runIDFromArtifact); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, strategy_runs.ErrMarketShareArtifactNotFound
			}
			return nil, err
		}
		marketShareRunID = &runIDFromArtifact
		inputMarketShareArtifactID = &artifactID
	} else if p.MarketShareRunID != nil && strings.TrimSpace(*p.MarketShareRunID) != "" {
		runID := strings.TrimSpace(*p.MarketShareRunID)
		var artifactID string
		if err := r.pool.QueryRow(ctx, `
			SELECT a.id::text
			FROM derived.market_share_runs r
			JOIN derived.run_artifacts a
				ON a.run_kind = 'market_share'
				AND a.run_id = r.id
				AND a.artifact_kind = 'metrics'
				AND a.deleted_at IS NULL
			WHERE r.id = $1::uuid
				AND r.calcutta_id = $2::uuid
				AND r.deleted_at IS NULL
			LIMIT 1
		`, runID, p.CalcuttaID).Scan(&artifactID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, strategy_runs.ErrMarketShareRunMissingMetrics
			}
			return nil, err
		}
		artifactID = strings.TrimSpace(artifactID)
		if artifactID != "" {
			inputMarketShareArtifactID = &artifactID
		}
		marketShareRunID = &runID
	} else {
		// handler enforces required-ness; keep internal behavior strict.
		return nil, strategy_runs.ErrMarketShareArtifactNotFound
	}

	params := map[string]any{}
	if marketShareRunID != nil {
		params["market_share_run_id"] = *marketShareRunID
	}
	if inputMarketShareArtifactID != nil {
		params["market_share_artifact_id"] = *inputMarketShareArtifactID
	}
	params["source"] = p.Source
	paramsJSON, _ := json.Marshal(params)

	gitSHAParam := any(nil)
	if p.GitSHA != nil && strings.TrimSpace(*p.GitSHA) != "" {
		gitSHAParam = strings.TrimSpace(*p.GitSHA)
	}

	var runID string
	if err := r.pool.QueryRow(ctx, `
		INSERT INTO derived.strategy_generation_runs (
			run_key,
			run_key_uuid,
			name,
			simulated_tournament_id,
			calcutta_id,
			purpose,
			returns_model_key,
			investment_model_key,
			optimizer_key,
			market_share_run_id,
			params_json,
			git_sha
		)
		VALUES (
			$1,
			$2::uuid,
			$3,
			NULL,
			$4::uuid,
			'go_strategy_generation_run_bids',
			'legacy',
			'predicted_market_share',
			$5,
			$6::uuid,
			$7::jsonb,
			$8
		)
		RETURNING id::text
	`, p.RunKey, p.RunKeyUUID, p.Name, p.CalcuttaID, p.OptimizerKey, marketShareRunID, string(paramsJSON), gitSHAParam).Scan(&runID); err != nil {
		return nil, err
	}

	resp := &strategy_runs.CreateRunResult{RunID: runID, RunKey: p.RunKey}

	return resp, nil
}

func (r *StrategyRunsRepository) ListRuns(ctx context.Context, calcuttaID *string, limit, offset int) ([]strategy_runs.RunListItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			id::text,
			run_key,
			name,
			calcutta_id::text,
			simulated_tournament_id::text,
			purpose,
			returns_model_key,
			investment_model_key,
			optimizer_key,
			created_at::text
		FROM derived.strategy_generation_runs
		WHERE deleted_at IS NULL
			AND ($1::uuid IS NULL OR calcutta_id = $1::uuid)
		ORDER BY created_at DESC
		LIMIT $2::int
		OFFSET $3::int
	`, uuidParamOrNil(calcuttaID), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]strategy_runs.RunListItem, 0)
	for rows.Next() {
		var it strategy_runs.RunListItem
		if err := rows.Scan(
			&it.ID,
			&it.RunKey,
			&it.Name,
			&it.CalcuttaID,
			&it.SimulatedTournamentID,
			&it.Purpose,
			&it.ReturnsModelKey,
			&it.InvestmentModelKey,
			&it.OptimizerKey,
			&it.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *StrategyRunsRepository) GetRun(ctx context.Context, id string) (*strategy_runs.RunListItem, error) {
	var it strategy_runs.RunListItem
	if err := r.pool.QueryRow(ctx, `
		SELECT
			id::text,
			run_key,
			name,
			calcutta_id::text,
			simulated_tournament_id::text,
			purpose,
			returns_model_key,
			investment_model_key,
			optimizer_key,
			created_at::text
		FROM derived.strategy_generation_runs
		WHERE id = $1::uuid
			AND deleted_at IS NULL
		LIMIT 1
	`, id).Scan(
		&it.ID,
		&it.RunKey,
		&it.Name,
		&it.CalcuttaID,
		&it.SimulatedTournamentID,
		&it.Purpose,
		&it.ReturnsModelKey,
		&it.InvestmentModelKey,
		&it.OptimizerKey,
		&it.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, strategy_runs.ErrEntryRunNotFound
		}
		return nil, err
	}
	return &it, nil
}

func (r *StrategyRunsRepository) GetEntryArtifact(ctx context.Context, artifactID string) (*strategy_runs.RunArtifactListItem, error) {
	var it strategy_runs.RunArtifactListItem
	var runKey *string
	var summaryText string
	if err := r.pool.QueryRow(ctx, `
		SELECT
			id::text,
			run_id::text,
			run_key::text,
			artifact_kind,
			schema_version,
			storage_uri,
			summary_json::text,
			input_market_share_artifact_id::text,
			input_advancement_artifact_id::text,
			created_at::text,
			updated_at::text
		FROM derived.run_artifacts
		WHERE id = $1::uuid
			AND run_kind = 'strategy_generation'
			AND deleted_at IS NULL
		LIMIT 1
	`, artifactID).Scan(
		&it.ID,
		&it.RunID,
		&runKey,
		&it.ArtifactKind,
		&it.SchemaVersion,
		&it.StorageURI,
		&summaryText,
		&it.InputMarketShareArtifactID,
		&it.InputAdvancementArtifactID,
		&it.CreatedAt,
		&it.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, strategy_runs.ErrEntryArtifactNotFound
		}
		return nil, err
	}
	if runKey != nil && strings.TrimSpace(*runKey) != "" {
		v := strings.TrimSpace(*runKey)
		it.RunKey = &v
	}
	it.SummaryJSON = json.RawMessage([]byte(summaryText))
	return &it, nil
}

func (r *StrategyRunsRepository) ListRunArtifacts(ctx context.Context, runID string) ([]strategy_runs.RunArtifactListItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT
			id::text,
			run_id::text,
			run_key::text,
			artifact_kind,
			schema_version,
			storage_uri,
			summary_json::text,
			input_market_share_artifact_id::text,
			input_advancement_artifact_id::text,
			created_at::text,
			updated_at::text
		FROM derived.run_artifacts
		WHERE run_kind = 'strategy_generation'
			AND run_id = $1::uuid
			AND deleted_at IS NULL
		ORDER BY created_at DESC
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]strategy_runs.RunArtifactListItem, 0)
	for rows.Next() {
		var it strategy_runs.RunArtifactListItem
		var runKey *string
		var summaryText string
		var inputMarketShareArtifactID *string
		var inputAdvancementArtifactID *string
		if err := rows.Scan(
			&it.ID,
			&it.RunID,
			&runKey,
			&it.ArtifactKind,
			&it.SchemaVersion,
			&it.StorageURI,
			&summaryText,
			&inputMarketShareArtifactID,
			&inputAdvancementArtifactID,
			&it.CreatedAt,
			&it.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if runKey != nil && strings.TrimSpace(*runKey) != "" {
			v := strings.TrimSpace(*runKey)
			it.RunKey = &v
		}
		if inputMarketShareArtifactID != nil && strings.TrimSpace(*inputMarketShareArtifactID) != "" {
			v := strings.TrimSpace(*inputMarketShareArtifactID)
			it.InputMarketShareArtifactID = &v
		}
		if inputAdvancementArtifactID != nil && strings.TrimSpace(*inputAdvancementArtifactID) != "" {
			v := strings.TrimSpace(*inputAdvancementArtifactID)
			it.InputAdvancementArtifactID = &v
		}
		it.SummaryJSON = json.RawMessage([]byte(summaryText))
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *StrategyRunsRepository) GetRunArtifact(ctx context.Context, runID, artifactKind string) (*strategy_runs.RunArtifactListItem, error) {
	var it strategy_runs.RunArtifactListItem
	var runKey *string
	var summaryText string
	if err := r.pool.QueryRow(ctx, `
		SELECT
			id::text,
			run_id::text,
			run_key::text,
			artifact_kind,
			schema_version,
			storage_uri,
			summary_json::text,
			input_market_share_artifact_id::text,
			input_advancement_artifact_id::text,
			created_at::text,
			updated_at::text
		FROM derived.run_artifacts
		WHERE run_kind = 'strategy_generation'
			AND run_id = $1::uuid
			AND artifact_kind = $2
			AND deleted_at IS NULL
		LIMIT 1
	`, runID, artifactKind).Scan(
		&it.ID,
		&it.RunID,
		&runKey,
		&it.ArtifactKind,
		&it.SchemaVersion,
		&it.StorageURI,
		&summaryText,
		&it.InputMarketShareArtifactID,
		&it.InputAdvancementArtifactID,
		&it.CreatedAt,
		&it.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, strategy_runs.ErrEntryArtifactNotFound
		}
		return nil, err
	}
	if runKey != nil && strings.TrimSpace(*runKey) != "" {
		v := strings.TrimSpace(*runKey)
		it.RunKey = &v
	}
	it.SummaryJSON = json.RawMessage([]byte(summaryText))
	return &it, nil
}
