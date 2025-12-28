from __future__ import annotations

from typing import Dict, List, Optional

import pandas as pd

import backtest_scaffold


def points_by_round_from_tables(
    *,
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    points_mode: str,
) -> Optional[Dict[int, float]]:
    if points_mode != "round_scoring":
        return None
    if "round_scoring" not in tables:
        return None

    rs = tables["round_scoring"].copy()
    rs = rs[rs["calcutta_key"] == calcutta_key].copy()
    rs["round"] = pd.to_numeric(rs["round"], errors="coerce")
    rs["points"] = pd.to_numeric(rs["points"], errors="coerce")
    rs = rs[rs["round"].notna() & rs["points"].notna()].copy()
    return {int(r["round"]): float(r["points"]) for _, r in rs.iterrows()}


def compute_team_mean_points(
    *,
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    market_bids: pd.DataFrame,
    points_mode: str,
    points_by_round: Optional[Dict[int, float]],
    expected_sims: int,
    expected_seed: int,
    kenpom_scale: float,
    budget: float,
    expected_use_historical_winners: bool,
    competitor_entry_keys: List[str],
    any_team_key: str,
) -> Dict[str, float]:
    dummy_sim_rows = pd.DataFrame(
        {
            "calcutta_key": [calcutta_key],
            "entry_key": ["simulated:entry"],
            "team_key": [str(any_team_key)],
            "bid_amount": [0.0],
        }
    )

    exp_pre = backtest_scaffold._expected_simulation(
        tables=tables,
        calcutta_key=calcutta_key,
        market_bids=market_bids,
        sim_rows=dummy_sim_rows,
        sim_entry_key="simulated:entry",
        points_mode=points_mode,
        points_by_round=points_by_round,
        n_sims=int(expected_sims),
        seed=int(expected_seed),
        kenpom_scale=float(kenpom_scale),
        budget=float(budget),
        use_historical_winners=bool(expected_use_historical_winners),
        competitor_entry_keys=competitor_entry_keys,
    )

    team_mean_points = exp_pre.get("team_mean_points")
    if not isinstance(team_mean_points, dict) or not team_mean_points:
        raise ValueError("expected simulation did not produce team_mean_points")

    out: Dict[str, float] = {}
    for k, v in team_mean_points.items():
        out[str(k)] = float(v)
    return out


def add_expected_columns(
    *,
    usable: pd.DataFrame,
    predicted_total_pool_bids: float,
    team_mean_points: Dict[str, float],
) -> pd.DataFrame:
    out = usable.copy()
    out["expected_team_points"] = out["team_key"].apply(
        lambda tk: float(team_mean_points.get(str(tk), 0.0))
    )
    out["predicted_team_total_bids"] = (
        pd.to_numeric(out["predicted_team_share_of_pool"], errors="coerce")
        .fillna(0.0)
        .apply(lambda s: float(s) * float(predicted_total_pool_bids))
    )
    return out
