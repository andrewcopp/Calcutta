-- Add run-lineage spine tables for tournament simulations and calcutta evaluations.

-- Tournament state snapshots (per-team wins/byes form)
CREATE TABLE IF NOT EXISTS analytics.tournament_state_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES core.tournaments(id),
    source TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_analytics_tournament_state_snapshots_tournament_id
ON analytics.tournament_state_snapshots(tournament_id);

CREATE TABLE IF NOT EXISTS analytics.tournament_state_snapshot_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_state_snapshot_id UUID NOT NULL REFERENCES analytics.tournament_state_snapshots(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES core.teams(id),
    wins INTEGER NOT NULL,
    byes INTEGER NOT NULL,
    eliminated BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_analytics_tournament_state_snapshot_teams_snapshot_team UNIQUE (tournament_state_snapshot_id, team_id)
);

CREATE INDEX IF NOT EXISTS idx_analytics_tournament_state_snapshot_teams_snapshot_id
ON analytics.tournament_state_snapshot_teams(tournament_state_snapshot_id);

CREATE INDEX IF NOT EXISTS idx_analytics_tournament_state_snapshot_teams_team_id
ON analytics.tournament_state_snapshot_teams(team_id);

-- Tournament simulation batches
CREATE TABLE IF NOT EXISTS analytics.tournament_simulation_batches (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_id UUID NOT NULL REFERENCES core.tournaments(id),
    tournament_state_snapshot_id UUID NOT NULL REFERENCES analytics.tournament_state_snapshots(id),
    n_sims INTEGER NOT NULL,
    seed INTEGER NOT NULL,
    probability_source_key TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_analytics_tournament_simulation_batches_tournament_id
ON analytics.tournament_simulation_batches(tournament_id);

CREATE INDEX IF NOT EXISTS idx_analytics_tournament_simulation_batches_snapshot_id
ON analytics.tournament_simulation_batches(tournament_state_snapshot_id);

-- Calcutta evaluation runs
CREATE TABLE IF NOT EXISTS analytics.calcutta_evaluation_runs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tournament_simulation_batch_id UUID NOT NULL REFERENCES analytics.tournament_simulation_batches(id),
    calcutta_snapshot_id UUID,
    purpose TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_analytics_calcutta_evaluation_runs_batch_id
ON analytics.calcutta_evaluation_runs(tournament_simulation_batch_id);

CREATE INDEX IF NOT EXISTS idx_analytics_calcutta_evaluation_runs_calcutta_snapshot_id
ON analytics.calcutta_evaluation_runs(calcutta_snapshot_id);

-- Add nullable lineage FKs to existing cached output tables
ALTER TABLE analytics.simulated_tournaments
    ADD COLUMN IF NOT EXISTS tournament_simulation_batch_id UUID REFERENCES analytics.tournament_simulation_batches(id);

CREATE INDEX IF NOT EXISTS idx_analytics_simulated_tournaments_batch_id
ON analytics.simulated_tournaments(tournament_simulation_batch_id);

ALTER TABLE analytics.entry_simulation_outcomes
    ADD COLUMN IF NOT EXISTS calcutta_evaluation_run_id UUID REFERENCES analytics.calcutta_evaluation_runs(id);

CREATE INDEX IF NOT EXISTS idx_analytics_entry_simulation_outcomes_eval_run_id
ON analytics.entry_simulation_outcomes(calcutta_evaluation_run_id);

ALTER TABLE analytics.entry_performance
    ADD COLUMN IF NOT EXISTS calcutta_evaluation_run_id UUID REFERENCES analytics.calcutta_evaluation_runs(id);

CREATE INDEX IF NOT EXISTS idx_analytics_entry_performance_eval_run_id
ON analytics.entry_performance(calcutta_evaluation_run_id);

-- updated_at triggers
DROP TRIGGER IF EXISTS set_updated_at ON analytics.tournament_state_snapshots;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON analytics.tournament_state_snapshots
FOR EACH ROW
EXECUTE FUNCTION public.set_updated_at();

DROP TRIGGER IF EXISTS set_updated_at ON analytics.tournament_state_snapshot_teams;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON analytics.tournament_state_snapshot_teams
FOR EACH ROW
EXECUTE FUNCTION public.set_updated_at();

DROP TRIGGER IF EXISTS set_updated_at ON analytics.tournament_simulation_batches;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON analytics.tournament_simulation_batches
FOR EACH ROW
EXECUTE FUNCTION public.set_updated_at();

DROP TRIGGER IF EXISTS set_updated_at ON analytics.calcutta_evaluation_runs;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON analytics.calcutta_evaluation_runs
FOR EACH ROW
EXECUTE FUNCTION public.set_updated_at();
