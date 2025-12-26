import argparse
import json
import math
from pathlib import Path
from typing import Dict, List, Tuple

import numpy as np
import pandas as pd

import backtest_scaffold


def _sigmoid(x: float) -> float:
    if x >= 0:
        z = math.exp(-x)
        return 1.0 / (1.0 + z)
    z = math.exp(x)
    return z / (1.0 + z)


def _log_loss(probs: np.ndarray, y: np.ndarray) -> float:
    eps = 1e-12
    p = np.clip(probs, eps, 1.0 - eps)
    return float(-(y * np.log(p) + (1.0 - y) * np.log(1.0 - p)).mean())


def _find_snapshots(out_root: Path) -> List[Path]:
    if not out_root.exists():
        return []

    snaps: List[Path] = []
    for p in sorted(out_root.iterdir()):
        if not p.is_dir():
            continue
        if (p / "games.parquet").exists() and (p / "teams.parquet").exists():
            snaps.append(p)
    return snaps


def _load_game_rows(snapshot_dir: Path) -> pd.DataFrame:
    games = pd.read_parquet(snapshot_dir / "games.parquet")
    teams = pd.read_parquet(snapshot_dir / "teams.parquet")

    required_games = [
        "game_id",
        "team1_key",
        "team2_key",
        "winner_team_key",
        "round",
        "sort_order",
        "next_game_id",
        "next_game_slot",
    ]
    missing = [c for c in required_games if c not in games.columns]
    if missing:
        raise ValueError(f"games.parquet missing columns: {missing}")

    if "kenpom_net" not in teams.columns:
        raise ValueError("teams.parquet missing kenpom_net")
    if "team_key" not in teams.columns:
        raise ValueError("teams.parquet missing team_key")

    teams = teams.copy()
    teams["kenpom_net"] = pd.to_numeric(teams["kenpom_net"], errors="coerce")
    teams = teams[teams["kenpom_net"].notna()].copy()
    net_by_team: Dict[str, float] = {
        str(r["team_key"]): float(r["kenpom_net"])
        for _, r in teams[["team_key", "kenpom_net"]].iterrows()
        if str(r["team_key"])
    }

    g_all = games[required_games].copy()
    g_all = g_all[g_all["winner_team_key"].notna()].copy()
    g_all["team1_key"] = g_all["team1_key"].fillna("")
    g_all["team2_key"] = g_all["team2_key"].fillna("")
    g_all["winner_team_key"] = g_all["winner_team_key"].fillna("")

    g_sorted, prev_by_next = backtest_scaffold._prepare_bracket_graph(g_all)

    winners_by_game: Dict[str, str] = {}
    rows: List[Dict[str, object]] = []
    for _, gr in g_sorted.iterrows():
        gid = str(gr.get("game_id") or "")
        if not gid:
            continue

        winner = str(gr.get("winner_team_key") or "")
        if winner:
            winners_by_game[gid] = winner

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

        if not t1 or not t2 or not winner:
            continue
        if winner not in (t1, t2):
            continue

        net1 = net_by_team.get(t1)
        net2 = net_by_team.get(t2)
        if net1 is None or net2 is None:
            continue

        y = 1 if winner == t1 else 0
        rows.append(
            {
                "game_id": gid,
                "diff": float(net1) - float(net2),
                "y": int(y),
            }
        )

    return pd.DataFrame(rows)


def _score_for_scale(diff: np.ndarray, y: np.ndarray, scale: float) -> float:
    if scale <= 0:
        return float("inf")
    probs = np.array([_sigmoid(float(d) / scale) for d in diff], dtype=float)
    return _log_loss(probs, y)


def _fit_scale(diff: np.ndarray, y: np.ndarray) -> Tuple[float, float]:
    # We optimize in log-space with a ternary search over a reasonable range.
    # Typical rating diffs are on the order of ~0-30.
    lo = math.log(0.5)
    hi = math.log(50.0)

    for _ in range(60):
        m1 = lo + (hi - lo) / 3.0
        m2 = hi - (hi - lo) / 3.0
        s1 = math.exp(m1)
        s2 = math.exp(m2)
        f1 = _score_for_scale(diff, y, s1)
        f2 = _score_for_scale(diff, y, s2)
        if f1 <= f2:
            hi = m2
        else:
            lo = m1

    best = math.exp((lo + hi) / 2.0)
    loss = _score_for_scale(diff, y, best)
    return float(best), float(loss)


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Calibrate a global kenpom_scale for win probability mapping: "
            "P(team1 wins)=sigmoid((net1-net2)/scale)."
        )
    )
    parser.add_argument(
        "out_root",
        help="Option-A out-root containing snapshot dirs",
    )
    parser.add_argument(
        "--out",
        dest="out_path",
        default=None,
        help="Write JSON results to this path (default: stdout)",
    )

    args = parser.parse_args()

    out_root = Path(args.out_root)
    snapshots = _find_snapshots(out_root)
    if not snapshots:
        raise FileNotFoundError(
            f"no snapshots with games+teams found under: {out_root}"
        )

    rows: List[pd.DataFrame] = []
    for s in snapshots:
        try:
            df = _load_game_rows(s)
        except Exception:
            continue
        if len(df) > 0:
            df["snapshot"] = s.name
            rows.append(df)

    if not rows:
        raise ValueError(
            "no usable game rows (need winner_team_key and kenpom_net)"
        )

    all_rows = pd.concat(rows, ignore_index=True)
    diff = all_rows["diff"].to_numpy(dtype=float)
    y = all_rows["y"].to_numpy(dtype=float)

    scale, loss = _fit_scale(diff, y)

    probs = np.array([_sigmoid(float(d) / scale) for d in diff], dtype=float)
    preds = (probs >= 0.5).astype(int)
    acc = float((preds == y.astype(int)).mean())

    out: Dict[str, object] = {
        "n_games": int(len(all_rows)),
        "n_snapshots": int(all_rows["snapshot"].nunique()),
        "kenpom_scale": float(scale),
        "log_loss": float(loss),
        "accuracy": acc,
        "notes": (
            "Fit is global across all snapshots containing games.parquet + "
            "teams.parquet. "
            "Only games with winner_team_key and both kenpom_net values are "
            "used."
        ),
    }

    out_json = json.dumps(out, indent=2) + "\n"
    if args.out_path:
        Path(args.out_path).write_text(out_json, encoding="utf-8")
    else:
        print(out_json, end="")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
