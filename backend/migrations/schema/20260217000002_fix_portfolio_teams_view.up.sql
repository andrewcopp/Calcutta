-- Fix portfolio_teams view regression from migration 0005.
-- The arithmetic (wins + byes + (tournament_rounds - wins - byes)) always equals
-- tournament_rounds, losing the actual scoring-rule-based point calculation.
-- Restore core.calcutta_points_for_progress() calls for correct scoring.

DROP VIEW IF EXISTS derived.portfolios;
DROP VIEW IF EXISTS derived.portfolio_teams;

CREATE VIEW derived.portfolio_teams AS
WITH entry_bids AS (
    SELECT ce.id AS entry_id,
        ce.calcutta_id,
        cet.team_id,
        (cet.bid_points)::double precision AS bid_points,
        cet.created_at AS entry_team_created_at,
        cet.updated_at AS entry_team_updated_at,
        sum((cet.bid_points)::double precision) OVER (PARTITION BY ce.calcutta_id, cet.team_id) AS team_total_bid_points,
        tt.school_id,
        tt.tournament_id,
        tt.seed,
        tt.region,
        tt.byes,
        tt.wins,
        tt.eliminated,
        tt.created_at AS team_created_at,
        tt.updated_at AS team_updated_at,
        t.rounds AS tournament_rounds,
        s.name AS school_name,
        GREATEST(ce.updated_at, cet.updated_at, tt.updated_at) AS derived_updated_at
    FROM ((((core.entries ce
        JOIN core.entry_teams cet ON (((cet.entry_id = ce.id) AND (cet.deleted_at IS NULL))))
        JOIN core.teams tt ON (((tt.id = cet.team_id) AND (tt.deleted_at IS NULL))))
        JOIN core.tournaments t ON (((t.id = tt.tournament_id) AND (t.deleted_at IS NULL))))
        LEFT JOIN core.schools s ON (((s.id = tt.school_id) AND (s.deleted_at IS NULL))))
    WHERE (ce.deleted_at IS NULL)
), entry_team_points AS (
    SELECT eb.entry_id,
        eb.calcutta_id,
        eb.team_id,
        eb.bid_points,
        eb.team_total_bid_points,
        CASE
            WHEN (eb.team_total_bid_points > (0)::double precision) THEN (eb.bid_points / eb.team_total_bid_points)
            ELSE (0)::double precision
        END AS ownership_percentage,
        (
            CASE
                WHEN (eb.team_total_bid_points > (0)::double precision) THEN (eb.bid_points / eb.team_total_bid_points)
                ELSE (0)::double precision
            END *
            CASE
                WHEN (eb.eliminated = true) THEN (core.calcutta_points_for_progress(eb.calcutta_id, eb.wins, eb.byes))::double precision
                ELSE (0)::double precision
            END
        ) AS actual_points,
        (
            CASE
                WHEN (eb.team_total_bid_points > (0)::double precision) THEN (eb.bid_points / eb.team_total_bid_points)
                ELSE (0)::double precision
            END *
            CASE
                WHEN (eb.eliminated = true)
                    THEN (core.calcutta_points_for_progress(eb.calcutta_id, eb.wins, eb.byes))::double precision
                ELSE (core.calcutta_points_for_progress(eb.calcutta_id, eb.tournament_rounds, 0))::double precision
            END
        ) AS expected_points,
        (0)::double precision AS predicted_points,
        eb.school_id,
        eb.tournament_id,
        eb.seed,
        eb.region,
        eb.byes,
        eb.wins,
        eb.eliminated,
        eb.team_created_at,
        eb.team_updated_at,
        eb.school_name,
        eb.entry_team_created_at AS created_at,
        eb.derived_updated_at AS updated_at,
        NULL::timestamp with time zone AS deleted_at,
        eb.entry_id AS portfolio_id
    FROM entry_bids eb
)
SELECT concat(etp.entry_id, '-', etp.team_id) AS id,
    etp.entry_id AS portfolio_id,
    etp.team_id,
    etp.ownership_percentage,
    etp.actual_points,
    etp.expected_points,
    etp.predicted_points,
    etp.created_at,
    etp.updated_at,
    etp.deleted_at
FROM entry_team_points etp;

CREATE VIEW derived.portfolios AS
WITH entry_totals AS (
    SELECT dpt.portfolio_id AS entry_id,
        sum(dpt.expected_points) AS maximum_points,
        max(dpt.updated_at) AS updated_at
    FROM derived.portfolio_teams dpt
    GROUP BY dpt.portfolio_id
)
SELECT ce.id,
    ce.id AS entry_id,
    COALESCE(et.maximum_points, (0)::double precision) AS maximum_points,
    ce.created_at,
    COALESCE(et.updated_at, ce.updated_at) AS updated_at,
    NULL::timestamp with time zone AS deleted_at
FROM (core.entries ce
    LEFT JOIN entry_totals et ON ((et.entry_id = ce.id)))
WHERE (ce.deleted_at IS NULL);
