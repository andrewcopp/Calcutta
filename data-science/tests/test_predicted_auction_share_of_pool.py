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
