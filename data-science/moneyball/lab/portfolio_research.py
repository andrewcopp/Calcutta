"""
Portfolio optimization research module for lab experiments.

This module provides optimizers for experimenting with different portfolio
allocation strategies. For production use, the Go DP allocator in
backend/internal/app/recommended_entry_bids/allocator.go provides exact
optimal solutions and should be preferred.

Optimizers in this module:
- optimize_portfolio_gekko: True MINLP using GEKKO's APOPT solver
- optimize_portfolio_maxmin: Max-min ROI optimization (conservative strategy)
"""
from __future__ import annotations

from typing import Any, Dict, List, Tuple

import numpy as np
import pandas as pd


def optimize_portfolio_gekko(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
) -> Tuple[pd.DataFrame, List[Dict[str, Any]]]:
    """
    True MINLP optimization using GEKKO's APOPT solver.

    Uses greedy solution as warm start, then tries to improve via MINLP.

    Args:
        teams_df: DataFrame with columns: team_key, expected_team_points,
                  predicted_team_total_bids
        budget_points: Total budget (typically 100)
        min_teams: Minimum number of teams (typically 3)
        max_teams: Maximum number of teams (typically 10)
        max_per_team_points: Maximum bid per team (typically 50)
        min_bid_points: Minimum bid per team (typically 1)

    Returns:
        Tuple of (chosen_df, portfolio_rows) matching the interface of other
        optimizers
    """
    from gekko import GEKKO

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

    n = len(teams_df)
    exp_pts = teams_df["expected_team_points"].fillna(0.0).values.astype(float)
    pred_markets = teams_df["predicted_team_total_bids"].fillna(0.0).values.astype(
        float
    )
    pred_markets = np.maximum(pred_markets, 0.0)

    # Get greedy solution as warm start
    greedy_bids = _get_greedy_warm_start(
        teams_df=teams_df,
        budget_points=budget_points,
        min_teams=min_teams,
        max_teams=max_teams,
        max_per_team_points=max_per_team_points,
        min_bid_points=min_bid_points,
    )

    m = GEKKO(remote=False)

    # Binary selection variables (warm started from greedy)
    x = []
    for i in range(n):
        init_val = 1 if greedy_bids[i] >= min_bid_points else 0
        var = m.Var(lb=0, ub=1, integer=True, value=init_val)
        x.append(var)

    # Integer bid variables (warm started from greedy)
    b = []
    for i in range(n):
        var = m.Var(
            lb=0, ub=max_per_team_points, integer=True, value=greedy_bids[i]
        )
        b.append(var)

    # Coupling constraints: bid is nonzero only if team is selected
    for i in range(n):
        m.Equation(b[i] >= min_bid_points * x[i])
        m.Equation(b[i] <= max_per_team_points * x[i])

    # Budget constraint: all bids sum to budget
    m.Equation(m.sum(b) == budget_points)

    # Team count constraints
    m.Equation(m.sum(x) >= min_teams)
    m.Equation(m.sum(x) <= max_teams)

    # Objective: maximize expected ownership-weighted return
    ownership = []
    for i in range(n):
        own = m.Intermediate(b[i] / (pred_markets[i] + b[i] + 1e-6))
        ownership.append(own)

    obj_terms = [exp_pts[i] * ownership[i] for i in range(n)]
    m.Maximize(m.sum(obj_terms))

    # Solver options
    m.options.SOLVER = 1  # APOPT solver for MINLP
    m.options.IMODE = 3  # Steady-state optimization
    m.options.MAX_ITER = 1000

    m.solve(disp=False)

    # Extract GEKKO solution
    gekko_bids = np.array([int(round(b[i].value[0])) for i in range(n)])

    # Calculate objective values for both solutions
    def calc_return(bids_arr: np.ndarray) -> float:
        ownership = bids_arr / (pred_markets + bids_arr + 1e-9)
        return float(np.sum(exp_pts * ownership))

    gekko_return = calc_return(gekko_bids)
    greedy_return = calc_return(greedy_bids)

    # Use the better solution (never do worse than greedy)
    if gekko_return >= greedy_return:
        bids = gekko_bids
    else:
        bids = greedy_bids

    selected_indices = np.where(bids >= min_bid_points)[0]
    chosen = teams_df.iloc[selected_indices].copy().reset_index(drop=True)
    chosen["bid_amount_points"] = bids[selected_indices]

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
        portfolio_rows.append(
            {
                "team_key": str(row["team_key"]),
                "bid_amount_points": int(row["bid_amount_points"]),
                "score": float(row.get("score", 0.0) or 0.0),
            }
        )

    return chosen, portfolio_rows


