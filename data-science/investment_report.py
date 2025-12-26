import argparse
import json
from pathlib import Path
from typing import Dict, List, Optional, Tuple

import numpy as np
import pandas as pd

import backtest_scaffold
import predict_market_share


def _zscore(s: pd.Series) -> pd.Series:
    v = pd.to_numeric(s, errors="coerce")
    mu = float(v.mean()) if v.notna().any() else 0.0
    sd = float(v.std(ddof=0)) if v.notna().any() else 0.0
    if sd <= 0:
        return v * 0.0
    return (v - mu) / sd


def _parse_snapshot_year(name: str) -> Optional[int]:
    try:
        return int(str(name).strip())
    except Exception:
        return None


def _choose_train_snapshots(
    all_snapshot_dirs: List[Path],
    predict_snapshot: str,
    train_mode: str,
) -> List[str]:
    if train_mode == "loo":
        return [
            s.name
            for s in all_snapshot_dirs
            if s.name != predict_snapshot
        ]

    if train_mode != "past_only":
        raise ValueError(f"unknown train_mode: {train_mode}")

    y_pred = _parse_snapshot_year(predict_snapshot)
    if y_pred is None:
        return [
            s.name
            for s in all_snapshot_dirs
            if s.name != predict_snapshot
        ]

    out: List[str] = []
    for s in all_snapshot_dirs:
        if s.name == predict_snapshot:
            continue
        ys = _parse_snapshot_year(s.name)
        if ys is None:
            continue
        if ys < y_pred:
            out.append(s.name)
    return sorted(out)


def _filter_market_bids(
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    exclude_entry_names: List[str],
) -> pd.DataFrame:
    market_bids = tables["entry_bids"].copy()
    market_bids = market_bids[
        market_bids["calcutta_key"] == calcutta_key
    ].copy()
    if not exclude_entry_names:
        return market_bids

    entries = tables.get("entries")
    if entries is None or "entry_key" not in entries.columns:
        return market_bids
    if "entry_name" not in entries.columns:
        return market_bids

    e = entries.copy()
    e = e[e["calcutta_key"] == calcutta_key].copy()
    e["entry_name"] = e["entry_name"].astype(str)
    exclude_norm = [
        str(n).strip().lower()
        for n in exclude_entry_names
        if str(n).strip()
    ]
    if not exclude_norm:
        return market_bids

    def _matches(name: str) -> bool:
        n = str(name).strip().lower()
        return any(x in n for x in exclude_norm)

    excluded_keys = set(
        e[e["entry_name"].apply(_matches)]["entry_key"].astype(str).tolist()
    )
    if not excluded_keys:
        return market_bids
    return market_bids[
        ~market_bids["entry_key"].astype(str).isin(excluded_keys)
    ].copy()


def _compute_team_shares_from_bids(
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    exclude_entry_names: List[str],
) -> Dict[str, float]:
    bids = _filter_market_bids(
        tables=tables,
        calcutta_key=calcutta_key,
        exclude_entry_names=exclude_entry_names,
    )
    bids = bids[bids["calcutta_key"] == calcutta_key].copy()
    bids["bid_amount"] = (
        pd.to_numeric(bids["bid_amount"], errors="coerce")
        .fillna(0.0)
    )

    teams = tables.get("teams")
    required = {
        "team_key",
        "wins",
        "byes",
        "calcutta_key",
    }
    if teams is not None and required.issubset(set(teams.columns)):
        t = teams[teams["calcutta_key"] == calcutta_key].copy()
        t["wins"] = (
            pd.to_numeric(
                t["wins"],
                errors="coerce",
            )
            .fillna(0)
            .astype(int)
        )
        t["byes"] = (
            pd.to_numeric(
                t["byes"],
                errors="coerce",
            )
            .fillna(0)
            .astype(int)
        )
        eligible_team_keys = set(
            t[(t["wins"] != 0) | (t["byes"] != 0)]["team_key"]
            .astype(str)
            .tolist()
        )
        bids = bids[
            bids["team_key"].astype(str).isin(eligible_team_keys)
        ].copy()

    totals = bids.groupby("team_key")["bid_amount"].sum()
    denom = float(totals.sum())
    if denom <= 0:
        return {}
    return {str(k): float(v) / denom for k, v in totals.items()}


