import argparse
import json
import math
import random
from pathlib import Path
from typing import Dict, List, Optional, Tuple

import pandas as pd


def _read_parquet(p: Path) -> pd.DataFrame:
    if not p.exists():
        raise FileNotFoundError(f"missing required file: {p}")
    return pd.read_parquet(p)


def _choose_calcutta_key(df: pd.DataFrame, requested: Optional[str]) -> str:
    if "calcutta_key" not in df.columns:
        raise ValueError("team_dataset missing calcutta_key")

    keys = [k for k in sorted(df["calcutta_key"].dropna().unique())]
    if requested:
        if requested not in keys:
            raise ValueError(f"calcutta_key not found: {requested}")
        return requested

    if len(keys) != 1:
        raise ValueError(
            "multiple calcutta_key values found; pass --calcutta-key"
        )
    return str(keys[0])


def _waterfill_equal(
    k: int,
    budget: float,
    max_per_team: float,
) -> List[float]:
    if k <= 0:
        return []
    if budget < 0:
        raise ValueError("budget must be non-negative")
    if max_per_team <= 0:
        raise ValueError("max_per_team must be positive")

    bids: List[float] = [0.0 for _ in range(k)]
    remaining = budget
    remaining_idx = list(range(k))

    while remaining_idx:
        per = remaining / float(len(remaining_idx))
        capped = [i for i in remaining_idx if per > max_per_team]
        if not capped:
            for i in remaining_idx:
                bids[i] = per
            remaining = 0.0
            break

        for i in capped:
            bids[i] = max_per_team
        remaining -= max_per_team * float(len(capped))
        remaining_idx = [i for i in remaining_idx if i not in capped]

        if remaining < -1e-9:
            raise ValueError("invalid allocation: remaining budget < 0")

    return bids


def _load_snapshot_tables(snapshot_dir: Path) -> Dict[str, pd.DataFrame]:
    tables: Dict[str, pd.DataFrame] = {}

    # Required
    tables["team_dataset"] = _read_parquet(
        snapshot_dir / "derived" / "team_dataset.parquet"
    )
    tables["teams"] = _read_parquet(snapshot_dir / "teams.parquet")
    tables["entries"] = _read_parquet(snapshot_dir / "entries.parquet")
    tables["entry_bids"] = _read_parquet(snapshot_dir / "entry_bids.parquet")
    tables["payouts"] = _read_parquet(snapshot_dir / "payouts.parquet")

    # Optional
    round_scoring_path = snapshot_dir / "round_scoring.parquet"
    if round_scoring_path.exists():
        tables["round_scoring"] = _read_parquet(round_scoring_path)

    games_path = snapshot_dir / "games.parquet"
    if games_path.exists():
        tables["games"] = _read_parquet(games_path)

    return tables


def _round_order(round_name: str) -> int:
    order = {
        "first_four": 1,
        "round_of_64": 2,
        "round_of_32": 3,
        "sweet_16": 4,
        "elite_8": 5,
        "final_four": 6,
        "championship": 7,
    }
    return int(order.get(str(round_name), 999))


def _sigmoid(x: float) -> float:
    # Numerically stable sigmoid.
    if x >= 0:
        z = math.exp(-x)
        return 1.0 / (1.0 + z)
    z = math.exp(x)
    return z / (1.0 + z)


def _win_prob(
    net1: float,
    net2: float,
    scale: float,
) -> float:
    # Simple calibration: logistic on rating diff.
    if scale <= 0:
        raise ValueError("kenpom_scale must be positive")
    return _sigmoid((net1 - net2) / scale)


