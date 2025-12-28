from __future__ import annotations

import json
from pathlib import Path
from typing import Dict, List, Optional

import numpy as np
import pandas as pd

import backtest_scaffold

from calcutta_ds.investment_report.debug_output import build_debug_output
from calcutta_ds.investment_report.expected_points import (
    add_expected_columns,
    compute_team_mean_points,
    points_by_round_from_tables,
)
from calcutta_ds.investment_report.market_bids import filter_market_bids
from calcutta_ds.investment_report.market_model import (
    predict_share_for_snapshot,
)
from calcutta_ds.investment_report.portfolio_allocation import select_portfolio
from calcutta_ds.investment_report.portfolio_contest import (
    optimize_portfolio_greedy_contest,
)
from calcutta_ds.investment_report.portfolio_greedy import (
    optimize_portfolio_greedy,
)
from calcutta_ds.investment_report.portfolio_knapsack import (
    optimize_portfolio_knapsack,
)
from calcutta_ds.investment_report.realized import compute_realized
from calcutta_ds.investment_report.utils import choose_train_snapshots, zscore


def _filter_eligible_teams(
    *,
    pred_df: pd.DataFrame,
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
) -> tuple[pd.DataFrame, Optional[set[str]]]:
    eligible_team_keys: Optional[set[str]] = None

    teams = tables.get("teams")
    required = {
        "team_key",
        "wins",
        "byes",
    }
    if teams is not None and required.issubset(set(teams.columns)):
        t = teams.copy()
        if "calcutta_key" in t.columns:
            t = t[t["calcutta_key"] == calcutta_key].copy()
        t["wins"] = (
            pd.to_numeric(t["wins"], errors="coerce")
            .fillna(0)
            .astype(int)
        )
        t["byes"] = (
            pd.to_numeric(t["byes"], errors="coerce")
            .fillna(0)
            .astype(int)
        )
        eligible_team_keys = set(
            t[(t["wins"] != 0) | (t["byes"] != 0)]["team_key"]
            .astype(str)
            .tolist()
        )

        pred_keys = set(pred_df["team_key"].astype(str).tolist())
        missing = sorted(eligible_team_keys - pred_keys)
        if missing:
            add_cols = set(pred_df.columns)
            add_rows: List[Dict[str, object]] = []
            t2 = t[t["team_key"].astype(str).isin(missing)].copy()
            for _, r in t2.iterrows():
                row: Dict[str, object] = {c: np.nan for c in add_cols}
                row["team_key"] = str(r.get("team_key"))
                if "calcutta_key" in add_cols:
                    row["calcutta_key"] = str(calcutta_key)
                if "school_name" in add_cols:
                    row["school_name"] = r.get("school_name")
                if "seed" in add_cols:
                    row["seed"] = r.get("seed")
                if "region" in add_cols:
                    row["region"] = r.get("region")
                if "kenpom_net" in add_cols:
                    row["kenpom_net"] = r.get("kenpom_net")
                if "predicted_team_share_of_pool" in add_cols:
                    row["predicted_team_share_of_pool"] = 0.0
                add_rows.append(row)
            if add_rows:
                pred_df = pd.concat(
                    [pred_df, pd.DataFrame(add_rows)],
                    ignore_index=True,
                )

        pred_df = pred_df[
            pred_df["team_key"].astype(str).isin(eligible_team_keys)
        ].copy()

        ps = pd.to_numeric(
            pred_df["predicted_team_share_of_pool"],
            errors="coerce",
        ).fillna(0.0)
        s = float(ps.sum())
        if s > 0:
            pred_df["predicted_team_share_of_pool"] = ps / s
        else:
            pred_df["predicted_team_share_of_pool"] = 1.0 / float(len(pred_df))

    return pred_df, eligible_team_keys