def _select_portfolio(
    df: pd.DataFrame,
    score_col: str,
    budget: float,
    min_teams: int,
    max_teams: int,
    max_per_team: float,
    min_bid: float,
    allocation_mode: str,
) -> Tuple[pd.DataFrame, List[Dict[str, object]]]:
    if min_teams <= 0 or max_teams <= 0:
        raise ValueError("min_teams and max_teams must be positive")
    if min_teams > max_teams:
        raise ValueError("min_teams cannot exceed max_teams")
    if budget <= 0:
        raise ValueError("budget must be positive")
    if min_bid <= 0:
        raise ValueError("min_bid must be positive")

    max_k_by_budget = int(budget // min_bid)
    max_k = min(max_teams, max_k_by_budget, int(len(df)))
    if max_k < min_teams:
        raise ValueError("budget too small to satisfy min_teams at min_bid")

    k = max_k
    ranked = df.sort_values(by=score_col, ascending=False)
    chosen = ranked.head(k).copy().reset_index(drop=True)

    if allocation_mode == "equal":
        bids = backtest_scaffold._waterfill_equal(
            k=k,
            budget=budget,
            max_per_team=max_per_team,
        )
        if any(b < min_bid for b in bids):
            raise ValueError("allocation violates min_bid constraint")
        chosen["bid_amount"] = bids
    elif allocation_mode == "expected_points":
        required_cols = ["expected_team_points", "predicted_team_total_bids"]
        missing = [c for c in required_cols if c not in chosen.columns]
        if missing:
            raise ValueError(
                "allocation_mode=expected_points requires columns: "
                + ", ".join(missing)
            )
        bids = _allocate_expected_points(
            df=chosen,
            budget=float(budget),
            min_bid=float(min_bid),
            max_per_team=float(max_per_team),
        )
        chosen["bid_amount"] = bids
    else:
        raise ValueError(f"unknown allocation_mode: {allocation_mode}")

    portfolio_rows: List[Dict[str, object]] = []
    for _, r in chosen.iterrows():
        portfolio_rows.append(
            {
                "team_key": str(r["team_key"]),
                "bid_amount": float(r["bid_amount"]),
                "score": float(r[score_col]),
            }
        )

    return chosen, portfolio_rows


def _allocate_expected_points(
    df: pd.DataFrame,
    budget: float,
    min_bid: float,
    max_per_team: float,
    step: float = 0.25,
) -> List[float]:
    if budget <= 0:
        return []
    if min_bid <= 0:
        raise ValueError("min_bid must be positive")
    if max_per_team <= 0:
        raise ValueError("max_per_team must be positive")
    if step <= 0:
        raise ValueError("step must be positive")

    k = int(len(df))
    if k <= 0:
        return []

    base_spend = float(k) * float(min_bid)
    if base_spend - budget > 1e-9:
        raise ValueError("budget too small for min_bid across selected teams")

    bids: List[float] = [float(min_bid) for _ in range(k)]
    remaining = float(budget) - base_spend

    exp_pts = (
        pd.to_numeric(df["expected_team_points"], errors="coerce")
        .fillna(0.0)
        .tolist()
    )
    market_totals = (
        pd.to_numeric(df["predicted_team_total_bids"], errors="coerce")
        .fillna(0.0)
        .tolist()
    )

    while remaining > 1e-9:
        best_i: Optional[int] = None
        best_val = -1e99
        for i in range(k):
            if bids[i] + step - max_per_team > 1e-9:
                continue
            m = float(market_totals[i])
            if m < 0:
                m = 0.0
            b0 = float(bids[i])
            b1 = float(b0 + step)

            denom0 = m + b0
            denom1 = m + b1
            s0 = (b0 / denom0) if denom0 > 0 else 0.0
            s1 = (b1 / denom1) if denom1 > 0 else 0.0
            delta = float(exp_pts[i]) * (s1 - s0)
            v = (delta / float(step)) if step > 0 else 0.0
            if v > best_val:
                best_val = v
                best_i = i

        if best_i is None:
            break

        inc = step if remaining >= step else remaining
        if bids[best_i] + inc - max_per_team > 1e-9:
            inc = max(0.0, max_per_team - bids[best_i])
        if inc <= 1e-12:
            break

        bids[best_i] += float(inc)
        remaining -= float(inc)

    return bids


def _optimize_portfolio_greedy(
    df: pd.DataFrame,
    score_col: str,
    budget: float,
    min_teams: int,
    max_teams: int,
    max_per_team: float,
    min_bid: float,
    step: float = 0.25,
) -> Tuple[pd.DataFrame, List[Dict[str, object]]]:
    required_cols = [
        "team_key",
        "expected_team_points",
        "predicted_team_total_bids",
    ]
    missing = [c for c in required_cols if c not in df.columns]
    if missing:
        raise ValueError(
            "greedy optimizer requires columns: " + ", ".join(missing)
        )
    if budget <= 0:
        raise ValueError("budget must be positive")
    if min_teams <= 0 or max_teams <= 0:
        raise ValueError("min_teams and max_teams must be positive")
    if min_teams > max_teams:
        raise ValueError("min_teams cannot exceed max_teams")
    if min_bid <= 0:
        raise ValueError("min_bid must be positive")
    if step <= 0:
        raise ValueError("step must be positive")

    pool = df.copy().reset_index(drop=True)
    pool["expected_team_points"] = pd.to_numeric(
        pool["expected_team_points"], errors="coerce"
    ).fillna(0.0)
    pool["predicted_team_total_bids"] = pd.to_numeric(
        pool["predicted_team_total_bids"], errors="coerce"
    ).fillna(0.0)

    n = int(len(pool))
    if n == 0:
        return pool, []

    if float(min_teams) * float(min_bid) - float(budget) > 1e-9:
        raise ValueError("budget too small to satisfy min_teams at min_bid")

    bids: List[float] = [0.0 for _ in range(n)]
    selected: set[int] = set()

    def _delta_for(i: int, b0: float, inc: float) -> float:
        if inc <= 0:
            return 0.0
        if b0 + inc - max_per_team > 1e-9:
            return -1e99
        m = float(pool.loc[i, "predicted_team_total_bids"])
        if m < 0:
            m = 0.0
        exp_pts = float(pool.loc[i, "expected_team_points"])
        denom0 = m + b0
        denom1 = m + b0 + inc
        s0 = (b0 / denom0) if denom0 > 0 else 0.0
        s1 = ((b0 + inc) / denom1) if denom1 > 0 else 0.0
        return exp_pts * (s1 - s0)

    remaining = float(budget)

    # Seed with min_teams by choosing best value for buying min_bid from 0.
    while len(selected) < int(min_teams):
        best_i: Optional[int] = None
        best_v = -1e99
        for i in range(n):
            if i in selected:
                continue
            v = _delta_for(i, 0.0, float(min_bid)) / float(min_bid)
            if v > best_v:
                best_v = v
                best_i = i
        if best_i is None:
            break
        bids[best_i] = float(min_bid)
        selected.add(best_i)
        remaining -= float(min_bid)

    # Greedy marginal allocation.
    while remaining > 1e-9:
        best_i: Optional[int] = None
        best_inc: float = 0.0
        best_val = -1e99

        # Increment an existing team.
        for i in selected:
            inc = float(step) if remaining >= step else float(remaining)
            if bids[i] + inc - max_per_team > 1e-9:
                inc = max(0.0, float(max_per_team) - float(bids[i]))
            if inc <= 1e-12:
                continue
            v = _delta_for(i, float(bids[i]), float(inc)) / float(inc)
            if v > best_val:
                best_val = v
                best_i = i
                best_inc = float(inc)

        # Add a new team (must allocate min_bid at once).
        if (
            len(selected) < int(max_teams)
            and remaining + 1e-9 >= float(min_bid)
        ):
            for i in range(n):
                if i in selected:
                    continue
                if float(min_bid) - float(max_per_team) > 1e-9:
                    continue
                v = _delta_for(i, 0.0, float(min_bid)) / float(min_bid)
                if v > best_val:
                    best_val = v
                    best_i = i
                    best_inc = float(min_bid)

        if best_i is None or best_inc <= 1e-12:
            break

        if best_i not in selected and abs(best_inc - float(min_bid)) < 1e-9:
            selected.add(best_i)
        bids[best_i] += float(best_inc)
        remaining -= float(best_inc)

    chosen = pool.loc[sorted(selected)].copy().reset_index(drop=True)
    chosen["bid_amount"] = [float(bids[i]) for i in sorted(selected)]

    portfolio_rows: List[Dict[str, object]] = []
    for _, r in chosen.iterrows():
        portfolio_rows.append(
            {
                "team_key": str(r["team_key"]),
                "bid_amount": float(r["bid_amount"]),
                "score": float(r.get(score_col, 0.0) or 0.0),
            }
        )
    return chosen, portfolio_rows


def _predict_share_for_snapshot(
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
                pd.to_numeric(
                    tt["wins"],
                    errors="coerce",
                )
                .fillna(0)
                .astype(int)
            )
            tt["byes"] = (
                pd.to_numeric(
                    tt["byes"],
                    errors="coerce",
                )
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
        shares = _compute_team_shares_from_bids(
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


def _run_single_year(
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
    market_mode: str,
    points_mode: str,
    expected_sims: int,
    expected_seed: int,
    expected_use_historical_winners: bool,
    kenpom_scale: float,
    kenpom_scale_source: str,
    allocation_mode: str,
    exclude_entry_names: List[str],
    train_mode: str,
) -> Dict[str, object]:
    snapshot = snapshot_dir.name
    train_snapshots = _choose_train_snapshots(
        all_snapshot_dirs=all_snapshot_dirs,
        predict_snapshot=snapshot,
        train_mode=str(train_mode),
    )

    pred_df = _predict_share_for_snapshot(
        out_root=out_root,
        snapshot_dirs=all_snapshot_dirs,
        predict_snapshot=snapshot,
        train_snapshots=train_snapshots,
        ridge_alpha=ridge_alpha,
        exclude_entry_names=exclude_entry_names,
    )

    if "calcutta_key" not in pred_df.columns:
        raise ValueError("team_dataset missing calcutta_key")

    calcutta_key = backtest_scaffold._choose_calcutta_key(pred_df, None)
    pred_df = pred_df[pred_df["calcutta_key"] == calcutta_key].copy()

    tables = backtest_scaffold._load_snapshot_tables(snapshot_dir)
    eligible_team_keys: Optional[set[str]] = None
    teams = tables.get("teams")
    required = {
        "team_key",
        "wins",
        "byes",
        "calcutta_key",
    }
    if teams is not None and required.issubset(set(teams.columns)):
        t = teams[teams["calcutta_key"] == calcutta_key].copy()
        t["wins"] = (
            pd.to_numeric(
                t["wins"],
                errors="coerce",
            )
            .fillna(0)
            .astype(int)
        )
        t["byes"] = (
            pd.to_numeric(
                t["byes"],
                errors="coerce",
            )
            .fillna(0)
            .astype(int)
        )
        eligible_team_keys = set(
            t[(t["wins"] != 0) | (t["byes"] != 0)]["team_key"]
            .astype(str)
            .tolist()
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

    usable["kenpom_z"] = _zscore(usable["kenpom_net"])
    usable["share_z"] = _zscore(usable["predicted_team_share_of_pool"])
    usable["value_score"] = usable["kenpom_z"] - usable["share_z"]

    market_bids = _filter_market_bids(
        tables=tables,
        calcutta_key=calcutta_key,
        exclude_entry_names=exclude_entry_names,
    )
    if eligible_team_keys is not None and not market_bids.empty:
        market_bids = market_bids[
            market_bids["team_key"].astype(str).isin(eligible_team_keys)
        ].copy()

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

    n_entries = int(eligible_entries["entry_key"].nunique())
    predicted_total_pool_bids = float(n_entries) * float(budget)
    competitor_entry_keys = [
        str(k)
        for k in eligible_entries["entry_key"].astype(str).unique().tolist()
    ]

    pmode = points_mode
    if pmode == "auto":
        pmode = "round_scoring" if "round_scoring" in tables else "fixed"

    if allocation_mode in ("expected_points", "greedy"):
        if not expected_sims or expected_sims <= 0:
            raise ValueError(
                "allocation_mode requires --expected-sims > 0"
            )

        points_by_round: Optional[Dict[int, float]] = None
        if pmode == "round_scoring" and "round_scoring" in tables:
            rs = tables["round_scoring"].copy()
            rs = rs[rs["calcutta_key"] == calcutta_key].copy()
            rs["round"] = pd.to_numeric(rs["round"], errors="coerce")
            rs["points"] = pd.to_numeric(rs["points"], errors="coerce")
            rs = rs[rs["round"].notna() & rs["points"].notna()].copy()
            points_by_round = {
                int(r["round"]): float(r["points"]) for _, r in rs.iterrows()
            }

        # Pre-pass expected simulation to estimate team_mean_points.
        # Provide a dummy simulated entry row so standings are always
        # computable.
        dummy_team_key = str(usable.iloc[0]["team_key"])
        dummy_sim_rows = pd.DataFrame(
            {
                "calcutta_key": [calcutta_key],
                "entry_key": ["simulated:entry"],
                "team_key": [dummy_team_key],
                "bid_amount": [0.0],
            }
        )
        exp_pre = backtest_scaffold._expected_simulation(
            tables=tables,
            calcutta_key=calcutta_key,
            market_bids=market_bids,
            sim_rows=dummy_sim_rows,
            sim_entry_key="simulated:entry",
            market_mode=str(market_mode),
            points_mode=pmode,
            points_by_round=points_by_round,
            n_sims=int(expected_sims),
            seed=int(expected_seed),
            kenpom_scale=float(kenpom_scale),
            budget=float(budget),
            use_historical_winners=bool(expected_use_historical_winners),
            competitor_entry_keys=competitor_entry_keys,
        )
        team_mean_points = exp_pre.get("team_mean_points")
        if not isinstance(team_mean_points, dict) or not team_mean_points:
            raise ValueError(
                "expected simulation did not produce team_mean_points"
            )

        usable["expected_team_points"] = usable["team_key"].apply(
            lambda tk: float(team_mean_points.get(str(tk), 0.0))
        )
        usable["predicted_team_total_bids"] = (
            pd.to_numeric(
                usable["predicted_team_share_of_pool"],
                errors="coerce",
            )
            .fillna(0.0)
            .apply(lambda s: float(s) * float(predicted_total_pool_bids))
        )
    if allocation_mode == "greedy":
        chosen, portfolio_rows = _optimize_portfolio_greedy(
            df=usable,
            score_col="value_score",
            budget=float(budget),
            min_teams=int(min_teams),
            max_teams=int(max_teams),
            max_per_team=float(max_per_team),
            min_bid=float(min_bid),
        )
    else:
        chosen, portfolio_rows = _select_portfolio(
            df=usable,
            score_col="value_score",
            budget=budget,
            min_teams=min_teams,
            max_teams=max_teams,
            max_per_team=max_per_team,
            min_bid=min_bid,
            allocation_mode=str(allocation_mode),
        )

    # Realized

    points_by_team = backtest_scaffold._build_points_by_team(
        teams=tables["teams"],
        calcutta_key=calcutta_key,
        points_mode=pmode,
        round_scoring=tables.get("round_scoring"),
    )

    sim_rows = pd.DataFrame(
        {
            "calcutta_key": [calcutta_key for _ in portfolio_rows],
            "entry_key": ["simulated:entry" for _ in portfolio_rows],
            "team_key": [r["team_key"] for r in portfolio_rows],
            "bid_amount": [r["bid_amount"] for r in portfolio_rows],
        }
    )

    if market_mode == "join":
        bids_all = pd.concat([market_bids, sim_rows], ignore_index=True)
        entry_points = backtest_scaffold._compute_entry_points(
            entry_bids=bids_all,
            points_by_team=points_by_team,
            calcutta_key=calcutta_key,
        )
    else:
        entry_points = backtest_scaffold._compute_entry_points(
            entry_bids=market_bids,
            points_by_team=points_by_team,
            calcutta_key=calcutta_key,
        )
        sim_points = backtest_scaffold._compute_sim_entry_points_shadow(
            sim_bids=sim_rows,
            market_entry_bids=market_bids,
            points_by_team=points_by_team,
            calcutta_key=calcutta_key,
            sim_entry_key="simulated:entry",
        )
        entry_points = pd.concat(
            [entry_points, sim_points],
            ignore_index=True,
        )

    entry_points = backtest_scaffold._ensure_entry_points_include_competitors(
        entry_points,
        competitor_entry_keys=competitor_entry_keys + ["simulated:entry"],
    )

    standings = backtest_scaffold._compute_finish_positions_and_payouts(
        entry_points=entry_points,
        payouts=tables["payouts"],
        calcutta_key=calcutta_key,
    )

    sim_row = standings[standings["entry_key"] == "simulated:entry"]
    if len(sim_row) != 1:
        raise ValueError("failed to compute simulated entry standing")

    sim = sim_row.iloc[0]
    payout_cents = int(sim["payout_cents"])
    roi = payout_cents / (float(budget) * 100.0) if budget else 0.0
    buyin_cents = int(round(float(real_buyin_dollars) * 100.0))
    roi_real_buyin = payout_cents / float(buyin_cents) if buyin_cents else 0.0
    points_per_fake_dollar = (
        float(sim["total_points"]) / float(budget) if budget else 0.0
    )

    year_out: Dict[str, object] = {
        "snapshot": snapshot,
        "calcutta_key": calcutta_key,
        "market_model": {
            "ridge_alpha": float(ridge_alpha),
            "train_mode": str(train_mode),
            "train_snapshots": train_snapshots,
            "n_train_snapshots": int(len(train_snapshots)),
            "is_cold_start": bool(len(train_snapshots) == 0),
            "target": "team_share_of_pool",
        },
        "strategy": {
            "score": "kenpom_z - predicted_share_z",
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
        "portfolio": chosen[
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
        ].to_dict(orient="records"),
        "realized": {
            "points_mode": pmode,
            "market_mode": str(market_mode),
            "budget": float(budget),
            "real_buyin_dollars": float(real_buyin_dollars),
            "total_points": float(sim["total_points"]),
            "points_per_fake_dollar": float(points_per_fake_dollar),
            "finish_position": int(sim["finish_position"]),
            "is_tied": bool(sim["is_tied"]),
            "payout_cents": payout_cents,
            "roi": float(roi),
            "roi_real_buyin": float(roi_real_buyin),
        },
    }

    if expected_sims and expected_sims > 0:
        points_by_round: Optional[Dict[int, float]] = None
        if pmode == "round_scoring" and "round_scoring" in tables:
            rs = tables["round_scoring"].copy()
            rs = rs[rs["calcutta_key"] == calcutta_key].copy()
            rs["round"] = pd.to_numeric(rs["round"], errors="coerce")
            rs["points"] = pd.to_numeric(rs["points"], errors="coerce")
            rs = rs[rs["round"].notna() & rs["points"].notna()].copy()
            points_by_round = {
                int(r["round"]): float(r["points"])
                for _, r in rs.iterrows()
            }

        exp = backtest_scaffold._expected_simulation(
            tables=tables,
            calcutta_key=calcutta_key,
            market_bids=market_bids,
            sim_rows=sim_rows,
            sim_entry_key="simulated:entry",
            market_mode=str(market_mode),
            points_mode=pmode,
            points_by_round=points_by_round,
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

    return year_out


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Generate a single JSON report that backtests an "
            "underpriced-team strategy across snapshots."
        )
    )
    parser.add_argument(
        "out_root",
        help="Path to Option-A out-root (contains snapshot dirs)",
    )
    parser.add_argument(
        "--ridge-alpha",
        dest="ridge_alpha",
        type=float,
        default=1.0,
        help="Ridge regularization strength for market model (default: 1.0)",
    )
    parser.add_argument(
        "--train-mode",
        dest="train_mode",
        default="past_only",
        choices=["past_only", "loo"],
        help=(
            "past_only trains only on earlier snapshots (default: past_only). "
            "loo trains on all snapshots except the one being predicted."
        ),
    )
    parser.add_argument(
        "--budget",
        dest="budget",
        type=float,
        default=100.0,
        help="Total bankroll (default: 100)",
    )
    parser.add_argument(
        "--min-teams",
        dest="min_teams",
        type=int,
        default=3,
        help="Minimum number of teams to buy (default: 3)",
    )
    parser.add_argument(
        "--max-teams",
        dest="max_teams",
        type=int,
        default=10,
        help="Maximum number of teams to buy (default: 10)",
    )
    parser.add_argument(
        "--max-per-team",
        dest="max_per_team",
        type=float,
        default=50.0,
        help="Max spend per team (default: 50)",
    )
    parser.add_argument(
        "--min-bid",
        dest="min_bid",
        type=float,
        default=1.0,
        help="Min bid per team used by the scaffold (default: 1)",
    )
    parser.add_argument(
        "--market-mode",
        dest="market_mode",
        default="join",
        choices=["join", "shadow"],
        help="join: simulated entry joins the auction (default: join)",
    )
    parser.add_argument(
        "--points-mode",
        dest="points_mode",
        default="auto",
        choices=["auto", "round_scoring", "fixed"],
        help="How to compute points (default: auto)",
    )
    parser.add_argument(
        "--allocation-mode",
        dest="allocation_mode",
        default="equal",
        choices=["equal", "expected_points", "greedy"],
        help=(
            "How to allocate bids across selected teams. equal matches the "
            "old behavior. expected_points concentrates bids by expected "
            "points per dollar (requires --expected-sims > 0). greedy "
            "optimizes both team selection and bid sizing (requires "
            "--expected-sims > 0)."
        ),
    )
    parser.add_argument(
        "--exclude-entry-name",
        dest="exclude_entry_names",
        action="append",
        default=["Andrew Copp"],
        help=(
            "Exclude entries whose entry_name contains this string when "
            "building market totals and standings (default: Andrew Copp). "
            "May be specified multiple times."
        ),
    )
    parser.add_argument(
        "--real-buyin-dollars",
        dest="real_buyin_dollars",
        type=float,
        default=25.0,
        help=(
            "Real entry fee in dollars for real-money ROI reporting "
            "(default: 25)"
        ),
    )
    parser.add_argument(
        "--expected-sims",
        dest="expected_sims",
        type=int,
        default=0,
        help="If >0, run Monte Carlo for expected payout/ROI",
    )
    parser.add_argument(
        "--expected-seed",
        dest="expected_seed",
        type=int,
        default=1,
        help="Random seed for expected Monte Carlo",
    )
    parser.add_argument(
        "--expected-use-historical-winners",
        dest="expected_use_historical_winners",
        action="store_true",
        help=(
            "If set, expected simulation will use winner_team_key values from "
            "games.parquet (leaky for true pre-tournament expectations)"
        ),
    )
    parser.add_argument(
        "--kenpom-scale",
        dest="kenpom_scale",
        type=float,
        default=10.0,
        help="Logistic scale for kenpom_net diff to win prob",
    )
    parser.add_argument(
        "--kenpom-scale-file",
        dest="kenpom_scale_file",
        default=None,
        help="If set, read a JSON file containing kenpom_scale",
    )
    parser.add_argument(
        "--summary-min-train-snapshots",
        dest="summary_min_train_snapshots",
        type=int,
        default=1,
        help=(
            "Exclude years from summary metrics if the market model had fewer "
            "than this many train snapshots (default: 1)."
        ),
    )
    parser.add_argument(
        "--out",
        dest="out_path",
        default=None,
        help=(
            "Write report JSON to this path "
            "(default: <out_root>/report.json)"
        ),
    )

    args = parser.parse_args()

    out_root = Path(args.out_root)
    snapshot_dirs = predict_market_share._find_snapshots(out_root)
    if not snapshot_dirs:
        raise FileNotFoundError(f"no snapshots found under: {out_root}")

    kenpom_scale = float(args.kenpom_scale)
    kenpom_scale_source = "arg"
    if args.kenpom_scale_file:
        p = Path(args.kenpom_scale_file)
        payload = json.loads(p.read_text(encoding="utf-8"))
        if "kenpom_scale" not in payload:
            raise ValueError("kenpom scale file missing kenpom_scale")
        kenpom_scale = float(payload["kenpom_scale"])
        kenpom_scale_source = str(p)

    years: List[Dict[str, object]] = []
    for sd in snapshot_dirs:
        years.append(
            _run_single_year(
                out_root=out_root,
                snapshot_dir=sd,
                all_snapshot_dirs=snapshot_dirs,
                ridge_alpha=float(args.ridge_alpha),
                budget=float(args.budget),
                real_buyin_dollars=float(args.real_buyin_dollars),
                min_teams=int(args.min_teams),
                max_teams=int(args.max_teams),
                max_per_team=float(args.max_per_team),
                min_bid=float(args.min_bid),
                market_mode=str(args.market_mode),
                points_mode=str(args.points_mode),
                expected_sims=int(args.expected_sims),
                expected_seed=int(args.expected_seed),
                expected_use_historical_winners=bool(
                    args.expected_use_historical_winners
                ),
                kenpom_scale=float(kenpom_scale),
                kenpom_scale_source=str(kenpom_scale_source),
                allocation_mode=str(args.allocation_mode),
                exclude_entry_names=list(args.exclude_entry_names or []),
                train_mode=str(args.train_mode),
            )
        )

    min_train = int(args.summary_min_train_snapshots)
    eligible_years = [
        y
        for y in years
        if int(y.get("market_model", {}).get("n_train_snapshots", 0))
        >= min_train
    ]
    eligible_snapshots = set(
        str(y.get("snapshot")) for y in eligible_years
    )
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
    points_per_fake_dollar = [
        float(y.get("realized", {}).get("points_per_fake_dollar", 0.0))
        for y in years
    ]
    payout_cents = [
        int(y.get("realized", {}).get("payout_cents", 0))
        for y in years
    ]

    rois_eligible = [
        float(y.get("realized", {}).get("roi", 0.0)) for y in eligible_years
    ]
    rois_real_buyin_eligible = [
        float(y.get("realized", {}).get("roi_real_buyin", 0.0))
        for y in eligible_years
    ]
    points_per_fake_dollar_eligible = [
        float(y.get("realized", {}).get("points_per_fake_dollar", 0.0))
        for y in eligible_years
    ]
    payout_cents_eligible = [
        int(y.get("realized", {}).get("payout_cents", 0))
        for y in eligible_years
    ]

    report: Dict[str, object] = {
        "snapshots": [sd.name for sd in snapshot_dirs],
        "config": {
            "ridge_alpha": float(args.ridge_alpha),
            "train_mode": str(args.train_mode),
            "budget": float(args.budget),
            "real_buyin_dollars": float(args.real_buyin_dollars),
            "min_teams": int(args.min_teams),
            "max_teams": int(args.max_teams),
            "max_per_team": float(args.max_per_team),
            "min_bid": float(args.min_bid),
            "market_mode": str(args.market_mode),
            "points_mode": str(args.points_mode),
            "allocation_mode": str(args.allocation_mode),
            "exclude_entry_names": list(args.exclude_entry_names or []),
            "expected_sims": int(args.expected_sims),
            "expected_seed": int(args.expected_seed),
            "expected_use_historical_winners": bool(
                args.expected_use_historical_winners
            ),
            "kenpom_scale": float(kenpom_scale),
            "kenpom_scale_source": str(kenpom_scale_source),
            "summary_min_train_snapshots": int(
                args.summary_min_train_snapshots
            ),
        },
        "summary": {
            "n_years": int(len(years)),
            "mean_roi": float(np.mean(rois)) if rois else 0.0,
            "median_roi": float(np.median(rois)) if rois else 0.0,
            "mean_roi_real_buyin": (
                float(np.mean(rois_real_buyin)) if rois_real_buyin else 0.0
            ),
            "median_roi_real_buyin": (
                float(np.median(rois_real_buyin)) if rois_real_buyin else 0.0
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
            "mean_payout_cents": (
                float(np.mean(payout_cents))
                if payout_cents
                else 0.0
            ),
            "median_payout_cents": (
                float(np.median(payout_cents))
                if payout_cents
                else 0.0
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

    out_path = (
        Path(args.out_path)
        if args.out_path
        else (out_root / "report.json")
    )
    out_path.write_text(json.dumps(report, indent=2) + "\n", encoding="utf-8")
    print(str(out_path))

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
