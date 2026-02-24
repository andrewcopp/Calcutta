"""
Prediction batch readers.

Functions for reading predicted team values from compute.* tables
(prediction_batches, predicted_team_values).

DEPENDENCY NOTE: This module reads from these compute.* tables:
  - compute.prediction_batches   (latest batch lookup by tournament)
  - compute.predicted_team_values (team-level predictions: expected_points, p_round_*)
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

    Queries compute.prediction_batches to find the latest batch, then returns
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
                    FROM compute.prediction_batches pb
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
                FROM compute.predicted_team_values ptv
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


def enrich_with_analytical_probabilities(
    df: pd.DataFrame,
) -> pd.DataFrame:
    """
    Enrich team dataset with analytical championship probabilities.

    Reads from Go-generated predictions in compute.predicted_team_values.
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
    rows = read_latest_predicted_team_values(tournament_id)
    analytical_values: Dict[str, Tuple[float, float]] = {
        row.team_id: (row.p_championship, row.expected_points)
        for row in rows
    }

    if not analytical_values:
        raise ValueError(
            f"No Go predictions found for tournament {tournament_id}. "
            "Run the prediction pipeline before training models that use "
            "analytical features (compute.prediction_batches / "
            "compute.predicted_team_values)."
        )

    # Warn about teams missing from Go predictions (will be zero-filled)
    df_team_ids = set(df[id_col].astype(str).unique())
    missing = df_team_ids - set(analytical_values.keys())
    if missing:
        logger.warning(
            "Teams missing from Go predictions (zero-filled): %s (tournament=%s, %d/%d missing)",
            sorted(missing)[:5], tournament_id, len(missing), len(df_team_ids),
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