def _prepare_bracket_graph(
    games: pd.DataFrame,
) -> Tuple[pd.DataFrame, Dict[str, Dict[int, str]]]:
    required = [
        "game_id",
        "round",
        "sort_order",
        "team1_key",
        "team2_key",
        "next_game_id",
        "next_game_slot",
    ]
    missing = [c for c in required if c not in games.columns]
    if missing:
        raise ValueError(f"games.parquet missing columns: {missing}")

    g = games.copy()
    g["sort_order"] = (
        pd.to_numeric(g["sort_order"], errors="coerce")
        .fillna(0)
        .astype(int)
    )
    g["round_order"] = g["round"].apply(_round_order)

    # prev_by_next[next_game_id][slot] = prev_game_id
    prev_by_next: Dict[str, Dict[int, str]] = {}
    for _, r in g.iterrows():
        nxt = str(r.get("next_game_id") or "")
        if not nxt:
            continue
        slot_raw = r.get("next_game_slot")
        try:
            slot = int(slot_raw)
        except Exception:
            continue
        if slot not in (1, 2):
            continue
        prev_by_next.setdefault(nxt, {})[slot] = str(r.get("game_id"))

    g = g.sort_values(
        by=["round_order", "sort_order", "game_id"],
    ).reset_index(drop=True)
    return g, prev_by_next


def _expected_simulation(
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    market_bids: pd.DataFrame,
    sim_rows: pd.DataFrame,
    sim_entry_key: str,
    market_mode: str,
    points_mode: str,
    points_by_round: Optional[Dict[int, float]],
    n_sims: int,
    seed: int,
    kenpom_scale: float,
    budget: float,
) -> Dict[str, object]:
    if n_sims <= 0:
        return {}
    if "games" not in tables:
        raise ValueError("expected simulation requires games.parquet")

    games, prev_by_next = _prepare_bracket_graph(tables["games"])

    teams = tables["teams"].copy()
    for c in ["wins", "byes"]:
        teams[c] = (
            pd.to_numeric(teams[c], errors="coerce")
            .fillna(0)
            .astype(int)
        )

    if "kenpom_net" not in teams.columns:
        raise ValueError(
            "teams.parquet missing kenpom_net "
            "(needed for simulation)"
        )
    teams["kenpom_net"] = pd.to_numeric(teams["kenpom_net"], errors="coerce")

    net_by_team: Dict[str, float] = {}
    byes_by_team: Dict[str, int] = {}
    for _, r in teams.iterrows():
        tk = str(r.get("team_key") or "")
        if not tk:
            continue
        net = r.get("kenpom_net")
        if pd.isna(net):
            continue
        net_by_team[tk] = float(net)
        byes_by_team[tk] = int(r.get("byes") or 0)

    rng = random.Random(seed)
    payouts: List[int] = []
    rois: List[float] = []
    points_list: List[float] = []
    finish_positions: List[int] = []

    for _ in range(n_sims):
        # Simulate winners.
        wins_sim: Dict[str, int] = {}
        winners_by_game: Dict[str, str] = {}

        for _, gr in games.iterrows():
            gid = str(gr.get("game_id"))

            winner_fixed = str(gr.get("winner_team_key") or "")
            if winner_fixed:
                winners_by_game[gid] = winner_fixed
                wins_sim[winner_fixed] = wins_sim.get(winner_fixed, 0) + 1
                continue

            t1 = str(gr.get("team1_key") or "")
            t2 = str(gr.get("team2_key") or "")

            if not t1:
                prev = prev_by_next.get(gid, {}).get(1)
                if prev:
                    t1 = winners_by_game.get(prev, "")
            if not t2:
                prev = prev_by_next.get(gid, {}).get(2)
                if prev:
                    t2 = winners_by_game.get(prev, "")

            if not t1 or not t2:
                continue

            net1 = float(net_by_team.get(t1, 0.0))
            net2 = float(net_by_team.get(t2, 0.0))
            p1 = _win_prob(net1, net2, kenpom_scale)
            w = t1 if rng.random() < p1 else t2
            winners_by_game[gid] = w
            wins_sim[w] = wins_sim.get(w, 0) + 1

        # Compute team points.
        team_points: Dict[str, float] = {}
        for team_key, byes in byes_by_team.items():
            progress = int(byes) + int(wins_sim.get(team_key, 0))
            if points_mode == "round_scoring" and points_by_round is not None:
                team_points[team_key] = _team_points_from_round_scoring(
                    progress,
                    points_by_round,
                )
            else:
                team_points[team_key] = _team_points_fixed(progress)

        # Entry points for historical entries.
        points_by_team_df = pd.DataFrame(
            {
                "team_key": list(team_points.keys()),
                "team_points": list(team_points.values()),
            }
        )
        entry_points = _compute_entry_points(
            entry_bids=market_bids,
            points_by_team=points_by_team_df,
            calcutta_key=calcutta_key,
        )

        # Sim entry points.
        if market_mode == "join":
            bids_all = pd.concat([market_bids, sim_rows], ignore_index=True)
            entry_points = _compute_entry_points(
                entry_bids=bids_all,
                points_by_team=points_by_team_df,
                calcutta_key=calcutta_key,
            )
        else:
            sim_points = _compute_sim_entry_points_shadow(
                sim_bids=sim_rows,
                market_entry_bids=market_bids,
                points_by_team=points_by_team_df,
                calcutta_key=calcutta_key,
                sim_entry_key=sim_entry_key,
            )
            entry_points = pd.concat(
                [entry_points, sim_points],
                ignore_index=True,
            )

        standings = _compute_finish_positions_and_payouts(
            entry_points=entry_points,
            payouts=tables["payouts"],
            calcutta_key=calcutta_key,
        )
        sim_row = standings[standings["entry_key"] == sim_entry_key]
        if len(sim_row) != 1:
            continue
        sim = sim_row.iloc[0]

        payout = int(sim["payout_cents"])
        total_points = float(sim["total_points"])
        finish_pos = int(sim["finish_position"])

        payouts.append(payout)
        points_list.append(total_points)
        finish_positions.append(finish_pos)
        if budget > 0:
            rois.append(payout / (float(budget) * 100.0))
        else:
            rois.append(0.0)

    if not payouts:
        return {}

    def _pct(xs: List[float], p: float) -> float:
        s = sorted(xs)
        idx = int(round((len(s) - 1) * p))
        return float(s[idx])

    expected: Dict[str, object] = {
        "sims": int(len(payouts)),
        "seed": int(seed),
        "budget": float(budget),
        "mean_payout_cents": float(sum(payouts) / len(payouts)),
        "mean_roi": float(sum(rois) / len(rois)),
        "p50_payout_cents": _pct([float(x) for x in payouts], 0.50),
        "p90_payout_cents": _pct([float(x) for x in payouts], 0.90),
        "p50_roi": _pct(rois, 0.50),
        "p90_roi": _pct(rois, 0.90),
        "p50_total_points": _pct(points_list, 0.50),
        "p90_total_points": _pct(points_list, 0.90),
        "finish_position_counts": {},
    }

    counts: Dict[str, int] = {}
    for fp in finish_positions:
        k = str(int(fp))
        counts[k] = counts.get(k, 0) + 1
    expected["finish_position_counts"] = counts
    return expected


