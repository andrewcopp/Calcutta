from __future__ import annotations

from typing import Dict, List, Optional, Tuple

import pandas as pd

from moneyball.utils import points


def expected_team_points_from_predicted_game_outcomes(
    *,
    predicted_game_outcomes: pd.DataFrame,
) -> pd.DataFrame:
    required = [
        "round_order",
        "team1_key",
        "team2_key",
        "p_matchup",
        "p_team1_wins_given_matchup",
        "p_team2_wins_given_matchup",
    ]
    missing = [c for c in required if c not in predicted_game_outcomes.columns]
    if missing:
        raise ValueError(
            "predicted_game_outcomes missing columns: " + ", ".join(missing)
        )

    df = predicted_game_outcomes.copy()
    df["round_order"] = pd.to_numeric(df["round_order"], errors="coerce")
    df = df[df["round_order"].notna()].copy()
    df["round_order"] = df["round_order"].astype(int)

    df["p_matchup"] = pd.to_numeric(df["p_matchup"], errors="coerce").fillna(0)
    df["p_team1_wins_given_matchup"] = pd.to_numeric(
        df["p_team1_wins_given_matchup"],
        errors="coerce",
    ).fillna(0)
    df["p_team2_wins_given_matchup"] = pd.to_numeric(
        df["p_team2_wins_given_matchup"],
        errors="coerce",
    ).fillna(0)

    df = df[(df["round_order"] >= 1) & (df["round_order"] <= 7)].copy()

    t1 = pd.DataFrame(
        {
            "team_key": df["team1_key"].astype(str),
            "round_order": df["round_order"].astype(int),
            "p_win": df["p_matchup"]
            * df["p_team1_wins_given_matchup"],
        }
    )
    t2 = pd.DataFrame(
        {
            "team_key": df["team2_key"].astype(str),
            "round_order": df["round_order"].astype(int),
            "p_win": df["p_matchup"]
            * df["p_team2_wins_given_matchup"],
        }
    )

    wins = pd.concat([t1, t2], ignore_index=True)
    wins = wins[wins["team_key"].astype(str).str.len() > 0].copy()
    wins["p_win"] = pd.to_numeric(wins["p_win"], errors="coerce").fillna(0)

    p_by_round = (
        wins.groupby(["team_key", "round_order"], as_index=False)["p_win"]
        .sum()
        .copy()
    )

    inc_by_round: Dict[int, float] = {}
    for r in range(1, 8):
        inc_by_round[int(r)] = float(points.team_points_fixed(int(r))) - float(
            points.team_points_fixed(int(r - 1))
        )

    p_by_round["points_inc"] = p_by_round["round_order"].apply(
        lambda ro: float(inc_by_round.get(int(ro), 0.0))
    )
    p_by_round["expected_points"] = p_by_round["p_win"] * p_by_round[
        "points_inc"
    ]

    out = (
        p_by_round.groupby("team_key", as_index=False)["expected_points"]
        .sum()
        .copy()
    )
    out = out.rename(columns={"expected_points": "expected_team_points"})
    out["expected_team_points"] = pd.to_numeric(
        out["expected_team_points"], errors="coerce"
    ).fillna(0.0)
    return out


