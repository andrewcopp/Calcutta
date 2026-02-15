-- Add immutable calcutta snapshots as core input artifacts.

CREATE TABLE core.calcutta_snapshots (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    base_calcutta_id UUID NOT NULL REFERENCES core.calcuttas(id),
    snapshot_type TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE INDEX idx_core_calcutta_snapshots_base_calcutta_id ON core.calcutta_snapshots(base_calcutta_id);

CREATE TABLE core.calcutta_snapshot_entries (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calcutta_snapshot_id UUID NOT NULL REFERENCES core.calcutta_snapshots(id) ON DELETE CASCADE,
    entry_id UUID REFERENCES core.entries(id),
    display_name TEXT NOT NULL,
    is_synthetic BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_core_calcutta_snapshot_entries_snapshot_display_name UNIQUE (calcutta_snapshot_id, display_name)
);

CREATE INDEX idx_core_calcutta_snapshot_entries_snapshot_id ON core.calcutta_snapshot_entries(calcutta_snapshot_id);

CREATE TABLE core.calcutta_snapshot_entry_teams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calcutta_snapshot_entry_id UUID NOT NULL REFERENCES core.calcutta_snapshot_entries(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES core.teams(id),
    bid_points INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_core_calcutta_snapshot_entry_teams_entry_team UNIQUE (calcutta_snapshot_entry_id, team_id)
);

CREATE INDEX idx_core_calcutta_snapshot_entry_teams_entry_id ON core.calcutta_snapshot_entry_teams(calcutta_snapshot_entry_id);

CREATE TABLE core.calcutta_snapshot_payouts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calcutta_snapshot_id UUID NOT NULL REFERENCES core.calcutta_snapshots(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    amount_cents INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_core_calcutta_snapshot_payouts_snapshot_position UNIQUE (calcutta_snapshot_id, position)
);

CREATE INDEX idx_core_calcutta_snapshot_payouts_snapshot_id ON core.calcutta_snapshot_payouts(calcutta_snapshot_id);

CREATE TABLE core.calcutta_snapshot_scoring_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calcutta_snapshot_id UUID NOT NULL REFERENCES core.calcutta_snapshots(id) ON DELETE CASCADE,
    win_index INTEGER NOT NULL,
    points_awarded INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT uq_core_calcutta_snapshot_scoring_rules_snapshot_win_index UNIQUE (calcutta_snapshot_id, win_index)
);

CREATE INDEX idx_core_calcutta_snapshot_scoring_rules_snapshot_id ON core.calcutta_snapshot_scoring_rules(calcutta_snapshot_id);

-- Wire calcutta evaluation runs to snapshots
ALTER TABLE analytics.calcutta_evaluation_runs
    ADD CONSTRAINT fk_analytics_calcutta_evaluation_runs_calcutta_snapshot_id
    FOREIGN KEY (calcutta_snapshot_id)
    REFERENCES core.calcutta_snapshots(id);

-- updated_at triggers (core)
DROP TRIGGER IF EXISTS trg_core_calcutta_snapshots_updated_at ON core.calcutta_snapshots;
CREATE TRIGGER trg_core_calcutta_snapshots_updated_at
BEFORE UPDATE ON core.calcutta_snapshots
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_calcutta_snapshot_entries_updated_at ON core.calcutta_snapshot_entries;
CREATE TRIGGER trg_core_calcutta_snapshot_entries_updated_at
BEFORE UPDATE ON core.calcutta_snapshot_entries
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_calcutta_snapshot_entry_teams_updated_at ON core.calcutta_snapshot_entry_teams;
CREATE TRIGGER trg_core_calcutta_snapshot_entry_teams_updated_at
BEFORE UPDATE ON core.calcutta_snapshot_entry_teams
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_calcutta_snapshot_payouts_updated_at ON core.calcutta_snapshot_payouts;
CREATE TRIGGER trg_core_calcutta_snapshot_payouts_updated_at
BEFORE UPDATE ON core.calcutta_snapshot_payouts
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

DROP TRIGGER IF EXISTS trg_core_calcutta_snapshot_scoring_rules_updated_at ON core.calcutta_snapshot_scoring_rules;
CREATE TRIGGER trg_core_calcutta_snapshot_scoring_rules_updated_at
BEFORE UPDATE ON core.calcutta_snapshot_scoring_rules
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();
