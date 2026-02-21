"""
Market prediction generation for lab investment models.

This module contains the core prediction logic extracted from
scripts/generate_lab_predictions.py. It handles:
- Loading training data across historical years
- Running ridge regression to predict market share
- Translating predictions into lab entries with expected points

The script (scripts/generate_lab_predictions.py) handles CLI orchestration
and imports from this module.
"""

from __future__ import annotations

import logging
from typing import List, Optional, Tuple

import pandas as pd

from moneyball.db.lab_helpers import (
    get_expected_points_map,
    get_historical_calcuttas,
    get_team_id_map,
    HistoricalCalcutta,
)
from moneyball.lab.models import (
    Entry,
    InvestmentModel,
    Prediction,
    create_entry_with_predictions,
)

logger = logging.getLogger(__name__)


def generate_market_predictions(
    model_name: str,
    year: int,
    excluded_entry_name: Optional[str] = None,
) -> Tuple[Optional[pd.DataFrame], Optional[str]]:
    """
    Generate market predictions for a model against a specific year.

    Uses leave-one-year-out cross-validation: trains on all years except
    the target year, then predicts market share for the target year.

    Args:
        model_name: Name of the investment model to use.
        year: The target year to predict market share for.
        excluded_entry_name: Optional entry name to exclude from training data.

    Returns:
        Tuple of (predictions_df, error_message). On success, predictions_df
        has columns including team_slug and predicted_market_share.
        On failure, predictions_df is None and error_message describes the issue.
    """
    from moneyball.db.readers import (
        initialize_default_scoring_rules_for_year,
        read_ridge_team_dataset_for_year,
    )
    from moneyball.models.predicted_market_share import (
        predict_market_share,
    )
    from moneyball.lab.models import get_investment_model

    model = get_investment_model(model_name)
    if not model:
        return None, f"Model '{model_name}' not found"

    ridge_alpha = model.params.get("alpha", 1.0)
    feature_set = model.params.get("feature_set", "optimal")
    target_transform = model.params.get("target_transform", "none")
    seed_prior_monotone = model.params.get("seed_prior_monotone", None)
    seed_prior_k = model.params.get("seed_prior_k", 0.0)
    program_prior_k = model.params.get("program_prior_k", 0.0)

    initialize_default_scoring_rules_for_year(year)

    # Training years - derive from database instead of hardcoding
    all_years = sorted({c.year for c in get_historical_calcuttas()})
    train_years = [y for y in all_years if y != year]

    exclude_entry_names = [excluded_entry_name] if excluded_entry_name else None

    # Load training data
    train_frames = []
    train_failures = 0
    for y in train_years:
        try:
            df = read_ridge_team_dataset_for_year(
                y,
                exclude_entry_names=exclude_entry_names,
                include_target=True,
            )
            train_frames.append(df)
        except Exception as e:
            train_failures += 1
            logger.warning("Could not load training data for %d: %s", y, e)
            continue

    if train_failures > 0:
        return None, f"{train_failures}/{len(train_years)} training years failed to load"

    if not train_frames:
        return None, "No training data available"

    train_ds = pd.concat(train_frames, ignore_index=True)
    train_ds = train_ds[train_ds["observed_team_share_of_pool"].notna()].copy()

    if train_ds.empty:
        return None, "No valid training rows"

    try:
        predict_ds = read_ridge_team_dataset_for_year(
            year,
            exclude_entry_names=None,
            include_target=False,
        )
    except Exception as e:
        return None, f"Cannot load data for year {year}: {e}"

    # Run ridge regression to predict market share
    predictions_df = predict_market_share(
        train_team_dataset=train_ds,
        predict_team_dataset=predict_ds,
        ridge_alpha=ridge_alpha,
        feature_set=feature_set,
        target_transform=target_transform,
        seed_prior_monotone=seed_prior_monotone,
        seed_prior_k=seed_prior_k,
        program_prior_k=program_prior_k,
    )

    predictions_df["team_slug"] = predictions_df["team_key"].str.split(":").str[-1]

    return predictions_df, None


