#!/usr/bin/env python
"""
Test seed indicators at different levels of granularity.
Compare 1-seeds vs 2-seeds vs combined top seeds.
"""
from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from scripts.feature_tests.test_framework import (
    compare_features,
    print_results,
    test_feature,
)


def add_seed_1_indicator(df):
    """Add indicator for 1-seeds only."""
    df = df.copy()
    df["is_seed_1"] = (df["seed"] == 1).astype(int)
    return df


def add_seed_2_indicator(df):
    """Add indicator for 2-seeds only."""
    df = df.copy()
    df["is_seed_2"] = (df["seed"] == 2).astype(int)
    return df


def add_seed_1_and_2_indicators(df):
    """Add separate indicators for 1-seeds and 2-seeds."""
    df = df.copy()
    df["is_seed_1"] = (df["seed"] == 1).astype(int)
    df["is_seed_2"] = (df["seed"] == 2).astype(int)
    return df


def add_top_seeds_indicator(df):
    """Add combined indicator for 1-2 seeds."""
    df = df.copy()
    df["is_top_seed"] = (df["seed"] <= 2).astype(int)
    return df


def main():
    """Test seed indicators at different granularities."""
    baseline_cols = [
        "seed", "kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"
    ]
    
    print("TESTING SEED INDICATOR GRANULARITY")
    print("=" * 90)
    print()
    print("Comparing different ways to capture top seed behavior:")
    print("  1. Single indicator for 1-seeds only")
    print("  2. Single indicator for 2-seeds only")
    print("  3. Separate indicators for 1-seeds AND 2-seeds")
    print("  4. Combined indicator for 1-2 seeds together")
    print()
    
    all_results = []
    
    # Test 1: 1-seeds only
    print("\n" + "=" * 90)
    print("TEST: 1-Seed Indicator Only")
    print("=" * 90)
    results = test_feature(
        feature_name="1-Seed Only",
        feature_columns=["is_seed_1"],
        feature_adder=add_seed_1_indicator,
        baseline_columns=baseline_cols,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 2: 2-seeds only
    print("\n" + "=" * 90)
    print("TEST: 2-Seed Indicator Only")
    print("=" * 90)
    results = test_feature(
        feature_name="2-Seed Only",
        feature_columns=["is_seed_2"],
        feature_adder=add_seed_2_indicator,
        baseline_columns=baseline_cols,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 3: Both 1-seeds and 2-seeds as separate features
    print("\n" + "=" * 90)
    print("TEST: Separate 1-Seed AND 2-Seed Indicators")
    print("=" * 90)
    results = test_feature(
        feature_name="1-Seed AND 2-Seed (separate)",
        feature_columns=["is_seed_1", "is_seed_2"],
        feature_adder=add_seed_1_and_2_indicators,
        baseline_columns=baseline_cols,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 4: Combined top seeds (1-2)
    print("\n" + "=" * 90)
    print("TEST: Combined Top Seeds (1-2) Indicator")
    print("=" * 90)
    results = test_feature(
        feature_name="Top Seeds (1-2) Combined",
        feature_columns=["is_top_seed"],
        feature_adder=add_top_seeds_indicator,
        baseline_columns=baseline_cols,
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
    print("RECOMMENDATION")
    print("=" * 90)
    print()
    
    best = comparison.iloc[0]
    print(f"Best approach: {best['feature']}")
    print(f"  MAE improvement: {best['mae_improvement_pct']:.2f}%")
    print(f"  Correlation improvement: {best['corr_improvement']:.4f}")
    print(f"  Years improved: {best['years_improved']}")
    print()
    
    if "separate" in best['feature'].lower():
        print("✓ Use SEPARATE indicators for 1-seeds and 2-seeds")
        print("  This allows the model to learn different coefficients for each")
        print("  Likely: 1-seeds are underbid, 2-seeds are overbid")
    elif "1-Seed Only" in best['feature']:
        print("✓ Use 1-SEED indicator only")
        print("  2-seeds don't have systematic market behavior")
    elif "2-Seed Only" in best['feature']:
        print("✓ Use 2-SEED indicator only")
        print("  1-seeds don't have systematic market behavior")
    else:
        print("✓ Use COMBINED top seeds (1-2) indicator")
        print("  1-seeds and 2-seeds have similar market behavior")


if __name__ == "__main__":
    main()
