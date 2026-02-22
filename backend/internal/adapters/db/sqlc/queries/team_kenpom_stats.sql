-- name: UpsertTeamKenPomStats :exec
INSERT INTO core.team_kenpom_stats (team_id, net_rtg, o_rtg, d_rtg, adj_t)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (team_id)
DO UPDATE SET
    net_rtg = EXCLUDED.net_rtg,
    o_rtg = EXCLUDED.o_rtg,
    d_rtg = EXCLUDED.d_rtg,
    adj_t = EXCLUDED.adj_t,
    updated_at = NOW(),
    deleted_at = NULL;
