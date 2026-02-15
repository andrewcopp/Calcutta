CREATE SCHEMA IF NOT EXISTS derived;

CREATE TABLE IF NOT EXISTS derived.synthetic_calcuttas (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    cohort_id UUID NOT NULL REFERENCES derived.suites(id),
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
    CONSTRAINT ck_derived_synthetic_calcuttas_starting_state_key
        CHECK (starting_state_key IS NULL OR starting_state_key IN ('post_first_four', 'current'))
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_synthetic_calcuttas_cohort_calcutta
ON derived.synthetic_calcuttas(cohort_id, calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcuttas_cohort_id
ON derived.synthetic_calcuttas(cohort_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcuttas_calcutta_id
ON derived.synthetic_calcuttas(calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcuttas_created_at
ON derived.synthetic_calcuttas(created_at DESC)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.synthetic_calcuttas;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.synthetic_calcuttas
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

INSERT INTO derived.synthetic_calcuttas (
    id,
    cohort_id,
    calcutta_id,
    calcutta_snapshot_id,
    focus_strategy_generation_run_id,
    focus_entry_name,
    starting_state_key,
    excluded_entry_name,
    params_json,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    sc.id,
    sc.suite_id,
    sc.calcutta_id,
    sc.calcutta_snapshot_id,
    sc.focus_strategy_generation_run_id,
    sc.focus_entry_name,
    sc.starting_state_key,
    sc.excluded_entry_name,
    sc.params_json,
    sc.created_at,
    sc.updated_at,
    sc.deleted_at
FROM derived.suite_scenarios sc
WHERE sc.deleted_at IS NULL
    AND NOT EXISTS (
        SELECT 1
        FROM derived.synthetic_calcuttas dst
        WHERE dst.id = sc.id
    );
