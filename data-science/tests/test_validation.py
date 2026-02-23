"""Tests for moneyball.validation module."""

import pandas as pd
import pytest

from moneyball.validation import validate_dataframe


def test_that_missing_column_raises_with_column_name():
    # GIVEN a DataFrame missing 'seed'
    df = pd.DataFrame({"region": ["East"]})

    # WHEN validating with 'seed' required
    # THEN ValueError mentions the missing column
    with pytest.raises(ValueError, match="seed"):
        validate_dataframe(df, required_columns=["seed"])


def test_that_empty_dataframe_raises_when_min_rows_required():
    # GIVEN an empty DataFrame with the right columns
    df = pd.DataFrame({"seed": pd.Series([], dtype=int)})

    # WHEN validating with min_rows=1
    # THEN ValueError mentions row count
    with pytest.raises(ValueError, match="at least 1 rows, got 0"):
        validate_dataframe(df, required_columns=["seed"], min_rows=1)


def test_that_valid_dataframe_passes():
    # GIVEN a DataFrame that meets all constraints
    df = pd.DataFrame({"seed": [1, 2], "region": ["East", "West"]})

    # WHEN validating
    # THEN no exception is raised
    validate_dataframe(
        df,
        required_columns=["seed", "region"],
        non_null_columns=["seed"],
        min_rows=1,
    )


def test_that_context_appears_in_error_message():
    # GIVEN a DataFrame missing a column
    df = pd.DataFrame({"a": [1]})

    # WHEN validating with context="ridge training"
    # THEN error message includes the context
    with pytest.raises(ValueError, match=r"\[ridge training\]"):
        validate_dataframe(
            df,
            required_columns=["b"],
            context="ridge training",
        )


def test_that_null_in_non_null_column_raises():
    # GIVEN a DataFrame with a null value in 'seed'
    df = pd.DataFrame({"seed": [1, None, 3]})

    # WHEN validating with seed as non-null
    # THEN ValueError mentions the null count
    with pytest.raises(ValueError, match="column 'seed' has 1 null"):
        validate_dataframe(
            df,
            required_columns=["seed"],
            non_null_columns=["seed"],
        )
