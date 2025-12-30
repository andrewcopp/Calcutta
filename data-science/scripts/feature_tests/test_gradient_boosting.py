#!/usr/bin/env python
"""
Test gradient boosting models vs Ridge regression.

Compare:
1. Ridge with optimal engineered features (baseline)
2. XGBoost with raw features
3. XGBoost with engineered features
4. LightGBM with raw features
5. LightGBM with engineered features
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

from sklearn.ensemble import GradientBoostingRegressor, RandomForestRegressor


def add_all_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add all engineered features."""
    df = df.copy()
    
    # Optimal features
    seed_title_prob = {
        1: 0.20, 2: 0.12, 3: 0.08, 4: 0.05, 5: 0.03, 6: 0.02,
        7: 0.01, 8: 0.01, 9: 0.005, 10: 0.003, 11: 0.002,
        12: 0.001, 13: 0.0005, 14: 0.0002, 15: 0.0001, 16: 0.00001
    }
    df['champ_equity'] = df['seed'].map(seed_title_prob)
    
    df['kenpom_net_pct'] = df['kenpom_net'].rank(pct=True)
    df['kenpom_o_pct'] = df['kenpom_o'].rank(pct=True)
    df['kenpom_d_pct'] = df['kenpom_d'].rank(pct=True)
    df['kenpom_balance'] = np.abs(df['kenpom_o_pct'] - df['kenpom_d_pct'])
    
    seed_expected_points = {
        1: 12, 2: 9, 3: 7, 4: 5, 5: 4, 6: 3, 7: 2, 8: 2,
        9: 1, 10: 1, 11: 1, 12: 1, 13: 0.5, 14: 0.3, 15: 0.2, 16: 0.1
    }
    df['expected_points'] = df['seed'].map(seed_expected_points)
    df['points_per_equity'] = df['expected_points'] / (df['champ_equity'] + 0.001)
    
    df['kenpom_pct_cubed'] = df['kenpom_net_pct'] ** 3
    
    # Additional engineered features
    df['seed_sq'] = df['seed'] ** 2
    df['seed_cubed'] = df['seed'] ** 3
    df['kenpom_sq'] = df['kenpom_net'] ** 2
    df['kenpom_x_seed'] = df['kenpom_net'] * df['seed']
    
    return df


def test_models():
    """Test all models with LOOCV."""
    years = [2017, 2018, 2019, 2021, 2022, 2023, 2024, 2025]
    
    # Feature sets
    raw_features = ['seed', 'kenpom_net', 'kenpom_o', 'kenpom_d', 'kenpom_adj_t']
    
    optimal_features = [
        'champ_equity', 'kenpom_net_pct', 'kenpom_balance',
        'points_per_equity', 'kenpom_pct_cubed'
    ]
    
    all_engineered = optimal_features + [
        'kenpom_o_pct', 'kenpom_d_pct', 'expected_points',
        'seed_sq', 'seed_cubed', 'kenpom_sq', 'kenpom_x_seed'
    ]
    
    print('GRADIENT BOOSTING vs RIDGE REGRESSION')
    print('='*90)
    print()
    
    results = []
    
    # 1. Ridge with optimal features (baseline)
    print('Testing Ridge with optimal features...')
    mae = evaluate_ridge(optimal_features, years)
    results.append(('Ridge (optimal 5 features)', mae, len(optimal_features)))
    print(f'  MAE: {mae:.6f}')
    print()
    
    # 2. Ridge with all engineered
    print('Testing Ridge with all engineered features...')
    mae = evaluate_ridge(all_engineered, years)
    results.append(('Ridge (all engineered)', mae, len(all_engineered)))
    print(f'  MAE: {mae:.6f}')
    print()
    
    # 3. Gradient Boosting with raw features
    print('Testing Gradient Boosting with raw features...')
    mae = evaluate_gradient_boosting(raw_features, years)
    results.append(('GradientBoosting (raw)', mae, len(raw_features)))
    print(f'  MAE: {mae:.6f}')
    print()
    
    # 4. Gradient Boosting with engineered features
    print('Testing Gradient Boosting with engineered features...')
    mae = evaluate_gradient_boosting(all_engineered, years)
    results.append(('GradientBoosting (engineered)', mae, len(all_engineered)))
    print(f'  MAE: {mae:.6f}')
    print()
    
    # 5. Random Forest with raw features
    print('Testing Random Forest with raw features...')
    mae = evaluate_random_forest(raw_features, years)
    results.append(('RandomForest (raw)', mae, len(raw_features)))
    print(f'  MAE: {mae:.6f}')
    print()
    
    # 6. Random Forest with engineered features
    print('Testing Random Forest with engineered features...')
    mae = evaluate_random_forest(all_engineered, years)
    results.append(('RandomForest (engineered)', mae, len(all_engineered)))
    print(f'  MAE: {mae:.6f}')
    print()
    
    # Summary
    print('='*90)
    print('RESULTS SUMMARY')
    print('='*90)
    print()
    
    results.sort(key=lambda x: x[1])
    
    print(f"{'Model':<30s} {'MAE':>12s} {'# Features':>12s} {'vs Best':>12s}")
    print('-'*90)
    
    best_mae = results[0][1]
    for model, mae, n_feat in results:
        diff = mae - best_mae
        symbol = '✓' if mae == best_mae else ' '
        print(f'{symbol} {model:<28s} {mae:>12.6f} {n_feat:>12d} {diff:>+12.6f}')
    
    print()
    print('='*90)
    print('RECOMMENDATION')
    print('='*90)
    print()
    
    best = results[0]
    print(f'Best model: {best[0]}')
    print(f'  MAE: {best[1]:.6f}')
    print(f'  Features: {best[2]}')
    print()
    
    ridge_optimal = [r for r in results if 'Ridge (optimal' in r[0]][0]
    improvement = (ridge_optimal[1] - best[1]) / ridge_optimal[1] * 100
    
    if best[0].startswith('Ridge'):
        print('✓ Ridge regression remains best choice')
        print('  - Simpler, more interpretable')
        print('  - Less prone to overfitting with small dataset')
    else:
        print(f'✓ {best[0]} improves over Ridge by {improvement:.1f}%')
        print('  - Consider switching if improvement is substantial')
        print('  - Trade-off: complexity vs performance')


