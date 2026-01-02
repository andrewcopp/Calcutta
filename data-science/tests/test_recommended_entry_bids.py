from __future__ import annotations

import unittest

import pandas as pd

from moneyball.models.recommended_entry_bids import recommend_entry_bids
from moneyball.utils import points


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


def _toy_predicted_game_outcomes() -> pd.DataFrame:
    return pd.DataFrame(
        {
            "game_id": ["g1", "g2", "g3"],
            "round": ["round_of_64", "round_of_64", "round_of_32"],
            "round_order": [2, 2, 3],
            "sort_order": [1, 2, 3],
            "team1_key": ["t1", "t3", "t1"],
            "team1_school_name": ["T1", "T3", "T1"],
            "team2_key": ["t2", "t4", "t3"],
            "team2_school_name": ["T2", "T4", "T3"],
            "p_matchup": [1.0, 1.0, 1.0],
            "p_team1_wins_given_matchup": [0.9, 0.6, 0.7],
            "p_team2_wins_given_matchup": [0.1, 0.4, 0.3],
        }
    )


def _toy_predicted_auction_share_of_pool() -> pd.DataFrame:
    return pd.DataFrame(
        {
            "team_key": ["t1", "t2", "t3", "t4"],
            "predicted_auction_share_of_pool": [0.4, 0.2, 0.3, 0.1],
        }
    )


class TestThatEntryBidsSumToBudget(unittest.TestCase):
    def test_that_recommended_entry_bids_sum_to_budget(self) -> None:
        points.set_default_points_by_win_index(_points_by_win_index_fixture())
        share_df = _toy_predicted_auction_share_of_pool()
        go_df = _toy_predicted_game_outcomes()
        out = recommend_entry_bids(
            predicted_auction_share_of_pool=share_df,
            predicted_game_outcomes=go_df,
            predicted_total_pool_bids_points=40.0,
            budget_points=10,
            min_teams=2,
            max_teams=3,
            max_per_team_points=7,
            min_bid_points=1,
        )

        self.assertEqual(
            int(out["bid_amount_points"].sum()),
            10,
        )


class TestThatEntryBidsRespectTeamCount(unittest.TestCase):
    def test_that_recommended_entry_bids_respect_team_count_constraints(
        self,
    ) -> None:
        points.set_default_points_by_win_index(_points_by_win_index_fixture())
        share_df = _toy_predicted_auction_share_of_pool()
        go_df = _toy_predicted_game_outcomes()
        out = recommend_entry_bids(
            predicted_auction_share_of_pool=share_df,
            predicted_game_outcomes=go_df,
            predicted_total_pool_bids_points=40.0,
            budget_points=10,
            min_teams=2,
            max_teams=3,
            max_per_team_points=7,
            min_bid_points=1,
        )

        self.assertTrue(2 <= int(len(out)) <= 3)


class TestThatEntryBidsRespectMaxPerTeam(unittest.TestCase):
    def test_that_recommended_entry_bids_respect_max_per_team(self) -> None:
        points.set_default_points_by_win_index(_points_by_win_index_fixture())
        share_df = _toy_predicted_auction_share_of_pool()
        go_df = _toy_predicted_game_outcomes()
        out = recommend_entry_bids(
            predicted_auction_share_of_pool=share_df,
            predicted_game_outcomes=go_df,
            predicted_total_pool_bids_points=40.0,
            budget_points=10,
            min_teams=2,
            max_teams=3,
            max_per_team_points=7,
            min_bid_points=1,
        )

        self.assertTrue(bool((out["bid_amount_points"] <= 7).all()))


class TestThatEntryBidsAreDeterministic(unittest.TestCase):
    def test_that_recommended_entry_bids_are_deterministic(self) -> None:
        points.set_default_points_by_win_index(_points_by_win_index_fixture())
        share_df = _toy_predicted_auction_share_of_pool()
        go_df = _toy_predicted_game_outcomes()
        out_a = recommend_entry_bids(
            predicted_auction_share_of_pool=share_df,
            predicted_game_outcomes=go_df,
            predicted_total_pool_bids_points=40.0,
            budget_points=10,
            min_teams=2,
            max_teams=3,
            max_per_team_points=7,
            min_bid_points=1,
        )
        out_b = recommend_entry_bids(
            predicted_auction_share_of_pool=share_df,
            predicted_game_outcomes=go_df,
            predicted_total_pool_bids_points=40.0,
            budget_points=10,
            min_teams=2,
            max_teams=3,
            max_per_team_points=7,
            min_bid_points=1,
        )

        self.assertEqual(
            out_a.to_dict(orient="records"),
            out_b.to_dict(orient="records"),
        )
