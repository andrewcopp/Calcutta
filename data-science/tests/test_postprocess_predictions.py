"""
Unit tests for _postprocess_predictions edge cases.

Tests the pure post-processing function directly (log inverse transform,
all-zero fallback, non-finite value handling).
"""
from __future__ import annotations

import numpy as np
import pandas as pd
import pytest

from moneyball.models.predicted_market_share import _postprocess_predictions


def _make_predict_dataset(n: int = 4) -> pd.DataFrame:
    """Create a minimal predict dataset with required columns."""
    return pd.DataFrame({
        "team_key": [f"team-{i}" for i in range(n)],
        "school_name": [f"School {i}" for i in range(n)],
        "seed": list(range(1, n + 1)),
        "region": ["East"] * n,
        "kenpom_net": [10.0] * n,
    })


class TestThatPostprocessHandlesLogTransform:
    def test_that_log_transform_applies_exp_inverse(self) -> None:
        # GIVEN predictions in log space (log(0.25) for each of 4 teams)
        yhat = np.array([np.log(0.25)] * 4)
        df = _make_predict_dataset(4)

        # WHEN postprocessing with log transform
        result = _postprocess_predictions(yhat, df, target_transform="log")

        # THEN each team gets 0.25 share (exp reverses log, then normalize)
        assert float(result["predicted_market_share"].iloc[0]) == pytest.approx(0.25, abs=1e-6)

    def test_that_log_transform_clamps_extreme_values(self) -> None:
        # GIVEN extreme log predictions that would overflow without clamping
        yhat = np.array([100.0, -100.0, 0.0, 0.0])
        df = _make_predict_dataset(4)

        # WHEN postprocessing with log transform
        result = _postprocess_predictions(yhat, df, target_transform="log")

        # THEN result is finite and sums to 1.0
        assert float(result["predicted_market_share"].sum()) == pytest.approx(1.0, abs=1e-8)


class TestThatPostprocessHandlesAllZeros:
    def test_that_all_zero_predictions_fall_back_to_uniform(self) -> None:
        # GIVEN all-zero predictions
        yhat = np.array([0.0, 0.0, 0.0, 0.0])
        df = _make_predict_dataset(4)

        # WHEN postprocessing
        result = _postprocess_predictions(yhat, df, target_transform="none")

        # THEN falls back to uniform distribution (1/4 each)
        assert float(result["predicted_market_share"].iloc[0]) == pytest.approx(0.25, abs=1e-8)


class TestThatPostprocessHandlesNonFiniteValues:
    def test_that_nan_values_are_replaced_with_zero(self) -> None:
        # GIVEN predictions containing NaN
        yhat = np.array([1.0, float("nan"), 1.0, 0.0])
        df = _make_predict_dataset(4)

        # WHEN postprocessing
        result = _postprocess_predictions(yhat, df, target_transform="none")

        # THEN result is finite and sums to 1.0
        assert float(result["predicted_market_share"].sum()) == pytest.approx(1.0, abs=1e-8)

    def test_that_inf_values_are_replaced_with_zero(self) -> None:
        # GIVEN predictions containing infinity
        yhat = np.array([1.0, float("inf"), 1.0, float("-inf")])
        df = _make_predict_dataset(4)

        # WHEN postprocessing
        result = _postprocess_predictions(yhat, df, target_transform="none")

        # THEN result is finite and sums to 1.0
        assert float(result["predicted_market_share"].sum()) == pytest.approx(1.0, abs=1e-8)


class TestThatPostprocessNormalizesOutput:
    def test_that_output_sums_to_one(self) -> None:
        # GIVEN unnormalized predictions
        yhat = np.array([3.0, 1.0, 1.0, 5.0])
        df = _make_predict_dataset(4)

        # WHEN postprocessing
        result = _postprocess_predictions(yhat, df, target_transform="none")

        # THEN output sums to 1.0
        assert float(result["predicted_market_share"].sum()) == pytest.approx(1.0, abs=1e-8)

    def test_that_negative_predictions_are_clamped_to_zero(self) -> None:
        # GIVEN predictions with negative values
        yhat = np.array([2.0, -1.0, 3.0, -0.5])
        df = _make_predict_dataset(4)

        # WHEN postprocessing
        result = _postprocess_predictions(yhat, df, target_transform="none")

        # THEN all values are non-negative
        assert bool((result["predicted_market_share"] >= 0.0).all())
