"""
Ridge model dataset readers.

Functions for reading team features from core.* tables to build
training/inference datasets for the ridge regression model.
"""
from __future__ import annotations

from typing import Optional, List, Any

import pandas as pd

from moneyball.db.connection import get_db_connection


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


def _build_team_dataset_query(
    *,
    include_target: bool,
    exclude_clause: str = "",
) -> str:
    """
    Build the SQL query for reading a ridge team dataset.

    The shared FROM/JOIN/WHERE/ORDER BY clause is defined once; the
    target variant prepends CTEs and adds a observed_team_share_of_pool column.

    Args:
        include_target: When True, include team_bids CTEs and the
            observed_team_share_of_pool computed column.
        exclude_clause: Optional SQL fragment for filtering entries
            (only relevant when include_target is True).

    Returns:
        A parameterized SQL query string.
    """
    ctes = ""
    extra_select = ""
    extra_join = ""
    calcutta_key_expr = "NULL::text"

    if include_target:
        ctes = f"""
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
        """
        extra_select = """,
                CASE
                    WHEN (SELECT total_bid FROM total) > 0 THEN
                        COALESCE(tb.total_bid, 0)::float8
                        / (SELECT total_bid FROM total)
                    ELSE NULL
                END AS observed_team_share_of_pool"""
        extra_join = "\n            LEFT JOIN team_bids tb ON tb.team_id = tt.id"
        calcutta_key_expr = "%s::text"

    return f"""
        {ctes}
        SELECT
            %s::text AS snapshot,
            %s::text AS tournament_key,
            {calcutta_key_expr} AS calcutta_key,
            %s::text AS tournament_id,
            (%s::text || ':' || s.slug)::text AS team_key,
            s.name::text AS school_name,
            s.slug::text AS school_slug,
            tt.seed::int AS seed,
            tt.region::text AS region,
            COALESCE(k.net_rtg, 0)::float8 AS kenpom_net,
            COALESCE(k.o_rtg, 0)::float8 AS kenpom_o,
            COALESCE(k.d_rtg, 0)::float8 AS kenpom_d{extra_select}
        FROM core.teams tt
        JOIN core.schools s
          ON s.id = tt.school_id
         AND s.deleted_at IS NULL
        LEFT JOIN core.team_kenpom_stats k
          ON k.team_id = tt.id
         AND k.deleted_at IS NULL{extra_join}
        WHERE tt.tournament_id = %s
          AND tt.deleted_at IS NULL
        ORDER BY tt.seed ASC, s.name ASC;
    """


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
            cte_params: List[Any] = [calcutta_id]
            if exclude:
                exclude_clause = " AND ce.name <> ALL(%s::text[]) "
                cte_params.append(exclude)

            query = _build_team_dataset_query(
                include_target=True,
                exclude_clause=exclude_clause,
            )
            params: List[Any] = [
                *cte_params,
                snapshot,
                tournament_key,
                str(calcutta_id),
                str(tournament_id),
                tournament_key,
                str(tournament_id),
            ]
            return pd.read_sql_query(query, conn, params=tuple(params))

        query = _build_team_dataset_query(include_target=False)
        params_tuple = (
            snapshot,
            tournament_key,
            str(tournament_id),
            tournament_key,
            str(tournament_id),
        )
        return pd.read_sql_query(query, conn, params=params_tuple)
