from __future__ import annotations

from dataclasses import dataclass
from typing import Callable, Dict, Iterable, List

import pandas as pd


@dataclass(frozen=True)
class ArtifactContract:
    name: str
    required_columns: List[str]
    validators: List[Callable[[pd.DataFrame], None]]


def _require_columns(
    df: pd.DataFrame,
    required: Iterable[str],
    *,
    name: str,
) -> None:
    missing = [c for c in required if c not in df.columns]
    if missing:
        raise ValueError(f"{name} missing columns: {missing}")


def _require_in_0_1(series: pd.Series, *, col: str) -> None:
    v = pd.to_numeric(series, errors="coerce")
    if v.isna().any():
        raise ValueError(f"{col} contains non-numeric values")
    if (v < 0.0).any() or (v > 1.0).any():
        raise ValueError(f"{col} must be in [0, 1]")


def _require_finite_non_negative(series: pd.Series, *, col: str) -> None:
    v = pd.to_numeric(series, errors="coerce")
    if v.isna().any():
        raise ValueError(f"{col} contains non-numeric values")
    if (~pd.Series(v).map(pd.api.types.is_number)).any():
        raise ValueError(f"{col} contains non-numeric values")
    if (v < 0.0).any():
        raise ValueError(f"{col} must be non-negative")


def _require_sum_to_one(series: pd.Series, *, col: str, tol: float) -> None:
    v = pd.to_numeric(series, errors="coerce")
    if v.isna().any():
        raise ValueError(f"{col} contains non-numeric values")
    s = float(v.sum())
    if abs(s - 1.0) > float(tol):
        raise ValueError(f"{col} must sum to 1 (got {s})")


def _require_rowwise_sum_to_one(
    a: pd.Series,
    b: pd.Series,
    *,
    col_a: str,
    col_b: str,
    tol: float,
) -> None:
    va = pd.to_numeric(a, errors="coerce")
    vb = pd.to_numeric(b, errors="coerce")
    if va.isna().any() or vb.isna().any():
        raise ValueError(f"{col_a} and {col_b} must be numeric")
    s = (va + vb).astype(float)
    if (abs(s - 1.0) > float(tol)).any():
        raise ValueError(
            f"{col_a} + {col_b} must equal 1 for every row (within {tol})"
        )


def _validate_predicted_game_outcomes(df: pd.DataFrame) -> None:
    _require_columns(
        df,
        [
            "game_id",
            "round",
            "round_order",
            "sort_order",
            "team1_key",
            "team1_school_name",
            "team2_key",
            "team2_school_name",
            "p_matchup",
            "p_team1_wins_given_matchup",
            "p_team2_wins_given_matchup",
        ],
        name="predicted_game_outcomes",
    )
    _require_in_0_1(df["p_matchup"], col="p_matchup")
    _require_in_0_1(
        df["p_team1_wins_given_matchup"],
        col="p_team1_wins_given_matchup",
    )
    _require_in_0_1(
        df["p_team2_wins_given_matchup"],
        col="p_team2_wins_given_matchup",
    )
    _require_rowwise_sum_to_one(
        df["p_team1_wins_given_matchup"],
        df["p_team2_wins_given_matchup"],
        col_a="p_team1_wins_given_matchup",
        col_b="p_team2_wins_given_matchup",
        tol=1e-8,
    )


def _validate_predicted_auction_share_of_pool(df: pd.DataFrame) -> None:
    _require_columns(
        df,
        ["predicted_auction_share_of_pool"],
        name="predicted_auction_share_of_pool",
    )
    col = "predicted_auction_share_of_pool"
    v = pd.to_numeric(df[col], errors="coerce")
    if v.isna().any():
        raise ValueError(f"{col} contains non-numeric values")
    if (v < 0.0).any():
        raise ValueError(f"{col} must be non-negative")
    _require_sum_to_one(v, col=col, tol=1e-8)


def _validate_recommended_entry_bids(df: pd.DataFrame) -> None:
    _require_columns(
        df,
        ["team_key", "bid_amount_points"],
        name="recommended_entry_bids",
    )
    if df.empty:
        raise ValueError("recommended_entry_bids must not be empty")

    team_key = df["team_key"].astype(str)
    if (team_key.str.len() == 0).any():
        raise ValueError("team_key must be non-empty")
    if team_key.duplicated().any():
        raise ValueError("team_key must be unique")

    bids = pd.to_numeric(df["bid_amount_points"], errors="coerce")
    if bids.isna().any():
        raise ValueError("bid_amount_points must be numeric")
    if (bids < 0).any():
        raise ValueError("bid_amount_points must be non-negative")
    if (bids.round() != bids).any():
        raise ValueError("bid_amount_points must be integer-valued")


