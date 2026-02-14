"""Tests for analytical tournament value computation."""
from __future__ import annotations

import unittest

import pandas as pd

from moneyball.models.analytical_tournament_value import (
    compute_analytical_tournament_values,
)
from moneyball.utils.points import set_default_points_by_win_index

# Standard March Madness scoring: 1, 2, 4, 8, 16, 32 per round
_DEFAULT_POINTS = {
    1: 1,   # Round of 64 win
    2: 2,   # Round of 32 win
    3: 4,   # Sweet 16 win
    4: 8,   # Elite 8 win
    5: 16,  # Final Four win
    6: 32,  # Championship win
    7: 0,   # No additional points for winning championship game twice
}


def setUpModule():
    """Set up default scoring rules before any tests run."""
    set_default_points_by_win_index(_DEFAULT_POINTS)


def _create_68_team_field() -> pd.DataFrame:
    """Create a realistic 68-team tournament field."""
    teams = []
    regions = ["East", "West", "South", "Midwest"]

    # Standard seeds 1-16 for each region (64 teams)
    # Assign KenPom ratings that decrease with seed
    for region in regions:
        for seed in range(1, 17):
            # KenPom net roughly correlates with seed
            # 1-seeds ~+25, 16-seeds ~-10
            kenpom_net = 25 - (seed - 1) * 2.2
            teams.append({
                "id": f"{region}-{seed}",
                "team_key": f"{region}-{seed}",
                "seed": seed,
                "region": region,
                "kenpom_net": kenpom_net,
            })

    # Add 4 First Four teams (extra 11-seeds and 16-seeds)
    # Two regions get an extra 11-seed, two get an extra 16-seed
    first_four = [
        ("East", 16, -12.0),
        ("West", 16, -11.0),
        ("South", 11, 5.0),
        ("Midwest", 11, 4.0),
    ]
    for region, seed, kenpom in first_four:
        teams.append({
            "id": f"{region}-{seed}-FF",
            "team_key": f"{region}-{seed}-FF",
            "seed": seed,
            "region": region,
            "kenpom_net": kenpom,
        })

    return pd.DataFrame(teams)


class TestThatChampionshipProbabilitiesSumToOne(unittest.TestCase):
    """Championship probabilities must sum to 1.0."""

    def test_that_championship_probabilities_sum_to_one(self) -> None:
        # GIVEN a 68-team tournament field
        teams = _create_68_team_field()

        # WHEN computing analytical tournament values
        result = compute_analytical_tournament_values(teams)

        # THEN championship probabilities sum to 1.0
        total_prob = result["analytical_p_championship"].sum()
        self.assertAlmostEqual(total_prob, 1.0, places=6)


class TestThatHigherKenPomMeansHigherChampionshipProbability(unittest.TestCase):
    """Teams with higher KenPom should have higher championship probability."""

    def test_that_one_seeds_have_highest_championship_probability(self) -> None:
        # GIVEN a 68-team tournament field
        teams = _create_68_team_field()

        # WHEN computing analytical tournament values
        result = compute_analytical_tournament_values(teams)

        # Merge seed info back
        result = result.merge(
            teams[["team_key", "seed"]],
            on="team_key",
            how="left",
        )

        # THEN 1-seeds have higher average championship probability than 2-seeds
        avg_by_seed = result.groupby("seed")["analytical_p_championship"].mean()
        self.assertGreater(avg_by_seed[1], avg_by_seed[2])

    def test_that_sixteen_seeds_have_lowest_championship_probability(self) -> None:
        # GIVEN a 68-team tournament field
        teams = _create_68_team_field()

        # WHEN computing analytical tournament values
        result = compute_analytical_tournament_values(teams)

        # Merge seed info back
        result = result.merge(
            teams[["team_key", "seed"]],
            on="team_key",
            how="left",
        )

        # THEN 16-seeds have lower average championship probability than 15-seeds
        avg_by_seed = result.groupby("seed")["analytical_p_championship"].mean()
        self.assertLess(avg_by_seed[16], avg_by_seed[15])


class TestThatExpectedPointsArePositive(unittest.TestCase):
    """Expected points should be positive for all teams."""

    def test_that_expected_points_are_positive(self) -> None:
        # GIVEN a 68-team tournament field
        teams = _create_68_team_field()

        # WHEN computing analytical tournament values
        result = compute_analytical_tournament_values(teams)

        # THEN all expected points are positive
        self.assertTrue((result["analytical_expected_points"] > 0).all())


class TestThatExpectedPointsAreReasonable(unittest.TestCase):
    """Expected points should be within reasonable bounds."""

    def test_that_expected_points_are_in_reasonable_range(self) -> None:
        # GIVEN a 68-team tournament field
        teams = _create_68_team_field()

        # WHEN computing analytical tournament values
        result = compute_analytical_tournament_values(teams)

        # THEN expected points are between 0 and max possible (63 for champion)
        max_points = result["analytical_expected_points"].max()
        min_points = result["analytical_expected_points"].min()

        self.assertLess(max_points, 63.0)  # Less than perfect champion
        self.assertGreater(min_points, 0.0)  # At least some expected points


class TestThatOneSeedsHaveHigherExpectedPoints(unittest.TestCase):
    """1-seeds should have higher expected points than lower seeds."""

    def test_that_one_seeds_have_higher_expected_points_than_sixteen_seeds(self) -> None:
        # GIVEN a 68-team tournament field
        teams = _create_68_team_field()

        # WHEN computing analytical tournament values
        result = compute_analytical_tournament_values(teams)

        # Merge seed info back
        result = result.merge(
            teams[["team_key", "seed"]],
            on="team_key",
            how="left",
        )

        # THEN 1-seeds have higher average expected points than 16-seeds
        avg_by_seed = result.groupby("seed")["analytical_expected_points"].mean()
        self.assertGreater(avg_by_seed[1], avg_by_seed[16])


class TestThatKenPomScaleAffectsResults(unittest.TestCase):
    """Different KenPom scales should produce different probabilities."""

    def test_that_larger_scale_reduces_favorite_advantage(self) -> None:
        # GIVEN a 68-team tournament field
        teams = _create_68_team_field()

        # WHEN computing with different scales
        result_small = compute_analytical_tournament_values(teams, kenpom_scale=5.0)
        result_large = compute_analytical_tournament_values(teams, kenpom_scale=20.0)

        # Merge seed info
        result_small = result_small.merge(
            teams[["team_key", "seed"]], on="team_key", how="left"
        )
        result_large = result_large.merge(
            teams[["team_key", "seed"]], on="team_key", how="left"
        )

        # THEN with larger scale, 1-seeds have lower championship probability
        # (more randomness = less predictable = lower favorites' edge)
        small_1seed = result_small[result_small["seed"] == 1][
            "analytical_p_championship"
        ].mean()
        large_1seed = result_large[result_large["seed"] == 1][
            "analytical_p_championship"
        ].mean()

        self.assertGreater(small_1seed, large_1seed)


if __name__ == "__main__":
    unittest.main()
