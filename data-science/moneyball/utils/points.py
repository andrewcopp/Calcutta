from typing import Dict, Optional

import pandas as pd


_DEFAULT_POINTS_BY_WIN_INDEX: Optional[Dict[int, float]] = None


def set_default_points_by_win_index(
    points_by_win_index: Dict[int, float],
) -> None:
    global _DEFAULT_POINTS_BY_WIN_INDEX
    _DEFAULT_POINTS_BY_WIN_INDEX = {
        int(k): float(v) for k, v in points_by_win_index.items()
    }


def points_by_win_index_from_scoring_rules(
    scoring_rules: pd.DataFrame,
) -> Dict[int, float]:
    if scoring_rules is None or scoring_rules.empty:
        return {}

    sr = scoring_rules.copy()
    if "win_index" in sr.columns and "points_awarded" in sr.columns:
        sr["win_index"] = pd.to_numeric(sr["win_index"], errors="coerce")
        sr["points_awarded"] = pd.to_numeric(
            sr["points_awarded"],
            errors="coerce",
        )
        sr = sr[
            sr["win_index"].notna() & sr["points_awarded"].notna()
        ].copy()
        return {
            int(r["win_index"]): float(r["points_awarded"])
            for _, r in sr.iterrows()
        }

    if "round" in sr.columns and "points" in sr.columns:
        sr["round"] = pd.to_numeric(sr["round"], errors="coerce")
        sr["points"] = pd.to_numeric(sr["points"], errors="coerce")
        sr = sr[sr["round"].notna() & sr["points"].notna()].copy()
        return {
            int(r["round"]): float(r["points"]) for _, r in sr.iterrows()
        }

    raise ValueError(
        "scoring_rules must contain (win_index, points_awarded) "
        "or (round, points)"
    )


def team_points_from_scoring_rules(
    progress: int,
    points_by_win_index: Dict[int, float],
) -> float:
    total = 0.0
    p = int(progress)
    if p <= 0:
        return 0.0
    for i in range(1, p + 1):
        total += float(
            points_by_win_index.get(int(i), 0.0)
        )
    return total


def team_points_fixed(progress: int) -> float:
    if _DEFAULT_POINTS_BY_WIN_INDEX is None:
        raise ValueError(
            "Default scoring rules not set; "
            "call set_default_points_by_win_index(...)"
        )
    return team_points_from_scoring_rules(
        int(progress),
        _DEFAULT_POINTS_BY_WIN_INDEX,
    )
