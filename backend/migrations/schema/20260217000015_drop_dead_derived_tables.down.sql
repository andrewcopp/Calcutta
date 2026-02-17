ALTER TABLE lab.entries ADD COLUMN IF NOT EXISTS state text NOT NULL DEFAULT 'complete';
CREATE INDEX IF NOT EXISTS idx_lab_entries_state ON lab.entries (state) WHERE state <> 'complete';

ALTER TABLE lab.investment_models ADD COLUMN IF NOT EXISTS is_benchmark boolean NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS derived.optimized_entries (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_key text,
    bids_json jsonb,
    market_share_run_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS derived.simulation_runs (
    id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_key text,
    calcutta_id uuid,
    game_outcome_run_id uuid,
    calcutta_evaluation_run_id uuid,
    starting_state_key text,
    excluded_entry_name text,
    n_sims int,
    seed int,
    status text,
    claimed_at timestamptz,
    claimed_by text,
    error_message text,
    simulated_calcutta_id uuid,
    game_outcome_spec_json jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);
