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
    n_sims: int,
    seed: int,
    budget_points: int,
    calcutta_id: Optional[str] = None
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
            cur.execute("""
                INSERT INTO gold_optimization_runs
                (run_id, calcutta_id, strategy, n_sims, seed, budget_points)
                VALUES (%s, %s, %s, %s, %s, %s)
                ON CONFLICT (run_id) DO UPDATE SET
                    calcutta_id = EXCLUDED.calcutta_id,
                    strategy = EXCLUDED.strategy,
                    n_sims = EXCLUDED.n_sims,
                    seed = EXCLUDED.seed,
                    budget_points = EXCLUDED.budget_points
            """, (run_id, calcutta_id, strategy, n_sims, seed, budget_points))
            
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
            # Clear existing bids for this run
            cur.execute("""
                DELETE FROM gold_recommended_entry_bids
                WHERE run_id = %s
            """, (run_id,))
            
            # Extract school slugs from team keys and map to IDs
            df = bids_df.copy()
            df['school_slug'] = df['team_key'].str.split(':').str[-1]
            df['team_id'] = df['school_slug'].map(team_id_map)
            
            # Check for unmapped teams
            if df['team_id'].isna().any():
                unmapped = df[df['team_id'].isna()]['school_slug'].unique()
                raise ValueError(f"Unmapped teams: {list(unmapped)}")
            
            values = [
                (
                    run_id,
                    int(row['team_id']),
                    int(row['bid_amount_points']),
                    float(row.get('score', row.get('expected_roi', 0.0)))
                )
                for _, row in df.iterrows()
            ]
            
            psycopg2.extras.execute_batch(cur, """
                INSERT INTO gold_recommended_entry_bids
                (run_id, team_id, recommended_bid_points, expected_roi)
                VALUES (%s, %s, %s, %s)
            """, values)
            
            conn.commit()
            logger.info(f"Inserted {len(values)} recommended bids")
            return len(values)
