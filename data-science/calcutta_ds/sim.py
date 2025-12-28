import random
from typing import Dict, List, Optional, Tuple

import numpy as np
import pandas as pd

from calcutta_ds.bracket import prepare_bracket_graph, win_prob
from calcutta_ds.points import (
    team_points_fixed,
    team_points_from_round_scoring,
)
from calcutta_ds.standings import (
    compute_entry_points,
    compute_finish_positions_and_payouts,
    ensure_entry_points_include_competitors,
)


def expected_simulation(
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    market_bids: pd.DataFrame,
    sim_rows: pd.DataFrame,
    sim_entry_key: str,
    points_mode: str,
    points_by_round: Optional[Dict[int, float]],
    n_sims: int,
    seed: int,
    kenpom_scale: float,
    budget: float,
    use_historical_winners: bool,
    competitor_entry_keys: Optional[List[str]] = None,
) -> Dict[str, object]:
    if n_sims <= 0:
        return {}
    if "games" not in tables:
        raise ValueError("expected simulation requires games.parquet")

    games, prev_by_next = prepare_bracket_graph(tables["games"])

    teams = tables["teams"].copy()
    for c in ["wins", "byes"]:
        teams[c] = (
            pd.to_numeric(teams[c], errors="coerce")
            .fillna(0)
            .astype(int)
        )

    if "kenpom_net" not in teams.columns:
        raise ValueError(
            "teams.parquet missing kenpom_net (needed for simulation)"
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
    payout_rois: List[float] = []
    points_list: List[float] = []
    finish_positions: List[int] = []
    team_points_sums: Dict[str, float] = {}

    for _ in range(n_sims):
        wins_sim: Dict[str, int] = {}
        winners_by_game: Dict[str, str] = {}

        for _, gr in games.iterrows():
            gid = str(gr.get("game_id"))

            if use_historical_winners:
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
            p1 = win_prob(net1, net2, kenpom_scale)
            w = t1 if rng.random() < p1 else t2
            winners_by_game[gid] = w
            wins_sim[w] = wins_sim.get(w, 0) + 1

        team_points: Dict[str, float] = {}
        for team_key, byes in byes_by_team.items():
            progress = int(byes) + int(wins_sim.get(team_key, 0))
            if points_mode == "round_scoring" and points_by_round is not None:
                team_points[team_key] = team_points_from_round_scoring(
                    progress,
                    points_by_round,
                )
            else:
                team_points[team_key] = team_points_fixed(progress)

        for team_key, pts in team_points.items():
            team_points_sums[team_key] = (
                team_points_sums.get(team_key, 0.0) + float(pts)
            )

        points_by_team_df = pd.DataFrame(
            {
                "team_key": list(team_points.keys()),
                "team_points": list(team_points.values()),
            }
        )
        bids_all = pd.concat([market_bids, sim_rows], ignore_index=True)
        entry_points = compute_entry_points(
            entry_bids=bids_all,
            points_by_team=points_by_team_df,
            calcutta_key=calcutta_key,
        )

        if competitor_entry_keys is not None:
            allowed = set(str(k) for k in competitor_entry_keys)
            allowed.add(str(sim_entry_key))
            entry_points = entry_points[
                entry_points["entry_key"].astype(str).isin(allowed)
            ].copy()
            entry_points = ensure_entry_points_include_competitors(
                entry_points,
                competitor_entry_keys=list(allowed),
            )

        standings = compute_finish_positions_and_payouts(
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
            rois.append(total_points / float(budget))
            payout_rois.append(payout / (float(budget) * 100.0))
        else:
            rois.append(0.0)
            payout_rois.append(0.0)

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
        "mean_payout_per_fake_dollar": float(
            sum(payout_rois) / len(payout_rois)
        ),
        "mean_total_points": float(sum(points_list) / len(points_list)),
        "p50_payout_cents": _pct([float(x) for x in payouts], 0.50),
        "p90_payout_cents": _pct([float(x) for x in payouts], 0.90),
        "p50_roi": _pct(rois, 0.50),
        "p90_roi": _pct(rois, 0.90),
        "p50_payout_per_fake_dollar": _pct(payout_rois, 0.50),
        "p90_payout_per_fake_dollar": _pct(payout_rois, 0.90),
        "p50_total_points": _pct(points_list, 0.50),
        "p90_total_points": _pct(points_list, 0.90),
        "mean_finish_position": float(
            sum(finish_positions) / len(finish_positions)
        ),
        "p50_finish_position": _pct(
            [float(x) for x in finish_positions],
            0.50,
        ),
        "p90_finish_position": _pct(
            [float(x) for x in finish_positions],
            0.90,
        ),
        "p_top1": 0.0,
        "p_top3": 0.0,
        "p_top6": 0.0,
        "p_top10": 0.0,
        "finish_position_counts": {},
        "team_mean_points": {},
    }

    team_means: Dict[str, float] = {}
    denom = float(len(payouts))
    if denom > 0:
        for team_key, s in team_points_sums.items():
            team_means[str(team_key)] = float(s) / denom
    expected["team_mean_points"] = team_means

    counts: Dict[str, int] = {}
    for fp in finish_positions:
        k = str(int(fp))
        counts[k] = counts.get(k, 0) + 1
    expected["finish_position_counts"] = counts

    denom_p = float(len(finish_positions))
    if denom_p > 0:
        expected["p_top1"] = float(
            sum(1 for fp in finish_positions if int(fp) <= 1) / denom_p
        )
        expected["p_top3"] = float(
            sum(1 for fp in finish_positions if int(fp) <= 3) / denom_p
        )
        expected["p_top6"] = float(
            sum(1 for fp in finish_positions if int(fp) <= 6) / denom_p
        )
        expected["p_top10"] = float(
            sum(1 for fp in finish_positions if int(fp) <= 10) / denom_p
        )
    return expected


def simulate_team_points_scenarios(
    tables: Dict[str, pd.DataFrame],
    calcutta_key: str,
    points_mode: str,
    points_by_round: Optional[Dict[int, float]],
    n_sims: int,
    seed: int,
    kenpom_scale: float,
    use_historical_winners: bool,
) -> Tuple[List[str], np.ndarray]:
    if n_sims <= 0:
        return [], np.zeros((0, 0), dtype=float)
    if "games" not in tables:
        raise ValueError("expected simulation requires games.parquet")

    games, prev_by_next = prepare_bracket_graph(tables["games"])

    teams = tables["teams"].copy()
    if "calcutta_key" in teams.columns:
        teams = teams[teams["calcutta_key"] == calcutta_key].copy()
    for c in ["wins", "byes"]:
        teams[c] = (
            pd.to_numeric(teams[c], errors="coerce")
            .fillna(0)
            .astype(int)
        )
    if "kenpom_net" not in teams.columns:
        raise ValueError(
            "teams.parquet missing kenpom_net (needed for simulation)"
        )
    teams["kenpom_net"] = pd.to_numeric(teams["kenpom_net"], errors="coerce")

    team_keys: List[str] = []
    net_by_team: Dict[str, float] = {}
    byes_by_team: Dict[str, int] = {}
    for _, r in teams.iterrows():
        tk = str(r.get("team_key") or "")
        if not tk:
            continue
        net = r.get("kenpom_net")
        if pd.isna(net):
            continue
        team_keys.append(tk)
        net_by_team[tk] = float(net)
        byes_by_team[tk] = int(r.get("byes") or 0)
    team_keys = sorted(set(team_keys))
    if not team_keys:
        return [], np.zeros((0, 0), dtype=float)
    t_index = {tk: i for i, tk in enumerate(team_keys)}

    rng = random.Random(seed)
    out = np.zeros((int(n_sims), int(len(team_keys))), dtype=float)

    for s in range(int(n_sims)):
        wins_sim: Dict[str, int] = {}
        winners_by_game: Dict[str, str] = {}

        for _, gr in games.iterrows():
            gid = str(gr.get("game_id"))

            if use_historical_winners:
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
            p1 = win_prob(net1, net2, kenpom_scale)
            w = t1 if rng.random() < p1 else t2
            winners_by_game[gid] = w
            wins_sim[w] = wins_sim.get(w, 0) + 1

        for team_key in team_keys:
            byes = int(byes_by_team.get(team_key, 0))
            progress = int(byes) + int(wins_sim.get(team_key, 0))
            if points_mode == "round_scoring" and points_by_round is not None:
                pts = team_points_from_round_scoring(
                    progress,
                    points_by_round,
                )
            else:
                pts = team_points_fixed(progress)
            out[s, t_index[team_key]] = float(pts)

    return team_keys, out
