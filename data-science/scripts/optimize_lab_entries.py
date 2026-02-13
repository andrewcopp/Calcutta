#!/usr/bin/env python3
"""
Optimize lab entries that have predictions but no bids yet.

NOTE: This script is deprecated for production use. The lab pipeline worker
now calls the Go DP allocator directly, which is provably optimal and faster.
This script remains for manual testing and research experiments only.

This is stage 3 of the lab pipeline:
1. Register model (lab.investment_models)
2. Generate predictions (generate_lab_predictions.py) -> predictions_json
3. Optimize entry (Go DP allocator via lab_pipeline_worker) -> bids_json
4. Evaluate (evaluate_lab_entries.py) -> lab.evaluations

Currently supports (for research only):
- predicted_market_share: Bid proportionally to predicted market share (baseline)
- edge_weighted: Bid more on teams with positive edge (expected ROI > 1)

REMOVED (2026-02-13):
- minlp: Removed because the Go DP allocator is provably optimal and the MINLP
  implementation had silent fallbacks that produced invalid results.

Usage:
    python optimize_lab_entries.py --model-name ridge-v1
    python optimize_lab_entries.py --all-pending
    python optimize_lab_entries.py --entry-id <uuid>
"""
import argparse
import json
import os
import sys
from pathlib import Path
from typing import List, Optional

import pandas as pd

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
    budget_points: int = 100,
    max_per_team: int = 50,
) -> List[Bid]:
    """
    Baseline optimizer: bid proportionally to predicted market share.

    This is equivalent to what the old generate_lab_entries.py did.
    No edge exploitation - just matches market expectations.
    """
    bids = []
    for pred in predictions:
        bid_points = int(round(pred.predicted_market_share * budget_points))
        # Enforce max per team constraint
        bid_points = min(bid_points, max_per_team)
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

    # Normalize to budget if max_per_team caused under-allocation
    total = sum(b.bid_points for b in bids)
    if total < budget_points and total > 0:
        scale = budget_points / total
        for b in bids:
            b.bid_points = min(int(round(b.bid_points * scale)), max_per_team)

    return bids


