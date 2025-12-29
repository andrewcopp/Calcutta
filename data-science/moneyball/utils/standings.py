from typing import Dict, List

import pandas as pd


def ensure_entry_points_include_competitors(
    entry_points: pd.DataFrame,
    competitor_entry_keys: List[str],
) -> pd.DataFrame:
    if entry_points is None or entry_points.empty:
        base = pd.DataFrame(columns=["entry_key", "total_points"])
    else:
        base = entry_points.copy()

    have = set(base["entry_key"].astype(str).tolist())
    missing = [str(k) for k in competitor_entry_keys if str(k) not in have]
    if not missing:
        return base

    add = pd.DataFrame(
        {
            "entry_key": missing,
            "total_points": [0.0 for _ in missing],
        }
    )
    return pd.concat([base, add], ignore_index=True)


def compute_entry_points(
    entry_bids: pd.DataFrame,
    points_by_team: pd.DataFrame,
    calcutta_key: str,
) -> pd.DataFrame:
    required = ["calcutta_key", "entry_key", "team_key", "bid_amount"]
    missing = [c for c in required if c not in entry_bids.columns]
    if missing:
        raise ValueError(f"entry_bids.parquet missing columns: {missing}")

    bids = entry_bids.copy()
    bids = bids[bids["calcutta_key"] == calcutta_key].copy()
    bids["bid_amount"] = (
        pd.to_numeric(bids["bid_amount"], errors="coerce")
        .fillna(0.0)
    )

    totals = (
        bids.groupby(["team_key"], dropna=False)
        .agg(total_team_bids=("bid_amount", "sum"))
        .reset_index()
    )

    b = bids.merge(totals, on="team_key", how="left")
    b = b.merge(points_by_team, on="team_key", how="left")
    b["team_points"] = (
        pd.to_numeric(b["team_points"], errors="coerce")
        .fillna(0.0)
    )
    b["total_team_bids"] = (
        pd.to_numeric(b["total_team_bids"], errors="coerce")
        .fillna(0.0)
    )
    b["ownership"] = b.apply(
        lambda r: (
            (r["bid_amount"] / r["total_team_bids"])
            if r["total_team_bids"]
            else 0.0
        ),
        axis=1,
    )
    b["points"] = b["team_points"] * b["ownership"]

    entry_points = (
        b.groupby(["entry_key"], dropna=False)
        .agg(total_points=("points", "sum"))
        .reset_index()
    )
    entry_points["total_points"] = (
        pd.to_numeric(entry_points["total_points"], errors="coerce")
        .fillna(0.0)
    )
    return entry_points


def compute_finish_positions_and_payouts(
    entry_points: pd.DataFrame,
    payouts: pd.DataFrame,
    calcutta_key: str,
    epsilon: float = 0.0001,
) -> pd.DataFrame:
    req = ["calcutta_key", "position", "amount_cents"]
    missing = [c for c in req if c not in payouts.columns]
    if missing:
        raise ValueError(f"payouts.parquet missing columns: {missing}")

    payout_map: Dict[int, int] = {}
    p = payouts[payouts["calcutta_key"] == calcutta_key].copy()
    p["position"] = pd.to_numeric(p["position"], errors="coerce")
    p["amount_cents"] = pd.to_numeric(p["amount_cents"], errors="coerce")
    p = p[p["position"].notna() & p["amount_cents"].notna()].copy()
    for _, r in p.iterrows():
        payout_map[int(r["position"])] = int(r["amount_cents"])

    ep = entry_points.copy()
    ep["total_points"] = (
        pd.to_numeric(ep["total_points"], errors="coerce")
        .fillna(0.0)
    )

    ep = ep.sort_values(
        by=["total_points", "entry_key"],
        ascending=[False, True],
    ).reset_index(drop=True)

    finish_pos: List[int] = []
    is_tied: List[bool] = []
    payout_cents: List[int] = []

    position = 1
    i = 0
    while i < len(ep):
        j = i + 1
        cur = float(ep.loc[i, "total_points"])
        while j < len(ep):
            nxt = float(ep.loc[j, "total_points"])
            if abs(nxt - cur) >= epsilon:
                break
            j += 1

        group_size = j - i
        tied = group_size > 1

        total_group_payout = 0
        for pos in range(position, position + group_size):
            total_group_payout += payout_map.get(pos, 0)

        base = 0
        remainder = 0
        if group_size > 0:
            base = total_group_payout // group_size
            remainder = total_group_payout % group_size

        for _ in range(group_size):
            finish_pos.append(position)
            is_tied.append(tied)
            amt = base
            if remainder > 0:
                amt += 1
                remainder -= 1
            payout_cents.append(int(amt))

        position += group_size
        i = j

    ep["finish_position"] = finish_pos
    ep["is_tied"] = is_tied
    ep["payout_cents"] = payout_cents
    ep["in_the_money"] = ep["payout_cents"].apply(lambda v: bool(int(v) > 0))
    return ep
