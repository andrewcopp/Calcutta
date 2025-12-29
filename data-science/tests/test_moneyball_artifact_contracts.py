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


class TestThatSimOutcomesContractAcceptsValid(unittest.TestCase):
    def test_that_contract_accepts_valid_frame(self) -> None:
        df = pd.DataFrame(
            {
                "entry_key": ["e1"],
                "sims": [100],
                "seed": [123],
                "budget_points": [100],
                "mean_payout_cents": [5000.0],
                "mean_total_points": [50.0],
                "mean_finish_position": [3.5],
                "mean_n_entries": [50.0],
                "p_top1": [0.1],
                "p_top3": [0.3],
                "p_top6": [0.6],
                "p_top10": [0.9],
            }
        )

        validate_artifact_df(
            artifact_name="simulated_entry_outcomes",
            df=df,
        )

        self.assertTrue(True)


class TestThatSimOutcomesContractRejectsNegativeSims(unittest.TestCase):
    def test_that_contract_rejects_negative_sims(self) -> None:
        df = pd.DataFrame(
            {
                "entry_key": ["e1"],
                "sims": [-1],
                "seed": [123],
                "budget_points": [100],
                "mean_payout_cents": [5000.0],
                "mean_total_points": [50.0],
                "mean_finish_position": [3.5],
                "mean_n_entries": [50.0],
                "p_top1": [0.1],
                "p_top3": [0.3],
                "p_top6": [0.6],
                "p_top10": [0.9],
            }
        )

        with self.assertRaises(ValueError):
            validate_artifact_df(
                artifact_name="simulated_entry_outcomes",
                df=df,
            )


class TestThatSimOutcomesContractRejectsInvalidProb(unittest.TestCase):
    def test_that_contract_rejects_invalid_probability(self) -> None:
        df = pd.DataFrame(
            {
                "entry_key": ["e1"],
                "sims": [100],
                "seed": [123],
                "budget_points": [100],
                "mean_payout_cents": [5000.0],
                "mean_total_points": [50.0],
                "mean_finish_position": [3.5],
                "p_top1": [1.5],
                "p_top3": [0.3],
                "p_top6": [0.6],
                "p_top10": [0.9],
            }
        )

        with self.assertRaises(ValueError):
            validate_artifact_df(
                artifact_name="simulated_entry_outcomes",
                df=df,
            )


class TestThatInvestmentReportContractAcceptsValid(unittest.TestCase):
    def test_that_contract_accepts_valid_frame(self) -> None:
        df = pd.DataFrame({
            "snapshot_name": ["test"],
            "budget_points": [100],
            "n_sims": [1000],
            "seed": [123],
            "portfolio_team_count": [5],
            "portfolio_total_bids": [100],
            "mean_expected_payout_cents": [12000.0],
            "mean_expected_points": [150.0],
            "mean_expected_finish_position": [3.0],
            "mean_n_entries": [50.0],
            "p_top1": [0.15],
            "p_top3": [0.5],
            "p_top6": [0.8],
            "p_top10": [0.95],
            "portfolio_concentration_hhi": [0.35],
            "portfolio_teams_json": ["[]"],
        })

        validate_artifact_df(
            artifact_name="investment_report",
            df=df,
        )

        self.assertTrue(True)


class TestThatInvestmentReportContractRejectsInvalidHHI(
    unittest.TestCase
):
    def test_that_contract_rejects_invalid_hhi(self) -> None:
        df = pd.DataFrame({
            "snapshot_name": ["test"],
            "budget_points": [100],
            "n_sims": [1000],
            "seed": [123],
            "portfolio_team_count": [5],
            "portfolio_total_bids": [100],
            "mean_expected_payout_cents": [12000.0],
            "mean_expected_points": [150.0],
            "mean_expected_finish_position": [3.0],
            "mean_n_entries": [50.0],
            "p_top1": [0.15],
            "p_top3": [0.5],
            "p_top6": [0.8],
            "p_top10": [0.95],
            "portfolio_concentration_hhi": [1.5],
            "portfolio_teams_json": ["[]"],
        })

        with self.assertRaises(ValueError):
            validate_artifact_df(
                artifact_name="investment_report",
                df=df,
            )
