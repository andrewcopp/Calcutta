"""
Calculate percentile rankings for all entries based on normalized payouts.

This compares our strategy against all actual entries to determine relative
performance independent of pool size.
"""
from __future__ import annotations

import pandas as pd
import numpy as np
from typing import Dict, Tuple

from moneyball.models.simulated_entry_outcomes import simulate_entry_outcomes


def calculate_entry_percentiles(
    *,
    games: pd.DataFrame,
    teams: pd.DataFrame,
    payouts: pd.DataFrame,
    entry_bids: pd.DataFrame,
    predicted_game_outcomes: pd.DataFrame,
    recommended_entry_bids: pd.DataFrame,
    simulated_tournaments: pd.DataFrame,
    calcutta_key: str,
    n_sims: int,
    seed: int,
    budget_points: int,
) -> pd.DataFrame:
    """
    Calculate normalized payout percentiles for all entries.
    
    Returns DataFrame with columns:
    - entry_key: Entry identifier
    - mean_normalized_payout: Average normalized payout across simulations
    - percentile_rank: Percentile rank (0-100, higher is better)
    - is_our_strategy: Boolean indicating if this is our recommended strategy
    """
    # Calculate our strategy's performance
    our_summary, _ = simulate_entry_outcomes(
        games=games,
        teams=teams,
        payouts=payouts,
        entry_bids=entry_bids,
        predicted_game_outcomes=predicted_game_outcomes,
        recommended_entry_bids=recommended_entry_bids,
        simulated_tournaments=simulated_tournaments,
        calcutta_key=calcutta_key,
        n_sims=n_sims,
        seed=seed,
        budget_points=budget_points,
        sim_entry_key="our_strategy",
        keep_sims=False,
    )
    
    our_normalized_payout = float(
        our_summary["mean_normalized_payout"].iloc[0]
    )
    
    # Calculate performance for all actual entries
    entry_performances = []
    
    for entry_key in entry_bids["entry_key"].unique():
        entry_data = entry_bids[entry_bids["entry_key"] == entry_key].copy()
        entry_data = entry_data.rename(
            columns={"bid_amount": "bid_amount_points"}
        )
        
        try:
            entry_summary, _ = simulate_entry_outcomes(
                games=games,
                teams=teams,
                payouts=payouts,
                entry_bids=entry_bids,
                predicted_game_outcomes=predicted_game_outcomes,
                recommended_entry_bids=entry_data,
                simulated_tournaments=simulated_tournaments,
                calcutta_key=calcutta_key,
                n_sims=n_sims,
                seed=seed,
                budget_points=budget_points,
                sim_entry_key=str(entry_key),
                keep_sims=False,
            )
            
            normalized_payout = float(
                entry_summary["mean_normalized_payout"].iloc[0]
            )
            
            entry_performances.append({
                "entry_key": str(entry_key),
                "mean_normalized_payout": normalized_payout,
                "is_our_strategy": False,
            })
        except Exception as e:
            print(f"Warning: Failed to simulate {entry_key}: {e}")
            continue
    
    # Add our strategy
    entry_performances.append({
        "entry_key": "our_strategy",
        "mean_normalized_payout": our_normalized_payout,
        "is_our_strategy": True,
    })
    
    # Calculate percentiles
    df = pd.DataFrame(entry_performances)
    
    # Percentile rank: what % of entries have lower normalized payout
    df["percentile_rank"] = df["mean_normalized_payout"].rank(
        pct=True
    ) * 100
    
    # Sort by performance
    df = df.sort_values("mean_normalized_payout", ascending=False)
    
    return df


def generate_percentile_report(
    percentiles_df: pd.DataFrame,
) -> str:
    """Generate a human-readable report of percentile rankings."""
    lines = []
    lines.append("=" * 80)
    lines.append("ENTRY PERFORMANCE PERCENTILE ANALYSIS")
    lines.append("=" * 80)
    lines.append("")
    
    our_row = percentiles_df[percentiles_df["is_our_strategy"]].iloc[0]
    our_rank = (percentiles_df["mean_normalized_payout"] >
                our_row["mean_normalized_payout"]).sum() + 1
    total_entries = len(percentiles_df)
    
    lines.append(f"Our Strategy Performance:")
    lines.append(f"  Normalized Payout: {our_row['mean_normalized_payout']:.3f}")
    lines.append(f"  Percentile Rank: {our_row['percentile_rank']:.1f}%")
    lines.append(f"  Absolute Rank: {our_rank} out of {total_entries}")
    lines.append("")
    
    lines.append("Top 10 Entries:")
    lines.append("-" * 80)
    lines.append(f"{'Rank':<6} {'Entry':<40} {'Norm Payout':>12} {'Percentile':>11}")
    lines.append("-" * 80)
    
    for i, (_, row) in enumerate(percentiles_df.head(10).iterrows(), 1):
        marker = " ← US" if row["is_our_strategy"] else ""
        entry_name = row["entry_key"][:39]
        lines.append(
            f"{i:<6} {entry_name:<40} "
            f"{row['mean_normalized_payout']:>12.3f} "
            f"{row['percentile_rank']:>10.1f}%{marker}"
        )
    
    lines.append("")
    
    return "\n".join(lines)
