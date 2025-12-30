#!/usr/bin/env python
"""
Test market memory features with proper year-over-year joins.

Since last_year columns don't exist, we need to:
1. Load current year data
2. Load prior year data
3. Join on school_name to get prior year market behavior
4. Create memory features from the join
"""
from __future__ import annotations

import sys
from pathlib import Path

import numpy as np
import pandas as pd
from sklearn.linear_model import Ridge
from sklearn.preprocessing import StandardScaler

sys.path.insert(0, str(Path(__file__).parent.parent.parent))


def load_year_with_prior(year: int) -> pd.DataFrame:
    """Load year data and join with prior year market data."""
    # Load current year
    current_file = Path(f'out/{year}/derived/team_dataset.parquet')
    if not current_file.exists():
        return None
    
    current = pd.read_parquet(current_file)
    
    # Load prior year
    prior_year = year - 1
    if prior_year == 2020:  # No tournament in 2020
        prior_year = 2019
    
    prior_file = Path(f'out/{prior_year}/derived/team_dataset.parquet')
    if not prior_file.exists():
        # No prior year data, return current with nulls
        current['prior_year_share'] = 0.0
        current['prior_year_seed'] = 0
        current['prior_year_progress'] = 0.0
        current['prior_year_wins'] = 0
        current['had_prior_year'] = 0
        return current
    
    prior = pd.read_parquet(prior_file)
    
    # Get share column name
    share_col = 'team_share_of_pool' if 'team_share_of_pool' in prior.columns else 'actual_share'
    
    # Select relevant prior year columns
    prior_cols = {
        'school_name': 'school_name',
        share_col: 'prior_year_share',
        'seed': 'prior_year_seed',
    }
    
    # Add progress/wins if available
    if 'progress' in prior.columns:
        prior_cols['progress'] = 'prior_year_progress'
    if 'wins' in prior.columns:
        prior_cols['wins'] = 'prior_year_wins'
    
    prior_subset = prior[list(prior_cols.keys())].copy()
    prior_subset = prior_subset.rename(columns=prior_cols)
    
    # Join
    merged = current.merge(
        prior_subset,
        on='school_name',
        how='left'
    )
    
    # Fill nulls (teams that didn't make prior tournament)
    merged['prior_year_share'] = merged['prior_year_share'].fillna(0)
    merged['prior_year_seed'] = merged['prior_year_seed'].fillna(0)
    if 'prior_year_progress' in merged.columns:
        merged['prior_year_progress'] = merged['prior_year_progress'].fillna(0)
    else:
        merged['prior_year_progress'] = 0
    if 'prior_year_wins' in merged.columns:
        merged['prior_year_wins'] = merged['prior_year_wins'].fillna(0)
    else:
        merged['prior_year_wins'] = 0
    merged['had_prior_year'] = (merged['prior_year_share'] > 0).astype(int)
    
    return merged


