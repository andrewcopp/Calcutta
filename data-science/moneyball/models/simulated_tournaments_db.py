"""
DB-first simulation functions.

Simulates NCAA tournament brackets using team/region/seed data from database.
Uses predictions table as a lookup for win probabilities.

This is a pure DB-first approach - no parquet files or games DataFrame needed.
The bracket structure is determined from seed/region, and predictions are
looked up dynamically during simulation.
"""
from __future__ import annotations

import random
from typing import Dict, Optional, Set, Tuple
import pandas as pd

from moneyball.utils import points


def simulate_tournaments_from_predictions(
    *,
    predictions_df: pd.DataFrame,
    teams_df: pd.DataFrame,
    points_by_win_index: Optional[Dict[int, float]] = None,
    n_sims: int = 5000,
    seed: int = 42,
) -> pd.DataFrame:
    """
    Simulate NCAA tournament brackets using DB-first approach.

    Uses team seed/region to determine bracket pairings, then looks up
    win probabilities from predictions_df (which contains all possible
    matchups as a lookup table).

    Args:
        predictions_df: All possible matchup predictions (lookup table)
            Required columns: team1_id, team2_id, p_team1_wins_given_matchup
        teams_df: Team data with seed and region
            Required columns: id, school_name, seed, region
        n_sims: Number of simulations
        seed: Random seed

    Returns:
        DataFrame with columns: team_id, sim_id, wins, byes, points,
        school_name, seed, region
    """
    rng = random.Random(seed)

    # Build team lookup by region and seed for bracket construction
    teams_by_region_seed: Dict[str, Dict[int, list[str]]] = {}
    team_info: Dict[str, Dict] = {}

    for _, row in teams_df.iterrows():
        team_id = str(row['id'])
        region = str(row['region'])
        seed_val = int(row['seed'])

        team_info[team_id] = {
            'school_name': row['school_name'],
            'seed': seed_val,
            'region': region,
        }

        if region not in teams_by_region_seed:
            teams_by_region_seed[region] = {}
        if seed_val not in teams_by_region_seed[region]:
            teams_by_region_seed[region][seed_val] = []
        teams_by_region_seed[region][seed_val].append(team_id)

    # Build prediction lookup: (team1_id, team2_id) -> probability
    pred_lookup: Dict[Tuple[str, str], float] = {}
    for _, row in predictions_df.iterrows():
        t1 = str(row['team1_id'])
        t2 = str(row['team2_id'])
        p = float(row['p_team1_wins_given_matchup'])
        pred_lookup[(t1, t2)] = p
        pred_lookup[(t2, t1)] = 1.0 - p  # Reverse matchup

    results = []

    for sim_id in range(n_sims):
        team_wins: Dict[str, int] = {tid: 0 for tid in team_info.keys()}
        team_byes: Dict[str, int] = {tid: 0 for tid in team_info.keys()}

        # Simulate bracket progression
        simulate_bracket(
            teams_by_region_seed=teams_by_region_seed,
            pred_lookup=pred_lookup,
            team_wins=team_wins,
            team_byes=team_byes,
            rng=rng,
        )

        # Record results for all teams
        for team_id in team_info.keys():
            wins = team_wins[team_id]
            byes = team_byes[team_id]
            points_scored = calculate_points(
                wins,
                byes,
                points_by_win_index=points_by_win_index,
            )

            results.append({
                'team_id': team_id,
                'sim_id': sim_id,
                'wins': wins,
                'byes': byes,
                'points': points_scored,
                'school_name': team_info[team_id]['school_name'],
                'seed': team_info[team_id]['seed'],
                'region': team_info[team_id]['region'],
            })

    return pd.DataFrame(results)


