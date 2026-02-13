#!/usr/bin/env python3
"""
Generate investment predictions for a lab model.

This script runs the investment model (e.g., ridge regression) against
historical calcutta data and outputs predictions comparing:
- Naive: EV-proportional allocation (what everyone would bid if rational)
- Predicted: Model's predicted market share allocation

Usage:
    python get_lab_model_predictions.py --model-name ridge-v1 --year 2024

Output: JSON with predictions for each team
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


def get_predictions(model_name: str, year: int, excluded_entry_name: str = "Andrew Copp"):
    """
    Generate predictions for a model against a specific year's calcutta.

    Returns a dict with:
    - model_name: The lab model name
    - year: Tournament year
    - predictions: List of {team_key, school_name, seed, naive, predicted, delta}
    """
    from moneyball.db.connection import get_db_connection
    from moneyball.db.readers import (
        initialize_default_scoring_rules_for_year,
        read_ridge_team_dataset_for_year,
    )
    from moneyball.models.predicted_auction_share_of_pool import (
        predict_auction_share_of_pool,
    )
    from moneyball.lab.models import get_investment_model
    import pandas as pd

    # Verify model exists
    model = get_investment_model(model_name)
    if not model:
        return {"error": f"Model '{model_name}' not found"}

    # Get model params
    ridge_alpha = model.params.get("alpha", 1.0)
    feature_set = model.params.get("feature_set", "optimal")

    # Initialize scoring rules
    initialize_default_scoring_rules_for_year(year)

    # Determine training years (all except target)
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
        except Exception as e:
            print(f"Skipping training year {y}: {e}", file=sys.stderr)
            continue

    if not train_frames:
        return {"error": "No training data available"}

    train_ds = pd.concat(train_frames, ignore_index=True)
    train_ds = train_ds[train_ds["team_share_of_pool"].notna()].copy()

    if train_ds.empty:
        return {"error": "No valid training rows"}

    # Load prediction data
    try:
        predict_ds = read_ridge_team_dataset_for_year(
            year,
            exclude_entry_names=None,
            include_target=False,
        )
    except Exception as e:
        return {"error": f"Cannot load data for year {year}: {e}"}

    # Run ridge regression
    predictions_df = predict_auction_share_of_pool(
        train_team_dataset=train_ds,
        predict_team_dataset=predict_ds,
        ridge_alpha=ridge_alpha,
        feature_set=feature_set,
    )

    # Calculate naive (EV-proportional) allocation
    # Naive assumes everyone bids proportional to expected tournament points
    # Use seed-based expected points as a proxy for expected value
    seed_expected_points = {
        1: 12, 2: 9, 3: 7, 4: 5, 5: 4, 6: 3, 7: 2, 8: 2,
        9: 1, 10: 1, 11: 1, 12: 1, 13: 0.5, 14: 0.3, 15: 0.2, 16: 0.1
    }
    predict_ds["expected_points"] = predict_ds["seed"].map(seed_expected_points)
    total_ev = predict_ds["expected_points"].sum()
    predict_ds["naive_share"] = predict_ds["expected_points"] / total_ev

    # Merge predictions with team info
    predictions_df["team_slug"] = predictions_df["team_key"].str.split(":").str[-1]

    # Build result
    results = []
    for _, row in predictions_df.iterrows():
        team_slug = row["team_slug"]
        predicted = row["predicted_auction_share_of_pool"]

        # Find matching team in predict_ds
        team_row = predict_ds[predict_ds["team_key"].str.endswith(f":{team_slug}")]
        if team_row.empty:
            continue

        team_row = team_row.iloc[0]
        naive = team_row.get("naive_share", 0)
        seed = int(team_row.get("seed", 0))
        school_name = team_row.get("school_name", team_slug.replace("-", " ").title())

        # Delta: how much predicted differs from naive (positive = overvalued by market)
        delta = ((predicted - naive) / naive * 100) if naive > 0 else 0

        results.append({
            "team_slug": team_slug,
            "school_name": school_name,
            "seed": seed,
            "naive": round(naive * 100, 2),  # as percentage
            "predicted": round(predicted * 100, 2),  # as percentage
            "delta": round(delta, 1),  # as percentage points
        })

    # Sort by seed, then by predicted (descending)
    results.sort(key=lambda x: (x["seed"], -x["predicted"]))

    return {
        "model_name": model_name,
        "model_kind": model.kind,
        "year": year,
        "excluded_entry_name": excluded_entry_name,
        "ridge_alpha": ridge_alpha,
        "feature_set": feature_set,
        "train_years": train_years,
        "predictions": results,
    }


def main():
    parser = argparse.ArgumentParser(description="Generate lab model predictions")
    parser.add_argument("--model-name", required=True, help="Lab model name (e.g., ridge-v1)")
    parser.add_argument("--year", type=int, required=True, help="Tournament year to predict")
    parser.add_argument("--excluded-entry", default="Andrew Copp", help="Entry name to exclude from training")
    parser.add_argument("--format", choices=["json", "table"], default="json", help="Output format")

    args = parser.parse_args()

    result = get_predictions(args.model_name, args.year, args.excluded_entry)

    if args.format == "json":
        print(json.dumps(result, indent=2))
    else:
        if "error" in result:
            print(f"Error: {result['error']}")
            sys.exit(1)

        print(f"\nModel: {result['model_name']} ({result['model_kind']})")
        print(f"Year: {result['year']}")
        print(f"Excluded: {result['excluded_entry_name']}")
        print(f"Training years: {result['train_years']}")
        print(f"\n{'Team':<25} {'Seed':>4} {'Naive':>8} {'Predicted':>10} {'Delta':>8}")
        print("-" * 60)
        for p in result["predictions"]:
            print(f"{p['school_name']:<25} {p['seed']:>4} {p['naive']:>7.2f}% {p['predicted']:>9.2f}% {p['delta']:>+7.1f}%")


if __name__ == "__main__":
    main()
