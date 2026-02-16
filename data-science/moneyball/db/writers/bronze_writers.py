"""
Bronze layer database lookups.

The lab_bronze schema no longer exists. Tournament and team data now lives
in core.tournaments / core.teams / core.schools.  These functions preserve
the original signatures so callers do not break, but they read from core.*
instead of writing to lab_bronze.*.
"""
import logging
import pandas as pd
from typing import Dict
from moneyball.db.connection import get_db_connection

logger = logging.getLogger(__name__)


def get_or_create_tournament(season: int) -> str:
    """
    Look up the core tournament id for a given season year.

    This previously created rows in lab_bronze.tournaments, but that schema
    no longer exists.  The data already lives in core.tournaments joined to
    core.seasons.

    Args:
        season: Tournament year (e.g., 2025)

    Returns:
        tournament_id (UUID as string)

    Raises:
        ValueError: If no tournament exists for the given season.
    """
    with get_db_connection() as conn:
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
                (season,),
            )
            result = cur.fetchone()
            if result:
                return str(result[0])

            raise ValueError(
                f"No core tournament found for season {season}. "
                "Tournaments must be created via the Go API or migrations."
            )


def write_teams(tournament_id: str, teams_df: pd.DataFrame) -> Dict[str, str]:
    """
    Look up team ids for a tournament, returning a school_slug -> team_id map.

    This previously wrote to lab_bronze.teams, but that schema no longer
    exists.  Team data already lives in core.teams / core.schools.

    Args:
        tournament_id: Tournament UUID
        teams_df: DataFrame (unused, kept for signature compatibility)

    Returns:
        Dict mapping school_slug to team_id (UUID as string)
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT s.slug, t.id
                FROM core.teams t
                JOIN core.schools s
                  ON s.id = t.school_id
                 AND s.deleted_at IS NULL
                WHERE t.tournament_id = %s
                  AND t.deleted_at IS NULL
                """,
                (tournament_id,),
            )
            return {str(row[0]): str(row[1]) for row in cur.fetchall()}
