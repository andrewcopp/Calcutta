"""
Calculate normalized payout metrics for all entries.

Metrics:
1. normalized_average_payout: Average payout divided by average pool payout
   - 1.0 = winning every simulation
   - 0.0 = never making money
   
2. percentile_normalized_payout: Percentile rank among all entries
   - 100 = best entry
   - 0 = worst entry
"""
from pathlib import Path
import pandas as pd
import numpy as np

POINTS_PER_WIN = {0: 0, 1: 1, 2: 2, 3: 4, 4: 8, 5: 16, 6: 32}


def calculate_entry_performance_metrics(
    year: int,
    strategy_name: str = "minlp",
) -> dict:
    """
    Calculate normalized payout metrics for all entries including ours.
    """
    year_dir = Path(f"out/{year}")
    
    # Load data
    entry_bids = pd.read_parquet(year_dir / "entry_bids.parquet")
    tournaments = pd.read_parquet(year_dir / "derived/tournaments.parquet")
    entries = pd.read_parquet(year_dir / "entries.parquet")
    
    # Get our strategy's bids
    calcutta_dir = year_dir / "derived/calcutta"
    latest_run = sorted(calcutta_dir.glob("20251230T193*"))[-1]
    our_bids = pd.read_parquet(latest_run / "recommended_entry_bids.parquet")
    
    # Calculate actual markets (sum of all actual bids)
    actual_markets = entry_bids.groupby('team_key')[
        'bid_amount'
    ].sum().to_dict()
    
    # Add points to tournaments
    tournaments['points'] = tournaments['wins'].map(POINTS_PER_WIN)
    
    # Calculate payout for each entry in each simulation
    print(f"Calculating metrics for {year} ({strategy_name})...")
    print("=" * 80)
    
    all_entries = {}
    sim_ids = tournaments['sim_id'].unique()
    
    # Calculate for all actual entries
    for entry_key in entry_bids['entry_key'].unique():
        entry_data = entry_bids[entry_bids['entry_key'] == entry_key]
        entry_payouts = []
        
        for sim_id in sim_ids:
            sim_results = tournaments[tournaments['sim_id'] == sim_id]
            
            total_payout = 0
            for _, bid_row in entry_data.iterrows():
                team_key = bid_row['team_key']
                bid = bid_row['bid_amount']
                
                team_points = sim_results[
                    sim_results['team_key'] == team_key
                ]['points'].values
                if len(team_points) == 0:
                    continue
                team_points = team_points[0]
                
                market = actual_markets.get(team_key, 0)
                ownership = bid / (market + bid)
                total_payout += ownership * team_points
            
            entry_payouts.append(total_payout)
        
        all_entries[entry_key] = np.array(entry_payouts)
    
    # Calculate for our strategy
    our_payouts = []
    for sim_id in sim_ids:
        sim_results = tournaments[tournaments['sim_id'] == sim_id]
        
        total_payout = 0
        for _, bid_row in our_bids.iterrows():
            team_key = bid_row['team_key']
            bid = bid_row['bid_amount_points']
            
            team_points = sim_results[
                sim_results['team_key'] == team_key
            ]['points'].values
            if len(team_points) == 0:
                continue
            team_points = team_points[0]
            
            market = actual_markets.get(team_key, 0)
            ownership = bid / (market + bid)
            total_payout += ownership * team_points
        
        our_payouts.append(total_payout)
    
    our_payouts = np.array(our_payouts)
    all_entries[f'our_{strategy_name}'] = our_payouts
    
    # Calculate normalized metrics
    # Maximum possible payout = winning 1st place in every simulation
    # 1st place gets the highest payout in each simulation
    max_possible_payout = 0
    for sim_id in sim_ids:
        # Find max payout in this simulation across all entries
        sim_payouts = [
            all_entries[k][sim_id] for k in all_entries.keys()
        ]
        max_possible_payout += max(sim_payouts)
    
    # Average max payout per simulation
    max_avg_payout = max_possible_payout / len(sim_ids)
    
    # For each entry, calculate normalized_average_payout
    entry_metrics = {}
    for entry_key, payouts in all_entries.items():
        avg_payout = payouts.mean()
        # Normalize by the average 1st place payout
        normalized_avg_payout = avg_payout / max_avg_payout
        
        entry_metrics[entry_key] = {
            'avg_payout': avg_payout,
            'normalized_avg_payout': normalized_avg_payout,
            'median_payout': np.median(payouts),
            'std_payout': np.std(payouts),
            'p_positive': (payouts > 0).mean(),
        }
    
    # Calculate percentile rankings
    normalized_payouts = [
        m['normalized_avg_payout'] for m in entry_metrics.values()
    ]
    
    for entry_key, metrics in entry_metrics.items():
        # Percentile: what % of entries are worse than this one
        percentile = (
            sum(1 for x in normalized_payouts
                if x < metrics['normalized_avg_payout'])
            / len(normalized_payouts) * 100
        )
        metrics['percentile_rank'] = percentile
    
    # Summary statistics
    print(f"\nPool Statistics:")
    print(f"  Total entries: {len(entries)}")
    print(f"  Total simulations: {len(sim_ids)}")
    print(f"  Average 1st place payout: {max_avg_payout:.2f} points")
    print()
    
    # Our performance
    our_key = f'our_{strategy_name}'
    our_metrics = entry_metrics[our_key]
    
    print(f"Our {strategy_name.upper()} Performance:")
    print(f"  Average payout: {our_metrics['avg_payout']:.2f} points")
    print(f"  Normalized avg payout: {our_metrics['normalized_avg_payout']:.3f}")
    print(f"  Percentile rank: {our_metrics['percentile_rank']:.1f}%")
    print(f"  P(positive payout): {our_metrics['p_positive']:.1%}")
    print()
    
    # Top 10 entries by normalized payout
    sorted_entries = sorted(
        entry_metrics.items(),
        key=lambda x: x[1]['normalized_avg_payout'],
        reverse=True
    )
    
    print("Top 10 Entries by Normalized Average Payout:")
    print("-" * 80)
    print(f"{'Rank':<6s} {'Entry':<30s} {'Norm Payout':>12s} "
          f"{'Percentile':>11s}")
    print("-" * 80)
    
    for i, (entry_key, metrics) in enumerate(sorted_entries[:10], 1):
        is_ours = entry_key == our_key
        marker = " ← US" if is_ours else ""
        print(f"{i:<6d} {entry_key[:29]:<30s} "
              f"{metrics['normalized_avg_payout']:>12.3f} "
              f"{metrics['percentile_rank']:>10.1f}%{marker}")
    
    print()
    
    # Find our rank
    our_rank = next(
        i for i, (k, _) in enumerate(sorted_entries, 1) if k == our_key
    )
    print(f"Our rank: {our_rank} out of {len(entry_metrics)}")
    print()
    
    # Identify entries to study (top performers that aren't us)
    print("Top Actual Entries to Study:")
    print("-" * 80)
    
    top_actual = [
        (k, m) for k, m in sorted_entries
        if not k.startswith('our_')
    ][:5]
    
    for i, (entry_key, metrics) in enumerate(top_actual, 1):
        print(f"{i}. {entry_key}")
        print(f"   Normalized payout: {metrics['normalized_avg_payout']:.3f}")
        print(f"   Average payout: {metrics['avg_payout']:.2f} points")
        print(f"   Percentile: {metrics['percentile_rank']:.1f}%")
        
        # Show their portfolio
        entry_data = entry_bids[entry_bids['entry_key'] == entry_key]
        print(f"   Teams: {len(entry_data)}")
        print(f"   Total bid: ${entry_data['bid_amount'].sum():.0f}")
        
        # Top 3 bids
        top_bids = entry_data.nlargest(3, 'bid_amount')
        print("   Top bids:")
        for _, row in top_bids.iterrows():
            team = row['team_key'].split(':')[-1].replace('-', ' ').title()
            print(f"     {team}: ${row['bid_amount']:.0f}")
        print()
    
    return entry_metrics


