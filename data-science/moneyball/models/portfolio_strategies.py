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
    variance_weight: float = 0.0,
) -> pd.DataFrame:
    """
    Greedy allocation: maximize expected value.

    This is the current default strategy. Greedily selects teams
    with highest expected value until budget is exhausted.

    Args:
        variance_weight: Weight for variance bonus (0.0 = pure expected value,
            >0 = favor high-variance teams for tail-risk optimization)
    """
    from moneyball.models.recommended_entry_bids import (
        _optimize_portfolio_greedy,
    )

    teams = teams_df.copy()

    if variance_weight > 0 and "std_team_points" in teams.columns:
        teams["variance_adjusted_score"] = teams.apply(
            lambda r: (
                (r.get("expected_team_points", 0) +
                 variance_weight * r.get("std_team_points", 0)) /
                (r.get("predicted_team_total_bids", 0) + min_bid_points)
                if (r.get("predicted_team_total_bids", 0) + min_bid_points) > 0
                else 0.0
            ),
            axis=1
        )
        score_col = "variance_adjusted_score"
    else:
        score_col = "score"

    chosen, _rows = _optimize_portfolio_greedy(
        df=teams,
        score_col=score_col,
        budget=float(budget_points),
        min_teams=min_teams,
        max_teams=max_teams,
        max_per_team=float(max_per_team_points),
        min_bid=float(min_bid_points),
    )

    if "score" not in chosen.columns:
        chosen["score"] = chosen.get(score_col, 0.0)

    return chosen[
        ["team_key", "bid_amount_points", "expected_team_points",
         "predicted_team_total_bids", "predicted_auction_share_of_pool",
         "score"]
    ]


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
        allocation = round(allocation)  # Round to integer
        
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
    
    # Distribute any remaining budget to existing selections
    if remaining_budget > 0 and len(selected) > 0:
        # Add to teams that haven't hit max_per_team_points
        for item in selected:
            if remaining_budget <= 0:
                break
            can_add = max_per_team_points - item["bid_amount_points"]
            add_amount = min(can_add, remaining_budget)
            item["bid_amount_points"] += add_amount
            remaining_budget -= add_amount
    
    # Convert to DataFrame and merge with original data to get required columns
    result_df = pd.DataFrame(selected)
    if len(result_df) == 0:
        return pd.DataFrame(columns=["team_key", "bid_amount_points", "expected_team_points", "predicted_team_total_bids", "predicted_auction_share_of_pool", "score"])
    
    # Merge with teams_df to get all required columns
    result_df = result_df.merge(teams_df, on="team_key", how="left")
    
    # Ensure score column exists
    if "score" not in result_df.columns:
        result_df["score"] = 0.0
    
    return result_df[
        ["team_key", "bid_amount_points", "expected_team_points",
         "predicted_team_total_bids", "predicted_auction_share_of_pool", "score"]
    ]


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
    
    # Ensure we have all required columns
    if "score" not in selected_teams.columns:
        selected_teams["score"] = 0.0
    
    return selected_teams[
        ["team_key", "bid_amount_points", "expected_team_points",
         "predicted_team_total_bids", "predicted_auction_share_of_pool", "score"]
    ]


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
        allocation = round(allocation)  # Round to integer
        
        if allocation >= min_bid_points:
            selected.append({
                "team_key": team["team_key"],
                "bid_amount_points": allocation,
            })
            remaining_budget -= allocation
    
    # Distribute any remaining budget to existing selections
    if remaining_budget > 0 and len(selected) > 0:
        # Add to teams that haven't hit max_per_team_points
        for item in selected:
            if remaining_budget <= 0:
                break
            can_add = max_per_team_points - item["bid_amount_points"]
            add_amount = min(can_add, remaining_budget)
            item["bid_amount_points"] += add_amount
            remaining_budget -= add_amount
    
    # Convert to DataFrame and merge with original data to get required columns
    result_df = pd.DataFrame(selected)
    if len(result_df) == 0:
        return pd.DataFrame(columns=["team_key", "bid_amount_points", "expected_team_points", "predicted_team_total_bids", "predicted_auction_share_of_pool", "score"])
    
    # Merge with teams_df to get all required columns
    result_df = result_df.merge(teams_df, on="team_key", how="left")
    
    # Ensure score column exists
    if "score" not in result_df.columns:
        result_df["score"] = 0.0
    
    return result_df[
        ["team_key", "bid_amount_points", "expected_team_points",
         "predicted_team_total_bids", "predicted_auction_share_of_pool", "score"]
    ]


