TRUNCATE TABLE
    analytics.entry_simulation_outcomes,
    analytics.entry_performance,
    analytics.simulated_tournaments,
    analytics.calcutta_evaluation_runs,
    analytics.tournament_simulation_batches,
    analytics.tournament_state_snapshot_teams,
    analytics.tournament_state_snapshots,
    lab_gold.detailed_investment_report,
    lab_gold.recommended_entry_bids,
    lab_gold.optimization_runs,
    lab_gold.strategy_generation_runs,
    lab_silver.predicted_market_share,
    lab_silver.predicted_game_outcomes
CASCADE;
