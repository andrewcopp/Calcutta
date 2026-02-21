from __future__ import annotations

import pandas as pd
import pytest

from moneyball.models.predicted_market_share import (
    predict_market_share,
)
from tests.conftest import SEED_MARKET_SHARES, create_68_team_field


class TestThatPredictedMarketShareIsNormalized:
    def test_that_predicted_market_share_sums_to_one(
        self, basic_train_dataset: pd.DataFrame, basic_predict_dataset: pd.DataFrame
    ) -> None:
        out = predict_market_share(
            train_team_dataset=basic_train_dataset,
            predict_team_dataset=basic_predict_dataset,
            ridge_alpha=1.0,
            feature_set="basic",
        )

        assert float(out["predicted_market_share"].sum()) == pytest.approx(
            1.0, abs=1e-8
        )


class TestThatPredictedMarketShareIsNonNegative:
    def test_that_predicted_share_of_pool_is_non_negative(
        self, basic_train_dataset: pd.DataFrame, basic_predict_dataset: pd.DataFrame
    ) -> None:
        out = predict_market_share(
            train_team_dataset=basic_train_dataset,
            predict_team_dataset=basic_predict_dataset,
            ridge_alpha=1.0,
            feature_set="basic",
        )

        assert bool((out["predicted_market_share"] >= 0.0).all())


class TestThatPredictedMarketShareIncludesKeyColumns:
    def test_that_output_includes_predicted_share_of_pool(
        self, basic_train_dataset: pd.DataFrame, basic_predict_dataset: pd.DataFrame
    ) -> None:
        out = predict_market_share(
            train_team_dataset=basic_train_dataset,
            predict_team_dataset=basic_predict_dataset,
            ridge_alpha=1.0,
            feature_set="basic",
        )

        assert "predicted_market_share" in out.columns


class TestThatOptimalV2RequiresSeedPriorK:
    """Test that optimal_v2 feature set requires seed_prior_k > 0."""

    def test_that_optimal_v2_with_zero_seed_prior_k_raises_error(
        self,
        extended_train_dataset: pd.DataFrame,
        extended_predict_dataset: pd.DataFrame,
    ) -> None:
        # WHEN/THEN using optimal_v2 with seed_prior_k=0 raises ValueError
        with pytest.raises(ValueError, match="seed_prior_k > 0") as exc_info:
            predict_market_share(
                train_team_dataset=extended_train_dataset,
                predict_team_dataset=extended_predict_dataset,
                ridge_alpha=1.0,
                feature_set="optimal_v2",
                seed_prior_k=0,
            )

        assert "optimal_v2" in str(exc_info.value)

    def test_that_optimal_v2_with_positive_seed_prior_k_works(
        self,
        extended_train_dataset: pd.DataFrame,
        extended_predict_dataset: pd.DataFrame,
    ) -> None:
        # WHEN using optimal_v2 with seed_prior_k=20 (recommended value)
        out = predict_market_share(
            train_team_dataset=extended_train_dataset,
            predict_team_dataset=extended_predict_dataset,
            ridge_alpha=1.0,
            feature_set="optimal_v2",
            seed_prior_k=20,
            program_prior_k=50,
        )

        # THEN predictions sum to 1.0
        assert float(out["predicted_market_share"].sum()) == pytest.approx(
            1.0, abs=1e-8
        )


class TestThatOptimalV3PredictionsSumToOne:
    """optimal_v3 predictions should sum to 1.0."""

    def test_that_optimal_v3_predictions_sum_to_one(self) -> None:
        # GIVEN 68-team training and prediction datasets
        train = create_68_team_field()
        train["observed_team_share_of_pool"] = 1.0 / len(train)

        pred = create_68_team_field()

        # WHEN using optimal_v3 feature set
        out = predict_market_share(
            train_team_dataset=train,
            predict_team_dataset=pred,
            ridge_alpha=1.0,
            feature_set="optimal_v3",
        )

        # THEN predictions sum to 1.0
        assert float(out["predicted_market_share"].sum()) == pytest.approx(
            1.0, abs=1e-8
        )


class TestThatOptimalV3PredictionsAreNonNegative:
    """optimal_v3 predictions should be non-negative."""

    def test_that_optimal_v3_predictions_are_non_negative(self) -> None:
        # GIVEN 68-team training and prediction datasets
        train = create_68_team_field()
        train["observed_team_share_of_pool"] = 1.0 / len(train)

        pred = create_68_team_field()

        # WHEN using optimal_v3 feature set
        out = predict_market_share(
            train_team_dataset=train,
            predict_team_dataset=pred,
            ridge_alpha=1.0,
            feature_set="optimal_v3",
        )

        # THEN all predictions are non-negative
        assert bool((out["predicted_market_share"] >= 0.0).all())


class TestThatOptimalV3WorksWithRealisticMarketShares:
    """optimal_v3 should work with realistic seed-based market shares."""

    def test_that_optimal_v3_produces_seed_correlated_predictions(self) -> None:
        # GIVEN 68-team training with seed-based market shares
        train = create_68_team_field()
        train["observed_team_share_of_pool"] = train["seed"].map(SEED_MARKET_SHARES)
        train["observed_team_share_of_pool"] = (
            train["observed_team_share_of_pool"] / train["observed_team_share_of_pool"].sum()
        )

        pred = create_68_team_field()

        # WHEN using optimal_v3 feature set
        out = predict_market_share(
            train_team_dataset=train,
            predict_team_dataset=pred,
            ridge_alpha=1.0,
            feature_set="optimal_v3",
        )

        # THEN 1-seeds should have higher predicted share than 16-seeds
        if "seed" in out.columns:
            out_with_seed = out
        else:
            out_with_seed = out.merge(
                pred[["team_key", "seed"]], on="team_key", how="left"
            )

        seed_col = "seed" if "seed" in out_with_seed.columns else "seed_y"
        avg_by_seed = out_with_seed.groupby(seed_col)[
            "predicted_market_share"
        ].mean()

        assert avg_by_seed[1] > avg_by_seed[16]


class TestThatOptimalV3WorksWithMultiYearTrainingData:
    """optimal_v3 must process each year separately to avoid broken analytics."""

    def test_that_optimal_v3_produces_non_uniform_predictions_with_multi_year_data(
        self,
    ) -> None:
        # GIVEN multi-year training data (simulating concatenated snapshots)
        train_2023 = create_68_team_field()
        train_2023["snapshot"] = "2023"
        train_2024 = create_68_team_field()
        train_2024["snapshot"] = "2024"

        train = pd.concat([train_2023, train_2024], ignore_index=True)

        train["observed_team_share_of_pool"] = train["seed"].map(SEED_MARKET_SHARES)
        train["observed_team_share_of_pool"] = (
            train["observed_team_share_of_pool"] / train["observed_team_share_of_pool"].sum()
        )

        pred = create_68_team_field()
        pred["snapshot"] = "2025"

        # WHEN using optimal_v3 with multi-year training data
        out = predict_market_share(
            train_team_dataset=train,
            predict_team_dataset=pred,
            ridge_alpha=1.0,
            feature_set="optimal_v3",
        )

        # THEN predictions should NOT be uniform (the original bug produced 1/68)
        uniform_share = 1.0 / 68
        predictions = out["predicted_market_share"].values

        max_deviation = max(abs(p - uniform_share) for p in predictions)
        assert max_deviation > 0.005, (
            "Predictions are nearly uniform - analytical enrichment likely failed"
        )

        # THEN predictions should still sum to 1.0
        assert float(out["predicted_market_share"].sum()) == pytest.approx(
            1.0, abs=1e-8
        )