def add_market_memory_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add market memory features from prior year data."""
    df = df.copy()
    
    # 1. Was expensive last year (top 25% of investment)
    if 'prior_year_share' in df.columns:
        threshold = df[df['prior_year_share'] > 0]['prior_year_share'].quantile(0.75)
        df['was_expensive_last_year'] = (
            df['prior_year_share'] > threshold
        ).astype(int)
    else:
        df['was_expensive_last_year'] = 0
    
    # 2. Disappointed last year (high investment, early exit)
    if 'prior_year_share' in df.columns and 'prior_year_progress' in df.columns:
        median_share = df[df['prior_year_share'] > 0]['prior_year_share'].median()
        high_investment = df['prior_year_share'] > median_share
        early_exit = df['prior_year_progress'] < 2
        df['disappointed_last_year'] = (high_investment & early_exit).astype(int)
    else:
        df['disappointed_last_year'] = 0
    
    # 3. Won championship last year
    if 'prior_year_wins' in df.columns:
        df['won_championship_last_year'] = (df['prior_year_wins'] >= 6).astype(int)
    else:
        df['won_championship_last_year'] = 0
    
    # 4. Missed tournament last year
    if 'had_prior_year' in df.columns:
        df['missed_tournament_last_year'] = (df['had_prior_year'] == 0).astype(int)
    else:
        df['missed_tournament_last_year'] = 0
    
    # 5. Last year investment percentile
    if 'prior_year_share' in df.columns:
        df['last_year_investment_pct'] = df['prior_year_share'].rank(pct=True)
    else:
        df['last_year_investment_pct'] = 0
    
    # 6. Overpriced last year
    if 'prior_year_share' in df.columns and 'prior_year_seed' in df.columns:
        seed_expected = {
            1: 0.055, 2: 0.040, 3: 0.030, 4: 0.022, 5: 0.016, 6: 0.012,
            7: 0.009, 8: 0.007, 9: 0.005, 10: 0.004, 11: 0.003, 12: 0.002,
            13: 0.001, 14: 0.001, 15: 0.0005, 16: 0.0002
        }
        df['expected_prior'] = df['prior_year_seed'].map(seed_expected).fillna(0)
        df['overpriced_last_year'] = (
            df['prior_year_share'] > df['expected_prior'] * 1.5
        ).astype(int)
        df = df.drop(columns=['expected_prior'])
    else:
        df['overpriced_last_year'] = 0
    
    return df


def test_market_memory_features():
    """Test market memory features with proper year joins."""
    years = [2018, 2019, 2021, 2022, 2023, 2024, 2025]  # Skip 2017 (no prior)
    
    # Optimal features
    optimal_features = [
        'champ_equity', 'kenpom_net_pct', 'kenpom_balance',
        'points_per_equity', 'kenpom_pct_cubed'
    ]
    
    # Memory features
    memory_features = [
        'was_expensive_last_year',
        'disappointed_last_year',
        'won_championship_last_year',
        'missed_tournament_last_year',
        'last_year_investment_pct',
        'overpriced_last_year',
    ]
    
    print('TESTING MARKET MEMORY FEATURES (with proper year joins)')
    print('='*90)
    print()
    
    def add_all_features(df):
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
        
        # Market memory
        df = add_market_memory_features(df)
        
        return df
    
    def evaluate(features):
        maes = []
        for test_year in years:
            train_years = [y for y in years if y != test_year]
            
            # Load with prior year joins
            train_data = [load_year_with_prior(y) for y in train_years]
            train_data = [d for d in train_data if d is not None]
            train_df = pd.concat(train_data, ignore_index=True)
            test_df = load_year_with_prior(test_year)
            
            if test_df is None:
                continue
            
            train_df = add_all_features(train_df)
            test_df = add_all_features(test_df)
            
            share_col = 'team_share_of_pool' if 'team_share_of_pool' in test_df.columns else 'actual_share'
            
            X_train = train_df[features].fillna(0)
            X_test = test_df[features].fillna(0)
            y_train = train_df[share_col].values
            y_test = test_df[share_col].values
            
            scaler = StandardScaler()
            X_train_scaled = scaler.fit_transform(X_train)
            X_test_scaled = scaler.transform(X_test)
            
            model = Ridge(alpha=1.0)
            model.fit(X_train_scaled, y_train)
            pred = model.predict(X_test_scaled)
            
            mae = np.abs(pred - y_test).mean()
            maes.append(mae)
        
        return np.mean(maes)
    
    # Baseline
    baseline_mae = evaluate(optimal_features)
    print(f'Baseline (optimal 5 features): MAE = {baseline_mae:.6f}')
    print()
    
    # Test each memory feature
    print('Testing each market memory feature:')
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
        print(f'Best: {best[0]}')
        print(f'  Improvement: {best[2]:+.6f} ({best[2]/baseline_mae*100:.1f}%)')
        print(f'  New MAE: {best[1]:.6f}')
        print()
        print('✓ Market memory effects exist!')
    else:
        print('No market memory features improved MAE by >0.0001')
        print('Market treats each year independently.')


def analyze_specific_cases():
    """Analyze specific cases with proper data."""
    print()
    print('='*90)
    print('SPECIFIC CASES')
    print('='*90)
    print()
    
    cases = [
        (2022, 'Baylor', '2021 champion'),
        (2024, 'North Carolina', 'Missed 2023'),
        (2024, 'Arizona', 'Overpriced in 2023'),
    ]
    
    for year, team, note in cases:
        df = load_year_with_prior(year)
        if df is None:
            continue
        
        df = add_market_memory_features(df)
        
        team_row = df[df['school_name'].str.contains(team, case=False, na=False)]
        
        if len(team_row) > 0:
            row = team_row.iloc[0]
            share_col = 'team_share_of_pool' if 'team_share_of_pool' in df.columns else 'actual_share'
            
            print(f'{year} {team}: {note}')
            print('-'*90)
            print(f"  Current investment: {row[share_col]*100:.2f}%")
            print(f"  Prior year investment: {row['prior_year_share']*100:.2f}%")
            print(f"  Prior year seed: {row['prior_year_seed']}")
            print(f"  Prior year progress: {row.get('prior_year_progress', 0):.1f}")
            print(f"  Prior year wins: {row.get('prior_year_wins', 0):.0f}")
            print(f"  Won championship: {row['won_championship_last_year']}")
            print(f"  Missed tournament: {row['missed_tournament_last_year']}")
            print(f"  Was expensive: {row['was_expensive_last_year']}")
            print(f"  Disappointed: {row['disappointed_last_year']}")
            print()


if __name__ == '__main__':
    test_market_memory_features()
    analyze_specific_cases()
