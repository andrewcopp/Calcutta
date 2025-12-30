"""
Find the optimal entry allocation given actual market and tournament outcomes.

This script performs post-hoc optimization to find the theoretical maximum
performance achievable with perfect knowledge of:
1. Actual market (what everyone else bid)
2. Actual tournament outcomes

This establishes a performance ceiling to compare our strategies against.
"""
from pathlib import Path
import pandas as pd
import numpy as np
from scipy.optimize import differential_evolution


POINTS_PER_WIN = {
    0: 0,
    1: 1,
    2: 2,
    3: 4,
    4: 8,
    5: 16,
    6: 32,
}


def calculate_entry_performance(
    bids: np.ndarray,
    team_keys: np.ndarray,
    actual_markets: pd.Series,
    tournaments: pd.DataFrame,
    other_entries: pd.DataFrame,
) -> dict:
    """
    Calculate P(1st) for a given entry allocation across all simulations.
    
    Args:
        bids: Array of bid amounts for each team
        team_keys: Array of team keys corresponding to bids
        actual_markets: Series of actual market totals (excluding our bid)
        tournaments: DataFrame with simulated tournament outcomes (wins)
        other_entries: DataFrame with other entries' bids
        
    Returns:
        Dictionary with performance metrics including P(1st)
    """
    # Create bid DataFrame
    our_bids = pd.DataFrame({
        'team_key': team_keys,
        'bid_amount': bids,
    })
    our_bids = our_bids[our_bids['bid_amount'] > 0]
    
    if len(our_bids) == 0:
        return {'p_first': 0.0, 'total_bid': 0.0, 'num_teams': 0}
    
    # Calculate points for each team in each simulation
    tournaments_with_points = tournaments.copy()
    tournaments_with_points['points'] = tournaments_with_points['wins'].map(
        POINTS_PER_WIN
    )
    
    # For each simulation, calculate our total points
    sim_ids = tournaments['sim_id'].unique()
    our_points_per_sim = []
    
    for sim_id in sim_ids:
        sim_results = tournaments_with_points[
            tournaments_with_points['sim_id'] == sim_id
        ]
        
        # Calculate our points
        our_sim_points = 0
        for _, bid_row in our_bids.iterrows():
            team_key = bid_row['team_key']
            bid = bid_row['bid_amount']
            
            # Get team's points in this sim
            team_points = sim_results[
                sim_results['team_key'] == team_key
            ]['points'].values
            if len(team_points) == 0:
                continue
            team_points = team_points[0]
            
            # Calculate ownership
            market = actual_markets.get(team_key, 0)
            ownership = bid / (market + bid)
            
            # Add to our points
            our_sim_points += ownership * team_points
        
        our_points_per_sim.append(our_sim_points)
    
    # Calculate other entries' points per sim
    other_points_per_sim = {}
    for entry_key in other_entries['entry_key'].unique():
        entry_bids = other_entries[
            other_entries['entry_key'] == entry_key
        ]
        entry_points = []
        
        for sim_id in sim_ids:
            sim_results = tournaments_with_points[
                tournaments_with_points['sim_id'] == sim_id
            ]
            
            entry_sim_points = 0
            for _, bid_row in entry_bids.iterrows():
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
                entry_sim_points += ownership * team_points
            
            entry_points.append(entry_sim_points)
        
        other_points_per_sim[entry_key] = entry_points
    
    # Calculate P(1st): fraction of sims where we beat all other entries
    wins = 0
    for i, our_pts in enumerate(our_points_per_sim):
        beat_all = True
        for entry_pts in other_points_per_sim.values():
            if entry_pts[i] >= our_pts:
                beat_all = False
                break
        if beat_all:
            wins += 1
    
    p_first = wins / len(sim_ids)
    
    return {
        'p_first': p_first,
        'total_bid': bids.sum(),
        'num_teams': (bids > 0).sum(),
        'mean_points': np.mean(our_points_per_sim),
    }


