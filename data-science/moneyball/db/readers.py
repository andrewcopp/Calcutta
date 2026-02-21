"""
Database readers for loading data from PostgreSQL.

This module provides functions to read data from the analytics database.
"""
from __future__ import annotations

import logging
from typing import Optional, Dict, Any, List, Tuple
import pandas as pd
from psycopg2.extras import RealDictCursor

from moneyball.db.connection import get_db_connection
from moneyball.utils import points

logger = logging.getLogger(__name__)


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
            COALESCE(k.d_rtg, 0)::float8 AS kenpom_d,
            COALESCE(k.adj_t, 0)::float8 AS kenpom_adj_t{extra_select}
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
    return read_points_by_win_index_for_year(year)


def read_latest_predicted_team_values(
    tournament_id: str,
) -> List[Dict[str, Any]]:
    """
    Get predicted team values from the latest prediction batch for a tournament.

    Queries derived.prediction_batches to find the latest batch, then returns
    all predicted_team_values rows joined with team/school metadata.

    This is the canonical function for reading prediction batch data.
    Other modules should use this rather than querying prediction_batches directly.

    Args:
        tournament_id: The tournament UUID to look up predictions for.

    Returns:
        List of dicts with keys: team_id, team_slug, p_championship,
        expected_points. Returns empty list if no prediction batch exists.
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("""
                WITH latest_batch AS (
                    SELECT pb.id
                    FROM derived.prediction_batches pb
                    WHERE pb.tournament_id = %s::uuid
                        AND pb.deleted_at IS NULL
                    ORDER BY pb.created_at DESC
                    LIMIT 1
                )
                SELECT
                    ptv.team_id::text,
                    s.slug::text AS team_slug,
                    COALESCE(ptv.p_round_7, 0)::float AS p_championship,
                    COALESCE(ptv.expected_points, 0)::float AS expected_points
                FROM derived.predicted_team_values ptv
                JOIN core.teams t ON t.id = ptv.team_id AND t.deleted_at IS NULL
                JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
                WHERE ptv.prediction_batch_id = (SELECT id FROM latest_batch)
                    AND ptv.deleted_at IS NULL
            """, (tournament_id,))
            return [
                {
                    "team_id": str(row[0]),
                    "team_slug": row[1],
                    "p_championship": float(row[2]),
                    "expected_points": float(row[3]),
                }
                for row in cur.fetchall()
            ]


def read_analytical_values_from_db(
    tournament_id: str,
) -> Dict[str, Tuple[float, float]]:
    """
    Get analytical values from Go-generated predictions in the database.

    Returns:
        Dict mapping team_id -> (p_championship, expected_points)
    """
    rows = read_latest_predicted_team_values(tournament_id)
    return {
        row["team_id"]: (row["p_championship"], row["expected_points"])
        for row in rows
    }


def enrich_with_analytical_probabilities(
    df: pd.DataFrame,
) -> pd.DataFrame:
    """
    Enrich team dataset with analytical championship probabilities.

    Reads from Go-generated predictions in derived.predicted_team_values.
    Requires predictions to be pre-generated for the tournament.

    Args:
        df: Team dataset with id/team_key and tournament_id

    Returns:
        DataFrame with added predicted_p_championship and
        predicted_expected_points columns
    """
    # Use team_key or id as identifier
    id_col = "team_key" if "team_key" in df.columns else "id"
    if id_col not in df.columns:
        raise ValueError("team_dataset must have 'team_key' or 'id' column")

    # Get tournament_id from the dataframe
    if "tournament_id" in df.columns:
        tournament_id = str(df["tournament_id"].iloc[0])
    elif "tournament_key" in df.columns:
        tournament_id = str(df["tournament_key"].iloc[0])
    else:
        raise ValueError(
            "team_dataset must have 'tournament_id' or 'tournament_key' column"
        )

    # Get analytical values from database
    analytical_values = read_analytical_values_from_db(tournament_id)

    if not analytical_values:
        logger.warning(
            "No Go predictions found for tournament %s. "
            "Setting analytical values to 0.",
            tournament_id,
        )

    # Add columns to dataframe
    result = df.copy()
    result["predicted_p_championship"] = result[id_col].astype(str).map(
        lambda tid: analytical_values.get(tid, (0.0, 0.0))[0]
    )
    result["predicted_expected_points"] = result[id_col].astype(str).map(
        lambda tid: analytical_values.get(tid, (0.0, 0.0))[1]
    )

    # Ensure numeric types
    result["predicted_p_championship"] = pd.to_numeric(
        result["predicted_p_championship"], errors="coerce"
    ).fillna(0.0)
    result["predicted_expected_points"] = pd.to_numeric(
        result["predicted_expected_points"], errors="coerce"
    ).fillna(0.0)

    return result