def optimize_edge_weighted(
    predictions: List[Prediction],
    budget_points: int = 100,
    max_per_team: int = 50,
    edge_multiplier: float = 2.0,
) -> List[Bid]:
    """
    Edge-weighted optimizer: overweight teams with positive edge.

    Edge = expected_roi - 1.0 (positive means undervalued by market)
    Teams with positive edge get more allocation, negative edge get less.
    Respects max_per_team constraint and redistributes excess.
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

    # Convert to bids with max constraint - iteratively redistribute overflow
    max_share = max_per_team / budget_points
    for _ in range(10):  # Max iterations to redistribute
        overflow = 0.0
        uncapped_share = 0.0
        for td in team_data:
            if td["final_share"] > max_share:
                overflow += td["final_share"] - max_share
                td["final_share"] = max_share
            else:
                uncapped_share += td["final_share"]

        if overflow == 0 or uncapped_share == 0:
            break

        # Redistribute overflow proportionally to uncapped teams
        for td in team_data:
            if td["final_share"] < max_share:
                td["final_share"] += overflow * (td["final_share"] / uncapped_share)

    # Convert to bids
    bids = []
    for td in team_data:
        bid_points = int(round(td["final_share"] * budget_points))
        bid_points = min(bid_points, max_per_team)  # Final safety check
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
    budget_points: int = 100,
    max_per_team: int = 50,
    **optimizer_params,
) -> List[Bid]:
    """
    Run optimizer on an entry's predictions to generate bids.
    """
    if not entry.predictions:
        return []

    if optimizer_kind == "predicted_market_share":
        return optimize_predicted_market_share(entry.predictions, budget_points, max_per_team)
    elif optimizer_kind == "edge_weighted":
        edge_multiplier = optimizer_params.get("edge_multiplier", 2.0)
        return optimize_edge_weighted(entry.predictions, budget_points, max_per_team, edge_multiplier)
    elif optimizer_kind == "minlp":
        raise ValueError(
            "MINLP optimizer was removed. Use the Go DP allocator via lab_pipeline_worker "
            "or use 'predicted_market_share' or 'edge_weighted' for research."
        )
    else:
        raise ValueError(f"Unknown optimizer kind: {optimizer_kind}")


def get_entry_constraints(entry: Entry, cli_args) -> dict:
    """
    Get optimization constraints, preferring entry's stored params over CLI defaults.
    CLI args override entry params when explicitly provided.
    """
    stored = entry.optimizer_params or {}

    return {
        "budget_points": cli_args.budget if cli_args.budget is not None else stored.get("budget_points", 100),
        "max_per_team": cli_args.max_per_team if cli_args.max_per_team is not None else stored.get("max_per_team", 50),
        "min_teams": cli_args.min_teams if cli_args.min_teams is not None else stored.get("min_teams", 3),
        "max_teams": cli_args.max_teams if cli_args.max_teams is not None else stored.get("max_teams", 10),
        "min_bid": cli_args.min_bid if cli_args.min_bid is not None else stored.get("min_bid", 1),
        "edge_multiplier": cli_args.edge_multiplier if cli_args.edge_multiplier is not None else stored.get("edge_multiplier", 2.0),
        "estimated_participants": stored.get("estimated_participants"),
        "excluded_entry_name": stored.get("excluded_entry_name"),
    }


def main():
    parser = argparse.ArgumentParser(
        description="Optimize lab entries with predictions. "
                    "Constraints default to values stored in entry's optimizer_params (from calcutta rules)."
    )
    parser.add_argument("--model-name", help="Lab model name (e.g., ridge-v1)")
    parser.add_argument("--model-id", help="Lab model ID (alternative to --model-name)")
    parser.add_argument("--entry-id", help="Specific entry ID to optimize")
    parser.add_argument("--all-pending", action="store_true", help="Optimize all pending entries")
    parser.add_argument("--optimizer", default="predicted_market_share",
                       choices=["predicted_market_share", "edge_weighted"],
                       help="Optimizer to use (minlp removed - use Go DP allocator)")
    # These args are optional - if not provided, use entry's stored params
    parser.add_argument("--budget", type=int, default=None,
                       help="Override budget points (default: from entry/calcutta rules)")
    parser.add_argument("--max-per-team", type=int, default=None,
                       help="Override maximum bid per team (default: from entry/calcutta rules)")
    parser.add_argument("--edge-multiplier", type=float, default=None,
                       help="Edge multiplier for edge_weighted optimizer (default: 2.0)")
    parser.add_argument("--min-teams", type=int, default=None,
                       help="Override minimum teams for minlp optimizer (default: from entry/calcutta rules)")
    parser.add_argument("--max-teams", type=int, default=None,
                       help="Override maximum teams for minlp optimizer (default: from entry/calcutta rules)")
    parser.add_argument("--min-bid", type=int, default=None,
                       help="Override minimum bid per team for minlp optimizer (default: 1)")
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
    log(f"Note: Constraints default to values from entry's optimizer_params (calcutta rules)")
    if args.budget or args.max_per_team or args.min_teams or args.max_teams or args.min_bid:
        log(f"CLI overrides: budget={args.budget}, max_per_team={args.max_per_team}, "
            f"min_teams={args.min_teams}, max_teams={args.max_teams}, min_bid={args.min_bid}")
    log("")

    entries_optimized = 0
    errors = []

    for entry in entries_to_optimize:
        log(f"Optimizing entry {entry.id}...")

        if not entry.predictions:
            log(f"  Skipping: No predictions")
            errors.append(f"{entry.id}: No predictions")
            continue

        # Get constraints from entry's stored params, with CLI overrides
        constraints = get_entry_constraints(entry, args)
        log(f"  Constraints: budget={constraints['budget_points']}, max_per_team={constraints['max_per_team']}, "
            f"min_teams={constraints['min_teams']}, max_teams={constraints['max_teams']}")

        try:
            bids = optimize_entry(
                entry,
                optimizer_kind=args.optimizer,
                budget_points=constraints["budget_points"],
                max_per_team=constraints["max_per_team"],
                edge_multiplier=constraints["edge_multiplier"],
                min_teams=constraints["min_teams"],
                max_teams=constraints["max_teams"],
                min_bid=constraints["min_bid"],
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
            # Build optimizer_params, preserving stored values and adding optimizer-specific ones
            optimizer_params = {
                "budget_points": constraints["budget_points"],
                "max_per_team": constraints["max_per_team"],
                "min_teams": constraints["min_teams"],
                "max_teams": constraints["max_teams"],
            }
            # Preserve prediction inputs from original entry
            if constraints.get("estimated_participants"):
                optimizer_params["estimated_participants"] = constraints["estimated_participants"]
            if constraints.get("excluded_entry_name"):
                optimizer_params["excluded_entry_name"] = constraints["excluded_entry_name"]
            # Add optimizer-specific params
            if args.optimizer == "edge_weighted":
                optimizer_params["edge_multiplier"] = constraints["edge_multiplier"]
            elif args.optimizer == "minlp":
                optimizer_params["min_bid"] = constraints["min_bid"]

            update_entry_with_bids(
                entry.id,
                bids,
                optimizer_kind=args.optimizer,
                optimizer_params=optimizer_params,
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
