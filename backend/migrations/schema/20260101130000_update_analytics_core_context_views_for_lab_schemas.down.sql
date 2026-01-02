-- Revert analytics core-context views to select from legacy (public) compatibility view names.

CREATE OR REPLACE VIEW public.bronze_tournaments_core_ctx AS
SELECT
    id,
    season,
    core_tournament_id,
    created_at
FROM bronze_tournaments;

CREATE OR REPLACE VIEW public.bronze_teams_core_ctx AS
SELECT
    t.id,
    t.tournament_id,
    bt.season,
    bt.core_tournament_id,
    t.school_slug,
    t.school_name,
    t.seed,
    t.region,
    t.kenpom_net,
    t.kenpom_adj_em,
    t.kenpom_adj_o,
    t.kenpom_adj_d,
    t.kenpom_adj_t,
    t.created_at
FROM bronze_teams t
JOIN bronze_tournaments bt ON bt.id = t.tournament_id;

CREATE OR REPLACE VIEW public.bronze_calcuttas_core_ctx AS
SELECT
    bc.id,
    bc.tournament_id,
    bt.season,
    bt.core_tournament_id,
    bc.core_calcutta_id,
    bc.name,
    bc.min_teams,
    bc.max_teams,
    bc.max_bid_points,
    bc.created_at
FROM bronze_calcuttas bc
JOIN bronze_tournaments bt ON bt.id = bc.tournament_id;
