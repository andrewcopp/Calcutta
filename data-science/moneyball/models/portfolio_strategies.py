"""
Portfolio allocation strategies for Calcutta entry optimization.

Different approaches to allocating budget across teams:
- greedy: Current approach - maximize expected value
- waterfill_equal: Baseline from OLD_ALGORITHMS - equal allocation
- kelly: Kelly criterion for optimal betting
- min_variance: Conservative approach minimizing variance
- max_sharpe: Optimize risk/reward ratio
"""
from __future__ import annotations

import pandas as pd
import numpy as np


def allocate_greedy(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
) -> pd.DataFrame:
    """
    Greedy allocation: maximize expected value.
    
    This is the current default strategy. Greedily selects teams
    with highest expected value until budget is exhausted.
    """
    from moneyball.models.recommended_entry_bids import recommend_entry_bids
    
    return recommend_entry_bids(
        teams=teams_df,
        budget_points=budget_points,
        min_teams=min_teams,
        max_teams=max_teams,
        max_per_team_points=max_per_team_points,
        min_bid_points=min_bid_points,
    )


def allocate_waterfill_equal(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
) -> pd.DataFrame:
    """
    Waterfill equal allocation: baseline strategy.
    
    Allocates budget equally across top teams by expected value,
    respecting constraints. This is the preserved OLD_ALGORITHMS approach.
    """
    # Sort teams by expected value
    teams_sorted = teams_df.sort_values(
        "expected_team_points", ascending=False
    ).copy()
    
    # Determine number of teams to select
    n_teams = min(max_teams, len(teams_sorted))
    n_teams = max(n_teams, min_teams)
    
    # Equal allocation
    equal_allocation = budget_points / n_teams
    
    # Cap at max per team
    allocation_per_team = min(equal_allocation, max_per_team_points)
    
    # Floor at min bid
    allocation_per_team = max(allocation_per_team, min_bid_points)
    
    # Select top teams
    selected_teams = teams_sorted.head(n_teams).copy()
    
    # Allocate equally
    selected_teams["bid_amount_points"] = allocation_per_team
    
    # Adjust if we're over budget
    total_allocated = allocation_per_team * n_teams
    if total_allocated > budget_points:
        # Scale down proportionally
        scale_factor = budget_points / total_allocated
        selected_teams["bid_amount_points"] *= scale_factor
        selected_teams["bid_amount_points"] = selected_teams[
            "bid_amount_points"
        ].round()
    
    # Ensure we have all required columns
    if "score" not in selected_teams.columns:
        selected_teams["score"] = 0.0
    
    return selected_teams[
        ["team_key", "bid_amount_points", "expected_team_points",
         "predicted_team_total_bids", "predicted_auction_share_of_pool", "score"]
    ]


def allocate_kelly(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
) -> pd.DataFrame:
    """
    Kelly criterion allocation: optimal betting strategy.
    
    Allocates based on Kelly criterion which maximizes long-term
    growth rate. More aggressive on high-edge opportunities.
    
    Kelly fraction = (expected_value - cost) / variance
    """
    teams = teams_df.copy()
    
    # Calculate Kelly fractions
    # For simplicity, assume variance proportional to expected_team_points
    # In reality, we'd want actual variance from simulations
    teams["kelly_fraction"] = (
        teams["expected_team_points"] / teams["expected_team_points"].sum()
    )
    
    # Sort by Kelly fraction
    teams_sorted = teams.sort_values("kelly_fraction", ascending=False)
    
    # Allocate based on Kelly fractions
    selected = []
    remaining_budget = budget_points
    
    for _, team in teams_sorted.iterrows():
        if len(selected) >= max_teams:
            break
        
        # Kelly allocation
        kelly_allocation = (
            team["kelly_fraction"] * budget_points
        )
        
        # Apply constraints
        allocation = min(kelly_allocation, max_per_team_points)
        allocation = max(allocation, min_bid_points)
        allocation = min(allocation, remaining_budget)
        
        if allocation >= min_bid_points:
            selected.append({
                "team_key": team["team_key"],
                "bid_amount_points": allocation,
            })
            remaining_budget -= allocation
    
    # Ensure minimum teams
    while len(selected) < min_teams and remaining_budget >= min_bid_points:
        # Add next best team
        remaining_teams = teams_sorted[
            ~teams_sorted["team_key"].isin(
                [s["team_key"] for s in selected]
            )
        ]
        if len(remaining_teams) == 0:
            break
        
        next_team = remaining_teams.iloc[0]
        allocation = min(min_bid_points, remaining_budget)
        selected.append({
            "team_key": next_team["team_key"],
            "bid_amount_points": allocation,
        })
        remaining_budget -= allocation
    
    return pd.DataFrame(selected)


