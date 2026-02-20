"""
Shared database helpers for lab pipeline scripts.
"""
from __future__ import annotations

import logging
from dataclasses import dataclass
from typing import Dict, List, Optional

from moneyball.db.connection import get_db_connection

logger = logging.getLogger(__name__)


@dataclass
class HistoricalCalcutta:
    """A historical calcutta with its rules and entry count."""

    id: str
    name: str
    year: int
    tournament_id: str
    budget_points: Optional[int]
    min_teams: Optional[int]
    max_teams: Optional[int]
    max_bid: Optional[int]
    entry_count: int


def get_historical_calcuttas() -> List[HistoricalCalcutta]:
    """Get all historical calcuttas from the database with their rules and entry counts."""
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT
                    c.id,
                    c.name,
                    s.year,
                    t.id as tournament_id,
                    c.budget_points,
                    c.min_teams,
                    c.max_teams,
                    c.max_bid,
                    (SELECT COUNT(*) FROM core.entries e WHERE e.calcutta_id = c.id AND e.deleted_at IS NULL) as entry_count
                FROM core.calcuttas c
                JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
                JOIN core.seasons s ON s.id = t.season_id AND s.deleted_at IS NULL
                WHERE c.deleted_at IS NULL
                ORDER BY s.year DESC
            """)
            return [
                HistoricalCalcutta(
                    id=str(row[0]),
                    name=row[1],
                    year=row[2],
                    tournament_id=str(row[3]),
                    budget_points=row[4],
                    min_teams=row[5],
                    max_teams=row[6],
                    max_bid=row[7],
                    entry_count=row[8],
                )
                for row in cur.fetchall()
            ]


def get_team_id_map(tournament_id: str) -> Dict[str, str]:
    """Get mapping from school_slug to team_id for a tournament."""
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT s.slug, t.id
                FROM core.teams t
                JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
                WHERE t.tournament_id = %s AND t.deleted_at IS NULL
            """, (tournament_id,))
            return {row[0]: str(row[1]) for row in cur.fetchall()}


def get_expected_points_map(calcutta_id: str) -> Dict[str, float]:
    """
    Get expected tournament points for each team.

    First tries to use Go-generated predictions from derived.predicted_team_values.
    Falls back to simulation-based calculation if no predictions exist.
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # First, try to get from Go-generated predictions
            cur.execute("""
                WITH calcutta_ctx AS (
                    SELECT c.id AS calcutta_id, t.id AS tournament_id
                    FROM core.calcuttas c
                    JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
                    WHERE c.id = %s AND c.deleted_at IS NULL
                ),
                latest_batch AS (
                    SELECT pb.id
                    FROM derived.prediction_batches pb
                    WHERE pb.tournament_id = (SELECT tournament_id FROM calcutta_ctx)
                        AND pb.deleted_at IS NULL
                    ORDER BY pb.created_at DESC
                    LIMIT 1
                )
                SELECT s.slug AS team_slug, ptv.expected_points::float
                FROM derived.predicted_team_values ptv
                JOIN core.teams t ON t.id = ptv.team_id AND t.deleted_at IS NULL
                JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
                WHERE ptv.prediction_batch_id = (SELECT id FROM latest_batch)
                    AND ptv.deleted_at IS NULL
            """, (calcutta_id,))
            result = {row[0]: row[1] for row in cur.fetchall()}

            if result:
                logger.info("Using Go-generated predictions for calcutta %s (%d teams)", calcutta_id, len(result))
                return result

            # Fallback: compute from simulations
            logger.info("No Go predictions found, falling back to simulation-based calculation for calcutta %s", calcutta_id)
            cur.execute("""
                WITH calcutta_ctx AS (
                    SELECT c.id AS calcutta_id, t.id AS tournament_id
                    FROM core.calcuttas c
                    JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
                    WHERE c.id = %s AND c.deleted_at IS NULL
                ),
                win_distribution AS (
                    SELECT
                        st.team_id,
                        st.wins,
                        st.byes,
                        COUNT(*)::float AS sim_count
                    FROM derived.simulated_teams st
                    WHERE st.tournament_id = (SELECT tournament_id FROM calcutta_ctx)
                      AND st.deleted_at IS NULL
                    GROUP BY st.team_id, st.wins, st.byes
                ),
                team_totals AS (
                    SELECT team_id, SUM(sim_count) AS total_sims
                    FROM win_distribution
                    GROUP BY team_id
                ),
                team_expected AS (
                    SELECT
                        s.slug AS team_slug,
                        SUM(
                            wd.sim_count * core.calcutta_points_for_progress(
                                (SELECT calcutta_id FROM calcutta_ctx),
                                wd.wins,
                                wd.byes
                            )
                        ) / tt.total_sims AS expected_points
                    FROM win_distribution wd
                    JOIN team_totals tt ON tt.team_id = wd.team_id
                    JOIN core.teams t ON t.id = wd.team_id AND t.deleted_at IS NULL
                    JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
                    GROUP BY s.slug, tt.total_sims
                )
                SELECT team_slug, expected_points::float FROM team_expected
            """, (calcutta_id,))
            result = {row[0]: row[1] for row in cur.fetchall()}

            if not result:
                raise ValueError(
                    f"No prediction or simulation data for calcutta {calcutta_id}. "
                    "Run predictions or simulations before generating market predictions."
                )

            return result
