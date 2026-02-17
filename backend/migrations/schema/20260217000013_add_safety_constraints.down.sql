-- Remove unique constraint on entry_teams
DROP INDEX IF EXISTS core.uq_core_entry_teams_entry_team;

-- Remove check constraint on calcuttas
ALTER TABLE core.calcuttas
    DROP CONSTRAINT IF EXISTS ck_core_calcuttas_max_bid_le_budget;

-- Remove simulation_runs index
DROP INDEX IF EXISTS derived.idx_derived_simulation_runs_calcutta_evaluation_run_id;

-- Restore ON DELETE CASCADE on simulated_teams FKs
ALTER TABLE derived.simulated_teams
    DROP CONSTRAINT IF EXISTS simulated_teams_team_id_fkey;
ALTER TABLE derived.simulated_teams
    ADD CONSTRAINT simulated_teams_team_id_fkey
    FOREIGN KEY (team_id) REFERENCES core.teams(id) ON DELETE CASCADE;

ALTER TABLE derived.simulated_teams
    DROP CONSTRAINT IF EXISTS simulated_teams_tournament_id_fkey;
ALTER TABLE derived.simulated_teams
    ADD CONSTRAINT simulated_teams_tournament_id_fkey
    FOREIGN KEY (tournament_id) REFERENCES core.tournaments(id) ON DELETE CASCADE;
