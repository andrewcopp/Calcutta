"""
Gold layer writers - Business metrics and analysis.

These functions are called by Python optimizer and analytics services
to write results directly to Postgres.
"""
import logging
import pandas as pd
import psycopg2.extras
from datetime import datetime
from moneyball.db.connection import get_db_connection

logger = logging.getLogger(__name__)


def write_optimization_run(
    run_id: str,
    calcutta_key: str,
    strategy: str,
    n_sims: int,
    seed: int,
    budget_points: int = 100,
    run_timestamp: datetime = None
) -> None:
    """
    Write optimization run metadata to gold_optimization_runs.

    Args:
        run_id: Unique run identifier (e.g., timestamp)
        calcutta_key: Calcutta identifier
        strategy: Strategy name (e.g., "minlp", "greedy")
        n_sims: Number of simulations
        seed: Random seed
        budget_points: Budget in points (default: 100)
        run_timestamp: Timestamp of run (default: now)
    """
    if run_timestamp is None:
        run_timestamp = datetime.now()

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("""
                INSERT INTO gold_optimization_runs
                (run_id, calcutta_key, strategy, n_sims, seed,
                 budget_points, run_timestamp)
                VALUES (%s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (run_id) DO NOTHING
            """, (run_id, calcutta_key, strategy, n_sims, seed,
                  budget_points, run_timestamp))

            conn.commit()
            logger.info(f"Inserted optimization run: {run_id}")


def write_recommended_entry_bids(
    run_id: str,
    bids_df: pd.DataFrame
) -> int:
    """
    Write recommended entry bids to gold_recommended_entry_bids.

    Args:
        run_id: Optimization run identifier
        bids_df: DataFrame with columns:
            - team_key (str)
            - bid_amount_points (int)

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing bids for this run
            cur.execute("""
                DELETE FROM gold_recommended_entry_bids WHERE run_id = %s
            """, (run_id,))

            values = [
                (run_id, row['team_key'], int(row['bid_amount_points']))
                for _, row in bids_df.iterrows()
            ]

            psycopg2.extras.execute_batch(cur, """
                INSERT INTO gold_recommended_entry_bids
                (run_id, team_key, bid_amount_points)
                VALUES (%s, %s, %s)
            """, values)

            conn.commit()
            logger.info(f"Inserted {len(values)} recommended bids")
            return len(values)


def write_entry_simulation_outcomes(
    run_id: str,
    outcomes_df: pd.DataFrame,
    batch_size: int = 10000
) -> int:
    """
    Write entry simulation outcomes to gold_entry_simulation_outcomes.

    This is a high-volume table (240K rows per year for 48 entries x 5000 sims).
    Uses batch inserts for performance.

    Args:
        run_id: Optimization run identifier
        outcomes_df: DataFrame with columns:
            - entry_key (str)
            - sim_id (int)
            - payout_cents (int)
            - total_points (float)
            - finish_position (int)
            - is_tied (bool)
            - n_entries (int)
            - normalized_payout (float)
        batch_size: Number of rows per batch (default: 10000)

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing outcomes for this run
            cur.execute("""
                DELETE FROM gold_entry_simulation_outcomes
                WHERE run_id = %s
            """, (run_id,))

            # Prepare all values
            values = [
                (
                    run_id,
                    row['entry_key'],
                    int(row['sim_id']),
                    int(row['payout_cents']),
                    float(row['total_points']),
                    int(row['finish_position']),
                    bool(row['is_tied']),
                    int(row['n_entries']),
                    float(row['normalized_payout'])
                )
                for _, row in outcomes_df.iterrows()
            ]

            # Insert in batches
            total_inserted = 0
            for i in range(0, len(values), batch_size):
                batch = values[i:i+batch_size]
                psycopg2.extras.execute_batch(cur, """
                    INSERT INTO gold_entry_simulation_outcomes
                    (run_id, entry_key, sim_id, payout_cents, total_points,
                     finish_position, is_tied, n_entries, normalized_payout)
                    VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
                """, batch)
                total_inserted += len(batch)
                logger.info(
                    f"Inserted batch {i//batch_size + 1}: "
                    f"{total_inserted}/{len(values)} rows"
                )

            conn.commit()
            logger.info(
                f"Completed: Inserted {total_inserted} simulation outcomes"
            )
            return total_inserted


def write_entry_performance(
    run_id: str,
    performance_df: pd.DataFrame
) -> int:
    """
    Write entry performance metrics to gold_entry_performance.

    Args:
        run_id: Optimization run identifier
        performance_df: DataFrame with columns:
            - entry_key (str)
            - is_our_strategy (bool)
            - n_teams (int)
            - total_bid_points (int)
            - mean_payout_cents (float)
            - mean_points (float)
            - mean_normalized_payout (float)
            - p50_normalized_payout (float)
            - p90_normalized_payout (float)
            - p_top1 (float)
            - p_in_money (float)
            - percentile_rank (float, optional)

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing performance for this run
            cur.execute("""
                DELETE FROM gold_entry_performance WHERE run_id = %s
            """, (run_id,))

            values = [
                (
                    run_id,
                    row['entry_key'],
                    bool(row['is_our_strategy']),
                    int(row['n_teams']),
                    int(row['total_bid_points']),
                    float(row['mean_payout_cents']),
                    float(row['mean_points']),
                    float(row['mean_normalized_payout']),
                    float(row['p50_normalized_payout']),
                    float(row['p90_normalized_payout']),
                    float(row['p_top1']),
                    float(row['p_in_money']),
                    float(row['percentile_rank']) if pd.notna(
                        row.get('percentile_rank')) else None,
                )
                for _, row in performance_df.iterrows()
            ]

            psycopg2.extras.execute_batch(cur, """
                INSERT INTO gold_entry_performance
                (run_id, entry_key, is_our_strategy, n_teams,
                 total_bid_points, mean_payout_cents, mean_points,
                 mean_normalized_payout, p50_normalized_payout,
                 p90_normalized_payout, p_top1, p_in_money,
                 percentile_rank)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """, values)

            conn.commit()
            logger.info(f"Inserted {len(values)} entry performance records")
            return len(values)


def write_detailed_investment_report(
    run_id: str,
    report_df: pd.DataFrame
) -> int:
    """
    Write detailed investment report to gold_detailed_investment_report.

    Args:
        run_id: Optimization run identifier
        report_df: DataFrame with columns:
            - team_key (str)
            - our_bid_points (int)
            - expected_points (float)
            - predicted_market_points (float)
            - actual_market_points (float)
            - our_ownership (float)
            - expected_roi (float)
            - our_roi (float)
            - roi_degradation (float)

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing report for this run
            cur.execute("""
                DELETE FROM gold_detailed_investment_report
                WHERE run_id = %s
            """, (run_id,))

            values = [
                (
                    run_id,
                    row['team_key'],
                    int(row['our_bid_points']),
                    float(row['expected_points']),
                    float(row['predicted_market_points']),
                    float(row['actual_market_points']),
                    float(row['our_ownership']),
                    float(row['expected_roi']),
                    float(row['our_roi']),
                    float(row['roi_degradation'])
                )
                for _, row in report_df.iterrows()
            ]

            psycopg2.extras.execute_batch(cur, """
                INSERT INTO gold_detailed_investment_report
                (run_id, team_key, our_bid_points, expected_points,
                 predicted_market_points, actual_market_points,
                 our_ownership, expected_roi, our_roi, roi_degradation)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
            """, values)

            conn.commit()
            logger.info(f"Inserted {len(values)} investment report records")
            return len(values)
