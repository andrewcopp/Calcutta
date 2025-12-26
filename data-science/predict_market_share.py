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


def _prepare_features(df: pd.DataFrame) -> pd.DataFrame:
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
    base["kenpom_adj_t"] = pd.to_numeric(base["kenpom_adj_t"], errors="coerce")

    base["seed_sq"] = base["seed"] ** 2
    base["kenpom_x_seed"] = base["kenpom_net"] * base["seed"]

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


def _choose_default_out_paths(
    out_root: Path,
    predict_snapshot: str,
) -> Tuple[Path, Path]:
    base = out_root / predict_snapshot / "derived"
    return (
        base / "predicted_team_share_of_pool.csv",
        base / "predicted_team_share_of_pool.json",
    )


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Train a no-leak Ridge model to predict team_share_of_pool "
            "from prior snapshots and output predictions for a target "
            "snapshot."
        )
    )
    parser.add_argument(
        "out_root",
        help="Path to Option-A out-root (contains snapshot dirs)",
    )
    parser.add_argument(
        "--predict-snapshot",
        dest="predict_snapshot",
        default=None,
        help=(
            "Snapshot dir name to generate predictions for "
            "(default: out_root/LATEST)"
        ),
    )
    parser.add_argument(
        "--train-snapshots",
        dest="train_snapshots",
        default=None,
        help=(
            "Comma-separated list of snapshot dir names to train on "
            "(default: all snapshots except predict snapshot)"
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
        "--out-csv",
        dest="out_csv",
        default=None,
        help=(
            "Write predictions CSV to this path "
            "(default: <predict>/derived/...)"
        ),
    )
    parser.add_argument(
        "--out-json",
        dest="out_json",
        default=None,
        help=(
            "Write predictions JSON to this path "
            "(default: <predict>/derived/...)"
        ),
    )

    args = parser.parse_args()

    out_root = Path(args.out_root)
    snapshots = _find_snapshots(out_root)
    if not snapshots:
        raise FileNotFoundError(
            f"no snapshots found under: {out_root}"
        )

    latest_path = out_root / "LATEST"
    default_predict = (
        latest_path.read_text(encoding="utf-8").strip()
        if latest_path.exists()
        else None
    )
    predict_snapshot = str(
        args.predict_snapshot
        or default_predict
        or ""
    ).strip()
    if not predict_snapshot:
        raise ValueError(
            "--predict-snapshot is required (or create out_root/LATEST)"
        )

    snapshot_dirs_by_name: Dict[str, Path] = {s.name: s for s in snapshots}
    if predict_snapshot not in snapshot_dirs_by_name:
        raise FileNotFoundError(
            (
                "predict snapshot not found under out_root: "
                f"{predict_snapshot}"
            )
        )

    if args.train_snapshots:
        train_names = [
            s.strip()
            for s in str(args.train_snapshots).split(",")
            if s.strip()
        ]
    else:
        train_names = [s.name for s in snapshots if s.name != predict_snapshot]

    missing_train = [s for s in train_names if s not in snapshot_dirs_by_name]
    if missing_train:
        raise FileNotFoundError(
            "train snapshots not found under out_root: "
            f"{missing_train}"
        )

    train_frames = [
        _load_team_dataset(snapshot_dirs_by_name[s])
        for s in train_names
    ]
    train_df = pd.concat(train_frames, ignore_index=True)

    predict_df = _load_team_dataset(snapshot_dirs_by_name[predict_snapshot])

    if "team_share_of_pool" not in train_df.columns:
        raise ValueError("team_dataset missing team_share_of_pool")

    X_train = _prepare_features(train_df)
    y_train = pd.to_numeric(train_df["team_share_of_pool"], errors="coerce")

    X_pred = _prepare_features(predict_df)
    X_train, X_pred = _align_columns(X_train, X_pred)

    coef = _fit_ridge(X_train, y_train, alpha=float(args.ridge_alpha))
    if coef is None:
        raise ValueError("not enough valid training rows to fit model")

    yhat = _predict_ridge(X_pred, coef)

    out = predict_df.copy()
    out["predicted_team_share_of_pool"] = yhat

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
            "predicted_team_share_of_pool",
            "team_share_of_pool",
        ]
        if c in out.columns
    ]
    out = out[cols].copy()

    default_csv, default_json = _choose_default_out_paths(
        out_root,
        predict_snapshot,
    )
    out_csv = Path(args.out_csv) if args.out_csv else default_csv
    out_json = Path(args.out_json) if args.out_json else default_json

    out_csv.parent.mkdir(parents=True, exist_ok=True)
    out_json.parent.mkdir(parents=True, exist_ok=True)

    out.to_csv(out_csv, index=False)

    payload = {
        "config": {
            "out_root": str(out_root),
            "train_snapshots": train_names,
            "predict_snapshot": predict_snapshot,
            "ridge_alpha": float(args.ridge_alpha),
            "target": "team_share_of_pool",
            "features": [c for c in X_train.columns],
        },
        "rows": out.to_dict(orient="records"),
    }
    out_json.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")

    print(str(out_csv))
    print(str(out_json))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
