#!/usr/bin/env python
"""Test upset chic feature: Are seeds 10-12 systematically overbid?"""
from __future__ import annotations

import numpy as np
import pandas as pd


def main():
    """Test if upset chic (seeds 10-12) improves market prediction."""
    years = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
    
    print("TESTING UPSET CHIC FEATURE (Seeds 10-12)")
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
        
        # Add upset chic indicator
        df["is_upset_seed"] = df["seed"].apply(lambda x: 1 if 10 <= x <= 12 else 0)
        
        # Calculate prediction errors
        df["prediction_error"] = df["actual_share"] - df["predicted_auction_share_of_pool"]
        
        # Analyze by upset seed status
        upset_seeds = df[df["is_upset_seed"] == 1]
        other_seeds = df[df["is_upset_seed"] == 0]
        
        upset_mean_error = upset_seeds["prediction_error"].mean()
        other_mean_error = other_seeds["prediction_error"].mean()
        
        upset_mae = np.abs(upset_seeds["prediction_error"]).mean()
        other_mae = np.abs(other_seeds["prediction_error"]).mean()
        
        all_results.append({
            "year": year,
            "n_upset_seeds": len(upset_seeds),
            "upset_mean_error": upset_mean_error,
            "other_mean_error": other_mean_error,
            "upset_mae": upset_mae,
            "other_mae": other_mae,
        })
        
        print(f"{year}:")
        print(f"  Seeds 10-12 ({len(upset_seeds)} teams):")
        print(f"    Mean prediction error: {upset_mean_error:+.6f} (positive = underpredicted)")
        print(f"    MAE: {upset_mae:.6f}")
        print(f"  Other seeds ({len(other_seeds)} teams):")
        print(f"    Mean prediction error: {other_mean_error:+.6f}")
        print(f"    MAE: {other_mae:.6f}")
        print()
    
    # Summary
    results_df = pd.DataFrame(all_results)
    
    print("=" * 90)
    print("SUMMARY ACROSS ALL YEARS")
    print("=" * 90)
    print()
    print(f"Average upset seed (10-12) prediction error: {results_df['upset_mean_error'].mean():+.6f}")
    print(f"Average other seed prediction error: {results_df['other_mean_error'].mean():+.6f}")
    print(f"Difference: {results_df['upset_mean_error'].mean() - results_df['other_mean_error'].mean():+.6f}")
    print()
    
    if results_df["upset_mean_error"].mean() > 0.001:
        print("✓ FINDING: Seeds 10-12 are systematically UNDERPREDICTED (overbid in market)")
        print("  Recommendation: Add positive coefficient for is_upset_seed feature")
    elif results_df["upset_mean_error"].mean() < -0.001:
        print("✓ FINDING: Seeds 10-12 are systematically OVERPREDICTED (underbid in market)")
        print("  Recommendation: Add negative coefficient for is_upset_seed feature")
    else:
        print("✗ NO EFFECT: Seeds 10-12 are not systematically mis-predicted")
        print("  Recommendation: Do not add upset chic feature")


if __name__ == "__main__":
    main()