def allocate_region_constrained(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
    max_teams_per_region: int = 1,
) -> pd.DataFrame:
    """
    Region-constrained greedy allocation: diversify across regions.
    
    Uses greedy allocation logic but enforces a maximum number of teams
    per region to avoid concentration risk. This prevents scenarios like
    picking 10 teams all from the same region or same side of bracket.
    
    Args:
        max_teams_per_region: Maximum teams to select from each region (default 1)
    """
    from moneyball.models.recommended_entry_bids import (
        _optimize_portfolio_greedy,
    )
    
    teams = teams_df.copy()
    
    # Check if region column exists
    if "region" not in teams.columns:
        # Fall back to unconstrained greedy if no region data
        return allocate_greedy(
            teams_df=teams_df,
            budget_points=budget_points,
            min_teams=min_teams,
            max_teams=max_teams,
            max_per_team_points=max_per_team_points,
            min_bid_points=min_bid_points,
        )
    
    # Sort by expected value (score)
    teams["expected_team_points"] = pd.to_numeric(
        teams["expected_team_points"], errors="coerce"
    ).fillna(0.0)
    teams["predicted_team_total_bids"] = pd.to_numeric(
        teams["predicted_team_total_bids"], errors="coerce"
    ).fillna(0.0)
    
    # Calculate score (expected points per dollar at min bid)
    teams["temp_score"] = teams.apply(
        lambda r: (
            r["expected_team_points"] / 
            (r["predicted_team_total_bids"] + min_bid_points)
            if (r["predicted_team_total_bids"] + min_bid_points) > 0 
            else 0.0
        ),
        axis=1
    )
    
    teams_sorted = teams.sort_values("temp_score", ascending=False)
    
    # Greedily select teams respecting region constraint
    selected = []
    region_counts = {}
    
    for _, team in teams_sorted.iterrows():
        region = team.get("region", "unknown")
        current_count = region_counts.get(region, 0)
        
        # Check if we can add this team
        if current_count < max_teams_per_region and len(selected) < max_teams:
            selected.append(team)
            region_counts[region] = current_count + 1
    
    # Ensure we have minimum teams
    if len(selected) < min_teams:
        # Relax constraint to meet minimum
        for _, team in teams_sorted.iterrows():
            if len(selected) >= min_teams:
                break
            if team["team_key"] not in [s["team_key"] for s in selected]:
                selected.append(team)
    
    # Convert to DataFrame
    selected_df = pd.DataFrame(selected)
    
    if len(selected_df) == 0:
        return pd.DataFrame(columns=[
            "team_key", "bid_amount_points", "expected_team_points",
            "predicted_team_total_bids", "predicted_auction_share_of_pool", "score"
        ])
    
    # Now run greedy optimizer on this constrained set
    chosen, _rows = _optimize_portfolio_greedy(
        df=selected_df,
        score_col="temp_score",
        budget=float(budget_points),
        min_teams=min(min_teams, len(selected_df)),
        max_teams=min(max_teams, len(selected_df)),
        max_per_team=float(max_per_team_points),
        min_bid=float(min_bid_points),
    )
    
    # Ensure score column exists
    if "score" not in chosen.columns:
        chosen["score"] = chosen.get("temp_score", 0.0)
    
    return chosen[
        ["team_key", "bid_amount_points", "expected_team_points",
         "predicted_team_total_bids", "predicted_auction_share_of_pool", "score"]
    ]


def allocate_one_per_region(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
) -> pd.DataFrame:
    """
    One team per region: maximum regional diversification.
    
    Selects exactly one team from each region (typically 4 regions in NCAA).
    This ensures no correlation from teams on the same side of the bracket.
    """
    return allocate_region_constrained(
        teams_df=teams_df,
        budget_points=budget_points,
        min_teams=min_teams,
        max_teams=max_teams,
        max_per_team_points=max_per_team_points,
        min_bid_points=min_bid_points,
        max_teams_per_region=1,
    )


def allocate_two_per_region(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
) -> pd.DataFrame:
    """
    Two teams per region: balanced regional diversification.
    
    Selects up to two teams from each region. Allows picking one from
    the top half and one from the bottom half of each region while
    maintaining diversification.
    """
    return allocate_region_constrained(
        teams_df=teams_df,
        budget_points=budget_points,
        min_teams=min_teams,
        max_teams=max_teams,
        max_per_team_points=max_per_team_points,
        min_bid_points=min_bid_points,
        max_teams_per_region=2,
    )


def allocate_variance_aware_light(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
) -> pd.DataFrame:
    """
    Variance-aware greedy (light): slight preference for high-variance teams.

    Uses variance_weight=0.3 to give a small bonus to teams with high
    variance (longshots). This helps capture tail-risk scenarios where
    favorites underperform.
    """
    return allocate_greedy(
        teams_df=teams_df,
        budget_points=budget_points,
        min_teams=min_teams,
        max_teams=max_teams,
        max_per_team_points=max_per_team_points,
        min_bid_points=min_bid_points,
        variance_weight=0.3,
    )


def allocate_variance_aware_medium(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
) -> pd.DataFrame:
    """
    Variance-aware greedy (medium): moderate preference for high-variance.

    Uses variance_weight=0.5 to balance expected value with variance.
    Favors portfolios that perform well in tail scenarios.
    """
    return allocate_greedy(
        teams_df=teams_df,
        budget_points=budget_points,
        min_teams=min_teams,
        max_teams=max_teams,
        max_per_team_points=max_per_team_points,
        min_bid_points=min_bid_points,
        variance_weight=0.5,
    )


def allocate_variance_aware_heavy(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
) -> pd.DataFrame:
    """
    Variance-aware greedy (heavy): strong preference for high-variance.

    Uses variance_weight=1.0 to heavily favor high-variance teams.
    Optimizes for winning in scenarios where favorites fail.
    """
    return allocate_greedy(
        teams_df=teams_df,
        budget_points=budget_points,
        min_teams=min_teams,
        max_teams=max_teams,
        max_per_team_points=max_per_team_points,
        min_bid_points=min_bid_points,
        variance_weight=1.0,
    )


STRATEGIES = {
    "greedy": allocate_greedy,
    "waterfill_equal": allocate_waterfill_equal,
    "kelly": allocate_kelly,
    "min_variance": allocate_min_variance,
    "max_sharpe": allocate_max_sharpe,
    "one_per_region": allocate_one_per_region,
    "two_per_region": allocate_two_per_region,
    "region_constrained": allocate_region_constrained,
    "variance_aware_light": allocate_variance_aware_light,
    "variance_aware_medium": allocate_variance_aware_medium,
    "variance_aware_heavy": allocate_variance_aware_heavy,
}


def get_strategy(strategy_name: str):
    """Get allocation strategy function by name."""
    if strategy_name not in STRATEGIES:
        raise ValueError(
            f"Unknown strategy: {strategy_name}. "
            f"Available: {list(STRATEGIES.keys())}"
        )
    return STRATEGIES[strategy_name]
