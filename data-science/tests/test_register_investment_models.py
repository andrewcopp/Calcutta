"""
Unit tests for scripts/register_investment_models.py.

Validates the INVESTMENT_MODELS registry data structure without
requiring a database connection.
"""

import unittest
import sys
from pathlib import Path

# The script uses sys.path manipulation; replicate it so we can import the module
_PROJECT_ROOT = str(Path(__file__).resolve().parents[1])
if _PROJECT_ROOT not in sys.path:
    sys.path.insert(0, _PROJECT_ROOT)

from scripts.register_investment_models import INVESTMENT_MODELS


class TestThatInvestmentModelRegistryHasNoDuplicateNames(unittest.TestCase):
    """Test that the INVESTMENT_MODELS list has unique model names."""

    def test_that_no_duplicate_model_names_exist(self) -> None:
        """Every model name in INVESTMENT_MODELS must be unique."""
        # GIVEN the list of model specs
        names = [spec.name for spec in INVESTMENT_MODELS]

        # WHEN checking for duplicates
        duplicates = [name for name in names if names.count(name) > 1]

        # THEN there should be no duplicates
        self.assertEqual(
            len(set(duplicates)),
            0,
            f"Duplicate model names found: {set(duplicates)}",
        )

    def test_that_model_names_are_non_empty(self) -> None:
        """Every model name must be a non-empty string."""
        # GIVEN the list of model specs
        for spec in INVESTMENT_MODELS:
            # THEN each name is a non-empty string
            self.assertIsInstance(spec.name, str)
            self.assertGreater(
                len(spec.name),
                0,
                f"Model spec has empty name: {spec}",
            )

    def test_that_model_kinds_are_non_empty(self) -> None:
        """Every model kind must be a non-empty string."""
        # GIVEN the list of model specs
        for spec in INVESTMENT_MODELS:
            # THEN each kind is a non-empty string
            self.assertIsInstance(spec.kind, str)
            self.assertGreater(
                len(spec.kind),
                0,
                f"Model spec '{spec.name}' has empty kind",
            )

    def test_that_at_least_one_model_is_registered(self) -> None:
        """The registry must contain at least one model."""
        # GIVEN the list of model specs
        # THEN it is non-empty
        self.assertGreater(
            len(INVESTMENT_MODELS),
            0,
            "INVESTMENT_MODELS is empty -- no models would be registered",
        )


if __name__ == "__main__":
    unittest.main()