def simulate_bracket(
    *,
    teams_by_region_seed: Dict[str, Dict[int, list[str]]],
    pred_lookup: Dict[Tuple[str, str], float],
    team_wins: Dict[str, int],
    team_byes: Dict[str, int],
    rng: random.Random,
) -> None:
    """
    Simulate NCAA tournament bracket progression.

    Standard NCAA bracket structure:
    - 4 regions with 16 teams each (seeds 1-16)
    - Round of 64: 32 games (1v16, 2v15, 3v14, 4v13, 5v12, 6v11, 7v10, 8v9)
    - Round of 32: 16 games (winners advance)
    - Sweet 16: 8 games
    - Elite 8: 4 games (regional finals)
    - Final Four: 2 games (regional champions)
    - Championship: 1 game

    Mutates team_wins and team_byes dictionaries in place.
    """
    # Track alive teams per region
    alive_by_region: Dict[str, Set[str]] = {}

    for region in sorted(teams_by_region_seed.keys()):
        teams = teams_by_region_seed[region]
        alive: Set[str] = set(
            team_id for seed_teams in teams.values() for team_id in seed_teams
        )

        resolved_teams_by_seed: Dict[int, str] = {}
        for seed_val in sorted(teams.keys()):
            team_ids = teams[seed_val]
            if not team_ids:
                continue
            if len(team_ids) == 1:
                resolved_teams_by_seed[seed_val] = team_ids[0]
                continue

            candidates = sorted(team_ids)
            while len(candidates) > 1:
                t1 = candidates[0]
                t2 = candidates[1]
                winner = simulate_game(t1, t2, pred_lookup, team_wins, rng)
                candidates = candidates[2:] + [winner]

            winner = candidates[0]
            resolved_teams_by_seed[seed_val] = winner
            for team_id in team_ids:
                if team_id != winner and team_id in alive:
                    alive.remove(team_id)

        # Round of 64 - standard bracket pairings (8 games per region)
        alive = simulate_round_by_seed(
            alive_teams=alive,
            pairings=[(1, 16), (8, 9), (5, 12), (4, 13),
                      (6, 11), (3, 14), (7, 10), (2, 15)],
            teams_by_seed=resolved_teams_by_seed,
            pred_lookup=pred_lookup,
            team_wins=team_wins,
            rng=rng,
        )

        # Round of 32 - winners advance (4 games per region)
        alive = simulate_round_any_matchup(
            alive_teams=alive,
            pred_lookup=pred_lookup,
            team_wins=team_wins,
            rng=rng,
            expected_games=4,
        )

        # Sweet 16 - winners advance (2 games per region)
        alive = simulate_round_any_matchup(
            alive_teams=alive,
            pred_lookup=pred_lookup,
            team_wins=team_wins,
            rng=rng,
            expected_games=2,
        )

        # Elite 8 - regional final (1 game per region)
        alive = simulate_round_any_matchup(
            alive_teams=alive,
            pred_lookup=pred_lookup,
            team_wins=team_wins,
            rng=rng,
            expected_games=1,
        )

        alive_by_region[region] = alive

    # Final Four - combine regional champions
    all_alive = set()
    for alive in alive_by_region.values():
        all_alive.update(alive)

    # Simulate Final Four (2 games) + Championship
    regions = sorted(alive_by_region.keys())
    if len(regions) >= 4 and len(all_alive) >= 4:
        # Semifinal 1
        game1_teams = (
            sorted(alive_by_region[regions[0]])
            + sorted(alive_by_region[regions[1]])
        )
        if len(game1_teams) == 2:
            winner1 = simulate_game(
                game1_teams[0], game1_teams[1],
                pred_lookup, team_wins, rng
            )
        else:
            winner1 = game1_teams[0] if game1_teams else None

        # Semifinal 2
        game2_teams = (
            sorted(alive_by_region[regions[2]])
            + sorted(alive_by_region[regions[3]])
        )
        if len(game2_teams) == 2:
            winner2 = simulate_game(
                game2_teams[0], game2_teams[1],
                pred_lookup, team_wins, rng
            )
        else:
            winner2 = game2_teams[0] if game2_teams else None

        # Championship
        if winner1 and winner2:
            simulate_game(winner1, winner2, pred_lookup, team_wins, rng)


def simulate_round_by_seed(
    *,
    alive_teams: Set[str],
    pairings: list,
    teams_by_seed: Dict[int, str],
    pred_lookup: Dict[Tuple[str, str], float],
    team_wins: Dict[str, int],
    rng: random.Random,
) -> Set[str]:
    """Simulate one round of bracket games based on seed pairings."""
    winners = set()

    for seed1, seed2 in pairings:
        team1 = teams_by_seed.get(seed1)
        team2 = teams_by_seed.get(seed2)

        if not team1 or not team2:
            continue
        if team1 not in alive_teams or team2 not in alive_teams:
            continue

        winner = simulate_game(team1, team2, pred_lookup, team_wins, rng)
        winners.add(winner)

    return winners


def simulate_round_any_matchup(
    *,
    alive_teams: Set[str],
    pred_lookup: Dict[Tuple[str, str], float],
    team_wins: Dict[str, int],
    rng: random.Random,
    expected_games: int,
) -> Set[str]:
    """
    Simulate a round where any alive team can play any other alive team.

    Pairs teams sequentially and simulates games until expected number
    of games is reached or we run out of teams.
    """
    winners = set()
    teams_list = sorted(alive_teams)

    # Pair teams sequentially (simple bracket progression)
    for i in range(0, len(teams_list) - 1, 2):
        if len(winners) >= expected_games:
            break

        team1 = teams_list[i]
        team2 = teams_list[i + 1]

        winner = simulate_game(team1, team2, pred_lookup, team_wins, rng)
        winners.add(winner)

    return winners


def simulate_game(
    team1_id: str,
    team2_id: str,
    pred_lookup: Dict[Tuple[str, str], float],
    team_wins: Dict[str, int],
    rng: random.Random,
) -> str:
    """Simulate a single game and return winner."""
    # Look up win probability
    p_team1_wins = pred_lookup.get((team1_id, team2_id), 0.5)

    # Simulate outcome
    if rng.random() < p_team1_wins:
        winner = team1_id
    else:
        winner = team2_id

    # Update wins
    team_wins[winner] += 1

    return winner


def calculate_points(
    wins: int,
    byes: int,
    *,
    points_by_win_index: Optional[Dict[int, float]] = None,
) -> int:
    """
    Calculate points for a team based on wins and byes.

    Uses scoring rules derived from core.calcutta_scoring_rules.

    Args:
        wins: Number of wins
        byes: Number of byes

    Returns:
        Total points
    """
    total_rounds = int(wins) + int(byes)
    if points_by_win_index is not None:
        return int(
            round(
                points.team_points_from_scoring_rules(
                    total_rounds,
                    points_by_win_index,
                )
            )
        )

    return int(round(points.team_points_fixed(total_rounds)))
