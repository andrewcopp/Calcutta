#!/usr/bin/env python
"""
Diagnose why model performance collapsed in 2023-2025.

Compares:
1. Prediction accuracy (predicted vs actual market shares)
2. Feature distributions
3. Market behavior changes
4. KenPom predictive power
"""
from __future__ import annotations

import sys
from pathlib import Path

import pandas as pd
import numpy as np

sys.path.insert(0, str(Path(__file__).parent.parent))

from moneyball.models.predicted_auction_share_of_pool import (
    predict_auction_share_of_pool_from_out_root,
)


def load_year_data(year: str, out_root: Path) -> dict:
    """Load all relevant data for a year."""
    snapshot_dir = out_root / year
    
    team_dataset = pd.read_parquet(
        snapshot_dir / "derived" / "team_dataset.parquet"
    )
    
    predictions = predict_auction_share_of_pool_from_out_root(
        out_root=out_root,
        predict_snapshot=year,
        ridge_alpha=1.0,
        feature_set="expanded_last_year_expected",
    )
    
    merged = team_dataset.merge(
        predictions[["team_key", "predicted_auction_share_of_pool"]],
        on="team_key",
        how="left"
    )
    
    return {
        "year": year,
        "data": merged,
        "n_teams": len(merged),
    }


def analyze_prediction_errors(years_data: list) -> pd.DataFrame:
    """Analyze prediction errors by year."""
    results = []
    
    for yd in years_data:
        df = yd["data"]
        
        if "team_share_of_pool" not in df.columns:
            continue
            
        df = df[df["team_share_of_pool"].notna()].copy()
        df = df[df["predicted_auction_share_of_pool"].notna()].copy()
        
        if df.empty:
            continue
        
        errors = df["predicted_auction_share_of_pool"] - df["team_share_of_pool"]
        abs_errors = errors.abs()
        
        mae = abs_errors.mean()
        rmse = np.sqrt((errors ** 2).mean())
        corr = df["predicted_auction_share_of_pool"].corr(
            df["team_share_of_pool"]
        )
        
        max_overpredict = errors.max()
        max_underpredict = errors.min()
        
        overpredict_teams = df[errors > 0.02].copy()
        underpredict_teams = df[errors < -0.02].copy()
        
        results.append({
            "year": yd["year"],
            "mae": mae,
            "rmse": rmse,
            "correlation": corr,
            "max_overpredict": max_overpredict,
            "max_underpredict": max_underpredict,
            "n_overpredicted": len(overpredict_teams),
            "n_underpredicted": len(underpredict_teams),
        })
    
    return pd.DataFrame(results)


def analyze_feature_distributions(years_data: list) -> pd.DataFrame:
    """Compare feature distributions across years."""
    results = []
    
    features = [
        "kenpom_net",
        "seed",
        "total_bid_amount",
        "team_share_of_pool",
    ]
    
    for yd in years_data:
        df = yd["data"]
        row = {"year": yd["year"]}
        
        for feat in features:
            if feat in df.columns:
                vals = pd.to_numeric(df[feat], errors="coerce").dropna()
                row[f"{feat}_mean"] = vals.mean()
                row[f"{feat}_std"] = vals.std()
                row[f"{feat}_min"] = vals.min()
                row[f"{feat}_max"] = vals.max()
        
        results.append(row)
    
    return pd.DataFrame(results)


def analyze_market_concentration(years_data: list) -> pd.DataFrame:
    """Analyze market concentration (Gini coefficient, top-N share)."""
    results = []
    
    for yd in years_data:
        df = yd["data"]
        
        if "team_share_of_pool" not in df.columns:
            continue
        
        shares = pd.to_numeric(
            df["team_share_of_pool"], errors="coerce"
        ).dropna().sort_values(ascending=False)
        
        if shares.empty:
            continue
        
        top1_share = shares.iloc[0] if len(shares) > 0 else 0
        top3_share = shares.iloc[:3].sum() if len(shares) >= 3 else 0
        top5_share = shares.iloc[:5].sum() if len(shares) >= 5 else 0
        
        sorted_shares = np.sort(shares.values)
        n = len(sorted_shares)
        cumsum = np.cumsum(sorted_shares)
        gini = (2 * np.sum((np.arange(1, n + 1)) * sorted_shares)) / (
            n * cumsum[-1]
        ) - (n + 1) / n
        
        results.append({
            "year": yd["year"],
            "top1_share": top1_share,
            "top3_share": top3_share,
            "top5_share": top5_share,
            "gini": gini,
        })
    
    return pd.DataFrame(results)


