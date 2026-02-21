#!/usr/bin/env python3
"""
Generate market predictions for a model across all historical calcuttas.

This is stage 2 of the lab pipeline:
1. Register model (lab.investment_models)
2. Generate predictions (this script) -> predictions_json
3. Evaluate (lab pipeline worker) -> lab.evaluations

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

# moneyball must be installed: pip install -e . from the data-science directory
from moneyball.db.lab_helpers import get_historical_calcuttas
from moneyball.lab.predictions import process_calcuttas

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s %(levelname)s %(name)s: %(message)s",
)


def parse_args(argv=None) -> argparse.Namespace:
    """Parse command-line arguments for the generate_lab_predictions script."""
    parser = argparse.ArgumentParser(
        description="Generate market predictions for a model",
    )
    parser.add_argument(
        "--model-name",
        help="Lab model name (e.g., ridge-v1)",
    )
    parser.add_argument(
        "--model-id",
        help="Lab model ID (alternative to --model-name)",
    )
    parser.add_argument(
        "--calcutta-id",
        help="Process only this specific calcutta (for pipeline worker)",
    )
    parser.add_argument(
        "--excluded-entry",
        default=os.environ.get("EXCLUDED_ENTRY_NAME"),
        help="Entry name to exclude from training (default: $EXCLUDED_ENTRY_NAME)",
    )
    parser.add_argument(
        "--years",
        type=str,
        help="Comma-separated years to process (default: all)",
    )
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Show what would be done without writing",
    )
    parser.add_argument(
        "--json-output",
        action="store_true",
        help="Output machine-readable JSON result",
    )

    args = parser.parse_args(argv)

    if not args.model_name and not args.model_id:
        parser.error("Either --model-name or --model-id is required")

    return args


def main():
    args = parse_args()

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
            print(json.dumps({"ok": False, "entriesCreated": 0, "errors": ["Model not found"]}))
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
                print(json.dumps({"ok": False, "entryId": None, "error": f"Calcutta {args.calcutta_id} not found"}))
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

    entries_created, last_entry_id, errors = process_calcuttas(
        model=model,
        calcuttas=calcuttas,
        excluded_entry=args.excluded_entry,
        dry_run=args.dry_run,
        log_fn=log,
    )

    log("")
    log("Done!")
    log(f"Created {entries_created} entries with predictions")

    if args.json_output:
        # For single calcutta mode (pipeline worker), return entry_id
        if args.calcutta_id and entries_created == 1:
            result = {
                "ok": True,
                "entryId": last_entry_id,
                "errors": errors if errors else [],
            }
        else:
            result = {
                "ok": True,
                "entriesCreated": entries_created,
                "errors": errors if errors else [],
            }
        print(json.dumps(result))


if __name__ == "__main__":
    main()
