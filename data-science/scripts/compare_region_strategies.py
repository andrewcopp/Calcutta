"""
Compare region-constrained strategies vs greedy baseline.

This script compares:
- greedy: Standard unconstrained greedy optimizer
- one_per_region: Max 1 team per region (4 teams total)
- two_per_region: Max 2 teams per region (8 teams total)

The goal is to test whether artificial regional diversification
improves outcomes by reducing correlation risk.
"""
import json
import subprocess
from pathlib import Path
import pandas as pd


STRATEGIES = [
    "greedy",
    "one_per_region",
    "two_per_region",
]
YEARS = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
N_SIMS = 5000
SEED = 123
BUDGET_POINTS = 100


def run_strategy_for_year(year: int, strategy: str) -> dict:
    """Run investment report for a specific year and strategy."""
    print(f"  Running {strategy}...")

    cmd = [
        "python", "-m", "moneyball.cli",
        "investment-report",
        f"out/{year}",
        "--snapshot-name", str(year),
        "--n-sims", str(N_SIMS),
        "--seed", str(SEED),
        "--budget-points", str(BUDGET_POINTS),
        "--strategy", strategy,
    ]

    result = subprocess.run(
        cmd,
        capture_output=True,
        text=True,
        cwd=Path(__file__).parent.parent,
    )

    if result.returncode != 0:
        print(f"    ERROR: Command failed with code {result.returncode}")
        print(f"    STDERR: {result.stderr[-200:]}")
        return None

    try:
        stdout = result.stdout
        json_start = stdout.find('{')
        json_end = stdout.rfind('}')

        if json_start == -1 or json_end == -1:
            print(f"    ERROR: No JSON found in output")
            return None

        json_str = stdout[json_start:json_end+1]
        output = json.loads(json_str)
        return output
    except json.JSONDecodeError:
        print(f"    ERROR: Failed to parse JSON output")
        return None


def load_portfolio_details(year: int, strategy: str) -> pd.DataFrame:
    """Load the recommended bids to see team selection."""
    base = Path(f"out/{year}/derived/calcutta")

    if not base.exists():
        return None

    runs = sorted([d for d in base.iterdir() if d.is_dir()], reverse=True)

    for run_dir in runs:
        manifest_path = run_dir / "manifest.json"
        if not manifest_path.exists():
            continue

        with open(manifest_path) as f:
            manifest = json.load(f)

        stages = manifest.get("stages", {})
        bids_stage = stages.get("recommended_entry_bids", {})
        config = bids_stage.get("stage_config", {})

        if config.get("strategy") == strategy:
            bids_path = run_dir / "recommended_entry_bids.parquet"
            if bids_path.exists():
                return pd.read_parquet(bids_path)

    return None


