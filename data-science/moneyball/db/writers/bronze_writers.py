"""
Bronze layer database writers.

Write raw tournament and simulation data using integer IDs.
"""
import logging
import pandas as pd
import psycopg2.extras
from typing import Dict
from moneyball.db.connection import get_db_connection

logger = logging.getLogger(__name__)


def get_or_create_tournament(season: int) -> int:
    """
    Get or create tournament by season, return tournament_id.
    
    Args:
        season: Tournament year (e.g., 2025)
    
    Returns:
        tournament_id
    """
    tournament_name = f"NCAA Tournament {season}"
    
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT id FROM bronze_tournaments
                WHERE season = %s AND tournament_name = %s
            """, (season, tournament_name))
            
            result = cur.fetchone()
            if result:
                return result[0]
            
            cur.execute("""
                INSERT INTO bronze_tournaments (season, tournament_name)
                VALUES (%s, %s)
                RETURNING id
            """, (season, tournament_name))
            
            tournament_id = cur.fetchone()[0]
            conn.commit()
            return tournament_id


def write_teams(tournament_id: int, teams_df: pd.DataFrame) -> Dict[str, int]:
    """
    Write teams for a tournament, return school_slug -> team_id mapping.
    
    Args:
        tournament_id: Tournament ID
        teams_df: DataFrame with columns:
            - school_slug, school_name, seed, region
            - byes (optional), kenpom_* (optional)
    
    Returns:
        Dict mapping school_slug to team_id
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            team_ids = {}
            
            for _, row in teams_df.iterrows():
                cur.execute("""
                    INSERT INTO bronze_teams
                    (tournament_id, school_slug, school_name, seed, region,
                     byes, kenpom_net, kenpom_o, kenpom_d, kenpom_adj_t)
                    VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                    ON CONFLICT (tournament_id, school_slug)
                    DO UPDATE SET
                        school_name = EXCLUDED.school_name,
                        seed = EXCLUDED.seed,
                        region = EXCLUDED.region,
                        byes = EXCLUDED.byes,
                        kenpom_net = EXCLUDED.kenpom_net,
                        kenpom_o = EXCLUDED.kenpom_o,
                        kenpom_d = EXCLUDED.kenpom_d,
                        kenpom_adj_t = EXCLUDED.kenpom_adj_t
                    RETURNING id
                """, (
                    tournament_id,
                    row['school_slug'],
                    row['school_name'],
                    int(row['seed']),
                    row['region'],
                    int(row.get('byes', 0)),
                    float(row['kenpom_net']) if pd.notna(row.get('kenpom_net')) else None,
                    float(row['kenpom_o']) if pd.notna(row.get('kenpom_o')) else None,
                    float(row['kenpom_d']) if pd.notna(row.get('kenpom_d')) else None,
                    float(row['kenpom_adj_t']) if pd.notna(row.get('kenpom_adj_t')) else None,
                ))
                
                team_id = cur.fetchone()[0]
                team_ids[row['school_slug']] = team_id
            
            conn.commit()
            return team_ids


def write_simulated_tournaments(
    tournament_id: int,
    simulations_df: pd.DataFrame,
    team_id_map: Dict[str, int]
) -> int:
    """
    Write simulated tournament outcomes.
    
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
                DELETE FROM bronze_simulated_tournaments
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
                    int(row['team_id']),
                    int(row['wins']),
                    int(row['byes']),
                    bool(row['eliminated'])
                )
                for _, row in df.iterrows()
            ]
            
            # Batch insert
            psycopg2.extras.execute_batch(cur, """
                INSERT INTO bronze_simulated_tournaments
                (tournament_id, sim_id, team_id, wins, byes, eliminated)
                VALUES (%s, %s, %s, %s, %s, %s)
            """, values, page_size=10000)
            
            conn.commit()
            return len(values)
