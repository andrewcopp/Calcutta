"""Database writers for analytics tables."""

from .bronze_writers import (
    get_or_create_tournament,
    write_teams,
)

from .silver_writers import (
    write_predicted_game_outcomes,
    write_simulated_tournaments,
    write_predicted_market_share,
)

from .gold_writers import (
    write_optimization_run,
    write_recommended_entry_bids,
)

__all__ = [
    # Bronze
    'get_or_create_tournament',
    'write_teams',
    # Silver
    'write_predicted_game_outcomes',
    'write_simulated_tournaments',
    'write_predicted_market_share',
    # Gold
    'write_optimization_run',
    'write_recommended_entry_bids',
]
