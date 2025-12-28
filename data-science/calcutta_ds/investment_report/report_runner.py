from __future__ import annotations

import json
from pathlib import Path
from typing import Dict, List, Optional

import numpy as np

import predict_market_share

from calcutta_ds.investment_report.year_runner import run_single_year


def _load_kenpom_scale(
    *,
    kenpom_scale: float,
    kenpom_scale_file: Optional[str],
) -> tuple[float, str]:
    scale = float(kenpom_scale)
    source = "arg"
    if kenpom_scale_file:
        p = Path(kenpom_scale_file)
        payload = json.loads(p.read_text(encoding="utf-8"))
        if "kenpom_scale" not in payload:
            raise ValueError("kenpom scale file missing kenpom_scale")
        scale = float(payload["kenpom_scale"])
        source = str(p)
    return scale, source


def build_report(
    *,
    out_root: Path,
    ridge_alpha: float,
    train_mode: str,
    budget: float,
    real_buyin_dollars: float,
    min_teams: int,
    max_teams: int,
    max_per_team: float,
    min_bid: float,
    points_mode: str,
    allocation_mode: str,
    greedy_objective: str,
    greedy_contest_sims: int,
    exclude_entry_names: List[str],
    debug_output: bool,
    expected_sims: int,
    expected_seed: int,
    expected_use_historical_winners: bool,
    kenpom_scale: float,
    kenpom_scale_source: str,
    summary_min_train_snapshots: int,
    artifacts_dir: Optional[str] = None,
    use_cache: bool = False,
    snapshot_filter: Optional[List[str]] = None,
) -> Dict[str, object]:
    snapshot_dirs = predict_market_share._find_snapshots(out_root)
    if (
        not snapshot_dirs
        and (out_root / "derived" / "team_dataset.parquet").exists()
    ):
        snapshot_dirs = [out_root]
    if not snapshot_dirs:
        raise FileNotFoundError(f"no snapshots found under: {out_root}")

    if snapshot_filter:
        allowed = set(str(s) for s in snapshot_filter if str(s).strip())
        snapshot_dirs = [sd for sd in snapshot_dirs if sd.name in allowed]
        if not snapshot_dirs:
            raise FileNotFoundError(
                "no snapshots matched snapshot_filter: "
                + ",".join(sorted(allowed))
            )

    greedy_top_k = 1
    if str(greedy_objective) == "p_top3":
        greedy_top_k = 3
    elif str(greedy_objective) == "p_top6":
        greedy_top_k = 6

    years: List[Dict[str, object]] = []
    for sd in snapshot_dirs:
        years.append(
            run_single_year(
                out_root=out_root,
                snapshot_dir=sd,
                all_snapshot_dirs=snapshot_dirs,
                ridge_alpha=float(ridge_alpha),
                budget=float(budget),
                real_buyin_dollars=float(real_buyin_dollars),
                min_teams=int(min_teams),
                max_teams=int(max_teams),
                max_per_team=float(max_per_team),
                min_bid=float(min_bid),
                points_mode=str(points_mode),
                expected_sims=int(expected_sims),
                expected_seed=int(expected_seed),
                expected_use_historical_winners=bool(
                    expected_use_historical_winners
                ),
                kenpom_scale=float(kenpom_scale),
                kenpom_scale_source=str(kenpom_scale_source),
                allocation_mode=str(allocation_mode),
                greedy_objective=str(greedy_objective),
                greedy_top_k=int(greedy_top_k),
                greedy_contest_sims=int(greedy_contest_sims),
                exclude_entry_names=list(exclude_entry_names or []),
                train_mode=str(train_mode),
                debug_output=bool(debug_output),
                artifacts_dir=str(artifacts_dir) if artifacts_dir else None,
                use_cache=bool(use_cache),
            )
        )

    min_train = int(summary_min_train_snapshots)
    eligible_years = [
        y
        for y in years
        if int(y.get("market_model", {}).get("n_train_snapshots", 0))
        >= min_train
    ]
    eligible_snapshots = set(str(y.get("snapshot")) for y in eligible_years)
    excluded_snapshots = [
        str(y.get("snapshot"))
        for y in years
        if str(y.get("snapshot")) not in eligible_snapshots
    ]

    rois = [float(y.get("realized", {}).get("roi", 0.0)) for y in years]
    rois_real_buyin = [
        float(y.get("realized", {}).get("roi_real_buyin", 0.0))
        for y in years
    ]
    total_points = [
        float(y.get("realized", {}).get("total_points", 0.0)) for y in years
    ]
    points_per_fake_dollar = [
        float(y.get("realized", {}).get("points_per_fake_dollar", 0.0))
        for y in years
    ]
    payout_per_fake_dollar = [
        float(y.get("realized", {}).get("payout_per_fake_dollar", 0.0))
        for y in years
    ]
    payout_cents = [
        int(y.get("realized", {}).get("payout_cents", 0))
        for y in years
    ]

    rois_eligible = [
        float(y.get("realized", {}).get("roi", 0.0))
        for y in eligible_years
    ]
    rois_real_buyin_eligible = [
        float(y.get("realized", {}).get("roi_real_buyin", 0.0))
        for y in eligible_years
    ]
    total_points_eligible = [
        float(y.get("realized", {}).get("total_points", 0.0))
        for y in eligible_years
    ]
    points_per_fake_dollar_eligible = [
        float(y.get("realized", {}).get("points_per_fake_dollar", 0.0))
        for y in eligible_years
    ]
    payout_per_fake_dollar_eligible = [
        float(y.get("realized", {}).get("payout_per_fake_dollar", 0.0))
        for y in eligible_years
    ]
    payout_cents_eligible = [
        int(y.get("realized", {}).get("payout_cents", 0))
        for y in eligible_years
    ]

    report: Dict[str, object] = {
        "snapshots": [sd.name for sd in snapshot_dirs],
        "config": {
            "ridge_alpha": float(ridge_alpha),
            "train_mode": str(train_mode),
            "budget": float(budget),
            "real_buyin_dollars": float(real_buyin_dollars),
            "min_teams": int(min_teams),
            "max_teams": int(max_teams),
            "max_per_team": float(max_per_team),
            "min_bid": float(min_bid),
            "points_mode": str(points_mode),
            "allocation_mode": str(allocation_mode),
            "greedy_objective": str(greedy_objective),
            "greedy_contest_sims": int(greedy_contest_sims),
            "exclude_entry_names": list(exclude_entry_names or []),
            "debug_output": bool(debug_output),
            "expected_sims": int(expected_sims),
            "expected_seed": int(expected_seed),
            "expected_use_historical_winners": bool(
                expected_use_historical_winners
            ),
            "kenpom_scale": float(kenpom_scale),
            "kenpom_scale_source": str(kenpom_scale_source),
            "summary_min_train_snapshots": int(summary_min_train_snapshots),
            "snapshot_filter": list(snapshot_filter) if snapshot_filter else None,
        },
        "summary": {
            "n_years": int(len(years)),
            "mean_roi": float(np.mean(rois)) if rois else 0.0,
            "median_roi": float(np.median(rois)) if rois else 0.0,
            "mean_roi_real_buyin": (
                float(np.mean(rois_real_buyin))
                if rois_real_buyin
                else 0.0
            ),
            "median_roi_real_buyin": (
                float(np.median(rois_real_buyin))
                if rois_real_buyin
                else 0.0
            ),
            "mean_total_points": (
                float(np.mean(total_points)) if total_points else 0.0
            ),
            "median_total_points": (
                float(np.median(total_points)) if total_points else 0.0
            ),
            "mean_points_per_fake_dollar": (
                float(np.mean(points_per_fake_dollar))
                if points_per_fake_dollar
                else 0.0
            ),
            "median_points_per_fake_dollar": (
                float(np.median(points_per_fake_dollar))
                if points_per_fake_dollar
                else 0.0
            ),
            "mean_payout_per_fake_dollar": (
                float(np.mean(payout_per_fake_dollar))
                if payout_per_fake_dollar
                else 0.0
            ),
            "median_payout_per_fake_dollar": (
                float(np.median(payout_per_fake_dollar))
                if payout_per_fake_dollar
                else 0.0
            ),
            "mean_payout_cents": (
                float(np.mean(payout_cents)) if payout_cents else 0.0
            ),
            "median_payout_cents": (
                float(np.median(payout_cents)) if payout_cents else 0.0
            ),
        },
        "summary_filtered": {
            "min_train_snapshots": int(min_train),
            "excluded_snapshots": excluded_snapshots,
            "n_years": int(len(eligible_years)),
            "mean_roi": (
                float(np.mean(rois_eligible)) if rois_eligible else 0.0
            ),
            "median_roi": (
                float(np.median(rois_eligible)) if rois_eligible else 0.0
            ),
            "mean_roi_real_buyin": (
                float(np.mean(rois_real_buyin_eligible))
                if rois_real_buyin_eligible
                else 0.0
            ),
            "median_roi_real_buyin": (
                float(np.median(rois_real_buyin_eligible))
                if rois_real_buyin_eligible
                else 0.0
            ),
            "mean_total_points": (
                float(np.mean(total_points_eligible))
                if total_points_eligible
                else 0.0
            ),
            "median_total_points": (
                float(np.median(total_points_eligible))
                if total_points_eligible
                else 0.0
            ),
            "mean_points_per_fake_dollar": (
                float(np.mean(points_per_fake_dollar_eligible))
                if points_per_fake_dollar_eligible
                else 0.0
            ),
            "median_points_per_fake_dollar": (
                float(np.median(points_per_fake_dollar_eligible))
                if points_per_fake_dollar_eligible
                else 0.0
            ),
            "mean_payout_per_fake_dollar": (
                float(np.mean(payout_per_fake_dollar_eligible))
                if payout_per_fake_dollar_eligible
                else 0.0
            ),
            "median_payout_per_fake_dollar": (
                float(np.median(payout_per_fake_dollar_eligible))
                if payout_per_fake_dollar_eligible
                else 0.0
            ),
            "mean_payout_cents": (
                float(np.mean(payout_cents_eligible))
                if payout_cents_eligible
                else 0.0
            ),
            "median_payout_cents": (
                float(np.median(payout_cents_eligible))
                if payout_cents_eligible
                else 0.0
            ),
        },
        "years": years,
    }

    return report


def write_report(
    *,
    out_root: Path,
    out_path: Optional[str],
    report: Dict[str, object],
) -> Path:
    final_path = Path(out_path) if out_path else (out_root / "report.json")
    final_path.write_text(
        json.dumps(report, indent=2) + "\n",
        encoding="utf-8",
    )
    return final_path


def load_scale_and_build_report(
    *,
    out_root: Path,
    kenpom_scale: float,
    kenpom_scale_file: Optional[str],
    **kwargs: object,
) -> Dict[str, object]:
    scale, source = _load_kenpom_scale(
        kenpom_scale=kenpom_scale,
        kenpom_scale_file=kenpom_scale_file,
    )
    return build_report(
        out_root=out_root,
        kenpom_scale=float(scale),
        kenpom_scale_source=str(source),
        **kwargs,
    )
