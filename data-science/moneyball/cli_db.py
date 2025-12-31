"""
DB-first CLI for moneyball pipeline.

This CLI runs the pipeline using PostgreSQL as the primary data source,
eliminating parquet file dependencies.
"""
from __future__ import annotations

import argparse
import sys

from moneyball.pipeline.runner_db import (
    stage_predicted_game_outcomes,
    stage_simulate_tournaments,
    stage_recommended_entry_bids,
    run_full_pipeline,
)


def main() -> int:
    parser = argparse.ArgumentParser(
        prog="moneyball-db",
        description="Run moneyball pipeline with database-first architecture"
    )

    sub = parser.add_subparsers(dest="cmd", required=True)

    # Predicted game outcomes
    p_go = sub.add_parser(
        "predicted-game-outcomes",
        help="Generate predicted game outcomes"
    )
    p_go.add_argument("year", type=int, help="Tournament year (e.g., 2025)")
    p_go.add_argument("--calcutta-id", dest="calcutta_id", default=None)
    p_go.add_argument(
        "--kenpom-scale",
        dest="kenpom_scale",
        type=float,
        default=10.0,
    )
    p_go.add_argument("--n-sims", dest="n_sims", type=int, default=5000)
    p_go.add_argument("--seed", dest="seed", type=int, default=42)
    p_go.add_argument(
        "--model-version",
        dest="model_version",
        default="kenpom-v1"
    )

    # Simulate tournaments
    p_st = sub.add_parser(
        "simulate-tournaments",
        help="Simulate tournament outcomes"
    )
    p_st.add_argument("year", type=int, help="Tournament year")
    p_st.add_argument("--n-sims", dest="n_sims", type=int, default=5000)
    p_st.add_argument("--seed", dest="seed", type=int, default=42)
    p_st.add_argument("--run-id", dest="run_id", default=None)

    # Recommended entry bids
    p_reb = sub.add_parser(
        "recommended-entry-bids",
        help="Generate recommended entry bids"
    )
    p_reb.add_argument("year", type=int, help="Tournament year")
    p_reb.add_argument("--strategy", dest="strategy", default="greedy")
    p_reb.add_argument(
        "--budget-points",
        dest="budget_points",
        type=int,
        default=100
    )
    p_reb.add_argument("--min-teams", dest="min_teams", type=int, default=3)
    p_reb.add_argument("--max-teams", dest="max_teams", type=int, default=10)
    p_reb.add_argument("--min-bid", dest="min_bid", type=int, default=1)
    p_reb.add_argument("--max-bid", dest="max_bid", type=int, default=50)
    p_reb.add_argument("--run-id", dest="run_id", default=None)
    p_reb.add_argument("--calcutta-id", dest="calcutta_id", default=None)

    # Full pipeline
    p_full = sub.add_parser(
        "full-pipeline",
        help="Run the complete pipeline"
    )
    p_full.add_argument("year", type=int, help="Tournament year")
    p_full.add_argument("--n-sims", dest="n_sims", type=int, default=5000)
    p_full.add_argument("--seed", dest="seed", type=int, default=42)
    p_full.add_argument("--strategy", dest="strategy", default="greedy")
    p_full.add_argument(
        "--kenpom-scale",
        dest="kenpom_scale",
        type=float,
        default=10.0
    )
    p_full.add_argument("--calcutta-id", dest="calcutta_id", default=None)

    args = parser.parse_args()

    try:
        if args.cmd == "predicted-game-outcomes":
            result = stage_predicted_game_outcomes(
                year=args.year,
                calcutta_id=args.calcutta_id,
                kenpom_scale=args.kenpom_scale,
                n_sims=args.n_sims,
                seed=args.seed,
                model_version=args.model_version,
            )
            print(f"\nResult: {result}")

        elif args.cmd == "simulate-tournaments":
            result = stage_simulate_tournaments(
                year=args.year,
                n_sims=args.n_sims,
                seed=args.seed,
                run_id=args.run_id,
            )
            print(f"\nResult: {result}")

        elif args.cmd == "recommended-entry-bids":
            result = stage_recommended_entry_bids(
                year=args.year,
                strategy=args.strategy,
                budget_points=args.budget_points,
                min_teams=args.min_teams,
                max_teams=args.max_teams,
                min_bid=args.min_bid,
                max_bid=args.max_bid,
                run_id=args.run_id,
                calcutta_id=args.calcutta_id,
            )
            print(f"\nResult: {result}")

        elif args.cmd == "full-pipeline":
            result = run_full_pipeline(
                year=args.year,
                n_sims=args.n_sims,
                seed=args.seed,
                strategy=args.strategy,
                kenpom_scale=args.kenpom_scale,
                calcutta_id=args.calcutta_id,
            )
            print(f"\nFull pipeline results: {result}")

        return 0

    except Exception as e:
        print(f"\nError: {e}", file=sys.stderr)
        import traceback
        traceback.print_exc()
        return 1


if __name__ == "__main__":
    raise SystemExit(main())
