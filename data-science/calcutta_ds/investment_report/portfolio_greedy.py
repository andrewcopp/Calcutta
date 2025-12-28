from __future__ import annotations

from typing import Dict, List, Optional, Tuple

import pandas as pd


def optimize_portfolio_greedy(
    *,
    df: pd.DataFrame,
    score_col: str,
    budget: float,
    min_teams: int,
    max_teams: int,
    max_per_team: float,
    min_bid: float,
    step: float = 1.0,
) -> Tuple[pd.DataFrame, List[Dict[str, object]]]:
    required_cols = [
        "team_key",
        "expected_team_points",
        "predicted_team_total_bids",
    ]
    missing = [c for c in required_cols if c not in df.columns]
    if missing:
        raise ValueError(
            "greedy optimizer requires columns: " + ", ".join(missing)
        )
    if budget <= 0:
        raise ValueError("budget must be positive")
    if min_teams <= 0 or max_teams <= 0:
        raise ValueError("min_teams and max_teams must be positive")
    if min_teams > max_teams:
        raise ValueError("min_teams cannot exceed max_teams")
    if min_bid <= 0:
        raise ValueError("min_bid must be positive")
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

    step = 1.0

    pool = df.copy().reset_index(drop=True)
    pool["expected_team_points"] = pd.to_numeric(
        pool["expected_team_points"],
        errors="coerce",
    ).fillna(0.0)
    pool["predicted_team_total_bids"] = pd.to_numeric(
        pool["predicted_team_total_bids"],
        errors="coerce",
    ).fillna(0.0)

    n = int(len(pool))
    if n == 0:
        return pool, []

    if float(min_teams) * float(min_int) - float(b_int) > 1e-9:
        raise ValueError("budget too small to satisfy min_teams at min_bid")

    bids: List[float] = [0.0 for _ in range(n)]
    selected: set[int] = set()

    def _delta_for(i: int, b0: float, inc: float) -> float:
        if inc <= 0:
            return 0.0
        if b0 + inc - float(max_int) > 1e-9:
            return -1e99
        m = float(pool.loc[i, "predicted_team_total_bids"])
        if m < 0:
            m = 0.0
        exp_pts = float(pool.loc[i, "expected_team_points"])
        denom0 = m + b0
        denom1 = m + b0 + inc
        s0 = (b0 / denom0) if denom0 > 0 else 0.0
        s1 = ((b0 + inc) / denom1) if denom1 > 0 else 0.0
        return exp_pts * (s1 - s0)

    remaining = float(b_int)

    while len(selected) < int(min_teams):
        best_i: Optional[int] = None
        best_v = -1e99
        for i in range(n):
            if i in selected:
                continue
            v = _delta_for(i, 0.0, float(min_int)) / float(min_int)
            if v > best_v:
                best_v = v
                best_i = i
        if best_i is None:
            break
        bids[best_i] = float(min_int)
        selected.add(best_i)
        remaining -= float(min_int)

    while remaining > 1e-9:
        best_i: Optional[int] = None
        best_inc: float = 0.0
        best_val = -1e99

        for i in selected:
            inc = float(step) if remaining >= step else float(remaining)
            if bids[i] + inc - float(max_int) > 1e-9:
                inc = max(0.0, float(max_int) - float(bids[i]))
            if inc <= 1e-12:
                continue
            v = _delta_for(i, float(bids[i]), float(inc)) / float(inc)
            if v > best_val:
                best_val = v
                best_i = i
                best_inc = float(inc)

        if (
            len(selected) < int(max_teams)
            and remaining + 1e-9 >= float(min_int)
        ):
            for i in range(n):
                if i in selected:
                    continue
                if float(min_int) - float(max_int) > 1e-9:
                    continue
                v = _delta_for(i, 0.0, float(min_int)) / float(min_int)
                if v > best_val:
                    best_val = v
                    best_i = i
                    best_inc = float(min_int)

        if best_i is None or best_inc <= 1e-12:
            break

        if (
            best_i not in selected
            and abs(best_inc - float(min_int)) < 1e-9
        ):
            selected.add(best_i)
        bids[best_i] += float(best_inc)
        remaining -= float(best_inc)

    chosen = pool.loc[sorted(selected)].copy().reset_index(drop=True)
    chosen_bids = [int(round(float(bids[i]))) for i in sorted(selected)]
    if any(x < int(min_int) for x in chosen_bids):
        raise ValueError("allocation violates min_bid constraint")
    if any(x > int(max_int) for x in chosen_bids):
        raise ValueError("allocation violates max_per_team constraint")
    if int(sum(chosen_bids)) != int(b_int):
        delta = int(b_int) - int(sum(chosen_bids))
        chosen_bids[0] += delta
    chosen["bid_amount"] = [float(x) for x in chosen_bids]

    portfolio_rows: List[Dict[str, object]] = []
    for _, r in chosen.iterrows():
        portfolio_rows.append(
            {
                "team_key": str(r["team_key"]),
                "bid_amount": float(r["bid_amount"]),
                "score": float(r.get(score_col, 0.0) or 0.0),
            }
        )
    return chosen, portfolio_rows
