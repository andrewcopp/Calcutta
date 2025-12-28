import json
from pathlib import Path
from typing import Dict, List, Optional, Tuple

import pandas as pd

from calcutta_ds.allocation import waterfill_equal
from calcutta_ds.io import choose_calcutta_key, load_snapshot_tables
from calcutta_ds.points import build_points_by_team
from calcutta_ds.sim import expected_simulation
from calcutta_ds.standings import (
    compute_entry_points,
    compute_finish_positions_and_payouts,
)


def _validate_constraints(
    *,
    budget: float,
    min_teams: int,
    max_teams: int,
    min_bid: float,
) -> None:
    if min_teams <= 0 or max_teams <= 0:
        raise ValueError("min_teams and max_teams must be positive")
    if min_teams > max_teams:
        raise ValueError("min_teams cannot exceed max_teams")
    if budget <= 0:
        raise ValueError("budget must be positive")
    if min_bid <= 0:
        raise ValueError("min_bid must be positive")


def _select_k(
    df_len: int,
    *,
    max_teams: int,
    budget: float,
    min_bid: float,
) -> int:
    max_k_by_budget = int(budget // min_bid)
    return min(int(max_teams), int(max_k_by_budget), int(df_len))


def _build_portfolio_equal(
    df: pd.DataFrame,
    *,
    score_col: str,
    k: int,
    budget: float,
    max_per_team: float,
    min_bid: float,
) -> Tuple[pd.DataFrame, List[Dict[str, object]]]:
    ranked = df.sort_values(by=score_col, ascending=False)
    chosen = ranked.head(int(k)).copy().reset_index(drop=True)

    bids = waterfill_equal(
        k=int(k),
        budget=float(budget),
        max_per_team=float(max_per_team),
    )

    if any(b < float(min_bid) for b in bids):
        raise ValueError("allocation violates min_bid constraint")

    chosen["bid_amount"] = bids

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


def _points_mode_auto(
    points_mode: str,
    tables: Dict[str, pd.DataFrame],
) -> str:
    if points_mode != "auto":
        return points_mode
    return "round_scoring" if "round_scoring" in tables else "fixed"


def _realized_results(
    *,
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    points_mode: str,
    portfolio_rows: List[Dict[str, object]],
    sim_entry_key: str,
    budget: float,
) -> Tuple[Dict[str, object], pd.DataFrame]:
    points_by_team = build_points_by_team(
        teams=tables["teams"],
        calcutta_key=calcutta_key,
        points_mode=points_mode,
        round_scoring=tables.get("round_scoring"),
    )

    market_bids = tables["entry_bids"].copy()
    sim_rows = pd.DataFrame(
        {
            "calcutta_key": [calcutta_key for _ in portfolio_rows],
            "entry_key": [sim_entry_key for _ in portfolio_rows],
            "team_key": [r["team_key"] for r in portfolio_rows],
            "bid_amount": [r["bid_amount"] for r in portfolio_rows],
        }
    )

    bids_all = pd.concat([market_bids, sim_rows], ignore_index=True)
    entry_points = compute_entry_points(
        entry_bids=bids_all,
        points_by_team=points_by_team,
        calcutta_key=calcutta_key,
    )

    standings = compute_finish_positions_and_payouts(
        entry_points=entry_points,
        payouts=tables["payouts"],
        calcutta_key=calcutta_key,
    )

    sim_row = standings[standings["entry_key"] == sim_entry_key]
    if len(sim_row) != 1:
        raise ValueError("failed to compute simulated entry standing")
    sim_r = sim_row.iloc[0]

    payout_cents = int(sim_r["payout_cents"])
    payout_per_fake_dollar = (
        float(payout_cents) / (float(budget) * 100.0)
        if float(budget)
        else 0.0
    )
    points_per_fake_dollar = (
        float(sim_r["total_points"]) / float(budget)
        if float(budget)
        else 0.0
    )
    roi = float(points_per_fake_dollar)

    realized = {
        "points_mode": points_mode,
        "total_points": float(sim_r["total_points"]),
        "points_per_fake_dollar": float(points_per_fake_dollar),
        "finish_position": int(sim_r["finish_position"]),
        "is_tied": bool(sim_r["is_tied"]),
        "payout_cents": payout_cents,
        "payout_per_fake_dollar": float(payout_per_fake_dollar),
        "roi": roi,
    }
    return realized, standings


def _points_by_round_from_tables(
    *,
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    points_mode: str,
) -> Optional[Dict[int, float]]:
    if points_mode != "round_scoring":
        return None
    if "round_scoring" not in tables:
        return None

    rs = tables["round_scoring"].copy()
    rs = rs[rs["calcutta_key"] == calcutta_key].copy()
    rs["round"] = pd.to_numeric(rs["round"], errors="coerce")
    rs["points"] = pd.to_numeric(rs["points"], errors="coerce")
    rs = rs[rs["round"].notna() & rs["points"].notna()].copy()
    return {int(r["round"]): float(r["points"]) for _, r in rs.iterrows()}


def _standings_with_names(
    standings_df: pd.DataFrame,
    entries: pd.DataFrame,
) -> pd.DataFrame:
    out = standings_df.copy()
    if "entry_key" in entries.columns and "entry_name" in entries.columns:
        out = out.merge(
            entries[["entry_key", "entry_name"]],
            on="entry_key",
            how="left",
        )
    return out


def build_output(
    *,
    snapshot_dir: Path,
    calcutta_key: str,
    score_col: str,
    constraints: Dict[str, object],
    portfolio_rows: List[Dict[str, object]],
    realized: Dict[str, object],
    standings_top: List[Dict[str, object]],
    expected: Optional[Dict[str, object]],
) -> Dict[str, object]:
    out: Dict[str, object] = {
        "snapshot": snapshot_dir.name,
        "calcutta_key": calcutta_key,
        "constraints": constraints,
        "score_col": str(score_col),
        "portfolio": portfolio_rows,
        "summary": {
            "n_teams": int(len(portfolio_rows)),
            "total_spend": float(sum(r["bid_amount"] for r in portfolio_rows)),
        },
        "realized": realized,
        "standings_top": standings_top,
    }
    if expected:
        out["expected"] = expected
    return out


def run_backtest_scaffold(
    *,
    snapshot_dir: Path,
    calcutta_key: Optional[str],
    score_col: str,
    budget: float,
    min_teams: int,
    max_teams: int,
    max_per_team: float,
    min_bid: float,
    points_mode: str,
    sim_entry_key: str,
    expected_sims: int,
    expected_seed: int,
    expected_use_historical_winners: bool,
    kenpom_scale: float,
    kenpom_scale_file: Optional[str],
) -> Dict[str, object]:
    tables = load_snapshot_tables(snapshot_dir)
    df = tables["team_dataset"]

    ck = choose_calcutta_key(df, calcutta_key)
    df = df[df["calcutta_key"] == ck].copy()

    if score_col not in df.columns:
        raise ValueError(f"score column not found: {score_col}")

    df[score_col] = pd.to_numeric(df[score_col], errors="coerce")
    df = df[df[score_col].notna()].copy()

    if "team_key" not in df.columns:
        raise ValueError("team_dataset missing team_key")

    _validate_constraints(
        budget=float(budget),
        min_teams=int(min_teams),
        max_teams=int(max_teams),
        min_bid=float(min_bid),
    )

    k = _select_k(len(df), max_teams=max_teams, budget=budget, min_bid=min_bid)
    if int(k) < int(min_teams):
        raise ValueError("budget too small to satisfy min_teams at min_bid")

    _, portfolio_rows = _build_portfolio_equal(
        df,
        score_col=score_col,
        k=k,
        budget=budget,
        max_per_team=max_per_team,
        min_bid=min_bid,
    )

    pm = _points_mode_auto(points_mode, tables)
    realized, standings_df = _realized_results(
        tables=tables,
        calcutta_key=ck,
        points_mode=pm,
        portfolio_rows=portfolio_rows,
        sim_entry_key=sim_entry_key,
        budget=budget,
    )

    standings_with_names = _standings_with_names(
        standings_df,
        tables["entries"],
    )

    exp_out: Optional[Dict[str, object]] = None
    if int(expected_sims) > 0:
        scale = float(kenpom_scale)
        scale_source = "arg"
        if kenpom_scale_file:
            p = Path(kenpom_scale_file)
            payload = json.loads(p.read_text(encoding="utf-8"))
            if "kenpom_scale" not in payload:
                raise ValueError("kenpom scale file missing kenpom_scale")
            scale = float(payload["kenpom_scale"])
            scale_source = str(p)

        points_by_round = _points_by_round_from_tables(
            tables=tables,
            calcutta_key=ck,
            points_mode=pm,
        )

        market_bids = tables["entry_bids"].copy()
        sim_rows = pd.DataFrame(
            {
                "calcutta_key": [ck for _ in portfolio_rows],
                "entry_key": [sim_entry_key for _ in portfolio_rows],
                "team_key": [r["team_key"] for r in portfolio_rows],
                "bid_amount": [r["bid_amount"] for r in portfolio_rows],
            }
        )

        exp = expected_simulation(
            tables=tables,
            calcutta_key=ck,
            market_bids=market_bids,
            sim_rows=sim_rows,
            sim_entry_key=sim_entry_key,
            points_mode=pm,
            points_by_round=points_by_round,
            n_sims=int(expected_sims),
            seed=int(expected_seed),
            kenpom_scale=float(scale),
            budget=float(budget),
            use_historical_winners=bool(expected_use_historical_winners),
        )
        if exp:
            exp["kenpom_scale_source"] = scale_source
            exp_out = exp

    constraints = {
        "budget": float(budget),
        "min_teams": int(min_teams),
        "max_teams": int(max_teams),
        "max_per_team": float(max_per_team),
        "min_bid": float(min_bid),
    }

    return build_output(
        snapshot_dir=snapshot_dir,
        calcutta_key=ck,
        score_col=score_col,
        constraints=constraints,
        portfolio_rows=portfolio_rows,
        realized=realized,
        standings_top=standings_with_names.head(10).to_dict(orient="records"),
        expected=exp_out,
    )
