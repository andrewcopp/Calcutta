CREATE TABLE IF NOT EXISTS derived.entry_evaluation_requests (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calcutta_id UUID NOT NULL REFERENCES core.calcuttas(id),
    entry_candidate_id UUID NOT NULL REFERENCES models.entry_candidates(id),
    excluded_entry_name TEXT,
    starting_state_key TEXT NOT NULL,
    n_sims INTEGER NOT NULL,
    seed INTEGER NOT NULL,
    experiment_key TEXT,
    request_source TEXT,
    status TEXT NOT NULL DEFAULT 'queued',
    claimed_at TIMESTAMPTZ,
    claimed_by TEXT,
    evaluation_run_id UUID REFERENCES derived.calcutta_evaluation_runs(id),
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_derived_entry_evaluation_requests_starting_state_key
        CHECK (starting_state_key IN ('post_first_four')),
    CONSTRAINT ck_derived_entry_evaluation_requests_status
        CHECK (status IN ('queued', 'running', 'succeeded', 'failed'))
);

CREATE INDEX IF NOT EXISTS idx_derived_entry_evaluation_requests_status
ON derived.entry_evaluation_requests(status)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_entry_evaluation_requests_experiment_key
ON derived.entry_evaluation_requests(experiment_key)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_entry_evaluation_requests_created_at
ON derived.entry_evaluation_requests(created_at DESC)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_entry_evaluation_requests_entry_candidate_id
ON derived.entry_evaluation_requests(entry_candidate_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_entry_evaluation_requests_calcutta_id
ON derived.entry_evaluation_requests(calcutta_id)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.entry_evaluation_requests;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.entry_evaluation_requests
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();
