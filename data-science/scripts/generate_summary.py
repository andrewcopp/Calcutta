#!/usr/bin/env python
"""Generate investment_report_summary.csv from all investment reports."""
from __future__ import annotations

import glob
from pathlib import Path

import pandas as pd

out_root = Path(__file__).parent.parent / "out"
years = ["2017", "2018", "2019", "2021", "2022", "2023", "2024", "2025"]

results = []

for year in years:
    # Find the most recent investment report for this year
    pattern = str(out_root / year / "derived" / "calcutta" / "*" / "investment_report.parquet")
    reports = sorted(glob.glob(pattern))
    
    if not reports:
        print(f"Warning: No investment report found for {year}")
        continue
    
    # Use the most recent one
    report_path = reports[-1]
    report = pd.read_parquet(report_path)
    
    results.append({
        "year": int(year),
        "budget_points": int(report["budget_points"].iloc[0]),
        "portfolio_teams": int(report["portfolio_team_count"].iloc[0]),
        "n_entries": float(report["mean_n_entries"].iloc[0]),
        "mean_payout_cents": float(report["mean_expected_payout_cents"].iloc[0]),
        "mean_position": float(report["mean_expected_finish_position"].iloc[0]),
        "p_top1": float(report["p_top1"].iloc[0]),
        "p_in_money": float(report["p_in_money"].iloc[0]),
        "concentration_hhi": float(report["portfolio_concentration_hhi"].iloc[0]),
    })

summary = pd.DataFrame(results)
summary = summary.sort_values("year")

output_path = out_root / "investment_report_summary.csv"
summary.to_csv(output_path, index=False)

print(f"Wrote summary to {output_path}")
print("\nSummary:")
print(summary.to_string(index=False))
