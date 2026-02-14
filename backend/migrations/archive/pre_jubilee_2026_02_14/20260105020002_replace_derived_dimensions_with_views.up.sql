-- Phase 3: Replace legacy derived dimension tables with compatibility views over core.

CREATE SCHEMA IF NOT EXISTS derived;

DO $$
BEGIN
    -- Rename legacy tables out of the way (only if they exist as base tables).
    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'derived'
          AND c.relname = 'tournaments'
          AND c.relkind = 'r'
    ) THEN
        ALTER TABLE derived.tournaments RENAME TO tournaments_legacy;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'derived'
          AND c.relname = 'teams'
          AND c.relkind = 'r'
    ) THEN
        ALTER TABLE derived.teams RENAME TO teams_legacy;
    END IF;

    IF EXISTS (
        SELECT 1
        FROM pg_class c
        JOIN pg_namespace n ON n.oid = c.relnamespace
        WHERE n.nspname = 'derived'
          AND c.relname = 'calcuttas'
          AND c.relkind = 'r'
    ) THEN
        ALTER TABLE derived.calcuttas RENAME TO calcuttas_legacy;
    END IF;
END $$;

-- Compatibility views. These intentionally keep common legacy columns.

CREATE OR REPLACE VIEW derived.tournaments AS
SELECT
    t.id,
    t.id AS core_tournament_id,
    seas.year AS season,
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
    t.deleted_at
FROM core.tournaments t
LEFT JOIN core.seasons seas ON seas.id = t.season_id;

CREATE OR REPLACE VIEW derived.teams AS
SELECT
    tm.id,
    tm.id AS core_team_id,
    tm.tournament_id,
    tm.tournament_id AS core_tournament_id,
    s.name AS school_name,
    tm.seed,
    tm.region,
    ks.net_rtg AS kenpom_net,
    tm.created_at,
    tm.updated_at,
    tm.deleted_at
FROM core.teams tm
LEFT JOIN core.schools s ON s.id = tm.school_id
LEFT JOIN core.team_kenpom_stats ks ON ks.team_id = tm.id;

CREATE OR REPLACE VIEW derived.calcuttas AS
SELECT
    c.id,
    c.id AS core_calcutta_id,
    c.tournament_id,
    c.tournament_id AS core_tournament_id,
    c.owner_id,
    c.name,
    c.min_teams,
    c.max_teams,
    c.max_bid,
    c.created_at,
    c.updated_at,
    c.deleted_at
FROM core.calcuttas c
;
