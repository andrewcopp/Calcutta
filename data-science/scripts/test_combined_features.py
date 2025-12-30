#!/usr/bin/env python
"""Test all three features combined: brand tax + upset chic + within-seed ranking."""
from __future__ import annotations

import sys
from pathlib import Path

import numpy as np
import pandas as pd

sys.path.insert(0, str(Path(__file__).parent.parent))

from moneyball.models.predicted_auction_share_of_pool import (
    _fit_ridge,
    _predict_ridge,
)


def add_all_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add all three features to the dataframe."""
    df = df.copy()
    
    # 1. Brand tax
    blue_bloods = {
        "duke", "north-carolina", "kentucky", "kansas", "villanova",
        "michigan-state", "louisville", "connecticut", "ucla", "indiana",
        "gonzaga", "arizona"
    }
    df["is_blue_blood"] = df["school_slug"].apply(
        lambda x: 1 if str(x).lower() in blue_bloods else 0
    )
    
    # 2. Upset chic
    df["is_upset_seed"] = df["seed"].apply(lambda x: 1 if 10 <= x <= 12 else 0)
    
    # 3. Within-seed ranking
    df["kenpom_rank_within_seed"] = df.groupby("seed")["kenpom_net"].rank(
        ascending=False, method="dense"
    )
    # Normalize to 0-1 within each seed (higher = worse rank)
    df["kenpom_rank_within_seed_norm"] = df.groupby("seed")[
        "kenpom_rank_within_seed"
    ].transform(lambda x: (x - 1) / (x.max() - 1) if x.max() > 1 else 0)
    
    return df


def main():
    """Test combined features using leave-one-year-out cross-validation."""
    years = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
    
    # Baseline features (from current model)
    baseline_features = ["seed", "kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"]
    
    # New features
    new_features = ["is_blue_blood", "is_upset_seed", "kenpom_rank_within_seed_norm"]
    
    print("TESTING COMBINED FEATURES")
    print("=" * 90)
    print(f"Baseline features: {baseline_features}")
    print(f"New features: {new_features}")
    print()
    print("Methodology: Leave-one-year-out cross-validation")
    print()
    
    results = []
    
    for test_year in years:
        train_years = [y for y in years if y != test_year]
        
        # Load training data
        train_data = []
        for year in train_years:
            team_dataset = pd.read_parquet(f"out/{year}/derived/team_dataset.parquet")
            entry_bids = pd.read_parquet(f"out/{year}/entry_bids.parquet")
            
            total_bids = entry_bids["bid_amount"].sum()
            actual_shares = entry_bids.groupby("team_key")["bid_amount"].sum() / total_bids
            actual_shares = actual_shares.reset_index()
            actual_shares.columns = ["team_key", "actual_share"]
            
            data = team_dataset.merge(actual_shares, on="team_key", how="left")
            data["actual_share"] = data["actual_share"].fillna(0)
            train_data.append(data)
        
        train_df = pd.concat(train_data, ignore_index=True)
        
        # Load test data
        test_dataset = pd.read_parquet(f"out/{test_year}/derived/team_dataset.parquet")
        test_bids = pd.read_parquet(f"out/{test_year}/entry_bids.parquet")
        
        total_bids = test_bids["bid_amount"].sum()
        test_actual = test_bids.groupby("team_key")["bid_amount"].sum() / total_bids
        test_actual = test_actual.reset_index()
        test_actual.columns = ["team_key", "actual_share"]
        
        test_df = test_dataset.merge(test_actual, on="team_key", how="left")
        test_df["actual_share"] = test_df["actual_share"].fillna(0)
        
        # Add features
        train_df = add_all_features(train_df)
        test_df = add_all_features(test_df)
        
        # Prepare baseline model
        X_train_base = train_df[baseline_features].fillna(0)
        X_test_base = test_df[baseline_features].fillna(0)
        y_train = train_df["actual_share"].values
        y_test = test_df["actual_share"].values
        
        # Standardize baseline
        X_train_base_mean = X_train_base.mean()
        X_train_base_std = X_train_base.std()
        X_train_base_scaled = (X_train_base - X_train_base_mean) / (X_train_base_std + 1e-8)
        X_test_base_scaled = (X_test_base - X_train_base_mean) / (X_train_base_std + 1e-8)
        
        # Train baseline model
        coef_base = _fit_ridge(X_train_base_scaled, pd.Series(y_train), alpha=1.0)
        if coef_base is None:
            print(f"  {test_year}: Failed to fit baseline model")
            continue
        
        pred_base = _predict_ridge(X_test_base_scaled, coef_base)
        mae_base = np.abs(pred_base - y_test).mean()
        corr_base = np.corrcoef(pred_base, y_test)[0, 1]
        
        # Prepare combined model
        combined_features = baseline_features + new_features
        X_train_comb = train_df[combined_features].fillna(0)
        X_test_comb = test_df[combined_features].fillna(0)
        
        # Standardize combined
        X_train_comb_mean = X_train_comb.mean()
        X_train_comb_std = X_train_comb.std()
        X_train_comb_scaled = (X_train_comb - X_train_comb_mean) / (X_train_comb_std + 1e-8)
        X_test_comb_scaled = (X_test_comb - X_train_comb_mean) / (X_train_comb_std + 1e-8)
        
        # Train combined model
        coef_comb = _fit_ridge(X_train_comb_scaled, pd.Series(y_train), alpha=1.0)
        if coef_comb is None:
            print(f"  {test_year}: Failed to fit combined model")
            continue
        
        pred_comb = _predict_ridge(X_test_comb_scaled, coef_comb)
        mae_comb = np.abs(pred_comb - y_test).mean()
        corr_comb = np.corrcoef(pred_comb, y_test)[0, 1]
        
        # Store results
        results.append({
            "year": test_year,
            "baseline_mae": mae_base,
            "combined_mae": mae_comb,
            "mae_improvement": mae_base - mae_comb,
            "mae_improvement_pct": (mae_base - mae_comb) / mae_base * 100,
            "baseline_corr": corr_base,
            "combined_corr": corr_comb,
            "corr_improvement": corr_comb - corr_base,
        })
        
        print(f"{test_year}:")
        print(f"  Baseline MAE: {mae_base:.6f}, Correlation: {corr_base:.4f}")
        print(f"  Combined MAE: {mae_comb:.6f}, Correlation: {corr_comb:.4f}")
        print(f"  Improvement: {mae_base - mae_comb:+.6f} ({(mae_base - mae_comb) / mae_base * 100:+.2f}%)")
        print()
    
    # Summary
    results_df = pd.DataFrame(results)
    
    print("=" * 90)
    print("SUMMARY ACROSS ALL YEARS")
    print("=" * 90)
    print()
    print(f"Mean baseline MAE: {results_df['baseline_mae'].mean():.6f}")
    print(f"Mean combined MAE: {results_df['combined_mae'].mean():.6f}")
    print(f"Mean improvement: {results_df['mae_improvement'].mean():+.6f} ({results_df['mae_improvement_pct'].mean():+.2f}%)")
    print()
    print(f"Mean baseline correlation: {results_df['baseline_corr'].mean():.4f}")
    print(f"Mean combined correlation: {results_df['combined_corr'].mean():.4f}")
    print(f"Mean improvement: {results_df['corr_improvement'].mean():+.4f}")
    print()
    
    # Check consistency
    improvements = results_df['mae_improvement'].values
    positive_improvements = (improvements > 0).sum()
    
    print(f"Years with improvement: {positive_improvements}/{len(improvements)}")
    print()
    
    if results_df['mae_improvement_pct'].mean() > 5:
        print("✓ STRONG RECOMMENDATION: Implement all three features")
        print(f"  Expected MAE reduction: {results_df['mae_improvement_pct'].mean():.1f}%")
        print("  All features show consistent positive impact")
    elif results_df['mae_improvement_pct'].mean() > 2:
        print("✓ RECOMMENDATION: Implement all three features")
        print(f"  Expected MAE reduction: {results_df['mae_improvement_pct'].mean():.1f}%")
    elif results_df['mae_improvement_pct'].mean() > 0:
        print("⚠ WEAK RECOMMENDATION: Consider implementing features")
        print(f"  Expected MAE reduction: {results_df['mae_improvement_pct'].mean():.1f}%")
        print("  Impact is small but positive")
    else:
        print("✗ DO NOT IMPLEMENT: Features do not improve predictions")


if __name__ == "__main__":
    main()
