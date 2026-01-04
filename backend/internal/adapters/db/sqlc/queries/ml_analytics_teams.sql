-- name: GetTeamPerformanceByID :one
WITH season_ctx AS (
    SELECT bt.core_tournament_id
    FROM derived.teams t
    JOIN derived.tournaments bt ON bt.id = t.tournament_id
    WHERE t.id = $1::uuid
    LIMIT 1
),
main_tournament AS (
    SELECT tr.id
    FROM core.tournaments tr
    JOIN season_ctx sc ON tr.id = sc.core_tournament_id
    WHERE tr.deleted_at IS NULL
    ORDER BY tr.created_at DESC
    LIMIT 1
),
calcutta_ctx AS (
    SELECT c.id AS calcutta_id
    FROM core.calcuttas c
    JOIN main_tournament mt ON mt.id = c.tournament_id
    WHERE c.deleted_at IS NULL
    ORDER BY c.created_at DESC
    LIMIT 1
),
round_distribution AS (
    SELECT 
        st.team_id,
        CASE (st.wins + st.byes)
            WHEN 0 THEN 'R64'
            WHEN 1 THEN 'R64'
            WHEN 2 THEN 'R32'
            WHEN 3 THEN 'S16'
            WHEN 4 THEN 'E8'
            WHEN 5 THEN 'F4'
            WHEN 6 THEN 'Finals'
            WHEN 7 THEN 'Champion'
            ELSE 'Unknown'
        END as round_name,
        COUNT(*)::int as count
    FROM derived.simulated_teams st
    JOIN derived.teams t ON t.id = st.team_id
    WHERE st.team_id = $1::uuid
    GROUP BY st.team_id, round_name
)
SELECT 
    t.id as team_id,
    t.school_name,
    t.seed,
    t.region,
    t.kenpom_net,
    COUNT(DISTINCT st.sim_id)::int as total_sims,
    AVG(st.wins)::float as avg_wins,
    AVG(
        CASE
            WHEN (SELECT calcutta_id FROM calcutta_ctx) IS NULL THEN 0
            ELSE core.calcutta_points_for_progress((SELECT calcutta_id FROM calcutta_ctx), st.wins, st.byes)
        END
    )::float as avg_points,
    jsonb_object_agg(rd.round_name, rd.count) as round_distribution
FROM derived.teams t
JOIN derived.simulated_teams st ON st.team_id = t.id
LEFT JOIN round_distribution rd ON rd.team_id = t.id
WHERE t.id = $1::uuid
GROUP BY t.id, t.school_name, t.seed, t.region, t.kenpom_net;

-- name: GetTeamPredictionsByYear :many
SELECT 
    t.id as team_id,
    t.school_name,
    t.seed,
    t.region,
    t.kenpom_net
FROM derived.teams t
JOIN derived.tournaments bt ON bt.id = t.tournament_id
WHERE bt.season = $1::int
ORDER BY t.seed;
