"""
Bronze layer database writers.

Write raw tournament and simulation data using UUIDs.
"""
import logging
import pandas as pd
from typing import Dict
from moneyball.db.connection import get_db_connection

logger = logging.getLogger(__name__)


def get_or_create_tournament(season: int) -> str:
    """
    Get or create tournament by season, return tournament UUID.
    
    Args:
        season: Tournament year (e.g., 2025)
    
    Returns:
        tournament_id (UUID as string)
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT id FROM bronze.tournaments
                WHERE season = %s
            """, (season,))
            
            result = cur.fetchone()
            if result:
                return str(result[0])
            
            cur.execute("""
                INSERT INTO bronze.tournaments (season)
                VALUES (%s)
                RETURNING id
            """, (season,))
            
            tournament_id = cur.fetchone()[0]
            conn.commit()
            return str(tournament_id)


def write_teams(tournament_id: str, teams_df: pd.DataFrame) -> Dict[str, str]:
    """
    Write teams for a tournament, return school_slug -> team_id mapping.
    
    Args:
        tournament_id: Tournament UUID
        teams_df: DataFrame with columns:
            - school_slug, school_name, seed, region
            - kenpom_* (optional)
    
    Returns:
        Dict mapping school_slug to team_id (UUID as string)
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            team_ids = {}
            
            for _, row in teams_df.iterrows():
                cur.execute("""
                    INSERT INTO bronze.teams
                    (tournament_id, school_slug, school_name, seed, region,
                     kenpom_net, kenpom_adj_em, kenpom_adj_o,
                     kenpom_adj_d, kenpom_adj_t)
                    VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                    ON CONFLICT (tournament_id, school_slug)
                    DO UPDATE SET
                        school_name = EXCLUDED.school_name,
                        seed = EXCLUDED.seed,
                        region = EXCLUDED.region,
                        kenpom_net = EXCLUDED.kenpom_net,
                        kenpom_adj_em = EXCLUDED.kenpom_adj_em,
                        kenpom_adj_o = EXCLUDED.kenpom_adj_o,
                        kenpom_adj_d = EXCLUDED.kenpom_adj_d,
                        kenpom_adj_t = EXCLUDED.kenpom_adj_t
                    RETURNING id
                """, (
                    tournament_id,
                    row['school_slug'],
                    row['school_name'],
                    int(row['seed']),
                    row['region'],
                    (
                        float(row['kenpom_net'])
                        if pd.notna(row.get('kenpom_net'))
                        else None
                    ),
                    (
                        float(row['kenpom_adj_em'])
                        if pd.notna(row.get('kenpom_adj_em'))
                        else None
                    ),
                    (
                        float(row['kenpom_adj_o'])
                        if pd.notna(row.get('kenpom_adj_o'))
                        else None
                    ),
                    (
                        float(row['kenpom_adj_d'])
                        if pd.notna(row.get('kenpom_adj_d'))
                        else None
                    ),
                    (
                        float(row['kenpom_adj_t'])
                        if pd.notna(row.get('kenpom_adj_t'))
                        else None
                    ),
                ))
                
                team_id = cur.fetchone()[0]
                team_ids[row['school_slug']] = str(team_id)
            
            conn.commit()
            return team_ids


