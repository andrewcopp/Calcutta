"""
Unit tests for lab portfolio research module (GEKKO optimizer).

Following testing guidelines:
- GIVEN / WHEN / THEN structure
- Exactly one reason to fail / one assertion per test
- Deterministic tests
- Test naming: TestThat{Scenario}
"""
from __future__ import annotations

import unittest

import pandas as pd


def _create_teams_df(n_teams: int = 10) -> pd.DataFrame:
    """Create a test DataFrame with n teams."""
    return pd.DataFrame({
        "team_key": [f"team_{i}" for i in range(n_teams)],
        "expected_team_points": [100.0 - i * 5 for i in range(n_teams)],
        "predicted_team_total_bids": [20.0 - i for i in range(n_teams)],
    })


class TestThatGekkoOptimizerRespectsBudgetConstraint(unittest.TestCase):
    """
    GIVEN a portfolio optimization problem with budget=100
    WHEN we run the GEKKO optimizer
    THEN total bids should equal exactly 100
    """

    def test(self):
        # GIVEN
        from moneyball.lab.portfolio_research import optimize_portfolio_gekko

        teams_df = _create_teams_df(10)
        budget = 100

        # WHEN
        result, _ = optimize_portfolio_gekko(
            teams_df=teams_df,
            budget_points=budget,
            min_teams=3,
            max_teams=10,
            max_per_team_points=50,
            min_bid_points=1,
        )

        # THEN
        self.assertEqual(int(result["bid_amount_points"].sum()), budget)


class TestThatGekkoOptimizerRespectsMinTeamsConstraint(unittest.TestCase):
    """
    GIVEN a portfolio optimization problem with min_teams=5
    WHEN we run the GEKKO optimizer
    THEN result should have at least 5 teams
    """

    def test(self):
        # GIVEN
        from moneyball.lab.portfolio_research import optimize_portfolio_gekko

        teams_df = _create_teams_df(10)
        min_teams = 5

        # WHEN
        result, _ = optimize_portfolio_gekko(
            teams_df=teams_df,
            budget_points=100,
            min_teams=min_teams,
            max_teams=10,
            max_per_team_points=50,
            min_bid_points=1,
        )

        # THEN
        self.assertGreaterEqual(len(result), min_teams)


class TestThatGekkoOptimizerRespectsMaxTeamsConstraint(unittest.TestCase):
    """
    GIVEN a portfolio optimization problem with max_teams=4
    WHEN we run the GEKKO optimizer
    THEN result should have at most 4 teams
    """

    def test(self):
        # GIVEN
        from moneyball.lab.portfolio_research import optimize_portfolio_gekko

        teams_df = _create_teams_df(10)
        max_teams = 4

        # WHEN
        result, _ = optimize_portfolio_gekko(
            teams_df=teams_df,
            budget_points=100,
            min_teams=2,
            max_teams=max_teams,
            max_per_team_points=50,
            min_bid_points=1,
        )

        # THEN
        self.assertLessEqual(len(result), max_teams)


class TestThatGekkoOptimizerProducesIntegerBids(unittest.TestCase):
    """
    GIVEN a portfolio optimization problem
    WHEN we run the GEKKO optimizer
    THEN all bids should be integers
    """

    def test(self):
        # GIVEN
        from moneyball.lab.portfolio_research import optimize_portfolio_gekko

        teams_df = _create_teams_df(10)

        # WHEN
        result, _ = optimize_portfolio_gekko(
            teams_df=teams_df,
            budget_points=100,
            min_teams=3,
            max_teams=10,
            max_per_team_points=50,
            min_bid_points=1,
        )

        # THEN
        for bid in result["bid_amount_points"]:
            self.assertEqual(bid, int(bid))


class TestThatGekkoOptimizerRespectsMaxPerTeam(unittest.TestCase):
    """
    GIVEN a portfolio optimization problem with max_per_team=30
    WHEN we run the GEKKO optimizer
    THEN no bid should exceed 30
    """

    def test(self):
        # GIVEN
        from moneyball.lab.portfolio_research import optimize_portfolio_gekko

        teams_df = _create_teams_df(10)
        max_per_team = 30

        # WHEN
        result, _ = optimize_portfolio_gekko(
            teams_df=teams_df,
            budget_points=100,
            min_teams=3,
            max_teams=10,
            max_per_team_points=max_per_team,
            min_bid_points=1,
        )

        # THEN
        self.assertTrue((result["bid_amount_points"] <= max_per_team).all())


class TestThatGekkoOptimizerRespectsMinBid(unittest.TestCase):
    """
    GIVEN a portfolio optimization problem with min_bid=2
    WHEN we run the GEKKO optimizer
    THEN all bids should be at least 2
    """

    def test(self):
        # GIVEN
        from moneyball.lab.portfolio_research import optimize_portfolio_gekko

        teams_df = _create_teams_df(10)
        min_bid = 2

        # WHEN
        result, _ = optimize_portfolio_gekko(
            teams_df=teams_df,
            budget_points=100,
            min_teams=3,
            max_teams=10,
            max_per_team_points=50,
            min_bid_points=min_bid,
        )

        # THEN
        self.assertTrue((result["bid_amount_points"] >= min_bid).all())


class TestThatGekkoOptimizerIsDeterministic(unittest.TestCase):
    """
    GIVEN a portfolio optimization problem
    WHEN we run the GEKKO optimizer twice with same inputs
    THEN results should be identical
    """

    def test(self):
        # GIVEN
        from moneyball.lab.portfolio_research import optimize_portfolio_gekko

        teams_df = _create_teams_df(10)

        # WHEN
        result_a, _ = optimize_portfolio_gekko(
            teams_df=teams_df,
            budget_points=100,
            min_teams=3,
            max_teams=10,
            max_per_team_points=50,
            min_bid_points=1,
        )
        result_b, _ = optimize_portfolio_gekko(
            teams_df=teams_df,
            budget_points=100,
            min_teams=3,
            max_teams=10,
            max_per_team_points=50,
            min_bid_points=1,
        )

        # THEN
        self.assertEqual(
            sorted(result_a.to_dict(orient="records"), key=lambda x: x["team_key"]),
            sorted(result_b.to_dict(orient="records"), key=lambda x: x["team_key"]),
        )


class TestThatGekkoOptimizerReturnsPortfolioRows(unittest.TestCase):
    """
    GIVEN a portfolio optimization problem
    WHEN we run the GEKKO optimizer
    THEN portfolio_rows should contain valid entries
    """

    def test(self):
        # GIVEN
        from moneyball.lab.portfolio_research import optimize_portfolio_gekko

        teams_df = _create_teams_df(10)

        # WHEN
        result, portfolio_rows = optimize_portfolio_gekko(
            teams_df=teams_df,
            budget_points=100,
            min_teams=3,
            max_teams=10,
            max_per_team_points=50,
            min_bid_points=1,
        )

        # THEN
        self.assertEqual(len(portfolio_rows), len(result))


class TestThatGekkoOptimizerHandlesSmallBudget(unittest.TestCase):
    """
    GIVEN a portfolio optimization problem with budget=10 and min_teams=3
    WHEN we run the GEKKO optimizer
    THEN it should still find a valid solution
    """

    def test(self):
        # GIVEN
        from moneyball.lab.portfolio_research import optimize_portfolio_gekko

        teams_df = _create_teams_df(10)

        # WHEN
        result, _ = optimize_portfolio_gekko(
            teams_df=teams_df,
            budget_points=10,
            min_teams=3,
            max_teams=10,
            max_per_team_points=50,
            min_bid_points=1,
        )

        # THEN
        self.assertGreaterEqual(len(result), 3)