def variance_team_points_from_predicted_game_outcomes(
    *,
    predicted_game_outcomes: pd.DataFrame,
) -> pd.DataFrame:
    """
    Calculate variance of team points from predicted game outcomes.

    Variance captures tail risk - high variance teams (longshots) have
    potential for extreme outcomes that help win in scenarios where
    favorites underperform.

    For each team, we calculate the variance of their total points across
    all possible tournament outcomes, weighted by probability.

    Returns:
        DataFrame with columns: team_key, variance_team_points
    """
    required = [
        "round_order",
        "team1_key",
        "team2_key",
        "p_matchup",
        "p_team1_wins_given_matchup",
        "p_team2_wins_given_matchup",
    ]
    missing = [c for c in required if c not in predicted_game_outcomes.columns]
    if missing:
        raise ValueError(
            "predicted_game_outcomes missing columns: " + ", ".join(missing)
        )

    df = predicted_game_outcomes.copy()
    df["round_order"] = pd.to_numeric(df["round_order"], errors="coerce")
    df = df[df["round_order"].notna()].copy()
    df["round_order"] = df["round_order"].astype(int)

    df["p_matchup"] = pd.to_numeric(df["p_matchup"], errors="coerce").fillna(0)
    df["p_team1_wins_given_matchup"] = pd.to_numeric(
        df["p_team1_wins_given_matchup"],
        errors="coerce",
    ).fillna(0)
    df["p_team2_wins_given_matchup"] = pd.to_numeric(
        df["p_team2_wins_given_matchup"],
        errors="coerce",
    ).fillna(0)

    df = df[(df["round_order"] >= 1) & (df["round_order"] <= 7)].copy()

    t1 = pd.DataFrame(
        {
            "team_key": df["team1_key"].astype(str),
            "round_order": df["round_order"].astype(int),
            "p_win": df["p_matchup"]
            * df["p_team1_wins_given_matchup"],
        }
    )
    t2 = pd.DataFrame(
        {
            "team_key": df["team2_key"].astype(str),
            "round_order": df["round_order"].astype(int),
            "p_win": df["p_matchup"]
            * df["p_team2_wins_given_matchup"],
        }
    )

    wins = pd.concat([t1, t2], ignore_index=True)
    wins = wins[wins["team_key"].astype(str).str.len() > 0].copy()
    wins["p_win"] = pd.to_numeric(wins["p_win"], errors="coerce").fillna(0)

    p_by_round = (
        wins.groupby(["team_key", "round_order"], as_index=False)["p_win"]
        .sum()
        .copy()
    )

    inc_by_round: Dict[int, float] = {}
    for r in range(1, 8):
        inc_by_round[int(r)] = float(points.team_points_fixed(int(r))) - float(
            points.team_points_fixed(int(r - 1))
        )

    p_by_round["points_inc"] = p_by_round["round_order"].apply(
        lambda ro: float(inc_by_round.get(int(ro), 0.0))
    )

    p_by_round["points_inc_sq"] = p_by_round["points_inc"] ** 2
    p_by_round["var_contribution"] = (
        p_by_round["p_win"] * p_by_round["points_inc_sq"]
    )

    var_by_team = (
        p_by_round.groupby("team_key", as_index=False)["var_contribution"]
        .sum()
        .copy()
    )

    expected_by_team = (
        p_by_round.groupby("team_key", as_index=False)
        .apply(lambda g: (g["p_win"] * g["points_inc"]).sum())
        .reset_index()
        .rename(columns={0: "expected_points"})
    )

    result = var_by_team.merge(expected_by_team, on="team_key", how="left")
    result["expected_points"] = result["expected_points"].fillna(0.0)

    result["variance_team_points"] = (
        result["var_contribution"] - result["expected_points"] ** 2
    )
    result["variance_team_points"] = (
        result["variance_team_points"].clip(lower=0.0)
    )

    out = result[["team_key", "variance_team_points"]].copy()
    out["variance_team_points"] = pd.to_numeric(
        out["variance_team_points"], errors="coerce"
    ).fillna(0.0)

    return out


