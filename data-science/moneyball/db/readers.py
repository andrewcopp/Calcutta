"""
Database readers for loading data from PostgreSQL.

This module provides functions to read data from the analytics database,
replacing parquet file dependencies with direct database queries.
"""
from __future__ import annotations

import os
from typing import Optional, Dict, Any, List
import pandas as pd
import psycopg2
from psycopg2.extras import RealDictCursor

from moneyball.utils import points


def get_db_connection():
    """Get a database connection using environment variables."""
    database_url = os.environ.get("DATABASE_URL", "").strip()
    try:
        if database_url:
            return psycopg2.connect(database_url)

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
    except psycopg2.OperationalError as e:
        msg = str(e)
        if 'could not translate host name "db"' in msg:
            raise psycopg2.OperationalError(
                msg
                + "\n\n"
                + (
                    "It looks like DB_HOST is set to 'db' (docker-compose "
                    "hostname). When running locally outside docker, set "
                    "DB_HOST=localhost (and DB_PORT/DB_NAME/DB_USER as "
                    "needed) or set DATABASE_URL."
                )
            )
        raise


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
                SELECT t.id, s.year AS season, t.created_at
                FROM core.tournaments t
                JOIN core.seasons s
                  ON s.id = t.season_id
                 AND s.deleted_at IS NULL
                WHERE s.year = %s
                  AND t.deleted_at IS NULL
                ORDER BY t.created_at DESC
                LIMIT 1
                """,
                (year,)
            )
            row = cur.fetchone()
            return dict(row) if row else None
    finally:
        conn.close()


def _read_latest_core_tournament_id_for_year(conn, year: int) -> Optional[str]:
    with conn.cursor() as cur:
        cur.execute(
            """
            SELECT t.id
            FROM core.tournaments t
            JOIN core.seasons s
              ON s.id = t.season_id
             AND s.deleted_at IS NULL
            WHERE s.year = %s
              AND t.deleted_at IS NULL
            ORDER BY t.created_at DESC
            LIMIT 1
            """,
            (year,),
        )
        row = cur.fetchone()
        return str(row[0]) if row and row[0] else None


def _read_latest_core_calcutta_id_for_tournament(
    conn,
    tournament_id: str,
) -> Optional[str]:
    with conn.cursor() as cur:
        cur.execute(
            """
            SELECT c.id
            FROM core.calcuttas c
            WHERE c.tournament_id = %s
              AND c.deleted_at IS NULL
            ORDER BY c.created_at DESC
            LIMIT 1
            """,
            (tournament_id,),
        )
        row = cur.fetchone()
        return str(row[0]) if row and row[0] else None


def read_ridge_team_dataset_for_year(
    year: int,
    exclude_entry_names: Optional[List[str]] = None,
    include_target: bool = True,
) -> pd.DataFrame:
    conn = get_db_connection()
    try:
        y = int(year)
        tournament_id = _read_latest_core_tournament_id_for_year(conn, y)
        if not tournament_id:
            raise ValueError(f"no core tournament found for year {year}")

        tournament_key = f"ncaa-tournament-{y}"
        snapshot = str(y)

        if include_target:
            calcutta_id = _read_latest_core_calcutta_id_for_tournament(
                conn,
                tournament_id,
            )
            if not calcutta_id:
                raise ValueError(
                    f"no core calcutta found for tournament_id={tournament_id}"
                )

            exclude = [
                str(n)
                for n in (exclude_entry_names or [])
                if str(n).strip()
            ]
            exclude_clause = ""
            params: List[Any] = [calcutta_id]
            if exclude:
                exclude_clause = " AND ce.name <> ALL(%s::text[]) "
                params.append(exclude)

            query = f"""
            WITH team_bids AS (
                SELECT
                    cet.team_id,
                    SUM(cet.bid_points)::float8 AS total_bid
                FROM core.entry_teams cet
                JOIN core.entries ce
                  ON ce.id = cet.entry_id
                 AND ce.deleted_at IS NULL
                WHERE ce.calcutta_id = %s
                  AND cet.deleted_at IS NULL
                  {exclude_clause}
                GROUP BY cet.team_id
            ),
            total AS (
                SELECT COALESCE(SUM(total_bid), 0)::float8 AS total_bid
                FROM team_bids
            )
            SELECT
                %s::text AS snapshot,
                %s::text AS tournament_key,
                %s::text AS calcutta_key,
                (%s::text || ':' || s.slug)::text AS team_key,
                s.name::text AS school_name,
                s.slug::text AS school_slug,
                tt.seed::int AS seed,
                tt.region::text AS region,
                COALESCE(k.net_rtg, 0)::float8 AS kenpom_net,
                COALESCE(k.o_rtg, 0)::float8 AS kenpom_o,
                COALESCE(k.d_rtg, 0)::float8 AS kenpom_d,
                COALESCE(k.adj_t, 0)::float8 AS kenpom_adj_t,
                CASE
                    WHEN (SELECT total_bid FROM total) > 0 THEN
                        COALESCE(tb.total_bid, 0)::float8
                        / (SELECT total_bid FROM total)
                    ELSE NULL
                END AS team_share_of_pool
            FROM core.teams tt
            JOIN core.schools s
              ON s.id = tt.school_id
             AND s.deleted_at IS NULL
            LEFT JOIN core.team_kenpom_stats k
              ON k.team_id = tt.id
             AND k.deleted_at IS NULL
            LEFT JOIN team_bids tb ON tb.team_id = tt.id
            WHERE tt.tournament_id = %s
              AND tt.deleted_at IS NULL
            ORDER BY tt.seed ASC, s.name ASC;
            """

            params = [
                *params,
                snapshot,
                tournament_key,
                str(calcutta_id),
                tournament_key,
                str(tournament_id),
            ]
            return pd.read_sql_query(query, conn, params=tuple(params))

        query = """
        SELECT
            %s::text AS snapshot,
            %s::text AS tournament_key,
            NULL::text AS calcutta_key,
            (%s::text || ':' || s.slug)::text AS team_key,
            s.name::text AS school_name,
            s.slug::text AS school_slug,
            tt.seed::int AS seed,
            tt.region::text AS region,
            COALESCE(k.net_rtg, 0)::float8 AS kenpom_net,
            COALESCE(k.o_rtg, 0)::float8 AS kenpom_o,
            COALESCE(k.d_rtg, 0)::float8 AS kenpom_d,
            COALESCE(k.adj_t, 0)::float8 AS kenpom_adj_t
        FROM core.teams tt
        JOIN core.schools s
          ON s.id = tt.school_id
         AND s.deleted_at IS NULL
        LEFT JOIN core.team_kenpom_stats k
          ON k.team_id = tt.id
         AND k.deleted_at IS NULL
        WHERE tt.tournament_id = %s
          AND tt.deleted_at IS NULL
        ORDER BY tt.seed ASC, s.name ASC;
        """
        params = (
            snapshot,
            tournament_key,
            tournament_key,
            str(tournament_id),
        )
        return pd.read_sql_query(query, conn, params=params)
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
        DataFrame with columns: id, tournament_id, school_slug, school_name,
        seed, region, kenpom_net, kenpom_adj_em, kenpom_adj_o, kenpom_adj_d,
        kenpom_adj_t, created_at
    """
    conn = get_db_connection()
    try:
        query = """
        SELECT
            t.id,
            t.tournament_id,
            s.slug AS school_slug,
            s.name AS school_name,
            t.seed,
            t.region,
            COALESCE(k.net_rtg, 0)::float8 AS kenpom_net,
            COALESCE(k.net_rtg, 0)::float8 AS kenpom_adj_em,
            COALESCE(k.o_rtg, 0)::float8 AS kenpom_adj_o,
            COALESCE(k.d_rtg, 0)::float8 AS kenpom_adj_d,
            COALESCE(k.adj_t, 0)::float8 AS kenpom_adj_t,
            t.created_at
        FROM core.teams t
        JOIN core.schools s
          ON s.id = t.school_id
         AND s.deleted_at IS NULL
        JOIN core.tournaments tour
          ON tour.id = t.tournament_id
         AND tour.deleted_at IS NULL
        JOIN core.seasons seas
          ON seas.id = tour.season_id
         AND seas.deleted_at IS NULL
        LEFT JOIN core.team_kenpom_stats k
          ON k.team_id = t.id
         AND k.deleted_at IS NULL
        WHERE seas.year = %s
          AND t.deleted_at IS NULL
        ORDER BY t.seed, s.name
        """
        return pd.read_sql_query(query, conn, params=(year,))
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
        calcutta_id: Optional calcutta id for scoring rules
        tournament_simulation_batch_id: Optional batch id to filter by

    Returns:
        DataFrame with simulation results
    """
    conn = get_db_connection()
    try:
        # Note: Schema doesn't have points column; we calculate it from
        # wins+byes.

        batch_id = tournament_simulation_batch_id
        if not batch_id:
            with conn.cursor() as cur:
                cur.execute(
                    """
                    SELECT st.id
                    FROM core.tournaments tour
                    JOIN core.seasons seas
                      ON seas.id = tour.season_id
                     AND seas.deleted_at IS NULL
                    JOIN derived.simulated_tournaments st
                      ON st.tournament_id = tour.id
                     AND st.deleted_at IS NULL
                    WHERE seas.year = %s
                      AND tour.deleted_at IS NULL
                    ORDER BY st.created_at DESC
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
                s.name AS school_name,
                t.seed,
                t.region
            FROM derived.simulated_teams st
            JOIN core.teams t
              ON t.id = st.team_id
             AND t.deleted_at IS NULL
            JOIN core.schools s
              ON s.id = t.school_id
             AND s.deleted_at IS NULL
            JOIN core.tournaments tour
              ON tour.id = st.tournament_id
             AND tour.deleted_at IS NULL
            JOIN core.seasons seas
              ON seas.id = tour.season_id
             AND seas.deleted_at IS NULL
            WHERE seas.year = %s
              AND st.simulated_tournament_id = %s
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
                s.name AS school_name,
                t.seed,
                t.region
            FROM derived.simulated_teams st
            JOIN core.teams t
              ON t.id = st.team_id
             AND t.deleted_at IS NULL
            JOIN core.schools s
              ON s.id = t.school_id
             AND s.deleted_at IS NULL
            JOIN core.tournaments tour
              ON tour.id = st.tournament_id
             AND tour.deleted_at IS NULL
            JOIN core.seasons seas
              ON seas.id = tour.season_id
             AND seas.deleted_at IS NULL
            WHERE seas.year = %s
              AND st.simulated_tournament_id IS NULL
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
