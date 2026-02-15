CREATE TABLE IF NOT EXISTS derived.suite_executions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    suite_id UUID NOT NULL REFERENCES derived.suites(id),
    name TEXT,
    optimizer_key TEXT,
    n_sims INT,
    seed INT,
    starting_state_key TEXT NOT NULL DEFAULT 'post_first_four',
    excluded_entry_name TEXT,
    status TEXT NOT NULL DEFAULT 'running',
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_derived_suite_executions_status
        CHECK (status IN ('queued', 'running', 'succeeded', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_derived_suite_executions_suite_id
ON derived.suite_executions(suite_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_suite_executions_created_at
ON derived.suite_executions(created_at DESC)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.suite_executions;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.suite_executions
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

ALTER TABLE IF EXISTS derived.suite_calcutta_evaluations
    ADD COLUMN IF NOT EXISTS suite_execution_id UUID REFERENCES derived.suite_executions(id),
    ADD COLUMN IF NOT EXISTS realized_finish_position INT,
    ADD COLUMN IF NOT EXISTS realized_is_tied BOOLEAN,
    ADD COLUMN IF NOT EXISTS realized_in_the_money BOOLEAN,
    ADD COLUMN IF NOT EXISTS realized_payout_cents INT,
    ADD COLUMN IF NOT EXISTS realized_total_points DOUBLE PRECISION;

CREATE INDEX IF NOT EXISTS idx_derived_suite_calcutta_evaluations_suite_execution_id
ON derived.suite_calcutta_evaluations(suite_execution_id)
WHERE deleted_at IS NULL;

INSERT INTO permissions (key, description)
VALUES
    ('analytics.suite_executions.write', 'Create suite executions'),
    ('analytics.suite_executions.read', 'Read suite executions')
ON CONFLICT (key) DO NOTHING;

INSERT INTO label_permissions (label_id, permission_id)
SELECT l.id, p.id
FROM labels l
JOIN permissions p ON p.key IN (
    'analytics.suite_executions.write',
    'analytics.suite_executions.read'
)
WHERE l.key IN ('global_admin')
ON CONFLICT DO NOTHING;
