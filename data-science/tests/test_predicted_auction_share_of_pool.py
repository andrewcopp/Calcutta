from __future__ import annotations

import unittest

import pandas as pd

from moneyball.models.predicted_auction_share_of_pool import (
    predict_auction_share_of_pool,
)
from moneyball.utils.points import set_default_points_by_win_index

# Standard March Madness scoring: 1, 2, 4, 8, 16, 32 per round
_DEFAULT_POINTS = {
    1: 1,   # Round of 64 win
    2: 2,   # Round of 32 win
    3: 4,   # Sweet 16 win
    4: 8,   # Elite 8 win
    5: 16,  # Final Four win
    6: 32,  # Championship win
    7: 0,   # No additional points for winning championship game twice
}


def setUpModule():
    """Set up default scoring rules before any tests run."""
    set_default_points_by_win_index(_DEFAULT_POINTS)


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


def _create_68_team_field_for_auction() -> pd.DataFrame:
    """Create a 68-team field with all required columns for auction prediction."""
    teams = []
    regions = ["East", "West", "South", "Midwest"]

    for region in regions:
        for seed in range(1, 17):
            kenpom_net = 25 - (seed - 1) * 2.2
            teams.append({
                "id": f"{region}-{seed}",
                "team_key": f"{region}-{seed}",
                "seed": seed,
                "region": region,
                "kenpom_net": kenpom_net,
                "kenpom_o": 110 + kenpom_net * 0.3,
                "kenpom_d": 100 - kenpom_net * 0.2,
                "kenpom_adj_t": 68 + seed * 0.1,
                "school_slug": f"school-{region.lower()}-{seed}",
            })

    # First Four teams
    first_four = [
        ("East", 16, -12.0),
        ("West", 16, -11.0),
        ("South", 11, 5.0),
        ("Midwest", 11, 4.0),
    ]
    for region, seed, kenpom in first_four:
        teams.append({
            "id": f"{region}-{seed}-FF",
            "team_key": f"{region}-{seed}-FF",
            "seed": seed,
            "region": region,
            "kenpom_net": kenpom,
            "kenpom_o": 110 + kenpom * 0.3,
            "kenpom_d": 100 - kenpom * 0.2,
            "kenpom_adj_t": 68 + seed * 0.1,
            "school_slug": f"school-{region.lower()}-{seed}-ff",
        })

    return pd.DataFrame(teams)


class TestThatOptimalV3PredictionsSumToOne(unittest.TestCase):
    """optimal_v3 predictions should sum to 1.0."""

    def test_that_optimal_v3_predictions_sum_to_one(self) -> None:
        # GIVEN 68-team training and prediction datasets
        train = _create_68_team_field_for_auction()
        # Assign fake market shares that sum to 1.0
        train["team_share_of_pool"] = 1.0 / len(train)

        pred = _create_68_team_field_for_auction()

        # WHEN using optimal_v3 feature set
        out = predict_auction_share_of_pool(
            train_team_dataset=train,
            predict_team_dataset=pred,
            ridge_alpha=1.0,
            feature_set="optimal_v3",
            kenpom_scale=10.0,
        )

        # THEN predictions sum to 1.0
        self.assertAlmostEqual(
            float(out["predicted_auction_share_of_pool"].sum()),
            1.0,
            places=8,
        )


class TestThatOptimalV3PredictionsAreNonNegative(unittest.TestCase):
    """optimal_v3 predictions should be non-negative."""

    def test_that_optimal_v3_predictions_are_non_negative(self) -> None:
        # GIVEN 68-team training and prediction datasets
        train = _create_68_team_field_for_auction()
        train["team_share_of_pool"] = 1.0 / len(train)

        pred = _create_68_team_field_for_auction()

        # WHEN using optimal_v3 feature set
        out = predict_auction_share_of_pool(
            train_team_dataset=train,
            predict_team_dataset=pred,
            ridge_alpha=1.0,
            feature_set="optimal_v3",
            kenpom_scale=10.0,
        )

        # THEN all predictions are non-negative
        self.assertTrue(
            bool((out["predicted_auction_share_of_pool"] >= 0.0).all())
        )


class TestThatOptimalV3WorksWithRealisticMarketShares(unittest.TestCase):
    """optimal_v3 should work with realistic seed-based market shares."""

    def test_that_optimal_v3_produces_seed_correlated_predictions(self) -> None:
        # GIVEN 68-team training with seed-based market shares
        train = _create_68_team_field_for_auction()
        # Higher seeds get larger market share
        seed_shares = {
            1: 0.08, 2: 0.06, 3: 0.05, 4: 0.04, 5: 0.03, 6: 0.025,
            7: 0.02, 8: 0.018, 9: 0.015, 10: 0.012, 11: 0.01,
            12: 0.008, 13: 0.005, 14: 0.003, 15: 0.002, 16: 0.001,
        }
        train["team_share_of_pool"] = train["seed"].map(seed_shares)
        # Normalize to sum to 1.0
        train["team_share_of_pool"] = (
            train["team_share_of_pool"] / train["team_share_of_pool"].sum()
        )

        pred = _create_68_team_field_for_auction()

        # WHEN using optimal_v3 feature set
        out = predict_auction_share_of_pool(
            train_team_dataset=train,
            predict_team_dataset=pred,
            ridge_alpha=1.0,
            feature_set="optimal_v3",
            kenpom_scale=10.0,
        )

        # THEN 1-seeds should have higher predicted share than 16-seeds
        # The output already contains seed from pred dataset, but we need to
        # handle case where seed column might already exist with suffix
        if "seed" in out.columns:
            out_with_seed = out
        else:
            out_with_seed = out.merge(
                pred[["team_key", "seed"]], on="team_key", how="left"
            )

        # Handle potential column suffix from merge
        seed_col = "seed" if "seed" in out_with_seed.columns else "seed_y"
        avg_by_seed = out_with_seed.groupby(seed_col)[
            "predicted_auction_share_of_pool"
        ].mean()

        self.assertGreater(avg_by_seed[1], avg_by_seed[16])
