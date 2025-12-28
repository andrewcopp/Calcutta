from __future__ import annotations

from pathlib import Path
from typing import Dict, List

import numpy as np
import pandas as pd

import backtest_scaffold
import predict_market_share

from calcutta_ds.investment_report.market_bids import (
    compute_team_shares_from_bids,
)


def predict_share_for_snapshot(
    *,
    out_root: Path,
    snapshot_dirs: List[Path],
    predict_snapshot: str,
    train_snapshots: List[str],
    ridge_alpha: float,
    exclude_entry_names: List[str],
) -> pd.DataFrame:
    snapshot_dirs_by_name: Dict[str, Path] = {s.name: s for s in snapshot_dirs}

    pred_df = predict_market_share._load_team_dataset(
        snapshot_dirs_by_name[predict_snapshot]
    )
    if pred_df.empty:
        raise ValueError(
            f"empty team_dataset for snapshot: {predict_snapshot}"
        )

    if not train_snapshots:
        out = pred_df.copy()
        out["predicted_team_share_of_pool"] = 1.0 / float(len(out))
        return out

    train_frames: List[pd.DataFrame] = []
    for s in train_snapshots:
        sd = snapshot_dirs_by_name[s]
        t = backtest_scaffold._load_snapshot_tables(sd)
        ds = predict_market_share._load_team_dataset(sd)
        ck = backtest_scaffold._choose_calcutta_key(ds, None)
        ds = ds[ds["calcutta_key"] == ck].copy()

        teams = t.get("teams")
        required = {
            "team_key",
            "wins",
            "byes",
            "calcutta_key",
        }
        if teams is not None and required.issubset(set(teams.columns)):
            tt = teams[teams["calcutta_key"] == ck].copy()
            tt["wins"] = (
                pd.to_numeric(tt["wins"], errors="coerce")
                .fillna(0)
                .astype(int)
            )
            tt["byes"] = (
                pd.to_numeric(tt["byes"], errors="coerce")
                .fillna(0)
                .astype(int)
            )
            eligible_team_keys = set(
                tt[(tt["wins"] != 0) | (tt["byes"] != 0)]["team_key"]
                .astype(str)
                .tolist()
            )
            ds = ds[
                ds["team_key"].astype(str).isin(eligible_team_keys)
            ].copy()

        shares = compute_team_shares_from_bids(
            tables=t,
            calcutta_key=ck,
            exclude_entry_names=exclude_entry_names,
        )
        ds["team_share_of_pool"] = ds["team_key"].apply(
            lambda tk: float(shares.get(str(tk), 0.0))
        )
        train_frames.append(ds)

    train_df = pd.concat(train_frames, ignore_index=True)

    X_train = predict_market_share._prepare_features(train_df)
    y_train = pd.to_numeric(train_df["team_share_of_pool"], errors="coerce")

    X_pred = predict_market_share._prepare_features(pred_df)
    X_train, X_pred = predict_market_share._align_columns(X_train, X_pred)

    coef = predict_market_share._fit_ridge(
        X_train,
        y_train,
        alpha=float(ridge_alpha),
    )
    if coef is None:
        raise ValueError("not enough valid training rows to fit market model")

    yhat = predict_market_share._predict_ridge(X_pred, coef)
    yhat = np.asarray(yhat, dtype=float)
    yhat = np.where(np.isfinite(yhat), yhat, 0.0)
    yhat = np.maximum(yhat, 0.0)
    s = float(yhat.sum())
    if s > 0:
        yhat = yhat / s
    else:
        yhat = np.ones_like(yhat, dtype=float) / float(len(yhat))

    pred_df = pred_df.copy()
    pred_df["predicted_team_share_of_pool"] = yhat
    return pred_df
