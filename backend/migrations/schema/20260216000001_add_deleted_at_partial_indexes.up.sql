-- Partial indexes on deleted_at IS NULL for hot-path core tables.
-- Every query in the codebase filters WHERE deleted_at IS NULL; these indexes prevent full table scans.

CREATE INDEX IF NOT EXISTS idx_users_active ON core.users (id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_calcuttas_active ON core.calcuttas (id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_entries_active ON core.entries (id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_entry_teams_active ON core.entry_teams (id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_teams_active ON core.teams (id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_payouts_active ON core.payouts (id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_calcutta_scoring_rules_active ON core.calcutta_scoring_rules (id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_calcutta_invitations_active ON core.calcutta_invitations (id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tournaments_active ON core.tournaments (id) WHERE deleted_at IS NULL;
