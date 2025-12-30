#!/usr/bin/env python
"""
Test seed as categorical variable (16 separate indicators) vs continuous.
"""
from __future__ import annotations

import sys
from pathlib import Path

import pandas as pd

sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from scripts.feature_tests.test_framework import (
    compare_features,
    print_results,
    test_feature,
)


def add_seed_dummies(df):
    """Add seed as 16 separate dummy variables (one-hot encoding)."""
    df = df.copy()
    # Create dummy variables for each seed (drop first to avoid collinearity)
    seed_dummies = pd.get_dummies(df["seed"], prefix="seed", drop_first=True)
    for col in seed_dummies.columns:
        df[col] = seed_dummies[col]
    return df


def main():
    """Test seed as categorical vs continuous."""
    # Baseline WITHOUT seed (just KenPom)
    baseline_no_seed = ["kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"]
    
    # Baseline WITH continuous seed
    baseline_with_seed = [
        "seed", "kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"
    ]
    
    print("TESTING SEED AS CATEGORICAL VS CONTINUOUS")
    print("=" * 90)
    print()
    print("Comparing three approaches:")
    print("  1. No seed (KenPom only)")
    print("  2. Seed as continuous (current approach)")
    print("  3. Seed as categorical (16 dummy variables)")
    print()
    
    all_results = []
    
    # Test 1: No seed (baseline)
    print("\n" + "=" * 90)
    print("TEST: No Seed (KenPom Only)")
    print("=" * 90)
    results = test_feature(
        feature_name="No Seed (KenPom only)",
        feature_columns=[],
        baseline_columns=baseline_no_seed,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 2: Continuous seed (current approach)
    print("\n" + "=" * 90)
    print("TEST: Seed as Continuous Variable")
    print("=" * 90)
    results = test_feature(
        feature_name="Seed (continuous)",
        feature_columns=["seed"],
        baseline_columns=baseline_no_seed,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 3: Categorical seed (16 dummies)
    print("\n" + "=" * 90)
    print("TEST: Seed as Categorical (16 Dummy Variables)")
    print("=" * 90)
    
    # Get all seed dummy column names
    seed_dummy_cols = [f"seed_{i}" for i in range(2, 17)]
    
    results = test_feature(
        feature_name="Seed (categorical - 16 dummies)",
        feature_columns=seed_dummy_cols,
        feature_adder=add_seed_dummies,
        baseline_columns=baseline_no_seed,
    )
    print_results(results)
    all_results.append(results)
    
    # Comparison
    print("\n" + "=" * 90)
    print("COMPARISON")
    print("=" * 90)
    print()
    
    comparison = compare_features(all_results)
    print(comparison.to_string(index=False))
    
    print("\n" + "=" * 90)
    print("INTERPRETATION")
    print("=" * 90)
    print()
    
    categorical_result = comparison[
        comparison["feature"].str.contains("categorical")
    ].iloc[0]
    continuous_result = comparison[
        comparison["feature"].str.contains("continuous")
    ].iloc[0]
    
    improvement_diff = (
        categorical_result["mae_improvement_pct"] -
        continuous_result["mae_improvement_pct"]
    )
    
    print(f"Categorical seed improvement: "
          f"{categorical_result['mae_improvement_pct']:.2f}%")
    print(f"Continuous seed improvement:  "
          f"{continuous_result['mae_improvement_pct']:.2f}%")
    print(f"Difference: {improvement_diff:+.2f}%")
    print()
    
    if improvement_diff > 2:
        print("✓ STRONG RECOMMENDATION: Use categorical seed encoding")
        print("  Seed is truly categorical - each seed has unique behavior")
        print("  This allows model to learn 16 separate coefficients")
    elif improvement_diff > 0:
        print("✓ RECOMMENDATION: Use categorical seed encoding")
        print("  Small but consistent improvement over continuous")
    else:
        print("⚠ Continuous seed is sufficient")
        print("  Categorical encoding doesn't add meaningful value")
    
    print()
    print("IMPLICATIONS:")
    print("  - Seed², seed³ are mathematically incorrect (treating "
          "categorical as continuous)")
    print("  - Should use dummy variables instead")
    print("  - This is the 'right' way to model categorical variables")
    print("  - May need more data to estimate 16 separate coefficients")


if __name__ == "__main__":
    main()
