"""
Generate detailed investment reports for all years.
"""
from pathlib import Path
import pandas as pd
from moneyball.models.detailed_investment_report import generate_detailed_investment_report

# Years to process
years = ['2017', '2018', '2019', '2021', '2022', '2023', '2024', '2025']

print('GENERATING DETAILED INVESTMENT REPORTS FOR ALL YEARS')
print('=' * 90)
print()

for year in years:
    print(f'Processing {year}...')
    
    year_dir = Path(f"out/{year}")
    
    # Check if required files exist
    if not (year_dir / "teams.parquet").exists():
        print(f'  ✗ Skipping {year} - teams.parquet not found')
        continue
    
    if not (year_dir / "entry_bids.parquet").exists():
        print(f'  ✗ Skipping {year} - entry_bids.parquet not found')
        continue
    
    # Load base data
    teams = pd.read_parquet(year_dir / "teams.parquet")
    entry_bids = pd.read_parquet(year_dir / "entry_bids.parquet")
    entries = pd.read_parquet(year_dir / "entries.parquet")
    
    # Load latest run artifacts
    run_dir = Path(f"out/{year}/derived/calcutta")
    if not run_dir.exists():
        print(f'  ✗ Skipping {year} - no calcutta runs found')
        continue
    
    runs = sorted(run_dir.glob("*"))
    if not runs:
        print(f'  ✗ Skipping {year} - no calcutta runs found')
        continue
    
    latest_run = runs[-1]
    
    rec_bids_path = latest_run / "recommended_entry_bids.parquet"
    pred_market_path = latest_run / "predicted_auction_share_of_pool.parquet"
    
    if not rec_bids_path.exists() or not pred_market_path.exists():
        print(f'  ✗ Skipping {year} - missing artifacts')
        continue
    
    recommended_bids = pd.read_parquet(rec_bids_path)
    predicted_market = pd.read_parquet(pred_market_path)
    
    # Load tournament value
    tournament_value_path = Path(f"out/{year}/derived/artifacts_v2/tournament_value.parquet")
    if tournament_value_path.exists():
        tournament_value = pd.read_parquet(tournament_value_path)
    else:
        # Fallback: calculate from predicted_game_outcomes
        from moneyball.models.tournament_value import generate_tournament_value
        pred_games_path = year_dir / "derived/predicted_game_outcomes.parquet"
        if not pred_games_path.exists():
            print(f'  ✗ Skipping {year} - no tournament value data')
            continue
        predicted_games = pd.read_parquet(pred_games_path)
        tournament_value, _ = generate_tournament_value(
            predicted_game_outcomes=predicted_games
        )
    
    # Calculate predicted total pool
    predicted_total_pool = len(entries) * 100
    
    # Generate report
    try:
        report = generate_detailed_investment_report(
            teams=teams,
            recommended_entry_bids=recommended_bids,
            predicted_auction_share_of_pool=predicted_market,
            entry_bids=entry_bids,
            predicted_total_pool_bids_points=predicted_total_pool,
            tournament_value=tournament_value,
        )
        
        # Save as CSV
        output_path = latest_run / "detailed_investment_report.csv"
        report.to_csv(output_path, index=False)
        
        # Print summary
        portfolio_teams = (report['our_bid'] > 0).sum()
        total_bid = report['our_bid'].sum()
        
        print(f'  ✓ Generated report: {output_path.name}')
        print(f'    Teams: {len(report)}, Portfolio: {portfolio_teams}, Total bid: {total_bid}')
        
    except Exception as e:
        print(f'  ✗ Error generating report: {e}')
    
    print()

print('=' * 90)
print('DONE')
