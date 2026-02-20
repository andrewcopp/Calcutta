"""
Feature engineering helpers for auction share prediction.

Extracted from predicted_auction_share_of_pool.py to reduce file size
and enable independent testing.
"""
from __future__ import annotations

from typing import Dict, Optional

import numpy as np
import pandas as pd


# -- Constants ----------------------------------------------------------------

SEED_TITLE_PROBABILITY = {
    1: 0.20, 2: 0.12, 3: 0.08, 4: 0.05, 5: 0.03, 6: 0.02,
    7: 0.01, 8: 0.01, 9: 0.005, 10: 0.003, 11: 0.002,
    12: 0.001, 13: 0.0005, 14: 0.0002, 15: 0.0001,
    16: 0.00001,
}

SEED_EXPECTED_POINTS = {
    1: 12, 2: 9, 3: 7, 4: 5, 5: 4, 6: 3, 7: 2, 8: 2,
    9: 1, 10: 1, 11: 1, 12: 1, 13: 0.5, 14: 0.3,
    15: 0.2, 16: 0.1,
}

BLUE_BLOODS = {
    "duke",
    "north-carolina",
    "kentucky",
    "kansas",
    "villanova",
    "michigan-state",
    "louisville",
    "connecticut",
    "ucla",
    "indiana",
    "gonzaga",
    "arizona",
}


# -- Shared helpers -----------------------------------------------------------

def compute_kenpom_net_zscore(kenpom_net: pd.Series) -> pd.Series:
    """Compute z-score of KenPom net rating within the field."""
    mean = kenpom_net.mean()
    std = kenpom_net.std()
    if std > 0:
        return (kenpom_net - mean) / std
    return pd.Series(0.0, index=kenpom_net.index)


def compute_kenpom_balance_zscore(
    kenpom_o: pd.Series,
    kenpom_d: pd.Series,
) -> pd.Series:
    """Compute offensive/defensive imbalance using z-score normalization."""
    ko = pd.to_numeric(kenpom_o, errors="coerce").fillna(0.0)
    kd = pd.to_numeric(kenpom_d, errors="coerce").fillna(0.0)
    ko_std = float(ko.std() or 0.0)
    ko_z = (ko - float(ko.mean() or 0.0)) / (
        ko_std if ko_std > 0 else 1.0
    )
    kd_inv = -kd
    kd_inv_std = float(kd_inv.std() or 0.0)
    kd_z = (kd_inv - float(kd_inv.mean() or 0.0)) / (
        kd_inv_std if kd_inv_std > 0 else 1.0
    )
    return np.abs(ko_z - kd_z)


def compute_kenpom_balance_percentile(
    kenpom_o: pd.Series,
    kenpom_d: pd.Series,
) -> pd.Series:
    """Compute offensive/defensive imbalance using percentile ranks."""
    kenpom_o_pct = kenpom_o.rank(pct=True)
    kenpom_d_pct = kenpom_d.rank(pct=True)
    return np.abs(kenpom_o_pct - kenpom_d_pct)


def compute_seed_interactions(
    seed: pd.Series,
    kenpom_net: pd.Series,
) -> pd.DataFrame:
    """Compute seed-based interaction features (seed^2, kenpom*seed)."""
    return pd.DataFrame({
        "seed_sq": seed ** 2,
        "kenpom_x_seed": kenpom_net * seed,
    }, index=seed.index)


def compute_market_behavior_features(
    df: pd.DataFrame,
    base: pd.DataFrame,
) -> pd.DataFrame:
    """
    Compute market behavior features: upset seeds and within-seed ranking.

    Args:
        df: Original dataframe (used for groupby on seed/kenpom_net).
        base: Feature dataframe to add columns to. Modified in-place
              and returned.

    Returns:
        Updated base dataframe with is_upset_seed and
        kenpom_rank_within_seed_norm columns. The intermediate
        kenpom_rank_within_seed column is dropped.
    """
    base["is_upset_seed"] = base["seed"].apply(
        lambda x: 1 if 10 <= x <= 12 else 0
    )

    base["kenpom_rank_within_seed"] = df.groupby("seed")[
        "kenpom_net"
    ].rank(ascending=False, method="dense")
    base["kenpom_rank_within_seed_norm"] = base.groupby("seed")[
        "kenpom_rank_within_seed"
    ].transform(
        lambda x: (x - 1) / (x.max() - 1) if x.max() > 1 else 0
    )
    base["kenpom_rank_within_seed_norm"] = base[
        "kenpom_rank_within_seed_norm"
    ].fillna(0.0)
    base = base.drop(columns=["kenpom_rank_within_seed"])
    return base


# -- Per-set assemblers -------------------------------------------------------

def prepare_optimal_features(
    df: pd.DataFrame,
    base: pd.DataFrame,
) -> pd.DataFrame:
    """
    Assemble the 'optimal' feature set.

    Args:
        df: Original dataframe (used for school_slug, groupby, etc.).
        base: Feature dataframe with seed, kenpom_net, kenpom_o, kenpom_d
              already coerced to numeric.

    Returns:
        Updated base dataframe with optimal features added.
    """
    # 1. Championship equity (smart seed encoding)
    base["champ_equity"] = base["seed"].map(SEED_TITLE_PROBABILITY)

    # 2. KenPom z-scores
    base["kenpom_net_zscore"] = compute_kenpom_net_zscore(base["kenpom_net"])
    base["kenpom_net_zscore_sq"] = base["kenpom_net_zscore"] ** 2
    base["kenpom_net_zscore_cubed"] = base["kenpom_net_zscore"] ** 3

    # 3. KenPom balance (percentile-based)
    base["kenpom_balance"] = compute_kenpom_balance_percentile(
        base["kenpom_o"], base["kenpom_d"],
    )

    # 4. Points per equity (value play indicator)
    base["expected_points"] = base["seed"].map(SEED_EXPECTED_POINTS)
    base["points_per_equity"] = (
        base["expected_points"] / (base["champ_equity"] + 0.001)
    )

    # 5. Seed interactions
    interactions = compute_seed_interactions(base["seed"], base["kenpom_net"])
    for col in interactions.columns:
        base[col] = interactions[col]

    # 6. Blue blood indicator
    base["is_blue_blood"] = df["school_slug"].apply(
        lambda x: 1 if str(x).lower() in BLUE_BLOODS else 0
    )

    # 7. Market behavior features
    base = compute_market_behavior_features(df, base)

    return base


