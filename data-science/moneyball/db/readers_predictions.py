"""
Prediction batch readers.

Functions for reading predicted team values from derived.* tables
(prediction_batches, predicted_team_values).

DEPENDENCY NOTE: This module reads from these derived.* tables:
  - derived.prediction_batches   (latest batch lookup by tournament)
  - derived.predicted_team_values (team-level predictions: expected_points, p_round_*)
These tables are populated by the Go prediction pipeline. If they are archived
or renamed, the optimal_v3 feature set will break silently (the enrichment
function returns zeros when no data is found).
"""
from __future__ import annotations

import logging
from dataclasses import dataclass
from typing import Dict, List, Tuple

import pandas as pd

from moneyball.db.connection import get_db_connection

logger = logging.getLogger(__name__)


@dataclass(frozen=True)
class PredictedTeamValue:
    team_id: str
    team_slug: str
    p_championship: float
    expected_points: float


def read_latest_predicted_team_values(
    tournament_id: str,
) -> List[PredictedTeamValue]:
    """
    Get predicted team values from the latest prediction batch for a tournament.

    Queries derived.prediction_batches to find the latest batch, then returns
    all predicted_team_values rows joined with team/school metadata.

    This is the canonical function for reading prediction batch data.
    Other modules should use this rather than querying prediction_batches directly.

    Args:
        tournament_id: The tournament UUID to look up predictions for.

    Returns:
        List of PredictedTeamValue dataclasses. Returns empty list if no
        prediction batch exists.
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
                PredictedTeamValue(
                    team_id=str(row[0]),
                    team_slug=row[1],
                    p_championship=float(row[2]),
                    expected_points=float(row[3]),
                )
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
        row.team_id: (row.p_championship, row.expected_points)
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