def evaluate_ridge(features, years):
    """Evaluate Ridge with LOOCV."""
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


def evaluate_gradient_boosting(features, years):
    """Evaluate sklearn GradientBoostingRegressor with LOOCV."""
    maes = []
    
    for test_year in years:
        train_years = [y for y in years if y != test_year]
        train_data = [load_year_data(y) for y in train_years]
        train_df = pd.concat(train_data, ignore_index=True)
        test_df = load_year_data(test_year)
        
        train_df = add_all_features(train_df)
        test_df = add_all_features(test_df)
        
        X_train = train_df[features].fillna(0).values
        X_test = test_df[features].fillna(0).values
        y_train = train_df['actual_share'].values
        y_test = test_df['actual_share'].values
        
        # Conservative hyperparameters to avoid overfitting
        model = GradientBoostingRegressor(
            n_estimators=100,
            max_depth=3,
            learning_rate=0.1,
            subsample=0.8,
            min_samples_split=10,
            min_samples_leaf=5,
            random_state=42
        )
        
        model.fit(X_train, y_train)
        pred = model.predict(X_test)
        
        mae = np.abs(pred - y_test).mean()
        maes.append(mae)
    
    return np.mean(maes)


def evaluate_random_forest(features, years):
    """Evaluate RandomForestRegressor with LOOCV."""
    maes = []
    
    for test_year in years:
        train_years = [y for y in years if y != test_year]
        train_data = [load_year_data(y) for y in train_years]
        train_df = pd.concat(train_data, ignore_index=True)
        test_df = load_year_data(test_year)
        
        train_df = add_all_features(train_df)
        test_df = add_all_features(test_df)
        
        X_train = train_df[features].fillna(0).values
        X_test = test_df[features].fillna(0).values
        y_train = train_df['actual_share'].values
        y_test = test_df['actual_share'].values
        
        # Conservative hyperparameters
        model = RandomForestRegressor(
            n_estimators=100,
            max_depth=5,
            min_samples_split=10,
            min_samples_leaf=5,
            max_features='sqrt',
            random_state=42
        )
        
        model.fit(X_train, y_train)
        pred = model.predict(X_test)
        
        mae = np.abs(pred - y_test).mean()
        maes.append(mae)
    
    return np.mean(maes)


if __name__ == '__main__':
    test_models()
