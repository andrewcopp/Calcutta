from __future__ import annotations

import random
from dataclasses import dataclass
from typing import Dict, List, Optional, Tuple

import pandas as pd

from moneyball.utils import bracket
from moneyball.utils.points import team_points_fixed
from moneyball.utils.standings import (
    compute_entry_points,
    compute_finish_positions_and_payouts,
)


@dataclass(frozen=True)
class SimulatedEntryOutcomesConfig:
    calcutta_key: str
    n_sims: int
    seed: int
    budget_points: int


def _pct(xs: List[float], p: float) -> float:
    if not xs:
        return 0.0
    s = sorted(xs)
    idx = int(round((len(s) - 1) * float(p)))
    return float(s[idx])


def _win_prob_from_predicted_game_outcomes(
    *,
    predicted_game_outcomes: pd.DataFrame,
    game_id: str,
    team1_key: str,
    team2_key: str,
) -> float:
    g = predicted_game_outcomes
    m = g[(g["game_id"].astype(str) == str(game_id))].copy()
    if m.empty:
        return 0.5

    direct = m[
        (m["team1_key"].astype(str) == str(team1_key))
        & (m["team2_key"].astype(str) == str(team2_key))
    ]
    if len(direct) == 1:
        p = direct.iloc[0].get("p_team1_wins_given_matchup")
        try:
            return float(p)
        except Exception:
            return 0.5

    swapped = m[
        (m["team1_key"].astype(str) == str(team2_key))
        & (m["team2_key"].astype(str) == str(team1_key))
    ]
    if len(swapped) == 1:
        p = swapped.iloc[0].get("p_team2_wins_given_matchup")
        try:
            return float(p)
        except Exception:
            return 0.5

    return 0.5


