"""Unit tests for feature engineering pure functions in predicted_auction_share_of_pool."""

import unittest

import numpy as np
import pandas as pd

from moneyball.models.predicted_auction_share_of_pool import (
    _align_columns,
    _fit_ridge,
    _predict_ridge,
    _prepare_features_set,
)


class TestThatAlignColumns(unittest.TestCase):
    """Tests for _align_columns."""

    def test_that_missing_columns_are_filled_with_zero(self) -> None:
        # GIVEN train has col_a, test has col_b
        train = pd.DataFrame({"col_a": [1.0]})
        test = pd.DataFrame({"col_b": [2.0]})

        # WHEN aligning
        train_aligned, test_aligned = _align_columns(train, test)

        # THEN both have both columns, missing filled with 0
        self.assertEqual(float(test_aligned["col_a"].iloc[0]), 0.0)

    def test_that_aligned_frames_have_matching_columns(self) -> None:
        # GIVEN train and test with different columns
        train = pd.DataFrame({"a": [1.0], "b": [2.0]})
        test = pd.DataFrame({"b": [3.0], "c": [4.0]})

        # WHEN aligning
        train_aligned, test_aligned = _align_columns(train, test)

        # THEN both have the same columns
        self.assertEqual(list(train_aligned.columns), list(test_aligned.columns))

    def test_that_existing_values_are_preserved(self) -> None:
        # GIVEN train with col_a=5.0
        train = pd.DataFrame({"col_a": [5.0]})
        test = pd.DataFrame({"col_a": [3.0], "col_b": [1.0]})

        # WHEN aligning
        train_aligned, _ = _align_columns(train, test)

        # THEN train's col_a value is preserved
        self.assertEqual(float(train_aligned["col_a"].iloc[0]), 5.0)


class TestThatFitRidge(unittest.TestCase):
    """Tests for _fit_ridge."""

    def test_that_fewer_than_five_rows_returns_none(self) -> None:
        # GIVEN only 3 rows
        X = pd.DataFrame({"x1": [1.0, 2.0, 3.0]})
        y = pd.Series([1.0, 2.0, 3.0])

        # WHEN fitting
        result = _fit_ridge(X, y, alpha=1.0)

        # THEN returns None
        self.assertIsNone(result)

    def test_that_negative_alpha_raises_value_error(self) -> None:
        # GIVEN valid data
        X = pd.DataFrame({"x1": range(10)})
        y = pd.Series(range(10), dtype=float)

        # WHEN fitting with negative alpha
        # THEN raises ValueError
        with self.assertRaises(ValueError):
            _fit_ridge(X, y, alpha=-1.0)

    def test_that_output_shape_has_intercept_plus_features(self) -> None:
        # GIVEN 10 rows with 2 features
        X = pd.DataFrame({"x1": range(10), "x2": range(10, 20)})
        y = pd.Series(range(10), dtype=float)

        # WHEN fitting
        coef = _fit_ridge(X, y, alpha=1.0)

        # THEN coef has 3 elements (intercept + 2 features)
        self.assertEqual(len(coef), 3)

    def test_that_simple_linear_relationship_is_recovered(self) -> None:
        # GIVEN y = 2*x (a simple linear relationship)
        x_vals = list(range(20))
        X = pd.DataFrame({"x": x_vals})
        y = pd.Series([2.0 * v for v in x_vals])

        # WHEN fitting with very small regularization
        coef = _fit_ridge(X, y, alpha=1e-10)

        # THEN the slope coefficient (index 1) is close to 2.0
        self.assertAlmostEqual(float(coef[1]), 2.0, places=2)


class TestThatPredictRidge(unittest.TestCase):
    """Tests for _predict_ridge."""

    def test_that_output_shape_matches_input_rows(self) -> None:
        # GIVEN 5 rows, 2 features, and a coefficient vector
        X = pd.DataFrame({"x1": range(5), "x2": range(5, 10)})
        coef = np.array([1.0, 0.5, 0.3])  # intercept + 2 features

        # WHEN predicting
        result = _predict_ridge(X, coef)

        # THEN output has 5 elements
        self.assertEqual(len(result), 5)

    def test_that_intercept_and_coefficients_are_applied(self) -> None:
        # GIVEN X with one row [x1=2, x2=3] and coef = [10, 1, 2]
        X = pd.DataFrame({"x1": [2.0], "x2": [3.0]})
        coef = np.array([10.0, 1.0, 2.0])

        # WHEN predicting
        result = _predict_ridge(X, coef)

        # THEN prediction = 10 + 1*2 + 2*3 = 18
        self.assertAlmostEqual(float(result[0]), 18.0)


class TestThatPrepareFeaturesSet(unittest.TestCase):
    """Tests for _prepare_features_set."""

    def test_that_unknown_feature_set_raises_value_error(self) -> None:
        # GIVEN valid DataFrame
        df = pd.DataFrame({
            "seed": [1], "region": ["East"], "kenpom_net": [25.0],
        })

        # WHEN preparing with unknown feature set
        # THEN raises ValueError
        with self.assertRaises(ValueError):
            _prepare_features_set(df, "nonexistent_set")

    def test_that_missing_required_columns_raises_value_error(self) -> None:
        # GIVEN DataFrame missing 'kenpom_net'
        df = pd.DataFrame({"seed": [1], "region": ["East"]})

        # WHEN preparing basic features
        # THEN raises ValueError
        with self.assertRaises(ValueError):
            _prepare_features_set(df, "basic")

    def test_that_basic_set_one_hot_encodes_region(self) -> None:
        # GIVEN DataFrame with two different regions
        df = pd.DataFrame({
            "seed": [1, 2],
            "region": ["East", "West"],
            "kenpom_net": [25.0, 20.0],
        })

        # WHEN preparing basic features
        result = _prepare_features_set(df, "basic")

        # THEN output contains one-hot encoded region columns
        region_cols = [c for c in result.columns if c.startswith("region_")]
        self.assertGreater(len(region_cols), 0)


if __name__ == "__main__":
    unittest.main()
