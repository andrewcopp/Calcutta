from __future__ import annotations

import unittest

import pandas as pd

from moneyball.models.investment_report import generate_investment_report


def _toy_recommended_entry_bids() -> pd.DataFrame:
    return pd.DataFrame({
        "team_key": ["t1", "t2", "t3"],
        "bid_amount_points": [50, 30, 20],
        "expected_team_points": [100.0, 80.0, 60.0],
    })


def _toy_simulated_entry_outcomes() -> pd.DataFrame:
    return pd.DataFrame({
        "entry_key": ["sim_entry"],
        "mean_payout_cents": [15000.0],
        "mean_total_points": [240.0],
        "mean_finish_position": [2.5],
        "mean_n_entries": [50.0],
        "p_top1": [0.2],
        "p_top3": [0.6],
        "p_top6": [0.9],
        "p_top10": [1.0],
    })


def _toy_predicted_game_outcomes() -> pd.DataFrame:
    return pd.DataFrame({
        "game_id": ["g1"],
        "team1_key": ["t1"],
        "team2_key": ["t2"],
        "p_matchup": [1.0],
        "p_team1_wins_given_matchup": [0.7],
        "p_team2_wins_given_matchup": [0.3],
    })


def _toy_predicted_auction_share_of_pool() -> pd.DataFrame:
    return pd.DataFrame({
        "team_key": ["t1", "t2", "t3"],
        "predicted_auction_share_of_pool": [0.4, 0.35, 0.25],
    })


class TestThatInvestmentReportIncludesKeyMetrics(unittest.TestCase):
    def test_that_report_includes_expected_columns(self) -> None:
        bids = _toy_recommended_entry_bids()
        outcomes = _toy_simulated_entry_outcomes()
        game_outcomes = _toy_predicted_game_outcomes()
        share = _toy_predicted_auction_share_of_pool()

        report = generate_investment_report(
            recommended_entry_bids=bids,
            simulated_entry_outcomes=outcomes,
            predicted_game_outcomes=game_outcomes,
            predicted_auction_share_of_pool=share,
            snapshot_name="test",
            budget_points=100,
            n_sims=1000,
            seed=123,
        )

        required = [
            "snapshot_name",
            "budget_points",
            "portfolio_team_count",
            "mean_expected_payout_cents",
        ]
        self.assertTrue(set(required).issubset(set(report.columns)))


class TestThatInvestmentReportIncludesPoints(unittest.TestCase):
    def test_that_report_includes_points_metrics(self) -> None:
        bids = _toy_recommended_entry_bids()
        outcomes = _toy_simulated_entry_outcomes()
        game_outcomes = _toy_predicted_game_outcomes()
        share = _toy_predicted_auction_share_of_pool()

        report = generate_investment_report(
            recommended_entry_bids=bids,
            simulated_entry_outcomes=outcomes,
            predicted_game_outcomes=game_outcomes,
            predicted_auction_share_of_pool=share,
            snapshot_name="test",
            budget_points=100,
            n_sims=1000,
            seed=123,
        )

        self.assertIn("mean_expected_points", report.columns)
        self.assertGreater(float(report["mean_expected_points"].iloc[0]), 0)


class TestThatInvestmentReportCalculatesConcentration(unittest.TestCase):
    def test_that_report_calculates_hhi(self) -> None:
        bids = _toy_recommended_entry_bids()
        outcomes = _toy_simulated_entry_outcomes()
        game_outcomes = _toy_predicted_game_outcomes()
        share = _toy_predicted_auction_share_of_pool()

        report = generate_investment_report(
            recommended_entry_bids=bids,
            simulated_entry_outcomes=outcomes,
            predicted_game_outcomes=game_outcomes,
            predicted_auction_share_of_pool=share,
            snapshot_name="test",
            budget_points=100,
            n_sims=1000,
            seed=123,
        )

        hhi = float(report["portfolio_concentration_hhi"].iloc[0])
        self.assertGreater(hhi, 0.0)
        self.assertLessEqual(hhi, 1.0)


class TestThatInvestmentReportCountsTeams(unittest.TestCase):
    def test_that_report_counts_portfolio_teams(self) -> None:
        bids = _toy_recommended_entry_bids()
        outcomes = _toy_simulated_entry_outcomes()
        game_outcomes = _toy_predicted_game_outcomes()
        share = _toy_predicted_auction_share_of_pool()

        report = generate_investment_report(
            recommended_entry_bids=bids,
            simulated_entry_outcomes=outcomes,
            predicted_game_outcomes=game_outcomes,
            predicted_auction_share_of_pool=share,
            snapshot_name="test",
            budget_points=100,
            n_sims=1000,
            seed=123,
        )

        self.assertEqual(
            int(report["portfolio_team_count"].iloc[0]),
            3,
        )


if __name__ == "__main__":
    unittest.main()
