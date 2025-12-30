#!/usr/bin/env python
"""
Comprehensive feature testing framework for market prediction model.

Tests individual features using leave-one-year-out cross-validation with sklearn Ridge.
"""
from __future__ import annotations

from pathlib import Path
from typing import Callable, Dict, List, Optional

import numpy as np
import pandas as pd
from sklearn.linear_model import Ridge
from sklearn.preprocessing import StandardScaler


def load_year_data(year: int) -> pd.DataFrame:
    """Load team dataset with actual auction shares for a year."""
    team_dataset = pd.read_parquet(f"out/{year}/derived/team_dataset.parquet")
    entry_bids = pd.read_parquet(f"out/{year}/entry_bids.parquet")
    
    # Calculate actual auction shares
    total_bids = entry_bids["bid_amount"].sum()
    actual_shares = entry_bids.groupby("team_key")["bid_amount"].sum() / total_bids
    actual_shares = actual_shares.reset_index()
    actual_shares.columns = ["team_key", "actual_share"]
    
    # Merge
    data = team_dataset.merge(actual_shares, on="team_key", how="left")
    data["actual_share"] = data["actual_share"].fillna(0)
    
    return data


def test_feature(
    feature_name: str,
    feature_columns: List[str],
    feature_adder: Optional[Callable[[pd.DataFrame], pd.DataFrame]] = None,
    baseline_columns: Optional[List[str]] = None,
    years: Optional[List[int]] = None,
    alpha: float = 1.0,
) -> Dict:
    """
    Test a feature using leave-one-year-out cross-validation.
    
    Args:
        feature_name: Name of the feature being tested
        feature_columns: List of column names that comprise the feature
        feature_adder: Optional function to add feature columns to dataframe
        baseline_columns: Baseline features to compare against (if None, uses seed only)
        years: Years to test on (if None, uses all available years)
        alpha: Ridge regularization parameter
        
    Returns:
        Dictionary with test results
    """
    if years is None:
        years = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
    
    if baseline_columns is None:
        baseline_columns = ["seed"]
    
    results = {
        "feature": feature_name,
        "baseline_mae": [],
        "with_feature_mae": [],
        "baseline_corr": [],
        "with_feature_corr": [],
    }
    
    for test_year in years:
        train_years = [y for y in years if y != test_year]
        
        # Load data
        train_data = [load_year_data(y) for y in train_years]
        train_df = pd.concat(train_data, ignore_index=True)
        test_df = load_year_data(test_year)
        
        # Add features if function provided
        if feature_adder is not None:
            train_df = feature_adder(train_df)
            test_df = feature_adder(test_df)
        
        # Prepare baseline features
        X_train_base = train_df[baseline_columns].copy()
        X_test_base = test_df[baseline_columns].copy()
        y_train = train_df["actual_share"].values
        y_test = test_df["actual_share"].values
        
        # Handle missing values
        X_train_base = X_train_base.fillna(0)
        X_test_base = X_test_base.fillna(0)
        
        # Standardize
        scaler_base = StandardScaler()
        X_train_base_scaled = scaler_base.fit_transform(X_train_base)
        X_test_base_scaled = scaler_base.transform(X_test_base)
        
        # Train baseline model
        model_base = Ridge(alpha=alpha)
        model_base.fit(X_train_base_scaled, y_train)
        pred_base = model_base.predict(X_test_base_scaled)
        
        # Calculate baseline metrics
        mae_base = np.abs(pred_base - y_test).mean()
        corr_base = np.corrcoef(pred_base, y_test)[0, 1]
        
        # Prepare features with new feature
        combined_columns = baseline_columns + feature_columns
        X_train_feat = train_df[combined_columns].copy()
        X_test_feat = test_df[combined_columns].copy()
        
        # Handle missing values
        X_train_feat = X_train_feat.fillna(0)
        X_test_feat = X_test_feat.fillna(0)
        
        # Standardize
        scaler_feat = StandardScaler()
        X_train_feat_scaled = scaler_feat.fit_transform(X_train_feat)
        X_test_feat_scaled = scaler_feat.transform(X_test_feat)
        
        # Train model with feature
        model_feat = Ridge(alpha=alpha)
        model_feat.fit(X_train_feat_scaled, y_train)
        pred_feat = model_feat.predict(X_test_feat_scaled)
        
        # Calculate metrics with feature
        mae_feat = np.abs(pred_feat - y_test).mean()
        corr_feat = np.corrcoef(pred_feat, y_test)[0, 1]
        
        # Store results
        results["baseline_mae"].append(mae_base)
        results["baseline_corr"].append(corr_base)
        results["with_feature_mae"].append(mae_feat)
        results["with_feature_corr"].append(corr_feat)
    
    # Calculate summary statistics
    results["mean_baseline_mae"] = np.mean(results["baseline_mae"])
    results["mean_with_feature_mae"] = np.mean(results["with_feature_mae"])
    results["mae_improvement"] = results["mean_baseline_mae"] - results["mean_with_feature_mae"]
    results["mae_improvement_pct"] = (results["mae_improvement"] / results["mean_baseline_mae"]) * 100
    
    results["mean_baseline_corr"] = np.mean(results["baseline_corr"])
    results["mean_with_feature_corr"] = np.mean(results["with_feature_corr"])
    results["corr_improvement"] = results["mean_with_feature_corr"] - results["mean_baseline_corr"]
    
    # Count years with improvement
    improvements = np.array(results["baseline_mae"]) - np.array(results["with_feature_mae"])
    results["years_improved"] = (improvements > 0).sum()
    results["total_years"] = len(improvements)
    
    return results


def print_results(results: Dict) -> None:
    """Print formatted results from feature test."""
    print(f"\n{results['feature']}")
    print("=" * 90)
    print(f"Baseline MAE:     {results['mean_baseline_mae']:.6f}")
    print(f"With Feature MAE: {results['mean_with_feature_mae']:.6f}")
    print(f"Improvement:      {results['mae_improvement']:+.6f} ({results['mae_improvement_pct']:+.2f}%)")
    print()
    print(f"Baseline Corr:    {results['mean_baseline_corr']:.4f}")
    print(f"With Feature Corr:{results['mean_with_feature_corr']:.4f}")
    print(f"Improvement:      {results['corr_improvement']:+.4f}")
    print()
    print(f"Years improved:   {results['years_improved']}/{results['total_years']}")
    
    # Recommendation
    if results['mae_improvement_pct'] > 5:
        print("\n✓ STRONG RECOMMENDATION: Add this feature")
    elif results['mae_improvement_pct'] > 2:
        print("\n✓ RECOMMENDATION: Add this feature")
    elif results['mae_improvement_pct'] > 0:
        print("\n⚠ WEAK RECOMMENDATION: Consider adding this feature")
    else:
        print("\n✗ DO NOT ADD: Feature does not improve predictions")


def compare_features(results_list: List[Dict]) -> pd.DataFrame:
    """Compare multiple feature test results."""
    comparison = pd.DataFrame([
        {
            "feature": r["feature"],
            "mae_improvement_pct": r["mae_improvement_pct"],
            "corr_improvement": r["corr_improvement"],
            "years_improved": f"{r['years_improved']}/{r['total_years']}",
        }
        for r in results_list
    ])
    
    comparison = comparison.sort_values("mae_improvement_pct", ascending=False)
    return comparison
