"""
MINLP-based portfolio optimizer for Calcutta entry construction.

This module implements a Mixed Integer Nonlinear Programming approach to
portfolio optimization that avoids the local maximum trap of the greedy algorithm.

The optimization problem:
    Maximize: Σ expected_points_i × (bid_i / (predicted_market_i + bid_i))
    
    Subject to:
        - Σ bid_i = budget
        - min_teams ≤ Σ(bid_i > 0) ≤ max_teams
        - min_bid ≤ bid_i ≤ max_per_team (for selected teams)
        - bid_i = 0 (for unselected teams)
"""
from __future__ import annotations

import numpy as np
import pandas as pd
from scipy.optimize import minimize, NonlinearConstraint, LinearConstraint
from typing import Tuple, List, Dict, Any


def _objective(bids: np.ndarray, exp_pts: np.ndarray, pred_markets: np.ndarray) -> float:
    """
    Objective function: negative total expected return (for minimization).
    
    Total return = Σ expected_points_i × ownership_i
    where ownership_i = bid_i / (predicted_market_i + bid_i)
    
    We return negative because scipy.optimize.minimize minimizes.
    """
    ownership = bids / (pred_markets + bids + 1e-9)  # Add epsilon to avoid division by zero
    total_return = np.sum(exp_pts * ownership)
    return -total_return  # Negative for minimization


def _objective_gradient(bids: np.ndarray, exp_pts: np.ndarray, pred_markets: np.ndarray) -> np.ndarray:
    """
    Gradient of objective function with respect to bids.
    
    d/d(bid_i) [exp_pts_i × bid_i / (pred_market_i + bid_i)]
    = exp_pts_i × pred_market_i / (pred_market_i + bid_i)^2
    
    Return negative gradient for minimization.
    """
    denom = (pred_markets + bids + 1e-9) ** 2
    gradient = exp_pts * pred_markets / denom
    return -gradient  # Negative for minimization


