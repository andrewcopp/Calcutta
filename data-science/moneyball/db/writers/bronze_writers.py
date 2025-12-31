"""
Bronze layer writers - Raw simulation data.

These functions are called by Go services to write tournament simulations
and related data directly to Postgres.
"""
import logging
import pandas as pd
import psycopg2.extras
from typing import Optional
from moneyball.db.connection import get_db_connection

logger = logging.getLogger(__name__)


def write_tournaments(tournaments_df: pd.DataFrame) -> int:
    """
    Write tournament metadata to bronze_tournaments.

    Args:
        tournaments_df: DataFrame with columns:
            - tournament_key (str)
            - season (int)
            - tournament_name (str)

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            values = [
                (row['tournament_key'], int(row['season']),
                 row['tournament_name'])
                for _, row in tournaments_df.iterrows()
            ]

            psycopg2.extras.execute_batch(cur, """
                INSERT INTO bronze_tournaments
                (tournament_key, season, tournament_name)
                VALUES (%s, %s, %s)
                ON CONFLICT (tournament_key) DO NOTHING
            """, values)

            conn.commit()
            logger.info(f"Inserted {len(values)} tournaments")
            return len(values)


def write_teams(teams_df: pd.DataFrame) -> int:
    """
    Write team data to bronze_teams.

    Args:
        teams_df: DataFrame with columns:
            - team_key (str)
            - tournament_key (str)
            - school_slug (str)
            - school_name (str)
            - seed (int)
            - region (str)
            - byes (int, optional)
            - kenpom_net (float, optional)
            - kenpom_o (float, optional)
            - kenpom_d (float, optional)
            - kenpom_adj_t (float, optional)

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            values = [
                (
                    row['team_key'],
                    row['tournament_key'],
                    row['school_slug'],
                    row['school_name'],
                    int(row['seed']),
                    row['region'],
                    int(row.get('byes', 0)),
                    float(row['kenpom_net']) if pd.notna(
                        row.get('kenpom_net')) else None,
                    float(row['kenpom_o']) if pd.notna(
                        row.get('kenpom_o')) else None,
                    float(row['kenpom_d']) if pd.notna(
                        row.get('kenpom_d')) else None,
                    float(row['kenpom_adj_t']) if pd.notna(
                        row.get('kenpom_adj_t')) else None,
                )
                for _, row in teams_df.iterrows()
            ]

            psycopg2.extras.execute_batch(cur, """
                INSERT INTO bronze_teams
                (team_key, tournament_key, school_slug, school_name,
                 seed, region, byes, kenpom_net, kenpom_o, kenpom_d,
                 kenpom_adj_t)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (team_key) DO UPDATE SET
                    seed = EXCLUDED.seed,
                    region = EXCLUDED.region,
                    byes = EXCLUDED.byes,
                    kenpom_net = EXCLUDED.kenpom_net,
                    kenpom_o = EXCLUDED.kenpom_o,
                    kenpom_d = EXCLUDED.kenpom_d,
                    kenpom_adj_t = EXCLUDED.kenpom_adj_t
            """, values)

            conn.commit()
            logger.info(f"Inserted {len(values)} teams")
            return len(values)


