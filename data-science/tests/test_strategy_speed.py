"""Quick test to verify strategy comparison is fast with cached tournaments."""
import time
import subprocess
from pathlib import Path
import sys
import os
import unittest

_YEAR = 2017
_STRATEGY = "waterfill_equal"


def _run_speed_check() -> tuple[int, str, str, float]:
    start = time.time()
    cmd = [
        sys.executable,
        "-m",
        "moneyball.cli",
        "investment-report",
        f"out/{_YEAR}",
        "--snapshot-name",
        str(_YEAR),
        "--n-sims",
        "5000",
        "--seed",
        "123",
        "--budget-points",
        "100",
        "--strategy",
        _STRATEGY,
        "--no-include-upstream",
    ]

    env = dict(os.environ)
    root = str(Path(__file__).resolve().parents[1])
    env["PYTHONPATH"] = root + (
        (os.pathsep + env["PYTHONPATH"]) if env.get("PYTHONPATH") else ""
    )

    result = subprocess.run(
        cmd,
        env=env,
        capture_output=True,
        text=True,
        cwd=root,
    )

    elapsed = time.time() - start
    return result.returncode, result.stdout, result.stderr, elapsed


class TestThatStrategySpeedIsAcceptable(unittest.TestCase):
    @unittest.skipUnless(
        os.environ.get("RUN_STRATEGY_SPEED_TEST") == "1",
        "Set RUN_STRATEGY_SPEED_TEST=1 to enable this test",
    )
    def test_that_strategy_speed_is_acceptable(self) -> None:
        code, _out, err, elapsed = _run_speed_check()
        self.assertEqual(code, 0, msg=err)
        self.assertLess(elapsed, 10.0)


if __name__ == "__main__":
    print(f"Testing {_STRATEGY} for {_YEAR} with cached tournaments...")
    code, out, err, elapsed = _run_speed_check()
    if code == 0:
        print(f" Success in {elapsed:.1f} seconds")
        print("  (Should be <10 seconds with cached tournaments)")
    else:
        print(f" Failed after {elapsed:.1f} seconds")
        print(err[-500:])
