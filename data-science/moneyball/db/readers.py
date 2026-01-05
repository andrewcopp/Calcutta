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

from moneyball.utils import points


def get_db_connection():
    """Get a database connection using environment variables."""
    password = os.environ.get("DB_PASSWORD", "").strip()
    if not password:
        raise RuntimeError(
            "DB_PASSWORD must be set (or use DATABASE_URL where supported)"
        )

    return psycopg2.connect(
        host=os.environ.get("DB_HOST", "localhost"),
        port=int(os.environ.get("DB_PORT", "5432")),
        dbname=os.environ.get("DB_NAME", "calcutta"),
        user=os.environ.get("DB_USER", "calcutta"),
        password=password,
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
                FROM lab_bronze.tournaments
                WHERE season = %s
                """,
                (year,)
            )
            row = cur.fetchone()
            return dict(row) if row else None
    finally:
        conn.close()


def read_points_by_win_index_for_year(year: int) -> Dict[int, float]:
    conn = get_db_connection()
    try:
        with conn.cursor(cursor_factory=RealDictCursor) as cur:
            cur.execute(
                """
                WITH season AS (
                    SELECT id
                    FROM core.seasons
                    WHERE year = %s AND deleted_at IS NULL
                ),
                tournament AS (
                    SELECT t.id
                    FROM core.tournaments t
                    JOIN season s ON s.id = t.season_id
                    WHERE t.deleted_at IS NULL
                    ORDER BY t.created_at DESC
                    LIMIT 1
                ),
                calcutta AS (
                    SELECT c.id
                    FROM core.calcuttas c
                    JOIN tournament t ON t.id = c.tournament_id
                    WHERE c.deleted_at IS NULL
                    ORDER BY c.created_at DESC
                    LIMIT 1
                )
                SELECT r.win_index, r.points_awarded
                FROM core.calcutta_scoring_rules r
                JOIN calcutta c ON c.id = r.calcutta_id
                WHERE r.deleted_at IS NULL
                ORDER BY r.win_index ASC
                """,
                (year,),
            )
            rows = cur.fetchall() or []

        df = pd.DataFrame(rows)
        return points.points_by_win_index_from_scoring_rules(df)
    finally:
        conn.close()


def read_points_by_win_index_for_calcutta(
    calcutta_id: str,
) -> Dict[int, float]:
    conn = get_db_connection()
    try:
        with conn.cursor(cursor_factory=RealDictCursor) as cur:
            cur.execute(
                """
                SELECT r.win_index, r.points_awarded
                FROM core.calcutta_scoring_rules r
                WHERE r.calcutta_id = %s AND r.deleted_at IS NULL
                ORDER BY r.win_index ASC
                """,
                (calcutta_id,),
            )
            rows = cur.fetchall() or []

        df = pd.DataFrame(rows)
        return points.points_by_win_index_from_scoring_rules(df)
    finally:
        conn.close()


def initialize_default_scoring_rules_for_year(year: int) -> Dict[int, float]:
    pbwi = read_points_by_win_index_for_year(year)
    points.set_default_points_by_win_index(pbwi)
    return pbwi


def initialize_default_scoring_rules_for_calcutta(
    calcutta_id: str,
) -> Dict[int, float]:
    pbwi = read_points_by_win_index_for_calcutta(calcutta_id)
    points.set_default_points_by_win_index(pbwi)
    return pbwi


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
        FROM lab_bronze.teams t
        JOIN lab_bronze.tournaments tour ON t.tournament_id = tour.id
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
        region_teams = teams_df[
            teams_df["region"] == region
        ].sort_values("seed")
        
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


def read_predicted_game_outcomes(
    year: int,
    model_version: str = "kenpom-v1",
) -> pd.DataFrame:
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
        FROM lab_silver.predicted_game_outcomes pgo
        JOIN lab_bronze.tournaments tour ON pgo.tournament_id = tour.id
        JOIN lab_bronze.teams t1 ON pgo.team1_id = t1.id
        JOIN lab_bronze.teams t2 ON pgo.team2_id = t2.id
        WHERE tour.season = %s AND pgo.model_version = %s
        ORDER BY pgo.round, t1.seed
        """
        return pd.read_sql_query(query, conn, params=(year, model_version))
    finally:
        conn.close()


def read_simulated_tournaments(
    year: int,
    run_id: Optional[str] = None,
    calcutta_id: Optional[str] = None,
    tournament_simulation_batch_id: Optional[str] = None,
) -> pd.DataFrame:
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
        # Note: Schema doesn't have points column; we calculate it from
        # wins+byes.
        # Also doesn't have run_id yet, so we ignore it for now.

        batch_id = tournament_simulation_batch_id
        if not batch_id:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT b.id
                    FROM lab_bronze.tournaments tour
                    JOIN analytics.tournament_simulation_batches b
                      ON b.tournament_id = tour.core_tournament_id
                     AND b.deleted_at IS NULL
                    WHERE tour.season = %s
                    ORDER BY b.created_at DESC
                    LIMIT 1
                    """,
                    (year,),
                )
                row = cur.fetchone()
                if row and row[0]:
                    batch_id = str(row[0])

        if batch_id:
            query = """
            SELECT
                st.id,
                st.tournament_id,
                st.team_id,
                st.sim_id,
                st.wins,
                st.byes,
                st.eliminated,
                st.created_at,
                t.school_name,
                t.seed,
                t.region
            FROM analytics.simulated_tournaments st
            JOIN lab_bronze.tournaments tour ON st.tournament_id = tour.id
            JOIN lab_bronze.teams t ON st.team_id = t.id
            WHERE tour.season = %s
              AND st.tournament_simulation_batch_id = %s
              AND st.deleted_at IS NULL
            ORDER BY st.sim_id, t.seed
            """
            df = pd.read_sql_query(query, conn, params=(year, batch_id))
        else:
            # Legacy fallback: use rows without batch id.
            query = """
            SELECT
                st.id,
                st.tournament_id,
                st.team_id,
                st.sim_id,
                st.wins,
                st.byes,
                st.eliminated,
                st.created_at,
                t.school_name,
                t.seed,
                t.region
            FROM analytics.simulated_tournaments st
            JOIN lab_bronze.tournaments tour ON st.tournament_id = tour.id
            JOIN lab_bronze.teams t ON st.team_id = t.id
            WHERE tour.season = %s
              AND st.tournament_simulation_batch_id IS NULL
              AND st.deleted_at IS NULL
            ORDER BY st.sim_id, t.seed
            """
            df = pd.read_sql_query(query, conn, params=(year,))

        pbwi = (
            read_points_by_win_index_for_calcutta(calcutta_id)
            if calcutta_id
            else read_points_by_win_index_for_year(year)
        )
        df["points"] = (df["wins"] + df["byes"]).apply(
            lambda p: points.team_points_from_scoring_rules(
                int(p),
                pbwi,
            )
        )
        return df
    finally:
        conn.close()


def read_recommended_entry_bids(year: int, run_id: str) -> pd.DataFrame:
    """
    Read recommended entry bids from the database.
    
    Args:
        year: Tournament year
        run_id: Run ID for the strategy generation run
        
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
            reb.expected_roi,
            reb.bid_points,
            reb.created_at,
            t.school_name,
            t.seed,
            t.region,
            seas.year AS season
        FROM lab_gold.recommended_entry_bids reb
        JOIN lab_gold.strategy_generation_runs sgr
          ON sgr.id = reb.strategy_generation_run_id
         AND sgr.deleted_at IS NULL
        JOIN core.calcuttas c
          ON c.id = sgr.calcutta_id
         AND c.deleted_at IS NULL
        JOIN core.tournaments tour
          ON tour.id = c.tournament_id
         AND tour.deleted_at IS NULL
        JOIN core.seasons seas
          ON seas.id = tour.season_id
        JOIN lab_bronze.teams t ON reb.team_id = t.id
        WHERE seas.year = %s
          AND sgr.run_key = %s
        ORDER BY reb.bid_points DESC
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
                SELECT
                    c.id,
                    c.tournament_id,
                    c.name,
                    c.created_at,
                    c.updated_at
                FROM lab_bronze.calcuttas c
                JOIN lab_bronze.tournaments t ON c.tournament_id = t.id
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
                FROM lab_gold.optimization_runs run
                JOIN lab_bronze.calcuttas bc ON run.calcutta_id = bc.id
                JOIN lab_bronze.tournaments tour ON bc.tournament_id = tour.id
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
