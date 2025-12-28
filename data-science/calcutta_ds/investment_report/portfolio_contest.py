from __future__ import annotations

from typing import Dict, List, Optional, Tuple

import numpy as np
import pandas as pd


def _utility(
    *,
    payout_cents: np.ndarray,
    utility: str,
    utility_gamma: float,
    utility_alpha: float,
    epsilon: float,
) -> np.ndarray:
    u = str(utility)
    x = np.asarray(payout_cents, dtype=float)
    if u == "linear":
        return x
    if u == "log":
        return np.log(np.maximum(x, 0.0) + float(epsilon))
    if u == "power":
        g = float(utility_gamma)
        if g <= 0:
            return np.zeros_like(x)
        return np.power(np.maximum(x, 0.0) + float(epsilon), g)
    if u == "exp":
        a = float(utility_alpha)
        return 1.0 - np.exp(-a * np.maximum(x, 0.0))
    raise ValueError(f"unknown utility: {utility}")


def contest_objective_from_sim_bids(
    *,
    team_points_scenarios: np.ndarray,
    market_entry_bids: np.ndarray,
    market_team_totals: np.ndarray,
    sim_team_bids: np.ndarray,
    objective: str,
    top_k: int,
    payout_map: Optional[Dict[int, int]] = None,
    utility: str = "linear",
    utility_gamma: float = 1.0,
    utility_alpha: float = 1.0,
    epsilon: float = 1e-6,
) -> float:
    if team_points_scenarios.size == 0:
        return 0.0

    denom = market_team_totals + sim_team_bids
    denom = np.where(denom > 0.0, denom, 1.0)

    comp_shares = market_entry_bids / denom[None, :]
    competitor_points = team_points_scenarios @ comp_shares.T

    sim_share = sim_team_bids / denom
    sim_points = team_points_scenarios @ sim_share

    diff = competitor_points - sim_points[:, None]
    gt = (diff > float(epsilon)).sum(axis=1).astype(float)
    eq = (np.abs(diff) <= float(epsilon)).sum(axis=1).astype(float)
    finish_pos = 1.0 + gt + 0.5 * eq

    if objective == "mean_finish_position":
        return -float(finish_pos.mean())
    if objective == "p_top1":
        return float((finish_pos <= 1.0 + 1e-9).mean())
    if objective == "p_topk":
        k = int(top_k)
        if k <= 0:
            return 0.0
        return float((finish_pos <= float(k) + 1e-9).mean())

    if objective in ("expected_payout", "expected_utility_payout"):
        if not payout_map:
            raise ValueError("payout objective requires payout_map")

        start_pos = (1.0 + gt).astype(int)
        group_size = (1.0 + eq).astype(int)

        max_pos = int(start_pos.max() + group_size.max()) if start_pos.size else 0
        payout_by_pos = np.zeros((max_pos + 2,), dtype=float)
        for pos, amt in payout_map.items():
            p = int(pos)
            if p < 0 or p >= len(payout_by_pos):
                continue
            payout_by_pos[p] = float(amt)

        payout_cents = np.zeros((len(start_pos),), dtype=float)
        for i in range(len(start_pos)):
            s = int(start_pos[i])
            g = int(group_size[i])
            tot = float(payout_by_pos[s : s + g].sum()) if g > 0 else 0.0
            payout_cents[i] = (tot / float(g)) if g > 0 else 0.0

        if objective == "expected_payout":
            return float(payout_cents.mean())

        util = _utility(
            payout_cents=payout_cents,
            utility=str(utility),
            utility_gamma=float(utility_gamma),
            utility_alpha=float(utility_alpha),
            epsilon=float(epsilon),
        )
        return float(np.asarray(util, dtype=float).mean())

    raise ValueError(f"unknown greedy_objective: {objective}")


