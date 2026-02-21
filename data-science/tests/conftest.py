"""Shared pytest fixtures for data-science tests."""

from __future__ import annotations

import pytest
import pandas as pd


@pytest.fixture()
def basic_train_dataset() -> pd.DataFrame:
    """A minimal 5-team training dataset with seed, region, kenpom_net, and target."""
    return pd.DataFrame(
        {
            "seed": [1, 2, 3, 4, 5],
            "region": ["East", "East", "West", "West", "South"],
            "kenpom_net": [30.0, 20.0, 10.0, 5.0, 15.0],
            "observed_team_share_of_pool": [0.3, 0.25, 0.15, 0.1, 0.2],
        }
    )


@pytest.fixture()
def basic_predict_dataset() -> pd.DataFrame:
    """A minimal 3-team prediction dataset with seed, region, and kenpom_net."""
    return pd.DataFrame(
        {
            "seed": [1, 10, 16],
            "region": ["East", "West", "South"],
            "kenpom_net": [35.0, 0.0, -5.0],
        }
    )


@pytest.fixture()
def extended_train_dataset() -> pd.DataFrame:
    """A 5-team training dataset with all KenPom columns needed for optimal_v2."""
    return pd.DataFrame(
        {
            "seed": [1, 2, 3, 4, 5],
            "region": ["East", "East", "West", "West", "South"],
            "kenpom_net": [30.0, 20.0, 10.0, 5.0, 15.0],
            "kenpom_o": [120.0, 115.0, 110.0, 108.0, 112.0],
            "kenpom_d": [95.0, 98.0, 100.0, 102.0, 99.0],
            "kenpom_adj_t": [70.0, 68.0, 66.0, 65.0, 67.0],
            "observed_team_share_of_pool": [0.3, 0.25, 0.15, 0.1, 0.2],
            "school_slug": ["duke", "unc", "kentucky", "kansas", "gonzaga"],
        }
    )


@pytest.fixture()
def extended_predict_dataset() -> pd.DataFrame:
    """A 3-team prediction dataset with all KenPom columns needed for optimal_v2."""
    return pd.DataFrame(
        {
            "seed": [1, 10, 16],
            "region": ["East", "West", "South"],
            "kenpom_net": [35.0, 0.0, -5.0],
            "kenpom_o": [125.0, 105.0, 95.0],
            "kenpom_d": [90.0, 105.0, 110.0],
            "kenpom_adj_t": [72.0, 64.0, 60.0],
            "school_slug": ["duke", "team10", "team16"],
        }
    )


def create_68_team_field() -> pd.DataFrame:
    """Create a 68-team field with all required columns for auction prediction.

    This is a plain function (not a fixture) so it can be called multiple times
    within a single test to create independent train/predict copies.
    """
    teams = []
    regions = ["East", "West", "South", "Midwest"]
    fake_tournament_id = "00000000-0000-0000-0000-000000000000"

    for region in regions:
        for seed in range(1, 17):
            kenpom_net = 25 - (seed - 1) * 2.2
            champ_prob = max(0.001, 0.20 - (seed - 1) * 0.013)
            expected_pts = max(0.5, 12.0 - (seed - 1) * 0.75)
            teams.append(
                {
                    "id": f"{region}-{seed}",
                    "team_key": f"{region}-{seed}",
                    "seed": seed,
                    "region": region,
                    "kenpom_net": kenpom_net,
                    "kenpom_o": 110 + kenpom_net * 0.3,
                    "kenpom_d": 100 - kenpom_net * 0.2,
                    "kenpom_adj_t": 68 + seed * 0.1,
                    "school_slug": f"school-{region.lower()}-{seed}",
                    "tournament_id": fake_tournament_id,
                    "predicted_p_championship": champ_prob,
                    "predicted_expected_points": expected_pts,
                }
            )

    first_four = [
        ("East", 16, -12.0),
        ("West", 16, -11.0),
        ("South", 11, 5.0),
        ("Midwest", 11, 4.0),
    ]
    for region, seed, kenpom in first_four:
        champ_prob = max(0.001, 0.20 - (seed - 1) * 0.013)
        expected_pts = max(0.5, 12.0 - (seed - 1) * 0.75)
        teams.append(
            {
                "id": f"{region}-{seed}-FF",
                "team_key": f"{region}-{seed}-FF",
                "seed": seed,
                "region": region,
                "kenpom_net": kenpom,
                "kenpom_o": 110 + kenpom * 0.3,
                "kenpom_d": 100 - kenpom * 0.2,
                "kenpom_adj_t": 68 + seed * 0.1,
                "school_slug": f"school-{region.lower()}-{seed}-ff",
                "tournament_id": fake_tournament_id,
                "predicted_p_championship": champ_prob,
                "predicted_expected_points": expected_pts,
            }
        )

    return pd.DataFrame(teams)


SEED_MARKET_SHARES: dict[int, float] = {
    1: 0.08,
    2: 0.06,
    3: 0.05,
    4: 0.04,
    5: 0.03,
    6: 0.025,
    7: 0.02,
    8: 0.018,
    9: 0.015,
    10: 0.012,
    11: 0.01,
    12: 0.008,
    13: 0.005,
    14: 0.003,
    15: 0.002,
    16: 0.001,
}
