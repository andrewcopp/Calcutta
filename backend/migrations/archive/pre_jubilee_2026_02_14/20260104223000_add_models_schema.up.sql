CREATE SCHEMA IF NOT EXISTS models;

CREATE TABLE IF NOT EXISTS models.runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT,
    season_year INTEGER,
    experiment_key TEXT,
    returns_model_key TEXT NOT NULL,
    investment_model_key TEXT NOT NULL,
    allocator_key TEXT NOT NULL,
    params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_models_runs_experiment_key
ON models.runs(experiment_key)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_models_runs_season_year
ON models.runs(season_year)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON models.runs;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON models.runs
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS models.entry_candidates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_id UUID NOT NULL REFERENCES models.runs(id),
    calcutta_id UUID REFERENCES core.calcuttas(id),
    budget_points INTEGER NOT NULL,
    min_teams INTEGER NOT NULL,
    max_teams INTEGER NOT NULL,
    min_bid_points INTEGER NOT NULL,
    max_bid_points INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_models_entry_candidates_run_id
ON models.entry_candidates(run_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_models_entry_candidates_calcutta_id
ON models.entry_candidates(calcutta_id)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON models.entry_candidates;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON models.entry_candidates
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS models.entry_candidate_bids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    entry_candidate_id UUID NOT NULL REFERENCES models.entry_candidates(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES core.teams(id),
    bid_points INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_models_entry_candidate_bids_candidate_team
ON models.entry_candidate_bids(entry_candidate_id, team_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_models_entry_candidate_bids_candidate_id
ON models.entry_candidate_bids(entry_candidate_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_models_entry_candidate_bids_team_id
ON models.entry_candidate_bids(team_id)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON models.entry_candidate_bids;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON models.entry_candidate_bids
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();
