-- Unique constraint: prevent duplicate team assignments within an entry
CREATE UNIQUE INDEX IF NOT EXISTS uq_core_entry_teams_entry_team
    ON core.entry_teams (entry_id, team_id)
    WHERE deleted_at IS NULL;

-- Check constraint: max_bid cannot exceed budget_points
ALTER TABLE core.calcuttas
    ADD CONSTRAINT ck_core_calcuttas_max_bid_le_budget
    CHECK (max_bid <= budget_points);

-- Index for calcutta_evaluation_run_id lookups on simulation_runs
CREATE INDEX IF NOT EXISTS idx_derived_simulation_runs_calcutta_evaluation_run_id
    ON derived.simulation_runs (calcutta_evaluation_run_id)
    WHERE deleted_at IS NULL;

-- Change ON DELETE CASCADE to ON DELETE RESTRICT on simulated_teams FKs to core tables
ALTER TABLE derived.simulated_teams
    DROP CONSTRAINT IF EXISTS simulated_teams_team_id_fkey;
ALTER TABLE derived.simulated_teams
    ADD CONSTRAINT simulated_teams_team_id_fkey
    FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE RESTRICT;

ALTER TABLE derived.simulated_teams
    DROP CONSTRAINT IF EXISTS simulated_teams_tournament_id_fkey;
ALTER TABLE derived.simulated_teams
    ADD CONSTRAINT simulated_teams_tournament_id_fkey
    FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE RESTRICT;
