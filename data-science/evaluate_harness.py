import argparse
import json
from pathlib import Path
from typing import Dict, List, Optional, Tuple

import numpy as np
import pandas as pd


def _find_snapshots(out_root: Path) -> List[Path]:
    if not out_root.exists():
        return []

    snapshots: List[Path] = []
    for p in sorted(out_root.iterdir()):
        if not p.is_dir():
            continue
        if (p / "derived" / "team_dataset.parquet").exists():
            snapshots.append(p)
    return snapshots


def _load_team_dataset(snapshot_dir: Path) -> pd.DataFrame:
    p = snapshot_dir / "derived" / "team_dataset.parquet"
    df = pd.read_parquet(p)
    df["snapshot"] = snapshot_dir.name
    return df


def _spearman(a: pd.Series, b: pd.Series) -> float:
    a_rank = a.rank(method="average")
    b_rank = b.rank(method="average")
    return float(a_rank.corr(b_rank))


def _mae(y: np.ndarray, yhat: np.ndarray) -> float:
    return float(np.mean(np.abs(y - yhat)))


def _rmse(y: np.ndarray, yhat: np.ndarray) -> float:
    return float(np.sqrt(np.mean((y - yhat) ** 2)))


def _prepare_features(
    df: pd.DataFrame,
) -> Tuple[pd.DataFrame, pd.Series, pd.Series]:
    required = [
        "seed",
        "region",
        "kenpom_net",
        "total_bid_amount",
        "team_share_of_pool",
    ]
    missing = [c for c in required if c not in df.columns]
    if missing:
        raise ValueError(f"team_dataset missing columns: {missing}")

    base = df[["seed", "region", "kenpom_net"]].copy()
    base["seed"] = pd.to_numeric(base["seed"], errors="coerce")
    base["kenpom_net"] = pd.to_numeric(base["kenpom_net"], errors="coerce")

    X = pd.get_dummies(base, columns=["region"], dummy_na=True)
    y_total = pd.to_numeric(df["total_bid_amount"], errors="coerce")
    y_share = pd.to_numeric(df["team_share_of_pool"], errors="coerce")
    return X, y_total, y_share


def _prepare_features_set(
    df: pd.DataFrame,
    feature_set: str,
) -> Tuple[pd.DataFrame, pd.Series, pd.Series]:
    X, y_total, y_share = _prepare_features(df)
    if feature_set == "basic":
        return X, y_total, y_share

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

    base = df[
        [
            "seed",
            "region",
            "kenpom_net",
            "kenpom_o",
            "kenpom_d",
            "kenpom_adj_t",
        ]
    ].copy()
    base["seed"] = pd.to_numeric(base["seed"], errors="coerce")
    base["kenpom_net"] = pd.to_numeric(base["kenpom_net"], errors="coerce")
    base["kenpom_o"] = pd.to_numeric(base["kenpom_o"], errors="coerce")
    base["kenpom_d"] = pd.to_numeric(base["kenpom_d"], errors="coerce")
    base["kenpom_adj_t"] = pd.to_numeric(
        base["kenpom_adj_t"],
        errors="coerce",
    )

    base["seed_sq"] = base["seed"] ** 2
    base["kenpom_x_seed"] = base["kenpom_net"] * base["seed"]

    X = pd.get_dummies(base, columns=["region"], dummy_na=True)
    return X, y_total, y_share


def _align_columns(
    train: pd.DataFrame,
    test: pd.DataFrame,
) -> Tuple[pd.DataFrame, pd.DataFrame]:
    cols = sorted(set(train.columns).union(set(test.columns)))
    return train.reindex(columns=cols, fill_value=0.0), test.reindex(
        columns=cols,
        fill_value=0.0,
    )


def _nanmean(xs: List[float]) -> float:
    v = [x for x in xs if x == x and np.isfinite(x)]
    if not v:
        return float("nan")
    return float(np.mean(v))


def _nanmedian(xs: List[float]) -> float:
    v = [x for x in xs if x == x and np.isfinite(x)]
    if not v:
        return float("nan")
    return float(np.median(v))


