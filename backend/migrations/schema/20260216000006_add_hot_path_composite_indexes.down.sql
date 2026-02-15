-- Rollback: restore original non-partial indexes, drop new ones.

DROP INDEX IF EXISTS core.idx_calcutta_invitations_calcutta_user_active;
DROP INDEX IF EXISTS core.idx_calcutta_invitations_calcutta_id_active;
DROP INDEX IF EXISTS core.idx_entry_teams_team_id_active;
DROP INDEX IF EXISTS core.idx_entry_teams_entry_id_active;
DROP INDEX IF EXISTS core.idx_entries_calcutta_id_active;

-- Restore original non-partial indexes
CREATE INDEX IF NOT EXISTS idx_core_entries_calcutta_id ON core.entries USING btree (calcutta_id);
CREATE INDEX IF NOT EXISTS idx_core_entry_teams_entry_id ON core.entry_teams USING btree (entry_id);
CREATE INDEX IF NOT EXISTS idx_core_entry_teams_team_id ON core.entry_teams USING btree (team_id);
