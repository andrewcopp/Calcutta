from __future__ import annotations

import logging
from typing import Dict, Optional, Tuple

import numpy as np
import pandas as pd

from moneyball.db.readers import (
    enrich_with_analytical_probabilities as _enrich_with_analytical_probabilities,
)
from moneyball.models.features import (
    prepare_optimal_v1_features,
    prepare_optimal_v2_features,
    prepare_optimal_v3_features,
)
from moneyball.models.market_priors import (
    compute_program_priors,
    compute_seed_priors,
)

logger = logging.getLogger(__name__)

FEATURE_SETS = [
    "basic",
    "optimal",
    "optimal_v2",
    "optimal_v3",
]

DEFAULT_FEATURE_SET = "optimal"


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

    if fs == "optimal":
        base = prepare_optimal_v1_features(df, base)

    if fs == "optimal_v2":
        base = prepare_optimal_v2_features(
            df,
            base,
            seed_prior_by_seed=seed_prior_by_seed,
            program_share_mean_by_slug=program_share_mean_by_slug,
            program_share_count_by_slug=program_share_count_by_slug,
        )

    if fs == "optimal_v3":
        base = prepare_optimal_v3_features(df, base)

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


def predict_market_share(
    *,
    train_team_dataset: pd.DataFrame,
    predict_team_dataset: pd.DataFrame,
    ridge_alpha: float = 1.0,
    feature_set: str = DEFAULT_FEATURE_SET,
    target_transform: str = "none",
    seed_prior_monotone: Optional[bool] = None,
    seed_prior_k: float = 0.0,
    program_prior_k: float = 0.0,
) -> pd.DataFrame:
    if train_team_dataset.empty:
        raise ValueError("train_team_dataset must not be empty")
    if predict_team_dataset.empty:
        raise ValueError("predict_team_dataset must not be empty")

    if "observed_team_share_of_pool" not in train_team_dataset.columns:
        raise ValueError("train team_dataset missing observed_team_share_of_pool")

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
        seed_prior_by_seed = compute_seed_priors(
            train_team_dataset,
            seed_prior_k=seed_prior_k,
            seed_prior_monotone=seed_prior_monotone,
        )
        program_share_mean_by_slug, program_share_count_by_slug = (
            compute_program_priors(
                train_team_dataset,
                program_prior_k=program_prior_k,
            )
        )

    # Enrich with analytical probabilities for optimal_v3
    # IMPORTANT: Must process each year/snapshot separately since the analytical
    # calculation expects a single 68-team tournament, not concatenated multi-year data
    # Skip enrichment when columns already exist (e.g. pre-populated by tests)
    if fs == "optimal_v3":
        train_needs = "predicted_p_championship" not in train_team_dataset.columns
        pred_needs = "predicted_p_championship" not in predict_team_dataset.columns

        if train_needs:
            if "snapshot" in train_team_dataset.columns:
                # Process each year separately, then concatenate
                enriched_frames = []
                for snapshot, group in train_team_dataset.groupby("snapshot"):
                    enriched = _enrich_with_analytical_probabilities(
                        group.copy()
                    )
                    enriched_frames.append(enriched)
                train_team_dataset = pd.concat(enriched_frames, ignore_index=True)
            else:
                # Single tournament - process directly
                train_team_dataset = _enrich_with_analytical_probabilities(
                    train_team_dataset
                )
        if pred_needs:
            predict_team_dataset = _enrich_with_analytical_probabilities(
                predict_team_dataset
            )

    X_train = _prepare_features_set(
        train_team_dataset,
        fs,
        seed_prior_by_seed=seed_prior_by_seed,
        program_share_mean_by_slug=program_share_mean_by_slug,
        program_share_count_by_slug=program_share_count_by_slug,
    )
    y_train = pd.to_numeric(
        train_team_dataset["observed_team_share_of_pool"],
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
    out["predicted_market_share"] = yhat.astype(float)

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
            "predicted_market_share",
        ]
        if c in out.columns
    ]
    return out[cols].copy()

