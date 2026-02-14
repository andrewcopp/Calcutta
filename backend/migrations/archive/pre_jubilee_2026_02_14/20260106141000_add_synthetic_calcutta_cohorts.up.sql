CREATE SCHEMA IF NOT EXISTS derived;

CREATE TABLE IF NOT EXISTS derived.synthetic_calcutta_cohorts (
    id UUID PRIMARY KEY REFERENCES derived.suites(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

DROP TRIGGER IF EXISTS set_updated_at ON derived.synthetic_calcutta_cohorts;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.synthetic_calcutta_cohorts
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

-- Backfill: every suite is treated as a cohort for now.
INSERT INTO derived.synthetic_calcutta_cohorts (id, created_at, updated_at, deleted_at)
SELECT s.id, s.created_at, s.updated_at, s.deleted_at
FROM derived.suites s
WHERE NOT EXISTS (
    SELECT 1
    FROM derived.synthetic_calcutta_cohorts c
    WHERE c.id = s.id
);

CREATE OR REPLACE FUNCTION derived.sync_synthetic_calcutta_cohort_from_suite()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO derived.synthetic_calcutta_cohorts (id, created_at, updated_at, deleted_at)
    VALUES (NEW.id, NEW.created_at, NEW.updated_at, NEW.deleted_at)
    ON CONFLICT (id) DO UPDATE
        SET deleted_at = EXCLUDED.deleted_at,
            updated_at = NOW();
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_sync_synthetic_calcutta_cohort_from_suite ON derived.suites;
CREATE TRIGGER trg_sync_synthetic_calcutta_cohort_from_suite
AFTER INSERT OR UPDATE ON derived.suites
FOR EACH ROW
EXECUTE FUNCTION derived.sync_synthetic_calcutta_cohort_from_suite();
