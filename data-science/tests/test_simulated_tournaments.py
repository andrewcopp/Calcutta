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


class TestThatSimulatedTournamentsWinsAreBounded(unittest.TestCase):
    def test_that_wins_cannot_exceed_number_of_games(self) -> None:
        """GIVEN a tournament bracket
        WHEN simulations are run
        THEN each team's wins in any simulation cannot exceed total games
        
        This test prevents the bug where merging only on game_id created
        duplicate games, causing teams to accumulate impossible win counts.
        """
        games = _toy_games()
        predicted = _toy_predicted_game_outcomes()
        
        result = simulate_tournaments(
            games=games,
            predicted_game_outcomes=predicted,
            n_sims=100,
            seed=42,
        )
        
        # Maximum possible wins is the number of games in the bracket
        max_possible_wins = len(games)
        
        # Check that no team ever exceeds this
        max_wins_observed = result["wins"].max()
        self.assertLessEqual(
            max_wins_observed,
            max_possible_wins,
            f"Team won {max_wins_observed} games but only {max_possible_wins} games exist"
        )


class TestThatSimulatedTournamentsMergeCorrectly(unittest.TestCase):
    def test_that_merge_produces_one_row_per_game(self) -> None:
        """GIVEN predicted_game_outcomes with multiple matchups per game_id
        WHEN simulate_tournaments merges with games
        THEN it should produce exactly one row per game, not multiple
        
        This test prevents the bug where merging only on game_id created
        a cartesian product when predicted_game_outcomes had multiple
        possible matchups for the same game_id.
        """
        # Create games with 3 games
        games = _toy_games()
        
        # Create predicted outcomes with EXTRA rows for the same game_id
        # (simulating what happens when we have multiple possible matchups)
        predicted = pd.DataFrame({
            "game_id": ["g1", "g1", "g2", "g3"],  # g1 appears twice!
            "team1_key": ["t1", "t1", "t2", "t1"],
            "team2_key": ["t3", "t5", "t4", "t2"],  # Different opponent for g1
            "p_team1_wins_given_matchup": [0.6, 0.8, 0.7, 0.5],
            "p_team2_wins_given_matchup": [0.4, 0.2, 0.3, 0.5],
        })
        
        result = simulate_tournaments(
            games=games,
            predicted_game_outcomes=predicted,
            n_sims=10,
            seed=123,
        )
        
        # Each simulation should have exactly 4 teams (t1, t2, t3, t4)
        # If the merge is wrong, we'd see more teams or inflated win counts
        for sim_id in range(10):
            sim_data = result[result["sim_id"] == sim_id]
            total_wins = sim_data["wins"].sum()
            
            # Total wins across all teams should equal number of games
            self.assertEqual(
                total_wins,
                len(games),
                f"Simulation {sim_id} has {total_wins} total wins but should have {len(games)}"
            )


class TestThatSimulatedTournamentsProducesReasonableWinDistribution(unittest.TestCase):
    def test_that_high_probability_team_wins_more_often(self) -> None:
        """GIVEN a bracket where one team has high win probability
        WHEN many simulations are run
        THEN that team should win more games on average than low probability teams
        
        This is a sanity check that simulations respect probabilities.
        """
        # Simple bracket where t1 is heavily favored
        games = pd.DataFrame({
            "game_id": ["g1", "g2"],
            "team1_key": ["t1", "t2"],
            "team2_key": ["t3", "t4"],
            "round": [1, 1],
        })
        
        # t1 has 90% to win, t2 has 50%
        predicted = pd.DataFrame({
            "game_id": ["g1", "g2"],
            "team1_key": ["t1", "t2"],
            "team2_key": ["t3", "t4"],
            "p_team1_wins_given_matchup": [0.9, 0.5],
            "p_team2_wins_given_matchup": [0.1, 0.5],
        })
        
        result = simulate_tournaments(
            games=games,
            predicted_game_outcomes=predicted,
            n_sims=1000,
            seed=42,
        )
        
        # Calculate average wins
        t1_wins = result[result["team_key"] == "t1"]["wins"].mean()
        t2_wins = result[result["team_key"] == "t2"]["wins"].mean()
        t3_wins = result[result["team_key"] == "t3"]["wins"].mean()
        
        # t1 should win more than t2, and t2 should win more than t3
        self.assertGreater(t1_wins, t2_wins, "t1 (90% prob) should win more than t2 (50% prob)")
        self.assertGreater(t2_wins, t3_wins, "t2 (50% prob) should win more than t3 (10% prob)")


if __name__ == "__main__":
    unittest.main()
