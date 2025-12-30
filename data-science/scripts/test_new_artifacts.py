"""
Test script to validate new artifact structure.

Generates production and debug artifacts for 2025 data.
"""
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent.parent))

import pandas as pd

from moneyball.models.tournament_value import (
    generate_tournament_value,
    generate_tournament_value_from_simulations,
)
from moneyball.models.market_prediction import generate_market_prediction
from moneyball.models.portfolio_construction import generate_recommended_bids
from moneyball.models.predicted_auction_share_of_pool import (
    predict_auction_share_of_pool_from_out_root,
)
from moneyball.models.recommended_entry_bids import recommend_entry_bids


def test_tournament_value():
    """Test tournament value artifact generation."""
    print("=" * 90)
    print("TESTING TOURNAMENT VALUE ARTIFACTS")
    print("=" * 90)
    print()

    # Load predicted game outcomes
    pred_games = pd.read_parquet(
        "out/2025/derived/predicted_game_outcomes.parquet"
    )

    # Generate artifacts
    prod, debug = generate_tournament_value(
        predicted_game_outcomes=pred_games
    )

    print(f"Production artifact shape: {prod.shape}")
    print(f"Production columns: {prod.columns.tolist()}")
    print()

    print(f"Debug artifact shape: {debug.shape}")
    print(f"Debug columns: {debug.columns.tolist()}")
    print()

    # Validate production artifact
    assert "team_key" in prod.columns
    assert "expected_points_per_entry" in prod.columns
    assert len(prod.columns) == 2, "Production should have exactly 2 columns"

    # Validate debug artifact
    assert "team_key" in debug.columns
    assert "expected_points_per_entry" in debug.columns
    assert "variance_points" in debug.columns
    assert "std_points" in debug.columns
    for r in range(1, 8):
        assert f"p_round_{r}" in debug.columns

    print("✓ Tournament value artifacts validated")
    print()

    # Show top 5 teams
    print("Top 5 teams by expected points:")
    print("-" * 90)
    top5 = debug.nlargest(5, "expected_points_per_entry")
    for _, row in top5.iterrows():
        print(
            f"  {row['team_key']}: "
            f"{row['expected_points_per_entry']:.2f} points, "
            f"P(champ)={row['p_round_7']*100:.1f}%"
        )
    print()

    # Check championship probabilities
    total_champ_prob = debug["p_round_7"].sum()
    print(f"Total championship probability: {total_champ_prob*100:.2f}%")
    if abs(total_champ_prob - 1.0) > 0.01:
        print(
            f"⚠️  WARNING: Championship probabilities sum to "
            f"{total_champ_prob*100:.2f}%, not 100%"
        )
    print()

    return prod, debug


def test_tournament_value_from_sims():
    """Test tournament value from simulations."""
    print("=" * 90)
    print("TESTING TOURNAMENT VALUE FROM SIMULATIONS")
    print("=" * 90)
    print()

    # Load simulated tournaments
    sim_tournaments = pd.read_parquet("out/2025/derived/tournaments.parquet")

    # Generate artifacts
    prod, debug = generate_tournament_value_from_simulations(
        simulated_tournaments=sim_tournaments
    )

    print(f"Production artifact shape: {prod.shape}")
    print(f"Debug artifact shape: {debug.shape}")
    print()

    # Show top 5 teams
    print("Top 5 teams by expected points (from sims):")
    print("-" * 90)
    top5 = debug.nlargest(5, "expected_points_per_entry")
    for _, row in top5.iterrows():
        print(
            f"  {row['team_key']}: "
            f"{row['expected_points_per_entry']:.2f} points, "
            f"P(champ)={row['p_round_7']*100:.1f}%, "
            f"max_wins={int(row['max_wins'])}"
        )
    print()

    # Check championship probabilities
    total_champ_prob = debug["p_round_7"].sum()
    print(f"Total championship probability: {total_champ_prob*100:.2f}%")
    if abs(total_champ_prob - 1.0) > 0.01:
        print(
            f"⚠️  BUG IDENTIFIED: Championship probabilities sum to "
            f"{total_champ_prob*100:.2f}%, not 100%"
        )
        print("   This confirms the tournament simulation is broken.")
    print()

    return prod, debug


def test_market_prediction():
    """Test market prediction artifact generation."""
    print("=" * 90)
    print("TESTING MARKET PREDICTION ARTIFACTS")
    print("=" * 90)
    print()

    # Generate market predictions
    out_root = Path("out")
    pred_share = predict_auction_share_of_pool_from_out_root(
        out_root=out_root,
        predict_snapshot="2025",
        train_snapshots=["2017", "2018", "2019", "2021", "2022", "2023", "2024"],
        ridge_alpha=1.0,
        feature_set="optimal",
    )

    # Generate artifacts
    prod, debug = generate_market_prediction(
        predicted_auction_share_of_pool=pred_share
    )

    print(f"Production artifact shape: {prod.shape}")
    print(f"Production columns: {prod.columns.tolist()}")
    print()

    print(f"Debug artifact shape: {debug.shape}")
    print(f"Debug columns: {debug.columns.tolist()[:5]}... (truncated)")
    print()

    # Validate
    assert "team_key" in prod.columns
    assert "predicted_market_share" in prod.columns
    assert len(prod.columns) == 2

    # Check sum
    total_share = prod["predicted_market_share"].sum()
    print(f"Total predicted market share: {total_share:.6f}")
    assert abs(total_share - 1.0) < 0.0001, "Market shares should sum to 1.0"

    print("✓ Market prediction artifacts validated")
    print()

    return prod, debug


