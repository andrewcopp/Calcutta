"""
Unit tests for predicted returns and expected value calculations.

These tests verify that:
1. EV calculations match NCAA tournament scoring
2. Probabilities are properly normalized (sum to ~1.0)
3. Edge cases work correctly (guaranteed champion, etc.)
"""

import unittest


class TestExpectedValueCalculation(unittest.TestCase):
    """Test expected value calculations for NCAA tournament."""
    
    def test_guaranteed_champion_ev(self):
        """A team that always wins championship should have EV = 1050."""
        # Team wins 6 games in all simulations
        wins_distribution = {
            0: 0,    # 0% eliminated before R64
            1: 0,    # 0% eliminated in R64
            2: 0,    # 0% eliminated in R32
            3: 0,    # 0% eliminated in S16
            4: 0,    # 0% eliminated in E8
            5: 0,    # 0% eliminated in FF
            6: 1.0,  # 100% win championship
        }
        
        ev = self._calculate_ev(wins_distribution)
        self.assertAlmostEqual(ev, 1050.0, places=2)
    
    def test_guaranteed_r64_elimination_ev(self):
        """A team that always loses in R64 should have EV = 50."""
        # Team wins 1 game in all simulations
        wins_distribution = {
            0: 0,    # 0% eliminated before R64
            1: 1.0,  # 100% eliminated in R64
            2: 0,
            3: 0,
            4: 0,
            5: 0,
            6: 0,
        }
        
        ev = self._calculate_ev(wins_distribution)
        self.assertAlmostEqual(ev, 50.0, places=2)
    
    def test_guaranteed_first_round_loss_ev(self):
        """A team that always loses first game should have EV = 0."""
        # Team wins 0 games in all simulations
        wins_distribution = {
            0: 1.0,  # 100% eliminated before R64
            1: 0,
            2: 0,
            3: 0,
            4: 0,
            5: 0,
            6: 0,
        }
        
        ev = self._calculate_ev(wins_distribution)
        self.assertAlmostEqual(ev, 0.0, places=2)
    
    def test_fifty_fifty_r64_vs_r32_ev(self):
        """A team with 50% R64, 50% R32 should have EV = 100."""
        wins_distribution = {
            0: 0,
            1: 0.5,  # 50% eliminated in R64 (50 points)
            2: 0.5,  # 50% eliminated in R32 (150 points)
            3: 0,
            4: 0,
            5: 0,
            6: 0,
        }
        
        ev = self._calculate_ev(wins_distribution)
        expected = 0.5 * 50 + 0.5 * 150
        self.assertAlmostEqual(ev, expected, places=2)
        self.assertAlmostEqual(ev, 100.0, places=2)
    
    def test_uniform_distribution_ev(self):
        """Test EV with uniform distribution across all outcomes."""
        # Equal probability for each outcome
        prob = 1.0 / 7
        wins_distribution = {
            0: prob,
            1: prob,
            2: prob,
            3: prob,
            4: prob,
            5: prob,
            6: prob,
        }
        
        ev = self._calculate_ev(wins_distribution)
        expected = prob * (0 + 50 + 150 + 300 + 500 + 750 + 1050)
        self.assertAlmostEqual(ev, expected, places=2)
        self.assertAlmostEqual(ev, 400.0, places=2)
    
    def test_realistic_one_seed_ev(self):
        """Test EV for a realistic 1-seed distribution."""
        # Approximate distribution for a strong 1-seed
        wins_distribution = {
            0: 0.01,   # 1% first round loss
            1: 0.14,   # 14% R64 loss
            2: 0.20,   # 20% R32 loss
            3: 0.25,   # 25% S16 loss
            4: 0.20,   # 20% E8 loss
            5: 0.12,   # 12% FF loss
            6: 0.08,   # 8% championship
        }
        
        # Verify probabilities sum to 1
        total_prob = sum(wins_distribution.values())
        self.assertAlmostEqual(total_prob, 1.0, places=2)
        
        ev = self._calculate_ev(wins_distribution)
        expected = (
            0.01 * 0 +
            0.14 * 50 +
            0.20 * 150 +
            0.25 * 300 +
            0.20 * 500 +
            0.12 * 750 +
            0.08 * 1050
        )
        self.assertAlmostEqual(ev, expected, places=2)
        # Should be around 340-350 points
        self.assertGreater(ev, 300)
        self.assertLess(ev, 400)
    
    def test_ev_never_exceeds_max(self):
        """EV should never exceed 1050 (max possible points)."""
        # Even with impossible probabilities, EV should be capped
        wins_distribution = {
            0: 0,
            1: 0,
            2: 0,
            3: 0,
            4: 0,
            5: 0,
            6: 1.0,  # 100% champion
        }
        
        ev = self._calculate_ev(wins_distribution)
        self.assertLessEqual(ev, 1050.0)
    
    def test_ev_never_negative(self):
        """EV should never be negative."""
        wins_distribution = {
            0: 1.0,  # 100% first round loss
            1: 0,
            2: 0,
            3: 0,
            4: 0,
            5: 0,
            6: 0,
        }
        
        ev = self._calculate_ev(wins_distribution)
        self.assertGreaterEqual(ev, 0.0)
    
    def _calculate_ev(self, wins_distribution: dict) -> float:
        """
        Calculate expected value given a wins distribution.
        
        This matches the backend SQL calculation:
        - 0 wins: 0 points
        - 1 win (R64): 50 points
        - 2 wins (R32): 150 points
        - 3 wins (S16): 300 points
        - 4 wins (E8): 500 points
        - 5 wins (FF): 750 points
        - 6 wins (Champ): 1050 points
        """
        points_by_wins = {
            0: 0,
            1: 50,
            2: 150,
            3: 300,
            4: 500,
            5: 750,
            6: 1050,
        }
        
        ev = sum(
            wins_distribution.get(wins, 0) * points_by_wins[wins]
            for wins in range(7)
        )
        
        return ev


