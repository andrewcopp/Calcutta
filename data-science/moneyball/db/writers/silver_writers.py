"""
Silver layer writers - ML predictions and enriched data.

These functions are called by Python ML services to write model outputs
directly to Postgres.
"""
import logging
import pandas as pd
import psycopg2.extras
from moneyball.db.connection import get_db_connection

logger = logging.getLogger(__name__)


def write_predicted_game_outcomes(
    tournament_key: str,
    predictions_df: pd.DataFrame,
    model_version: str = None
) -> int:
    """
    Write game outcome predictions to silver_predicted_game_outcomes.

    Args:
        tournament_key: Tournament identifier
        predictions_df: DataFrame with columns:
            - game_id (str)
            - round (int)
            - team1_key (str)
            - team2_key (str)
            - p_team1_wins (float)
            - p_matchup (float, optional, default: 1.0)
        model_version: Optional model version string

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing predictions for this tournament
            cur.execute("""
                DELETE FROM silver_predicted_game_outcomes
                WHERE tournament_key = %s
            """, (tournament_key,))

            values = [
                (
                    tournament_key,
                    row['game_id'],
                    int(row['round']),
                    row['team1_key'],
                    row['team2_key'],
                    float(row['p_team1_wins']),
                    float(row.get('p_matchup', 1.0)),
                    model_version
                )
                for _, row in predictions_df.iterrows()
            ]

            psycopg2.extras.execute_batch(cur, """
                INSERT INTO silver_predicted_game_outcomes
                (tournament_key, game_id, round, team1_key, team2_key,
                 p_team1_wins, p_matchup, model_version)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
            """, values)

            conn.commit()
            logger.info(f"Inserted {len(values)} game predictions")
            return len(values)


def write_predicted_market_share(
    calcutta_key: str,
    predictions_df: pd.DataFrame,
    model_version: str = None
) -> int:
    """
    Write market share predictions to silver_predicted_market_share.

    Args:
        calcutta_key: Calcutta identifier
        predictions_df: DataFrame with columns:
            - team_key (str)
            - predicted_share_of_pool (float)
        model_version: Optional model version string

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing predictions for this calcutta
            cur.execute("""
                DELETE FROM silver_predicted_market_share
                WHERE calcutta_key = %s
            """, (calcutta_key,))

            values = [
                (
                    calcutta_key,
                    row['team_key'],
                    float(row['predicted_share_of_pool']),
                    model_version
                )
                for _, row in predictions_df.iterrows()
            ]

            psycopg2.extras.execute_batch(cur, """
                INSERT INTO silver_predicted_market_share
                (calcutta_key, team_key, predicted_share_of_pool,
                 model_version)
                VALUES (%s, %s, %s, %s)
            """, values)

            conn.commit()
            logger.info(f"Inserted {len(values)} market predictions")
            return len(values)


def write_team_tournament_value(
    tournament_key: str,
    values_df: pd.DataFrame
) -> int:
    """
    Write team tournament values to silver_team_tournament_value.

    Args:
        tournament_key: Tournament identifier
        values_df: DataFrame with columns:
            - team_key (str)
            - expected_points (float)
            - p_champion (float, optional)
            - p_finals (float, optional)
            - p_final_four (float, optional)
            - p_elite_eight (float, optional)
            - p_sweet_sixteen (float, optional)
            - p_round_32 (float, optional)

    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing values for this tournament
            cur.execute("""
                DELETE FROM silver_team_tournament_value
                WHERE tournament_key = %s
            """, (tournament_key,))

            values = [
                (
                    tournament_key,
                    row['team_key'],
                    float(row['expected_points']),
                    float(row['p_champion']) if pd.notna(
                        row.get('p_champion')) else None,
                    float(row['p_finals']) if pd.notna(
                        row.get('p_finals')) else None,
                    float(row['p_final_four']) if pd.notna(
                        row.get('p_final_four')) else None,
                    float(row['p_elite_eight']) if pd.notna(
                        row.get('p_elite_eight')) else None,
                    float(row['p_sweet_sixteen']) if pd.notna(
                        row.get('p_sweet_sixteen')) else None,
                    float(row['p_round_32']) if pd.notna(
                        row.get('p_round_32')) else None,
                )
                for _, row in values_df.iterrows()
            ]

            psycopg2.extras.execute_batch(cur, """
                INSERT INTO silver_team_tournament_value
                (tournament_key, team_key, expected_points, p_champion,
                 p_finals, p_final_four, p_elite_eight, p_sweet_sixteen,
                 p_round_32)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
            """, values)

            conn.commit()
            logger.info(f"Inserted {len(values)} team values")
            return len(values)
