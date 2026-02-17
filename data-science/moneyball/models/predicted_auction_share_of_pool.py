from __future__ import annotations

from typing import Dict, List, Optional, Tuple

import numpy as np
import pandas as pd

from moneyball.models import feature_engineering
from moneyball.models.analytical_tournament_value import (
    compute_analytical_tournament_values,
)

FEATURE_SETS = [
    "basic",
    "expanded",
    "enhanced",
    "optimal",
    "optimal_v2",
    "optimal_v3",
]

DEFAULT_FEATURE_SET = "optimal"


def _enrich_with_analytical_probabilities(
    df: pd.DataFrame,
    kenpom_scale: float = 10.0,
) -> pd.DataFrame:
    """
    Enrich team dataset with analytical championship probabilities.

    Requires a complete 68-team field with proper tournament structure.

    Args:
        df: Team dataset with id/team_key, seed, region, kenpom_net
        kenpom_scale: Scale parameter for win probability sigmoid

    Returns:
        DataFrame with added analytical_p_championship and
        analytical_expected_points columns
    """
    # Ensure we have required columns
    required = ["seed", "region", "kenpom_net"]
    missing = [c for c in required if c not in df.columns]
    if missing:
        raise ValueError(f"team_dataset missing columns for analytical: {missing}")

    # Use team_key or id as identifier
    id_col = "team_key" if "team_key" in df.columns else "id"
    if id_col not in df.columns:
        raise ValueError("team_dataset must have 'team_key' or 'id' column")

    # Compute analytical values
    analytical = compute_analytical_tournament_values(
        df,
        kenpom_scale=kenpom_scale,
    )

    # Merge back to original DataFrame
    result = df.copy()
    result["_merge_key"] = result[id_col].astype(str)
    analytical["_merge_key"] = analytical["team_key"].astype(str)

    result = result.merge(
        analytical[["_merge_key", "analytical_p_championship", "analytical_expected_points"]],
        on="_merge_key",
        how="left",
    )
    result = result.drop(columns=["_merge_key"])

    # Fill any missing values with 0
    result["analytical_p_championship"] = (
        pd.to_numeric(result["analytical_p_championship"], errors="coerce")
        .fillna(0.0)
    )
    result["analytical_expected_points"] = (
        pd.to_numeric(result["analytical_expected_points"], errors="coerce")
        .fillna(0.0)
    )

    return result


