from __future__ import annotations

import unittest

import pandas as pd

from moneyball.pipeline.contracts import validate_artifact_df


class TestGOContractAccepts(unittest.TestCase):
    def test_that_contract_accepts_valid_frame(self) -> None:
        df = pd.DataFrame(
            {
                "game_id": ["g1"],
                "round": ["R1"],
                "round_order": [1],
                "sort_order": [1],
                "team1_key": ["t1"],
                "team1_school_name": ["Team 1"],
                "team2_key": ["t2"],
                "team2_school_name": ["Team 2"],
                "p_matchup": [1.0],
                "p_team1_wins_given_matchup": [0.6],
                "p_team2_wins_given_matchup": [0.4],
            }
        )

        validate_artifact_df(artifact_name="predicted_game_outcomes", df=df)

        self.assertTrue(True)


class TestGOContractRejects(unittest.TestCase):
    def test_that_contract_rejects_bad_rowwise_sum(
        self,
    ) -> None:
        df = pd.DataFrame(
            {
                "game_id": ["g1"],
                "round": ["R1"],
                "round_order": [1],
                "sort_order": [1],
                "team1_key": ["t1"],
                "team1_school_name": ["Team 1"],
                "team2_key": ["t2"],
                "team2_school_name": ["Team 2"],
                "p_matchup": [1.0],
                "p_team1_wins_given_matchup": [0.6],
                "p_team2_wins_given_matchup": [0.6],
            }
        )

        with self.assertRaises(ValueError):
            validate_artifact_df(
                artifact_name="predicted_game_outcomes",
                df=df,
            )


class TestThatAuctionShareContractRejectsNonNormalized(unittest.TestCase):
    def test_that_contract_rejects_non_normalized(
        self,
    ) -> None:
        df = pd.DataFrame({"predicted_auction_share_of_pool": [0.2, 0.2]})

        with self.assertRaises(ValueError):
            validate_artifact_df(
                artifact_name="predicted_auction_share_of_pool",
                df=df,
            )


class TestThatEntryBidsContractAcceptsValid(unittest.TestCase):
    def test_that_contract_accepts_valid_frame(self) -> None:
        df = pd.DataFrame(
            {
                "team_key": ["t1", "t2"],
                "bid_amount_points": [7, 3],
            }
        )

        validate_artifact_df(artifact_name="recommended_entry_bids", df=df)

        self.assertTrue(True)


class TestThatEntryBidsContractRejectsDupTeamKey(unittest.TestCase):
    def test_that_contract_rejects_duplicate_team_key(self) -> None:
        df = pd.DataFrame(
            {
                "team_key": ["t1", "t1"],
                "bid_amount_points": [5, 5],
            }
        )

        with self.assertRaises(ValueError):
            validate_artifact_df(
                artifact_name="recommended_entry_bids",
                df=df,
            )


