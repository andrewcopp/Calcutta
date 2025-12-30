"""
Regenerate investment reports for all years using max-min strategy.
"""
from pathlib import Path
import pandas as pd
from moneyball.models.recommended_entry_bids import recommend_entry_bids
from moneyball.models.investment_report import generate_investment_report
from moneyball.models.simulated_entry_outcomes import simulate_entry_outcomes

YEARS = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]

print("REGENERATING INVESTMENT REPORTS WITH MAX-MIN STRATEGY")
print("="*90)
print()

for year in YEARS:
    print(f"Processing {year}...")
    
    year_dir = Path(f"out/{year}")
    
    # Check if required files exist
    required_files = [
        year_dir / "derived/predicted_game_outcomes.parquet",
        year_dir / "entries.parquet",
    ]
    
    missing = [f for f in required_files if not f.exists()]
    if missing:
        print(f"  Skipping {year}: missing files")
        continue
    
    # Load data
    try:
        predicted_games = pd.read_parquet(
            year_dir / "derived/predicted_game_outcomes.parquet"
        )
        entries = pd.read_parquet(year_dir / "entries.parquet")
        
        # Load or find predicted market share
        calcutta_dir = year_dir / "derived/calcutta"
        if not calcutta_dir.exists():
            print(f"  Skipping {year}: no calcutta directory")
            continue
            
        latest_run = sorted(calcutta_dir.glob("*"))
        if not latest_run:
            print(f"  Skipping {year}: no runs found")
            continue
        latest_run = latest_run[-1]
        
        predicted_market = pd.read_parquet(
            latest_run / "predicted_auction_share_of_pool.parquet"
        )
        
    except Exception as e:
        print(f"  Skipping {year}: error loading data: {e}")
        continue
    
    # Calculate predicted total pool
    predicted_total_pool = len(entries) * 100
    
    # Generate recommended bids with max-min strategy
    try:
        recommended_bids = recommend_entry_bids(
            predicted_auction_share_of_pool=predicted_market,
            predicted_game_outcomes=predicted_games,
            predicted_total_pool_bids_points=predicted_total_pool,
            budget_points=100,
            min_teams=3,
            max_teams=10,
            max_per_team_points=50,
            min_bid_points=1,
            strategy="maxmin",
        )
        
        print(f"  Generated recommended bids: {len(recommended_bids)} teams")
        
        # Simulate entry outcomes
        simulated_outcomes = simulate_entry_outcomes(
            recommended_entry_bids=recommended_bids,
            predicted_game_outcomes=predicted_games,
            n_sims=10000,
            seed=42,
        )
        
        print(f"  Simulated entry outcomes: {len(simulated_outcomes)} simulations")
        
        # Generate investment report
        report = generate_investment_report(
            recommended_entry_bids=recommended_bids,
            simulated_entry_outcomes=simulated_outcomes,
            predicted_game_outcomes=predicted_games,
            predicted_auction_share_of_pool=predicted_market,
            predicted_total_pool_bids_points=predicted_total_pool,
            strategy="maxmin",
        )
        
        # Display key metrics
        print(f"  P(1st): {report['p_top1'].iloc[0]:.3f}")
        print(f"  P(Money): {report['p_in_money'].iloc[0]:.3f}")
        print(f"  Expected Points: {report['mean_expected_points'].iloc[0]:.1f}")
        print(f"  Expected Finish: {report['mean_expected_finish'].iloc[0]:.1f}")
        print()
        
    except Exception as e:
        print(f"  Error generating report: {e}")
        import traceback
        traceback.print_exc()
        continue

print()
print("="*90)
print("DONE")
