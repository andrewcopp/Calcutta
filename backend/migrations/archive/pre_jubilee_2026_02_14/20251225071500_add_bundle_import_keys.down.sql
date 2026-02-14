DROP INDEX IF EXISTS uq_calcutta_entry_teams_entry_team;
DROP INDEX IF EXISTS uq_calcutta_rounds_calcutta_round;
DROP INDEX IF EXISTS uq_tournament_teams_tournament_school;

DROP INDEX IF EXISTS uq_calcutta_entries_calcutta_key;
ALTER TABLE calcutta_entries
    DROP COLUMN IF EXISTS key;

DROP INDEX IF EXISTS uq_calcuttas_tournament_key;
ALTER TABLE calcuttas
    DROP COLUMN IF EXISTS key;