def main():
    """Run region-constrained strategies and compare results."""
    import time

    print("=" * 80)
    print("REGION-CONSTRAINED STRATEGY COMPARISON")
    print("=" * 80)
    print(f"Strategies: {', '.join(STRATEGIES)}")
    print(f"Years: {', '.join(map(str, YEARS))}")
    print(f"Simulations per year: {N_SIMS:,}")
    print("=" * 80)
    print()

    start_time = time.time()
    completed = 0
    total = len(STRATEGIES) * len(YEARS)

    # Run all strategies for all years
    for year in YEARS:
        print(f"\n{'='*80}")
        print(f"YEAR: {year}")
        print(f"{'='*80}")

        for strategy in STRATEGIES:
            run_strategy_for_year(year, strategy)
            completed += 1
            elapsed = time.time() - start_time
            avg_time = elapsed / completed
            remaining = (total - completed) * avg_time
            print(f"  Progress: {completed}/{total} "
                  f"({completed/total*100:.0f}%) | "
                  f"Elapsed: {elapsed:.0f}s | ETA: {remaining:.0f}s")

    print("\n" + "=" * 80)
    print("COLLECTING RESULTS")
    print("=" * 80)
    print()

    # Collect results
    results = []
    for year in YEARS:
        for strategy in STRATEGIES:
            # Load portfolio details
            portfolio = load_portfolio_details(year, strategy)

            if portfolio is not None:
                # Count teams and regions
                n_teams = len(portfolio)
                if "region" in portfolio.columns:
                    regions = portfolio["region"].value_counts().to_dict()
                    n_regions = len(regions)
                else:
                    regions = {}
                    n_regions = 0

                # Load outcomes
                base = Path(f"out/{year}/derived/calcutta")
                runs = sorted(
                    [d for d in base.iterdir() if d.is_dir()],
                    reverse=True
                )

                for run_dir in runs:
                    manifest_path = run_dir / "manifest.json"
                    if not manifest_path.exists():
                        continue

                    with open(manifest_path) as f:
                        manifest = json.load(f)

                    stages = manifest.get("stages", {})
                    bids_stage = stages.get("recommended_entry_bids", {})
                    config = bids_stage.get("stage_config", {})

                    if config.get("strategy") == strategy:
                        outcomes_path = run_dir / "simulated_entry_outcomes.parquet"
                        if outcomes_path.exists():
                            df = pd.read_parquet(outcomes_path)
                            outcome = df.iloc[0].to_dict()

                            results.append({
                                "year": year,
                                "strategy": strategy,
                                "n_teams": n_teams,
                                "n_regions": n_regions,
                                "regions": str(regions),
                                "mean_normalized_payout": outcome.get(
                                    "mean_normalized_payout", 0.0
                                ),
                                "p_top1": outcome.get("p_top1", 0.0),
                                "p_in_money": outcome.get("p_in_money", 0.0),
                                "mean_payout_cents": outcome.get(
                                    "mean_payout_cents", 0.0
                                ),
                            })
                        break

    if not results:
        print("No results found!")
        return

    df = pd.DataFrame(results)

    # Generate comparison report
    print("\n" + "=" * 80)
    print("RESULTS BY YEAR")
    print("=" * 80)
    print()

    for year in YEARS:
        year_df = df[df["year"] == year].copy()
        if len(year_df) == 0:
            continue

        print(f"\n{year}:")
        print("-" * 80)

        # Sort by normalized payout
        year_df = year_df.sort_values(
            "mean_normalized_payout",
            ascending=False
        )

        for _, row in year_df.iterrows():
            print(f"  {row['strategy']:16s} | "
                  f"Teams: {row['n_teams']:2.0f} | "
                  f"Regions: {row['n_regions']:1.0f} | "
                  f"Norm: {row['mean_normalized_payout']:.3f} | "
                  f"P(1st): {row['p_top1']*100:5.1f}% | "
                  f"P($>0): {row['p_in_money']*100:5.1f}%")

    # Summary by strategy
    print("\n" + "=" * 80)
    print("SUMMARY BY STRATEGY (across all years)")
    print("=" * 80)
    print()

    summary = df.groupby("strategy").agg({
        "mean_normalized_payout": ["mean", "std"],
        "p_top1": "mean",
        "p_in_money": "mean",
        "mean_payout_cents": "mean",
        "n_teams": "mean",
        "n_regions": "mean",
    }).round(4)

    summary = summary.sort_values(
        ("mean_normalized_payout", "mean"),
        ascending=False
    )

    print(f"{'Strategy':<16} | {'Teams':>5} | {'Regions':>7} | "
          f"{'Norm μ':>7} | {'Norm σ':>7} | "
          f"{'P(1st)':>7} | {'P($>0)':>7}")
    print("-" * 80)

    for strategy in summary.index:
        row = summary.loc[strategy]
        print(f"{strategy:<16} | "
              f"{row[('n_teams', 'mean')]:>5.1f} | "
              f"{row[('n_regions', 'mean')]:>7.1f} | "
              f"{row[('mean_normalized_payout', 'mean')]:>7.3f} | "
              f"{row[('mean_normalized_payout', 'std')]:>7.3f} | "
              f"{row[('p_top1', 'mean')]*100:>6.1f}% | "
              f"{row[('p_in_money', 'mean')]*100:>6.1f}%")

    print("\n" + "=" * 80)
    winner = summary.index[0]
    winner_norm = summary.loc[winner, ("mean_normalized_payout", "mean")]
    greedy_norm = summary.loc["greedy", ("mean_normalized_payout", "mean")]
    improvement = (winner_norm - greedy_norm) / greedy_norm * 100

    print(f"WINNER: {winner}")
    if winner != "greedy":
        print(f"Improvement over greedy: {improvement:+.1f}%")
    print("=" * 80)

    # Save to CSV
    output_path = Path("out/region_strategy_comparison.csv")
    output_path.parent.mkdir(exist_ok=True)
    df.to_csv(output_path, index=False)
    print(f"\nFull results saved to: {output_path}")

    summary_path = Path("out/region_strategy_summary.csv")
    summary.to_csv(summary_path)
    print(f"Summary saved to: {summary_path}")

    # Show example portfolios
    print("\n" + "=" * 80)
    print("EXAMPLE PORTFOLIOS (2024)")
    print("=" * 80)

    for strategy in STRATEGIES:
        portfolio = load_portfolio_details(2024, strategy)
        if portfolio is not None:
            print(f"\n{strategy}:")
            print("-" * 80)

            # Sort by bid amount
            portfolio = portfolio.sort_values(
                "bid_amount_points",
                ascending=False
            )

            for _, team in portfolio.iterrows():
                region = team.get("region", "?")
                print(f"  ${team['bid_amount_points']:2.0f} | "
                      f"{region:10s} | "
                      f"{team['team_key']}")


if __name__ == "__main__":
    main()
