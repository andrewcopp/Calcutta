-- Add composite partial indexes for hot FK-lookup query paths.
-- These replace non-partial btree indexes with partial variants that
-- match the WHERE deleted_at IS NULL filter present in every query.

-- entries looked up by calcutta_id (every calcutta detail page)
DROP INDEX IF EXISTS core.idx_core_entries_calcutta_id;
CREATE INDEX idx_entries_calcutta_id_active ON core.entries (calcutta_id) WHERE deleted_at IS NULL;

-- entry_teams looked up by entry_id (every portfolio/bidding view)
DROP INDEX IF EXISTS core.idx_core_entry_teams_entry_id;
CREATE INDEX idx_entry_teams_entry_id_active ON core.entry_teams (entry_id) WHERE deleted_at IS NULL;

-- entry_teams looked up by team_id (portfolio teams view, scoring)
DROP INDEX IF EXISTS core.idx_core_entry_teams_team_id;
CREATE INDEX idx_entry_teams_team_id_active ON core.entry_teams (team_id) WHERE deleted_at IS NULL;

-- calcutta_invitations looked up by calcutta_id (invitation list page)
CREATE INDEX IF NOT EXISTS idx_calcutta_invitations_calcutta_id_active
ON core.calcutta_invitations (calcutta_id) WHERE deleted_at IS NULL;

-- calcutta_invitations looked up by (calcutta_id, user_id) (invitation check)
CREATE INDEX IF NOT EXISTS idx_calcutta_invitations_calcutta_user_active
ON core.calcutta_invitations (calcutta_id, user_id) WHERE deleted_at IS NULL;
