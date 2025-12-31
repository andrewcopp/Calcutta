"""
Database readers for loading data from PostgreSQL.

This module provides functions to read data from the analytics database,
replacing parquet file dependencies with direct database queries.
"""
from __future__ import annotations

import os
from typing import Optional, Dict, Any
import pandas as pd
import psycopg2
from psycopg2.extras import RealDictCursor


def get_db_connection():
    """Get a database connection using environment variables."""
    return psycopg2.connect(
        host=os.environ.get("DB_HOST", "localhost"),
        port=int(os.environ.get("DB_PORT", "5432")),
        dbname=os.environ.get("DB_NAME", "calcutta"),
        user=os.environ.get("DB_USER", "calcutta"),
        password=os.environ.get("DB_PASSWORD", "calcutta"),
    )


def read_tournament(year: int) -> Optional[Dict[str, Any]]:
    """
    Read tournament metadata for a given year.
    
    Args:
        year: Tournament year (e.g., 2025)
        
    Returns:
        Dictionary with tournament metadata including id, season, created_at
        Returns None if tournament not found
    """
    conn = get_db_connection()
    try:
        with conn.cursor(cursor_factory=RealDictCursor) as cur:
            cur.execute(
                """
                SELECT id, season, created_at
                FROM bronze_tournaments
                WHERE season = %s
                """,
                (year,)
            )
            row = cur.fetchone()
            return dict(row) if row else None
    finally:
        conn.close()


def read_teams(year: int) -> pd.DataFrame:
    """
    Read teams for a given tournament year.
    
    Args:
        year: Tournament year
        
    Returns:
        DataFrame with columns: id, tournament_id, school_name, seed, region,
        kenpom_net, adj_o, adj_d, adj_t, created_at, updated_at
    """
    conn = get_db_connection()
    try:
        query = """
        SELECT
            t.id,
            t.tournament_id,
            t.school_slug,
            t.school_name,
            t.seed,
            t.region,
            t.kenpom_net,
            t.kenpom_adj_em,
            t.kenpom_adj_o,
            t.kenpom_adj_d,
            t.kenpom_adj_t,
            t.created_at
        FROM bronze_teams t
        JOIN bronze_tournaments tour ON t.tournament_id = tour.id
        WHERE tour.season = %s
        ORDER BY t.seed, t.school_name
        """
        return pd.read_sql_query(query, conn, params=(year,))
    finally:
        conn.close()


def read_games(year: int) -> pd.DataFrame:
    """
    Read games for a given tournament year.
    
    For now, this generates the bracket structure from teams.
    In the future, this could read from a bronze_games table if we add one.
    
    Args:
        year: Tournament year
        
    Returns:
        DataFrame with bracket matchups
    """
    teams_df = read_teams(year)
    
    # Generate bracket structure from teams
    # This is a simplified version - you may want to enhance this
    games = []
    regions = teams_df['region'].unique()
    
    for region in regions:
        region_teams = teams_df[teams_df['region'] == region].sort_values('seed')
        
        # Round 1 matchups (1v16, 8v9, 5v12, 4v13, 6v11, 3v14, 7v10, 2v15)
        matchups = [
            (1, 16), (8, 9), (5, 12), (4, 13),
            (6, 11), (3, 14), (7, 10), (2, 15)
        ]
        
        for seed1, seed2 in matchups:
            team1 = region_teams[region_teams['seed'] == seed1]
            team2 = region_teams[region_teams['seed'] == seed2]
            
            if len(team1) > 0 and len(team2) > 0:
                games.append({
                    'round': 1,
                    'region': region,
                    'team1_id': team1.iloc[0]['id'],
                    'team1_seed': seed1,
                    'team1_school': team1.iloc[0]['school_name'],
                    'team2_id': team2.iloc[0]['id'],
                    'team2_seed': seed2,
                    'team2_school': team2.iloc[0]['school_name'],
                })
    
    return pd.DataFrame(games)


def read_predicted_game_outcomes(year: int, model_version: str = "kenpom-v1") -> pd.DataFrame:
    """
    Read predicted game outcomes from the database.
    
    Args:
        year: Tournament year
        model_version: Model version to filter by
        
    Returns:
        DataFrame with game predictions
    """
    conn = get_db_connection()
    try:
        query = """
        SELECT 
            pgo.id,
            pgo.tournament_id,
            pgo.team1_id,
            pgo.team2_id,
            pgo.round,
            pgo.p_team1_wins,
            pgo.model_version,
            pgo.created_at,
            t1.school_name as team1_school,
            t1.seed as team1_seed,
            t2.school_name as team2_school,
            t2.seed as team2_seed
        FROM silver_predicted_game_outcomes pgo
        JOIN bronze_tournaments tour ON pgo.tournament_id = tour.id
        JOIN bronze_teams t1 ON pgo.team1_id = t1.id
        JOIN bronze_teams t2 ON pgo.team2_id = t2.id
        WHERE tour.season = %s AND pgo.model_version = %s
        ORDER BY pgo.round, t1.seed
        """
        return pd.read_sql_query(query, conn, params=(year, model_version))
    finally:
        conn.close()