def create_predictions_for_calcutta(
    model_id: str,
    calcutta_id: str,
    predictions_df: pd.DataFrame,
    team_id_map: dict,
    expected_points_map: dict,
) -> Optional[Entry]:
    """
    Create a lab entry with market predictions for a single calcutta.

    Maps model predictions (by team_slug) to team IDs and expected points,
    then writes them to the lab.entries table.

    Args:
        model_id: UUID of the investment model.
        calcutta_id: UUID of the calcutta to create an entry for.
        predictions_df: DataFrame with team_slug and predicted_market_share.
        team_id_map: Mapping from school_slug to team UUID.
        expected_points_map: Mapping from school_slug to expected tournament points.

    Returns:
        The created Entry, or None if no valid predictions could be mapped.

    Raises:
        ValueError: If any team_slug in predictions has no team_id mapping,
            or if expected points are missing for a team.
    """
    predictions = []
    skipped_slugs = []
    for _, row in predictions_df.iterrows():
        team_slug = row["team_slug"]
        predicted_share = row["predicted_market_share"]

        team_id = team_id_map.get(team_slug)
        if not team_id:
            skipped_slugs.append(team_slug)
            continue

        expected_points = expected_points_map.get(team_slug)
        if expected_points is None:
            raise ValueError(
                f"No expected points for team {team_slug}. "
                "Run tournament simulations before generating predictions."
            )

        predictions.append(
            Prediction(
                team_id=team_id,
                predicted_market_share=predicted_share,
                expected_points=expected_points,
            )
        )

    if skipped_slugs:
        raise ValueError(
            f"Slug mismatch: {len(skipped_slugs)} teams in model output "
            f"have no team_id mapping: {skipped_slugs}. "
            "Check school slug consistency between model training data and core.schools."
        )

    if not predictions:
        return None

    entry = create_entry_with_predictions(
        investment_model_id=model_id,
        calcutta_id=calcutta_id,
        predictions=predictions,
        game_outcome_kind="kenpom",
        game_outcome_params={},
        starting_state_key="post_first_four",
    )

    return entry


def process_calcuttas(
    model: InvestmentModel,
    calcuttas: List[HistoricalCalcutta],
    excluded_entry: Optional[str] = None,
    dry_run: bool = False,
    log_fn: Optional[callable] = None,
) -> Tuple[int, Optional[str], List[str]]:
    """
    Iterate over calcuttas, generate predictions, and create lab entries.

    Args:
        model: The investment model object (has .id, .name).
        calcuttas: List of HistoricalCalcutta to process.
        excluded_entry: Optional entry name to exclude from training.
        dry_run: If True, show what would be done without writing.
        log_fn: Optional callback for progress messages. Called with a
            single string argument. If None, messages are suppressed.

    Returns:
        Tuple of (entries_created, last_entry_id, errors).
    """
    def log(msg: str) -> None:
        if log_fn is not None:
            log_fn(msg)

    entries_created = 0
    errors: List[str] = []
    last_entry_id: Optional[str] = None

    for calcutta in calcuttas:
        year = calcutta.year
        log(f"Processing {calcutta.name} ({year})...")

        predictions_df, error = generate_market_predictions(
            model.name, year, excluded_entry
        )

        if error:
            log(f"  Skipping: {error}")
            errors.append(f"{calcutta.name}: {error}")
            continue

        if predictions_df is None or predictions_df.empty:
            log(f"  Skipping: No predictions generated")
            continue

        team_id_map = get_team_id_map(calcutta.tournament_id)
        if not team_id_map:
            log(f"  Skipping: No teams found")
            continue

        # Get expected points for each team from simulation data
        expected_points_map = get_expected_points_map(calcutta.id)

        if dry_run:
            pred_count = len(
                [
                    1
                    for _, row in predictions_df.iterrows()
                    if team_id_map.get(row["team_slug"])
                ]
            )
            log(f"  Would create entry with {pred_count} predictions")
        else:
            entry = create_predictions_for_calcutta(
                model.id,
                calcutta.id,
                predictions_df,
                team_id_map,
                expected_points_map,
            )
            if entry:
                log(
                    f"  Created entry {entry.id} with {len(entry.predictions)} predictions"
                )
                entries_created += 1
                last_entry_id = entry.id
            else:
                log(f"  No entry created (no valid predictions)")

    return entries_created, last_entry_id, errors