def optimize_portfolio_minlp(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
    initial_solution: str = "greedy",
    max_iterations: int = 1000,
) -> Tuple[pd.DataFrame, List[Dict[str, Any]]]:
    """
    Optimize portfolio using MINLP approach.
    
    Args:
        teams_df: DataFrame with columns: team_key, expected_team_points, predicted_team_total_bids
        budget_points: Total budget (typically 100)
        min_teams: Minimum number of teams (typically 3)
        max_teams: Maximum number of teams (typically 10)
        max_per_team_points: Maximum bid per team (typically 50)
        min_bid_points: Minimum bid per team (typically 1)
        initial_solution: How to initialize ("greedy", "uniform", or "random")
        max_iterations: Maximum solver iterations
        
    Returns:
        Tuple of (chosen_df, portfolio_rows) matching the interface of _optimize_portfolio_greedy
    """
    # Validate inputs
    required_cols = ["team_key", "expected_team_points", "predicted_team_total_bids"]
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
    pred_markets = teams_df["predicted_team_total_bids"].fillna(0.0).values.astype(float)
    pred_markets = np.maximum(pred_markets, 0.0)  # Ensure non-negative
    
    # Generate initial solution
    if initial_solution == "greedy":
        # Use current greedy algorithm as starting point
        from moneyball.models.recommended_entry_bids import _optimize_portfolio_greedy
        try:
            greedy_result, _ = _optimize_portfolio_greedy(
                df=teams_df,
                score_col="score" if "score" in teams_df.columns else "expected_team_points",
                budget=float(budget_points),
                min_teams=min_teams,
                max_teams=max_teams,
                max_per_team=float(max_per_team_points),
                min_bid=float(min_bid_points),
            )
            x0 = np.zeros(n_teams)
            for idx, row in greedy_result.iterrows():
                team_key = row["team_key"]
                bid = row["bid_amount_points"]
                team_idx = teams_df[teams_df["team_key"] == team_key].index[0]
                x0[team_idx] = float(bid)
        except Exception:
            # Fallback to uniform if greedy fails
            x0 = np.zeros(n_teams)
            teams_to_select = min(max_teams, n_teams)
            bid_per_team = budget_points / teams_to_select
            x0[:teams_to_select] = bid_per_team
    elif initial_solution == "uniform":
        # Distribute budget uniformly across max_teams
        x0 = np.zeros(n_teams)
        teams_to_select = min(max_teams, n_teams)
        bid_per_team = budget_points / teams_to_select
        x0[:teams_to_select] = bid_per_team
    else:  # random
        x0 = np.random.uniform(0, max_per_team_points, n_teams)
        x0 = x0 * (budget_points / np.sum(x0))  # Normalize to budget
    
    # Define constraints
    # 1. Budget constraint: Σ bid_i = budget
    budget_constraint = LinearConstraint(
        A=np.ones(n_teams),
        lb=budget_points,
        ub=budget_points
    )
    
    # 2. Bounds: 0 ≤ bid_i ≤ max_per_team
    bounds = [(0, max_per_team_points) for _ in range(n_teams)]
    
    # 3. Team count constraint: min_teams ≤ Σ(bid_i > 0) ≤ max_teams
    # This is tricky because it's discrete. We'll handle it in post-processing.
    
    # Run optimization
    result = minimize(
        fun=lambda x: _objective(x, exp_pts, pred_markets),
        x0=x0,
        method='SLSQP',
        jac=lambda x: _objective_gradient(x, exp_pts, pred_markets),
        constraints=[budget_constraint],
        bounds=bounds,
        options={'maxiter': max_iterations, 'ftol': 1e-9}
    )
    
    if not result.success:
        # If optimization failed, fall back to initial solution
        print(f"Warning: MINLP optimization did not converge: {result.message}")
        bids = x0
    else:
        bids = result.x
    
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
    
    # Calculate score (expected return per dollar at final bid)
    chosen_bids = bids[selected_indices]
    chosen_exp_pts = exp_pts[selected_indices]
    chosen_pred_markets = pred_markets[selected_indices]
    chosen_ownership = chosen_bids / (chosen_pred_markets + chosen_bids + 1e-9)
    chosen_return = chosen_exp_pts * chosen_ownership
    chosen["score"] = chosen_return / (chosen_bids + 1e-9)
    
    # Build portfolio rows
    portfolio_rows = []
    for _, row in chosen.iterrows():
        portfolio_rows.append({
            "team_key": str(row["team_key"]),
            "bid_amount_points": int(row["bid_amount_points"]),
            "score": float(row.get("score", 0.0) or 0.0),
        })
    
    return chosen, portfolio_rows


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
    """
    Enforce discrete constraints on continuous solution.
    
    1. Round bids to integers
    2. Set bids < min_bid to 0
    3. Cap bids at max_per_team
    4. Adjust to satisfy min_teams ≤ num_teams ≤ max_teams
    5. Adjust to satisfy budget constraint
    """
    # Round to integers
    bids = np.round(bids).astype(int)
    
    # Set small bids to 0
    bids[bids < min_bid] = 0
    
    # Cap at max_per_team
    bids = np.minimum(bids, max_per_team)
    
    # Count selected teams
    num_teams = np.sum(bids > 0)
    
    # If too few teams, add teams with best marginal return
    while num_teams < min_teams:
        # Find best unselected team
        unselected = np.where(bids == 0)[0]
        if len(unselected) == 0:
            break
        
        marginal_returns = np.zeros(len(unselected))
        for i, idx in enumerate(unselected):
            marginal_returns[i] = exp_pts[idx] * min_bid / (pred_markets[idx] + min_bid + 1e-9)
        
        best_idx = unselected[np.argmax(marginal_returns)]
        bids[best_idx] = min_bid
        num_teams += 1
    
    # If too many teams, remove teams with worst marginal return
    while num_teams > max_teams:
        selected = np.where(bids > 0)[0]
        if len(selected) == 0:
            break
        
        marginal_returns = np.zeros(len(selected))
        for i, idx in enumerate(selected):
            # Marginal return of REMOVING this team
            marginal_returns[i] = exp_pts[idx] * bids[idx] / (pred_markets[idx] + bids[idx] + 1e-9)
        
        worst_idx = selected[np.argmin(marginal_returns)]
        bids[worst_idx] = 0
        num_teams -= 1
    
    # Adjust to satisfy budget constraint
    current_total = np.sum(bids)
    
    if current_total < budget:
        # Add remaining budget to teams with best marginal return
        remaining = budget - current_total
        while remaining > 0:
            selected = np.where(bids > 0)[0]
            if len(selected) == 0:
                break
            
            # Calculate marginal return for adding $1 to each team
            marginal_returns = np.zeros(len(selected))
            for i, idx in enumerate(selected):
                if bids[idx] >= max_per_team:
                    marginal_returns[i] = -1e99  # Can't add more
                else:
                    marginal_returns[i] = (
                        exp_pts[idx] * pred_markets[idx] / 
                        ((pred_markets[idx] + bids[idx] + 1) ** 2 + 1e-9)
                    )
            
            best_idx = selected[np.argmax(marginal_returns)]
            if marginal_returns[np.argmax(marginal_returns)] <= -1e99:
                break  # All teams at max
            
            bids[best_idx] += 1
            remaining -= 1
    
    elif current_total > budget:
        # Remove excess budget from teams with worst marginal return
        excess = current_total - budget
        while excess > 0:
            selected = np.where(bids > 0)[0]
            if len(selected) == 0:
                break
            
            # Calculate marginal return for removing $1 from each team
            marginal_returns = np.zeros(len(selected))
            for i, idx in enumerate(selected):
                if bids[idx] <= min_bid:
                    marginal_returns[i] = 1e99  # Can't remove (would violate min_bid)
                else:
                    marginal_returns[i] = (
                        exp_pts[idx] * pred_markets[idx] / 
                        ((pred_markets[idx] + bids[idx]) ** 2 + 1e-9)
                    )
            
            worst_idx = selected[np.argmin(marginal_returns)]
            if marginal_returns[np.argmin(marginal_returns)] >= 1e99:
                break  # All teams at min_bid
            
            bids[worst_idx] -= 1
            if bids[worst_idx] < min_bid:
                bids[worst_idx] = 0
            excess -= 1
    
    return bids
