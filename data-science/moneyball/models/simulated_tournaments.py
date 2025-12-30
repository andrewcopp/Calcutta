"""
Simulated tournaments model for Moneyball pipeline.

This pure function simulates tournament outcomes by running Monte Carlo
simulations of game results. The output is tournament-agnostic and can be
reused across different Calcutta configurations.

Output: DataFrame with columns [sim_id, team_key, wins]
"""
from __future__ import annotations

import random
from typing import Dict, List

import pandas as pd


def simulate_tournaments(
    *,
    games: pd.DataFrame,
    predicted_game_outcomes: pd.DataFrame,
    n_sims: int,
    seed: int,
) -> pd.DataFrame:
    """
    Simulate tournament outcomes via Monte Carlo simulation.

    This generates a distribution of possible tournament results by
    simulating each game according to predicted win probabilities.
    The output is Calcutta-agnostic - it only tracks wins per team
    per simulation.

    Args:
        games: Tournament bracket structure
            Required columns: game_id, team1_key, team2_key, round
        predicted_game_outcomes: Win probabilities for each matchup
            Required columns: game_id, team1_key, team2_key,
                p_team1_wins_given_matchup, p_team2_wins_given_matchup
        n_sims: Number of Monte Carlo simulations to run
        seed: Random seed for reproducibility

    Returns:
        DataFrame with columns:
            - sim_id: Simulation number (0 to n_sims-1)
            - team_key: Team identifier
            - wins: Number of games won in this simulation

    Example:
        >>> games = pd.DataFrame({
        ...     "game_id": ["g1", "g2"],
        ...     "team1_key": ["t1", "t2"],
        ...     "team2_key": ["t3", "t4"],
        ...     "round": [1, 1],
        ... })
        >>> predicted = pd.DataFrame({
        ...     "game_id": ["g1", "g2"],
        ...     "team1_key": ["t1", "t2"],
        ...     "team2_key": ["t3", "t4"],
        ...     "p_team1_wins_given_matchup": [0.6, 0.7],
        ...     "p_team2_wins_given_matchup": [0.4, 0.3],
        ... })
        >>> result = simulate_tournaments(
        ...     games=games,
        ...     predicted_game_outcomes=predicted,
        ...     n_sims=100,
        ...     seed=42,
        ... )
        >>> result.columns.tolist()
        ['sim_id', 'team_key', 'wins']
    """
    if games.empty:
        raise ValueError("games must not be empty")
    if predicted_game_outcomes.empty:
        raise ValueError("predicted_game_outcomes must not be empty")
    if n_sims <= 0:
        raise ValueError("n_sims must be positive")

    required_games = ["game_id", "team1_key", "team2_key", "round"]
    missing = [c for c in required_games if c not in games.columns]
    if missing:
        raise ValueError(f"games missing columns: {missing}")

    required_pred = [
        "game_id",
        "team1_key",
        "team2_key",
        "p_team1_wins_given_matchup",
        "p_team2_wins_given_matchup",
    ]
    missing = [
        c for c in required_pred
        if c not in predicted_game_outcomes.columns
    ]
    if missing:
        raise ValueError(f"predicted_game_outcomes missing columns: {missing}")

    # Build bracket graph with next_game tracking
    from moneyball.utils import bracket
    games_graph, prev_by_next = bracket.prepare_bracket_graph(games)

    # Collect all teams from round 1 and 2 only (actual participants)
    all_teams = set()
    for _, gr in games_graph.iterrows():
        round_order = int(gr.get("round_order") or 999)
        if round_order <= 2:
            t1 = str(gr.get("team1_key") or "")
            t2 = str(gr.get("team2_key") or "")
            if t1:
                all_teams.add(t1)
            if t2:
                all_teams.add(t2)

    rng = random.Random(int(seed))
    sim_rows: List[Dict[str, object]] = []

    for sim_i in range(int(n_sims)):
        wins_sim: Dict[str, int] = {}
        winners_by_game: Dict[str, str] = {}

        # Initialize all teams with 0 wins
        for team in all_teams:
            wins_sim[team] = 0

        # Simulate games in bracket order
        for _, gr in games_graph.iterrows():
            gid = str(gr.get("game_id") or "")
            if not gid:
                continue

            round_order = int(gr.get("round_order") or 999)

            # For rounds 1-2, use pre-filled team keys
            if round_order <= 2:
                t1 = str(gr.get("team1_key") or "")
                t2 = str(gr.get("team2_key") or "")
            else:
                # For later rounds, determine teams from previous winners
                t1 = ""
                t2 = ""
                prev = prev_by_next.get(gid, {}).get(1)
                if prev:
                    t1 = winners_by_game.get(prev, "")
                prev = prev_by_next.get(gid, {}).get(2)
                if prev:
                    t2 = winners_by_game.get(prev, "")

            if not t1 or not t2:
                continue

            # Look up win probability for this matchup
            matchup = predicted_game_outcomes[
                (predicted_game_outcomes["game_id"] == gid) &
                (predicted_game_outcomes["team1_key"] == t1) &
                (predicted_game_outcomes["team2_key"] == t2)
            ]

            if len(matchup) == 0:
                p1 = 0.5
            else:
                p1 = float(matchup.iloc[0]["p_team1_wins_given_matchup"])

            # Simulate game outcome
            roll = rng.random()
            winner = t1 if roll < p1 else t2

            # Track winner for bracket progression
            winners_by_game[gid] = winner

            # Increment wins for winner
            wins_sim[winner] = wins_sim.get(winner, 0) + 1

        # Record results for this simulation - include ALL teams
        for team_key, wins in wins_sim.items():
            sim_rows.append({
                "sim_id": int(sim_i),
                "team_key": str(team_key),
                "wins": int(wins),
            })

    if not sim_rows:
        raise ValueError("simulation produced no results")

    return pd.DataFrame(sim_rows)
