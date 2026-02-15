CREATE TABLE IF NOT EXISTS derived.candidate_bids (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    candidate_id UUID NOT NULL REFERENCES derived.candidates(id) ON DELETE CASCADE,
    team_id UUID NOT NULL REFERENCES core.teams(id) ON DELETE CASCADE,
    bid_points INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_derived_candidate_bids_candidate_team
ON derived.candidate_bids(candidate_id, team_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_candidate_bids_candidate_id
ON derived.candidate_bids(candidate_id)
WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_derived_candidate_bids_team_id
ON derived.candidate_bids(team_id)
WHERE deleted_at IS NULL;

DROP TRIGGER IF EXISTS set_updated_at ON derived.candidate_bids;
CREATE TRIGGER set_updated_at
BEFORE UPDATE ON derived.candidate_bids
FOR EACH ROW
EXECUTE FUNCTION core.set_updated_at();

INSERT INTO derived.candidate_bids (candidate_id, team_id, bid_points)
SELECT c.id, b.team_id, b.bid_points
FROM derived.candidates c
JOIN derived.recommended_entry_bids b
    ON b.strategy_generation_run_id = c.strategy_generation_run_id
    AND b.deleted_at IS NULL
WHERE c.deleted_at IS NULL
    AND c.strategy_generation_run_id IS NOT NULL
ON CONFLICT (candidate_id, team_id) WHERE deleted_at IS NULL
DO UPDATE
SET bid_points = EXCLUDED.bid_points,
    updated_at = NOW(),
    deleted_at = NULL;
