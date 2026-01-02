-- Backfill core.* tables from existing public tables.
-- This migration preserves IDs so future cutover can be done with stable references.

-- 1) Seasons
WITH tournament_years AS (
    SELECT DISTINCT
        COALESCE(
            CAST(SUBSTRING(t.name FROM '[0-9]{4}') AS INTEGER),
            CAST(EXTRACT(YEAR FROM t.starting_at) AS INTEGER)
        ) AS year
    FROM tournaments t
    WHERE t.deleted_at IS NULL
)
INSERT INTO core.seasons (year)
SELECT ty.year
FROM tournament_years ty
WHERE ty.year IS NOT NULL
ON CONFLICT (year) DO NOTHING;

-- 2) Competitions (seed a single competition for now)
INSERT INTO core.competitions (name)
VALUES ('NCAA Men''s')
ON CONFLICT (name) DO NOTHING;

-- 3) Tournaments
WITH competition AS (
    SELECT id
    FROM core.competitions
    WHERE name = 'NCAA Men''s'
    LIMIT 1
),
source AS (
    SELECT
        t.id,
        t.name,
        t.import_key,
        t.rounds,
        t.starting_at,
        t.final_four_top_left,
        t.final_four_bottom_left,
        t.final_four_top_right,
        t.final_four_bottom_right,
        t.created_at,
        t.updated_at,
        t.deleted_at,
        COALESCE(
            CAST(SUBSTRING(t.name FROM '[0-9]{4}') AS INTEGER),
            CAST(EXTRACT(YEAR FROM t.starting_at) AS INTEGER)
        ) AS year
    FROM tournaments t
    WHERE t.deleted_at IS NULL
)
INSERT INTO core.tournaments (
    id,
    competition_id,
    season_id,
    name,
    import_key,
    rounds,
    starting_at,
    final_four_top_left,
    final_four_bottom_left,
    final_four_top_right,
    final_four_bottom_right,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    s.id,
    c.id,
    seas.id,
    s.name,
    s.import_key,
    s.rounds,
    s.starting_at,
    s.final_four_top_left,
    s.final_four_bottom_left,
    s.final_four_top_right,
    s.final_four_bottom_right,
    s.created_at,
    s.updated_at,
    s.deleted_at
FROM source s
JOIN core.seasons seas ON seas.year = s.year
CROSS JOIN competition c
ON CONFLICT (id) DO NOTHING;

-- 4) Schools
INSERT INTO core.schools (
    id,
    name,
    slug,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    s.id,
    s.name,
    s.slug,
    s.created_at,
    s.updated_at,
    s.deleted_at
FROM schools s
WHERE s.deleted_at IS NULL
ON CONFLICT (id) DO NOTHING;

-- 5) Teams (tournament-scoped)
INSERT INTO core.teams (
    id,
    tournament_id,
    school_id,
    seed,
    region,
    byes,
    wins,
    eliminated,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    tt.id,
    tt.tournament_id,
    tt.school_id,
    tt.seed,
    tt.region,
    tt.byes,
    tt.wins,
    tt.eliminated,
    tt.created_at,
    tt.updated_at,
    tt.deleted_at
FROM tournament_teams tt
WHERE tt.deleted_at IS NULL
ON CONFLICT (id) DO NOTHING;

-- 6) KenPom stats (protected cleaned data)
INSERT INTO core.team_kenpom_stats (
    team_id,
    net_rtg,
    o_rtg,
    d_rtg,
    adj_t,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    kps.tournament_team_id,
    kps.net_rtg,
    kps.o_rtg,
    kps.d_rtg,
    kps.adj_t,
    kps.created_at,
    kps.updated_at,
    kps.deleted_at
FROM tournament_team_kenpom_stats kps
WHERE kps.deleted_at IS NULL
ON CONFLICT (team_id) DO NOTHING;

-- 7) Calcuttas
INSERT INTO core.calcuttas (
    id,
    tournament_id,
    owner_id,
    name,
    min_teams,
    max_teams,
    max_bid,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    c.id,
    c.tournament_id,
    c.owner_id,
    c.name,
    c.min_teams,
    c.max_teams,
    c.max_bid,
    c.created_at,
    c.updated_at,
    c.deleted_at
FROM calcuttas c
WHERE c.deleted_at IS NULL
ON CONFLICT (id) DO NOTHING;

-- 8) Entries
INSERT INTO core.entries (
    id,
    name,
    user_id,
    calcutta_id,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    ce.id,
    ce.name,
    ce.user_id,
    ce.calcutta_id,
    ce.created_at,
    ce.updated_at,
    ce.deleted_at
FROM calcutta_entries ce
WHERE ce.deleted_at IS NULL
ON CONFLICT (id) DO NOTHING;

-- 9) Entry teams (bids)
INSERT INTO core.entry_teams (
    id,
    entry_id,
    team_id,
    bid_points,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    cet.id,
    cet.entry_id,
    cet.team_id,
    cet.bid,
    cet.created_at,
    cet.updated_at,
    cet.deleted_at
FROM calcutta_entry_teams cet
WHERE cet.deleted_at IS NULL
ON CONFLICT (id) DO NOTHING;

-- 10) Payouts (real money)
INSERT INTO core.payouts (
    id,
    calcutta_id,
    position,
    amount_cents,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    cp.id,
    cp.calcutta_id,
    cp.position,
    cp.amount_cents,
    cp.created_at,
    cp.updated_at,
    cp.deleted_at
FROM calcutta_payouts cp
WHERE cp.deleted_at IS NULL
ON CONFLICT (id) DO NOTHING;

-- 11) Scoring rules (incremental per win) from calcutta_rounds
-- Map (calcutta_id, round ASC) -> win_index (1..N)
WITH ranked AS (
    SELECT
        cr.id,
        cr.calcutta_id,
        ROW_NUMBER() OVER (PARTITION BY cr.calcutta_id ORDER BY cr.round ASC) AS win_index,
        cr.points AS points_awarded,
        cr.created_at,
        cr.updated_at,
        cr.deleted_at
    FROM calcutta_rounds cr
    WHERE cr.deleted_at IS NULL
)
INSERT INTO core.calcutta_scoring_rules (
    id,
    calcutta_id,
    win_index,
    points_awarded,
    created_at,
    updated_at,
    deleted_at
)
SELECT
    r.id,
    r.calcutta_id,
    r.win_index,
    r.points_awarded,
    r.created_at,
    r.updated_at,
    r.deleted_at
FROM ranked r
ON CONFLICT (id) DO NOTHING;