if __name__ == "__main__":
    years = [2024, 2023, 2022, 2021]
    
    all_results = {}
    for year in years:
        try:
            metrics = calculate_entry_performance_metrics(year, "minlp")
            all_results[year] = metrics
            print()
            print()
        except Exception as e:
            print(f"Error processing {year}: {e}")
            import traceback
            traceback.print_exc()
            print()
    
    # Summary across all years
    print("=" * 80)
    print("SUMMARY ACROSS ALL YEARS")
    print("=" * 80)
    print()
    
    print(f"{'Year':<6s} {'Norm Payout':>12s} {'Percentile':>11s} "
          f"{'Rank':>6s}")
    print("-" * 80)
    
    for year in years:
        if year not in all_results:
            continue
        
        our_metrics = all_results[year].get(f'our_minlp')
        if not our_metrics:
            continue
        
        # Calculate rank
        sorted_entries = sorted(
            all_results[year].items(),
            key=lambda x: x[1]['normalized_avg_payout'],
            reverse=True
        )
        our_rank = next(
            i for i, (k, _) in enumerate(sorted_entries, 1)
            if k == 'our_minlp'
        )
        total_entries = len(all_results[year])
        
        print(f"{year:<6d} {our_metrics['normalized_avg_payout']:>12.3f} "
              f"{our_metrics['percentile_rank']:>10.1f}% "
              f"{our_rank:>3d}/{total_entries:<3d}")
    
    print()
    
    # Average performance
    avg_norm_payout = np.mean([
        all_results[y]['our_minlp']['normalized_avg_payout']
        for y in years if y in all_results
    ])
    avg_percentile = np.mean([
        all_results[y]['our_minlp']['percentile_rank']
        for y in years if y in all_results
    ])
    
    print(f"Average normalized payout: {avg_norm_payout:.3f}")
    print(f"Average percentile rank: {avg_percentile:.1f}%")
