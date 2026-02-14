from __future__ import annotations

import unittest

import pandas as pd

from moneyball.models.predicted_auction_share_of_pool import (
    predict_auction_share_of_pool,
)


class TestThatPredictedAuctionShareOfPoolIsNormalized(unittest.TestCase):
    def test_that_predicted_auction_share_of_pool_sums_to_one(self) -> None:
        train = pd.DataFrame(
            {
                "seed": [1, 2, 3, 4, 5],
                "region": ["East", "East", "West", "West", "South"],
                "kenpom_net": [30.0, 20.0, 10.0, 5.0, 15.0],
                "team_share_of_pool": [0.3, 0.25, 0.15, 0.1, 0.2],
            }
        )
        pred = pd.DataFrame(
            {
                "seed": [1, 10, 16],
                "region": ["East", "West", "South"],
                "kenpom_net": [35.0, 0.0, -5.0],
            }
        )

        out = predict_auction_share_of_pool(
            train_team_dataset=train,
            predict_team_dataset=pred,
            ridge_alpha=1.0,
            feature_set="basic",
        )

        self.assertAlmostEqual(
            float(out["predicted_auction_share_of_pool"].sum()),
            1.0,
            places=8,
        )


class TestThatPredictedAuctionShareOfPoolIsNonNegative(unittest.TestCase):
    def test_that_predicted_share_of_pool_is_non_negative(self) -> None:
        train = pd.DataFrame(
            {
                "seed": [1, 2, 3, 4, 5],
                "region": ["East", "East", "West", "West", "South"],
                "kenpom_net": [30.0, 20.0, 10.0, 5.0, 15.0],
                "team_share_of_pool": [0.3, 0.25, 0.15, 0.1, 0.2],
            }
        )
        pred = pd.DataFrame(
            {
                "seed": [1, 10, 16],
                "region": ["East", "West", "South"],
                "kenpom_net": [35.0, 0.0, -5.0],
            }
        )

        out = predict_auction_share_of_pool(
            train_team_dataset=train,
            predict_team_dataset=pred,
            ridge_alpha=1.0,
            feature_set="basic",
        )

        self.assertTrue(
            bool((out["predicted_auction_share_of_pool"] >= 0.0).all())
        )


class TestThatPredictedAuctionShareOfPoolIncludesKeyColumns(unittest.TestCase):
    def test_that_output_includes_predicted_share_of_pool(self) -> None:
        train = pd.DataFrame(
            {
                "seed": [1, 2, 3, 4, 5],
                "region": ["East", "East", "West", "West", "South"],
                "kenpom_net": [30.0, 20.0, 10.0, 5.0, 15.0],
                "team_share_of_pool": [0.3, 0.25, 0.15, 0.1, 0.2],
            }
        )
        pred = pd.DataFrame(
            {
                "seed": [1, 10, 16],
                "region": ["East", "West", "South"],
                "kenpom_net": [35.0, 0.0, -5.0],
            }
        )

        out = predict_auction_share_of_pool(
            train_team_dataset=train,
            predict_team_dataset=pred,
            ridge_alpha=1.0,
            feature_set="basic",
        )

        self.assertIn("predicted_auction_share_of_pool", out.columns)


class TestThatOptimalV2RequiresSeedPriorK(unittest.TestCase):
    """Test that optimal_v2 feature set requires seed_prior_k > 0."""

    def test_that_optimal_v2_with_zero_seed_prior_k_raises_error(self) -> None:
        # GIVEN training and prediction data
        train = pd.DataFrame(
            {
                "seed": [1, 2, 3, 4, 5],
                "region": ["East", "East", "West", "West", "South"],
                "kenpom_net": [30.0, 20.0, 10.0, 5.0, 15.0],
                "kenpom_o": [120.0, 115.0, 110.0, 108.0, 112.0],
                "kenpom_d": [95.0, 98.0, 100.0, 102.0, 99.0],
                "kenpom_adj_t": [70.0, 68.0, 66.0, 65.0, 67.0],
                "team_share_of_pool": [0.3, 0.25, 0.15, 0.1, 0.2],
                "school_slug": ["duke", "unc", "kentucky", "kansas", "gonzaga"],
            }
        )
        pred = pd.DataFrame(
            {
                "seed": [1, 10, 16],
                "region": ["East", "West", "South"],
                "kenpom_net": [35.0, 0.0, -5.0],
                "kenpom_o": [125.0, 105.0, 95.0],
                "kenpom_d": [90.0, 105.0, 110.0],
                "kenpom_adj_t": [72.0, 64.0, 60.0],
                "school_slug": ["duke", "team10", "team16"],
            }
        )

        # WHEN/THEN using optimal_v2 with seed_prior_k=0 raises ValueError
        with self.assertRaises(ValueError) as context:
            predict_auction_share_of_pool(
                train_team_dataset=train,
                predict_team_dataset=pred,
                ridge_alpha=1.0,
                feature_set="optimal_v2",
                seed_prior_k=0,
            )

        self.assertIn("seed_prior_k > 0", str(context.exception))
        self.assertIn("optimal_v2", str(context.exception))

    def test_that_optimal_v2_with_positive_seed_prior_k_works(self) -> None:
        # GIVEN training and prediction data
        train = pd.DataFrame(
            {
                "seed": [1, 2, 3, 4, 5],
                "region": ["East", "East", "West", "West", "South"],
                "kenpom_net": [30.0, 20.0, 10.0, 5.0, 15.0],
                "kenpom_o": [120.0, 115.0, 110.0, 108.0, 112.0],
                "kenpom_d": [95.0, 98.0, 100.0, 102.0, 99.0],
                "kenpom_adj_t": [70.0, 68.0, 66.0, 65.0, 67.0],
                "team_share_of_pool": [0.3, 0.25, 0.15, 0.1, 0.2],
                "school_slug": ["duke", "unc", "kentucky", "kansas", "gonzaga"],
            }
        )
        pred = pd.DataFrame(
            {
                "seed": [1, 10, 16],
                "region": ["East", "West", "South"],
                "kenpom_net": [35.0, 0.0, -5.0],
                "kenpom_o": [125.0, 105.0, 95.0],
                "kenpom_d": [90.0, 105.0, 110.0],
                "kenpom_adj_t": [72.0, 64.0, 60.0],
                "school_slug": ["duke", "team10", "team16"],
            }
        )

        # WHEN using optimal_v2 with seed_prior_k=20 (recommended value)
        out = predict_auction_share_of_pool(
            train_team_dataset=train,
            predict_team_dataset=pred,
            ridge_alpha=1.0,
            feature_set="optimal_v2",
            seed_prior_k=20,
            program_prior_k=50,
        )

        # THEN predictions sum to 1.0
        self.assertAlmostEqual(
            float(out["predicted_auction_share_of_pool"].sum()),
            1.0,
            places=8,
        )
