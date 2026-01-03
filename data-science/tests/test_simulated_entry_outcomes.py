from __future__ import annotations

import unittest

import pandas as pd

from moneyball.models.simulated_entry_outcomes import simulate_entry_outcomes
from moneyball.utils import points

raise unittest.SkipTest(
    "Python simulated_entry_outcomes is deprecated; use Go evaluation instead."
)


def _points_by_win_index_fixture() -> dict:
    return {
        1: 0,
        2: 50,
        3: 100,
        4: 150,
        5: 200,
        6: 250,
        7: 300,
    }


class TestThatSimulatedEntryOutcomesAreDeterministic(unittest.TestCase):
    def test_that_simulated_entry_outcomes_are_deterministic(self) -> None:
        points.set_default_points_by_win_index(_points_by_win_index_fixture())
        games = pd.DataFrame(
            {
                "game_id": ["g1"],
                "round": ["round_of_64"],
                "sort_order": [1],
                "team1_key": ["t1"],
                "team2_key": ["t2"],
                "next_game_id": [""],
                "next_game_slot": [0],
            }
        )
        teams = pd.DataFrame(
            {
                "calcutta_key": ["ck", "ck"],
                "team_key": ["t1", "t2"],
                "wins": [0, 0],
                "byes": [1, 1],
            }
        )
        payouts = pd.DataFrame(
            {
                "calcutta_key": ["ck", "ck"],
                "position": [1, 2],
                "amount_cents": [10000, 0],
            }
        )
        entry_bids = pd.DataFrame(
            {
                "calcutta_key": ["ck"],
                "entry_key": ["e1"],
                "team_key": ["t2"],
                "bid_amount": [10.0],
            }
        )
        predicted_game_outcomes = pd.DataFrame(
            {
                "game_id": ["g1"],
                "team1_key": ["t1"],
                "team2_key": ["t2"],
                "p_team1_wins_given_matchup": [1.0],
                "p_team2_wins_given_matchup": [0.0],
            }
        )
        recommended_bids = pd.DataFrame(
            {
                "team_key": ["t1"],
                "bid_amount_points": [10],
            }
        )

        out_a, _ = simulate_entry_outcomes(
            games=games,
            teams=teams,
            payouts=payouts,
            entry_bids=entry_bids,
            predicted_game_outcomes=predicted_game_outcomes,
            recommended_entry_bids=recommended_bids,
            simulated_tournaments=None,
            calcutta_key="ck",
            n_sims=5,
            seed=123,
            budget_points=10,
            keep_sims=False,
        )
        out_b, _ = simulate_entry_outcomes(
            games=games,
            teams=teams,
            payouts=payouts,
            entry_bids=entry_bids,
            predicted_game_outcomes=predicted_game_outcomes,
            recommended_entry_bids=recommended_bids,
            simulated_tournaments=None,
            calcutta_key="ck",
            n_sims=5,
            seed=123,
            budget_points=10,
            keep_sims=False,
        )

        self.assertEqual(
            out_a.to_dict(orient="records"),
            out_b.to_dict(orient="records"),
        )


class TestThatSimulatedEntryOutcomesProduceExpectedPayout(unittest.TestCase):
    def test_that_simulation_produces_expected_payout(self) -> None:
        points.set_default_points_by_win_index(_points_by_win_index_fixture())
        games = pd.DataFrame(
            {
                "game_id": ["g1"],
                "round": ["round_of_64"],
                "sort_order": [1],
                "team1_key": ["t1"],
                "team2_key": ["t2"],
                "next_game_id": [""],
                "next_game_slot": [0],
            }
        )
        teams = pd.DataFrame(
            {
                "calcutta_key": ["ck", "ck"],
                "team_key": ["t1", "t2"],
                "wins": [0, 0],
                "byes": [1, 1],
            }
        )
        payouts = pd.DataFrame(
            {
                "calcutta_key": ["ck", "ck"],
                "position": [1, 2],
                "amount_cents": [10000, 0],
            }
        )
        entry_bids = pd.DataFrame(
            {
                "calcutta_key": ["ck"],
                "entry_key": ["e1"],
                "team_key": ["t2"],
                "bid_amount": [10.0],
            }
        )
        predicted_game_outcomes = pd.DataFrame(
            {
                "game_id": ["g1"],
                "team1_key": ["t1"],
                "team2_key": ["t2"],
                "p_team1_wins_given_matchup": [1.0],
                "p_team2_wins_given_matchup": [0.0],
            }
        )
        recommended_bids = pd.DataFrame(
            {
                "team_key": ["t1"],
                "bid_amount_points": [10],
            }
        )

        out, _ = simulate_entry_outcomes(
            games=games,
            teams=teams,
            payouts=payouts,
            entry_bids=entry_bids,
            predicted_game_outcomes=predicted_game_outcomes,
            recommended_entry_bids=recommended_bids,
            simulated_tournaments=None,
            calcutta_key="ck",
            n_sims=5,
            seed=123,
            budget_points=10,
            keep_sims=False,
        )

        self.assertEqual(int(out.iloc[0]["mean_payout_cents"]), 10000)
