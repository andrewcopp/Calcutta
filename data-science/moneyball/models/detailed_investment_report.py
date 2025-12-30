"""
Detailed investment report for Moneyball pipeline.

Generates a comprehensive report showing all 64 tournament teams with:
- Expected performance metrics
- Market predictions vs actuals
- Portfolio allocation decisions
- ROI analysis
"""
from __future__ import annotations

import pandas as pd


def generate_detailed_investment_report(
    *,
    teams: pd.DataFrame,
    recommended_entry_bids: pd.DataFrame,
    predicted_auction_share_of_pool: pd.DataFrame,
    entry_bids: pd.DataFrame,
    predicted_total_pool_bids_points: float,
    tournament_value: pd.DataFrame,
) -> pd.DataFrame:
    """
    Generate detailed investment report with all 64 tournament teams.

    Args:
        teams: Team dataset
            Required columns: team_key, school_name, seed, region
        recommended_entry_bids: Our portfolio allocation
            Required columns: team_key, bid_amount_points
        predicted_auction_share_of_pool: Market share predictions
            Required columns: team_key, predicted_auction_share_of_pool
        entry_bids: Actual market bids from all entries
            Required columns: team_key, bid_amount, entry_key
        predicted_total_pool_bids_points: Predicted total pool size
        tournament_value: Tournament performance predictions
            Required columns: team_key, expected_points_per_entry

    Returns:
        DataFrame with one row per team (64 rows) containing:
            - team_key
            - school_name
            - seed
            - region
            - expected_points
            - expected_market
            - expected_roi
            - our_bid
            - actual_market (excluding our bid)
            - actual_roi
    """
    # Validate inputs
    required_teams = ["team_key", "school_name", "seed", "region"]
    missing = [c for c in required_teams if c not in teams.columns]
    if missing:
        raise ValueError(f"teams missing columns: {missing}")

    required_bids = ["team_key", "bid_amount_points"]
    missing = [c for c in required_bids if c not in recommended_entry_bids.columns]
    if missing:
        raise ValueError(f"recommended_entry_bids missing columns: {missing}")

    required_market = ["team_key", "predicted_auction_share_of_pool"]
    missing = [c for c in required_market if c not in predicted_auction_share_of_pool.columns]
    if missing:
        raise ValueError(f"predicted_auction_share_of_pool missing columns: {missing}")

    required_entry_bids = ["team_key", "bid_amount"]
    missing = [c for c in required_entry_bids if c not in entry_bids.columns]
    if missing:
        raise ValueError(f"entry_bids missing columns: {missing}")

    required_value = ["team_key", "expected_points_per_entry"]
    missing = [c for c in required_value if c not in tournament_value.columns]
    if missing:
        raise ValueError(f"tournament_value missing columns: {missing}")

    # Start with all 64 teams
    report = teams[["team_key", "school_name", "seed", "region"]].copy()

    # Add expected points from tournament value
    report = report.merge(
        tournament_value[["team_key", "expected_points_per_entry"]],
        on="team_key",
        how="left"
    )
    report = report.rename(columns={"expected_points_per_entry": "expected_points"})
    report["expected_points"] = report["expected_points"].fillna(0.0)

    # Add expected market from predictions
    market_pred = predicted_auction_share_of_pool[["team_key", "predicted_auction_share_of_pool"]].copy()
    market_pred["expected_market"] = (
        market_pred["predicted_auction_share_of_pool"] * predicted_total_pool_bids_points
    )
    report = report.merge(
        market_pred[["team_key", "expected_market"]],
        on="team_key",
        how="left"
    )
    report["expected_market"] = report["expected_market"].fillna(0.0)

    # Calculate expected ROI
    report["expected_roi"] = report.apply(
        lambda r: r["expected_points"] / r["expected_market"] if r["expected_market"] > 0 else 0.0,
        axis=1
    )

    # Add our bids
    our_bids = recommended_entry_bids[["team_key", "bid_amount_points"]].copy()
    our_bids = our_bids.rename(columns={"bid_amount_points": "our_bid"})
    report = report.merge(our_bids, on="team_key", how="left")
    report["our_bid"] = report["our_bid"].fillna(0).astype(int)

    # Calculate actual market (excluding our bid)
    actual_market = entry_bids.groupby("team_key")["bid_amount"].sum().reset_index()
    actual_market.columns = ["team_key", "actual_market_total"]
    report = report.merge(actual_market, on="team_key", how="left")
    report["actual_market_total"] = report["actual_market_total"].fillna(0.0)

    # Subtract our bid from actual market to get "actual market excluding us"
    report["actual_market"] = report["actual_market_total"] - report["our_bid"]
    report["actual_market"] = report["actual_market"].clip(lower=0.0)

    # Calculate actual ROI
    report["actual_roi"] = report.apply(
        lambda r: r["expected_points"] / r["actual_market"] if r["actual_market"] > 0 else 0.0,
        axis=1
    )

    # Select and order final columns
    final_cols = [
        "team_key",
        "school_name",
        "seed",
        "region",
        "expected_points",
        "expected_market",
        "expected_roi",
        "our_bid",
        "actual_market",
        "actual_roi",
    ]

    # Sort by seed, then by school name
    report = report.sort_values(["seed", "school_name"])

    return report[final_cols]