def _team_points_fixed(progress: int) -> float:
    # Mirrors backend CalculatePoints() mapping.
    if progress in (0, 1):
        return 0.0
    if progress == 2:
        return 50.0
    if progress == 3:
        return 150.0
    if progress == 4:
        return 300.0
    if progress == 5:
        return 500.0
    if progress == 6:
        return 750.0
    if progress == 7:
        return 1050.0
    return 0.0


def _team_points_from_round_scoring(
    progress: int,
    points_by_round: Dict[int, float],
) -> float:
    # progress = wins + byes
    # scoring_rounds = progress - 1 (see backend mapping)
    scoring_rounds = max(progress - 1, 0)
    total = 0.0
    for r in range(1, scoring_rounds + 1):
        total += float(points_by_round.get(r, 0.0))
    return total


def _build_points_by_team(
    teams: pd.DataFrame,
    calcutta_key: str,
    points_mode: str,
    round_scoring: Optional[pd.DataFrame],
) -> pd.DataFrame:
    required = ["team_key", "wins", "byes"]
    missing = [c for c in required if c not in teams.columns]
    if missing:
        raise ValueError(f"teams.parquet missing columns: {missing}")

    points_by_round: Dict[int, float] = {}
    if points_mode == "round_scoring":
        if round_scoring is None:
            raise ValueError(
                "points_mode=round_scoring but round_scoring not found"
            )
        rs = round_scoring.copy()
        rs = rs[rs["calcutta_key"] == calcutta_key].copy()
        rs["round"] = pd.to_numeric(rs["round"], errors="coerce")
        rs["points"] = pd.to_numeric(rs["points"], errors="coerce")
        rs = rs[rs["round"].notna() & rs["points"].notna()].copy()
        for _, r in rs.iterrows():
            points_by_round[int(r["round"])] = float(r["points"])

    t = teams.copy()
    t["wins"] = pd.to_numeric(t["wins"], errors="coerce").fillna(0).astype(int)
    t["byes"] = pd.to_numeric(t["byes"], errors="coerce").fillna(0).astype(int)
    t["progress"] = t["wins"] + t["byes"]

    if points_mode == "round_scoring":
        t["team_points"] = t["progress"].apply(
            lambda p: _team_points_from_round_scoring(int(p), points_by_round)
        )
    else:
        t["team_points"] = t["progress"].apply(
            lambda p: _team_points_fixed(int(p))
        )

    return t[["team_key", "team_points"]].copy()