def _summarize_folds(results: Dict[str, object]) -> Dict[str, object]:
    # results["folds"][i]["targets"][target][model][metric]
    folds = results.get("folds")
    if not isinstance(folds, list):
        return {}

    summary: Dict[str, object] = {
        "targets": {},
    }

    for fold in folds:
        if not isinstance(fold, dict):
            continue
        targets = fold.get("targets")
        if not isinstance(targets, dict):
            continue

        for target_name, models in targets.items():
            if not isinstance(models, dict):
                continue
            t = summary["targets"].setdefault(str(target_name), {})
            for model_name, metrics in models.items():
                if not isinstance(metrics, dict):
                    continue
                m = t.setdefault(str(model_name), {})
                for metric_name in ["mae", "rmse", "spearman", "n"]:
                    m.setdefault(metric_name, []).append(
                        float(metrics.get(metric_name, float("nan")))
                    )

    out: Dict[str, object] = {"targets": {}}
    for target_name, models in summary["targets"].items():
        out_models: Dict[str, object] = {}
        for model_name, metric_lists in models.items():
            out_models[model_name] = {
                "mae_mean": _nanmean(metric_lists.get("mae", [])),
                "mae_median": _nanmedian(metric_lists.get("mae", [])),
                "rmse_mean": _nanmean(metric_lists.get("rmse", [])),
                "rmse_median": _nanmedian(metric_lists.get("rmse", [])),
                "spearman_mean": _nanmean(metric_lists.get("spearman", [])),
                "spearman_median": _nanmedian(
                    metric_lists.get("spearman", [])
                ),
                "n_mean": _nanmean(metric_lists.get("n", [])),
            }
        out["targets"][target_name] = out_models
    return out


def _fit_ols(X: pd.DataFrame, y: pd.Series) -> Optional[np.ndarray]:
    m = X.copy()
    m.insert(0, "intercept", 1.0)

    valid = y.notna()
    for c in m.columns:
        valid &= m[c].notna()

    if int(valid.sum()) < 5:
        return None

    Xv = m.loc[valid].to_numpy(dtype=float)
    yv = y.loc[valid].to_numpy(dtype=float)

    coef, *_ = np.linalg.lstsq(Xv, yv, rcond=None)
    return coef


def _predict_ols(X: pd.DataFrame, coef: np.ndarray) -> np.ndarray:
    m = X.copy()
    m.insert(0, "intercept", 1.0)
    Xv = m.to_numpy(dtype=float)
    return Xv @ coef


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

    # Ridge closed form: (X'X + alpha*I)^{-1} X'y
    # Intercept is not regularized.
    xtx = Xv.T @ Xv
    reg = np.eye(xtx.shape[0], dtype=float) * float(alpha)
    reg[0, 0] = 0.0
    coef = np.linalg.solve(xtx + reg, Xv.T @ yv)
    return coef


def _predict_ridge(X: pd.DataFrame, coef: np.ndarray) -> np.ndarray:
    return _predict_ols(X, coef)


def _predict_seed_mean(
    train: pd.DataFrame,
    test: pd.DataFrame,
    target: str,
) -> np.ndarray:
    tmp = train.copy()
    tmp["seed"] = pd.to_numeric(tmp["seed"], errors="coerce")
    tmp[target] = pd.to_numeric(tmp[target], errors="coerce")

    means = tmp.groupby("seed")[target].mean()
    global_mean = (
        float(tmp[target].mean())
        if tmp[target].notna().any()
        else 0.0
    )

    test_seed = pd.to_numeric(test["seed"], errors="coerce")
    pred = test_seed.map(means).fillna(global_mean).to_numpy(dtype=float)
    return pred


