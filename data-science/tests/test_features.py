"""Unit tests for moneyball.models.features helpers."""

import unittest

import numpy as np
import pandas as pd

from moneyball.models.features import (
    BLUE_BLOODS,
    SEED_EXPECTED_POINTS,
    SEED_TITLE_PROBABILITY,
    compute_kenpom_balance_percentile,
    compute_kenpom_balance_zscore,
    compute_kenpom_net_zscore,
    compute_market_behavior_features,
    compute_seed_interactions,
    prepare_optimal_v1_features,
    prepare_optimal_v2_features,
    prepare_optimal_v3_features,
)


class TestThatKenpomNetZscoreIsCorrect(unittest.TestCase):
    """Tests for compute_kenpom_net_zscore."""

    def test_that_zscore_has_zero_mean(self) -> None:
        # GIVEN a series of KenPom net ratings
        s = pd.Series([30.0, 20.0, 10.0, 0.0, -10.0])

        # WHEN computing z-scores
        z = compute_kenpom_net_zscore(s)

        # THEN the mean is approximately zero
        self.assertAlmostEqual(float(z.mean()), 0.0, places=10)

    def test_that_zscore_has_unit_std(self) -> None:
        # GIVEN a series of KenPom net ratings
        s = pd.Series([30.0, 20.0, 10.0, 0.0, -10.0])

        # WHEN computing z-scores
        z = compute_kenpom_net_zscore(s)

        # THEN the standard deviation is approximately one
        self.assertAlmostEqual(float(z.std()), 1.0, places=10)

    def test_that_constant_input_returns_zeros(self) -> None:
        # GIVEN a constant series (std=0)
        s = pd.Series([5.0, 5.0, 5.0])

        # WHEN computing z-scores
        z = compute_kenpom_net_zscore(s)

        # THEN all values are 0.0
        self.assertTrue((z == 0.0).all())


class TestThatKenpomBalanceZscoreIsNonNegative(unittest.TestCase):
    """Tests for compute_kenpom_balance_zscore."""

    def test_that_balance_is_non_negative(self) -> None:
        # GIVEN offensive and defensive ratings
        ko = pd.Series([120.0, 110.0, 105.0])
        kd = pd.Series([95.0, 100.0, 110.0])

        # WHEN computing balance
        balance = compute_kenpom_balance_zscore(ko, kd)

        # THEN all values are non-negative
        self.assertTrue((balance >= 0.0).all())


class TestThatKenpomBalancePercentileIsNonNegative(unittest.TestCase):
    """Tests for compute_kenpom_balance_percentile."""

    def test_that_balance_is_non_negative(self) -> None:
        # GIVEN offensive and defensive ratings
        ko = pd.Series([120.0, 110.0, 105.0])
        kd = pd.Series([95.0, 100.0, 110.0])

        # WHEN computing balance
        balance = compute_kenpom_balance_percentile(ko, kd)

        # THEN all values are non-negative
        self.assertTrue((balance >= 0.0).all())


class TestThatSeedInteractionsHaveExpectedColumns(unittest.TestCase):
    """Tests for compute_seed_interactions."""

    def test_that_output_has_seed_sq_and_kenpom_x_seed(self) -> None:
        # GIVEN seed and kenpom series
        seed = pd.Series([1, 5, 16])
        kenpom = pd.Series([30.0, 10.0, -5.0])

        # WHEN computing interactions
        result = compute_seed_interactions(seed, kenpom)

        # THEN output has the expected columns
        self.assertIn("seed_sq", result.columns)
        self.assertIn("kenpom_x_seed", result.columns)

    def test_that_seed_sq_is_correct(self) -> None:
        # GIVEN seed=4
        seed = pd.Series([4])
        kenpom = pd.Series([10.0])

        # WHEN computing interactions
        result = compute_seed_interactions(seed, kenpom)

        # THEN seed_sq = 16
        self.assertEqual(float(result["seed_sq"].iloc[0]), 16.0)

    def test_that_kenpom_x_seed_is_correct(self) -> None:
        # GIVEN seed=3, kenpom_net=10.0
        seed = pd.Series([3])
        kenpom = pd.Series([10.0])

        # WHEN computing interactions
        result = compute_seed_interactions(seed, kenpom)

        # THEN kenpom_x_seed = 30.0
        self.assertEqual(float(result["kenpom_x_seed"].iloc[0]), 30.0)


class TestThatMarketBehaviorFeaturesAreCorrect(unittest.TestCase):
    """Tests for compute_market_behavior_features."""

    def test_that_upset_seed_flag_is_set_for_seeds_10_to_12(self) -> None:
        # GIVEN a dataframe with seeds including 10, 11, 12
        df = pd.DataFrame({
            "seed": [1, 10, 11, 12, 16],
            "kenpom_net": [30.0, 5.0, 4.0, 3.0, -10.0],
        })
        base = df[["seed"]].copy()

        # WHEN computing market behavior features
        result = compute_market_behavior_features(df, base)

        # THEN seeds 10-12 have is_upset_seed=1, others 0
        self.assertEqual(result["is_upset_seed"].tolist(), [0, 1, 1, 1, 0])

    def test_that_kenpom_rank_within_seed_norm_is_between_0_and_1(self) -> None:
        # GIVEN multiple teams per seed
        df = pd.DataFrame({
            "seed": [1, 1, 1, 1],
            "kenpom_net": [30.0, 28.0, 25.0, 20.0],
        })
        base = df[["seed"]].copy()

        # WHEN computing market behavior features
        result = compute_market_behavior_features(df, base)

        # THEN norm values are in [0, 1]
        vals = result["kenpom_rank_within_seed_norm"]
        self.assertTrue((vals >= 0.0).all())
        self.assertTrue((vals <= 1.0).all())


class TestThatConstantsAreComplete(unittest.TestCase):
    """Tests for module-level constants."""

    def test_that_seed_title_probability_has_all_16_seeds(self) -> None:
        self.assertEqual(set(SEED_TITLE_PROBABILITY.keys()), set(range(1, 17)))

    def test_that_seed_expected_points_has_all_16_seeds(self) -> None:
        self.assertEqual(set(SEED_EXPECTED_POINTS.keys()), set(range(1, 17)))

    def test_that_seed_title_probabilities_decrease_with_seed(self) -> None:
        # GIVEN the seed title probability map
        # THEN probabilities are monotonically non-increasing
        for s in range(1, 16):
            self.assertGreaterEqual(
                SEED_TITLE_PROBABILITY[s],
                SEED_TITLE_PROBABILITY[s + 1],
            )

    def test_that_blue_bloods_contains_known_programs(self) -> None:
        self.assertIn("duke", BLUE_BLOODS)
        self.assertIn("kansas", BLUE_BLOODS)
        self.assertIn("kentucky", BLUE_BLOODS)


class TestThatPrepareOptimalV3RaisesWithoutAnalyticalColumns(unittest.TestCase):
    """Tests for prepare_optimal_v3_features validation."""

    def test_that_missing_predicted_p_championship_raises(self) -> None:
        # GIVEN a dataframe without predicted_p_championship
        df = pd.DataFrame({
            "seed": [1],
            "kenpom_net": [25.0],
            "kenpom_o": [120.0],
            "kenpom_d": [95.0],
        })
        base = df[["seed", "kenpom_net", "kenpom_o", "kenpom_d"]].copy()

        # WHEN/THEN calling prepare_optimal_v3_features raises ValueError
        with self.assertRaises(ValueError) as ctx:
            prepare_optimal_v3_features(df, base)

        self.assertIn("predicted_p_championship", str(ctx.exception))


if __name__ == "__main__":
    unittest.main()
