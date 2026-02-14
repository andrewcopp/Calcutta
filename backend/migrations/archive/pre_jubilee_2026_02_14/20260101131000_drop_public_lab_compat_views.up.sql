-- Drop public compatibility views for lab-tier tables.
-- These views were temporary shims during the cutover to schema-qualified lab tables.

DROP VIEW IF EXISTS public.gold_detailed_investment_report;
DROP VIEW IF EXISTS public.gold_entry_performance;
DROP VIEW IF EXISTS public.gold_entry_simulation_outcomes;
DROP VIEW IF EXISTS public.gold_recommended_entry_bids;
DROP VIEW IF EXISTS public.gold_optimization_runs;

DROP VIEW IF EXISTS public.silver_predicted_market_share;
DROP VIEW IF EXISTS public.silver_simulated_tournaments;
DROP VIEW IF EXISTS public.silver_predicted_game_outcomes;

DROP VIEW IF EXISTS public.bronze_payouts;
DROP VIEW IF EXISTS public.bronze_entry_bids;
DROP VIEW IF EXISTS public.bronze_calcuttas;
DROP VIEW IF EXISTS public.bronze_teams;
DROP VIEW IF EXISTS public.bronze_tournaments;
