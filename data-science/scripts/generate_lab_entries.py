#!/usr/bin/env python3
"""
Generate lab entries for a model across all historical calcuttas.

This runs the investment model against each historical calcutta and stores
the predicted bids as lab.entries.

Usage:
    python generate_lab_entries.py --model-name ridge-v1

The script will:
1. Find all historical calcuttas
2. For each calcutta, run the model to predict market share
3. Convert predictions to bid points (using 10000 point budget)
4. Store as lab.entries
"""
import argparse
import json
import os
import sys
from pathlib import Path

# Add project root to path
project_root = Path(__file__).resolve().parents[1]
if str(project_root) not in sys.path:
    sys.path.insert(0, str(project_root))


def get_historical_calcuttas():
    """Get all historical calcuttas from the database."""
    from moneyball.db.connection import get_db_connection

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT c.id, c.name, s.year, t.id as tournament_id
                FROM core.calcuttas c
                JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
                JOIN core.seasons s ON s.id = t.season_id AND s.deleted_at IS NULL
                WHERE c.deleted_at IS NULL
                ORDER BY s.year DESC
            """)
            return [
                {"id": str(row[0]), "name": row[1], "year": row[2], "tournament_id": str(row[3])}
                for row in cur.fetchall()
            ]


def get_team_id_map(tournament_id: str):
    """Get mapping from school_slug to team_id for a tournament."""
    from moneyball.db.connection import get_db_connection

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            cur.execute("""
                SELECT s.slug, t.id
                FROM core.teams t
                JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
                WHERE t.tournament_id = %s AND t.deleted_at IS NULL
            """, (tournament_id,))
            return {row[0]: str(row[1]) for row in cur.fetchall()}


def generate_predictions(model_name: str, year: int, excluded_entry_name: str = "Andrew Copp"):
    """Generate predictions for a model against a specific year."""
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

    initialize_default_scoring_rules_for_year(year)

    # Training years
    all_years = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
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
        except Exception:
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

    # Run ridge regression
    predictions_df = predict_auction_share_of_pool(
        train_team_dataset=train_ds,
        predict_team_dataset=predict_ds,
        ridge_alpha=ridge_alpha,
        feature_set=feature_set,
    )

    predictions_df["team_slug"] = predictions_df["team_key"].str.split(":").str[-1]

    return predictions_df, None


def create_entry_for_calcutta(
    model_id: str,
    calcutta_id: str,
    predictions_df,
    team_id_map: dict,
    budget_points: int = 10000,
):
    """Create a lab entry with bids from predictions."""
    from moneyball.lab.models import Bid, create_entry

    bids = []
    for _, row in predictions_df.iterrows():
        team_slug = row["team_slug"]
        predicted_share = row["predicted_auction_share_of_pool"]

        team_id = team_id_map.get(team_slug)
        if not team_id:
            continue

        bid_points = int(round(predicted_share * budget_points))
        if bid_points <= 0:
            continue

        bids.append(Bid(
            team_id=team_id,
            bid_points=bid_points,
            expected_roi=0.0,  # We'll compute this later during evaluation
        ))

    if not bids:
        return None

    entry = create_entry(
        investment_model_id=model_id,
        calcutta_id=calcutta_id,
        bids=bids,
        game_outcome_kind="kenpom",
        game_outcome_params={},
        optimizer_kind="predicted_market_share",
        optimizer_params={"budget_points": budget_points},
        starting_state_key="post_first_four",
    )

    return entry


def main():
    parser = argparse.ArgumentParser(description="Generate lab entries for a model")
    parser.add_argument("--model-name", required=True, help="Lab model name (e.g., ridge-v1)")
    parser.add_argument("--excluded-entry", default="Andrew Copp", help="Entry name to exclude from training")
    parser.add_argument("--budget", type=int, default=10000, help="Budget points for bids")
    parser.add_argument("--years", type=str, help="Comma-separated years to process (default: all)")
    parser.add_argument("--dry-run", action="store_true", help="Show what would be done without writing")

    args = parser.parse_args()

    from moneyball.lab.models import get_investment_model

    model = get_investment_model(args.model_name)
    if not model:
        print(f"Error: Model '{args.model_name}' not found")
        sys.exit(1)

    print(f"Model: {model.name} ({model.kind})")
    print(f"Params: {model.params}")
    print(f"Excluded entry: {args.excluded_entry}")
    print(f"Budget: {args.budget} points")
    print()

    calcuttas = get_historical_calcuttas()
    print(f"Found {len(calcuttas)} historical calcuttas")

    # Filter by years if specified
    if args.years:
        target_years = [int(y.strip()) for y in args.years.split(",")]
        calcuttas = [c for c in calcuttas if c["year"] in target_years]
        print(f"Filtered to {len(calcuttas)} calcuttas for years: {target_years}")

    print()

    for calcutta in calcuttas:
        year = calcutta["year"]
        print(f"Processing {calcutta['name']} ({year})...")

        predictions_df, error = generate_predictions(
            args.model_name, year, args.excluded_entry
        )

        if error:
            print(f"  Skipping: {error}")
            continue

        if predictions_df is None or predictions_df.empty:
            print(f"  Skipping: No predictions generated")
            continue

        team_id_map = get_team_id_map(calcutta["tournament_id"])
        if not team_id_map:
            print(f"  Skipping: No teams found")
            continue

        if args.dry_run:
            bid_count = len([
                1 for _, row in predictions_df.iterrows()
                if team_id_map.get(row["team_slug"])
                and int(round(row["predicted_auction_share_of_pool"] * args.budget)) > 0
            ])
            print(f"  Would create entry with {bid_count} bids")
        else:
            entry = create_entry_for_calcutta(
                model.id,
                calcutta["id"],
                predictions_df,
                team_id_map,
                args.budget,
            )
            if entry:
                print(f"  Created entry {entry.id} with {len(entry.bids)} bids")
            else:
                print(f"  No entry created (no valid bids)")

    print()
    print("Done!")


if __name__ == "__main__":
    main()
