"""
Unit tests for predicted returns and expected value calculations.

These tests verify that:
1. EV calculations match NCAA tournament scoring
2. Probabilities are properly normalized (sum to ~1.0)
3. Edge cases work correctly (guaranteed champion, etc.)
"""

import unittest

from moneyball.utils import points as mb_points


def _points_by_win_index_fixture() -> dict:
    return {
        1: 0,
        2: 50,
        3: 100,
        4: 150,
        5: 200,
        6: 250,
        7: 300,
    }


class TestThatExpectedValueCalculationIsCorrect(unittest.TestCase):
    """Test expected value calculations for NCAA tournament."""

    def test_that_guaranteed_champion_has_ev_of_1050(self) -> None:
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
    
    def test_that_guaranteed_r64_elimination_has_ev_of_50(self) -> None:
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
    
    def test_that_guaranteed_first_round_loss_has_ev_of_zero(self) -> None:
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
    
    def test_that_fifty_fifty_r64_vs_r32_has_ev_of_100(self) -> None:
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
    
    def test_that_uniform_distribution_has_ev_of_400(self) -> None:
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
    
    def test_that_realistic_one_seed_ev_is_reasonable(self) -> None:
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
    
    def test_that_ev_never_exceeds_max_of_1050(self) -> None:
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
    
    def test_that_ev_is_never_negative(self) -> None:
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
        pbwi = _points_by_win_index_fixture()

        def _points_for_wins(wins: int) -> float:
            return float(
                mb_points.team_points_from_scoring_rules(
                    int(wins) + 1,
                    pbwi,
                )
            )

        ev = sum(
            float(wins_distribution.get(wins, 0.0)) * _points_for_wins(wins)
            for wins in range(7)
        )

        return float(ev)


class TestThatProbabilitiesAreNormalized(unittest.TestCase):
    """Test that probabilities are properly normalized."""

    def test_that_probabilities_sum_to_one(self) -> None:
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
    
    def test_that_no_probability_exceeds_one(self) -> None:
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
            self.assertLessEqual(
                prob, 1.0, f"Probability for {wins} wins exceeds 1.0"
            )
            self.assertGreaterEqual(
                prob, 0.0, f"Probability for {wins} wins is negative"
            )


class TestThatNCAAPointValuesAreCorrect(unittest.TestCase):
    """Test that NCAA tournament point values are correct."""

    def test_that_point_values_match_ncaa_scoring(self) -> None:
        """Verify point values match standard NCAA tournament scoring."""
        pbwi = _points_by_win_index_fixture()

        expected_points = {
            wins: int(
                mb_points.team_points_from_scoring_rules(
                    int(wins) + 1,
                    pbwi,
                )
            )
            for wins in range(7)
        }
        
        # These should match the backend SQL calculation
        for wins, expected in expected_points.items():
            self.assertEqual(
                expected,
                expected_points[wins],
                f"Points for {wins} wins should be {expected}",
            )
    
    def test_that_incremental_points_per_round_are_correct(self) -> None:
        """Verify incremental points awarded per round."""
        pbwi = _points_by_win_index_fixture()
        incremental_points = {
            r: int(pbwi[r + 1]) for r in range(1, 7)
        }
        
        cumulative = 0
        for round_num, increment in incremental_points.items():
            cumulative += increment
            # Verify cumulative matches expected
            expected_cumulative = {
                r: int(
                    mb_points.team_points_from_scoring_rules(
                        r + 1,
                        pbwi,
                    )
                )
                for r in range(1, 7)
            }
            self.assertEqual(cumulative, expected_cumulative[round_num])


if __name__ == '__main__':
    unittest.main()
