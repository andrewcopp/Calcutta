import argparse
import json
import sys
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

import numpy as np
import pandas as pd


def _ensure_project_root_on_path() -> Path:
    project_root = Path(__file__).resolve().parents[1]
    if str(project_root) not in sys.path:
        sys.path.insert(0, str(project_root))
    return project_root


def _seed_band(seed: Any) -> str:
    try:
        s = int(seed)
    except Exception:
        return "unknown"
    if 1 <= s <= 2:
        return "1-2"
    if 3 <= s <= 4:
        return "3-4"
    if 5 <= s <= 8:
        return "5-8"
    if 9 <= s <= 12:
        return "9-12"
    if 13 <= s <= 16:
        return "13-16"
    return "unknown"


def _top_k_overlap(
    *, actual: pd.Series, pred: pd.Series, k: int
) -> float:
    k = int(k)
    if k <= 0:
        return 0.0
    a_idx = set(actual.nlargest(k).index.tolist())
    p_idx = set(pred.nlargest(k).index.tolist())
    if not a_idx:
        return 0.0
    return float(len(a_idx.intersection(p_idx))) / float(k)


def _metrics_for_year(
    *,
    train_years: List[int],
    predict_year: int,
    exclude_entry_name: str,
    ridge_alpha: float,
    feature_set: str,
    target_transform: str,
    seed_prior_monotone: Optional[bool],
    seed_prior_k: float,
    program_prior_k: float,
) -> Dict[str, Any]:
    from moneyball.db.readers import read_ridge_team_dataset_for_year
    from moneyball.models.predicted_auction_share_of_pool import (
        predict_auction_share_of_pool,
    )

    frames: List[pd.DataFrame] = []
    train_errors: List[str] = []
    for y in train_years:
        try:
            df = read_ridge_team_dataset_for_year(
                y,
                exclude_entry_names=[exclude_entry_name],
                include_target=True,
            )
        except Exception as e:
            train_errors.append(f"train_year={y}: {type(e).__name__}: {e}")
            continue
        frames.append(df)

    if not frames:
        raise ValueError(
            "no training data loaded; "
            f"predict_year={predict_year}; "
            f"errors={train_errors[:3]}"
        )

    train_ds = pd.concat(frames, ignore_index=True)
    train_ds = train_ds[train_ds["team_share_of_pool"].notna()].copy()
    if train_ds.empty:
        raise ValueError("no training rows (team_share_of_pool all NULL)")

    pred_ds = read_ridge_team_dataset_for_year(
        int(predict_year),
        exclude_entry_names=None,
        include_target=True,
    )
    pred_ds = pred_ds[pred_ds["team_share_of_pool"].notna()].copy()

    preds = predict_auction_share_of_pool(
        train_team_dataset=train_ds,
        predict_team_dataset=pred_ds,
        ridge_alpha=float(ridge_alpha),
        feature_set=str(feature_set),
        target_transform=str(target_transform),
        seed_prior_monotone=seed_prior_monotone,
        seed_prior_k=float(seed_prior_k),
        program_prior_k=float(program_prior_k),
    )

    merged = pred_ds.merge(
        preds[["team_key", "predicted_auction_share_of_pool"]],
        on="team_key",
        how="inner",
    )

    y = pd.to_numeric(
        merged["team_share_of_pool"],
        errors="coerce",
    ).fillna(0.0)
    yhat = pd.to_numeric(
        merged["predicted_auction_share_of_pool"],
        errors="coerce",
    ).fillna(0.0)

    eps = 1e-12
    mae = float(np.mean(np.abs(y - yhat)))
    rmse = float(np.sqrt(np.mean((y - yhat) ** 2)))
    kl = float(
        np.sum((y + eps) * np.log((y + eps) / (yhat + eps)))
    )

    bands = (
        merged.assign(
            seed_band=merged["seed"].apply(_seed_band),
            err=(yhat - y),
            abs_err=np.abs(yhat - y),
        )
        .groupby("seed_band")
        .agg(
            mean_err=("err", "mean"),
            mae=("abs_err", "mean"),
        )
        .reset_index()
        .sort_values("seed_band")
    )

    top10 = _top_k_overlap(actual=y, pred=yhat, k=10)
    top5 = _top_k_overlap(actual=y, pred=yhat, k=5)

    return {
        "predict_year": int(predict_year),
        "feature_set": str(feature_set),
        "target_transform": str(target_transform),
        "seed_prior_monotone": seed_prior_monotone,
        "seed_prior_k": float(seed_prior_k),
        "program_prior_k": float(program_prior_k),
        "mae": mae,
        "rmse": rmse,
        "kl": kl,
        "top5_overlap": top5,
        "top10_overlap": top10,
        "seed_band": bands.to_dict(orient="records"),
    }


