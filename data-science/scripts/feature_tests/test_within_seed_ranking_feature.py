#!/usr/bin/env python
"""Test within-seed ranking: Is the 3rd/4th best team in a seed undervalued?"""
from __future__ import annotations

import numpy as np
import pandas as pd


def main():
    """Test if within-seed KenPom ranking improves market prediction."""
    years = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
    
    print("TESTING WITHIN-SEED KENPOM RANKING FEATURE")
    print("=" * 90)
    print()
    
    all_results = []
    
    for year in years:
        # Load data
        team_dataset = pd.read_parquet(f"out/{year}/derived/team_dataset.parquet")
        entry_bids = pd.read_parquet(f"out/{year}/entry_bids.parquet")
        
        # Calculate actual auction shares
        total_bids = entry_bids["bid_amount"].sum()
        actual_shares = entry_bids.groupby("team_key")["bid_amount"].sum() / total_bids
        actual_shares = actual_shares.reset_index()
        actual_shares.columns = ["team_key", "actual_share"]
        
        # Load predicted shares
        import glob
        shares_file = glob.glob(f"out/{year}/derived/calcutta/20251230T04*/predicted_auction_share_of_pool.parquet")[0]
        predicted = pd.read_parquet(shares_file)
        
        # Merge everything
        df = team_dataset.merge(actual_shares, on="team_key", how="left")
        df = df.merge(predicted[["team_key", "predicted_auction_share_of_pool"]], on="team_key", how="left")
        df["actual_share"] = df["actual_share"].fillna(0)
        
        # Add within-seed KenPom ranking
        df["kenpom_rank_within_seed"] = df.groupby("seed")["kenpom_net"].rank(
            ascending=False, method="dense"
        )
        
        # Calculate prediction errors
        df["prediction_error"] = df["actual_share"] - df["predicted_auction_share_of_pool"]
        
        # Focus on seeds with 4 teams (1-4 seeds typically)
        seed_counts = df.groupby("seed").size()
        seeds_with_4 = seed_counts[seed_counts == 4].index.tolist()
        
        df_4team_seeds = df[df["seed"].isin(seeds_with_4)]
        
        if len(df_4team_seeds) == 0:
            continue
        
        # Analyze by within-seed rank
        rank_1_2 = df_4team_seeds[df_4team_seeds["kenpom_rank_within_seed"] <= 2]
        rank_3_4 = df_4team_seeds[df_4team_seeds["kenpom_rank_within_seed"] >= 3]
        
        rank_1_2_error = rank_1_2["prediction_error"].mean()
        rank_3_4_error = rank_3_4["prediction_error"].mean()
        
        all_results.append({
            "year": year,
            "n_rank_1_2": len(rank_1_2),
            "n_rank_3_4": len(rank_3_4),
            "rank_1_2_error": rank_1_2_error,
            "rank_3_4_error": rank_3_4_error,
        })
        
        print(f"{year}:")
        print(f"  Top 2 within seed ({len(rank_1_2)} teams):")
        print(f"    Mean prediction error: {rank_1_2_error:+.6f}")
        print(f"  Bottom 2 within seed ({len(rank_3_4)} teams):")
        print(f"    Mean prediction error: {rank_3_4_error:+.6f}")
        print(f"  Difference: {rank_3_4_error - rank_1_2_error:+.6f}")
        print()
    
    # Summary
    results_df = pd.DataFrame(all_results)
    
    print("=" * 90)
    print("SUMMARY ACROSS ALL YEARS")
    print("=" * 90)
    print()
    print(f"Average top-2 within seed prediction error: {results_df['rank_1_2_error'].mean():+.6f}")
    print(f"Average bottom-2 within seed prediction error: {results_df['rank_3_4_error'].mean():+.6f}")
    print(f"Difference: {results_df['rank_3_4_error'].mean() - results_df['rank_1_2_error'].mean():+.6f}")
    print()
    
    diff = results_df['rank_3_4_error'].mean() - results_df['rank_1_2_error'].mean()
    
    if diff < -0.001:
        print("✓ FINDING: Bottom-2 within seed are systematically OVERPREDICTED (undervalued in market)")
        print("  Recommendation: Add negative coefficient for lower within-seed rank")
        print("  This means the 3rd/4th best team in a seed is UNDERVALUED - good arbitrage opportunity!")
    elif diff > 0.001:
        print("✓ FINDING: Bottom-2 within seed are systematically UNDERPREDICTED (overvalued in market)")
        print("  Recommendation: Add positive coefficient for lower within-seed rank")
    else:
        print("✗ NO EFFECT: Within-seed ranking does not systematically affect predictions")
        print("  Recommendation: Do not add within-seed ranking feature")


if __name__ == "__main__":
    main()
