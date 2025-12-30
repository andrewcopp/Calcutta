from __future__ import annotations

import argparse
import json
from pathlib import Path

from moneyball.pipeline.runner import run


def main() -> int:
    parser = argparse.ArgumentParser(prog="moneyball")

    sub = parser.add_subparsers(dest="cmd", required=True)

    p_go = sub.add_parser("predicted-game-outcomes")
    p_go.add_argument("snapshot_dir")
    p_go.add_argument("--snapshot-name", dest="snapshot_name", default=None)
    p_go.add_argument("--artifacts-root", dest="artifacts_root", default=None)
    p_go.add_argument("--run-id", dest="run_id", default=None)
    p_go.add_argument("--calcutta-key", dest="calcutta_key", default=None)
    p_go.add_argument(
        "--kenpom-scale",
        dest="kenpom_scale",
        type=float,
        default=10.0,
    )
    p_go.add_argument("--n-sims", dest="n_sims", type=int, default=5000)
    p_go.add_argument("--seed", dest="seed", type=int, default=123)
    p_go.add_argument(
        "--no-cache",
        dest="use_cache",
        action="store_false",
        default=True,
    )

    p_ms = sub.add_parser("predicted-auction-share-of-pool")
    p_ms.add_argument("snapshot_dir")
    p_ms.add_argument("--snapshot-name", dest="snapshot_name", default=None)
    p_ms.add_argument("--artifacts-root", dest="artifacts_root", default=None)
    p_ms.add_argument("--run-id", dest="run_id", default=None)
    p_ms.add_argument(
        "--train-snapshot",
        dest="train_snapshots",
        action="append",
        default=None,
    )
    p_ms.add_argument(
        "--ridge-alpha",
        dest="ridge_alpha",
        type=float,
        default=1.0,
    )
    p_ms.add_argument(
        "--feature-set",
        dest="feature_set",
        default="expanded_last_year_expected",
    )
    p_ms.add_argument(
        "--exclude-entry-name",
        dest="exclude_entry_names",
        action="append",
        default=None,
    )
    p_ms.add_argument(
        "--no-cache",
        dest="use_cache",
        action="store_false",
        default=True,
    )

    p_rb = sub.add_parser("recommended-entry-bids")
    p_rb.add_argument("snapshot_dir")
    p_rb.add_argument("--snapshot-name", dest="snapshot_name", default=None)
    p_rb.add_argument("--artifacts-root", dest="artifacts_root", default=None)
    p_rb.add_argument("--run-id", dest="run_id", default=None)
    p_rb.add_argument(
        "--budget-points",
        dest="budget_points",
        type=int,
        default=100,
    )
    p_rb.add_argument(
        "--min-teams",
        dest="min_teams",
        type=int,
        default=3,
    )
    p_rb.add_argument(
        "--max-teams",
        dest="max_teams",
        type=int,
        default=10,
    )
    p_rb.add_argument(
        "--max-per-team-points",
        dest="max_per_team_points",
        type=int,
        default=50,
    )
    p_rb.add_argument(
        "--min-bid-points",
        dest="min_bid_points",
        type=int,
        default=1,
    )
    p_rb.add_argument(
        "--predicted-total-pool-bids-points",
        dest="predicted_total_pool_bids_points",
        type=float,
        default=None,
    )
    p_rb.add_argument(
        "--no-include-upstream",
        dest="include_upstream",
        action="store_false",
        default=True,
    )
    p_rb.add_argument(
        "--no-cache",
        dest="use_cache",
        action="store_false",
        default=True,
    )

    p_sim = sub.add_parser("simulated-entry-outcomes")
    p_sim.add_argument("snapshot_dir")
    p_sim.add_argument("--snapshot-name", dest="snapshot_name", default=None)
    p_sim.add_argument("--artifacts-root", dest="artifacts_root", default=None)
    p_sim.add_argument("--run-id", dest="run_id", default=None)
    p_sim.add_argument("--calcutta-key", dest="calcutta_key", default=None)
    p_sim.add_argument("--n-sims", dest="n_sims", type=int, default=5000)
    p_sim.add_argument("--seed", dest="seed", type=int, default=123)
    p_sim.add_argument(
        "--budget-points",
        dest="budget_points",
        type=int,
        default=100,
    )
    p_sim.add_argument(
        "--keep-sims",
        dest="keep_sims",
        action="store_true",
        default=False,
    )
    p_sim.add_argument(
        "--no-include-upstream",
        dest="include_upstream",
        action="store_false",
        default=True,
    )
    p_sim.add_argument(
        "--no-cache",
        dest="use_cache",
        action="store_false",
        default=True,
    )

    p_tournaments = sub.add_parser("simulate-tournaments")
    p_tournaments.add_argument("snapshot_dir")
    p_tournaments.add_argument(
        "--snapshot-name", dest="snapshot_name", default=None
    )
    p_tournaments.add_argument(
        "--artifacts-root", dest="artifacts_root", default=None
    )
    p_tournaments.add_argument("--run-id", dest="run_id", default=None)
    p_tournaments.add_argument(
        "--n-sims", dest="n_sims", type=int, default=5000
    )
    p_tournaments.add_argument("--seed", dest="seed", type=int, default=123)
    p_tournaments.add_argument(
        "--regenerate",
        dest="regenerate_tournaments",
        action="store_true",
        default=False,
        help="Force regeneration of tournaments (ignore cache)",
    )
    p_tournaments.add_argument(
        "--no-cache",
        dest="use_cache",
        action="store_false",
        default=True,
    )

    p_report = sub.add_parser("investment-report")
    p_report.add_argument("snapshot_dir")
    p_report.add_argument(
        "--snapshot-name", dest="snapshot_name", default=None
    )
    p_report.add_argument(
        "--artifacts-root", dest="artifacts_root", default=None
    )
    p_report.add_argument("--run-id", dest="run_id", default=None)
    p_report.add_argument("--n-sims", dest="n_sims", type=int, default=5000)
    p_report.add_argument("--seed", dest="seed", type=int, default=123)
    p_report.add_argument(
        "--budget-points",
        dest="budget_points",
        type=int,
        default=100,
    )
    p_report.add_argument(
        "--strategy",
        dest="strategy",
        type=str,
        default="greedy",
        choices=["greedy", "waterfill_equal", "kelly", "min_variance", "max_sharpe"],
        help="Portfolio allocation strategy",
    )
    p_report.add_argument(
        "--no-include-upstream",
        dest="include_upstream",
        action="store_false",
        default=True,
    )
    p_report.add_argument(
        "--no-cache",
        dest="use_cache",
        action="store_false",
        default=True,
    )
    p_report.add_argument(
        "--regenerate-tournaments",
        dest="regenerate_tournaments",
        action="store_true",
        default=False,
        help="Force regeneration of tournament simulations (ignore cache)",
    )

    args = parser.parse_args()

    if args.cmd == "predicted-game-outcomes":
        out = run(
            snapshot_dir=Path(args.snapshot_dir),
            snapshot_name=args.snapshot_name,
            artifacts_root=Path(args.artifacts_root)
            if args.artifacts_root
            else None,
            run_id=args.run_id,
            stages=["predicted_game_outcomes"],
            calcutta_key=args.calcutta_key,
            kenpom_scale=float(args.kenpom_scale),
            n_sims=int(args.n_sims),
            seed=int(args.seed),
            use_cache=bool(args.use_cache),
        )
        print(json.dumps(out, indent=2))
        return 0

    if args.cmd == "simulated-entry-outcomes":
        run_stages = ["simulated_entry_outcomes"]
        if bool(args.include_upstream):
            run_stages = [
                "predicted_game_outcomes",
                "predicted_auction_share_of_pool",
                "recommended_entry_bids",
                "simulated_entry_outcomes",
            ]

        out = run(
            snapshot_dir=Path(args.snapshot_dir),
            snapshot_name=args.snapshot_name,
            artifacts_root=Path(args.artifacts_root)
            if args.artifacts_root
            else None,
            run_id=args.run_id,
            stages=run_stages,
            calcutta_key=args.calcutta_key,
            sim_n_sims=int(args.n_sims),
            sim_seed=int(args.seed),
            sim_budget_points=int(args.budget_points),
            sim_keep_sims=bool(args.keep_sims),
            use_cache=bool(args.use_cache),
        )
        print(json.dumps(out, indent=2))
        return 0

    if args.cmd == "predicted-auction-share-of-pool":
        out = run(
            snapshot_dir=Path(args.snapshot_dir),
            snapshot_name=args.snapshot_name,
            artifacts_root=Path(args.artifacts_root)
            if args.artifacts_root
            else None,
            run_id=args.run_id,
            stages=["predicted_auction_share_of_pool"],
            auction_train_snapshots=args.train_snapshots,
            auction_ridge_alpha=float(args.ridge_alpha),
            auction_feature_set=str(args.feature_set),
            auction_exclude_entry_names=args.exclude_entry_names,
            use_cache=bool(args.use_cache),
        )
        print(json.dumps(out, indent=2))
        return 0

    if args.cmd == "recommended-entry-bids":
        run_stages = ["recommended_entry_bids"]
        if bool(args.include_upstream):
            run_stages = [
                "predicted_game_outcomes",
                "predicted_auction_share_of_pool",
                "recommended_entry_bids",
            ]

        out = run(
            snapshot_dir=Path(args.snapshot_dir),
            snapshot_name=args.snapshot_name,
            artifacts_root=Path(args.artifacts_root)
            if args.artifacts_root
            else None,
            run_id=args.run_id,
            stages=run_stages,
            bids_budget_points=int(args.budget_points),
            bids_min_teams=int(args.min_teams),
            bids_max_teams=int(args.max_teams),
            bids_max_per_team_points=int(args.max_per_team_points),
            bids_min_bid_points=int(args.min_bid_points),
            bids_predicted_total_pool_bids_points=(
                float(args.predicted_total_pool_bids_points)
                if args.predicted_total_pool_bids_points
                else None
            ),
            use_cache=bool(args.use_cache),
        )
        print(json.dumps(out, indent=2))
        return 0

    if args.cmd == "simulate-tournaments":
        out = run(
            snapshot_dir=Path(args.snapshot_dir),
            snapshot_name=args.snapshot_name,
            artifacts_root=Path(args.artifacts_root)
            if args.artifacts_root
            else None,
            run_id=args.run_id,
            stages=["predicted_game_outcomes", "simulated_tournaments"],
            n_sims=int(args.n_sims),
            seed=int(args.seed),
            regenerate_tournaments=bool(args.regenerate_tournaments),
            use_cache=bool(args.use_cache),
        )
        print(json.dumps(out, indent=2))
        return 0

    if args.cmd == "investment-report":
        run_stages = ["investment_report"]
        if bool(args.include_upstream):
            run_stages = [
                "predicted_game_outcomes",
                "predicted_auction_share_of_pool",
                "recommended_entry_bids",
                "simulated_tournaments",
                "simulated_entry_outcomes",
                "investment_report",
            ]

        out = run(
            snapshot_dir=Path(args.snapshot_dir),
            snapshot_name=args.snapshot_name,
            artifacts_root=Path(args.artifacts_root)
            if args.artifacts_root
            else None,
            run_id=args.run_id,
            stages=run_stages,
            sim_n_sims=int(args.n_sims),
            sim_seed=int(args.seed),
            sim_budget_points=int(args.budget_points),
            bids_strategy=str(args.strategy),
            regenerate_tournaments=bool(args.regenerate_tournaments),
            use_cache=bool(args.use_cache),
        )
        print(json.dumps(out, indent=2))
        return 0

    raise ValueError(f"unknown cmd: {args.cmd}")


if __name__ == "__main__":
    raise SystemExit(main())
