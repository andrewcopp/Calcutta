#!/usr/bin/env python
"""
Test all candidate features for market prediction model.

Systematically tests each feature individually to determine which ones
improve prediction accuracy.
"""
from __future__ import annotations

import sys
from pathlib import Path

import pandas as pd

# Add parent directory to path
sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from scripts.feature_tests.test_framework import (
    compare_features,
    print_results,
    test_feature,
)


# Feature definitions
def add_brand_tax(df: pd.DataFrame) -> pd.DataFrame:
    """Add brand tax feature."""
    blue_bloods = {
        "duke", "north-carolina", "kentucky", "kansas", "villanova",
        "michigan-state", "louisville", "connecticut", "ucla", "indiana",
        "gonzaga", "arizona"
    }
    df = df.copy()
    df["is_blue_blood"] = df["school_slug"].apply(
        lambda x: 1 if str(x).lower() in blue_bloods else 0
    )
    return df


def add_upset_chic(df: pd.DataFrame) -> pd.DataFrame:
    """Add upset chic feature."""
    df = df.copy()
    df["is_upset_seed"] = df["seed"].apply(lambda x: 1 if 10 <= x <= 12 else 0)
    return df


def add_within_seed_ranking(df: pd.DataFrame) -> pd.DataFrame:
    """Add within-seed ranking feature."""
    df = df.copy()
    df["kenpom_rank_within_seed"] = df.groupby("seed")["kenpom_net"].rank(
        ascending=False, method="dense"
    )
    df["kenpom_rank_within_seed_norm"] = df.groupby("seed")[
        "kenpom_rank_within_seed"
    ].transform(lambda x: (x - 1) / (x.max() - 1) if x.max() > 1 else 0)
    df["kenpom_rank_within_seed_norm"] = df["kenpom_rank_within_seed_norm"].fillna(0.0)
    return df


def add_region_strength(df: pd.DataFrame) -> pd.DataFrame:
    """Add region strength feature."""
    df = df.copy()
    region_strength = df.groupby("region").apply(
        lambda g: g.nlargest(4, "kenpom_net")["kenpom_net"].mean()
    )
    df["region_strength"] = df["region"].map(region_strength)
    df["region_strength_norm"] = (
        df["region_strength"] - df["region_strength"].mean()
    ) / (df["region_strength"].std() + 1e-8)
    return df


def add_kenpom_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add KenPom interaction features."""
    df = df.copy()
    df["seed_sq"] = df["seed"] ** 2
    df["kenpom_x_seed"] = df["kenpom_net"] * df["seed"]
    return df


def add_last_year_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add last year performance features (if available)."""
    df = df.copy()
    last_year_cols = [
        "has_last_year",
        "wins_last_year",
        "progress_last_year",
        "team_share_of_pool_last_year",
    ]
    for col in last_year_cols:
        if col not in df.columns:
            df[col] = 0
    return df


def main():
    """Test all features and report results."""
    print("COMPREHENSIVE FEATURE TESTING")
    print("=" * 90)
    print()
    print("Testing individual features against seed-only baseline")
    print("Using leave-one-year-out cross-validation with sklearn Ridge (alpha=1.0)")
    print()
    
    all_results = []
    
    # Test 1: Seed only (baseline)
    print("\n" + "=" * 90)
    print("BASELINE: Seed only")
    print("=" * 90)
    results = test_feature(
        feature_name="Seed only (baseline)",
        feature_columns=[],
        baseline_columns=["seed"],
    )
    print_results(results)
    all_results.append(results)
    
    # Test 2: KenPom net rating
    print("\n" + "=" * 90)
    print("TEST: KenPom Net Rating")
    print("=" * 90)
    results = test_feature(
        feature_name="KenPom Net Rating",
        feature_columns=["kenpom_net"],
        baseline_columns=["seed"],
    )
    print_results(results)
    all_results.append(results)
    
    # Test 3: All KenPom features
    print("\n" + "=" * 90)
    print("TEST: All KenPom Features (net, o, d, adj_t)")
    print("=" * 90)
    results = test_feature(
        feature_name="All KenPom Features",
        feature_columns=["kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"],
        baseline_columns=["seed"],
    )
    print_results(results)
    all_results.append(results)
    
    # Test 4: KenPom + interactions
    print("\n" + "=" * 90)
    print("TEST: KenPom + Interaction Features")
    print("=" * 90)
    results = test_feature(
        feature_name="KenPom + Interactions",
        feature_columns=["kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t", "seed_sq", "kenpom_x_seed"],
        feature_adder=add_kenpom_features,
        baseline_columns=["seed"],
    )
    print_results(results)
    all_results.append(results)
    
    # Test 5: Brand tax
    print("\n" + "=" * 90)
    print("TEST: Brand Tax")
    print("=" * 90)
    results = test_feature(
        feature_name="Brand Tax",
        feature_columns=["is_blue_blood"],
        feature_adder=add_brand_tax,
        baseline_columns=["seed", "kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"],
    )
    print_results(results)
    all_results.append(results)
    
    # Test 6: Upset chic
    print("\n" + "=" * 90)
    print("TEST: Upset Chic (Seeds 10-12)")
    print("=" * 90)
    results = test_feature(
        feature_name="Upset Chic",
        feature_columns=["is_upset_seed"],
        feature_adder=add_upset_chic,
        baseline_columns=["seed", "kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"],
    )
    print_results(results)
    all_results.append(results)
    
    # Test 7: Within-seed ranking
    print("\n" + "=" * 90)
    print("TEST: Within-Seed KenPom Ranking")
    print("=" * 90)
    results = test_feature(
        feature_name="Within-Seed Ranking",
        feature_columns=["kenpom_rank_within_seed_norm"],
        feature_adder=add_within_seed_ranking,
        baseline_columns=["seed", "kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"],
    )
    print_results(results)
    all_results.append(results)
    
    # Test 8: Region strength
    print("\n" + "=" * 90)
    print("TEST: Region Strength")
    print("=" * 90)
    results = test_feature(
        feature_name="Region Strength",
        feature_columns=["region_strength_norm"],
        feature_adder=add_region_strength,
        baseline_columns=["seed", "kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"],
    )
    print_results(results)
    all_results.append(results)
    
    # Test 9: Last year features (if available)
    print("\n" + "=" * 90)
    print("TEST: Last Year Performance")
    print("=" * 90)
    results = test_feature(
        feature_name="Last Year Performance",
        feature_columns=["has_last_year", "wins_last_year", "progress_last_year", "team_share_of_pool_last_year"],
        feature_adder=add_last_year_features,
        baseline_columns=["seed", "kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"],
    )
    print_results(results)
    all_results.append(results)
    
    # Final comparison
    print("\n" + "=" * 90)
    print("FINAL COMPARISON")
    print("=" * 90)
    print()
    
    comparison = compare_features(all_results)
    print(comparison.to_string(index=False))
    
    print("\n" + "=" * 90)
    print("RECOMMENDATIONS")
    print("=" * 90)
    print()
    
    # Filter to features with positive improvement
    positive = comparison[comparison["mae_improvement_pct"] > 0]
    
    if len(positive) > 0:
        print("Features to implement (in order of impact):")
        for _, row in positive.iterrows():
            print(f"  {row['feature']:40s}: {row['mae_improvement_pct']:+.2f}% MAE improvement")
    else:
        print("No features showed consistent improvement over baseline")


if __name__ == "__main__":
    main()
