"""
Compute analytical tournament values from KenPom ratings.

This module bridges the all_matchups enumeration with the tournament_value
probability calculator to produce analytically-derived championship probabilities
and expected points for each team.
"""
from __future__ import annotations

import pandas as pd

from moneyball.models.all_matchups import generate_all_theoretical_matchups
from moneyball.models.tournament_value import generate_tournament_value


def compute_analytical_tournament_values(
    teams_df: pd.DataFrame,
    kenpom_scale: float = 10.0,
) -> pd.DataFrame:
    """
    Compute analytical championship probabilities and expected points.

    Uses deterministic KenPom-based calculations (NOT Monte Carlo).

    Args:
        teams_df: DataFrame with columns:
            - id or team_key: Unique team identifier
            - seed: Tournament seed (1-16)
            - region: Region name (East, West, South, Midwest)
            - kenpom_net: KenPom net rating
        kenpom_scale: Scale parameter for win probability sigmoid

    Returns:
        DataFrame with columns:
            - team_key: Unique team identifier
            - analytical_p_championship: Probability of winning championship
            - analytical_expected_points: Expected tournament points
    """
    # Prepare teams DataFrame with required columns
    teams = teams_df.copy()

    # Ensure we have an 'id' column for all_matchups
    if "id" not in teams.columns:
        if "team_key" in teams.columns:
            teams["id"] = teams["team_key"]
        else:
            raise ValueError("teams_df must have 'id' or 'team_key' column")

    required = ["id", "seed", "region", "kenpom_net"]
    missing = [c for c in required if c not in teams.columns]
    if missing:
        raise ValueError(f"teams_df missing columns: {missing}")

    # Generate all theoretical matchups
    matchups = generate_all_theoretical_matchups(teams, kenpom_scale=kenpom_scale)

    # Transform matchups to tournament_value format
    # Rename columns: round_int -> round_order, team1_id/team2_id -> team1_key/team2_key
    matchups = matchups.rename(columns={
        "round_int": "round_order",
        "team1_id": "team1_key",
        "team2_id": "team2_key",
    })

    # Add p_team2_wins_given_matchup (complement of p_team1_wins_given_matchup)
    matchups["p_team2_wins_given_matchup"] = (
        1.0 - matchups["p_team1_wins_given_matchup"]
    )

    # Generate tournament values
    production_df, debug_df = generate_tournament_value(
        predicted_game_outcomes=matchups,
    )

    # Build output with analytical_ prefix
    # p_round_6 = probability of winning the championship game (6th game)
    result = pd.DataFrame({
        "team_key": debug_df["team_key"],
        "analytical_p_championship": debug_df["p_round_6"],
        "analytical_expected_points": debug_df["expected_points_per_entry"],
    })

    return result
