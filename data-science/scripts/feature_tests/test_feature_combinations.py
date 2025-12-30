#!/usr/bin/env python
"""
Efficiently test feature combinations using forward selection.

Forward selection algorithm:
1. Start with baseline features
2. Test adding each candidate feature individually
3. Add the best feature to the set
4. Repeat until no improvement or max features reached
"""
from __future__ import annotations

import sys
from pathlib import Path
from typing import Dict, List, Set

import numpy as np
import pandas as pd
from sklearn.linear_model import Ridge
from sklearn.preprocessing import StandardScaler

sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from scripts.feature_tests.test_framework import load_year_data


def add_all_candidate_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add all candidate features to dataframe."""
    df = df.copy()
    
    # Categorical seed (16 dummies, drop first)
    seed_dummies = pd.get_dummies(df["seed"], prefix="seed", drop_first=True)
    for col in seed_dummies.columns:
        df[col] = seed_dummies[col]
    
    # KenPom percentiles
    df["kenpom_net_pct"] = df["kenpom_net"].rank(pct=True)
    df["kenpom_o_pct"] = df["kenpom_o"].rank(pct=True)
    df["kenpom_d_pct"] = df["kenpom_d"].rank(pct=True)
    
    # KenPom balance
    df["kenpom_balance"] = np.abs(
        df["kenpom_o"].rank(pct=True) - df["kenpom_d"].rank(pct=True)
    )
    
    # Brand tax
    blue_bloods = {
        "duke", "north-carolina", "kentucky", "kansas", "villanova",
        "michigan-state", "louisville", "connecticut", "ucla",
        "indiana", "gonzaga", "arizona"
    }
    df["is_blue_blood"] = df["school_slug"].apply(
        lambda x: 1 if str(x).lower() in blue_bloods else 0
    )
    
    # Upset chic
    df["is_upset_seed"] = df["seed"].apply(lambda x: 1 if 10 <= x <= 12 else 0)
    
    # Within-seed ranking
    df["kenpom_rank_within_seed"] = df.groupby("seed")["kenpom_net"].rank(
        ascending=False, method="dense"
    )
    df["kenpom_rank_within_seed_norm"] = df.groupby("seed")[
        "kenpom_rank_within_seed"
    ].transform(lambda x: (x - 1) / (x.max() - 1) if x.max() > 1 else 0)
    df["kenpom_rank_within_seed_norm"] = df[
        "kenpom_rank_within_seed_norm"
    ].fillna(0.0)
    df = df.drop(columns=["kenpom_rank_within_seed"])
    
    # Championship equity (approximate)
    seed_title_prob = {
        1: 0.20, 2: 0.12, 3: 0.08, 4: 0.05, 5: 0.03, 6: 0.02,
        7: 0.01, 8: 0.01, 9: 0.005, 10: 0.003, 11: 0.002, 12: 0.001,
        13: 0.0005, 14: 0.0002, 15: 0.0001, 16: 0.00001
    }
    df["champ_equity"] = df["seed"].map(seed_title_prob)
    
    # Bubble fraud
    seed_kenpom_expected = {
        1: 25, 2: 20, 3: 15, 4: 12, 5: 10, 6: 8, 7: 6, 8: 4,
        9: 2, 10: 0, 11: -2, 12: -4, 13: -6, 14: -8, 15: -10, 16: -12
    }
    df["expected_kenpom_for_seed"] = df["seed"].map(seed_kenpom_expected)
    df["bubble_fraud"] = df["kenpom_net"] - df["expected_kenpom_for_seed"]
    df["bubble_fraud_norm"] = (
        df["bubble_fraud"] - df["bubble_fraud"].mean()
    ) / (df["bubble_fraud"].std() + 1e-8)
    df = df.drop(columns=["expected_kenpom_for_seed", "bubble_fraud"])
    
    # KenPom interactions
    df["kenpom_o_x_d"] = df["kenpom_o"] * df["kenpom_d"]
    
    return df


def evaluate_features(
    feature_cols: List[str],
    years: List[int],
    alpha: float = 1.0,
) -> Dict:
    """Evaluate a feature set using LOOCV."""
    maes = []
    corrs = []
    
    for test_year in years:
        train_years = [y for y in years if y != test_year]
        
        # Load data
        train_data = [load_year_data(y) for y in train_years]
        train_df = pd.concat(train_data, ignore_index=True)
        test_df = load_year_data(test_year)
        
        # Add features
        train_df = add_all_candidate_features(train_df)
        test_df = add_all_candidate_features(test_df)
        
        # Prepare data
        X_train = train_df[feature_cols].fillna(0)
        X_test = test_df[feature_cols].fillna(0)
        y_train = train_df["actual_share"].values
        y_test = test_df["actual_share"].values
        
        # Standardize
        scaler = StandardScaler()
        X_train_scaled = scaler.fit_transform(X_train)
        X_test_scaled = scaler.transform(X_test)
        
        # Train and predict
        model = Ridge(alpha=alpha)
        model.fit(X_train_scaled, y_train)
        pred = model.predict(X_test_scaled)
        
        # Metrics
        mae = np.abs(pred - y_test).mean()
        corr = np.corrcoef(pred, y_test)[0, 1]
        
        maes.append(mae)
        corrs.append(corr)
    
    return {
        "mae": np.mean(maes),
        "corr": np.mean(corrs),
        "maes_by_year": maes,
        "corrs_by_year": corrs,
    }


def forward_selection(
    baseline_features: List[str],
    candidate_features: List[str],
    years: List[int],
    max_features: int = 20,
    min_improvement: float = 0.001,
) -> List[str]:
    """
    Forward selection: greedily add features that improve MAE.
    
    Returns:
        List of selected features in order of addition
    """
    selected = baseline_features.copy()
    remaining = set(candidate_features) - set(baseline_features)
    
    print("FORWARD SELECTION")
    print("=" * 90)
    print(f"Baseline features: {baseline_features}")
    print(f"Candidate features: {len(remaining)}")
    print(f"Max features: {max_features}")
    print(f"Min improvement: {min_improvement:.4f}")
    print()
    
    # Evaluate baseline
    baseline_result = evaluate_features(selected, years)
    current_mae = baseline_result["mae"]
    
    print(f"Baseline MAE: {current_mae:.6f}")
    print()
    
    iteration = 1
    while remaining and len(selected) < max_features:
        print(f"Iteration {iteration}: Testing {len(remaining)} candidates...")
        
        best_feature = None
        best_mae = current_mae
        best_improvement = 0
        
        # Test each remaining feature
        for feature in remaining:
            test_features = selected + [feature]
            result = evaluate_features(test_features, years)
            improvement = current_mae - result["mae"]
            
            if result["mae"] < best_mae:
                best_mae = result["mae"]
                best_feature = feature
                best_improvement = improvement
        
        # Check if improvement meets threshold
        if best_improvement < min_improvement:
            print(f"  No feature improved MAE by >{min_improvement:.4f}")
            print(f"  Stopping.")
            break
        
        # Add best feature
        selected.append(best_feature)
        remaining.remove(best_feature)
        current_mae = best_mae
        
        print(f"  Added: {best_feature}")
        print(f"  New MAE: {current_mae:.6f} (improvement: {best_improvement:+.6f})")
        print()
        
        iteration += 1
    
    print("=" * 90)
    print("FINAL FEATURE SET")
    print("=" * 90)
    print(f"Total features: {len(selected)}")
    print(f"Final MAE: {current_mae:.6f}")
    print(f"Improvement from baseline: {baseline_result['mae'] - current_mae:+.6f}")
    print()
    print("Features (in order of addition):")
    for i, feat in enumerate(selected, 1):
        print(f"  {i:2d}. {feat}")
    
    return selected


def main():
    """Run forward selection to find optimal feature combination."""
    years = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
    
    # Baseline: just KenPom (continuous variables)
    baseline = ["kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"]
    
    # All candidate features
    candidates = baseline + [
        # Categorical seed (15 dummies, drop seed_1)
        "seed_2", "seed_3", "seed_4", "seed_5", "seed_6", "seed_7", "seed_8",
        "seed_9", "seed_10", "seed_11", "seed_12", "seed_13", "seed_14",
        "seed_15", "seed_16",
        # KenPom derived
        "kenpom_net_pct", "kenpom_o_pct", "kenpom_d_pct",
        "kenpom_balance", "kenpom_o_x_d",
        # Market behavior
        "is_blue_blood", "is_upset_seed", "kenpom_rank_within_seed_norm",
        "champ_equity", "bubble_fraud_norm",
    ]
    
    # Run forward selection
    selected_features = forward_selection(
        baseline_features=baseline,
        candidate_features=candidates,
        years=years,
        max_features=25,
        min_improvement=0.0005,  # Stop if improvement < 0.05%
    )
    
    # Final evaluation
    print("\n" + "=" * 90)
    print("FINAL EVALUATION")
    print("=" * 90)
    
    final_result = evaluate_features(selected_features, years)
    
    print(f"Mean MAE: {final_result['mae']:.6f}")
    print(f"Mean Correlation: {final_result['corr']:.4f}")
    print()
    print("MAE by year:")
    for year, mae in zip(years, final_result['maes_by_year']):
        print(f"  {year}: {mae:.6f}")


if __name__ == "__main__":
    main()