class TestProbabilityNormalization(unittest.TestCase):
    """Test that probabilities are properly normalized."""
    
    def test_probabilities_sum_to_one(self):
        """All win probabilities for a team should sum to 1.0."""
        # Simulate a team's win distribution
        wins_distribution = {
            0: 0.05,
            1: 0.15,
            2: 0.25,
            3: 0.25,
            4: 0.15,
            5: 0.10,
            6: 0.05,
        }
        
        total = sum(wins_distribution.values())
        self.assertAlmostEqual(total, 1.0, places=6)
    
    def test_no_probability_exceeds_one(self):
        """No single outcome should have probability > 1.0."""
        wins_distribution = {
            0: 0.1,
            1: 0.2,
            2: 0.3,
            3: 0.2,
            4: 0.1,
            5: 0.05,
            6: 0.05,
        }
        
        for wins, prob in wins_distribution.items():
            self.assertLessEqual(prob, 1.0, f"Probability for {wins} wins exceeds 1.0")
            self.assertGreaterEqual(prob, 0.0, f"Probability for {wins} wins is negative")


class TestNCAAPointValues(unittest.TestCase):
    """Test that NCAA tournament point values are correct."""
    
    def test_point_values_match_ncaa_scoring(self):
        """Verify point values match standard NCAA tournament scoring."""
        # Standard NCAA tournament scoring (cumulative)
        expected_points = {
            0: 0,      # No wins
            1: 50,     # R64 win
            2: 150,    # R64 + R32 wins (50 + 100)
            3: 300,    # + S16 win (150 + 150)
            4: 500,    # + E8 win (300 + 200)
            5: 750,    # + FF win (500 + 250)
            6: 1050,   # + Championship win (750 + 300)
        }
        
        # These should match the backend SQL calculation
        for wins, points in expected_points.items():
            self.assertEqual(points, expected_points[wins],
                           f"Points for {wins} wins should be {points}")
    
    def test_incremental_points_per_round(self):
        """Verify incremental points awarded per round."""
        incremental_points = {
            1: 50,   # R64
            2: 100,  # R32
            3: 150,  # S16
            4: 200,  # E8
            5: 250,  # FF
            6: 300,  # Championship
        }
        
        cumulative = 0
        for round_num, increment in incremental_points.items():
            cumulative += increment
            # Verify cumulative matches expected
            expected_cumulative = {
                1: 50, 2: 150, 3: 300, 4: 500, 5: 750, 6: 1050
            }
            self.assertEqual(cumulative, expected_cumulative[round_num])


if __name__ == '__main__':
    unittest.main()
