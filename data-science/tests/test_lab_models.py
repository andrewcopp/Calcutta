"""
Unit tests for moneyball.lab.models pure functions.

Tests the JSON serialization logic extracted from create_entry_with_predictions().
"""

import json
import unittest

from moneyball.lab.models import Prediction, serialize_predictions


class TestThatSerializePredictionsProducesCorrectJson(unittest.TestCase):
    """Test serialize_predictions() produces the expected JSON structure."""

    def test_that_single_prediction_serializes_all_fields(self) -> None:
        """A single Prediction should serialize to a dict with all three fields."""
        # GIVEN a single prediction
        predictions = [
            Prediction(
                team_id="team-abc",
                predicted_market_share=0.08,
                expected_points=12.5,
            )
        ]

        # WHEN serializing
        result = serialize_predictions(predictions)

        # THEN the result has one dict with all three fields
        self.assertEqual(len(result), 1)
        self.assertEqual(result[0]["team_id"], "team-abc")
        self.assertAlmostEqual(result[0]["predicted_market_share"], 0.08)
        self.assertAlmostEqual(result[0]["expected_points"], 12.5)

    def test_that_multiple_predictions_preserve_order(self) -> None:
        """Multiple Predictions should serialize in the same order they appear."""
        # GIVEN three predictions in a specific order
        predictions = [
            Prediction(team_id="team-a", predicted_market_share=0.10, expected_points=15.0),
            Prediction(team_id="team-b", predicted_market_share=0.05, expected_points=8.0),
            Prediction(team_id="team-c", predicted_market_share=0.02, expected_points=3.0),
        ]

        # WHEN serializing
        result = serialize_predictions(predictions)

        # THEN the order is preserved
        self.assertEqual([r["team_id"] for r in result], ["team-a", "team-b", "team-c"])

    def test_that_empty_predictions_list_produces_empty_list(self) -> None:
        """An empty predictions list should serialize to an empty list."""
        # GIVEN no predictions
        predictions = []

        # WHEN serializing
        result = serialize_predictions(predictions)

        # THEN the result is an empty list
        self.assertEqual(result, [])

    def test_that_serialized_predictions_are_json_serializable(self) -> None:
        """The serialized output must be valid JSON (no dataclass objects, no numpy types)."""
        # GIVEN predictions with various numeric values
        predictions = [
            Prediction(team_id="team-1", predicted_market_share=0.123456789, expected_points=0.0),
            Prediction(team_id="team-2", predicted_market_share=0.0, expected_points=999.99),
        ]

        # WHEN serializing and converting to JSON string
        result = serialize_predictions(predictions)
        json_str = json.dumps(result)

        # THEN it round-trips cleanly through JSON
        round_tripped = json.loads(json_str)
        self.assertEqual(len(round_tripped), 2)
        self.assertEqual(round_tripped[0]["team_id"], "team-1")
        self.assertAlmostEqual(round_tripped[1]["expected_points"], 999.99)

    def test_that_serialized_dict_has_exactly_three_keys(self) -> None:
        """Each serialized prediction should have exactly team_id, predicted_market_share, expected_points."""
        # GIVEN a prediction
        predictions = [
            Prediction(team_id="team-x", predicted_market_share=0.05, expected_points=7.0),
        ]

        # WHEN serializing
        result = serialize_predictions(predictions)

        # THEN each dict has exactly the three expected keys
        expected_keys = {"team_id", "predicted_market_share", "expected_points"}
        self.assertEqual(set(result[0].keys()), expected_keys)

    def test_that_zero_market_share_is_preserved(self) -> None:
        """A prediction with zero market share should not be dropped or altered."""
        # GIVEN a prediction with zero market share
        predictions = [
            Prediction(team_id="team-16", predicted_market_share=0.0, expected_points=0.5),
        ]

        # WHEN serializing
        result = serialize_predictions(predictions)

        # THEN the zero value is preserved
        self.assertEqual(result[0]["predicted_market_share"], 0.0)


if __name__ == "__main__":
    unittest.main()
