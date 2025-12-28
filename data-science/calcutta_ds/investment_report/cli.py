from __future__ import annotations

import argparse
from pathlib import Path

from calcutta_ds.investment_report.report_runner import (
    load_scale_and_build_report,
    write_report,
)


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Generate a single JSON report that backtests an "
            "underpriced-team strategy across snapshots."
        )
    )
    parser.add_argument(
        "out_root",
        help="Path to Option-A out-root (contains snapshot dirs)",
    )
    parser.add_argument(
        "--ridge-alpha",
        dest="ridge_alpha",
        type=float,
        default=1.0,
        help="Ridge regularization strength for market model (default: 1.0)",
    )
    parser.add_argument(
        "--train-mode",
        dest="train_mode",
        default="past_only",
        choices=["past_only", "loo"],
        help=(
            "past_only trains only on earlier snapshots (default: past_only). "
            "loo trains on all snapshots except the one being predicted."
        ),
    )
    parser.add_argument(
        "--budget",
        dest="budget",
        type=float,
        default=100.0,
        help="Total bankroll (default: 100)",
    )
    parser.add_argument(
        "--min-teams",
        dest="min_teams",
        type=int,
        default=3,
        help="Minimum number of teams to buy (default: 3)",
    )
    parser.add_argument(
        "--max-teams",
        dest="max_teams",
        type=int,
        default=10,
        help="Maximum number of teams to buy (default: 10)",
    )
    parser.add_argument(
        "--max-per-team",
        dest="max_per_team",
        type=float,
        default=50.0,
        help="Max spend per team (default: 50)",
    )
    parser.add_argument(
        "--min-bid",
        dest="min_bid",
        type=float,
        default=1.0,
        help="Min bid per team used by the scaffold (default: 1)",
    )
    parser.add_argument(
        "--points-mode",
        dest="points_mode",
        default="auto",
        choices=["auto", "round_scoring", "fixed"],
        help=(
            "How to compute points (default: fixed)"
        ),
    )
    parser.add_argument(
        "--allocation-mode",
        dest="allocation_mode",
        default="equal",
        choices=["equal", "expected_points", "greedy", "knapsack"],
        help=(
            "How to allocate bids across selected teams. equal matches the "
            "old behavior. expected_points concentrates bids by expected "
            "points per dollar (requires --expected-sims > 0). greedy "
            "optimizes both team selection and bid sizing (requires "
            "--expected-sims > 0). knapsack uses a DP to find the globally "
            "optimal integer-dollar allocation for expected captured points "
            "(requires --expected-sims > 0)."
        ),
    )
    parser.add_argument(
        "--exclude-entry-name",
        dest="exclude_entry_names",
        action="append",
        default=["Andrew Copp"],
        help=(
            "Exclude entries whose entry_name contains this string when "
            "building market totals and standings (default: Andrew Copp). "
            "May be specified multiple times."
        ),
    )
    parser.add_argument(
        "--real-buyin-dollars",
        dest="real_buyin_dollars",
        type=float,
        default=25.0,
        help=(
            "Real entry fee in dollars for real-money ROI reporting "
            "(default: 25)"
        ),
    )
    parser.add_argument(
        "--debug-output",
        dest="debug_output",
        action="store_true",
        help=(
            "Include debug fields in the report JSON"
        ),
    )
    parser.add_argument(
        "--greedy-objective",
        dest="greedy_objective",
        default="expected_points",
        choices=[
            "expected_points",
            "expected_payout",
            "expected_utility_payout",
            "mean_finish_position",
            "p_top1",
            "p_top3",
            "p_top6",
        ],
        help=(
            "Objective used by --allocation-mode greedy. expected_points "
            "matches old behavior. mean_finish_position/p_top* optimize "
            "contest outcome. expected_payout/expected_utility_payout optimize "
            "expected payout using a payout table."
        ),
    )

    parser.add_argument(
        "--payout-snapshot",
        dest="payout_snapshot",
        default="2025",
        help=(
            "Snapshot year to use for payout table during contest optimization "
            "(default: 2025)."
        ),
    )
    parser.add_argument(
        "--payout-utility",
        dest="payout_utility",
        default="power",
        choices=["linear", "log", "power", "exp"],
        help=(
            "Utility function applied to per-sim payout when using "
            "expected_utility_payout (default: power)."
        ),
    )
    parser.add_argument(
        "--payout-utility-gamma",
        dest="payout_utility_gamma",
        type=float,
        default=1.2,
        help=(
            "Gamma parameter for payout utility when --payout-utility=power "
            "(default: 1.2)."
        ),
    )
    parser.add_argument(
        "--payout-utility-alpha",
        dest="payout_utility_alpha",
        type=float,
        default=1.0,
        help=(
            "Alpha parameter for payout utility when --payout-utility=exp "
            "(default: 1.0)."
        ),
    )
    parser.add_argument(
        "--greedy-contest-sims",
        dest="greedy_contest_sims",
        type=int,
        default=500,
        help=(
            "Number of tournament sims used during greedy contest "
            "optimization "
            "(default: 500). "
            "Uses --expected-sims if <= 0."
        ),
    )
    parser.add_argument(
        "--expected-sims",
        dest="expected_sims",
        type=int,
        default=2000,
        help="If >0, run Monte Carlo for expected payout/ROI",
    )
    parser.add_argument(
        "--expected-seed",
        dest="expected_seed",
        type=int,
        default=1,
        help="Random seed for expected Monte Carlo",
    )
    parser.add_argument(
        "--expected-use-historical-winners",
        dest="expected_use_historical_winners",
        action="store_true",
        help=(
            "If set, expected simulation will use winner_team_key values from "
            "games.parquet (leaky for true pre-tournament expectations)"
        ),
    )
    parser.add_argument(
        "--kenpom-scale",
        dest="kenpom_scale",
        type=float,
        default=10.0,
        help="Logistic scale for kenpom_net diff to win prob",
    )
    parser.add_argument(
        "--kenpom-scale-file",
        dest="kenpom_scale_file",
        default=None,
        help="If set, read a JSON file containing kenpom_scale",
    )
    parser.add_argument(
        "--summary-min-train-snapshots",
        dest="summary_min_train_snapshots",
        type=int,
        default=2,
        help=(
            "Exclude years from summary metrics if the market model had fewer "
            "than this many train snapshots (default: 1)."
        ),
    )
    parser.add_argument(
        "--out",
        dest="out_path",
        default=None,
        help=(
            "Write report JSON to this path "
            "(default: <out_root>/report.json)"
        ),
    )

    parser.add_argument(
        "--only-snapshot",
        dest="only_snapshot",
        default=None,
        help=(
            "If set, run the report for only this snapshot (e.g. 2025)"
        ),
    )
    parser.add_argument(
        "--snapshots",
        dest="snapshots",
        default=None,
        help=(
            "If set, run the report only for this comma-separated list of "
            "snapshots (e.g. 2023,2024,2025)"
        ),
    )

    parser.add_argument(
        "--artifacts-dir",
        dest="artifacts_dir",
        default=None,
        help=(
            "If set, write per-snapshot stage artifacts under this directory "
            "(scores/investments/roi/portfolio/evaluation)"
        ),
    )
    parser.add_argument(
        "--use-cache",
        dest="use_cache",
        action="store_true",
        help=(
            "If set, reuse cached artifacts from --artifacts-dir when present"
        ),
    )

    args = parser.parse_args()

    only_snapshot = str(args.only_snapshot) if args.only_snapshot else None
    snapshots = str(args.snapshots) if args.snapshots else None
    if only_snapshot and snapshots:
        raise ValueError("cannot set both --only-snapshot and --snapshots")

    snapshot_filter = None
    if only_snapshot:
        snapshot_filter = [str(only_snapshot)]
    elif snapshots:
        snapshot_filter = [
            s.strip() for s in str(snapshots).split(",") if s.strip()
        ]

    out_root = Path(args.out_root)
    report = load_scale_and_build_report(
        out_root=out_root,
        kenpom_scale=float(args.kenpom_scale),
        kenpom_scale_file=args.kenpom_scale_file,
        ridge_alpha=float(args.ridge_alpha),
        train_mode=str(args.train_mode),
        budget=float(args.budget),
        real_buyin_dollars=float(args.real_buyin_dollars),
        min_teams=int(args.min_teams),
        max_teams=int(args.max_teams),
        max_per_team=float(args.max_per_team),
        min_bid=float(args.min_bid),
        points_mode=str(args.points_mode),
        allocation_mode=str(args.allocation_mode),
        greedy_objective=str(args.greedy_objective),
        greedy_contest_sims=int(args.greedy_contest_sims),
        payout_snapshot=(
            str(args.payout_snapshot) if args.payout_snapshot is not None else None
        ),
        payout_utility=str(args.payout_utility),
        payout_utility_gamma=float(args.payout_utility_gamma),
        payout_utility_alpha=float(args.payout_utility_alpha),
        exclude_entry_names=list(args.exclude_entry_names or []),
        debug_output=bool(args.debug_output),
        expected_sims=int(args.expected_sims),
        expected_seed=int(args.expected_seed),
        expected_use_historical_winners=bool(
            args.expected_use_historical_winners
        ),
        summary_min_train_snapshots=int(args.summary_min_train_snapshots),
        artifacts_dir=(
            str(args.artifacts_dir) if args.artifacts_dir is not None else None
        ),
        use_cache=bool(args.use_cache),
        snapshot_filter=snapshot_filter,
    )

    out_path = write_report(
        out_root=out_root,
        out_path=args.out_path,
        report=report,
    )
    print(str(out_path))
    return 0