def _compute_entry_points(
    entry_bids: pd.DataFrame,
    points_by_team: pd.DataFrame,
    calcutta_key: str,
) -> pd.DataFrame:
    required = ["calcutta_key", "entry_key", "team_key", "bid_amount"]
    missing = [c for c in required if c not in entry_bids.columns]
    if missing:
        raise ValueError(f"entry_bids.parquet missing columns: {missing}")

    bids = entry_bids.copy()
    bids = bids[bids["calcutta_key"] == calcutta_key].copy()
    bids["bid_amount"] = (
        pd.to_numeric(bids["bid_amount"], errors="coerce")
        .fillna(0.0)
    )

    totals = (
        bids.groupby(["team_key"], dropna=False)
        .agg(total_team_bids=("bid_amount", "sum"))
        .reset_index()
    )

    b = bids.merge(totals, on="team_key", how="left")
    b = b.merge(points_by_team, on="team_key", how="left")
    b["team_points"] = (
        pd.to_numeric(b["team_points"], errors="coerce")
        .fillna(0.0)
    )
    b["total_team_bids"] = (
        pd.to_numeric(b["total_team_bids"], errors="coerce")
        .fillna(0.0)
    )
    b["ownership"] = b.apply(
        lambda r: (
            (r["bid_amount"] / r["total_team_bids"])
            if r["total_team_bids"]
            else 0.0
        ),
        axis=1,
    )
    b["points"] = b["team_points"] * b["ownership"]

    entry_points = (
        b.groupby(["entry_key"], dropna=False)
        .agg(total_points=("points", "sum"))
        .reset_index()
    )
    entry_points["total_points"] = (
        pd.to_numeric(entry_points["total_points"], errors="coerce")
        .fillna(0.0)
    )
    return entry_points


