"""
Compare greedy vs MINLP strategies across all historical years.

Runs both strategies on each year and compares performance metrics.
"""
from pathlib import Path
import pandas as pd
from moneyball.models.recommended_entry_bids import (
    expected_team_points_from_predicted_game_outcomes,
)
from moneyball.models.portfolio_strategies import get_strategy

# Years to test
YEARS = [2017, 2018, 2019, 2020, 2021, 2022, 2023, 2024, 2025]

# Parameters
BUDGET = 100
MIN_TEAMS = 3
MAX_TEAMS = 10
MAX_PER_TEAM = 50
MIN_BID = 1

results = []

for year in YEARS:
    print(f"\n{'='*90}")
    print(f"YEAR {year}")
    print(f"{'='*90}")
    
    year_dir = Path(f"out/{year}")
    run_dir = Path(f"out/{year}/derived/calcutta")
    
    # Check if data exists
    if not year_dir.exists():
        print(f"  Skipping {year}: no data directory")
        continue
    
    if not run_dir.exists():
        print(f"  Skipping {year}: no run directory")
        continue
    
    latest_run = sorted(run_dir.glob("*"))
    if not latest_run:
        print(f"  Skipping {year}: no runs found")
        continue
    latest_run = latest_run[-1]
    
    # Load data
    try:
        predicted_games_path = year_dir / "derived/predicted_game_outcomes.parquet"
        if not predicted_games_path.exists():
            print(f"  Skipping {year}: no predicted_game_outcomes")
            continue
            
        predicted_games = pd.read_parquet(predicted_games_path)
        predicted_market = pd.read_parquet(
            latest_run / "predicted_auction_share_of_pool.parquet"
        )
        entries = pd.read_parquet(year_dir / "entries.parquet")
    except Exception as e:
        print(f"  Skipping {year}: error loading data: {e}")
        continue
    
    # Calculate expected points
    exp_pts = expected_team_points_from_predicted_game_outcomes(
        predicted_game_outcomes=predicted_games
    )
    
    # Merge data
    df = predicted_market.merge(exp_pts, on="team_key", how="left")
    df["expected_team_points"] = df["expected_team_points"].fillna(0.0)
    
    # Calculate predicted total bids
    predicted_total_pool = len(entries) * 100
    df["predicted_team_total_bids"] = (
        df["predicted_auction_share_of_pool"] * predicted_total_pool
    )
    
    # Calculate initial score (for greedy)
    df["score"] = df.apply(
        lambda r: (
            r["expected_team_points"] / (r["predicted_team_total_bids"] + 1)
            if (r["predicted_team_total_bids"] + 1) > 0 else 0.0
        ),
        axis=1
    )
    
    # Test all three strategies
    for strategy_name in ["greedy", "minlp", "maxmin"]:
        print(f"\n  Testing {strategy_name}...")
        
        try:
            strategy_func = get_strategy(strategy_name)
            result = strategy_func(
                teams_df=df,
                budget_points=BUDGET,
                min_teams=MIN_TEAMS,
                max_teams=MAX_TEAMS,
                max_per_team_points=MAX_PER_TEAM,
                min_bid_points=MIN_BID,
            )
            
            # Calculate metrics
            bids = result["bid_amount_points"].values
            exp_pts_vals = result["expected_team_points"].values
            pred_markets = result["predicted_team_total_bids"].values
            
            ownership = bids / (pred_markets + bids)
            total_return = (exp_pts_vals * ownership).sum()
            our_roi = exp_pts_vals / (pred_markets + bids)
            
            num_bad_roi = (our_roi < 1.0).sum()
            min_roi = our_roi.min()
            mean_roi = our_roi.mean()
            
            print(f"    Expected return: {total_return:.2f} points")
            print(f"    Teams: {len(result)}")
            print(f"    Teams with our_roi < 1.0: {num_bad_roi}")
            print(f"    Min our_roi: {min_roi:.2f}")
            print(f"    Mean our_roi: {mean_roi:.2f}")
            
            results.append({
                "year": year,
                "strategy": strategy_name,
                "expected_return": total_return,
                "num_teams": len(result),
                "num_bad_roi": num_bad_roi,
                "min_roi": min_roi,
                "mean_roi": mean_roi,
            })
            
        except Exception as e:
            print(f"    Error: {e}")
            continue

# Create comparison DataFrame
if results:
    df_results = pd.DataFrame(results)
    
    print(f"\n{'='*90}")
    print("SUMMARY: GREEDY VS MINLP")
    print(f"{'='*90}\n")
    
    # Pivot to compare strategies
    pivot = df_results.pivot(
        index="year",
        columns="strategy",
        values=["expected_return", "num_bad_roi", "min_roi"]
    )
    
    print("Expected Return (points):")
    print(pivot["expected_return"])
    print()
    
    print("Teams with our_roi < 1.0:")
    print(pivot["num_bad_roi"])
    print()
    
    print("Minimum our_roi:")
    print(pivot["min_roi"])
    print()
    
    # Calculate improvements
    greedy_results = df_results[df_results["strategy"] == "greedy"]
    minlp_results = df_results[df_results["strategy"] == "minlp"]
    
    if len(greedy_results) > 0 and len(minlp_results) > 0:
        merged = greedy_results.merge(
            minlp_results,
            on="year",
            suffixes=("_greedy", "_minlp")
        )
        
        merged["return_improvement"] = (
            merged["expected_return_minlp"] - merged["expected_return_greedy"]
        )
        merged["return_improvement_pct"] = (
            100 * merged["return_improvement"] / merged["expected_return_greedy"]
        )
        
        print(f"{'='*90}")
        print("IMPROVEMENTS (MINLP vs GREEDY)")
        print(f"{'='*90}\n")
        
        print(f"{'Year':<6s} {'Return Δ':>10s} {'Return Δ%':>12s} "
              f"{'Bad ROI Δ':>12s} {'Min ROI Δ':>12s}")
        print("-"*60)
        
        for _, row in merged.iterrows():
            year = int(row["year"])
            ret_delta = row["return_improvement"]
            ret_pct = row["return_improvement_pct"]
            bad_delta = int(row["num_bad_roi_minlp"] - row["num_bad_roi_greedy"])
            min_delta = row["min_roi_minlp"] - row["min_roi_greedy"]
            
            print(f"{year:<6d} {ret_delta:>+10.2f} {ret_pct:>+11.1f}% "
                  f"{bad_delta:>+12d} {min_delta:>+12.2f}")
        
        print()
        print(f"Average return improvement: {merged['return_improvement'].mean():+.2f} points "
              f"({merged['return_improvement_pct'].mean():+.1f}%)")
        print(f"Years with improvement: {(merged['return_improvement'] > 0).sum()}/{len(merged)}")
        print(f"Years with zero bad ROI (MINLP): {(merged['num_bad_roi_minlp'] == 0).sum()}/{len(merged)}")
        print(f"Years with zero bad ROI (greedy): {(merged['num_bad_roi_greedy'] == 0).sum()}/{len(merged)}")
else:
    print("\nNo results to display")
