"""
Generate tournament value artifacts (production + debug).

Production artifact: Minimal fields for portfolio construction
Debug artifact: All fields for inspection and debugging
"""
from __future__ import annotations

from typing import Dict, Tuple

import pandas as pd
import numpy as np

from moneyball.utils import points


def generate_tournament_value(
    *,
    predicted_game_outcomes: pd.DataFrame,
) -> Tuple[pd.DataFrame, pd.DataFrame]:
    """
    Generate tournament value artifacts from predicted game outcomes.
    
    Returns:
        (production_df, debug_df)
        
        production_df columns:
            - team_key: Unique team identifier
            - expected_points_per_entry: Expected points for 100% ownership
            
        debug_df columns:
            - team_key: Unique team identifier
            - expected_points_per_entry: Expected points for 100% ownership
            - variance_points: Variance of points
            - std_points: Standard deviation of points
            - p_round_2: Probability of reaching round 2 (R32)
            - p_round_3: Probability of reaching round 3 (S16)
            - p_round_4: Probability of reaching round 4 (E8)
            - p_round_5: Probability of reaching round 5 (F4)
            - p_round_6: Probability of reaching round 6 (Finals)
            - p_round_7: Probability of reaching round 7 (Champion)
    """
    required = [
        "round_order",
        "team1_key",
        "team2_key",
        "p_matchup",
        "p_team1_wins_given_matchup",
        "p_team2_wins_given_matchup",
    ]
    missing = [c for c in required if c not in predicted_game_outcomes.columns]
    if missing:
        raise ValueError(
            "predicted_game_outcomes missing columns: " + ", ".join(missing)
        )

    df = predicted_game_outcomes.copy()
    df["round_order"] = pd.to_numeric(df["round_order"], errors="coerce")
    df = df[df["round_order"].notna()].copy()
    df["round_order"] = df["round_order"].astype(int)

    df["p_matchup"] = pd.to_numeric(df["p_matchup"], errors="coerce").fillna(0)
    df["p_team1_wins_given_matchup"] = pd.to_numeric(
        df["p_team1_wins_given_matchup"],
        errors="coerce",
    ).fillna(0)
    df["p_team2_wins_given_matchup"] = pd.to_numeric(
        df["p_team2_wins_given_matchup"],
        errors="coerce",
    ).fillna(0)

    df = df[(df["round_order"] >= 1) & (df["round_order"] <= 7)].copy()

    # Calculate win probabilities by round for each team
    t1 = pd.DataFrame(
        {
            "team_key": df["team1_key"].astype(str),
            "round_order": df["round_order"].astype(int),
            "p_win": df["p_matchup"] * df["p_team1_wins_given_matchup"],
        }
    )
    t2 = pd.DataFrame(
        {
            "team_key": df["team2_key"].astype(str),
            "round_order": df["round_order"].astype(int),
            "p_win": df["p_matchup"] * df["p_team2_wins_given_matchup"],
        }
    )

    wins = pd.concat([t1, t2], ignore_index=True)
    wins = wins[wins["team_key"].astype(str).str.len() > 0].copy()
    wins["p_win"] = pd.to_numeric(wins["p_win"], errors="coerce").fillna(0)

    p_by_round = (
        wins.groupby(["team_key", "round_order"], as_index=False)["p_win"]
        .sum()
        .copy()
    )

    # Calculate incremental points for each round
    inc_by_round: Dict[int, float] = {}
    for r in range(1, 8):
        inc_by_round[int(r)] = float(points.team_points_fixed(int(r))) - float(
            points.team_points_fixed(int(r - 1))
        )

    p_by_round["points_inc"] = p_by_round["round_order"].apply(
        lambda ro: float(inc_by_round.get(int(ro), 0.0))
    )
    p_by_round["expected_points"] = p_by_round["p_win"] * p_by_round["points_inc"]

    # Calculate expected points per team
    expected = (
        p_by_round.groupby("team_key", as_index=False)["expected_points"]
        .sum()
        .copy()
    )
    expected = expected.rename(columns={"expected_points": "expected_points_per_entry"})
    expected["expected_points_per_entry"] = pd.to_numeric(
        expected["expected_points_per_entry"], errors="coerce"
    ).fillna(0.0)

    # Calculate variance for debug
    p_by_round["variance_contribution"] = (
        p_by_round["p_win"]
        * (1 - p_by_round["p_win"])
        * (p_by_round["points_inc"] ** 2)
    )
    variance = (
        p_by_round.groupby("team_key", as_index=False)["variance_contribution"]
        .sum()
        .copy()
    )
    variance = variance.rename(columns={"variance_contribution": "variance_points"})

    # Pivot to get probabilities by round for debug
    p_wide = p_by_round.pivot(
        index="team_key", columns="round_order", values="p_win"
    ).reset_index()
    p_wide.columns = ["team_key"] + [
        f"p_round_{int(r)}" for r in p_wide.columns[1:]
    ]

    # Build production artifact (minimal)
    production = expected[["team_key", "expected_points_per_entry"]].copy()

    # Build debug artifact (comprehensive)
    debug = expected.merge(variance, on="team_key", how="left")
    debug = debug.merge(p_wide, on="team_key", how="left")
    debug["std_points"] = np.sqrt(debug["variance_points"].fillna(0))

    # Fill missing round probabilities with 0
    for r in range(1, 8):
        col = f"p_round_{r}"
        if col not in debug.columns:
            debug[col] = 0.0
        else:
            debug[col] = debug[col].fillna(0.0)

    # Order debug columns
    debug_cols = ["team_key", "expected_points_per_entry", "variance_points", "std_points"]
    debug_cols += [f"p_round_{r}" for r in range(1, 8)]
    debug = debug[debug_cols].copy()

    return production, debug
