"""
Generate portfolio construction artifacts (production + debug).

Production artifact: Minimal fields for simulation
Debug artifact: All fields for inspection and debugging
"""
from __future__ import annotations

from typing import Tuple

import pandas as pd


def generate_recommended_bids(
    *,
    recommended_entry_bids: pd.DataFrame,
    tournament_value: pd.DataFrame,
    market_prediction: pd.DataFrame,
    predicted_total_pool_bids_points: float,
) -> Tuple[pd.DataFrame, pd.DataFrame]:
    """
    Generate recommended bids artifacts.

    Takes the output from recommend_entry_bids and enriches it with
    tournament value and market prediction data for debug artifact.

    Returns:
        (production_df, debug_df)

        production_df columns:
            - team_key: Unique team identifier
            - bid_amount_points: How many points to bid

        debug_df columns:
            - team_key: Unique team identifier
            - bid_amount_points: How many points to bid
            - expected_points_per_entry: From tournament_value
            - predicted_market_share: From market_prediction
            - predicted_cost_points: predicted_market_share * total_pool
            - ownership_fraction: bid_amount / predicted_cost
            - expected_return_points: expected_points * ownership_fraction
            - roi: expected_return / bid_amount
            - score: Internal optimization score
    """
    if "team_key" not in recommended_entry_bids.columns:
        raise ValueError("recommended_entry_bids missing team_key")
    if "bid_amount_points" not in recommended_entry_bids.columns:
        raise ValueError("recommended_entry_bids missing bid_amount_points")

    bids = recommended_entry_bids.copy()
    bids["team_key"] = bids["team_key"].astype(str)
    bids["bid_amount_points"] = pd.to_numeric(
        bids["bid_amount_points"],
        errors="coerce"
    ).fillna(0).astype(int)

    # Production artifact (minimal)
    production = bids[["team_key", "bid_amount_points"]].copy()

    # Debug artifact (comprehensive)
    debug = bids.copy()

    # Merge with tournament value
    if "expected_points_per_entry" not in debug.columns:
        tv = tournament_value[["team_key", "expected_points_per_entry"]].copy()
        debug = debug.merge(tv, on="team_key", how="left")

    # Merge with market prediction
    if "predicted_market_share" not in debug.columns:
        mp = market_prediction[["team_key", "predicted_market_share"]].copy()
        debug = debug.merge(mp, on="team_key", how="left")

    # Calculate derived fields
    debug["predicted_cost_points"] = (
        debug["predicted_market_share"] * predicted_total_pool_bids_points
    )

    # Ownership calculation: our bid moves the market
    # ownership = our_bid / (predicted_market_cost + our_bid)
    debug["ownership_fraction"] = debug.apply(
        lambda r: (
            r["bid_amount_points"] / (
                r["predicted_cost_points"] + r["bid_amount_points"]
            )
            if (r["predicted_cost_points"] + r["bid_amount_points"]) > 0
            else 0.0
        ),
        axis=1
    )

    debug["expected_return_points"] = (
        debug["expected_points_per_entry"] * debug["ownership_fraction"]
    )

    debug["roi"] = debug.apply(
        lambda r: (
            r["expected_return_points"] / r["bid_amount_points"]
            if r["bid_amount_points"] > 0
            else 0.0
        ),
        axis=1
    )

    # Ensure score column exists
    if "score" not in debug.columns:
        debug["score"] = 0.0

    # Order debug columns
    debug_cols = [
        "team_key",
        "bid_amount_points",
        "expected_points_per_entry",
        "predicted_market_share",
        "predicted_cost_points",
        "ownership_fraction",
        "expected_return_points",
        "roi",
        "score",
    ]

    # Add any other columns that exist
    other_cols = [c for c in debug.columns if c not in debug_cols]
    debug = debug[debug_cols + other_cols]

    return production, debug
