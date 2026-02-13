#!/usr/bin/env python3
"""
Optimize lab entries that have predictions but no bids yet.

This is stage 3 of the lab pipeline:
1. Register model (lab.investment_models)
2. Generate predictions (generate_lab_predictions.py) -> predictions_json
3. Optimize entry (this script) -> bids_json
4. Evaluate (evaluate_lab_entries.py) -> lab.evaluations

Given market predictions (what others will bid) and expected points (from KenPom),
this script runs an optimizer to determine our optimal bid allocation.

Currently supports:
- predicted_market_share: Bid proportionally to predicted market share (baseline)
- edge_weighted: Bid more on teams with positive edge (expected ROI > 1)

Future:
- minlp: Mixed-integer nonlinear programming optimizer

Usage:
    python optimize_lab_entries.py --model-name ridge-v1
    python optimize_lab_entries.py --all-pending
    python optimize_lab_entries.py --entry-id <uuid>
"""
import argparse
import json
import sys
from pathlib import Path
from typing import List

# Add project root to path
project_root = Path(__file__).resolve().parents[1]
if str(project_root) not in sys.path:
    sys.path.insert(0, str(project_root))

from moneyball.lab.models import (
    Bid,
    Entry,
    Prediction,
    get_entries_pending_optimization,
    get_entry,
    get_investment_model,
    get_investment_model_by_id,
    update_entry_with_bids,
)


def optimize_predicted_market_share(
    predictions: List[Prediction],
    budget_points: int = 10000,
) -> List[Bid]:
    """
    Baseline optimizer: bid proportionally to predicted market share.

    This is equivalent to what the old generate_lab_entries.py did.
    No edge exploitation - just matches market expectations.
    """
    bids = []
    for pred in predictions:
        bid_points = int(round(pred.predicted_market_share * budget_points))
        if bid_points <= 0:
            continue

        # ROI = expected_points / our_cost
        # our_cost = predicted_market + our_bid (in a competitive market)
        # For this baseline, assume our bid = predicted market, so ROI = exp_pts / (2 * predicted)
        # But actually for proportional bidding, we just track expected_roi = exp_pts / predicted_cost
        predicted_cost = pred.predicted_market_share * budget_points
        expected_roi = pred.expected_points / predicted_cost if predicted_cost > 0 else 0.0

        bids.append(Bid(
            team_id=pred.team_id,
            bid_points=bid_points,
            expected_roi=expected_roi,
        ))

    return bids


def optimize_edge_weighted(
    predictions: List[Prediction],
    budget_points: int = 10000,
    edge_multiplier: float = 2.0,
) -> List[Bid]:
    """
    Edge-weighted optimizer: overweight teams with positive edge.

    Edge = expected_roi - 1.0 (positive means undervalued by market)
    Teams with positive edge get more allocation, negative edge get less.
    """
    # First compute expected ROI for each team
    team_data = []
    for pred in predictions:
        predicted_cost = pred.predicted_market_share * budget_points
        expected_roi = pred.expected_points / predicted_cost if predicted_cost > 0 else 0.0
        edge = expected_roi - 1.0

        team_data.append({
            "prediction": pred,
            "expected_roi": expected_roi,
            "edge": edge,
            "base_share": pred.predicted_market_share,
        })

    # Adjust shares based on edge
    for td in team_data:
        # Apply edge multiplier: positive edge increases share, negative decreases
        edge_factor = 1.0 + (td["edge"] * edge_multiplier)
        edge_factor = max(0.0, edge_factor)  # Can't go negative
        td["adjusted_share"] = td["base_share"] * edge_factor

    # Normalize to sum to 1.0
    total_adjusted = sum(td["adjusted_share"] for td in team_data)
    if total_adjusted > 0:
        for td in team_data:
            td["final_share"] = td["adjusted_share"] / total_adjusted
    else:
        for td in team_data:
            td["final_share"] = 1.0 / len(team_data)

    # Convert to bids
    bids = []
    for td in team_data:
        bid_points = int(round(td["final_share"] * budget_points))
        if bid_points <= 0:
            continue

        bids.append(Bid(
            team_id=td["prediction"].team_id,
            bid_points=bid_points,
            expected_roi=td["expected_roi"],
        ))

    return bids