def _optimize_portfolio_greedy(
    *,
    df: pd.DataFrame,
    score_col: str,
    budget: float,
    min_teams: int,
    max_teams: int,
    max_per_team: float,
    min_bid: float,
    step: float = 1.0,
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

    b_int = int(round(float(budget)))
    min_int = int(round(float(min_bid)))
    max_int = int(round(float(max_per_team)))
    if abs(float(budget) - float(b_int)) > 1e-9:
        raise ValueError("budget must be an integer number of dollars")
    if abs(float(min_bid) - float(min_int)) > 1e-9:
        raise ValueError("min_bid must be an integer number of dollars")
    if abs(float(max_per_team) - float(max_int)) > 1e-9:
        raise ValueError("max_per_team must be an integer number of dollars")

    step = 1.0

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

    if float(min_teams) * float(min_int) - float(b_int) > 1e-9:
        raise ValueError("budget too small to satisfy min_teams at min_bid")

    bids: List[float] = [0.0 for _ in range(n)]
    selected: set[int] = set()

    def _delta_for(i: int, b0: float, inc: float) -> float:
        if inc <= 0:
            return 0.0
        if b0 + inc - float(max_int) > 1e-9:
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

    remaining = float(b_int)

    while len(selected) < int(min_teams):
        best_i: Optional[int] = None
        best_v = -1e99
        for i in range(n):
            if i in selected:
                continue
            v = _delta_for(i, 0.0, float(min_int)) / float(min_int)
            if v > best_v:
                best_v = v
                best_i = i
        if best_i is None:
            break
        bids[best_i] = float(min_int)
        selected.add(best_i)
        remaining -= float(min_int)

    while remaining > 1e-9:
        best_i: Optional[int] = None
        best_inc: float = 0.0
        best_val = -1e99

        for i in selected:
            inc = float(step) if remaining >= step else float(remaining)
            if bids[i] + inc - float(max_int) > 1e-9:
                inc = max(0.0, float(max_int) - float(bids[i]))
            if inc <= 1e-12:
                continue
            v = _delta_for(i, float(bids[i]), float(inc)) / float(inc)
            if v > best_val:
                best_val = v
                best_i = i
                best_inc = float(inc)

        if (
            len(selected) < int(max_teams)
            and remaining + 1e-9 >= float(min_int)
        ):
            for i in range(n):
                if i in selected:
                    continue
                if float(min_int) - float(max_int) > 1e-9:
                    continue
                v = _delta_for(i, 0.0, float(min_int)) / float(min_int)
                if v > best_val:
                    best_val = v
                    best_i = i
                    best_inc = float(min_int)

        if best_i is None or best_inc <= 1e-12:
            break

        if best_i not in selected and abs(best_inc - float(min_int)) < 1e-9:
            selected.add(best_i)
        bids[best_i] += float(best_inc)
        remaining -= float(best_inc)

    chosen = pool.loc[sorted(selected)].copy().reset_index(drop=True)
    chosen_bids = [int(round(float(bids[i]))) for i in sorted(selected)]
    if any(x < int(min_int) for x in chosen_bids):
        raise ValueError("allocation violates min_bid constraint")
    if any(x > int(max_int) for x in chosen_bids):
        raise ValueError("allocation violates max_per_team constraint")
    if int(sum(chosen_bids)) != int(b_int):
        delta = int(b_int) - int(sum(chosen_bids))
        chosen_bids[0] += delta
    chosen["bid_amount_points"] = [int(x) for x in chosen_bids]

    portfolio_rows: List[Dict[str, object]] = []
    for _, r in chosen.iterrows():
        portfolio_rows.append(
            {
                "team_key": str(r["team_key"]),
                "bid_amount_points": int(r["bid_amount_points"]),
                "score": float(r.get(score_col, 0.0) or 0.0),
            }
        )
    return chosen, portfolio_rows


def recommend_entry_bids(
    *,
    predicted_auction_share_of_pool: pd.DataFrame,
    predicted_game_outcomes: pd.DataFrame,
    predicted_total_pool_bids_points: float,
    budget_points: int = 100,
    min_teams: int = 3,
    max_teams: int = 10,
    max_per_team_points: int = 50,
    min_bid_points: int = 1,
    strategy: str = "minlp",
    variance_weight: float = 0.0,
) -> pd.DataFrame:
    if predicted_total_pool_bids_points <= 0:
        raise ValueError("predicted_total_pool_bids_points must be positive")

    if "team_key" not in predicted_auction_share_of_pool.columns:
        raise ValueError("predicted_auction_share_of_pool missing team_key")
    if (
        "predicted_auction_share_of_pool"
        not in predicted_auction_share_of_pool.columns
    ):
        raise ValueError(
            "predicted_auction_share_of_pool missing "
            "predicted_auction_share_of_pool"
        )

    shares = predicted_auction_share_of_pool.copy()
    shares["team_key"] = shares["team_key"].astype(str)
    shares["predicted_auction_share_of_pool"] = pd.to_numeric(
        shares["predicted_auction_share_of_pool"],
        errors="coerce",
    ).fillna(0.0)

    pts = expected_team_points_from_predicted_game_outcomes(
        predicted_game_outcomes=predicted_game_outcomes
    )

    # Preserve region column if it exists for region-constrained strategies
    df = shares.merge(pts, on="team_key", how="left")
    df["expected_team_points"] = pd.to_numeric(
        df["expected_team_points"], errors="coerce"
    ).fillna(0.0)

    if variance_weight > 0:
        var = variance_team_points_from_predicted_game_outcomes(
            predicted_game_outcomes=predicted_game_outcomes
        )
        df = df.merge(var, on="team_key", how="left")
        df["variance_team_points"] = pd.to_numeric(
            df["variance_team_points"], errors="coerce"
        ).fillna(0.0)
        df["std_team_points"] = df["variance_team_points"] ** 0.5
    else:
        df["variance_team_points"] = 0.0
        df["std_team_points"] = 0.0

    df["predicted_team_total_bids"] = df[
        "predicted_auction_share_of_pool"
    ].apply(lambda s: float(s) * float(predicted_total_pool_bids_points))

    df = df.sort_values(by=["team_key"]).reset_index(drop=True)

    def _ppd_at_min_bid(r: pd.Series) -> float:
        exp_pts = float(r.get("expected_team_points") or 0.0)
        std_pts = float(r.get("std_team_points") or 0.0)
        m = float(r.get("predicted_team_total_bids") or 0.0)
        if m < 0:
            m = 0.0
        denom = m + float(min_bid_points)
        if denom <= 0:
            return 0.0
        value_with_variance = exp_pts + variance_weight * std_pts
        return value_with_variance / denom

    df["score"] = df.apply(_ppd_at_min_bid, axis=1)

    # Use strategy-specific allocation
    if strategy == "greedy":
        chosen, _rows = _optimize_portfolio_greedy(
            df=df,
            score_col="score",
            budget=float(int(budget_points)),
            min_teams=int(min_teams),
            max_teams=int(max_teams),
            max_per_team=float(int(max_per_team_points)),
            min_bid=float(int(min_bid_points)),
        )
    else:
        # Use portfolio_strategies module for other strategies
        from moneyball.models.portfolio_strategies import get_strategy
        
        strategy_func = get_strategy(strategy)
        chosen = strategy_func(
            teams_df=df,
            budget_points=int(budget_points),
            min_teams=int(min_teams),
            max_teams=int(max_teams),
            max_per_team_points=int(max_per_team_points),
            min_bid_points=int(min_bid_points),
        )

    out_cols = [
        "team_key",
        "bid_amount_points",
        "expected_team_points",
        "predicted_team_total_bids",
        "predicted_auction_share_of_pool",
        "score",
    ]
    out = chosen[out_cols].copy()
    out["bid_amount_points"] = pd.to_numeric(
        out["bid_amount_points"], errors="coerce"
    ).fillna(0).astype(int)
    out = out.sort_values(
        by=["bid_amount_points", "team_key"],
        ascending=[False, True],
    )
    out = out.reset_index(drop=True)

    if int(out["bid_amount_points"].sum()) != int(budget_points):
        raise ValueError("recommended bids do not sum to budget")

    return out