def prepare_optimal_v2_features(
    df: pd.DataFrame,
    base: pd.DataFrame,
    *,
    seed_prior_by_seed: Optional[Dict[int, float]] = None,
    program_share_mean_by_slug: Optional[Dict[str, float]] = None,
    program_share_count_by_slug: Optional[Dict[str, int]] = None,
) -> pd.DataFrame:
    """
    Assemble the 'optimal_v2' feature set.

    Args:
        df: Original dataframe (used for school_slug, groupby, etc.).
        base: Feature dataframe with seed, kenpom_net, kenpom_o, kenpom_d
              already coerced to numeric.
        seed_prior_by_seed: Shrinkage-regularized seed priors.
        program_share_mean_by_slug: Shrinkage-regularized program means.
        program_share_count_by_slug: Count of observations per program.

    Returns:
        Updated base dataframe with optimal_v2 features added.
    """
    base["expected_points"] = base["seed"].map(SEED_EXPECTED_POINTS)

    if seed_prior_by_seed:
        base["seed_market_prior"] = base["seed"].apply(
            lambda s: float(seed_prior_by_seed.get(int(s), 0.0))
            if pd.notna(s)
            else 0.0
        )
    else:
        base["seed_market_prior"] = 0.0

    # KenPom z-scores
    base["kenpom_net_zscore"] = compute_kenpom_net_zscore(base["kenpom_net"])
    base["kenpom_net_zscore_sq"] = base["kenpom_net_zscore"] ** 2
    base["kenpom_net_zscore_cubed"] = base["kenpom_net_zscore"] ** 3

    # KenPom balance (z-score based)
    base["kenpom_balance"] = compute_kenpom_balance_zscore(
        base["kenpom_o"], base["kenpom_d"],
    )

    base["points_per_seed_market_prior"] = (
        base["expected_points"] / (base["seed_market_prior"] + 0.001)
    )

    slug = (
        df.get("school_slug")
        if "school_slug" in df.columns
        else pd.Series([""] * len(df), index=df.index)
    )
    slug = slug.astype(str).str.lower()
    if program_share_mean_by_slug:
        base["program_share_mean"] = slug.apply(
            lambda s: float(program_share_mean_by_slug.get(str(s), 0.0))
        )
    else:
        base["program_share_mean"] = 0.0

    if program_share_count_by_slug:
        base["program_share_count"] = slug.apply(
            lambda s: float(program_share_count_by_slug.get(str(s), 0))
        )
    else:
        base["program_share_count"] = 0.0

    # Seed interactions
    interactions = compute_seed_interactions(base["seed"], base["kenpom_net"])
    for col in interactions.columns:
        base[col] = interactions[col]

    # Market behavior features (no blue blood for v2)
    base = compute_market_behavior_features(df, base)

    return base


def prepare_optimal_v3_features(
    df: pd.DataFrame,
    base: pd.DataFrame,
) -> pd.DataFrame:
    """
    Assemble the 'optimal_v3' feature set.

    Uses analytical KenPom-based probabilities (NOT hard-coded seed tables).
    Requires analytical_p_championship and analytical_expected_points to be
    pre-computed via _enrich_with_analytical_probabilities().

    Args:
        df: Original dataframe with analytical columns.
        base: Feature dataframe with seed, kenpom_net, kenpom_o, kenpom_d
              already coerced to numeric.

    Returns:
        Updated base dataframe with optimal_v3 features added.
    """
    # 1. Championship probability from analytical calculation
    if "analytical_p_championship" not in df.columns:
        raise ValueError(
            "optimal_v3 requires analytical_p_championship column. "
            "Call _enrich_with_analytical_probabilities() first."
        )
    base["p_championship"] = pd.to_numeric(
        df["analytical_p_championship"], errors="coerce"
    ).fillna(0.0)

    # 2. Expected points from analytical calculation
    if "analytical_expected_points" not in df.columns:
        raise ValueError(
            "optimal_v3 requires analytical_expected_points column. "
            "Call _enrich_with_analytical_probabilities() first."
        )
    base["expected_points"] = pd.to_numeric(
        df["analytical_expected_points"], errors="coerce"
    ).fillna(0.0)

    # 3. KenPom z-scores
    base["kenpom_net_zscore"] = compute_kenpom_net_zscore(base["kenpom_net"])
    base["kenpom_net_zscore_sq"] = base["kenpom_net_zscore"] ** 2
    base["kenpom_net_zscore_cubed"] = base["kenpom_net_zscore"] ** 3

    # 4. KenPom balance (z-score based)
    base["kenpom_balance"] = compute_kenpom_balance_zscore(
        base["kenpom_o"], base["kenpom_d"],
    )

    # 5. Points per championship probability (value play indicator)
    base["points_per_p_champ"] = (
        base["expected_points"] / (base["p_championship"] + 1e-9)
    )

    # 6. Seed interactions
    interactions = compute_seed_interactions(base["seed"], base["kenpom_net"])
    for col in interactions.columns:
        base[col] = interactions[col]

    # 7. Market behavior features
    base = compute_market_behavior_features(df, base)

    return base
