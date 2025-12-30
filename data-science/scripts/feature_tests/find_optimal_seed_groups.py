#!/usr/bin/env python
"""
Use decision tree to mathematically determine optimal seed groupings.

Instead of arbitrary groupings, let the data tell us where the natural
breakpoints are in how the market prices different seeds.
"""
from __future__ import annotations

import sys
from pathlib import Path

import numpy as np
import pandas as pd
from sklearn.tree import DecisionTreeRegressor

sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from scripts.feature_tests.test_framework import load_year_data


def find_seed_splits_with_tree(max_depth: int = 3) -> dict:
    """
    Use decision tree to find optimal seed groupings.
    
    Train a decision tree to predict market share using only seed,
    then extract the split points to define groupings.
    """
    years = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
    
    # Load all data
    all_data = []
    for year in years:
        df = load_year_data(year)
        all_data.append(df)
    
    combined = pd.concat(all_data, ignore_index=True)
    
    # Train decision tree on seed only
    X = combined[["seed"]].values
    y = combined["actual_share"].values
    
    tree = DecisionTreeRegressor(
        max_depth=max_depth,
        min_samples_split=20,
        min_samples_leaf=10,
        random_state=42,
    )
    tree.fit(X, y)
    
    # Extract split points
    tree_structure = tree.tree_
    feature = tree_structure.feature
    threshold = tree_structure.threshold
    
    # Find all splits on seed (feature 0)
    splits = []
    for i in range(tree_structure.node_count):
        if feature[i] == 0:  # Split on seed
            splits.append(threshold[i])
    
    splits = sorted(set(splits))
    
    print("DECISION TREE SEED SPLITS")
    print("=" * 90)
    print(f"Max depth: {max_depth}")
    print(f"Found {len(splits)} split points:")
    for i, split in enumerate(splits, 1):
        print(f"  {i}. Seed <= {split:.1f}")
    print()
    
    # Create groupings from splits
    splits_int = [int(np.ceil(s)) for s in splits]
    
    # Define groups
    groups = []
    prev = 1
    for split in splits_int:
        if split > prev:
            groups.append((prev, split))
            prev = split + 1
    groups.append((prev, 16))
    
    print("RESULTING SEED GROUPS:")
    for i, (start, end) in enumerate(groups, 1):
        if start == end:
            print(f"  Group {i}: Seed {start}")
        else:
            print(f"  Group {i}: Seeds {start}-{end}")
    print()
    
    # Analyze each group
    print("GROUP STATISTICS:")
    print("-" * 90)
    for i, (start, end) in enumerate(groups, 1):
        group_data = combined[
            (combined["seed"] >= start) & (combined["seed"] <= end)
        ]
        mean_share = group_data["actual_share"].mean()
        std_share = group_data["actual_share"].std()
        n_teams = len(group_data)
        
        if start == end:
            label = f"Seed {start}"
        else:
            label = f"Seeds {start}-{end}"
        
        print(f"{label:20s}: "
              f"Mean={mean_share:.4f}, "
              f"Std={std_share:.4f}, "
              f"N={n_teams}")
    
    return {
        "splits": splits,
        "groups": groups,
        "tree": tree,
    }


def create_group_indicators(df: pd.DataFrame, groups: list) -> pd.DataFrame:
    """Create indicator variables for each seed group."""
    df = df.copy()
    
    for i, (start, end) in enumerate(groups, 1):
        col_name = f"seed_group_{i}"
        df[col_name] = (
            (df["seed"] >= start) & (df["seed"] <= end)
        ).astype(int)
    
    return df


def test_seed_groupings(groups: list) -> None:
    """Test the data-driven seed groupings."""
    from scripts.feature_tests.test_framework import (
        test_feature,
        print_results,
    )
    
    # Create feature adder
    def add_groups(df):
        return create_group_indicators(df, groups)
    
    # Get group column names
    group_cols = [f"seed_group_{i}" for i in range(1, len(groups) + 1)]
    
    # Test
    baseline = ["kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"]
    
    print("\n" + "=" * 90)
    print("TESTING DATA-DRIVEN SEED GROUPS")
    print("=" * 90)
    
    results = test_feature(
        feature_name="Data-Driven Seed Groups",
        feature_columns=group_cols,
        feature_adder=add_groups,
        baseline_columns=baseline,
    )
    
    print_results(results)
    
    return results


def compare_approaches() -> None:
    """Compare different seed encoding approaches."""
    from scripts.feature_tests.test_framework import (
        test_feature,
        compare_features,
    )
    
    baseline = ["kenpom_net", "kenpom_o", "kenpom_d", "kenpom_adj_t"]
    
    all_results = []
    
    # 1. Continuous seed
    print("\n" + "=" * 90)
    print("TEST: Continuous Seed")
    print("=" * 90)
    results = test_feature(
        feature_name="Continuous Seed",
        feature_columns=["seed"],
        baseline_columns=baseline,
    )
    all_results.append(results)
    
    # 2. Categorical seed (16 dummies)
    def add_seed_dummies(df):
        df = df.copy()
        seed_dummies = pd.get_dummies(df["seed"], prefix="seed", drop_first=True)
        for col in seed_dummies.columns:
            df[col] = seed_dummies[col]
        return df
    
    seed_dummy_cols = [f"seed_{i}" for i in range(2, 17)]
    
    print("\n" + "=" * 90)
    print("TEST: Categorical Seed (16 dummies)")
    print("=" * 90)
    results = test_feature(
        feature_name="Categorical Seed (16 dummies)",
        feature_columns=seed_dummy_cols,
        feature_adder=add_seed_dummies,
        baseline_columns=baseline,
    )
    all_results.append(results)
    
    # 3. Data-driven groups (from decision tree)
    print("\n" + "=" * 90)
    print("FINDING OPTIMAL SEED GROUPS...")
    print("=" * 90)
    result = find_seed_splits_with_tree(max_depth=3)
    groups = result["groups"]
    
    results = test_seed_groupings(groups)
    all_results.append(results)
    
    # Comparison
    print("\n" + "=" * 90)
    print("FINAL COMPARISON")
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
    print(f"  Years improved: {best['years_improved']}")
    print()
    
    if "Data-Driven" in best['feature']:
        print("✓ Use DATA-DRIVEN seed groups")
        print("  Groups determined mathematically by decision tree")
        print("  Balances granularity with statistical stability")
        print()
        print("Optimal groups:")
        for i, (start, end) in enumerate(groups, 1):
            if start == end:
                print(f"  Group {i}: Seed {start}")
            else:
                print(f"  Group {i}: Seeds {start}-{end}")


def main():
    """Run complete analysis."""
    compare_approaches()


if __name__ == "__main__":
    main()
