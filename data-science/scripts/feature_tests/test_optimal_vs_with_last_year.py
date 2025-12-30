#!/usr/bin/env python
"""
Compare optimal feature set (no last year) vs with last year features.
"""
from __future__ import annotations

import sys
from pathlib import Path

import numpy as np
import pandas as pd
from sklearn.linear_model import Ridge
from sklearn.preprocessing import StandardScaler

sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from scripts.feature_tests.test_framework import load_year_data


def main():
    """Compare feature sets with and without last year features."""
    years = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
    
    # Core features (always included)
    core_features = ["seed", "kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"]
    
    # Interaction features
    interaction_features = ["seed_sq", "kenpom_x_seed"]
    
    # Market behavior features
    market_features = ["is_blue_blood", "is_upset_seed", "kenpom_rank_within_seed_norm"]
    
    # Last year features
    last_year_features = [
        "has_last_year", "wins_last_year", "byes_last_year", "progress_last_year",
        "points_last_year", "total_bid_amount_last_year", "team_share_of_pool_last_year",
        "points_per_dollar_last_year", "bid_per_point_last_year",
        "expected_progress_last_year", "expected_points_last_year",
        "expected_points_per_dollar_last_year", "progress_ratio_last_year",
        "progress_residual_last_year", "roi_ratio_last_year"
    ]
    
    print("COMPARING FEATURE SETS")
    print("=" * 90)
    print()
    print("Optimal: Core + Interactions + Market (NO last year)")
    print("With Last Year: Core + Interactions + Market + Last Year")
    print()
    
    results_optimal = []
    results_with_last_year = []
    
    for test_year in years:
        train_years = [y for y in years if y != test_year]
        
        # Load data
        train_data = [load_year_data(y) for y in train_years]
        train_df = pd.concat(train_data, ignore_index=True)
        test_df = load_year_data(test_year)
        
        # Add interaction features
        for df in [train_df, test_df]:
            df["seed_sq"] = df["seed"] ** 2
            df["kenpom_x_seed"] = df["kenpom_net"] * df["seed"]
            
            # Add market features
            blue_bloods = {
                "duke", "north-carolina", "kentucky", "kansas", "villanova",
                "michigan-state", "louisville", "connecticut", "ucla",
                "indiana", "gonzaga", "arizona"
            }
            df["is_blue_blood"] = df["school_slug"].apply(
                lambda x: 1 if str(x).lower() in blue_bloods else 0
            )
            df["is_upset_seed"] = df["seed"].apply(lambda x: 1 if 10 <= x <= 12 else 0)
            
            df["kenpom_rank_within_seed"] = df.groupby("seed")["kenpom_net"].rank(
                ascending=False, method="dense"
            )
            df["kenpom_rank_within_seed_norm"] = df.groupby("seed")[
                "kenpom_rank_within_seed"
            ].transform(lambda x: (x - 1) / (x.max() - 1) if x.max() > 1 else 0)
            df["kenpom_rank_within_seed_norm"] = df["kenpom_rank_within_seed_norm"].fillna(0.0)
        
        y_train = train_df["actual_share"].values
        y_test = test_df["actual_share"].values
        
        # Test 1: Optimal (no last year)
        optimal_features = core_features + interaction_features + market_features
        X_train_opt = train_df[optimal_features].fillna(0)
        X_test_opt = test_df[optimal_features].fillna(0)
        
        scaler_opt = StandardScaler()
        X_train_opt_scaled = scaler_opt.fit_transform(X_train_opt)
        X_test_opt_scaled = scaler_opt.transform(X_test_opt)
        
        model_opt = Ridge(alpha=1.0)
        model_opt.fit(X_train_opt_scaled, y_train)
        pred_opt = model_opt.predict(X_test_opt_scaled)
        
        mae_opt = np.abs(pred_opt - y_test).mean()
        corr_opt = np.corrcoef(pred_opt, y_test)[0, 1]
        
        # Test 2: With last year
        all_features = optimal_features + last_year_features
        X_train_all = train_df[all_features].fillna(0)
        X_test_all = test_df[all_features].fillna(0)
        
        scaler_all = StandardScaler()
        X_train_all_scaled = scaler_all.fit_transform(X_train_all)
        X_test_all_scaled = scaler_all.transform(X_test_all)
        
        model_all = Ridge(alpha=1.0)
        model_all.fit(X_train_all_scaled, y_train)
        pred_all = model_all.predict(X_test_all_scaled)
        
        mae_all = np.abs(pred_all - y_test).mean()
        corr_all = np.corrcoef(pred_all, y_test)[0, 1]
        
        # Store results
        results_optimal.append({"year": test_year, "mae": mae_opt, "corr": corr_opt})
        results_with_last_year.append({"year": test_year, "mae": mae_all, "corr": corr_all})
        
        improvement = mae_all - mae_opt
        print(f"{test_year}:")
        print(f"  Optimal MAE:        {mae_opt:.6f}")
        print(f"  With Last Year MAE: {mae_all:.6f}")
        print(f"  Difference:         {improvement:+.6f} ({'worse' if improvement > 0 else 'better'} with last year)")
        print()
    
    # Summary
    df_opt = pd.DataFrame(results_optimal)
    df_all = pd.DataFrame(results_with_last_year)
    
    print("=" * 90)
    print("SUMMARY")
    print("=" * 90)
    print()
    print(f"Mean Optimal MAE:        {df_opt['mae'].mean():.6f}")
    print(f"Mean With Last Year MAE: {df_all['mae'].mean():.6f}")
    print(f"Difference:              {df_all['mae'].mean() - df_opt['mae'].mean():+.6f}")
    print()
    
    improvements = (df_all['mae'] - df_opt['mae']).values
    years_better_without = (improvements > 0).sum()
    
    print(f"Years better WITHOUT last year: {years_better_without}/8")
    print()
    
    if df_opt['mae'].mean() < df_all['mae'].mean():
        print("✓ RECOMMENDATION: Use OPTIMAL (no last year features)")
        print("  Last year features add noise and hurt predictions")
    else:
        print("⚠ Keep last year features - they help slightly")


if __name__ == "__main__":
    main()
