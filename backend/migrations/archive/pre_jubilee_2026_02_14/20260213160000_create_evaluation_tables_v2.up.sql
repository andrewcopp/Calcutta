-- Create new evaluation tables with better naming conventions
-- Replaces: archive.algorithms -> derived.prediction_models
-- Replaces: archive.strategy_generation_runs + archive.strategy_generation_run_bids -> derived.optimized_entries

-- =============================================================================
-- 1. Create derived.prediction_models (replaces archive.algorithms)
-- =============================================================================
CREATE TABLE IF NOT EXISTS derived.prediction_models (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    kind TEXT NOT NULL,                    -- 'game_outcomes', 'market_share'
    name TEXT NOT NULL,                    -- 'kenpom', 'kenpom-v1-sigma11-go'
    params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Unique constraint on (kind, name) for active records
CREATE UNIQUE INDEX IF NOT EXISTS prediction_models_kind_name_uniq
    ON derived.prediction_models (kind, name)
    WHERE deleted_at IS NULL;

-- Index for listing by kind
CREATE INDEX IF NOT EXISTS prediction_models_kind_idx
    ON derived.prediction_models (kind)
    WHERE deleted_at IS NULL;

-- Add updated_at trigger
DROP TRIGGER IF EXISTS set_updated_at ON derived.prediction_models;
CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON derived.prediction_models
    FOR EACH ROW
    EXECUTE FUNCTION core.set_updated_at();

-- =============================================================================
-- 2. Create derived.optimized_entries (replaces strategy_generation_runs + bids)
-- =============================================================================
CREATE TABLE IF NOT EXISTS derived.optimized_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),

    -- Identity
    run_key TEXT,                          -- For external references/idempotency
    name TEXT,                             -- Display name

    -- Calcutta context
    calcutta_id UUID NOT NULL REFERENCES core.calcuttas(id),

    -- Upstream references (lineage)
    simulated_tournament_id UUID,
    game_outcome_run_id UUID REFERENCES derived.game_outcome_runs(id),
    market_share_run_id UUID,              -- No FK since market_share_runs is in archive schema

    -- Optimizer configuration
    optimizer_kind TEXT NOT NULL DEFAULT 'minlp',
    optimizer_params_json JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- The bids: [{team_id, bid_points, expected_roi}]
    bids_json JSONB NOT NULL,

    -- Context
    purpose TEXT,                          -- 'evaluation', 'what_if', 'lab_entries_generation'
    excluded_entry_name TEXT,
    starting_state_key TEXT,

    -- Legacy compatibility fields (to support existing queries during migration)
    returns_model_key TEXT,
    investment_model_key TEXT,

    -- Metadata
    git_sha TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

-- Unique constraint on run_key for active records
CREATE UNIQUE INDEX IF NOT EXISTS optimized_entries_run_key_uniq
    ON derived.optimized_entries (run_key)
    WHERE run_key IS NOT NULL AND deleted_at IS NULL;

-- Index for lookups by calcutta
CREATE INDEX IF NOT EXISTS optimized_entries_calcutta_id_idx
    ON derived.optimized_entries (calcutta_id)
    WHERE deleted_at IS NULL;

-- Index for lookups by game_outcome_run
CREATE INDEX IF NOT EXISTS optimized_entries_game_outcome_run_id_idx
    ON derived.optimized_entries (game_outcome_run_id)
    WHERE deleted_at IS NULL;

-- Add updated_at trigger
DROP TRIGGER IF EXISTS set_updated_at ON derived.optimized_entries;
CREATE TRIGGER set_updated_at
    BEFORE UPDATE ON derived.optimized_entries
    FOR EACH ROW
    EXECUTE FUNCTION core.set_updated_at();

-- =============================================================================
-- 3. Add prediction_model_id to game_outcome_runs (optional FK to new table)
-- =============================================================================
ALTER TABLE derived.game_outcome_runs
    ADD COLUMN IF NOT EXISTS prediction_model_id UUID REFERENCES derived.prediction_models(id);

-- =============================================================================
-- 4. Migrate data from existing tables (in archive schema)
-- =============================================================================

-- 4a. Migrate algorithms to prediction_models (preserving IDs)
INSERT INTO derived.prediction_models (id, kind, name, params_json, description, created_at, updated_at, deleted_at)
SELECT
    id,
    kind,
    name,
    COALESCE(params_json, '{}'::jsonb),
    description,
    created_at,
    updated_at,
    deleted_at
FROM archive.algorithms
ON CONFLICT (id) DO NOTHING;

-- 4b. Update game_outcome_runs to reference prediction_models
UPDATE derived.game_outcome_runs gor
SET prediction_model_id = gor.algorithm_id
WHERE gor.algorithm_id IS NOT NULL
  AND gor.prediction_model_id IS NULL
  AND EXISTS (
    SELECT 1 FROM derived.prediction_models pm WHERE pm.id = gor.algorithm_id
  );

-- 4c. Migrate strategy_generation_runs + _bids to optimized_entries
INSERT INTO derived.optimized_entries (
    id,
    run_key,
    name,
    calcutta_id,
    simulated_tournament_id,
    game_outcome_run_id,
    market_share_run_id,
    optimizer_kind,
    optimizer_params_json,
    bids_json,
    purpose,
    excluded_entry_name,
    starting_state_key,
    returns_model_key,
    investment_model_key,
    git_sha,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    sgr.id,
    sgr.run_key,
    sgr.name,
    sgr.calcutta_id,
    sgr.simulated_tournament_id,
    sgr.game_outcome_run_id,
    sgr.market_share_run_id,
    COALESCE(NULLIF(sgr.optimizer_key, ''), 'minlp'),
    COALESCE(sgr.params_json, '{}'::jsonb),
    -- Aggregate bids into JSON array
    COALESCE(
        (
            SELECT jsonb_agg(
                jsonb_build_object(
                    'team_id', b.team_id::text,
                    'bid_points', b.bid_points,
                    'expected_roi', COALESCE(b.expected_roi, 0)
                )
                ORDER BY b.bid_points DESC
            )
            FROM archive.strategy_generation_run_bids b
            WHERE b.strategy_generation_run_id = sgr.id
              AND b.deleted_at IS NULL
        ),
        '[]'::jsonb
    ),
    sgr.purpose,
    sgr.excluded_entry_name,
    sgr.starting_state_key,
    sgr.returns_model_key,
    sgr.investment_model_key,
    sgr.git_sha,
    sgr.created_at,
    sgr.updated_at,
    sgr.deleted_at
FROM archive.strategy_generation_runs sgr
ON CONFLICT (id) DO NOTHING;

-- =============================================================================
-- 5. Create views for backward compatibility during migration
-- =============================================================================

-- View: derived.v_strategy_generation_run_bids
-- Extracts bids from optimized_entries.bids_json for queries that need per-row access
CREATE OR REPLACE VIEW derived.v_strategy_generation_run_bids AS
SELECT
    oe.id AS strategy_generation_run_id,
    oe.run_key AS run_id,
    (bid->>'team_id')::uuid AS team_id,
    (bid->>'bid_points')::int AS bid_points,
    COALESCE((bid->>'expected_roi')::double precision, 0) AS expected_roi,
    oe.created_at,
    oe.deleted_at
FROM derived.optimized_entries oe,
     jsonb_array_elements(oe.bids_json) AS bid;

-- View: derived.v_algorithms (maps to prediction_models for compatibility)
CREATE OR REPLACE VIEW derived.v_algorithms AS
SELECT
    id,
    kind,
    name,
    description,
    params_json,
    created_at,
    updated_at,
    deleted_at
FROM derived.prediction_models;
