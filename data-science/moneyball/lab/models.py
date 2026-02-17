"""
Lab data models and database writers.

Provides dataclasses and functions for creating lab entities.
"""

import json
import logging
import uuid
from dataclasses import dataclass, field
from datetime import datetime
from typing import Any, Dict, List, Optional

from moneyball.db.connection import get_db_connection

logger = logging.getLogger(__name__)


@dataclass
class InvestmentModel:
    """An investment prediction model being tested."""

    id: str
    name: str
    kind: str
    params: Dict[str, Any] = field(default_factory=dict)
    notes: Optional[str] = None
    created_at: Optional[datetime] = None


@dataclass
class Prediction:
    """A market prediction for a single team.

    This represents what the model predicts THE MARKET will bid.
    """

    team_id: str
    predicted_market_share: float  # Expected share of total pool (0.0-1.0)
    expected_points: float  # Expected tournament points from KenPom simulation


@dataclass
class Bid:
    """A single optimized bid in an entry.

    This represents OUR optimal bid allocation.
    """

    team_id: str
    bid_points: int
    expected_roi: float = 0.0  # expected_points / (predicted_market + our_bid)


@dataclass
class Entry:
    """An entry with market predictions and optimized bids.

    Pipeline stages:
    1. predictions: What the model predicts the market will bid
    2. bids: Our optimal allocation given predictions + expected points
    """

    id: str
    investment_model_id: str
    calcutta_id: str
    game_outcome_kind: str
    game_outcome_params: Dict[str, Any]
    optimizer_kind: str
    optimizer_params: Dict[str, Any]
    starting_state_key: str
    predictions: List[Prediction]  # Market predictions
    bids: List[Bid]  # Optimized bids
    created_at: Optional[datetime] = None


