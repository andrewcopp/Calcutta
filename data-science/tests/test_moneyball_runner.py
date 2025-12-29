from __future__ import annotations

import unittest

from moneyball.pipeline.artifacts import (
    FileFingerprint,
    manifest_matches,
    sha256_jsonable,
)


class TestThatManifestMatchesIsCacheHit(unittest.TestCase):

    def test_manifest_matches_with_same_config(self) -> None:
        stage = "predicted_game_outcomes"
        stage_config = {"n_sims": 10, "seed": 1}

        games_fp = FileFingerprint(
            path="games.parquet",
            size=123,
            mtime_ns=1,
        )
        teams_fp = FileFingerprint(
            path="teams.parquet",
            size=456,
            mtime_ns=2,
        )
        inputs = {"games": games_fp, "teams": teams_fp}

        existing = {
            "stages": {
                stage: {
                    "stage_config_hash": "ignored_in_test",
                    "inputs": {
                        "games": {
                            "path": games_fp.path,
                            "size": games_fp.size,
                            "mtime_ns": games_fp.mtime_ns,
                        },
                        "teams": {
                            "path": teams_fp.path,
                            "size": teams_fp.size,
                            "mtime_ns": teams_fp.mtime_ns,
                        },
                    },
                }
            }
        }

        existing["stages"][stage]["stage_config_hash"] = sha256_jsonable(
            stage_config
        )

        out = manifest_matches(
            existing=existing,
            stage=stage,
            stage_config=stage_config,
            input_fingerprints=inputs,
        )

        self.assertTrue(out)


def test_that_cache_misses_when_config_changes(self) -> None:
    def test_cache_misses_when_config_changes(self) -> None:
        stage = "predicted_game_outcomes"
        stage_config_a = {"n_sims": 10, "seed": 1}
        stage_config_b = {"n_sims": 999, "seed": 1}

        games_fp = FileFingerprint(
            path="games.parquet",
            size=123,
            mtime_ns=1,
        )
        inputs = {"games": games_fp}

        existing = {
            "stages": {
                stage: {
                    "stage_config_hash": sha256_jsonable(stage_config_a),
                    "inputs": {
                        "games": {
                            "path": games_fp.path,
                            "size": games_fp.size,
                            "mtime_ns": games_fp.mtime_ns,
                        },
                    },
                }
            }
        }

        out = manifest_matches(
            existing=existing,
            stage=stage,
            stage_config=stage_config_b,
            input_fingerprints=inputs,
        )

        self.assertFalse(out)
