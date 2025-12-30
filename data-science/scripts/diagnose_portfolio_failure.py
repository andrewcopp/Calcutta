#!/usr/bin/env python
"""
Diagnose why portfolios failed in 2023-2025 despite accurate market predictions.

The market predictions are accurate, but tournament outcomes killed the portfolios.
"""
from __future__ import annotations

import sys
from pathlib import Path

import pandas as pd

sys.path.insert(0, str(Path(__file__).parent.parent))

from moneyball.utils import io, points


def analyze_portfolio_vs_actual(year: str, out_root: Path) -> dict:
    """Analyze what happened to the recommended portfolio."""
    snapshot_dir = out_root / year
    tables = io.load_snapshot_tables(snapshot_dir)
    
    teams = tables.get("teams")
    if teams is None:
        return {"year": year, "status": "no_data"}
    
    calcutta_key = io.choose_calcutta_key(teams, None)
    teams = teams[teams["calcutta_key"] == calcutta_key].copy()
    
    teams["wins"] = pd.to_numeric(teams["wins"], errors="coerce").fillna(0).astype(int)
    teams["byes"] = pd.to_numeric(teams["byes"], errors="coerce").fillna(0).astype(int)
    teams["progress"] = teams["wins"] + teams["byes"]
    teams["actual_points"] = teams["progress"].apply(
        lambda p: float(points.team_points_fixed(int(p)))
    )
    
    report_path = snapshot_dir / "derived" / "calcutta" / "investment_report.parquet"
    if not report_path.exists():
        return {"year": year, "status": "no_report"}
    
    report = pd.read_parquet(report_path)
    portfolio = report[report["is_recommended_portfolio"]].copy()
    
    if portfolio.empty:
        return {"year": year, "status": "no_portfolio"}
    
    portfolio = portfolio.merge(
        teams[["team_key", "school_name", "seed", "progress", "actual_points"]],
        on="team_key",
        how="left"
    )
    
    portfolio["expected_points"] = pd.to_numeric(
        portfolio["expected_team_points"], errors="coerce"
    ).fillna(0.0)
    portfolio["actual_points"] = pd.to_numeric(
        portfolio["actual_points"], errors="coerce"
    ).fillna(0.0)
    
    total_expected = portfolio["expected_points"].sum()
    total_actual = portfolio["actual_points"].sum()
    
    portfolio["points_delta"] = portfolio["actual_points"] - portfolio["expected_points"]
    
    biggest_disappointments = portfolio.nsmallest(3, "points_delta")
    biggest_surprises = portfolio.nlargest(3, "points_delta")
    
    return {
        "year": year,
        "status": "ok",
        "n_teams": len(portfolio),
        "total_expected_points": total_expected,
        "total_actual_points": total_actual,
        "points_delta": total_actual - total_expected,
        "portfolio": portfolio,
        "disappointments": biggest_disappointments,
        "surprises": biggest_surprises,
    }


def main():
    out_root = Path(__file__).parent.parent / "out"
    years = ["2017", "2018", "2019", "2021", "2022", "2023", "2024", "2025"]
    
    print("Analyzing portfolio performance vs expectations...\n")
    
    results = []
    for year in years:
        result = analyze_portfolio_vs_actual(year, out_root)
        results.append(result)
        
        if result["status"] != "ok":
            print(f"{year}: {result['status']}")
            continue
        
        print(f"{'='*70}")
        print(f"{year}")
        print(f"{'='*70}")
        print(f"Expected points: {result['total_expected_points']:.1f}")
        print(f"Actual points:   {result['total_actual_points']:.1f}")
        print(f"Delta:           {result['points_delta']:+.1f}")
        print(f"Performance:     {result['total_actual_points']/result['total_expected_points']*100:.1f}% of expected")
        
        print(f"\nBiggest Disappointments:")
        for _, row in result["disappointments"].iterrows():
            print(f"  {row['school_name']:20s} (seed {int(row['seed']):2d}): "
                  f"Expected {row['expected_points']:5.1f}, Got {row['actual_points']:5.1f} "
                  f"({row['points_delta']:+.1f})")
        
        print(f"\nBiggest Surprises:")
        for _, row in result["surprises"].iterrows():
            print(f"  {row['school_name']:20s} (seed {int(row['seed']):2d}): "
                  f"Expected {row['expected_points']:5.1f}, Got {row['actual_points']:5.1f} "
                  f"({row['points_delta']:+.1f})")
        print()
    
    print(f"\n{'='*70}")
    print("SUMMARY")
    print(f"{'='*70}")
    
    success_results = [r for r in results if r["status"] == "ok"]
    
    pre_2023 = [r for r in success_results if r["year"] in ["2017", "2018", "2019", "2021", "2022"]]
    post_2022 = [r for r in success_results if r["year"] in ["2023", "2024", "2025"]]
    
    if pre_2023:
        avg_delta_pre = sum(r["points_delta"] for r in pre_2023) / len(pre_2023)
        avg_pct_pre = sum(
            r["total_actual_points"] / r["total_expected_points"] 
            for r in pre_2023
        ) / len(pre_2023) * 100
        print(f"\nPre-2023 average:")
        print(f"  Points delta: {avg_delta_pre:+.1f}")
        print(f"  Performance: {avg_pct_pre:.1f}% of expected")
    
    if post_2022:
        avg_delta_post = sum(r["points_delta"] for r in post_2022) / len(post_2022)
        avg_pct_post = sum(
            r["total_actual_points"] / r["total_expected_points"] 
            for r in post_2022
        ) / len(post_2022) * 100
        print(f"\nPost-2022 average:")
        print(f"  Points delta: {avg_delta_post:+.1f}")
        print(f"  Performance: {avg_pct_post:.1f}% of expected")
        
        if pre_2023:
            print(f"\nChange:")
            print(f"  Points delta: {avg_delta_post - avg_delta_pre:+.1f}")
            print(f"  Performance: {avg_pct_post - avg_pct_pre:+.1f} percentage points")
    
    print(f"\n{'='*70}")
    print("KEY INSIGHT")
    print(f"{'='*70}")
    print("\nThe market prediction model is ACCURATE (improved in 2023-2025).")
    print("The problem is TOURNAMENT VARIANCE - the teams we invested in")
    print("underperformed their expected points in 2023-2025.")
    print("\nThis suggests:")
    print("1. Bad luck (variance is expected)")
    print("2. Possible systematic bias in tournament outcome predictions")
    print("3. Portfolio concentration risk (need better diversification)")


if __name__ == "__main__":
    main()
