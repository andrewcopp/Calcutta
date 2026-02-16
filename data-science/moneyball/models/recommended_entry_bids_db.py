"""
DB-first recommendation functions.

Helper functions for generating entry bid recommendations from simulations.
"""
from __future__ import annotations

import pandas as pd


def recommend_entry_bids_from_simulations(
    *,
    simulations_df: pd.DataFrame,
    year: int,
    strategy: str = "greedy",
    budget_points: int = 100,
    min_teams: int = 3,
    max_teams: int = 10,
    min_bid: int = 1,
    max_bid: int = 50,
) -> pd.DataFrame:
    """
    Generate recommended entry bids from simulation results.
    
    Args:
        simulations_df: DataFrame with simulation results
        strategy: Portfolio strategy (greedy, minlp, etc.)
        budget_points: Total budget in points
        min_teams: Minimum number of teams
        max_teams: Maximum number of teams
        min_bid: Minimum bid per team
        max_bid: Maximum bid per team
        
    Returns:
        DataFrame with recommendations (team_id, bid_amount_points, 
        expected_points, expected_roi)
    """
    # Calculate expected points for each team
    team_stats = simulations_df.groupby('team_id').agg({
        'points': 'mean',
        'school_name': 'first',
        'seed': 'first',
        'region': 'first',
    }).reset_index()
    
    team_stats.rename(columns={'points': 'expected_points'}, inplace=True)
    
    # Greedy strategy: allocate budget proportional to expected points
    team_stats = team_stats.sort_values('expected_points', ascending=False)
    
    # Select top teams
    selected_teams = team_stats.head(max_teams).copy()
    
    # Allocate budget proportionally
    total_expected = selected_teams['expected_points'].sum()
    
    if total_expected > 0:
        selected_teams['bid_amount_points'] = (
            selected_teams['expected_points'] / total_expected * budget_points
        ).round().astype(int)
    else:
        # Fallback: equal distribution
        per_team = budget_points // len(selected_teams)
        selected_teams['bid_amount_points'] = per_team
    
    # Enforce constraints
    selected_teams['bid_amount_points'] = selected_teams[
        'bid_amount_points'
    ].clip(lower=min_bid, upper=max_bid)
    
    # Adjust to meet budget
    total_bid = selected_teams['bid_amount_points'].sum()
    if total_bid > budget_points:
        # Scale down proportionally
        scale = budget_points / total_bid
        selected_teams['bid_amount_points'] = (
            selected_teams['bid_amount_points'] * scale
        ).round().astype(int)
    
    # Ensure minimum teams
    if len(selected_teams) < min_teams:
        # Add more teams
        remaining = team_stats[
            ~team_stats['team_id'].isin(selected_teams['team_id'])
        ].head(min_teams - len(selected_teams))
        
        remaining['bid_amount_points'] = min_bid
        selected_teams = pd.concat([selected_teams, remaining])
    
    # Calculate expected ROI (simplified)
    # ROI = expected_points / bid_amount_points
    selected_teams['expected_roi'] = (
        selected_teams['expected_points'] / 
        selected_teams['bid_amount_points'].replace(0, 1)
    )
    
    # Final adjustment to exactly meet budget
    current_total = selected_teams['bid_amount_points'].sum()
    if current_total < budget_points:
        # Add remaining budget to highest ROI team
        diff = budget_points - current_total
        idx = selected_teams['expected_roi'].idxmax()
        selected_teams.loc[idx, 'bid_amount_points'] += diff
        selected_teams.loc[idx, 'expected_roi'] = (
            selected_teams.loc[idx, 'expected_points'] /
            selected_teams.loc[idx, 'bid_amount_points']
        )
    
    return selected_teams[[
        'team_id',
        'bid_amount_points',
        'expected_points',
        'expected_roi',
        'school_name',
        'seed',
        'region',
    ]]