def _fold_metrics(y: pd.Series, yhat: np.ndarray) -> Dict[str, float]:
    valid = y.notna() & np.isfinite(yhat)
    if int(valid.sum()) == 0:
        return {
            "n": 0,
            "mae": float("nan"),
            "rmse": float("nan"),
            "spearman": float("nan"),
        }

    yy = y.loc[valid].to_numpy(dtype=float)
    yh = yhat[valid.to_numpy()]
    return {
        "n": int(valid.sum()),
        "mae": _mae(yy, yh),
        "rmse": _rmse(yy, yh),
        "spearman": _spearman(pd.Series(yy), pd.Series(yh)),
    }


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Evaluate simple baseline models over snapshots in an Option-A "
            "out-root. Reads derived/team_dataset.parquet."
        )
    )
    parser.add_argument(
        "out_root",
        help=(
            "Path to Option-A out-root (contains snapshot dirs)"
        ),
    )
    parser.add_argument(
        "--ridge-alpha",
        dest="ridge_alpha",
        type=float,
        default=1.0,
        help="Ridge regularization strength (default: 1.0)",
    )
    parser.add_argument(
        "--feature-set",
        dest="feature_set",
        default="basic",
        choices=["basic", "expanded"],
        help="Which feature set to use (default: basic)",
    )
    parser.add_argument(
        "--out",
        dest="out_path",
        default=None,
        help=(
            "Write metrics JSON to this path (default: stdout)"
        ),
    )

    args = parser.parse_args()

    out_root = Path(args.out_root)
    snapshots = _find_snapshots(out_root)
    if not snapshots:
        raise FileNotFoundError(f"no snapshots found under: {out_root}")

    all_rows = pd.concat(
        [_load_team_dataset(s) for s in snapshots],
        ignore_index=True,
    )

    if "tournament_key" not in all_rows.columns:
        raise ValueError(
            "team_dataset missing tournament_key "
            "(needed for leakage-safe folds)"
        )

    tournament_keys = all_rows["tournament_key"].dropna().unique()
    tournaments = [
        t
        for t in sorted(tournament_keys)
    ]
    if not tournaments:
        raise ValueError("no tournament_key values found")

    results: Dict[str, object] = {
        "snapshots": [s.name for s in snapshots],
        "folds": [],
        "config": {
            "ridge_alpha": float(args.ridge_alpha),
            "feature_set": str(args.feature_set),
        },
    }

    for held_out in tournaments:
        train = all_rows[all_rows["tournament_key"] != held_out].copy()
        test = all_rows[all_rows["tournament_key"] == held_out].copy()

        X_train, y_total_train, y_share_train = _prepare_features_set(
            train,
            str(args.feature_set),
        )
        X_test, y_total_test, y_share_test = _prepare_features_set(
            test,
            str(args.feature_set),
        )

        X_train, X_test = _align_columns(X_train, X_test)

        fold: Dict[str, object] = {
            "held_out_tournament_key": str(held_out),
            "n_train": int(len(train)),
            "n_test": int(len(test)),
            "targets": {},
        }

        for target_name, y_train, y_test, target_col in [
            (
                "total_bid_amount",
                y_total_train,
                y_total_test,
                "total_bid_amount",
            ),
            (
                "team_share_of_pool",
                y_share_train,
                y_share_test,
                "team_share_of_pool",
            ),
        ]:
            seed_pred = _predict_seed_mean(train, test, target_col)
            seed_metrics = _fold_metrics(y_test, seed_pred)

            coef = _fit_ols(X_train, y_train)
            if coef is None:
                ols_pred = np.full(
                    shape=(len(test),),
                    fill_value=np.nan,
                    dtype=float,
                )
            else:
                ols_pred = _predict_ols(X_test, coef)

            ols_metrics = _fold_metrics(y_test, ols_pred)

            ridge_coef = _fit_ridge(
                X_train,
                y_train,
                alpha=float(args.ridge_alpha),
            )
            if ridge_coef is None:
                ridge_pred = np.full(
                    shape=(len(test),),
                    fill_value=np.nan,
                    dtype=float,
                )
            else:
                ridge_pred = _predict_ridge(X_test, ridge_coef)

            ridge_metrics = _fold_metrics(y_test, ridge_pred)

            fold["targets"][target_name] = {
                "seed_mean": seed_metrics,
                "ols": ols_metrics,
                "ridge": ridge_metrics,
            }

        results["folds"].append(fold)

    results["summary"] = _summarize_folds(results)

    out_json = json.dumps(results, indent=2) + "\n"
    if args.out_path:
        Path(args.out_path).write_text(out_json, encoding="utf-8")
    else:
        print(out_json, end="")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
