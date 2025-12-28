import argparse
import json
from pathlib import Path

from calcutta_ds.backtest_scaffold_runner import run_backtest_scaffold


def main() -> int:
    parser = argparse.ArgumentParser(
        description=(
            "Backtesting scaffold: construct a feasible portfolio "
            "under auction constraints using a simple ranking score."
        )
    )
    parser.add_argument(
        "snapshot_dir",
        help=(
            "Path to an ingested snapshot directory "
            "(expects derived/team_dataset.parquet)"
        ),
    )
    parser.add_argument(
        "--calcutta-key",
        dest="calcutta_key",
        default=None,
        help="Which calcutta_key to backtest (required if multiple exist)",
    )
    parser.add_argument(
        "--score",
        dest="score_col",
        default="kenpom_net",
        help="Column to rank teams by (default: kenpom_net)",
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
        default="fixed",
        choices=["auto", "round_scoring", "fixed"],
        help=(
            "How to convert wins+byes into team points. "
            "auto uses round_scoring if available."
        ),
    )
    parser.add_argument(
        "--sim-entry-key",
        dest="sim_entry_key",
        default="simulated:entry",
        help="Entry key used for the simulated portfolio",
    )
    parser.add_argument(
        "--expected-sims",
        dest="expected_sims",
        type=int,
        default=0,
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
        help=(
            "If set, read a JSON file containing kenpom_scale and override "
            "--kenpom-scale (used for expected simulation)"
        ),
    )
    parser.add_argument(
        "--out",
        dest="out_path",
        default=None,
        help="Write portfolio JSON to this path (default: stdout)",
    )

    args = parser.parse_args()

    out = run_backtest_scaffold(
        snapshot_dir=Path(args.snapshot_dir),
        calcutta_key=args.calcutta_key,
        score_col=args.score_col,
        budget=args.budget,
        min_teams=args.min_teams,
        max_teams=args.max_teams,
        max_per_team=args.max_per_team,
        min_bid=args.min_bid,
        points_mode=args.points_mode,
        sim_entry_key=args.sim_entry_key,
        expected_sims=args.expected_sims,
        expected_seed=args.expected_seed,
        expected_use_historical_winners=args.expected_use_historical_winners,
        kenpom_scale=args.kenpom_scale,
        kenpom_scale_file=args.kenpom_scale_file,
    )

    out_json = json.dumps(out, indent=2) + "\n"
    if args.out_path:
        Path(args.out_path).write_text(out_json, encoding="utf-8")
    else:
        print(out_json, end="")

    return 0