def optimize_portfolio_maxmin(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
) -> Tuple[pd.DataFrame, List[Dict[str, Any]]]:
    """
    Optimize portfolio using max-min approach.

    Maximizes the minimum ROI across all teams in the portfolio. This is a
    conservative strategy that ensures every dollar invested has a guaranteed
    minimum return.

    Args:
        teams_df: DataFrame with columns: team_key, expected_team_points,
                  predicted_team_total_bids
        budget_points: Total budget (typically 100)
        min_teams: Minimum number of teams (typically 3)
        max_teams: Maximum number of teams (typically 10)
        max_per_team_points: Maximum bid per team (typically 50)
        min_bid_points: Minimum bid per team (typically 1)

    Returns:
        Tuple of (chosen_df, portfolio_rows)
    """
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

    n_teams = len(teams_df)
    exp_pts = teams_df["expected_team_points"].fillna(0.0).values.astype(float)
    pred_markets = teams_df["predicted_team_total_bids"].fillna(0.0).values.astype(
        float
    )
    pred_markets = np.maximum(pred_markets, 0.0)

    # Binary search for the maximum achievable min ROI
    max_possible_roi = np.max(exp_pts / (pred_markets + min_bid_points + 1e-9))
    min_possible_roi = 0.5

    best_solution = None
    best_min_roi = 0.0

    for _ in range(20):
        target_roi = (min_possible_roi + max_possible_roi) / 2

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
            bids = solution
            actual_min_roi = np.min(
                exp_pts[bids > 0] / (pred_markets[bids > 0] + bids[bids > 0])
            )

            if actual_min_roi > best_min_roi:
                best_solution = bids
                best_min_roi = actual_min_roi

            min_possible_roi = target_roi
        else:
            max_possible_roi = target_roi

        if max_possible_roi - min_possible_roi < 0.01:
            break

    if best_solution is None:
        best_solution = np.zeros(n_teams)
        teams_to_select = min(max_teams, n_teams)
        bid_per_team = budget_points / teams_to_select
        best_solution[:teams_to_select] = bid_per_team

    bids = _enforce_discrete_constraints(
        bids=best_solution,
        budget=budget_points,
        min_teams=min_teams,
        max_teams=max_teams,
        min_bid=min_bid_points,
        max_per_team=max_per_team_points,
        exp_pts=exp_pts,
        pred_markets=pred_markets,
    )

    selected_indices = np.where(bids >= min_bid_points)[0]
    chosen = teams_df.iloc[selected_indices].copy().reset_index(drop=True)
    chosen["bid_amount_points"] = bids[selected_indices].astype(int)

    chosen_bids = bids[selected_indices]
    chosen_exp_pts = exp_pts[selected_indices]
    chosen_pred_markets = pred_markets[selected_indices]
    chosen["score"] = chosen_exp_pts / (chosen_pred_markets + chosen_bids + 1e-9)

    portfolio_rows = []
    for _, row in chosen.iterrows():
        portfolio_rows.append(
            {
                "team_key": str(row["team_key"]),
                "bid_amount_points": int(row["bid_amount_points"]),
                "score": float(row.get("score", 0.0) or 0.0),
            }
        )

    return chosen, portfolio_rows


# -----------------------------------------------------------------------------
# Helper functions
# -----------------------------------------------------------------------------


