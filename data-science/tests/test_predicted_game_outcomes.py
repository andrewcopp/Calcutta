from __future__ import annotations

import unittest

import pandas as pd

from moneyball.models.predicted_game_outcomes import predict_game_outcomes


def _toy_games_and_teams() -> tuple[pd.DataFrame, pd.DataFrame]:
    teams = pd.DataFrame(
        {
            "team_key": [
                "A",
                "B",
                "C",
                "D",
                "E",
                "F",
                "G",
                "H",
            ],
            "school_name": [
                "A",
                "B",
                "C",
                "D",
                "E",
                "F",
                "G",
                "H",
            ],
            "kenpom_net": [10, 9, 8, 7, 6, 5, 4, 3],
        }
    )

    games = pd.DataFrame(
        {
            "game_id": [
                "G1",
                "G2",
                "G3",
                "G4",
                "G5",
                "G6",
                "G7",
            ],
            "round": [
                "round_of_64",
                "round_of_64",
                "round_of_64",
                "round_of_64",
                "sweet_16",
                "sweet_16",
                "elite_8",
            ],
            "sort_order": [1, 2, 3, 4, 5, 6, 7],
            "team1_key": ["A", "C", "E", "G", "", "", ""],
            "team2_key": ["B", "D", "F", "H", "", "", ""],
            "next_game_id": [
                "G5",
                "G5",
                "G6",
                "G6",
                "G7",
                "G7",
                "",
            ],
            "next_game_slot": [1, 2, 1, 2, 1, 2, 0],
        }
    )

    return games, teams


class TestThatPredictedGameOutcomesIsWellFormed(unittest.TestCase):
    def test_that_p_matchup_sums_to_1_for_each_game(self) -> None:
        games, teams = _toy_games_and_teams()
        df = predict_game_outcomes(
            games=games,
            teams=teams,
            calcutta_key=None,
            kenpom_scale=10.0,
            n_sims=500,
            seed=123,
        )
        sums = (
            df.groupby("game_id")["p_matchup"].sum().round(8).to_dict()
        )
        self.assertEqual(set(sums.values()), {1.0})


class TestThatPredictedGameOutcomesHasExpectedColumns(unittest.TestCase):
    def test_that_output_includes_probability_fields(self) -> None:
        games, teams = _toy_games_and_teams()
        df = predict_game_outcomes(
            games=games,
            teams=teams,
            calcutta_key=None,
            kenpom_scale=10.0,
            n_sims=10,
            seed=1,
        )
        self.assertTrue(
            {
                "p_matchup",
                "p_team1_wins_given_matchup",
                "p_team2_wins_given_matchup",
            }.issubset(set(df.columns))
        )
