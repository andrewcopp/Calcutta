"""
Test script for detailed investment report.
"""
from pathlib import Path
import pandas as pd
from moneyball.models.detailed_investment_report import generate_detailed_investment_report

# Load 2025 data
year = "2025"
year_dir = Path(f"out/{year}")

# Load base data
teams = pd.read_parquet(year_dir / "teams.parquet")
entry_bids = pd.read_parquet(year_dir / "entry_bids.parquet")
entries = pd.read_parquet(year_dir / "entries.parquet")

# Load latest run artifacts
run_dir = Path(f"out/{year}/derived/calcutta")
latest_run = sorted(run_dir.glob("*"))[-1]

recommended_bids = pd.read_parquet(latest_run / "recommended_entry_bids.parquet")
predicted_market = pd.read_parquet(latest_run / "predicted_auction_share_of_pool.parquet")

# Load tournament value from new artifacts
tournament_value_path = Path(f"out/{year}/derived/artifacts_v2/tournament_value.parquet")
if tournament_value_path.exists():
    tournament_value = pd.read_parquet(tournament_value_path)
else:
    # Fallback: calculate from predicted_game_outcomes
    from moneyball.models.tournament_value import generate_tournament_value
    predicted_games = pd.read_parquet(year_dir / "derived/predicted_game_outcomes.parquet")
    tournament_value, _ = generate_tournament_value(
        predicted_game_outcomes=predicted_games
    )

# Calculate predicted total pool
predicted_total_pool = len(entries) * 100  # Each entry has 100 points

# Generate report
report = generate_detailed_investment_report(
    teams=teams,
    recommended_entry_bids=recommended_bids,
    predicted_auction_share_of_pool=predicted_market,
    entry_bids=entry_bids,
    predicted_total_pool_bids_points=predicted_total_pool,
    tournament_value=tournament_value,
)

print(f"DETAILED INVESTMENT REPORT - {year}")
print("=" * 100)
print()

# Show summary stats
print("Summary Statistics:")
print(f"  Total teams: {len(report)}")
print(f"  Teams in our portfolio: {(report['our_bid'] > 0).sum()}")
print(f"  Total bid: {report['our_bid'].sum()}")
print()

# Show our portfolio
print("Our Portfolio (teams with bids > 0):")
print("-" * 110)
portfolio = report[report['our_bid'] > 0].copy()
print(f"{'Team':<30s} {'Seed':>4s} {'Exp Pts':>8s} {'Exp Mkt':>8s} {'Exp ROI':>8s} {'Our Bid':>8s} {'Our ROI':>8s} {'Act Mkt':>8s} {'Act ROI':>8s}")
print("-" * 110)
for _, row in portfolio.iterrows():
    print(f"{row['school_name'][:29]:<30s} "
          f"{int(row['seed']):>4d} "
          f"{row['expected_points']:>8.1f} "
          f"{row['expected_market']:>8.1f} "
          f"{row['expected_roi']:>8.2f} "
          f"{int(row['our_bid']):>8d} "
          f"{row['our_roi']:>8.2f} "
          f"{row['actual_market']:>8.1f} "
          f"{row['actual_roi']:>8.2f}")
print()

# Show top 10 by expected ROI
print("Top 10 Teams by Expected ROI:")
print("-" * 100)
top_roi = report.nlargest(10, 'expected_roi')
print(f"{'Team':<30s} {'Seed':>4s} {'Exp Pts':>8s} {'Exp Mkt':>8s} {'Exp ROI':>8s} {'Our Bid':>8s}")
print("-" * 100)
for _, row in top_roi.iterrows():
    bid_marker = " *" if row['our_bid'] > 0 else ""
    print(f"{row['school_name'][:29]:<30s} "
          f"{int(row['seed']):>4d} "
          f"{row['expected_points']:>8.1f} "
          f"{row['expected_market']:>8.1f} "
          f"{row['expected_roi']:>8.2f} "
          f"{int(row['our_bid']):>8d}{bid_marker}")
print()

# Show top 10 by actual ROI
print("Top 10 Teams by Actual ROI:")
print("-" * 100)
top_actual_roi = report[report['actual_market'] > 0].nlargest(10, 'actual_roi')
print(f"{'Team':<30s} {'Seed':>4s} {'Exp Pts':>8s} {'Act Mkt':>8s} {'Act ROI':>8s} {'Our Bid':>8s}")
print("-" * 100)
for _, row in top_actual_roi.iterrows():
    bid_marker = " *" if row['our_bid'] > 0 else ""
    print(f"{row['school_name'][:29]:<30s} "
          f"{int(row['seed']):>4d} "
          f"{row['expected_points']:>8.1f} "
          f"{row['actual_market']:>8.1f} "
          f"{row['actual_roi']:>8.2f} "
          f"{int(row['our_bid']):>8d}{bid_marker}")
print()

# Save report as CSV
output_path = latest_run / "detailed_investment_report.csv"
report.to_csv(output_path, index=False)
print(f"Report saved to: {output_path}")