def simulate_entry_outcomes(
    *,
    games: pd.DataFrame,
    teams: pd.DataFrame,
    payouts: pd.DataFrame,
    entry_bids: pd.DataFrame,
    predicted_game_outcomes: pd.DataFrame,
    recommended_entry_bids: pd.DataFrame,
    simulated_tournaments: Optional[pd.DataFrame],
    calcutta_key: str,
    n_sims: int,
    seed: int,
    budget_points: int,
    sim_entry_key: str = "simulated:entry",
    keep_sims: bool = False,
) -> Tuple[pd.DataFrame, Optional[pd.DataFrame]]:
    if n_sims <= 0:
        raise ValueError("n_sims must be positive")
    if budget_points <= 0:
        raise ValueError("budget_points must be positive")

    # If simulated_tournaments is provided, use cached simulations
    use_cached = (
        simulated_tournaments is not None and
        not simulated_tournaments.empty
    )
    if not use_cached:
        required_go = [
            "game_id",
            "team1_key",
            "team2_key",
            "p_team1_wins_given_matchup",
            "p_team2_wins_given_matchup",
        ]
        missing_go = [
            c
            for c in required_go
            if c not in predicted_game_outcomes.columns
        ]
        if missing_go:
            raise ValueError(
                "predicted_game_outcomes missing columns: "
                + ", ".join(missing_go)
            )
    else:
        # Validate cached tournaments
        required_cached = ["sim_id", "team_key", "wins"]
        missing_cached = [
            c for c in required_cached
            if c not in simulated_tournaments.columns
        ]
        if missing_cached:
            raise ValueError(
                "simulated_tournaments missing columns: "
                + ", ".join(missing_cached)
            )

    required_bids = ["team_key", "bid_amount_points"]
    missing_bids = [
        c
        for c in required_bids
        if c not in recommended_entry_bids.columns
    ]
    if missing_bids:
        raise ValueError(
            "recommended_entry_bids missing columns: "
            + ", ".join(missing_bids)
        )

    teams_f = teams.copy()
    if "calcutta_key" in teams_f.columns:
        teams_f = teams_f[teams_f["calcutta_key"] == calcutta_key].copy()

    if "team_key" not in teams_f.columns:
        raise ValueError("teams missing team_key")
    for c in ["wins", "byes"]:
        if c not in teams_f.columns:
            raise ValueError(f"teams missing {c}")

    teams_f["byes"] = pd.to_numeric(teams_f["byes"], errors="coerce").fillna(0)
    byes_by_team = {
        str(r.get("team_key") or ""): int(r.get("byes") or 0)
        for _, r in teams_f.iterrows()
        if str(r.get("team_key") or "")
    }

    games_graph, prev_by_next = bracket.prepare_bracket_graph(games)

    sim_rows = recommended_entry_bids[
        ["team_key", "bid_amount_points"]
    ].copy()
    sim_rows["calcutta_key"] = str(calcutta_key)
    sim_rows["entry_key"] = str(sim_entry_key)
    sim_rows["team_key"] = sim_rows["team_key"].astype(str)
    sim_rows["bid_amount"] = pd.to_numeric(
        sim_rows["bid_amount_points"],
        errors="coerce",
    ).fillna(0.0)
    sim_rows = sim_rows[
        ["calcutta_key", "entry_key", "team_key", "bid_amount"]
    ]

    market_bids = entry_bids.copy()

    rng = random.Random(int(seed))
    payouts_cents: List[int] = []
    normalized_payouts: List[float] = []
    total_points_list: List[float] = []
    finish_positions: List[int] = []
    is_tied_list: List[bool] = []
    n_entries_list: List[int] = []

    sims_rows: List[Dict[str, object]] = []

    for sim_i in range(int(n_sims)):
        if use_cached:
            # Use cached tournament results
            sim_data = simulated_tournaments[
                simulated_tournaments["sim_id"] == sim_i
            ]
            wins_sim = {
                str(row["team_key"]): int(row["wins"])
                for _, row in sim_data.iterrows()
            }
        else:
            # Simulate games from scratch
            wins_sim: Dict[str, int] = {}
            winners_by_game: Dict[str, str] = {}

            for _, gr in games_graph.iterrows():
                gid = str(gr.get("game_id") or "")
                if not gid:
                    continue

                t1 = str(gr.get("team1_key") or "")
                t2 = str(gr.get("team2_key") or "")

                if int(gr.get("round_order") or 999) > 2:
                    t1 = ""
                    t2 = ""

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

                p1 = _win_prob_from_predicted_game_outcomes(
                    predicted_game_outcomes=predicted_game_outcomes,
                    game_id=gid,
                    team1_key=t1,
                    team2_key=t2,
                )
                if p1 < 0.0:
                    p1 = 0.0
                if p1 > 1.0:
                    p1 = 1.0

                w = t1 if rng.random() < float(p1) else t2
                winners_by_game[gid] = w
                wins_sim[w] = wins_sim.get(w, 0) + 1

        team_points = []
        for team_key, byes in byes_by_team.items():
            progress = int(byes) + int(wins_sim.get(team_key, 0))
            team_points.append(
                {
                    "team_key": str(team_key),
                    "team_points": float(team_points_fixed(int(progress))),
                }
            )

        points_by_team_df = pd.DataFrame(team_points)

        bids_all = pd.concat([market_bids, sim_rows], ignore_index=True)
        entry_points = compute_entry_points(
            entry_bids=bids_all,
            points_by_team=points_by_team_df,
            calcutta_key=str(calcutta_key),
        )

        standings = compute_finish_positions_and_payouts(
            entry_points=entry_points,
            payouts=payouts,
            calcutta_key=str(calcutta_key),
        )

        sim_row = standings[
            standings["entry_key"].astype(str) == str(sim_entry_key)
        ]
        if len(sim_row) != 1:
            continue
        sim_r = sim_row.iloc[0]

        payout = int(sim_r.get("payout_cents") or 0)
        total_points = float(sim_r.get("total_points") or 0.0)
        finish_pos = int(sim_r.get("finish_position") or 0)
        is_tied = bool(sim_r.get("is_tied"))

        n_entries = len(standings)

        # Calculate normalized payout by maximum prize money
        # 1.0 = winning max prize ($650), 0.5 = winning half of max prize ($325)
        # This is independent of pool size and comparable across years
        max_payout_cents = payouts["amount_cents"].max()
        normalized_payout = (
            float(payout) / float(max_payout_cents)
            if max_payout_cents > 0
            else 0.0
        )

        payouts_cents.append(payout)
        normalized_payouts.append(normalized_payout)
        total_points_list.append(total_points)
        finish_positions.append(finish_pos)
        is_tied_list.append(is_tied)
        n_entries_list.append(n_entries)

        if keep_sims:
            sims_rows.append(
                {
                    "sim": int(sim_i),
                    "payout_cents": int(payout),
                    "normalized_payout": float(normalized_payout),
                    "total_points": float(total_points),
                    "finish_position": int(finish_pos),
                    "is_tied": bool(is_tied),
                }
            )

    if not payouts_cents:
        raise ValueError("simulation produced no valid simulated entry rows")

    denom = float(len(payouts_cents))
    p_top1 = float(
        sum(1 for fp in finish_positions if int(fp) <= 1) / denom
    )
    p_top3 = float(
        sum(1 for fp in finish_positions if int(fp) <= 3) / denom
    )
    p_top6 = float(
        sum(1 for fp in finish_positions if int(fp) <= 6) / denom
    )
    p_top10 = float(
        sum(1 for fp in finish_positions if int(fp) <= 10) / denom
    )
    p_in_money = float(
        sum(1 for p in payouts_cents if int(p) > 0) / denom
    )

    payout_per_fake_dollar = [
        float(p) / (float(budget_points) * 100.0) for p in payouts_cents
    ]

    mean_n_entries = (
        float(sum(n_entries_list) / denom) if n_entries_list else 0.0
    )

    summary = {
        "entry_key": str(sim_entry_key),
        "sims": int(len(payouts_cents)),
        "seed": int(seed),
        "budget_points": int(budget_points),
        "mean_payout_cents": float(sum(payouts_cents) / denom),
        "p50_payout_cents": _pct([float(x) for x in payouts_cents], 0.50),
        "p90_payout_cents": _pct([float(x) for x in payouts_cents], 0.90),
        "mean_normalized_payout": float(sum(normalized_payouts) / denom),
        "p50_normalized_payout": _pct(normalized_payouts, 0.50),
        "p90_normalized_payout": _pct(normalized_payouts, 0.90),
        "mean_total_points": float(sum(total_points_list) / denom),
        "p50_total_points": _pct(total_points_list, 0.50),
        "p90_total_points": _pct(total_points_list, 0.90),
        "mean_finish_position": float(sum(finish_positions) / denom),
        "p50_finish_position": _pct(
            [float(x) for x in finish_positions],
            0.50,
        ),
        "p90_finish_position": _pct(
            [float(x) for x in finish_positions],
            0.90,
        ),
        "mean_n_entries": mean_n_entries,
        "p_top1": float(p_top1),
        "p_top3": float(p_top3),
        "p_top6": float(p_top6),
        "p_top10": float(p_top10),
        "p_in_money": float(p_in_money),
        "mean_payout_per_fake_dollar": float(
            sum(payout_per_fake_dollar) / denom
        ),
        "p50_payout_per_fake_dollar": _pct(payout_per_fake_dollar, 0.50),
        "p90_payout_per_fake_dollar": _pct(payout_per_fake_dollar, 0.90),
    }

    summary_df = pd.DataFrame([summary])
    sims_df = pd.DataFrame(sims_rows) if keep_sims else None
    return summary_df, sims_df