def _compute_sim_entry_points_shadow(
    sim_bids: pd.DataFrame,
    market_entry_bids: pd.DataFrame,
    points_by_team: pd.DataFrame,
    calcutta_key: str,
    sim_entry_key: str,
) -> pd.DataFrame:
    # Compute simulated entry points using historical market totals as the
    # ownership denominator, without affecting historical entries.
    sim = sim_bids.copy()
    sim = sim[sim["calcutta_key"] == calcutta_key].copy()
    sim["bid_amount"] = (
        pd.to_numeric(sim["bid_amount"], errors="coerce")
        .fillna(0.0)
    )
    sim["entry_key"] = sim_entry_key

    market = market_entry_bids.copy()
    market = market[market["calcutta_key"] == calcutta_key].copy()
    market["bid_amount"] = (
        pd.to_numeric(market["bid_amount"], errors="coerce")
        .fillna(0.0)
    )
    totals = (
        market.groupby(["team_key"], dropna=False)
        .agg(total_team_bids=("bid_amount", "sum"))
        .reset_index()
    )

    b = sim.merge(totals, on="team_key", how="left")
    b = b.merge(points_by_team, on="team_key", how="left")
    b["team_points"] = (
        pd.to_numeric(b["team_points"], errors="coerce")
        .fillna(0.0)
    )
    b["total_team_bids"] = (
        pd.to_numeric(b["total_team_bids"], errors="coerce")
        .fillna(0.0)
    )
    b["ownership"] = b.apply(
        lambda r: (
            (r["bid_amount"] / r["total_team_bids"])
            if r["total_team_bids"]
            else 0.0
        ),
        axis=1,
    )
    b["points"] = b["team_points"] * b["ownership"]

    pts = (
        b.groupby(["entry_key"], dropna=False)
        .agg(total_points=("points", "sum"))
        .reset_index()
    )
    pts["total_points"] = (
        pd.to_numeric(pts["total_points"], errors="coerce")
        .fillna(0.0)
    )
    return pts


