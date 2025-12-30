"""
Generate market prediction artifacts (production + debug).

Production artifact: Minimal fields for portfolio construction
Debug artifact: All fields for inspection and debugging
"""
from __future__ import annotations

from typing import Tuple

import pandas as pd


def generate_market_prediction(
    *,
    predicted_auction_share_of_pool: pd.DataFrame,
) -> Tuple[pd.DataFrame, pd.DataFrame]:
    """
    Generate market prediction artifacts.

    Takes the output from predicted_auction_share_of_pool model and
    splits it into production (minimal) and debug (comprehensive) artifacts.

    Returns:
        (production_df, debug_df)

        production_df columns:
            - team_key: Unique team identifier
            - predicted_market_share: Predicted fraction of pool (0.0-1.0)

        debug_df columns:
            - team_key: Unique team identifier
            - predicted_market_share: Predicted fraction of pool (0.0-1.0)
            - (all other columns from input for debugging)
    """
    if "team_key" not in predicted_auction_share_of_pool.columns:
        raise ValueError("predicted_auction_share_of_pool missing team_key")
    if (
        "predicted_auction_share_of_pool"
        not in predicted_auction_share_of_pool.columns
    ):
        raise ValueError(
            "predicted_auction_share_of_pool missing "
            "predicted_auction_share_of_pool column"
        )

    df = predicted_auction_share_of_pool.copy()

    # Production artifact (minimal)
    production = pd.DataFrame({
        "team_key": df["team_key"].astype(str),
        "predicted_market_share": pd.to_numeric(
            df["predicted_auction_share_of_pool"],
            errors="coerce"
        ).fillna(0.0)
    })

    # Debug artifact (comprehensive - all columns)
    debug = df.copy()
    debug = debug.rename(
        columns={"predicted_auction_share_of_pool": "predicted_market_share"}
    )

    # Ensure team_key is first column in debug
    cols = ["team_key", "predicted_market_share"]
    other_cols = [c for c in debug.columns if c not in cols]
    debug = debug[cols + other_cols]

    return production, debug
