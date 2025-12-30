"""
Analyze normalized payout percentiles for 2025 entries.
"""
from pathlib import Path
import pandas as pd
import numpy as np
from moneyball.utils import points


def calculate_entry_normalized_payout(
    entry_bids: pd.DataFrame,
    tournaments: pd.DataFrame,
    payouts: pd.DataFrame,
    entry_key: str,
) -> float:
    """Calculate mean normalized payout for a single entry."""
    # Get this entry's bids
    bids = entry_bids[entry_bids['entry_key'] == entry_key].copy()
    
    # Calculate market for each team (sum of all bids)
    market_by_team = entry_bids.groupby('team_key')['bid_amount'].sum()
    
    # Get max prize
    max_prize = payouts['amount_cents'].max()
    
    # Calculate payout for each simulation
    normalized_payouts = []
    
    for sim_id in tournaments['sim_id'].unique():
        sim_results = tournaments[tournaments['sim_id'] == sim_id]
        
        # Calculate this entry's total points
        entry_points = 0
        for _, bid_row in bids.iterrows():
            team_key = bid_row['team_key']
            bid = bid_row['bid_amount']
            
            # Get team's wins in this sim
            team_data = sim_results[sim_results['team_key'] == team_key]
            if len(team_data) == 0:
                continue
            
            wins = team_data['wins'].iloc[0]
            byes = team_data['byes'].iloc[0] if 'byes' in team_data.columns else 0
            progress = wins + byes
            
            # Calculate points using the actual point system
            team_points = points.team_points_fixed(int(progress))
            
            # Calculate ownership
            market = market_by_team.get(team_key, 0)
            ownership = bid / (market + bid) if (market + bid) > 0 else 0
            
            entry_points += ownership * team_points
        
        # Calculate all entries' points to determine payout
        all_entry_points = []
        for ek in entry_bids['entry_key'].unique():
            ek_bids = entry_bids[entry_bids['entry_key'] == ek]
            ek_points = 0
            
            for _, bid_row in ek_bids.iterrows():
                team_key = bid_row['team_key']
                bid = bid_row['bid_amount']
                
                team_data = sim_results[sim_results['team_key'] == team_key]
                if len(team_data) == 0:
                    continue
                
                wins = team_data['wins'].iloc[0]
                byes = team_data['byes'].iloc[0] if 'byes' in team_data.columns else 0
                progress = wins + byes
                team_points = points.team_points_fixed(int(progress))
                
                market = market_by_team.get(team_key, 0)
                ownership = bid / (market + bid) if (market + bid) > 0 else 0
                ek_points += ownership * team_points
            
            all_entry_points.append(ek_points)
        
        # Determine finish position
        all_entry_points_sorted = sorted(all_entry_points, reverse=True)
        finish_position = all_entry_points_sorted.index(entry_points) + 1
        
        # Get payout for this finish position
        payout_row = payouts[payouts['position'] == finish_position]
        payout_cents = payout_row['amount_cents'].iloc[0] if len(payout_row) > 0 else 0
        
        # Normalize by max prize
        normalized_payout = payout_cents / max_prize if max_prize > 0 else 0
        normalized_payouts.append(normalized_payout)
    
    return np.mean(normalized_payouts)


def main():
    year = 2025
    year_dir = Path(f"out/{year}")
    
    print(f"\n{'='*80}")
    print(f"2025 NORMALIZED PAYOUT ANALYSIS")
    print(f"{'='*80}\n")
    
    # Load data
    entry_bids = pd.read_parquet(year_dir / "entry_bids.parquet")
    tournaments = pd.read_parquet(year_dir / "derived/tournaments.parquet")
    payouts = pd.read_parquet(year_dir / "payouts.parquet")
    
    # Get our MINLP entry
    calcutta_dir = year_dir / "derived/calcutta"
    latest_run = sorted(calcutta_dir.glob("202512*"))[-1]
    our_bids = pd.read_parquet(latest_run / "recommended_entry_bids.parquet")
    our_bids = our_bids.rename(columns={'bid_amount_points': 'bid_amount'})
    our_bids['entry_key'] = 'our_minlp_strategy'
    
    # Add our entry to the pool
    all_bids = pd.concat([entry_bids, our_bids[['entry_key', 'team_key', 'bid_amount']]], ignore_index=True)
    
    print(f"Calculating normalized payouts for {len(all_bids['entry_key'].unique())} entries...")
    print(f"Using {len(tournaments['sim_id'].unique())} simulations")
    print()
    
    # Calculate normalized payout for each entry
    results = []
    for i, entry_key in enumerate(all_bids['entry_key'].unique(), 1):
        print(f"  [{i}/{len(all_bids['entry_key'].unique())}] {entry_key[:50]}...", end='\r')
        
        norm_payout = calculate_entry_normalized_payout(
            entry_bids=all_bids,
            tournaments=tournaments,
            payouts=payouts,
            entry_key=entry_key,
        )
        
        results.append({
            'entry_key': entry_key,
            'mean_normalized_payout': norm_payout,
            'is_our_strategy': entry_key == 'our_minlp_strategy',
        })
    
    print()
    
    # Create DataFrame and calculate percentiles
    df = pd.DataFrame(results)
    df['percentile_rank'] = df['mean_normalized_payout'].rank(pct=True) * 100
    df = df.sort_values('mean_normalized_payout', ascending=False)
    
    # Display results
    print(f"\n{'='*80}")
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
    
    # Save results
    output_path = latest_run / "entry_percentiles.parquet"
    df.to_parquet(output_path, index=False)
    print(f"✓ Saved results to {output_path}")
    print()


if __name__ == "__main__":
    main()