def write_simulated_tournaments(
    tournament_key: str,
    simulations_df: pd.DataFrame,
    batch_size: int = 10000
) -> int:
    """
    Write simulated tournament outcomes to bronze_simulated_tournaments.

    This is a high-volume table (320K rows per year for 5000 sims).
    Uses batch inserts for performance.

    Args:
        tournament_key: Tournament identifier
        simulations_df: DataFrame with columns:
            - sim_id (int)
            - team_key (str)
            - wins (int)
            - byes (int)
            - eliminated (bool)
        batch_size: Number of rows per batch (default: 10000)

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing data for this tournament
            cur.execute("""
                DELETE FROM bronze_simulated_tournaments
                WHERE tournament_key = %s
            """, (tournament_key,))

            # Prepare all values
            values = [
                (
                    tournament_key,
                    int(row['sim_id']),
                    row['team_key'],
                    int(row['wins']),
                    int(row['byes']),
                    bool(row['eliminated'])
                )
                for _, row in simulations_df.iterrows()
            ]

            # Insert in batches
            total_inserted = 0
            for i in range(0, len(values), batch_size):
                batch = values[i:i+batch_size]
                psycopg2.extras.execute_batch(cur, """
                    INSERT INTO bronze_simulated_tournaments
                    (tournament_key, sim_id, team_key, wins, byes,
                     eliminated)
                    VALUES (%s, %s, %s, %s, %s, %s)
                """, batch)
                total_inserted += len(batch)
                logger.info(
                    f"Inserted batch {i//batch_size + 1}: "
                    f"{total_inserted}/{len(values)} rows"
                )

            conn.commit()
            logger.info(
                f"Completed: Inserted {total_inserted} simulation records"
            )
            return total_inserted


def write_calcuttas(calcuttas_df: pd.DataFrame) -> int:
    """
    Write calcutta metadata to bronze_calcuttas.

    Args:
        calcuttas_df: DataFrame with columns:
            - calcutta_key (str)
            - tournament_key (str)
            - calcutta_name (str)
            - budget_points (int, optional, default: 100)

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            values = [
                (
                    row['calcutta_key'],
                    row['tournament_key'],
                    row['calcutta_name'],
                    int(row.get('budget_points', 100))
                )
                for _, row in calcuttas_df.iterrows()
            ]

            psycopg2.extras.execute_batch(cur, """
                INSERT INTO bronze_calcuttas
                (calcutta_key, tournament_key, calcutta_name, budget_points)
                VALUES (%s, %s, %s, %s)
                ON CONFLICT (calcutta_key) DO NOTHING
            """, values)

            conn.commit()
            logger.info(f"Inserted {len(values)} calcuttas")
            return len(values)


def write_entry_bids(
    calcutta_key: str,
    entry_bids_df: pd.DataFrame
) -> int:
    """
    Write actual entry bids to bronze_entry_bids.

    Args:
        calcutta_key: Calcutta identifier
        entry_bids_df: DataFrame with columns:
            - entry_key (str)
            - team_key (str)
            - bid_amount (int)

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing data for this calcutta
            cur.execute("""
                DELETE FROM bronze_entry_bids WHERE calcutta_key = %s
            """, (calcutta_key,))

            values = [
                (
                    calcutta_key,
                    row['entry_key'],
                    row['team_key'],
                    int(row['bid_amount'])
                )
                for _, row in entry_bids_df.iterrows()
            ]

            psycopg2.extras.execute_batch(cur, """
                INSERT INTO bronze_entry_bids
                (calcutta_key, entry_key, team_key, bid_amount)
                VALUES (%s, %s, %s, %s)
            """, values)

            conn.commit()
            logger.info(f"Inserted {len(values)} entry bids")
            return len(values)


def write_payouts(calcutta_key: str, payouts_df: pd.DataFrame) -> int:
    """
    Write payout structure to bronze_payouts.

    Args:
        calcutta_key: Calcutta identifier
        payouts_df: DataFrame with columns:
            - position (int)
            - amount_cents (int)

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing data for this calcutta
            cur.execute("""
                DELETE FROM bronze_payouts WHERE calcutta_key = %s
            """, (calcutta_key,))

            values = [
                (
                    calcutta_key,
                    int(row['position']),
                    int(row['amount_cents'])
                )
                for _, row in payouts_df.iterrows()
            ]

            psycopg2.extras.execute_batch(cur, """
                INSERT INTO bronze_payouts
                (calcutta_key, position, amount_cents)
                VALUES (%s, %s, %s)
            """, values)

            conn.commit()
            logger.info(f"Inserted {len(values)} payout positions")
            return len(values)
