CREATE OR REPLACE FUNCTION core.calcutta_points_for_progress(
    p_calcutta_id UUID,
    p_wins INTEGER,
    p_byes INTEGER DEFAULT 0
)
RETURNS INTEGER
LANGUAGE sql
STABLE
AS $$
    SELECT COALESCE(SUM(r.points_awarded), 0)::int
    FROM core.calcutta_scoring_rules r
    WHERE r.calcutta_id = p_calcutta_id
      AND r.deleted_at IS NULL
      AND r.win_index <= (COALESCE(p_wins, 0) + COALESCE(p_byes, 0));
$$;