def create_investment_model(
    name: str,
    kind: str,
    params: Optional[Dict[str, Any]] = None,
    notes: Optional[str] = None,
) -> InvestmentModel:
    """
    Create a new investment model.

    Args:
        name: Unique name for the model (e.g. "ridge-v3-shrunk")
        kind: Model type (ridge, random_forest, xgboost, oracle, naive_ev)
        params: Model hyperparameters
        notes: Free-form notes about the model

    Returns:
        The created InvestmentModel

    Raises:
        ValueError: If a model with the same name already exists
    """
    model_id = str(uuid.uuid4())
    params = params or {}

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO lab.investment_models (id, name, kind, params_json, notes)
                VALUES (%s, %s, %s, %s::jsonb, %s)
                RETURNING id, created_at
                """,
                (model_id, name, kind, json.dumps(params), notes),
            )
            row = cur.fetchone()

    logger.info(f"Created investment model: {name} ({kind})")
    return InvestmentModel(
        id=row[0] if row else model_id,
        name=name,
        kind=kind,
        params=params,
        notes=notes,
        created_at=row[1] if row else None,
    )


def get_investment_model(name: str) -> Optional[InvestmentModel]:
    """
    Get an investment model by name.

    Args:
        name: Model name to look up

    Returns:
        InvestmentModel if found, None otherwise
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT id, name, kind, params_json, notes, created_at
                FROM lab.investment_models
                WHERE name = %s AND deleted_at IS NULL
                """,
                (name,),
            )
            row = cur.fetchone()

    if not row:
        return None

    return InvestmentModel(
        id=str(row[0]),
        name=row[1],
        kind=row[2],
        params=row[3] if row[3] else {},
        notes=row[4],
        created_at=row[5],
    )


def get_investment_model_by_id(model_id: str) -> Optional[InvestmentModel]:
    """
    Get an investment model by ID.

    Args:
        model_id: Model UUID to look up

    Returns:
        InvestmentModel if found, None otherwise
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT id, name, kind, params_json, notes, created_at
                FROM lab.investment_models
                WHERE id = %s AND deleted_at IS NULL
                """,
                (model_id,),
            )
            row = cur.fetchone()

    if not row:
        return None

    return InvestmentModel(
        id=str(row[0]),
        name=row[1],
        kind=row[2],
        params=row[3] if row[3] else {},
        notes=row[4],
        created_at=row[5],
    )


def get_or_create_investment_model(
    name: str,
    kind: str,
    params: Optional[Dict[str, Any]] = None,
    notes: Optional[str] = None,
) -> InvestmentModel:
    """
    Get an existing investment model by name, or create if it doesn't exist.

    Args:
        name: Unique name for the model
        kind: Model type (only used if creating)
        params: Model hyperparameters (only used if creating)
        notes: Free-form notes (only used if creating)

    Returns:
        The existing or newly created InvestmentModel
    """
    existing = get_investment_model(name)
    if existing:
        return existing
    return create_investment_model(name, kind, params, notes)


def create_entry_with_predictions(
    investment_model_id: str,
    calcutta_id: str,
    predictions: List[Prediction],
    game_outcome_kind: str = "kenpom",
    game_outcome_params: Optional[Dict[str, Any]] = None,
    starting_state_key: str = "post_first_four",
) -> Entry:
    """
    Create a new entry with market predictions (stage 2 of pipeline).

    This stores what the model predicts the market will bid. The entry
    will have predictions_json populated but bids_json will be empty
    until optimization runs.

    Args:
        investment_model_id: ID of the investment model that generated predictions
        calcutta_id: ID of the calcutta
        predictions: List of Prediction objects with market predictions
        game_outcome_kind: Game outcome model type (default: kenpom)
        game_outcome_params: Game outcome model parameters
        starting_state_key: Tournament state

    Returns:
        The created Entry (with empty bids)
    """
    entry_id = str(uuid.uuid4())
    game_outcome_params = game_outcome_params or {}

    predictions_json = [
        {
            "team_id": p.team_id,
            "predicted_market_share": p.predicted_market_share,
            "expected_points": p.expected_points,
        }
        for p in predictions
    ]

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO lab.entries (
                    id, investment_model_id, calcutta_id,
                    game_outcome_kind, game_outcome_params_json,
                    optimizer_kind, optimizer_params_json,
                    starting_state_key, predictions_json, bids_json
                )
                VALUES (%s, %s, %s, %s, %s::jsonb, %s, %s::jsonb, %s, %s::jsonb, %s::jsonb)
                ON CONFLICT (investment_model_id, calcutta_id, starting_state_key)
                WHERE deleted_at IS NULL
                DO UPDATE SET
                    game_outcome_kind = EXCLUDED.game_outcome_kind,
                    game_outcome_params_json = EXCLUDED.game_outcome_params_json,
                    predictions_json = EXCLUDED.predictions_json,
                    updated_at = NOW()
                RETURNING id, created_at
                """,
                (
                    entry_id,
                    investment_model_id,
                    calcutta_id,
                    game_outcome_kind,
                    json.dumps(game_outcome_params),
                    "pending",  # optimizer_kind - not yet optimized
                    json.dumps({}),
                    starting_state_key,
                    json.dumps(predictions_json),
                    json.dumps([]),  # Empty bids until optimization
                ),
            )
            row = cur.fetchone()

    logger.info(f"Created entry with {len(predictions)} predictions for calcutta {calcutta_id}")
    return Entry(
        id=str(row[0]) if row else entry_id,
        investment_model_id=investment_model_id,
        calcutta_id=calcutta_id,
        game_outcome_kind=game_outcome_kind,
        game_outcome_params=game_outcome_params,
        optimizer_kind="pending",
        optimizer_params={},
        starting_state_key=starting_state_key,
        predictions=predictions,
        bids=[],
        created_at=row[1] if row else None,
    )


