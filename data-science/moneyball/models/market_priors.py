"""
Market prior computation helpers for auction share prediction.

Extracted from predict_auction_share_of_pool() to reduce function size
and enable independent testing.
"""
from __future__ import annotations

from typing import Dict, List, Optional, Tuple

import numpy as np
import pandas as pd


def compute_seed_priors(
    train_team_dataset: pd.DataFrame,
    *,
    seed_prior_k: float = 0.0,
    seed_prior_monotone: Optional[bool] = None,
) -> Dict[int, float]:
    """
    Compute shrinkage-regularized seed-level market share priors.

    Args:
        train_team_dataset: Training data with 'seed' and 'team_share_of_pool'.
        seed_prior_k: Shrinkage strength toward the global mean.
        seed_prior_monotone: If True (default), enforce monotonically
            non-increasing priors from seed 1 to 16.

    Returns:
        Dict mapping seed (1-16) to prior market share.
    """
    y_train_raw = pd.to_numeric(
        train_team_dataset["team_share_of_pool"],
        errors="coerce",
    )
    tmp_y = train_team_dataset.assign(_y=y_train_raw).dropna(
        subset=["seed", "_y"]
    )
    y_global_mean = float(tmp_y["_y"].mean() or 0.0)

    seed_agg = tmp_y.groupby("seed")["_y"].agg(["sum", "count"])
    seed_prior_k_f = float(seed_prior_k or 0.0)
    vals: List[float] = []
    for i in range(1, 17):
        if i in seed_agg.index:
            s = float(seed_agg.loc[i, "sum"])
            c = float(seed_agg.loc[i, "count"])
        else:
            s = 0.0
            c = 0.0

        if seed_prior_k_f > 0.0:
            denom = c + seed_prior_k_f
            vals.append(
                float((s + seed_prior_k_f * y_global_mean) / denom)
                if denom > 0.0
                else 0.0
            )
        else:
            vals.append(float(s / c) if c > 0.0 else 0.0)

    use_monotone = (
        bool(seed_prior_monotone)
        if seed_prior_monotone is not None
        else True
    )
    if use_monotone and vals:
        vals = list(np.minimum.accumulate(np.asarray(vals, dtype=float)))
    return {i: float(vals[i - 1]) for i in range(1, 17)}


def compute_program_priors(
    train_team_dataset: pd.DataFrame,
    *,
    program_prior_k: float = 0.0,
) -> Tuple[Optional[Dict[str, float]], Optional[Dict[str, int]]]:
    """
    Compute shrinkage-regularized program-level market share priors.

    Args:
        train_team_dataset: Training data with 'school_slug' and
            'team_share_of_pool'.
        program_prior_k: Shrinkage strength toward the global mean.

    Returns:
        Tuple of (program_share_mean_by_slug, program_share_count_by_slug).
        Returns (None, None) if 'school_slug' is not present.
    """
    if "school_slug" not in train_team_dataset.columns:
        return None, None

    y_train_raw = pd.to_numeric(
        train_team_dataset["team_share_of_pool"],
        errors="coerce",
    )
    tmp = train_team_dataset.copy()
    tmp["_y"] = y_train_raw
    tmp["school_slug"] = tmp["school_slug"].astype(str).str.lower()
    tmp = tmp.dropna(subset=["_y"])
    y_global_mean = float(tmp["_y"].mean() or 0.0)

    grp = tmp.groupby("school_slug")["_y"]
    program_share_count_by_slug: Dict[str, int] = grp.size().to_dict()

    program_prior_k_f = float(program_prior_k or 0.0)
    if program_prior_k_f > 0.0:
        program_sum_by_slug = grp.sum().to_dict()
        program_share_mean_by_slug: Dict[str, float] = {
            str(slug): float(
                (
                    float(program_sum_by_slug.get(slug, 0.0))
                    + program_prior_k_f * y_global_mean
                )
                / (
                    float(program_share_count_by_slug.get(slug, 0))
                    + program_prior_k_f
                )
            )
            for slug in program_share_count_by_slug.keys()
        }
    else:
        program_share_mean_by_slug = grp.mean().to_dict()

    return program_share_mean_by_slug, program_share_count_by_slug
