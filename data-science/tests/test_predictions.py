"""
Unit tests for moneyball.lab.predictions._map_predictions().

Tests the mapping logic that resolves team slugs to team IDs and expected
points, producing MarketPrediction objects from a raw predictions DataFrame.
"""

from __future__ import annotations

import pandas as pd
import pytest

from moneyball.lab.models import MarketPrediction
from moneyball.lab.predictions import _map_predictions


class TestThatMapPredictionsReturnsValidPredictions:
    """Valid inputs produce the expected list of MarketPrediction objects."""

    def test_that_all_slugs_map_to_predictions(self) -> None:
        """Every row in predictions_df with a matching team_id and expected points
        produces a MarketPrediction with the correct fields."""
        # GIVEN a predictions DataFrame with three teams
        predictions_df = pd.DataFrame(
            {
                "team_slug": ["duke", "unc", "kentucky"],
                "predicted_market_share": [0.10, 0.08, 0.05],
            }
        )
        team_id_map = {
            "duke": "id-duke",
            "unc": "id-unc",
            "kentucky": "id-kentucky",
        }
        expected_points_map = {
            "duke": 12.5,
            "unc": 9.0,
            "kentucky": 6.5,
        }

        # WHEN mapping predictions
        result = _map_predictions(predictions_df, team_id_map, expected_points_map)

        # THEN three MarketPrediction objects are returned with correct values
        assert len(result) == 3
        assert all(isinstance(p, MarketPrediction) for p in result)
        assert result[0].team_id == "id-duke"
        assert result[0].predicted_market_share == pytest.approx(0.10)
        assert result[0].expected_points == pytest.approx(12.5)
        assert result[1].team_id == "id-unc"
        assert result[2].team_id == "id-kentucky"


class TestThatMapPredictionsRaisesOnSlugMismatch:
    """Slugs with no team_id mapping cause a ValueError."""

    def test_that_unmapped_slugs_raise_value_error(self) -> None:
        """When some team_slugs have no entry in team_id_map, a ValueError
        is raised listing the missing slugs."""
        # GIVEN a predictions DataFrame where one slug has no team_id mapping
        predictions_df = pd.DataFrame(
            {
                "team_slug": ["duke", "unknown-team"],
                "predicted_market_share": [0.10, 0.05],
            }
        )
        team_id_map = {"duke": "id-duke"}
        expected_points_map = {"duke": 12.5, "unknown-team": 3.0}

        # WHEN mapping predictions
        # THEN a ValueError is raised mentioning the missing slug
        with pytest.raises(ValueError, match="Slug mismatch") as exc_info:
            _map_predictions(predictions_df, team_id_map, expected_points_map)

        assert "unknown-team" in str(exc_info.value)


class TestThatMapPredictionsRaisesOnMissingExpectedPoints:
    """Slugs that map to a team_id but lack expected points cause a ValueError."""

    def test_that_missing_expected_points_raises_value_error(self) -> None:
        """When a team_slug resolves to a team_id but expected_points_map has
        no entry for it, a ValueError is raised."""
        # GIVEN a predictions DataFrame where one team has no expected points
        predictions_df = pd.DataFrame(
            {
                "team_slug": ["duke", "unc"],
                "predicted_market_share": [0.10, 0.08],
            }
        )
        team_id_map = {"duke": "id-duke", "unc": "id-unc"}
        expected_points_map = {"duke": 12.5}  # missing unc

        # WHEN mapping predictions
        # THEN a ValueError is raised mentioning the team with missing points
        with pytest.raises(ValueError, match="No expected points for team unc"):
            _map_predictions(predictions_df, team_id_map, expected_points_map)


class TestThatMapPredictionsHandlesEmptyDataFrame:
    """An empty predictions DataFrame produces an empty list."""

    def test_that_empty_dataframe_returns_empty_list(self) -> None:
        """When predictions_df has no rows, _map_predictions returns an empty list
        without raising."""
        # GIVEN an empty predictions DataFrame with the correct columns
        predictions_df = pd.DataFrame(
            {
                "team_slug": pd.Series([], dtype="str"),
                "predicted_market_share": pd.Series([], dtype="float"),
            }
        )
        team_id_map = {"duke": "id-duke"}
        expected_points_map = {"duke": 12.5}

        # WHEN mapping predictions
        result = _map_predictions(predictions_df, team_id_map, expected_points_map)

        # THEN the result is an empty list
        assert result == []
