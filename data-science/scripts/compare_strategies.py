"""
Compare different portfolio allocation strategies across all years.

Runs investment reports for each strategy and generates a comparison report.
"""
import json
import subprocess
from pathlib import Path
import pandas as pd


STRATEGIES = [
    "greedy",
    "waterfill_equal",
    "kelly",
    "min_variance",
    "max_sharpe",
    "one_per_region",
    "two_per_region",
    "variance_aware_light",
    "variance_aware_medium",
    "variance_aware_heavy",
]
YEARS = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
N_SIMS = 5000
SEED = 123
BUDGET_POINTS = 100


def run_strategy_for_year(year: int, strategy: str) -> dict:
    """Run investment report for a specific year and strategy."""
    print(f"Running {strategy} for {year}...")
    
    cmd = [
        "python", "-m", "moneyball.cli",
        "investment-report",
        f"out/{year}",
        "--snapshot-name", str(year),
        "--n-sims", str(N_SIMS),
        "--seed", str(SEED),
        "--budget-points", str(BUDGET_POINTS),
        "--strategy", strategy,
        # Default behavior uses cached tournaments - no flag needed!
    ]
    
    result = subprocess.run(
        cmd,
        capture_output=True,
        text=True,
        cwd=Path(__file__).parent,
    )
    
    if result.returncode != 0:
        print(f"  ERROR: Command failed with code {result.returncode}")
        print(f"  STDERR: {result.stderr[-200:]}")
        return None
    
    # Try to parse JSON output
    # The CLI outputs progress messages before JSON, so find the JSON part
    try:
        # Find the JSON object (starts with { and ends with })
        stdout = result.stdout
        json_start = stdout.find('{')
        json_end = stdout.rfind('}')
        
        if json_start == -1 or json_end == -1:
            print(f"  ERROR: No JSON found in output")
            print(f"  Stdout: {stdout[:200]}")
            return None
        
        json_str = stdout[json_start:json_end+1]
        output = json.loads(json_str)
        return output
    except json.JSONDecodeError:
        print(f"  ERROR: Failed to parse JSON output")
        print(f"  Attempted to parse: {json_str[:200]}")
        return None


def load_results(year: int, strategy: str) -> dict:
    """Load results from the most recent run for a year/strategy."""
    base = Path(f"out/{year}/derived/calcutta")
    
    if not base.exists():
        return None
    
    # Find runs with this strategy
    runs = sorted([d for d in base.iterdir() if d.is_dir()], reverse=True)
    
    for run_dir in runs:
        manifest_path = run_dir / "manifest.json"
        if not manifest_path.exists():
            continue
        
        with open(manifest_path) as f:
            manifest = json.load(f)
        
        # Check if this run used the right strategy
        stages = manifest.get("stages", {})
        bids_stage = stages.get("recommended_entry_bids", {})
        config = bids_stage.get("stage_config", {})
        
        if config.get("strategy") == strategy:
            # Load simulated_entry_outcomes
            outcomes_path = run_dir / "simulated_entry_outcomes.parquet"
            if outcomes_path.exists():
                df = pd.read_parquet(outcomes_path)
                return df.iloc[0].to_dict()
    
    return None


def main():
    """Run all strategies across all years and generate comparison report."""
    import time
    
    print("=" * 80)
    print("STRATEGY COMPARISON ACROSS ALL YEARS")
    print("=" * 80)
    print(f"Running {len(STRATEGIES)} strategies x {len(YEARS)} years = {len(STRATEGIES) * len(YEARS)} reports")
    print("Using cached tournaments - should be fast!")
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
            print(f"  Progress: {completed}/{total} ({completed/total*100:.0f}%) | "
                  f"Elapsed: {elapsed:.0f}s | ETA: {remaining:.0f}s")
    
    print("\n" + "=" * 80)
    print("COLLECTING RESULTS")
    print("=" * 80)
    print()
    
    # Collect results
    results = []
    for year in YEARS:
        for strategy in STRATEGIES:
            data = load_results(year, strategy)
            if data:
                results.append({
                    "year": year,
                    "strategy": strategy,
                    "mean_normalized_payout": data.get("mean_normalized_payout", 0.0),
                    "p_top1": data.get("p_top1", 0.0),
                    "p_in_money": data.get("p_in_money", 0.0),
                    "mean_payout_cents": data.get("mean_payout_cents", 0.0),
                })
    
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
        year_df = df[df["year"] == year]
        if len(year_df) == 0:
            continue
        
        print(f"\n{year}:")
        print("-" * 80)
        for _, row in year_df.iterrows():
            print(f"  {row['strategy']:15s} | "
                  f"Norm: {row['mean_normalized_payout']:.3f} | "
                  f"P(1st): {row['p_top1']*100:5.1f}% | "
                  f"P($>0): {row['p_in_money']*100:5.1f}% | "
                  f"${row['mean_payout_cents']/100:6.2f}")
    
    # Summary by strategy
    print("\n" + "=" * 80)
    print("SUMMARY BY STRATEGY (across all years)")
    print("=" * 80)
    print()
    
    summary = df.groupby("strategy").agg({
        "mean_normalized_payout": "mean",
        "p_top1": "mean",
        "p_in_money": "mean",
        "mean_payout_cents": "mean",
    }).round(4)
    
    summary = summary.sort_values("mean_normalized_payout", ascending=False)
    
    print(f"{'Strategy':<15} | {'Norm Payout':>11} | {'P(1st)':>7} | {'P($>0)':>7} | {'Mean $':>8}")
    print("-" * 80)
    for strategy, row in summary.iterrows():
        print(f"{strategy:<15} | "
              f"{row['mean_normalized_payout']:>11.3f} | "
              f"{row['p_top1']*100:>6.1f}% | "
              f"{row['p_in_money']*100:>6.1f}% | "
              f"${row['mean_payout_cents']/100:>7.2f}")
    
    print("\n" + "=" * 80)
    print("WINNER: " + summary.index[0])
    print("=" * 80)
    
    # Save to CSV
    output_path = Path("out/strategy_comparison.csv")
    df.to_csv(output_path, index=False)
    print(f"\nFull results saved to: {output_path}")
    
    summary_path = Path("out/strategy_summary.csv")
    summary.to_csv(summary_path)
    print(f"Summary saved to: {summary_path}")


if __name__ == "__main__":
    main()
