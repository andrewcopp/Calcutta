"""
Unit tests for market share writer validation.

Tests that predictions must sum to 1.0 before being written to the database.
"""
from __future__ import annotations

import unittest
from unittest.mock import patch, MagicMock

import pandas as pd

from moneyball.db.writers.silver_writers import write_predicted_market_share_with_run


class TestThatMarketShareWriterValidatesSumToOne(unittest.TestCase):
    """Test that write_predicted_market_share_with_run validates prediction sum."""

    def test_that_predictions_summing_to_less_than_one_raise_error(self) -> None:
        # GIVEN predictions that sum to 0.38 (the bug case)
        predictions_df = pd.DataFrame({
            "team_key": ["duke", "unc", "kentucky"],
            "predicted_auction_share_of_pool": [0.2, 0.1, 0.08],  # sums to 0.38
        })
        team_id_map = {
            "duke": "uuid-duke",
            "unc": "uuid-unc",
            "kentucky": "uuid-kentucky",
        }

        # WHEN/THEN writing raises ValueError with helpful message
        with self.assertRaises(ValueError) as context:
            write_predicted_market_share_with_run(
                predictions_df=predictions_df,
                team_id_map=team_id_map,
                tournament_id="uuid-tournament",
            )

        self.assertIn("must sum to 1.0", str(context.exception))
        self.assertIn("0.38", str(context.exception))

    def test_that_predictions_summing_to_more_than_one_raise_error(self) -> None:
        # GIVEN predictions that sum to > 1.0
        predictions_df = pd.DataFrame({
            "team_key": ["duke", "unc"],
            "predicted_auction_share_of_pool": [0.8, 0.8],  # sums to 1.6
        })
        team_id_map = {
            "duke": "uuid-duke",
            "unc": "uuid-unc",
        }

        # WHEN/THEN writing raises ValueError
        with self.assertRaises(ValueError) as context:
            write_predicted_market_share_with_run(
                predictions_df=predictions_df,
                team_id_map=team_id_map,
                tournament_id="uuid-tournament",
            )

        self.assertIn("must sum to 1.0", str(context.exception))
        self.assertIn("1.6", str(context.exception))

    def test_that_validation_error_includes_hint_about_optimal_v2(self) -> None:
        # GIVEN predictions that don't sum to 1.0
        predictions_df = pd.DataFrame({
            "team_key": ["duke"],
            "predicted_auction_share_of_pool": [0.5],
        })
        team_id_map = {"duke": "uuid-duke"}

        # WHEN/THEN error message includes hint about optimal_v2 hyperparameters
        with self.assertRaises(ValueError) as context:
            write_predicted_market_share_with_run(
                predictions_df=predictions_df,
                team_id_map=team_id_map,
                tournament_id="uuid-tournament",
            )

        self.assertIn("seed_prior_k", str(context.exception))
        self.assertIn("optimal_v2", str(context.exception))

    def test_that_predictions_within_tolerance_pass_validation(self) -> None:
        # GIVEN predictions that sum to 1.0005 (within 0.001 tolerance)
        predictions_df = pd.DataFrame({
            "team_key": ["duke", "unc"],
            "predicted_auction_share_of_pool": [0.50025, 0.50025],  # sums to 1.0005
        })
        team_id_map = {
            "duke": "uuid-duke",
            "unc": "uuid-unc",
        }

        # WHEN/THEN no ValueError is raised (DB error is expected, but not validation)
        # We expect a RuntimeError from DB connection, not ValueError from validation
        with self.assertRaises(RuntimeError):
            write_predicted_market_share_with_run(
                predictions_df=predictions_df,
                team_id_map=team_id_map,
                tournament_id="uuid-tournament",
            )


if __name__ == "__main__":
    unittest.main()
