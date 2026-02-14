ALTER TABLE calcuttas
    ADD COLUMN key TEXT;

CREATE UNIQUE INDEX uq_calcuttas_tournament_key ON calcuttas (tournament_id, key)
WHERE deleted_at IS NULL AND key IS NOT NULL;

ALTER TABLE calcutta_entries
    ADD COLUMN key TEXT;

CREATE UNIQUE INDEX uq_calcutta_entries_calcutta_key ON calcutta_entries (calcutta_id, key)
WHERE deleted_at IS NULL AND key IS NOT NULL;

CREATE UNIQUE INDEX uq_tournament_teams_tournament_school ON tournament_teams (tournament_id, school_id)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_calcutta_rounds_calcutta_round ON calcutta_rounds (calcutta_id, round)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX uq_calcutta_entry_teams_entry_team ON calcutta_entry_teams (entry_id, team_id)
WHERE deleted_at IS NULL;
