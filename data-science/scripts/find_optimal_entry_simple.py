"""
Find the theoretical maximum P(1st) given actual market distribution.

Simplified approach: Use win probabilities from simulations and optimize
for maximum weighted ownership.
"""
from pathlib import Path
import pandas as pd
import numpy as np
from scipy.optimize import differential_evolution


POINTS_PER_WIN = {0: 0, 1: 1, 2: 2, 3: 4, 4: 8, 5: 16, 6: 32}


def optimize_for_max_p1st(
    actual_entry_bids: pd.DataFrame,
    tournaments: pd.DataFrame,
    budget: float = 100.0,
    min_teams: int = 3,
    max_teams: int = 10,
    max_bid: float = 50.0,
    min_bid: float = 1.0,
) -> pd.DataFrame:
    """
    Find optimal entry that maximizes P(1st) given actual market.
    """
    # Get unique teams
    teams = sorted(actual_entry_bids['team_key'].unique())
    n_teams = len(teams)
    
    # Calculate actual market for each team (sum of all actual bids)
    actual_markets = actual_entry_bids.groupby('team_key')[
        'bid_amount'
    ].sum().reindex(teams, fill_value=0)
    
    # Calculate win probability for each team from tournaments
    tournaments_with_points = tournaments.copy()
    tournaments_with_points['points'] = tournaments_with_points[
        'wins'
    ].map(POINTS_PER_WIN)
    
    # A team "wins" if they get 192+ points (champion = 63 points)
    # Actually, let's use a more nuanced approach: weight by points
    total_sims = tournaments['sim_id'].nunique()
    
    # Calculate average points per team
    avg_points = tournaments_with_points.groupby('team_key')[
        'points'
    ].mean().reindex(teams, fill_value=0)
    
    # Calculate championship probability (6 wins = champion)
    champ_prob = tournaments[
        tournaments['wins'] == 6
    ].groupby('team_key').size().reindex(teams, fill_value=0) / total_sims
    
    print(f"Optimizing for {n_teams} teams")
    print(f"Actual market range: ${actual_markets.min():.0f} - "
          f"${actual_markets.max():.0f}")
    print(f"Championship prob range: {champ_prob.min():.4f} - "
          f"{champ_prob.max():.4f}")
    print()
    
    # Objective: maximize expected ownership-weighted championship prob
    # with penalty for constraint violations
    def objective(x):
        bids = x[:n_teams]
        ownership = bids / (actual_markets.values + bids + 1e-9)
        value = (ownership * champ_prob.values).sum()
        
        # Penalties for constraint violations
        penalty = 0
        
        # Budget constraint (should sum to budget)
        budget_diff = abs(bids.sum() - budget)
        penalty += budget_diff * 1000
        
        # Team count constraints
        num_teams_selected = (bids >= min_bid).sum()
        if num_teams_selected < min_teams:
            penalty += (min_teams - num_teams_selected) * 1000
        if num_teams_selected > max_teams:
            penalty += (num_teams_selected - max_teams) * 1000
        
        return -value + penalty
    
    # Bounds
    bounds = [(0, max_bid) for _ in range(n_teams)]
    
    # Optimize
    print("Running optimization...")
    result = differential_evolution(
        objective,
        bounds,
        seed=42,
        maxiter=500,
        atol=0.01,
        tol=0.01,
        workers=1,
        disp=True,
    )
    
    if not result.success:
        print(f"Warning: {result.message}")
    
    # Extract and round bids
    optimal_bids = np.round(result.x[:n_teams])
    optimal_bids[optimal_bids < min_bid] = 0
    optimal_bids[optimal_bids > max_bid] = max_bid
    
    # Adjust to meet budget
    total_bid = optimal_bids.sum()
    if total_bid > budget:
        scale = budget / total_bid
        optimal_bids = np.floor(optimal_bids * scale)
    elif total_bid < budget:
        remaining = int(budget - optimal_bids.sum())
        ownership = optimal_bids / (actual_markets.values + optimal_bids + 1e-9)
        marginal_value = champ_prob.values / (ownership + 1e-9)
        sorted_idx = np.argsort(-marginal_value)
        for i in range(remaining):
            idx = sorted_idx[i % len(sorted_idx)]
            if optimal_bids[idx] < max_bid:
                optimal_bids[idx] += 1
    
    # Create result DataFrame
    result_df = pd.DataFrame({
        'team_key': teams,
        'optimal_bid': optimal_bids,
        'actual_market': actual_markets.values,
        'champ_prob': champ_prob.values,
        'avg_points': avg_points.values,
    })
    
    result_df['ownership'] = result_df['optimal_bid'] / (
        result_df['actual_market'] + result_df['optimal_bid'] + 1e-9
    )
    
    result_df = result_df[result_df['optimal_bid'] > 0].sort_values(
        'optimal_bid', ascending=False
    )
    
    return result_df


if __name__ == "__main__":
    year = 2024
    year_dir = Path(f"out/{year}")
    
    print(f"THEORETICAL MAXIMUM ANALYSIS - {year}")
    print("=" * 80)
    print()
    
    # Load data
    entry_bids = pd.read_parquet(year_dir / "entry_bids.parquet")
    tournaments = pd.read_parquet(year_dir / "derived/tournaments.parquet")
    
    # Find optimal
    optimal = optimize_for_max_p1st(
        actual_entry_bids=entry_bids,
        tournaments=tournaments,
    )
    
    print()
    print("THEORETICAL OPTIMAL ENTRY:")
    print("-" * 80)
    print(f"{'Team':<30s} {'Bid':>6s} {'Market':>8s} {'P(Champ)':>9s} "
          f"{'Own%':>7s}")
    print("-" * 80)
    
    for _, row in optimal.iterrows():
        team = row['team_key'].split(':')[-1].replace('-', ' ').title()
        print(f"{team:<30s} ${row['optimal_bid']:>5.0f} "
              f"${row['actual_market']:>7.0f} "
              f"{row['champ_prob']:>8.2%} {row['ownership']:>6.1%}")
    
    print("-" * 80)
    print(f"Total bid: ${optimal['optimal_bid'].sum():.0f}")
    print(f"Teams: {len(optimal)}")
    
    # Calculate theoretical max P(1st)
    theoretical_p1st = (optimal['ownership'] * optimal['champ_prob']).sum()
    print(f"Theoretical Max P(1st): {theoretical_p1st:.2%}")
    print()
    
    # Load our MINLP result for comparison
    calcutta_dir = year_dir / "derived/calcutta"
    latest_run = sorted(calcutta_dir.glob("20251230T193*"))[-1]
    our_report = pd.read_parquet(latest_run / "investment_report.parquet")
    our_p1st = our_report['p_top1'].iloc[0]
    
    print(f"Our MINLP P(1st): {our_p1st:.2%}")
    print(f"Gap to theoretical max: {theoretical_p1st - our_p1st:.2%}")
    print(f"Efficiency: {our_p1st / theoretical_p1st:.1%}")
