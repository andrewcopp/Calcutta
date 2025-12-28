from __future__ import annotations

from typing import Dict, List, Optional, Tuple

import pandas as pd

import backtest_scaffold


def allocate_expected_points(
    *,
    df: pd.DataFrame,
    budget: float,
    min_bid: float,
    max_per_team: float,
    step: float = 1.0,
) -> List[float]:
    if budget <= 0:
        return []
    if min_bid <= 0:
        raise ValueError("min_bid must be positive")
    if max_per_team <= 0:
        raise ValueError("max_per_team must be positive")
    if step <= 0:
        raise ValueError("step must be positive")

    b_int = int(round(float(budget)))
    min_int = int(round(float(min_bid)))
    max_int = int(round(float(max_per_team)))
    if abs(float(budget) - float(b_int)) > 1e-9:
        raise ValueError("budget must be an integer number of dollars")
    if abs(float(min_bid) - float(min_int)) > 1e-9:
        raise ValueError("min_bid must be an integer number of dollars")
    if abs(float(max_per_team) - float(max_int)) > 1e-9:
        raise ValueError("max_per_team must be an integer number of dollars")
    if min_int <= 0:
        raise ValueError("min_bid must be positive")
    if max_int <= 0:
        raise ValueError("max_per_team must be positive")

    step = 1.0

    k = int(len(df))
    if k <= 0:
        return []

    base_spend = int(k) * int(min_int)
    if float(base_spend) - float(b_int) > 1e-9:
        raise ValueError("budget too small for min_bid across selected teams")

    bids: List[float] = [float(min_int) for _ in range(k)]
    remaining = float(b_int - base_spend)

    exp_pts = (
        pd.to_numeric(df["expected_team_points"], errors="coerce")
        .fillna(0.0)
        .tolist()
    )
    market_totals = (
        pd.to_numeric(df["predicted_team_total_bids"], errors="coerce")
        .fillna(0.0)
        .tolist()
    )

    while remaining > 1e-9:
        best_i: Optional[int] = None
        best_val = -1e99
        for i in range(k):
            if bids[i] + step - max_per_team > 1e-9:
                continue
            m = float(market_totals[i])
            if m < 0:
                m = 0.0
            b0 = float(bids[i])
            b1 = float(b0 + step)

            denom0 = m + b0
            denom1 = m + b1
            s0 = (b0 / denom0) if denom0 > 0 else 0.0
            s1 = (b1 / denom1) if denom1 > 0 else 0.0
            delta = float(exp_pts[i]) * (s1 - s0)
            v = (delta / float(step)) if step > 0 else 0.0
            if v > best_val:
                best_val = v
                best_i = i

        if best_i is None:
            break

        inc = step if remaining >= step else remaining
        if bids[best_i] + inc - float(max_int) > 1e-9:
            inc = max(0.0, float(max_int) - bids[best_i])
        if inc <= 1e-12:
            break

        bids[best_i] += float(inc)
        remaining -= float(inc)

    bids_int = [int(round(float(x))) for x in bids]
    if sum(bids_int) != int(b_int):
        delta = int(b_int) - int(sum(bids_int))
        if delta != 0:
            bids_int[0] += delta
    if any(x < int(min_int) for x in bids_int):
        raise ValueError("allocation violates min_bid constraint")
    if any(x > int(max_int) for x in bids_int):
        raise ValueError("allocation violates max_per_team constraint")
    return [float(x) for x in bids_int]


def select_portfolio(
    *,
    df: pd.DataFrame,
    score_col: str,
    budget: float,
    min_teams: int,
    max_teams: int,
    max_per_team: float,
    min_bid: float,
    allocation_mode: str,
) -> Tuple[pd.DataFrame, List[Dict[str, object]]]:
    if min_teams <= 0 or max_teams <= 0:
        raise ValueError("min_teams and max_teams must be positive")
    if min_teams > max_teams:
        raise ValueError("min_teams cannot exceed max_teams")
    if budget <= 0:
        raise ValueError("budget must be positive")
    if min_bid <= 0:
        raise ValueError("min_bid must be positive")

    b_int = int(round(float(budget)))
    min_int = int(round(float(min_bid)))
    max_int = int(round(float(max_per_team)))
    if abs(float(budget) - float(b_int)) > 1e-9:
        raise ValueError("budget must be an integer number of dollars")
    if abs(float(min_bid) - float(min_int)) > 1e-9:
        raise ValueError("min_bid must be an integer number of dollars")
    if abs(float(max_per_team) - float(max_int)) > 1e-9:
        raise ValueError("max_per_team must be an integer number of dollars")

    max_k_by_budget = int(budget // min_bid)
    max_k = min(max_teams, max_k_by_budget, int(len(df)))
    if max_k < min_teams:
        raise ValueError("budget too small to satisfy min_teams at min_bid")

    k = max_k
    ranked = df.sort_values(by=score_col, ascending=False)
    chosen = ranked.head(k).copy().reset_index(drop=True)

    if allocation_mode == "equal":
        bids = backtest_scaffold._waterfill_equal(
            k=k,
            budget=budget,
            max_per_team=max_per_team,
        )
        if any(b < min_bid for b in bids):
            raise ValueError("allocation violates min_bid constraint")
        chosen["bid_amount"] = bids
    elif allocation_mode == "expected_points":
        required_cols = ["expected_team_points", "predicted_team_total_bids"]
        missing = [c for c in required_cols if c not in chosen.columns]
        if missing:
            raise ValueError(
                "allocation_mode=expected_points requires columns: "
                + ", ".join(missing)
            )
        bids = allocate_expected_points(
            df=chosen,
            budget=float(budget),
            min_bid=float(min_bid),
            max_per_team=float(max_per_team),
        )
        chosen["bid_amount"] = bids
    else:
        raise ValueError(f"unknown allocation_mode: {allocation_mode}")

    portfolio_rows: List[Dict[str, object]] = []
    for _, r in chosen.iterrows():
        portfolio_rows.append(
            {
                "team_key": str(r["team_key"]),
                "bid_amount": float(r["bid_amount"]),
                "score": float(r[score_col]),
            }
        )

    return chosen, portfolio_rows