def test_portfolio_construction():
    """Test portfolio construction artifact generation."""
    print("=" * 90)
    print("TESTING PORTFOLIO CONSTRUCTION ARTIFACTS")
    print("=" * 90)
    print()

    # Load inputs
    pred_games = pd.read_parquet(
        "out/2025/derived/predicted_game_outcomes.parquet"
    )
    out_root = Path("out")
    pred_share = predict_auction_share_of_pool_from_out_root(
        out_root=out_root,
        predict_snapshot="2025",
        train_snapshots=["2017", "2018", "2019", "2021", "2022", "2023", "2024"],
        ridge_alpha=1.0,
        feature_set="optimal",
    )

    # Generate tournament value
    tv_prod, tv_debug = generate_tournament_value(
        predicted_game_outcomes=pred_games
    )

    # Generate market prediction
    mp_prod, mp_debug = generate_market_prediction(
        predicted_auction_share_of_pool=pred_share
    )

    # Generate recommended bids
    entries = pd.read_parquet("out/2025/entries.parquet")
    n_entries = int(entries["entry_key"].nunique())
    total_pool = n_entries * 100

    rec_bids = recommend_entry_bids(
        predicted_auction_share_of_pool=pred_share,
        predicted_game_outcomes=pred_games,
        predicted_total_pool_bids_points=total_pool,
        budget_points=100,
        min_teams=3,
        max_teams=10,
        max_per_team_points=50,
        min_bid_points=1,
        strategy="greedy",
    )

    # Generate artifacts
    prod, debug = generate_recommended_bids(
        recommended_entry_bids=rec_bids,
        tournament_value=tv_prod,
        market_prediction=mp_prod,
        predicted_total_pool_bids_points=total_pool,
    )

    print(f"Production artifact shape: {prod.shape}")
    print(f"Production columns: {prod.columns.tolist()}")
    print()

    print(f"Debug artifact shape: {debug.shape}")
    print(f"Debug columns: {debug.columns.tolist()}")
    print()

    # Validate
    assert "team_key" in prod.columns
    assert "bid_amount_points" in prod.columns
    assert len(prod.columns) == 2

    # Check budget
    total_bid = prod["bid_amount_points"].sum()
    print(f"Total bid amount: ${total_bid}")
    assert total_bid == 100, "Bids should sum to budget"

    print("✓ Portfolio construction artifacts validated")
    print()

    # Show portfolio with ROI
    print("Portfolio (sorted by bid amount):")
    print("-" * 90)
    print(
        f"{'Team':<40s} {'Bid':>6s} {'Expected':>10s} {'ROI':>8s} "
        f"{'Own%':>8s}"
    )
    print("-" * 90)

    portfolio = debug.sort_values("bid_amount_points", ascending=False)
    for _, row in portfolio.iterrows():
        team = row["team_key"][:39]
        bid = row["bid_amount_points"]
        exp_ret = row["expected_return_points"]
        roi = row["roi"]
        own = row["ownership_fraction"] * 100
        print(
            f"{team:<40s} ${bid:>5.0f} {exp_ret:>10.2f} {roi:>7.2f}x "
            f"{own:>7.1f}%"
        )

    print()
    print(f"Portfolio expected return: {debug['expected_return_points'].sum():.2f}")
    print(
        f"Portfolio ROI: "
        f"{debug['expected_return_points'].sum() / total_bid:.2f}x"
    )
    print()

    return prod, debug


if __name__ == "__main__":
    print("TESTING NEW ARTIFACT STRUCTURE")
    print("=" * 90)
    print()

    # Test each step
    tv_prod, tv_debug = test_tournament_value()
    tv_sim_prod, tv_sim_debug = test_tournament_value_from_sims()
    mp_prod, mp_debug = test_market_prediction()
    pc_prod, pc_debug = test_portfolio_construction()

    print("=" * 90)
    print("ALL TESTS PASSED")
    print("=" * 90)
    print()

    # Save artifacts to new location
    out_dir = Path("out/2025/derived/artifacts_v2")
    out_dir.mkdir(parents=True, exist_ok=True)

    tv_prod.to_parquet(out_dir / "tournament_value.parquet", index=False)
    tv_debug.to_parquet(out_dir / "tournament_value_debug.parquet", index=False)

    mp_prod.to_parquet(out_dir / "market_prediction.parquet", index=False)
    mp_debug.to_parquet(out_dir / "market_prediction_debug.parquet", index=False)

    pc_prod.to_parquet(out_dir / "recommended_bids.parquet", index=False)
    pc_debug.to_parquet(out_dir / "recommended_bids_debug.parquet", index=False)

    print(f"✓ Artifacts saved to {out_dir}")
    print()
    print("Use debug artifacts to inspect:")
    print(f"  - tournament_value_debug.parquet: Championship probabilities, variance")
    print(f"  - market_prediction_debug.parquet: All model features")
    print(f"  - recommended_bids_debug.parquet: ROI, ownership, expected returns")
