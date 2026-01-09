CREATE SCHEMA IF NOT EXISTS derived;

CREATE TABLE IF NOT EXISTS derived.simulated_calcuttas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name TEXT NOT NULL,
    description TEXT,
    tournament_id UUID NOT NULL REFERENCES core.tournaments(id),
    base_calcutta_id UUID REFERENCES core.calcuttas(id),
    starting_state_key TEXT NOT NULL DEFAULT 'post_first_four',
    excluded_entry_name TEXT,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_derived_simulated_calcuttas_starting_state_key
        CHECK (starting_state_key IN ('post_first_four', 'current'))
);

CREATE INDEX IF NOT EXISTS idx_derived_simulated_calcuttas_tournament_id
ON derived.simulated_calcuttas(tournament_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulated_calcuttas_base_calcutta_id
ON derived.simulated_calcuttas(base_calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulated_calcuttas_created_at
ON derived.simulated_calcuttas(created_at DESC)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.simulated_calcuttas;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.simulated_calcuttas
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS derived.simulated_calcutta_payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    simulated_calcutta_id UUID NOT NULL REFERENCES derived.simulated_calcuttas(id) ON DELETE CASCADE,
    position INT NOT NULL,
    amount_cents INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_simulated_calcutta_payouts_simulated_position
ON derived.simulated_calcutta_payouts(simulated_calcutta_id, position)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulated_calcutta_payouts_simulated_calcutta_id
ON derived.simulated_calcutta_payouts(simulated_calcutta_id)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.simulated_calcutta_payouts;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.simulated_calcutta_payouts
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS derived.simulated_calcutta_scoring_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    simulated_calcutta_id UUID NOT NULL REFERENCES derived.simulated_calcuttas(id) ON DELETE CASCADE,
    win_index INT NOT NULL,
    points_awarded INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_simulated_calcutta_scoring_rules_simulated_win_index
ON derived.simulated_calcutta_scoring_rules(simulated_calcutta_id, win_index)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulated_calcutta_scoring_rules_simulated_calcutta_id
ON derived.simulated_calcutta_scoring_rules(simulated_calcutta_id)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.simulated_calcutta_scoring_rules;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.simulated_calcutta_scoring_rules
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS derived.simulated_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    simulated_calcutta_id UUID NOT NULL REFERENCES derived.simulated_calcuttas(id) ON DELETE CASCADE,
    display_name TEXT NOT NULL,
    source_kind TEXT NOT NULL,
    source_entry_id UUID REFERENCES core.entries(id),
    source_candidate_id UUID REFERENCES derived.candidates(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_derived_simulated_entries_source_kind
        CHECK (source_kind IN ('manual', 'from_real_entry', 'from_candidate'))
);

CREATE INDEX IF NOT EXISTS idx_derived_simulated_entries_simulated_calcutta_id
ON derived.simulated_entries(simulated_calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulated_entries_created_at
ON derived.simulated_entries(created_at DESC)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.simulated_entries;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.simulated_entries
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS derived.simulated_entry_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    simulated_entry_id UUID NOT NULL REFERENCES derived.simulated_entries(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES core.teams(id),
    bid_points INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_simulated_entry_teams_entry_team
ON derived.simulated_entry_teams(simulated_entry_id, team_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulated_entry_teams_simulated_entry_id
ON derived.simulated_entry_teams(simulated_entry_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_simulated_entry_teams_team_id
ON derived.simulated_entry_teams(team_id)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.simulated_entry_teams;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.simulated_entry_teams
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

ALTER TABLE derived.simulated_calcuttas
    ADD COLUMN IF NOT EXISTS highlighted_simulated_entry_id UUID REFERENCES derived.simulated_entries(id);

CREATE INDEX IF NOT EXISTS idx_derived_simulated_calcuttas_highlighted_simulated_entry_id
ON derived.simulated_calcuttas(highlighted_simulated_entry_id)
WHERE deleted_at IS NULL;