def _probe_year(
    *,
    year: int,
    exclude_entry_name: str,
) -> Tuple[bool, str]:
    from moneyball.db.readers import read_ridge_team_dataset_for_year

    try:
        df = read_ridge_team_dataset_for_year(
            int(year),
            exclude_entry_names=[exclude_entry_name],
            include_target=True,
        )
    except Exception as e:
        return False, f"{type(e).__name__}: {e}"

    if df is None or df.empty:
        return False, "empty dataset"
    if "team_share_of_pool" not in df.columns:
        return False, "missing team_share_of_pool"
    if df["team_share_of_pool"].notna().sum() == 0:
        return False, "team_share_of_pool all NULL"
    return True, ""


def _discover_snapshots(*, out_root: Path) -> List[str]:
    if not out_root.exists():
        return []
    out: List[str] = []
    for p in sorted(out_root.iterdir()):
        if not p.is_dir():
            continue
        if (p / "derived" / "team_dataset.parquet").exists():
            out.append(p.name)
    return out


def _read_snapshot_team_dataset(
    *,
    out_root: Path,
    snapshot: str,
) -> pd.DataFrame:
    p = out_root / str(snapshot) / "derived" / "team_dataset.parquet"
    df = pd.read_parquet(p)
    df["snapshot"] = str(snapshot)
    return df


def _read_snapshot_actual_shares(
    *,
    out_root: Path,
    snapshot: str,
    exclude_entry_name: str,
) -> pd.DataFrame:
    from moneyball.utils import io
    from moneyball.utils.market_bids import compute_team_shares_from_bids

    sd = out_root / str(snapshot)
    ds = _read_snapshot_team_dataset(out_root=out_root, snapshot=str(snapshot))
    ck = io.choose_calcutta_key(ds, None)

    tables = io.load_snapshot_tables(sd)
    shares = compute_team_shares_from_bids(
        tables=tables,
        calcutta_key=ck,
        exclude_entry_names=[exclude_entry_name] if exclude_entry_name else [],
    )

    out = ds[ds["calcutta_key"] == ck].copy()
    out["team_share_of_pool"] = out["team_key"].apply(
        lambda tk: float(shares.get(str(tk), 0.0))
    )
    return out


