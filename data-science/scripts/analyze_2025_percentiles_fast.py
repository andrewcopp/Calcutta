"""
Fast analysis of normalized payout percentiles for 2025 entries.
Uses pre-computed simulation results instead of recalculating.
"""
from pathlib import Path
import pandas as pd
import numpy as np
from moneyball.models.simulated_entry_outcomes import simulate_entry_outcomes


def main():
    year = 2025
    year_dir = Path(f"out/{year}")
    
    print(f"\n{'='*80}")
    print(f"2025 NORMALIZED PAYOUT ANALYSIS")
    print(f"{'='*80}\n")
    
    # Load data
    games = pd.read_parquet(year_dir / "games.parquet")
    teams = pd.read_parquet(year_dir / "teams.parquet")
    payouts = pd.read_parquet(year_dir / "payouts.parquet")
    entry_bids = pd.read_parquet(year_dir / "entry_bids.parquet")
    predicted_game_outcomes = pd.read_parquet(
        year_dir / "derived/predicted_game_outcomes.parquet"
    )
    tournaments = pd.read_parquet(year_dir / "derived/tournaments.parquet")
    
    # Get our MINLP entry
    calcutta_dir = year_dir / "derived/calcutta"
    latest_run = sorted(calcutta_dir.glob("202512*"))[-1]
    our_bids = pd.read_parquet(latest_run / "recommended_entry_bids.parquet")
    
    # Get calcutta_key from payouts
    calcutta_key = payouts["calcutta_key"].iloc[0]
    
    print(f"Analyzing {len(entry_bids['entry_key'].unique())} actual entries + our MINLP strategy")
    print(f"Using {len(tournaments['sim_id'].unique())} simulations\n")
    
    # Calculate normalized payout for each entry
    results = []
    
    # First, calculate for our MINLP strategy
    print("Calculating our MINLP strategy...")
    our_summary, _ = simulate_entry_outcomes(
        games=games,
        teams=teams,
        payouts=payouts,
        entry_bids=entry_bids,
        predicted_game_outcomes=predicted_game_outcomes,
        recommended_entry_bids=our_bids,
        simulated_tournaments=tournaments,
        calcutta_key=calcutta_key,
        n_sims=5000,
        seed=42,
        budget_points=100,
        sim_entry_key="our_minlp_strategy",
        keep_sims=False,
    )
    
    results.append({
        'entry_key': 'our_minlp_strategy',
        'mean_normalized_payout': float(our_summary['mean_normalized_payout'].iloc[0]),
        'is_our_strategy': True,
    })
    
    # Now calculate for each actual entry
    total_entries = len(entry_bids['entry_key'].unique())
    for i, entry_key in enumerate(entry_bids['entry_key'].unique(), 1):
        print(f"  [{i}/{total_entries}] {str(entry_key)[:50]}...", end='\r')
        
        # Get this entry's bids
        entry_data = entry_bids[entry_bids['entry_key'] == entry_key].copy()
        entry_data = entry_data.rename(columns={'bid_amount': 'bid_amount_points'})
        
        try:
            entry_summary, _ = simulate_entry_outcomes(
                games=games,
                teams=teams,
                payouts=payouts,
                entry_bids=entry_bids,
                predicted_game_outcomes=predicted_game_outcomes,
                recommended_entry_bids=entry_data,
                simulated_tournaments=tournaments,
                calcutta_key=calcutta_key,
                n_sims=5000,
                seed=42,
                budget_points=100,
                sim_entry_key=str(entry_key),
                keep_sims=False,
            )
            
            results.append({
                'entry_key': str(entry_key),
                'mean_normalized_payout': float(entry_summary['mean_normalized_payout'].iloc[0]),
                'is_our_strategy': False,
            })
        except Exception as e:
            print(f"\nWarning: Failed to simulate {entry_key}: {e}")
            continue
    
    print("\n")
    
    # Create DataFrame and calculate percentiles
    df = pd.DataFrame(results)
    df['percentile_rank'] = df['mean_normalized_payout'].rank(pct=True) * 100
    df = df.sort_values('mean_normalized_payout', ascending=False)
    
    # Display results
    print(f"{'='*80}")
    print("RESULTS")
    print(f"{'='*80}\n")
    
    our_row = df[df['is_our_strategy']].iloc[0]
    our_rank = (df['mean_normalized_payout'] > our_row['mean_normalized_payout']).sum() + 1
    total_entries = len(df)
    
    print(f"Our MINLP Strategy Performance:")
    print(f"  Mean Normalized Payout: {our_row['mean_normalized_payout']:.3f}")
    print(f"  Percentile Rank: {our_row['percentile_rank']:.1f}%")
    print(f"  Absolute Rank: {our_rank} out of {total_entries}")
    print()
    
    max_prize = payouts['amount_cents'].max()
    print(f"  In dollar terms (max prize = ${max_prize/100:.2f}):")
    print(f"    Average winnings: ${our_row['mean_normalized_payout'] * max_prize / 100:.2f}")
    print()
    
    # Top 10
    print(f"Top 10 Entries by Normalized Payout:")
    print("-" * 80)
    print(f"{'Rank':<6} {'Entry':<45} {'Norm Payout':>12} {'Percentile':>11}")
    print("-" * 80)
    
    for i, (_, row) in enumerate(df.head(10).iterrows(), 1):
        marker = " ← US" if row['is_our_strategy'] else ""
        entry_name = row['entry_key'][:44]
        print(f"{i:<6} {entry_name:<45} "
              f"{row['mean_normalized_payout']:>12.3f} "
              f"{row['percentile_rank']:>10.1f}%{marker}")
    
    print()
    
    # Bottom 5 for context
    print(f"Bottom 5 Entries:")
    print("-" * 80)
    for i, (_, row) in enumerate(df.tail(5).iterrows(), len(df) - 4):
        entry_name = row['entry_key'][:44]
        print(f"{i:<6} {entry_name:<45} "
              f"{row['mean_normalized_payout']:>12.3f} "
              f"{row['percentile_rank']:>10.1f}%")
    
    print()
    
    # Save results
    output_path = latest_run / "entry_percentiles.parquet"
    df.to_parquet(output_path, index=False)
    print(f"✓ Saved results to {output_path}")
    
    # Also save as CSV for easy viewing
    csv_path = latest_run / "entry_percentiles.csv"
    df.to_csv(csv_path, index=False)
    print(f"✓ Saved results to {csv_path}")
    print()


if __name__ == "__main__":
    main()
