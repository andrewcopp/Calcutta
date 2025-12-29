from __future__ import annotations

import unittest

import pandas as pd

from moneyball.models.simulated_tournaments import simulate_tournaments


def _toy_games() -> pd.DataFrame:
    return pd.DataFrame({
        "game_id": ["g1", "g2", "g3"],
        "team1_key": ["t1", "t2", "t1"],
        "team2_key": ["t3", "t4", "t2"],
        "round": [1, 1, 2],
    })


def _toy_predicted_game_outcomes() -> pd.DataFrame:
    return pd.DataFrame({
        "game_id": ["g1", "g2", "g3"],
        "team1_key": ["t1", "t2", "t1"],
        "team2_key": ["t3", "t4", "t2"],
        "p_team1_wins_given_matchup": [0.6, 0.7, 0.5],
        "p_team2_wins_given_matchup": [0.4, 0.3, 0.5],
    })


class TestThatSimulatedTournamentsAreDeterministic(unittest.TestCase):
    def test_that_simulations_are_deterministic(self) -> None:
        games = _toy_games()
        predicted = _toy_predicted_game_outcomes()

        result1 = simulate_tournaments(
            games=games,
            predicted_game_outcomes=predicted,
            n_sims=100,
            seed=42,
        )

        result2 = simulate_tournaments(
            games=games,
            predicted_game_outcomes=predicted,
            n_sims=100,
            seed=42,
        )

        pd.testing.assert_frame_equal(result1, result2)


class TestThatSimulatedTournamentsHaveExpectedColumns(unittest.TestCase):
    def test_that_output_has_expected_columns(self) -> None:
        games = _toy_games()
        predicted = _toy_predicted_game_outcomes()

        result = simulate_tournaments(
            games=games,
            predicted_game_outcomes=predicted,
            n_sims=10,
            seed=123,
        )

        expected_cols = ["sim_id", "team_key", "wins"]
        self.assertEqual(sorted(result.columns.tolist()), sorted(expected_cols))


class TestThatSimulatedTournamentsProduceValidWins(unittest.TestCase):
    def test_that_wins_are_non_negative(self) -> None:
        games = _toy_games()
        predicted = _toy_predicted_game_outcomes()

        result = simulate_tournaments(
            games=games,
            predicted_game_outcomes=predicted,
            n_sims=50,
            seed=999,
        )

        self.assertTrue((result["wins"] >= 0).all())
        self.assertTrue((result["sim_id"] >= 0).all())


class TestThatSimulatedTournamentsTrackAllSims(unittest.TestCase):
    def test_that_all_simulations_are_tracked(self) -> None:
        games = _toy_games()
        predicted = _toy_predicted_game_outcomes()
        n_sims = 25

        result = simulate_tournaments(
            games=games,
            predicted_game_outcomes=predicted,
            n_sims=n_sims,
            seed=456,
        )

        unique_sims = result["sim_id"].nunique()
        self.assertEqual(unique_sims, n_sims)


class TestThatSimulatedTournamentsIncludeAllTeams(unittest.TestCase):
    def test_that_all_teams_are_included_even_with_zero_wins(self) -> None:
        games = _toy_games()
        predicted = _toy_predicted_game_outcomes()

        result = simulate_tournaments(
            games=games,
            predicted_game_outcomes=predicted,
            n_sims=50,
            seed=999,
        )

        # Get all unique teams from games
        all_teams = set()
        for _, row in games.iterrows():
            all_teams.add(row["team1_key"])
            all_teams.add(row["team2_key"])

        # Each simulation should have ALL teams
        for sim_id in range(50):
            sim_teams = set(
                result[result["sim_id"] == sim_id]["team_key"].tolist()
            )
            self.assertEqual(
                sim_teams,
                all_teams,
                f"Simulation {sim_id} missing teams",
            )


if __name__ == "__main__":
    unittest.main()
