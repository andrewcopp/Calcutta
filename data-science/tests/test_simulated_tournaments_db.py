"""
Unit tests for DB-first tournament simulation.

Verifies that the DB-first simulation (using team/region/seed data)
produces correct results with proper win bounds.
"""
from __future__ import annotations

import unittest
import pandas as pd

from moneyball.models.simulated_tournaments_db import (
    simulate_tournaments_from_predictions,
    calculate_points,
)


class TestDBFirstSimulationMaxWins(unittest.TestCase):
    """Test that DB-first simulation respects maximum win bounds."""
    
    def test_max_wins_is_six_for_64_team_bracket(self):
        """
        GIVEN a 64-team tournament (4 regions, 16 teams each)
        WHEN simulations are run
        THEN no team should win more than 6 games
        
        This is the critical test - ensures we fixed the bug where
        wins were capped at 5 instead of 6.
        """
        teams_df, predictions_df = _create_64_team_bracket()
        
        result = simulate_tournaments_from_predictions(
            predictions_df=predictions_df,
            teams_df=teams_df,
            n_sims=100,
            seed=42,
        )
        
        max_wins = result['wins'].max()
        self.assertLessEqual(
            max_wins,
            6,
            f"Team won {max_wins} games but max should be 6"
        )
        self.assertGreaterEqual(
            max_wins,
            5,
            "At least one team should win 5+ games in 100 simulations"
        )
    
    def test_champion_wins_exactly_six_games(self):
        """
        GIVEN a 64-team tournament
        WHEN a team wins the championship
        THEN they should have exactly 6 wins (R64, R32, S16, E8, FF, Champ)
        """
        teams_df, predictions_df = _create_64_team_bracket()
        
        result = simulate_tournaments_from_predictions(
            predictions_df=predictions_df,
            teams_df=teams_df,
            n_sims=1000,
            seed=42,
        )
        
        # Check that some teams win exactly 6 games
        six_win_count = len(result[result['wins'] == 6])
        self.assertGreater(
            six_win_count,
            0,
            "Should have at least one champion with 6 wins in 1000 sims"
        )
    
    def test_all_teams_included_in_results(self):
        """
        GIVEN a tournament with N teams
        WHEN simulations are run
        THEN all N teams should appear in results (even with 0 wins)
        """
        teams_df, predictions_df = _create_64_team_bracket()
        n_teams = len(teams_df)
        
        result = simulate_tournaments_from_predictions(
            predictions_df=predictions_df,
            teams_df=teams_df,
            n_sims=10,
            seed=42,
        )
        
        # Each simulation should have all teams
        for sim_id in range(10):
            sim_teams = result[result['sim_id'] == sim_id]['team_id'].nunique()
            self.assertEqual(
                sim_teams,
                n_teams,
                f"Simulation {sim_id} has {sim_teams} teams, expected {n_teams}"
            )


class TestDBFirstSimulationDeterminism(unittest.TestCase):
    """Test that simulations are deterministic with same seed."""
    
    def test_same_seed_produces_same_results(self):
        """
        GIVEN the same teams and predictions
        WHEN simulations are run twice with the same seed
        THEN results should be identical
        """
        teams_df, predictions_df = _create_64_team_bracket()
        
        result1 = simulate_tournaments_from_predictions(
            predictions_df=predictions_df,
            teams_df=teams_df,
            n_sims=50,
            seed=123,
        )
        
        result2 = simulate_tournaments_from_predictions(
            predictions_df=predictions_df,
            teams_df=teams_df,
            n_sims=50,
            seed=123,
        )
        
        pd.testing.assert_frame_equal(result1, result2)


class TestCalculatePoints(unittest.TestCase):
    """Test NCAA tournament point calculation."""
    
    def test_point_values_by_round(self):
        """Test standard NCAA tournament scoring."""
        # Round of 64 (1 win) = 0 points
        self.assertEqual(calculate_points(1, 0), 0)
        
        # Round of 32 (2 wins) = 50 points
        self.assertEqual(calculate_points(2, 0), 50)
        
        # Sweet 16 (3 wins) = 150 points
        self.assertEqual(calculate_points(3, 0), 150)
        
        # Elite 8 (4 wins) = 300 points
        self.assertEqual(calculate_points(4, 0), 300)
        
        # Final Four (5 wins) = 500 points
        self.assertEqual(calculate_points(5, 0), 500)
        
        # Championship (6 wins) = 750 points
        self.assertEqual(calculate_points(6, 0), 750)
        
        # Winner (7 wins with bye) = 1050 points
        self.assertEqual(calculate_points(6, 1), 1050)
        self.assertEqual(calculate_points(7, 0), 1050)


def _create_64_team_bracket():
    """
    Create a simple 64-team bracket for testing.
    
    Returns:
        Tuple of (teams_df, predictions_df)
    """
    # Create 4 regions with 16 teams each
    regions = ['East', 'West', 'South', 'Midwest']
    teams = []
    team_id_counter = 0
    
    for region in regions:
        for seed in range(1, 17):
            teams.append({
                'id': f'team_{team_id_counter}',
                'school_name': f'{region} Seed {seed}',
                'seed': seed,
                'region': region,
            })
            team_id_counter += 1
    
    teams_df = pd.DataFrame(teams)
    
    # Create predictions for all possible matchups
    # Simple model: lower seed has 60% chance to win
    predictions = []
    for i, team1 in enumerate(teams):
        for j, team2 in enumerate(teams):
            if i >= j:
                continue
            
            # Lower seed number = better team
            if team1['seed'] < team2['seed']:
                p_team1_wins = 0.6
            elif team1['seed'] > team2['seed']:
                p_team1_wins = 0.4
            else:
                p_team1_wins = 0.5
            
            predictions.append({
                'team1_id': team1['id'],
                'team2_id': team2['id'],
                'p_team1_wins': p_team1_wins,
                'round': 0,  # Not used in DB-first simulation
            })
    
    predictions_df = pd.DataFrame(predictions)
    
    return teams_df, predictions_df


if __name__ == '__main__':
    unittest.main()