def allocate_min_variance(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
) -> pd.DataFrame:
    """
    Minimum variance allocation: conservative strategy.
    
    Spreads budget across more teams to reduce variance.
    Prefers diversification over concentration.
    """
    teams_sorted = teams_df.sort_values(
        "expected_team_points", ascending=False
    ).copy()
    
    # Use maximum teams for diversification
    n_teams = min(max_teams, len(teams_sorted))
    
    # Equal allocation for minimum variance
    allocation_per_team = budget_points / n_teams
    allocation_per_team = min(allocation_per_team, max_per_team_points)
    allocation_per_team = max(allocation_per_team, min_bid_points)
    
    selected_teams = teams_sorted.head(n_teams).copy()
    selected_teams["bid_amount_points"] = allocation_per_team
    
    # Adjust to budget
    total_allocated = allocation_per_team * n_teams
    if total_allocated > budget_points:
        scale_factor = budget_points / total_allocated
        selected_teams["bid_amount_points"] *= scale_factor
        selected_teams["bid_amount_points"] = selected_teams[
            "bid_amount_points"
        ].round()
    
    return selected_teams[["team_key", "bid_amount_points"]]


def allocate_max_sharpe(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
) -> pd.DataFrame:
    """
    Maximum Sharpe ratio allocation: optimize risk/reward.
    
    Allocates to maximize expected return per unit of risk.
    Balances concentration and diversification.
    """
    teams = teams_df.copy()
    
    # Calculate Sharpe-like metric
    # Sharpe = (expected_return - risk_free_rate) / std_dev
    # For simplicity, assume std_dev proportional to sqrt(expected_team_points)
    teams["sharpe"] = teams["expected_team_points"] / np.sqrt(
        teams["expected_team_points"] + 1
    )
    
    teams_sorted = teams.sort_values("sharpe", ascending=False)
    
    # Allocate proportional to Sharpe ratio
    selected = []
    remaining_budget = budget_points
    
    # Select top teams by Sharpe
    n_teams = min(max_teams, len(teams_sorted))
    top_teams = teams_sorted.head(n_teams)
    
    # Allocate proportional to Sharpe ratio
    total_sharpe = top_teams["sharpe"].sum()
    
    for _, team in top_teams.iterrows():
        if remaining_budget < min_bid_points:
            break
        
        # Proportional allocation
        allocation = (team["sharpe"] / total_sharpe) * budget_points
        
        # Apply constraints
        allocation = min(allocation, max_per_team_points)
        allocation = max(allocation, min_bid_points)
        allocation = min(allocation, remaining_budget)
        
        if allocation >= min_bid_points:
            selected.append({
                "team_key": team["team_key"],
                "bid_amount_points": allocation,
            })
            remaining_budget -= allocation
    
    return pd.DataFrame(selected)


STRATEGIES = {
    "greedy": allocate_greedy,
    "waterfill_equal": allocate_waterfill_equal,
    "kelly": allocate_kelly,
    "min_variance": allocate_min_variance,
    "max_sharpe": allocate_max_sharpe,
}


def get_strategy(strategy_name: str):
    """Get allocation strategy function by name."""
    if strategy_name not in STRATEGIES:
        raise ValueError(
            f"Unknown strategy: {strategy_name}. "
            f"Available: {list(STRATEGIES.keys())}"
        )
    return STRATEGIES[strategy_name]
