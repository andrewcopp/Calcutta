import math
from typing import Dict, Tuple

import pandas as pd


def round_order(round_name: str) -> int:
    order = {
        "first_four": 1,
        "round_of_64": 2,
        "round_of_32": 3,
        "sweet_16": 4,
        "elite_8": 5,
        "final_four": 6,
        "championship": 7,
    }
    return int(order.get(str(round_name), 999))


def sigmoid(x: float) -> float:
    if x >= 0:
        z = math.exp(-x)
        return 1.0 / (1.0 + z)
    z = math.exp(x)
    return z / (1.0 + z)


def win_prob(
    net1: float,
    net2: float,
    scale: float,
) -> float:
    if scale <= 0:
        raise ValueError("kenpom_scale must be positive")
    return sigmoid((net1 - net2) / scale)


def prepare_bracket_graph(
    games: pd.DataFrame,
) -> Tuple[pd.DataFrame, Dict[str, Dict[int, str]]]:
    required = [
        "game_id",
        "round",
        "sort_order",
        "team1_key",
        "team2_key",
        "next_game_id",
        "next_game_slot",
    ]
    missing = [c for c in required if c not in games.columns]
    if missing:
        raise ValueError(f"games DataFrame missing columns: {missing}")

    g = games.copy()
    g["sort_order"] = (
        pd.to_numeric(g["sort_order"], errors="coerce")
        .fillna(0)
        .astype(int)
    )
    g["round_order"] = g["round"].apply(round_order)

    prev_by_next: Dict[str, Dict[int, str]] = {}
    for _, r in g.iterrows():
        nxt = str(r.get("next_game_id") or "")
        if not nxt:
            continue
        slot_raw = r.get("next_game_slot")
        try:
            slot = int(slot_raw)
        except Exception:
            continue
        if slot not in (1, 2):
            continue
        prev_by_next.setdefault(nxt, {})[slot] = str(r.get("game_id"))

    g = g.sort_values(
        by=["round_order", "sort_order", "game_id"],
    ).reset_index(drop=True)
    return g, prev_by_next