def analyze_seed_bias(years_data: list) -> pd.DataFrame:
    """Analyze prediction bias by seed."""
    results = []
    
    for yd in years_data:
        df = yd["data"]
        
        if "team_share_of_pool" not in df.columns or "seed" not in df.columns:
            continue
        
        df = df[df["team_share_of_pool"].notna()].copy()
        df = df[df["predicted_auction_share_of_pool"].notna()].copy()
        df["seed"] = pd.to_numeric(df["seed"], errors="coerce")
        df = df[df["seed"].notna()].copy()
        
        if df.empty:
            continue
        
        df["error"] = (
            df["predicted_auction_share_of_pool"] - df["team_share_of_pool"]
        )
        
        for seed_group in [(1, 4), (5, 8), (9, 12), (13, 16)]:
            mask = (df["seed"] >= seed_group[0]) & (df["seed"] <= seed_group[1])
            subset = df[mask]
            
            if subset.empty:
                continue
            
            results.append({
                "year": yd["year"],
                "seed_group": f"{seed_group[0]}-{seed_group[1]}",
                "mean_error": subset["error"].mean(),
                "mae": subset["error"].abs().mean(),
                "n_teams": len(subset),
            })
    
    return pd.DataFrame(results)


def main():
    out_root = Path(__file__).parent.parent / "out"
    
    years = ["2017", "2018", "2019", "2021", "2022", "2023", "2024", "2025"]
    
    print("Loading data for all years...")
    years_data = []
    for year in years:
        try:
            yd = load_year_data(year, out_root)
            years_data.append(yd)
            print(f"  {year}: {yd['n_teams']} teams")
        except Exception as e:
            print(f"  {year}: ERROR - {e}")
    
    print("\n" + "="*70)
    print("1. PREDICTION ERRORS BY YEAR")
    print("="*70)
    
    errors_df = analyze_prediction_errors(years_data)
    print(errors_df.to_string(index=False))
    
    print("\n" + "="*70)
    print("2. FEATURE DISTRIBUTIONS")
    print("="*70)
    
    features_df = analyze_feature_distributions(years_data)
    key_cols = [
        "year",
        "kenpom_net_mean",
        "seed_mean",
        "total_bid_amount_mean",
        "team_share_of_pool_mean",
    ]
    available_cols = [c for c in key_cols if c in features_df.columns]
    print(features_df[available_cols].to_string(index=False))
    
    print("\n" + "="*70)
    print("3. MARKET CONCENTRATION")
    print("="*70)
    
    concentration_df = analyze_market_concentration(years_data)
    print(concentration_df.to_string(index=False))
    
    print("\n" + "="*70)
    print("4. PREDICTION BIAS BY SEED GROUP")
    print("="*70)
    
    seed_bias_df = analyze_seed_bias(years_data)
    pivot = seed_bias_df.pivot(
        index="seed_group", columns="year", values="mean_error"
    )
    print(pivot.to_string())
    
    print("\n" + "="*70)
    print("KEY FINDINGS")
    print("="*70)
    
    if not errors_df.empty:
        pre_2023 = errors_df[errors_df["year"].isin(["2017", "2018", "2019", "2021", "2022"])]
        post_2022 = errors_df[errors_df["year"].isin(["2023", "2024", "2025"])]
        
        if not pre_2023.empty and not post_2022.empty:
            print(f"\nPre-2023 average MAE: {pre_2023['mae'].mean():.6f}")
            print(f"Post-2022 average MAE: {post_2022['mae'].mean():.6f}")
            print(f"Change: {(post_2022['mae'].mean() / pre_2023['mae'].mean() - 1) * 100:+.1f}%")
            
            print(f"\nPre-2023 average correlation: {pre_2023['correlation'].mean():.4f}")
            print(f"Post-2022 average correlation: {post_2022['correlation'].mean():.4f}")
            print(f"Change: {post_2022['correlation'].mean() - pre_2023['correlation'].mean():+.4f}")
    
    if not concentration_df.empty:
        pre_2023 = concentration_df[
            concentration_df["year"].isin(["2017", "2018", "2019", "2021", "2022"])
        ]
        post_2022 = concentration_df[
            concentration_df["year"].isin(["2023", "2024", "2025"])
        ]
        
        if not pre_2023.empty and not post_2022.empty:
            print(f"\nPre-2023 average Gini: {pre_2023['gini'].mean():.4f}")
            print(f"Post-2022 average Gini: {post_2022['gini'].mean():.4f}")
            print(f"Change: {post_2022['gini'].mean() - pre_2023['gini'].mean():+.4f}")


if __name__ == "__main__":
    main()