def _validate_simulated_tournaments(df: pd.DataFrame) -> None:
    _require_columns(
        df,
        [
            "sim_id",
            "team_key",
            "wins",
        ],
        name="simulated_tournaments",
    )
    if df.empty:
        raise ValueError("simulated_tournaments must not be empty")

    sim_ids = pd.to_numeric(df["sim_id"], errors="coerce")
    if sim_ids.isna().any() or (sim_ids < 0).any():
        raise ValueError("sim_id must be non-negative integer")

    wins = pd.to_numeric(df["wins"], errors="coerce")
    if wins.isna().any() or (wins < 0).any():
        raise ValueError("wins must be non-negative integer")


def _validate_simulated_entry_outcomes(df: pd.DataFrame) -> None:
    _require_columns(
        df,
        [
            "entry_key",
            "sims",
            "seed",
            "budget_points",
            "mean_payout_cents",
            "mean_total_points",
            "mean_finish_position",
            "mean_n_entries",
            "p_top1",
            "p_top3",
            "p_top6",
            "p_top10",
        ],
        name="simulated_entry_outcomes",
    )
    if df.empty:
        raise ValueError("simulated_entry_outcomes must not be empty")

    sims = pd.to_numeric(df["sims"], errors="coerce")
    if sims.isna().any() or (sims <= 0).any():
        raise ValueError("sims must be positive")

    budget = pd.to_numeric(df["budget_points"], errors="coerce")
    if budget.isna().any() or (budget <= 0).any():
        raise ValueError("budget_points must be positive")

    for c in ["p_top1", "p_top3", "p_top6", "p_top10"]:
        _require_in_0_1(df[c], col=c)

    payout = pd.to_numeric(df["mean_payout_cents"], errors="coerce")
    if payout.isna().any() or (payout < 0).any():
        raise ValueError("mean_payout_cents must be non-negative")


def _validate_investment_report(df: pd.DataFrame) -> None:
    if (df["portfolio_team_count"] <= 0).any():
        raise ValueError("portfolio_team_count must be positive")
    hhi_col = "portfolio_concentration_hhi"
    if ((df[hhi_col] < 0) | (df[hhi_col] > 1)).any():
        raise ValueError(f"{hhi_col} must be in [0, 1]")
    prob_cols = ["p_top1", "p_top3", "p_top6", "p_top10"]
    for col in prob_cols:
        if ((df[col] < 0) | (df[col] > 1)).any():
            raise ValueError(f"{col} must be in [0, 1]")


CONTRACTS: Dict[str, ArtifactContract] = {
    "predicted_game_outcomes": ArtifactContract(
        name="predicted_game_outcomes",
        required_columns=[
            "game_id",
            "round",
            "round_order",
            "sort_order",
            "team1_key",
            "team1_school_name",
            "team2_key",
            "team2_school_name",
            "p_matchup",
            "p_team1_wins_given_matchup",
            "p_team2_wins_given_matchup",
        ],
        validators=[_validate_predicted_game_outcomes],
    ),
    "predicted_auction_share_of_pool": ArtifactContract(
        name="predicted_auction_share_of_pool",
        required_columns=["predicted_auction_share_of_pool"],
        validators=[_validate_predicted_auction_share_of_pool],
    ),
    "recommended_entry_bids": ArtifactContract(
        name="recommended_entry_bids",
        required_columns=["team_key", "bid_amount_points"],
        validators=[_validate_recommended_entry_bids],
    ),
    "simulated_tournaments": ArtifactContract(
        name="simulated_tournaments",
        required_columns=["sim_id", "team_key", "wins"],
        validators=[_validate_simulated_tournaments],
    ),
    "simulated_entry_outcomes": ArtifactContract(
        name="simulated_entry_outcomes",
        required_columns=[
            "entry_key",
            "sims",
            "seed",
            "budget_points",
            "mean_payout_cents",
            "mean_total_points",
            "mean_finish_position",
            "p_top1",
            "p_top3",
            "p_top6",
            "p_top10",
        ],
        validators=[_validate_simulated_entry_outcomes],
    ),
    "investment_report": ArtifactContract(
        name="investment_report",
        required_columns=[
            "snapshot_name",
            "budget_points",
            "n_sims",
            "seed",
            "portfolio_team_count",
            "portfolio_total_bids",
            "mean_expected_payout_cents",
            "mean_expected_points",
            "mean_expected_finish_position",
            "mean_n_entries",
            "p_top1",
            "p_top3",
            "p_top6",
            "p_top10",
            "portfolio_concentration_hhi",
            "portfolio_teams_json",
        ],
        validators=[_validate_investment_report],
    ),
}


def validate_artifact_df(
    *,
    artifact_name: str,
    df: pd.DataFrame,
) -> None:
    c = CONTRACTS.get(str(artifact_name))
    if c is None:
        raise ValueError(f"unknown artifact contract: {artifact_name}")
    _require_columns(df, c.required_columns, name=c.name)
    for v in c.validators:
        v(df)
