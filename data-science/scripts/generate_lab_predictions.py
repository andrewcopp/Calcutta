#!/usr/bin/env python3
"""
Generate market predictions for a model across all historical calcuttas.

This is stage 2 of the lab pipeline:
1. Register model (lab.investment_models)
2. Generate predictions (this script) -> predictions_json
3. Optimize entry (optimize_lab_entries.py) -> bids_json
4. Evaluate (evaluate_lab_entries.py) -> lab.evaluations

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


def get_expected_points_map(calcutta_id: str):
    """
    Get expected tournament points for each team from simulation data.

    Uses pre-computed simulations from derived.simulated_teams and the calcutta's
    scoring rules via core.calcutta_points_for_progress() to calculate:
        expected_points = AVG(calcutta_points_for_progress(wins, byes))

    Optimized to pre-aggregate win distributions first, reducing function calls
    from millions to ~1000.
    """
    from moneyball.db.connection import get_db_connection

    with get_db_connection() as conn:
        with conn.cursor() as cur:
            # Optimized: pre-aggregate win distributions, then compute points once per bucket
            # This reduces function calls from millions to (teams * win_levels * bye_levels)
            cur.execute("""
                WITH calcutta_ctx AS (
                    SELECT c.id AS calcutta_id, t.id AS tournament_id
                    FROM core.calcuttas c
                    JOIN core.tournaments t ON t.id = c.tournament_id AND t.deleted_at IS NULL
                    WHERE c.id = %s AND c.deleted_at IS NULL
                ),
                win_distribution AS (
                    -- Aggregate simulation results into (team, wins, byes, count) buckets
                    SELECT
                        st.team_id,
                        st.wins,
                        st.byes,
                        COUNT(*)::float AS sim_count
                    FROM derived.simulated_teams st
                    WHERE st.tournament_id = (SELECT tournament_id FROM calcutta_ctx)
                    GROUP BY st.team_id, st.wins, st.byes
                ),
                team_totals AS (
                    SELECT team_id, SUM(sim_count) AS total_sims
                    FROM win_distribution
                    GROUP BY team_id
                ),
                team_expected AS (
                    -- Compute weighted average: SUM(count * points) / total_count
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

            # If no simulation data, fall back to seed-based estimates
            if not result:
                seed_expected_points = {
                    1: 80.0, 2: 55.0, 3: 42.0, 4: 35.0,
                    5: 28.0, 6: 23.0, 7: 19.0, 8: 16.0,
                    9: 14.0, 10: 12.0, 11: 10.0, 12: 9.0,
                    13: 5.0, 14: 4.0, 15: 2.5, 16: 1.0,
                }
                cur.execute("""
                    SELECT s.slug, t.seed
                    FROM core.teams t
                    JOIN core.schools s ON s.id = t.school_id AND s.deleted_at IS NULL
                    JOIN core.calcuttas c ON c.tournament_id = t.tournament_id AND c.deleted_at IS NULL
                    WHERE c.id = %s AND t.deleted_at IS NULL
                """, (calcutta_id,))
                result = {row[0]: seed_expected_points.get(row[1], 10.0) for row in cur.fetchall()}

            return result


def generate_market_predictions(model_name: str, year: int, excluded_entry_name: str = "Andrew Copp"):
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

        expected_points = expected_points_map.get(team_slug, 10.0)

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
    parser.add_argument("--excluded-entry", default="Andrew Copp", help="Entry name to exclude from training")
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
        calcuttas = [c for c in calcuttas if c["id"] == args.calcutta_id]
        if not calcuttas:
            if args.json_output:
                print(json.dumps({"ok": False, "entry_id": None, "error": f"Calcutta {args.calcutta_id} not found"}))
                sys.exit(1)
            else:
                print(f"Error: Calcutta {args.calcutta_id} not found")
                sys.exit(1)
        log(f"Processing single calcutta: {calcuttas[0]['name']}")

    # Filter by years if specified
    if args.years:
        target_years = [int(y.strip()) for y in args.years.split(",")]
        calcuttas = [c for c in calcuttas if c["year"] in target_years]
        log(f"Filtered to {len(calcuttas)} calcuttas for years: {target_years}")

    log("")

    entries_created = 0
    errors = []
    last_entry_id = None

    for calcutta in calcuttas:
        year = calcutta["year"]
        log(f"Processing {calcutta['name']} ({year})...")

        predictions_df, error = generate_market_predictions(
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

        # Get expected points for each team from simulation data
        expected_points_map = get_expected_points_map(calcutta["id"])

        if args.dry_run:
            pred_count = len([
                1 for _, row in predictions_df.iterrows()
                if team_id_map.get(row["team_slug"])
            ])
            log(f"  Would create entry with {pred_count} predictions")
        else:
            entry = create_predictions_for_calcutta(
                model.id,
                calcutta["id"],
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
