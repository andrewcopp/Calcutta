"""
Lab query helpers for comparing investment model performance.

Provides pandas-based queries for analysis and comparison.
"""

import logging
from typing import List, Optional

import pandas as pd

from moneyball.db.connection import get_db_connection

logger = logging.getLogger(__name__)


def get_model_leaderboard() -> pd.DataFrame:
    """
    Get the investment model leaderboard sorted by mean normalized payout.

    Returns:
        DataFrame with columns:
            - model_name: Investment model name
            - model_kind: Model type (ridge, etc.)
            - n_entries: Number of entries generated
            - n_evaluations: Number of evaluations completed
            - avg_mean_payout: Average mean normalized payout
            - avg_median_payout: Average median normalized payout
            - avg_p_top1: Average probability of finishing 1st
            - avg_p_in_money: Average probability of finishing in money
            - first_eval_at: When first evaluation was run
            - last_eval_at: When most recent evaluation was run
    """
    with get_db_connection() as conn:
        df = pd.read_sql(
            """
            SELECT
                investment_model_id,
                model_name,
                model_kind,
                n_entries,
                n_evaluations,
                avg_mean_payout,
                avg_median_payout,
                avg_p_top1,
                avg_p_in_money,
                first_eval_at,
                last_eval_at
            FROM lab.model_leaderboard
            """,
            conn,
        )
    return df


def get_entry_evaluations(
    model_name: Optional[str] = None,
    calcutta_id: Optional[str] = None,
) -> pd.DataFrame:
    """
    Get detailed evaluation results, optionally filtered.

    Args:
        model_name: Filter to specific investment model
        calcutta_id: Filter to specific calcutta

    Returns:
        DataFrame with evaluation details per entry
    """
    query = """
        SELECT
            entry_id,
            model_name,
            model_kind,
            calcutta_name,
            starting_state_key,
            game_outcome_kind,
            optimizer_kind,
            n_sims,
            seed,
            mean_normalized_payout,
            median_normalized_payout,
            p_top1,
            p_in_money,
            our_rank,
            eval_created_at
        FROM lab.entry_evaluations
        WHERE 1=1
    """
    params = []

    if model_name:
        query += " AND model_name = %s"
        params.append(model_name)

    if calcutta_id:
        query += " AND entry_id IN (SELECT id FROM lab.entries WHERE calcutta_id = %s)"
        params.append(calcutta_id)

    query += " ORDER BY eval_created_at DESC NULLS LAST"

    with get_db_connection() as conn:
        df = pd.read_sql(query, conn, params=params if params else None)
    return df


def compare_models(
    model_names: List[str],
    calcutta_ids: Optional[List[str]] = None,
) -> pd.DataFrame:
    """
    Compare multiple investment models across calcuttas.

    Args:
        model_names: List of model names to compare
        calcutta_ids: Optional list of calcutta IDs to filter to

    Returns:
        DataFrame with comparison metrics pivoted by model
    """
    placeholders = ",".join(["%s"] * len(model_names))
    query = f"""
        SELECT
            c.id AS calcutta_id,
            c.name AS calcutta_name,
            im.name AS model_name,
            e.starting_state_key,
            ev.n_sims,
            ev.seed,
            ev.mean_normalized_payout,
            ev.median_normalized_payout,
            ev.p_top1,
            ev.p_in_money
        FROM lab.evaluations ev
        JOIN lab.entries e ON e.id = ev.entry_id AND e.deleted_at IS NULL
        JOIN lab.investment_models im ON im.id = e.investment_model_id AND im.deleted_at IS NULL
        JOIN core.calcuttas c ON c.id = e.calcutta_id
        WHERE im.name IN ({placeholders})
          AND ev.deleted_at IS NULL
    """
    params = list(model_names)

    if calcutta_ids:
        calcutta_placeholders = ",".join(["%s"] * len(calcutta_ids))
        query += f" AND c.id IN ({calcutta_placeholders})"
        params.extend(calcutta_ids)

    query += " ORDER BY c.name, im.name"

    with get_db_connection() as conn:
        df = pd.read_sql(query, conn, params=params)
    return df


def get_best_model_per_calcutta() -> pd.DataFrame:
    """
    Get the best-performing investment model for each calcutta.

    Returns:
        DataFrame with the top model per calcutta by mean normalized payout
    """
    query = """
        WITH ranked AS (
            SELECT
                c.id AS calcutta_id,
                c.name AS calcutta_name,
                im.name AS model_name,
                im.kind AS model_kind,
                ev.mean_normalized_payout,
                ev.p_top1,
                ROW_NUMBER() OVER (
                    PARTITION BY c.id
                    ORDER BY ev.mean_normalized_payout DESC NULLS LAST
                ) AS rank
            FROM lab.evaluations ev
            JOIN lab.entries e ON e.id = ev.entry_id AND e.deleted_at IS NULL
            JOIN lab.investment_models im ON im.id = e.investment_model_id AND im.deleted_at IS NULL
            JOIN core.calcuttas c ON c.id = e.calcutta_id
            WHERE ev.deleted_at IS NULL
              AND ev.mean_normalized_payout IS NOT NULL
        )
        SELECT
            calcutta_id,
            calcutta_name,
            model_name,
            model_kind,
            mean_normalized_payout,
            p_top1
        FROM ranked
        WHERE rank = 1
        ORDER BY calcutta_name
    """
    with get_db_connection() as conn:
        df = pd.read_sql(query, conn)
    return df


def get_model_variance(model_name: str) -> pd.DataFrame:
    """
    Analyze variance in a model's performance across calcuttas.

    Args:
        model_name: Investment model name

    Returns:
        DataFrame with per-calcutta performance and overall stats
    """
    query = """
        SELECT
            c.name AS calcutta_name,
            ev.mean_normalized_payout,
            ev.median_normalized_payout,
            ev.p_top1,
            ev.p_in_money,
            ev.n_sims,
            ev.seed
        FROM lab.evaluations ev
        JOIN lab.entries e ON e.id = ev.entry_id AND e.deleted_at IS NULL
        JOIN lab.investment_models im ON im.id = e.investment_model_id AND im.deleted_at IS NULL
        JOIN core.calcuttas c ON c.id = e.calcutta_id
        WHERE im.name = %s
          AND ev.deleted_at IS NULL
          AND ev.mean_normalized_payout IS NOT NULL
        ORDER BY ev.mean_normalized_payout DESC
    """
    with get_db_connection() as conn:
        df = pd.read_sql(query, conn, params=[model_name])

    if not df.empty:
        stats = df["mean_normalized_payout"].describe()
        logger.info(
            f"Model '{model_name}' stats: "
            f"mean={stats['mean']:.4f}, "
            f"std={stats['std']:.4f}, "
            f"min={stats['min']:.4f}, "
            f"max={stats['max']:.4f}"
        )

    return df