def _metrics_for_snapshot(
    *,
    out_root: Path,
    snapshots: List[str],
    predict_snapshot: str,
    exclude_entry_name: str,
    ridge_alpha: float,
    feature_set: str,
    target_transform: str,
    seed_prior_monotone: Optional[bool],
    seed_prior_k: float,
    program_prior_k: float,
) -> Dict[str, Any]:
    from moneyball.models.predicted_auction_share_of_pool import (
        predict_auction_share_of_pool_from_out_root,
        predict_auction_share_of_pool,
    )

    actual = _read_snapshot_actual_shares(
        out_root=out_root,
        snapshot=str(predict_snapshot),
        exclude_entry_name=str(exclude_entry_name),
    )
    actual = actual[actual["team_share_of_pool"].notna()].copy()
    if actual.empty:
        raise ValueError(
            f"empty actual shares for snapshot={predict_snapshot}"
        )

    train = [s for s in snapshots if str(s) != str(predict_snapshot)]
    pred = predict_auction_share_of_pool_from_out_root(
        out_root=Path(out_root),
        predict_snapshot=str(predict_snapshot),
        train_snapshots=[str(s) for s in train],
        ridge_alpha=float(ridge_alpha),
        feature_set=str(feature_set),
        exclude_entry_names=(
            [exclude_entry_name] if exclude_entry_name else None
        ),
    )

    if (
        str(target_transform).strip().lower() != "none"
        or feature_set == "optimal_v2"
    ):
        train_frames: List[pd.DataFrame] = []
        for s in train:
            df = _read_snapshot_actual_shares(
                out_root=out_root,
                snapshot=str(s),
                exclude_entry_name=str(exclude_entry_name),
            )
            train_frames.append(df)
        train_ds = pd.concat(train_frames, ignore_index=True)

        pred_ds = actual.drop(columns=["team_share_of_pool"], errors="ignore")
        pred = predict_auction_share_of_pool(
            train_team_dataset=train_ds,
            predict_team_dataset=pred_ds,
            ridge_alpha=float(ridge_alpha),
            feature_set=str(feature_set),
            target_transform=str(target_transform),
            seed_prior_monotone=seed_prior_monotone,
            seed_prior_k=float(seed_prior_k),
            program_prior_k=float(program_prior_k),
        )

    merged = actual.merge(
        pred[["team_key", "predicted_auction_share_of_pool"]],
        on="team_key",
        how="inner",
    )

    y = pd.to_numeric(
        merged["team_share_of_pool"],
        errors="coerce",
    ).fillna(0.0)
    yhat = pd.to_numeric(
        merged["predicted_auction_share_of_pool"],
        errors="coerce",
    ).fillna(0.0)

    eps = 1e-12
    mae = float(np.mean(np.abs(y - yhat)))
    rmse = float(np.sqrt(np.mean((y - yhat) ** 2)))
    kl = float(
        np.sum((y + eps) * np.log((y + eps) / (yhat + eps)))
    )

    bands = (
        merged.assign(
            seed_band=merged["seed"].apply(_seed_band),
            err=(yhat - y),
            abs_err=np.abs(yhat - y),
        )
        .groupby("seed_band")
        .agg(
            mean_err=("err", "mean"),
            mae=("abs_err", "mean"),
        )
        .reset_index()
        .sort_values("seed_band")
    )

    top10 = _top_k_overlap(actual=y, pred=yhat, k=10)
    top5 = _top_k_overlap(actual=y, pred=yhat, k=5)

    return {
        "predict_year": int(str(predict_snapshot)),
        "feature_set": str(feature_set),
        "target_transform": str(target_transform),
        "seed_prior_monotone": seed_prior_monotone,
        "seed_prior_k": float(seed_prior_k),
        "program_prior_k": float(program_prior_k),
        "mae": mae,
        "rmse": rmse,
        "kl": kl,
        "top5_overlap": top5,
        "top10_overlap": top10,
        "seed_band": bands.to_dict(orient="records"),
    }


