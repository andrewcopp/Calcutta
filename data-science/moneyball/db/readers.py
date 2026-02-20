"""
Database readers for loading data from PostgreSQL.

This module provides functions to read data from the analytics database.
"""
from __future__ import annotations

from typing import Optional, Dict, Any, List
import pandas as pd
from psycopg2.extras import RealDictCursor

from moneyball.db.connection import get_db_connection
from moneyball.utils import points


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
    with get_db_connection() as conn:
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
                %s::text AS tournament_id,
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
                str(tournament_id),
                tournament_key,
                str(tournament_id),
            ]
            return pd.read_sql_query(query, conn, params=tuple(params))

        query = """
        SELECT
            %s::text AS snapshot,
            %s::text AS tournament_key,
            NULL::text AS calcutta_key,
            %s::text AS tournament_id,
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
            str(tournament_id),
            tournament_key,
            str(tournament_id),
        )
        return pd.read_sql_query(query, conn, params=params)


def read_points_by_win_index_for_year(year: int) -> Dict[int, float]:
    with get_db_connection() as conn:
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


def initialize_default_scoring_rules_for_year(year: int) -> Dict[int, float]:
    pbwi = read_points_by_win_index_for_year(year)
    points.set_default_points_by_win_index(pbwi)
    return pbwi
