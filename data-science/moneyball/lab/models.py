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
class Bid:
    """A single bid in an entry."""

    team_id: str
    bid_points: int
    expected_roi: float = 0.0


@dataclass
class Entry:
    """An optimized entry produced by an investment model."""

    id: str
    investment_model_id: str
    calcutta_id: str
    game_outcome_kind: str
    game_outcome_params: Dict[str, Any]
    optimizer_kind: str
    optimizer_params: Dict[str, Any]
    starting_state_key: str
    bids: List[Bid]
    created_at: Optional[datetime] = None


@dataclass
class Evaluation:
    """Simulation evaluation results for an entry."""

    id: str
    entry_id: str
    n_sims: int
    seed: int
    mean_normalized_payout: Optional[float] = None
    median_normalized_payout: Optional[float] = None
    p_top1: Optional[float] = None
    p_in_money: Optional[float] = None
    our_rank: Optional[int] = None
    simulated_calcutta_id: Optional[str] = None
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
            conn.commit()

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


def create_entry(
    investment_model_id: str,
    calcutta_id: str,
    bids: List[Bid],
    game_outcome_kind: str = "kenpom",
    game_outcome_params: Optional[Dict[str, Any]] = None,
    optimizer_kind: str = "minlp",
    optimizer_params: Optional[Dict[str, Any]] = None,
    starting_state_key: str = "post_first_four",
) -> Entry:
    """
    Create a new entry (set of optimized bids).

    Args:
        investment_model_id: ID of the investment model that generated this entry
        calcutta_id: ID of the calcutta this entry is for
        bids: List of Bid objects representing team bids
        game_outcome_kind: Game outcome model type (default: kenpom)
        game_outcome_params: Game outcome model parameters
        optimizer_kind: Optimizer type (default: minlp)
        optimizer_params: Optimizer parameters
        starting_state_key: Tournament state (pre_tournament, post_first_four, current)

    Returns:
        The created Entry
    """
    entry_id = str(uuid.uuid4())
    game_outcome_params = game_outcome_params or {}
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
                INSERT INTO lab.entries (
                    id, investment_model_id, calcutta_id,
                    game_outcome_kind, game_outcome_params_json,
                    optimizer_kind, optimizer_params_json,
                    starting_state_key, bids_json
                )
                VALUES (%s, %s, %s, %s, %s::jsonb, %s, %s::jsonb, %s, %s::jsonb)
                ON CONFLICT ON CONSTRAINT uq_lab_entries_model_calcutta_state
                DO UPDATE SET
                    game_outcome_kind = EXCLUDED.game_outcome_kind,
                    game_outcome_params_json = EXCLUDED.game_outcome_params_json,
                    optimizer_kind = EXCLUDED.optimizer_kind,
                    optimizer_params_json = EXCLUDED.optimizer_params_json,
                    bids_json = EXCLUDED.bids_json,
                    updated_at = NOW()
                RETURNING id, created_at
                """,
                (
                    entry_id,
                    investment_model_id,
                    calcutta_id,
                    game_outcome_kind,
                    json.dumps(game_outcome_params),
                    optimizer_kind,
                    json.dumps(optimizer_params),
                    starting_state_key,
                    json.dumps(bids_json),
                ),
            )
            row = cur.fetchone()
            conn.commit()

    logger.info(f"Created entry for calcutta {calcutta_id} with {len(bids)} bids")
    return Entry(
        id=str(row[0]) if row else entry_id,
        investment_model_id=investment_model_id,
        calcutta_id=calcutta_id,
        game_outcome_kind=game_outcome_kind,
        game_outcome_params=game_outcome_params,
        optimizer_kind=optimizer_kind,
        optimizer_params=optimizer_params,
        starting_state_key=starting_state_key,
        bids=bids,
        created_at=row[1] if row else None,
    )


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
                       starting_state_key, bids_json, created_at
                FROM lab.entries
                WHERE id = %s AND deleted_at IS NULL
                """,
                (entry_id,),
            )
            row = cur.fetchone()

    if not row:
        return None

    bids_data = row[8] if row[8] else []
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
        bids=bids,
        created_at=row[9],
    )


def create_evaluation(
    entry_id: str,
    n_sims: int,
    seed: int,
    mean_normalized_payout: Optional[float] = None,
    median_normalized_payout: Optional[float] = None,
    p_top1: Optional[float] = None,
    p_in_money: Optional[float] = None,
    our_rank: Optional[int] = None,
    simulated_calcutta_id: Optional[str] = None,
) -> Evaluation:
    """
    Record evaluation results for an entry.

    Args:
        entry_id: ID of the entry being evaluated
        n_sims: Number of simulations run
        seed: Random seed used
        mean_normalized_payout: Average normalized payout (THE metric)
        median_normalized_payout: Median normalized payout
        p_top1: Probability of finishing 1st
        p_in_money: Probability of finishing in the money
        our_rank: Rank among all entries
        simulated_calcutta_id: Optional link to simulation infrastructure

    Returns:
        The created Evaluation
    """
    eval_id = str(uuid.uuid4())

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute(
                """
                INSERT INTO lab.evaluations (
                    id, entry_id, n_sims, seed,
                    mean_normalized_payout, median_normalized_payout,
                    p_top1, p_in_money, our_rank, simulated_calcutta_id
                )
                VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                ON CONFLICT ON CONSTRAINT uq_lab_evaluations_entry_sims_seed
                DO UPDATE SET
                    mean_normalized_payout = EXCLUDED.mean_normalized_payout,
                    median_normalized_payout = EXCLUDED.median_normalized_payout,
                    p_top1 = EXCLUDED.p_top1,
                    p_in_money = EXCLUDED.p_in_money,
                    our_rank = EXCLUDED.our_rank,
                    simulated_calcutta_id = EXCLUDED.simulated_calcutta_id,
                    updated_at = NOW()
                RETURNING id, created_at
                """,
                (
                    eval_id,
                    entry_id,
                    n_sims,
                    seed,
                    mean_normalized_payout,
                    median_normalized_payout,
                    p_top1,
                    p_in_money,
                    our_rank,
                    simulated_calcutta_id,
                ),
            )
            row = cur.fetchone()
            conn.commit()

    logger.info(
        f"Created evaluation for entry {entry_id}: "
        f"mean_payout={mean_normalized_payout}"
    )
    return Evaluation(
        id=str(row[0]) if row else eval_id,
        entry_id=entry_id,
        n_sims=n_sims,
        seed=seed,
        mean_normalized_payout=mean_normalized_payout,
        median_normalized_payout=median_normalized_payout,
        p_top1=p_top1,
        p_in_money=p_in_money,
        our_rank=our_rank,
        simulated_calcutta_id=simulated_calcutta_id,
        created_at=row[1] if row else None,
    )
