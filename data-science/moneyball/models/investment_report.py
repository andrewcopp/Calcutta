"""
Investment report model for Moneyball pipeline.

This pure function aggregates pipeline outputs into a human-readable report
containing portfolio details, expected performance, and ROI diagnostics.
"""
from __future__ import annotations

import pandas as pd


def generate_investment_report(
    *,
    recommended_entry_bids: pd.DataFrame,
    simulated_entry_outcomes: pd.DataFrame,
    predicted_game_outcomes: pd.DataFrame,
    predicted_auction_share_of_pool: pd.DataFrame,
    snapshot_name: str,
    budget_points: int,
    n_sims: int,
    seed: int,
) -> pd.DataFrame:
    """
    Generate investment report from pipeline artifacts.

    Args:
        recommended_entry_bids: Portfolio allocation
            Required columns: team_key, bid_amount_points
            Optional: expected_team_points, predicted_auction_share_of_pool
        simulated_entry_outcomes: Simulated performance distribution
            Required columns: entry_key, mean_payout_cents,
                mean_total_points, mean_finish_position,
                p_top1, p_top3, p_top6, p_top10
        predicted_game_outcomes: Game outcome predictions
            Required columns: team1_key, team2_key, p_matchup,
                p_team1_wins_given_matchup, p_team2_wins_given_matchup
        predicted_auction_share_of_pool: Market share predictions
            Required columns: team_key, predicted_auction_share_of_pool
        snapshot_name: Snapshot identifier
        budget_points: Entry budget
        n_sims: Number of simulations
        seed: Random seed

    Returns:
        DataFrame with single row containing report summary:
            - snapshot_name
            - budget_points
            - n_sims
            - seed
            - portfolio_team_count
            - portfolio_total_bids
            - mean_expected_payout_cents
            - mean_expected_points
            - mean_expected_finish_position
            - p_top1, p_top3, p_top6, p_top10
            - portfolio_concentration_hhi
            - portfolio_teams (JSON list)
    """
    bids = recommended_entry_bids.copy()
    outcomes = simulated_entry_outcomes.copy()

    if bids.empty:
        raise ValueError("recommended_entry_bids is empty")
    if outcomes.empty:
        raise ValueError("simulated_entry_outcomes is empty")

    required_bids = ["team_key", "bid_amount_points"]
    missing = [c for c in required_bids if c not in bids.columns]
    if missing:
        raise ValueError(
            f"recommended_entry_bids missing columns: {missing}"
        )

    required_outcomes = [
        "mean_payout_cents",
        "mean_total_points",
        "mean_finish_position",
        "mean_n_entries",
        "p_top1",
        "p_top3",
        "p_top6",
        "p_top10",
    ]
    missing = [c for c in required_outcomes if c not in outcomes.columns]
    if missing:
        raise ValueError(
            f"simulated_entry_outcomes missing columns: {missing}"
        )

    portfolio_team_count = len(bids)
    portfolio_total_bids = int(bids["bid_amount_points"].sum())

    outcome_row = outcomes.iloc[0]
    mean_payout_cents = float(outcome_row["mean_payout_cents"])
    mean_points = float(outcome_row["mean_total_points"])
    mean_finish = float(outcome_row["mean_finish_position"])
    mean_n_entries = float(outcome_row["mean_n_entries"])
    p_top1 = float(outcome_row["p_top1"])
    p_top3 = float(outcome_row["p_top3"])
    p_top6 = float(outcome_row["p_top6"])
    p_top10 = float(outcome_row["p_top10"])

    bid_shares = bids["bid_amount_points"] / portfolio_total_bids
    hhi = float((bid_shares ** 2).sum())

    portfolio_teams = bids.to_dict(orient="records")

    report = pd.DataFrame([{
        "snapshot_name": snapshot_name,
        "budget_points": budget_points,
        "n_sims": n_sims,
        "seed": seed,
        "portfolio_team_count": portfolio_team_count,
        "portfolio_total_bids": portfolio_total_bids,
        "mean_expected_payout_cents": mean_payout_cents,
        "mean_expected_points": mean_points,
        "mean_expected_finish_position": mean_finish,
        "mean_n_entries": mean_n_entries,
        "p_top1": p_top1,
        "p_top3": p_top3,
        "p_top6": p_top6,
        "p_top10": p_top10,
        "portfolio_concentration_hhi": hhi,
        "portfolio_teams_json": str(portfolio_teams),
    }])

    return report