def _compute_finish_positions_and_payouts(
    entry_points: pd.DataFrame,
    payouts: pd.DataFrame,
    calcutta_key: str,
    epsilon: float = 0.0001,
) -> pd.DataFrame:
    req = ["calcutta_key", "position", "amount_cents"]
    missing = [c for c in req if c not in payouts.columns]
    if missing:
        raise ValueError(f"payouts.parquet missing columns: {missing}")

    payout_map: Dict[int, int] = {}
    p = payouts[payouts["calcutta_key"] == calcutta_key].copy()
    p["position"] = pd.to_numeric(p["position"], errors="coerce")
    p["amount_cents"] = pd.to_numeric(p["amount_cents"], errors="coerce")
    p = p[p["position"].notna() & p["amount_cents"].notna()].copy()
    for _, r in p.iterrows():
        payout_map[int(r["position"])] = int(r["amount_cents"])

    ep = entry_points.copy()
    ep["total_points"] = (
        pd.to_numeric(ep["total_points"], errors="coerce")
        .fillna(0.0)
    )

    # Stable tie-breaker: entry_key asc.
    ep = ep.sort_values(
        by=["total_points", "entry_key"],
        ascending=[False, True],
    ).reset_index(drop=True)

    finish_pos: List[int] = []
    is_tied: List[bool] = []
    payout_cents: List[int] = []

    position = 1
    i = 0
    while i < len(ep):
        j = i + 1
        cur = float(ep.loc[i, "total_points"])
        while j < len(ep):
            nxt = float(ep.loc[j, "total_points"])
            if abs(nxt - cur) >= epsilon:
                break
            j += 1

        group_size = j - i
        tied = group_size > 1

        total_group_payout = 0
        for pos in range(position, position + group_size):
            total_group_payout += payout_map.get(pos, 0)

        base = 0
        remainder = 0
        if group_size > 0:
            base = total_group_payout // group_size
            remainder = total_group_payout % group_size

        for k in range(group_size):
            finish_pos.append(position)
            is_tied.append(tied)
            amt = base
            if remainder > 0:
                amt += 1
                remainder -= 1
            payout_cents.append(int(amt))

        position += group_size
        i = j

    ep["finish_position"] = finish_pos
    ep["is_tied"] = is_tied
    ep["payout_cents"] = payout_cents
    ep["in_the_money"] = ep["payout_cents"].apply(lambda v: bool(int(v) > 0))
    return ep


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Backtesting scaffold: construct a feasible portfolio "
            "under auction "
            "constraints using a simple ranking score."
        )
    )
    parser.add_argument(
        "snapshot_dir",
        help=(
            "Path to an ingested snapshot directory "
            "(expects derived/team_dataset.parquet)"
        ),
    )
    parser.add_argument(
        "--calcutta-key",
        dest="calcutta_key",
        default=None,
        help="Which calcutta_key to backtest (required if multiple exist)",
    )
    parser.add_argument(
        "--score",
        dest="score_col",
        default="kenpom_net",
        help="Column to rank teams by (default: kenpom_net)",
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
        "--points-mode",
        dest="points_mode",
        default="fixed",
        choices=["auto", "round_scoring", "fixed"],
        help=(
            "How to convert wins+byes into team points. "
            "auto uses round_scoring if available."
        ),
    )
    parser.add_argument(
        "--market-mode",
        dest="market_mode",
        default="shadow",
        choices=["join", "shadow"],
        help=(
            "join: simulated entry joins the auction and dilutes ownership "
            "for all. "
            "shadow: do not dilute historical entries (what-if)."
        ),
    )
    parser.add_argument(
        "--sim-entry-key",
        dest="sim_entry_key",
        default="simulated:entry",
        help="Entry key used for the simulated portfolio",
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
        "--kenpom-scale",
        dest="kenpom_scale",
        type=float,
        default=10.0,
        help="Logistic scale for kenpom_net diff to win prob",
    )
    parser.add_argument(
        "--out",
        dest="out_path",
        default=None,
        help="Write portfolio JSON to this path (default: stdout)",
    )

    args = parser.parse_args()

    snapshot_dir = Path(args.snapshot_dir)
    tables = _load_snapshot_tables(snapshot_dir)
    df = tables["team_dataset"]

    calcutta_key = _choose_calcutta_key(df, args.calcutta_key)
    df = df[df["calcutta_key"] == calcutta_key].copy()

    if args.score_col not in df.columns:
        raise ValueError(f"score column not found: {args.score_col}")

    df[args.score_col] = pd.to_numeric(df[args.score_col], errors="coerce")
    df = df[df[args.score_col].notna()].copy()

    if "team_key" not in df.columns:
        raise ValueError("team_dataset missing team_key")

    if args.min_teams <= 0 or args.max_teams <= 0:
        raise ValueError("min_teams and max_teams must be positive")
    if args.min_teams > args.max_teams:
        raise ValueError("min_teams cannot exceed max_teams")
    if args.budget <= 0:
        raise ValueError("budget must be positive")
    if args.min_bid <= 0:
        raise ValueError("min_bid must be positive")

    max_k_by_budget = int(args.budget // args.min_bid)
    max_k = min(args.max_teams, max_k_by_budget, int(len(df)))
    if max_k < args.min_teams:
        raise ValueError("budget too small to satisfy min_teams at min_bid")

    k = max_k

    ranked = df.sort_values(by=args.score_col, ascending=False)
    chosen = ranked.head(k).copy()

    bids = _waterfill_equal(
        k=k,
        budget=args.budget,
        max_per_team=args.max_per_team,
    )

    if any(b < args.min_bid for b in bids):
        raise ValueError("allocation violates min_bid constraint")

    chosen = chosen.reset_index(drop=True)
    chosen["bid_amount"] = bids

    portfolio_rows: List[Dict[str, object]] = []
    for _, r in chosen.iterrows():
        portfolio_rows.append(
            {
                "team_key": str(r["team_key"]),
                "bid_amount": float(r["bid_amount"]),
                "score": float(r[args.score_col]),
            }
        )

    # Realized results: points, finish position, payouts.
    points_mode = args.points_mode
    if points_mode == "auto":
        points_mode = "round_scoring" if "round_scoring" in tables else "fixed"

    points_by_team = _build_points_by_team(
        teams=tables["teams"],
        calcutta_key=calcutta_key,
        points_mode=points_mode,
        round_scoring=tables.get("round_scoring"),
    )

    market_bids = tables["entry_bids"].copy()
    sim_rows = pd.DataFrame(
        {
            "calcutta_key": [calcutta_key for _ in portfolio_rows],
            "entry_key": [args.sim_entry_key for _ in portfolio_rows],
            "team_key": [r["team_key"] for r in portfolio_rows],
            "bid_amount": [r["bid_amount"] for r in portfolio_rows],
        }
    )

    if args.market_mode == "join":
        bids_all = pd.concat([market_bids, sim_rows], ignore_index=True)
        entry_points = _compute_entry_points(
            entry_bids=bids_all,
            points_by_team=points_by_team,
            calcutta_key=calcutta_key,
        )
    else:
        entry_points = _compute_entry_points(
            entry_bids=market_bids,
            points_by_team=points_by_team,
            calcutta_key=calcutta_key,
        )
        sim_points = _compute_sim_entry_points_shadow(
            sim_bids=sim_rows,
            market_entry_bids=market_bids,
            points_by_team=points_by_team,
            calcutta_key=calcutta_key,
            sim_entry_key=args.sim_entry_key,
        )
        entry_points = pd.concat(
            [entry_points, sim_points],
            ignore_index=True,
        )

    standings = _compute_finish_positions_and_payouts(
        entry_points=entry_points,
        payouts=tables["payouts"],
        calcutta_key=calcutta_key,
    )

    sim_row = standings[standings["entry_key"] == args.sim_entry_key]
    if len(sim_row) != 1:
        raise ValueError("failed to compute simulated entry standing")
    sim = sim_row.iloc[0]

    payout_cents = int(sim["payout_cents"])
    roi = (
        float(payout_cents) / (float(args.budget) * 100.0)
        if args.budget
        else 0.0
    )

    standings_with_names = standings.copy()
    entries = tables["entries"].copy()
    if "entry_key" in entries.columns and "entry_name" in entries.columns:
        standings_with_names = standings_with_names.merge(
            entries[["entry_key", "entry_name"]],
            on="entry_key",
            how="left",
        )

    out: Dict[str, object] = {
        "snapshot": snapshot_dir.name,
        "calcutta_key": calcutta_key,
        "constraints": {
            "budget": float(args.budget),
            "min_teams": int(args.min_teams),
            "max_teams": int(args.max_teams),
            "max_per_team": float(args.max_per_team),
            "min_bid": float(args.min_bid),
        },
        "score_col": str(args.score_col),
        "portfolio": portfolio_rows,
        "summary": {
            "n_teams": int(len(portfolio_rows)),
            "total_spend": float(sum(r["bid_amount"] for r in portfolio_rows)),
        },
        "realized": {
            "points_mode": points_mode,
            "market_mode": str(args.market_mode),
            "total_points": float(sim["total_points"]),
            "finish_position": int(sim["finish_position"]),
            "is_tied": bool(sim["is_tied"]),
            "payout_cents": payout_cents,
            "roi": roi,
        },
        "standings_top": standings_with_names.head(10).to_dict(
            orient="records"
        ),
    }

    if args.expected_sims and args.expected_sims > 0:
        points_by_round: Optional[Dict[int, float]] = None
        if points_mode == "round_scoring" and "round_scoring" in tables:
            rs = tables["round_scoring"].copy()
            rs = rs[rs["calcutta_key"] == calcutta_key].copy()
            rs["round"] = pd.to_numeric(rs["round"], errors="coerce")
            rs["points"] = pd.to_numeric(rs["points"], errors="coerce")
            rs = rs[rs["round"].notna() & rs["points"].notna()].copy()
            points_by_round = {
                int(r["round"]): float(r["points"])
                for _, r in rs.iterrows()
            }

        exp = _expected_simulation(
            tables=tables,
            calcutta_key=calcutta_key,
            market_bids=market_bids,
            sim_rows=sim_rows,
            sim_entry_key=args.sim_entry_key,
            market_mode=str(args.market_mode),
            points_mode=points_mode,
            points_by_round=points_by_round,
            n_sims=int(args.expected_sims),
            seed=int(args.expected_seed),
            kenpom_scale=float(args.kenpom_scale),
            budget=float(args.budget),
        )
        if exp:
            out["expected"] = exp

    out_json = json.dumps(out, indent=2) + "\n"
    if args.out_path:
        Path(args.out_path).write_text(out_json, encoding="utf-8")
    else:
        print(out_json, end="")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
