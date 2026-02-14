CREATE TABLE IF NOT EXISTS derived.run_progress_events (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    run_kind TEXT NOT NULL,
    run_id UUID NOT NULL,
    run_key UUID,
    event_kind TEXT NOT NULL,
    status TEXT,
    percent DOUBLE PRECISION,
    phase TEXT,
    message TEXT,
    source TEXT NOT NULL DEFAULT 'unknown',
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT ck_derived_run_progress_events_event_kind
        CHECK (event_kind IN ('status', 'progress')),
    CONSTRAINT ck_derived_run_progress_events_status
        CHECK (status IS NULL OR status IN ('queued', 'running', 'succeeded', 'failed')),
    CONSTRAINT ck_derived_run_progress_events_percent
        CHECK (percent IS NULL OR (percent >= 0 AND percent <= 1))
);

CREATE INDEX IF NOT EXISTS idx_derived_run_progress_events_kind_run_created_at
ON derived.run_progress_events(run_kind, run_id, created_at);

CREATE INDEX IF NOT EXISTS idx_derived_run_progress_events_run_key
ON derived.run_progress_events(run_key);

CREATE OR REPLACE FUNCTION derived.enqueue_run_progress_event_from_run_jobs()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO derived.run_progress_events (
            run_kind,
            run_id,
            run_key,
            event_kind,
            status,
            percent,
            phase,
            message,
            source,
            payload_json,
            created_at
        )
        VALUES (
            NEW.run_kind,
            NEW.run_id,
            NEW.run_key,
            'status',
            NEW.status,
            NULL,
            NEW.status,
            NULL,
            'db_trigger',
            '{}'::jsonb,
            COALESCE(NEW.created_at, NOW())
        );

        RETURN NEW;
    END IF;

    IF TG_OP = 'UPDATE' THEN
        IF NEW.status IS DISTINCT FROM OLD.status THEN
            INSERT INTO derived.run_progress_events (
                run_kind,
                run_id,
                run_key,
                event_kind,
                status,
                percent,
                phase,
                message,
                source,
                payload_json,
                created_at
            )
            VALUES (
                NEW.run_kind,
                NEW.run_id,
                NEW.run_key,
                'status',
                NEW.status,
                NULL,
                NEW.status,
                CASE WHEN NEW.status = 'failed' THEN NEW.error_message ELSE NULL END,
                'db_trigger',
                '{}'::jsonb,
                COALESCE(NEW.finished_at, NEW.updated_at, NOW())
            );
        END IF;

        RETURN NEW;
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS trg_derived_run_jobs_enqueue_run_progress_event_insert ON derived.run_jobs;
CREATE TRIGGER trg_derived_run_jobs_enqueue_run_progress_event_insert
AFTER INSERT ON derived.run_jobs
FOR EACH ROW
EXECUTE FUNCTION derived.enqueue_run_progress_event_from_run_jobs();

DROP TRIGGER IF EXISTS trg_derived_run_jobs_enqueue_run_progress_event_update ON derived.run_jobs;
CREATE TRIGGER trg_derived_run_jobs_enqueue_run_progress_event_update
AFTER UPDATE OF status ON derived.run_jobs
FOR EACH ROW
EXECUTE FUNCTION derived.enqueue_run_progress_event_from_run_jobs();

-- Backfill lifecycle events for existing runs (best-effort, idempotent).
INSERT INTO derived.run_progress_events (
    run_kind,
    run_id,
    run_key,
    event_kind,
    status,
    percent,
    phase,
    message,
    source,
    payload_json,
    created_at
)
SELECT
    j.run_kind,
    j.run_id,
    j.run_key,
    'status',
    'queued',
    NULL,
    'queued',
    NULL,
    'migration_backfill',
    '{}'::jsonb,
    j.created_at
FROM derived.run_jobs j
WHERE NOT EXISTS (
    SELECT 1
    FROM derived.run_progress_events e
    WHERE e.run_kind = j.run_kind
        AND e.run_id = j.run_id
        AND e.event_kind = 'status'
        AND e.status = 'queued'
);

INSERT INTO derived.run_progress_events (
    run_kind,
    run_id,
    run_key,
    event_kind,
    status,
    percent,
    phase,
    message,
    source,
    payload_json,
    created_at
)
SELECT
    j.run_kind,
    j.run_id,
    j.run_key,
    'status',
    'running',
    NULL,
    'running',
    NULL,
    'migration_backfill',
    '{}'::jsonb,
    j.started_at
FROM derived.run_jobs j
WHERE j.started_at IS NOT NULL
    AND NOT EXISTS (
        SELECT 1
        FROM derived.run_progress_events e
        WHERE e.run_kind = j.run_kind
            AND e.run_id = j.run_id
            AND e.event_kind = 'status'
            AND e.status = 'running'
    );

INSERT INTO derived.run_progress_events (
    run_kind,
    run_id,
    run_key,
    event_kind,
    status,
    percent,
    phase,
    message,
    source,
    payload_json,
    created_at
)
SELECT
    j.run_kind,
    j.run_id,
    j.run_key,
    'status',
    j.status,
    NULL,
    j.status,
    CASE WHEN j.status = 'failed' THEN j.error_message ELSE NULL END,
    'migration_backfill',
    '{}'::jsonb,
    COALESCE(j.finished_at, j.updated_at)
FROM derived.run_jobs j
WHERE j.status IN ('succeeded', 'failed')
    AND COALESCE(j.finished_at, j.updated_at) IS NOT NULL
    AND NOT EXISTS (
        SELECT 1
        FROM derived.run_progress_events e
        WHERE e.run_kind = j.run_kind
            AND e.run_id = j.run_id
            AND e.event_kind = 'status'
            AND e.status = j.status
    );
