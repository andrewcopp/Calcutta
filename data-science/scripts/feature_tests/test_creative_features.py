#!/usr/bin/env python
"""
Test creative new features to see if anything else improves the model.

Current best: champ_equity + kenpom_net_pct + kenpom_balance (MAE: 0.006499)

New ideas to test:
1. Interaction terms (champ_equity × kenpom features)
2. Non-linear KenPom transformations (log, sqrt, squared)
3. Seed × KenPom interactions
4. Recent tournament success proxies
5. Extreme value indicators
6. Ratio features
"""
from __future__ import annotations

import sys
from pathlib import Path

import numpy as np
import pandas as pd

sys.path.insert(0, str(Path(__file__).parent.parent.parent))

from scripts.feature_tests.test_framework import load_year_data


def add_creative_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add creative new feature ideas."""
    df = df.copy()
    
    # Existing features we know work
    df["kenpom_net_pct"] = df["kenpom_net"].rank(pct=True)
    df["kenpom_o_pct"] = df["kenpom_o"].rank(pct=True)
    df["kenpom_d_pct"] = df["kenpom_d"].rank(pct=True)
    
    df["kenpom_balance"] = np.abs(
        df["kenpom_o"].rank(pct=True) - df["kenpom_d"].rank(pct=True)
    )
    
    seed_title_prob = {
        1: 0.20, 2: 0.12, 3: 0.08, 4: 0.05, 5: 0.03, 6: 0.02,
        7: 0.01, 8: 0.01, 9: 0.005, 10: 0.003, 11: 0.002, 12: 0.001,
        13: 0.0005, 14: 0.0002, 15: 0.0001, 16: 0.00001
    }
    df["champ_equity"] = df["seed"].map(seed_title_prob)
    
    # NEW IDEA 1: Interaction between championship equity and strength
    df["champ_x_kenpom"] = df["champ_equity"] * df["kenpom_net_pct"]
    df["champ_x_balance"] = df["champ_equity"] * df["kenpom_balance"]
    
    # NEW IDEA 2: Disagreement features (seed vs KenPom)
    # Expected KenPom percentile by seed
    seed_expected_pct = {}
    for seed in range(1, 17):
        # 1-seeds should be ~top 6% (4/68), 16-seeds bottom 6%
        seed_expected_pct[seed] = 1.0 - (seed - 0.5) / 16.0
    
    df["expected_kenpom_pct"] = df["seed"].map(seed_expected_pct)
    df["kenpom_surprise"] = df["kenpom_net_pct"] - df["expected_kenpom_pct"]
    df["kenpom_surprise_abs"] = np.abs(df["kenpom_surprise"])
    
    # NEW IDEA 3: Elite vs non-elite indicators
    df["is_elite"] = (df["seed"] <= 2).astype(int)
    df["is_top_kenpom"] = (df["kenpom_net_pct"] >= 0.9).astype(int)
    df["elite_and_strong"] = df["is_elite"] * df["is_top_kenpom"]
    
    # NEW IDEA 4: Offensive/defensive style extremes
    df["offense_dominant"] = (df["kenpom_o_pct"] > 0.75).astype(int)
    df["defense_dominant"] = (df["kenpom_d_pct"] > 0.75).astype(int)
    df["balanced_elite"] = (
        (df["kenpom_balance"] < 0.1) & (df["kenpom_net_pct"] > 0.7)
    ).astype(int)
    
    # NEW IDEA 5: Non-linear KenPom transformations
    # Log transform (shift to avoid log(0))
    kenpom_shifted = df["kenpom_net"] - df["kenpom_net"].min() + 1
    df["kenpom_log"] = np.log(kenpom_shifted)
    df["kenpom_sqrt"] = np.sqrt(kenpom_shifted)
    df["kenpom_squared"] = df["kenpom_net"] ** 2
    
    # NEW IDEA 6: Ratio features
    df["kenpom_o_to_d_ratio"] = df["kenpom_o"] / (df["kenpom_d"] + 100)
    df["efficiency_per_tempo"] = df["kenpom_net"] / (df["kenpom_adj_t"] + 1)
    
    # NEW IDEA 7: Seed-based expected points
    # Approximate expected tournament points by seed
    seed_expected_points = {
        1: 12, 2: 9, 3: 7, 4: 5, 5: 4, 6: 3, 7: 2, 8: 2,
        9: 1, 10: 1, 11: 1, 12: 1, 13: 0.5, 14: 0.3, 15: 0.2, 16: 0.1
    }
    df["expected_points"] = df["seed"].map(seed_expected_points)
    df["points_per_equity"] = df["expected_points"] / (df["champ_equity"] + 0.001)
    
    # NEW IDEA 8: Percentile interactions
    df["kenpom_pct_squared"] = df["kenpom_net_pct"] ** 2
    df["kenpom_pct_cubed"] = df["kenpom_net_pct"] ** 3
    
    # NEW IDEA 9: Extreme seed indicators
    df["is_1_seed"] = (df["seed"] == 1).astype(int)
    df["is_16_seed"] = (df["seed"] == 16).astype(int)
    df["is_double_digit"] = (df["seed"] >= 10).astype(int)
    
    # NEW IDEA 10: Championship equity non-linearity
    df["champ_equity_squared"] = df["champ_equity"] ** 2
    df["champ_equity_log"] = np.log(df["champ_equity"] + 0.0001)
    
    return df


def test_new_features():
    """Test all new creative features."""
    import numpy as np
    import pandas as pd
    from sklearn.linear_model import Ridge
    from sklearn.preprocessing import StandardScaler
    
    years = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
    
    # Current best 3 features
    current_best = ["champ_equity", "kenpom_net_pct", "kenpom_balance"]
    
    # All new creative features
    new_features = [
        "champ_x_kenpom", "champ_x_balance",
        "kenpom_surprise", "kenpom_surprise_abs",
        "is_elite", "is_top_kenpom", "elite_and_strong",
        "offense_dominant", "defense_dominant", "balanced_elite",
        "kenpom_log", "kenpom_sqrt", "kenpom_squared",
        "kenpom_o_to_d_ratio", "efficiency_per_tempo",
        "expected_points", "points_per_equity",
        "kenpom_pct_squared", "kenpom_pct_cubed",
        "is_1_seed", "is_16_seed", "is_double_digit",
        "champ_equity_squared", "champ_equity_log",
    ]
    
    print("TESTING CREATIVE NEW FEATURES")
    print("=" * 90)
    print(f"Current best (3 features): {current_best}")
    print(f"Testing {len(new_features)} new creative features")
    print()
    
    # Evaluate current best
    def evaluate(features):
        maes = []
        for test_year in years:
            train_years = [y for y in years if y != test_year]
            train_data = [load_year_data(y) for y in train_years]
            train_df = pd.concat(train_data, ignore_index=True)
            test_df = load_year_data(test_year)
            
            train_df = add_creative_features(train_df)
            test_df = add_creative_features(test_df)
            
            X_train = train_df[features].fillna(0)
            X_test = test_df[features].fillna(0)
            y_train = train_df["actual_share"].values
            y_test = test_df["actual_share"].values
            
            scaler = StandardScaler()
            X_train_scaled = scaler.fit_transform(X_train)
            X_test_scaled = scaler.transform(X_test)
            
            model = Ridge(alpha=1.0)
            model.fit(X_train_scaled, y_train)
            pred = model.predict(X_test_scaled)
            
            mae = np.abs(pred - y_test).mean()
            maes.append(mae)
        
        return np.mean(maes)
    
    baseline_mae = evaluate(current_best)
    print(f"Baseline MAE (3 features): {baseline_mae:.6f}")
    print()
    
    # Test each new feature added to current best
    print("Testing each new feature added to current best:")
    print("-" * 90)
    
    improvements = []
    for feat in new_features:
        test_features = current_best + [feat]
        mae = evaluate(test_features)
        improvement = baseline_mae - mae
        improvements.append((feat, mae, improvement))
    
    # Sort by improvement
    improvements.sort(key=lambda x: x[2], reverse=True)
    
    print(f"{'Feature':<30s} {'MAE':>10s} {'Improvement':>12s}")
    print("-" * 90)
    for feat, mae, imp in improvements[:15]:  # Top 15
        symbol = "✓" if imp > 0.0001 else " "
        print(f"{symbol} {feat:<28s} {mae:>10.6f} {imp:>+12.6f}")
    
    print()
    print("=" * 90)
    print("SUMMARY")
    print("=" * 90)
    
    best_new = improvements[0]
    if best_new[2] > 0.0001:
        print(f"Best new feature: {best_new[0]}")
        print(f"  Improvement: {best_new[2]:+.6f}")
        print(f"  New MAE: {best_new[1]:.6f}")
        print()
        print("✓ This feature adds value! Consider including it.")
    else:
        print("No new features improved MAE by >0.0001")
        print()
        print("The 3-feature set appears optimal:")
        print("  1. champ_equity")
        print("  2. kenpom_net_pct")
        print("  3. kenpom_balance")


if __name__ == "__main__":
    test_new_features()
