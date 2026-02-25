-- name: ListPredictionBatches :many
SELECT id::text, probability_source_key, through_round, created_at
FROM compute.prediction_batches
WHERE tournament_id = $1::uuid
    AND deleted_at IS NULL
ORDER BY created_at DESC;

-- name: GetLatestPredictionBatch :one
SELECT id::text, probability_source_key, through_round, created_at
FROM compute.prediction_batches
WHERE tournament_id = $1::uuid
    AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 1;

-- name: GetPredictionBatchByID :one
SELECT id::text, probability_source_key, through_round, created_at
FROM compute.prediction_batches
WHERE id = $1::uuid
    AND deleted_at IS NULL;

-- name: GetPredictedTeamValues :many
SELECT
    team_id::text,
    COALESCE(actual_points, 0) AS actual_points,
    expected_points,
    COALESCE(variance_points, 0) AS variance_points,
    COALESCE(std_points, 0) AS std_points,
    COALESCE(p_round_1, 0) AS p_round_1,
    COALESCE(p_round_2, 0) AS p_round_2,
    COALESCE(p_round_3, 0) AS p_round_3,
    COALESCE(p_round_4, 0) AS p_round_4,
    COALESCE(p_round_5, 0) AS p_round_5,
    COALESCE(p_round_6, 0) AS p_round_6,
    COALESCE(p_round_7, 0) AS p_round_7,
    COALESCE(favorites_total_points, 0) AS favorites_total_points
FROM compute.predicted_team_values
WHERE prediction_batch_id = $1::uuid
    AND deleted_at IS NULL;

-- name: GetFinalFourConfig :one
SELECT final_four_top_left, final_four_bottom_left,
       final_four_top_right, final_four_bottom_right
FROM core.tournaments
WHERE id = $1::uuid AND deleted_at IS NULL;

-- name: GetTeamsWithKenpomForPrediction :many
SELECT
    t.id::text,
    t.seed,
    t.region,
    COALESCE(ks.net_rtg, 0) AS kenpom_net,
    t.wins,
    COALESCE(t.byes, 0) AS byes
FROM core.teams t
LEFT JOIN core.team_kenpom_stats ks
    ON ks.team_id = t.id
    AND ks.deleted_at IS NULL
WHERE t.tournament_id = $1::uuid
    AND t.deleted_at IS NULL
ORDER BY t.region, t.seed;

-- name: GetScoringRulesForTournament :many
SELECT csr.win_index::int, csr.points_awarded::int
FROM core.calcutta_scoring_rules csr
JOIN core.calcuttas c ON c.id = csr.calcutta_id AND c.deleted_at IS NULL
WHERE c.tournament_id = $1::uuid
    AND csr.deleted_at IS NULL
ORDER BY csr.win_index ASC
LIMIT 10;

-- name: CreatePredictionBatch :one
INSERT INTO compute.prediction_batches (
    tournament_id,
    probability_source_key,
    game_outcome_spec_json,
    through_round
)
VALUES (sqlc.arg(tournament_id)::uuid, sqlc.arg(probability_source_key), sqlc.arg(game_outcome_spec_json)::jsonb, sqlc.arg(through_round))
RETURNING id::text;

-- name: CreatePredictedTeamValue :exec
INSERT INTO compute.predicted_team_values (
    prediction_batch_id,
    tournament_id,
    team_id,
    actual_points,
    expected_points,
    variance_points,
    std_points,
    p_round_1,
    p_round_2,
    p_round_3,
    p_round_4,
    p_round_5,
    p_round_6,
    p_round_7,
    favorites_total_points
)
VALUES (
    sqlc.arg(prediction_batch_id)::uuid,
    sqlc.arg(tournament_id)::uuid,
    sqlc.arg(team_id)::uuid,
    sqlc.arg(actual_points),
    sqlc.arg(expected_points),
    sqlc.arg(variance_points),
    sqlc.arg(std_points),
    sqlc.arg(p_round_1),
    sqlc.arg(p_round_2),
    sqlc.arg(p_round_3),
    sqlc.arg(p_round_4),
    sqlc.arg(p_round_5),
    sqlc.arg(p_round_6),
    sqlc.arg(p_round_7),
    sqlc.arg(favorites_total_points)
);

-- name: BulkCreatePredictedTeamValues :copyfrom
INSERT INTO compute.predicted_team_values (
    prediction_batch_id, tournament_id, team_id,
    actual_points, expected_points, variance_points, std_points,
    p_round_1, p_round_2, p_round_3, p_round_4, p_round_5, p_round_6, p_round_7,
    favorites_total_points
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15);

-- name: PruneOldBatchesForCheckpoint :execrows
DELETE FROM compute.prediction_batches pb_outer
WHERE pb_outer.tournament_id = sqlc.arg(tournament_id)::uuid
    AND pb_outer.through_round = sqlc.arg(through_round)
    AND pb_outer.deleted_at IS NULL
    AND pb_outer.id NOT IN (
        SELECT pb_inner.id FROM compute.prediction_batches pb_inner
        WHERE pb_inner.tournament_id = sqlc.arg(tournament_id)::uuid
            AND pb_inner.through_round = sqlc.arg(through_round)
            AND pb_inner.deleted_at IS NULL
        ORDER BY pb_inner.created_at DESC
        LIMIT sqlc.arg(keep_n)
    );

-- name: ListEligibleTournamentsForBackfill :many
SELECT t.id::text
FROM core.tournaments t
WHERE t.deleted_at IS NULL
    AND NOT EXISTS (
        SELECT 1 FROM compute.prediction_batches pb
        WHERE pb.tournament_id = t.id AND pb.deleted_at IS NULL
    )
    AND (
        SELECT COUNT(*) FROM core.teams tt
        WHERE tt.tournament_id = t.id AND tt.deleted_at IS NULL
    ) = 68
    AND EXISTS (
        SELECT 1 FROM core.team_kenpom_stats ks
        JOIN core.teams tt ON tt.id = ks.team_id AND tt.deleted_at IS NULL
        WHERE tt.tournament_id = t.id AND ks.deleted_at IS NULL
    )
    AND EXISTS (
        SELECT 1 FROM core.calcutta_scoring_rules csr
        JOIN core.calcuttas c ON c.id = csr.calcutta_id AND c.deleted_at IS NULL
        WHERE c.tournament_id = t.id AND csr.deleted_at IS NULL
    );
