"""
Gold layer database writers.

Write optimization results and recommendations using UUIDs.
"""
import logging
import pandas as pd
import psycopg2.extras
from typing import Dict, Optional
from moneyball.db.connection import get_db_connection

logger = logging.getLogger(__name__)


def write_optimization_run(
    run_id: str,
    strategy: str,
    n_sims: int = 0,
    seed: int = 0,
    budget_points: int = 100,
    calcutta_id: Optional[str] = None,
    simulated_tournament_id: Optional[str] = None,
) -> None:
    """
    Write optimization run metadata.

    Args:
        calcutta_id: Calcutta ID
        run_id: Unique run identifier
        strategy: Strategy name
        n_sims: Number of simulations
        seed: Random seed
        budget_points: Budget in points
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO derived.strategy_generation_runs (
                    run_key,
                    name,
                    simulated_tournament_id,
                    calcutta_id,
                    purpose,
                    returns_model_key,
                    investment_model_key,
                    optimizer_key,
                    params_json,
                    git_sha
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, '{}'::jsonb, NULL)
                ON CONFLICT (run_key) DO UPDATE SET
                    simulated_tournament_id =
                        EXCLUDED.simulated_tournament_id,
                    calcutta_id = EXCLUDED.calcutta_id,
                    optimizer_key = EXCLUDED.optimizer_key,
                    name = EXCLUDED.name,
                    updated_at = NOW()
                """,
                (
                    run_id,
                    strategy,
                    simulated_tournament_id,
                    calcutta_id,
                    'moneyball_pipeline',
                    'legacy',
                    'legacy',
                    strategy,
                ),
            )
            conn.commit()
            logger.info(f"Wrote optimization run: {run_id}")


def write_recommended_entry_bids(
    run_id: str,
    bids_df: pd.DataFrame,
    team_id_map: Dict[str, str]
) -> int:
    """
    Write recommended entry bids.

    Args:
        run_id: Optimization run ID
        bids_df: DataFrame with columns:
            - team_key, bid_amount_points, score (expected_roi)
        team_id_map: Dict mapping school_slug to team_id

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT id
                FROM derived.strategy_generation_runs
                WHERE run_key = %s
                  AND deleted_at IS NULL
                ORDER BY created_at DESC
                LIMIT 1
                """,
                (run_id,),
            )
            row = cur.fetchone()
            if not row or not row[0]:
                raise ValueError(
                    f"No strategy_generation_run found for run_key={run_id}"
                )
            strategy_generation_run_id = str(row[0])

            # Clear existing bids for this run
            cur.execute("""
                DELETE FROM derived.strategy_generation_run_bids
                WHERE run_id = %s
            """, (run_id,))

            # Extract school slugs from team keys and map to IDs
            df = bids_df.copy()

            # Handle different column formats
            if 'team_key' in df.columns:
                df['school_slug'] = df['team_key'].str.split(':').str[-1]
                df['team_id'] = df['school_slug'].map(team_id_map)
            elif 'school_slug' in df.columns:
                df['team_id'] = df['school_slug'].map(team_id_map)
            elif 'team_id' not in df.columns:
                raise ValueError(
                    "DataFrame must have team_key, school_slug, or "
                    "team_id column"
                )

            # Check for unmapped teams
            if df['team_id'].isna().any():
                unmapped = df[df['team_id'].isna()]['school_slug'].unique()
                raise ValueError(f"Unmapped teams: {list(unmapped)}")

            values = [
                (
                    run_id,
                    strategy_generation_run_id,
                    str(row['team_id']),
                    int(float(row['bid_amount_points'])),
                    float(row.get('score', row.get('expected_roi', 0.0)))
                )
                for _, row in df.iterrows()
            ]

            psycopg2.extras.execute_batch(cur, """
                INSERT INTO derived.strategy_generation_run_bids
                (
                    run_id,
                    strategy_generation_run_id,
                    team_id,
                    bid_points,
                    expected_roi
                )
                VALUES (%s, %s, %s, %s, %s)
            """, values)

            conn.commit()
            logger.info(f"Inserted {len(values)} recommended bids")
            return len(values)
