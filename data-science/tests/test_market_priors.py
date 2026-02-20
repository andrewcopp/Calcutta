"""Unit tests for moneyball.models.market_priors helpers."""

import unittest

import pandas as pd

from moneyball.models.market_priors import (
    compute_program_priors,
    compute_seed_priors,
)


def _make_train_data() -> pd.DataFrame:
    """Create minimal training data with seeds 1-5 and market shares."""
    return pd.DataFrame({
        "seed": [1, 2, 3, 4, 5],
        "team_share_of_pool": [0.30, 0.25, 0.20, 0.15, 0.10],
        "school_slug": ["duke", "north-carolina", "kentucky", "kansas", "gonzaga"],
    })


class TestThatComputeSeedPriorsCoversAllSeeds(unittest.TestCase):
    """Tests for compute_seed_priors."""

    def test_that_output_has_seeds_1_through_16(self) -> None:
        # GIVEN training data
        train = _make_train_data()

        # WHEN computing seed priors
        priors = compute_seed_priors(train, seed_prior_k=10.0)

        # THEN all 16 seeds are present
        self.assertEqual(set(priors.keys()), set(range(1, 17)))

    def test_that_priors_are_non_negative(self) -> None:
        # GIVEN training data
        train = _make_train_data()

        # WHEN computing seed priors
        priors = compute_seed_priors(train, seed_prior_k=10.0)

        # THEN all values are non-negative
        for seed, val in priors.items():
            self.assertGreaterEqual(val, 0.0, f"seed {seed} has negative prior")

    def test_that_monotone_enforcement_makes_priors_non_increasing(self) -> None:
        # GIVEN training data
        train = _make_train_data()

        # WHEN computing seed priors with monotone enforcement
        priors = compute_seed_priors(
            train, seed_prior_k=10.0, seed_prior_monotone=True
        )

        # THEN priors are monotonically non-increasing
        vals = [priors[s] for s in range(1, 17)]
        for i in range(len(vals) - 1):
            self.assertGreaterEqual(vals[i], vals[i + 1])

    def test_that_zero_k_uses_raw_means(self) -> None:
        # GIVEN training data with known seed=1 share of 0.30
        train = _make_train_data()

        # WHEN computing seed priors without shrinkage
        priors = compute_seed_priors(
            train, seed_prior_k=0.0, seed_prior_monotone=False
        )

        # THEN seed 1 prior equals the raw mean (0.30)
        self.assertAlmostEqual(priors[1], 0.30, places=6)


class TestThatComputeProgramPriorsIsCorrect(unittest.TestCase):
    """Tests for compute_program_priors."""

    def test_that_programs_are_returned_when_school_slug_exists(self) -> None:
        # GIVEN training data with school_slug
        train = _make_train_data()

        # WHEN computing program priors
        means, counts = compute_program_priors(train, program_prior_k=10.0)

        # THEN both dicts are not None
        self.assertIsNotNone(means)
        self.assertIsNotNone(counts)

    def test_that_returns_none_when_school_slug_missing(self) -> None:
        # GIVEN training data without school_slug
        train = _make_train_data().drop(columns=["school_slug"])

        # WHEN computing program priors
        means, counts = compute_program_priors(train, program_prior_k=10.0)

        # THEN both are None
        self.assertIsNone(means)
        self.assertIsNone(counts)

    def test_that_counts_match_number_of_observations(self) -> None:
        # GIVEN training data with one observation per program
        train = _make_train_data()

        # WHEN computing program priors
        _, counts = compute_program_priors(train, program_prior_k=0.0)

        # THEN each program has count 1
        for slug, count in counts.items():
            self.assertEqual(count, 1, f"slug '{slug}' has count {count}")

    def test_that_zero_k_uses_raw_means(self) -> None:
        # GIVEN training data with duke at 0.30
        train = _make_train_data()

        # WHEN computing program priors without shrinkage
        means, _ = compute_program_priors(train, program_prior_k=0.0)

        # THEN duke's mean matches raw value
        self.assertAlmostEqual(means["duke"], 0.30, places=6)


if __name__ == "__main__":
    unittest.main()