def optimize_portfolio_greedy_contest(
    *,
    df: pd.DataFrame,
    budget: float,
    min_teams: int,
    max_teams: int,
    max_per_team: float,
    min_bid: float,
    objective: str,
    top_k: int,
    payout_map: Optional[Dict[int, int]] = None,
    utility: str = "linear",
    utility_gamma: float = 1.0,
    utility_alpha: float = 1.0,
    team_keys: List[str],
    team_points_scenarios: np.ndarray,
    market_entry_bids: np.ndarray,
    market_team_totals: np.ndarray,
    trace_out: Optional[List[Dict[str, object]]] = None,
) -> Tuple[pd.DataFrame, List[Dict[str, object]]]:
    if df.empty:
        return df.copy(), []

    b_int = int(round(float(budget)))
    min_int = int(round(float(min_bid)))
    max_int = int(round(float(max_per_team)))
    if abs(float(budget) - float(b_int)) > 1e-9:
        raise ValueError("budget must be an integer number of dollars")
    if abs(float(min_bid) - float(min_int)) > 1e-9:
        raise ValueError("min_bid must be an integer number of dollars")
    if abs(float(max_per_team) - float(max_int)) > 1e-9:
        raise ValueError("max_per_team must be an integer number of dollars")

    pool = df.copy().reset_index(drop=True)
    pool["team_key"] = pool["team_key"].astype(str)

    t_index = {str(tk): i for i, tk in enumerate(team_keys)}
    pool_idx: List[int] = []
    for _, r in pool.iterrows():
        tk = str(r.get("team_key"))
        pool_idx.append(int(t_index.get(tk, -1)))
    pool["_t_idx"] = pool_idx
    pool = pool[pool["_t_idx"] >= 0].copy().reset_index(drop=True)
    if pool.empty:
        return pool, []

    n = int(len(pool))
    if float(min_teams) * float(min_int) - float(b_int) > 1e-9:
        raise ValueError("budget too small to satisfy min_teams at min_bid")

    bids: List[int] = [0 for _ in range(n)]
    selected: set[int] = set()
    remaining = int(b_int)

    sim_team_bids = np.zeros((int(len(team_keys)),), dtype=float)

    def _objective_for_current() -> float:
        return contest_objective_from_sim_bids(
            team_points_scenarios=team_points_scenarios,
            market_entry_bids=market_entry_bids,
            market_team_totals=market_team_totals,
            sim_team_bids=sim_team_bids,
            objective=str(objective),
            top_k=int(top_k),
            payout_map=payout_map,
            utility=str(utility),
            utility_gamma=float(utility_gamma),
            utility_alpha=float(utility_alpha),
        )

    cur_obj = _objective_for_current()

    def _try_apply(i: int, inc: int) -> float:
        if inc <= 0:
            return -1e99
        if bids[i] + inc > int(max_int):
            return -1e99
        tk_i = int(pool.loc[i, "_t_idx"])
        sim_team_bids[tk_i] += float(inc)
        new_obj = _objective_for_current()
        sim_team_bids[tk_i] -= float(inc)
        if np.isnan(new_obj):
            return -1e99
        return float(new_obj)

    step = 0

    while len(selected) < int(min_teams):
        if remaining < int(min_int):
            break
        best_i: Optional[int] = None
        best_new = -1e99
        tied: List[int] = []
        for i in range(n):
            if i in selected:
                continue
            new_obj = _try_apply(i, int(min_int))
            if new_obj > best_new:
                best_new = new_obj
                best_i = i
                tied = [i]
            elif abs(float(new_obj) - float(best_new)) <= 1e-12:
                tied.append(i)
        if best_i is None:
            break

        if trace_out is not None:
            tk = str(pool.loc[best_i, "team_key"])
            trace_out.append(
                {
                    "step": int(step),
                    "phase": "seed_min_teams",
                    "remaining_before": int(remaining),
                    "selected_before": int(len(selected)),
                    "action": {
                        "pool_i": int(best_i),
                        "team_key": tk,
                        "school_name": str(pool.loc[best_i].get("school_name") or ""),
                        "inc": int(min_int),
                    },
                    "objective_before": float(cur_obj),
                    "objective_after": float(best_new),
                    "delta": float(best_new - float(cur_obj)),
                    "n_tied_best": int(len(tied)),
                    "tied_best_team_keys": [
                        str(pool.loc[i, "team_key"]) for i in tied[:10]
                    ],
                }
            )

        bids[best_i] = int(min_int)
        selected.add(best_i)
        sim_team_bids[int(pool.loc[best_i, "_t_idx"])] += float(min_int)
        remaining -= int(min_int)
        cur_obj = float(best_new)
        step += 1

    while remaining > 0:
        best_i: Optional[int] = None
        best_inc: int = 0
        # For contest objectives (especially p_topk), the metric can be flat
        # across many $1 increments. We still want to spend the full budget
        # (if feasible), so pick the best available action even if it does not
        # strictly improve the objective.
        best_new = -1e99
        tied_keys: List[str] = []

        for i in selected:
            if bids[i] >= int(max_int):
                continue
            new_obj = _try_apply(i, 1)
            if new_obj > best_new:
                best_new = new_obj
                best_i = i
                best_inc = 1
                tied_keys = [str(pool.loc[i, "team_key"])]
            elif abs(float(new_obj) - float(best_new)) <= 1e-12:
                tied_keys.append(str(pool.loc[i, "team_key"]))

        if len(selected) < int(max_teams) and remaining >= int(min_int):
            for i in range(n):
                if i in selected:
                    continue
                if int(min_int) > int(max_int):
                    continue
                new_obj = _try_apply(i, int(min_int))
                if new_obj > best_new:
                    best_new = new_obj
                    best_i = i
                    best_inc = int(min_int)
                    tied_keys = [str(pool.loc[i, "team_key"])]
                elif abs(float(new_obj) - float(best_new)) <= 1e-12:
                    tied_keys.append(str(pool.loc[i, "team_key"]))

        if best_i is None or best_inc <= 0:
            raise ValueError(
                "unable to allocate remaining budget under constraints"
            )

        if trace_out is not None:
            tk = str(pool.loc[best_i, "team_key"])
            trace_out.append(
                {
                    "step": int(step),
                    "phase": "allocate",
                    "remaining_before": int(remaining),
                    "selected_before": int(len(selected)),
                    "action": {
                        "pool_i": int(best_i),
                        "team_key": tk,
                        "school_name": str(pool.loc[best_i].get("school_name") or ""),
                        "inc": int(best_inc),
                        "new_bid": int(bids[best_i] + int(best_inc)),
                        "is_new_team": bool(best_i not in selected),
                    },
                    "objective_before": float(cur_obj),
                    "objective_after": float(best_new),
                    "delta": float(best_new - float(cur_obj)),
                    "n_tied_best": int(len(tied_keys)),
                    "tied_best_team_keys": tied_keys[:10],
                }
            )

        if best_i not in selected and best_inc == int(min_int):
            selected.add(best_i)
        bids[best_i] += int(best_inc)
        sim_team_bids[int(pool.loc[best_i, "_t_idx"])] += float(best_inc)
        remaining -= int(best_inc)
        cur_obj = float(best_new)
        step += 1

    chosen = pool.loc[sorted(selected)].copy().reset_index(drop=True)
    chosen_bids = [int(bids[i]) for i in sorted(selected)]

    spent = int(sum(chosen_bids))
    if spent < int(b_int) and chosen_bids:
        extra = int(b_int) - spent
        chosen_bids[0] = min(int(max_int), int(chosen_bids[0]) + extra)

    chosen["bid_amount"] = [float(x) for x in chosen_bids]

    portfolio_rows: List[Dict[str, object]] = []
    for _, r in chosen.iterrows():
        portfolio_rows.append(
            {
                "team_key": str(r["team_key"]),
                "bid_amount": float(r["bid_amount"]),
                "score": 0.0,
            }
        )
    return chosen, portfolio_rows
