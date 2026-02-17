-- Re-create derived.simulated_calcuttas and children
CREATE TABLE IF NOT EXISTS derived.simulated_calcuttas (
    id uuid DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    calcutta_id uuid NOT NULL,
    starting_state_key text NOT NULL DEFAULT 'post_first_four',
    highlighted_simulated_entry_id uuid,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    CONSTRAINT ck_derived_simulated_calcuttas_starting_state_key CHECK (starting_state_key = ANY (ARRAY['post_first_four','current']))
);

CREATE TABLE IF NOT EXISTS derived.simulated_entries (
    id uuid DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    simulated_calcutta_id uuid NOT NULL REFERENCES derived.simulated_calcuttas(id) ON DELETE CASCADE,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS derived.simulated_entry_teams (
    id uuid DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    simulated_entry_id uuid NOT NULL REFERENCES derived.simulated_entries(id) ON DELETE CASCADE,
    team_id uuid NOT NULL,
    bid_points integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS derived.simulated_calcutta_payouts (
    id uuid DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    simulated_calcutta_id uuid NOT NULL REFERENCES derived.simulated_calcuttas(id) ON DELETE CASCADE,
    position integer NOT NULL,
    payout_percent numeric(5,2) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS derived.simulated_calcutta_scoring_rules (
    id uuid DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    simulated_calcutta_id uuid NOT NULL REFERENCES derived.simulated_calcuttas(id) ON DELETE CASCADE,
    win_index integer NOT NULL,
    points integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

-- Re-create core.calcutta_snapshot_* tables
CREATE TABLE IF NOT EXISTS core.calcutta_snapshots (
    id uuid DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    calcutta_id uuid NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS core.calcutta_snapshot_entries (
    id uuid DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    calcutta_snapshot_id uuid NOT NULL REFERENCES core.calcutta_snapshots(id) ON DELETE CASCADE,
    name text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS core.calcutta_snapshot_entry_teams (
    id uuid DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    calcutta_snapshot_entry_id uuid NOT NULL REFERENCES core.calcutta_snapshot_entries(id) ON DELETE CASCADE,
    team_id uuid NOT NULL,
    bid_points integer NOT NULL DEFAULT 0,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS core.calcutta_snapshot_payouts (
    id uuid DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    calcutta_snapshot_id uuid NOT NULL REFERENCES core.calcutta_snapshots(id) ON DELETE CASCADE,
    position integer NOT NULL,
    payout_percent numeric(5,2) NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

CREATE TABLE IF NOT EXISTS core.calcutta_snapshot_scoring_rules (
    id uuid DEFAULT uuid_generate_v4() NOT NULL PRIMARY KEY,
    calcutta_snapshot_id uuid NOT NULL REFERENCES core.calcutta_snapshots(id) ON DELETE CASCADE,
    win_index integer NOT NULL,
    points integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz
);

-- Re-add lab.evaluations.simulated_calcutta_id column
ALTER TABLE lab.evaluations
    ADD COLUMN IF NOT EXISTS simulated_calcutta_id uuid;
ALTER TABLE lab.evaluations
    ADD CONSTRAINT evaluations_simulated_calcutta_id_fkey
    FOREIGN KEY (simulated_calcutta_id) REFERENCES derived.simulated_calcuttas(id);
