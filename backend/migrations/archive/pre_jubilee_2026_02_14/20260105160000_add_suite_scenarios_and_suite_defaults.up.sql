ALTER TABLE IF EXISTS derived.suites
    ADD COLUMN IF NOT EXISTS starting_state_key TEXT NOT NULL DEFAULT 'post_first_four';

ALTER TABLE IF EXISTS derived.suites
    ADD COLUMN IF NOT EXISTS excluded_entry_name TEXT;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'ck_derived_suites_starting_state_key'
    ) THEN
        ALTER TABLE derived.suites
            ADD CONSTRAINT ck_derived_suites_starting_state_key
            CHECK (starting_state_key IN ('post_first_four', 'current'));
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS derived.suite_scenarios (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    suite_id UUID NOT NULL REFERENCES derived.suites(id),
    calcutta_id UUID NOT NULL REFERENCES core.calcuttas(id),
    calcutta_snapshot_id UUID REFERENCES core.calcutta_snapshots(id),
    focus_strategy_generation_run_id UUID REFERENCES derived.strategy_generation_runs(id),
    focus_entry_name TEXT,
    starting_state_key TEXT,
    excluded_entry_name TEXT,
    params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_derived_suite_scenarios_starting_state_key
        CHECK (starting_state_key IS NULL OR starting_state_key IN ('post_first_four', 'current'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_suite_scenarios_suite_calcutta
ON derived.suite_scenarios(suite_id, calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_suite_scenarios_suite_id
ON derived.suite_scenarios(suite_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_suite_scenarios_calcutta_id
ON derived.suite_scenarios(calcutta_id)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.suite_scenarios;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.suite_scenarios
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

INSERT INTO permissions (key, description)
VALUES
    ('analytics.suites.write', 'Create/update suites'),
    ('analytics.suites.read', 'Read suites'),
    ('analytics.suite_scenarios.write', 'Create/update suite scenarios'),
    ('analytics.suite_scenarios.read', 'Read suite scenarios'),
    ('analytics.strategy_generation_runs.write', 'Create strategy generation runs'),
    ('analytics.strategy_generation_runs.read', 'Read strategy generation runs')
ON CONFLICT (key) DO NOTHING;

INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key IN (
    'analytics.suites.write',
    'analytics.suites.read',
    'analytics.suite_scenarios.write',
    'analytics.suite_scenarios.read',
    'analytics.strategy_generation_runs.write',
    'analytics.strategy_generation_runs.read'
)
WHERE l.key IN ('global_admin')
ON CONFLICT DO NOTHING;
