#!/usr/bin/env python
"""
Test even more feature candidates to find all possible signals.
"""
from __future__ import annotations

import sys
from pathlib import Path

import numpy as np

sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from scripts.feature_tests.test_framework import (
    compare_features,
    print_results,
    test_feature,
)


def add_seed_3_indicator(df):
    """3-seeds only."""
    df = df.copy()
    df["is_seed_3"] = (df["seed"] == 3).astype(int)
    return df


def add_seed_4_indicator(df):
    """4-seeds only."""
    df = df.copy()
    df["is_seed_4"] = (df["seed"] == 4).astype(int)
    return df


def add_double_digit_seeds(df):
    """Seeds 10+."""
    df = df.copy()
    df["is_double_digit_seed"] = (df["seed"] >= 10).astype(int)
    return df


def add_kenpom_offensive_rank(df):
    """Offensive efficiency percentile."""
    df = df.copy()
    df["kenpom_o_percentile"] = df["kenpom_o"].rank(pct=True)
    return df


def add_kenpom_defensive_rank(df):
    """Defensive efficiency percentile."""
    df = df.copy()
    df["kenpom_d_percentile"] = df["kenpom_d"].rank(pct=True)
    return df


def add_kenpom_tempo_percentile(df):
    """Tempo percentile."""
    df = df.copy()
    df["kenpom_tempo_percentile"] = df["kenpom_adj_t"].rank(pct=True)
    return df


def add_offense_heavy(df):
    """Offense-heavy teams (O >> D)."""
    df = df.copy()
    df["offense_heavy"] = (
        df["kenpom_o"].rank(pct=True) - df["kenpom_d"].rank(pct=True)
    )
    return df


def add_defense_heavy(df):
    """Defense-heavy teams (D >> O)."""
    df = df.copy()
    df["defense_heavy"] = (
        df["kenpom_d"].rank(pct=True) - df["kenpom_o"].rank(pct=True)
    )
    return df


def add_kenpom_squared(df):
    """KenPom squared (non-linearity)."""
    df = df.copy()
    df["kenpom_net_sq"] = df["kenpom_net"] ** 2
    return df


def add_seed_cubed(df):
    """Seed cubed (more non-linearity)."""
    df = df.copy()
    df["seed_cubed"] = df["seed"] ** 3
    return df


def add_kenpom_o_x_d(df):
    """KenPom O × D interaction."""
    df = df.copy()
    df["kenpom_o_x_d"] = df["kenpom_o"] * df["kenpom_d"]
    return df


def add_seed_x_tempo(df):
    """Seed × Tempo interaction."""
    df = df.copy()
    df["seed_x_tempo"] = df["seed"] * df["kenpom_adj_t"]
    return df


def main():
    """Test additional feature candidates."""
    baseline_cols = [
        "seed", "kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"
    ]
    
    print("TESTING MORE FEATURE CANDIDATES")
    print("=" * 90)
    print()
    
    all_results = []
    
    # Individual seed indicators
    tests = [
        ("3-Seed Only", ["is_seed_3"], add_seed_3_indicator),
        ("4-Seed Only", ["is_seed_4"], add_seed_4_indicator),
        ("Double-Digit Seeds (10+)", ["is_double_digit_seed"], add_double_digit_seeds),
        
        # KenPom percentiles
        ("KenPom Offensive Percentile", ["kenpom_o_percentile"], add_kenpom_offensive_rank),
        ("KenPom Defensive Percentile", ["kenpom_d_percentile"], add_kenpom_defensive_rank),
        ("KenPom Tempo Percentile", ["kenpom_tempo_percentile"], add_kenpom_tempo_percentile),
        
        # Stylistic features
        ("Offense-Heavy (O >> D)", ["offense_heavy"], add_offense_heavy),
        ("Defense-Heavy (D >> O)", ["defense_heavy"], add_defense_heavy),
        
        # Non-linear transformations
        ("KenPom Net Squared", ["kenpom_net_sq"], add_kenpom_squared),
        ("Seed Cubed", ["seed_cubed"], add_seed_cubed),
        
        # Additional interactions
        ("KenPom O × D", ["kenpom_o_x_d"], add_kenpom_o_x_d),
        ("Seed × Tempo", ["seed_x_tempo"], add_seed_x_tempo),
    ]
    
    for name, cols, adder in tests:
        print("\n" + "=" * 90)
        print(f"TEST: {name}")
        print("=" * 90)
        results = test_feature(
            feature_name=name,
            feature_columns=cols,
            feature_adder=adder,
            baseline_columns=baseline_cols,
        )
        print_results(results)
        all_results.append(results)
    
    # Comparison
    print("\n" + "=" * 90)
    print("FINAL COMPARISON")
    print("=" * 90)
    print()
    
    comparison = compare_features(all_results)
    print(comparison.to_string(index=False))
    
    print("\n" + "=" * 90)
    print("FEATURES WITH >2% IMPROVEMENT")
    print("=" * 90)
    print()
    
    meaningful = comparison[comparison["mae_improvement_pct"] > 2.0]
    
    if len(meaningful) > 0:
        for _, row in meaningful.iterrows():
            improvement_pct = row['mae_improvement_pct']
            years = row['years_improved']
            print(f"  {row['feature']:40s}: {improvement_pct:+6.2f}% ({years})")
    else:
        print("No features showed >2% improvement")


if __name__ == "__main__":
    main()
