from __future__ import annotations

import unittest

import numpy as np

from calcutta_ds.investment_report.portfolio_contest import (
    contest_objective_from_sim_bids,
)


class TestThatTieSplittingIsCorrect(unittest.TestCase):
    def test_that_expected_payout_splits_tie_group_evenly(self) -> None:
        # GIVEN a single simulation where the simulated entry ties for 1st
        # with exactly one competitor (so the tie group covers positions 1-2).
        team_points_scenarios = np.zeros((1, 1), dtype=float)
        market_entry_bids = np.zeros((1, 1), dtype=float)
        market_team_totals = np.ones((1,), dtype=float)
        sim_team_bids = np.ones((1,), dtype=float)

        # Force a tie: competitor_points == sim_points in the only scenario.
        # With one competitor: start_pos = 1 + gt = 1, group_size = 1 + eq = 2.
        payout_map = {1: 100, 2: 50}

        # WHEN we compute expected_payout with tie splitting
        out = contest_objective_from_sim_bids(
            team_points_scenarios=team_points_scenarios,
            market_entry_bids=market_entry_bids,
            market_team_totals=market_team_totals,
            sim_team_bids=sim_team_bids,
            objective="expected_payout",
            top_k=1,
            payout_map=payout_map,
        )

        # THEN payout is the average of positions 1 and 2
        self.assertEqual(out, 75.0)


class TestThatUtilityIsAppliedToPayout(unittest.TestCase):
    def test_that_expected_utility_payout_uses_power_utility(self) -> None:
        # GIVEN a single simulation with a 2-way tie for 1st
        team_points_scenarios = np.zeros((1, 1), dtype=float)
        market_entry_bids = np.zeros((1, 1), dtype=float)
        market_team_totals = np.ones((1,), dtype=float)
        sim_team_bids = np.ones((1,), dtype=float)
        payout_map = {1: 100, 2: 50}

        # WHEN using power utility with gamma=2
        out = contest_objective_from_sim_bids(
            team_points_scenarios=team_points_scenarios,
            market_entry_bids=market_entry_bids,
            market_team_totals=market_team_totals,
            sim_team_bids=sim_team_bids,
            objective="expected_utility_payout",
            top_k=1,
            payout_map=payout_map,
            utility="power",
            utility_gamma=2.0,
        )

        # THEN utility is mean((payout + eps)^gamma) with default epsilon=1e-6
        expected = (75.0 + 1e-6) ** 2
        self.assertEqual(out, expected)
