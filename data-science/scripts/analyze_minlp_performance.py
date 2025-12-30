"""
Analyze MINLP performance to identify potential improvements.

This script examines:
1. Which tournament scenarios MINLP wins vs loses
2. What teams would have helped in losing scenarios
3. Whether there are systematic patterns we're missing
"""
from pathlib import Path
import pandas as pd
import numpy as np

POINTS_PER_WIN = {0: 0, 1: 1, 2: 2, 3: 4, 4: 8, 5: 16, 6: 32}


def analyze_minlp_performance(year: int):
    """Analyze MINLP performance for a given year."""
    year_dir = Path(f"out/{year}")
    
    # Load data
    entry_bids = pd.read_parquet(year_dir / "entry_bids.parquet")
    tournaments = pd.read_parquet(year_dir / "derived/tournaments.parquet")
    
    # Get our MINLP bids
    calcutta_dir = year_dir / "derived/calcutta"
    latest_run = sorted(calcutta_dir.glob("20251230T193*"))[-1]
    our_bids = pd.read_parquet(latest_run / "recommended_entry_bids.parquet")
    
    # Calculate actual markets (excluding our bid)
    actual_markets = entry_bids.groupby('team_key')[
        'bid_amount'
    ].sum().to_dict()
    
    # Add points to tournaments
    tournaments['points'] = tournaments['wins'].map(POINTS_PER_WIN)
    
    # Calculate our points in each simulation
    print(f"Analyzing {year} MINLP performance...")
    print("=" * 80)
    print()
    
    print("Our Portfolio:")
    print("-" * 80)
    for _, row in our_bids.iterrows():
        team = row['team_key'].split(':')[-1].replace('-', ' ').title()
        bid = row['bid_amount_points']
        market = actual_markets.get(row['team_key'], 0)
        ownership = bid / (market + bid) * 100
        print(f"  {team:<30s} ${bid:>3.0f}  (market: ${market:>4.0f}, "
              f"own: {ownership:>4.1f}%)")
    print()
    
    # Calculate points per simulation
    sim_ids = tournaments['sim_id'].unique()
    our_points = []
    
    for sim_id in sim_ids:
        sim_results = tournaments[tournaments['sim_id'] == sim_id]
        
        total_points = 0
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
            total_points += ownership * team_points
        
        our_points.append(total_points)
    
    our_points = np.array(our_points)
    
    # Calculate other entries' points
    other_entries = {}
    for entry_key in entry_bids['entry_key'].unique():
        entry_data = entry_bids[entry_bids['entry_key'] == entry_key]
        entry_points = []
        
        for sim_id in sim_ids:
            sim_results = tournaments[tournaments['sim_id'] == sim_id]
            
            total_points = 0
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
                total_points += ownership * team_points
            
            entry_points.append(total_points)
        
        other_entries[entry_key] = np.array(entry_points)
    
    # Find scenarios where we win
    wins = 0
    win_sims = []
    loss_sims = []
    
    for i, our_pts in enumerate(our_points):
        beat_all = True
        for entry_pts in other_entries.values():
            if entry_pts[i] >= our_pts:
                beat_all = False
                break
        
        if beat_all:
            wins += 1
            win_sims.append(i)
        else:
            loss_sims.append(i)
    
    p_first = wins / len(sim_ids)
    
    print(f"Performance Summary:")
    print(f"  P(1st): {p_first:.2%}")
    print(f"  Wins: {wins}/{len(sim_ids)} simulations")
    print(f"  Mean points: {our_points.mean():.1f}")
    print(f"  Median points: {np.median(our_points):.1f}")
    print()
    
    # Analyze winning vs losing scenarios
    print("Scenario Analysis:")
    print("-" * 80)
    
    # In winning scenarios, which teams contributed most?
    win_contributions = {}
    for sim_id in win_sims[:100]:  # Sample first 100 wins
        sim_results = tournaments[tournaments['sim_id'] == sim_id]
        
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
            contribution = ownership * team_points
            
            if team_key not in win_contributions:
                win_contributions[team_key] = []
            win_contributions[team_key].append(contribution)
    
    print("Top contributors in WINNING scenarios:")
    avg_contributions = {
        k: np.mean(v) for k, v in win_contributions.items()
    }
    for team_key in sorted(avg_contributions, key=avg_contributions.get,
                          reverse=True)[:5]:
        team = team_key.split(':')[-1].replace('-', ' ').title()
        avg_contrib = avg_contributions[team_key]
        print(f"  {team:<30s} {avg_contrib:>6.1f} pts/sim")
    print()
    
    # In losing scenarios, what went wrong?
    print("Analysis of LOSING scenarios (sample of 100):")
    loss_margins = []
    for sim_id in loss_sims[:100]:
        max_other = max(other_entries[k][sim_id]
                       for k in other_entries.keys())
        margin = max_other - our_points[sim_id]
        loss_margins.append(margin)
    
    print(f"  Average loss margin: {np.mean(loss_margins):.1f} points")
    print(f"  Median loss margin: {np.median(loss_margins):.1f} points")
    print(f"  Close losses (<10 pts): "
          f"{sum(1 for m in loss_margins if m < 10)}/100")
    print()
    
    # Key insight: Are we spread too thin or too concentrated?
    print("Portfolio Characteristics:")
    print("-" * 80)
    total_bid = our_bids['bid_amount_points'].sum()
    num_teams = len(our_bids)
    avg_bid = total_bid / num_teams
    
    print(f"  Total bid: ${total_bid:.0f}")
    print(f"  Number of teams: {num_teams}")
    print(f"  Average bid per team: ${avg_bid:.1f}")
    print(f"  Min bid: ${our_bids['bid_amount_points'].min():.0f}")
    print(f"  Max bid: ${our_bids['bid_amount_points'].max():.0f}")
    print()
    
    # Calculate concentration
    bids_array = our_bids['bid_amount_points'].values
    herfindahl = (bids_array / bids_array.sum()) ** 2
    concentration = herfindahl.sum()
    
    print(f"  Concentration (Herfindahl): {concentration:.3f}")
    print(f"    (1.0 = all in one team, 1/{num_teams}={1/num_teams:.3f} = "
          f"perfectly equal)")
    print()


if __name__ == "__main__":
    for year in [2024, 2023, 2022]:
        analyze_minlp_performance(year)
        print()
        print()
