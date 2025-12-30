#!/usr/bin/env python
"""
Tune ridge regularization parameter for enhanced features.
"""
from __future__ import annotations

import sys
from pathlib import Path

import pandas as pd

sys.path.insert(0, str(Path(__file__).parent.parent))

from moneyball.models.predicted_auction_share_of_pool import (
    predict_auction_share_of_pool_from_out_root,
)


def test_alpha(year: str, out_root: Path, alpha: float) -> dict:
    """Test a specific alpha value."""
    try:
        pred = predict_auction_share_of_pool_from_out_root(
            out_root=out_root,
            predict_snapshot=year,
            ridge_alpha=alpha,
            feature_set="enhanced",
        )
        
        team_dataset_path = out_root / year / "derived" / "team_dataset.parquet"
        teams = pd.read_parquet(team_dataset_path)
        teams = teams[["team_key", "team_share_of_pool"]].copy()
        teams["team_key"] = teams["team_key"].astype(str)
        
        merged = pred.merge(teams, on="team_key", how="inner")
        
        mae = (
            (merged["predicted_auction_share_of_pool"] - merged["team_share_of_pool"])
            .abs()
            .mean()
        )
        corr = merged["predicted_auction_share_of_pool"].corr(
            merged["team_share_of_pool"]
        )
        
        return {"year": year, "alpha": alpha, "mae": mae, "corr": corr, "status": "ok"}
    except Exception as e:
        return {"year": year, "alpha": alpha, "status": "error", "error": str(e)}


def main():
    out_root = Path(__file__).parent.parent / "out"
    test_years = ["2023", "2024", "2025"]
    
    alphas = [0.1, 0.5, 1.0, 5.0, 10.0, 50.0, 100.0, 500.0, 1000.0]
    
    print("Testing ridge alpha values for enhanced features...")
    print(f"Years: {test_years}")
    print(f"Alphas: {alphas}\n")
    
    results = []
    for alpha in alphas:
        print(f"\nTesting alpha={alpha}")
        for year in test_years:
            result = test_alpha(year, out_root, alpha)
            results.append(result)
            if result["status"] == "ok":
                print(f"  {year}: MAE={result['mae']:.6f}, Corr={result['corr']:.4f}")
            else:
                print(f"  {year}: ERROR - {result.get('error', 'unknown')}")
    
    df = pd.DataFrame([r for r in results if r["status"] == "ok"])
    
    if df.empty:
        print("\nNo successful results")
        return
    
    summary = df.groupby("alpha").agg({
        "mae": "mean",
        "corr": "mean"
    }).reset_index()
    
    print("\n" + "="*60)
    print("SUMMARY BY ALPHA")
    print("="*60)
    print(summary.to_string(index=False))
    
    best_mae_alpha = summary.loc[summary["mae"].idxmin(), "alpha"]
    best_corr_alpha = summary.loc[summary["corr"].idxmax(), "alpha"]
    
    print(f"\nBest alpha for MAE: {best_mae_alpha}")
    print(f"Best alpha for correlation: {best_corr_alpha}")
    
    baseline_mae = 0.006135
    baseline_corr = 0.8746
    
    print(f"\nBaseline (expanded_last_year_expected, alpha=1.0):")
    print(f"  MAE: {baseline_mae:.6f}")
    print(f"  Correlation: {baseline_corr:.4f}")
    
    best_row = summary.loc[summary["mae"].idxmin()]
    print(f"\nBest enhanced (alpha={best_row['alpha']}):")
    print(f"  MAE: {best_row['mae']:.6f} ({(best_row['mae']/baseline_mae - 1)*100:+.1f}%)")
    print(f"  Correlation: {best_row['corr']:.4f} ({best_row['corr'] - baseline_corr:+.4f})")


if __name__ == "__main__":
    main()
