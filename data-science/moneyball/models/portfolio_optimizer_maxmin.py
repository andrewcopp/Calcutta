"""
Max-min portfolio optimizer for Calcutta entry construction.

This module implements a max-min optimization approach that maximizes
the minimum ROI across all teams in the portfolio.

The optimization problem:
    Maximize: min(our_roi_1, our_roi_2, ..., our_roi_n)
    
    Where: our_roi_i = expected_points_i / (predicted_market_i + bid_i)
    
    Subject to:
        - Σ bid_i = budget
        - min_teams ≤ Σ(bid_i > 0) ≤ max_teams
        - min_bid ≤ bid_i ≤ max_per_team (for selected teams)

This is reformulated as:
    Maximize: t
    
    Subject to:
        - t ≤ expected_points_i / (predicted_market_i + bid_i) for all i
        - (all other constraints)
"""
from __future__ import annotations

import numpy as np
import pandas as pd
from scipy.optimize import minimize, LinearConstraint
from typing import Tuple, List, Dict, Any


def optimize_portfolio_maxmin(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
    initial_solution: str = "uniform",
    max_iterations: int = 1000,
) -> Tuple[pd.DataFrame, List[Dict[str, Any]]]:
    """
    Optimize portfolio using max-min approach.
    
    Maximizes the minimum ROI across all teams in the portfolio.
    
    Args:
        teams_df: DataFrame with columns: team_key, expected_team_points,
                  predicted_team_total_bids
        budget_points: Total budget (typically 100)
        min_teams: Minimum number of teams (typically 3)
        max_teams: Maximum number of teams (typically 10)
        max_per_team_points: Maximum bid per team (typically 50)
        min_bid_points: Minimum bid per team (typically 1)
        initial_solution: How to initialize ("uniform" or "greedy")
        max_iterations: Maximum solver iterations
        
    Returns:
        Tuple of (chosen_df, portfolio_rows)
    """
    # Validate inputs
    required_cols = ["team_key", "expected_team_points",
                     "predicted_team_total_bids"]
    missing = [c for c in required_cols if c not in teams_df.columns]
    if missing:
        raise ValueError(f"Missing required columns: {', '.join(missing)}")
    
    if budget_points <= 0:
        raise ValueError("budget_points must be positive")
    if min_teams <= 0 or max_teams <= 0:
        raise ValueError("min_teams and max_teams must be positive")
    if min_teams > max_teams:
        raise ValueError("min_teams cannot exceed max_teams")
    if min_bid_points <= 0:
        raise ValueError("min_bid_points must be positive")
    
    # Extract data
    n_teams = len(teams_df)
    exp_pts = teams_df["expected_team_points"].fillna(0.0).values.astype(float)
    pred_markets = (teams_df["predicted_team_total_bids"]
                    .fillna(0.0).values.astype(float))
    pred_markets = np.maximum(pred_markets, 0.0)
    
    # Use iterative approach: try different target min ROIs
    # Start with a reasonable range based on the data
    max_possible_roi = np.max(exp_pts / (pred_markets + min_bid_points + 1e-9))
    min_possible_roi = 0.5
    
    best_solution = None
    best_min_roi = 0.0
    
    # Binary search for the maximum achievable min ROI
    for iteration in range(20):
        target_roi = (min_possible_roi + max_possible_roi) / 2
        
        # Try to find a portfolio with min ROI >= target_roi
        solution = _find_portfolio_with_min_roi(
            exp_pts=exp_pts,
            pred_markets=pred_markets,
            target_min_roi=target_roi,
            budget=budget_points,
            min_teams=min_teams,
            max_teams=max_teams,
            max_per_team=max_per_team_points,
            min_bid=min_bid_points,
        )
        
        if solution is not None:
            # Feasible solution found
            bids = solution
            actual_min_roi = np.min(
                exp_pts[bids > 0] / (pred_markets[bids > 0] + bids[bids > 0])
            )
            
            if actual_min_roi > best_min_roi:
                best_solution = bids
                best_min_roi = actual_min_roi
            
            # Try higher target
            min_possible_roi = target_roi
        else:
            # No feasible solution, try lower target
            max_possible_roi = target_roi
        
        # Convergence check
        if max_possible_roi - min_possible_roi < 0.01:
            break
    
    if best_solution is None:
        # Fallback: use uniform allocation
        best_solution = np.zeros(n_teams)
        teams_to_select = min(max_teams, n_teams)
        bid_per_team = budget_points / teams_to_select
        best_solution[:teams_to_select] = bid_per_team
    
    bids = best_solution
    
    # Post-processing: enforce discrete constraints
    bids = _enforce_discrete_constraints(
        bids=bids,
        budget=budget_points,
        min_teams=min_teams,
        max_teams=max_teams,
        min_bid=min_bid_points,
        max_per_team=max_per_team_points,
        exp_pts=exp_pts,
        pred_markets=pred_markets,
    )
    
    # Build output DataFrame
    selected_indices = np.where(bids >= min_bid_points)[0]
    chosen = teams_df.iloc[selected_indices].copy().reset_index(drop=True)
    chosen["bid_amount_points"] = bids[selected_indices].astype(int)
    
    # Calculate score (our ROI at final bid)
    chosen_bids = bids[selected_indices]
    chosen_exp_pts = exp_pts[selected_indices]
    chosen_pred_markets = pred_markets[selected_indices]
    chosen["score"] = chosen_exp_pts / (chosen_pred_markets + chosen_bids + 1e-9)
    
    # Build portfolio rows
    portfolio_rows = []
    for _, row in chosen.iterrows():
        portfolio_rows.append({
            "team_key": str(row["team_key"]),
            "bid_amount_points": int(row["bid_amount_points"]),
            "score": float(row.get("score", 0.0) or 0.0),
        })
    
    return chosen, portfolio_rows