def _get_greedy_warm_start(
    *,
    teams_df: pd.DataFrame,
    budget_points: int,
    min_teams: int,
    max_teams: int,
    max_per_team_points: int,
    min_bid_points: int,
) -> np.ndarray:
    """Get greedy solution as warm start for GEKKO."""
    from moneyball.models.recommended_entry_bids import _optimize_portfolio_greedy

    n = len(teams_df)
    bids = np.zeros(n, dtype=int)

    try:
        score_col = (
            "score" if "score" in teams_df.columns else "expected_team_points"
        )
        greedy_result, _ = _optimize_portfolio_greedy(
            df=teams_df,
            score_col=score_col,
            budget=float(budget_points),
            min_teams=min_teams,
            max_teams=max_teams,
            max_per_team=float(max_per_team_points),
            min_bid=float(min_bid_points),
        )
        for _, row in greedy_result.iterrows():
            team_key = row["team_key"]
            bid = int(row["bid_amount_points"])
            idx = teams_df[teams_df["team_key"] == team_key].index
            if len(idx) > 0:
                bids[idx[0]] = bid
    except Exception:
        teams_to_select = min(max_teams, n)
        bid_per_team = budget_points // teams_to_select
        remainder = budget_points % teams_to_select
        for i in range(teams_to_select):
            bids[i] = bid_per_team + (1 if i < remainder else 0)

    return bids


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
) -> np.ndarray | None:
    """Find a portfolio where all teams have our_roi >= target_min_roi."""
    n_teams = len(exp_pts)

    max_bid_for_roi = np.zeros(n_teams)
    for i in range(n_teams):
        if target_min_roi > 0:
            max_bid_for_roi[i] = (
                exp_pts[i] - target_min_roi * pred_markets[i]
            ) / target_min_roi
        else:
            max_bid_for_roi[i] = max_per_team

    viable_teams = np.where(max_bid_for_roi >= min_bid)[0]

    if len(viable_teams) < min_teams:
        return None

    viable_teams_sorted = viable_teams[np.argsort(-max_bid_for_roi[viable_teams])]
    selected_teams = viable_teams_sorted[: min(max_teams, len(viable_teams_sorted))]

    if len(selected_teams) < min_teams:
        return None

    bids = np.zeros(n_teams)
    for i in selected_teams:
        bids[i] = min_bid

    remaining = budget - np.sum(bids)

    while remaining > 0:
        best_team = None
        best_capacity = 0

        for i in selected_teams:
            capacity = min(max_bid_for_roi[i] - bids[i], max_per_team - bids[i])
            if capacity > best_capacity:
                best_capacity = capacity
                best_team = i

        if best_team is None or best_capacity <= 0:
            break

        bids[best_team] += 1
        remaining -= 1

    if np.sum(bids) < budget * 0.95:
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
    bids = np.round(bids).astype(int)
    bids[bids < min_bid] = 0
    bids = np.minimum(bids, max_per_team)

    current_total = np.sum(bids)

    if current_total < budget:
        remaining = budget - current_total
        while remaining > 0:
            selected = np.where(bids > 0)[0]
            if len(selected) == 0:
                break

            marginal_roi = np.zeros(len(selected))
            for idx, i in enumerate(selected):
                if bids[i] >= max_per_team:
                    marginal_roi[idx] = -1e99
                else:
                    marginal_roi[idx] = exp_pts[i] / (pred_markets[i] + bids[i] + 1)

            best_idx = selected[np.argmax(marginal_roi)]
            if marginal_roi[np.argmax(marginal_roi)] <= -1e99:
                break

            bids[best_idx] += 1
            remaining -= 1

    elif current_total > budget:
        excess = current_total - budget
        while excess > 0:
            selected = np.where(bids > 0)[0]
            if len(selected) == 0:
                break

            marginal_roi = np.zeros(len(selected))
            for idx, i in enumerate(selected):
                if bids[i] <= min_bid:
                    marginal_roi[idx] = 1e99
                else:
                    marginal_roi[idx] = exp_pts[i] / (pred_markets[i] + bids[i])

            worst_idx = selected[np.argmin(marginal_roi)]
            if marginal_roi[np.argmin(marginal_roi)] >= 1e99:
                break

            bids[worst_idx] -= 1
            if bids[worst_idx] < min_bid:
                bids[worst_idx] = 0
            excess -= 1

    return bids
