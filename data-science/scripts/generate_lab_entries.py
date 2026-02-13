#!/usr/bin/env python3
"""
Generate lab entries for a model across all historical calcuttas.

This is a convenience wrapper that runs the full pipeline:
1. Generate predictions (generate_lab_predictions.py)
2. Optimize entries (optimize_lab_entries.py)

For more control, run the scripts separately:
    python generate_lab_predictions.py --model-name ridge-v1
    python optimize_lab_entries.py --model-name ridge-v1 --optimizer edge_weighted

Usage:
    python generate_lab_entries.py --model-name ridge-v1
    python generate_lab_entries.py --model-id <uuid> --optimizer edge_weighted
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
    """Get all historical calcuttas from the database with their rules and entry counts."""
    from moneyball.db.connection import get_db_connection

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
                {
                    "id": str(row[0]),
                    "name": row[1],
                    "year": row[2],
                    "tournament_id": str(row[3]),
                    "budget_points": row[4],
                    "min_teams": row[5],
                    "max_teams": row[6],
                    "max_bid": row[7],
                    "entry_count": row[8],
                }
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


def get_expected_points_map(year: int):
    """Get expected tournament points for each team in a year."""
    from moneyball.db.connection import get_db_connection

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Simple fallback: use seed-based expected points
            cur.execute("""
                SELECT
                    s.slug as team_slug,
                    CASE t.seed
                        WHEN 1 THEN 25.0
                        WHEN 2 THEN 18.0
                        WHEN 3 THEN 14.0
                        WHEN 4 THEN 12.0
                        WHEN 5 THEN 9.0
                        WHEN 6 THEN 8.0
                        WHEN 7 THEN 7.0
                        WHEN 8 THEN 6.0
                        WHEN 9 THEN 5.5
                        WHEN 10 THEN 5.0
                        WHEN 11 THEN 4.5
                        WHEN 12 THEN 4.0
                        WHEN 13 THEN 2.5
                        WHEN 14 THEN 2.0
                        WHEN 15 THEN 1.5
                        WHEN 16 THEN 1.0
                        ELSE 5.0
                    END as expected_points
                FROM core.teams t
                JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
                JOIN core.tournaments tour ON tour.id = t.tournament_id AND tour.deleted_at IS NULL
                JOIN core.seasons sea ON sea.id = tour.season_id AND sea.deleted_at IS NULL
                WHERE sea.year = %s AND t.deleted_at IS NULL
            """, (year,))
            return {row[0]: float(row[1]) for row in cur.fetchall()}


def generate_predictions(model_name: str, year: int, excluded_entry_name: str = "Andrew Copp"):
    """Generate market predictions for a model against a specific year."""
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
    expected_points_map: dict,
    budget_points: int,
    max_per_team: int,
    min_teams: int,
    max_teams: int,
    optimizer_kind: str,
    estimated_participants: int,
    excluded_entry_name: str,
):
    """Create a lab entry with predictions and optimized bids."""
    from moneyball.lab.models import Bid, Prediction, create_entry

    # Build predictions
    predictions = []
    for _, row in predictions_df.iterrows():
        team_slug = row["team_slug"]
        predicted_share = row["predicted_auction_share_of_pool"]

        team_id = team_id_map.get(team_slug)
        if not team_id:
            continue

        expected_points = expected_points_map.get(team_slug, 10.0)

        predictions.append(Prediction(
            team_id=team_id,
            predicted_market_share=predicted_share,
            expected_points=expected_points,
        ))

    if not predictions:
        return None

    # Generate optimized bids based on predictions
    bids = []
    for pred in predictions:
        if optimizer_kind == "predicted_market_share":
            # Baseline: bid proportionally to predicted market share
            bid_points = int(round(pred.predicted_market_share * budget_points))
        elif optimizer_kind == "edge_weighted":
            # Edge-weighted: adjust by expected ROI
            predicted_cost = pred.predicted_market_share * budget_points
            expected_roi = pred.expected_points / predicted_cost if predicted_cost > 0 else 0.0
            edge = expected_roi - 1.0
            edge_factor = max(0.0, 1.0 + edge * 2.0)
            bid_points = int(round(pred.predicted_market_share * edge_factor * budget_points))
        else:
            bid_points = int(round(pred.predicted_market_share * budget_points))

        # Enforce max per team constraint
        bid_points = min(bid_points, max_per_team)

        if bid_points <= 0:
            continue

        predicted_cost = pred.predicted_market_share * budget_points
        expected_roi = pred.expected_points / predicted_cost if predicted_cost > 0 else 0.0

        bids.append(Bid(
            team_id=pred.team_id,
            bid_points=bid_points,
            expected_roi=expected_roi,
        ))

    # Normalize bids to budget if edge_weighted caused drift, respecting max constraint
    if optimizer_kind == "edge_weighted" and bids:
        total = sum(b.bid_points for b in bids)
        if total > 0:
            scale = budget_points / total
            for b in bids:
                b.bid_points = min(int(round(b.bid_points * scale)), max_per_team)

    if not bids:
        return None

    entry = create_entry(
        investment_model_id=model_id,
        calcutta_id=calcutta_id,
        bids=bids,
        predictions=predictions,
        game_outcome_kind="kenpom",
        game_outcome_params={},
        optimizer_kind=optimizer_kind,
        optimizer_params={
            "budget_points": budget_points,
            "max_per_team": max_per_team,
            "min_teams": min_teams,
            "max_teams": max_teams,
            "estimated_participants": estimated_participants,
            "excluded_entry_name": excluded_entry_name,
        },
        starting_state_key="post_first_four",
    )

    return entry


def main():
    # Read defaults from environment variables
    default_excluded_entry = os.environ.get("EXCLUDED_ENTRY_NAME", "Andrew Copp")
    default_estimated_participants = int(os.environ.get("DEFAULT_ESTIMATED_PARTICIPANTS", "42"))

    parser = argparse.ArgumentParser(description="Generate lab entries for a model")
    parser.add_argument("--model-name", help="Lab model name (e.g., ridge-v1)")
    parser.add_argument("--model-id", help="Lab model ID (alternative to --model-name)")
    parser.add_argument("--excluded-entry", default=default_excluded_entry,
                       help=f"Entry name to exclude from training (env: EXCLUDED_ENTRY_NAME, default: {default_excluded_entry})")
    parser.add_argument("--optimizer", default="predicted_market_share",
                       choices=["predicted_market_share", "edge_weighted", "minlp"],
                       help="Optimizer to use for bid generation")
    parser.add_argument("--estimated-participants-override", type=int,
                       help=f"Override estimated participants (env: DEFAULT_ESTIMATED_PARTICIPANTS={default_estimated_participants}). "
                            "If not set, uses actual entry count from calcutta or env default.")
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
    log(f"Optimizer: {args.optimizer}")
    log(f"Note: Budget, min/max teams, and max bid come from each calcutta's rules")
    log(f"Note: Estimated participants uses actual entry count (fallback: {default_estimated_participants})")
    log("")

    calcuttas = get_historical_calcuttas()
    log(f"Found {len(calcuttas)} historical calcuttas")

    # Filter by years if specified
    if args.years:
        target_years = [int(y.strip()) for y in args.years.split(",")]
        calcuttas = [c for c in calcuttas if c["year"] in target_years]
        log(f"Filtered to {len(calcuttas)} calcuttas for years: {target_years}")

    log("")

    entries_created = 0
    errors = []

    for calcutta in calcuttas:
        year = calcutta["year"]
        log(f"Processing {calcutta['name']} ({year})...")

        # Get calcutta rules from database
        budget_points = calcutta["budget_points"]
        min_teams = calcutta["min_teams"]
        max_teams = calcutta["max_teams"]
        max_per_team = calcutta["max_bid"]

        # Use actual entry count, or override, or env default
        if args.estimated_participants_override:
            estimated_participants = args.estimated_participants_override
        elif calcutta["entry_count"] > 0:
            estimated_participants = calcutta["entry_count"]
        else:
            estimated_participants = default_estimated_participants

        log(f"  Rules: budget={budget_points}, min_teams={min_teams}, max_teams={max_teams}, max_bid={max_per_team}")
        log(f"  Estimated participants: {estimated_participants} (actual entries: {calcutta['entry_count']})")

        predictions_df, error = generate_predictions(
            model.name, year, args.excluded_entry
        )

        if error:
            log(f"  Skipping: {error}")
            errors.append(f"{calcutta['name']}: {error}")
            continue

        if predictions_df is None or predictions_df.empty:
            log(f"  Skipping: No predictions generated")
            continue

        team_id_map = get_team_id_map(calcutta["tournament_id"])
        if not team_id_map:
            log(f"  Skipping: No teams found")
            continue

        expected_points_map = get_expected_points_map(year)

        if args.dry_run:
            pred_count = len([
                1 for _, row in predictions_df.iterrows()
                if team_id_map.get(row["team_slug"])
            ])
            log(f"  Would create entry with {pred_count} predictions/bids")
        else:
            entry = create_entry_for_calcutta(
                model.id,
                calcutta["id"],
                predictions_df,
                team_id_map,
                expected_points_map,
                budget_points,
                max_per_team,
                min_teams,
                max_teams,
                args.optimizer,
                estimated_participants,
                args.excluded_entry,
            )
            if entry:
                log(f"  Created entry {entry.id} with {len(entry.predictions)} predictions and {len(entry.bids)} bids")
                entries_created += 1
            else:
                log(f"  No entry created (no valid predictions)")

    log("")
    log("Done!")

    if args.json_output:
        result = {
            "ok": True,
            "entries_created": entries_created,
            "errors": errors if errors else [],
        }
        print(json.dumps(result))


if __name__ == "__main__":
    main()
