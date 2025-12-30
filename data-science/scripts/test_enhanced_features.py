#!/usr/bin/env python
"""
Test enhanced feature set on historical data (2023-2025).

Compares old feature set vs new enhanced feature set.
"""
from __future__ import annotations

import sys
from pathlib import Path

import pandas as pd

sys.path.insert(0, str(Path(__file__).parent.parent))

from moneyball.models.predicted_auction_share_of_pool import (
    predict_auction_share_of_pool_from_out_root,
)


def test_year(
    year: str,
    out_root: Path,
    old_feature_set: str,
    new_feature_set: str,
) -> dict:
    """Test both feature sets on a given year."""
    print(f"\n{'='*60}")
    print(f"Testing {year}")
    print(f"{'='*60}")

    try:
        old_pred = predict_auction_share_of_pool_from_out_root(
            out_root=out_root,
            predict_snapshot=year,
            ridge_alpha=1.0,
            feature_set=old_feature_set,
        )
        print(f"✓ Old feature set ({old_feature_set}): {len(old_pred)} teams")
    except Exception as e:
        print(f"✗ Old feature set failed: {e}")
        old_pred = None

    try:
        new_pred = predict_auction_share_of_pool_from_out_root(
            out_root=out_root,
            predict_snapshot=year,
            ridge_alpha=1.0,
            feature_set=new_feature_set,
        )
        print(f"✓ New feature set ({new_feature_set}): {len(new_pred)} teams")
    except Exception as e:
        print(f"✗ New feature set failed: {e}")
        new_pred = None

    if old_pred is None or new_pred is None:
        return {"year": year, "status": "failed"}

    snapshot_dir = out_root / year
    
    team_dataset_path = snapshot_dir / "derived" / "team_dataset.parquet"
    if not team_dataset_path.exists():
        print("✗ No team_dataset available")
        return {"year": year, "status": "no_data"}
    
    teams = pd.read_parquet(team_dataset_path)
    
    if "team_share_of_pool" not in teams.columns:
        print("✗ No actual market data available")
        return {"year": year, "status": "no_data"}

    teams = teams[["team_key", "team_share_of_pool"]].copy()
    teams["team_key"] = teams["team_key"].astype(str)
    teams["team_share_of_pool"] = pd.to_numeric(
        teams["team_share_of_pool"], errors="coerce"
    ).fillna(0.0)

    old_merged = old_pred.merge(teams, on="team_key", how="inner")
    new_merged = new_pred.merge(teams, on="team_key", how="inner")

    if old_merged.empty or new_merged.empty:
        print("✗ Merge failed")
        return {"year": year, "status": "merge_failed"}

    old_mae = (
        (
            old_merged["predicted_auction_share_of_pool"]
            - old_merged["team_share_of_pool"]
        )
        .abs()
        .mean()
    )
    new_mae = (
        (
            new_merged["predicted_auction_share_of_pool"]
            - new_merged["team_share_of_pool"]
        )
        .abs()
        .mean()
    )

    old_corr = old_merged["predicted_auction_share_of_pool"].corr(
        old_merged["team_share_of_pool"]
    )
    new_corr = new_merged["predicted_auction_share_of_pool"].corr(
        new_merged["team_share_of_pool"]
    )

    print(f"\nOld MAE: {old_mae:.6f} | Correlation: {old_corr:.4f}")
    print(f"New MAE: {new_mae:.6f} | Correlation: {new_corr:.4f}")

    improvement_mae = ((old_mae - new_mae) / old_mae * 100) if old_mae > 0 else 0
    improvement_corr = new_corr - old_corr

    if improvement_mae > 0:
        print(f"✓ MAE improved by {improvement_mae:.1f}%")
    else:
        print(f"✗ MAE worsened by {-improvement_mae:.1f}%")

    if improvement_corr > 0:
        print(f"✓ Correlation improved by {improvement_corr:.4f}")
    else:
        print(f"✗ Correlation worsened by {improvement_corr:.4f}")

    return {
        "year": year,
        "status": "success",
        "old_mae": old_mae,
        "new_mae": new_mae,
        "old_corr": old_corr,
        "new_corr": new_corr,
        "mae_improvement_pct": improvement_mae,
        "corr_improvement": improvement_corr,
    }


def main():
    out_root = Path(__file__).parent.parent / "out"
    old_feature_set = "expanded_last_year_expected"
    new_feature_set = "enhanced"

    test_years = ["2023", "2024", "2025"]

    results = []
    for year in test_years:
        result = test_year(year, out_root, old_feature_set, new_feature_set)
        results.append(result)

    print(f"\n{'='*60}")
    print("SUMMARY")
    print(f"{'='*60}")

    df = pd.DataFrame(results)
    successful = df[df["status"] == "success"]

    if successful.empty:
        print("No successful tests")
        return

    print(f"\nTested {len(successful)} years successfully")
    print(f"\nAverage MAE improvement: {successful['mae_improvement_pct'].mean():.1f}%")
    print(f"Average correlation improvement: {successful['corr_improvement'].mean():.4f}")

    print("\nDetailed Results:")
    print(successful[["year", "old_mae", "new_mae", "old_corr", "new_corr"]].to_string(index=False))


if __name__ == "__main__":
    main()
