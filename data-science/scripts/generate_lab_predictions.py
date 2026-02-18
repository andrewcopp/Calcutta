#!/usr/bin/env python3
"""
Generate market predictions for a model across all historical calcuttas.

This is stage 2 of the lab pipeline:
1. Register model (lab.investment_models)
2. Generate predictions (this script) -> predictions_json
3. Optimize entry (optimize_lab_entries.py) -> bids_json
4. Evaluate (lab pipeline worker) -> lab.evaluations

The script predicts what THE MARKET will bid on each team based on
the investment model. It stores:
- predicted_market_share: fraction of pool the market will spend on this team
- expected_points: expected tournament points from KenPom simulation

Usage:
    python generate_lab_predictions.py --model-name ridge-v1
    python generate_lab_predictions.py --model-id <uuid> --years 2023,2024
"""
import argparse
import json
import logging
import os
import sys
from pathlib import Path
from typing import Optional

# Add project root to path
project_root = Path(__file__).resolve().parents[1]
if str(project_root) not in sys.path:
    sys.path.insert(0, str(project_root))


from moneyball.db.lab_helpers import (
    get_historical_calcuttas,
    get_team_id_map,
    get_expected_points_map,
)


def generate_market_predictions(model_name: str, year: int, excluded_entry_name: Optional[str] = None):
    """
    Generate market predictions for a model against a specific year.

    Returns DataFrame with columns: team_slug, predicted_auction_share_of_pool
    """
    from moneyball.db.readers import (
        initialize_default_scoring_rules_for_year,
        read_ridge_team_dataset_for_year,
    )
    from moneyball.models.predicted_auction_share_of_pool import (
        predict_auction_share_of_pool,
    )
    from moneyball.lab.models import get_investment_model
    import pandas as pd

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
    for y in train_years:
        try:
            df = read_ridge_team_dataset_for_year(
                y,
                exclude_entry_names=exclude_entry_names,
                include_target=True,
            )
            train_frames.append(df)
        except Exception as e:
            print(f"  Warning: Could not load training data for {y}: {e}")
            continue

    if not train_frames:
        return None, "No training data available"

    train_ds = pd.concat(train_frames, ignore_index=True)
    train_ds = train_ds[train_ds["team_share_of_pool"].notna()].copy()

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
    predictions_df = predict_auction_share_of_pool(
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
    predictions_df,
    team_id_map: dict,
    expected_points_map: dict,
):
    """Create a lab entry with market predictions."""
    from moneyball.lab.models import Prediction, create_entry_with_predictions

    predictions = []
    for _, row in predictions_df.iterrows():
        team_slug = row["team_slug"]
        predicted_share = row["predicted_auction_share_of_pool"]

        team_id = team_id_map.get(team_slug)
        if not team_id:
            continue

        expected_points = expected_points_map.get(team_slug)
        if expected_points is None:
            logging.warning("No expected points for team %s, using default 10.0", team_slug)
            expected_points = 10.0

        predictions.append(Prediction(
            team_id=team_id,
            predicted_market_share=predicted_share,
            expected_points=expected_points,
        ))

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


def main():
    parser = argparse.ArgumentParser(description="Generate market predictions for a model")
    parser.add_argument("--model-name", help="Lab model name (e.g., ridge-v1)")
    parser.add_argument("--model-id", help="Lab model ID (alternative to --model-name)")
    parser.add_argument("--calcutta-id", help="Process only this specific calcutta (for pipeline worker)")
    parser.add_argument("--excluded-entry", default=os.environ.get("EXCLUDED_ENTRY_NAME"), help="Entry name to exclude from training (default: $EXCLUDED_ENTRY_NAME)")
    parser.add_argument("--years", type=str, help="Comma-separated years to process (default: all)")
    parser.add_argument("--dry-run", action="store_true", help="Show what would be done without writing")
    parser.add_argument("--json-output", action="store_true", help="Output machine-readable JSON result")

    args = parser.parse_args()

    if not args.model_name and not args.model_id:
        parser.error("Either --model-name or --model-id is required")

    # Helper for logging (suppressed when --json-output)
    def log(msg):
        if not args.json_output:
            print(msg)

    from moneyball.lab.models import get_investment_model, get_investment_model_by_id

    # Load model by ID or name
    if args.model_id:
        model = get_investment_model_by_id(args.model_id)
    else:
        model = get_investment_model(args.model_name)

    if not model:
        if args.json_output:
            print(json.dumps({"ok": False, "entries_created": 0, "errors": ["Model not found"]}))
            sys.exit(1)
        else:
            print("Error: Model not found")
            sys.exit(1)

    log(f"Model: {model.name} ({model.kind})")
    log(f"Params: {model.params}")
    log(f"Excluded entry: {args.excluded_entry}")
    log("")

    calcuttas = get_historical_calcuttas()
    log(f"Found {len(calcuttas)} historical calcuttas")

    # Filter by calcutta_id if specified (for pipeline worker)
    if args.calcutta_id:
        calcuttas = [c for c in calcuttas if c.id == args.calcutta_id]
        if not calcuttas:
            if args.json_output:
                print(json.dumps({"ok": False, "entry_id": None, "error": f"Calcutta {args.calcutta_id} not found"}))
                sys.exit(1)
            else:
                print(f"Error: Calcutta {args.calcutta_id} not found")
                sys.exit(1)
        log(f"Processing single calcutta: {calcuttas[0].name}")

    # Filter by years if specified
    if args.years:
        target_years = [int(y.strip()) for y in args.years.split(",")]
        calcuttas = [c for c in calcuttas if c.year in target_years]
        log(f"Filtered to {len(calcuttas)} calcuttas for years: {target_years}")

    log("")

    entries_created = 0
    errors = []
    last_entry_id = None

    for calcutta in calcuttas:
        year = calcutta.year
        log(f"Processing {calcutta.name} ({year})...")

        predictions_df, error = generate_market_predictions(
            model.name, year, args.excluded_entry
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

        if args.dry_run:
            pred_count = len([
                1 for _, row in predictions_df.iterrows()
                if team_id_map.get(row["team_slug"])
            ])
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
                log(f"  Created entry {entry.id} with {len(entry.predictions)} predictions")
                entries_created += 1
                last_entry_id = entry.id
            else:
                log(f"  No entry created (no valid predictions)")

    log("")
    log("Done!")
    log(f"Created {entries_created} entries with predictions")
    log("Run optimize_lab_entries.py to generate optimized bids")

    if args.json_output:
        # For single calcutta mode (pipeline worker), return entry_id
        if args.calcutta_id and entries_created == 1:
            result = {
                "ok": True,
                "entry_id": last_entry_id,
                "errors": errors if errors else [],
            }
        else:
            result = {
                "ok": True,
                "entries_created": entries_created,
                "errors": errors if errors else [],
            }
        print(json.dumps(result))


if __name__ == "__main__":
    main()
