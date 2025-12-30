"""Quick test to verify strategy comparison is fast with cached tournaments."""
import time
import subprocess
from pathlib import Path

year = 2017
strategy = "waterfill_equal"

print(f"Testing {strategy} for {year} with cached tournaments...")
start = time.time()

cmd = [
    "python", "-m", "moneyball.cli",
    "investment-report",
    f"out/{year}",
    "--snapshot-name", str(year),
    "--n-sims", "5000",
    "--seed", "123",
    "--budget-points", "100",
    "--strategy", strategy,
    "--no-include-upstream",
]

result = subprocess.run(
    cmd,
    capture_output=True,
    text=True,
    cwd=Path(__file__).parent,
)

elapsed = time.time() - start

if result.returncode == 0:
    print(f"✓ Success in {elapsed:.1f} seconds")
    print(f"  (Should be <10 seconds with cached tournaments)")
else:
    print(f"✗ Failed after {elapsed:.1f} seconds")
    print(result.stderr[-500:])
