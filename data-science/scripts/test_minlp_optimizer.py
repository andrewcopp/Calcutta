"""
Test script for MINLP portfolio optimizer.

Compares MINLP vs greedy on 2024 data to verify it fixes the overbidding bug.
"""
from pathlib import Path
import pandas as pd
from moneyball.models.portfolio_optimizer_minlp import optimize_portfolio_minlp
from moneyball.models.recommended_entry_bids import _optimize_portfolio_greedy

# Load 2024 data
year = "2024"
year_dir = Path(f"out/{year}")
run_dir = Path(f"out/{year}/derived/calcutta")
latest_run = sorted(run_dir.glob("*"))[-1]

# Load necessary data
predicted_games = pd.read_parquet(year_dir / "derived/predicted_game_outcomes.parquet")
predicted_market = pd.read_parquet(latest_run / "predicted_auction_share_of_pool.parquet")
entries = pd.read_parquet(year_dir / "entries.parquet")

# Calculate expected points
from moneyball.models.recommended_entry_bids import (
    expected_team_points_from_predicted_game_outcomes,
)

exp_pts = expected_team_points_from_predicted_game_outcomes(
    predicted_game_outcomes=predicted_games
)

# Merge data
df = predicted_market.merge(exp_pts, on="team_key", how="left")
df["expected_team_points"] = df["expected_team_points"].fillna(0.0)

# Calculate predicted total bids
predicted_total_pool = len(entries) * 100
df["predicted_team_total_bids"] = (
    df["predicted_auction_share_of_pool"] * predicted_total_pool
)

# Calculate initial score (for greedy)
df["score"] = df.apply(
    lambda r: (
        r["expected_team_points"] / (r["predicted_team_total_bids"] + 1)
        if (r["predicted_team_total_bids"] + 1) > 0 else 0.0
    ),
    axis=1
)

# Parameters
budget = 100
min_teams = 3
max_teams = 10
max_per_team = 50
min_bid = 1

print("="*90)
print("COMPARING GREEDY VS MINLP ON 2024 DATA")
print("="*90)
print()

# Run greedy
print("Running greedy optimizer...")
greedy_result, _ = _optimize_portfolio_greedy(
    df=df,
    score_col="score",
    budget=float(budget),
    min_teams=min_teams,
    max_teams=max_teams,
    max_per_team=float(max_per_team),
    min_bid=float(min_bid),
)

# Calculate greedy metrics
greedy_bids = greedy_result["bid_amount_points"].values
greedy_exp_pts = greedy_result["expected_team_points"].values
greedy_pred_markets = greedy_result["predicted_team_total_bids"].values
greedy_ownership = greedy_bids / (greedy_pred_markets + greedy_bids)
greedy_return = (greedy_exp_pts * greedy_ownership).sum()
greedy_our_roi = greedy_exp_pts / (greedy_pred_markets + greedy_bids)

print(f"Greedy total expected return: {greedy_return:.2f} points")
print(f"Greedy teams: {len(greedy_result)}")
print(f"Greedy total bid: {greedy_bids.sum()}")
print()

# Show greedy portfolio
print("Greedy Portfolio:")
print(f"{'Team':<30s} {'Bid':>5s} {'Exp Pts':>8s} {'Pred Mkt':>9s} {'Our ROI':>8s}")
print("-"*70)
for _, row in greedy_result.sort_values("bid_amount_points", ascending=False).iterrows():
    team = row["team_key"].split(":")[-1].replace("-", " ").title()[:29]
    bid = row["bid_amount_points"]
    exp = row["expected_team_points"]
    pred = row["predicted_team_total_bids"]
    our_roi = exp / (pred + bid)
    marker = " ⚠️" if our_roi < 1.0 else ""
    print(f"{team:<30s} {bid:>5d} {exp:>8.1f} {pred:>9.1f} {our_roi:>8.2f}{marker}")
print()
print(f"Teams with our_roi < 1.0: {(greedy_our_roi < 1.0).sum()}")
print()

# Run MINLP
print("Running MINLP optimizer...")
minlp_result, _ = optimize_portfolio_minlp(
    teams_df=df,
    budget_points=budget,
    min_teams=min_teams,
    max_teams=max_teams,
    max_per_team_points=max_per_team,
    min_bid_points=min_bid,
    initial_solution="greedy",
)

# Calculate MINLP metrics
minlp_bids = minlp_result["bid_amount_points"].values
minlp_exp_pts = minlp_result["expected_team_points"].values
minlp_pred_markets = minlp_result["predicted_team_total_bids"].values
minlp_ownership = minlp_bids / (minlp_pred_markets + minlp_bids)
minlp_return = (minlp_exp_pts * minlp_ownership).sum()
minlp_our_roi = minlp_exp_pts / (minlp_pred_markets + minlp_bids)

print(f"MINLP total expected return: {minlp_return:.2f} points")
print(f"MINLP teams: {len(minlp_result)}")
print(f"MINLP total bid: {minlp_bids.sum()}")
print()

# Show MINLP portfolio
print("MINLP Portfolio:")
print(f"{'Team':<30s} {'Bid':>5s} {'Exp Pts':>8s} {'Pred Mkt':>9s} {'Our ROI':>8s}")
print("-"*70)
for _, row in minlp_result.sort_values("bid_amount_points", ascending=False).iterrows():
    team = row["team_key"].split(":")[-1].replace("-", " ").title()[:29]
    bid = row["bid_amount_points"]
    exp = row["expected_team_points"]
    pred = row["predicted_team_total_bids"]
    our_roi = exp / (pred + bid)
    marker = " ⚠️" if our_roi < 1.0 else ""
    print(f"{team:<30s} {bid:>5d} {exp:>8.1f} {pred:>9.1f} {our_roi:>8.2f}{marker}")
print()
print(f"Teams with our_roi < 1.0: {(minlp_our_roi < 1.0).sum()}")
print()

# Compare
print("="*90)
print("COMPARISON")
print("="*90)
print(f"Expected return improvement: {minlp_return - greedy_return:+.2f} points ({(minlp_return/greedy_return - 1)*100:+.1f}%)")
print(f"Teams with our_roi < 1.0: Greedy={int((greedy_our_roi < 1.0).sum())}, MINLP={int((minlp_our_roi < 1.0).sum())}")
print()

if minlp_return > greedy_return:
    print("✅ MINLP found a better solution!")
elif abs(minlp_return - greedy_return) < 0.01:
    print("⚠️  MINLP found same solution as greedy")
else:
    print("❌ MINLP found worse solution than greedy (this shouldn't happen)")
