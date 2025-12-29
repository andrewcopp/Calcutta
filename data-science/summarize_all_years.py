#!/usr/bin/env python
"""
Summarize investment report results across all years.
"""
from __future__ import annotations
import pandas as pd
from pathlib import Path
from typing import Optional


def find_latest_report(year: int) -> Optional[Path]:
    """Find the most recent investment report for a given year."""
    base = Path(f"out/{year}/derived/moneyball/{year}")
    if not base.exists():
        return None
    
    # Find all run directories
    runs = sorted([d for d in base.iterdir() if d.is_dir()], reverse=True)
    if not runs:
        return None
    
    # Get the most recent run
    latest = runs[0]
    report_path = latest / "investment_report.parquet"
    
    return report_path if report_path.exists() else None


def main():
    years = list(range(2017, 2026))
    results = []
    
    for year in years:
        report_path = find_latest_report(year)
        if report_path is None:
            print(f"⚠️  No report found for {year}")
            continue
        
        df = pd.read_parquet(report_path)
        if df.empty:
            print(f"⚠️  Empty report for {year}")
            continue
        
        row = df.iloc[0]
        results.append({
            "year": year,
            "budget_points": int(row["budget_points"]),
            "portfolio_teams": int(row["portfolio_team_count"]),
            "n_entries": float(row["mean_n_entries"]),
            "mean_payout_cents": float(row["mean_expected_payout_cents"]),
            "mean_position": float(row["mean_expected_finish_position"]),
            "p_top1": float(row["p_top1"]),
            "p_top3": float(row["p_top3"]),
            "concentration_hhi": float(row["portfolio_concentration_hhi"]),
        })
    
    if not results:
        print("No results found!")
        return
    
    # Create summary DataFrame
    summary = pd.DataFrame(results)
    
    print("\n" + "="*80)
    print("INVESTMENT REPORT SUMMARY (2017-2025)")
    print("="*80)
    print()
    print("Key Metrics:")
    print(f"  • Budget: {summary['budget_points'].iloc[0]} points per year")
    print(f"  • Years analyzed: {len(summary)}")
    print()
    
    print("Performance by Year (Payout-Based Metrics):")
    print("-" * 80)
    for _, row in summary.iterrows():
        print(f"\n{int(row['year'])} ({int(row['n_entries'])} entries):")
        print(f"  Portfolio: {int(row['portfolio_teams'])} teams")
        print(f"  Expected payout: ${row['mean_payout_cents']/100:.2f}")
        print(f"  Expected position: {row['mean_position']:.1f}")
        print(f"  P(1st place): {row['p_top1']:.1%}")
        print(f"  P(top 3): {row['p_top3']:.1%}")
        print(f"  Concentration (HHI): {row['concentration_hhi']:.3f}")
    
    print("\n" + "="*80)
    print("AGGREGATE STATISTICS")
    print("="*80)
    print(f"\nMean expected payout: ${summary['mean_payout_cents'].mean()/100:.2f}")
    print(f"Mean expected position: {summary['mean_position'].mean():.1f}")
    print(f"Mean P(1st): {summary['p_top1'].mean():.1%}")
    print(f"Mean P(top 3): {summary['p_top3'].mean():.1%}")
    print(f"Mean concentration: {summary['concentration_hhi'].mean():.3f}")
    print()
    
    # Save summary
    summary_path = Path("out/investment_report_summary.csv")
    summary.to_csv(summary_path, index=False)
    print(f"Summary saved to: {summary_path}")
    print()


if __name__ == "__main__":
    main()