def read_simulated_tournaments(year: int, run_id: Optional[str] = None) -> pd.DataFrame:
    """
    Read simulated tournament outcomes from the database.
    
    Args:
        year: Tournament year
        run_id: Optional run_id to filter by specific simulation run
        
    Returns:
        DataFrame with simulation results
    """
    conn = get_db_connection()
    try:
        if run_id:
            query = """
            SELECT 
                st.id,
                st.tournament_id,
                st.team_id,
                st.sim_id,
                st.wins,
                st.byes,
                st.points,
                st.run_id,
                st.created_at,
                t.school_name,
                t.seed,
                t.region
            FROM silver_simulated_tournaments st
            JOIN bronze_tournaments tour ON st.tournament_id = tour.id
            JOIN bronze_teams t ON st.team_id = t.id
            WHERE tour.season = %s AND st.run_id = %s
            ORDER BY st.sim_id, t.seed
            """
            params = (year, run_id)
        else:
            query = """
            SELECT 
                st.id,
                st.tournament_id,
                st.team_id,
                st.sim_id,
                st.wins,
                st.byes,
                st.points,
                st.run_id,
                st.created_at,
                t.school_name,
                t.seed,
                t.region
            FROM silver_simulated_tournaments st
            JOIN bronze_tournaments tour ON st.tournament_id = tour.id
            JOIN bronze_teams t ON st.team_id = t.id
            WHERE tour.season = %s
            ORDER BY st.sim_id, t.seed
            """
            params = (year,)
        
        return pd.read_sql_query(query, conn, params=params)
    finally:
        conn.close()


def read_recommended_entry_bids(year: int, run_id: str) -> pd.DataFrame:
    """
    Read recommended entry bids from the database.
    
    Args:
        year: Tournament year
        run_id: Run ID for the optimization run
        
    Returns:
        DataFrame with recommended bids
    """
    conn = get_db_connection()
    try:
        query = """
        SELECT 
            reb.id,
            reb.run_id,
            reb.team_id,
            reb.bid_amount_points,
            reb.expected_points,
            reb.expected_roi,
            reb.created_at,
            t.school_name,
            t.seed,
            t.region,
            tour.season
        FROM gold_recommended_entry_bids reb
        JOIN gold_optimization_runs run ON reb.run_id = run.run_id
        JOIN bronze_tournaments tour ON run.tournament_id = tour.id
        JOIN bronze_teams t ON reb.team_id = t.id
        WHERE tour.season = %s AND reb.run_id = %s
        ORDER BY reb.bid_amount_points DESC
        """
        return pd.read_sql_query(query, conn, params=(year, run_id))
    finally:
        conn.close()


def read_calcutta(year: int) -> Optional[Dict[str, Any]]:
    """
    Read calcutta metadata for a given year.
    
    Args:
        year: Tournament year
        
    Returns:
        Dictionary with calcutta metadata or None if not found
    """
    conn = get_db_connection()
    try:
        with conn.cursor(cursor_factory=RealDictCursor) as cur:
            cur.execute(
                """
                SELECT c.id, c.tournament_id, c.name, c.created_at, c.updated_at
                FROM bronze_calcuttas c
                JOIN bronze_tournaments t ON c.tournament_id = t.id
                WHERE t.season = %s
                LIMIT 1
                """,
                (year,)
            )
            row = cur.fetchone()
            return dict(row) if row else None
    finally:
        conn.close()


def tournament_exists(year: int) -> bool:
    """
    Check if a tournament exists for the given year.
    
    Args:
        year: Tournament year
        
    Returns:
        True if tournament exists, False otherwise
    """
    tournament = read_tournament(year)
    return tournament is not None


def get_latest_run_id(year: int) -> Optional[str]:
    """
    Get the most recent run_id for a given tournament year.
    
    Args:
        year: Tournament year
        
    Returns:
        Latest run_id or None if no runs exist
    """
    conn = get_db_connection()
    try:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT run.run_id
                FROM gold_optimization_runs run
                JOIN bronze_tournaments tour ON run.tournament_id = tour.id
                WHERE tour.season = %s
                ORDER BY run.created_at DESC
                LIMIT 1
                """,
                (year,)
            )
            row = cur.fetchone()
            return row[0] if row else None
    finally:
        conn.close()
