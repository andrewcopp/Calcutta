CREATE SCHEMA IF NOT EXISTS legacy;

ALTER TABLE IF EXISTS public.schools SET SCHEMA legacy;
ALTER TABLE IF EXISTS public.tournaments SET SCHEMA legacy;
ALTER TABLE IF EXISTS public.tournament_teams SET SCHEMA legacy;
ALTER TABLE IF EXISTS public.tournament_team_kenpom_stats SET SCHEMA legacy;
ALTER TABLE IF EXISTS public.tournament_games SET SCHEMA legacy;

ALTER TABLE IF EXISTS public.calcuttas SET SCHEMA legacy;
ALTER TABLE IF EXISTS public.calcutta_rounds SET SCHEMA legacy;
ALTER TABLE IF EXISTS public.calcutta_entries SET SCHEMA legacy;
ALTER TABLE IF EXISTS public.calcutta_entry_teams SET SCHEMA legacy;
ALTER TABLE IF EXISTS public.calcutta_payouts SET SCHEMA legacy;
ALTER TABLE IF EXISTS public.calcutta_portfolios SET SCHEMA legacy;
ALTER TABLE IF EXISTS public.calcutta_portfolio_teams SET SCHEMA legacy;
