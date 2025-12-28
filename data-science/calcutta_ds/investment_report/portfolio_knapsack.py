from __future__ import annotations

from typing import Dict, List, Optional, Tuple

import pandas as pd


def optimize_portfolio_knapsack(
    *,
    df: pd.DataFrame,
    score_col: str,
    budget: float,
    min_teams: int,
    max_teams: int,
    max_per_team: float,
    min_bid: float,
) -> Tuple[pd.DataFrame, List[Dict[str, object]]]:
    required_cols = [
        "team_key",
        "expected_team_points",
        "predicted_team_total_bids",
    ]
    missing = [c for c in required_cols if c not in df.columns]
    if missing:
        raise ValueError(
            "knapsack optimizer requires columns: " + ", ".join(missing)
        )
    if budget <= 0:
        raise ValueError("budget must be positive")
    if min_teams <= 0 or max_teams <= 0:
        raise ValueError("min_teams and max_teams must be positive")
    if min_teams > max_teams:
        raise ValueError("min_teams cannot exceed max_teams")
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

    if float(min_teams) * float(min_int) - float(b_int) > 1e-9:
        raise ValueError("budget too small to satisfy min_teams at min_bid")

    pool = df.copy().reset_index(drop=True)
    pool["expected_team_points"] = pd.to_numeric(
        pool["expected_team_points"],
        errors="coerce",
    ).fillna(0.0)
    pool["predicted_team_total_bids"] = pd.to_numeric(
        pool["predicted_team_total_bids"],
        errors="coerce",
    ).fillna(0.0)

    if score_col not in pool.columns:
        pool[score_col] = 0.0
    pool[score_col] = pd.to_numeric(
        pool[score_col],
        errors="coerce",
    ).fillna(0.0)

    n = int(len(pool))
    if n == 0:
        return pool, []

    K = int(max_teams)
    neg = -1e99
    dp = [[neg for _ in range(K + 1)] for _ in range(b_int + 1)]
    dp[0][0] = 0.0

    keep_bid: List[List[List[int]]] = [
        [[0 for _ in range(K + 1)] for _ in range(b_int + 1)]
        for _ in range(n)
    ]

    def _team_value(i: int, bid: int) -> float:
        if bid <= 0:
            return 0.0
        m = float(pool.loc[i, "predicted_team_total_bids"])
        if m < 0:
            m = 0.0
        exp_pts = float(pool.loc[i, "expected_team_points"])
        denom = m + float(bid)
        own = (float(bid) / denom) if denom > 0 else 0.0
        return exp_pts * own

    for i in range(n):
        newdp = [[neg for _ in range(K + 1)] for _ in range(b_int + 1)]

        for spent_prev in range(b_int + 1):
            for k_prev in range(K + 1):
                prev = dp[spent_prev][k_prev]
                if prev <= neg / 2:
                    continue

                if prev > newdp[spent_prev][k_prev] + 1e-12:
                    newdp[spent_prev][k_prev] = prev
                    keep_bid[i][spent_prev][k_prev] = 0

                if k_prev + 1 > K:
                    continue

                for bid in range(min_int, max_int + 1):
                    spent = int(spent_prev) + int(bid)
                    if spent > b_int:
                        break
                    k = int(k_prev) + 1
                    v = float(prev) + _team_value(i, int(bid))
                    cur = newdp[spent][k]
                    if v > cur + 1e-12:
                        newdp[spent][k] = v
                        keep_bid[i][spent][k] = int(bid)
                    elif abs(v - cur) <= 1e-12:
                        if int(bid) > int(keep_bid[i][spent][k]):
                            keep_bid[i][spent][k] = int(bid)

        dp = newdp

    best_k: Optional[int] = None
    best_v = neg
    for k in range(int(min_teams), int(max_teams) + 1):
        v = dp[b_int][k]
        if v > best_v + 1e-12:
            best_v = v
            best_k = int(k)
        elif abs(v - best_v) <= 1e-12 and best_k is not None:
            if int(k) < int(best_k):
                best_k = int(k)

    if best_k is None or best_v <= neg / 2:
        raise ValueError("knapsack failed to find feasible portfolio")

    bids: List[int] = [0 for _ in range(n)]
    spent = int(b_int)
    k = int(best_k)
    for i in range(n - 1, -1, -1):
        bid = int(keep_bid[i][spent][k])
        bids[i] = int(bid)
        spent -= int(bid)
        if bid > 0:
            k -= 1
        if spent < 0 or k < 0:
            raise ValueError("knapsack backtrack failed")

    if spent != 0 or k != 0:
        raise ValueError("knapsack backtrack did not consume full state")

    selected_idx = [i for i, b in enumerate(bids) if int(b) > 0]
    chosen = pool.loc[selected_idx].copy().reset_index(drop=True)
    chosen["bid_amount"] = [float(bids[i]) for i in selected_idx]

    if chosen.empty:
        raise ValueError("knapsack produced empty portfolio")

    if any(float(x) < float(min_int) - 1e-9 for x in chosen["bid_amount"]):
        raise ValueError("allocation violates min_bid constraint")
    if any(float(x) > float(max_int) + 1e-9 for x in chosen["bid_amount"]):
        raise ValueError("allocation violates max_per_team constraint")
    if int(round(float(chosen["bid_amount"].sum()))) != int(b_int):
        raise ValueError("allocation violates budget constraint")
    if int(len(chosen)) < int(min_teams) or int(len(chosen)) > int(max_teams):
        raise ValueError("allocation violates team-count constraint")

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
