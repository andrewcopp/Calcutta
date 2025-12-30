#!/usr/bin/env python
"""Test brand tax feature: Do blue-blood programs get systematically overbid?"""
from __future__ import annotations

import numpy as np
import pandas as pd


def main():
    """Test if brand tax improves market prediction."""
    years = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
    
    # Define blue-blood programs
    blue_bloods = {
        "duke", "north-carolina", "kentucky", "kansas", "villanova",
        "michigan-state", "louisville", "connecticut", "ucla", "indiana",
        "gonzaga", "arizona"
    }
    
    print("TESTING BRAND TAX FEATURE")
    print("=" * 90)
    print(f"Blue-blood programs: {', '.join(sorted(blue_bloods))}")
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
        
        # Add brand indicator
        df["is_blue_blood"] = df["school_slug"].apply(
            lambda x: 1 if str(x).lower() in blue_bloods else 0
        )
        
        # Calculate prediction errors
        df["prediction_error"] = df["actual_share"] - df["predicted_auction_share_of_pool"]
        
        # Analyze by brand status
        blue_blood_teams = df[df["is_blue_blood"] == 1]
        non_blue_blood_teams = df[df["is_blue_blood"] == 0]
        
        bb_mean_error = blue_blood_teams["prediction_error"].mean()
        non_bb_mean_error = non_blue_blood_teams["prediction_error"].mean()
        
        bb_mae = np.abs(blue_blood_teams["prediction_error"]).mean()
        non_bb_mae = np.abs(non_blue_blood_teams["prediction_error"]).mean()
        
        all_results.append({
            "year": year,
            "n_blue_bloods": len(blue_blood_teams),
            "bb_mean_error": bb_mean_error,
            "non_bb_mean_error": non_bb_mean_error,
            "bb_mae": bb_mae,
            "non_bb_mae": non_bb_mae,
        })
        
        print(f"{year}:")
        print(f"  Blue-bloods ({len(blue_blood_teams)} teams):")
        print(f"    Mean prediction error: {bb_mean_error:+.6f} (positive = underpredicted)")
        print(f"    MAE: {bb_mae:.6f}")
        print(f"  Non-blue-bloods ({len(non_blue_blood_teams)} teams):")
        print(f"    Mean prediction error: {non_bb_mean_error:+.6f}")
        print(f"    MAE: {non_bb_mae:.6f}")
        print()
    
    # Summary
    results_df = pd.DataFrame(all_results)
    
    print("=" * 90)
    print("SUMMARY ACROSS ALL YEARS")
    print("=" * 90)
    print()
    print(f"Average blue-blood prediction error: {results_df['bb_mean_error'].mean():+.6f}")
    print(f"Average non-blue-blood prediction error: {results_df['non_bb_mean_error'].mean():+.6f}")
    print(f"Difference: {results_df['bb_mean_error'].mean() - results_df['non_bb_mean_error'].mean():+.6f}")
    print()
    
    if results_df["bb_mean_error"].mean() > 0.001:
        print("✓ FINDING: Blue-bloods are systematically UNDERPREDICTED (overbid in market)")
        print("  Recommendation: Add positive coefficient for is_blue_blood feature")
    elif results_df["bb_mean_error"].mean() < -0.001:
        print("✓ FINDING: Blue-bloods are systematically OVERPREDICTED (underbid in market)")
        print("  Recommendation: Add negative coefficient for is_blue_blood feature")
    else:
        print("✗ NO EFFECT: Blue-bloods are not systematically mis-predicted")
        print("  Recommendation: Do not add brand tax feature")


if __name__ == "__main__":
    main()
