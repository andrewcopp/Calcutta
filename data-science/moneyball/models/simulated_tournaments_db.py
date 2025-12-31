"""
DB-first simulation functions.

Helper functions for simulating tournaments using data from database.
"""
from __future__ import annotations

import random
from typing import Dict
import pandas as pd


def simulate_tournaments_from_predictions(
    *,
    predictions_df: pd.DataFrame,
    teams_df: pd.DataFrame,
    n_sims: int = 5000,
    seed: int = 42,
) -> pd.DataFrame:
    """
    Simulate tournaments from predicted game outcomes.
    
    Args:
        predictions_df: DataFrame with game predictions
        teams_df: DataFrame with team data
        n_sims: Number of simulations
        seed: Random seed
        
    Returns:
        DataFrame with simulation results (team_id, sim_id, wins, byes, points)
    """
    rng = random.Random(seed)
    
    # Create team lookup
    team_info = {}
    for _, row in teams_df.iterrows():
        team_info[str(row['id'])] = {
            'school_name': row['school_name'],
            'seed': row['seed'],
            'region': row['region'],
        }
    
    # Simple simulation: for each sim, pick winners based on probabilities
    results = []
    
    for sim_id in range(n_sims):
        team_wins = {}
        team_byes = {}
        
        # For each team, simulate their performance
        for team_id in team_info.keys():
            # Simple model: wins based on seed (better seed = more wins)
            # This is a placeholder - you'd want to use actual bracket simulation
            seed_val = team_info[team_id]['seed']
            
            # Expected wins based on seed (rough approximation)
            if seed_val <= 2:
                expected_wins = 4 + rng.random() * 2
            elif seed_val <= 4:
                expected_wins = 3 + rng.random() * 2
            elif seed_val <= 8:
                expected_wins = 2 + rng.random() * 2
            else:
                expected_wins = 1 + rng.random() * 2
            
            wins = int(expected_wins)
            byes = 0
            
            # Calculate points (standard NCAA tournament scoring)
            points = calculate_points(wins, byes)
            
            team_wins[team_id] = wins
            team_byes[team_id] = byes
            
            results.append({
                'team_id': team_id,
                'sim_id': sim_id,
                'wins': wins,
                'byes': byes,
                'points': points,
                'school_name': team_info[team_id]['school_name'],
                'seed': team_info[team_id]['seed'],
                'region': team_info[team_id]['region'],
            })
    
    return pd.DataFrame(results)


def calculate_points(wins: int, byes: int) -> int:
    """
    Calculate points for a team based on wins and byes.
    
    Standard NCAA tournament scoring:
    - Round of 64: 0 points
    - Round of 32: 50 points
    - Sweet 16: 150 points
    - Elite 8: 300 points
    - Final 4: 500 points
    - Championship: 750 points
    - Winner: 1050 points
    
    Args:
        wins: Number of wins
        byes: Number of byes
        
    Returns:
        Total points
    """
    total_rounds = wins + byes
    
    if total_rounds == 0:
        return 0
    elif total_rounds == 1:
        return 0
    elif total_rounds == 2:
        return 50
    elif total_rounds == 3:
        return 150
    elif total_rounds == 4:
        return 300
    elif total_rounds == 5:
        return 500
    elif total_rounds == 6:
        return 750
    elif total_rounds >= 7:
        return 1050
    else:
        return 0