def optimize_entry(
    entry: Entry,
    optimizer_kind: str = "predicted_market_share",
    budget_points: int = 10000,
    **optimizer_params,
) -> List[Bid]:
    """
    Run optimizer on an entry's predictions to generate bids.
    """
    if not entry.predictions:
        return []

    if optimizer_kind == "predicted_market_share":
        return optimize_predicted_market_share(entry.predictions, budget_points)
    elif optimizer_kind == "edge_weighted":
        edge_multiplier = optimizer_params.get("edge_multiplier", 2.0)
        return optimize_edge_weighted(entry.predictions, budget_points, edge_multiplier)
    elif optimizer_kind == "minlp":
        # TODO: Implement MINLP optimizer
        # For now, fall back to edge_weighted
        print(f"  Warning: MINLP optimizer not yet implemented, using edge_weighted")
        return optimize_edge_weighted(entry.predictions, budget_points)
    else:
        raise ValueError(f"Unknown optimizer kind: {optimizer_kind}")


def main():
    parser = argparse.ArgumentParser(description="Optimize lab entries with predictions")
    parser.add_argument("--model-name", help="Lab model name (e.g., ridge-v1)")
    parser.add_argument("--model-id", help="Lab model ID (alternative to --model-name)")
    parser.add_argument("--entry-id", help="Specific entry ID to optimize")
    parser.add_argument("--all-pending", action="store_true", help="Optimize all pending entries")
    parser.add_argument("--optimizer", default="predicted_market_share",
                       choices=["predicted_market_share", "edge_weighted", "minlp"],
                       help="Optimizer to use")
    parser.add_argument("--budget", type=int, default=10000, help="Budget points for bids")
    parser.add_argument("--edge-multiplier", type=float, default=2.0,
                       help="Edge multiplier for edge_weighted optimizer")
    parser.add_argument("--dry-run", action="store_true", help="Show what would be done")
    parser.add_argument("--json-output", action="store_true", help="Output JSON result")

    args = parser.parse_args()

    if not args.model_name and not args.model_id and not args.entry_id and not args.all_pending:
        parser.error("One of --model-name, --model-id, --entry-id, or --all-pending is required")

    def log(msg):
        if not args.json_output:
            print(msg)

    # Get entries to optimize
    entries_to_optimize = []

    if args.entry_id:
        entry = get_entry(args.entry_id)
        if not entry:
            if args.json_output:
                print(json.dumps({"ok": False, "entries_optimized": 0, "errors": ["Entry not found"]}))
            else:
                print(f"Error: Entry {args.entry_id} not found")
            sys.exit(1)
        entries_to_optimize = [entry]
    elif args.all_pending:
        entries_to_optimize = get_entries_pending_optimization()
        log(f"Found {len(entries_to_optimize)} entries pending optimization")
    else:
        # Get by model
        if args.model_id:
            model = get_investment_model_by_id(args.model_id)
        else:
            model = get_investment_model(args.model_name)

        if not model:
            if args.json_output:
                print(json.dumps({"ok": False, "entries_optimized": 0, "errors": ["Model not found"]}))
            else:
                print("Error: Model not found")
            sys.exit(1)

        log(f"Model: {model.name} ({model.kind})")
        entries_to_optimize = get_entries_pending_optimization(model.id)
        log(f"Found {len(entries_to_optimize)} entries pending optimization")

    if not entries_to_optimize:
        log("No entries to optimize")
        if args.json_output:
            print(json.dumps({"ok": True, "entries_optimized": 0, "errors": []}))
        return

    log(f"Optimizer: {args.optimizer}")
    log(f"Budget: {args.budget} points")
    log("")

    entries_optimized = 0
    errors = []

    for entry in entries_to_optimize:
        log(f"Optimizing entry {entry.id}...")

        if not entry.predictions:
            log(f"  Skipping: No predictions")
            errors.append(f"{entry.id}: No predictions")
            continue

        try:
            bids = optimize_entry(
                entry,
                optimizer_kind=args.optimizer,
                budget_points=args.budget,
                edge_multiplier=args.edge_multiplier,
            )
        except Exception as e:
            log(f"  Error: {e}")
            errors.append(f"{entry.id}: {e}")
            continue

        if not bids:
            log(f"  Skipping: No bids generated")
            continue

        total_bid = sum(b.bid_points for b in bids)
        log(f"  Generated {len(bids)} bids totaling {total_bid} points")

        if args.dry_run:
            log(f"  (dry run - not saving)")
        else:
            update_entry_with_bids(
                entry.id,
                bids,
                optimizer_kind=args.optimizer,
                optimizer_params={
                    "budget_points": args.budget,
                    "edge_multiplier": args.edge_multiplier if args.optimizer == "edge_weighted" else None,
                },
            )
            log(f"  Saved")
            entries_optimized += 1

    log("")
    log(f"Done! Optimized {entries_optimized} entries")

    if args.json_output:
        print(json.dumps({
            "ok": True,
            "entries_optimized": entries_optimized,
            "errors": errors if errors else [],
        }))


if __name__ == "__main__":
    main()