def _eligible_entries(
    *,
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    exclude_entry_names: List[str],
) -> pd.DataFrame:
    eligible_entries = tables.get("entries")
    if eligible_entries is None:
        raise ValueError("snapshot missing entries.parquet")

    eligible_entries = eligible_entries[
        eligible_entries["calcutta_key"] == calcutta_key
    ].copy()

    if exclude_entry_names and "entry_name" in eligible_entries.columns:
        ex = [
            str(n).strip().lower()
            for n in exclude_entry_names
            if str(n).strip()
        ]
        if ex:
            eligible_entries["_name_norm"] = (
                eligible_entries["entry_name"]
                .astype(str)
                .str.strip()
                .str.lower()
            )
            mask = eligible_entries["_name_norm"].apply(
                lambda n: any(x in n for x in ex)
            )
            eligible_entries = eligible_entries[~mask].copy()

    return eligible_entries


def _map_greedy_objective(
    greedy_objective: str,
    greedy_top_k: int,
) -> tuple[str, int]:
    mapped_obj = str(greedy_objective)
    mapped_k = int(greedy_top_k)
    if mapped_obj in ("p_top3", "p_top6"):
        mapped_k = 3 if mapped_obj == "p_top3" else 6
        mapped_obj = "p_topk"
    return mapped_obj, mapped_k


def _build_payout_map(*, payouts: pd.DataFrame) -> Dict[int, int]:
    req = {"position", "amount_cents"}
    if payouts is None or payouts.empty or not req.issubset(set(payouts.columns)):
        raise ValueError("payouts table missing required columns")

    p = payouts.copy()
    p["position"] = pd.to_numeric(p["position"], errors="coerce")
    p["amount_cents"] = pd.to_numeric(p["amount_cents"], errors="coerce")
    p = p[p["position"].notna() & p["amount_cents"].notna()].copy()

    payout_map: Dict[int, int] = {}
    for _, r in p.iterrows():
        payout_map[int(r["position"])] = int(r["amount_cents"])
    if not payout_map:
        raise ValueError("payouts table produced empty payout_map")
    return payout_map


def _load_payout_map(
    *,
    out_root: Path,
    all_snapshot_dirs: List[Path],
    payout_snapshot: Optional[str],
    fallback_tables: Dict[str, pd.DataFrame],
) -> Dict[int, int]:
    if payout_snapshot:
        name = str(payout_snapshot)
        sd = next((p for p in all_snapshot_dirs if p.name == name), None)
        if sd is None:
            raise FileNotFoundError(f"payout_snapshot not found: {name}")
        ptables = backtest_scaffold._load_snapshot_tables(sd)
        payouts = ptables.get("payouts")
        if payouts is None:
            raise ValueError(f"payout_snapshot missing payouts.parquet: {name}")
        return _build_payout_map(payouts=payouts)

    payouts = fallback_tables.get("payouts")
    if payouts is None:
        raise ValueError("snapshot missing payouts.parquet")
    return _build_payout_map(payouts=payouts)