def _find_portfolio_with_min_roi(
    *,
    exp_pts: np.ndarray,
    pred_markets: np.ndarray,
    target_min_roi: float,
    budget: int,
    min_teams: int,
    max_teams: int,
    max_per_team: int,
    min_bid: int,
) -> np.ndarray:
    """
    Find a portfolio where all teams have our_roi >= target_min_roi.
    
    For each team i, we need:
        exp_pts_i / (pred_markets_i + bid_i) >= target_min_roi
    
    Rearranging:
        exp_pts_i >= target_min_roi * (pred_markets_i + bid_i)
        exp_pts_i >= target_min_roi * pred_markets_i + target_min_roi * bid_i
        exp_pts_i - target_min_roi * pred_markets_i >= target_min_roi * bid_i
        bid_i <= (exp_pts_i - target_min_roi * pred_markets_i) / target_min_roi
    
    So each team has a maximum bid to maintain the target ROI.
    """
    n_teams = len(exp_pts)
    
    # Calculate max bid for each team to maintain target ROI
    max_bid_for_roi = np.zeros(n_teams)
    for i in range(n_teams):
        if target_min_roi > 0:
            max_bid_for_roi[i] = (
                (exp_pts[i] - target_min_roi * pred_markets[i]) / target_min_roi
            )
        else:
            max_bid_for_roi[i] = max_per_team
    
    # Filter teams that can achieve target ROI with at least min_bid
    viable_teams = np.where(max_bid_for_roi >= min_bid)[0]
    
    if len(viable_teams) < min_teams:
        return None  # Not enough viable teams
    
    # Greedy allocation: pick teams with best capacity (highest max_bid_for_roi)
    viable_teams_sorted = viable_teams[np.argsort(-max_bid_for_roi[viable_teams])]
    
    # Select up to max_teams
    selected_teams = viable_teams_sorted[:min(max_teams, len(viable_teams_sorted))]
    
    if len(selected_teams) < min_teams:
        return None
    
    # Allocate budget to selected teams
    bids = np.zeros(n_teams)
    
    # Start with min_bid for each selected team
    for i in selected_teams:
        bids[i] = min_bid
    
    remaining = budget - np.sum(bids)
    
    # Distribute remaining budget proportionally to capacity
    while remaining > 0:
        # Find team with most capacity remaining
        best_team = None
        best_capacity = 0
        
        for i in selected_teams:
            capacity = min(
                max_bid_for_roi[i] - bids[i],
                max_per_team - bids[i]
            )
            if capacity > best_capacity:
                best_capacity = capacity
                best_team = i
        
        if best_team is None or best_capacity <= 0:
            break
        
        # Add $1 to best team
        bids[best_team] += 1
        remaining -= 1
    
    # Check if we used the full budget
    if np.sum(bids) < budget * 0.95:  # Allow 5% slack
        return None
    
    return bids


def _enforce_discrete_constraints(
    *,
    bids: np.ndarray,
    budget: int,
    min_teams: int,
    max_teams: int,
    min_bid: int,
    max_per_team: int,
    exp_pts: np.ndarray,
    pred_markets: np.ndarray,
) -> np.ndarray:
    """Enforce discrete constraints on solution."""
    # Round to integers
    bids = np.round(bids).astype(int)
    
    # Set small bids to 0
    bids[bids < min_bid] = 0
    
    # Cap at max_per_team
    bids = np.minimum(bids, max_per_team)
    
    # Adjust to satisfy budget constraint
    current_total = np.sum(bids)
    
    if current_total < budget:
        # Add remaining budget to teams with best marginal ROI
        remaining = budget - current_total
        while remaining > 0:
            selected = np.where(bids > 0)[0]
            if len(selected) == 0:
                break
            
            # Calculate marginal ROI for adding $1
            marginal_roi = np.zeros(len(selected))
            for idx, i in enumerate(selected):
                if bids[i] >= max_per_team:
                    marginal_roi[idx] = -1e99
                else:
                    marginal_roi[idx] = (
                        exp_pts[i] / (pred_markets[i] + bids[i] + 1)
                    )
            
            best_idx = selected[np.argmax(marginal_roi)]
            if marginal_roi[np.argmax(marginal_roi)] <= -1e99:
                break
            
            bids[best_idx] += 1
            remaining -= 1
    
    elif current_total > budget:
        # Remove excess budget from teams with worst marginal ROI
        excess = current_total - budget
        while excess > 0:
            selected = np.where(bids > 0)[0]
            if len(selected) == 0:
                break
            
            # Calculate marginal ROI for removing $1
            marginal_roi = np.zeros(len(selected))
            for idx, i in enumerate(selected):
                if bids[i] <= min_bid:
                    marginal_roi[idx] = 1e99
                else:
                    marginal_roi[idx] = (
                        exp_pts[i] / (pred_markets[i] + bids[i])
                    )
            
            worst_idx = selected[np.argmin(marginal_roi)]
            if marginal_roi[np.argmin(marginal_roi)] >= 1e99:
                break
            
            bids[worst_idx] -= 1
            if bids[worst_idx] < min_bid:
                bids[worst_idx] = 0
            excess -= 1
    
    return bids