def optimize_entry_for_p1st(
    actual_entry_bids: pd.DataFrame,
    simulated_tournaments: pd.DataFrame,
    budget: float = 100.0,
    min_teams: int = 3,
    max_teams: int = 10,
    max_bid: float = 50.0,
    min_bid: float = 1.0,
) -> pd.DataFrame:
    """
    Find the optimal entry allocation that maximizes P(1st) given actual market.
    
    Args:
        actual_entry_bids: Actual bids from all entries
        simulated_tournaments: Simulated tournament outcomes
        budget: Total budget
        min_teams: Minimum number of teams
        max_teams: Maximum number of teams
        max_bid: Maximum bid per team
        min_bid: Minimum bid per team
        
    Returns:
        DataFrame with optimal bids
    """
    # Get unique teams
    teams = actual_entry_bids['team_key'].unique()
    n_teams = len(teams)
    
    # Calculate actual market for each team (sum of all bids)
    actual_markets = actual_entry_bids.groupby('team_key')['bid_amount'].sum().reindex(teams, fill_value=0).values
    
    # Calculate actual points for each team from tournament outcomes
    # For now, use expected points from simulated tournaments
    team_points = simulated_tournaments.groupby('team_key')['points'].mean().reindex(teams, fill_value=0).values
    
    # Calculate win probability for each team
    total_sims = simulated_tournaments['sim_id'].nunique()
    team_wins = simulated_tournaments[simulated_tournaments['points'] >= 192].groupby('team_key').size().reindex(teams, fill_value=0).values / total_sims
    
    print(f"Optimizing for {n_teams} teams")
    print(f"Actual market range: ${actual_markets.min():.0f} - ${actual_markets.max():.0f}")
    print(f"Team win probabilities range: {team_wins.min():.4f} - {team_wins.max():.4f}")
    
    # Objective: maximize P(1st)
    # We'll use a simplified approach: maximize sum of (ownership * win_probability)
    def objective(x):
        # x is a vector of bids for each team
        bids = x[:n_teams]
        
        # Calculate ownership
        ownership = bids / (actual_markets + bids + 1e-9)
        
        # Calculate P(1st) as sum of (ownership * win_probability)
        p_first = (ownership * team_wins).sum()
        
        # Negative because we're minimizing
        return -p_first
    
    # Constraints
    def budget_constraint(x):
        return budget - x[:n_teams].sum()
    
    def min_teams_constraint(x):
        return (x[:n_teams] >= min_bid).sum() - min_teams
    
    def max_teams_constraint(x):
        return max_teams - (x[:n_teams] >= min_bid).sum()
    
    # Bounds: each bid between 0 and max_bid
    bounds = [(0, max_bid) for _ in range(n_teams)]
    
    # Use differential evolution for global optimization
    result = differential_evolution(
        objective,
        bounds,
        constraints=[
            {'type': 'eq', 'fun': budget_constraint},
            {'type': 'ineq', 'fun': min_teams_constraint},
            {'type': 'ineq', 'fun': max_teams_constraint},
        ],
        seed=42,
        maxiter=1000,
        atol=0.01,
        tol=0.01,
        workers=1,
    )
    
    if not result.success:
        print(f"Warning: Optimization did not converge: {result.message}")
    
    # Extract optimal bids
    optimal_bids = result.x[:n_teams]
    
    # Round to nearest dollar and ensure constraints
    optimal_bids = np.round(optimal_bids)
    optimal_bids[optimal_bids < min_bid] = 0
    optimal_bids[optimal_bids > max_bid] = max_bid
    
    # Adjust to meet budget exactly
    total_bid = optimal_bids.sum()
    if total_bid > budget:
        # Scale down
        scale = budget / total_bid
        optimal_bids = np.floor(optimal_bids * scale)
    elif total_bid < budget:
        # Add remaining budget to highest ROI teams
        remaining = budget - optimal_bids.sum()
        ownership = optimal_bids / (actual_markets + optimal_bids + 1e-9)
        roi = team_wins / (ownership + 1e-9)
        sorted_idx = np.argsort(-roi)
        for i in range(int(remaining)):
            idx = sorted_idx[i % len(sorted_idx)]
            if optimal_bids[idx] < max_bid:
                optimal_bids[idx] += 1
    
    # Create result DataFrame
    result_df = pd.DataFrame({
        'team_key': teams,
        'optimal_bid': optimal_bids,
        'actual_market': actual_markets,
        'win_probability': team_wins,
        'ownership': optimal_bids / (actual_markets + optimal_bids + 1e-9),
    })
    
    # Filter to only teams with bids
    result_df = result_df[result_df['optimal_bid'] > 0].sort_values('optimal_bid', ascending=False)
    
    return result_df


if __name__ == "__main__":
    # Test on 2024 data
    year = 2024
    year_dir = Path(f"out/{year}")
    
    print(f"Finding optimal entry for {year}")
    print("=" * 80)
    print()
    
    # Load actual entry bids
    entry_bids = pd.read_parquet(year_dir / "entry_bids.parquet")
    
    # Load simulated tournaments
    tournaments = pd.read_parquet(year_dir / "derived/tournaments.parquet")
    
    # Find optimal entry
    optimal = optimize_entry_for_p1st(
        actual_entry_bids=entry_bids,
        simulated_tournaments=tournaments,
    )
    
    print()
    print("OPTIMAL ENTRY ALLOCATION:")
    print("-" * 80)
    print(f"{'Team':<20s} {'Bid':>8s} {'Market':>10s} {'Win Prob':>10s} {'Ownership':>10s}")
    print("-" * 80)
    
    for _, row in optimal.iterrows():
        print(f"{row['team_key']:<20s} "
              f"${row['optimal_bid']:>7.0f} "
              f"${row['actual_market']:>9.0f} "
              f"{row['win_probability']:>9.2%} "
              f"{row['ownership']:>9.2%}")
    
    print("-" * 80)
    print(f"Total bid: ${optimal['optimal_bid'].sum():.0f}")
    print(f"Number of teams: {len(optimal)}")
    print(f"Estimated P(1st): {(optimal['ownership'] * optimal['win_probability']).sum():.2%}")
