"""
Lab module for simplified R&D iteration on investment models.

This module provides a clean interface for:
1. Creating/managing investment models
2. Generating entries (optimized bids)
3. Recording evaluation results
4. Comparing model performance
5. Portfolio optimization research (maxmin)

The lab schema replaces the over-complicated derived.algorithms,
derived.candidates, and derived.suite_* tables with just 3 tables:
- lab.investment_models
- lab.entries
- lab.evaluations

For production portfolio optimization, use the Go DP allocator in
backend/internal/app/recommended_entry_bids/allocator.go which provides
exact optimal solutions.
"""

from moneyball.lab.models import (
    InvestmentModel,
    Entry,
    Evaluation,
    create_investment_model,
    get_investment_model,
    create_entry,
    get_entry,
    create_evaluation,
)
from moneyball.lab.queries import (
    get_model_leaderboard,
    get_entry_evaluations,
    compare_models,
)
from moneyball.lab.portfolio_research import (
    optimize_portfolio_maxmin,
)

__all__ = [
    # Models
    "InvestmentModel",
    "Entry",
    "Evaluation",
    # Writers
    "create_investment_model",
    "get_investment_model",
    "create_entry",
    "get_entry",
    "create_evaluation",
    # Queries
    "get_model_leaderboard",
    "get_entry_evaluations",
    "compare_models",
    # Portfolio research
    "optimize_portfolio_maxmin",
]
