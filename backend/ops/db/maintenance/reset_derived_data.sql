TRUNCATE TABLE
    derived.entry_simulation_outcomes,
    derived.entry_performance,
    derived.simulated_teams,
    derived.calcutta_evaluation_runs,
    derived.simulated_tournaments,
    derived.simulation_state_teams,
    derived.simulation_states,
    lab_gold.detailed_investment_report,
    lab_gold.recommended_entry_bids,
    lab_gold.optimization_runs,
    lab_gold.strategy_generation_runs,
    lab_silver.predicted_market_share,
    lab_silver.predicted_game_outcomes
CASCADE;
