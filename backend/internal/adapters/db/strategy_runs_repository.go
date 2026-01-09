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

	if p.SyntheticCalcuttaID != nil && strings.TrimSpace(*p.SyntheticCalcuttaID) != "" {
		synthID := strings.TrimSpace(*p.SyntheticCalcuttaID)
		tx, err := r.pool.Begin(ctx)
		if err != nil {
			return nil, err
		}
		committed := false
		defer func() {
			if committed {
				return
			}
			_ = tx.Rollback(ctx)
		}()

		var scenarioCalcuttaID string
		var excludedEntryName *string
		var startingStateKey *string
		if err := tx.QueryRow(ctx, `
			SELECT calcutta_id::text, excluded_entry_name, starting_state_key
			FROM derived.synthetic_calcuttas
			WHERE id = $1::uuid
				AND deleted_at IS NULL
			LIMIT 1
		`, synthID).Scan(&scenarioCalcuttaID, &excludedEntryName, &startingStateKey); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, strategy_runs.ErrSyntheticCalcuttaNotFound
			}
			return nil, err
		}
		if scenarioCalcuttaID != p.CalcuttaID {
			return nil, strategy_runs.ErrSyntheticCalcuttaMismatch
		}

		if _, err := tx.Exec(ctx, `
			UPDATE derived.strategy_generation_runs
			SET excluded_entry_name = COALESCE(excluded_entry_name, $2),
				starting_state_key = COALESCE(starting_state_key, $3),
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, runID, excludedEntryName, startingStateKey); err != nil {
			return nil, err
		}

		snapshotID, err := r.createSyntheticCalcuttaSnapshot(ctx, tx, scenarioCalcuttaID, excludedEntryName, runID)
		if err != nil {
			return nil, err
		}

		if _, err := tx.Exec(ctx, `
			UPDATE derived.synthetic_calcuttas
			SET focus_strategy_generation_run_id = $2::uuid,
				calcutta_snapshot_id = $3::uuid,
				focus_entry_name = COALESCE(focus_entry_name, 'Our Strategy'),
				updated_at = NOW()
			WHERE id = $1::uuid
				AND deleted_at IS NULL
		`, synthID, runID, snapshotID); err != nil {
			return nil, err
		}

		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		committed = true

		resp.SyntheticCalcuttaID = &synthID
		resp.CalcuttaSnapshotID = &snapshotID
	}

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

func (r *StrategyRunsRepository) createSyntheticCalcuttaSnapshot(ctx context.Context, tx pgx.Tx, calcuttaID string, excludedEntryName *string, focusStrategyGenerationRunID string) (string, error) {
	var snapshotID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO core.calcutta_snapshots (base_calcutta_id, snapshot_type, description)
		VALUES ($1::uuid, 'synthetic_calcutta', 'Synthetic calcutta snapshot')
		RETURNING id
	`, calcuttaID).Scan(&snapshotID); err != nil {
		return "", err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO core.calcutta_snapshot_payouts (calcutta_snapshot_id, position, amount_cents)
		SELECT $2, position, amount_cents
		FROM core.payouts
		WHERE calcutta_id = $1
			AND deleted_at IS NULL
	`, calcuttaID, snapshotID); err != nil {
		return "", err
	}

	if _, err := tx.Exec(ctx, `
		INSERT INTO core.calcutta_snapshot_scoring_rules (calcutta_snapshot_id, win_index, points_awarded)
		SELECT $2, win_index, points_awarded
		FROM core.calcutta_scoring_rules
		WHERE calcutta_id = $1
			AND deleted_at IS NULL
	`, calcuttaID, snapshotID); err != nil {
		return "", err
	}

	excluded := ""
	if excludedEntryName != nil {
		excluded = *excludedEntryName
	}

	entryRows, err := tx.Query(ctx, `
		SELECT id::text, name
		FROM core.entries
		WHERE calcutta_id = $1::uuid
			AND deleted_at IS NULL
			AND (name != $2 OR $2 = '')
		ORDER BY created_at ASC
	`, calcuttaID, excluded)
	if err != nil {
		return "", err
	}
	defer entryRows.Close()

	type entryRow struct {
		id   string
		name string
	}
	entries := make([]entryRow, 0)
	for entryRows.Next() {
		var id, name string
		if err := entryRows.Scan(&id, &name); err != nil {
			return "", err
		}
		entries = append(entries, entryRow{id: id, name: name})
	}
	if err := entryRows.Err(); err != nil {
		return "", err
	}

	for _, e := range entries {
		var snapshotEntryID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
			VALUES ($1::uuid, $2::uuid, $3, false)
			RETURNING id
		`, snapshotID, e.id, e.name).Scan(&snapshotEntryID); err != nil {
			return "", err
		}

		if _, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			SELECT $1::uuid, team_id, bid_points
			FROM core.entry_teams
			WHERE entry_id = $2::uuid
				AND deleted_at IS NULL
		`, snapshotEntryID, e.id); err != nil {
			return "", err
		}
	}

	if focusStrategyGenerationRunID != "" {
		var snapshotEntryID string
		if err := tx.QueryRow(ctx, `
			INSERT INTO core.calcutta_snapshot_entries (calcutta_snapshot_id, entry_id, display_name, is_synthetic)
			VALUES ($1::uuid, NULL, 'Our Strategy', true)
			RETURNING id
		`, snapshotID).Scan(&snapshotEntryID); err != nil {
			return "", err
		}

		ct, err := tx.Exec(ctx, `
			INSERT INTO core.calcutta_snapshot_entry_teams (calcutta_snapshot_entry_id, team_id, bid_points)
			SELECT $1::uuid, reb.team_id, reb.bid_points
			FROM derived.strategy_generation_run_bids reb
			WHERE reb.strategy_generation_run_id = $2::uuid
				AND reb.deleted_at IS NULL
		`, snapshotEntryID, focusStrategyGenerationRunID)
		if err != nil {
			return "", err
		}
		if ct.RowsAffected() == 0 {
			return "", pgx.ErrNoRows
		}
	}

	return snapshotID, nil
}
