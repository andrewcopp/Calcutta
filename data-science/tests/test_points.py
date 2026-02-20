"""Unit tests for moneyball.utils.points pure functions."""

import unittest

import pandas as pd

from moneyball.utils.points import (
    points_by_win_index_from_scoring_rules,
    team_points_from_scoring_rules,
)


class TestThatPointsByWinIndexFromScoringRules(unittest.TestCase):
    """Tests for points_by_win_index_from_scoring_rules."""

    def test_that_none_returns_empty_dict(self) -> None:
        # GIVEN None scoring_rules
        # WHEN parsing
        result = points_by_win_index_from_scoring_rules(None)

        # THEN returns empty dict
        self.assertEqual(result, {})

    def test_that_empty_dataframe_returns_empty_dict(self) -> None:
        # GIVEN empty DataFrame with expected columns
        df = pd.DataFrame(columns=["win_index", "points_awarded"])

        # WHEN parsing
        result = points_by_win_index_from_scoring_rules(df)

        # THEN returns empty dict
        self.assertEqual(result, {})

    def test_that_win_index_points_awarded_format_is_parsed(self) -> None:
        # GIVEN DataFrame with win_index/points_awarded columns
        df = pd.DataFrame({
            "win_index": [1, 2, 3],
            "points_awarded": [50, 100, 150],
        })

        # WHEN parsing
        result = points_by_win_index_from_scoring_rules(df)

        # THEN returns correct mapping
        self.assertEqual(result, {1: 50.0, 2: 100.0, 3: 150.0})

    def test_that_round_points_format_is_parsed(self) -> None:
        # GIVEN DataFrame with round/points columns
        df = pd.DataFrame({
            "round": [1, 2],
            "points": [25, 75],
        })

        # WHEN parsing
        result = points_by_win_index_from_scoring_rules(df)

        # THEN returns correct mapping
        self.assertEqual(result, {1: 25.0, 2: 75.0})

    def test_that_unknown_columns_raises_value_error(self) -> None:
        # GIVEN DataFrame with wrong columns
        df = pd.DataFrame({"foo": [1], "bar": [2]})

        # WHEN parsing
        # THEN raises ValueError
        with self.assertRaises(ValueError):
            points_by_win_index_from_scoring_rules(df)

    def test_that_string_numerics_are_coerced(self) -> None:
        # GIVEN DataFrame with string values
        df = pd.DataFrame({
            "win_index": ["1", "2"],
            "points_awarded": ["50", "100"],
        })

        # WHEN parsing
        result = points_by_win_index_from_scoring_rules(df)

        # THEN values are coerced to numeric
        self.assertEqual(result, {1: 50.0, 2: 100.0})


class TestThatTeamPointsFromScoringRules(unittest.TestCase):
    """Tests for team_points_from_scoring_rules."""

    def test_that_zero_progress_returns_zero(self) -> None:
        # GIVEN zero progress
        points_map = {1: 50.0, 2: 100.0, 3: 150.0}

        # WHEN calculating points
        result = team_points_from_scoring_rules(0, points_map)

        # THEN returns 0.0
        self.assertEqual(result, 0.0)

    def test_that_negative_progress_returns_zero(self) -> None:
        # GIVEN negative progress
        points_map = {1: 50.0, 2: 100.0}

        # WHEN calculating points
        result = team_points_from_scoring_rules(-1, points_map)

        # THEN returns 0.0
        self.assertEqual(result, 0.0)

    def test_that_points_accumulate_through_multiple_rounds(self) -> None:
        # GIVEN progress of 3 and scoring rules for 3 rounds
        points_map = {1: 50.0, 2: 100.0, 3: 150.0}

        # WHEN calculating points
        result = team_points_from_scoring_rules(3, points_map)

        # THEN returns sum of rounds 1+2+3 = 300
        self.assertEqual(result, 300.0)

    def test_that_missing_win_index_is_treated_as_zero(self) -> None:
        # GIVEN progress of 3 but only rounds 1 and 3 have points
        points_map = {1: 50.0, 3: 150.0}

        # WHEN calculating points
        result = team_points_from_scoring_rules(3, points_map)

        # THEN round 2 contributes 0, total = 50 + 0 + 150 = 200
        self.assertEqual(result, 200.0)


if __name__ == "__main__":
    unittest.main()
