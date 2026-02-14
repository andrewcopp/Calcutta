-- Lab Schema: Simplified R&D environment for investment model iteration
-- Replaces over-complicated derived.algorithms, derived.candidates, derived.suite_* tables

CREATE SCHEMA IF NOT EXISTS lab;

--------------------------------------------------------------------------------
-- Investment Models: What investment prediction approach are we testing?
--------------------------------------------------------------------------------
CREATE TABLE lab.investment_models (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    kind TEXT NOT NULL,  -- ridge, random_forest, xgboost, oracle, naive_ev, etc.
    params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX uq_lab_investment_models_name
ON lab.investment_models(name)
WHERE deleted_at IS NULL;

CREATE INDEX idx_lab_investment_models_kind
ON lab.investment_models(kind)
WHERE deleted_at IS NULL;

CREATE INDEX idx_lab_investment_models_created_at
ON lab.investment_models(created_at DESC)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON lab.investment_models;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON lab.investment_models
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

--------------------------------------------------------------------------------
-- Entries: What entry did an investment model produce for a specific calcutta?
--------------------------------------------------------------------------------
CREATE TABLE lab.entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    investment_model_id UUID NOT NULL REFERENCES lab.investment_models(id),
    calcutta_id UUID NOT NULL REFERENCES core.calcuttas(id),

    -- Fixed components (tracked for reproducibility)
    game_outcome_kind TEXT NOT NULL DEFAULT 'kenpom',
    game_outcome_params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    optimizer_kind TEXT NOT NULL DEFAULT 'minlp',
    optimizer_params_json JSONB NOT NULL DEFAULT '{}'::jsonb,

    -- State at which entry was generated
    starting_state_key TEXT NOT NULL DEFAULT 'post_first_four',

    -- The actual bids: [{team_id, bid_points, expected_roi}]
    bids_json JSONB NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT ck_lab_entries_starting_state_key
        CHECK (starting_state_key IN ('pre_tournament', 'post_first_four', 'current'))
);

CREATE INDEX idx_lab_entries_investment_model_id
ON lab.entries(investment_model_id)
WHERE deleted_at IS NULL;

CREATE INDEX idx_lab_entries_calcutta_id
ON lab.entries(calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX idx_lab_entries_created_at
ON lab.entries(created_at DESC)
WHERE deleted_at IS NULL;

-- One entry per (model, calcutta, starting_state) combo - allows re-running with different states
CREATE UNIQUE INDEX uq_lab_entries_model_calcutta_state
ON lab.entries(investment_model_id, calcutta_id, starting_state_key)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON lab.entries;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON lab.entries
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

--------------------------------------------------------------------------------
-- Evaluations: How did the entry perform in simulation?
--------------------------------------------------------------------------------
CREATE TABLE lab.evaluations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entry_id UUID NOT NULL REFERENCES lab.entries(id),

    -- Simulation params
    n_sims INT NOT NULL,
    seed INT NOT NULL,

    -- Results (THE metric that matters)
    mean_normalized_payout DOUBLE PRECISION,
    median_normalized_payout DOUBLE PRECISION,
    p_top1 DOUBLE PRECISION,
    p_in_money DOUBLE PRECISION,
    our_rank INT,

    -- Link to simulation infrastructure (optional, for debugging)
    simulated_calcutta_id UUID REFERENCES derived.simulated_calcuttas(id),

    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,

    CONSTRAINT ck_lab_evaluations_n_sims CHECK (n_sims > 0),
    CONSTRAINT ck_lab_evaluations_seed CHECK (seed <> 0)
);

-- Prevent duplicate evaluations with same params
CREATE UNIQUE INDEX uq_lab_evaluations_entry_sims_seed
ON lab.evaluations(entry_id, n_sims, seed)
WHERE deleted_at IS NULL;

CREATE INDEX idx_lab_evaluations_entry_id
ON lab.evaluations(entry_id)
WHERE deleted_at IS NULL;

CREATE INDEX idx_lab_evaluations_created_at
ON lab.evaluations(created_at DESC)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON lab.evaluations;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON lab.evaluations
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

--------------------------------------------------------------------------------
-- Convenience views for common queries
--------------------------------------------------------------------------------

-- Model leaderboard: Compare all investment models by mean normalized payout
CREATE OR REPLACE VIEW lab.model_leaderboard AS
SELECT
    im.id AS investment_model_id,
    im.name AS model_name,
    im.kind AS model_kind,
    COUNT(DISTINCT e.id) AS n_entries,
    COUNT(ev.id) AS n_evaluations,
    AVG(ev.mean_normalized_payout) AS avg_mean_payout,
    AVG(ev.median_normalized_payout) AS avg_median_payout,
    AVG(ev.p_top1) AS avg_p_top1,
    AVG(ev.p_in_money) AS avg_p_in_money,
    MIN(ev.created_at) AS first_eval_at,
    MAX(ev.created_at) AS last_eval_at
FROM lab.investment_models im
LEFT JOIN lab.entries e ON e.investment_model_id = im.id AND e.deleted_at IS NULL
LEFT JOIN lab.evaluations ev ON ev.entry_id = e.id AND ev.deleted_at IS NULL
WHERE im.deleted_at IS NULL
GROUP BY im.id, im.name, im.kind
ORDER BY avg_mean_payout DESC NULLS LAST;

-- Entry detail: Full evaluation results per entry
CREATE OR REPLACE VIEW lab.entry_evaluations AS
SELECT
    e.id AS entry_id,
    im.name AS model_name,
    im.kind AS model_kind,
    c.name AS calcutta_name,
    e.starting_state_key,
    e.game_outcome_kind,
    e.optimizer_kind,
    ev.n_sims,
    ev.seed,
    ev.mean_normalized_payout,
    ev.median_normalized_payout,
    ev.p_top1,
    ev.p_in_money,
    ev.our_rank,
    ev.created_at AS eval_created_at
FROM lab.entries e
JOIN lab.investment_models im ON im.id = e.investment_model_id
JOIN core.calcuttas c ON c.id = e.calcutta_id
LEFT JOIN lab.evaluations ev ON ev.entry_id = e.id AND ev.deleted_at IS NULL
WHERE e.deleted_at IS NULL
  AND im.deleted_at IS NULL;
