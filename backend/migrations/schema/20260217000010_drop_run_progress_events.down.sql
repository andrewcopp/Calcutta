-- Recreate the run_progress_events table
CREATE TABLE derived.run_progress_events (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    run_kind text NOT NULL,
    run_id uuid NOT NULL,
    run_key uuid,
    event_kind text NOT NULL,
    status text,
    percent double precision,
    phase text,
    message text,
    source text NOT NULL DEFAULT 'unknown'::text,
    payload_json jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT ck_derived_run_progress_events_event_kind CHECK ((event_kind = ANY (ARRAY['status'::text, 'progress'::text]))),
    CONSTRAINT ck_derived_run_progress_events_percent CHECK (((percent IS NULL) OR ((percent >= (0)::double precision) AND (percent <= (1)::double precision)))),
    CONSTRAINT ck_derived_run_progress_events_status CHECK (((status IS NULL) OR (status = ANY (ARRAY['queued'::text, 'running'::text, 'succeeded'::text, 'failed'::text]))))
);

ALTER TABLE ONLY derived.run_progress_events ADD CONSTRAINT run_progress_events_pkey PRIMARY KEY (id);
CREATE INDEX idx_derived_run_progress_events_kind_run_created_at ON derived.run_progress_events USING btree (run_kind, run_id, created_at);
CREATE INDEX idx_derived_run_progress_events_run_key ON derived.run_progress_events USING btree (run_key);

-- Recreate the trigger function
CREATE FUNCTION derived.enqueue_run_progress_event_from_run_jobs() RETURNS trigger
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
            payload_json
        )
        VALUES (
            NEW.run_kind,
            NEW.run_id,
            NEW.run_key,
            'status',
            NEW.status,
            NULL,
            NULL,
            NULL,
            'trigger',
            '{}'::jsonb
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
                payload_json
            )
            VALUES (
                NEW.run_kind,
                NEW.run_id,
                NEW.run_key,
                'status',
                NEW.status,
                NULL,
                NULL,
                NULL,
                'trigger',
                '{}'::jsonb
            );
        END IF;
        RETURN NEW;
    END IF;

    RETURN NEW;
END;
$$;

-- Recreate triggers
CREATE TRIGGER trg_derived_run_jobs_enqueue_run_progress_event_insert AFTER INSERT ON derived.run_jobs FOR EACH ROW EXECUTE FUNCTION derived.enqueue_run_progress_event_from_run_jobs();
CREATE TRIGGER trg_derived_run_jobs_enqueue_run_progress_event_update AFTER UPDATE ON derived.run_jobs FOR EACH ROW EXECUTE FUNCTION derived.enqueue_run_progress_event_from_run_jobs();
