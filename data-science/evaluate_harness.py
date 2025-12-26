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
    }

    for held_out in tournaments:
        train = all_rows[all_rows["tournament_key"] != held_out].copy()
        test = all_rows[all_rows["tournament_key"] == held_out].copy()

        X_train, y_total_train, y_share_train = _prepare_features(train)
        X_test, y_total_test, y_share_test = _prepare_features(test)

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

            fold["targets"][target_name] = {
                "seed_mean": seed_metrics,
                "ols": ols_metrics,
            }

        results["folds"].append(fold)

    out_json = json.dumps(results, indent=2) + "\n"
    if args.out_path:
        Path(args.out_path).write_text(out_json, encoding="utf-8")
    else:
        print(out_json, end="")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