def run_single_year(
    *,
    out_root: Path,
    snapshot_dir: Path,
    all_snapshot_dirs: List[Path],
    ridge_alpha: float,
    budget: float,
    real_buyin_dollars: float,
    min_teams: int,
    max_teams: int,
    max_per_team: float,
    min_bid: float,
    points_mode: str,
    expected_sims: int,
    expected_seed: int,
    expected_use_historical_winners: bool,
    kenpom_scale: float,
    kenpom_scale_source: str,
    allocation_mode: str,
    greedy_objective: str,
    greedy_top_k: int,
    greedy_contest_sims: int,
    exclude_entry_names: List[str],
    train_mode: str,
    debug_output: bool,
    payout_snapshot: Optional[str] = None,
    payout_utility: str = "power",
    payout_utility_gamma: float = 1.2,
    payout_utility_alpha: float = 1.0,
    artifacts_dir: Optional[str] = None,
    use_cache: bool = False,
) -> Dict[str, object]:
    snapshot = snapshot_dir.name

    artifact_root: Optional[Path] = None
    if artifacts_dir:
        artifact_root = Path(artifacts_dir) / str(snapshot)
        artifact_root.mkdir(parents=True, exist_ok=True)

    def _write_json(p: Path, obj: object) -> None:
        p.write_text(json.dumps(obj, indent=2) + "\n", encoding="utf-8")

    def _meta_matches(p: Path, meta: Dict[str, object]) -> bool:
        if not p.exists():
            return False
        try:
            cur = json.loads(p.read_text(encoding="utf-8"))
        except Exception:
            return False
        if not isinstance(cur, dict):
            return False
        for k, v in meta.items():
            if cur.get(k) != v:
                return False
        return True
    train_snapshots = choose_train_snapshots(
        all_snapshot_dirs=all_snapshot_dirs,
        predict_snapshot=snapshot,
        train_mode=str(train_mode),
    )

    pred_df: pd.DataFrame
    pred_cache_meta = {
        "snapshot": str(snapshot),
        "ridge_alpha": float(ridge_alpha),
        "train_mode": str(train_mode),
        "train_snapshots": list(train_snapshots),
        "exclude_entry_names": list(exclude_entry_names or []),
    }

    pred_cache_meta_path: Optional[Path] = None
    pred_cache_data_path: Optional[Path] = None
    if artifact_root is not None:
        pred_cache_meta_path = artifact_root / "predicted_market_meta.json"
        pred_cache_data_path = artifact_root / "predicted_market.json"

    if (
        bool(use_cache)
        and pred_cache_meta_path is not None
        and pred_cache_data_path is not None
        and _meta_matches(pred_cache_meta_path, pred_cache_meta)
        and pred_cache_data_path.exists()
    ):
        payload = json.loads(pred_cache_data_path.read_text(encoding="utf-8"))
        if not isinstance(payload, list):
            raise ValueError("invalid predicted_market cache payload")
        pred_df = pd.DataFrame(payload)
    else:
        pred_df = predict_share_for_snapshot(
            out_root=out_root,
            snapshot_dirs=all_snapshot_dirs,
            predict_snapshot=snapshot,
            train_snapshots=train_snapshots,
            ridge_alpha=ridge_alpha,
            exclude_entry_names=exclude_entry_names,
        )
        if pred_cache_meta_path is not None and pred_cache_data_path is not None:
            _write_json(pred_cache_meta_path, pred_cache_meta)
            _write_json(pred_cache_data_path, pred_df.to_dict(orient="records"))

    if "calcutta_key" not in pred_df.columns:
        raise ValueError("team_dataset missing calcutta_key")

    calcutta_key = backtest_scaffold._choose_calcutta_key(pred_df, None)
    pred_df = pred_df[pred_df["calcutta_key"] == calcutta_key].copy()

    tables = backtest_scaffold._load_snapshot_tables(snapshot_dir)

    pred_df, eligible_team_keys = _filter_eligible_teams(
        pred_df=pred_df,
        tables=tables,
        calcutta_key=calcutta_key,
    )

    if artifact_root is not None:
        _write_json(
            artifact_root / "predicted_market_stage.json",
            pred_df.to_dict(orient="records"),
        )

    if artifact_root is not None:
        _write_json(
            artifact_root / "stage_meta.json",
            {
                "snapshot": str(snapshot),
                "calcutta_key": str(calcutta_key),
                "train_snapshots": list(train_snapshots),
                "exclude_entry_names": list(exclude_entry_names or []),
                "ridge_alpha": float(ridge_alpha),
                "train_mode": str(train_mode),
                "budget": float(budget),
                "min_teams": int(min_teams),
                "max_teams": int(max_teams),
                "min_bid": float(min_bid),
                "max_per_team": float(max_per_team),
                "points_mode": str(points_mode),
                "expected_sims": int(expected_sims),
                "expected_seed": int(expected_seed),
                "expected_use_historical_winners": bool(
                    expected_use_historical_winners
                ),
                "kenpom_scale": float(kenpom_scale),
                "kenpom_scale_source": str(kenpom_scale_source),
                "allocation_mode": str(allocation_mode),
                "greedy_objective": str(greedy_objective),
            },
        )

    pred_df["kenpom_net"] = pd.to_numeric(
        pred_df["kenpom_net"],
        errors="coerce",
    )
    pred_df["predicted_team_share_of_pool"] = pd.to_numeric(
        pred_df["predicted_team_share_of_pool"],
        errors="coerce",
    )

    usable = pred_df[
        pred_df["team_key"].notna()
        & pred_df["kenpom_net"].notna()
        & pred_df["predicted_team_share_of_pool"].notna()
    ].copy()

    usable["kenpom_z"] = zscore(usable["kenpom_net"])
    usable["share_z"] = zscore(usable["predicted_team_share_of_pool"])
    usable["value_score"] = usable["kenpom_z"] - usable["share_z"]

    score_label = "kenpom_z - predicted_share_z"

    market_bids = filter_market_bids(
        tables=tables,
        calcutta_key=calcutta_key,
        exclude_entry_names=exclude_entry_names,
    )
    if eligible_team_keys is not None and not market_bids.empty:
        market_bids = market_bids[
            market_bids["team_key"].astype(str).isin(eligible_team_keys)
        ].copy()

    eligible_entries = _eligible_entries(
        tables=tables,
        calcutta_key=calcutta_key,
        exclude_entry_names=exclude_entry_names,
    )

    n_entries = int(eligible_entries["entry_key"].nunique())
    predicted_total_pool_bids = float(n_entries) * float(budget)
    competitor_entry_keys = [
        str(k)
        for k in eligible_entries["entry_key"].astype(str).unique().tolist()
    ]

    pmode = points_mode
    if pmode == "auto":
        pmode = "round_scoring" if "round_scoring" in tables else "fixed"

    if allocation_mode in ("expected_points", "greedy", "knapsack"):
        if not expected_sims or expected_sims <= 0:
            raise ValueError("allocation_mode requires --expected-sims > 0")

        points_by_round = points_by_round_from_tables(
            tables=tables,
            calcutta_key=calcutta_key,
            points_mode=pmode,
        )

        team_mean_points: Dict[str, float]
        tmp_cache_meta = {
            "snapshot": str(snapshot),
            "calcutta_key": str(calcutta_key),
            "points_mode": str(pmode),
            "expected_sims": int(expected_sims),
            "expected_seed": int(expected_seed),
            "kenpom_scale": float(kenpom_scale),
            "budget": float(budget),
            "expected_use_historical_winners": bool(
                expected_use_historical_winners
            ),
            "exclude_entry_names": list(exclude_entry_names or []),
        }

        tmp_cache_meta_path: Optional[Path] = None
        tmp_cache_data_path: Optional[Path] = None
        if artifact_root is not None:
            tmp_cache_meta_path = artifact_root / "team_mean_points_meta.json"
            tmp_cache_data_path = artifact_root / "team_mean_points.json"

        if (
            bool(use_cache)
            and tmp_cache_meta_path is not None
            and tmp_cache_data_path is not None
            and _meta_matches(tmp_cache_meta_path, tmp_cache_meta)
            and tmp_cache_data_path.exists()
        ):
            payload = json.loads(tmp_cache_data_path.read_text(encoding="utf-8"))
            if not isinstance(payload, dict):
                raise ValueError("invalid team_mean_points cache payload")
            team_mean_points = {str(k): float(v) for k, v in payload.items()}
        else:
            team_mean_points = compute_team_mean_points(
                tables=tables,
                calcutta_key=calcutta_key,
                market_bids=market_bids,
                points_mode=pmode,
                points_by_round=points_by_round,
                expected_sims=expected_sims,
                expected_seed=expected_seed,
                kenpom_scale=kenpom_scale,
                budget=budget,
                expected_use_historical_winners=bool(
                    expected_use_historical_winners
                ),
                competitor_entry_keys=competitor_entry_keys,
                any_team_key=str(usable.iloc[0]["team_key"]),
            )
            if tmp_cache_meta_path is not None and tmp_cache_data_path is not None:
                _write_json(tmp_cache_meta_path, tmp_cache_meta)
                _write_json(tmp_cache_data_path, team_mean_points)

        usable = add_expected_columns(
            usable=usable,
            predicted_total_pool_bids=predicted_total_pool_bids,
            team_mean_points=team_mean_points,
        )

        if artifact_root is not None:
            score_cols = [
                c
                for c in [
                    "team_key",
                    "school_name",
                    "seed",
                    "region",
                    "expected_team_points",
                    "expected_points_share",
                ]
                if c in usable.columns
            ]
            inv_cols = [
                c
                for c in [
                    "team_key",
                    "school_name",
                    "seed",
                    "region",
                    "predicted_team_share_of_pool",
                    "predicted_team_total_bids",
                ]
                if c in usable.columns
            ]

            usable[score_cols].to_csv(
                artifact_root / "scores.csv",
                index=False,
            )
            usable[inv_cols].to_csv(
                artifact_root / "predicted_investments.csv",
                index=False,
            )

            roi_df = usable.copy()
            roi_df["predicted_total_team_bids"] = pd.to_numeric(
                roi_df.get("predicted_team_total_bids"),
                errors="coerce",
            ).fillna(0.0)
            roi_df["predicted_ownership_at_min_bid"] = roi_df.apply(
                lambda r: (
                    float(min_bid)
                    / (
                        float(r.get("predicted_total_team_bids") or 0.0)
                        + float(min_bid)
                    )
                    if (
                        float(r.get("predicted_total_team_bids") or 0.0)
                        + float(min_bid)
                    )
                    > 0
                    else 0.0
                ),
                axis=1,
            )
            roi_df["predicted_points_at_min_bid"] = roi_df.apply(
                lambda r: float(r.get("expected_team_points") or 0.0)
                * float(r.get("predicted_ownership_at_min_bid") or 0.0),
                axis=1,
            )
            roi_df["predicted_points_per_dollar_at_min_bid"] = roi_df.apply(
                lambda r: (
                    float(r.get("predicted_points_at_min_bid") or 0.0)
                    / float(min_bid)
                    if float(min_bid) > 0
                    else 0.0
                ),
                axis=1,
            )

            roi_cols = [
                c
                for c in [
                    "team_key",
                    "school_name",
                    "seed",
                    "region",
                    "expected_team_points",
                    "predicted_team_share_of_pool",
                    "predicted_total_team_bids",
                    "predicted_ownership_at_min_bid",
                    "predicted_points_at_min_bid",
                    "predicted_points_per_dollar_at_min_bid",
                    "value_score",
                ]
                if c in roi_df.columns
            ]
            roi_df = roi_df.sort_values(
                by="predicted_points_per_dollar_at_min_bid",
                ascending=False,
            )
            roi_df[roi_cols].to_csv(
                artifact_root / "predicted_roi.csv",
                index=False,
            )

    usable["expected_team_points"] = pd.to_numeric(
        usable["expected_team_points"],
        errors="coerce",
    ).fillna(0.0)
    usable["predicted_team_share_of_pool"] = pd.to_numeric(
        usable["predicted_team_share_of_pool"],
        errors="coerce",
    ).fillna(0.0)
    usable["predicted_team_total_bids"] = pd.to_numeric(
        usable.get("predicted_team_total_bids"),
        errors="coerce",
    ).fillna(0.0)

    total_exp = float(usable["expected_team_points"].sum())
    if total_exp > 0:
        usable["expected_points_share"] = (
            usable["expected_team_points"] / total_exp
        )
    else:
        usable["expected_points_share"] = 0.0

    def _ratio_row(r: pd.Series) -> float:
        denom = float(r.get("predicted_team_share_of_pool") or 0.0)
        if denom <= 0:
            return float("inf")
        return float(r.get("expected_points_share") or 0.0) / denom

    usable["value_ratio"] = usable.apply(_ratio_row, axis=1)

    def _ppd_row(r: pd.Series) -> float:
        exp_pts = float(r.get("expected_team_points") or 0.0)
        m = float(r.get("predicted_team_total_bids") or 0.0)
        if m < 0:
            m = 0.0
        denom = m + float(min_bid)
        return (exp_pts / denom) if denom > 0 else 0.0

    usable["value_score"] = usable.apply(_ppd_row, axis=1)
    score_label = "expected_points_per_dollar_at_min_bid"

    contest_greedy_trace: Optional[List[Dict[str, object]]] = None

    if allocation_mode == "greedy":
        if str(greedy_objective) == "expected_points":
            chosen, portfolio_rows = optimize_portfolio_greedy(
                df=usable,
                score_col="value_score",
                budget=float(budget),
                min_teams=int(min_teams),
                max_teams=int(max_teams),
                max_per_team=float(max_per_team),
                min_bid=float(min_bid),
            )
        else:
            sim_n = (
                int(greedy_contest_sims)
                if int(greedy_contest_sims) > 0
                else int(expected_sims)
            )

            team_keys_s, team_pts_s = (
                backtest_scaffold._simulate_team_points_scenarios(
                    tables=tables,
                    calcutta_key=calcutta_key,
                    points_mode=pmode,
                    points_by_round=points_by_round,
                    n_sims=int(sim_n),
                    seed=int(expected_seed),
                    kenpom_scale=float(kenpom_scale),
                    use_historical_winners=bool(
                        expected_use_historical_winners
                    ),
                )
            )

            entry_keys = [str(k) for k in competitor_entry_keys]
            e_index = {k: i for i, k in enumerate(entry_keys)}
            t_index = {str(tk): i for i, tk in enumerate(team_keys_s)}
            market_entry_bids = np.zeros(
                (len(entry_keys), len(team_keys_s)),
                dtype=float,
            )
            market_team_totals = np.zeros((len(team_keys_s),), dtype=float)

            mb = market_bids.copy()
            mb["entry_key"] = mb["entry_key"].astype(str)
            mb["team_key"] = mb["team_key"].astype(str)
            mb["bid_amount"] = (
                pd.to_numeric(mb["bid_amount"], errors="coerce")
                .fillna(0.0)
            )

            for _, r in mb.iterrows():
                ek = str(r.get("entry_key"))
                tk = str(r.get("team_key"))
                if ek not in e_index or tk not in t_index:
                    continue
                amt = float(r.get("bid_amount") or 0.0)
                if amt <= 0:
                    continue
                market_entry_bids[e_index[ek], t_index[tk]] += amt
                market_team_totals[t_index[tk]] += amt

            mapped_obj, mapped_k = _map_greedy_objective(
                greedy_objective,
                greedy_top_k,
            )

            payout_map = None
            if mapped_obj in ("expected_payout", "expected_utility_payout"):
                payout_map = _load_payout_map(
                    out_root=out_root,
                    all_snapshot_dirs=all_snapshot_dirs,
                    payout_snapshot=payout_snapshot,
                    fallback_tables=tables,
                )

            if bool(debug_output):
                contest_greedy_trace = []

            chosen, portfolio_rows = optimize_portfolio_greedy_contest(
                df=usable,
                budget=float(budget),
                min_teams=int(min_teams),
                max_teams=int(max_teams),
                max_per_team=float(max_per_team),
                min_bid=float(min_bid),
                objective=mapped_obj,
                top_k=mapped_k,
                payout_map=payout_map,
                utility=str(payout_utility),
                utility_gamma=float(payout_utility_gamma),
                utility_alpha=float(payout_utility_alpha),
                team_keys=list(team_keys_s),
                team_points_scenarios=np.asarray(team_pts_s, dtype=float),
                market_entry_bids=market_entry_bids,
                market_team_totals=market_team_totals,
                trace_out=contest_greedy_trace,
            )
    elif allocation_mode == "knapsack":
        chosen, portfolio_rows = optimize_portfolio_knapsack(
            df=usable,
            score_col="value_score",
            budget=float(budget),
            min_teams=int(min_teams),
            max_teams=int(max_teams),
            max_per_team=float(max_per_team),
            min_bid=float(min_bid),
        )
    else:
        chosen, portfolio_rows = select_portfolio(
            df=usable,
            score_col="value_score",
            budget=float(budget),
            min_teams=min_teams,
            max_teams=max_teams,
            max_per_team=max_per_team,
            min_bid=min_bid,
            allocation_mode=str(allocation_mode),
        )

    realized, standings, ctx_df = compute_realized(
        tables=tables,
        calcutta_key=calcutta_key,
        points_mode=pmode,
        market_bids=market_bids,
        portfolio_rows=portfolio_rows,
        competitor_entry_keys=competitor_entry_keys,
        budget=budget,
        real_buyin_dollars=real_buyin_dollars,
    )

    if artifact_root is not None:
        standings.to_csv(
            artifact_root / "standings.csv",
            index=False,
        )
        _write_json(
            artifact_root / "realized.json",
            realized,
        )

    portfolio_out = chosen[
        [
            c
            for c in [
                "team_key",
                "school_name",
                "seed",
                "region",
                "kenpom_net",
                "predicted_team_share_of_pool",
                "value_score",
                "bid_amount",
            ]
            if c in chosen.columns
        ]
    ].to_dict(orient="records")

    if artifact_root is not None:
        _write_json(
            artifact_root / "portfolio.json",
            portfolio_out,
        )

    year_out: Dict[str, object] = {
        "snapshot": snapshot,
        "calcutta_key": calcutta_key,
        "market_model": {
            "ridge_alpha": float(ridge_alpha),
            "train_mode": str(train_mode),
            "train_snapshots": train_snapshots,
            "n_train_snapshots": int(len(train_snapshots)),
            "exclude_entry_names": list(exclude_entry_names or []),
        },
        "portfolio": portfolio_out,
        "strategy": {
            "score": str(score_label),
            "constraints": {
                "budget": float(budget),
                "min_teams": int(min_teams),
                "max_teams": int(max_teams),
                "max_per_team": float(max_per_team),
                "min_bid": float(min_bid),
            },
            "allocation_mode": str(allocation_mode),
            "exclude_entry_names": list(exclude_entry_names),
        },
        "realized": realized,
    }

    if expected_sims and expected_sims > 0:
        buyin_cents = int(ctx_df.iloc[0]["buyin_cents"])

        exp = backtest_scaffold._expected_simulation(
            tables=tables,
            calcutta_key=calcutta_key,
            market_bids=market_bids,
            sim_rows=pd.DataFrame(
                {
                    "calcutta_key": [calcutta_key for _ in portfolio_rows],
                    "entry_key": ["simulated:entry" for _ in portfolio_rows],
                    "team_key": [r["team_key"] for r in portfolio_rows],
                    "bid_amount": [r["bid_amount"] for r in portfolio_rows],
                }
            ),
            sim_entry_key="simulated:entry",
            points_mode=pmode,
            points_by_round=points_by_round_from_tables(
                tables=tables,
                calcutta_key=calcutta_key,
                points_mode=pmode,
            ),
            n_sims=int(expected_sims),
            seed=int(expected_seed),
            kenpom_scale=float(kenpom_scale),
            budget=float(budget),
            use_historical_winners=bool(expected_use_historical_winners),
            competitor_entry_keys=competitor_entry_keys,
        )

        if exp:
            exp["kenpom_scale_source"] = str(kenpom_scale_source)
            exp["real_buyin_dollars"] = float(real_buyin_dollars)
            exp["points_per_fake_dollar_mean"] = (
                float(exp.get("mean_total_points") or 0.0) / float(budget)
                if budget
                else 0.0
            )
            exp["mean_roi_real_buyin"] = (
                float(exp.get("mean_payout_cents") or 0.0) / float(buyin_cents)
                if buyin_cents
                else 0.0
            )
            exp["p50_roi_real_buyin"] = (
                float(exp.get("p50_payout_cents") or 0.0) / float(buyin_cents)
                if buyin_cents
                else 0.0
            )
            exp["p90_roi_real_buyin"] = (
                float(exp.get("p90_payout_cents") or 0.0) / float(buyin_cents)
                if buyin_cents
                else 0.0
            )
            year_out["expected"] = exp

            if artifact_root is not None:
                _write_json(
                    artifact_root / "expected.json",
                    exp,
                )

    if debug_output:
        year_out["debug"] = build_debug_output(
            usable=usable,
            chosen=chosen,
            market_bids=market_bids,
            min_bid=float(min_bid),
            n_entries=int(n_entries),
            predicted_total_pool_bids=float(predicted_total_pool_bids),
        )

        if contest_greedy_trace is not None:
            year_out["debug"]["contest_greedy_trace"] = contest_greedy_trace

    return year_out
