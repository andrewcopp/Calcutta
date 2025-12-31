"""Database writers for analytics tables."""

from .bronze_writers import (
    write_tournaments,
    write_teams,
    write_simulated_tournaments,
    write_calcuttas,
    write_entry_bids,
    write_payouts,
)

from .silver_writers import (
    write_predicted_game_outcomes,
    write_predicted_market_share,
    write_team_tournament_value,
)

from .gold_writers import (
    write_optimization_run,
    write_recommended_entry_bids,
    write_entry_simulation_outcomes,
    write_entry_performance,
    write_detailed_investment_report,
)

__all__ = [
    # Bronze
    'write_tournaments',
    'write_teams',
    'write_simulated_tournaments',
    'write_calcuttas',
    'write_entry_bids',
    'write_payouts',
    # Silver
    'write_predicted_game_outcomes',
    'write_predicted_market_share',
    'write_team_tournament_value',
    # Gold
    'write_optimization_run',
    'write_recommended_entry_bids',
    'write_entry_simulation_outcomes',
    'write_entry_performance',
    'write_detailed_investment_report',
]
