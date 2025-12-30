"""
Unit tests for portfolio construction artifacts.

Following testing guidelines:
- GIVEN / WHEN / THEN structure
- Exactly one reason to fail / one assertion per test
- Deterministic tests
- Test naming: TestThat{Scenario}
"""
import unittest
import pandas as pd

from moneyball.models.portfolio_construction import generate_recommended_bids


class TestThatOwnershipAccountsForBidMovingMarket(unittest.TestCase):
    """
    GIVEN a team with predicted cost of 2 points and our bid of 5 points
    WHEN we calculate ownership fraction
    THEN ownership should be 5/(2+5) = 71.4%, not 5/2 = 250%
    """

    def test(self):
        # GIVEN
        recommended_bids = pd.DataFrame({
            "team_key": ["team_a"],
            "bid_amount_points": [5],
        })

        tournament_value = pd.DataFrame({
            "team_key": ["team_a"],
            "expected_points_per_entry": [10.0],
        })

        market_prediction = pd.DataFrame({
            "team_key": ["team_a"],
            "predicted_market_share": [0.02],
        })

        predicted_total_pool = 100.0

        # WHEN
        _, debug = generate_recommended_bids(
            recommended_entry_bids=recommended_bids,
            tournament_value=tournament_value,
            market_prediction=market_prediction,
            predicted_total_pool_bids_points=predicted_total_pool,
        )

        # THEN
        expected_ownership = 5.0 / (2.0 + 5.0)
        actual_ownership = debug.loc[0, "ownership_fraction"]
        self.assertAlmostEqual(actual_ownership, expected_ownership, places=3)


class TestThatOwnershipIsZeroWhenNoBid(unittest.TestCase):
    """
    GIVEN a team with no bid
    WHEN we calculate ownership fraction
    THEN ownership should be 0%
    """

    def test(self):
        # GIVEN
        recommended_bids = pd.DataFrame({
            "team_key": ["team_a"],
            "bid_amount_points": [0],
        })

        tournament_value = pd.DataFrame({
            "team_key": ["team_a"],
            "expected_points_per_entry": [10.0],
        })

        market_prediction = pd.DataFrame({
            "team_key": ["team_a"],
            "predicted_market_share": [0.05],
        })

        # WHEN
        _, debug = generate_recommended_bids(
            recommended_entry_bids=recommended_bids,
            tournament_value=tournament_value,
            market_prediction=market_prediction,
            predicted_total_pool_bids_points=100.0,
        )

        # THEN
        self.assertEqual(debug.loc[0, "ownership_fraction"], 0.0)


class TestThatOwnershipIsFiftyPercentWhenBidEqualsPredictedCost(
    unittest.TestCase
):
    """
    GIVEN a team with predicted cost of 10 and our bid of 10
    WHEN we calculate ownership fraction
    THEN ownership should be 10/(10+10) = 50%
    """

    def test(self):
        # GIVEN
        recommended_bids = pd.DataFrame({
            "team_key": ["team_a"],
            "bid_amount_points": [10],
        })

        tournament_value = pd.DataFrame({
            "team_key": ["team_a"],
            "expected_points_per_entry": [20.0],
        })

        market_prediction = pd.DataFrame({
            "team_key": ["team_a"],
            "predicted_market_share": [0.10],
        })

        # WHEN
        _, debug = generate_recommended_bids(
            recommended_entry_bids=recommended_bids,
            tournament_value=tournament_value,
            market_prediction=market_prediction,
            predicted_total_pool_bids_points=100.0,
        )

        # THEN
        self.assertAlmostEqual(
            debug.loc[0, "ownership_fraction"], 0.5, places=3
        )


class TestThatRoiCalculationUsesCorrectOwnership(unittest.TestCase):
    """
    GIVEN a team with expected points of 100, predicted cost of 20,
    and bid of 10
    WHEN we calculate ROI
    THEN ROI should be (100 * 10/30) / 10 = 3.33x
    """

    def test(self):
        # GIVEN
        recommended_bids = pd.DataFrame({
            "team_key": ["team_a"],
            "bid_amount_points": [10],
        })

        tournament_value = pd.DataFrame({
            "team_key": ["team_a"],
            "expected_points_per_entry": [100.0],
        })

        market_prediction = pd.DataFrame({
            "team_key": ["team_a"],
            "predicted_market_share": [0.20],
        })

        # WHEN
        _, debug = generate_recommended_bids(
            recommended_entry_bids=recommended_bids,
            tournament_value=tournament_value,
            market_prediction=market_prediction,
            predicted_total_pool_bids_points=100.0,
        )

        # THEN
        ownership = 10.0 / (20.0 + 10.0)
        expected_return = 100.0 * ownership
        expected_roi = expected_return / 10.0

        actual_roi = debug.loc[0, "roi"]
        self.assertAlmostEqual(actual_roi, expected_roi, places=2)
