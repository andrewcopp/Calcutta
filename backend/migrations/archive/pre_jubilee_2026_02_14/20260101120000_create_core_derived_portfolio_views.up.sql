CREATE OR REPLACE VIEW core.derived_portfolio_teams AS
WITH entry_bids AS (
    SELECT
        ce.id AS entry_id,
        ce.calcutta_id,
        cet.team_id,
        cet.bid_points::float8 AS bid_points,
        cet.created_at AS entry_team_created_at,
        cet.updated_at AS entry_team_updated_at,
        SUM(cet.bid_points::float8) OVER (
            PARTITION BY ce.calcutta_id, cet.team_id
        ) AS team_total_bid_points,
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
    FROM core.entries ce
    JOIN core.entry_teams cet ON cet.entry_id = ce.id AND cet.deleted_at IS NULL
    JOIN core.teams tt ON tt.id = cet.team_id AND tt.deleted_at IS NULL
    JOIN core.tournaments t ON t.id = tt.tournament_id AND t.deleted_at IS NULL
    LEFT JOIN core.schools s ON s.id = tt.school_id AND s.deleted_at IS NULL
    WHERE ce.deleted_at IS NULL
),
entry_team_points AS (
    SELECT
        eb.entry_id,
        eb.calcutta_id,
        eb.team_id,
        CASE
            WHEN eb.team_total_bid_points > 0 THEN (eb.bid_points / eb.team_total_bid_points)
            ELSE 0
        END AS ownership_percentage,
        core.calcutta_points_for_progress(eb.calcutta_id, eb.wins, eb.byes)::float8 AS team_points,
        core.calcutta_points_for_progress(eb.calcutta_id, eb.tournament_rounds, 0)::float8 AS team_max_points,
        eb.school_id,
        eb.tournament_id,
        eb.seed,
        eb.region,
        eb.byes,
        eb.wins,
        eb.eliminated,
        eb.entry_team_created_at,
        eb.entry_team_updated_at,
        eb.team_created_at,
        eb.team_updated_at,
        eb.school_name,
        eb.derived_updated_at
    FROM entry_bids eb
)
SELECT
    md5(concat(etp.entry_id::text, ':', etp.team_id::text)) AS id,
    etp.entry_id AS portfolio_id,
    etp.team_id,
    etp.ownership_percentage::float8 AS ownership_percentage,
    (etp.team_points * etp.ownership_percentage)::float8 AS actual_points,
    (CASE
        WHEN etp.eliminated THEN (etp.team_points * etp.ownership_percentage)
        ELSE (etp.team_max_points * etp.ownership_percentage)
    END)::float8 AS expected_points,
    (CASE
        WHEN etp.eliminated THEN (etp.team_points * etp.ownership_percentage)
        ELSE (etp.team_max_points * etp.ownership_percentage)
    END)::float8 AS predicted_points,
    etp.entry_team_created_at AS created_at,
    etp.derived_updated_at AS updated_at,
    NULL::timestamptz AS deleted_at,
    etp.team_id AS tournament_team_id,
    etp.school_id,
    etp.tournament_id,
    etp.seed,
    etp.region,
    etp.byes,
    etp.wins,
    etp.eliminated,
    etp.team_created_at,
    etp.team_updated_at,
    etp.school_name
FROM entry_team_points etp;

CREATE OR REPLACE VIEW core.derived_portfolios AS
WITH entry_totals AS (
    SELECT
        dpt.portfolio_id AS entry_id,
        SUM(dpt.expected_points)::float8 AS maximum_points,
        MAX(dpt.updated_at) AS updated_at
    FROM core.derived_portfolio_teams dpt
    GROUP BY dpt.portfolio_id
)
SELECT
    ce.id AS id,
    ce.id AS entry_id,
    COALESCE(et.maximum_points, 0)::float8 AS maximum_points,
    ce.created_at,
    COALESCE(et.updated_at, ce.updated_at) AS updated_at,
    NULL::timestamptz AS deleted_at
FROM core.entries ce
LEFT JOIN entry_totals et ON et.entry_id = ce.id
WHERE ce.deleted_at IS NULL;
