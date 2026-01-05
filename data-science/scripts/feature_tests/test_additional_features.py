#!/usr/bin/env python
"""
Test additional feature candidates from
data-science/docs/design/ideas.md backlog.
"""
from __future__ import annotations

import sys
from pathlib import Path

import numpy as np
import pandas as pd

sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from scripts.feature_tests.test_framework import (
    compare_features,
    print_results,
    test_feature,
)


def add_conference_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add conference-based features."""
    df = df.copy()
    
    # Top conferences (historically strong)
    power_conferences = {
        "acc", "big-12", "big-ten", "sec", "pac-12", "big-east"
    }
    
    if "conference_slug" in df.columns:
        df["is_power_conference"] = df["conference_slug"].apply(
            lambda x: 1 if str(x).lower() in power_conferences else 0
        )
    else:
        df["is_power_conference"] = 0
    
    return df


def add_bubble_fraud_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add bubble fraud: disagreement between seed and KenPom."""
    df = df.copy()
    
    # Expected KenPom by seed (approximate)
    seed_kenpom_expected = {
        1: 25, 2: 20, 3: 15, 4: 12, 5: 10, 6: 8, 7: 6, 8: 4,
        9: 2, 10: 0, 11: -2, 12: -4, 13: -6, 14: -8, 15: -10, 16: -12
    }
    
    df["expected_kenpom_for_seed"] = df["seed"].map(seed_kenpom_expected)
    df["kenpom_vs_seed"] = df["kenpom_net"] - df["expected_kenpom_for_seed"]
    
    # Normalize
    df["kenpom_vs_seed_norm"] = (
        df["kenpom_vs_seed"] - df["kenpom_vs_seed"].mean()
    ) / (df["kenpom_vs_seed"].std() + 1e-8)
    
    return df


def add_championship_equity_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add championship equity features based on seed."""
    df = df.copy()
    
    # Approximate championship probability by seed (historical)
    seed_title_prob = {
        1: 0.20, 2: 0.12, 3: 0.08, 4: 0.05, 5: 0.03, 6: 0.02,
        7: 0.01, 8: 0.01, 9: 0.005, 10: 0.003, 11: 0.002, 12: 0.001,
        13: 0.0005, 14: 0.0002, 15: 0.0001, 16: 0.00001
    }
    
    df["approx_title_prob"] = df["seed"].map(seed_title_prob)
    
    # Title equity relative to expected points
    # Higher seeds have disproportionate title equity
    df["title_equity_ratio"] = df["approx_title_prob"] * 100
    
    return df


def add_extreme_seed_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add features for extreme seeds (very high or very low)."""
    df = df.copy()
    
    # Top seeds (1-2)
    df["is_top_seed"] = df["seed"].apply(lambda x: 1 if x <= 2 else 0)
    
    # Bottom seeds (13-16)
    df["is_bottom_seed"] = df["seed"].apply(lambda x: 1 if x >= 13 else 0)
    
    # Middle seeds (5-8)
    df["is_middle_seed"] = df["seed"].apply(
        lambda x: 1 if 5 <= x <= 8 else 0
    )
    
    return df


def add_kenpom_percentile_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add KenPom percentile within tournament field."""
    df = df.copy()
    
    # Percentile rank within field
    df["kenpom_percentile"] = df["kenpom_net"].rank(pct=True)
    
    # Offensive/defensive balance
    df["kenpom_balance"] = np.abs(
        df["kenpom_o"].rank(pct=True) - df["kenpom_d"].rank(pct=True)
    )
    
    return df


def add_tempo_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add tempo-based features."""
    df = df.copy()
    
    # High tempo teams
    df["is_high_tempo"] = (
        df["kenpom_adj_t"] > df["kenpom_adj_t"].quantile(0.75)
    ).astype(int)
    
    # Low tempo teams
    df["is_low_tempo"] = (
        df["kenpom_adj_t"] < df["kenpom_adj_t"].quantile(0.25)
    ).astype(int)
    
    return df


