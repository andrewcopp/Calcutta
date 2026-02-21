"""
Unit tests for moneyball.lab.models pure functions.

Tests the JSON serialization logic extracted from create_entry_with_predictions().
"""

import json
import unittest

from moneyball.lab.models import Prediction, serialize_predictions, validate_model_params


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
        self.assertEqual(result[0]["teamId"], "team-abc")
        self.assertAlmostEqual(result[0]["predictedMarketShare"], 0.08)
        self.assertAlmostEqual(result[0]["expectedPoints"], 12.5)

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
        self.assertEqual([r["teamId"] for r in result], ["team-a", "team-b", "team-c"])

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
        self.assertEqual(round_tripped[0]["teamId"], "team-1")
        self.assertAlmostEqual(round_tripped[1]["expectedPoints"], 999.99)

    def test_that_serialized_dict_has_exactly_three_keys(self) -> None:
        """Each serialized prediction should have exactly teamId, predictedMarketShare, expectedPoints."""
        # GIVEN a prediction
        predictions = [
            Prediction(team_id="team-x", predicted_market_share=0.05, expected_points=7.0),
        ]

        # WHEN serializing
        result = serialize_predictions(predictions)

        # THEN each dict has exactly the three expected keys
        expected_keys = {"teamId", "predictedMarketShare", "expectedPoints"}
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
        self.assertEqual(result[0]["predictedMarketShare"], 0.0)


class TestThatValidateModelParamsRejectsInvalidInput(unittest.TestCase):
    """Test validate_model_params() catches typos and unknown kinds."""

    def test_that_valid_ridge_params_pass(self) -> None:
        """Known ridge params should not raise."""
        # GIVEN valid ridge params
        params = {"feature_set": "optimal_v2", "seed_prior_k": 20.0}

        # WHEN validating
        # THEN no exception is raised
        validate_model_params("ridge", params)

    def test_that_empty_params_pass_for_any_kind(self) -> None:
        """Empty params should be valid for any known kind."""
        # GIVEN empty params
        # WHEN validating for ridge
        # THEN no exception is raised
        validate_model_params("ridge", {})

    def test_that_unknown_kind_raises(self) -> None:
        """An unrecognized model kind should raise ValueError."""
        # GIVEN an unknown kind
        # WHEN validating
        # THEN ValueError is raised
        with self.assertRaises(ValueError) as ctx:
            validate_model_params("random_forest", {})
        self.assertIn("unknown model kind", str(ctx.exception))

    def test_that_typo_in_param_key_raises(self) -> None:
        """A misspelled param key should raise ValueError."""
        # GIVEN a typo in the param key
        params = {"featureset": "optimal_v2"}

        # WHEN validating
        # THEN ValueError is raised mentioning the unexpected key
        with self.assertRaises(ValueError) as ctx:
            validate_model_params("ridge", params)
        self.assertIn("featureset", str(ctx.exception))

    def test_that_extra_param_on_paramless_kind_raises(self) -> None:
        """Kinds like naive_ev that take no params should reject any keys."""
        # GIVEN params for a paramless kind
        params = {"feature_set": "optimal"}

        # WHEN validating
        # THEN ValueError is raised
        with self.assertRaises(ValueError):
            validate_model_params("naive_ev", params)


if __name__ == "__main__":
    unittest.main()
