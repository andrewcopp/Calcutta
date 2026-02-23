"""
Reusable DataFrame validators for data boundary checks.

Centralizes column-existence, non-null, and minimum-row validations
that were previously scattered as ad-hoc ValueError raises across the
prediction and lab modules.
"""

from __future__ import annotations

from typing import List, Optional

import pandas as pd


# -- Named column schemas ----------------------------------------------------

RIDGE_TRAINING_COLUMNS = [
    "seed",
    "region",
    "kenpom_net",
    "observed_team_share_of_pool",
]

RIDGE_PREDICTION_COLUMNS = [
    "seed",
    "region",
    "kenpom_net",
]

PREDICTION_OUTPUT_COLUMNS = [
    "team_slug",
    "predicted_market_share",
]


# -- Core validator -----------------------------------------------------------

def validate_dataframe(
    df: pd.DataFrame,
    *,
    required_columns: List[str],
    non_null_columns: Optional[List[str]] = None,
    min_rows: int = 0,
    context: str = "",
) -> None:
    """Validate a DataFrame against column, non-null, and row-count constraints.

    Args:
        df: The DataFrame to validate.
        required_columns: Columns that must exist in the DataFrame.
        non_null_columns: Columns that must not contain any null values.
            Defaults to None (no null checks).
        min_rows: Minimum number of rows required. Defaults to 0.
        context: Optional context string included in error messages
            (e.g. "ridge training data").

    Raises:
        ValueError: If any validation constraint is violated.
    """
    prefix = f"[{context}] " if context else ""

    missing = [c for c in required_columns if c not in df.columns]
    if missing:
        raise ValueError(f"{prefix}missing required columns: {missing}")

    if len(df) < min_rows:
        raise ValueError(
            f"{prefix}expected at least {min_rows} rows, got {len(df)}"
        )

    if non_null_columns:
        for col in non_null_columns:
            if col in df.columns and df[col].isna().any():
                n_null = int(df[col].isna().sum())
                raise ValueError(
                    f"{prefix}column '{col}' has {n_null} null values"
                )
