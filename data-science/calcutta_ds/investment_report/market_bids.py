from __future__ import annotations

from typing import Dict, List

import pandas as pd


def filter_market_bids(
    *,
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    exclude_entry_names: List[str],
) -> pd.DataFrame:
    market_bids = tables["entry_bids"].copy()
    market_bids = market_bids[
        market_bids["calcutta_key"] == calcutta_key
    ].copy()
    if not exclude_entry_names:
        return market_bids

    entries = tables.get("entries")
    if entries is None or "entry_key" not in entries.columns:
        return market_bids
    if "entry_name" not in entries.columns:
        return market_bids

    e = entries.copy()
    e = e[e["calcutta_key"] == calcutta_key].copy()
    e["entry_name"] = e["entry_name"].astype(str)
    exclude_norm = [
        str(n).strip().lower()
        for n in exclude_entry_names
        if str(n).strip()
    ]
    if not exclude_norm:
        return market_bids

    def _matches(name: str) -> bool:
        n = str(name).strip().lower()
        return any(x in n for x in exclude_norm)

    excluded_keys = set(
        e[e["entry_name"].apply(_matches)]["entry_key"].astype(str).tolist()
    )
    if not excluded_keys:
        return market_bids

    return market_bids[
        ~market_bids["entry_key"].astype(str).isin(excluded_keys)
    ].copy()


def compute_team_shares_from_bids(
    *,
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    exclude_entry_names: List[str],
) -> Dict[str, float]:
    bids = filter_market_bids(
        tables=tables,
        calcutta_key=calcutta_key,
        exclude_entry_names=exclude_entry_names,
    )
    bids = bids[bids["calcutta_key"] == calcutta_key].copy()
    bids["bid_amount"] = (
        pd.to_numeric(bids["bid_amount"], errors="coerce")
        .fillna(0.0)
    )

    teams = tables.get("teams")
    required = {
        "team_key",
        "wins",
        "byes",
        "calcutta_key",
    }
    if teams is not None and required.issubset(set(teams.columns)):
        t = teams[teams["calcutta_key"] == calcutta_key].copy()
        t["wins"] = (
            pd.to_numeric(t["wins"], errors="coerce")
            .fillna(0)
            .astype(int)
        )
        t["byes"] = (
            pd.to_numeric(t["byes"], errors="coerce")
            .fillna(0)
            .astype(int)
        )
        eligible_team_keys = set(
            t[(t["wins"] != 0) | (t["byes"] != 0)]["team_key"]
            .astype(str)
            .tolist()
        )
        bids = bids[
            bids["team_key"].astype(str).isin(eligible_team_keys)
        ].copy()

    totals = bids.groupby("team_key")["bid_amount"].sum()
    denom = float(totals.sum())
    if denom <= 0:
        return {}
    return {str(k): float(v) / denom for k, v in totals.items()}
