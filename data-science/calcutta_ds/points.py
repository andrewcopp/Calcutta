from typing import Dict, Optional

import pandas as pd


def team_points_fixed(progress: int) -> float:
    if progress in (0, 1):
        return 0.0
    if progress == 2:
        return 50.0
    if progress == 3:
        return 150.0
    if progress == 4:
        return 300.0
    if progress == 5:
        return 500.0
    if progress == 6:
        return 750.0
    if progress == 7:
        return 1050.0
    return 0.0


def team_points_from_round_scoring(
    progress: int,
    points_by_round: Dict[int, float],
) -> float:
    scoring_rounds = max(progress - 1, 0)
    total = 0.0
    for r in range(1, scoring_rounds + 1):
        total += float(points_by_round.get(r, 0.0))
    return total


def build_points_by_team(
    teams: pd.DataFrame,
    calcutta_key: str,
    points_mode: str,
    round_scoring: Optional[pd.DataFrame],
) -> pd.DataFrame:
    required = ["team_key", "wins", "byes"]
    missing = [c for c in required if c not in teams.columns]
    if missing:
        raise ValueError(f"teams.parquet missing columns: {missing}")

    points_by_round: Dict[int, float] = {}
    if points_mode == "round_scoring":
        if round_scoring is None:
            raise ValueError(
                "points_mode=round_scoring but round_scoring not found"
            )
        rs = round_scoring.copy()
        rs = rs[rs["calcutta_key"] == calcutta_key].copy()
        rs["round"] = pd.to_numeric(rs["round"], errors="coerce")
        rs["points"] = pd.to_numeric(rs["points"], errors="coerce")
        rs = rs[rs["round"].notna() & rs["points"].notna()].copy()
        for _, r in rs.iterrows():
            points_by_round[int(r["round"])] = float(r["points"])

    t = teams.copy()
    t["wins"] = pd.to_numeric(t["wins"], errors="coerce").fillna(0).astype(int)
    t["byes"] = pd.to_numeric(t["byes"], errors="coerce").fillna(0).astype(int)
    t["progress"] = t["wins"] + t["byes"]

    if points_mode == "round_scoring":
        t["team_points"] = t["progress"].apply(
            lambda p: team_points_from_round_scoring(int(p), points_by_round)
        )
    else:
        t["team_points"] = t["progress"].apply(
            lambda p: team_points_fixed(int(p))
        )

    return t[["team_key", "team_points"]].copy()
