CREATE SCHEMA IF NOT EXISTS derived;

CREATE TABLE IF NOT EXISTS derived.candidates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    source_kind TEXT NOT NULL,
    source_entry_artifact_id UUID REFERENCES derived.run_artifacts(id),
    display_name TEXT NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ,
    CONSTRAINT ck_derived_candidates_source_kind
        CHECK (source_kind IN ('manual', 'entry_artifact', 'other'))
);

CREATE INDEX IF NOT EXISTS idx_derived_candidates_source_kind
ON derived.candidates(source_kind)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_candidates_source_entry_artifact_id
ON derived.candidates(source_entry_artifact_id)
WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_candidates_source_kind_source_entry_artifact
ON derived.candidates(source_kind, source_entry_artifact_id)
WHERE deleted_at IS NULL
  AND source_entry_artifact_id IS NOT NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.candidates;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.candidates
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

CREATE TABLE IF NOT EXISTS derived.synthetic_calcutta_candidates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    synthetic_calcutta_id UUID NOT NULL REFERENCES derived.synthetic_calcuttas(id) ON DELETE CASCADE,
    candidate_id UUID NOT NULL REFERENCES derived.candidates(id) ON DELETE CASCADE,
    snapshot_entry_id UUID NOT NULL REFERENCES core.calcutta_snapshot_entries(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_synthetic_calcutta_candidates_synthetic_candidate
ON derived.synthetic_calcutta_candidates(synthetic_calcutta_id, candidate_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcutta_candidates_synthetic_calcutta_id
ON derived.synthetic_calcutta_candidates(synthetic_calcutta_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcutta_candidates_candidate_id
ON derived.synthetic_calcutta_candidates(candidate_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_synthetic_calcutta_candidates_snapshot_entry_id
ON derived.synthetic_calcutta_candidates(snapshot_entry_id)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.synthetic_calcutta_candidates;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.synthetic_calcutta_candidates
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

WITH rows_to_attach AS (
    SELECT
        sc.id AS synthetic_calcutta_id,
        e.id AS snapshot_entry_id,
        uuid_generate_v4() AS candidate_id,
        e.display_name AS display_name,
        jsonb_build_object(
            'teams',
            COALESCE(
                jsonb_agg(
                    jsonb_build_object(
                        'team_id', et.team_id::text,
                        'bid_points', et.bid_points
                    )
                ) FILTER (WHERE et.id IS NOT NULL),
                '[]'::jsonb
            )
        ) AS metadata_json
    FROM derived.synthetic_calcuttas sc
    JOIN core.calcutta_snapshot_entries e
        ON e.calcutta_snapshot_id = sc.calcutta_snapshot_id
        AND e.deleted_at IS NULL
    LEFT JOIN core.calcutta_snapshot_entry_teams et
        ON et.calcutta_snapshot_entry_id = e.id
        AND et.deleted_at IS NULL
    WHERE sc.deleted_at IS NULL
        AND e.is_synthetic = TRUE
        AND e.entry_id IS NULL
        AND NOT EXISTS (
            SELECT 1
            FROM derived.synthetic_calcutta_candidates scc
            WHERE scc.synthetic_calcutta_id = sc.id
                AND scc.snapshot_entry_id = e.id
                AND scc.deleted_at IS NULL
        )
    GROUP BY sc.id, e.id, e.display_name
),
ins_candidates AS (
    INSERT INTO derived.candidates (id, source_kind, source_entry_artifact_id, display_name, metadata_json)
    SELECT candidate_id, 'manual', NULL, display_name, metadata_json
    FROM rows_to_attach
    RETURNING id
)
INSERT INTO derived.synthetic_calcutta_candidates (synthetic_calcutta_id, candidate_id, snapshot_entry_id)
SELECT r.synthetic_calcutta_id, r.candidate_id, r.snapshot_entry_id
FROM rows_to_attach r
JOIN ins_candidates c
    ON c.id = r.candidate_id;
