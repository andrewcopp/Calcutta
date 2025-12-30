#!/usr/bin/env python
"""
Test market memory features.

Hypothesis: Market has memory of prior year's investment behavior
- Teams that were expensive last year may be avoided this year
- Teams that disappointed (high investment, early loss) may be underpriced
- Teams that won championship may face "back-to-back" skepticism

This is DIFFERENT from our previous "last year performance" features which
looked at ROI/progress. This looks at MARKET BEHAVIOR memory.
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


def add_market_memory_features(df: pd.DataFrame) -> pd.DataFrame:
    """
    Add features capturing market memory from prior year.
    
    Features:
    1. was_expensive_last_year - High investment last year (top quartile)
    2. disappointed_last_year - High investment but early exit
    3. won_championship_last_year - Won it all last year
    4. missed_tournament_last_year - Didn't make tournament last year
    5. last_year_investment_percentile - Where they ranked in investment
    """
    df = df.copy()
    
    # These should already be in the data from team_dataset
    # But let's create derived features
    
    # 1. Was expensive last year (top 25% of investment)
    if 'team_share_of_pool_last_year' in df.columns:
        df['was_expensive_last_year'] = (
            df['team_share_of_pool_last_year'] > 
            df['team_share_of_pool_last_year'].quantile(0.75)
        ).astype(int)
    else:
        df['was_expensive_last_year'] = 0
    
    # 2. Disappointed last year (high investment, early exit)
    # High investment = top 50%, early exit = progress < 2 (lost in R1 or R2)
    if 'team_share_of_pool_last_year' in df.columns and 'progress_last_year' in df.columns:
        high_investment = df['team_share_of_pool_last_year'] > df['team_share_of_pool_last_year'].median()
        early_exit = df['progress_last_year'] < 2
        df['disappointed_last_year'] = (high_investment & early_exit).astype(int)
    else:
        df['disappointed_last_year'] = 0
    
    # 3. Won championship last year
    if 'wins_last_year' in df.columns:
        df['won_championship_last_year'] = (df['wins_last_year'] >= 6).astype(int)
    else:
        df['won_championship_last_year'] = 0
    
    # 4. Missed tournament last year
    if 'has_last_year' in df.columns:
        df['missed_tournament_last_year'] = (df['has_last_year'] == 0).astype(int)
    else:
        df['missed_tournament_last_year'] = 0
    
    # 5. Last year investment percentile
    if 'team_share_of_pool_last_year' in df.columns:
        df['last_year_investment_pct'] = df['team_share_of_pool_last_year'].rank(pct=True)
    else:
        df['last_year_investment_pct'] = 0
    
    # 6. Overpriced last year (investment > expected based on seed)
    if 'team_share_of_pool_last_year' in df.columns and 'seed' in df.columns:
        # Expected investment by seed (rough approximation)
        seed_expected = {
            1: 0.055, 2: 0.040, 3: 0.030, 4: 0.022, 5: 0.016, 6: 0.012,
            7: 0.009, 8: 0.007, 9: 0.005, 10: 0.004, 11: 0.003, 12: 0.002,
            13: 0.001, 14: 0.001, 15: 0.0005, 16: 0.0002
        }
        df['expected_investment_last_year'] = df['seed'].map(seed_expected).fillna(0)
        df['overpriced_last_year'] = (
            df['team_share_of_pool_last_year'] > df['expected_investment_last_year'] * 1.5
        ).astype(int)
    else:
        df['overpriced_last_year'] = 0
    
    return df


def test_market_memory_features():
    """Test market memory features individually."""
    years = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
    
    # Current optimal features
    optimal_features = [
        'champ_equity', 'kenpom_net_pct', 'kenpom_balance',
        'points_per_equity', 'kenpom_pct_cubed'
    ]
    
    # Market memory features to test
    memory_features = [
        'was_expensive_last_year',
        'disappointed_last_year',
        'won_championship_last_year',
        'missed_tournament_last_year',
        'last_year_investment_pct',
        'overpriced_last_year',
    ]
    
    print('TESTING MARKET MEMORY FEATURES')
    print('='*90)
    print()
    print('Hypothesis: Market has memory of prior year investment behavior')
    print('  - Expensive teams last year may be avoided')
    print('  - Disappointing teams may be underpriced')
    print('  - Champions may face "back-to-back" skepticism')
    print()
    
    def add_all_features(df):
        # Add optimal features
        df = df.copy()
        
        # Championship equity
        seed_title_prob = {
            1: 0.20, 2: 0.12, 3: 0.08, 4: 0.05, 5: 0.03, 6: 0.02,
            7: 0.01, 8: 0.01, 9: 0.005, 10: 0.003, 11: 0.002,
            12: 0.001, 13: 0.0005, 14: 0.0002, 15: 0.0001, 16: 0.00001
        }
        df['champ_equity'] = df['seed'].map(seed_title_prob)
        
        # KenPom percentiles
        df['kenpom_net_pct'] = df['kenpom_net'].rank(pct=True)
        kenpom_o_pct = df['kenpom_o'].rank(pct=True)
        kenpom_d_pct = df['kenpom_d'].rank(pct=True)
        df['kenpom_balance'] = np.abs(kenpom_o_pct - kenpom_d_pct)
        
        # Points per equity
        seed_expected_points = {
            1: 12, 2: 9, 3: 7, 4: 5, 5: 4, 6: 3, 7: 2, 8: 2,
            9: 1, 10: 1, 11: 1, 12: 1, 13: 0.5, 14: 0.3, 15: 0.2, 16: 0.1
        }
        df['expected_points'] = df['seed'].map(seed_expected_points)
        df['points_per_equity'] = df['expected_points'] / (df['champ_equity'] + 0.001)
        
        # KenPom cubed
        df['kenpom_pct_cubed'] = df['kenpom_net_pct'] ** 3
        
        # Market memory features
        df = add_market_memory_features(df)
        
        return df
    
    def evaluate(features):
        maes = []
        for test_year in years:
            train_years = [y for y in years if y != test_year]
            train_data = [load_year_data(y) for y in train_years]
            train_df = pd.concat(train_data, ignore_index=True)
            test_df = load_year_data(test_year)
            
            train_df = add_all_features(train_df)
            test_df = add_all_features(test_df)
            
            X_train = train_df[features].fillna(0)
            X_test = test_df[features].fillna(0)
            y_train = train_df['actual_share'].values
            y_test = test_df['actual_share'].values
            
            scaler = StandardScaler()
            X_train_scaled = scaler.fit_transform(X_train)
            X_test_scaled = scaler.transform(X_test)
            
            model = Ridge(alpha=1.0)
            model.fit(X_train_scaled, y_train)
            pred = model.predict(X_test_scaled)
            
            mae = np.abs(pred - y_test).mean()
            maes.append(mae)
        
        return np.mean(maes)
    
    # Baseline with optimal features
    baseline_mae = evaluate(optimal_features)
    print(f'Baseline (optimal 5 features): MAE = {baseline_mae:.6f}')
    print()
    
    # Test each memory feature
    print('Testing each market memory feature added to optimal set:')
    print('-'*90)
    print(f"{'Feature':<35s} {'MAE':>10s} {'Improvement':>12s}")
    print('-'*90)
    
    results = []
    for feat in memory_features:
        test_features = optimal_features + [feat]
        mae = evaluate(test_features)
        improvement = baseline_mae - mae
        results.append((feat, mae, improvement))
        
        symbol = '✓' if improvement > 0.0001 else ' '
        print(f'{symbol} {feat:<33s} {mae:>10.6f} {improvement:>+12.6f}')
    
    print()
    print('='*90)
    print('ANALYSIS')
    print('='*90)
    
    best = max(results, key=lambda x: x[2])
    if best[2] > 0.0001:
        print(f'Best market memory feature: {best[0]}')
        print(f'  Improvement: {best[2]:+.6f} ({best[2]/baseline_mae*100:.1f}%)')
        print(f'  New MAE: {best[1]:.6f}')
        print()
        print('✓ Market memory effects exist! This feature adds value.')
    else:
        print('No market memory features improved MAE by >0.0001')
        print()
        print('Market appears to have no systematic memory of prior year behavior.')
        print('Each year is treated independently by the market.')


def analyze_specific_cases():
    """Analyze the specific cases user mentioned."""
    print()
    print('='*90)
    print('SPECIFIC CASE ANALYSIS')
    print('='*90)
    print()
    
    cases = [
        (2022, 'Baylor', '2021 champion, underpriced in 2022'),
        (2024, 'North Carolina', 'Missed 2023 tournament, underpriced in 2024'),
        (2024, 'Arizona', 'Overpriced in 2023, avoided in 2024?'),
    ]
    
    for year, team, hypothesis in cases:
        print(f'{year} {team}: {hypothesis}')
        print('-'*90)
        
        # Load data
        df = load_year_data(year)
        df = add_market_memory_features(df)
        
        team_row = df[df['school_name'].str.contains(team, case=False, na=False)]
        
        if len(team_row) > 0:
            row = team_row.iloc[0]
            
            print(f"  Current year investment: {row.get('actual_share', 0)*100:.2f}%")
            print(f"  Last year investment: {row.get('team_share_of_pool_last_year', 0)*100:.2f}%")
            print(f"  Last year progress: {row.get('progress_last_year', 0):.1f}")
            print(f"  Last year wins: {row.get('wins_last_year', 0):.0f}")
            print(f"  Won championship last year: {row.get('won_championship_last_year', 0)}")
            print(f"  Missed tournament last year: {row.get('missed_tournament_last_year', 0)}")
            print(f"  Was expensive last year: {row.get('was_expensive_last_year', 0)}")
            print(f"  Disappointed last year: {row.get('disappointed_last_year', 0)}")
        
        print()


if __name__ == '__main__':
    test_market_memory_features()
    analyze_specific_cases()
