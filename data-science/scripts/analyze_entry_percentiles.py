"""
Analyze percentile rankings for our MINLP strategy vs all actual entries.
"""
from pathlib import Path
import pandas as pd
from moneyball.models.entry_percentile_analysis import (
    calculate_entry_percentiles,
    generate_percentile_report,
)


def analyze_year(year: int):
    """Analyze percentile rankings for a given year."""
    year_dir = Path(f"out/{year}")
    
    # Load required data
    games = pd.read_parquet(year_dir / "games.parquet")
    teams = pd.read_parquet(year_dir / "teams.parquet")
    payouts = pd.read_parquet(year_dir / "payouts.parquet")
    entry_bids = pd.read_parquet(year_dir / "entry_bids.parquet")
    predicted_game_outcomes = pd.read_parquet(
        year_dir / "derived/predicted_game_outcomes.parquet"
    )
    simulated_tournaments = pd.read_parquet(
        year_dir / "derived/tournaments.parquet"
    )
    
    # Get latest MINLP run
    calcutta_dir = year_dir / "derived/calcutta"
    latest_run = sorted(calcutta_dir.glob("202512*"))[-1]
    recommended_entry_bids = pd.read_parquet(
        latest_run / "recommended_entry_bids.parquet"
    )
    
    # Get calcutta_key
    calcutta_key = teams["calcutta_key"].iloc[0]
    
    print(f"\n{'='*80}")
    print(f"YEAR {year}")
    print(f"{'='*80}")
    
    # Calculate percentiles
    percentiles_df = calculate_entry_percentiles(
        games=games,
        teams=teams,
        payouts=payouts,
        entry_bids=entry_bids,
        predicted_game_outcomes=predicted_game_outcomes,
        recommended_entry_bids=recommended_entry_bids,
        simulated_tournaments=simulated_tournaments,
        calcutta_key=calcutta_key,
        n_sims=5000,
        seed=42,
        budget_points=100,
    )
    
    # Generate report
    report = generate_percentile_report(percentiles_df)
    print(report)
    
    # Save results
    output_path = latest_run / "entry_percentiles.parquet"
    percentiles_df.to_parquet(output_path, index=False)
    print(f"\n✓ Saved percentiles to {output_path}")
    
    return percentiles_df


if __name__ == "__main__":
    years = [2024, 2023, 2022, 2021]
    
    all_results = {}
    for year in years:
        try:
            percentiles_df = analyze_year(year)
            all_results[year] = percentiles_df
        except Exception as e:
            print(f"\nError analyzing {year}: {e}")
            import traceback
            traceback.print_exc()
    
    # Summary across years
    print(f"\n{'='*80}")
    print("SUMMARY ACROSS ALL YEARS")
    print(f"{'='*80}\n")
    
    print(f"{'Year':<6} {'Norm Payout':>12} {'Percentile':>11} {'Rank':>8}")
    print("-" * 80)
    
    for year in years:
        if year not in all_results:
            continue
        
        df = all_results[year]
        our_row = df[df["is_our_strategy"]].iloc[0]
        our_rank = (df["mean_normalized_payout"] >
                   our_row["mean_normalized_payout"]).sum() + 1
        total = len(df)
        
        print(f"{year:<6} {our_row['mean_normalized_payout']:>12.3f} "
              f"{our_row['percentile_rank']:>10.1f}% "
              f"{our_rank:>3}/{total:<3}")
    
    # Calculate average performance
    if all_results:
        avg_norm = sum(
            all_results[y][all_results[y]["is_our_strategy"]].iloc[0][
                "mean_normalized_payout"
            ]
            for y in all_results
        ) / len(all_results)
        
        avg_percentile = sum(
            all_results[y][all_results[y]["is_our_strategy"]].iloc[0][
                "percentile_rank"
            ]
            for y in all_results
        ) / len(all_results)
        
        print("-" * 80)
        print(f"{'AVG':<6} {avg_norm:>12.3f} {avg_percentile:>10.1f}%")
