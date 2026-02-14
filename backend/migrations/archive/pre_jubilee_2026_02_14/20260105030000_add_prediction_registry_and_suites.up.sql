CREATE SCHEMA IF NOT EXISTS derived;

CREATE TABLE IF NOT EXISTS derived.algorithms (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    kind TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_algorithms_kind_name
ON derived.algorithms(kind, name)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.algorithms;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.algorithms
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS derived.game_outcome_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    algorithm_id UUID NOT NULL REFERENCES derived.algorithms(id),
    tournament_id UUID NOT NULL REFERENCES core.tournaments(id),
    params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    git_sha TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_derived_game_outcome_runs_tournament_id
ON derived.game_outcome_runs(tournament_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_game_outcome_runs_algorithm_id
ON derived.game_outcome_runs(algorithm_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_game_outcome_runs_created_at
ON derived.game_outcome_runs(created_at DESC)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.game_outcome_runs;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.game_outcome_runs
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS derived.market_share_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    algorithm_id UUID NOT NULL REFERENCES derived.algorithms(id),
    calcutta_id UUID NOT NULL REFERENCES core.calcuttas(id),
    calcutta_group_id UUID,
    params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    git_sha TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_derived_market_share_runs_calcutta_id
ON derived.market_share_runs(calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_market_share_runs_algorithm_id
ON derived.market_share_runs(algorithm_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_market_share_runs_created_at
ON derived.market_share_runs(created_at DESC)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.market_share_runs;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.market_share_runs
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS derived.suites (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    description TEXT,
    game_outcomes_algorithm_id UUID NOT NULL REFERENCES derived.algorithms(id),
    market_share_algorithm_id UUID NOT NULL REFERENCES derived.algorithms(id),
    optimizer_key TEXT NOT NULL,
    n_sims INTEGER NOT NULL,
    seed INTEGER NOT NULL,
    params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_suites_name
ON derived.suites(name)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.suites;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.suites
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS derived.suite_calcutta_evaluations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    suite_id UUID NOT NULL REFERENCES derived.suites(id),
    calcutta_id UUID NOT NULL REFERENCES core.calcuttas(id),
    game_outcome_run_id UUID REFERENCES derived.game_outcome_runs(id),
    market_share_run_id UUID REFERENCES derived.market_share_runs(id),
    strategy_generation_run_id UUID REFERENCES derived.strategy_generation_runs(id),
    calcutta_evaluation_run_id UUID REFERENCES derived.calcutta_evaluation_runs(id),
    status TEXT NOT NULL DEFAULT 'queued',
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_derived_suite_calcutta_evaluations_status
        CHECK (status IN ('queued', 'running', 'succeeded', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_derived_suite_calcutta_evaluations_suite_id
ON derived.suite_calcutta_evaluations(suite_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_suite_calcutta_evaluations_calcutta_id
ON derived.suite_calcutta_evaluations(calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_suite_calcutta_evaluations_created_at
ON derived.suite_calcutta_evaluations(created_at DESC)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.suite_calcutta_evaluations;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.suite_calcutta_evaluations
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

ALTER TABLE IF EXISTS derived.predicted_game_outcomes
    ADD COLUMN IF NOT EXISTS run_id UUID REFERENCES derived.game_outcome_runs(id);

ALTER TABLE IF EXISTS derived.predicted_market_share
    ADD COLUMN IF NOT EXISTS run_id UUID REFERENCES derived.market_share_runs(id);

ALTER TABLE IF EXISTS derived.predicted_game_outcomes
    DROP CONSTRAINT IF EXISTS silver_predicted_game_outcomes_unique_matchup;

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_predicted_game_outcomes_legacy_matchup
ON derived.predicted_game_outcomes(tournament_id, game_id, team1_id, team2_id)
WHERE run_id IS NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_predicted_game_outcomes_run_matchup
ON derived.predicted_game_outcomes(run_id, game_id, team1_id, team2_id)
WHERE run_id IS NOT NULL AND deleted_at IS NULL;

DROP INDEX IF EXISTS silver_predicted_market_share_calcutta_team_uniq;
DROP INDEX IF EXISTS silver_predicted_market_share_tournament_team_uniq;

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_predicted_market_share_legacy_calcutta_team
ON derived.predicted_market_share(calcutta_id, team_id)
WHERE calcutta_id IS NOT NULL AND run_id IS NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_predicted_market_share_legacy_tournament_team
ON derived.predicted_market_share(tournament_id, team_id)
WHERE tournament_id IS NOT NULL AND run_id IS NULL AND deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_predicted_market_share_run_team
ON derived.predicted_market_share(run_id, team_id)
WHERE run_id IS NOT NULL AND deleted_at IS NULL;