def _prepare_features_set(
    df: pd.DataFrame,
    feature_set: str,
    *,
    seed_prior_by_seed: Optional[Dict[int, float]] = None,
    program_share_mean_by_slug: Optional[Dict[str, float]] = None,
    program_share_count_by_slug: Optional[Dict[str, int]] = None,
) -> pd.DataFrame:
    fs = str(feature_set)
    if fs not in FEATURE_SETS:
        raise ValueError(f"unknown feature_set: {fs}")

    if fs == "enhanced":
        required = [
            "seed",
            "region",
            "kenpom_net",
            "kenpom_o",
            "kenpom_d",
            "kenpom_adj_t",
        ]
        missing = [c for c in required if c not in df.columns]
        if missing:
            raise ValueError(f"team_dataset missing columns: {missing}")

        base = df.copy()
        base = feature_engineering.add_all_enhanced_features(base)

        core_features = [
            "seed",
            "kenpom_net",
            "kenpom_o",
            "kenpom_d",
            "kenpom_adj_t",
        ]
        enhanced_features = (
            feature_engineering.get_enhanced_feature_columns()
        )

        last_year_features = [
            "has_last_year",
            "wins_last_year",
            "byes_last_year",
            "progress_last_year",
            "points_last_year",
            "total_bid_amount_last_year",
            "team_share_of_pool_last_year",
            "points_per_dollar_last_year",
            "bid_per_point_last_year",
            "expected_progress_last_year",
            "expected_points_last_year",
            "expected_points_per_dollar_last_year",
            "progress_ratio_last_year",
            "progress_residual_last_year",
            "roi_ratio_last_year",
        ]

        all_features = core_features + enhanced_features + last_year_features
        available = [c for c in all_features if c in base.columns]
        base = base[available + ["region"]].copy()

        for c in available:
            base[c] = pd.to_numeric(base[c], errors="coerce").fillna(0.0)

        X = pd.get_dummies(base, columns=["region"], dummy_na=True)
        return X

    if fs == "basic":
        required = ["seed", "region", "kenpom_net"]
    else:
        required = [
            "seed",
            "region",
            "kenpom_net",
            "kenpom_o",
            "kenpom_d",
            "kenpom_adj_t",
        ]

    missing = [c for c in required if c not in df.columns]
    if missing:
        raise ValueError(f"team_dataset missing columns: {missing}")

    base = df[required].copy()

    base["seed"] = pd.to_numeric(base["seed"], errors="coerce")
    base["kenpom_net"] = pd.to_numeric(base["kenpom_net"], errors="coerce")

    if fs != "basic":
        base["kenpom_o"] = pd.to_numeric(base["kenpom_o"], errors="coerce")
        base["kenpom_d"] = pd.to_numeric(base["kenpom_d"], errors="coerce")
        base["kenpom_adj_t"] = pd.to_numeric(
            base["kenpom_adj_t"],
            errors="coerce",
        )

    # Optimal feature set (data-driven from systematic feature testing)
    # Updated to use z-score normalization for better value detection
    if fs == "optimal":
        # 1. Championship equity (smart seed encoding)
        seed_title_prob = {
            1: 0.20, 2: 0.12, 3: 0.08, 4: 0.05, 5: 0.03, 6: 0.02,
            7: 0.01, 8: 0.01, 9: 0.005, 10: 0.003, 11: 0.002,
            12: 0.001, 13: 0.0005, 14: 0.0002, 15: 0.0001,
            16: 0.00001
        }
        base["champ_equity"] = base["seed"].map(seed_title_prob)

        # 2. KenPom z-score (captures magnitude of differences)
        kenpom_mean = base["kenpom_net"].mean()
        kenpom_std = base["kenpom_net"].std()
        base["kenpom_net_zscore"] = (
            (base["kenpom_net"] - kenpom_mean) / kenpom_std
        )

        # 3. KenPom z-score squared (emphasizes elite teams)
        base["kenpom_net_zscore_sq"] = base["kenpom_net_zscore"] ** 2

        # 4. KenPom z-score cubed (non-linear effect for extremes)
        base["kenpom_net_zscore_cubed"] = base["kenpom_net_zscore"] ** 3

        # 5. KenPom balance (offensive/defensive imbalance)
        kenpom_o_pct = base["kenpom_o"].rank(pct=True)
        kenpom_d_pct = base["kenpom_d"].rank(pct=True)
        base["kenpom_balance"] = np.abs(kenpom_o_pct - kenpom_d_pct)

        # 6. Points per equity (value play indicator)
        seed_expected_points = {
            1: 12, 2: 9, 3: 7, 4: 5, 5: 4, 6: 3, 7: 2, 8: 2,
            9: 1, 10: 1, 11: 1, 12: 1, 13: 0.5, 14: 0.3,
            15: 0.2, 16: 0.1
        }
        base["expected_points"] = base["seed"].map(seed_expected_points)
        base["points_per_equity"] = (
            base["expected_points"] / (base["champ_equity"] + 0.001)
        )

    if fs == "optimal_v2":
        seed_expected_points = {
            1: 12, 2: 9, 3: 7, 4: 5, 5: 4, 6: 3, 7: 2, 8: 2,
            9: 1, 10: 1, 11: 1, 12: 1, 13: 0.5, 14: 0.3,
            15: 0.2, 16: 0.1
        }
        base["expected_points"] = base["seed"].map(seed_expected_points)

        if seed_prior_by_seed:
            base["seed_market_prior"] = base["seed"].apply(
                lambda s: float(seed_prior_by_seed.get(int(s), 0.0))
                if pd.notna(s)
                else 0.0
            )
        else:
            base["seed_market_prior"] = 0.0

        kenpom_mean = base["kenpom_net"].mean()
        kenpom_std = base["kenpom_net"].std()
        base["kenpom_net_zscore"] = (
            (base["kenpom_net"] - kenpom_mean) / kenpom_std
        )
        base["kenpom_net_zscore_sq"] = base["kenpom_net_zscore"] ** 2
        base["kenpom_net_zscore_cubed"] = base["kenpom_net_zscore"] ** 3

        ko = pd.to_numeric(base["kenpom_o"], errors="coerce").fillna(0.0)
        kd = pd.to_numeric(base["kenpom_d"], errors="coerce").fillna(0.0)
        ko_std = float(ko.std() or 0.0)
        ko_z = (ko - float(ko.mean() or 0.0)) / (
            ko_std if ko_std > 0 else 1.0
        )
        kd_inv = -kd
        kd_inv_std = float(kd_inv.std() or 0.0)
        kd_z = (kd_inv - float(kd_inv.mean() or 0.0)) / (
            kd_inv_std if kd_inv_std > 0 else 1.0
        )
        base["kenpom_balance"] = np.abs(ko_z - kd_z)

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

    if fs == "optimal_v3":
        # Use analytical KenPom-based probabilities (NOT hard-coded seed tables)
        # Requires analytical_p_championship and analytical_expected_points
        # to be pre-computed via _enrich_with_analytical_probabilities()

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

        # 3. KenPom z-scores (same as optimal)
        kenpom_mean = base["kenpom_net"].mean()
        kenpom_std = base["kenpom_net"].std()
        if kenpom_std > 0:
            base["kenpom_net_zscore"] = (
                (base["kenpom_net"] - kenpom_mean) / kenpom_std
            )
        else:
            base["kenpom_net_zscore"] = 0.0
        base["kenpom_net_zscore_sq"] = base["kenpom_net_zscore"] ** 2
        base["kenpom_net_zscore_cubed"] = base["kenpom_net_zscore"] ** 3

        # 4. KenPom balance (offensive/defensive imbalance)
        ko = pd.to_numeric(base["kenpom_o"], errors="coerce").fillna(0.0)
        kd = pd.to_numeric(base["kenpom_d"], errors="coerce").fillna(0.0)
        ko_std = float(ko.std() or 0.0)
        ko_z = (ko - float(ko.mean() or 0.0)) / (
            ko_std if ko_std > 0 else 1.0
        )
        kd_inv = -kd
        kd_inv_std = float(kd_inv.std() or 0.0)
        kd_z = (kd_inv - float(kd_inv.mean() or 0.0)) / (
            kd_inv_std if kd_inv_std > 0 else 1.0
        )
        base["kenpom_balance"] = np.abs(ko_z - kd_z)

        # 5. Points per championship probability (value play indicator)
        base["points_per_p_champ"] = (
            base["expected_points"] / (base["p_championship"] + 1e-9)
        )

        # 6. Seed interactions (same as optimal_v2)
        base["seed_sq"] = base["seed"] ** 2
        base["kenpom_x_seed"] = base["kenpom_net"] * base["seed"]

        # 7. Market behavior features (upset chic, within-seed ranking)
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

    if fs != "basic" and fs != "optimal" and fs != "optimal_v3":
        base["seed_sq"] = base["seed"] ** 2
        base["kenpom_x_seed"] = base["kenpom_net"] * base["seed"]

    if fs in (
        "optimal",
        "optimal_v2",
    ):
        # Add KenPom interaction features (27.7% improvement)
        base["seed_sq"] = base["seed"] ** 2
        base["kenpom_x_seed"] = base["kenpom_net"] * base["seed"]

        # Add market behavior features

        if fs != "optimal_v2":
            blue_bloods = {
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
            base["is_blue_blood"] = df["school_slug"].apply(
                lambda x: 1 if str(x).lower() in blue_bloods else 0
            )

        # 2. Upset chic: seeds 10-12 systematically overbid (+1.6%)
        base["is_upset_seed"] = base["seed"].apply(
            lambda x: 1 if 10 <= x <= 12 else 0
        )

        # 3. Within-seed ranking: 3rd/4th teams undervalued (+3.6%)
        base["kenpom_rank_within_seed"] = df.groupby("seed")[
            "kenpom_net"
        ].rank(ascending=False, method="dense")
        # Normalize to 0-1 within each seed (higher = worse rank)
        base["kenpom_rank_within_seed_norm"] = base.groupby("seed")[
            "kenpom_rank_within_seed"
        ].transform(
            lambda x: (x - 1) / (x.max() - 1) if x.max() > 1 else 0
        )
        base["kenpom_rank_within_seed_norm"] = base[
            "kenpom_rank_within_seed_norm"
        ].fillna(0.0)

        # Drop the intermediate rank column
        base = base.drop(columns=["kenpom_rank_within_seed"])

    X = pd.get_dummies(base, columns=["region"], dummy_na=True)
    return X


def _align_columns(
    train: pd.DataFrame,
    test: pd.DataFrame,
) -> Tuple[pd.DataFrame, pd.DataFrame]:
    cols = sorted(set(train.columns).union(set(test.columns)))
    return train.reindex(columns=cols, fill_value=0.0), test.reindex(
        columns=cols,
        fill_value=0.0,
    )


def _fit_ridge(
    X: pd.DataFrame,
    y: pd.Series,
    alpha: float,
) -> Optional[np.ndarray]:
    if alpha < 0:
        raise ValueError("ridge alpha must be non-negative")

    m = X.copy()
    m.insert(0, "intercept", 1.0)

    valid = y.notna()
    for c in m.columns:
        valid &= m[c].notna()

    if int(valid.sum()) < 5:
        return None

    Xv = m.loc[valid].to_numpy(dtype=float)
    yv = y.loc[valid].to_numpy(dtype=float)

    xtx = Xv.T @ Xv
    reg = np.eye(xtx.shape[0], dtype=float) * float(alpha)
    reg[0, 0] = 0.0
    coef = np.linalg.solve(xtx + reg, Xv.T @ yv)
    return coef


def _predict_ridge(X: pd.DataFrame, coef: np.ndarray) -> np.ndarray:
    m = X.copy()
    m.insert(0, "intercept", 1.0)
    Xv = m.to_numpy(dtype=float)
    return Xv @ coef


def predict_auction_share_of_pool(
    *,
    train_team_dataset: pd.DataFrame,
    predict_team_dataset: pd.DataFrame,
    ridge_alpha: float = 1.0,
    feature_set: str = DEFAULT_FEATURE_SET,
    target_transform: str = "none",
    seed_prior_monotone: Optional[bool] = None,
    seed_prior_k: float = 0.0,
    program_prior_k: float = 0.0,
    kenpom_scale: float = 10.0,
) -> pd.DataFrame:
    if train_team_dataset.empty:
        raise ValueError("train_team_dataset must not be empty")
    if predict_team_dataset.empty:
        raise ValueError("predict_team_dataset must not be empty")

    if "team_share_of_pool" not in train_team_dataset.columns:
        raise ValueError("train team_dataset missing team_share_of_pool")

    fs = str(feature_set)
    tt = str(target_transform or "none").strip().lower()

    if tt not in ("none", "log"):
        raise ValueError(f"unknown target_transform: {tt}")

    # Validate hyperparameters for optimal_v2 feature set
    if fs == "optimal_v2" and seed_prior_k <= 0:
        raise ValueError(
            "optimal_v2 feature set requires seed_prior_k > 0 to stabilize "
            "market prior estimates for low seeds. Without shrinkage priors, "
            "the model produces unstable predictions that may not sum to 1.0. "
            "Recommended: seed_prior_k=20, program_prior_k=50."
        )

    seed_prior_by_seed: Optional[Dict[int, float]] = None
    program_share_mean_by_slug: Optional[Dict[str, float]] = None
    program_share_count_by_slug: Optional[Dict[str, int]] = None

    if fs == "optimal_v2":
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
        seed_prior_by_seed = {i: float(vals[i - 1]) for i in range(1, 17)}

        if "school_slug" in train_team_dataset.columns:
            tmp = train_team_dataset.copy()
            tmp["_y"] = y_train_raw
            tmp["school_slug"] = tmp["school_slug"].astype(str).str.lower()
            tmp = tmp.dropna(subset=["_y"])
            grp = tmp.groupby("school_slug")["_y"]
            program_share_count_by_slug = grp.size().to_dict()

            program_prior_k_f = float(program_prior_k or 0.0)
            if program_prior_k_f > 0.0:
                program_sum_by_slug = grp.sum().to_dict()
                program_share_mean_by_slug = {
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

    # Enrich with analytical probabilities for optimal_v3
    # IMPORTANT: Must process each year/snapshot separately since the analytical
    # calculation expects a single 68-team tournament, not concatenated multi-year data
    if fs == "optimal_v3":
        if "snapshot" in train_team_dataset.columns:
            # Process each year separately, then concatenate
            enriched_frames = []
            for snapshot, group in train_team_dataset.groupby("snapshot"):
                enriched = _enrich_with_analytical_probabilities(
                    group.copy(), kenpom_scale=kenpom_scale
                )
                enriched_frames.append(enriched)
            train_team_dataset = pd.concat(enriched_frames, ignore_index=True)
        else:
            # Single tournament - process directly
            train_team_dataset = _enrich_with_analytical_probabilities(
                train_team_dataset, kenpom_scale=kenpom_scale
            )
        predict_team_dataset = _enrich_with_analytical_probabilities(
            predict_team_dataset, kenpom_scale=kenpom_scale
        )

    X_train = _prepare_features_set(
        train_team_dataset,
        fs,
        seed_prior_by_seed=seed_prior_by_seed,
        program_share_mean_by_slug=program_share_mean_by_slug,
        program_share_count_by_slug=program_share_count_by_slug,
    )
    y_train = pd.to_numeric(
        train_team_dataset["team_share_of_pool"],
        errors="coerce",
    )

    X_pred = _prepare_features_set(
        predict_team_dataset,
        fs,
        seed_prior_by_seed=seed_prior_by_seed,
        program_share_mean_by_slug=program_share_mean_by_slug,
        program_share_count_by_slug=program_share_count_by_slug,
    )
    X_train, X_pred = _align_columns(X_train, X_pred)

    if tt == "log":
        y_train = np.log(
            pd.to_numeric(y_train, errors="coerce").fillna(0.0) + 1e-9
        )

    coef = _fit_ridge(X_train, y_train, alpha=float(ridge_alpha))
    if coef is None:
        raise ValueError("not enough valid training rows to fit model")

    yhat = _predict_ridge(X_pred, coef)
    yhat = np.asarray(yhat, dtype=float)
    yhat = np.where(np.isfinite(yhat), yhat, 0.0)

    if tt == "log":
        yhat = np.exp(np.clip(yhat, -20.0, 20.0))

    yhat = np.maximum(yhat, 0.0)

    s = float(yhat.sum())
    if s > 0:
        yhat = yhat / s
    else:
        yhat = np.ones_like(yhat, dtype=float) / float(len(yhat))

    out = predict_team_dataset.copy()
    out["predicted_auction_share_of_pool"] = yhat.astype(float)

    cols = [
        c
        for c in [
            "snapshot",
            "tournament_key",
            "calcutta_key",
            "team_key",
            "school_name",
            "school_slug",
            "seed",
            "region",
            "kenpom_net",
            "predicted_auction_share_of_pool",
        ]
        if c in out.columns
    ]
    return out[cols].copy()

