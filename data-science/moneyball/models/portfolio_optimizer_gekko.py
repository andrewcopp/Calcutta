"""
True MINLP portfolio optimizer using GEKKO.

This module implements a Mixed Integer Nonlinear Programming approach using
GEKKO's APOPT solver, which handles integer constraints natively without
post-processing hacks.

The optimization problem:
    Maximize: sum_i( expected_points[i] * bid[i] / (predicted_market[i] + bid[i]) )

    Subject to:
        - sum(bid[i]) = budget
        - min_teams <= sum(selected[i]) <= max_teams
        - bid[i] >= min_bid * selected[i]  (coupling)
        - bid[i] <= max_per_team * selected[i]  (coupling)
        - selected[i] in {0, 1}  (binary)
        - bid[i] integer
"""
from __future__ import annotations

import numpy as np
import pandas as pd
from typing import Tuple, List, Dict, Any


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
        score_col = "score" if "score" in teams_df.columns else "expected_team_points"
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
        # Fallback: uniform distribution across max_teams
        teams_to_select = min(max_teams, n)
        bid_per_team = budget_points // teams_to_select
        remainder = budget_points % teams_to_select
        for i in range(teams_to_select):
            bids[i] = bid_per_team + (1 if i < remainder else 0)

    return bids


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
        teams_df: DataFrame with columns: team_key, expected_team_points, predicted_team_total_bids
        budget_points: Total budget (typically 100)
        min_teams: Minimum number of teams (typically 3)
        max_teams: Maximum number of teams (typically 10)
        max_per_team_points: Maximum bid per team (typically 50)
        min_bid_points: Minimum bid per team (typically 1)

    Returns:
        Tuple of (chosen_df, portfolio_rows) matching the interface of other optimizers
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
    pred_markets = teams_df["predicted_team_total_bids"].fillna(0.0).values.astype(float)
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
        var = m.Var(lb=0, ub=max_per_team_points, integer=True, value=greedy_bids[i])
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
    # ownership[i] = b[i] / (pred_markets[i] + b[i])
    # Using intermediate variables for clarity
    ownership = []
    for i in range(n):
        # Add small epsilon to avoid division by zero
        own = m.Intermediate(b[i] / (pred_markets[i] + b[i] + 1e-6))
        ownership.append(own)

    # Objective: sum of expected_points * ownership
    obj_terms = [exp_pts[i] * ownership[i] for i in range(n)]
    m.Maximize(m.sum(obj_terms))

    # Solver options
    m.options.SOLVER = 1  # APOPT solver for MINLP
    m.options.IMODE = 3   # Steady-state optimization
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
        portfolio_rows.append({
            "team_key": str(row["team_key"]),
            "bid_amount_points": int(row["bid_amount_points"]),
            "score": float(row.get("score", 0.0) or 0.0),
        })

    return chosen, portfolio_rows
