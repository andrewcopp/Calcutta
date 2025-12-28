from __future__ import annotations

from typing import Dict, List, Tuple

import pandas as pd

import backtest_scaffold


def compute_realized(
    *,
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    points_mode: str,
    market_bids: pd.DataFrame,
    portfolio_rows: List[Dict[str, object]],
    competitor_entry_keys: List[str],
    budget: float,
    real_buyin_dollars: float,
) -> Tuple[Dict[str, object], pd.DataFrame, pd.DataFrame]:
    points_by_team = backtest_scaffold._build_points_by_team(
        teams=tables["teams"],
        calcutta_key=calcutta_key,
        points_mode=points_mode,
        round_scoring=tables.get("round_scoring"),
    )

    sim_rows = pd.DataFrame(
        {
            "calcutta_key": [calcutta_key for _ in portfolio_rows],
            "entry_key": ["simulated:entry" for _ in portfolio_rows],
            "team_key": [r["team_key"] for r in portfolio_rows],
            "bid_amount": [r["bid_amount"] for r in portfolio_rows],
        }
    )

    bids_all = pd.concat([market_bids, sim_rows], ignore_index=True)
    entry_points = backtest_scaffold._compute_entry_points(
        entry_bids=bids_all,
        points_by_team=points_by_team,
        calcutta_key=calcutta_key,
    )

    entry_points = backtest_scaffold._ensure_entry_points_include_competitors(
        entry_points,
        competitor_entry_keys=(
            list(competitor_entry_keys) + ["simulated:entry"]
        ),
    )

    standings = backtest_scaffold._compute_finish_positions_and_payouts(
        entry_points=entry_points,
        payouts=tables["payouts"],
        calcutta_key=calcutta_key,
    )

    sim_row = standings[standings["entry_key"] == "simulated:entry"]
    if len(sim_row) != 1:
        raise ValueError("failed to compute simulated entry standing")

    sim = sim_row.iloc[0]
    payout_cents = int(sim["payout_cents"])
    payout_per_fake_dollar = (
        payout_cents / (float(budget) * 100.0) if budget else 0.0
    )

    buyin_cents = int(round(float(real_buyin_dollars) * 100.0))
    roi_real_buyin = payout_cents / float(buyin_cents) if buyin_cents else 0.0
    points_per_fake_dollar = (
        float(sim["total_points"]) / float(budget) if budget else 0.0
    )
    roi = float(points_per_fake_dollar)

    realized = {
        "points_mode": str(points_mode),
        "budget": float(budget),
        "real_buyin_dollars": float(real_buyin_dollars),
        "total_points": float(sim["total_points"]),
        "points_per_fake_dollar": float(points_per_fake_dollar),
        "finish_position": int(sim["finish_position"]),
        "is_tied": bool(sim["is_tied"]),
        "payout_cents": payout_cents,
        "payout_per_fake_dollar": float(payout_per_fake_dollar),
        "roi": float(roi),
        "roi_real_buyin": float(roi_real_buyin),
    }

    ctx = {
        "payout_cents": payout_cents,
        "buyin_cents": buyin_cents,
    }

    return realized, standings, pd.DataFrame([ctx])
