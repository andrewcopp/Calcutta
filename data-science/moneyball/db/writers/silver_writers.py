"""
Silver layer database writers.

Write ML predictions, simulations, and enriched data using UUIDs.
"""
import logging
import pandas as pd
import psycopg2.extras
from typing import Dict
from moneyball.db.connection import get_db_connection

logger = logging.getLogger(__name__)


def write_predicted_game_outcomes(
    tournament_id: str,
    predictions_df: pd.DataFrame,
    team_id_map: Dict[str, str],
    model_version: str = None
) -> int:
    """
    Write game outcome predictions.
    
    Args:
        tournament_id: Tournament ID
        predictions_df: DataFrame with columns:
            - game_id, round, team1_key, team2_key
            - p_team1_wins_given_matchup, p_matchup
        team_id_map: Dict mapping school_slug to team_id
        model_version: Optional model version
    
    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing predictions
            cur.execute("""
                DELETE FROM silver_predicted_game_outcomes
                WHERE tournament_id = %s
            """, (tournament_id,))
            
            # Extract school slugs from team keys and map to IDs
            df = predictions_df.copy()
            df['team1_slug'] = df['team1_key'].str.split(':').str[-1]
            df['team2_slug'] = df['team2_key'].str.split(':').str[-1]
            df['team1_id'] = df['team1_slug'].map(team_id_map)
            df['team2_id'] = df['team2_slug'].map(team_id_map)
            
            # Map round names to inverted integers (championship = 0)
            round_mapping = {
                'championship': 0,
                'final_four': 1,
                'elite_8': 2,
                'sweet_16': 3,
                'round_of_32': 4,
                'round_of_64': 5,
                'first_four': 6,
            }
            df['round_int'] = df['round'].map(round_mapping)
            
            # Check for unmapped teams
            if df['team1_id'].isna().any() or df['team2_id'].isna().any():
                unmapped = set()
                if df['team1_id'].isna().any():
                    unmapped.update(df[df['team1_id'].isna()]['team1_slug'])
                if df['team2_id'].isna().any():
                    unmapped.update(df[df['team2_id'].isna()]['team2_slug'])
                raise ValueError(f"Unmapped teams: {list(unmapped)}")
            
            values = [
                (
                    tournament_id,
                    row['game_id'],
                    int(row['round_int']),
                    str(row['team1_id']),  # team_id is UUID string
                    str(row['team2_id']),  # team_id is UUID string
                    float(row.get('p_team1_wins_given_matchup',
                          row.get('p_team1_wins', 0.5))),
                    float(row.get('p_matchup', 1.0)),
                    model_version
                )
                for _, row in df.iterrows()
            ]
            
            psycopg2.extras.execute_batch(cur, """
                INSERT INTO silver_predicted_game_outcomes
                (tournament_id, game_id, round, team1_id, team2_id,
                 p_team1_wins, p_matchup, model_version)
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT (tournament_id, game_id) DO UPDATE SET
                    round = EXCLUDED.round,
                    team1_id = EXCLUDED.team1_id,
                    team2_id = EXCLUDED.team2_id,
                    p_team1_wins = EXCLUDED.p_team1_wins,
                    p_matchup = EXCLUDED.p_matchup,
                    model_version = EXCLUDED.model_version
            """, values)
            
            conn.commit()
            return len(values)


def write_simulated_tournaments(
    tournament_id: str,
    simulations_df: pd.DataFrame,
    team_id_map: Dict[str, str]
) -> int:
    """
    Write simulated tournament outcomes to silver layer.
    
    Simulated tournaments are derived data (Monte Carlo simulations),
    not raw data, so they belong in the silver layer.
    
    Args:
        tournament_id: Tournament ID
        simulations_df: DataFrame with columns:
            - sim_id, school_slug, wins, byes, eliminated
        team_id_map: Dict mapping school_slug to team_id
    
    Returns:
        Number of rows inserted
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Clear existing simulations
            cur.execute("""
                DELETE FROM silver_simulated_tournaments
                WHERE tournament_id = %s
            """, (tournament_id,))
            
            # Map school_slug to team_id
            df = simulations_df.copy()
            df['team_id'] = df['school_slug'].map(team_id_map)
            
            # Check for unmapped teams
            if df['team_id'].isna().any():
                unmapped = df[df['team_id'].isna()]['school_slug'].unique()
                raise ValueError(f"Unmapped teams: {list(unmapped)}")
            
            # Prepare values
            values = [
                (
                    tournament_id,
                    int(row['sim_id']),
                    str(row['team_id']),  # team_id is UUID string
                    int(row['wins']),
                    int(row['byes']),
                    bool(row['eliminated'])
                )
                for _, row in df.iterrows()
            ]
            
            # Batch insert
            psycopg2.extras.execute_batch(cur, """
                INSERT INTO silver_simulated_tournaments
                (tournament_id, sim_id, team_id, wins, byes, eliminated)
                VALUES (%s, %s, %s, %s, %s, %s)
            """, values, page_size=10000)
            
            conn.commit()
            return len(values)