def main() -> int:
    _ensure_project_root_on_path()

    parser = argparse.ArgumentParser(
        description=(
            "Evaluate market-share model variants via leave-one-year-out "
            "(real historical market shares)."
        )
    )
    parser.add_argument(
        "--years",
        default="2017,2018,2019,2021,2022,2023,2024,2025",
        help="Comma-separated years to include",
    )
    parser.add_argument(
        "--out-root",
        default="",
        help=(
            "Optional path to out/ snapshots (avoids DB). If set, years are "
            "interpreted as snapshot names under out_root."
        ),
    )
    parser.add_argument(
        "--excluded-entry-name",
        default="Andrew Copp",
        help="Entry name to exclude from training targets",
    )
    parser.add_argument("--ridge-alpha", type=float, default=1.0)
    parser.add_argument(
        "--variants",
        default="optimal:none,optimal_v2:none,optimal:log,optimal_v2:log",
        help=(
            "Comma-separated variants as feature_set:target_transform, e.g. "
            "optimal:none,optimal_v2:log"
        ),
    )
    parser.add_argument(
        "--seed-prior-monotone",
        default="auto",
        choices=["auto", "true", "false"],
        help="Only used by optimal_v2",
    )
    parser.add_argument("--seed-prior-k", type=float, default=0.0)
    parser.add_argument("--program-prior-k", type=float, default=0.0)
    parser.add_argument(
        "--train-mode",
        default="leave_one_out",
        choices=["leave_one_out", "prior_only"],
    )
    parser.add_argument("--train-years-window", type=int, default=0)

    args = parser.parse_args()

    years = [int(x.strip()) for x in str(args.years).split(",") if x.strip()]
    if len(years) < 2:
        raise ValueError("need at least 2 years")

    variants: List[Tuple[str, str]] = []
    for part in str(args.variants).split(","):
        p = part.strip()
        if not p:
            continue
        if ":" not in p:
            raise ValueError(f"invalid variant: {p}")
        fs, tt = p.split(":", 1)
        variants.append((fs.strip(), tt.strip()))

    seed_prior_monotone: Optional[bool]
    if str(args.seed_prior_monotone) == "auto":
        seed_prior_monotone = None
    elif str(args.seed_prior_monotone) == "true":
        seed_prior_monotone = True
    else:
        seed_prior_monotone = False

    out_root = (
        Path(str(args.out_root)).expanduser()
        if str(args.out_root)
        else None
    )

    skipped_years: Dict[str, str] = {}
    available_years: List[int] = []

    if out_root:
        snaps = _discover_snapshots(out_root=out_root)
        if not snaps:
            raise ValueError(f"no snapshots found under out_root={out_root}")

        for y in years:
            if str(y) in snaps:
                available_years.append(int(y))
            else:
                skipped_years[str(y)] = "snapshot not found"
    else:
        for y in years:
            ok, reason = _probe_year(
                year=int(y),
                exclude_entry_name=str(args.excluded_entry_name),
            )
            if ok:
                available_years.append(int(y))
            else:
                skipped_years[str(y)] = reason

    if len(available_years) < 2:
        raise ValueError(
            "need at least 2 loadable years; "
            f"skipped_years={skipped_years}"
        )

    results: List[Dict[str, Any]] = []
    errors: List[Dict[str, Any]] = []

    for predict_year in available_years:
        if str(args.train_mode) == "prior_only":
            train_years = [
                y
                for y in available_years
                if int(y) < int(predict_year)
            ]
        else:
            train_years = [y for y in available_years if y != predict_year]

        w = int(args.train_years_window or 0)
        if w > 0 and len(train_years) > w:
            train_years = sorted(train_years)[-w:]

        if len(train_years) < 1:
            continue
        for fs, tt in variants:
            try:
                if out_root:
                    res = _metrics_for_snapshot(
                        out_root=Path(out_root),
                        snapshots=[str(y) for y in available_years],
                        predict_snapshot=str(predict_year),
                        exclude_entry_name=str(args.excluded_entry_name),
                        ridge_alpha=float(args.ridge_alpha),
                        feature_set=str(fs),
                        target_transform=str(tt),
                        seed_prior_monotone=seed_prior_monotone,
                        seed_prior_k=float(args.seed_prior_k),
                        program_prior_k=float(args.program_prior_k),
                    )
                else:
                    res = _metrics_for_year(
                        train_years=train_years,
                        predict_year=predict_year,
                        exclude_entry_name=str(args.excluded_entry_name),
                        ridge_alpha=float(args.ridge_alpha),
                        feature_set=str(fs),
                        target_transform=str(tt),
                        seed_prior_monotone=seed_prior_monotone,
                        seed_prior_k=float(args.seed_prior_k),
                        program_prior_k=float(args.program_prior_k),
                    )
                results.append(res)
            except Exception as e:
                errors.append(
                    {
                        "predict_year": int(predict_year),
                        "feature_set": str(fs),
                        "target_transform": str(tt),
                        "error": str(e),
                        "error_type": type(e).__name__,
                    }
                )

    sys.stdout.write(
        json.dumps(
            {
                "available_years": available_years,
                "skipped_years": skipped_years,
                "results": results,
                "errors": errors,
            }
        )
        + "\n"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