def main():
    """Test additional features from IDEAS backlog."""
    print("TESTING ADDITIONAL FEATURES FROM IDEAS BACKLOG")
    print("=" * 90)
    print()
    print("Baseline: seed + kenpom_net + kenpom_o + kenpom_d + kenpom_adj_t")
    print("          + seed_sq + kenpom_x_seed")
    print()
    
    # Baseline includes KenPom interactions (already validated as best)
    baseline_cols = [
        "seed", "kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"
    ]
    
    all_results = []
    
    # Test 1: Conference effects
    print("\n" + "=" * 90)
    print("TEST: Power Conference Indicator")
    print("=" * 90)
    results = test_feature(
        feature_name="Power Conference",
        feature_columns=["is_power_conference"],
        feature_adder=add_conference_features,
        baseline_columns=baseline_cols,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 2: Bubble fraud (KenPom vs seed disagreement)
    print("\n" + "=" * 90)
    print("TEST: Bubble Fraud (KenPom vs Seed Disagreement)")
    print("=" * 90)
    results = test_feature(
        feature_name="Bubble Fraud",
        feature_columns=["kenpom_vs_seed_norm"],
        feature_adder=add_bubble_fraud_features,
        baseline_columns=baseline_cols,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 3: Championship equity
    print("\n" + "=" * 90)
    print("TEST: Championship Equity (Title Probability)")
    print("=" * 90)
    results = test_feature(
        feature_name="Championship Equity",
        feature_columns=["title_equity_ratio"],
        feature_adder=add_championship_equity_features,
        baseline_columns=baseline_cols,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 4: Top seeds (1-2)
    print("\n" + "=" * 90)
    print("TEST: Top Seeds (1-2) Indicator")
    print("=" * 90)
    results = test_feature(
        feature_name="Top Seeds (1-2)",
        feature_columns=["is_top_seed"],
        feature_adder=add_extreme_seed_features,
        baseline_columns=baseline_cols,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 5: Bottom seeds (13-16)
    print("\n" + "=" * 90)
    print("TEST: Bottom Seeds (13-16) Indicator")
    print("=" * 90)
    results = test_feature(
        feature_name="Bottom Seeds (13-16)",
        feature_columns=["is_bottom_seed"],
        feature_adder=add_extreme_seed_features,
        baseline_columns=baseline_cols,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 6: Middle seeds (5-8)
    print("\n" + "=" * 90)
    print("TEST: Middle Seeds (5-8) Indicator")
    print("=" * 90)
    results = test_feature(
        feature_name="Middle Seeds (5-8)",
        feature_columns=["is_middle_seed"],
        feature_adder=add_extreme_seed_features,
        baseline_columns=baseline_cols,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 7: KenPom percentile
    print("\n" + "=" * 90)
    print("TEST: KenPom Percentile Rank")
    print("=" * 90)
    results = test_feature(
        feature_name="KenPom Percentile",
        feature_columns=["kenpom_percentile"],
        feature_adder=add_kenpom_percentile_features,
        baseline_columns=baseline_cols,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 8: Offensive/Defensive balance
    print("\n" + "=" * 90)
    print("TEST: KenPom Offensive/Defensive Balance")
    print("=" * 90)
    results = test_feature(
        feature_name="KenPom Balance",
        feature_columns=["kenpom_balance"],
        feature_adder=add_kenpom_percentile_features,
        baseline_columns=baseline_cols,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 9: High tempo
    print("\n" + "=" * 90)
    print("TEST: High Tempo Indicator")
    print("=" * 90)
    results = test_feature(
        feature_name="High Tempo",
        feature_columns=["is_high_tempo"],
        feature_adder=add_tempo_features,
        baseline_columns=baseline_cols,
    )
    print_results(results)
    all_results.append(results)
    
    # Test 10: Low tempo
    print("\n" + "=" * 90)
    print("TEST: Low Tempo Indicator")
    print("=" * 90)
    results = test_feature(
        feature_name="Low Tempo",
        feature_columns=["is_low_tempo"],
        feature_adder=add_tempo_features,
        baseline_columns=baseline_cols,
    )
    print_results(results)
    all_results.append(results)
    
    # Final comparison
    print("\n" + "=" * 90)
    print("FINAL COMPARISON - ADDITIONAL FEATURES")
    print("=" * 90)
    print()
    
    comparison = compare_features(all_results)
    print(comparison.to_string(index=False))
    
    print("\n" + "=" * 90)
    print("SUMMARY")
    print("=" * 90)
    print()
    
    # Filter to features with meaningful improvement (>1%)
    meaningful = comparison[comparison["mae_improvement_pct"] > 1.0]
    
    if len(meaningful) > 0:
        print("Features with >1% improvement:")
        for _, row in meaningful.iterrows():
            print(f"  {row['feature']:40s}: {row['mae_improvement_pct']:+.2f}%")
    else:
        print("No additional features showed >1% improvement")
        print()
        print("Best performers (even if small):")
        top_5 = comparison.head(5)
        for _, row in top_5.iterrows():
            print(f"  {row['feature']:40s}: {row['mae_improvement_pct']:+.2f}%")


if __name__ == "__main__":
    main()
