#!/usr/bin/env python3
"""
Register investment models to lab.investment_models table.

This script replaces the Go algorithm_registry.SyncToDatabase() function.
Run this script to ensure all baseline models are registered in the database.

Usage:
    python scripts/register_investment_models.py
"""

import logging
import sys
from typing import Any, Dict, List, NamedTuple

# moneyball must be installed: pip install -e . from the data-science directory
from moneyball.lab.models import get_or_create_investment_model

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s - %(levelname)s - %(message)s",
)
logger = logging.getLogger(__name__)


class ModelSpec(NamedTuple):
    """Specification for an investment model to register."""
    name: str
    kind: str
    params: Dict[str, Any]
    notes: str


# Investment models to register
# These correspond to market_share algorithms from the old Go registry
INVESTMENT_MODELS: List[ModelSpec] = [
    ModelSpec(
        name="ridge",
        kind="ridge",
        params={},
        notes="Ridge Regression baseline (Python runner)",
    ),
    ModelSpec(
        name="ridge-v1",
        kind="ridge",
        params={"feature_set": "optimal"},
        notes="Ridge Regression V1 with optimal feature set",
    ),
    ModelSpec(
        name="ridge-v1-recent",
        kind="ridge",
        params={"feature_set": "optimal"},
        # TODO: implement train_years_window in training pipeline before re-adding param
        notes="Ridge Regression V1 with recent training window (3 years)",
    ),
    ModelSpec(
        name="ridge-v2",
        kind="ridge",
        params={
            "feature_set": "optimal_v2",
            "seed_prior_k": 20.0,
            "program_prior_k": 50.0,
        },
        notes="Ridge Regression V2 with optimal_v2 feature set and stabilized priors",
    ),
    ModelSpec(
        name="ridge-v2-shrunk",
        kind="ridge",
        params={
            "feature_set": "optimal_v2",
            "seed_prior_monotone": True,
            "seed_prior_k": 20.0,
            "program_prior_k": 50.0,
        },
        notes="Ridge Regression V2 with shrunk priors (seed monotone + program)",
    ),
    ModelSpec(
        name="ridge-v2-underbid-1sigma",
        kind="ridge",
        params={
            "feature_set": "optimal_v2",
            "seed_prior_k": 20.0,
            "program_prior_k": 50.0,
        },
        notes="Ridge Regression V2 for underbid detection (1-sigma, seed/log, unscaled)",
    ),
    ModelSpec(
        name="ridge-v2-log",
        kind="ridge",
        params={
            "feature_set": "optimal_v2",
            "target_transform": "log",
            "seed_prior_monotone": True,
            "seed_prior_k": 20.0,
            "program_prior_k": 50.0,
        },
        notes="Ridge Regression V2 with log target transform and stabilized priors",
    ),
    ModelSpec(
        name="ridge-v3",
        kind="ridge",
        params={
            "feature_set": "optimal_v3",
        },
        notes="Ridge Regression V3 with analytical KenPom championship probabilities",
    ),
    ModelSpec(
        name="naive-ev-baseline",
        kind="naive_ev",
        params={},
        notes="Naive EV Baseline - bids proportional to expected value",
    ),
    ModelSpec(
        name="oracle-actual-market",
        kind="oracle",
        params={},
        notes="Oracle using actual historical market shares (for testing)",
    ),
]


def main():
    """Register all investment models."""
    logger.info("Registering investment models...")

    created = 0
    existing = 0

    for spec in INVESTMENT_MODELS:
        try:
            model, was_created = get_or_create_investment_model(
                name=spec.name,
                kind=spec.kind,
                params=spec.params,
                notes=spec.notes,
            )

            if was_created:
                logger.info(f"  {spec.name}: registered (id={model.id})")
                created += 1
            else:
                logger.info(f"  {spec.name}: already exists (id={model.id})")
                existing += 1

        except Exception as e:
            logger.error(f"  {spec.name}: FAILED - {e}")
            raise

    logger.info(f"Done: {created} created, {existing} already existed")
    return 0


if __name__ == "__main__":
    sys.exit(main())