def update_entry_with_bids(
    entry_id: str,
    bids: List[Bid],
    optimizer_kind: str = "minlp",
    optimizer_params: Optional[Dict[str, Any]] = None,
) -> None:
    """
    Update an entry with optimized bids (stage 3 of pipeline).

    Args:
        entry_id: ID of the entry to update
        bids: List of optimized Bid objects
        optimizer_kind: Optimizer type used
        optimizer_params: Optimizer parameters
    """
    optimizer_params = optimizer_params or {}

    bids_json = [
        {
            "team_id": b.team_id,
            "bid_points": b.bid_points,
            "expected_roi": b.expected_roi,
        }
        for b in bids
    ]

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                UPDATE lab.entries
                SET
                    optimizer_kind = %s,
                    optimizer_params_json = %s::jsonb,
                    bids_json = %s::jsonb,
                    updated_at = NOW()
                WHERE id = %s AND deleted_at IS NULL
                """,
                (
                    optimizer_kind,
                    json.dumps(optimizer_params),
                    json.dumps(bids_json),
                    entry_id,
                ),
            )

    logger.info(f"Updated entry {entry_id} with {len(bids)} optimized bids")


def get_entry(entry_id: str) -> Optional[Entry]:
    """
    Get an entry by ID.

    Args:
        entry_id: Entry UUID

    Returns:
        Entry if found, None otherwise
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                SELECT id, investment_model_id, calcutta_id,
                       game_outcome_kind, game_outcome_params_json,
                       optimizer_kind, optimizer_params_json,
                       starting_state_key, predictions_json, bids_json, created_at
                FROM lab.entries
                WHERE id = %s AND deleted_at IS NULL
                """,
                (entry_id,),
            )
            row = cur.fetchone()

    if not row:
        return None

    predictions_data = row[8] if row[8] else []
    predictions = [
        Prediction(
            team_id=p["team_id"],
            predicted_market_share=p["predicted_market_share"],
            expected_points=p["expected_points"],
        )
        for p in predictions_data
    ]

    bids_data = row[9] if row[9] else []
    bids = [
        Bid(
            team_id=b["team_id"],
            bid_points=b["bid_points"],
            expected_roi=b.get("expected_roi", 0.0),
        )
        for b in bids_data
    ]

    return Entry(
        id=str(row[0]),
        investment_model_id=str(row[1]),
        calcutta_id=str(row[2]),
        game_outcome_kind=row[3],
        game_outcome_params=row[4] if row[4] else {},
        optimizer_kind=row[5],
        optimizer_params=row[6] if row[6] else {},
        starting_state_key=row[7],
        predictions=predictions,
        bids=bids,
        created_at=row[10],
    )


def get_entries_pending_optimization(
    investment_model_id: Optional[str] = None,
) -> List[Entry]:
    """
    Get entries that have predictions but no optimized bids yet.

    Args:
        investment_model_id: Optionally filter by model ID

    Returns:
        List of Entry objects pending optimization
    """
    with get_db_connection() as conn:
        with conn.cursor() as cur:
            if investment_model_id:
                cur.execute(
                    """
                    SELECT id, investment_model_id, calcutta_id,
                           game_outcome_kind, game_outcome_params_json,
                           optimizer_kind, optimizer_params_json,
                           starting_state_key, predictions_json, bids_json, created_at
                    FROM lab.entries
                    WHERE investment_model_id = %s
                      AND deleted_at IS NULL
                      AND predictions_json IS NOT NULL
                      AND (bids_json IS NULL OR bids_json = '[]'::jsonb)
                    ORDER BY created_at
                    """,
                    (investment_model_id,),
                )
            else:
                cur.execute(
                    """
                    SELECT id, investment_model_id, calcutta_id,
                           game_outcome_kind, game_outcome_params_json,
                           optimizer_kind, optimizer_params_json,
                           starting_state_key, predictions_json, bids_json, created_at
                    FROM lab.entries
                    WHERE deleted_at IS NULL
                      AND predictions_json IS NOT NULL
                      AND (bids_json IS NULL OR bids_json = '[]'::jsonb)
                    ORDER BY created_at
                    """
                )
            rows = cur.fetchall()

    entries = []
    for row in rows:
        predictions_data = row[8] if row[8] else []
        predictions = [
            Prediction(
                team_id=p["team_id"],
                predicted_market_share=p["predicted_market_share"],
                expected_points=p["expected_points"],
            )
            for p in predictions_data
        ]

        bids_data = row[9] if row[9] else []
        bids = [
            Bid(
                team_id=b["team_id"],
                bid_points=b["bid_points"],
                expected_roi=b.get("expected_roi", 0.0),
            )
            for b in bids_data
        ]

        entries.append(Entry(
            id=str(row[0]),
            investment_model_id=str(row[1]),
            calcutta_id=str(row[2]),
            game_outcome_kind=row[3],
            game_outcome_params=row[4] if row[4] else {},
            optimizer_kind=row[5],
            optimizer_params=row[6] if row[6] else {},
            starting_state_key=row[7],
            predictions=predictions,
            bids=bids,
            created_at=row[10],
        ))

    return entries
